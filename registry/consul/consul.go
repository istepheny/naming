package git

import (
	"git.ucloudadmin.com/monkey/naming/app"
	"git.ucloudadmin.com/monkey/naming/config"
	"git.ucloudadmin.com/monkey/naming/registry"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
)

const Driver = "consul"

func init() {
	registry.Set(Driver, NewConsul)
}

type Consul struct {
	client    *api.Client
	watchChan chan *registry.WatchResponse
}

func NewConsul(c config.Config) registry.Registry {
	consulConfig := api.DefaultConfig()
	consulConfig.Address = c.Servers[0]

	if c.Password != "" {
		consulConfig.Token = c.Password
	}

	client, err := api.NewClient(consulConfig)
	if err != nil {
		panic(err)
	}

	return &Consul{
		client:    client,
		watchChan: make(chan *registry.WatchResponse),
	}
}

func (c *Consul) Register(app *app.App) (err error) {
	r := &api.AgentServiceRegistration{
		ID:      app.ID,
		Name:    app.Name,
		Port:    app.Port,
		Address: app.Host,
		Meta:    app.Metadata,
		Check: &api.AgentServiceCheck{
			TCP:                            app.Address,
			Timeout:                        "5s",
			Interval:                       "10s",
			DeregisterCriticalServiceAfter: "30s",
		},
	}

	return c.client.Agent().ServiceRegister(r)
}

func (c *Consul) Deregister(app *app.App) (err error) {
	err = c.client.Agent().ServiceDeregister(app.ID)

	e := err.(api.StatusError)
	if e.Code == 404 {
		return nil
	}

	return e
}

func (c *Consul) Discover(appName string) (apps []*app.App, err error) {
	serviceEntries, _, err := c.client.Health().Service(appName, "", false, nil)
	if err != nil {
		return nil, err
	}

	for _, serviceEntry := range serviceEntries {
		a := app.New(
			serviceEntry.Service.Service,
			serviceEntry.Service.Address,
			serviceEntry.Service.Port,
		)

		apps = append(apps, a)
	}

	return apps, nil
}

func (c *Consul) Watch(appName string) (watchChan chan *registry.WatchResponse, err error) {
	params := map[string]interface{}{"type": "service", "service": appName}
	plan, err := watch.Parse(params)
	if err != nil {
		return nil, err
	}

	plan.Handler = c.serviceHandler
	go plan.RunWithClientAndHclog(c.client, nil)
	return c.watchChan, nil
}

func (c *Consul) serviceHandler(_ uint64, result interface{}) {
	serviceEntries, ok := result.([]*api.ServiceEntry)
	if !ok {
		return
	}

	apps := []*app.App{}
	for _, s := range serviceEntries {
		if s.Checks.AggregatedStatus() != api.HealthPassing {
			continue
		}

		a := app.New(
			s.Service.Service,
			s.Service.Address,
			s.Service.Port,
			s.Service.Meta,
		)

		apps = append(apps, a)
	}

	c.watchChan <- &registry.WatchResponse{
		Apps:     apps,
		Canceled: false,
		Error:    nil,
	}
}
