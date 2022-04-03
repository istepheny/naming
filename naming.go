package naming

import (
	"sync"
	"time"

	"git.ucloudadmin.com/monkey/naming/app"
	"git.ucloudadmin.com/monkey/naming/balancer"
	"git.ucloudadmin.com/monkey/naming/balancer/random"
	"git.ucloudadmin.com/monkey/naming/config"
	"git.ucloudadmin.com/monkey/naming/container"
	"git.ucloudadmin.com/monkey/naming/registry"
)

type Naming struct {
	mu        sync.RWMutex
	watched   map[string]struct{}
	registry  registry.Registry
	balancer  balancer.Balancer
	container *container.Container
}

func New(config config.Config) *Naming {
	return &Naming{
		watched:   make(map[string]struct{}),
		registry:  registry.Get(config.Driver, config),
		balancer:  balancer.Get(random.Driver),
		container: container.New(),
	}
}

func (n *Naming) Register(app *app.App) (err error) {
	return n.registry.Register(app)
}

func (n *Naming) Deregister(app *app.App) (err error) {
	return n.registry.Deregister(app)
}

func (n *Naming) Discover(appName string) (a *app.App, err error) {
	apps, err := n.DiscoverAll(appName)
	if err != nil {
		return nil, err
	}

	return n.balancer.Pick(apps)
}

func (n *Naming) DiscoverAll(appName string) (apps []*app.App, err error) {
	if n.isWatched(appName) == false {
		n.mu.Lock()
		n.watched[appName] = struct{}{}
		n.mu.Unlock()

		go n.watch(appName)
		go n.sync(appName)
	}

	apps = n.container.Get(appName)
	if len(apps) > 0 {
		return apps, nil
	}

	apps, err = n.syncContainer(appName)
	if err != nil {
		return nil, err
	}

	return apps, nil
}

func (n *Naming) watch(appName string) {
	watchChan, err := n.registry.Watch(appName)
	if err != nil {
		return
	}

	for {
		select {
		case appsMap, ok := <-watchChan:
			if !ok {
				continue
			}
			for appName, apps := range appsMap {
				n.container.Set(appName, apps)
			}
		}
	}
}

func (n *Naming) sync(appName string) {
	t := time.NewTicker(60 * time.Second)

	for {
		select {
		case <-t.C:
			_, _ = n.syncContainer(appName)
		}
	}
}

func (n *Naming) syncContainer(appName string) (apps []*app.App, err error) {
	apps, err = n.registry.Discover(appName)
	if err != nil {
		return nil, err
	}

	n.container.Set(appName, apps)

	return apps, nil
}

func (n *Naming) isWatched(appName string) bool {
	n.mu.RLock()
	_, ok := n.watched[appName]
	n.mu.RUnlock()
	return ok
}
