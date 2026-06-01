package entity

import "time"

type RolePermission struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	RoleID       uint       `gorm:"not null;index;uniqueIndex:idx_role_permission" json:"role_id"`
	PermissionID uint       `gorm:"not null;index;uniqueIndex:idx_role_permission" json:"permission_id"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	Role         Role       `gorm:"foreignKey:RoleID;constraint:OnDelete:CASCADE;" json:"-"`
	Permission   Permission `gorm:"foreignKey:PermissionID;constraint:OnDelete:CASCADE;" json:"-"`
}
