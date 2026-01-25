package v1

import (
	"backend-hotlines3/internal/dto"
	"backend-hotlines3/internal/models"
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

// List - GET /v1/peas
func (h *PEAHandler) List(c *gin.Context) {
	var peas []models.PEA
	if err := h.db.Preload("OperationCenter").Find(&peas).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "DATABASE_ERROR",
				Message: err.Error(),
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

// GetByID - GET /v1/peas/:id
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
	if err := h.db.Preload("OperationCenter").First(&pea, id).Error; err != nil {
		c.JSON(http.StatusNotFound, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "NOT_FOUND",
				Message: "PEA not found",
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

// Create - POST /v1/peas
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
	if err := h.db.Create(&pea).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "DATABASE_ERROR",
				Message: err.Error(),
			},
		})
		return
	}

	// Reload with relations
	h.db.Preload("OperationCenter").First(&pea, pea.ID)

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

// BulkCreate - POST /v1/peas/bulk
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

	if err := h.db.Create(&peas).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "DATABASE_ERROR",
				Message: err.Error(),
			},
		})
		return
	}

	// Reload with relations
	var peaIDs []int64
	for _, p := range peas {
		peaIDs = append(peaIDs, p.ID)
	}
	h.db.Preload("OperationCenter").Where("id IN ?", peaIDs).Find(&peas)

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

// Update - PUT /v1/peas/:id
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
	if err := h.db.First(&pea, id).Error; err != nil {
		c.JSON(http.StatusNotFound, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "NOT_FOUND",
				Message: "PEA not found",
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

	if err := h.db.Save(&pea).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "DATABASE_ERROR",
				Message: err.Error(),
			},
		})
		return
	}

	// Reload with relations
	h.db.Preload("OperationCenter").First(&pea, pea.ID)

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

// Delete - DELETE /v1/peas/:id
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

	result := h.db.Delete(&models.PEA{}, id)
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
				Message: "PEA not found",
			},
		})
		return
	}

	c.Status(http.StatusNoContent)
}
