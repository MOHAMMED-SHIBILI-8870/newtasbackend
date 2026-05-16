package handler

import (
	"backend/internal/entity"
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

//
// ➕ Admin Create User
//
func (h *AdminHandler) CreateUserByAdmin(c *fiber.Ctx) error {

	var req entity.AdminCreateUserRequest

	// Parse JSON body
	if err := c.BodyParser(&req); err != nil {

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "invalid request body",
			"error":   err.Error(),
		})
	}

	// Basic validation
	if req.FullName == "" ||
		req.Email == "" ||
		req.Password == "" ||
		req.Role == "" {

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "all fields are required",
		})
	}

	// Create user
	user, err := h.usecase.CreateUserByAdmin(
		c.Context(),
		req,
	)

	if err != nil {

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	// Success response
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "user created successfully",
		"data": fiber.Map{
			"id":          user.ID,
			"full_name":   user.FullName,
			"email":       user.Email,
			"role":        user.Role,
			"is_blocked":  user.IsBlocked,
			"is_verified": user.IsVerified,
		},
	})
}

func (h *AdminHandler) DeleteUser(c *fiber.Ctx) error {
    idParam := c.Params("id")
    targetID, _ := strconv.ParseUint(idParam, 10, 32)

    // --- FIX START ---
    // Try different common keys that your middleware might be using
    var adminIDVal interface{}
    
    keysToTry := []string{"admin_id", "user_id", "id", "sub"}
    for _, key := range keysToTry {
        if val := c.Locals(key); val != nil {
            adminIDVal = val
            break
        }
    }

    if adminIDVal == nil {
        // If still nil, let's return a clearer error to help you debug
        return c.Status(401).JSON(fiber.Map{
            "error": "Unauthorized: No user identifier found in request context",
            "hint":  "Check if your middleware is correctly setting locals like 'user_id' or 'admin_id'",
        })
    }

    // Attempt to convert whatever is there to uint safely
    var adminID uint
    switch v := adminIDVal.(type) {
    case uint:
        adminID = v
    case int:
        adminID = uint(v)
    case float64:
        adminID = uint(v)
    case string:
        // In case the middleware stored it as a string
        parsed, _ := strconv.ParseUint(v, 10, 32)
        adminID = uint(parsed)
    default:
        return c.Status(500).JSON(fiber.Map{"error": "admin_id type mismatch"})
    }

    err := h.usecase.RemoveUser(adminID, uint(targetID))
    if err != nil {
        status := 400
        if err.Error() == "user not found" {
            status = 404
        }
        return c.Status(status).JSON(fiber.Map{"error": err.Error()})
    }

    return c.Status(200).JSON(fiber.Map{
        "message": "User deleted successfully",
    })
}