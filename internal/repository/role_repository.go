package repository

import (
	"backend/internal/entity"
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type RoleRepository interface {
	Create(ctx context.Context, role *entity.Role) error
	Update(ctx context.Context, role *entity.Role) error
	Delete(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*entity.Role, error)
	GetByName(ctx context.Context, name string) (*entity.Role, error)
	List(ctx context.Context) ([]entity.Role, error)
	AssignRoleToUser(ctx context.Context, userID uint, roleID uint) error
	RemoveUserRoles(ctx context.Context, userID uint) error
	GetUserRoles(ctx context.Context, userID uint) ([]entity.Role, error)
	GetUserRoleIDs(ctx context.Context, userID uint) ([]uint, error)
}

type roleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) Create(ctx context.Context, role *entity.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *roleRepository) Update(ctx context.Context, role *entity.Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}

func (r *roleRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.Role{}, id).Error
}

func (r *roleRepository) GetByID(ctx context.Context, id uint) (*entity.Role, error) {
	var role entity.Role
	err := r.db.WithContext(ctx).First(&role, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("role not found")
		}
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) GetByName(ctx context.Context, name string) (*entity.Role, error) {
	var role entity.Role
	err := r.db.WithContext(ctx).
		Where("LOWER(name) = LOWER(?)", strings.TrimSpace(name)).
		First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) List(ctx context.Context) ([]entity.Role, error) {
	var roles []entity.Role
	err := r.db.WithContext(ctx).Order("name ASC").Find(&roles).Error
	return roles, err
}

func (r *roleRepository) AssignRoleToUser(ctx context.Context, userID uint, roleID uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&entity.UserRole{}).Error; err != nil {
			return err
		}

		userRole := entity.UserRole{
			UserID:    userID,
			RoleID:    roleID,
			IsPrimary: true,
		}

		return tx.Create(&userRole).Error
	})
}

func (r *roleRepository) RemoveUserRoles(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&entity.UserRole{}).Error
}

func (r *roleRepository) GetUserRoles(ctx context.Context, userID uint) ([]entity.Role, error) {
	var userRoles []entity.UserRole
	err := r.db.WithContext(ctx).
		Preload("Role").
		Where("user_id = ?", userID).
		Order("is_primary DESC, id ASC").
		Find(&userRoles).Error
	if err != nil {
		return nil, err
	}

	roles := make([]entity.Role, 0, len(userRoles))
	for _, link := range userRoles {
		if link.Role.ID != 0 {
			roles = append(roles, link.Role)
		}
	}

	return roles, nil
}

func (r *roleRepository) GetUserRoleIDs(ctx context.Context, userID uint) ([]uint, error) {
	var userRoles []entity.UserRole
	err := r.db.WithContext(ctx).
		Select("role_id").
		Where("user_id = ?", userID).
		Order("is_primary DESC, id ASC").
		Find(&userRoles).Error
	if err != nil {
		return nil, err
	}

	roleIDs := make([]uint, 0, len(userRoles))
	for _, link := range userRoles {
		if link.RoleID != 0 {
			roleIDs = append(roleIDs, link.RoleID)
		}
	}

	return roleIDs, nil
}
