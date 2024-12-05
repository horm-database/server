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
