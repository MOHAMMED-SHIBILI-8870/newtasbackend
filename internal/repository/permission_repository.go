package repository

import (
	"backend/internal/entity"
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type PermissionRepository interface {
	Create(ctx context.Context, permission *entity.Permission) error
	Update(ctx context.Context, permission *entity.Permission) error
	Delete(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*entity.Permission, error)
	GetByKey(ctx context.Context, key string) (*entity.Permission, error)
	List(ctx context.Context) ([]entity.Permission, error)
	AssignPermissionToRole(ctx context.Context, roleID uint, permissionID uint) error
	RemovePermissionFromRole(ctx context.Context, roleID uint, permissionID uint) error
	GetByRoleID(ctx context.Context, roleID uint) ([]entity.Permission, error)
	GetByRoleIDs(ctx context.Context, roleIDs []uint) ([]entity.Permission, error)
}

type permissionRepository struct {
	db *gorm.DB
}

func NewPermissionRepository(db *gorm.DB) PermissionRepository {
	return &permissionRepository{db: db}
}

func (r *permissionRepository) Create(ctx context.Context, permission *entity.Permission) error {
	return r.db.WithContext(ctx).Create(permission).Error
}

func (r *permissionRepository) Update(ctx context.Context, permission *entity.Permission) error {
	return r.db.WithContext(ctx).Save(permission).Error
}

func (r *permissionRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.Permission{}, id).Error
}

func (r *permissionRepository) GetByID(ctx context.Context, id uint) (*entity.Permission, error) {
	var permission entity.Permission
	err := r.db.WithContext(ctx).First(&permission, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("permission not found")
		}
		return nil, err
	}
	return &permission, nil
}

func (r *permissionRepository) GetByKey(ctx context.Context, key string) (*entity.Permission, error) {
	var permission entity.Permission
	err := r.db.WithContext(ctx).
		Where("LOWER(key) = LOWER(?)", strings.TrimSpace(key)).
		First(&permission).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &permission, nil
}

func (r *permissionRepository) List(ctx context.Context) ([]entity.Permission, error) {
	var permissions []entity.Permission
	err := r.db.WithContext(ctx).Order("key ASC").Find(&permissions).Error
	return permissions, err
}

func (r *permissionRepository) AssignPermissionToRole(ctx context.Context, roleID uint, permissionID uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		link := entity.RolePermission{RoleID: roleID, PermissionID: permissionID}
		return tx.Where("role_id = ? AND permission_id = ?", roleID, permissionID).
			FirstOrCreate(&link).Error
	})
}

func (r *permissionRepository) RemovePermissionFromRole(ctx context.Context, roleID uint, permissionID uint) error {
	return r.db.WithContext(ctx).
		Where("role_id = ? AND permission_id = ?", roleID, permissionID).
		Delete(&entity.RolePermission{}).Error
}

func (r *permissionRepository) GetByRoleID(ctx context.Context, roleID uint) ([]entity.Permission, error) {
	var links []entity.RolePermission
	err := r.db.WithContext(ctx).
		Preload("Permission").
		Where("role_id = ?", roleID).
		Find(&links).Error
	if err != nil {
		return nil, err
	}

	permissions := make([]entity.Permission, 0, len(links))
	for _, link := range links {
		if link.Permission.ID != 0 {
			permissions = append(permissions, link.Permission)
		}
	}

	return permissions, nil
}

func (r *permissionRepository) GetByRoleIDs(ctx context.Context, roleIDs []uint) ([]entity.Permission, error) {
	if len(roleIDs) == 0 {
		return []entity.Permission{}, nil
	}

	var links []entity.RolePermission
	err := r.db.WithContext(ctx).
		Preload("Permission").
		Where("role_id IN ?", roleIDs).
		Find(&links).Error
	if err != nil {
		return nil, err
	}

	seen := make(map[uint]struct{}, len(links))
	permissions := make([]entity.Permission, 0, len(links))
	for _, link := range links {
		if link.Permission.ID == 0 {
			continue
		}
		if _, ok := seen[link.Permission.ID]; ok {
			continue
		}
		seen[link.Permission.ID] = struct{}{}
		permissions = append(permissions, link.Permission)
	}

	return permissions, nil
}
