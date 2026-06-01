package usecase

import (
	"backend/internal/entity"
	"backend/internal/repository"
	"context"
	"errors"
	"fmt"
	"strings"
)

type PermissionUsecase struct {
	permissionRepo repository.PermissionRepository
	roleRepo       repository.RoleRepository
	userRepo       repository.UserRepository
}

func NewPermissionUsecase(
	permissionRepo repository.PermissionRepository,
	roleRepo repository.RoleRepository,
	userRepo repository.UserRepository,
) *PermissionUsecase {
	return &PermissionUsecase{
		permissionRepo: permissionRepo,
		roleRepo:       roleRepo,
		userRepo:       userRepo,
	}
}

func (u *PermissionUsecase) ListPermissions(ctx context.Context) ([]entity.Permission, error) {
	return u.permissionRepo.List(ctx)
}

func (u *PermissionUsecase) CreatePermission(ctx context.Context, permission *entity.Permission) error {
	if permission == nil {
		return errors.New("permission is required")
	}

	permission.Key = strings.ToLower(strings.TrimSpace(permission.Key))
	permission.Name = strings.TrimSpace(permission.Name)
	if permission.Key == "" || permission.Name == "" {
		return errors.New("permission key and name are required")
	}

	existing, err := u.permissionRepo.GetByKey(ctx, permission.Key)
	if err != nil {
		return err
	}
	if existing != nil {
		return fmt.Errorf("permission already exists")
	}

	return u.permissionRepo.Create(ctx, permission)
}

func (u *PermissionUsecase) UpdatePermission(ctx context.Context, id uint, permission *entity.Permission) error {
	if id == 0 {
		return errors.New("permission id is required")
	}
	if permission == nil {
		return errors.New("permission is required")
	}

	existing, err := u.permissionRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("permission not found")
	}

	if strings.TrimSpace(permission.Key) != "" {
		existing.Key = strings.ToLower(strings.TrimSpace(permission.Key))
	}
	if strings.TrimSpace(permission.Name) != "" {
		existing.Name = strings.TrimSpace(permission.Name)
	}
	if strings.TrimSpace(permission.Description) != "" {
		existing.Description = strings.TrimSpace(permission.Description)
	}

	return u.permissionRepo.Update(ctx, existing)
}

func (u *PermissionUsecase) DeletePermission(ctx context.Context, id uint) error {
	if id == 0 {
		return errors.New("permission id is required")
	}
	return u.permissionRepo.Delete(ctx, id)
}

func (u *PermissionUsecase) AssignPermissionToRole(ctx context.Context, roleID uint, permissionID uint) error {
	if roleID == 0 || permissionID == 0 {
		return errors.New("role and permission are required")
	}

	role, err := u.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return err
	}
	if role == nil {
		return fmt.Errorf("role not found")
	}

	permission, err := u.permissionRepo.GetByID(ctx, permissionID)
	if err != nil {
		return err
	}
	if permission == nil {
		return fmt.Errorf("permission not found")
	}

	return u.permissionRepo.AssignPermissionToRole(ctx, roleID, permissionID)
}

func (u *PermissionUsecase) RemovePermissionFromRole(ctx context.Context, roleID uint, permissionID uint) error {
	if roleID == 0 || permissionID == 0 {
		return errors.New("role and permission are required")
	}
	return u.permissionRepo.RemovePermissionFromRole(ctx, roleID, permissionID)
}

func (u *PermissionUsecase) GetPermissionsByRoleID(ctx context.Context, roleID uint) ([]entity.Permission, error) {
	return u.permissionRepo.GetByRoleID(ctx, roleID)
}

func (u *PermissionUsecase) GetUserPermissions(ctx context.Context, userID uint) ([]entity.Permission, error) {
	if userID == 0 {
		return nil, errors.New("user id is required")
	}

	roles, err := u.roleRepo.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}

	if len(roles) == 0 && u.userRepo != nil {
		user, err := u.userRepo.GetByID(ctx, userID)
		if err == nil && user != nil {
			if role, roleErr := u.roleRepo.GetByName(ctx, NormalizeRole(user.Role)); roleErr == nil && role != nil {
				roles = append(roles, *role)
			}
		}
	}

	if len(roles) == 0 {
		return []entity.Permission{}, nil
	}

	roleIDs := make([]uint, 0, len(roles))
	for _, role := range roles {
		if role.ID != 0 {
			roleIDs = append(roleIDs, role.ID)
		}
	}

	permissions, err := u.permissionRepo.GetByRoleIDs(ctx, roleIDs)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]struct{}, len(permissions))
	result := make([]entity.Permission, 0, len(permissions))
	for _, permission := range permissions {
		key := strings.ToLower(strings.TrimSpace(permission.Key))
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, permission)
	}

	return result, nil
}

func (u *PermissionUsecase) GetUserPermissionKeys(ctx context.Context, userID uint) ([]string, error) {
	permissions, err := u.GetUserPermissions(ctx, userID)
	if err != nil {
		return nil, err
	}

	keys := make([]string, 0, len(permissions))
	for _, permission := range permissions {
		if permission.Key != "" {
			keys = append(keys, strings.ToLower(strings.TrimSpace(permission.Key)))
		}
	}

	return keys, nil
}

func (u *PermissionUsecase) HasPermission(ctx context.Context, userID uint, permissionKey string) (bool, error) {
	if userID == 0 {
		return false, errors.New("user id is required")
	}

	user, err := u.userRepo.GetByID(ctx, userID)
	if err == nil && user != nil && NormalizeRole(user.Role) == "admin" {
		return true, nil
	}

	keys, err := u.GetUserPermissionKeys(ctx, userID)
	if err != nil {
		return false, err
	}

	needle := strings.ToLower(strings.TrimSpace(permissionKey))
	for _, key := range keys {
		if key == needle {
			return true, nil
		}
	}

	return false, nil
}
