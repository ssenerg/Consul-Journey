package dashboard

import "github.com/gofiber/fiber/v3"

func (h *DashboardHandler) nodePeersHandler(c fiber.Ctx) error {
	peers := h.node.GetPeers()
	_ = peers
	return c.Status(fiber.StatusOK).JSON(peers)
}

func (h *DashboardHandler) nodePeersIDHandler(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return ErrPeerIDRequired
	}
	peer := h.node.GetPeer(id)
	if peer == nil {
		return ErrPeerNotFound
	}
	_ = peer
	return nil // TODO
}
