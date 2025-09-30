package auth

import (
	"todo-api/internal/domain/auth"
	authService "todo-api/internal/service/auth"
	"todo-api/pkg/config"

	"github.com/gofiber/fiber/v2"
)

// Handler handles authentication HTTP requests
type Handler struct {
	authService authService.Service
}

// NewHandler creates a new auth handler instance
func NewHandler(config *config.Config) *Handler {
	// Initialize service
	authSvc := authService.NewService(config)

	return &Handler{
		authService: authSvc,
	}
}

// Login handles user login
func (h *Handler) Login(c *fiber.Ctx) error {
	var req auth.LoginRequest

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid request body",
		})
	}

	// Login user
	tokenResponse, err := h.authService.Login(&req)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"error":   false,
		"message": "Login successful",
		"data":    tokenResponse,
	})
}
