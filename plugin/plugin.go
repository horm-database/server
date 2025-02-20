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

package plugin

import (
	"context"
	"fmt"

	"github.com/horm-database/common/errs"
	"github.com/horm-database/common/proto/plugin"
	"github.com/horm-database/common/types"
	"github.com/horm-database/server/plugin/conf"
)

// Plugin 插件
type Plugin interface {
	// Handle 插件处理函数。
	// input param: ctx context 上下文。
	// input param: req 请求参数。
	// input param: rsp 返回参数。
	// input param: extend 客户端送的扩展信息，也可以将信息从上一个插件传递到下一个插件，另外请求头部信息也会通过 extend 带进来。
	// input param: conf 插件配置。
	// output param: err 插件处理异常，err 非空会直接返回客户端 error，不再执行后续逻辑。
	// output param: f 执行函数
	Handle(ctx context.Context,
		req *plugin.Request,
		rsp *plugin.Response,
		extend types.Map,
		conf conf.PluginConfig, f conf.HandleFunc) error
}

// GetRequestHeader get request header from extend
func GetRequestHeader(extend types.Map) *plugin.Header {
	header, _ := extend["request_header"].(*plugin.Header)
	return header
}

var Func = map[string]Plugin{}

func register(name string, plugin Plugin, version ...int) {
	var ver int

	if len(version) > 0 {
		ver = version[0]
	}

	name = fmt.Sprintf("%s_%d", name, ver)

	_, exits := Func[name]
	if exits {
		panic(errs.Newf(1, "plugin %s has already registered", name))
	}

	Func[name] = plugin
}
