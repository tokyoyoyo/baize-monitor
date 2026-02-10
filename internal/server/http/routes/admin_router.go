package routes

import (
	alert_handler "baize-monitor/internal/server/alert/handler"
	"baize-monitor/internal/server/http/middleware"
	user_handler "baize-monitor/internal/server/user/handler"
	"baize-monitor/pkg/constants"
	"fmt"

	"time"

	"github.com/gin-gonic/gin"
)

type AdminRouter interface {
	GetAdminRouter() *gin.Engine
}

type AdminRouterImpl struct {
	router *gin.Engine
}

func NewAdminRouter(bmcTrapParserH alert_handler.BMCTrapParserHandler, userH *user_handler.UserHandler) AdminRouter {
	router := gin.New()
	router.Use(middleware.RequestID())
	router.Use(middleware.Auth())

	ar := &AdminRouterImpl{router: router}
	router.POST(fmt.Sprintf("/%s", constants.PermissionHealthCheckPass), ar.healthCheck)

	// API v1 分组
	apiV1 := router.Group("/api/v1")

	bmcTrapParserRouter := apiV1.Group(fmt.Sprintf("/%s", constants.PermissionBMCTrapParserManagement))
	{
		bmcTrapParserRouter.POST("/add", bmcTrapParserH.Create)
		bmcTrapParserRouter.POST("/update", bmcTrapParserH.Update)
		bmcTrapParserRouter.POST("/delete", bmcTrapParserH.Delete)
		bmcTrapParserRouter.POST("/list", bmcTrapParserH.List)
		bmcTrapParserRouter.POST("/activate", bmcTrapParserH.Activate)
		bmcTrapParserRouter.POST("/deactivate", bmcTrapParserH.Deactivate)
	}

	users := apiV1.Group(fmt.Sprintf("/%s", constants.PermissionUserManagement))
	{
		users.POST("/create", userH.CreateUser)
		users.POST("/delete", userH.DeleteUser)
		users.POST("/update_status", userH.UpdateUserStatus)
		users.POST("/grant_permissions", userH.GrantPermissions)
		users.POST("/revoke_permissions", userH.RevokePermissions)
		users.POST("/list", userH.ListUsers)
	}

	auth := apiV1.Group(fmt.Sprintf("/%s", constants.PermissionUserAuthPass))
	{
		auth.POST("/login", userH.Login)
		auth.POST("/logout", userH.Logout)
		auth.POST("/refresh_token", userH.RefreshToken)
	}

	return ar
}

func (ar *AdminRouterImpl) GetAdminRouter() *gin.Engine {
	return ar.router
}

func (ar *AdminRouterImpl) healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
	})
}
