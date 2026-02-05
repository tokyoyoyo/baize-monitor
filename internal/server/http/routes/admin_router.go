package routes

import (
	alert_handler "baize-monitor/internal/server/alert/handler"
	"time"

	"github.com/gin-gonic/gin"
)

type AdminRouter interface {
	GetAdminRouter() *gin.Engine
}

type AdminRouterImpl struct {
	router *gin.Engine
}

func NewAdminRouter(bmcTrapParserH alert_handler.BMCTrapParserHandler) AdminRouter {
	router := gin.New()

	ar := &AdminRouterImpl{router: router}

	// API v1 分组
	apiV1 := router.Group("/api/v1")

	router.GET("/health", ar.healthCheck)

	bmcTrapParserRouter := apiV1.Group("/bmc_trap_parsers")
	bmcTrapParserRouter.POST("/add", bmcTrapParserH.Create)
	bmcTrapParserRouter.POST("/update", bmcTrapParserH.Update)
	bmcTrapParserRouter.POST("/delete", bmcTrapParserH.Delete)
	bmcTrapParserRouter.POST("/list", bmcTrapParserH.List)
	bmcTrapParserRouter.POST("/activate", bmcTrapParserH.Activate)
	bmcTrapParserRouter.POST("/deactivate", bmcTrapParserH.Deactivate)

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
