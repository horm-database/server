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

// Package http provides support for http protocol by default,
// provides rpc server with http protocol, and provides rpc database
// for calling http protocol.
package http

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/horm-database/common/codec"
	"github.com/horm-database/common/errs"
	"github.com/horm-database/common/log"
	"github.com/horm-database/common/log/logger"
	"github.com/horm-database/common/proto"
	"github.com/horm-database/common/snowflake"
	"github.com/horm-database/common/types"
	"github.com/horm-database/server/consts"
	"github.com/horm-database/server/model/table"
	cc "github.com/horm-database/server/srv/codec"
	"github.com/horm-database/server/srv/transport"

	"github.com/evanphx/wildcat"
	"github.com/panjf2000/gnet/v2"
)

// DefaultHttpTransport default server http transport.
var DefaultHttpTransport = NewHttpTransport()

type transportHttp struct{}

// NewHttpTransport create a new http transport.
func NewHttpTransport() transport.Transport {
	return &transportHttp{}
}

// Serve starts listening and serve
func (t *transportHttp) Serve(ctx context.Context, opts *transport.Options) (err error) {
	if opts.Handler == nil {
		return errors.New("http server client handler empty")
	}

	serverOpts := []gnet.Option{
		gnet.WithMulticore(true),
		gnet.WithReusePort(true),
		gnet.WithLogger(logger.DefaultLogger),
	}

	if opts.EventLoopNum > 0 {
		serverOpts = append(serverOpts, gnet.WithNumEventLoop(opts.EventLoopNum))
	}

	if opts.KeepAlivePeriod > 0 {
		serverOpts = append(serverOpts, gnet.WithTCPKeepAlive(opts.KeepAlivePeriod))
	}

	hs := &httpServer{
		ctx:     ctx,
		address: fmt.Sprintf("tcp://%s", opts.Address),
		opts:    opts,
	}

	go func() {
		err = gnet.Run(hs, hs.address, serverOpts...)
	}()

	time.Sleep(10 * time.Millisecond)
	return nil
}

type httpServer struct {
	gnet.BuiltinEventEngine
	ctx       context.Context
	address   string
	opts      *transport.Options
	eng       gnet.Engine
	closeOnce sync.Once
}

func (h *httpServer) OnBoot(eng gnet.Engine) gnet.Action {
	h.eng = eng

	go func() {
		<-h.ctx.Done()
		log.Debug(cc.GCtx, "receive server stop event")

		h.closeOnce.Do(
			func() {
				_ = eng.Stop(context.TODO())
			},
		)
	}()

	log.Infof(h.ctx, "http server listening on %s", h.address)
	return gnet.None
}

func (h *httpServer) OnShutdown(_ gnet.Engine) {
}

func (h *httpServer) OnOpen(c gnet.Conn) ([]byte, gnet.Action) {
	c.SetContext(&frameCodec{Parser: wildcat.NewHTTPParser(), opts: h.opts})
	return nil, gnet.None
}

func (h *httpServer) OnClose(_ gnet.Conn, _ error) (action gnet.Action) {
	return
}

func (h *httpServer) OnTraffic(c gnet.Conn) (action gnet.Action) {
	fc := c.Context().(*frameCodec)

	defer func() {
		if e := recover(); e != nil {
			action = writeError(c, fc, fmt.Errorf("receive buf panic: %v", e))
		}
	}()

	buf, err := c.Next(-1)
	if err != nil {
		return writeError(c, fc, fmt.Errorf("next buf error: %v", err))
	}

	bufLen := len(buf)
	if bufLen == 0 {
		return writeError(c, fc, errors.New("receive empty buf"))
	}

	var body []byte

	if fc.bufLen == 0 { // 初次收包
		headLen, err := fc.Parser.Parse(buf)
		if err != nil {
			return writeError(c, fc, fmt.Errorf("http parse error: %v", err))
		}

		bodyLen := int(fc.Parser.ContentLength())
		if bodyLen < 0 {
			return writeError(c, fc, errors.New("body is empty"))
		}

		fc.headLen = headLen
		fc.bodyLen = bodyLen
		fc.bufLen = fc.bufLen + bufLen

		if bufLen < headLen+bodyLen { // 处理超大被拆分的数据包
			fc.buf = make([]byte, headLen+bodyLen)
			copy(fc.buf, buf)
			return gnet.None
		}

		body = buf[fc.headLen:]
	} else { // 继续接收超大被拆分的数据包
		if bufLen+fc.bufLen < fc.headLen+fc.bodyLen { // 继续处理超大被拆分的数据包
			copy(fc.buf[fc.bufLen:], buf)
			fc.bufLen = fc.bufLen + bufLen
			return gnet.None
		} else if bufLen+fc.bufLen > fc.headLen+fc.bodyLen { // 包大小异常
			return writeError(c, fc, errors.New("body length is invalid"))
		} else {
			copy(fc.buf[fc.bufLen:], buf)
			fc.bufLen = fc.bufLen + bufLen
		}

		copy(fc.buf[fc.bufLen:], buf)
		_, err = fc.Parser.Parse(fc.buf)
		if err != nil {
			return writeError(c, fc, fmt.Errorf("http parse error: %v", err))
		}

		body = fc.buf[fc.headLen:]
	}

	workspace := table.GetWorkspace()
	if workspace.EnforceSign == consts.WorkspaceEnforceSignYes {
		return writeError(c, fc, errors.New("workspace enforce signature, not support http"))
	}

	return h.handle(c, fc, body)
}

func (h *httpServer) handle(c gnet.Conn, fc *frameCodec, body []byte) gnet.Action {
	ctx, msg := codec.NewMessage(h.ctx)

	defer func() {
		codec.RecycleMessage(msg)
		fc.resetBuf()
	}()

	msg.WithFrameCodec(fc)
	msg.WithLocalAddr(c.LocalAddr())
	msg.WithRemoteAddr(c.RemoteAddr())
	msg.WithSpanID(snowflake.GenerateID())

	rsp, err := h.opts.Handler.Handle(ctx, body)
	if err != nil {
		log.Error(cc.GCtx, errs.Code(err), "http server handle fail:", err)
		return writeError(c, fc, err)
	}

	_, err = c.Write(rsp)
	if err != nil {
		return gnet.Close
	}

	return gnet.None
}

func writeError(c gnet.Conn, fc *frameCodec, err error) gnet.Action {
	respBuilder := strings.Builder{}
	respBuilder.WriteString("HTTP/1.1 500 Internal Server Error\r\nServer: http.")
	respBuilder.WriteString(fc.opts.ServiceName)
	respBuilder.WriteString("\r\nDate: ")
	respBuilder.WriteString(time.Now().Format("Mon, 02 Jan 2006 15:04:05 GMT"))
	respBuilder.WriteString("\r\nX-Content-Type-Options: nosniff\r\nContent-Type: application/json")

	errMsg := types.QuickReplaceLFCR2Space(types.StringToBytes(errs.Msg(err)))

	respBuilder.WriteString("\r\n" + proto.HeaderErrorType + ": ")
	respBuilder.WriteString(strconv.Itoa(int(errs.ETypeSystem)))
	respBuilder.WriteString("\r\n" + proto.HeaderErrorCode + ": ")
	respBuilder.WriteString(strconv.Itoa(errs.ErrServerReadFrame))
	respBuilder.WriteString("\r\n" + proto.HeaderErrorMessage + ": ")
	respBuilder.WriteString(errMsg)

	respBuilder.WriteString("\r\nContent-Length: 0\r\n\r\n")

	fc.resetBuf()

	c.Write(types.StringToBytes(respBuilder.String()))
	return gnet.Close
}
