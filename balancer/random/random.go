package random

import (
	"errors"
	"math/rand"
	"sync"
	"time"

	"git.ucloudadmin.com/monkey/naming/app"
	"git.ucloudadmin.com/monkey/naming/balancer"
)

var (
	once   sync.Once
	random *Random
)

const Driver = "random"

func init() {
	balancer.Set(Driver, NewRandom)
}

type Random struct {
	rd *rand.Rand
}

func NewRandom() balancer.Balancer {
	once.Do(func() {
		random = &Random{
			rd: rand.New(rand.NewSource(time.Now().UnixNano())),
		}
	})

	return random
}

func (r *Random) Pick(apps []*app.App) (app *app.App, err error) {
	if len(apps) == 0 {
		return nil, errors.New("no endpoints available")
	}

	return apps[r.rd.Intn(len(apps))], nil
}
