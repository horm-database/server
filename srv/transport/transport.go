// Package transport is the network client layer. It is only used for basic binary body network
// communication without any business logic.
// By default, there is only one pluggable ServerTransport.
package transport

import (
	"context"
	"net"
	"time"
)

// Transport defines the server client layer interface.
type Transport interface {
	Serve(ctx context.Context, opts *Options) error
}

// Handler is the process function when server client receive a package.
type Handler interface {
	Handle(ctx context.Context, req []byte) (rsp []byte, err error)
}

// Options is the server options on start.
type Options struct {
	ServiceName string
	Protocol    string
	Address     string
	Network     string
	Handler     Handler
	Listener    net.Listener

	EventLoopNum    int           // epoll loop 大小，默认取 CPU 核数
	IdleTimeout     time.Duration // 连接最大空闲时间
	KeepAlivePeriod time.Duration
	EnableH2C       bool

	CACertFile  string // ca certification file
	TLSCertFile string // server certification file
	TLSKeyFile  string // server key file
}
