package service

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"baize-monitor/internal/server/user/repository"
	"baize-monitor/pkg/models"
)

type UserService interface {
	// User authentication
	Authenticate(username, password string) (*models.User, error)

	// User management
	CreateUser(username, password string) (*models.User, error)
	DeleteUser(userID int64) error
	UpdateUserStatus(userID int64, isActive bool) error
	GetUserByID(userID int64) (*models.User, error)
	ListUsers(page, pageSize int) ([]*models.User, int64, error)

	// Permission management
	GrantPermissions(userID int64, permissions map[string]bool) error
	RevokePermissions(userID int64, permissionKeys []string) error
	GetUserPermissions(userID int64) (map[string]bool, error)
}

// userServiceImpl user service implementation
type userServiceImpl struct {
	userRepo *repository.UserRepo
}

// NewUserService creates a new user service
func NewUserService(db *gorm.DB) UserService {
	return &userServiceImpl{
		userRepo: repository.NewUserRepo(db),
	}
}

// Authenticate user authentication
func (s *userServiceImpl) Authenticate(username, password string) (*models.User, error) {
	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("username or password is incorrect")
	}

	if !user.IsActive {
		return nil, fmt.Errorf("user account has been disabled")
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("username or password is incorrect")
	}

	return user, nil
}

// CreateUser create user
func (s *userServiceImpl) CreateUser(username, password string) (*models.User, error) {
	// Check if username already exists
	existingUser, _ := s.userRepo.FindByUsername(username)
	if existingUser != nil {
		return nil, fmt.Errorf("username already exists")
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("password hashing failed: %w", err)
	}

	user := &models.User{
		Username:     username,
		PasswordHash: string(hashedPassword),
		IsAdmin:      false,
		IsActive:     true,
		IsDeleted:    false,
		Permissions:  make(map[string]bool), // No permissions initially
	}

	err = s.userRepo.Create(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// DeleteUser delete user
func (s *userServiceImpl) DeleteUser(userID int64) error {
	// Check if it's an admin account
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	if user.IsAdmin {
		return fmt.Errorf("cannot delete admin account")
	}

	return s.userRepo.Delete(userID)
}

// UpdateUserStatus update user status
func (s *userServiceImpl) UpdateUserStatus(userID int64, isActive bool) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	if user.IsAdmin {
		return fmt.Errorf("cannot modify admin account status")
	}

	user.IsActive = isActive
	return s.userRepo.Update(user)
}

// GetUserByID get user by ID
func (s *userServiceImpl) GetUserByID(userID int64) (*models.User, error) {
	return s.userRepo.FindByID(userID)
}

// GetAllUsersWithPagination get all users with pagination
func (s *userServiceImpl) ListUsers(page, pageSize int) ([]*models.User, int64, error) {
	return s.userRepo.ListUsers(page, pageSize)
}

// GrantPermissions grant permissions
func (s *userServiceImpl) GrantPermissions(userID int64, permissions map[string]bool) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	if user.IsAdmin {
		return fmt.Errorf("admin account permissions cannot be modified")
	}

	// Merge permissions
	if user.Permissions == nil {
		user.Permissions = make(map[string]bool)
	}

	for key, value := range permissions {
		user.Permissions[key] = value
	}

	return s.userRepo.Update(user)
}

// RevokePermissions revoke permissions
func (s *userServiceImpl) RevokePermissions(userID int64, permissionKeys []string) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	if user.IsAdmin {
		return fmt.Errorf("admin account permissions cannot be modified")
	}

	if user.Permissions == nil {
		return nil // User has no permissions, nothing to revoke
	}

	for _, key := range permissionKeys {
		delete(user.Permissions, key)
	}

	return s.userRepo.Update(user)
}

// GetUserPermissions get user permissions
func (s *userServiceImpl) GetUserPermissions(userID int64) (map[string]bool, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	if user.IsAdmin {
		// Admin has all permissions
		return map[string]bool{"*": true}, nil
	}

	if user.Permissions == nil {
		return make(map[string]bool), nil
	}

	return user.Permissions, nil
}
