package main

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"sync"
	"time"

	"github.com/hashicorp/consul/api"
)

// checkIDs derives the deterministic check IDs for this node's service.
func (c Config) httpCheckID() string { return c.NodeID + ":http" }
func (c Config) ttlCheckID() string  { return c.NodeID + ":ttl" }

// Peer is a snapshot of one instance as seen through Consul's health API.
type Peer struct {
	ID      string            `json:"id"`
	Node    string            `json:"node"`
	Address string            `json:"address"`
	Port    int               `json:"port"`
	Status  string            `json:"status"` // passing | warning | critical | maintenance
	Tags    []string          `json:"tags"`
	Meta    map[string]string `json:"meta"`
}

// LeaderInfo is the payload written to the leader KV key by whoever holds it.
type LeaderInfo struct {
	NodeID     string    `json:"node_id"`
	Address    string    `json:"address"`
	HTTPAddr   string    `json:"http_addr"`
	SessionID  string    `json:"session_id"`
	AcquiredAt time.Time `json:"acquired_at"`
}

// Node is a single running instance. It owns the Consul client and the
// mutable state shared across the registration, health, discovery and
// election goroutines.
type Node struct {
	cfg    Config
	client *api.Client
	log    *slog.Logger
	start  time.Time

	// discovery state
	peersMu sync.RWMutex
	peers   []Peer

	// election state
	mu        sync.RWMutex
	isLeader  bool
	leader    *LeaderInfo // last observed leader (may be self, another node, or nil)
	sessionID string
}

// NewNode builds a Consul client from the config and returns a ready Node.
func NewNode(cfg Config, log *slog.Logger) (*Node, error) {
	apiCfg := api.DefaultConfig()
	apiCfg.Address = cfg.ConsulAddr
	if cfg.Datacenter != "" {
		apiCfg.Datacenter = cfg.Datacenter
	}

	client, err := api.NewClient(apiCfg)
	if err != nil {
		return nil, fmt.Errorf("build consul client: %w", err)
	}

	// Fail fast if the agent is unreachable rather than limping along.
	if _, err := client.Agent().Self(); err != nil {
		return nil, fmt.Errorf("consul agent unreachable at %s: %w", cfg.ConsulAddr, err)
	}

	return &Node{
		cfg:    cfg,
		client: client,
		log:    log,
		start:  time.Now(),
	}, nil
}

// setPeers atomically replaces the discovered peer set (sorted for stable output).
func (n *Node) setPeers(peers []Peer) {
	sort.Slice(peers, func(i, j int) bool { return peers[i].ID < peers[j].ID })
	n.peersMu.Lock()
	n.peers = peers
	n.peersMu.Unlock()
}

// Peers returns a copy of the current peer view.
func (n *Node) Peers() []Peer {
	n.peersMu.RLock()
	defer n.peersMu.RUnlock()
	out := make([]Peer, len(n.peers))
	copy(out, n.peers)
	return out
}

func (n *Node) setSession(id string) {
	n.mu.Lock()
	n.sessionID = id
	n.mu.Unlock()
}

func (n *Node) session() string {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.sessionID
}

// setLeadership records whether this node currently holds the lock and who the
// observed leader is. Transitions are logged so leadership changes are visible.
func (n *Node) setLeadership(isLeader bool, leader *LeaderInfo) {
	n.mu.Lock()
	was := n.isLeader
	n.isLeader = isLeader
	n.leader = leader
	n.mu.Unlock()

	if was != isLeader {
		if isLeader {
			n.log.Info("leadership acquired — this node is now the leader")
		} else {
			n.log.Info("leadership lost — this node is now a follower")
		}
	}
}

// LeadershipStatus returns whether this node leads and the observed leader info.
func (n *Node) LeadershipStatus() (bool, *LeaderInfo) {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.isLeader, n.leader
}

// IsLeader reports whether this node currently holds the lock.
func (n *Node) IsLeader() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.isLeader
}

// Uptime since the node started.
func (n *Node) Uptime() time.Duration { return time.Since(n.start) }

// Run wires the HTTP server, registration, health, discovery and election
// together and blocks until ctx is cancelled, then shuts down cleanly.
func (n *Node) Run(ctx context.Context) error {
	// The HTTP server must be up before registration so Consul's first
	// health scrape succeeds immediately.
	srv := n.newHTTPServer()
	srvErr := make(chan error, 1)
	go func() { srvErr <- n.serveHTTP(srv) }()

	if err := n.register(); err != nil {
		return fmt.Errorf("service registration: %w", err)
	}
	n.log.Info("registered with consul",
		"service", n.cfg.ServiceName,
		"id", n.cfg.NodeID,
		"addr", n.cfg.HTTPListenAddr(),
		"leader_key", n.cfg.LeaderKey,
	)

	var wg sync.WaitGroup
	wg.Add(3)
	go func() { defer wg.Done(); n.runHeartbeat(ctx) }()
	go func() { defer wg.Done(); n.runDiscovery(ctx) }()
	go func() { defer wg.Done(); n.runElection(ctx) }()

	// Block until shutdown is requested or the HTTP server dies.
	select {
	case <-ctx.Done():
		n.log.Info("shutdown requested")
	case err := <-srvErr:
		n.log.Error("http server stopped", "err", err)
	}

	// Give background goroutines a moment to observe cancellation.
	wg.Wait()

	n.shutdownHTTP(srv)
	n.deregister()
	return nil
}
