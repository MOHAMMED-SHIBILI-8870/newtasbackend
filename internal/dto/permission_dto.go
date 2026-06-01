package dto

import "time"

type PermissionRequest struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type AssignPermissionRequest struct {
	PermissionID uint `json:"permission_id"`
}

type PermissionResponse struct {
	ID          uint      `json:"id"`
	Key         string    `json:"key"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
