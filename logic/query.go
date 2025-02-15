// Copyright (c) 2024 The horm-database Authors. All rights reserved.
// This file Author:  CaoHao <18500482693@163.com> .
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logic

import (
	"context"
	"fmt"
	"time"

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
	"github.com/horm-database/server/model/table"
	"github.com/horm-database/server/plugin"
	"github.com/horm-database/server/plugin/conf"
)

// 节点查询
func query(ctx context.Context, appid uint64,
	node *obj.Tree) (result interface{}, detail *proto.Detail, isNil bool, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = errs.Newf(errs.ErrPanic, "query panic: %v", e)
			log.Error(ctx, errs.ErrPanic, err.Error())
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
	where, having, data, datas, key, field, val, args, isNil, err := referHandle(dbInfo.Addr.Type, unit, node)
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
	req.Field = field
	req.Bytes = unit.Bytes
	req.Val = val

	req.Params = unit.Params

	req.Query = unit.Query
	req.Args = args

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

	// 获取插件执行链
	chain, err := getPluginChain(ctx, appid, tblTable)
	if err != nil {
		return
	}

	dbExecFilter := func(ctx context.Context) error {
		// 校验是否有执行权限
		if req.Op != op || req.Query != unit.Query {
			err = auth.PermissionCheck(realNode, appid, req.Op, req.Query, true)
			if err != nil {
				return err
			}
		}

		// 校验表
		err = auth.TableVerify(realNode, appid, req.Tables, tblTable.TableVerify)
		if err != nil {
			return err
		}

		// 走 db 查询
		result, detail, isNil, err = database.QueryResult(ctx, req, realNode, dbInfo.Addr, node.TransInfo)

		rsp.IsNil = isNil
		rsp.Detail = detail
		rsp.Result = result
		rsp.Error = err
		return nil
	}

	err = chain.Handle(ctx, req, rsp, unit.Extend, dbExecFilter)
	if err != nil {
		return
	}

	return rsp.Result, rsp.Detail, rsp.IsNil, rsp.Error
}

// 获取插件链
func getPluginChain(ctx context.Context, appid uint64, tblTable *obj.TblTable) (Chain, error) {
	tablePlugins := table.GetTablePlugins(tblTable.Id)

	ret := Chain{}
	for _, tablePlugin := range tablePlugins {
		tblPlugin := table.GetPlugin(tablePlugin.PluginID)

		if tblPlugin == nil {
			e := errs.NewPluginf(errs.ErrPluginNotFound, "not find plugin : %d", tablePlugin.PluginID)
			if tablePlugin.ScheduleConf.SkipError {
				log.Error(ctx, errs.ErrPluginNotFound, e.Error())
				continue
			} else {
				return nil, e
			}
		}

		funcName := fmt.Sprintf("%s_%d", tblPlugin.Name, tablePlugin.PluginVersion)
		f := plugin.Func[funcName]

		if f == nil {
			e := errs.NewPluginf(errs.ErrPluginFuncNotRegister, "plugin %s functions "+
				"for version %d are not registered", tblPlugin.Name, tablePlugin.PluginVersion)

			if tablePlugin.ScheduleConf.SkipError {
				log.Error(ctx, errs.ErrPluginFuncNotRegister, e.Error())
				continue
			} else {
				return nil, e
			}
		}

		ret = append(ret, &PluginHandler{
			appid: appid,
			tp:    tablePlugin,
			f:     f,
		})
	}

	return ret, nil
}

type PluginHandler struct {
	appid uint64
	tp    *table.TblTablePlugin
	f     plugin.Plugin
}

// Chain chains of plugin handler
type Chain []*PluginHandler

// Handle invokes every server side filters in the chain.
func (c Chain) Handle(ctx context.Context, req *pf.Request, rsp *pf.Response, extend types.Map, next conf.HandleFunc) error {
	for i := len(c) - 1; i >= 0; i-- {
		curHandleFunc, curPlugin := next, c[i]
		next = func(ctx context.Context) error {
			e := pluginHandle(ctx, req, rsp, extend, curPlugin.tp, curPlugin.f, curHandleFunc)
			if e != nil {
				if curPlugin.tp.ScheduleConf.SkipError {
					log.Error(ctx, errs.ErrPluginNotFound, e.Error())
				} else {
					return getPluginError(e)
				}
			}

			return nil
		}
	}

	return next(ctx)
}

func pluginHandle(ctx context.Context, req *pf.Request, resp *pf.Response,
	extend types.Map, tablePlugin *table.TblTablePlugin, f plugin.Plugin, next conf.HandleFunc) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = errs.NewPluginf(errs.ErrPanic,
				"plugin handle panic: [%v], req=%s, resp=%s, extend=%s ",
				e, json.MarshalToString(req),
				json.MarshalToString(resp),
				json.MarshalToString(extend))

			log.Error(ctx, errs.ErrPanic, err.Error())
		}
	}()

	return f.Handle(ctx, req, resp, extend, tablePlugin.Conf, next)
}

func getPluginError(err error) error {
	if err == nil {
		return nil
	}

	if errs.Type(err) == errs.ETypePlugin || errs.Type(err) == errs.ETypeDatabase {
		return err
	}

	e := errs.NewPluginf(errs.Code(err), errs.Msg(err))

	if errs.Code(e) == errs.ErrUnknown {
		e = errs.SetErrorCode(e, errs.ErrPluginExec)
	}

	return e
}

