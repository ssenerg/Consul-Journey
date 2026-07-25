package grpc

import (
	"errors"
	"net"
	"strconv"

	"consul-journey/internal/utils"
)

type Config struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

func (c *Config) Validate() error {
	if c.Host == "" {
		return errors.New("host is required")
	}
	if err := utils.ValidatePort(c.Port); err != nil {
		return err
	}
	return nil
}

func (c *Config) srvAddr() string {
	return net.JoinHostPort(c.Host, strconv.Itoa(c.Port))
}
