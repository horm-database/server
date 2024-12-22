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

package http

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/evanphx/wildcat"
	cc "github.com/horm-database/common/codec"
	"github.com/horm-database/common/consts"
	"github.com/horm-database/common/errs"
	"github.com/horm-database/common/proto"
	"github.com/horm-database/common/types"
	"github.com/horm-database/common/util"
	"github.com/horm-database/server/srv/codec"
	"github.com/horm-database/server/srv/transport"
)

var (
	// DefaultServerCodec is the default http server codec.
	DefaultServerCodec = &ServerCodec{}
)

// ServerCodec http server side codec. used for http serverside codec.
type ServerCodec struct{}

// Decode implements codec.Codec.
// get request header from frameCodec`s HTTPParser and return http reqBody to handler
func (sc *ServerCodec) Decode(msg *cc.Msg, reqBody []byte) ([]byte, error) {
	fc := msg.FrameCodec().(*frameCodec)
	if fc == nil {
		return nil, errors.New("server decode missing frame codec in context")
	}

	if types.BytesToString(fc.Parser.Method) != http.MethodPost {
		return nil, fmt.Errorf("http method is not POST")
	}

	if len(reqBody) != fc.bodyLen {
		return nil, fmt.Errorf("http body length illegal")
	}

	if err := sc.setReqHeader(fc, msg); err != nil {
		return nil, err
	}

	return reqBody, nil
}

// Encode implements codec.Codec.
// check if respond http error buffer to client. or encode respHeader and respBody to http response buffer.
func (sc *ServerCodec) Encode(msg *cc.Msg, respBody []byte) (b []byte, err error) {
	fc := msg.FrameCodec().(*frameCodec)
	if fc == nil {
		return nil, errors.New("server encode frame codec is missing in context")
	}

	// response error
	if e := msg.ServerRespError(); e != nil {
		return getErrorRespBuf(fc, e), nil
	}

	return getRespBuf(fc, respBody), nil
}

// get request header from frameCodec`s HTTPParser
func (sc *ServerCodec) setReqHeader(fc *frameCodec, msg *cc.Msg) error {
	reqHeader := &proto.RequestHeader{}
	msg.WithServerReqHead(reqHeader)
	msg.WithSerializationType(codec.SerializationTypeJSON)

	urlPath := fc.Parser.Path
	if urlPath[0] == '/' {
		urlPath = urlPath[1:]
	}

	msg.WithCallRPCName(types.BytesToString(urlPath))

	initHeader(fc)

	ct, _ := fc.header["Content-Type"]
	if len(ct) == 0 {
		ct, _ = fc.header["content-type"]
	}

	if ct != "application/json" {
		return fmt.Errorf("content-type is not application/json")
	}

	reqHeader.RequestType = consts.RequestTypeHTTP
	reqHeader.Callee = msg.CallRPCName()

	if v, _ := fc.header[proto.HeaderVersion]; v != "" {
		i, _ := strconv.Atoi(v)
		reqHeader.Version = uint32(i)
	}

	if v := fc.header[proto.HeaderQueryMode]; v != "" {
		i, _ := strconv.Atoi(v)
		reqHeader.QueryMode = uint32(i)
	}
	if v := fc.header[proto.HeaderRequestID]; v != "" {
		i, _ := strconv.ParseUint(v, 10, 64)
		reqHeader.RequestId = i
		msg.WithRequestID(i)
	}
	if v := fc.header[proto.HeaderTraceID]; v != "" {
		reqHeader.TraceId = v
		msg.WithTraceID(v)
	}
	if v := fc.header[proto.HeaderTimestamp]; v != "" {
		reqHeader.Timestamp, _ = strconv.ParseUint(v, 10, 64)
	}
	if v := fc.header[proto.HeaderTimeout]; v != "" {
		i, _ := strconv.Atoi(v)
		reqHeader.Timeout = uint32(i)
		msg.WithRequestTimeout(time.Millisecond * time.Duration(i))
	}
	if v := fc.header[proto.HeaderCaller]; v != "" {
		reqHeader.Caller = v
		msg.WithCallerServiceName(v)
	}
	if v := fc.header[proto.HeaderAppid]; v != "" {
		reqHeader.Appid, _ = strconv.ParseUint(v, 10, 64)
	}

	reqHeader.Ip = util.GetIpFromAddr(msg.RemoteAddr())

	if v := fc.header[proto.HeaderAuthRand]; v != "" {
		i, _ := strconv.Atoi(v)
		reqHeader.AuthRand = uint32(i)
	}
	if v := fc.header[proto.HeaderSign]; v != "" {
		reqHeader.Sign = v
	}

	msg.WithServerRespHead(codec.GetRespFromReqHeader(reqHeader))
	return nil
}

// encode error to http error response buffer
func getErrorRespBuf(fc *frameCodec, e *errs.Error) []byte {
	respBuilder := strings.Builder{}
	respBuilder.Write(fc.Parser.Version)

	code, ok := errs.ErrToHTTPStatus[e.Code]
	if !ok {
		code = http.StatusInternalServerError
	}

	respBuilder.WriteString(fmt.Sprintf(" %d ", code))
	respBuilder.WriteString(wildcat.StatusText(code))
	respBuilder.WriteString("\r\nServer: ")
	respBuilder.WriteString(fc.opts.ServiceName)
	respBuilder.WriteString("\r\nDate: ")
	respBuilder.WriteString(time.Now().Format("Mon, 02 Jan 2006 15:04:05 GMT"))
	respBuilder.WriteString("\r\nX-Content-Type-Options: nosniff")
	respBuilder.WriteString("\r\nContent-Type: application/json")

	errMsg := types.QuickReplaceLFCR2Space(types.StringToBytes(e.Msg))
	respBuilder.WriteString("\r\n" + proto.HeaderErrorType + ": ")
	respBuilder.WriteString(strconv.Itoa(int(e.Type)))
	respBuilder.WriteString("\r\n" + proto.HeaderErrorCode + ": ")
	respBuilder.WriteString(strconv.Itoa(e.Code))
	respBuilder.WriteString("\r\n" + proto.HeaderErrorMessage + ": ")
	respBuilder.WriteString(errMsg)

	respBuilder.WriteString("\r\nContent-Length: 0\r\n\r\n")

	return types.StringToBytes(respBuilder.String())
}

// encode success respBody to http response buffer
func getRespBuf(fc *frameCodec, respBody []byte) []byte {
	respBuilder := strings.Builder{}
	respBuilder.Write(fc.Parser.Version)
	respBuilder.WriteString(" 200 OK \r\nServer: ")
	respBuilder.WriteString(fc.opts.ServiceName)
	respBuilder.WriteString("\r\nDate: ")
	respBuilder.WriteString(time.Now().Format("Mon, 02 Jan 2006 15:04:05 GMT"))
	respBuilder.WriteString("\r\nX-Content-Type-Options: nosniff\r\nContent-Type: application/json\r\nhead-error-code: 0\r\nContent-Length: ")
	respBuilder.WriteString(fmt.Sprint(len(respBody)))
	respBuilder.WriteString("\r\n\r\n")
	respBuilder.Write(respBody)

	return types.StringToBytes(respBuilder.String())
}

type frameCodec struct {
	opts    *transport.Options
	Parser  *wildcat.HTTPParser
	header  map[string]string
	buf     []byte
	bufLen  int
	headLen int
	bodyLen int
}

func (fc *frameCodec) resetBuf() {
	if len(fc.buf) > 0 {
		fc.buf = fc.buf[:0]
	}
	fc.bufLen = 0
	fc.headLen = 0
	fc.bodyLen = 0
}

func initHeader(fc *frameCodec) {
	fc.header = map[string]string{}
	for _, header := range fc.Parser.Headers {
		fc.header[types.BytesToString(header.Name)] = types.BytesToString(header.Value)
	}
}
