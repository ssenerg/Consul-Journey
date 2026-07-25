package node

import (
	"context"
	"fmt"

	"consul-journey/internal"
	"consul-journey/internal/utils"

	capi "github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
	"go.uber.org/zap"
)

func (n *Node) runElection() {
	if !n.cfg.LeaderElection.Enabled {
		return
	}
	// can not fail
	watchPlan, err := watch.Parse(
		map[string]any{
			"type": "key",
			"key":  leaderElectionKey(),
		},
	)
	if err != nil {
		n.logger.Panic("failed to parse watch plan", zap.Error(err))
	}
	watchPlan.Handler = func(idx uint64, data any) {
		if data == nil {
			return
		}
		kvPair, ok := data.(*capi.KVPair)
		if !ok {
			return
		}
		if kvPair.Session == "" {
			n.logger.Info("leader lock released")
			n.setLeader("")
			n.leaderLockRelease <- struct{}{}
			return
		}
		newLeaderID := string(kvPair.Value)
		n.setLeader(newLeaderID)
		n.logger.Info("new leader elected", zap.String("leader_id", newLeaderID))
	}
	n.wg.Go(func() {
		if err := watchPlan.RunWithClientAndHclog(n.client, nil); err != nil {
			n.logger.Error("failed to run watch plan", zap.Error(err))
		}
		n.logger.Info("watch plan stopped")
	})
	n.wg.Go(func() {
		<-n.ctx.Done()
		watchPlan.Stop()
	})

	backoff := n.cfg.LeaderElection.Backoff
	for {
		if utils.IsContextDone(n.ctx) {
			return
		}
		sessionID, _, err := n.client.Session().Create(
			&capi.SessionEntry{
				Name:      leaderElectionKey(),
				Behavior:  "release",
				TTL:       n.cfg.LeaderElection.TTL.String(),
				LockDelay: n.cfg.LeaderElection.LockDelay,
			},
			(&capi.WriteOptions{
				Namespace:  n.cfg.Namespace,
				Partition:  n.cfg.Partition,
				Datacenter: n.cfg.Datacenter,
			}).WithContext(n.ctx),
		)
		if utils.IsContextDone(n.ctx) {
			return
		}
		if err != nil {
			n.logger.Warn("failed to create leader election session", zap.Error(err))
			if utils.Sleep(n.ctx, backoff) {
				return
			}
			backoff = utils.NextBackoff(n.cfg.LeaderElection.Coefficient, n.cfg.LeaderElection.Max, backoff)
			continue
		}
		backoff = n.cfg.LeaderElection.Backoff

		sessCtx, sessCancel := context.WithCancel(n.ctx)
		renewStop := make(chan struct{})
		renewDone := make(chan struct{})
		go func() {
			defer close(renewDone)
			if err := n.client.Session().RenewPeriodic(
				n.cfg.LeaderElection.TTL.String(),
				sessionID,
				&capi.WriteOptions{
					Namespace:  n.cfg.Namespace,
					Partition:  n.cfg.Partition,
					Datacenter: n.cfg.Datacenter,
				},
				renewStop,
			); err != nil {
				n.logger.Warn("leader election session renewal ended", zap.Error(err))
			}
			sessCancel()
		}()

		n.holdLeadership(sessCtx, sessionID)

		close(renewStop)
		<-renewDone
		sessCancel()
	}
}

func (n *Node) holdLeadership(ctx context.Context, sessionID string) {
	kvPair := &capi.KVPair{
		Key:     leaderElectionKey(),
		Value:   []byte(n.id),
		Session: sessionID,
	}
	backoff := n.cfg.LeaderElection.Backoff
	for {
		if utils.IsContextDone(ctx) {
			return
		}
		acquired, _, err := n.client.KV().Acquire(
			kvPair,
			(&capi.WriteOptions{
				Namespace:  n.cfg.Namespace,
				Partition:  n.cfg.Partition,
				Datacenter: n.cfg.Datacenter,
			}).WithContext(ctx),
		)
		if utils.IsContextDone(ctx) {
			return
		}
		if err != nil {
			n.logger.Warn("failed to acquire leader election lock", zap.Error(err))
			if utils.Sleep(ctx, backoff) {
				return
			}
			backoff = utils.NextBackoff(n.cfg.LeaderElection.Coefficient, n.cfg.LeaderElection.Max, backoff)
			continue
		}
		backoff = n.cfg.LeaderElection.Backoff
		if acquired {
			n.logger.Info("acquired leadership", zap.String("session_id", sessionID))
		}
		select {
		case <-n.leaderLockRelease:
		case <-ctx.Done():
			return
		}
	}
}

func (n *Node) IsLeader() bool {
	if !n.cfg.LeaderElection.Enabled {
		return false
	}
	n.pmu.RLock()
	defer n.pmu.RUnlock()
	return n.leaderID == n.id
}

func (n *Node) setLeader(leaderID string) {
	n.pmu.Lock()
	defer n.pmu.Unlock()
	n.leaderID = leaderID
}

func leaderElectionKey() string {
	return fmt.Sprintf("services/%s/leader", internal.Module())
}
