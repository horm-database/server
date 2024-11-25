package filter

import (
	"github.com/horm-database/server/filter/official/cache"
	"github.com/horm-database/server/filter/official/uniquekey"
)

// Register 注册插件函数
func Register() {
	registerOfficial()
	registerThirdParty()
	registerPrivate()
}

// registerOfficial 注册官方插件
func registerOfficial() {
	register("unique_key", &uniquekey.Filter{})
	//filter.RegisterFilter("batch_insert", &batch.Filter{})
	register("cache_front_handle", &cache.FrontFilter{})
	register("cache_post_handle", &cache.PostFilter{})
}
