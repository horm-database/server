package logic

import (
	"context"
	"fmt"
	"time"

	"github.com/horm-database/common/codec"
	cc "github.com/horm-database/common/consts"
	"github.com/horm-database/common/errs"
	"github.com/horm-database/common/json"
	"github.com/horm-database/common/log"
	"github.com/horm-database/common/proto"
	pf "github.com/horm-database/common/proto/plugin"
	"github.com/horm-database/common/types"
	"github.com/horm-database/common/util"
	"github.com/horm-database/orm/database"
	"github.com/horm-database/orm/obj"
	"github.com/horm-database/server/auth"
	"github.com/horm-database/server/consts"
	"github.com/horm-database/server/model/table"
	"github.com/horm-database/server/plugin"
	ut "github.com/horm-database/server/util"

	"github.com/barkimedes/go-deepcopy"
)

// 节点查询
func query(ctx context.Context, appid uint64,
	node *obj.Tree) (result interface{}, detail *proto.Detail, isNil bool, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = ut.LogErrorf(ctx, errs.ErrPanic, "query panic: %v", e)
			return
		}
	}()

	realNode := node.GetReal()
	op := realNode.GetOp()
	unit := realNode.GetUnit()
	dbInfo := realNode.GetDB()
	tblTable := realNode.GetTable()

	// 查看表权限
	err = auth.PermissionCheck(realNode, appid, op, unit.Query, false)
	if err != nil {
		return
	}

	// 引用处理
	where, having, data, datas, key, args, isNil, err := referHandle(dbInfo.Addr.Type, unit, node)
	if err != nil || isNil {
		return nil, nil, isNil, err
	}

	// 插件请求/返回参数初始化
	req := &pf.Request{}
	rsp := &pf.Response{}

	req.Op = op
	req.Tables = realNode.Tables()
	req.Where = where
	req.Column = unit.Column
	req.Group = unit.Group
	req.Having = having
	req.Order = unit.Order
	req.Page = unit.Page
	req.Size = unit.Size
	req.From = unit.From

	req.Join = unit.Join

	req.Type = unit.Type
	req.Scroll = unit.Scroll

	req.Prefix = unit.Prefix
	req.Key = key
	req.Args = args
	req.Bytes = unit.Bytes

	req.Params = unit.Params
	req.Query = unit.Query

	defer func() {
		deferPluginsHandle(ctx, req, rsp, appid, unit.Extend, tblTable)
	}()

	req.Data, err = util.FormatData(data, unit.DataType)
	if err != nil {
		err = errs.Newf(errs.ErrFormatData, "[%s] format insert data error: %v", realNode.GetPath(), err)
		return
	}

	req.Datas, err = util.FormatDatas(datas, unit.DataType)
	if err != nil {
		err = errs.Newf(errs.ErrFormatData, "[%s] format insert datas error: %v", realNode.GetPath(), err)
		return
	}

	var response bool
	response, result, detail, isNil, err = pluginsHandle(ctx, req, rsp, appid, unit.Extend, tblTable, consts.PrePlugin)
	if response {
		return result, detail, isNil, err
	}

	// change db address by plugin
	addr := dbInfo.Addr
	if req.Addr != nil && req.Addr.Address != "" {
		err = util.ParseConnFromAddress(req.Addr)
		if err != nil {
			err = errs.Newf(errs.ErrDBAddressParse,
				"[%s] db address [%s] parse error: %v", realNode.GetPath(), req.Addr.Address, err)
			return
		}

		if req.Addr.WriteTimeout == 0 {
			req.Addr.WriteTimeout = addr.WriteTimeout
		}

		if req.Addr.ReadTimeout == 0 {
			req.Addr.ReadTimeout = addr.ReadTimeout
		}

		if req.Addr.WarnTimeout == 0 {
			req.Addr.WarnTimeout = addr.WarnTimeout
		}

		log.Info(ctx, "change db address by plugin: from [%s.%s %s:%s/%s] to [%s.%s %s:%s/%s]",
			cc.DBTypeDesc[addr.Type], addr.Version, addr.Network, addr.Conn.Target, addr.Conn.DB,
			cc.DBTypeDesc[req.Addr.Type], req.Addr.Version, req.Addr.Network, req.Addr.Conn.Target, req.Addr.Conn.DB)
		addr = req.Addr
	}

	// 校验是否有执行权限
	if req.Op != op || req.Query != unit.Query {
		err = auth.PermissionCheck(realNode, appid, req.Op, req.Query, true)
		if err != nil {
			return
		}
	}

	// 校验表
	err = auth.TableVerify(realNode, appid, req.Tables, tblTable.TableVerify)
	if err != nil {
		return
	}

	// 走 db 查询
	result, detail, isNil, err = database.QueryResult(ctx, req, realNode, addr, node.TransInfo)

	rsp.IsNil = isNil
	rsp.Detail = detail
	rsp.Result = result
	rsp.Error = err

	response, result, detail, isNil, err = pluginsHandle(ctx, req, rsp, appid, unit.Extend, tblTable, consts.PostPlugin)
	if response {
		return result, detail, isNil, err
	}

	return rsp.Result, rsp.Detail, rsp.IsNil, rsp.Error
}

// 插件编排处理
func pluginsHandle(ctx context.Context, req *pf.Request, resp *pf.Response, appid uint64, extend types.Map,
	tblTable *obj.TblTable, typ int8) (response bool, result interface{}, detail *proto.Detail, isNil bool, err error) {
	tablePlugins := table.GetTablePlugins(tblTable.Id, typ)

	for _, tablePlugin := range tablePlugins {
		tblPlugin := table.GetPlugin(tablePlugin.PluginID)

		if tblPlugin == nil {
			e := errs.Newf(errs.ErrPluginNotFound, "not find plugin : %d", tablePlugin.PluginID)
			if tablePlugin.ScheduleConf.SkipError {
				log.Error(ctx, errs.ErrPluginNotFound, e.Error())
				continue
			} else {
				return true, nil, nil, false, e
			}
		}

		funcName := fmt.Sprintf("%s_%d", tblPlugin.Func, tablePlugin.PluginVersion)
		f := plugin.Func[funcName]

		if f == nil {
			e := errs.Newf(errs.ErrPluginFuncNotRegister, "plugin %s`s function %s "+
				"version %d not register", tblPlugin.Name, tblPlugin.Func, tablePlugin.PluginVersion)

			if tablePlugin.ScheduleConf.SkipError {
				log.Error(ctx, errs.ErrPluginFuncNotRegister, e.Error())
				continue
			} else {
				return true, nil, nil, false, e
			}
		}

		if tablePlugin.ScheduleConf.Async { // 异步执行
			go pluginAsyncHandle(ctx, req, resp, extend, f, tablePlugin)
			continue
		}

		isResponse, e := pluginHandle(ctx, req, resp, extend, tablePlugin, f)
		if e != nil {
			if tablePlugin.ScheduleConf.SkipError {
				log.Error(ctx, errs.ErrPluginNotFound, e.Error())
			} else {
				return true, nil, nil, false, getPluginError(e)
			}
		}

		if isResponse {
			return true, resp.Result, resp.Detail, resp.IsNil, getPluginError(resp.Error)
		}
	}

	return false, nil, nil, false, nil
}

func deferPluginsHandle(ctx context.Context,
	req *pf.Request, resp *pf.Response, appid uint64, extend types.Map, tblTable *obj.TblTable) {
	tablePlugins := table.GetTablePlugins(tblTable.Id, consts.DeferPlugin)

	if tablePlugins == nil {
		return
	}

	for _, tablePlugin := range tablePlugins {
		tblPlugin := table.GetPlugin(tablePlugin.PluginID)

		if tblPlugin == nil {
			log.Error(ctx, errs.ErrPluginNotFound,
				errs.Newf(errs.ErrPluginNotFound, "not find plugin : %d", tablePlugin.PluginID).Error())
			continue
		}

		funcName := fmt.Sprintf("%s_%d", tblPlugin.Func, tablePlugin.PluginVersion)
		f := plugin.Func[funcName]
		if f == nil {
			log.Error(ctx, errs.ErrPluginFuncNotRegister,
				errs.Newf(errs.ErrPluginFuncNotRegister, "plugin %s`s function %s version %d not register",
					tblPlugin.Name, tblPlugin.Func, tablePlugin.PluginVersion).Error())
			continue
		}

		if tablePlugin.ScheduleConf.Async { // 异步执行
			go pluginAsyncHandle(ctx, req, resp, extend, f, tablePlugin)
			continue
		}

		_, err := pluginHandle(ctx, req, resp, extend, tablePlugin, f)
		if err != nil {
			log.Error(ctx, errs.ErrPluginNotFound, err.Error())
		}
	}
}

func pluginAsyncHandle(ctx context.Context, req *pf.Request, resp *pf.Response,
	extend types.Map, f plugin.Plugin, tablePlugin *table.TblTablePlugin) {
	asyncCtx, cancel, _ := codec.NewAsyncMessage(ctx,
		time.Duration(tablePlugin.ScheduleConf.Timeout)*time.Millisecond)

	defer cancel()

	// 异步处理不允许改变入参
	iReq, err := deepcopy.Anything(req)
	if err != nil || iReq == nil {
		log.Error(asyncCtx, errs.ErrPluginParamCopy, "plugin async handle deep copy request error: %v", err)
		return
	}

	iResp, err := deepcopy.Anything(resp)
	if err != nil || iResp == nil {
		log.Error(asyncCtx, errs.ErrPluginParamCopy, "plugin async handle deep copy response error: %v", err)
		return
	}

	iExtend, err := deepcopy.Anything(extend)
	if err != nil {
		log.Error(asyncCtx, errs.ErrPluginParamCopy, "plugin async handle deep copy extend error: %v", err)
		return
	}

	_, err = pluginHandle(asyncCtx, iReq.(*pf.Request), iResp.(*pf.Response), iExtend.(types.Map), tablePlugin, f)
	if err != nil {
		e := getPluginError(err)
		log.Error(asyncCtx, errs.Code(e), e.Error())
	}

	if resp.Error != nil {
		e := getPluginError(resp.Error)
		log.Error(asyncCtx, errs.Code(e), e.Error())
	}
}

func pluginHandle(ctx context.Context, req *pf.Request, resp *pf.Response,
	extend types.Map, tablePlugin *table.TblTablePlugin, f plugin.Plugin) (response bool, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = ut.LogErrorf(ctx, errs.ErrPanic,
				"plugin handle panic: [%v], req=%s, resp=%s, extend=%s ",
				e, json.MarshalToString(req),
				json.MarshalToString(resp),
				json.MarshalToString(extend))
		}
	}()

	return f.Handle(ctx, req, resp, extend, tablePlugin.Conf)
}

func getPluginError(err error) error {
	if err == nil {
		return nil
	}

	e := errs.Error{
		Type: errs.ETypePlugin,
		Code: errs.Code(err),
		Msg:  errs.Msg(err),
	}

	if e.Code == errs.ErrUnknown {
		e.Code = errs.ErrPluginExec
	}

	return &e
}
