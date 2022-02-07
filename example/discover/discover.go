package main

import (
	"fmt"
	"log"
	"net/http"

	"git.ucloudadmin.com/monkey/naming"
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

	app, err := n.Discover("user")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(app.String())

	http.HandleFunc("/discover", func(writer http.ResponseWriter, request *http.Request) {
		app, err := n.Discover("user")
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println(app.Address)
		writer.Write([]byte(app.Address))
	})

	err = http.ListenAndServe(":80", nil)
	if err != nil {
		log.Println(err)
		return
	}
}
