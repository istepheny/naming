package registry

import "github.com/istepheny/naming/app"

type Reason string

const (
	AppDeregister  Reason = "app deregister"
	KeepAliveError Reason = "keep alive error"
)

type WatchResponse struct {
	Apps     []*app.App
	Canceled bool
	Error    error
}

type NotifyMessage struct {
	Stopped bool

	StopReason Reason
}

type Registry interface {
	Register(app *app.App) (notifyChan chan *NotifyMessage, err error)
	Deregister(app *app.App) (err error)
	Discover(appName string) (apps []*app.App, err error)
	Watch(appName string) (watchChan chan *WatchResponse, err error)
}
