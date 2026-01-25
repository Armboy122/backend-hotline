package v1

import (
	"backend-hotlines3/internal/dto"
	"backend-hotlines3/internal/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type OperationCenterHandler struct {
	db *gorm.DB
}

func NewOperationCenterHandler(db *gorm.DB) *OperationCenterHandler {
	return &OperationCenterHandler{db: db}
}

// List - GET /v1/operation-centers
func (h *OperationCenterHandler) List(c *gin.Context) {
	var operationCenters []models.OperationCenter
	if err := h.db.Find(&operationCenters).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "DATABASE_ERROR",
				Message: err.Error(),
			},
		})
		return
	}

	var response []dto.OperationCenterResponse
	for _, oc := range operationCenters {
		response = append(response, dto.OperationCenterResponse{
			ID:   oc.ID,
			Name: oc.Name,
		})
	}

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data:    response,
	})
}

// GetByID - GET /v1/operation-centers/:id
func (h *OperationCenterHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INVALID_ID",
				Message: "Invalid operation center ID",
			},
		})
		return
	}

	var operationCenter models.OperationCenter
	if err := h.db.First(&operationCenter, id).Error; err != nil {
		c.JSON(http.StatusNotFound, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "NOT_FOUND",
				Message: "Operation center not found",
			},
		})
		return
	}

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data: dto.OperationCenterResponse{
			ID:   operationCenter.ID,
			Name: operationCenter.Name,
		},
	})
}

// Create - POST /v1/operation-centers
func (h *OperationCenterHandler) Create(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "VALIDATION_ERROR",
				Message: err.Error(),
			},
		})
		return
	}

	operationCenter := models.OperationCenter{Name: req.Name}
	if err := h.db.Create(&operationCenter).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "DATABASE_ERROR",
				Message: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusCreated, dto.StandardResponse{
		Success: true,
		Data: dto.OperationCenterResponse{
			ID:   operationCenter.ID,
			Name: operationCenter.Name,
		},
	})
}

// Update - PUT /v1/operation-centers/:id
func (h *OperationCenterHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INVALID_ID",
				Message: "Invalid operation center ID",
			},
		})
		return
	}

	var operationCenter models.OperationCenter
	if err := h.db.First(&operationCenter, id).Error; err != nil {
		c.JSON(http.StatusNotFound, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "NOT_FOUND",
				Message: "Operation center not found",
			},
		})
		return
	}

	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "VALIDATION_ERROR",
				Message: err.Error(),
			},
		})
		return
	}

	operationCenter.Name = req.Name
	if err := h.db.Save(&operationCenter).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "DATABASE_ERROR",
				Message: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data: dto.OperationCenterResponse{
			ID:   operationCenter.ID,
			Name: operationCenter.Name,
		},
	})
}

// Delete - DELETE /v1/operation-centers/:id
func (h *OperationCenterHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INVALID_ID",
				Message: "Invalid operation center ID",
			},
		})
		return
	}

	result := h.db.Delete(&models.OperationCenter{}, id)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "DATABASE_ERROR",
				Message: result.Error.Error(),
			},
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "NOT_FOUND",
				Message: "Operation center not found",
			},
		})
		return
	}

	c.Status(http.StatusNoContent)
}
