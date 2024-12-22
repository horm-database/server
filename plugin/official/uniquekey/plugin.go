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

package uniquekey

import (
	"context"

	"github.com/horm-database/common/consts"
	"github.com/horm-database/common/proto/plugin"
	"github.com/horm-database/common/snowflake"
	"github.com/horm-database/common/types"
	"github.com/horm-database/server/plugin/conf"
)

// Plugin 表主键生成插件
type Plugin struct{}

func (ft *Plugin) Handle(ctx context.Context,
	req *plugin.Request,
	rsp *plugin.Response,
	extend types.Map,
	conf conf.PluginConfig) (response bool, err error) {
	ukAutoGenerate, _, _ := conf.GetInt("uk_auto_generate")
	uniqueKey, _ := conf.GetString("unique_key")
	if (ukAutoGenerate == UKAutoGenByUStorage) && uniqueKey != "" && req.Op == consts.OpInsert {
		if len(req.Datas) > 0 {
			for k := range req.Datas {
				req.Datas[k][uniqueKey] = snowflake.GenerateID()
			}
		} else {
			req.Data[uniqueKey] = snowflake.GenerateID()
		}
	}

	return false, nil
}
