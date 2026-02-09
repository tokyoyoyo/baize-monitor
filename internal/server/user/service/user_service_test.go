package service

import (
	"baize-monitor/internal/server/user/repository"
	"baize-monitor/pkg/config"
	"baize-monitor/pkg/models"
	storage "baize-monitor/pkg/storage/postgres"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"
)

type UserServiceTestSuite struct {
	suite.Suite
	db          *storage.Client
	userRepo    *repository.UserRepo
	userService UserService
}

// SetupSuite initializes test database and service
func (suite *UserServiceTestSuite) SetupSuite() {
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
	suite.userRepo = repository.NewUserRepo(db.DB)
	suite.userService = NewUserService(db.DB)
}

// SetupTest cleans users table before each test
func (suite *UserServiceTestSuite) SetupTest() {
	suite.db.DB.Exec("DELETE FROM users")
}

// TearDownTest cleans users table after each test
func (suite *UserServiceTestSuite) TearDownTest() {
	suite.db.DB.Exec("DELETE FROM users")
}

// TearDownSuite closes database connection
func (suite *UserServiceTestSuite) TearDownSuite() {
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

func (suite *UserServiceTestSuite) TestAuthenticate_Success() {
	// Create a test user with password
	password := "testpassword123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	assert.NoError(suite.T(), err)

	user := &models.User{
		Username:     "authuser",
		PasswordHash: string(hashedPassword),
		IsActive:     true,
		IsDeleted:    false,
		Permissions:  map[string]bool{"read": true},
	}
	err = suite.userRepo.Create(user)
	assert.NoError(suite.T(), err)

	// Authenticate with correct credentials
	authUser, err := suite.userService.Authenticate("authuser", "testpassword123")
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), authUser)
	assert.Equal(suite.T(), "authuser", authUser.Username)
	assert.Equal(suite.T(), map[string]bool{"read": true}, authUser.Permissions)
}

func (suite *UserServiceTestSuite) TestAuthenticate_WrongPassword() {
	// Create a test user
	password := "correctpassword"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	assert.NoError(suite.T(), err)

	user := &models.User{
		Username:     "wrongpassuser",
		PasswordHash: string(hashedPassword),
		IsActive:     true,
	}
	err = suite.userRepo.Create(user)
	assert.NoError(suite.T(), err)

	// Authenticate with wrong password
	authUser, err := suite.userService.Authenticate("wrongpassuser", "wrongpassword")
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), authUser)
	assert.Contains(suite.T(), err.Error(), "username or password is incorrect")
}

func (suite *UserServiceTestSuite) TestAuthenticate_UserNotFound() {
	// Try to authenticate non-existent user
	authUser, err := suite.userService.Authenticate("nonexistent", "password")
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), authUser)
	assert.Contains(suite.T(), err.Error(), "username or password is incorrect")
}

func (suite *UserServiceTestSuite) TestAuthenticate_InactiveUser() {
	// Create a user first
	user, err := suite.userService.CreateUser("inactiveuser", "inactivepassword")
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), user)

	// Disable the user using UpdateUserStatus
	err = suite.userService.UpdateUserStatus(user.ID, false)
	assert.NoError(suite.T(), err)

	// Try to authenticate inactive user
	authUser, err := suite.userService.Authenticate("inactiveuser", "inactivepassword")
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), authUser)
	assert.Contains(suite.T(), err.Error(), "user account has been disabled")
}

func (suite *UserServiceTestSuite) TestCreateUser_Success() {
	// Create a new user
	user, err := suite.userService.CreateUser("newuser", "newpassword123")
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), user)
	assert.Equal(suite.T(), "newuser", user.Username)
	assert.False(suite.T(), user.IsAdmin)
	assert.True(suite.T(), user.IsActive)
	assert.False(suite.T(), user.IsDeleted)
	assert.Equal(suite.T(), map[string]bool{}, user.Permissions)

	// Verify user exists in database
	var foundUser models.User
	err = suite.db.DB.First(&foundUser, user.ID).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "newuser", foundUser.Username)
}

func (suite *UserServiceTestSuite) TestCreateUser_UsernameExists() {
	// Create first user
	_, err := suite.userService.CreateUser("existinguser", "password1")
	assert.NoError(suite.T(), err)

	// Try to create user with same username
	user, err := suite.userService.CreateUser("existinguser", "password2")
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), user)
	assert.Contains(suite.T(), err.Error(), "username already exists")
}

func (suite *UserServiceTestSuite) TestCreateUser_PasswordHashing() {
	// Create a user and verify password is hashed
	user, err := suite.userService.CreateUser("hashuser", "securepassword")
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), user)

	// Verify password is not stored in plain text
	assert.NotEqual(suite.T(), "securepassword", user.PasswordHash)
	assert.True(suite.T(), len(user.PasswordHash) > 0)

	// Verify we can authenticate with the original password
	authUser, err := suite.userService.Authenticate("hashuser", "securepassword")
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), authUser)
}

func (suite *UserServiceTestSuite) TestDeleteUser_Success() {
	// Create a regular user
	user, err := suite.userService.CreateUser("todelete", "deletepassword")
	assert.NoError(suite.T(), err)

	// Delete the user
	err = suite.userService.DeleteUser(user.ID)
	assert.NoError(suite.T(), err)

	// Verify user cannot be found
	_, err = suite.userService.GetUserByID(user.ID)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "user not found")
}

func (suite *UserServiceTestSuite) TestDeleteUser_AdminNotAllowed() {
	// Create admin user
	adminUser := &models.User{
		Username:     "adminuser",
		PasswordHash: "hashedpassword",
		IsAdmin:      true,
		IsActive:     true,
		IsDeleted:    false,
		Permissions:  map[string]bool{"*": true},
	}
	err := suite.userRepo.Create(adminUser)
	assert.NoError(suite.T(), err)

	// Try to delete admin user
	err = suite.userService.DeleteUser(adminUser.ID)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "cannot delete admin account")
}

func (suite *UserServiceTestSuite) TestDeleteUser_NotFound() {
	// Try to delete non-existent user
	err := suite.userService.DeleteUser(999999)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "user not found")
}

func (suite *UserServiceTestSuite) TestUpdateUserStatus_Success() {
	// Create a regular user
	user, err := suite.userService.CreateUser("statususer", "statuspassword")
	assert.NoError(suite.T(), err)

	// Disable the user
	err = suite.userService.UpdateUserStatus(user.ID, false)
	assert.NoError(suite.T(), err)

	// Verify status is updated
	updatedUser, err := suite.userService.GetUserByID(user.ID)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), updatedUser.IsActive)

	// Enable the user again
	err = suite.userService.UpdateUserStatus(user.ID, true)
	assert.NoError(suite.T(), err)

	// Verify status is updated
	updatedUser, err = suite.userService.GetUserByID(user.ID)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), updatedUser.IsActive)
}

func (suite *UserServiceTestSuite) TestUpdateUserStatus_AdminNotAllowed() {
	// Create admin user
	adminUser := &models.User{
		Username:     "adminstatus",
		PasswordHash: "hashedpassword",
		IsAdmin:      true,
		IsActive:     true,
		IsDeleted:    false,
		Permissions:  map[string]bool{"*": true},
	}
	err := suite.userRepo.Create(adminUser)
	assert.NoError(suite.T(), err)

	// Try to update admin status
	err = suite.userService.UpdateUserStatus(adminUser.ID, false)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "cannot modify admin account status")
}

func (suite *UserServiceTestSuite) TestUpdateUserStatus_NotFound() {
	// Try to update non-existent user
	err := suite.userService.UpdateUserStatus(999999, false)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "user not found")
}

func (suite *UserServiceTestSuite) TestGetUserByID_Success() {
	// Create a user
	user, err := suite.userService.CreateUser("getbyid", "getpassword")
	assert.NoError(suite.T(), err)

	// Get user by ID
	foundUser, err := suite.userService.GetUserByID(user.ID)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), foundUser)
	assert.Equal(suite.T(), user.ID, foundUser.ID)
	assert.Equal(suite.T(), "getbyid", foundUser.Username)
}

func (suite *UserServiceTestSuite) TestGetUserByID_NotFound() {
	// Try to get non-existent user
	user, err := suite.userService.GetUserByID(999999)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), user)
	assert.Contains(suite.T(), err.Error(), "user not found")
}

func (suite *UserServiceTestSuite) TestListUsersWithPagination_Success() {
	// Create multiple users
	users := []struct {
		username string
		password string
	}{
		{"user1", "pass1"},
		{"user2", "pass2"},
		{"user3", "pass3"},
		{"user4", "pass4"},
		{"user5", "pass5"},
	}

	for _, u := range users {
		_, err := suite.userService.CreateUser(u.username, u.password)
		assert.NoError(suite.T(), err)
	}

	// Test first page with page size 2
	paginatedUsers, total, err := suite.userService.ListUsers(1, 2)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), paginatedUsers, 2)
	assert.Equal(suite.T(), int64(5), total)

	usernames := make(map[string]bool)
	for _, user := range paginatedUsers {
		usernames[user.Username] = true
	}
	assert.True(suite.T(), usernames["user1"] || usernames["user2"] || usernames["user3"] || usernames["user4"] || usernames["user5"])

	// Test second page with page size 2
	paginatedUsers2, total2, err := suite.userService.ListUsers(2, 2)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), paginatedUsers2, 2)
	assert.Equal(suite.T(), int64(5), total2)

	// Test third page with page size 2 (should have 1 user)
	paginatedUsers3, total3, err := suite.userService.ListUsers(3, 2)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), paginatedUsers3, 1)
	assert.Equal(suite.T(), int64(5), total3)

	// Test page beyond total pages (should return empty slice)
	paginatedUsers4, total4, err := suite.userService.ListUsers(10, 2)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), paginatedUsers4, 0)
	assert.Equal(suite.T(), int64(5), total4)
}

func (suite *UserServiceTestSuite) TestListUsersWithPagination_Empty() {
	// Test pagination when no users exist
	paginatedUsers, total, err := suite.userService.ListUsers(1, 10)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), paginatedUsers, 0)
	assert.Equal(suite.T(), int64(0), total)
}

func (suite *UserServiceTestSuite) TestListUsersWithPagination_ExcludesDeleted() {
	// Create regular users
	_, err := suite.userService.CreateUser("active1", "pass1")
	assert.NoError(suite.T(), err)
	_, err = suite.userService.CreateUser("active2", "pass2")
	assert.NoError(suite.T(), err)

	// Create a deleted user (should not be included)
	deletedUser := &models.User{
		Username:     "deleteduser",
		PasswordHash: "deletedpass",
		IsDeleted:    true,
	}
	err = suite.userRepo.Create(deletedUser)
	assert.NoError(suite.T(), err)

	// Get paginated users
	paginatedUsers, total, err := suite.userService.ListUsers(1, 10)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), paginatedUsers, 2) // Should exclude deleted user
	assert.Equal(suite.T(), int64(2), total)

	// Verify users are active and not deleted
	for _, user := range paginatedUsers {
		assert.False(suite.T(), user.IsDeleted)
	}
}

func (suite *UserServiceTestSuite) TestGrantPermissions_Success() {
	// Create a regular user
	user, err := suite.userService.CreateUser("grantuser", "grantpassword")
	assert.NoError(suite.T(), err)

	// Grant permissions
	permissions := map[string]bool{
		"read":  true,
		"write": true,
	}
	err = suite.userService.GrantPermissions(user.ID, permissions)
	assert.NoError(suite.T(), err)

	// Verify permissions are granted
	grantedUser, err := suite.userService.GetUserByID(user.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), map[string]bool{"read": true, "write": true}, grantedUser.Permissions)
}

func (suite *UserServiceTestSuite) TestGrantPermissions_AdminNotAllowed() {
	// Create admin user
	adminUser := &models.User{
		Username:     "admingrant",
		PasswordHash: "hashedpassword",
		IsAdmin:      true,
		IsActive:     true,
		IsDeleted:    false,
		Permissions:  map[string]bool{"*": true},
	}
	err := suite.userRepo.Create(adminUser)
	assert.NoError(suite.T(), err)

	// Try to grant permissions to admin
	permissions := map[string]bool{"newperm": true}
	err = suite.userService.GrantPermissions(adminUser.ID, permissions)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "admin account permissions cannot be modified")
}

func (suite *UserServiceTestSuite) TestGrantPermissions_NotFound() {
	// Try to grant permissions to non-existent user
	permissions := map[string]bool{"read": true}
	err := suite.userService.GrantPermissions(999999, permissions)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "user not found")
}

func (suite *UserServiceTestSuite) TestRevokePermissions_Success() {
	// Create a user with permissions
	user, err := suite.userService.CreateUser("revokeuser", "revokepassword")
	assert.NoError(suite.T(), err)

	// Grant some permissions first
	permissions := map[string]bool{
		"read":   true,
		"write":  true,
		"delete": true,
	}
	err = suite.userService.GrantPermissions(user.ID, permissions)
	assert.NoError(suite.T(), err)

	// Revoke some permissions
	err = suite.userService.RevokePermissions(user.ID, []string{"write", "delete"})
	assert.NoError(suite.T(), err)

	// Verify remaining permissions
	remainingUser, err := suite.userService.GetUserByID(user.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), map[string]bool{"read": true}, remainingUser.Permissions)
}

func (suite *UserServiceTestSuite) TestRevokePermissions_AdminNotAllowed() {
	// Create admin user
	adminUser := &models.User{
		Username:     "adminrevoke",
		PasswordHash: "hashedpassword",
		IsAdmin:      true,
		IsActive:     true,
		IsDeleted:    false,
		Permissions:  map[string]bool{"*": true},
	}
	err := suite.userRepo.Create(adminUser)
	assert.NoError(suite.T(), err)

	// Try to revoke permissions from admin
	err = suite.userService.RevokePermissions(adminUser.ID, []string{"someperm"})
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "admin account permissions cannot be modified")
}

func (suite *UserServiceTestSuite) TestRevokePermissions_NotFound() {
	// Try to revoke permissions from non-existent user
	err := suite.userService.RevokePermissions(999999, []string{"read"})
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "user not found")
}

func (suite *UserServiceTestSuite) TestRevokePermissions_NoPermissions() {
	// Create a user with no permissions
	user, err := suite.userService.CreateUser("nopermuser", "nopermpassword")
	assert.NoError(suite.T(), err)

	// Try to revoke permissions from user with no permissions
	err = suite.userService.RevokePermissions(user.ID, []string{"nonexistent"})
	assert.NoError(suite.T(), err) // Should succeed silently
}

func (suite *UserServiceTestSuite) TestGetUserPermissions_Success() {
	// Create a user with permissions
	user, err := suite.userService.CreateUser("permuser", "permpassword")
	assert.NoError(suite.T(), err)

	// Grant permissions
	permissions := map[string]bool{"read": true, "write": false}
	err = suite.userService.GrantPermissions(user.ID, permissions)
	assert.NoError(suite.T(), err)

	// Get user permissions
	userPerms, err := suite.userService.GetUserPermissions(user.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), map[string]bool{"read": true, "write": false}, userPerms)
}

func (suite *UserServiceTestSuite) TestGetUserPermissions_Admin() {
	// Create admin user
	adminUser := &models.User{
		Username:     "adminperms",
		PasswordHash: "hashedpassword",
		IsAdmin:      true,
		IsActive:     true,
		IsDeleted:    false,
		Permissions:  map[string]bool{"original": true},
	}
	err := suite.userRepo.Create(adminUser)
	assert.NoError(suite.T(), err)

	// Get admin permissions (should return all permissions)
	userPerms, err := suite.userService.GetUserPermissions(adminUser.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), map[string]bool{"*": true}, userPerms)
}

func (suite *UserServiceTestSuite) TestGetUserPermissions_NoPermissions() {
	// Create a user with no permissions
	user, err := suite.userService.CreateUser("emptypermuser", "emptypassword")
	assert.NoError(suite.T(), err)

	// Get user permissions
	userPerms, err := suite.userService.GetUserPermissions(user.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), map[string]bool{}, userPerms)
}

func (suite *UserServiceTestSuite) TestGetUserPermissions_NotFound() {
	// Try to get permissions for non-existent user
	userPerms, err := suite.userService.GetUserPermissions(999999)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), userPerms)
	assert.Contains(suite.T(), err.Error(), "user not found")
}

func TestUserServiceTestSuite(t *testing.T) {
	suite.Run(t, new(UserServiceTestSuite))
}
