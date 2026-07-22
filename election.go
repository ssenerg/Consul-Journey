package main

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/hashicorp/consul/api"
)

// runElection continuously participates in leader election. The lifecycle is:
//
//  1. Create a Consul session bound to this node's health checks. While the
//     node is healthy the session is renewed; if the node becomes unhealthy or
//     the process dies, the session is invalidated by Consul.
//  2. Campaign: watch the leader KV key with blocking queries. Whenever the
//     key is unheld, attempt to acquire it with our session. Holding the key
//     == being the leader. Because the lock is tied to the session, losing
//     health automatically releases leadership to another node.
//  3. If the session is lost, tear down and rebuild it after a short backoff.
//
// This is the canonical Consul leader-election pattern (session + KV Acquire),
// implemented explicitly so leadership changes can be observed and reacted to.
func (n *Node) runElection(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if err := n.campaign(ctx); err != nil && ctx.Err() == nil {
			n.log.Warn("election campaign ended, rebuilding session", "err", err)
		}
		n.setLeadership(false, n.observedLeader())

		if sleep(ctx, n.cfg.CampaignBackoff) {
			return
		}
	}
}

// createSession opens a session whose liveness is gated on this node's health.
//
//   - Behavior "release": on invalidation the held lock is released (freed for
//     others) rather than deleted, preserving the key's value for observers.
//   - Checks serfHealth + our TTL check: the session dies if the node leaves
//     the gossip pool OR the app stops heartbeating. This is what ties
//     leadership to real health.
//   - LockDelay: after a session is invalidated, Consul refuses new acquires
//     of the freed lock for this window, preventing a split brain from a
//     partition flapping the lock rapidly.
func (n *Node) createSession(ctx context.Context) (string, error) {
	entry := &api.SessionEntry{
		Name:      "leader-election:" + n.cfg.NodeID,
		Behavior:  api.SessionBehaviorRelease,
		TTL:       n.cfg.SessionTTL.String(),
		LockDelay: n.cfg.SessionLockDelay,
		Checks:    []string{"serfHealth", n.cfg.ttlCheckID()},
	}
	id, _, err := n.client.Session().Create(entry, (&api.WriteOptions{}).WithContext(ctx))
	if err != nil {
		return "", err
	}
	return id, nil
}

func (n *Node) campaign(ctx context.Context) error {
	sessionID, err := n.createSession(ctx)
	if err != nil {
		return err
	}
	n.setSession(sessionID)
	n.log.Info("session created", "session", short(sessionID))

	// Renew the session's TTL in the background until this campaign ends.
	// RenewPeriodic returns (closing renewDone) when the session can no longer
	// be renewed — our signal that the session is gone and we must rebuild.
	renewDone := make(chan struct{})
	renewCtx, cancelRenew := context.WithCancel(ctx)
	go func() {
		defer close(renewDone)
		_ = n.client.Session().RenewPeriodic(
			n.cfg.SessionTTL.String(), sessionID, nil, renewCtx.Done(),
		)
	}()

	defer func() {
		cancelRenew()
		<-renewDone
		// Best-effort explicit destroy so leadership frees immediately on a
		// clean exit rather than waiting for TTL expiry.
		_, _ = n.client.Session().Destroy(sessionID, nil)
		n.setSession("")
		n.log.Info("session destroyed", "session", short(sessionID))
	}()

	kv := n.client.KV()
	var lastIndex uint64

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-renewDone:
			return errors.New("session invalidated")
		default:
		}

		opts := (&api.QueryOptions{
			WaitIndex: lastIndex,
			WaitTime:  n.cfg.WaitTime,
		}).WithContext(ctx)

		pair, meta, err := kv.Get(n.cfg.LeaderKey, opts)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}
		if meta != nil {
			lastIndex = meta.LastIndex
		}

		held := pair != nil && pair.Session != ""

		if !held {
			// The lock is free — try to grab it. A successful acquire binds
			// the key to our session and makes us leader.
			acquired, _, err := kv.Acquire(&api.KVPair{
				Key:     n.cfg.LeaderKey,
				Value:   n.leaderPayload(sessionID),
				Session: sessionID,
			}, nil)
			if err != nil {
				if ctx.Err() != nil {
					return nil
				}
				return err
			}
			if acquired {
				n.log.Info("won leadership election")
				// Loop back around; the next Get will confirm us as holder.
				continue
			}
			// Lost the race to another candidate; the next blocking Get will
			// wake us when the situation changes.
			n.setLeadership(false, nil)
			continue
		}

		// Key is held by someone (possibly us). Record who.
		info := parseLeader(pair.Value)
		n.setLeadership(pair.Session == sessionID, info)
	}
}

// leaderPayload is the value stored under the leader key while we hold it.
func (n *Node) leaderPayload(sessionID string) []byte {
	info := LeaderInfo{
		NodeID:     n.cfg.NodeID,
		Address:    n.cfg.AdvertiseAddr,
		HTTPAddr:   n.cfg.AdvertiseHTTPAddr(),
		SessionID:  sessionID,
		AcquiredAt: time.Now().UTC(),
	}
	b, _ := json.Marshal(info)
	return b
}

// observedLeader reads the current leader key without blocking, for status.
func (n *Node) observedLeader() *LeaderInfo {
	pair, _, err := n.client.KV().Get(n.cfg.LeaderKey, nil)
	if err != nil || pair == nil || pair.Session == "" {
		return nil
	}
	return parseLeader(pair.Value)
}

func parseLeader(b []byte) *LeaderInfo {
	if len(b) == 0 {
		return nil
	}
	var info LeaderInfo
	if err := json.Unmarshal(b, &info); err != nil {
		return nil
	}
	return &info
}
