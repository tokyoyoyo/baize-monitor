package handler

import (
	"baize-monitor/internal/server/alert/service"
	"baize-monitor/pkg/dto/request"
	"baize-monitor/pkg/models"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type BMCTrapHandler interface {
	ReceiveTrap(c *gin.Context)
	List(c *gin.Context)
	Update(c *gin.Context)
}

type BMCTrapHandlerImpl struct {
	s service.BMCTrapService
}

func NewBMCTrapHandler(BMCTrapService service.BMCTrapService) BMCTrapHandler {
	return &BMCTrapHandlerImpl{s: BMCTrapService}
}

func (h *BMCTrapHandlerImpl) ReceiveTrap(c *gin.Context) {
	var req models.TrapMessage

	// Bind JSON request body to struct
	err := c.BindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request parameters",
			"details": err.Error(),
		})
		return
	}

	if req.SourceType != models.TrapSourceTypeBMC {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request parameters",
			"details": "only BMC trap supported",
		})
		return
	}

	code, err := h.s.Process(&req)
	if err != nil {
		c.JSON(code, gin.H{
			"error":   "Invalid request parameters",
			"details": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{"status": "success"})
}

func (h *BMCTrapHandlerImpl) List(c *gin.Context) {
	var req request.AlertFilter
	err := c.BindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request parameters",
			"details": err.Error(),
		})
		return
	}

	resp, code, err := h.s.List(&req)
	if err != nil {
		c.JSON(code, gin.H{
			"message": fmt.Sprintf("failed to query Alerts list %v", err),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "List retrieved successfully",
		"data":    resp,
	})
}

func (h *BMCTrapHandlerImpl) Update(c *gin.Context) {
	var req request.AlertUpdate
	err := c.BindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request parameters",
			"details": err.Error(),
		})
		return
	}

	code, err := h.s.Update(req)
	if err != nil {
		c.JSON(code, gin.H{
			"message": fmt.Sprintf("Update failed: %v", err),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Update successful",
	})
}
