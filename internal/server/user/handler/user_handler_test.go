package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"baize-monitor/internal/server/user/repository"
	"baize-monitor/internal/server/user/service"
	"baize-monitor/pkg/config"
	"baize-monitor/pkg/dto/request"
	"baize-monitor/pkg/dto/response"
	"baize-monitor/pkg/models"
	"baize-monitor/pkg/storage"
	"baize-monitor/pkg/util"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTest(t *testing.T) (service.UserService, repository.UserRepo, util.JWTManager) {
	testConf, err := config.LoadTestMockServerConfig()
	if err != nil {
		t.Fatalf("Failed to load test config: %v", err)
	}
	db, err := storage.NewClient(testConf.PostGresConfig)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// Clean and migrate User table
	err = db.DB.Migrator().DropTable(&models.User{})
	if err != nil {
		t.Fatalf("Failed to drop User table: %v", err)
	}
	err = db.AutoMigrate(&models.User{})
	if err != nil {
		t.Fatalf("Failed to migrate User table: %v", err)
	}

	userService := service.NewUserService(db.DB)
	userRepo := repository.NewUserRepo(db.DB)
	jwtManager := util.NewJWTManager()

	return userService, *userRepo, jwtManager
}

// createAdminUser creates a real admin user for testing
func createAdminUser(userRepo repository.UserRepo) *models.User {
	password := "adminpassword"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}

	adminUser := &models.User{
		Username:     "admin_test",
		PasswordHash: string(hashedPassword),
		IsAdmin:      true,
		IsActive:     true,
		IsDeleted:    false,
		Permissions:  map[string]bool{"*": true},
	}

	err = userRepo.Create(adminUser)
	if err != nil {
		panic(err)
	}
	return adminUser
}

// createRegularUser creates a regular user for testing
func createRegularUser(userService service.UserService, username, password string) *models.User {
	user, err := userService.CreateUser(username, password)
	if err != nil {
		panic(err)
	}
	return user
}

func TestLogin_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userService, _, jwtManager := setupTest(t)

	// Create a test user
	testUser, err := userService.CreateUser("testuser", "testpassword")
	assert.NoError(t, err)
	assert.NotNil(t, testUser)

	handler := NewUserHandler(userService, jwtManager)
	router := gin.Default()
	router.POST("/login", handler.Login)

	// Prepare login request
	loginReq := request.LoginRequest{
		Username: "testuser",
		Password: "testpassword",
	}
	jsonData, _ := json.Marshal(loginReq)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var successResp response.SuccessResponse
	err = json.Unmarshal(resp.Body.Bytes(), &successResp)
	fmt.Println(successResp)
	assert.NoError(t, err)
	assert.Equal(t, 200, successResp.Code)
	assert.True(t, successResp.Success)
	assert.Equal(t, "Login successful", successResp.Message)

	// Extract login response from Data field
	loginRespBytes, err := json.Marshal(successResp.Data)
	assert.NoError(t, err)

	var loginResp response.LoginResponse
	err = json.Unmarshal(loginRespBytes, &loginResp)
	assert.NoError(t, err)
	assert.NotEmpty(t, loginResp.AccessToken)
	assert.NotEmpty(t, loginResp.RefreshToken)
	assert.Equal(t, testUser.ID, loginResp.ID)
	assert.Equal(t, "testuser", loginResp.Username)
	assert.Equal(t, map[string]bool{}, loginResp.Permissions)
}

func TestLogin_InvalidCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userService, _, jwtManager := setupTest(t)

	handler := NewUserHandler(userService, jwtManager)
	router := gin.Default()
	router.POST("/login", handler.Login)

	// Prepare login request with wrong password
	loginReq := request.LoginRequest{
		Username: "nonexistent",
		Password: "wrongpassword",
	}
	jsonData, _ := json.Marshal(loginReq)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusUnauthorized, resp.Code)

	var errorResp response.ErrorResponse
	err := json.Unmarshal(resp.Body.Bytes(), &errorResp)
	assert.NoError(t, err)
	assert.Equal(t, 401, errorResp.Code)
	assert.Contains(t, errorResp.Message, "Username or password is incorrect")
}

func TestCreateUser_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userService, _, jwtManager := setupTest(t)

	handler := NewUserHandler(userService, jwtManager)
	router := gin.Default()
	router.Use(func(c *gin.Context) {
		c.Set("is_admin", true)
		c.Next()
	})
	router.POST("/users/create", handler.CreateUser)

	// Prepare create user request
	createReq := request.CreateUserRequest{
		Username: "newuser",
		Password: "newpassword123",
	}
	jsonData, _ := json.Marshal(createReq)

	req, _ := http.NewRequest("POST", "/users/create", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var successResp response.SuccessResponse
	err := json.Unmarshal(resp.Body.Bytes(), &successResp)
	assert.NoError(t, err)
	assert.Equal(t, 200, successResp.Code)
	assert.True(t, successResp.Success)
	assert.Contains(t, successResp.Message, "User created successfully")

	// Extract user data from response
	userDataBytes, err := json.Marshal(successResp.Data)
	assert.NoError(t, err)

	var userResp response.UserResponse
	err = json.Unmarshal(userDataBytes, &userResp)
	assert.NoError(t, err)
	assert.Equal(t, "newuser", userResp.Username)

	// Verify the user was actually created
	createdUser, err := userService.GetUserByID(userResp.ID)
	assert.NoError(t, err)
	assert.NotNil(t, createdUser)
	assert.Equal(t, "newuser", createdUser.Username)
}

func TestCreateUser_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userService, _, jwtManager := setupTest(t)

	handler := NewUserHandler(userService, jwtManager)
	router := gin.Default()
	router.Use(func(c *gin.Context) {
		c.Set("is_admin", false) // Non-admin user
		c.Next()
	})
	router.POST("/users/create", handler.CreateUser)

	// Prepare create user request
	createReq := request.CreateUserRequest{
		Username: "unauthorizeduser",
		Password: "password123",
	}
	jsonData, _ := json.Marshal(createReq)

	req, _ := http.NewRequest("POST", "/users/create", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusForbidden, resp.Code)

	var errorResp response.ErrorResponse
	err := json.Unmarshal(resp.Body.Bytes(), &errorResp)
	assert.NoError(t, err)
	assert.Equal(t, 403, errorResp.Code)
	assert.Contains(t, errorResp.Message, "Only administrators can create users")
}

func TestDeleteUser_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userService, _, jwtManager := setupTest(t)

	// Create user to delete
	userToDelete := createRegularUser(userService, "todelete", "deletepass")

	handler := NewUserHandler(userService, jwtManager)
	router := gin.Default()
	router.Use(func(c *gin.Context) {
		c.Set("is_admin", true)
		c.Next()
	})
	router.POST("/users/delete", handler.DeleteUser)

	// Prepare delete user request
	deleteReq := request.DeleteUserRequest{
		UserID: userToDelete.ID,
	}
	jsonData, _ := json.Marshal(deleteReq)

	req, _ := http.NewRequest("POST", "/users/delete", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var successResp response.SuccessResponse
	err := json.Unmarshal(resp.Body.Bytes(), &successResp)
	assert.NoError(t, err)
	assert.Equal(t, 200, successResp.Code)
	assert.True(t, successResp.Success)
	assert.Contains(t, successResp.Message, "User deleted successfully")

	// Verify the user was actually deleted
	_, err = userService.GetUserByID(userToDelete.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}

func TestDeleteUser_AdminNotAllowed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userService, userRepo, jwtManager := setupTest(t)

	// Create admin user (this will be the admin that we try to delete)
	adminUser := createAdminUser(userRepo)

	handler := NewUserHandler(userService, jwtManager)
	router := gin.Default()
	router.Use(func(c *gin.Context) {
		c.Set("is_admin", true)
		c.Next()
	})
	router.POST("/users/delete", handler.DeleteUser)

	// Try to delete admin user
	deleteReq := request.DeleteUserRequest{
		UserID: adminUser.ID,
	}
	jsonData, _ := json.Marshal(deleteReq)

	req, _ := http.NewRequest("POST", "/users/delete", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)

	var errorResp response.ErrorResponse
	err := json.Unmarshal(resp.Body.Bytes(), &errorResp)
	assert.NoError(t, err)
	assert.Equal(t, 400, errorResp.Code)
	assert.Contains(t, errorResp.Message, "cannot delete admin account")
}

func TestUpdateUserStatus_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userService, _, jwtManager := setupTest(t)

	// Create user to update
	userToUpdate := createRegularUser(userService, "toupdate", "updatepass")

	handler := NewUserHandler(userService, jwtManager)
	router := gin.Default()
	router.Use(func(c *gin.Context) {
		c.Set("is_admin", true)
		c.Next()
	})
	router.POST("/users/update_status", handler.UpdateUserStatus)

	// Prepare update status request (disable user)
	updateReq := request.UpdateUserStatusRequest{
		UserID:   userToUpdate.ID,
		IsActive: false,
	}
	jsonData, _ := json.Marshal(updateReq)

	req, _ := http.NewRequest("POST", "/users/update_status", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var successResp response.SuccessResponse
	err := json.Unmarshal(resp.Body.Bytes(), &successResp)
	assert.NoError(t, err)
	assert.Equal(t, 200, successResp.Code)
	assert.True(t, successResp.Success)
	assert.Contains(t, successResp.Message, "User status updated successfully")

	// Verify the user status was updated
	updatedUser, err := userService.GetUserByID(userToUpdate.ID)
	assert.NoError(t, err)
	assert.False(t, updatedUser.IsActive)
}

func TestListUsers_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userService, _, jwtManager := setupTest(t)

	// Create multiple regular users
	users := []struct {
		username string
		password string
	}{
		{"user1", "pass1"},
		{"user2", "pass2"},
		{"user3", "pass3"},
	}

	for _, u := range users {
		createRegularUser(userService, u.username, u.password)
	}

	handler := NewUserHandler(userService, jwtManager)
	router := gin.Default()
	router.Use(func(c *gin.Context) {
		c.Set("is_admin", true)
		c.Next()
	})
	router.GET("/users/list", handler.ListUsers)

	// Test pagination
	req, _ := http.NewRequest("GET", "/users/list?page=1&page_size=2", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var successResp response.SuccessResponse
	err := json.Unmarshal(resp.Body.Bytes(), &successResp)
	assert.NoError(t, err)
	assert.Equal(t, 200, successResp.Code)
	assert.True(t, successResp.Success)
	assert.Contains(t, successResp.Message, "User list retrieved successfully")

	// Extract list response from Data field
	listDataBytes, err := json.Marshal(successResp.Data)
	assert.NoError(t, err)

	var listResp response.UsersListResponse
	err = json.Unmarshal(listDataBytes, &listResp)
	assert.NoError(t, err)
	assert.Len(t, listResp.Users, 2) // Should have 2 users due to page_size=2
	assert.Equal(t, 1, listResp.Page)
	assert.Equal(t, 2, listResp.PageSize)
	assert.Equal(t, int64(3), listResp.Total) // 3 regular users
	assert.Equal(t, 2, listResp.TotalPages)

	// Verify user data structure
	for _, user := range listResp.Users {
		assert.NotZero(t, user.ID)
		assert.NotEmpty(t, user.Username)
		assert.False(t, user.IsDeleted)
	}
}

func TestListUsers_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userService, _, jwtManager := setupTest(t)

	handler := NewUserHandler(userService, jwtManager)
	router := gin.Default()
	router.Use(func(c *gin.Context) {
		c.Set("is_admin", false) // Non-admin user
		c.Next()
	})
	router.GET("/users/list", handler.ListUsers)

	req, _ := http.NewRequest("GET", "/users/list", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusForbidden, resp.Code)

	var errorResp response.ErrorResponse
	err := json.Unmarshal(resp.Body.Bytes(), &errorResp)
	assert.NoError(t, err)
	assert.Equal(t, 403, errorResp.Code)
	assert.Contains(t, errorResp.Message, "Only administrators can view user list")
}

func TestGrantPermissions_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userService, _, jwtManager := setupTest(t)

	// Create user to grant permissions
	userToGrant := createRegularUser(userService, "grantuser", "grantpass")

	handler := NewUserHandler(userService, jwtManager)
	router := gin.Default()
	router.Use(func(c *gin.Context) {
		c.Set("is_admin", true)
		c.Next()
	})
	router.POST("/users/grant_permissions", handler.GrantPermissions)

	// Prepare grant permissions request
	permissions := map[string]bool{
		"read":  true,
		"write": true,
	}
	grantReq := request.GrantPermissionsRequest{
		UserID:      userToGrant.ID,
		Permissions: permissions,
	}
	jsonData, _ := json.Marshal(grantReq)

	req, _ := http.NewRequest("POST", "/users/grant_permissions", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var successResp response.SuccessResponse
	err := json.Unmarshal(resp.Body.Bytes(), &successResp)
	assert.NoError(t, err)
	assert.Equal(t, 200, successResp.Code)
	assert.True(t, successResp.Success)
	assert.Contains(t, successResp.Message, "Permissions granted successfully")

	// Verify permissions were granted
	grantedPerms, err := userService.GetUserPermissions(userToGrant.ID)
	assert.NoError(t, err)
	assert.Equal(t, permissions, grantedPerms)
}

func TestRevokePermissions_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userService, _, jwtManager := setupTest(t)

	// Create user and grant permissions first
	userToRevoke := createRegularUser(userService, "revokeuser", "revokepass")
	permissions := map[string]bool{
		"read":   true,
		"write":  true,
		"delete": true,
	}
	err := userService.GrantPermissions(userToRevoke.ID, permissions)
	assert.NoError(t, err)

	handler := NewUserHandler(userService, jwtManager)
	router := gin.Default()
	router.Use(func(c *gin.Context) {
		c.Set("is_admin", true)
		c.Next()
	})
	router.POST("/users/revoke_permissions", handler.RevokePermissions)

	// Prepare revoke permissions request
	revokeReq := request.RevokePermissionsRequest{
		UserID:         userToRevoke.ID,
		PermissionKeys: []string{"write", "delete"},
	}
	jsonData, _ := json.Marshal(revokeReq)

	req, _ := http.NewRequest("POST", "/users/revoke_permissions", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var successResp response.SuccessResponse
	err = json.Unmarshal(resp.Body.Bytes(), &successResp)
	assert.NoError(t, err)
	assert.Equal(t, 200, successResp.Code)
	assert.True(t, successResp.Success)
	assert.Contains(t, successResp.Message, "Permissions revoked successfully")

	// Verify permissions were revoked
	remainingPerms, err := userService.GetUserPermissions(userToRevoke.ID)
	assert.NoError(t, err)
	assert.Equal(t, map[string]bool{"read": true}, remainingPerms)
}
