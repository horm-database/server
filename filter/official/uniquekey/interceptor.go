package uniquekey

import (
	"context"

	"github.com/horm-database/common/consts"
	"github.com/horm-database/common/proto/filter"
	"github.com/horm-database/common/snowflake"
	"github.com/horm-database/common/types"
	"github.com/horm-database/server/filter/conf"
)

// Filter 表主键生成插件
type Filter struct{}

func (ft *Filter) Handle(ctx context.Context,
	req *filter.Request,
	rsp *filter.Response,
	extend types.Map,
	conf conf.FilterConfig) (response bool, err error) {
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
