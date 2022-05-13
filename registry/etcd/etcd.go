package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"

	"git.ucloudadmin.com/monkey/naming/app"
	"git.ucloudadmin.com/monkey/naming/config"
	"git.ucloudadmin.com/monkey/naming/registry"
)

const (
	Driver = "etcd"
	prefix = "/naming"
)

func init() {
	registry.Set(Driver, NewEtcd)
}

type Etcd struct {
	ctx    context.Context
	cancel context.CancelFunc

	client    *clientv3.Client
	lease     clientv3.Lease
	watchChan chan *registry.WatchResponse

	mu           sync.RWMutex
	registerApps map[string]struct{}
}

func NewEtcd(c config.Config) registry.Registry {
	client, err := clientv3.New(clientv3.Config{
		Username:             c.Username,
		Password:             c.Password,
		Endpoints:            c.Servers,
		DialTimeout:          5 * time.Second,
		DialKeepAliveTime:    5 * time.Second,
		DialKeepAliveTimeout: 5 * time.Second,
	})

	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Etcd{
		ctx:          ctx,
		cancel:       cancel,
		client:       client,
		lease:        clientv3.NewLease(client),
		watchChan:    make(chan *registry.WatchResponse),
		registerApps: make(map[string]struct{}),
	}
}

func (e *Etcd) newAppsMemSnapshot(appName string) (ams *AppsMemSnapshot, err error) {
	apps, err := e.fetchApps(appName)
	if err != nil {
		return
	}

	container := make(map[string]*app.App)
	for _, app := range apps {
		container[app.Address] = app
	}

	ams = &AppsMemSnapshot{
		container: container,
	}

	return
}

func (e *Etcd) Register(app *app.App) (notifyChan chan *registry.NotifyMessage, err error) {
	notifyChan = make(chan *registry.NotifyMessage, 1)

	grantResp, err := e.lease.Grant(e.ctx, 10)
	if err != nil {
		return
	}

	_, err = e.client.Put(e.ctx, e.appRegisterKey(app), app.String(), clientv3.WithLease(grantResp.ID))
	if err != nil {
		return
	}

	kaChan, err := e.lease.KeepAlive(e.ctx, grantResp.ID)
	if err != nil {
		return
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	e.registerApps[app.ID] = struct{}{}

	go func(app string) {
		for {
			select {
			case _, ok := <-kaChan:
				if !ok {
					e.mu.RLock()
					defer e.mu.RUnlock()

					if _, ok := e.registerApps[app]; !ok {
						notifyChan <- &registry.NotifyMessage{
							Stopped:    true,
							StopReason: registry.AppDeregister,
						}
					} else {
						notifyChan <- &registry.NotifyMessage{
							Stopped:    true,
							StopReason: registry.KeepAliveError,
						}
					}

					close(notifyChan)

					return
				}
			}
		}
	}(app.ID)

	return
}

func (e *Etcd) Deregister(app *app.App) (err error) {
	response, err := e.client.Get(e.ctx, e.appRegisterKey(app))
	if err != nil {
		return
	}

	if len(response.Kvs) == 0 {
		return nil
	}

	leaseID := response.Kvs[0].Lease

	_, err = e.client.Delete(e.ctx, e.appRegisterKey(app))
	if err != nil {
		return
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	_, err = e.lease.Revoke(e.ctx, clientv3.LeaseID(leaseID))
	if err != nil {
		return
	}

	delete(e.registerApps, app.ID)

	return nil
}

func (e *Etcd) Discover(appName string) (apps []*app.App, err error) {
	return e.fetchApps(appName)
}

func (e *Etcd) Watch(appName string) (watchChan chan *registry.WatchResponse, err error) {
	go func(appName string) {
		ams, err := e.newAppsMemSnapshot(appName)
		if err != nil {
			watchResponse := &registry.WatchResponse{
				Canceled: true,
				Error:    err,
			}
			e.watchChan <- watchResponse
			return
		}

		watchChan := e.client.Watch(e.ctx, e.appRegisterPrefix(appName), clientv3.WithPrefix(), clientv3.WithPrevKV(), clientv3.WithProgressNotify())

		for {
			select {
			case watchRes := <-watchChan:
				if watchRes.Canceled == true {
					watchResponse := &registry.WatchResponse{
						Canceled: true,
						Error:    watchRes.Err(),
					}
					e.watchChan <- watchResponse
					return
				}

				if watchRes.Err() != nil {
					watchResponse := &registry.WatchResponse{
						Canceled: false,
						Error:    watchRes.Err(),
					}
					e.watchChan <- watchResponse
					continue
				}

				e.handleEvents(watchRes.Events, ams)
				watchResponse := &registry.WatchResponse{
					Apps: ams.AppsList(),
				}

				e.watchChan <- watchResponse
			}
		}
	}(appName)

	return e.watchChan, nil
}

func (e *Etcd) handleEvents(events []*clientv3.Event, ams *AppsMemSnapshot) {
	for _, event := range events {
		var app *app.App

		switch event.Type {
		case mvccpb.PUT:
			err := json.Unmarshal(event.Kv.Value, &app)
			if err != nil {
				continue
			}

			ams.Put(app)

		case mvccpb.DELETE:
			err := json.Unmarshal(event.PrevKv.Value, &app)
			if err != nil {
				continue
			}

			ams.Delete(app)
		}
	}
}

func (e *Etcd) fetchApps(appName string) (apps []*app.App, err error) {
	response, err := e.client.Get(e.ctx, e.appRegisterPrefix(appName), clientv3.WithPrefix())
	if err != nil {
		return
	}

	for _, kv := range response.Kvs {
		var app *app.App

		err := json.Unmarshal(kv.Value, &app)
		if err != nil {
			continue
		}

		apps = append(apps, app)
	}

	return
}

func (e *Etcd) appRegisterKey(app *app.App) string {
	return fmt.Sprintf("%s/%s/%s_%d", prefix, app.Name, app.Host, app.Port)
}

func (e *Etcd) appRegisterPrefix(appName string) string {
	return fmt.Sprintf("%s/%s/", prefix, appName)
}

type AppsMemSnapshot struct {
	container map[string]*app.App
}

func (ams *AppsMemSnapshot) Put(app *app.App) {
	ams.container[app.Address] = app
}

func (ams *AppsMemSnapshot) Delete(app *app.App) {
	delete(ams.container, app.Address)
}

func (ams *AppsMemSnapshot) AppsList() (apps []*app.App) {
	for _, app := range ams.container {
		apps = append(apps, app)
	}

	return
}
