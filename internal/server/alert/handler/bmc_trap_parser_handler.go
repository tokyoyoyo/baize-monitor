package handler

import (
	"baize-monitor/internal/server/alert/service"
	"baize-monitor/pkg/dto/request"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type BMCTrapParserHandler interface {
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	List(c *gin.Context)
	Activate(c *gin.Context)
	Deactivate(c *gin.Context)
}

type BMCTrapParserHandlerImpl struct {
	s service.BMCTrapParserService
}

func NewBMCTrapParserHandler(srv service.BMCTrapParserService) BMCTrapParserHandler {
	return &BMCTrapParserHandlerImpl{
		s: srv,
	}
}

// Create creates a BMC Trap parser
func (h *BMCTrapParserHandlerImpl) Create(c *gin.Context) {
	var req request.BMCTrapParserCreate

	// Bind JSON request body to struct
	err := c.BindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request parameters",
			"details": err.Error(),
		})
		return
	}

	// Call Service layer to handle business logic
	code, err := h.s.Create(&req)
	if err != nil {
		c.JSON(code, gin.H{
			"message": fmt.Sprintf("Creation failed: %v", err),
		})
		return
	}

	// Return success response
	c.JSON(http.StatusCreated, gin.H{
		"message": "Creation successful",
	})
}

func (h *BMCTrapParserHandlerImpl) Update(c *gin.Context) {
	var req request.BMCTrapParserUpdate

	// Bind JSON request body to struct
	err := c.BindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request parameters",
			"details": err.Error(),
		})
		return
	}

	code, err := h.s.Update(&req)
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

func (h *BMCTrapParserHandlerImpl) Delete(c *gin.Context) {
	var req request.BMCTrapParserUpdate

	// Bind JSON request body to struct
	err := c.BindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request parameters",
			"details": err.Error(),
		})
		return
	}

	code, err := h.s.Delete(req.ID)
	if err != nil {
		c.JSON(code, gin.H{
			"message": fmt.Sprintf("Deletion failed: %v", err),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Deletion successful",
	})

}

func (h *BMCTrapParserHandlerImpl) Activate(c *gin.Context) {
	var req request.BMCTrapParserUpdate
	err := c.BindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request parameters",
			"details": err.Error(),
		})
		return
	}

	code, err := h.s.Activate(req.ID)
	if err != nil {
		c.JSON(code, gin.H{
			"message": fmt.Sprintf("Activation failed: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Activation successful",
	})
}

func (h *BMCTrapParserHandlerImpl) Deactivate(c *gin.Context) {
	var req request.BMCTrapParserUpdate
	err := c.BindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request parameters",
			"details": err.Error(),
		})
		return
	}

	code, err := h.s.Deactivate(req.ID)
	if err != nil {
		c.JSON(code, gin.H{
			"message": fmt.Sprintf("Deactivation failed: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Deactivation successful",
	})
}

func (h *BMCTrapParserHandlerImpl) List(c *gin.Context) {
	var req request.BMCTrapParserFilter

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
			"message": fmt.Sprintf("Failed to get list: %v", err),
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "List retrieved successfully",
		"data":    resp,
	})
}
