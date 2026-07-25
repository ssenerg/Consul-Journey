package http

import (
	"errors"
	"net"
	"strconv"
	"time"

	"consul-journey/internal/utils"
)

type Config struct {
	Host              string        `mapstructure:"host"`
	Port              int           `mapstructure:"port"`
	BodyLimit         int           `mapstructure:"body_limit"`
	Concurrency       int           `mapstructure:"concurrency"`
	ReadTimeout       time.Duration `mapstructure:"read_timeout"`
	WriteTimeout      time.Duration `mapstructure:"write_timeout"`
	IdleTimeout       time.Duration `mapstructure:"idle_timeout"`
	ReadBufferSize    int           `mapstructure:"read_buffer_size"`
	WriteBufferSize   int           `mapstructure:"write_buffer_size"`
	ProxyHeader       string        `mapstructure:"proxy_header"`
	Keepalive         bool          `mapstructure:"keepalive"`
	StreamRequestBody bool          `mapstructure:"stream_request_body"`
	ShutdownTimeout   time.Duration `mapstructure:"shutdown_timeout"`
	PrintRoutes       bool          `mapstructure:"print_routes"`
}

func (c *Config) Validate() error {
	if c.Host == "" {
		return errors.New("host is required")
	}
	if err := utils.ValidatePort(c.Port); err != nil {
		return err
	}
	if c.BodyLimit <= 0 {
		return errors.New("body_limit must be greater than 0")
	}
	if c.Concurrency <= 0 {
		return errors.New("concurrency must be greater than 0")
	}
	if c.ReadTimeout <= 0 {
		return errors.New("read_timeout must be greater than 0")
	}
	if c.WriteTimeout <= 0 {
		return errors.New("write_timeout must be greater than 0")
	}
	if c.IdleTimeout <= 0 {
		return errors.New("idle_timeout must be greater than 0")
	}
	if c.ReadBufferSize <= 0 {
		return errors.New("read_buffer_size must be greater than 0")
	}
	if c.WriteBufferSize <= 0 {
		return errors.New("write_buffer_size must be greater than 0")
	}
	if c.ShutdownTimeout <= 0 {
		return errors.New("shutdown_timeout must be greater than 0")
	}
	if c.ShutdownTimeout <= c.ReadTimeout {
		return errors.New("shutdown_timeout must be greater than read_timeout")
	}
	return nil
}

func (c *Config) srvAddr() string {
	return net.JoinHostPort(c.Host, strconv.Itoa(c.Port))
}
