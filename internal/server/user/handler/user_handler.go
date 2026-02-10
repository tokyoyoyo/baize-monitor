package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"baize-monitor/internal/server/user/service"
	"baize-monitor/pkg/dto/request"
	"baize-monitor/pkg/dto/response"
	"baize-monitor/pkg/utils"
)

// UserHandler user handler
type UserHandler struct {
	userService service.UserService
	jwtManager  utils.JWTManager
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
		jwtManager:  utils.JWTManagerInstance,
	}
}

// Login user login
func (h *UserHandler) Login(c *gin.Context) {
	var req request.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Code:    400,
			Success: false,
			Message: "Invalid request parameters",
		})
		return
	}

	// Authenticate user
	user, err := h.userService.Authenticate(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Code:    401,
			Success: false,
			Message: "Username or password is incorrect",
		})
		return
	}

	// Get user permissions
	permissions, err := h.userService.GetUserPermissions(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Code:    500,
			Success: false,
			Message: "Failed to get user permissions",
		})
		return
	}

	// Generate access token and refresh token
	accessToken, err := h.jwtManager.GenerateAccessToken(user.ID, user.Username,
		user.IsAdmin, user.IsActive, user.IsDeleted,
		permissions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Code:    500,
			Success: false,
			Message: "Failed to generate access token",
		})
		return
	}

	refreshToken, err := h.jwtManager.GenerateRefreshToken(user.ID, user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Code:    500,
			Success: false,
			Message: "Failed to generate refresh token",
		})
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Code:    200,
		Success: true,
		Message: "Login successful",
		Data: response.LoginResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ID:           user.ID,
			Username:     user.Username,
			IsAdmin:      user.IsAdmin,
			Permissions:  permissions,
		},
	})
}

// RefreshToken refresh token
func (h *UserHandler) RefreshToken(c *gin.Context) {
	var req request.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Code:    400,
			Success: false,
			Message: "Invalid request parameters",
		})
		return
	}

	// Validate refresh token
	refreshClaims, err := h.jwtManager.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Code:    401,
			Success: false,
			Message: "Invalid refresh token",
		})
		return
	}

	// Get user info and permissions
	user, err := h.userService.GetUserByID(refreshClaims.UserID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Code:    401,
			Success: false,
			Message: "User not found",
		})
		return
	}

	permissions, err := h.userService.GetUserPermissions(refreshClaims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Code:    500,
			Success: false,
			Message: "Failed to get user permissions",
		})
		return
	}

	// Generate new access token
	accessToken, err := h.jwtManager.GenerateAccessToken(refreshClaims.UserID, refreshClaims.Username,
		user.IsAdmin, user.IsActive, user.IsDeleted,
		permissions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Code:    500,
			Success: false,
			Message: "Failed to generate access token",
		})
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Code:    200,
		Success: true,
		Message: "Token refreshed successfully",
		Data: map[string]string{
			"access_token": accessToken,
		},
	})
}

// Logout user logout
func (h *UserHandler) Logout(c *gin.Context) {
	// TODO: Implement token blacklist mechanism
	// Currently JWT is stateless, logout is mainly handled by the client
	c.JSON(http.StatusOK, response.SuccessResponse{
		Code:    200,
		Success: true,
		Message: "Logout successful",
	})
}

// CreateUser create user (admin only)
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req request.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Code:    400,
			Success: false,
			Message: "Invalid request parameters",
		})
		return
	}

	user, err := h.userService.CreateUser(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Code:    400,
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Code:    200,
		Success: true,
		Message: "User created successfully",
		Data: response.UserResponse{
			ID:          user.ID,
			Username:    user.Username,
			IsAdmin:     user.IsAdmin,
			IsActive:    user.IsActive,
			IsDeleted:   user.IsDeleted,
			CreatedAt:   user.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   user.UpdatedAt.Format(time.RFC3339),
			Permissions: user.Permissions,
		},
	})
}

// DeleteUser delete user (admin only)
func (h *UserHandler) DeleteUser(c *gin.Context) {
	var req request.DeleteUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Code:    400,
			Success: false,
			Message: "Invalid request parameters",
		})
		return
	}

	err := h.userService.DeleteUser(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Code:    400,
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Code:    200,
		Success: true,
		Message: "User deleted successfully",
	})
}

// UpdateUserStatus update user status (admin only)
func (h *UserHandler) UpdateUserStatus(c *gin.Context) {
	var req request.UpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Code:    400,
			Success: false,
			Message: "Invalid request parameters",
		})
		return
	}

	err := h.userService.UpdateUserStatus(req.UserID, req.IsActive)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Code:    400,
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Code:    200,
		Success: true,
		Message: "User status updated successfully",
	})
}

// GrantPermissions grant permissions (admin only)
func (h *UserHandler) GrantPermissions(c *gin.Context) {
	var req request.GrantPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Code:    400,
			Success: false,
			Message: "Invalid request parameters",
		})
		return
	}

	err := h.userService.GrantPermissions(req.UserID, req.Permissions)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Code:    400,
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Code:    200,
		Success: true,
		Message: "Permissions granted successfully",
	})
}

// RevokePermissions revoke permissions (admin only)
func (h *UserHandler) RevokePermissions(c *gin.Context) {
	var req request.RevokePermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Code:    400,
			Success: false,
			Message: "Invalid request parameters",
		})
		return
	}

	err := h.userService.RevokePermissions(req.UserID, req.PermissionKeys)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Code:    400,
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Code:    200,
		Success: true,
		Message: "Permissions revoked successfully",
	})
}

// ListUsers get user list (admin only)
func (h *UserHandler) ListUsers(c *gin.Context) {
	// Bind pagination parameters
	var req request.ListUsersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Code:    400,
			Success: false,
			Message: "Invalid pagination parameters",
		})
		return
	}

	users, total, err := h.userService.ListUsers(req.Page, req.PageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Code:    500,
			Success: false,
			Message: "Failed to get user list",
		})
		return
	}

	// Calculate total pages
	totalPages := int(total / int64(req.PageSize))
	if total%int64(req.PageSize) > 0 {
		totalPages++
	}

	// Convert to response format
	var userResponses []response.UserResponse
	for _, user := range users {
		userResponses = append(userResponses, response.UserResponse{
			ID:          user.ID,
			Username:    user.Username,
			IsAdmin:     user.IsAdmin,
			IsActive:    user.IsActive,
			IsDeleted:   user.IsDeleted,
			CreatedAt:   user.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   user.UpdatedAt.Format(time.RFC3339),
			Permissions: user.Permissions,
		})
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Code:    200,
		Success: true,
		Message: "User list retrieved successfully",
		Data: response.UsersListResponse{
			Users:      userResponses,
			Page:       req.Page,
			PageSize:   req.PageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}
