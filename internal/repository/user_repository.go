package repository

import (
	"backend/internal/entity"
	"context"
	"fmt"

	"gorm.io/gorm"
)

type UserRepository interface {
	GetByID(ctx context.Context, id uint) (*entity.User, error)
	GetUsers(ctx context.Context, role string, search string) ([]entity.User, error)
	UpdateUserStatus(ctx context.Context, id uint, isBlocked bool) error
	UpdateUserRole(ctx context.Context, id uint, newRole string) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

//
// 🔍 Get user by ID
//
func (r *userRepository) GetByID(ctx context.Context, id uint) (*entity.User, error) {
	var user entity.User

	err := r.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	return &user, nil
}

//
// 📄 Get users with filters (role + search)
//
func (r *userRepository) GetUsers(ctx context.Context, role string, search string) ([]entity.User, error) {
	var users []entity.User

	query := r.db.WithContext(ctx).
		Model(&entity.User{}).
		Omit("password") // hide sensitive field

	// Filter by role
	if role != "" {
		query = query.Where("role = ?", role)
	}

	// Search by name or email (case-insensitive, DB-safe)
	if search != "" {
		searchTerm := "%" + search + "%"
		query = query.Where(
			"LOWER(name) LIKE LOWER(?) OR LOWER(email) LIKE LOWER(?)",
			searchTerm, searchTerm,
		)
	}

	// Order for consistency
	err := query.Order("id ASC").Find(&users).Error
	if err != nil {
		return nil, err
	}

	return users, nil
}

//
// 🔒 Block / Unblock user
//
func (r *userRepository) UpdateUserStatus(ctx context.Context, id uint, isBlocked bool) error {
	result := r.db.WithContext(ctx).
		Model(&entity.User{}).
		Where("id = ?", id).
		Update("is_blocked", isBlocked)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

//
// 🔄 Change user role
//
func (r *userRepository) UpdateUserRole(ctx context.Context, id uint, newRole string) error {
	result := r.db.WithContext(ctx).
		Model(&entity.User{}).
		Where("id = ?", id).
		Update("role", newRole)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}
