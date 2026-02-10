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

type PEAHandler struct {
	db *gorm.DB
}

func NewPEAHandler(db *gorm.DB) *PEAHandler {
	return &PEAHandler{db: db}
}

// List retrieves all PEAs with their operation center information.
func (h *PEAHandler) List(c *gin.Context) {
	var peas []models.PEA
	if err := h.db.WithContext(c.Request.Context()).Preload("OperationCenter").Find(&peas).Error; err != nil {
		log.Printf("Failed to fetch PEAs: %v", err)
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "An error occurred while fetching PEAs",
			},
		})
		return
	}

	var response []dto.PEAResponse
	for _, p := range peas {
		peaResp := dto.PEAResponse{
			ID:          p.ID,
			Shortname:   p.Shortname,
			Fullname:    p.Fullname,
			OperationID: p.OperationID,
		}
		if p.OperationCenter != nil {
			peaResp.OperationCenter = &dto.OperationCenterNested{
				ID:   p.OperationCenter.ID,
				Name: p.OperationCenter.Name,
			}
		}
		response = append(response, peaResp)
	}

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data:    response,
	})
}

// GetByID retrieves a specific PEA by ID with its operation center information.
func (h *PEAHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INVALID_ID",
				Message: "Invalid PEA ID",
			},
		})
		return
	}

	var pea models.PEA
	if err := h.db.WithContext(c.Request.Context()).Preload("OperationCenter").First(&pea, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, dto.StandardResponse{
				Success: false,
				Error: &dto.ErrorInfo{
					Code:    "NOT_FOUND",
					Message: "PEA not found",
				},
			})
			return
		}
		log.Printf("Failed to fetch PEA %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "An error occurred while fetching the PEA",
			},
		})
		return
	}

	response := dto.PEAResponse{
		ID:          pea.ID,
		Shortname:   pea.Shortname,
		Fullname:    pea.Fullname,
		OperationID: pea.OperationID,
	}
	if pea.OperationCenter != nil {
		response.OperationCenter = &dto.OperationCenterNested{
			ID:   pea.OperationCenter.ID,
			Name: pea.OperationCenter.Name,
		}
	}

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data:    response,
	})
}

// Create creates a new PEA with the provided shortname, fullname, and operation ID.
func (h *PEAHandler) Create(c *gin.Context) {
	var req struct {
		Shortname   string `json:"shortname" binding:"required"`
		Fullname    string `json:"fullname" binding:"required"`
		OperationID int64  `json:"operationId" binding:"required"`
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

	pea := models.PEA{
		Shortname:   req.Shortname,
		Fullname:    req.Fullname,
		OperationID: req.OperationID,
	}
	if err := h.db.WithContext(c.Request.Context()).Create(&pea).Error; err != nil {
		log.Printf("Failed to create PEA: %v", err)
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "An error occurred while creating the PEA",
			},
		})
		return
	}

	// Reload with relations
	h.db.WithContext(c.Request.Context()).Preload("OperationCenter").First(&pea, pea.ID)

	response := dto.PEAResponse{
		ID:          pea.ID,
		Shortname:   pea.Shortname,
		Fullname:    pea.Fullname,
		OperationID: pea.OperationID,
	}
	if pea.OperationCenter != nil {
		response.OperationCenter = &dto.OperationCenterNested{
			ID:   pea.OperationCenter.ID,
			Name: pea.OperationCenter.Name,
		}
	}

	c.JSON(http.StatusCreated, dto.StandardResponse{
		Success: true,
		Data:    response,
	})
}

// BulkCreate creates multiple PEAs in a single request.
func (h *PEAHandler) BulkCreate(c *gin.Context) {
	var req []struct {
		Shortname   string `json:"shortname" binding:"required"`
		Fullname    string `json:"fullname" binding:"required"`
		OperationID int64  `json:"operationId" binding:"required"`
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

	if len(req) == 0 {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "VALIDATION_ERROR",
				Message: "At least one PEA is required",
			},
		})
		return
	}

	var peas []models.PEA
	for _, r := range req {
		peas = append(peas, models.PEA{
			Shortname:   r.Shortname,
			Fullname:    r.Fullname,
			OperationID: r.OperationID,
		})
	}

	if err := h.db.WithContext(c.Request.Context()).Create(&peas).Error; err != nil {
		log.Printf("Failed to bulk create PEAs: %v", err)
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "An error occurred while creating PEAs",
			},
		})
		return
	}

	// Reload with relations
	var peaIDs []int64
	for _, p := range peas {
		peaIDs = append(peaIDs, p.ID)
	}
	h.db.WithContext(c.Request.Context()).Preload("OperationCenter").Where("id IN ?", peaIDs).Find(&peas)

	var response []dto.PEAResponse
	for _, p := range peas {
		peaResp := dto.PEAResponse{
			ID:          p.ID,
			Shortname:   p.Shortname,
			Fullname:    p.Fullname,
			OperationID: p.OperationID,
		}
		if p.OperationCenter != nil {
			peaResp.OperationCenter = &dto.OperationCenterNested{
				ID:   p.OperationCenter.ID,
				Name: p.OperationCenter.Name,
			}
		}
		response = append(response, peaResp)
	}

	c.JSON(http.StatusCreated, dto.StandardResponse{
		Success: true,
		Data:    response,
	})
}

// Update updates an existing PEA's shortname, fullname, and/or operation ID.
func (h *PEAHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INVALID_ID",
				Message: "Invalid PEA ID",
			},
		})
		return
	}

	var pea models.PEA
	if err := h.db.WithContext(c.Request.Context()).First(&pea, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, dto.StandardResponse{
				Success: false,
				Error: &dto.ErrorInfo{
					Code:    "NOT_FOUND",
					Message: "PEA not found",
				},
			})
			return
		}
		log.Printf("Failed to fetch PEA %d for update: %v", id, err)
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "An error occurred while fetching the PEA",
			},
		})
		return
	}

	var req struct {
		Shortname   string `json:"shortname"`
		Fullname    string `json:"fullname"`
		OperationID int64  `json:"operationId"`
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

	if req.Shortname != "" {
		pea.Shortname = req.Shortname
	}
	if req.Fullname != "" {
		pea.Fullname = req.Fullname
	}
	if req.OperationID != 0 {
		pea.OperationID = req.OperationID
	}

	if err := h.db.WithContext(c.Request.Context()).Save(&pea).Error; err != nil {
		log.Printf("Failed to update PEA %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "An error occurred while updating the PEA",
			},
		})
		return
	}

	// Reload with relations
	h.db.WithContext(c.Request.Context()).Preload("OperationCenter").First(&pea, pea.ID)

	response := dto.PEAResponse{
		ID:          pea.ID,
		Shortname:   pea.Shortname,
		Fullname:    pea.Fullname,
		OperationID: pea.OperationID,
	}
	if pea.OperationCenter != nil {
		response.OperationCenter = &dto.OperationCenterNested{
			ID:   pea.OperationCenter.ID,
			Name: pea.OperationCenter.Name,
		}
	}

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data:    response,
	})
}

// Delete removes a PEA by ID.
func (h *PEAHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INVALID_ID",
				Message: "Invalid PEA ID",
			},
		})
		return
	}

	result := h.db.WithContext(c.Request.Context()).Delete(&models.PEA{}, id)
	if result.Error != nil {
		log.Printf("Failed to delete PEA %d: %v", id, result.Error)
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "An error occurred while deleting the PEA",
			},
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "NOT_FOUND",
				Message: "PEA not found",
			},
		})
		return
	}

	c.Status(http.StatusNoContent)
}
