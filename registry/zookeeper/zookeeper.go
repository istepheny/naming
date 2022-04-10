package zookeeper

import (
	"git.ucloudadmin.com/monkey/naming/app"
	"git.ucloudadmin.com/monkey/naming/config"
	"git.ucloudadmin.com/monkey/naming/registry"
)

const Driver = "zookeeper"

func init() {
	registry.Set(Driver, NewZookeeper)
}

type Zookeeper struct{}

func NewZookeeper(c config.Config) registry.Registry {
	return &Zookeeper{}
}

func (z *Zookeeper) Register(app *app.App) (err error) {
	//TODO implement me
	panic("implement me")
}

func (z *Zookeeper) Deregister(app *app.App) (err error) {
	//TODO implement me
	panic("implement me")
}

func (z *Zookeeper) Discover(appName string) (apps []*app.App, err error) {
	//TODO implement me
	panic("implement me")
}

func (z *Zookeeper) Watch(appName string) (watchChan chan *registry.WatchResponse, err error) {
	//TODO implement me
	panic("implement me")
}
