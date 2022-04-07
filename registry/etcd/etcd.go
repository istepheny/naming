package etcd

import (
	"context"
	"encoding/json"
	"fmt"
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
	client    *clientv3.Client
	lease     clientv3.Lease
	watchChan chan map[string][]*app.App
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

	return &Etcd{
		client:    client,
		lease:     clientv3.NewLease(client),
		watchChan: make(chan map[string][]*app.App),
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
		appName:   appName,
		container: container,
	}

	return
}

func (e *Etcd) Register(app *app.App) (err error) {
	grantResp, err := e.lease.Grant(context.TODO(), 10)

	_, err = e.client.Put(context.TODO(), e.appRegisterKey(app), app.String(), clientv3.WithLease(grantResp.ID))
	if err != nil {
		return
	}

	kaChan, err := e.lease.KeepAlive(context.TODO(), grantResp.ID)
	if err != nil {
		return
	}

	go func() {
		for {
			select {
			case _, ok := <-kaChan:
				if !ok {
					return
				}
			}
		}
	}()

	return nil
}

func (e *Etcd) Deregister(app *app.App) (err error) {
	response, err := e.client.Get(context.TODO(), e.appRegisterKey(app))
	if err != nil {
		return
	}

	if len(response.Kvs) == 0 {
		err = fmt.Errorf("app not register")
		return
	}

	leaseID := response.Kvs[0].Lease
	_, err = e.lease.Revoke(context.TODO(), clientv3.LeaseID(leaseID))
	if err != nil {
		return
	}

	_, err = e.client.Delete(context.TODO(), e.appRegisterKey(app))
	if err != nil {
		return
	}

	return nil
}

func (e *Etcd) Discover(appName string) (apps []*app.App, err error) {
	return e.fetchApps(appName)
}

func (e *Etcd) Watch(appName string) (watchChan chan map[string][]*app.App, err error) {
	go func(appName string) {
		for i := 0; i < 5; i++ {
			canceled := e.watch(appName)
			if !canceled {
				break
			}
			time.Sleep(30 * time.Second)
		}
	}(appName)

	return e.watchChan, nil
}

func (e *Etcd) watch(appName string) bool {
	ams, err := e.newAppsMemSnapshot(appName)
	if err != nil {
		return true
	}

	ticker := time.NewTicker(time.Duration(5) * time.Second)
	watchChan := e.client.Watch(context.TODO(), e.appRegisterPrefix(appName), clientv3.WithPrefix(), clientv3.WithPrevKV(), clientv3.WithProgressNotify())

	var events []*clientv3.Event

	for {
		select {
		case watchRes := <-watchChan:
			if watchRes.Canceled == true {
				return true
			}

			if watchRes.Err() != nil {
				continue
			}

			events = append(events, watchRes.Events...)

			if len(events) >= 100 {
				ticker.Stop()
				e.handleEvents(events, ams)
				events = make([]*clientv3.Event, 0)
				ticker = time.NewTicker(time.Duration(5) * time.Second)
			}

		case <-ticker.C:
			if len(events) > 0 {
				e.handleEvents(events, ams)
				events = make([]*clientv3.Event, 0)
			}
		}
	}
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

	appsMap := make(map[string][]*app.App)
	appsMap[ams.AppName()] = ams.AppsList()

	e.watchChan <- appsMap
}

func (e *Etcd) fetchApps(appName string) (apps []*app.App, err error) {
	response, err := e.client.Get(context.TODO(), e.appRegisterPrefix(appName), clientv3.WithPrefix())
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
	return fmt.Sprintf("%s/%s", prefix, appName)
}

type AppsMemSnapshot struct {
	appName   string
	container map[string]*app.App
}

func (ams *AppsMemSnapshot) Put(app *app.App) {
	ams.container[app.Address] = app
}

func (ams *AppsMemSnapshot) Delete(app *app.App) {
	delete(ams.container, app.Address)
}

func (ams *AppsMemSnapshot) AppName() string {
	return ams.appName
}

func (ams *AppsMemSnapshot) AppsList() (apps []*app.App) {
	for _, app := range ams.container {
		apps = append(apps, app)
	}

	return
}
