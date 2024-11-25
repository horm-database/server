package util

import (
	"context"

	"github.com/horm-database/common/errs"
	"github.com/horm-database/common/log"
	"github.com/horm-database/common/proto"
)

func ErrorToRspError(err error) *proto.Error {
	return &proto.Error{
		Type: int32(errs.Type(err)),
		Code: int32(errs.Code(err)),
		Msg:  errs.Msg(err),
	}
}

// LogErrorf 生成错误，打印错误日志，如果错误码不存在，则使用输入的默认错误码
func LogErrorf(ctx context.Context, code int, format string, params ...interface{}) error {
	err := errs.Newf(code, format, params...)
	log.Errorf(ctx, code, format, params...)
	return err
}
