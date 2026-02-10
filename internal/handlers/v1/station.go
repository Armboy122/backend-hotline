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

type StationHandler struct {
	db *gorm.DB
}

func NewStationHandler(db *gorm.DB) *StationHandler {
	return &StationHandler{db: db}
}

// List retrieves all stations with their operation center information.
func (h *StationHandler) List(c *gin.Context) {
	var stations []models.Station
	if err := h.db.WithContext(c.Request.Context()).Preload("OperationCenter").Find(&stations).Error; err != nil {
		log.Printf("Failed to fetch stations: %v", err)
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "An error occurred while fetching stations",
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

// GetByID retrieves a specific station by ID with its operation center information.
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
	if err := h.db.WithContext(c.Request.Context()).Preload("OperationCenter").First(&station, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, dto.StandardResponse{
				Success: false,
				Error: &dto.ErrorInfo{
					Code:    "NOT_FOUND",
					Message: "Station not found",
				},
			})
			return
		}
		log.Printf("Failed to fetch station %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "An error occurred while fetching the station",
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

// Create creates a new station with the provided name, code name, and operation ID.
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
	if err := h.db.WithContext(c.Request.Context()).Create(&station).Error; err != nil {
		log.Printf("Failed to create station: %v", err)
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "An error occurred while creating the station",
			},
		})
		return
	}

	// Reload with relations
	h.db.WithContext(c.Request.Context()).Preload("OperationCenter").First(&station, station.ID)

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

// Update updates an existing station's name, code name, and/or operation ID.
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
	if err := h.db.WithContext(c.Request.Context()).First(&station, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, dto.StandardResponse{
				Success: false,
				Error: &dto.ErrorInfo{
					Code:    "NOT_FOUND",
					Message: "Station not found",
				},
			})
			return
		}
		log.Printf("Failed to fetch station %d for update: %v", id, err)
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "An error occurred while fetching the station",
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

	if err := h.db.WithContext(c.Request.Context()).Save(&station).Error; err != nil {
		log.Printf("Failed to update station %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "An error occurred while updating the station",
			},
		})
		return
	}

	// Reload with relations
	h.db.WithContext(c.Request.Context()).Preload("OperationCenter").First(&station, station.ID)

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

// Delete removes a station by ID.
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

	result := h.db.WithContext(c.Request.Context()).Delete(&models.Station{}, id)
	if result.Error != nil {
		log.Printf("Failed to delete station %d: %v", id, result.Error)
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "An error occurred while deleting the station",
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
