package v1

import (
	"backend-hotlines3/internal/dto"
	"backend-hotlines3/internal/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type StationHandler struct {
	db *gorm.DB
}

func NewStationHandler(db *gorm.DB) *StationHandler {
	return &StationHandler{db: db}
}

// List - GET /v1/stations
func (h *StationHandler) List(c *gin.Context) {
	var stations []models.Station
	if err := h.db.Preload("OperationCenter").Find(&stations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "DATABASE_ERROR",
				Message: err.Error(),
			},
		})
		return
	}

	var response []dto.StationResponse
	for _, s := range stations {
		stationResp := dto.StationResponse{
			ID:          s.ID,
			Name:        s.Name,
			CodeName:    s.CodeName,
			OperationID: s.OperationID,
		}
		if s.OperationCenter != nil {
			stationResp.OperationCenter = &dto.OperationCenterNested{
				ID:   s.OperationCenter.ID,
				Name: s.OperationCenter.Name,
			}
		}
		response = append(response, stationResp)
	}

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data:    response,
	})
}

// GetByID - GET /v1/stations/:id
func (h *StationHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INVALID_ID",
				Message: "Invalid station ID",
			},
		})
		return
	}

	var station models.Station
	if err := h.db.Preload("OperationCenter").First(&station, id).Error; err != nil {
		c.JSON(http.StatusNotFound, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "NOT_FOUND",
				Message: "Station not found",
			},
		})
		return
	}

	response := dto.StationResponse{
		ID:          station.ID,
		Name:        station.Name,
		CodeName:    station.CodeName,
		OperationID: station.OperationID,
	}
	if station.OperationCenter != nil {
		response.OperationCenter = &dto.OperationCenterNested{
			ID:   station.OperationCenter.ID,
			Name: station.OperationCenter.Name,
		}
	}

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data:    response,
	})
}

// Create - POST /v1/stations
func (h *StationHandler) Create(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		CodeName    string `json:"codeName" binding:"required"`
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

	station := models.Station{
		Name:        req.Name,
		CodeName:    req.CodeName,
		OperationID: req.OperationID,
	}
	if err := h.db.Create(&station).Error; err != nil {
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
	h.db.Preload("OperationCenter").First(&station, station.ID)

	response := dto.StationResponse{
		ID:          station.ID,
		Name:        station.Name,
		CodeName:    station.CodeName,
		OperationID: station.OperationID,
	}
	if station.OperationCenter != nil {
		response.OperationCenter = &dto.OperationCenterNested{
			ID:   station.OperationCenter.ID,
			Name: station.OperationCenter.Name,
		}
	}

	c.JSON(http.StatusCreated, dto.StandardResponse{
		Success: true,
		Data:    response,
	})
}

// Update - PUT /v1/stations/:id
func (h *StationHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INVALID_ID",
				Message: "Invalid station ID",
			},
		})
		return
	}

	var station models.Station
	if err := h.db.First(&station, id).Error; err != nil {
		c.JSON(http.StatusNotFound, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "NOT_FOUND",
				Message: "Station not found",
			},
		})
		return
	}

	var req struct {
		Name        string `json:"name"`
		CodeName    string `json:"codeName"`
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

	if req.Name != "" {
		station.Name = req.Name
	}
	if req.CodeName != "" {
		station.CodeName = req.CodeName
	}
	if req.OperationID != 0 {
		station.OperationID = req.OperationID
	}

	if err := h.db.Save(&station).Error; err != nil {
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
	h.db.Preload("OperationCenter").First(&station, station.ID)

	response := dto.StationResponse{
		ID:          station.ID,
		Name:        station.Name,
		CodeName:    station.CodeName,
		OperationID: station.OperationID,
	}
	if station.OperationCenter != nil {
		response.OperationCenter = &dto.OperationCenterNested{
			ID:   station.OperationCenter.ID,
			Name: station.OperationCenter.Name,
		}
	}

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data:    response,
	})
}

// Delete - DELETE /v1/stations/:id
func (h *StationHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INVALID_ID",
				Message: "Invalid station ID",
			},
		})
		return
	}

	result := h.db.Delete(&models.Station{}, id)
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
				Message: "Station not found",
			},
		})
		return
	}

	c.Status(http.StatusNoContent)
}
