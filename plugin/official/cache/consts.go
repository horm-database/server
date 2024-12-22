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
package cache

const ( // 缓存一致性方案
	ConsistencyTypeNone    = 0 //无需保障一致性（最终一致性，缓存到期就没了）
	ConsistencyTypeCount   = 1 //记录条数（针对的是只有新增类型的数据）
	ConsistencyTypeVersion = 2 //版本号（没有单条记录的并发问题）
	ConsistencyTypeQueue   = 3 //队列
	ConsistencyTypeLock    = 4 //锁
)

const ( // redis 缓存前缀
	PreFindCache = "data_" //数据缓存
)
