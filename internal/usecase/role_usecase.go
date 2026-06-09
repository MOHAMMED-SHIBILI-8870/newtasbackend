package usecase

import (
	"backend/internal/entity"
	"backend/internal/repository"
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type RoleUsecase struct {
	roleRepo repository.RoleRepository
	userRepo repository.UserRepository
	db       *gorm.DB
}

func NewRoleUsecase(roleRepo repository.RoleRepository, userRepo repository.UserRepository, db *gorm.DB) *RoleUsecase {
	return &RoleUsecase{
		roleRepo: roleRepo,
		userRepo: userRepo,
		db:       db,
	}
}

// syncPrimaryRoleTx keeps the join table and cached role column aligned in one transaction.
func syncPrimaryRoleTx(tx *gorm.DB, userID uint, roleID uint, roleName string) error {
	if err := tx.Where("user_id = ?", userID).Delete(&entity.UserRole{}).Error; err != nil {
		return err
	}

	if err := tx.Create(&entity.UserRole{
		UserID:    userID,
		RoleID:    roleID,
		IsPrimary: true,
	}).Error; err != nil {
		return err
	}

	return tx.Model(&entity.User{}).
		Where("id = ?", userID).
		Update("role", roleName).Error
}

func (u *RoleUsecase) ListRoles(ctx context.Context) ([]entity.Role, error) {
	return u.roleRepo.List(ctx)
}

func (u *RoleUsecase) CreateRole(ctx context.Context, role *entity.Role) error {
	if role == nil {
		return errors.New("role is required")
	}

	normalized := strings.ToLower(strings.TrimSpace(role.Name))
	if !IsValidRole(normalized) {
		return fmt.Errorf("invalid role")
	}
	role.Name = NormalizeRole(normalized)

	existing, err := u.roleRepo.GetByName(ctx, role.Name)
	if err != nil {
		return err
	}
	if existing != nil {
		return fmt.Errorf("role already exists")
	}

	if strings.TrimSpace(role.Description) == "" {
		role.Description = role.Name + " role"
	}

	return u.roleRepo.Create(ctx, role)
}

func (u *RoleUsecase) UpdateRole(ctx context.Context, id uint, role *entity.Role) error {
	if id == 0 {
		return errors.New("role id is required")
	}
	if role == nil {
		return errors.New("role is required")
	}

	existing, err := u.roleRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("role not found")
	}

	if strings.TrimSpace(role.Name) != "" {
		normalized := strings.ToLower(strings.TrimSpace(role.Name))
		if !IsValidRole(normalized) {
			return fmt.Errorf("invalid role")
		}
		normalized = NormalizeRole(normalized)
		if normalized != existing.Name {
			return errors.New("role name cannot be changed")
		}
		existing.Name = normalized
	}

	if strings.TrimSpace(role.Description) != "" {
		existing.Description = strings.TrimSpace(role.Description)
	}

	return u.roleRepo.Update(ctx, existing)
}

func (u *RoleUsecase) DeleteRole(ctx context.Context, id uint) error {
	if id == 0 {
		return errors.New("role id is required")
	}
	if u.db == nil {
		return errors.New("database unavailable")
	}

	role, err := u.roleRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if role == nil {
		return fmt.Errorf("role not found")
	}

	if role.Name == "admin" || role.Name == "user" {
		return fmt.Errorf("default roles cannot be deleted")
	}

	var count int64
	if err := u.db.WithContext(ctx).Model(&entity.UserRole{}).Where("role_id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("role is assigned to users")
	}

	return u.roleRepo.Delete(ctx, id)
}

func (u *RoleUsecase) AssignRoleToUser(ctx context.Context, userID uint, roleID uint) error {
	if userID == 0 || roleID == 0 {
		return errors.New("user and role are required")
	}
	if u.db == nil {
		return errors.New("database unavailable")
	}

	role, err := u.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return err
	}
	if role == nil {
		return fmt.Errorf("role not found")
	}

	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	return u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return syncPrimaryRoleTx(tx, userID, role.ID, role.Name)
	})
}

func (u *RoleUsecase) GetUserRoles(ctx context.Context, userID uint) ([]entity.Role, error) {
	roles, err := u.roleRepo.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}

	if len(roles) > 0 {
		return roles, nil
	}

	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil || user == nil {
		return roles, err
	}

	role, err := u.roleRepo.GetByName(ctx, NormalizeRole(user.Role))
	if err != nil {
		return nil, err
	}
	if role != nil {
		return []entity.Role{*role}, nil
	}

	return roles, nil
}
