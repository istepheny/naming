package balancer

import "sync"

type factory func() Balancer

var (
	mu        sync.RWMutex
	factories = make(map[string]factory)
)

func Set(name string, factory factory) {
	mu.Lock()
	defer mu.Unlock()
	if factory == nil {
		panic("register Balancer is nil")
	}

	if _, dup := factories[name]; dup {
		panic("register called twice for Balancer " + name)
	}

	factories[name] = factory
}

func Get(name string) Balancer {
	mu.RLock()
	defer mu.RUnlock()
	factory, ok := factories[name]
	if !ok {
		panic("Balancer " + name + " is not available")
	}
	return factory()
}
