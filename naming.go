package naming

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/istepheny/naming/app"
	"github.com/istepheny/naming/balancer"
	"github.com/istepheny/naming/balancer/random"
	"github.com/istepheny/naming/config"
	"github.com/istepheny/naming/container"
	"github.com/istepheny/naming/registry"
	"github.com/jpillora/backoff"
)

const ErrMaxRetry = 3

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

func (n *Naming) Register(app *app.App) (notifyChan chan *registry.NotifyMessage, err error) {
	b := &backoff.Backoff{}

	for i := 0; i < ErrMaxRetry; i++ {
		notifyChan, err = n.registry.Register(app)
		if err != nil {
			log.Printf("naming: Register error: %v, app: %v", err, app.String())
			d := b.Duration()
			time.Sleep(d)
			continue
		}
		b.Reset()
		return
	}

	b.Reset()
	return
}

func (n *Naming) Deregister(app *app.App) (err error) {
	b := &backoff.Backoff{}

	for i := 0; i < ErrMaxRetry; i++ {
		err = n.registry.Deregister(app)
		if err != nil {
			d := b.Duration()
			log.Printf("naming: Deregister error: %v, app: %v", err, app.String())
			time.Sleep(d)
			continue
		}
		b.Reset()
		return
	}

	b.Reset()
	return
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
	b := &backoff.Backoff{}

	for {
		err := n.watchLoop(appName)
		if err != nil {
			d := b.Duration()
			time.Sleep(d)
			continue
		}
		b.Reset()
	}
}

func (n *Naming) watchLoop(appName string) error {
	watchChan, err := n.registry.Watch(appName)
	if err != nil {
		return err
	}

	for {
		select {
		case watchResponse, ok := <-watchChan:
			if !ok {
				return errors.New("naming: watch chan closed")
			}

			if watchResponse.Canceled == true {
				return fmt.Errorf("naming: watch chan closed, %v", watchResponse.Error)
			}

			if watchResponse.Error != nil {
				log.Printf("naming: watch response error: %v", err)
				continue
			}

			n.container.Set(appName, watchResponse.Apps)
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
