package etcd

import (
	"git.ucloudadmin.com/monkey/naming/app"
	"git.ucloudadmin.com/monkey/naming/config"
	"git.ucloudadmin.com/monkey/naming/registry"
)

const Driver = "etcd"

func init() {
	registry.Set(Driver, NewEtcd)
}

type Etcd struct{}

func NewEtcd(c config.Config) registry.Registry {
	return &Etcd{}
}

func (e *Etcd) Register(app *app.App) (err error) {
	//TODO implement me
	panic("implement me")
}

func (e *Etcd) Deregister(app *app.App) (err error) {
	//TODO implement me
	panic("implement me")
}

func (e *Etcd) Discover(appName string) (apps []*app.App, err error) {
	//TODO implement me
	panic("implement me")
}

func (e *Etcd) Watch(appName string) (watchChan chan map[string][]*app.App, err error) {
	//TODO implement me
	panic("implement me")
}
