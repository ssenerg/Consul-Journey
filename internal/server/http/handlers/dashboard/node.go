package dashboard

import "github.com/gofiber/fiber/v3"

func (h *DashboardHandler) nodePeersHandler(c fiber.Ctx) error {
	peers := h.node.GetPeers()
	_ = peers
	return nil // TODO
}
