package node

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	capi "github.com/hashicorp/consul/api"
)

type Config struct {
	ConsulAddress  string                `mapstructure:"consul_address"`
	PathPrefix     string                `mapstructure:"path_prefix"`
	Datacenter     string                `mapstructure:"datacenter"`
	Username       string                `mapstructure:"username"`
	Password       string                `mapstructure:"password"`
	WaitTime       time.Duration         `mapstructure:"wait_time"`
	Token          string                `mapstructure:"token"`
	TokenFile      string                `mapstructure:"token_file"`
	Namespace      string                `mapstructure:"namespace"`
	Partition      string                `mapstructure:"partition"`
	Locality       *LocalityConfig       `mapstructure:"locality"`
	TLS            *TLSConfig            `mapstructure:"tls"`
	HTTPCheck      *HTTPCheckConfig      `mapstructure:"http_check"`
	HeartbeatCheck *HeartbeatCheckConfig `mapstructure:"heartbeat_check"`
	Discovery      *DiscoveryConfig      `mapstructure:"discovery"`
	LeaderElection *LeaderElectionConfig `mapstructure:"leader_election"`
}

func (c *Config) Validate() error {
	if c.ConsulAddress == "" {
		return errors.New("consul_address is required")
	} else if _, err := url.Parse("http://" + c.ConsulAddress); err != nil {
		return fmt.Errorf("invalid consul_address: %w", err)
	}
	if c.WaitTime <= 0 {
		return errors.New("wait_time must be greater than 0")
	}
	if err := c.Locality.validate(); err != nil {
		return fmt.Errorf("invalid locality config: %w", err)
	}
	if err := c.TLS.validate(); err != nil {
		return fmt.Errorf("invalid tls config: %w", err)
	}
	if err := c.HTTPCheck.validate(); err != nil {
		return fmt.Errorf("invalid http check config: %w", err)
	}
	if err := c.HeartbeatCheck.validate(); err != nil {
		return fmt.Errorf("invalid heartbeat check config: %w", err)
	}
	if !c.HeartbeatCheck.Enabled && !c.HTTPCheck.Enabled {
		return errors.New("at least one check must be enabled")
	}
	if err := c.Discovery.validate(); err != nil {
		return fmt.Errorf("invalid discovery config: %w", err)
	}
	if err := c.LeaderElection.validate(); err != nil {
		return fmt.Errorf("invalid leader election config: %w", err)
	}
	return nil
}

type TLSConfig struct {
	Enabled    bool   `mapstructure:"enabled"`
	ServerName string `mapstructure:"server_name"`
	CAFile     string `mapstructure:"ca_file"`
	CAPath     string `mapstructure:"ca_path"`
	CertFile   string `mapstructure:"cert_file"`
	KeyFile    string `mapstructure:"key_file"`
	Verify     bool   `mapstructure:"verify"`
}

func (c *TLSConfig) validate() error {
	if !c.Enabled {
		return nil
	}
	if c.ServerName == "" {
		return errors.New("server_name is required")
	}
	if c.CAFile == "" && c.CAPath == "" {
		return errors.New("either ca_file or ca_path is required")
	}
	if c.CertFile == "" {
		return errors.New("cert_file is required")
	}
	if c.KeyFile == "" {
		return errors.New("key_file is required")
	}
	return nil
}

type LocalityConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Region  string `mapstructure:"region"`
	Zone    string `mapstructure:"zone"`
}

func (c *LocalityConfig) validate() error {
	if !c.Enabled {
		return nil
	}
	if c.Region == "" {
		return errors.New("region is required")
	}
	if c.Zone == "" {
		return errors.New("zone is required")
	}
	return nil
}

func (c *LocalityConfig) toCapiLocality() *capi.Locality {
	if !c.Enabled {
		return nil
	}
	return &capi.Locality{
		Region: c.Region,
		Zone:   c.Zone,
	}
}

type HTTPCheckConfig struct {
	Enabled                bool          `mapstructure:"enabled"`
	AddressOverride        string        `mapstructure:"address_override"`
	Interval               time.Duration `mapstructure:"interval"`
	Timeout                time.Duration `mapstructure:"timeout"`
	SuccessBeforePassing   int           `mapstructure:"success_before_passing"`
	FailuresBeforeWarning  int           `mapstructure:"failures_before_warning"`
	FailuresBeforeCritical int           `mapstructure:"failures_before_critical"`
	DeregisterAfter        time.Duration `mapstructure:"deregister_after"`
}

func (c *HTTPCheckConfig) validate() error {
	if !c.Enabled {
		return nil
	}
	if c.AddressOverride != "" {
		if _, err := url.Parse(c.AddressOverride); err != nil {
			return fmt.Errorf("invalid address_override: %w", err)
		}
	}
	if c.Interval <= 0 {
		return errors.New("interval must be greater than 0")
	}
	if c.Timeout <= 0 {
		return errors.New("timeout must be greater than 0")
	}
	if c.SuccessBeforePassing < 0 {
		return errors.New("success_before_passing must be greater than or equal to 0")
	}
	if c.FailuresBeforeWarning < 0 {
		return errors.New("failures_before_warning must be greater than or equal to 0")
	}
	if c.FailuresBeforeCritical < 0 {
		return errors.New("failures_before_critical must be greater than or equal to 0")
	}
	if c.DeregisterAfter <= 0 {
		return errors.New("deregister_after must be greater than 0")
	}
	return nil
}

func (c *HTTPCheckConfig) toCapiAgentServiceCheck(checkID, method, address string) *capi.AgentServiceCheck {
	if !c.Enabled {
		return nil
	}
	return &capi.AgentServiceCheck{
		CheckID:                        checkID,
		Name:                           "HTTP Health",
		Interval:                       c.Interval.String(),
		Timeout:                        c.Timeout.String(),
		HTTP:                           address,
		Method:                         method,
		SuccessBeforePassing:           c.SuccessBeforePassing,
		FailuresBeforeWarning:          c.FailuresBeforeWarning,
		FailuresBeforeCritical:         c.FailuresBeforeCritical,
		DeregisterCriticalServiceAfter: c.DeregisterAfter.String(),
	}
}

type HeartbeatCheckConfig struct {
	Enabled         bool          `mapstructure:"enabled"`
	Interval        time.Duration `mapstructure:"interval"`
	TTL             time.Duration `mapstructure:"ttl"`
	DeregisterAfter time.Duration `mapstructure:"deregister_after"`
}

func (c *HeartbeatCheckConfig) validate() error {
	if !c.Enabled {
		return nil
	}
	if c.Interval <= 0 {
		return errors.New("interval must be greater than 0")
	}
	if c.TTL <= 0 {
		return errors.New("ttl must be greater than 0")
	}
	if c.TTL <= c.Interval {
		return errors.New("ttl must be greater than interval")
	}
	if c.DeregisterAfter <= 0 {
		return errors.New("deregister_after must be greater than 0")
	}
	return nil
}

func (c *HeartbeatCheckConfig) toCapiAgentServiceCheck(checkID string) *capi.AgentServiceCheck {
	if !c.Enabled {
		return nil
	}
	return &capi.AgentServiceCheck{
		CheckID:                        checkID,
		Name:                           "Heartbeat",
		TTL:                            c.TTL.String(),
		Status:                         capi.HealthPassing,
		DeregisterCriticalServiceAfter: c.DeregisterAfter.String(),
	}
}

type DiscoveryConfig struct {
	Backoff     time.Duration `mapstructure:"backoff"`
	Coefficient float64       `mapstructure:"coefficient"`
	Max         time.Duration `mapstructure:"max"`
	PassingOnly bool          `mapstructure:"passing_only"`
}

func (c *DiscoveryConfig) validate() error {
	if c.Backoff <= 0 {
		return errors.New("backoff must be greater than 0")
	}
	if c.Coefficient < 1 {
		return errors.New("coefficient must be greater than or equal to 1")
	}
	if c.Max < c.Backoff {
		return errors.New("max must be greater than or equal to backoff")
	}
	return nil
}

type LeaderElectionConfig struct {
	Enabled     bool          `mapstructure:"enabled"`
	TTL         time.Duration `mapstructure:"ttl"`
	LockDelay   time.Duration `mapstructure:"lock_delay"`
	Backoff     time.Duration `mapstructure:"backoff"`
	Coefficient float64       `mapstructure:"coefficient"`
	Max         time.Duration `mapstructure:"max"`
}

func (c *LeaderElectionConfig) validate() error {
	if !c.Enabled {
		return nil
	}
	if c.TTL < 10*time.Second {
		return errors.New("ttl must be greater than or equal to 10 seconds")
	}
	if c.LockDelay <= 0 {
		return errors.New("lock_delay must be greater than 0")
	}
	if c.LockDelay >= c.TTL {
		return errors.New("lock_delay must be less than ttl")
	}
	if c.Backoff <= 0 {
		return errors.New("backoff must be greater than 0")
	}
	if c.Coefficient < 1 {
		return errors.New("coefficient must be greater than or equal to 1")
	}
	if c.Max < c.Backoff {
		return errors.New("max must be greater than or equal to backoff")
	}
	if c.Max >= c.TTL {
		return errors.New("max must be less than ttl")
	}
	return nil
}
