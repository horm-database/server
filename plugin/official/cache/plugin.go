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

package cache

import (
	"context"
	"fmt"

	"github.com/horm-database/common/compress"
	"github.com/horm-database/common/consts"
	"github.com/horm-database/common/proto"
	"github.com/horm-database/common/proto/plugin"
	"github.com/horm-database/common/types"
	"github.com/horm-database/orm"
	"github.com/horm-database/server/plugin/conf"
)

type CacheData struct {
	IsNil   bool                     `json:"is_nil,omitempty"`
	IsArray bool                     `json:"array,omitempty"`
	Version int                      `json:"ver,omitempty"`
	Total   uint64                   `json:"total,omitempty"`
	Detail  *proto.Detail            `json:"detail,omitempty"`
	Data    map[string]interface{}   `json:"data,omitempty"`
	Datas   []map[string]interface{} `json:"datas,omitempty"`
}

type Plugin struct{}     // 缓存前置插件
type PostPlugin struct{} // 缓存后置插件

func (ft *Plugin) Handle(ctx context.Context,
	req *plugin.Request,
	rsp *plugin.Response,
	extend types.Map,
	conf conf.PluginConfig, hf conf.HandleFunc) (err error) {
	var limitField string
	var writeCache bool

	cacheKey, _ := extend.GetString("key")
	cacheOP, _ := extend.GetString("op")

	if cacheKey != "" {
		limit, offset := req.Size, req.From
		if req.Page > 0 {
			offset = uint64((req.Page - 1) * req.Size)
		}

		if limit > 0 {
			limitField = fmt.Sprintf("%d_%d", limit, offset)
		}

		if req.Op == consts.OpFind || req.Op == consts.OpFindAll {
			writeCache = true
			cacheResult := getFromCache(ctx, 1111, cacheKey, limitField)
			if cacheResult != nil {
				if cacheResult.IsArray {
					rsp.Result = cacheResult.Datas
				} else {
					rsp.Result = cacheResult.Data
				}

				rsp.IsNil = cacheResult.IsNil
				rsp.Detail = cacheResult.Detail
				return nil
			}
		} else { //缓存变更
			switch cacheOP {
			case "add": //用新数据替换缓存
				if req.Op == consts.OpInsert || req.Op == consts.OpReplace {

				}
			case "mod": //修改缓存部分字段
			default: //默认全部清理老缓存
				deleteCache(ctx, 1111, cacheKey)
			}
		}
	}

	extend["limitField"] = limitField
	extend["writeCache"] = writeCache

	return hf(ctx)
}

func (ft *PostPlugin) Handle(ctx context.Context,
	_ *plugin.Request,
	rsp *plugin.Response,
	extend types.Map,
	conf conf.PluginConfig) (response bool, err error) {
	writeCache, exists := extend.GetBool("writeCache")
	if !exists {
		return
	}

	limitField, exists := extend.GetString("limitField")
	if !exists {
		return
	}

	cacheKey, _ := extend.GetString("key")
	ttl, _, _ := extend.GetInt("ttl")

	if writeCache { //数据写入缓存
		setToCache(ctx, 1111, cacheKey, limitField, ttl, rsp.Result, rsp.Detail, rsp.IsNil)
	}

	return false, nil
}

func setToCache(ctx context.Context, tableId int, key, field string,
	ttl int, data interface{}, detail *proto.Detail, isNil bool) {
	cacheRedis := orm.NewORM("cache")

	key = fmt.Sprintf("%s_%d_%s", PreFindCache, tableId, key)
	result := CacheData{
		IsNil:  isNil,
		Detail: detail,
		Total:  detail.Total,
	}

	if !isNil {
		if datas, ok := data.([]map[string]interface{}); ok {
			result.Datas = datas
			result.IsArray = true
		} else if data, ok := data.(map[string]interface{}); ok {
			result.Data = data
		} else {
			return
		}
	}

	gzipData, err := compress.JsonMarshalAndCompress(result)
	if err == nil {
		if field == "" {
			_, _ = cacheRedis.SetEX(key, gzipData, ttl).Exec(ctx)
		} else {
			_, _ = cacheRedis.HSet(key, field, gzipData).Exec(ctx)
			_, _ = cacheRedis.Expire(key, ttl).Exec(ctx)
		}
		return
	}

	if field == "" {
		_, _ = cacheRedis.SetEX(key, &result, ttl).Exec(ctx)
	} else {
		_, _ = cacheRedis.HSet(key, field, &result).Exec(ctx)
		_, _ = cacheRedis.Expire(key, ttl).Exec(ctx)
	}
	return
}

func getFromCache(ctx context.Context, tableId int, key, field string) *CacheData {
	cacheRedis := orm.NewORM("cache")

	key = fmt.Sprintf("%s_%d_%s", PreFindCache, tableId, key)

	var err error
	bts := []byte{}

	if field == "" {
		_, err = cacheRedis.Get(key).Exec(ctx, &bts)
	} else {
		_, err = cacheRedis.HGet(key, field).Exec(ctx, &bts)
	}

	if err != nil {
		return nil
	}

	result := CacheData{}
	_ = compress.DecompressJsonUnmarshal(bts, &result)
	return &result
}

func deleteCache(ctx context.Context, tableId int, key string) {
	cacheRedis := orm.NewORM("cache")
	key = fmt.Sprintf("%s_%d_%s", PreFindCache, tableId, key)
	_, _ = cacheRedis.Del(key).Exec(ctx)
	return
}
