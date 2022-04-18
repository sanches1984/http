package server

import "time"

type Specs struct {
	Host            string        `envconfig:"SERVER_HTTP_HOST" default:"localhost" required:"true"`
	Port            int           `envconfig:"SERVER_HTTP_PORT" default:"7784" required:"true"`
	GracefulDelay   time.Duration `envconfig:"SERVER_HTTP_GRACEFUL_DELAY" default:"3s"`
	GracefulTimeout time.Duration `envconfig:"SERVER_HTTP_GRACEFUL_TIMEOUT" default:"5s"`
}

var specs Specs
