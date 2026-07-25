package dashboard

import (
	"strings"

	"consul-journey/internal/node"
	"consul-journey/internal/server/http/handlers"

	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	node *node.Node
}

var _ handlers.Handler = (*Handler)(nil)

func New(node *node.Node) *Handler {
	return &Handler{
		node: node,
	}
}

func (h *Handler) Register(router fiber.Router) {
	router.Get("/", h.homeHandler)
	router.Get("/healthz", h.healthzHandler)
	h.registerNode(router.Group("/node"))
}

func (h *Handler) registerNode(router fiber.Router) {
	router.Get("/peers", h.nodePeersHandler)
	router.Get("/peers/:id", h.nodePeersIDHandler)
}

func (h *Handler) homeHandler(c fiber.Ctx) error {
	return c.Redirect().Status(fiber.StatusTemporaryRedirect).To(strings.TrimSuffix(c.Path(), "/") + "/node/peers")
}

func (h *Handler) healthzHandler(c fiber.Ctx) error {
	return c.SendStatus(fiber.StatusOK) // TODO: implement health check
}
