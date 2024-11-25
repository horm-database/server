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
