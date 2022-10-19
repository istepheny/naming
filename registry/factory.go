package registry

import (
	"sync"

	"github.com/istepheny/naming/config"
)

type factory func(config config.Config) Registry

var (
	mu        sync.RWMutex
	factories = make(map[string]factory)
)

func Set(name string, factory factory) {
	mu.Lock()
	defer mu.Unlock()
	if factory == nil {
		panic("set Registry is nil")
	}

	if _, dup := factories[name]; dup {
		panic("set called twice for Registry " + name)
	}

	factories[name] = factory
}

func Get(name string, config config.Config) Registry {
	mu.RLock()
	defer mu.RUnlock()
	factory, ok := factories[name]
	if !ok {
		panic("Registry " + name + " is not available")
	}
	return factory(config)
}
