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

package consts

const ( // 插件类型
	PrePlugin   = 1 // 前置插件
	PostPlugin  = 2 // 后置插件
	DeferPlugin = 3 // 延迟插件（query 最后执行，不改变返回结果，而且每个 defer 都会被执行）
)

const ( // 插件配置类型
	PluginConfigTypeBool        = 1  // bool
	PluginConfigTypeString      = 2  // string
	PluginConfigTypeInt         = 3  // int
	PluginConfigTypeUint        = 4  // uint
	PluginConfigTypeFloat       = 5  // float
	PluginConfigTypeBytes       = 6  // bytes
	PluginConfigTypeEnum        = 7  // 枚举（单选）
	PluginConfigTypeMultiChoice = 8  // 多选
	PluginConfigTypeTime        = 9  // 时间
	PluginConfigTypeArray       = 10 // 数组
	PluginConfigTypeMap         = 11 // map
	PluginConfigTypeMultiConf   = 12 // 配置数组
)

const ( // 插件动作类型
	ActionTypeExec = 1 // 1-执行插件
	ActionTypeSkip = 2 // 2-跳过插件
)

const ( // 规则类型
	CondTypeAny = 1 // 1-任一规则(条件)
	CondTypeAll = 2 // 2-所有规则（条件）
)
