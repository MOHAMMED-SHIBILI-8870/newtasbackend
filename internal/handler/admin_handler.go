package handler

import (
	"backend/internal/entity"
	"backend/internal/middleware"
	"backend/internal/response"
	"backend/internal/usecase"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type AdminHandler struct {
	usecase usecase.AdminUsecase
}

func NewAdminHandler(u usecase.AdminUsecase) *AdminHandler {
	return &AdminHandler{usecase: u}
}

func adminStatusFromErr(err error) int {
	if err == nil {
		return fiber.StatusBadRequest
	}

	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "not found"):
		return fiber.StatusNotFound
	case strings.Contains(msg, "invalid role"):
		return fiber.StatusBadRequest
	case strings.Contains(msg, "security risk"):
		return fiber.StatusForbidden
	default:
		return fiber.StatusBadRequest
	}
}

func (h *AdminHandler) ListUsers(c *fiber.Ctx) error {
	role := c.Query("role")
	search := c.Query("search")

	users, err := h.usecase.FetchUsers(c.Context(), role, search)
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "failed to fetch users", err)
	}

	return response.Success(c, fiber.StatusOK, "users loaded successfully", users)
}

func (h *AdminHandler) BlockUser(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid user id", err)
	}

	name, isBlocked, err := h.usecase.ToggleUserBlock(c.Context(), uint(id))
	if err != nil {
		return response.Fail(c, adminStatusFromErr(err), "failed to update user status", err)
	}

	return response.Success(c, fiber.StatusOK, "user status updated successfully", fiber.Map{
		"name":       name,
		"is_blocked": isBlocked,
	})
}

func (h *AdminHandler) UpdateRole(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid user id", err)
	}

	var body struct {
		Role string `json:"role"`
	}
	if err := c.BodyParser(&body); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid request body", err)
	}

	if body.Role == "" {
		return response.Fail(c, fiber.StatusBadRequest, "role is required", nil)
	}

	if err := h.usecase.ChangeUserRole(c.Context(), uint(id), body.Role); err != nil {
		return response.Fail(c, adminStatusFromErr(err), "failed to update role", err)
	}

	return response.Success(c, fiber.StatusOK, "user role updated successfully", fiber.Map{
		"role": usecase.NormalizeRole(body.Role),
	})
}

func (h *AdminHandler) CreateUserByAdmin(c *fiber.Ctx) error {
	var req entity.AdminCreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid request body", err)
	}

	if req.FullName == "" || req.Email == "" || req.Password == "" || req.Role == "" {
		return response.Fail(c, fiber.StatusBadRequest, "all fields are required", nil)
	}

	user, err := h.usecase.CreateUserByAdmin(c.Context(), req)
	if err != nil {
		return response.Fail(c, adminStatusFromErr(err), "failed to create user", err)
	}

	return response.Success(c, fiber.StatusCreated, "user created successfully", fiber.Map{
		"id":          user.ID,
		"full_name":   user.FullName,
		"email":       user.Email,
		"role":        usecase.NormalizeRole(user.Role),
		"is_blocked":  user.IsBlocked,
		"is_verified": user.IsVerified,
	})
}

func (h *AdminHandler) DeleteUser(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid user id", err)
	}

	adminID := middleware.GetAuthUserID(c)
	if adminID == 0 {
		return response.Fail(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	if err := h.usecase.RemoveUser(c.Context(), adminID, uint(id)); err != nil {
		return response.Fail(c, adminStatusFromErr(err), "failed to delete user", err)
	}

	return response.Success(c, fiber.StatusOK, "user deleted successfully", nil)
}
