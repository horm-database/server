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

package codec

import (
	"context"

	"github.com/horm-database/common/codec"
	"github.com/horm-database/common/log/logger"
)

var (
	GCtx context.Context
)

// InitGlobalContext 获取初始化的 context、msg
func InitGlobalContext(env, container, callee string) {
	ctx, msg := codec.NewMessage(context.Background())

	msg.WithEnv(env)
	msg.WithCalleeServiceName(callee)

	msg.WithLogger(logger.DefaultLogger.With(
		logger.Field{"callee", callee},
		logger.Field{"container", container}))

	GCtx = ctx
}
