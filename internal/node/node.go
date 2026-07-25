package node

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"

	"consul-journey/internal"
	"consul-journey/internal/utils"

	capi "github.com/hashicorp/consul/api"
	"github.com/hashicorp/go-cleanhttp"
	"go.uber.org/zap"
)

type Node struct {
	logger    *zap.Logger
	cfg       *Config
	client    *capi.Client
	id        string
	httpPort  int
	startedAt time.Time

	leaderLockRelease chan struct{}

	pmu      sync.RWMutex
	peers    []*Peer
	leaderID string

	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

func New(logger *zap.Logger, cfg *Config, httpPort int) (*Node, error) {
	logger = logger.Named("node")
	client, err := capi.NewClient(newCapiConfig(cfg))
	if err != nil {
		return nil, fmt.Errorf("failed to build consul client: %w", err)
	}
	if _, err := client.Agent().Self(); err != nil {
		return nil, fmt.Errorf("consul agent unreachable at %s: %w", cfg.ConsulAddress, err)
	}
	n := &Node{
		logger:    logger,
		cfg:       cfg,
		client:    client,
		id:        internal.Module() + "-" + utils.RandomString(16),
		httpPort:  httpPort,
		startedAt: time.Now().UTC(),

		leaderLockRelease: make(chan struct{}, 1),

		pmu:      sync.RWMutex{},
		peers:    make([]*Peer, 0),
		leaderID: "",

		wg: sync.WaitGroup{},
	}
	n.ctx, n.cancel = context.WithCancel(context.Background())
	return n, nil
}

func (n *Node) Start() (err error) {
	n.logger.Info("starting ...")
	if err := n.register(); err != nil {
		n.logger.Error("failed to register service", zap.Error(err))
		return fmt.Errorf("failed to register service: %w", err)
	}
	defer func() {
		if err != nil {
			_ = n.client.Agent().ServiceDeregister(n.id)
		}
	}()
	n.logger.Info(
		"registered service",
		zap.String("id", n.id),
		zap.String("address", utils.Hostname()),
		zap.String("os", runtime.GOOS),
		zap.String("arch", runtime.GOARCH),
		zap.String("pid", utils.PID()),
		zap.Time("started_at", n.startedAt),
	)

	n.wg.Go(n.runHeartbeat)
	n.wg.Go(n.runDiscovery)
	n.wg.Go(n.runElection)

	n.logger.Info("started")
	return nil
}

func (n *Node) Stop() {
	n.logger.Info("stopping ...")
	n.cancel()
	n.wg.Wait()
	close(n.leaderLockRelease)
	if err := n.client.Agent().ServiceDeregister(n.id); err != nil {
		n.logger.Error("failed to deregister service", zap.Error(err))
		return
	}
	n.logger.Info("stopped")
}

func (n *Node) ID() string {
	return n.id
}

func (n *Node) LeaderElectionEnabled() bool {
	return n.cfg.LeaderElection.Enabled
}

func (n *Node) Self() *Peer {
	return n.GetPeer(n.id)
}

func (n *Node) PeersCount() int {
	n.pmu.RLock()
	defer n.pmu.RUnlock()
	count := 0
	for _, p := range n.peers {
		if p.ID != n.id {
			count++
		}
	}
	return count
}

func (n *Node) register() error {
	// build registration
	reg := &capi.AgentServiceRegistration{
		Kind: capi.ServiceKindTypical,
		ID:   n.id,
		Name: internal.Module(),
		Tags: []string{
			internal.Module(),
			internal.Version(),
			internal.Revision(),
			runtime.GOOS + "/" + runtime.GOARCH,
		},
		Ports: capi.ServicePorts{
			{
				Name:    "http",
				Port:    n.httpPort,
				Default: true,
			},
		},
		Address:           utils.Hostname(),
		EnableTagOverride: false,
		Meta: map[string]string{
			"version":    internal.Version(),
			"revision":   internal.Revision(),
			"os":         runtime.GOOS,
			"arch":       runtime.GOARCH,
			"pid":        utils.PID(),
			"started_at": n.startedAt.Format(time.RFC3339),
		},
		Checks:    n.buildChecks(),
		Namespace: n.cfg.Namespace,
		Partition: n.cfg.Partition,
		Locality:  n.cfg.Locality.toCapiLocality(),
	}
	return n.client.Agent().ServiceRegister(reg)
}

func newCapiConfig(cfg *Config) *capi.Config {
	capiCfg := &capi.Config{
		Address:    cfg.ConsulAddress,
		Scheme:     "http",
		PathPrefix: cfg.PathPrefix,
		Datacenter: cfg.Datacenter,
		Transport:  cleanhttp.DefaultTransport(),
		HttpClient: http.DefaultClient,
		WaitTime:   cfg.WaitTime,
		Token:      cfg.Token,
		TokenFile:  cfg.TokenFile,
		Namespace:  cfg.Namespace,
		Partition:  cfg.Partition,
	}
	if cfg.Username != "" {
		capiCfg.HttpAuth = &capi.HttpBasicAuth{
			Username: cfg.Username,
			Password: cfg.Password,
		}
	}
	if cfg.TLS.Enabled {
		capiCfg.Scheme = "https"
		capiCfg.TLSConfig = capi.TLSConfig{
			Address:            cfg.TLS.ServerName,
			CAFile:             cfg.TLS.CAFile,
			CAPath:             cfg.TLS.CAPath,
			CertFile:           cfg.TLS.CertFile,
			KeyFile:            cfg.TLS.KeyFile,
			InsecureSkipVerify: !cfg.TLS.Verify,
		}
	}
	return capiCfg
}
