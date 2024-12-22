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

package main

import (
	"time"

	_ "go.uber.org/automaxprocs"

	"github.com/horm-database/common/log"
	"github.com/horm-database/server/api"
	"github.com/horm-database/server/model"
	"github.com/horm-database/server/plugin"
	"github.com/horm-database/server/srv"
	"github.com/horm-database/server/srv/codec"
)

func main() {
	server := srv.NewServer(api.ServerDesc)

	// 注册插件处理函数
	plugin.Register()

	model.Init(codec.GCtx, srv.Config().MachineID)

	go func() {
		for {
			go model.SyncDbNewToLocal(codec.GCtx)
			//go batch.InsertHandle(codec.GCtx)
			//go batch.FailedCheck(codec.GCtx)
			time.Sleep(2 * time.Second)
		}
	}()

	if err := server.Serve(); err != nil {
		log.Fatal(codec.GCtx, err)
	}
}
