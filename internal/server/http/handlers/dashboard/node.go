package dashboard

import (
	"sort"
	"strings"

	"consul-journey/internal/node"

	"github.com/gofiber/fiber/v3"
)

func (h *Handler) nodePeersHandler(c fiber.Ctx) error {
	currentID := h.node.ID()
	self := h.node.Self()
	peers := h.node.GetPeers()
	leaderID := h.node.LeaderID()

	sort.Slice(peers, func(i, j int) bool {
		if peers[i].Node != peers[j].Node {
			return peers[i].Node < peers[j].Node
		}
		return peers[i].ID < peers[j].ID
	})

	rows := make([]peerRow, 0, len(peers)+1)
	if self != nil {
		rows = append(rows, newRow(self, currentID, leaderID))
	}
	for _, p := range peers {
		rows = append(rows, newRow(p, currentID, leaderID))
	}

	healthy := 0
	leaderNode := ""
	for _, r := range rows {
		if isPassing(r.Status) {
			healthy++
		}
		if r.IsLeader {
			leaderNode = r.Node
		}
	}

	view := peersView{
		baseView:     h.base(),
		Rows:         rows,
		Total:        len(rows),
		Healthy:      healthy,
		ElectionOn:   h.node.LeaderElectionEnabled(),
		LeaderKnown:  leaderID != "",
		LeaderNode:   leaderNode,
		LeaderID:     leaderID,
		PeerBasePath: c.Path(),
	}
	return h.render(c, fiber.StatusOK, "peers", view)
}

func (h *Handler) nodePeersIDHandler(peersPath string) fiber.Handler {
	return func(c fiber.Ctx) error {
		id := c.Params("id")
		if id == "" {
			return h.renderError(c, ErrPeerIDRequired, peersPath)
		}

		peer := h.node.GetPeer(id)
		if peer == nil {
			return h.renderError(c, ErrPeerNotFound, peersPath)
		}

		leaderID := h.node.LeaderID()
		row := newRow(peer, h.node.ID(), leaderID)

		view := peerView{
			baseView:   h.base(),
			Row:        row,
			Meta:       sortedMeta(peer.Meta),
			ElectionOn: h.node.LeaderElectionEnabled(),
			ListPath:   strings.TrimSuffix(c.Path(), "/"+id),
		}
		return h.render(c, fiber.StatusOK, "peer", view)
	}
}

func newRow(p *node.Peer, currentID, leaderID string) peerRow {
	return peerRow{
		Peer:      p,
		IsCurrent: p.ID == currentID,
		IsLeader:  leaderID != "" && p.ID == leaderID,
	}
}
