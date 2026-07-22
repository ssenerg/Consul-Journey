package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// newHTTPServer builds this instance's HTTP surface. It exposes the health
// endpoint Consul scrapes plus operator-facing status/dashboard views.
func (n *Node) newHTTPServer() *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", n.handleHealth)
	mux.HandleFunc("/status", n.handleStatus)
	mux.HandleFunc("/", n.handleDashboard)

	return &http.Server{
		Addr:              n.cfg.HTTPListenAddr(),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
}

func (n *Node) serveHTTP(srv *http.Server) error {
	n.log.Info("http server listening", "addr", srv.Addr)
	err := srv.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

func (n *Node) shutdownHTTP(srv *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}

// handleHealth is the endpoint Consul's HTTP check scrapes. It returns 200 when
// the node is passing and 503 otherwise, so Consul flips the check on the
// status code alone (the body is for humans).
func (n *Node) handleHealth(w http.ResponseWriter, r *http.Request) {
	snap := n.assessHealth()
	code := http.StatusOK
	if snap.Status == "critical" {
		code = http.StatusServiceUnavailable
	}
	writeJSON(w, code, snap)
}

// statusResponse is the full operational view of this node.
type statusResponse struct {
	NodeID    string      `json:"node_id"`
	Service   string      `json:"service"`
	Address   string      `json:"address"`
	Uptime    string      `json:"uptime"`
	IsLeader  bool        `json:"is_leader"`
	Leader    *LeaderInfo `json:"leader"`
	SessionID string      `json:"session_id"`
	PeerCount int         `json:"peer_count"`
	Peers     []Peer      `json:"peers"`
}

func (n *Node) handleStatus(w http.ResponseWriter, r *http.Request) {
	isLeader, leader := n.LeadershipStatus()
	resp := statusResponse{
		NodeID:    n.cfg.NodeID,
		Service:   n.cfg.ServiceName,
		Address:   n.cfg.HTTPListenAddr(),
		Uptime:    n.Uptime().Round(time.Second).String(),
		IsLeader:  isLeader,
		Leader:    leader,
		SessionID: n.session(),
		Peers:     n.Peers(),
	}
	resp.PeerCount = len(resp.Peers)
	writeJSON(w, http.StatusOK, resp)
}

// handleDashboard renders a self-refreshing HTML view of the cluster as seen
// by this node: its role, the current leader, and every known peer.
func (n *Node) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	isLeader, leader := n.LeadershipStatus()
	peers := n.Peers()

	role := "follower"
	if isLeader {
		role = "LEADER"
	}
	leaderID := "(none)"
	if leader != nil {
		leaderID = leader.NodeID
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	_, _ = w.Write([]byte(`<!doctype html><meta http-equiv="refresh" content="2">` +
		`<style>body{font:14px/1.5 system-ui,sans-serif;margin:2rem;background:#0b0e14;color:#c9d1d9}` +
		`h1{font-size:1.2rem}code{color:#7ee787}table{border-collapse:collapse;margin-top:1rem;width:100%}` +
		`th,td{border:1px solid #30363d;padding:.4rem .6rem;text-align:left}th{background:#161b22}` +
		`.passing{color:#3fb950}.warning{color:#d29922}.critical{color:#f85149}.me{background:#132a1a}` +
		`.badge{padding:.1rem .5rem;border-radius:.4rem;background:#1f6feb;color:#fff}</style>`))

	writeHTMLf(w, `<h1>consul-journey &mdash; <span class="badge">%s</span> <code>%s</code></h1>`, role, n.cfg.NodeID)
	writeHTMLf(w, `<p>service <code>%s</code> &middot; uptime %s &middot; current leader <code>%s</code> &middot; session <code>%s</code></p>`,
		n.cfg.ServiceName, n.Uptime().Round(time.Second), leaderID, short(n.session()))

	writeHTMLf(w, `<table><tr><th>ID</th><th>Address</th><th>Port</th><th>Status</th><th>Role</th></tr>`)
	for _, p := range peers {
		rowClass := ""
		if p.ID == n.cfg.NodeID {
			rowClass = "me"
		}
		peerRole := ""
		if leader != nil && leader.NodeID == p.ID {
			peerRole = "leader"
		}
		writeHTMLf(w, `<tr class="%s"><td>%s</td><td>%s</td><td>%d</td><td class="%s">%s</td><td>%s</td></tr>`,
			rowClass, p.ID, p.Address, p.Port, p.Status, p.Status, peerRole)
	}
	writeHTMLf(w, `</table>`)
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}
