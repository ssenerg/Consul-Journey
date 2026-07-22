package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

// Config holds all tunables for a single node. Every field is derived from
// flags with environment-variable fallbacks so the same binary can be run
// identically from a shell, a process manager, or a container.
type Config struct {
	// Identity
	ServiceName string // logical service all instances share (used for discovery)
	NodeID      string // unique per-instance service ID registered in Consul
	Datacenter  string // optional; empty means the agent's default DC

	// Networking
	ConsulAddr    string // address of the local Consul agent (host:port)
	BindAddr      string // address the HTTP server binds/listens on
	AdvertiseAddr string // address registered with Consul + used in the health-check URL (reachable from the Consul agent)
	HTTPPort      int    // port for this instance's own HTTP API

	// Health
	HTTPCheckInterval time.Duration // how often Consul scrapes /health
	HTTPCheckTimeout  time.Duration // per-scrape timeout
	TTLCheckTTL       time.Duration // window before the TTL check goes critical
	HeartbeatInterval time.Duration // how often the app pushes TTL heartbeat data
	DeregisterAfter   time.Duration // auto-deregister once critical this long

	// Discovery
	WaitTime time.Duration // blocking-query max wait for watches

	// Election
	LeaderKey        string        // KV key contended for leadership
	SessionTTL       time.Duration // session TTL (auto-renewed while healthy)
	SessionLockDelay time.Duration // grace period before a freed lock is re-acquirable
	CampaignBackoff  time.Duration // wait before rebuilding a lost session
}

// envOr returns the environment variable if set, otherwise the fallback.
func envOr(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

// LoadConfig parses flags (with CJ_* env fallbacks) and validates the result.
func LoadConfig(args []string) (Config, error) {
	fs := flag.NewFlagSet("consul-journey", flag.ContinueOnError)

	host, _ := os.Hostname()
	if host == "" {
		host = "node"
	}

	var (
		serviceName = fs.String("service", envOr("CJ_SERVICE", "consul-journey"), "logical service name shared by all instances")
		nodeID      = fs.String("id", envOr("CJ_NODE_ID", ""), "unique node/service id (default: <host>-<port>)")
		datacenter  = fs.String("dc", envOr("CJ_DATACENTER", ""), "consul datacenter (default: agent default)")
		consulAddr  = fs.String("consul", envOr("CJ_CONSUL_ADDR", "127.0.0.1:8500"), "consul agent address host:port")
		bindAddr    = fs.String("bind", envOr("CJ_BIND_ADDR", "127.0.0.1"), "address the http server binds/listens on")
		advertise   = fs.String("advertise", envOr("CJ_ADVERTISE_ADDR", ""), "address registered with consul + scraped for health (default: bind address)")
		httpPort    = fs.Int("port", envInt("CJ_HTTP_PORT", 8080), "http api port for this instance")
		leaderKey   = fs.String("leader-key", envOr("CJ_LEADER_KEY", ""), "kv key used for leader election (default: service/<name>/leader)")
	)

	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}

	cfg := Config{
		ServiceName:       *serviceName,
		NodeID:            *nodeID,
		Datacenter:        *datacenter,
		ConsulAddr:        *consulAddr,
		BindAddr:          *bindAddr,
		AdvertiseAddr:     *advertise,
		HTTPPort:          *httpPort,
		HTTPCheckInterval: 10 * time.Second,
		HTTPCheckTimeout:  3 * time.Second,
		TTLCheckTTL:       30 * time.Second,
		HeartbeatInterval: 8 * time.Second,
		DeregisterAfter:   60 * time.Second,
		WaitTime:          10 * time.Second,
		LeaderKey:         *leaderKey,
		SessionTTL:        15 * time.Second,
		SessionLockDelay:  5 * time.Second,
		CampaignBackoff:   2 * time.Second,
	}

	if cfg.AdvertiseAddr == "" {
		cfg.AdvertiseAddr = cfg.BindAddr
	}
	if cfg.NodeID == "" {
		cfg.NodeID = fmt.Sprintf("%s-%d", host, cfg.HTTPPort)
	}
	if cfg.LeaderKey == "" {
		cfg.LeaderKey = fmt.Sprintf("service/%s/leader", cfg.ServiceName)
	}

	return cfg, cfg.validate()
}

func (c Config) validate() error {
	if c.ServiceName == "" {
		return fmt.Errorf("service name must not be empty")
	}
	if c.HTTPPort <= 0 || c.HTTPPort > 65535 {
		return fmt.Errorf("http port %d out of range", c.HTTPPort)
	}
	if c.BindAddr == "" {
		return fmt.Errorf("bind address must not be empty")
	}
	// The advertise address is what Consul registers and scrapes for the HTTP
	// check, so it must be a concrete, reachable address. An unspecified
	// address (0.0.0.0 / ::) is fine to *bind* but Consul rejects it as a
	// service address (400 Invalid service address) — catch it early with
	// guidance instead of a cryptic registration failure.
	if ip := net.ParseIP(c.AdvertiseAddr); ip != nil && ip.IsUnspecified() {
		return fmt.Errorf("advertise address %q is unspecified; set -advertise (or CJ_ADVERTISE_ADDR) to a concrete address reachable from the Consul agent (e.g. host.docker.internal when Consul runs in Docker)", c.AdvertiseAddr)
	}
	// TTL must comfortably exceed the heartbeat so a single missed beat
	// does not immediately flap the check to critical.
	if c.TTLCheckTTL <= c.HeartbeatInterval {
		return fmt.Errorf("ttl (%s) must be greater than heartbeat interval (%s)", c.TTLCheckTTL, c.HeartbeatInterval)
	}
	return nil
}

// HTTPListenAddr is the address the local HTTP server binds to.
func (c Config) HTTPListenAddr() string {
	return net.JoinHostPort(c.BindAddr, strconv.Itoa(c.HTTPPort))
}

// AdvertiseHTTPAddr is the address peers and the Consul agent use to reach
// this instance (may differ from the bind address, e.g. behind Docker).
func (c Config) AdvertiseHTTPAddr() string {
	return net.JoinHostPort(c.AdvertiseAddr, strconv.Itoa(c.HTTPPort))
}

// HealthCheckURL is the URL Consul scrapes for the HTTP health check.
func (c Config) HealthCheckURL() string {
	return fmt.Sprintf("http://%s/health", c.AdvertiseHTTPAddr())
}

func envInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
