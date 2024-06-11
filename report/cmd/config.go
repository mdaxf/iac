package cmd

import (
	"fmt"
	"time"
)

type Config struct {
	TemplatesPath string
	RenderTimeout time.Duration
	ServerHost    string
	ServerPort    int

	ViewportHeight int
	ViewportWidth  int
}

func (config Config) BaseUrl() string {
	return fmt.Sprintf("%s:%d", config.ServerHost, config.ServerPort)
}

func ParseConfig() (Config, error) {
	var out Config
	/*	err := env.Parse(&out)

		if out.RenderTimeout == 0 {
			out.RenderTimeout = 3 * time.Second
		}
		return out, err  */
	out = Config{
		TemplatesPath:  "cmd/testdata/",
		RenderTimeout:  time.Duration(10) * time.Second,
		ServerHost:     "localhost",
		ServerPort:     8080,
		ViewportHeight: 2048,
		ViewportWidth:  1920,
	}
	return out, nil
}
