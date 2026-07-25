package main

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"strings"

	"consul-journey/internal/node"
	"consul-journey/internal/server"
	"consul-journey/internal/utils/logging"

	"github.com/spf13/viper"
)

//go:embed default.yaml
var defaultConfig []byte

type Config struct {
	Logging *logging.Config `mapstructure:"logging"`
	Node    *node.Config    `mapstructure:"node"`
	Server  *server.Config  `mapstructure:"server"`
}

func (c *Config) Validate() error {
	if err := c.Logging.Validate(); err != nil {
		return fmt.Errorf("logging: %w", err)
	}
	if err := c.Node.Validate(); err != nil {
		return fmt.Errorf("node: %w", err)
	}
	if err := c.Server.Validate(); err != nil {
		return fmt.Errorf("server: %w", err)
	}
	return nil
}

func LoadConfig(envPrefix, path, name string) (cfg *Config, err error) {
	v := viper.New()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AddConfigPath(path)
	v.SetConfigName(name)
	v.SetConfigType("yaml")
	v.SetEnvPrefix(envPrefix)
	v.AutomaticEnv()
	if err := v.ReadConfig(bytes.NewReader(defaultConfig)); err != nil {
		return nil, err
	}
	if err := v.MergeInConfig(); err != nil {
		_, okViper := errors.AsType[viper.ConfigFileNotFoundError](err)
		_, okFs := errors.AsType[*fs.PathError](err)
		if !okViper && !okFs {
			slog.Error("failed to read config file", "error", err)
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		slog.Warn("config file not found, falling back to defaults and using environment variables")
	} else {
		slog.Info("config file found", "file", v.ConfigFileUsed())
	}
	cfg = &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		slog.Error("failed to unmarshal config", "error", err)
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		slog.Error("config is invalid", "error", err)
		return nil, fmt.Errorf("config: %w", err)
	}
	slog.Info("config loaded")
	return cfg, nil
}
