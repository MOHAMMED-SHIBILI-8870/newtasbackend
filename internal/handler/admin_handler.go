package handler

import (
	"backend/internal/usecase"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type AdminHandler struct {
	usecase usecase.AdminUsecase
}

func NewAdminHandler(u usecase.AdminUsecase) *AdminHandler {
	return &AdminHandler{usecase: u}
}

func (h *AdminHandler) ListUsers(c *fiber.Ctx) error {
	// Capture queries: ?role=guide&search=shibili
	role := c.Query("role")
	search := c.Query("search")

	users, err := h.usecase.FetchUsers(c.Context(), role, search)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}
	return c.JSON(users)
}

// 🔒 Block / Unblock User
func (h *AdminHandler) BlockUser(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid user ID",
		})
	}

	name, isBlocked, err := h.usecase.ToggleUserBlock(c.Context(), uint(id))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message":    "user status toggled successfully",
		"name":       name,
		"is_blocked": isBlocked,
	})
}

func (h *AdminHandler) UpdateRole(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	type RequestBody struct {
		Role string `json:"role"`
	}

	var body RequestBody
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON body"})
	}

	// Quick check: don't call usecase if role is empty
	if body.Role == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Role is required"})
	}

	// Now this will not throw an error because it's in the interface!
	if err := h.usecase.ChangeUserRole(c.Context(), uint(id), body.Role); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message": fmt.Sprintf("User role successfully updated to %s", body.Role),
	})
}
