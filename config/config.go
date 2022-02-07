package config

import "encoding/json"

type Config struct {
	Driver   string   `json:"driver"`
	Servers  []string `json:"servers"`
	Username string   `json:"username"`
	Password string   `json:"password"`
}

func (c *Config) String() string {
	b, _ := json.Marshal(c)

	return string(b)
}
