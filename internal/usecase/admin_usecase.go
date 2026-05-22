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

	"golang.org/x/crypto/bcrypt"
)

type AdminUsecase interface {
	FetchUsers(ctx context.Context, role string, search string) ([]entity.User, error)
	ToggleUserBlock(ctx context.Context, targetID uint) (string, bool, error)
	ChangeUserRole(ctx context.Context, targetID uint, newRole string) error
	CreateUserByAdmin(ctx context.Context, req entity.AdminCreateUserRequest) (entity.User, error)
	RemoveUser(ctx context.Context, adminID uint, targetID uint) error
}

type adminUsecase struct {
	repo repository.UserRepository
}

func NewAdminUsecase(r repository.UserRepository) AdminUsecase {
	return &adminUsecase{
		repo: r,
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

	err = u.repo.UpdateUserStatus(ctx, targetID, newStatus)
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
	normalized := NormalizeRole(newRole)
	if normalized != "user" && normalized != "admin" {
		return fmt.Errorf("invalid role: %s", newRole)
	}

	return u.repo.UpdateUserRole(ctx, targetID, normalized)
}

// ➕ Create new user by admin
func (u *adminUsecase) CreateUserByAdmin(
	ctx context.Context,
	req entity.AdminCreateUserRequest,
) (entity.User, error) {
	normalized := NormalizeRole(req.Role)
	if normalized != "user" && normalized != "admin" {
		return entity.User{}, fmt.Errorf("invalid role")
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
		Role:         normalized,
		IsBlocked:    false,
		IsVerified:   true,
	}

	// Save user
	err = u.repo.CreateUser(ctx, &newUser)
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

	return u.repo.DeleteUser(targetID)
}
