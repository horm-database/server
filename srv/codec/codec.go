// Copyright (c) 2024 The horm-database Authors (such as CaoHao <18500482693@163.com>). All rights reserved.
//
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
// Package codec defines the business communication protocol of
// packing and unpacking.
package codec

import (
	"github.com/horm-database/common/codec"
	"github.com/horm-database/common/proto"
)

// Codec defines the interface of business communication protocol,
// which contains head and body. It only parses the body in binary
type Codec interface {
	// Encode pack the body into binary buffer.
	Encode(msg *codec.Msg, body []byte) (respBody []byte, err error)

	// Decode unpack the body from binary buffer
	Decode(message *codec.Msg, buf []byte) (reqBody []byte, err error)
}

func GetRespFromReqHeader(reqHeader *proto.RequestHeader) *proto.ResponseHeader {
	return &proto.ResponseHeader{
		Version:   reqHeader.Version,
		QueryMode: reqHeader.QueryMode,
		RequestId: reqHeader.RequestId,
	}
}
