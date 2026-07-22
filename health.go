package main

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"time"

	"github.com/hashicorp/consul/api"
)

// register publishes this instance to the local Consul agent with two
// complementary health checks:
//
//   - An HTTP check that Consul actively scrapes (pull model). It proves the
//     process is reachable and its HTTP surface is serving.
//   - A TTL check the app pushes into (push model). It lets the app report
//     rich, self-assessed health it alone can compute (queue depth, GC
//     pressure, dependency status, …) and fails fast if the app hangs.
//
// Registering both gives defence in depth: a wedged process that still accepts
// TCP connections is caught by the TTL check; a crashed process is caught by
// the HTTP check. Either going critical for DeregisterAfter reaps the service.
func (n *Node) register() error {
	reg := &api.AgentServiceRegistration{
		ID:      n.cfg.NodeID,
		Name:    n.cfg.ServiceName,
		Address: n.cfg.AdvertiseAddr,
		Port:    n.cfg.HTTPPort,
		Tags:    []string{"consul-journey", "v1", runtime.GOOS + "/" + runtime.GOARCH},
		Meta: map[string]string{
			"version":    "1.0.0",
			"started_at": n.start.UTC().Format(time.RFC3339),
			"pid":        fmt.Sprintf("%d", pid()),
		},
		Checks: api.AgentServiceChecks{
			{
				CheckID:                        n.cfg.httpCheckID(),
				Name:                           "HTTP /health",
				HTTP:                           n.cfg.HealthCheckURL(),
				Method:                         "GET",
				Interval:                       n.cfg.HTTPCheckInterval.String(),
				Timeout:                        n.cfg.HTTPCheckTimeout.String(),
				DeregisterCriticalServiceAfter: n.cfg.DeregisterAfter.String(),
			},
			{
				CheckID: n.cfg.ttlCheckID(),
				Name:    "app heartbeat (TTL)",
				TTL:     n.cfg.TTLCheckTTL.String(),
				// Start healthy so the election session (which is tied to this
				// check) can be created immediately at boot.
				Status:                         api.HealthPassing,
				DeregisterCriticalServiceAfter: n.cfg.DeregisterAfter.String(),
			},
		},
	}

	// EnableTagOverride left false: tags are authoritative from the service.
	if err := n.client.Agent().ServiceRegister(reg); err != nil {
		return err
	}
	// Push an initial heartbeat so the TTL check is immediately passing with
	// real data rather than the default registration status.
	return n.pushHeartbeat()
}

// deregister removes the service from the agent. Best-effort: on shutdown we
// want to try even if the context is already cancelled.
func (n *Node) deregister() {
	if err := n.client.Agent().ServiceDeregister(n.cfg.NodeID); err != nil {
		n.log.Warn("deregister failed", "err", err)
		return
	}
	n.log.Info("deregistered from consul", "id", n.cfg.NodeID)
}

// runHeartbeat periodically pushes self-assessed health into the TTL check.
// This is the "send healthcheck data to Consul" path: the output field carries
// a JSON snapshot of live runtime metrics that shows up in `consul monitor`
// and the UI, and the status downgrades to warning/critical under pressure.
func (n *Node) runHeartbeat(ctx context.Context) {
	ticker := time.NewTicker(n.cfg.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := n.pushHeartbeat(); err != nil {
				// Transient agent errors are expected during restarts; keep
				// trying. If they persist, the TTL will lapse to critical.
				n.log.Warn("heartbeat push failed", "err", err)
			}
		}
	}
}

// healthSnapshot is the self-assessed health report embedded in the TTL check.
type healthSnapshot struct {
	Status      string    `json:"status"`
	NodeID      string    `json:"node_id"`
	Uptime      string    `json:"uptime"`
	Goroutines  int       `json:"goroutines"`
	HeapAllocMB float64   `json:"heap_alloc_mb"`
	NumGC       uint32    `json:"num_gc"`
	IsLeader    bool      `json:"is_leader"`
	Peers       int       `json:"peers_known"`
	Timestamp   time.Time `json:"timestamp"`
}

// assessHealth computes this node's current health. Real services would fold
// in dependency probes (db, cache, downstream APIs); here we derive a status
// from runtime pressure to demonstrate warning/critical reporting.
func (n *Node) assessHealth() healthSnapshot {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	snap := healthSnapshot{
		NodeID:      n.cfg.NodeID,
		Uptime:      n.Uptime().Round(time.Second).String(),
		Goroutines:  runtime.NumGoroutine(),
		HeapAllocMB: float64(m.HeapAlloc) / (1024 * 1024),
		NumGC:       m.NumGC,
		IsLeader:    n.IsLeader(),
		Peers:       len(n.Peers()),
		Timestamp:   time.Now().UTC(),
	}

	// Example thresholds: goroutine leaks are a classic silent failure mode.
	switch {
	case snap.Goroutines > 10000:
		snap.Status = "critical"
	case snap.Goroutines > 2000:
		snap.Status = "warning"
	default:
		snap.Status = "passing"
	}
	return snap
}

// pushHeartbeat writes the current snapshot into the TTL check.
func (n *Node) pushHeartbeat() error {
	snap := n.assessHealth()

	var status string
	switch snap.Status {
	case "critical":
		status = api.HealthCritical
	case "warning":
		status = api.HealthWarning
	default:
		status = api.HealthPassing
	}

	output, _ := json.MarshalIndent(snap, "", "  ")
	// UpdateTTLOpts is the non-deprecated form; the check id, human-readable
	// output and status land in Consul for operators and the UI to see.
	return n.client.Agent().UpdateTTLOpts(n.cfg.ttlCheckID(), string(output), status, nil)
}
