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

package rpc

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/horm-database/common/codec"
	"github.com/horm-database/common/crypto"
	"github.com/horm-database/common/errs"
	"github.com/horm-database/common/log"
	"github.com/horm-database/common/log/logger"
	"github.com/horm-database/common/metrics"
	cp "github.com/horm-database/common/proto"
	"github.com/horm-database/common/snowflake"
	"github.com/horm-database/common/types"
	"github.com/horm-database/server/consts"
	"github.com/horm-database/server/model/table"
	cc "github.com/horm-database/server/srv/codec"
	"github.com/horm-database/server/srv/transport"

	"github.com/golang/protobuf/proto"
	"github.com/panjf2000/gnet/v2"
)

// DefaultRpcTransport is the default implementation of ServerStreamTransport.
var DefaultRpcTransport = NewRpcTransport()

// transportRPC is the implementation details of server client, may be tcp or udp.
type transportRPC struct{}

// NewRpcTransport creates a new rpc transport.
func NewRpcTransport() transport.Transport {
	return &transportRPC{}
}

// Serve starts listening and serve
func (s *transportRPC) Serve(ctx context.Context, opts *transport.Options) (err error) {
	if opts.Network == "unix" || opts.Network == "tcp" || opts.Network == "tcp4" || opts.Network == "tcp6" {
		rs := &rpcServer{
			ctx:     ctx,
			address: fmt.Sprintf("%s://%s", opts.Network, opts.Address),
			opts:    opts,
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

		go func() {
			err = gnet.Run(rs, rs.address, serverOpts...)
		}()

		time.Sleep(10 * time.Millisecond)
		return
	}

	return fmt.Errorf("server client: not support network type %s", opts.Network)
}

// rpcServer A network event processing model based on gnet
type rpcServer struct {
	gnet.BuiltinEventEngine
	ctx       context.Context
	address   string
	opts      *transport.Options
	eng       gnet.Engine
	closeOnce sync.Once
}

func (r *rpcServer) OnBoot(eng gnet.Engine) gnet.Action {
	r.eng = eng

	go func() {
		<-r.ctx.Done()
		log.Debug(cc.GCtx, "receive server stop event")

		r.closeOnce.Do(
			func() {
				_ = eng.Stop(context.TODO())
			},
		)
	}()

	log.Infof(r.ctx, "rpc server listening on %s", r.address)
	return gnet.None
}

func (r *rpcServer) OnShutdown(_ gnet.Engine) {
}

func (r *rpcServer) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	c.SetContext(&frameCodec{opts: r.opts})
	return
}

func (r *rpcServer) OnClose(_ gnet.Conn, _ error) (action gnet.Action) {
	return
}

func (r *rpcServer) OnTraffic(c gnet.Conn) (action gnet.Action) {
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

	var reqBuf []byte

	if fc.bufLen == 0 { // 初次收包
		fc.frameType = buf[0]

		if fc.frameType == codec.FrameTypeSignature {
			fc.signFrameHead = codec.NewSignFrameHead()
			fc.signFrameHead.Extract(buf)
			fc.totalLen = int(fc.signFrameHead.TotalLen)
		} else if fc.frameType == codec.FrameTypeEncrypt {
			fc.encryptFrameHead = codec.NewEncryptFrameHead()
			fc.encryptFrameHead.Extract(buf)
			fc.totalLen = int(fc.encryptFrameHead.TotalLen)
		} else {
			fc.frameHead = codec.NewFrameHead()
			fc.frameHead.Extract(buf)
			fc.totalLen = int(fc.frameHead.TotalLen)
		}

		fc.bufLen = bufLen

		if bufLen < fc.totalLen { // 处理超大被拆分的数据包
			fc.buf = make([]byte, fc.totalLen)
			copy(fc.buf, buf)
			return gnet.None
		}

		fc.buf = buf
	} else { // 继续接收超大被拆分的数据包
		if bufLen+fc.bufLen < fc.totalLen { // 继续处理超大被拆分的数据包
			copy(fc.buf[fc.bufLen:], buf)
			fc.bufLen = fc.bufLen + bufLen
			return gnet.None
		} else if bufLen+fc.bufLen > fc.totalLen { // 包大小异常
			return writeError(c, fc, errors.New("body length is invalid"))
		} else {
			copy(fc.buf[fc.bufLen:], buf)
			fc.bufLen = fc.bufLen + bufLen
		}
	}

	workspace := table.GetWorkspace()
	if workspace.EnforceSign == consts.WorkspaceEnforceSignYes &&
		(fc.frameType != codec.FrameTypeSignature && fc.frameType != codec.FrameTypeEncrypt) {
		return writeError(c, fc, errors.New("enforce signature, buf input frame is not signature and encrypt"))
	}

	if fc.frameType == codec.FrameTypeSignature {
		if fc.signFrameHead.WorkSpaceID != uint32(workspace.Id) {
			return writeError(c, fc, errors.New("workspace get from signature frame is illegal"))
		}

		frameBuf := fc.buf[codec.SignFrameHeadLen:]

		if workspace.EnforceSign == consts.WorkspaceEnforceSignYes { // 校验签名
			md5Bytes := append(types.StringToBytes(workspace.Token), frameBuf...)
			if bytes.Compare(crypto.MD5Bytes(md5Bytes), fc.signFrameHead.Sign) != 0 {
				// frame signature verification failed
				return writeError(c, fc, errors.New("frame signature failed"))
			}
		}

		fc.frameHead = codec.NewFrameHead()
		fc.frameHead.Extract(frameBuf)
		reqBuf = frameBuf[codec.FrameHeadLen:]

		if fc.frameHead.TotalLen != uint32(len(reqBuf))+codec.FrameHeadLen {
			return writeError(c, fc, errors.New("signature frame request buffer length is invalid"))
		}
	} else if fc.frameType == codec.FrameTypeEncrypt {
		if fc.encryptFrameHead.WorkspaceID != uint32(workspace.Id) {
			return writeError(c, fc, errors.New("workspace get from encrypt frame is illegal"))
		}

		encryptFrameBuf := buf[codec.EncryptFrameHeadLen:]
		frameBuf, err := aesDecrypt(encryptFrameBuf, types.StringToBytes(workspace.Token))
		if err != nil {
			return writeError(c, fc, fmt.Errorf("frame buffer AES decrypt failed: %v", err))
		}

		fc.frameHead = codec.NewFrameHead()
		fc.frameHead.Extract(frameBuf)
		reqBuf = frameBuf[codec.FrameHeadLen:]

		if fc.frameHead.TotalLen != uint32(len(reqBuf))+codec.FrameHeadLen {
			return writeError(c, fc, errors.New("AES decrypt failed, request buffer length is invalid"))
		}
	} else {
		reqBuf = fc.buf[codec.FrameHeadLen:]
	}

	return r.handle(c, fc, reqBuf)
}

func (r *rpcServer) handle(c gnet.Conn, fc *frameCodec, reqBuf []byte) gnet.Action {
	ctx, msg := codec.NewMessage(r.ctx)

	defer func() {
		codec.RecycleMessage(msg)
		fc.resetBuf()
	}()

	msg.WithFrameCodec(fc)
	msg.WithLocalAddr(c.LocalAddr())
	msg.WithRemoteAddr(c.RemoteAddr())
	msg.WithSpanID(snowflake.GenerateID())

	rsp, err := r.opts.Handler.Handle(ctx, reqBuf)
	if err != nil {
		metrics.TCPServerTransportHandleFail.Incr()
		log.Debug(cc.GCtx, "client: tcpConn serve handle fail ", err)
		return gnet.Close
	}

	if len(rsp) > 0 {
		if _, err = c.Write(rsp); err != nil {
			metrics.TCPServerTransportWriteFail.Incr()
			log.Debug(cc.GCtx, "client: tcpConn write fail ", err)
			return gnet.Close
		}
	}

	return gnet.None
}

func aesDecrypt(cryted, key []byte) (buf []byte, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("AES Decrypt Panic: %v", e)
		}
	}()

	dbuf := make([]byte, base64.StdEncoding.DecodedLen(len(cryted)))
	n, err := base64.StdEncoding.Decode(dbuf, cryted)
	if err != nil {
		return nil, err
	}

	crytedByte := dbuf[:n]
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	buf = make([]byte, len(crytedByte))
	blockMode.CryptBlocks(buf, crytedByte)
	buf = pkcs7UnPadding(buf)
	return buf, nil
}

// 去码
func pkcs7UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

func writeError(c gnet.Conn, fc *frameCodec, err error) gnet.Action {
	frameHead := codec.NewFrameHead()

	respHeader := cp.ResponseHeader{
		Err: &cp.Error{
			Type: int32(errs.ETypeSystem),
			Code: errs.ErrServerReadFrame,
			Msg:  err.Error(),
		},
	}

	fc.resetBuf()

	respHeaderBuf, err := proto.Marshal(&respHeader)
	if err != nil {
		return gnet.Close
	}

	errRespBuf, err := frameHead.Construct(respHeaderBuf, nil)
	if err == nil {
		c.Write(errRespBuf)
	}

	return gnet.Close
}
