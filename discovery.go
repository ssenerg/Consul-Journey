package main

import (
	"context"
	"time"

	"github.com/hashicorp/consul/api"
)

// runDiscovery keeps an up-to-date view of every instance of the service by
// watching Consul's health endpoint with blocking queries. Rather than polling
// on a timer, a blocking query parks on the server until the result set's
// index advances (a peer joins, leaves, or changes health) or WaitTime
// elapses. This gives near-real-time membership updates with almost no load.
func (n *Node) runDiscovery(ctx context.Context) {
	var lastIndex uint64
	backoff := time.Second

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		opts := (&api.QueryOptions{
			WaitIndex: lastIndex,
			WaitTime:  n.cfg.WaitTime,
		}).WithContext(ctx)

		// passingOnly=false so we observe unhealthy peers too and can surface
		// their status, rather than having them silently vanish.
		entries, meta, err := n.client.Health().Service(n.cfg.ServiceName, "", false, opts)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			n.log.Warn("discovery query failed", "err", err, "retry_in", backoff)
			if sleep(ctx, backoff) {
				return
			}
			backoff = nextBackoff(backoff)
			continue
		}
		backoff = time.Second

		// Consul guarantees indexes only move forward; guard against a reset.
		if meta.LastIndex < lastIndex {
			lastIndex = 0
			continue
		}
		// Unchanged index means the wait timed out with no updates.
		if meta.LastIndex == lastIndex {
			continue
		}
		lastIndex = meta.LastIndex

		peers := make([]Peer, 0, len(entries))
		for _, e := range entries {
			peers = append(peers, Peer{
				ID:      e.Service.ID,
				Node:    e.Node.Node,
				Address: serviceAddr(e),
				Port:    e.Service.Port,
				Status:  e.Checks.AggregatedStatus(),
				Tags:    e.Service.Tags,
				Meta:    e.Service.Meta,
			})
		}
		n.setPeers(peers)

		healthy := 0
		for _, p := range peers {
			if p.Status == api.HealthPassing {
				healthy++
			}
		}
		n.log.Info("peer view updated", "total", len(peers), "healthy", healthy, "index", lastIndex)
	}
}

// serviceAddr prefers the service-level address, falling back to the node
// address when the service did not advertise its own.
func serviceAddr(e *api.ServiceEntry) string {
	if e.Service.Address != "" {
		return e.Service.Address
	}
	return e.Node.Address
}
