package routes

import (
	"baize-monitor/internal/server/alert/handler"
	"baize-monitor/internal/server/http/middleware"
	"baize-monitor/pkg/constants"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

type AlertRouter interface {
	GetAlertRouter() *gin.Engine
}

type AlertRouterImpl struct {
	router *gin.Engine
}

func NewAlertRouter(bmcTraprH handler.BMCTrapHandler) AlertRouter {
	router := gin.New()
	router.Use(middleware.RequestID())
	router.Use(middleware.Auth())

	ar := &AlertRouterImpl{router: router}
	router.POST(fmt.Sprintf("/%s", constants.PermissionHealthCheckPass), ar.healthCheck)

	// API v1 分组
	apiV1 := router.Group(constants.APIV1Prefix)

	alerttManagement := apiV1.Group(fmt.Sprintf("/%s", constants.PermissionAlertManagement))
	{
		alerttManagement.POST("/update", bmcTraprH.Update)
	}

	AlertRead := apiV1.Group(fmt.Sprintf("/%s", constants.PermissionAlertRead))
	{
		AlertRead.POST("/list", bmcTraprH.List)
	}

	alertPass := apiV1.Group(fmt.Sprintf("/%s", constants.PermissionAlertPass))
	{
		alertPass.POST(constants.BMCAlertUploadEndpoint, bmcTraprH.ReceiveTrap)
	}

	return ar
}

func (ar *AlertRouterImpl) GetAlertRouter() *gin.Engine {
	return ar.router
}

func (ar *AlertRouterImpl) healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
	})
}
