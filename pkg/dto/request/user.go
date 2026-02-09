package request

// LoginRequest login request
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// CreateUserRequest create user request
type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// DeleteUserRequest delete user request
type DeleteUserRequest struct {
	UserID int64 `json:"user_id" binding:"required"`
}

// UpdateUserStatusRequest update user status request
type UpdateUserStatusRequest struct {
	UserID   int64 `json:"user_id" binding:"required"`
	IsActive bool  `json:"is_active"`
}

// GrantPermissionsRequest grant permissions request
type GrantPermissionsRequest struct {
	UserID      int64           `json:"user_id" binding:"required"`
	Permissions map[string]bool `json:"permissions" binding:"required"`
}

// RevokePermissionsRequest revoke permissions request
type RevokePermissionsRequest struct {
	UserID         int64    `json:"user_id" binding:"required"`
	PermissionKeys []string `json:"permission_keys" binding:"required"`
}

// RefreshTokenRequest refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// ListUsersRequest list users request with pagination
type ListUsersRequest struct {
	Page     int `form:"page" binding:"omitempty,min=1" default:"1"`
	PageSize int `form:"page_size" binding:"omitempty,min=1" default:"10"`
}