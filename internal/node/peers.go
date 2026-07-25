package node

import (
	"maps"

	"consul-journey/internal"
	"consul-journey/internal/utils"

	capi "github.com/hashicorp/consul/api"
	"go.uber.org/zap"
)

type Peer struct {
	ID       string
	Address  string
	Node     string
	HTTPPort int
	GRPCPort int
	Status   string
	Tags     []string
	Meta     map[string]string
}

func (p *Peer) Clone() *Peer {
	return &Peer{
		ID:       p.ID,
		Address:  p.Address,
		Node:     p.Node,
		HTTPPort: p.HTTPPort,
		GRPCPort: p.GRPCPort,
		Status:   p.Status,
		Tags:     p.Tags[:],
		Meta:     maps.Clone(p.Meta),
	}
}

func (n *Node) runDiscovery() {
	var (
		lastIndex = uint64(0)
		backoff   = n.cfg.Discovery.Backoff
	)
	for {
		if utils.IsContextDone(n.ctx) {
			return
		}
		entries, meta, err := n.client.
			Health().
			Service(
				internal.Module(),
				"",
				n.cfg.Discovery.PassingOnly,
				(&capi.QueryOptions{
					Namespace:  n.cfg.Namespace,
					Partition:  n.cfg.Partition,
					Datacenter: n.cfg.Datacenter,
					WaitIndex:  lastIndex,
					WaitTime:   n.cfg.WaitTime,
				}).WithContext(n.ctx),
			)
		if utils.IsContextDone(n.ctx) {
			return
		}
		if err != nil {
			n.logger.Warn(
				"discovery query failed",
				zap.Error(err),
				zap.Uint64("last_index", lastIndex),
				zap.Duration("retry_in", backoff),
			)
			if utils.Sleep(n.ctx, backoff) {
				return
			}
			backoff = utils.NextBackoff(n.cfg.Discovery.Coefficient, n.cfg.Discovery.Max, backoff)
			continue
		}
		backoff = n.cfg.Discovery.Backoff
		// Consul guarantees indexes only move forward; guard against a reset.
		if meta.LastIndex < lastIndex {
			lastIndex = 0
			continue
		}
		// Unchanged index means the wait timed out with no updates.
		if meta.LastIndex == lastIndex {
			n.logger.Debug(
				"no peer updates since last index",
				zap.Uint64("last_index", lastIndex),
				zap.Duration("retry_in", backoff),
			)
			continue
		}
		lastIndex = meta.LastIndex

		healthy := 0
		peers := make([]*Peer, 0, len(entries))
		for _, e := range entries {
			httpPort, grpcPort := 0, 0
			for _, p := range e.Service.Ports {
				if p.Name == "http" {
					httpPort = p.Port
					if grpcPort != 0 {
						break
					}
				}
				if p.Name == "grpc" {
					grpcPort = p.Port
					if httpPort != 0 {
						break
					}
				}
			}
			if httpPort == 0 {
				n.logger.Warn(
					"no http port found for peer",
					zap.String("id", e.Service.ID),
					zap.String("node", e.Node.Node),
					zap.String("address", e.Service.Address),
				)
				continue
			}
			if grpcPort == 0 {
				n.logger.Warn(
					"no grpc port found for peer",
					zap.String("id", e.Service.ID),
					zap.String("node", e.Node.Node),
					zap.String("address", e.Service.Address),
				)
				continue
			}
			peer := &Peer{
				ID:       e.Service.ID,
				Address:  e.Service.Address,
				Node:     e.Node.Node,
				HTTPPort: httpPort,
				GRPCPort: grpcPort,
				Status:   e.Checks.AggregatedStatus(),
				Tags:     e.Service.Tags,
				Meta:     maps.Clone(e.Service.Meta),
			}
			if peer.Status == capi.HealthPassing {
				healthy++
			}
			peers = append(peers, peer)
		}
		n.setPeers(peers)
		n.logger.Debug(
			"updated peer view",
			zap.Int("total", len(peers)),
			zap.Int("healthy", healthy),
			zap.Uint64("index", lastIndex),
			zap.Duration("retry_in", backoff),
		)
	}
}

func (n *Node) setPeers(peers []*Peer) {
	n.pmu.Lock()
	defer n.pmu.Unlock()
	n.peers = peers
}

func (n *Node) GetPeers() []*Peer {
	n.pmu.RLock()
	defer n.pmu.RUnlock()
	peers := make([]*Peer, 0, len(n.peers))
	for _, p := range n.peers {
		if p.ID == n.id {
			continue
		}
		peers = append(peers, p.Clone())
	}
	return peers
}

func (n *Node) GetPeer(id string) *Peer {
	n.pmu.RLock()
	defer n.pmu.RUnlock()
	for _, p := range n.peers {
		if p.ID == id {
			return p.Clone()
		}
	}
	return nil
}
