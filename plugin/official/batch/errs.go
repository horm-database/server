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

package batch

const (
	RetBatchInsert        = 101 // 批量插入异常
	RetBatchDataUnMarshal = 102 // 批量结果解压缩失败
	RetHasBatchFailed     = 103 // 有异常待处理批量插入记录
	RetBatchFailedHandle  = 104 // 批量插入异常处理失败
	ErrFormatData         = 105 // 格式化数据 error
)
