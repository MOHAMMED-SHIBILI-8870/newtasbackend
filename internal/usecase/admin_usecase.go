// ============================
// internal/usecase/admin_usecase.go
// ============================

package usecase

import (
	"backend/internal/entity"
	"backend/internal/repository"
	"context"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AdminUsecase interface {
	FetchUsers(ctx context.Context, role string, search string) ([]entity.User, error)
	ToggleUserBlock(ctx context.Context, targetID uint) (string, bool, error)
	ChangeUserRole(ctx context.Context, targetID uint, newRole string) error
	CreateUserByAdmin(ctx context.Context, req entity.AdminCreateUserRequest) (entity.User, error)
	RemoveUser(ctx context.Context, adminID uint, targetID uint) error
}

type adminUsecase struct {
	repo     repository.UserRepository
	roleRepo repository.RoleRepository
	db       *gorm.DB
}

func NewAdminUsecase(r repository.UserRepository, roleRepo repository.RoleRepository, db *gorm.DB) AdminUsecase {
	return &adminUsecase{
		repo:     r,
		roleRepo: roleRepo,
		db:       db,
	}
}

// 📄 Fetch users
func (u *adminUsecase) FetchUsers(
	ctx context.Context,
	role string,
	search string,
) ([]entity.User, error) {
	return u.repo.GetUsers(ctx, NormalizeRole(role), search)
}

// 🔒 Toggle block / unblock user
func (u *adminUsecase) ToggleUserBlock(
	ctx context.Context,
	targetID uint,
) (string, bool, error) {
	if u.repo == nil || u.db == nil {
		return "", false, errors.New("admin service unavailable")
	}

	user, err := u.repo.GetByID(ctx, targetID)
	if err != nil {
		return "", false, err
	}

	if user == nil {
		return "", false, fmt.Errorf("user not found")
	}

	if NormalizeRole(user.Role) == "admin" {
		return "", false, fmt.Errorf("cannot block admin user")
	}

	newStatus := !user.IsBlocked
	// Blocking a user must also revoke every live refresh token so the account
	// cannot continue minting access tokens after the status change.
	err = u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&entity.User{}).
			Where("id = ?", targetID).
			Update("is_blocked", newStatus)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("user not found")
		}

		if newStatus {
			if err := tx.Where("user_id = ?", targetID).Delete(&entity.RefreshToken{}).Error; err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return "", false, err
	}

	return user.FullName, newStatus, nil
}

// 🔄 Change user role
func (u *adminUsecase) ChangeUserRole(
	ctx context.Context,
	targetID uint,
	newRole string,
) error {
	if u.repo == nil || u.roleRepo == nil || u.db == nil {
		return errors.New("admin service unavailable")
	}

	normalized := strings.ToLower(strings.TrimSpace(newRole))
	if !IsValidRole(normalized) {
		return fmt.Errorf("invalid role: %s", newRole)
	}

	role, err := u.roleRepo.GetByName(ctx, normalized)
	if err != nil {
		return err
	}
	if role == nil {
		return fmt.Errorf("role not found")
	}

	user, err := u.repo.GetByID(ctx, targetID)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	return u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return syncPrimaryRoleTx(tx, user.ID, role.ID, role.Name)
	})
}

// ➕ Create new user by admin
func (u *adminUsecase) CreateUserByAdmin(
	ctx context.Context,
	req entity.AdminCreateUserRequest,
) (entity.User, error) {
	if u.repo == nil || u.roleRepo == nil || u.db == nil {
		return entity.User{}, errors.New("admin service unavailable")
	}

	normalized := strings.ToLower(strings.TrimSpace(req.Role))
	if !IsValidRole(normalized) {
		return entity.User{}, fmt.Errorf("invalid role")
	}

	roleName := NormalizeRole(normalized)
	role, err := u.roleRepo.GetByName(ctx, roleName)
	if err != nil {
		return entity.User{}, err
	}
	if role == nil {
		return entity.User{}, fmt.Errorf("role not found")
	}

	existingUser, err := u.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		return entity.User{}, err
	}
	if existingUser != nil {
		return entity.User{}, fmt.Errorf("email already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(req.Password),
		bcrypt.DefaultCost,
	)

	if err != nil {
		return entity.User{}, err
	}

	newUser := entity.User{
		FullName:     req.FullName,
		Email:        req.Email,
		HashPassword: string(hashedPassword),
		Role:         role.Name,
		IsBlocked:    false,
		IsVerified:   true,
	}

	// The admin-created user is written together with the user_roles row so the
	// cached role column and the join-table source of truth never drift apart.
	err = u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&newUser).Error; err != nil {
			return err
		}

		return syncPrimaryRoleTx(tx, newUser.ID, role.ID, role.Name)
	})
	if err != nil {
		return entity.User{}, err
	}

	return newUser, nil
}

// usecase/admin_usecase.go
func (u *adminUsecase) RemoveUser(ctx context.Context, adminID uint, targetID uint) error {
	if adminID == targetID {
		return errors.New("security risk: you cannot delete your own account")
	}

	target, err := u.repo.GetByID(ctx, targetID)
	if err == nil && target != nil && NormalizeRole(target.Role) == "admin" {
		return errors.New("cannot delete admin user")
	}

	return u.repo.DeleteUser(ctx, targetID)
}
