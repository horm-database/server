package main

import (
	"time"

	"github.com/horm/server/test"
	_ "go.uber.org/automaxprocs"

	"github.com/horm/common/log"
	"github.com/horm/go-horm/horm"
	"github.com/horm/server/api"
	"github.com/horm/server/filter"
	"github.com/horm/server/model"
	"github.com/horm/server/srv"
	"github.com/horm/server/srv/codec"
)

func main() {
	server := srv.NewServer(api.ServerDesc)

	// 注册插件处理函数
	filter.Register()

	model.Init(codec.GCtx, srv.Config().MachineID)

	go func() {
		for {
			go model.SyncDbNewToLocal(codec.GCtx)
			//go batch.InsertHandle(codec.GCtx)
			//go batch.FailedCheck(codec.GCtx)
			time.Sleep(2 * time.Second)
		}
	}()

	go test.TestMySQL()

	if err := server.Serve(); err != nil {
		log.Fatal(codec.GCtx, err)
	}
}

// init 配置全局统一接入协议执行器
func init() {
	horm.SetGlobalClient("default/app.server.service1",
		horm.WithAppID(10002),
		horm.WithSecret("S959223456"),
		horm.WithTarget("ip://127.0.0.1:8180"),
		horm.WithTimeout(332320),
		horm.WithCaller("app.server.service"))

	//horm.SetGlobalProxy("10002", "S959223456", "polaris://rpc.server.access.goapi", 332320)
}
