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
	pf "github.com/horm-database/common/proto/filter"
	"github.com/horm-database/common/types"
	"github.com/horm-database/common/util"
	"github.com/horm-database/orm/database"
	"github.com/horm-database/orm/obj"
	"github.com/horm-database/server/auth"
	"github.com/horm-database/server/consts"
	"github.com/horm-database/server/filter"
	"github.com/horm-database/server/model/table"
	ut "github.com/horm-database/server/util"

	"github.com/barkimedes/go-deepcopy"
)

// 节点查询
func query(ctx context.Context, appid uint64,
	node *obj.Tree) (result interface{}, detail *proto.Detail, isNil bool, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = ut.LogErrorf(ctx, errs.RetPanic, "query panic: %v", e)
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

	req.Type = unit.Type
	req.Scroll = unit.Scroll

	req.Prefix = unit.Prefix
	req.Key = key
	req.Args = args
	req.Bytes = unit.Bytes

	req.Params = unit.Params
	req.Query = unit.Query

	defer func() {
		deferFiltersHandle(ctx, req, rsp, appid, unit.Extend, tblTable)
	}()

	req.Data, err = util.FormatData(data, unit.DataType)
	if err != nil {
		err = errs.Newf(errs.RetFormatDataError, "[%s] format insert data error: %v", realNode.GetPath(), err)
		return
	}

	req.Datas, err = util.FormatDatas(datas, unit.DataType)
	if err != nil {
		err = errs.Newf(errs.RetFormatDataError, "[%s] format insert datas error: %v", realNode.GetPath(), err)
		return
	}

	var response bool
	response, result, detail, isNil, err = filtersHandle(ctx, req, rsp, appid, unit.Extend, tblTable, consts.PreFilter)
	if response {
		return result, detail, isNil, err
	}

	// change db address by filter
	addr := dbInfo.Addr
	if req.Addr != nil && req.Addr.Address != "" {
		err = util.ParseConnFromAddress(req.Addr)
		if err != nil {
			err = errs.Newf(errs.RetDBAddressParseError,
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

		log.Info(ctx, "change db address by filter: from [%s.%s %s:%s/%s] to [%s.%s %s:%s/%s]",
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

	response, result, detail, isNil, err = filtersHandle(ctx, req, rsp, appid, unit.Extend, tblTable, consts.PostFilter)
	if response {
		return result, detail, isNil, err
	}

	return rsp.Result, rsp.Detail, rsp.IsNil, rsp.Error
}

func filtersHandle(ctx context.Context, req *pf.Request, resp *pf.Response, appid uint64, extend types.Map,
	tblTable *obj.TblTable, typ int8) (response bool, result interface{}, detail *proto.Detail, isNil bool, err error) {
	tableFilters := table.GetTableFilters(tblTable.Id, typ)

	for _, tableFilter := range tableFilters {
		tblFilter := table.GetFilter(tableFilter.FilterId)

		if tblFilter == nil {
			e := errs.Newf(errs.RetFilterNotFind, "not find filter : %d", tableFilter.FilterId)
			if tableFilter.ScheduleConf.SkipError {
				log.Error(ctx, errs.RetFilterNotFind, e.Error())
				continue
			} else {
				return true, nil, nil, false, e
			}
		}

		funcName := fmt.Sprintf("%s_%d", tblFilter.Func, tableFilter.FilterVersion)
		f := filter.Func[funcName]

		if f == nil {
			e := errs.Newf(errs.RetFilterFuncNotRegister, "filter %s`s function %s "+
				"version %d not register", tblFilter.Name, tblFilter.Func, tableFilter.FilterVersion)

			if tableFilter.ScheduleConf.SkipError {
				log.Error(ctx, errs.RetFilterFuncNotRegister, e.Error())
				continue
			} else {
				return true, nil, nil, false, e
			}
		}

		if tableFilter.ScheduleConf.Async { // 异步执行
			go filterAsyncHandle(ctx, req, resp, extend, f, tableFilter)
			continue
		}

		isResponse, e := filterHandle(ctx, req, resp, extend, tableFilter, f)
		if err != nil {
			if tableFilter.ScheduleConf.SkipError {
				log.Error(ctx, errs.RetFilterNotFind, e.Error())
			} else {
				return true, nil, nil, false, getFilterError(err)
			}
		}

		if isResponse {
			return true, resp.Result, resp.Detail, resp.IsNil, getFilterError(resp.Error)
		}
	}

	return false, nil, nil, false, nil
}

func deferFiltersHandle(ctx context.Context,
	req *pf.Request, resp *pf.Response, appid uint64, extend types.Map, tblTable *obj.TblTable) {
	tableFilters := table.GetTableFilters(tblTable.Id, consts.DeferFilter)

	if tableFilters == nil {
		return
	}

	for _, tableFilter := range tableFilters {
		tblFilter := table.GetFilter(tableFilter.FilterId)

		if tblFilter == nil {
			log.Error(ctx, errs.RetFilterNotFind,
				errs.Newf(errs.RetFilterNotFind, "not find filter : %d", tableFilter.FilterId).Error())
			continue
		}

		funcName := fmt.Sprintf("%s_%d", tblFilter.Func, tableFilter.FilterVersion)
		f := filter.Func[funcName]
		if f == nil {
			log.Error(ctx, errs.RetFilterFuncNotRegister,
				errs.Newf(errs.RetFilterFuncNotRegister, "filter %s`s function %s version %d not register",
					tblFilter.Name, tblFilter.Func, tableFilter.FilterVersion).Error())
			continue
		}

		if tableFilter.ScheduleConf.Async { // 异步执行
			go filterAsyncHandle(ctx, req, resp, extend, f, tableFilter)
			continue
		}

		_, err := filterHandle(ctx, req, resp, extend, tableFilter, f)
		if err != nil {
			log.Error(ctx, errs.RetFilterNotFind, err.Error())
		}
	}
}

func filterAsyncHandle(ctx context.Context, req *pf.Request, resp *pf.Response,
	extend types.Map, f filter.Filter, tableFilter *table.TblTableFilter) {
	asyncCtx, cancel, _ := codec.NewAsyncMessage(ctx,
		time.Duration(tableFilter.ScheduleConf.Timeout)*time.Millisecond)

	defer cancel()

	// 异步处理不允许改变入参
	iReq, err := deepcopy.Anything(req)
	if err != nil || iReq == nil {
		log.Error(asyncCtx, errs.RetFilterParamCopy, "filter async handle deep copy request error: %v", err)
		return
	}

	iResp, err := deepcopy.Anything(req)
	if err != nil || iResp == nil {
		log.Error(asyncCtx, errs.RetFilterParamCopy, "filter async handle deep copy response error: %v", err)
		return
	}

	iExtend, err := deepcopy.Anything(extend)
	if err != nil {
		log.Error(asyncCtx, errs.RetFilterParamCopy, "filter async handle deep copy extend error: %v", err)
		return
	}

	_, err = filterHandle(asyncCtx, iReq.(*pf.Request), iResp.(*pf.Response), iExtend.(types.Map), tableFilter, f)
	if err != nil {
		e := getFilterError(err)
		log.Error(asyncCtx, errs.Code(e), e.Error())
	}

	if resp.Error != nil {
		e := getFilterError(resp.Error)
		log.Error(asyncCtx, errs.Code(e), e.Error())
	}
}

func filterHandle(ctx context.Context, req *pf.Request, resp *pf.Response,
	extend types.Map, tableFilter *table.TblTableFilter, f filter.Filter) (response bool, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = ut.LogErrorf(ctx, errs.RetPanic,
				"filter handle panic: [%v], req=%s, resp=%s, extend=%s ",
				e, json.MarshalToString(req),
				json.MarshalToString(resp),
				json.MarshalToString(extend))
		}
	}()

	return f.Handle(ctx, req, resp, extend, tableFilter.Conf)
}

func getFilterError(err error) error {
	if err == nil {
		return nil
	}

	e := errs.Error{
		Type: errs.ErrorTypeFilter,
		Code: errs.Code(err),
		Msg:  errs.Msg(err),
	}

	if e.Code == errs.RetUnknown {
		e.Code = errs.RetFilterHandle
	}

	return &e
}
