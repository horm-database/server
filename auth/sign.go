package auth

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/horm/common/crypto"
	"github.com/horm/common/proto"
	"github.com/horm/server/model/table"
)

var SameRequestLock = new(sync.RWMutex)
var SameRequest = map[string]bool{}

func init() {
	go func() {
		for {
			SameRequestLock.Lock()
			SameRequest = map[string]bool{}
			SameRequestLock.Unlock()
			time.Sleep(time.Duration(30+rand.Intn(30)) * time.Second) // 0.5~1 分钟清空一次
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
