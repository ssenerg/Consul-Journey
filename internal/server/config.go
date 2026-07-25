package server

import (
	"errors"
	"fmt"

	"consul-journey/internal/server/grpc"
	"consul-journey/internal/server/http"
)

type Config struct {
	HTTP *http.Config `mapstructure:"http"`
	GRPC *grpc.Config `mapstructure:"grpc"`
}

func (c *Config) Validate() error {
	if err := c.HTTP.Validate(); err != nil {
		return fmt.Errorf("http: %w", err)
	}
	if err := c.GRPC.Validate(); err != nil {
		return fmt.Errorf("grpc: %w", err)
	}
	if c.HTTP.Port == c.GRPC.Port {
		return errors.New("http and grpc ports must be different")
	}
	return nil
}
