package dto

import "time"

type RoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type AssignRoleRequest struct {
	RoleID uint `json:"role_id"`
}

type RoleResponse struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type UserAccessResponse struct {
	UserID      uint     `json:"user_id"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
}
