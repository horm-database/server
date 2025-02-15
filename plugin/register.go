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
	"github.com/horm-database/server/plugin/official/cache"
	"github.com/horm-database/server/plugin/official/uniquekey"
)

// Register 注册插件函数
func Register() {
	registerOfficial()
	registerThirdParty()
	registerPrivate()
}

// registerOfficial 注册官方插件
func registerOfficial() {
	register("unique_key", &uniquekey.Plugin{})
	//plugin.RegisterPlugin("batch_insert", &batch.Plugin{})
	register("cache_handle", &cache.Plugin{})
}
