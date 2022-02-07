package balancer

import (
	"git.ucloudadmin.com/monkey/naming/app"
)

type Balancer interface {
	Pick(apps []*app.App) (*app.App, error)
}
