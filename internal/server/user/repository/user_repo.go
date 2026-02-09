package repository

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"baize-monitor/pkg/models"
)

// UserRepo user repository
type UserRepo struct {
	db *gorm.DB
}

// NewUserRepo creates a new user repository
func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}

// FindByUsername finds a user by username
func (r *UserRepo) FindByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.db.Where("username = ? AND is_deleted = ?", username, false).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to query user: %w", err)
	}
	return &user, nil
}

// FindByID finds a user by ID
func (r *UserRepo) FindByID(userID int64) (*models.User, error) {
	var user models.User
	err := r.db.Where("id = ? AND is_deleted = ?", userID, false).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to query user: %w", err)
	}
	return &user, nil
}

// Create creates a user
func (r *UserRepo) Create(user *models.User) error {
	result := r.db.Create(user)
	if result.Error != nil {
		return fmt.Errorf("failed to create user: %w", result.Error)
	}
	return nil
}

// Update updates a user
func (r *UserRepo) Update(user *models.User) error {
	result := r.db.Save(user)
	if result.Error != nil {
		return fmt.Errorf("failed to update user: %w", result.Error)
	}
	return nil
}

// Delete deletes a user (soft delete)
func (r *UserRepo) Delete(userID int64) error {
	result := r.db.Model(&models.User{}).Where("id = ?", userID).Update("is_deleted", true)
	if result.Error != nil {
		return fmt.Errorf("failed to delete user: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

// ListUsers gets all users (excluding deleted ones)
func (r *UserRepo) ListUsers(page, pageSize int) ([]*models.User, int64, error) {
	var users []*models.User
	var total int64

	// Get total count
	if err := r.db.Model(&models.User{}).Where("is_deleted = ?", false).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get total user count: %w", err)
	}

	// Apply pagination
	offset := (page - 1) * pageSize
	if err := r.db.Where("is_deleted = ?", false).
		Offset(offset).
		Limit(pageSize).
		Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get paginated user list: %w", err)
	}

	return users, total, nil
}
