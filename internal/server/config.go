package server

import (
	"fmt"

	"consul-journey/internal/server/http"
)

type Config struct {
	HTTP *http.Config `mapstructure:"http"`
}

func (c *Config) Validate() error {
	if err := c.HTTP.Validate(); err != nil {
		return fmt.Errorf("http: %w", err)
	}
	// TODO: if other servers added check ports duplicates
	return nil
}
