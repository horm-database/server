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

package conf

// ScheduleConfig 插件调度配置
type ScheduleConfig struct {
	Async         bool        `json:"async"`          // 是否异步执行，默认 false
	SkipError     bool        `json:"skip_error"`     // 是否跳过 error，默认 false（插件返回报错是返回客户端，还是继续执行）
	Timeout       int         `json:"timeout"`        // 单个插件的超时时间，默认 1000 ms
	RequestSource []string    `json:"request_source"` // 指定请求来源，API 接口、WEB 管理，默认都生效 ["api","web"]
	OpType        []string    `json:"op_type"`        // 指定操作类型，默认增删改查都生效: ["read", "mod", "del"]
	GrayScale     int         `json:"gray_scale"`     // 灰度比例，0-100，默认 100（仅针对 API 接口）
	AppRule       *AppRule    `json:"app_rule"`       // 指定 app 执行/跳过插件
	CustomRule    *CustomRule `json:"custom_rule"`    // 自定义规则
}

type AppRule struct {
	ActType int8     `json:"act_type"` // 动作类别：1-执行插件 2-跳过插件
	AppIDs  []uint64 `json:"appids"`   // 生效的 appid
}

type CustomRule struct {
	ActType  int8    `json:"act_type"`  // 动作类别：1-执行插件 2-跳过插件
	RuleType int8    `json:"rule_type"` // 满足以下：1-任一规则 2-所有规则
	Rules    []*Rule `json:"rules"`     // 规则
}

type Rule struct {
	Name     string       `json:"name"`      // 规则名
	CondType int8         `json:"cond_type"` // 满足以下：1-任一条件 2-所有条件
	Cond     []*Condition `json:"cond"`      // 条件
}

type Condition struct {
	Key   string `json:"key"`   // Extend["key"]
	Op    int8   `json:"op"`    // 操作符 1-等于 2-不等于 3-大于 4-大于等于 5-小于 6-小于等于 7-类似于 8-不类似于 9-开头类似于 10-结尾类似于 11-存在于集合(in) 12-不存在于集合(not in)
	Value string `json:"value"` // 当 Extend["key"] ${op} value 时，才会执行插件
}
