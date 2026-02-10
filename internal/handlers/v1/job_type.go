package v1

import (
	"log"
	"net/http"
	"strconv"

	"backend-hotlines3/internal/dto"
	"backend-hotlines3/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type JobTypeHandler struct {
	db *gorm.DB
}

func NewJobTypeHandler(db *gorm.DB) *JobTypeHandler {
	return &JobTypeHandler{db: db}
}

// List retrieves all job types with their task counts.
func (h *JobTypeHandler) List(c *gin.Context) {
	var jobTypes []models.JobType
	if err := h.db.WithContext(c.Request.Context()).Find(&jobTypes).Error; err != nil {
		log.Printf("Failed to fetch job types: %v", err)
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "An error occurred while fetching job types",
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

// GetByID retrieves a specific job type by ID with its task count.
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
	if err := h.db.WithContext(c.Request.Context()).First(&jobType, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, dto.StandardResponse{
				Success: false,
				Error: &dto.ErrorInfo{
					Code:    "NOT_FOUND",
					Message: "Job type not found",
				},
			})
			return
		}
		log.Printf("Failed to fetch job type %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "An error occurred while fetching the job type",
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

// Create creates a new job type with the provided name.
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
	if err := h.db.WithContext(c.Request.Context()).Create(&jobType).Error; err != nil {
		log.Printf("Failed to create job type: %v", err)
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "An error occurred while creating the job type",
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

// Update updates an existing job type's name.
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
	if err := h.db.WithContext(c.Request.Context()).First(&jobType, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, dto.StandardResponse{
				Success: false,
				Error: &dto.ErrorInfo{
					Code:    "NOT_FOUND",
					Message: "Job type not found",
				},
			})
			return
		}
		log.Printf("Failed to fetch job type %d for update: %v", id, err)
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "An error occurred while fetching the job type",
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
	if err := h.db.WithContext(c.Request.Context()).Save(&jobType).Error; err != nil {
		log.Printf("Failed to update job type %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "An error occurred while updating the job type",
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

// Delete removes a job type by ID.
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

	result := h.db.WithContext(c.Request.Context()).Delete(&models.JobType{}, id)
	if result.Error != nil {
		log.Printf("Failed to delete job type %d: %v", id, result.Error)
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "An error occurred while deleting the job type",
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
