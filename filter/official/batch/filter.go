package batch

import (
	"context"

	"github.com/horm-database/common/consts"
	"github.com/horm-database/common/proto/filter"
	"github.com/horm-database/orm/obj"
	"github.com/horm-database/server/srv/codec"
)

// Filter 表的唯一键生成插件
type Filter struct{}

func (ft *Filter) Handle(ctx context.Context, req *filter.Request, resp *filter.Response,
	dbConf obj.TblDB, tableConf obj.TblTable, config map[string]interface{}) error {
	batchFlush, ok1 := config["batch_flush"].(int)
	batchNum, ok2 := config["batch_num"].(int)
	// 异步批量插入
	if req.Op == consts.OpInsert && (ok1 && batchFlush == FlushOpen) {
		_, err := PushBatchInsert(ctx, BufferRedis,
			tableConf.Id, req.Tables[0], req.Data, req.Datas, req.DataType)

		//成功则直接返回，失败则走直接插入
		if err == nil {
			if ok2 && batchNum > 0 { //如果缓冲区阈值
				go func() {
					l := BufferLen(codec.GCtx, tableConf.Id, req.Tables[0])
					if l > batchNum {
						Insert(codec.GCtx, &tableConf, nil)
					}
				}()
			}

			resp.Return = true
			return nil
		}
	}

	return nil
}
