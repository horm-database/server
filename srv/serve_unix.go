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
