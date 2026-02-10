package util

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims JWT声明结构体
type JWTClaims struct {
	UserID      int64           `json:"user_id"`
	Username    string          `json:"username"`
	IsAdmin     bool            `json:"is_admin"`
	IsActive    bool            `json:"is_active"`
	IsDelete    bool            `json:"is_delete"`
	Permissions map[string]bool `json:"permissions"`
	TokenType   string          `json:"token_type,omitempty"` // "access" or "refresh"
	jwt.RegisteredClaims
}

// RefreshTokenClaims 刷新令牌声明结构体
type RefreshTokenClaims struct {
	UserID    int64  `json:"user_id"`
	Username  string `json:"username"`
	TokenType string `json:"token_type"` // always "refresh"
	jwt.RegisteredClaims
}

// JWTManager JWT管理器接口
type JWTManager interface {
	GenerateAccessToken(userID int64, username string,
		isAdmin, IsActive, IsDelete bool,
		permissions map[string]bool) (string, error)
	GenerateRefreshToken(userID int64, username string) (string, error)
	ValidateAccessToken(tokenString string) (*JWTClaims, error)
	ValidateRefreshToken(tokenString string) (*RefreshTokenClaims, error)
}

// jwtManager JWT管理器实现
type jwtManager struct {
	secretKey []byte
}

// NewJWTManager 创建新的JWT管理器
func NewJWTManager() JWTManager {
	secretKey := "baize_JWT_secret_key"
	baize_JWT_secret_key := os.Getenv("baize_JWT_secret_key")

	if baize_JWT_secret_key != "" {
		secretKey = baize_JWT_secret_key
	}

	return &jwtManager{
		secretKey: []byte(secretKey),
	}
}

// GenerateAccessToken 生成访问令牌
func (j *jwtManager) GenerateAccessToken(userID int64, username string,
	isAdmin, isActive, isDelete bool,
	permissions map[string]bool) (string, error) {
	claims := &JWTClaims{
		UserID:      userID,
		Username:    username,
		IsAdmin:     isAdmin,
		IsActive:    isActive,
		IsDelete:    isDelete,
		Permissions: permissions,
		TokenType:   "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * time.Minute)), // 30分钟过期
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

// GenerateRefreshToken 生成刷新令牌
func (j *jwtManager) GenerateRefreshToken(userID int64, username string) (string, error) {
	claims := &RefreshTokenClaims{
		UserID:    userID,
		Username:  username,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)), // 7天过期
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

// ValidateAccessToken 验证访问令牌
func (j *jwtManager) ValidateAccessToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return j.secretKey, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		// 验证是否是访问令牌
		if claims.TokenType != "access" {
			return nil, errors.New("invalid token type")
		}
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// ValidateRefreshToken 验证刷新令牌
func (j *jwtManager) ValidateRefreshToken(tokenString string) (*RefreshTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &RefreshTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return j.secretKey, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*RefreshTokenClaims); ok && token.Valid {
		// 验证是否是刷新令牌
		if claims.TokenType != "refresh" {
			return nil, errors.New("invalid token type")
		}
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
