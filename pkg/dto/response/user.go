package response

// LoginResponse login response
type LoginResponse struct {
	AccessToken  string          `json:"access_token"`
	RefreshToken string          `json:"refresh_token"`
	ID           int64           `json:"id"`
	Username     string          `json:"username"`
	IsAdmin      bool            `json:"is_admin"`
	Permissions  map[string]bool `json:"permissions"`
}

// UserResponse user information response
type UserResponse struct {
	ID          int64           `json:"id"`
	Username    string          `json:"username"`
	IsAdmin     bool            `json:"is_admin"`
	IsActive    bool            `json:"is_active"`
	IsDeleted   bool            `json:"is_deleted"`
	CreatedAt   string          `json:"created_at"`
	UpdatedAt   string          `json:"updated_at"`
	Permissions map[string]bool `json:"permissions"`
}

// UsersListResponse user list response with pagination
type UsersListResponse struct {
	Users      []UserResponse `json:"users"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	Total      int64          `json:"total"`
	TotalPages int            `json:"total_pages"`
}

// SuccessResponse generic success response
type SuccessResponse struct {
	Code    int         `json:"code"`
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorResponse generic error response
type ErrorResponse struct {
	Code    int    `json:"code"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}
