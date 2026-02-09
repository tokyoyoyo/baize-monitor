package repository

import (
	"baize-monitor/pkg/config"
	"baize-monitor/pkg/models"
	"baize-monitor/pkg/storage"

	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type UserRepositoryTestSuite struct {
	suite.Suite
	db   *storage.Client
	repo *UserRepo
}

// SetupSuite initializes test database and repository
func (suite *UserRepositoryTestSuite) SetupSuite() {
	db, err := getUserTestDB()
	if err != nil {
		suite.T().Fatalf("Failed to get test DB: %v", err)
	}

	// Clean and migrate User table
	db.DB.Migrator().DropTable(&models.User{})
	if err := db.AutoMigrate(&models.User{}); err != nil {
		suite.T().Fatalf("Failed to migrate User: %v", err)
	}

	suite.db = db
	suite.repo = NewUserRepo(db.DB)
}

// SetupTest cleans users table before each test
func (suite *UserRepositoryTestSuite) SetupTest() {
	suite.db.DB.Exec("DELETE FROM users")
}

// TearDownTest cleans users table after each test
func (suite *UserRepositoryTestSuite) TearDownTest() {
	suite.db.DB.Exec("DELETE FROM users")
}

// TearDownSuite closes database connection
func (suite *UserRepositoryTestSuite) TearDownSuite() {
	if suite.db != nil {
		sqlDB, err := suite.db.DB.DB()
		if err == nil {
			sqlDB.Close()
		}
	}
}

// Helper to get test database client
func getUserTestDB() (*storage.Client, error) {
	testConf, err := config.LoadTestMockServerConfig()
	if err != nil {
		return nil, err
	}
	return storage.NewClient(testConf.PostGresConfig)
}

// ==================== TEST CASES ====================

func (suite *UserRepositoryTestSuite) TestCreate_Success() {
	user := &models.User{
		Username:     "testuser",
		PasswordHash: "hashed_password",
		IsAdmin:      false,
		IsActive:     true,
		IsDeleted:    false,
		Permissions:  map[string]bool{"read": true},
	}

	err := suite.repo.Create(user)
	assert.NoError(suite.T(), err)
	assert.NotZero(suite.T(), user.ID)

	// Verify user exists in database
	var foundUser models.User
	err = suite.db.DB.First(&foundUser, user.ID).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "testuser", foundUser.Username)
	assert.Equal(suite.T(), "hashed_password", foundUser.PasswordHash)
	assert.Equal(suite.T(), false, foundUser.IsAdmin)
	assert.Equal(suite.T(), true, foundUser.IsActive)
	assert.Equal(suite.T(), false, foundUser.IsDeleted)
	assert.Equal(suite.T(), map[string]bool{"read": true}, foundUser.Permissions)
}

func (suite *UserRepositoryTestSuite) TestCreate_UsernameConflict() {
	// Create first user
	user1 := &models.User{
		Username:     "conflictuser",
		PasswordHash: "password1",
	}
	err := suite.repo.Create(user1)
	assert.NoError(suite.T(), err)

	// Try to create second user with same username
	user2 := &models.User{
		Username:     "conflictuser",
		PasswordHash: "password2",
	}
	err = suite.repo.Create(user2)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to create user")
}

func (suite *UserRepositoryTestSuite) TestFindByUsername_Success() {
	// Create test user
	user := &models.User{
		Username:     "findme",
		PasswordHash: "password",
		IsAdmin:      true,
		IsActive:     true,
		IsDeleted:    false,
		Permissions:  map[string]bool{"*": true},
	}
	err := suite.repo.Create(user)
	assert.NoError(suite.T(), err)

	// Find by username
	foundUser, err := suite.repo.FindByUsername("findme")
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), foundUser)
	assert.Equal(suite.T(), user.ID, foundUser.ID)
	assert.Equal(suite.T(), "findme", foundUser.Username)
	assert.Equal(suite.T(), true, foundUser.IsAdmin)
	assert.Equal(suite.T(), true, foundUser.IsActive)
	assert.Equal(suite.T(), false, foundUser.IsDeleted)
	assert.Equal(suite.T(), map[string]bool{"*": true}, foundUser.Permissions)
}

func (suite *UserRepositoryTestSuite) TestFindByUsername_NotFound() {
	// Try to find non-existent user
	foundUser, err := suite.repo.FindByUsername("nonexistent")
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), foundUser)
	assert.Contains(suite.T(), err.Error(), "user not found")
}

func (suite *UserRepositoryTestSuite) TestFindByUsername_DeletedUser() {
	// Create and delete a user
	user := &models.User{
		Username:     "deleteduser",
		PasswordHash: "password",
		IsDeleted:    true,
	}
	err := suite.repo.Create(user)
	assert.NoError(suite.T(), err)

	// Try to find deleted user
	foundUser, err := suite.repo.FindByUsername("deleteduser")
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), foundUser)
	assert.Contains(suite.T(), err.Error(), "user not found")
}

func (suite *UserRepositoryTestSuite) TestFindByID_Success() {
	// Create test user
	user := &models.User{
		Username:     "findbyid",
		PasswordHash: "password",
		IsAdmin:      false,
		IsActive:     true,
		IsDeleted:    false,
		Permissions:  map[string]bool{"read": true, "write": false},
	}
	err := suite.repo.Create(user)
	assert.NoError(suite.T(), err)

	// Find by ID
	foundUser, err := suite.repo.FindByID(user.ID)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), foundUser)
	assert.Equal(suite.T(), user.ID, foundUser.ID)
	assert.Equal(suite.T(), "findbyid", foundUser.Username)
	assert.Equal(suite.T(), false, foundUser.IsAdmin)
	assert.Equal(suite.T(), true, foundUser.IsActive)
	assert.Equal(suite.T(), false, foundUser.IsDeleted)
	assert.Equal(suite.T(), map[string]bool{"read": true, "write": false}, foundUser.Permissions)
}

func (suite *UserRepositoryTestSuite) TestFindByID_NotFound() {
	// Try to find non-existent user
	foundUser, err := suite.repo.FindByID(999999)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), foundUser)
	assert.Contains(suite.T(), err.Error(), "user not found")
}

func (suite *UserRepositoryTestSuite) TestFindByID_DeletedUser() {
	// Create and delete a user
	user := &models.User{
		Username:     "deletedbyid",
		PasswordHash: "password",
		IsDeleted:    true,
	}
	err := suite.repo.Create(user)
	assert.NoError(suite.T(), err)

	// Try to find deleted user by ID
	foundUser, err := suite.repo.FindByID(user.ID)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), foundUser)
	assert.Contains(suite.T(), err.Error(), "user not found")
}

func (suite *UserRepositoryTestSuite) TestUpdate_Success() {
	// Create test user
	user := &models.User{
		Username:     "updateuser",
		PasswordHash: "oldpassword",
		IsAdmin:      false,
		IsActive:     true,
		IsDeleted:    false,
		Permissions:  map[string]bool{"read": true},
	}
	err := suite.repo.Create(user)
	assert.NoError(suite.T(), err)

	// Update user
	user.Username = "updateduser"
	user.PasswordHash = "newpassword"
	user.IsAdmin = true
	user.IsActive = false
	user.Permissions = map[string]bool{"read": true, "write": true, "delete": true}

	err = suite.repo.Update(user)
	assert.NoError(suite.T(), err)

	// Verify update in database
	var updatedUser models.User
	err = suite.db.DB.First(&updatedUser, user.ID).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "updateduser", updatedUser.Username)
	assert.Equal(suite.T(), "newpassword", updatedUser.PasswordHash)
	assert.Equal(suite.T(), true, updatedUser.IsAdmin)
	assert.Equal(suite.T(), false, updatedUser.IsActive)
	assert.Equal(suite.T(), map[string]bool{"read": true, "write": true, "delete": true}, updatedUser.Permissions)
}

func (suite *UserRepositoryTestSuite) TestDelete_Success() {
	// Create test user
	user := &models.User{
		Username:     "deleteuser",
		PasswordHash: "password",
	}
	err := suite.repo.Create(user)
	assert.NoError(suite.T(), err)

	// Delete user
	err = suite.repo.Delete(user.ID)
	assert.NoError(suite.T(), err)

	// Verify user is soft-deleted (not actually removed from DB)
	var foundUser models.User
	err = suite.db.DB.Unscoped().First(&foundUser, user.ID).Error
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), foundUser.IsDeleted)

	// Verify user cannot be found by normal queries
	_, err = suite.repo.FindByID(user.ID)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "user not found")
}

func (suite *UserRepositoryTestSuite) TestDelete_NotFound() {
	// Try to delete non-existent user
	err := suite.repo.Delete(999999)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "user not found")
}

func (suite *UserRepositoryTestSuite) TestGetAllUsersWithPagination_Success() {
	// Create multiple users
	users := []*models.User{
		{Username: "user1", PasswordHash: "pass1", IsActive: true, IsDeleted: false},
		{Username: "user2", PasswordHash: "pass2", IsActive: true, IsDeleted: false},
		{Username: "user3", PasswordHash: "pass3", IsActive: false, IsDeleted: false},
		{Username: "user4", PasswordHash: "pass4", IsActive: true, IsDeleted: false},
		{Username: "deleteduser", PasswordHash: "pass5", IsDeleted: true}, // This should not be included
	}

	for _, user := range users {
		err := suite.repo.Create(user)
		assert.NoError(suite.T(), err)
	}

	// Test first page with page size 2
	paginatedUsers, total, err := suite.repo.ListUsers(1, 2)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), paginatedUsers, 2)
	assert.Equal(suite.T(), int64(4), total) // Should exclude deleted user

	// Verify users are correct and not deleted
	for _, user := range paginatedUsers {
		assert.False(suite.T(), user.IsDeleted)
	}

	// Test second page with page size 2
	paginatedUsers2, total2, err := suite.repo.ListUsers(2, 2)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), paginatedUsers2, 2)
	assert.Equal(suite.T(), int64(4), total2)

	// Test third page with page size 2 (should return empty since only 4 users)
	paginatedUsers3, total3, err := suite.repo.ListUsers(3, 2)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), paginatedUsers3, 0)
	assert.Equal(suite.T(), int64(4), total3)
}

func (suite *UserRepositoryTestSuite) TestListUsersWithPagination_Empty() {
	// Test pagination when no users exist
	paginatedUsers, total, err := suite.repo.ListUsers(1, 10)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), paginatedUsers, 0)
	assert.Equal(suite.T(), int64(0), total)
}

func (suite *UserRepositoryTestSuite) TestListUsersWithPagination_ExcludesDeleted() {
	// Create a deleted user only
	deletedUser := &models.User{
		Username:     "deletedonly",
		PasswordHash: "deletedpass",
		IsDeleted:    true,
	}
	err := suite.repo.Create(deletedUser)
	assert.NoError(suite.T(), err)

	// Get paginated users
	paginatedUsers, total, err := suite.repo.ListUsers(1, 10)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), paginatedUsers, 0) // Should exclude deleted user
	assert.Equal(suite.T(), int64(0), total)
}

func TestUserRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(UserRepositoryTestSuite))
}
