// Copyright (c) 2024 The horm-database Authors (such as CaoHao <18500482693@163.com>). All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//go:build !windows
// +build !windows

package srv

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/horm-database/common/log/logger"
)

// DefaultServerCloseSIG are signals that close server.
var DefaultServerCloseSIG = []os.Signal{syscall.SIGINT, syscall.SIGTERM, syscall.SIGSEGV, syscall.SIGUSR2}

// Serve implements Service, starting all services that belong to the server.
func (s *Server) Serve() error {
	defer logger.Sync()

	if len(s.services) == 0 {
		panic("not have any service")
	}

	s.signalCh = make(chan os.Signal)

	var serveErr error

	for name, svc := range s.services {
		go func(name string, svc Service) {
			if err := svc.Serve(); err != nil {
				serveErr = err
				s.failedServices.Store(name, svc)
				time.Sleep(300 * time.Millisecond)
				s.signalCh <- syscall.SIGTERM
			}
		}(name, svc)
	}

	signal.Notify(s.signalCh, DefaultServerCloseSIG...)

	select {
	case <-s.signalCh:
	}

	// close server.
	s.Close()

	if serveErr != nil {
		panic(serveErr)
	}

	return nil
}
