package node

import (
	"fmt"
	"runtime"
	"time"

	"consul-journey/internal/utils"

	"github.com/bytedance/sonic"
	capi "github.com/hashicorp/consul/api"
	"go.uber.org/zap"
)

func (n *Node) pushHeartbeat() {
	snap := n.assessHealth()
	output, err := sonic.MarshalIndent(snap, "", "  ")
	if err != nil {
		n.logger.Warn("failed to marshal health snapshot", zap.Error(err))
		return
	}
	err = n.client.
		Agent().
		UpdateTTLOpts(
			n.heartbeatCheckID(),
			string(output),
			snap.Status,
			(&capi.QueryOptions{
				Namespace:  n.cfg.Namespace,
				Partition:  n.cfg.Partition,
				Datacenter: n.cfg.Datacenter,
			}).WithContext(n.ctx),
		)
	if err != nil {
		n.logger.Warn("failed to update heartbeat check", zap.Error(err))
		return
	}
	n.logger.Debug("updated heartbeat check", zap.String("status", snap.Status))
}

func (n *Node) runHeartbeat() {
	if !n.cfg.HeartbeatCheck.Enabled {
		return
	}
	n.pushHeartbeat() // fresh heartbeat
	ticker := time.NewTicker(n.cfg.HeartbeatCheck.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-n.ctx.Done():
			return
		case <-ticker.C:
			n.pushHeartbeat()
		}
	}
}

func (n *Node) buildChecks() capi.AgentServiceChecks {
	checks := make(capi.AgentServiceChecks, 0)
	httpAddress := utils.Hostname()
	if n.cfg.HTTPCheck.AddressOverride != "" {
		httpAddress = n.cfg.HTTPCheck.AddressOverride
	}
	httpCheck := n.cfg.HTTPCheck.toCapiAgentServiceCheck(
		n.httpCheckID(),
		"GET",
		fmt.Sprintf("http://%s:%d/healthz", httpAddress, n.httpPort),
	)
	if httpCheck != nil {
		checks = append(checks, httpCheck)
	}
	heartbeatCheck := n.cfg.HeartbeatCheck.toCapiAgentServiceCheck(n.heartbeatCheckID())
	if heartbeatCheck != nil {
		checks = append(checks, heartbeatCheck)
	}
	return checks
}

type healthSnapshot struct {
	Status      string    `json:"status"`
	Uptime      string    `json:"uptime"`
	Timestamp   time.Time `json:"timestamp"`
	Goroutines  int       `json:"goroutines"`
	HeapAllocMB float64   `json:"heap_alloc_mb"`
	NumGC       uint32    `json:"num_gc"`
	PeersCount  int       `json:"peers_count"`
	// TODO: add isLeader status
}

func (n *Node) assessHealth() healthSnapshot {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	snap := healthSnapshot{
		Uptime:      time.Since(n.startedAt).Round(time.Second).String(),
		Timestamp:   time.Now().UTC(),
		Goroutines:  runtime.NumGoroutine(),
		HeapAllocMB: float64(m.HeapAlloc) / (1024 * 1024),
		NumGC:       m.NumGC,
		PeersCount:  n.PeersCount(),
	}
	// TODO: ask underlying services for health or calculate from metrics
	snap.Status = capi.HealthPassing
	return snap
}

func (n *Node) httpCheckID() string {
	return n.id + ":http"
}

func (n *Node) heartbeatCheckID() string {
	return n.id + ":heartbeat"
}
