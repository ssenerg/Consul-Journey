package dashboard

import (
	"consul-journey/internal/node"
	"consul-journey/internal/server/http/handlers"

	"github.com/gofiber/fiber/v3"
)

type DashboardHandler struct {
	node *node.Node
}

var _ handlers.Handler = (*DashboardHandler)(nil)

func New(node *node.Node) *DashboardHandler {
	return &DashboardHandler{
		node: node,
	}
}

func (h *DashboardHandler) Register(router fiber.Router) {
	h.registerNode(router.Group("/node"))
}

func (h *DashboardHandler) registerNode(router fiber.Router) {
	router.Get("/peers", h.nodePeersHandler)
}
