package usecase

import (
	"backend/internal/entity"
	"backend/internal/repository"
	"context"
	"errors"
	"fmt"
	"strings"
)

type RoleUsecase struct {
	roleRepo repository.RoleRepository
	userRepo repository.UserRepository
}

func NewRoleUsecase(roleRepo repository.RoleRepository, userRepo repository.UserRepository) *RoleUsecase {
	return &RoleUsecase{
		roleRepo: roleRepo,
		userRepo: userRepo,
	}
}

func (u *RoleUsecase) ListRoles(ctx context.Context) ([]entity.Role, error) {
	return u.roleRepo.List(ctx)
}

func (u *RoleUsecase) CreateRole(ctx context.Context, role *entity.Role) error {
	if role == nil {
		return errors.New("role is required")
	}

	role.Name = NormalizeRole(role.Name)
	if !IsValidRole(role.Name) {
		return fmt.Errorf("invalid role")
	}

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
		normalized := NormalizeRole(role.Name)
		if !IsValidRole(normalized) {
			return fmt.Errorf("invalid role")
		}
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

	users, err := u.userRepo.GetUsers(ctx, role.Name, "")
	if err != nil {
		return err
	}
	if len(users) > 0 {
		return errors.New("role is assigned to users")
	}

	return u.roleRepo.Delete(ctx, id)
}

func (u *RoleUsecase) AssignRoleToUser(ctx context.Context, userID uint, roleID uint) error {
	if userID == 0 || roleID == 0 {
		return errors.New("user and role are required")
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

	if err := u.userRepo.UpdateUserRole(ctx, userID, role.Name); err != nil {
		return err
	}

	return u.roleRepo.AssignRoleToUser(ctx, userID, roleID)
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
