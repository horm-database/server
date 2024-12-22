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

package util

import (
	"context"

	"github.com/horm-database/common/errs"
	"github.com/horm-database/common/log"
	"github.com/horm-database/common/proto"
)

func ErrorToRspError(err error) *proto.Error {
	return &proto.Error{
		Type: int32(errs.Type(err)),
		Code: int32(errs.Code(err)),
		Msg:  errs.Msg(err),
	}
}

// LogErrorf 生成错误，打印错误日志，如果错误码不存在，则使用输入的默认错误码
func LogErrorf(ctx context.Context, code int, format string, params ...interface{}) error {
	err := errs.Newf(code, format, params...)
	log.Errorf(ctx, code, format, params...)
	return err
}
