package v1

import (
	"backend-hotlines3/internal/dto"
	"backend-hotlines3/internal/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type TaskHandler struct {
	db *gorm.DB
}

func NewTaskHandler(db *gorm.DB) *TaskHandler {
	return &TaskHandler{db: db}
}

// convertTaskToResponse converts a TaskDaily model to TaskResponse DTO
func convertTaskToResponse(task *models.TaskDaily) dto.TaskResponse {
	response := dto.TaskResponse{
		ID:          task.ID,
		WorkDate:    task.WorkDate.Format("2006-01-02"),
		TeamID:      task.TeamID,
		JobTypeID:   task.JobTypeID,
		JobDetailID: task.JobDetailID,
		FeederID:    task.FeederID,
		NumPole:     task.NumPole,
		DeviceCode:  task.DeviceCode,
		Detail:      task.Detail,
		URLsBefore:  []string(task.URLsBefore),
		URLsAfter:   []string(task.URLsAfter),
		CreatedAt:   task.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   task.UpdatedAt.Format(time.RFC3339),
	}

	// Handle coordinates
	if task.Latitude != nil {
		lat, _ := task.Latitude.Float64()
		response.Latitude = &lat
	}
	if task.Longitude != nil {
		lng, _ := task.Longitude.Float64()
		response.Longitude = &lng
	}

	// Handle deleted_at
	if task.DeletedAt != nil {
		formatted := task.DeletedAt.Format(time.RFC3339)
		response.DeletedAt = &formatted
	}

	// Handle relations
	if task.Team != nil {
		response.Team = &dto.TeamNested{
			ID:   task.Team.ID,
			Name: task.Team.Name,
		}
	}

	if task.JobType != nil {
		response.JobType = &dto.JobTypeNested{
			ID:   task.JobType.ID,
			Name: task.JobType.Name,
		}
	}

	if task.JobDetail != nil {
		response.JobDetail = &dto.JobDetailNested{
			ID:   task.JobDetail.ID,
			Name: task.JobDetail.Name,
		}
	}

	if task.Feeder != nil {
		response.Feeder = &dto.FeederNestedForTask{
			ID:   task.Feeder.ID,
			Code: task.Feeder.Code,
		}
		if task.Feeder.Station != nil {
			response.Feeder.Station = &dto.StationNestedSimple{
				Name: task.Feeder.Station.Name,
			}
			if task.Feeder.Station.OperationCenter != nil {
				response.Feeder.Station.OperationCenter = &dto.OperationCenterNested{
					ID:   task.Feeder.Station.OperationCenter.ID,
					Name: task.Feeder.Station.OperationCenter.Name,
				}
			}
		}
	}

	return response
}

// List - GET /v1/tasks
func (h *TaskHandler) List(c *gin.Context) {
	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}
	offset := (page - 1) * limit

	// Build query
	query := h.db.Model(&models.TaskDaily{}).Where("deleted_at IS NULL")

	// Apply filters
	if workDate := c.Query("workDate"); workDate != "" {
		query = query.Where("work_date = ?", workDate)
	}
	if teamID := c.Query("teamId"); teamID != "" {
		query = query.Where("team_id = ?", teamID)
	}
	if jobTypeID := c.Query("jobTypeId"); jobTypeID != "" {
		query = query.Where("job_type_id = ?", jobTypeID)
	}
	if feederID := c.Query("feederId"); feederID != "" {
		query = query.Where("feeder_id = ?", feederID)
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Get tasks with pagination
	var tasks []models.TaskDaily
	if err := query.
		Preload("Team").
		Preload("JobType").
		Preload("JobDetail").
		Preload("Feeder.Station.OperationCenter").
		Order("work_date DESC, created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&tasks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "DATABASE_ERROR",
				Message: err.Error(),
			},
		})
		return
	}

	// Convert to response
	var response []dto.TaskResponse
	for _, task := range tasks {
		response = append(response, convertTaskToResponse(&task))
	}

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data:    response,
		Meta: &dto.Meta{
			Page:  page,
			Limit: limit,
			Total: total,
		},
	})
}

// GetByID - GET /v1/tasks/:id
func (h *TaskHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INVALID_ID",
				Message: "Invalid task ID",
			},
		})
		return
	}

	var task models.TaskDaily
	if err := h.db.
		Preload("Team").
		Preload("JobType").
		Preload("JobDetail").
		Preload("Feeder.Station.OperationCenter").
		First(&task, id).Error; err != nil {
		c.JSON(http.StatusNotFound, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "NOT_FOUND",
				Message: "Task not found",
			},
		})
		return
	}

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data:    convertTaskToResponse(&task),
	})
}

// Create - POST /v1/tasks
func (h *TaskHandler) Create(c *gin.Context) {
	var req dto.CreateTaskRequest
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

	// Parse work date
	workDate, err := time.Parse("2006-01-02", req.WorkDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INVALID_DATE",
				Message: "Invalid work date format. Use YYYY-MM-DD",
			},
		})
		return
	}

	now := time.Now()
	task := models.TaskDaily{
		WorkDate:    workDate,
		TeamID:      req.TeamID,
		JobTypeID:   req.JobTypeID,
		JobDetailID: req.JobDetailID,
		FeederID:    req.FeederID,
		NumPole:     req.NumPole,
		DeviceCode:  req.DeviceCode,
		Detail:      req.Detail,
		URLsBefore:  models.StringArray(req.URLsBefore),
		URLsAfter:   models.StringArray(req.URLsAfter),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Handle coordinates
	if req.Latitude != nil && req.Longitude != nil {
		lat := decimal.NewFromFloat(*req.Latitude)
		lng := decimal.NewFromFloat(*req.Longitude)
		task.Latitude = &lat
		task.Longitude = &lng
	}

	if err := h.db.Create(&task).Error; err != nil {
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
	h.db.
		Preload("Team").
		Preload("JobType").
		Preload("JobDetail").
		Preload("Feeder.Station.OperationCenter").
		First(&task, task.ID)

	c.JSON(http.StatusCreated, dto.StandardResponse{
		Success: true,
		Data:    convertTaskToResponse(&task),
	})
}

// Update - PUT /v1/tasks/:id
func (h *TaskHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INVALID_ID",
				Message: "Invalid task ID",
			},
		})
		return
	}

	var task models.TaskDaily
	if err := h.db.First(&task, id).Error; err != nil {
		c.JSON(http.StatusNotFound, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "NOT_FOUND",
				Message: "Task not found",
			},
		})
		return
	}

	var req dto.UpdateTaskRequest
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

	// Update fields if provided
	if req.WorkDate != nil {
		workDate, err := time.Parse("2006-01-02", *req.WorkDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, dto.StandardResponse{
				Success: false,
				Error: &dto.ErrorInfo{
					Code:    "INVALID_DATE",
					Message: "Invalid work date format. Use YYYY-MM-DD",
				},
			})
			return
		}
		task.WorkDate = workDate
	}
	if req.TeamID != nil {
		task.TeamID = *req.TeamID
	}
	if req.JobTypeID != nil {
		task.JobTypeID = *req.JobTypeID
	}
	if req.JobDetailID != nil {
		task.JobDetailID = *req.JobDetailID
	}
	if req.FeederID != nil {
		task.FeederID = req.FeederID
	}
	if req.NumPole != nil {
		task.NumPole = req.NumPole
	}
	if req.DeviceCode != nil {
		task.DeviceCode = req.DeviceCode
	}
	if req.Detail != nil {
		task.Detail = req.Detail
	}
	if req.URLsBefore != nil {
		task.URLsBefore = models.StringArray(req.URLsBefore)
	}
	if req.URLsAfter != nil {
		task.URLsAfter = models.StringArray(req.URLsAfter)
	}
	if req.Latitude != nil && req.Longitude != nil {
		lat := decimal.NewFromFloat(*req.Latitude)
		lng := decimal.NewFromFloat(*req.Longitude)
		task.Latitude = &lat
		task.Longitude = &lng
	}

	task.UpdatedAt = time.Now()

	if err := h.db.Save(&task).Error; err != nil {
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
	h.db.
		Preload("Team").
		Preload("JobType").
		Preload("JobDetail").
		Preload("Feeder.Station.OperationCenter").
		First(&task, task.ID)

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data:    convertTaskToResponse(&task),
	})
}

// Delete - DELETE /v1/tasks/:id (Soft Delete)
func (h *TaskHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INVALID_ID",
				Message: "Invalid task ID",
			},
		})
		return
	}

	var task models.TaskDaily
	if err := h.db.First(&task, id).Error; err != nil {
		c.JSON(http.StatusNotFound, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "NOT_FOUND",
				Message: "Task not found",
			},
		})
		return
	}

	// Soft delete
	now := time.Now()
	task.DeletedAt = &now
	if err := h.db.Save(&task).Error; err != nil {
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

// ListByFilter - GET /v1/tasks/by-filter
func (h *TaskHandler) ListByFilter(c *gin.Context) {
	year := c.Query("year")
	month := c.Query("month")

	if year == "" || month == "" {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "VALIDATION_ERROR",
				Message: "year and month are required",
			},
		})
		return
	}

	// Build query
	query := h.db.Model(&models.TaskDaily{}).
		Where("deleted_at IS NULL").
		Where("EXTRACT(YEAR FROM work_date) = ?", year).
		Where("EXTRACT(MONTH FROM work_date) = ?", month)

	if teamID := c.Query("teamId"); teamID != "" {
		query = query.Where("team_id = ?", teamID)
	}

	// Get tasks
	var tasks []models.TaskDaily
	if err := query.
		Preload("Team").
		Preload("JobType").
		Preload("JobDetail").
		Preload("Feeder.Station.OperationCenter").
		Order("work_date DESC, created_at DESC").
		Find(&tasks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "DATABASE_ERROR",
				Message: err.Error(),
			},
		})
		return
	}

	// Group by team
	teamMap := make(map[string]dto.TasksByTeamResponse)
	for _, task := range tasks {
		teamName := "Unknown"
		teamID := int64(0)
		if task.Team != nil {
			teamName = task.Team.Name
			teamID = task.Team.ID
		}

		if _, exists := teamMap[teamName]; !exists {
			teamMap[teamName] = dto.TasksByTeamResponse{
				Team: dto.TeamNested{
					ID:   teamID,
					Name: teamName,
				},
				Tasks: []dto.TaskResponse{},
			}
		}

		entry := teamMap[teamName]
		entry.Tasks = append(entry.Tasks, convertTaskToResponse(&task))
		teamMap[teamName] = entry
	}

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data:    teamMap,
	})
}

// ListByTeam - GET /v1/tasks/by-team
func (h *TaskHandler) ListByTeam(c *gin.Context) {
	// Get tasks
	var tasks []models.TaskDaily
	if err := h.db.
		Where("deleted_at IS NULL").
		Preload("Team").
		Preload("JobType").
		Preload("JobDetail").
		Preload("Feeder.Station.OperationCenter").
		Order("work_date DESC, created_at DESC").
		Find(&tasks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "DATABASE_ERROR",
				Message: err.Error(),
			},
		})
		return
	}

	// Group by team
	teamMap := make(map[string]dto.TasksByTeamResponse)
	for _, task := range tasks {
		teamName := "Unknown"
		teamID := int64(0)
		if task.Team != nil {
			teamName = task.Team.Name
			teamID = task.Team.ID
		}

		if _, exists := teamMap[teamName]; !exists {
			teamMap[teamName] = dto.TasksByTeamResponse{
				Team: dto.TeamNested{
					ID:   teamID,
					Name: teamName,
				},
				Tasks: []dto.TaskResponse{},
			}
		}

		entry := teamMap[teamName]
		entry.Tasks = append(entry.Tasks, convertTaskToResponse(&task))
		teamMap[teamName] = entry
	}

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data:    teamMap,
	})
}
