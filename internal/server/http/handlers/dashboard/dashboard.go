package dashboard

import (
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

func (h *Handler) Register(router *fiber.Group) {
	router.Get("/", h.homeHandler(router.Prefix+"/node/peers"))
	router.Get("/healthz", h.healthzHandler)
	h.registerNode(router.Group("/node").(*fiber.Group))
}

func (h *Handler) registerNode(router *fiber.Group) {
	router.Get("/peers", h.nodePeersHandler)
	router.Get("/peers/:id", h.nodePeersIDHandler(router.Prefix+"/peers"))
}

func (h *Handler) homeHandler(redirectTo string) fiber.Handler {
	return func(c fiber.Ctx) error {
		return c.Redirect().Status(fiber.StatusTemporaryRedirect).To(redirectTo)
	}
}

func (h *Handler) healthzHandler(c fiber.Ctx) error {
	return c.SendStatus(fiber.StatusOK) // TODO: implement health check
}
