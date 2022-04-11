package registry

import "git.ucloudadmin.com/monkey/naming/app"

type WatchResponse struct {
	Apps     []*app.App
	Canceled bool
	Error    error
}

type Registry interface {
	Register(app *app.App) (err error)
	Deregister(app *app.App) (err error)
	Discover(appName string) (apps []*app.App, err error)
	Watch(appName string) (watchChan chan *WatchResponse, err error)
}
