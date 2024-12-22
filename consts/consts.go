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

const ( //超级权限，任意权限
	DBRootAll       = 1 // 超级权限（所有权限，包含DDL）
	DBRootTableData = 2 // 表数据权限（库下表的所有增删改查权限，不包含 DDL）  3-无
	DBRootNone      = 3 // 无
)

const ( // 是否支持所有的 query 语句
	TableQueryAllTrue  = 1 // 是，query
	TableQueryAllFalse = 2 // 否
)

const ( // 访问表/库权限状态
	AuthStatusNormal   = 1 // 正常
	AuthStatusOffline  = 2 // 下线
	AuthStatusChecking = 3 // 审核中
	AuthStatusCancel   = 4 // 撤销
	AuthStatusReject   = 5 // 拒绝
)

const (
	QueryFinishedNo       = 0 // query 未完成
	QueryFinishedYes      = 1 // query 已完成
	QueryFinishedRollback = 2 // 待回滚，不执行 query
)

const ( //是否强制签名
	WorkspaceEnforceSignNo  = 0
	WorkspaceEnforceSignYes = 1
)
