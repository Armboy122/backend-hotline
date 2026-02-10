package v1

import (
	"backend-hotlines3/internal/dto"
	"backend-hotlines3/internal/models"
	"log"
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

// List retrieves all operation centers.
func (h *OperationCenterHandler) List(c *gin.Context) {
	var operationCenters []models.OperationCenter
	if err := h.db.WithContext(c.Request.Context()).Find(&operationCenters).Error; err != nil {
		log.Printf("Failed to fetch operation centers: %v", err)
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "An error occurred while fetching operation centers",
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

// GetByID retrieves a specific operation center by ID.
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
	if err := h.db.WithContext(c.Request.Context()).First(&operationCenter, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, dto.StandardResponse{
				Success: false,
				Error: &dto.ErrorInfo{
					Code:    "NOT_FOUND",
					Message: "Operation center not found",
				},
			})
			return
		}
		log.Printf("Failed to fetch operation center %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "An error occurred while fetching the operation center",
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

// Create creates a new operation center with the provided name.
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
	if err := h.db.WithContext(c.Request.Context()).Create(&operationCenter).Error; err != nil {
		log.Printf("Failed to create operation center: %v", err)
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "An error occurred while creating the operation center",
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

// Update updates an existing operation center's name.
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
	if err := h.db.WithContext(c.Request.Context()).First(&operationCenter, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, dto.StandardResponse{
				Success: false,
				Error: &dto.ErrorInfo{
					Code:    "NOT_FOUND",
					Message: "Operation center not found",
				},
			})
			return
		}
		log.Printf("Failed to fetch operation center %d for update: %v", id, err)
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "An error occurred while fetching the operation center",
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
	if err := h.db.WithContext(c.Request.Context()).Save(&operationCenter).Error; err != nil {
		log.Printf("Failed to update operation center %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "An error occurred while updating the operation center",
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

// Delete removes an operation center by ID.
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

	result := h.db.WithContext(c.Request.Context()).Delete(&models.OperationCenter{}, id)
	if result.Error != nil {
		log.Printf("Failed to delete operation center %d: %v", id, result.Error)
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "An error occurred while deleting the operation center",
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
