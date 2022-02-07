package naming

import (
	"sync"

	"git.ucloudadmin.com/monkey/naming/app"
	"git.ucloudadmin.com/monkey/naming/balancer"
	"git.ucloudadmin.com/monkey/naming/balancer/random"
	"git.ucloudadmin.com/monkey/naming/config"
	"git.ucloudadmin.com/monkey/naming/container"
	"git.ucloudadmin.com/monkey/naming/registry"
)

type Naming struct {
	once      sync.Once
	registry  registry.Registry
	balancer  balancer.Balancer
	container *container.Container
}

func New(config config.Config) *Naming {
	return &Naming{
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
	n.once.Do(func() {
		go n.watch()
	})

	var apps []*app.App
	apps = n.container.Get(appName)
	if len(apps) > 0 {
		return n.balancer.Pick(apps)
	}

	apps, err = n.registry.Discover(appName)
	if err != nil {
		return nil, err
	}

	n.container.Set(appName, apps)
	return n.balancer.Pick(apps)
}

func (n *Naming) watch() {
	watchChan, err := n.registry.Watch()
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
