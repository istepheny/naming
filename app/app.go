package app

import (
	"encoding/json"
	"fmt"
)

type App struct {
	ID       string            `json:"id,omitempty"`
	Name     string            `json:"name,omitempty"`
	Host     string            `json:"host,omitempty"`
	Port     int               `json:"port,omitempty"`
	Address  string            `json:"address,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

func New(name string, host string, port int, metadata ...map[string]string) *App {
	md := make(map[string]string)
	if len(metadata) > 0 {
		md = metadata[0]
	}

	return &App{
		ID:       fmt.Sprintf("%s:%s:%d", name, host, port),
		Name:     name,
		Host:     host,
		Port:     port,
		Address:  fmt.Sprintf("%s:%d", host, port),
		Metadata: md,
	}
}

func (a *App) String() string {
	b, _ := json.Marshal(a)

	return string(b)
}
