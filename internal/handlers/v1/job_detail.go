package v1

import (
	"backend-hotlines3/internal/dto"
	"backend-hotlines3/internal/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type JobDetailHandler struct {
	db *gorm.DB
}

func NewJobDetailHandler(db *gorm.DB) *JobDetailHandler {
	return &JobDetailHandler{db: db}
}

// List - GET /v1/job-details
func (h *JobDetailHandler) List(c *gin.Context) {
	var jobDetails []models.JobDetail
	// Only get non-deleted records
	if err := h.db.Scopes(models.JobDetailNotDeleted).Find(&jobDetails).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "DATABASE_ERROR",
				Message: err.Error(),
			},
		})
		return
	}

	// Get task counts for each job detail
	var jobDetailIDs []int64
	for _, jd := range jobDetails {
		jobDetailIDs = append(jobDetailIDs, jd.ID)
	}

	countMap := models.CountTasksBy(h.db, models.TaskCol.JobDetailID, jobDetailIDs)

	// Build response
	var response []dto.JobDetailResponse
	for _, jd := range jobDetails {
		var deletedAt *string
		if jd.DeletedAt != nil {
			formatted := jd.DeletedAt.Format(time.RFC3339)
			deletedAt = &formatted
		}

		response = append(response, dto.JobDetailResponse{
			ID:        jd.ID,
			Name:      jd.Name,
			JobTypeID: jd.JobTypeID,
			CreatedAt: jd.CreatedAt.Format(time.RFC3339),
			UpdatedAt: jd.UpdatedAt.Format(time.RFC3339),
			DeletedAt: deletedAt,
			Count: &dto.Count{
				Tasks: countMap[jd.ID],
			},
		})
	}

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data:    response,
	})
}

// GetByID - GET /v1/job-details/:id
func (h *JobDetailHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INVALID_ID",
				Message: "Invalid job detail ID",
			},
		})
		return
	}

	var jobDetail models.JobDetail
	if err := h.db.First(&jobDetail, id).Error; err != nil {
		c.JSON(http.StatusNotFound, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "NOT_FOUND",
				Message: "Job detail not found",
			},
		})
		return
	}

	count := models.CountTasksFor(h.db, models.TaskCol.JobDetailID, id)

	var deletedAt *string
	if jobDetail.DeletedAt != nil {
		formatted := jobDetail.DeletedAt.Format(time.RFC3339)
		deletedAt = &formatted
	}

	response := dto.JobDetailResponse{
		ID:        jobDetail.ID,
		Name:      jobDetail.Name,
		JobTypeID: jobDetail.JobTypeID,
		CreatedAt: jobDetail.CreatedAt.Format(time.RFC3339),
		UpdatedAt: jobDetail.UpdatedAt.Format(time.RFC3339),
		DeletedAt: deletedAt,
		Count: &dto.Count{
			Tasks: count,
		},
	}

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data:    response,
	})
}

// Create - POST /v1/job-details
func (h *JobDetailHandler) Create(c *gin.Context) {
	var req struct {
		Name      string `json:"name" binding:"required"`
		JobTypeID *int64 `json:"jobTypeId"`
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

	now := time.Now()
	jobDetail := models.JobDetail{
		Name:      req.Name,
		JobTypeID: req.JobTypeID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := h.db.Create(&jobDetail).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "DATABASE_ERROR",
				Message: err.Error(),
			},
		})
		return
	}

	response := dto.JobDetailResponse{
		ID:        jobDetail.ID,
		Name:      jobDetail.Name,
		JobTypeID: jobDetail.JobTypeID,
		CreatedAt: jobDetail.CreatedAt.Format(time.RFC3339),
		UpdatedAt: jobDetail.UpdatedAt.Format(time.RFC3339),
		DeletedAt: nil,
		Count: &dto.Count{
			Tasks: 0,
		},
	}

	c.JSON(http.StatusCreated, dto.StandardResponse{
		Success: true,
		Data:    response,
	})
}

// Update - PUT /v1/job-details/:id
func (h *JobDetailHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INVALID_ID",
				Message: "Invalid job detail ID",
			},
		})
		return
	}

	var jobDetail models.JobDetail
	if err := h.db.First(&jobDetail, id).Error; err != nil {
		c.JSON(http.StatusNotFound, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "NOT_FOUND",
				Message: "Job detail not found",
			},
		})
		return
	}

	var req struct {
		Name      string `json:"name"`
		JobTypeID *int64 `json:"jobTypeId"`
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
		jobDetail.Name = req.Name
	}
	if req.JobTypeID != nil {
		jobDetail.JobTypeID = req.JobTypeID
	}
	jobDetail.UpdatedAt = time.Now()

	if err := h.db.Save(&jobDetail).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "DATABASE_ERROR",
				Message: err.Error(),
			},
		})
		return
	}

	count := models.CountTasksFor(h.db, models.TaskCol.JobDetailID, id)

	var deletedAt *string
	if jobDetail.DeletedAt != nil {
		formatted := jobDetail.DeletedAt.Format(time.RFC3339)
		deletedAt = &formatted
	}

	response := dto.JobDetailResponse{
		ID:        jobDetail.ID,
		Name:      jobDetail.Name,
		JobTypeID: jobDetail.JobTypeID,
		CreatedAt: jobDetail.CreatedAt.Format(time.RFC3339),
		UpdatedAt: jobDetail.UpdatedAt.Format(time.RFC3339),
		DeletedAt: deletedAt,
		Count: &dto.Count{
			Tasks: count,
		},
	}

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data:    response,
	})
}

// Delete - DELETE /v1/job-details/:id (Soft Delete)
func (h *JobDetailHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INVALID_ID",
				Message: "Invalid job detail ID",
			},
		})
		return
	}

	var jobDetail models.JobDetail
	if err := h.db.First(&jobDetail, id).Error; err != nil {
		c.JSON(http.StatusNotFound, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "NOT_FOUND",
				Message: "Job detail not found",
			},
		})
		return
	}

	// Soft delete
	now := time.Now()
	jobDetail.DeletedAt = &now
	if err := h.db.Save(&jobDetail).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "DATABASE_ERROR",
				Message: err.Error(),
			},
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// Restore - POST /v1/job-details/:id/restore
func (h *JobDetailHandler) Restore(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INVALID_ID",
				Message: "Invalid job detail ID",
			},
		})
		return
	}

	var jobDetail models.JobDetail
	// Find including soft deleted
	if err := h.db.Unscoped().First(&jobDetail, id).Error; err != nil {
		c.JSON(http.StatusNotFound, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "NOT_FOUND",
				Message: "Job detail not found",
			},
		})
		return
	}

	if jobDetail.DeletedAt == nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "NOT_DELETED",
				Message: "Job detail is not deleted",
			},
		})
		return
	}

	// Restore
	jobDetail.DeletedAt = nil
	jobDetail.UpdatedAt = time.Now()
	if err := h.db.Save(&jobDetail).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "DATABASE_ERROR",
				Message: err.Error(),
			},
		})
		return
	}

	count := models.CountTasksFor(h.db, models.TaskCol.JobDetailID, id)

	response := dto.JobDetailResponse{
		ID:        jobDetail.ID,
		Name:      jobDetail.Name,
		JobTypeID: jobDetail.JobTypeID,
		CreatedAt: jobDetail.CreatedAt.Format(time.RFC3339),
		UpdatedAt: jobDetail.UpdatedAt.Format(time.RFC3339),
		DeletedAt: nil,
		Count: &dto.Count{
			Tasks: count,
		},
	}

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data:    response,
	})
}
