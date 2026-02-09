package v1

import (
	"backend-hotlines3/internal/dto"
	"backend-hotlines3/internal/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type JobTypeHandler struct {
	db *gorm.DB
}

func NewJobTypeHandler(db *gorm.DB) *JobTypeHandler {
	return &JobTypeHandler{db: db}
}

// List - GET /v1/job-types
func (h *JobTypeHandler) List(c *gin.Context) {
	var jobTypes []models.JobType
	if err := h.db.Find(&jobTypes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "DATABASE_ERROR",
				Message: err.Error(),
			},
		})
		return
	}

	// Get task counts for each job type
	var jobTypeIDs []int64
	for _, jt := range jobTypes {
		jobTypeIDs = append(jobTypeIDs, jt.ID)
	}
	countMap := models.CountTasksBy(h.db, models.TaskCol.JobTypeID, jobTypeIDs)

	// Build response
	var response []dto.JobTypeResponse
	for _, jt := range jobTypes {
		response = append(response, dto.JobTypeResponse{
			ID:   jt.ID,
			Name: jt.Name,
			Count: &dto.Count{
				Tasks: countMap[jt.ID],
			},
		})
	}

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data:    response,
	})
}

// GetByID - GET /v1/job-types/:id
func (h *JobTypeHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INVALID_ID",
				Message: "Invalid job type ID",
			},
		})
		return
	}

	var jobType models.JobType
	if err := h.db.First(&jobType, id).Error; err != nil {
		c.JSON(http.StatusNotFound, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "NOT_FOUND",
				Message: "Job type not found",
			},
		})
		return
	}

	count := models.CountTasksFor(h.db, models.TaskCol.JobTypeID, id)

	response := dto.JobTypeResponse{
		ID:   jobType.ID,
		Name: jobType.Name,
		Count: &dto.Count{
			Tasks: count,
		},
	}

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data:    response,
	})
}

// Create - POST /v1/job-types
func (h *JobTypeHandler) Create(c *gin.Context) {
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

	jobType := models.JobType{Name: req.Name}
	if err := h.db.Create(&jobType).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "DATABASE_ERROR",
				Message: err.Error(),
			},
		})
		return
	}

	response := dto.JobTypeResponse{
		ID:   jobType.ID,
		Name: jobType.Name,
		Count: &dto.Count{
			Tasks: 0,
		},
	}

	c.JSON(http.StatusCreated, dto.StandardResponse{
		Success: true,
		Data:    response,
	})
}

// Update - PUT /v1/job-types/:id
func (h *JobTypeHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INVALID_ID",
				Message: "Invalid job type ID",
			},
		})
		return
	}

	var jobType models.JobType
	if err := h.db.First(&jobType, id).Error; err != nil {
		c.JSON(http.StatusNotFound, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "NOT_FOUND",
				Message: "Job type not found",
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

	jobType.Name = req.Name
	if err := h.db.Save(&jobType).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "DATABASE_ERROR",
				Message: err.Error(),
			},
		})
		return
	}

	count := models.CountTasksFor(h.db, models.TaskCol.JobTypeID, id)

	response := dto.JobTypeResponse{
		ID:   jobType.ID,
		Name: jobType.Name,
		Count: &dto.Count{
			Tasks: count,
		},
	}

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data:    response,
	})
}

// Delete - DELETE /v1/job-types/:id
func (h *JobTypeHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INVALID_ID",
				Message: "Invalid job type ID",
			},
		})
		return
	}

	result := h.db.Delete(&models.JobType{}, id)
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
				Message: "Job type not found",
			},
		})
		return
	}

	c.Status(http.StatusNoContent)
}
