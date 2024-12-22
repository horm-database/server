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

package api

import (
	"context"

	"github.com/horm-database/common/compress"
	"github.com/horm-database/common/consts"
	"github.com/horm-database/common/errs"
	"github.com/horm-database/common/proto"
	"github.com/horm-database/server/auth"
	"github.com/horm-database/server/logic"
	"github.com/horm-database/server/srv/codec"
)

// Query data query api
func Query(ctx context.Context, head *proto.RequestHeader, reqBuf []byte) (interface{}, error) {
	if !auth.SignSuccess(head) {
		//return nil, errs.Newf(errs.ErrAuthFail, "signature failed")
	}

	var err error
	if head.Compress == consts.Compression {
		reqBuf, err = compress.Decompress(reqBuf)
		if err != nil {
			return nil, errs.Newf(errs.ErrServerDecompress, "request body decompress error: %s", err.Error())
		}
	}

	// unmarshal request body
	var units = []*proto.Unit{}
	if reqBuf[0] == '[' {
		err = codec.Deserialize(ctx, reqBuf, &units)
		if err == nil {
			//校验 parse mode
			parseQueryMode := getQueryMode(units)
			if uint32(parseQueryMode) != head.QueryMode {
				parseQueryModeDesc, _ := consts.QueryModeDesc[parseQueryMode]
				inputQueryModeDesc, _ := consts.QueryModeDesc[int8(head.QueryMode)]
				return nil, errs.Newf(errs.ErrParamInvalid, "query mode is invalid, "+
					"it should be [%s], but input is [%s]", parseQueryModeDesc, inputQueryModeDesc)
			}
		}
	} else { // 单执行单元
		head.QueryMode = consts.QueryModeSingle
		tmp := proto.Unit{}
		err = codec.Deserialize(ctx, reqBuf, &tmp)
		units = append(units, &tmp)
	}

	if err != nil {
		return nil, errs.Newf(errs.ErrServerDecode, "request body codec unmarshal error: %s", err.Error())
	}

	return logic.Parse(ctx, head, units)
}

// 根据 units 获取 query mode.
func getQueryMode(units []*proto.Unit) int8 {
	var unitNum int
	for _, unit := range units {
		if len(unit.Sub) > 0 {
			return consts.QueryModeCompound
		}

		if len(unit.Trans) > 0 {
			for _, transUnit := range unit.Trans {
				if len(transUnit.Sub) > 0 {
					return consts.QueryModeCompound
				}
			}

			unitNum = unitNum + len(unit.Trans)
		} else {
			unitNum++
		}
	}

	if unitNum <= 1 {
		return consts.QueryModeSingle
	} else {
		return consts.QueryModeParallel
	}
}
