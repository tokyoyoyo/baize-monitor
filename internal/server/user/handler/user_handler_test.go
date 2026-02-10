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
	"baize-monitor/pkg/constants"
	"baize-monitor/pkg/dto/request"
	"baize-monitor/pkg/dto/response"
	"baize-monitor/pkg/models"
	"baize-monitor/pkg/storage"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTest(t *testing.T) (service.UserService, repository.UserRepo) {
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

	return userService, *userRepo
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
	userService, _ := setupTest(t)

	// Create a test user
	testUser, err := userService.CreateUser("testuser", "testpassword")
	assert.NoError(t, err)
	assert.NotNil(t, testUser)

	handler := NewUserHandler(userService)
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
	assert.Equal(t, constants.DefaultPermissions, loginResp.Permissions)
}

func TestLogin_InvalidCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userService, _ := setupTest(t)

	handler := NewUserHandler(userService)
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
	userService, _ := setupTest(t)

	handler := NewUserHandler(userService)
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

func TestDeleteUser_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userService, _ := setupTest(t)

	// Create user to delete
	userToDelete := createRegularUser(userService, "todelete", "deletepass")

	handler := NewUserHandler(userService)
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
	userService, userRepo := setupTest(t)

	// Create admin user (this will be the admin that we try to delete)
	adminUser := createAdminUser(userRepo)

	handler := NewUserHandler(userService)
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
	userService, _ := setupTest(t)

	// Create user to update
	userToUpdate := createRegularUser(userService, "toupdate", "updatepass")

	handler := NewUserHandler(userService)
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
	userService, _ := setupTest(t)

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

	handler := NewUserHandler(userService)
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

func TestListUsers_badRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userService, _ := setupTest(t)

	handler := NewUserHandler(userService)
	router := gin.Default()
	router.Use(func(c *gin.Context) {
		c.Set("is_admin", true)
		c.Next()
	})
	router.GET("/users/list", handler.ListUsers)

	// Test with invalid page_size=0
	req, _ := http.NewRequest("GET", "/users/list?page=1&page_size=0", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	// Should return 400 for invalid parameters
	assert.Equal(t, http.StatusBadRequest, resp.Code)

	var errorResp response.ErrorResponse
	err := json.Unmarshal(resp.Body.Bytes(), &errorResp)
	assert.NoError(t, err)
	assert.Equal(t, 400, errorResp.Code)
	assert.False(t, errorResp.Success)
	assert.Contains(t, errorResp.Message, "Invalid pagination parameters")
}

func TestGrantPermissions_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userService, _ := setupTest(t)

	// Create user to grant permissions
	userToGrant := createRegularUser(userService, "grantuser", "grantpass")

	handler := NewUserHandler(userService)
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

	// Expected permissions should include both default permissions and granted permissions
	expectedPerms := make(map[string]bool)
	// Add default permissions
	for k, v := range constants.DefaultPermissions {
		expectedPerms[k] = v
	}
	// Add granted permissions
	for k, v := range permissions {
		expectedPerms[k] = v
	}
	assert.Equal(t, expectedPerms, grantedPerms)
}

func TestRevokePermissions_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userService, _ := setupTest(t)

	// Create user and grant permissions first
	userToRevoke := createRegularUser(userService, "revokeuser", "revokepass")
	permissions := map[string]bool{
		"read":   true,
		"write":  true,
		"delete": true,
	}
	err := userService.GrantPermissions(userToRevoke.ID, permissions)
	assert.NoError(t, err)

	handler := NewUserHandler(userService)
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

	// Expected remaining permissions should include default permissions plus remaining custom permissions
	expectedPerms := make(map[string]bool)
	// Add default permissions
	for k, v := range constants.DefaultPermissions {
		expectedPerms[k] = v
	}
	// Add remaining custom permission ("read")
	expectedPerms["read"] = true
	assert.Equal(t, expectedPerms, remainingPerms)
}

// TestRefreshToken_Success tests successful token refresh
func TestRefreshToken_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userService, _ := setupTest(t)

	// Create a test user
	testUser, err := userService.CreateUser("refreshtest", "testpassword")
	assert.NoError(t, err)
	assert.NotNil(t, testUser)

	handler := NewUserHandler(userService)
	router := gin.Default()
	router.POST("/refresh", handler.RefreshToken)

	// First, get a refresh token by logging in
	loginReq := request.LoginRequest{
		Username: "refreshtest",
		Password: "testpassword",
	}
	loginData, _ := json.Marshal(loginReq)
	loginResp := httptest.NewRecorder()
	loginReqObj, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(loginData))
	loginReqObj.Header.Set("Content-Type", "application/json")
	
	loginRouter := gin.Default()
	loginRouter.POST("/login", handler.Login)
	loginRouter.ServeHTTP(loginResp, loginReqObj)
	
	assert.Equal(t, http.StatusOK, loginResp.Code)
	
	var loginSuccessResp response.SuccessResponse
	err = json.Unmarshal(loginResp.Body.Bytes(), &loginSuccessResp)
	assert.NoError(t, err)
	
	loginRespBytes, err := json.Marshal(loginSuccessResp.Data)
	assert.NoError(t, err)
	
	var loginRespData response.LoginResponse
	err = json.Unmarshal(loginRespBytes, &loginRespData)
	assert.NoError(t, err)
	assert.NotEmpty(t, loginRespData.RefreshToken)

	// Now test refresh token
	refreshReq := request.RefreshTokenRequest{
		RefreshToken: loginRespData.RefreshToken,
	}
	refreshData, _ := json.Marshal(refreshReq)

	req, _ := http.NewRequest("POST", "/refresh", bytes.NewBuffer(refreshData))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var successResp response.SuccessResponse
	err = json.Unmarshal(resp.Body.Bytes(), &successResp)
	assert.NoError(t, err)
	assert.Equal(t, 200, successResp.Code)
	assert.True(t, successResp.Success)
	assert.Contains(t, successResp.Message, "Token refreshed successfully")

	// Extract new access token from response
	refreshRespBytes, err := json.Marshal(successResp.Data)
	assert.NoError(t, err)

	var refreshRespData map[string]string
	err = json.Unmarshal(refreshRespBytes, &refreshRespData)
	assert.NoError(t, err)
	assert.NotEmpty(t, refreshRespData["access_token"])
}

// TestRefreshToken_InvalidToken tests refresh token with invalid token
func TestRefreshToken_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userService, _ := setupTest(t)

	handler := NewUserHandler(userService)
	router := gin.Default()
	router.POST("/refresh", handler.RefreshToken)

	// Test with invalid refresh token
	refreshReq := request.RefreshTokenRequest{
		RefreshToken: "invalid.token.here",
	}
	refreshData, _ := json.Marshal(refreshReq)

	req, _ := http.NewRequest("POST", "/refresh", bytes.NewBuffer(refreshData))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusUnauthorized, resp.Code)

	var errorResp response.ErrorResponse
	err := json.Unmarshal(resp.Body.Bytes(), &errorResp)
	assert.NoError(t, err)
	assert.Equal(t, 401, errorResp.Code)
	assert.Contains(t, errorResp.Message, "Invalid refresh token")
}

// TestLogout_Success tests logout functionality
func TestLogout_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userService, _ := setupTest(t)

	handler := NewUserHandler(userService)
	router := gin.Default()
	router.POST("/logout", handler.Logout)

	req, _ := http.NewRequest("POST", "/logout", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var successResp response.SuccessResponse
	err := json.Unmarshal(resp.Body.Bytes(), &successResp)
	assert.NoError(t, err)
	assert.Equal(t, 200, successResp.Code)
	assert.True(t, successResp.Success)
	assert.Contains(t, successResp.Message, "Logout successful")
}

// TestLogin_InvalidParameters tests login with invalid parameters
func TestLogin_InvalidParameters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userService, _ := setupTest(t)

	handler := NewUserHandler(userService)
	router := gin.Default()
	router.POST("/login", handler.Login)

	// Test with invalid JSON
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)

	var errorResp response.ErrorResponse
	err := json.Unmarshal(resp.Body.Bytes(), &errorResp)
	assert.NoError(t, err)
	assert.Equal(t, 400, errorResp.Code)
	assert.Contains(t, errorResp.Message, "Invalid request parameters")
}

// TestCreateUser_InvalidParameters tests create user with invalid parameters
func TestCreateUser_InvalidParameters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userService, _ := setupTest(t)

	handler := NewUserHandler(userService)
	router := gin.Default()
	router.Use(func(c *gin.Context) {
		c.Set("is_admin", true)
		c.Next()
	})
	router.POST("/users/create", handler.CreateUser)

	// Test with invalid JSON
	req, _ := http.NewRequest("POST", "/users/create", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)

	var errorResp response.ErrorResponse
	err := json.Unmarshal(resp.Body.Bytes(), &errorResp)
	assert.NoError(t, err)
	assert.Equal(t, 400, errorResp.Code)
	assert.Contains(t, errorResp.Message, "Invalid request parameters")
}

// TestCreateUser_EmptyUsername tests create user with empty username
func TestCreateUser_EmptyUsername(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userService, _ := setupTest(t)

	handler := NewUserHandler(userService)
	router := gin.Default()
	router.Use(func(c *gin.Context) {
		c.Set("is_admin", true)
		c.Next()
	})
	router.POST("/users/create", handler.CreateUser)

	createReq := request.CreateUserRequest{
		Username: "",
		Password: "password123",
	}
	jsonData, _ := json.Marshal(createReq)

	req, _ := http.NewRequest("POST", "/users/create", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

// TestDeleteUser_InvalidParameters tests delete user with invalid parameters
func TestDeleteUser_InvalidParameters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userService, _ := setupTest(t)

	handler := NewUserHandler(userService)
	router := gin.Default()
	router.Use(func(c *gin.Context) {
		c.Set("is_admin", true)
		c.Next()
	})
	router.POST("/users/delete", handler.DeleteUser)

	// Test with invalid JSON
	req, _ := http.NewRequest("POST", "/users/delete", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)

	var errorResp response.ErrorResponse
	err := json.Unmarshal(resp.Body.Bytes(), &errorResp)
	assert.NoError(t, err)
	assert.Equal(t, 400, errorResp.Code)
	assert.Contains(t, errorResp.Message, "Invalid request parameters")
}

// TestDeleteUser_NonExistentUser tests delete non-existent user
func TestDeleteUser_NonExistentUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userService, _ := setupTest(t)

	handler := NewUserHandler(userService)
	router := gin.Default()
	router.Use(func(c *gin.Context) {
		c.Set("is_admin", true)
		c.Next()
	})
	router.POST("/users/delete", handler.DeleteUser)

	deleteReq := request.DeleteUserRequest{
		UserID: 999999, // Non-existent user ID
	}
	jsonData, _ := json.Marshal(deleteReq)

	req, _ := http.NewRequest("POST", "/users/delete", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

// TestUpdateUserStatus_InvalidParameters tests update user status with invalid parameters
func TestUpdateUserStatus_InvalidParameters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userService, _ := setupTest(t)

	handler := NewUserHandler(userService)
	router := gin.Default()
	router.Use(func(c *gin.Context) {
		c.Set("is_admin", true)
		c.Next()
	})
	router.POST("/users/update_status", handler.UpdateUserStatus)

	// Test with invalid JSON
	req, _ := http.NewRequest("POST", "/users/update_status", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)

	var errorResp response.ErrorResponse
	err := json.Unmarshal(resp.Body.Bytes(), &errorResp)
	assert.NoError(t, err)
	assert.Equal(t, 400, errorResp.Code)
	assert.Contains(t, errorResp.Message, "Invalid request parameters")
}

// TestUpdateUserStatus_NonExistentUser tests update non-existent user status
func TestUpdateUserStatus_NonExistentUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userService, _ := setupTest(t)

	handler := NewUserHandler(userService)
	router := gin.Default()
	router.Use(func(c *gin.Context) {
		c.Set("is_admin", true)
		c.Next()
	})
	router.POST("/users/update_status", handler.UpdateUserStatus)

	updateReq := request.UpdateUserStatusRequest{
		UserID:   999999, // Non-existent user ID
		IsActive: false,
	}
	jsonData, _ := json.Marshal(updateReq)

	req, _ := http.NewRequest("POST", "/users/update_status", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

// TestGrantPermissions_InvalidParameters tests grant permissions with invalid parameters
func TestGrantPermissions_InvalidParameters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userService, _ := setupTest(t)

	handler := NewUserHandler(userService)
	router := gin.Default()
	router.Use(func(c *gin.Context) {
		c.Set("is_admin", true)
		c.Next()
	})
	router.POST("/users/grant_permissions", handler.GrantPermissions)

	// Test with invalid JSON
	req, _ := http.NewRequest("POST", "/users/grant_permissions", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)

	var errorResp response.ErrorResponse
	err := json.Unmarshal(resp.Body.Bytes(), &errorResp)
	assert.NoError(t, err)
	assert.Equal(t, 400, errorResp.Code)
	assert.Contains(t, errorResp.Message, "Invalid request parameters")
}

// TestGrantPermissions_NonExistentUser tests grant permissions to non-existent user
func TestGrantPermissions_NonExistentUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userService, _ := setupTest(t)

	handler := NewUserHandler(userService)
	router := gin.Default()
	router.Use(func(c *gin.Context) {
		c.Set("is_admin", true)
		c.Next()
	})
	router.POST("/users/grant_permissions", handler.GrantPermissions)

	grantReq := request.GrantPermissionsRequest{
		UserID: 999999, // Non-existent user ID
		Permissions: map[string]bool{
			"test": true,
		},
	}
	jsonData, _ := json.Marshal(grantReq)

	req, _ := http.NewRequest("POST", "/users/grant_permissions", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

// TestRevokePermissions_InvalidParameters tests revoke permissions with invalid parameters
func TestRevokePermissions_InvalidParameters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userService, _ := setupTest(t)

	handler := NewUserHandler(userService)
	router := gin.Default()
	router.Use(func(c *gin.Context) {
		c.Set("is_admin", true)
		c.Next()
	})
	router.POST("/users/revoke_permissions", handler.RevokePermissions)

	// Test with invalid JSON
	req, _ := http.NewRequest("POST", "/users/revoke_permissions", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)

	var errorResp response.ErrorResponse
	err := json.Unmarshal(resp.Body.Bytes(), &errorResp)
	assert.NoError(t, err)
	assert.Equal(t, 400, errorResp.Code)
	assert.Contains(t, errorResp.Message, "Invalid request parameters")
}

// TestRevokePermissions_NonExistentUser tests revoke permissions from non-existent user
func TestRevokePermissions_NonExistentUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userService, _ := setupTest(t)

	handler := NewUserHandler(userService)
	router := gin.Default()
	router.Use(func(c *gin.Context) {
		c.Set("is_admin", true)
		c.Next()
	})
	router.POST("/users/revoke_permissions", handler.RevokePermissions)

	revokeReq := request.RevokePermissionsRequest{
		UserID:         999999, // Non-existent user ID
		PermissionKeys: []string{"test"},
	}
	jsonData, _ := json.Marshal(revokeReq)

	req, _ := http.NewRequest("POST", "/users/revoke_permissions", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

// TestListUsers_InvalidPage tests list users with invalid page parameter
func TestListUsers_InvalidPage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userService, _ := setupTest(t)

	handler := NewUserHandler(userService)
	router := gin.Default()
	router.Use(func(c *gin.Context) {
		c.Set("is_admin", true)
		c.Next()
	})
	router.GET("/users/list", handler.ListUsers)

	// Test with invalid page=-1
	req, _ := http.NewRequest("GET", "/users/list?page=-1&page_size=10", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)

	var errorResp response.ErrorResponse
	err := json.Unmarshal(resp.Body.Bytes(), &errorResp)
	assert.NoError(t, err)
	assert.Equal(t, 400, errorResp.Code)
	assert.Contains(t, errorResp.Message, "Invalid pagination parameters")
}

// TestListUsers_ExceedMaxPageSize tests list users with page_size exceeding maximum
func TestListUsers_ExceedMaxPageSize(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userService, _ := setupTest(t)

	handler := NewUserHandler(userService)
	router := gin.Default()
	router.Use(func(c *gin.Context) {
		c.Set("is_admin", true)
		c.Next()
	})
	router.GET("/users/list", handler.ListUsers)

	// Test with page_size=101 (assuming max is 100)
	req, _ := http.NewRequest("GET", "/users/list?page=1&page_size=101", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)

	var errorResp response.ErrorResponse
	err := json.Unmarshal(resp.Body.Bytes(), &errorResp)
	assert.NoError(t, err)
	assert.Equal(t, 400, errorResp.Code)
	assert.Contains(t, errorResp.Message, "Invalid pagination parameters")
}


