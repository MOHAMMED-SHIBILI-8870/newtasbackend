package usecase

import (
	"backend/internal/entity"
	"backend/internal/repository"
	"context"
	"fmt"
)

type AdminUsecase interface {
	FetchUsers(ctx context.Context, role string, search string) ([]entity.User, error)
	ToggleUserBlock(ctx context.Context, targetID uint) (string, bool, error)
	ChangeUserRole(ctx context.Context, targetID uint, newRole string) error
}

type adminUsecase struct {
	repo repository.UserRepository
}

func NewAdminUsecase(r repository.UserRepository) AdminUsecase {
	return &adminUsecase{repo: r}
}

//
// 📄 Fetch users (with filters)
//
func (u *adminUsecase) FetchUsers(ctx context.Context, role string, search string) ([]entity.User, error) {
	return u.repo.GetUsers(ctx, role, search)
}

//
// 🔒 Toggle Block / Unblock user
//
func (u *adminUsecase) ToggleUserBlock(ctx context.Context, targetID uint) (string, bool, error) {
	user, err := u.repo.GetByID(ctx, targetID)
	if err != nil {
		return "", false, err
	}

	if user == nil {
		return "", false, fmt.Errorf("user not found")
	}

	if user.Role == "admin" {
		return "", false, fmt.Errorf("cannot block admin user")
	}

	newStatus := !user.IsBlocked

	err = u.repo.UpdateUserStatus(ctx, targetID, newStatus)
	if err != nil {
		return "", false, err
	}

	return user.FullName, newStatus, nil
}
//
// 🔄 Change user role
//
func (u *adminUsecase) ChangeUserRole(ctx context.Context, targetID uint, newRole string) error {

	validRoles := map[string]bool{
		"user": true,
		"guide": true,
		"manager": true,
		"admin": true,
	}

	if !validRoles[newRole] {
		return fmt.Errorf("invalid role: %s", newRole)
	}

	return u.repo.UpdateUserRole(ctx, targetID, newRole)
}