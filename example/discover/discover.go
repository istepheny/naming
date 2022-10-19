package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/istepheny/naming"
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

	http.HandleFunc("/discover", func(writer http.ResponseWriter, request *http.Request) {
		app, err := n.Discover("user")
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println(app.String())
		writer.Write([]byte(app.String()))
	})

	err := http.ListenAndServe(":80", nil)
	if err != nil {
		log.Println(err)
		return
	}
}
