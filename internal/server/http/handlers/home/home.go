package home

import (
	"strings"

	"consul-journey/internal/server/http/handlers"

	"github.com/gofiber/fiber/v3"
)

type Handler struct{}

var _ handlers.Handler = (*Handler)(nil)

func New() *Handler {
	return &Handler{}
}

func (h *Handler) Register(router fiber.Router) {
	router.Get("/", h.homeHandler)
}

func (h *Handler) homeHandler(c fiber.Ctx) error {
	return c.Redirect().Status(fiber.StatusTemporaryRedirect).To(strings.TrimSuffix(c.Path(), "/") + "/dashboard")
}
