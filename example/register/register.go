package main

import (
	"git.ucloudadmin.com/monkey/naming"
	"git.ucloudadmin.com/monkey/naming/app"
	"git.ucloudadmin.com/monkey/naming/config"
)

func main() {
	c := config.Config{
		Driver:   "consul",
		Servers:  []string{"10.72.137.14:8500"},
		Username: "",
		Password: "my-token",
	}

	n := naming.New(c)

	n.Register(app.New("user", "127.0.0.1", 81))
	n.Register(app.New("user", "127.0.0.1", 82))
	n.Register(app.New("user", "127.0.0.1", 83))

	n.Register(app.New("order", "127.0.0.1", 81))
	n.Register(app.New("order", "127.0.0.1", 82))
	n.Register(app.New("order", "127.0.0.1", 83))
}
