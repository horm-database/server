package uniquekey

import (
	"context"

	"github.com/horm-database/common/consts"
	"github.com/horm-database/common/proto/plugin"
	"github.com/horm-database/common/snowflake"
	"github.com/horm-database/common/types"
	"github.com/horm-database/server/plugin/conf"
)

// Plugin 表主键生成插件
type Plugin struct{}

func (ft *Plugin) Handle(ctx context.Context,
	req *plugin.Request,
	rsp *plugin.Response,
	extend types.Map,
	conf conf.PluginConfig) (response bool, err error) {
	ukAutoGenerate, _, _ := conf.GetInt("uk_auto_generate")
	uniqueKey, _ := conf.GetString("unique_key")
	if (ukAutoGenerate == UKAutoGenByUStorage) && uniqueKey != "" && req.Op == consts.OpInsert {
		if len(req.Datas) > 0 {
			for k := range req.Datas {
				req.Datas[k][uniqueKey] = snowflake.GenerateID()
			}
		} else {
			req.Data[uniqueKey] = snowflake.GenerateID()
		}
	}

	return false, nil
}
