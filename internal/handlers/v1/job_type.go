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

	// Query task counts
	type TaskCount struct {
		JobTypeID int64
		Count     int64
	}
	var taskCounts []TaskCount
	h.db.Model(&models.TaskDaily{}).
		Select("job_type_id as job_type_id, count(*) as count").
		Where("job_type_id IN ? AND deleted_at IS NULL", jobTypeIDs).
		Group("job_type_id").
		Find(&taskCounts)

	// Create count map
	countMap := make(map[int64]int64)
	for _, tc := range taskCounts {
		countMap[tc.JobTypeID] = tc.Count
	}

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

	// Get task count
	var count int64
	h.db.Model(&models.TaskDaily{}).
		Where("job_type_id = ? AND deleted_at IS NULL", id).
		Count(&count)

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

	// Get task count
	var count int64
	h.db.Model(&models.TaskDaily{}).
		Where("job_type_id = ? AND deleted_at IS NULL", id).
		Count(&count)

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
