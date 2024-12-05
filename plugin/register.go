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
	register("cache_front_handle", &cache.FrontPlugin{})
	register("cache_post_handle", &cache.PostPlugin{})
}
