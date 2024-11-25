package srv

import (
	"context"
	"fmt"
	"os"
	"sync/atomic"
	"time"

	cc "github.com/horm-database/common/codec"
	"github.com/horm-database/common/errs"
	"github.com/horm-database/common/json"
	"github.com/horm-database/common/log"
	"github.com/horm-database/common/log/logger"
	"github.com/horm-database/common/metrics"
	"github.com/horm-database/common/proto"
	"github.com/horm-database/common/types"
	"github.com/horm-database/server/srv/codec"
	"github.com/horm-database/server/srv/naming"
	"github.com/horm-database/server/srv/transport"
)

// Service is the interface that provides services.
type Service interface {
	// Register registers service handle func
	Register(funcs []Func) error
	// Serve start serving.
	Serve() error
	// Close stop serving.
	Close(chan struct{})
}

// Options are server side options.
type Options struct {
	Machine     string        // 机器名（容器名）
	Env         string        // 环境名
	ServiceName string        // 服务名
	Address     string        // 监听地址
	Protocol    string        // 服务协议，如 rpc、http、web
	Timeout     time.Duration // 请求处理超时时间

	Registry         naming.Registry     // 名字服务注册接口
	Transport        transport.Transport // 传输层
	TransportOptions transport.Options   // 传输层配置
	Codec            codec.Codec         // 编解码接口

	CloseWaitTime    time.Duration // 注销名字服务之后的等待时间，让名字服务更新实例列表。 (单位 ms) 最大: 10s.
	MaxCloseWaitTime time.Duration // 进程结束之前等待请求完成的最大等待时间。(单位 ms)
}

// Func provides the information of an RPC Method.
type Func struct {
	Name    string
	Handler Handler
}

// Handler is the default handler.
type Handler func(ctx context.Context,
	reqHeader *proto.RequestHeader, reqBodyBuf []byte) (rsp interface{}, err error)

// service is an implementation of Service
type service struct {
	ctx         context.Context    // context of this service
	cancel      context.CancelFunc // function that cancels this service
	opts        *Options           // options of this service
	handlers    map[string]Handler // func => handler
	activeCount int64              // 在设置了 MaxCloseWaitTime 参数时，优雅关闭的活跃请求数
}

// New create service
var New = func(opts *Options) Service {
	s := &service{
		opts:     opts,
		handlers: make(map[string]Handler),
	}

	s.opts.TransportOptions.Handler = s

	s.ctx, s.cancel = context.WithCancel(context.Background())
	return s
}

// Serve implements Service, starting serving.
func (s *service) Serve() error {
	pid := os.Getpid()

	if err := s.opts.Transport.Serve(s.ctx, &s.opts.TransportOptions); err != nil {
		log.Errorf(codec.GCtx, errs.RetSystem,
			"[%d] service %s ListenAndServe fail:%v", pid, s.opts.ServiceName, err)
		return err
	}

	if s.opts.Registry != nil {
		if err := s.opts.Registry.Register(s.opts.ServiceName, s.opts.Address); err != nil {
			// if registry fails, service needs to be closed and error should be returned.
			log.Errorf(codec.GCtx, errs.RetSystem,
				"[%d] service %s register fail: %v", pid, s.opts.ServiceName, err)
			return err
		}
	}

	log.Infof(codec.GCtx, "[%d] %s service %s start success, listening on [%s] ...",
		pid, s.opts.Protocol, s.opts.ServiceName, s.opts.Address)

	<-s.ctx.Done()
	return nil
}

// Handle implements transport.Handler.
func (s *service) Handle(ctx context.Context, reqBuf []byte) (respBuf []byte, err error) {
	if s.opts.MaxCloseWaitTime > s.opts.CloseWaitTime {
		atomic.AddInt64(&s.activeCount, 1)
		defer atomic.AddInt64(&s.activeCount, -1)
	}

	msg := cc.Message(ctx)

	reqBodyBuf, err := s.decode(msg, reqBuf)
	if err != nil {
		return s.encode(msg, nil, err)
	}

	// ServerRespError is already set, response error to client.
	if err := msg.ServerRespError(); err != nil {
		return s.encode(msg, nil, err)
	}

	respBody, err := s.handle(ctx, msg, reqBodyBuf)
	if err != nil {
		// failed to handle, should respond to database with error code, ignore respBody.
		metrics.ServiceHandleFail.Incr()
		return s.encode(msg, nil, err)
	}

	return s.handleResponse(ctx, msg, respBody)
}

// Register implements Service interface, registering a proto service impl for the service.
func (s *service) Register(funcs []Func) error {
	for _, f := range funcs {
		if _, ok := s.handlers[f.Name]; ok {
			continue
		}

		s.handlers[f.Name] = f.Handler
	}

	return nil
}

// Close closes the service，registry.Deregister will be called.
func (s *service) Close(ch chan struct{}) {
	closeWaitTime := s.opts.MaxCloseWaitTime
	if closeWaitTime < maxCloseWaitTime {
		closeWaitTime = maxCloseWaitTime
	}

	pid := os.Getpid()

	log.Infof(codec.GCtx, "[%d] %s service %s, closing ...", pid, s.opts.Protocol, s.opts.ServiceName)

	if s.opts.Registry != nil {
		if err := s.opts.Registry.Deregister(s.opts.ServiceName); err != nil {
			log.Errorf(codec.GCtx, errs.RetSystem,
				"[%d] deregister service %s fail: %s", pid, s.opts.ServiceName, err.Error())
		}
	}

	waitingTime := s.waitBeforeClose()

	// the remaining waiting time is the max_close_wait_time minus the time already waiting
	remainingTime := closeWaitTime - waitingTime
	if remainingTime > 0 {
		time.Sleep(remainingTime)
	}

	// this will cancel all children ctx.
	s.cancel()

	time.Sleep(100 * time.Millisecond)

	log.Infof(codec.GCtx, "[%d] %s service %s, closed", pid, s.opts.Protocol, s.opts.ServiceName)

	ch <- struct{}{}
	return
}

func (s *service) waitBeforeClose() (waitingTime time.Duration) {
	if s.opts.CloseWaitTime > 0 {
		// After registry.Deregister() is called,
		// sleep a while to let Naming Service finish updating instance list.
		// Otherwise, request would still arrive while the service had already been closed.
		log.Infof(codec.GCtx, "[%d] service %s remain %d requests wait %v time when closing service",
			os.Getpid(), s.opts.ServiceName, atomic.LoadInt64(&s.activeCount), s.opts.CloseWaitTime)

		waitingTime += s.opts.CloseWaitTime
		time.Sleep(s.opts.CloseWaitTime)
	}

	const sleepTime = 100 * time.Millisecond

	// wait for all active requests to be finished.
	if s.opts.MaxCloseWaitTime > s.opts.CloseWaitTime {
		spinCount := int((s.opts.MaxCloseWaitTime - s.opts.CloseWaitTime) / sleepTime)
		for i := 0; i < spinCount; i++ {
			if atomic.LoadInt64(&s.activeCount) <= 0 {
				break
			}

			waitingTime = waitingTime + sleepTime
			time.Sleep(sleepTime)
		}

		log.Infof(codec.GCtx, "[%d] service %s remain %d requests when closing service",
			os.Getpid(), s.opts.ServiceName, atomic.LoadInt64(&s.activeCount))
	}

	return
}

func (s *service) decode(msg *cc.Msg, reqBuf []byte) ([]byte, error) {
	reqBodyBuf, err := s.opts.Codec.Decode(msg, reqBuf)
	if err != nil {
		metrics.ServiceCodecDecodeFail.Incr()
		return nil, errs.Newf(errs.RetServerDecodeFail, "service codec Decode: %v", err)
	}

	msg.WithEnv(s.opts.Env)
	msg.WithCalleeServiceName(s.opts.ServiceName)

	return reqBodyBuf, nil
}
func (s *service) encode(msg *cc.Msg, respBodyBuf []byte, e error) (respBuf []byte, err error) {
	if e != nil {
		msg.WithServerRespError(e)
	}

	respBuf, err = s.opts.Codec.Encode(msg, respBodyBuf)
	if err != nil {
		metrics.ServiceCodecEncodeFail.Incr()
		log.Errorf(msg.Context(), 1, "service %s encode fail: %v", s.opts.ServiceName, err)
		return nil, err
	}

	return respBuf, nil
}

// handleResponse handles response.
func (s *service) handleResponse(ctx context.Context, msg *cc.Msg, respBody interface{}) ([]byte, error) {
	// serialize response body
	respBodyBuf, err := codec.Serialize(ctx, respBody)
	if err != nil {
		metrics.ServiceCodecMarshalFail.Incr()
		err = errs.Newf(errs.RetServerEncodeFail, "service codec Marshal: %v", err)

		// respBodyBuf will be nil if marshalling fails, respond only error code to database.
		return s.encode(msg, respBodyBuf, err)
	}

	return s.encode(msg, respBodyBuf, nil)
}

func (s *service) handle(ctx context.Context, msg *cc.Msg, reqBodyBuf []byte) (interface{}, error) {
	handler, ok := s.handlers[msg.CalleeMethod()]
	if !ok {
		handler, ok = s.handlers["default"] // 默认路由
		if !ok {
			metrics.ServiceHandleRPCNameInvalid.Incr()
			return nil, errs.New(errs.RetServerNoFunc,
				fmt.Sprintf("service handle: rpc name %s invalid, current service:%s",
					msg.CalleeMethod(), msg.CalleeServiceName()))
		}
	}

	timeout := s.opts.Timeout

	reqTimeout := msg.RequestTimeout()
	if reqTimeout > 0 { // 请求参数的超时和服务端配置的超时，取最小值。
		if reqTimeout < timeout || timeout == 0 {
			timeout = reqTimeout
		}
	}

	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	return apiHandle(ctx, handler, msg, reqBodyBuf)
}

// apiHandle 接口处理，函数名切勿修改，避免 log trace 路径异常
func apiHandle(ctx context.Context, handler Handler, msg *cc.Msg, reqBody []byte) (rsp interface{}, err error) {
	reqHeader := msg.ServerReqHead().(*proto.RequestHeader)

	initLogger(msg, reqHeader)

	log.InfoWith(ctx, []logger.Field{{"type", "HEADER"}}, json.MarshalToString(reqHeader, json.EncodeTypeFast))
	log.InfoWith(ctx, []logger.Field{{"type", "REQUEST"}}, types.QuickReplaceLFCR2Space(reqBody))

	start := time.Now()

	defer func() {
		if e := recover(); e != nil {
			metrics.IncrCounter("PanicNum", 1)
			err = errs.New(errs.RetPanic, fmt.Sprint(e))
		}

		fields := []logger.Field{}
		fields = append(fields, logger.Field{"type", "RESPONSE"})
		fields = append(fields, logger.Field{"during", time.Since(start).Milliseconds()})
		fields = append(fields, logger.Field{"seq", msg.LogSeq()})

		if err != nil {
			fields = append(fields, logger.Field{"code", errs.Code(err)})
			fields = append(fields, logger.Field{"files", "srv/service.go func=apiHandle()"})
			log.GetLogger(msg).Error(err.Error(), fields...)
		} else {
			respStr, _ := json.MarshalBaseToString(rsp, json.EncodeTypeFast)
			log.GetLogger(msg).Info(respStr, fields...)
		}
	}()

	rsp, err = handler(ctx, reqHeader, reqBody)
	if err != nil {
		return
	}

	respHeader, ok := msg.ServerRespHead().(*proto.ResponseHeader)
	if !ok {
		respHeader = codec.GetRespFromReqHeader(reqHeader)
		msg.WithServerRespHead(respHeader)
	}

	resp, ok := rsp.(*proto.QueryResp)
	if ok {
		respHeader.IsNil = resp.IsNil
		respHeader.RspErrs = resp.RspErrs
		respHeader.RspNils = resp.RspNils
		rsp = resp.RspData
	}

	return
}

func initLogger(msg *cc.Msg, h *proto.RequestHeader) {
	remote := ""
	if msg.RemoteAddr() != nil {
		remote = msg.RemoteAddr().String()
	}

	l := msg.Logger()
	if l == nil {
		l = logger.DefaultLogger
	}

	filed := []logger.Field{}
	filed = append(filed, logger.Field{"callee", msg.CalleeServiceName()})
	filed = append(filed, logger.Field{"remote", remote})
	filed = append(filed, logger.Field{"request_id", msg.RequestID()})
	filed = append(filed, logger.Field{"trace_id", msg.TraceID()})
	filed = append(filed, logger.Field{"span_id", msg.SpanID()})
	filed = append(filed, logger.Field{"machine", Config().Machine})
	filed = append(filed, logger.Field{"appid", h.Appid})

	if msg.CallerServiceName() != "" {
		filed = append(filed, logger.Field{"caller", msg.CallerServiceName()})
	}

	msg.WithLogger(l.With(filed...))
}
