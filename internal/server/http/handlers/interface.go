package handlers

import "github.com/gofiber/fiber/v3"

type Handler interface {
	Register(router *fiber.Group)
}
