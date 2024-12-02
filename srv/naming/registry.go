// Package registry registers servers. A server report itself on start.
package naming

import (
	"sync"
)

// Registry is the interface that defines a register.
type Registry interface {
	Register(service string, address string) error
	Deregister(service string) error
}

var (
	registries = make(map[string]Registry)
	lock       = sync.RWMutex{}
)

// Register registers a named registry. Each service has its own registry.
func Register(name string, s Registry) {
	lock.Lock()
	registries[name] = s
	lock.Unlock()
}

// Get gets a named registry.
func Get(name string) Registry {
	lock.RLock()
	r := registries[name]
	lock.RUnlock()
	return r
}
