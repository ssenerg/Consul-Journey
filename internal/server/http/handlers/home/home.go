package home

import (
	"consul-journey/internal/server/http/handlers"

	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	redirectTo string
}

var _ handlers.Handler = (*Handler)(nil)

func New(redirectTo string) *Handler {
	return &Handler{
		redirectTo: redirectTo,
	}
}

func (h *Handler) Register(router *fiber.Group) {
	router.Get("/", h.homeHandler)
}

func (h *Handler) homeHandler(c fiber.Ctx) error {
	return c.Redirect().Status(fiber.StatusTemporaryRedirect).To(h.redirectTo)
}
