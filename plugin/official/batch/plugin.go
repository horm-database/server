// Copyright (c) 2024 The horm-database Authors (such as CaoHao <18500482693@163.com>). All rights reserved.
//
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
package batch

import (
	"context"

	"github.com/horm-database/common/consts"
	"github.com/horm-database/common/proto/plugin"
	"github.com/horm-database/orm/obj"
	"github.com/horm-database/server/srv/codec"
)

// Plugin 表的唯一键生成插件
type Plugin struct{}

func (ft *Plugin) Handle(ctx context.Context, req *plugin.Request, resp *plugin.Response,
	dbConf obj.TblDB, tableConf obj.TblTable, config map[string]interface{}) error {
	batchFlush, ok1 := config["batch_flush"].(int)
	batchNum, ok2 := config["batch_num"].(int)
	// 异步批量插入
	if req.Op == consts.OpInsert && (ok1 && batchFlush == FlushOpen) {
		_, err := PushBatchInsert(ctx, BufferRedis,
			tableConf.Id, req.Tables[0], req.Data, req.Datas, req.DataType)

		//成功则直接返回，失败则走直接插入
		if err == nil {
			if ok2 && batchNum > 0 { //如果缓冲区阈值
				go func() {
					l := BufferLen(codec.GCtx, tableConf.Id, req.Tables[0])
					if l > batchNum {
						Insert(codec.GCtx, &tableConf, nil)
					}
				}()
			}

			resp.Return = true
			return nil
		}
	}

	return nil
}
