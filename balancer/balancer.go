package balancer

import (
	"github.com/istepheny/naming/app"
)

type Balancer interface {
	Pick(apps []*app.App) (*app.App, error)
}
