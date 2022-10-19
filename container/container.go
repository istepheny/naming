package container

import (
	"sync"

	"github.com/istepheny/naming/app"
)

type Container struct {
	mu        sync.RWMutex
	container map[string][]*app.App
}

func New() *Container {
	return &Container{
		container: make(map[string][]*app.App),
	}
}

func (c *Container) Set(appName string, apps []*app.App) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.container[appName] = apps
}

func (c *Container) Get(appName string) []*app.App {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.container[appName]
}
