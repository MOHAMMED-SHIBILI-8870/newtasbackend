package entity

import "time"

type Role struct {
	ID              uint             `gorm:"primaryKey" json:"id"`
	Name            string           `gorm:"size:50;uniqueIndex;not null" json:"name"`
	Description     string           `gorm:"type:text" json:"description,omitempty"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
	UserRoles       []UserRole       `gorm:"foreignKey:RoleID;constraint:OnDelete:CASCADE;" json:"-"`
	RolePermissions []RolePermission `gorm:"foreignKey:RoleID;constraint:OnDelete:CASCADE;" json:"-"`
}

type UserRole struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;index;uniqueIndex:idx_user_role" json:"user_id"`
	RoleID    uint      `gorm:"not null;index;uniqueIndex:idx_user_role" json:"role_id"`
	IsPrimary bool      `gorm:"default:true;index" json:"is_primary"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	User      User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;" json:"-"`
	Role      Role      `gorm:"foreignKey:RoleID;constraint:OnDelete:CASCADE;" json:"-"`
}
