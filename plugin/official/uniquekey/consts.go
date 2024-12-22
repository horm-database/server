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

const ( // 唯一键自动生成类型
	UKAutoGenNo         = 0 //不自动生成
	UKAutoGenByDB       = 1 //存储引擎自增，比如 mysql 的 auto createment
	UKAutoGenByUStorage = 2 //由统一存储自动生成全局唯一的值（注意，如果需要统一存储生成，字段类型必须是字符长，长度必须>=32）
)
