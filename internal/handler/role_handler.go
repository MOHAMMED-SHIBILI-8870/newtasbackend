package handler

import (
	"backend/internal/dto"
	"backend/internal/entity"
	"backend/internal/middleware"
	"backend/internal/response"
	"backend/internal/usecase"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type RoleHandler struct {
	roleUsecase       *usecase.RoleUsecase
	permissionUsecase *usecase.PermissionUsecase
}

func NewRoleHandler(roleUsecase *usecase.RoleUsecase, permissionUsecase *usecase.PermissionUsecase) *RoleHandler {
	return &RoleHandler{
		roleUsecase:       roleUsecase,
		permissionUsecase: permissionUsecase,
	}
}

func roleStatusFromErr(err error) int {
	if err == nil {
		return fiber.StatusBadRequest
	}

	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "not found"):
		return fiber.StatusNotFound
	case strings.Contains(msg, "access denied"):
		return fiber.StatusForbidden
	case strings.Contains(msg, "already exists"):
		return fiber.StatusConflict
	default:
		return fiber.StatusBadRequest
	}
}

func mapRoleResponse(role entity.Role) dto.RoleResponse {
	return dto.RoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	}
}

func (h *RoleHandler) ListRoles(c *fiber.Ctx) error {
	roles, err := h.roleUsecase.ListRoles(c.Context())
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "failed to load roles", err)
	}

	items := make([]dto.RoleResponse, 0, len(roles))
	for _, role := range roles {
		items = append(items, mapRoleResponse(role))
	}

	return response.Success(c, fiber.StatusOK, "roles loaded successfully", items)
}

func (h *RoleHandler) CreateRole(c *fiber.Ctx) error {
	var input dto.RoleRequest
	if err := c.BodyParser(&input); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid request body", err)
	}

	role := &entity.Role{
		Name:        input.Name,
		Description: input.Description,
	}

	if err := h.roleUsecase.CreateRole(c.Context(), role); err != nil {
		return response.Fail(c, roleStatusFromErr(err), "failed to create role", err)
	}

	return response.Success(c, fiber.StatusCreated, "role created successfully", mapRoleResponse(*role))
}

func (h *RoleHandler) UpdateRole(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid role id", err)
	}

	var input dto.RoleRequest
	if err := c.BodyParser(&input); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid request body", err)
	}

	role := &entity.Role{
		Name:        input.Name,
		Description: input.Description,
	}

	if err := h.roleUsecase.UpdateRole(c.Context(), uint(id), role); err != nil {
		return response.Fail(c, roleStatusFromErr(err), "failed to update role", err)
	}

	return response.Success(c, fiber.StatusOK, "role updated successfully", nil)
}

func (h *RoleHandler) DeleteRole(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid role id", err)
	}

	if err := h.roleUsecase.DeleteRole(c.Context(), uint(id)); err != nil {
		return response.Fail(c, roleStatusFromErr(err), "failed to delete role", err)
	}

	return response.Success(c, fiber.StatusOK, "role deleted successfully", nil)
}

func (h *RoleHandler) AssignRoleToUser(c *fiber.Ctx) error {
	userID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid user id", err)
	}

	var input dto.AssignRoleRequest
	if err := c.BodyParser(&input); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid request body", err)
	}

	if err := h.roleUsecase.AssignRoleToUser(c.Context(), uint(userID), input.RoleID); err != nil {
		return response.Fail(c, roleStatusFromErr(err), "failed to assign role", err)
	}

	return response.Success(c, fiber.StatusOK, "role assigned successfully", nil)
}

func (h *RoleHandler) GetMyAccess(c *fiber.Ctx) error {
	userID := middleware.GetAuthUserID(c)
	if userID == 0 {
		return response.Fail(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	roles, err := h.roleUsecase.GetUserRoles(c.Context(), userID)
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "failed to load role access", err)
	}

	permissions := []string{}
	if h.permissionUsecase != nil {
		permissions, err = h.permissionUsecase.GetUserPermissionKeys(c.Context(), userID)
		if err != nil {
			return response.Fail(c, fiber.StatusInternalServerError, "failed to load permissions", err)
		}
	}

	roleNames := make([]string, 0, len(roles))
	for _, role := range roles {
		roleNames = append(roleNames, role.Name)
	}

	return response.Success(c, fiber.StatusOK, "access loaded successfully", dto.UserAccessResponse{
		UserID:      userID,
		Roles:       roleNames,
		Permissions: permissions,
	})
}
