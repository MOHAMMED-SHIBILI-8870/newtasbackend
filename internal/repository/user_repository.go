// ============================
// internal/repository/user_repository.go
// ============================

package repository

import (
	"backend/internal/entity"
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type UserRepository interface {
	GetByID(ctx context.Context, id uint) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	GetUsers(ctx context.Context, role string, search string) ([]entity.User, error)
	UpdateUserStatus(ctx context.Context, id uint, isBlocked bool) error
	UpdateUserRole(ctx context.Context, id uint, newRole string) error
	CreateUser(ctx context.Context, user *entity.User) error
	DeleteUser(id uint) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

// 🔍 Get user by ID
func (r *userRepository) GetByID(
	ctx context.Context,
	id uint,
) (*entity.User, error) {

	var user entity.User

	err := r.db.WithContext(ctx).
		First(&user, id).Error

	if err != nil {

		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}

		return nil, err
	}

	return &user, nil
}

// 🔍 Get user by email
func (r *userRepository) GetByEmail(
	ctx context.Context,
	email string,
) (*entity.User, error) {

	var user entity.User

	err := r.db.WithContext(ctx).
		Where("email = ?", email).
		First(&user).Error

	if err != nil {

		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}

		return nil, err
	}

	return &user, nil
}

// 📄 Get users with filters
func (r *userRepository) GetUsers(
	ctx context.Context,
	role string,
	search string,
) ([]entity.User, error) {

	var users []entity.User

	query := r.db.WithContext(ctx).
		Model(&entity.User{}).
		Omit("HashPassword")

	// Filter by role
	if role != "" {
		query = query.Where("role = ?", role)
	}

	// Search by fullname or email
	if search != "" {

		searchTerm := "%" + search + "%"

		query = query.Where(
			`LOWER(full_name) LIKE LOWER(?) 
			OR LOWER(email) LIKE LOWER(?)`,
			searchTerm,
			searchTerm,
		)
	}

	// Order results
	err := query.
		Order("id ASC").
		Find(&users).Error

	if err != nil {
		return nil, err
	}

	return users, nil
}

// 🔒 Block / unblock user
func (r *userRepository) UpdateUserStatus(
	ctx context.Context,
	id uint,
	isBlocked bool,
) error {

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

// 🔄 Change role
func (r *userRepository) UpdateUserRole(
	ctx context.Context,
	id uint,
	newRole string,
) error {

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

// ➕ Create user
func (r *userRepository) CreateUser(
	ctx context.Context,
	user *entity.User,
) error {

	return r.db.WithContext(ctx).
		Create(user).Error
}

// infrastructure/repository/admin_repo_impl.go

func (r *userRepository) DeleteUser(userID uint) error {
    // Start a transaction
    return r.db.Transaction(func(tx *gorm.DB) error {
        
        // 1. Delete associated refresh tokens first
        if err := tx.Where("user_id = ?", userID).Delete(&entity.RefreshToken{}).Error; err != nil {
            return err // Rollback if this fails
        }

        // 2. Now delete the user
        result := tx.Delete(&entity.User{}, userID)
        if result.Error != nil {
            return result.Error
        }

        if result.RowsAffected == 0 {
            return errors.New("user not found")
        }

        return nil // Commit the transaction
    })
}