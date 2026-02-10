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

type FeederHandler struct {
	db *gorm.DB
}

func NewFeederHandler(db *gorm.DB) *FeederHandler {
	return &FeederHandler{db: db}
}

// List retrieves all feeders with their station and operation center information, and task counts.
func (h *FeederHandler) List(c *gin.Context) {
	// ใช้ Joins แทน Preload เพื่อลดจาก 3 queries เป็น 1 query
	type feederRow struct {
		ID              int64  `gorm:"column:id"`
		Code            string `gorm:"column:code"`
		StationID       int64  `gorm:"column:stationId"`
		StationName     string `gorm:"column:station_name"`
		StationCodeName string `gorm:"column:station_code_name"`
		OpCenterID      int64  `gorm:"column:op_center_id"`
		OpCenterName    string `gorm:"column:op_center_name"`
	}

	var rows []feederRow
	err := h.db.WithContext(c.Request.Context()).Table(`"Feeder"`).
		Select(`"Feeder"."id", "Feeder"."code", "Feeder"."stationId", "Station"."name" as station_name, "Station"."codeName" as station_code_name, "OperationCenter"."id" as op_center_id, "OperationCenter"."name" as op_center_name`).
		Joins(`LEFT JOIN "Station" ON "Station"."id" = "Feeder"."stationId"`).
		Joins(`LEFT JOIN "OperationCenter" ON "OperationCenter"."id" = "Station"."operationId"`).
		Find(&rows).Error

	if err != nil {
		log.Printf("Failed to fetch feeders: %v", err)
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "An error occurred while fetching feeders",
			},
		})
		return
	}

	// Get task counts for each feeder
	var feederIDs []int64
	for _, f := range rows {
		feederIDs = append(feederIDs, f.ID)
	}

	countMap := models.CountTasksBy(h.db, models.TaskCol.FeederID, feederIDs)

	// Build response
	var response []dto.FeederResponse
	for _, f := range rows {
		feederResp := dto.FeederResponse{
			ID:        f.ID,
			Code:      f.Code,
			StationID: f.StationID,
			Count: &dto.Count{
				Tasks: countMap[f.ID],
			},
		}

		if f.StationID != 0 {
			feederResp.Station = &dto.StationNested{
				ID:       f.StationID,
				Name:     f.StationName,
				CodeName: f.StationCodeName,
			}
			if f.OpCenterID != 0 {
				feederResp.Station.OperationCenter = &dto.OperationCenterNested{
					ID:   f.OpCenterID,
					Name: f.OpCenterName,
				}
			}
		}

		response = append(response, feederResp)
	}

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data:    response,
	})
}

// GetByID retrieves a specific feeder by ID with its station, operation center, and task count.
func (h *FeederHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INVALID_ID",
				Message: "Invalid feeder ID",
			},
		})
		return
	}

	var feeder models.Feeder
	if err := h.db.WithContext(c.Request.Context()).Preload("Station.OperationCenter").First(&feeder, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, dto.StandardResponse{
				Success: false,
				Error: &dto.ErrorInfo{
					Code:    "NOT_FOUND",
					Message: "Feeder not found",
				},
			})
			return
		}
		log.Printf("Failed to fetch feeder %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "An error occurred while fetching the feeder",
			},
		})
		return
	}

	count := models.CountTasksFor(h.db, models.TaskCol.FeederID, id)

	response := dto.FeederResponse{
		ID:        feeder.ID,
		Code:      feeder.Code,
		StationID: feeder.StationID,
		Count: &dto.Count{
			Tasks: count,
		},
	}

	if feeder.Station != nil {
		response.Station = &dto.StationNested{
			ID:       feeder.Station.ID,
			Name:     feeder.Station.Name,
			CodeName: feeder.Station.CodeName,
		}
		if feeder.Station.OperationCenter != nil {
			response.Station.OperationCenter = &dto.OperationCenterNested{
				ID:   feeder.Station.OperationCenter.ID,
				Name: feeder.Station.OperationCenter.Name,
			}
		}
	}

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data:    response,
	})
}

// Create creates a new feeder with the provided code and station ID.
func (h *FeederHandler) Create(c *gin.Context) {
	var req struct {
		Code      string `json:"code" binding:"required"`
		StationID int64  `json:"stationId" binding:"required"`
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

	feeder := models.Feeder{
		Code:      req.Code,
		StationID: req.StationID,
	}
	if err := h.db.WithContext(c.Request.Context()).Create(&feeder).Error; err != nil {
		log.Printf("Failed to create feeder: %v", err)
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "An error occurred while creating the feeder",
			},
		})
		return
	}

	// Reload with relations
	h.db.WithContext(c.Request.Context()).Preload("Station.OperationCenter").First(&feeder, feeder.ID)

	response := dto.FeederResponse{
		ID:        feeder.ID,
		Code:      feeder.Code,
		StationID: feeder.StationID,
		Count: &dto.Count{
			Tasks: 0,
		},
	}

	if feeder.Station != nil {
		response.Station = &dto.StationNested{
			ID:       feeder.Station.ID,
			Name:     feeder.Station.Name,
			CodeName: feeder.Station.CodeName,
		}
		if feeder.Station.OperationCenter != nil {
			response.Station.OperationCenter = &dto.OperationCenterNested{
				ID:   feeder.Station.OperationCenter.ID,
				Name: feeder.Station.OperationCenter.Name,
			}
		}
	}

	c.JSON(http.StatusCreated, dto.StandardResponse{
		Success: true,
		Data:    response,
	})
}

// Update updates an existing feeder's code and/or station ID.
func (h *FeederHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INVALID_ID",
				Message: "Invalid feeder ID",
			},
		})
		return
	}

	var feeder models.Feeder
	if err := h.db.WithContext(c.Request.Context()).First(&feeder, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, dto.StandardResponse{
				Success: false,
				Error: &dto.ErrorInfo{
					Code:    "NOT_FOUND",
					Message: "Feeder not found",
				},
			})
			return
		}
		log.Printf("Failed to fetch feeder %d for update: %v", id, err)
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "An error occurred while fetching the feeder",
			},
		})
		return
	}

	var req struct {
		Code      string `json:"code"`
		StationID int64  `json:"stationId"`
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

	if req.Code != "" {
		feeder.Code = req.Code
	}
	if req.StationID != 0 {
		feeder.StationID = req.StationID
	}

	if err := h.db.WithContext(c.Request.Context()).Save(&feeder).Error; err != nil {
		log.Printf("Failed to update feeder %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "An error occurred while updating the feeder",
			},
		})
		return
	}

	// Reload with relations
	h.db.WithContext(c.Request.Context()).Preload("Station.OperationCenter").First(&feeder, feeder.ID)

	count := models.CountTasksFor(h.db, models.TaskCol.FeederID, id)

	response := dto.FeederResponse{
		ID:        feeder.ID,
		Code:      feeder.Code,
		StationID: feeder.StationID,
		Count: &dto.Count{
			Tasks: count,
		},
	}

	if feeder.Station != nil {
		response.Station = &dto.StationNested{
			ID:       feeder.Station.ID,
			Name:     feeder.Station.Name,
			CodeName: feeder.Station.CodeName,
		}
		if feeder.Station.OperationCenter != nil {
			response.Station.OperationCenter = &dto.OperationCenterNested{
				ID:   feeder.Station.OperationCenter.ID,
				Name: feeder.Station.OperationCenter.Name,
			}
		}
	}

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data:    response,
	})
}

// Delete removes a feeder by ID.
func (h *FeederHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INVALID_ID",
				Message: "Invalid feeder ID",
			},
		})
		return
	}

	result := h.db.WithContext(c.Request.Context()).Delete(&models.Feeder{}, id)
	if result.Error != nil {
		log.Printf("Failed to delete feeder %d: %v", id, result.Error)
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "An error occurred while deleting the feeder",
			},
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "NOT_FOUND",
				Message: "Feeder not found",
			},
		})
		return
	}

	c.Status(http.StatusNoContent)
}
