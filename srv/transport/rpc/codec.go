package rpc

import (
	"errors"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/horm-database/common/codec"
	"github.com/horm-database/common/errs"
	cp "github.com/horm-database/common/proto"
	"github.com/horm-database/common/util"
	cc "github.com/horm-database/server/srv/codec"
	"github.com/horm-database/server/srv/transport"
)

var (
	// DefaultServerCodec is the default rpc server codec.
	DefaultServerCodec = &ServerCodec{}
)

// ServerCodec is an implementation of codec.Codec. used for rpc serverside codec.
type ServerCodec struct{}

// Decode implements codec.Codec.
// It decodes the request buffer into reqHeader and reqBody,
// and updates the msg already initialized by rpcServer.
func (s *ServerCodec) Decode(msg *codec.Msg, reqBuf []byte) ([]byte, error) {
	fc := msg.FrameCodec().(*frameCodec)

	if fc.frameHead.HeaderLen == 0 { // header not allowed to be empty for unary rpc
		return nil, errors.New("server decode pb head len empty")
	}

	if int(fc.frameHead.HeaderLen) > len(reqBuf) {
		return nil, errors.New("server pb header len is long than request buffer")
	}

	reqHeader := &cp.RequestHeader{}
	if err := proto.Unmarshal(reqBuf[:fc.frameHead.HeaderLen], reqHeader); err != nil {
		return nil, err
	}

	reqHeader.Ip = util.GetIpFromAddr(msg.RemoteAddr())

	// 根据解析后的 request_header 更新 msg
	msg.WithServerReqHead(reqHeader)
	msg.WithServerRespHead(cc.GetRespFromReqHeader(reqHeader))
	msg.WithRequestTimeout(time.Duration(reqHeader.Timeout) * time.Millisecond)
	msg.WithCallerServiceName(reqHeader.Caller)
	msg.WithCallRPCName(reqHeader.Callee)
	msg.WithSerializationType(cc.SerializationTypeJSON)
	msg.WithRequestID(reqHeader.RequestId)
	msg.WithTraceID(reqHeader.TraceId)

	return reqBuf[fc.frameHead.HeaderLen:], nil
}

// Encode implements codec.Codec.
// check if respond http error buffer to client .
// else it encodes the respHeader and respBody to binary body.
func (s *ServerCodec) Encode(msg *codec.Msg, respBody []byte) ([]byte, error) {
	frameHead := codec.NewFrameHead()

	respHeader := s.getResponseHeader(msg)

	// convert error returned by server handler to Error in response protocol head
	if err := msg.ServerRespError(); err != nil {
		respHeader.Err = &cp.Error{
			Type: int32(err.Type),
			Code: int32(err.Code),
			Msg:  err.Msg,
		}
	}

	respHeaderBuf, err := proto.Marshal(respHeader)
	if err != nil {
		return nil, errs.Newf(errs.RetServerEncodeFail, "rpc proto marshal response header error: %v", err)
	}

	respBuf, err := frameHead.Construct(respHeaderBuf, respBody)
	if errors.Is(err, codec.ErrHeadOverflowsUint16) {
		return s.handleEncodeErr(respHeader, frameHead, respBody, err)
	}

	if errors.Is(err, codec.ErrFrameTooLarge) || errors.Is(err, codec.ErrHeadOverflowsUint32) {
		// if length of frame is larger than MaxFrameSize or overflows uint32 set respBody nil
		return s.handleEncodeErr(respHeader, frameHead, nil, err)
	}

	return respBuf, err
}

// getResponseHeader returns response header from msg.
// If response header is not found from msg, a new response header will be created and returned.
func (s *ServerCodec) getResponseHeader(msg *codec.Msg) *cp.ResponseHeader {
	respHeader, ok := msg.ServerRespHead().(*cp.ResponseHeader)
	if !ok {
		respHeader = &cp.ResponseHeader{}

		request, ok := msg.ServerReqHead().(*cp.RequestHeader)
		if ok {
			respHeader.QueryMode = request.QueryMode
			respHeader.RequestId = request.RequestId
		}
	}
	return respHeader
}

// handleEncodeErr handle encode err and return RetServerEncodeFail
func (s *ServerCodec) handleEncodeErr(respHeader *cp.ResponseHeader,
	frameHead *codec.FrameHead, respBody []byte, encodeErr error) ([]byte, error) {
	// cover the original error.
	respHeader.Err = &cp.Error{
		Type: errs.ErrorTypeSystem,
		Code: errs.RetServerEncodeFail,
		Msg:  encodeErr.Error(),
	}

	respHeader.RspErrs = nil

	respHeaderBuf, err := proto.Marshal(respHeader)
	if err != nil {
		return nil, err
	}

	// if error still occurs, response will be discarded. client will be notified as pool closed
	return frameHead.Construct(respHeaderBuf, respBody)
}

type frameCodec struct {
	opts             *transport.Options
	frameType        uint8
	frameHead        *codec.FrameHead
	signFrameHead    *codec.SignFrameHead
	encryptFrameHead *codec.EncryptFrameHead
	buf              []byte
	bufLen           int
	totalLen         int
}

func (fc *frameCodec) resetBuf() {
	if len(fc.buf) > 0 {
		fc.buf = fc.buf[:0]
	}
	fc.bufLen = 0
	fc.frameHead = nil
}
