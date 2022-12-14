package main

import (
	"github.com/istepheny/naming"
	"github.com/istepheny/naming/app"
	"github.com/istepheny/naming/config"
)

func main() {
	c := config.Config{
		Driver:   "consul",
		Servers:  []string{"10.72.137.14:8500"},
		Username: "",
		Password: "my-token",
	}

	n := naming.New(c)

	n.Deregister(app.New("user", "127.0.0.1", 81))
	n.Deregister(app.New("user", "127.0.0.1", 82))
	n.Deregister(app.New("user", "127.0.0.1", 83))

	n.Deregister(app.New("order", "127.0.0.1", 81))
	n.Deregister(app.New("order", "127.0.0.1", 82))
	n.Deregister(app.New("order", "127.0.0.1", 83))
}
