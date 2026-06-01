package handler

import (
	"backend/internal/dto"
	"backend/internal/entity"
	"backend/internal/response"
	"backend/internal/usecase"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type PermissionHandler struct {
	usecase *usecase.PermissionUsecase
}

func NewPermissionHandler(usecase *usecase.PermissionUsecase) *PermissionHandler {
	return &PermissionHandler{usecase: usecase}
}

func permissionStatusFromErr(err error) int {
	if err == nil {
		return fiber.StatusBadRequest
	}

	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "not found"):
		return fiber.StatusNotFound
	case strings.Contains(msg, "already exists"):
		return fiber.StatusConflict
	case strings.Contains(msg, "access denied"):
		return fiber.StatusForbidden
	default:
		return fiber.StatusBadRequest
	}
}

func mapPermissionResponse(permission entity.Permission) dto.PermissionResponse {
	return dto.PermissionResponse{
		ID:          permission.ID,
		Key:         permission.Key,
		Name:        permission.Name,
		Description: permission.Description,
		CreatedAt:   permission.CreatedAt,
		UpdatedAt:   permission.UpdatedAt,
	}
}

func (h *PermissionHandler) ListPermissions(c *fiber.Ctx) error {
	permissions, err := h.usecase.ListPermissions(c.Context())
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "failed to load permissions", err)
	}

	items := make([]dto.PermissionResponse, 0, len(permissions))
	for _, permission := range permissions {
		items = append(items, mapPermissionResponse(permission))
	}

	return response.Success(c, fiber.StatusOK, "permissions loaded successfully", items)
}

func (h *PermissionHandler) CreatePermission(c *fiber.Ctx) error {
	var input dto.PermissionRequest
	if err := c.BodyParser(&input); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid request body", err)
	}

	permission := &entity.Permission{
		Key:         input.Key,
		Name:        input.Name,
		Description: input.Description,
	}

	if err := h.usecase.CreatePermission(c.Context(), permission); err != nil {
		return response.Fail(c, permissionStatusFromErr(err), "failed to create permission", err)
	}

	return response.Success(c, fiber.StatusCreated, "permission created successfully", mapPermissionResponse(*permission))
}

func (h *PermissionHandler) UpdatePermission(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid permission id", err)
	}

	var input dto.PermissionRequest
	if err := c.BodyParser(&input); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid request body", err)
	}

	permission := &entity.Permission{
		Key:         input.Key,
		Name:        input.Name,
		Description: input.Description,
	}

	if err := h.usecase.UpdatePermission(c.Context(), uint(id), permission); err != nil {
		return response.Fail(c, permissionStatusFromErr(err), "failed to update permission", err)
	}

	return response.Success(c, fiber.StatusOK, "permission updated successfully", nil)
}

func (h *PermissionHandler) DeletePermission(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid permission id", err)
	}

	if err := h.usecase.DeletePermission(c.Context(), uint(id)); err != nil {
		return response.Fail(c, permissionStatusFromErr(err), "failed to delete permission", err)
	}

	return response.Success(c, fiber.StatusOK, "permission deleted successfully", nil)
}

func (h *PermissionHandler) AssignPermissionToRole(c *fiber.Ctx) error {
	roleID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid role id", err)
	}

	var input dto.AssignPermissionRequest
	if err := c.BodyParser(&input); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid request body", err)
	}

	if err := h.usecase.AssignPermissionToRole(c.Context(), uint(roleID), input.PermissionID); err != nil {
		return response.Fail(c, permissionStatusFromErr(err), "failed to assign permission", err)
	}

	return response.Success(c, fiber.StatusOK, "permission assigned successfully", nil)
}

func (h *PermissionHandler) RemovePermissionFromRole(c *fiber.Ctx) error {
	roleID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid role id", err)
	}

	permissionID, err := strconv.Atoi(c.Params("permission_id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid permission id", err)
	}

	if err := h.usecase.RemovePermissionFromRole(c.Context(), uint(roleID), uint(permissionID)); err != nil {
		return response.Fail(c, permissionStatusFromErr(err), "failed to remove permission", err)
	}

	return response.Success(c, fiber.StatusOK, "permission removed successfully", nil)
}

func (h *PermissionHandler) GetRolePermissions(c *fiber.Ctx) error {
	roleID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid role id", err)
	}

	permissions, err := h.usecase.GetPermissionsByRoleID(c.Context(), uint(roleID))
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "failed to load role permissions", err)
	}

	items := make([]dto.PermissionResponse, 0, len(permissions))
	for _, permission := range permissions {
		items = append(items, mapPermissionResponse(permission))
	}

	return response.Success(c, fiber.StatusOK, "role permissions loaded successfully", items)
}
