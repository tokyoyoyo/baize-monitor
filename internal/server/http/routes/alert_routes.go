package routes

import (
	"baize-monitor/internal/server/alert/handler"
	"baize-monitor/pkg/constants"
	"time"

	"github.com/gin-gonic/gin"
)

type AlertRouter interface {
	GetAlertRouter() *gin.Engine
}

type AlertRouterImpl struct {
	router *gin.Engine
}

func NewAlertRouter(
	bmcTraprH handler.BMCTrapHandler,
) AlertRouter {
	router := gin.New()

	ar := &AlertRouterImpl{router: router}

	// API v1 分组
	apiV1 := router.Group(constants.APIV1Prefix)

	router.GET("/health", ar.healthCheck)

	apiV1.POST(constants.BMCAlertUploadEndpoint, bmcTraprH.ReceiveTrap)
	apiV1.POST("/update", bmcTraprH.Update)
	apiV1.POST("/list", bmcTraprH.List)

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
