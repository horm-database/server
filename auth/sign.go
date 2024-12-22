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
package auth

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/horm-database/common/crypto"
	"github.com/horm-database/common/proto"
	"github.com/horm-database/server/model/table"
)

var SameRequestLock = new(sync.RWMutex)
var SameRequest = map[string]bool{}

func init() {
	go func() {
		for {
			SameRequestLock.Lock()
			SameRequest = map[string]bool{}
			SameRequestLock.Unlock()
			time.Sleep(time.Duration(15+rand.Intn(15)) * time.Second) // 15 ~ 30 S清空一次
		}
	}()
}

// SignSuccess 签名是否正确
func SignSuccess(head *proto.RequestHeader) bool {
	if head.Appid == 0 {
		return false
	}

	secret := getSecretByAppid(head.Appid)
	if secret == "" {
		return false
	}

	md5Str := fmt.Sprintf("%d%s%d%d%d%s%d%d%s%d%d%d", head.Appid, secret,
		head.RequestType, head.QueryMode, head.RequestId, head.TraceId, head.Timestamp,
		head.Timeout, head.Caller, head.Compress, head.AuthRand, head.Version)

	sign := crypto.MD5Str(md5Str)

	if sign != head.Sign {
		return false
	}

	requestUniq := fmt.Sprintf("%d%s%d", head.Timestamp, head.Ip, head.AuthRand)

	SameRequestLock.Lock()
	isSame := SameRequest[requestUniq]
	isSame2 := SameRequest[sign]
	SameRequest[requestUniq] = true
	SameRequest[sign] = true
	SameRequestLock.Unlock()

	if isSame || isSame2 {
		return false
	}

	return true
}

func getSecretByAppid(appid uint64) string {
	appInfo := table.GetAppInfo(appid)
	if appInfo != nil {
		return appInfo.Info.Secret
	}
	return ""
}
