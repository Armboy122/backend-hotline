package v1

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	taskusecase "backend-hotlines3/internal/app/task/usecase"
	taskdomain "backend-hotlines3/internal/domain/task"
	"backend-hotlines3/internal/dto"
	"backend-hotlines3/internal/port/outbound/repository"

	"github.com/gin-gonic/gin"
)

// Use case ports for handler wiring — each is a single-method interface
// so the handler can be tested with fakes without needing the real usecase.
type listTasksPort interface {
	Execute(ctx context.Context, input taskusecase.ListTasksInput) (*taskusecase.ListTasksOutput, error)
}
type getTaskPort interface {
	Execute(ctx context.Context, input taskusecase.GetTaskInput) (*taskdomain.Entity, error)
}
type createTaskPort interface {
	Execute(ctx context.Context, input taskusecase.CreateTaskInput) (*taskdomain.Entity, error)
}
type updateTaskPort interface {
	Execute(ctx context.Context, input taskusecase.UpdateTaskInput) (*taskdomain.Entity, error)
}
type deleteTaskPort interface {
	Execute(ctx context.Context, input taskusecase.DeleteTaskInput) error
}
type listTasksByTeamPort interface {
	Execute(ctx context.Context, input taskusecase.ListTasksByTeamInput) (*taskusecase.ListTasksByTeamOutput, error)
}
type listTasksByFilterPort interface {
	Execute(ctx context.Context, input taskusecase.ListTasksByFilterInput) (*taskusecase.ListTasksByFilterOutput, error)
}

type TaskHandler struct {
	listUC         listTasksPort
	getUC          getTaskPort
	createUC       createTaskPort
	updateUC       updateTaskPort
	deleteUC       deleteTaskPort
	listByTeamUC   listTasksByTeamPort
	listByFilterUC listTasksByFilterPort
}

func NewTaskHandler(repo repository.TaskRepository) *TaskHandler {
	return &TaskHandler{
		listUC:         taskusecase.NewListTasksUseCase(repo),
		getUC:          taskusecase.NewGetTaskUseCase(repo),
		createUC:       taskusecase.NewCreateTaskUseCase(repo),
		updateUC:       taskusecase.NewUpdateTaskUseCase(repo),
		deleteUC:       taskusecase.NewDeleteTaskUseCase(repo),
		listByTeamUC:   taskusecase.NewListTasksByTeamUseCase(repo),
		listByFilterUC: taskusecase.NewListTasksByFilterUseCase(repo),
	}
}

// isValidationError checks if the error is a known usecase validation sentinel.
func isValidationError(err error) bool {
	return errors.Is(err, taskusecase.ErrWorkDateRequired) ||
		errors.Is(err, taskusecase.ErrTeamIDRequired) ||
		errors.Is(err, taskusecase.ErrJobTypeIDRequired) ||
		errors.Is(err, taskusecase.ErrJobDetailIDRequired) ||
		errors.Is(err, taskusecase.ErrInvalidTaskID) ||
		errors.Is(err, taskusecase.ErrYearMonthRequired)
}

// groupTasksByTeam is the shared helper for ListByFilter and ListByTeam responses. (SMELL-1 fix)
func groupTasksByTeam(tasks []taskdomain.Entity) []dto.TasksByTeamResponse {
	teamMap := make(map[string]dto.TasksByTeamResponse)
	order := make([]string, 0, len(tasks))

	for _, task := range tasks {
		teamName := "Unknown"
		teamID := int64(0)
		if task.TeamName != nil {
			teamName = *task.TeamName
			teamID = task.TeamID
		}
		if _, exists := teamMap[teamName]; !exists {
			teamMap[teamName] = dto.TasksByTeamResponse{
				Team:  dto.TeamNested{ID: teamID, Name: teamName},
				Tasks: []dto.TaskResponse{},
			}
			order = append(order, teamName)
		}
		entry := teamMap[teamName]
		entry.Tasks = append(entry.Tasks, convertDomainTaskToResponse(task))
		teamMap[teamName] = entry
	}

	response := make([]dto.TasksByTeamResponse, 0, len(order))
	for _, name := range order {
		response = append(response, teamMap[name])
	}
	return response
}

// convertDomainTaskToResponse converts a domain entity to TaskResponse DTO.
func convertDomainTaskToResponse(task taskdomain.Entity) dto.TaskResponse {
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
		URLsBefore:  task.URLsBefore,
		URLsAfter:   task.URLsAfter,
		Latitude:    task.Latitude,
		Longitude:   task.Longitude,
		CreatedAt:   task.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   task.UpdatedAt.Format(time.RFC3339),
	}

	if task.DeletedAt != nil {
		formatted := task.DeletedAt.Format(time.RFC3339)
		response.DeletedAt = &formatted
	}
	if task.TeamName != nil {
		response.Team = &dto.TeamNested{ID: task.TeamID, Name: *task.TeamName}
	}
	if task.JobTypeName != nil {
		response.JobType = &dto.JobTypeNested{ID: task.JobTypeID, Name: *task.JobTypeName}
	}
	if task.JobDetailName != nil {
		response.JobDetail = &dto.JobDetailNested{ID: task.JobDetailID, Name: *task.JobDetailName}
	}
	if task.FeederID != nil && task.FeederCode != nil {
		response.Feeder = &dto.FeederNestedForTask{
			ID:   *task.FeederID,
			Code: *task.FeederCode,
		}
		if task.StationName != nil {
			response.Feeder.Station = &dto.StationNestedSimple{
				Name: *task.StationName,
			}
			if task.OperationCenterID != nil && task.OperationCenterName != nil {
				response.Feeder.Station.OperationCenter = &dto.OperationCenterNested{
					ID:   *task.OperationCenterID,
					Name: *task.OperationCenterName,
				}
			}
		}
	}

	return response
}

// List - GET /v1/tasks
func (h *TaskHandler) List(c *gin.Context) {
	ctx := c.Request.Context()

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	filter := repository.TaskListFilter{}
	if workDate := c.Query("workDate"); workDate != "" {
		parsedDate, _ := time.Parse("2006-01-02", workDate)
		filter.WorkDate = &parsedDate
	}
	if teamID := c.Query("teamId"); teamID != "" {
		id, _ := strconv.ParseInt(teamID, 10, 64)
		filter.TeamID = &id
	}
	if jobTypeID := c.Query("jobTypeId"); jobTypeID != "" {
		id, _ := strconv.ParseInt(jobTypeID, 10, 64)
		filter.JobTypeID = &id
	}
	if feederID := c.Query("feederId"); feederID != "" {
		id, _ := strconv.ParseInt(feederID, 10, 64)
		filter.FeederID = &id
	}

	result, err := h.listUC.Execute(ctx, taskusecase.ListTasksInput{
		Page:   page,
		Limit:  limit,
		Filter: filter,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "INTERNAL_ERROR", Message: err.Error()},
		})
		return
	}

	response := make([]dto.TaskResponse, 0, len(result.Tasks))
	for _, task := range result.Tasks {
		response = append(response, convertDomainTaskToResponse(task))
	}

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data:    response,
		Meta:    &dto.Meta{Page: result.Page, Limit: result.Limit, Total: result.Total},
	})
}

// GetByID - GET /v1/tasks/:id
func (h *TaskHandler) GetByID(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "INVALID_ID", Message: "Invalid task ID"},
		})
		return
	}

	entity, err := h.getUC.Execute(ctx, taskusecase.GetTaskInput{ID: id})
	if err != nil {
		if errors.Is(err, taskusecase.ErrTaskNotFound) {
			c.JSON(http.StatusNotFound, dto.StandardResponse{
				Success: false,
				Error:   &dto.ErrorInfo{Code: "NOT_FOUND", Message: "Task not found"},
			})
			return
		}
		if errors.Is(err, taskusecase.ErrInvalidTaskID) {
			c.JSON(http.StatusBadRequest, dto.StandardResponse{
				Success: false,
				Error:   &dto.ErrorInfo{Code: "INVALID_ID", Message: "Invalid task ID"},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "INTERNAL_ERROR", Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data:    convertDomainTaskToResponse(*entity),
	})
}

// Create - POST /v1/tasks
func (h *TaskHandler) Create(c *gin.Context) {
	ctx := c.Request.Context()
	var req dto.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "VALIDATION_ERROR", Message: err.Error()},
		})
		return
	}

	workDate, err := time.Parse("2006-01-02", req.WorkDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "INVALID_DATE", Message: "Invalid work date format. Use YYYY-MM-DD"},
		})
		return
	}

	entity, err := h.createUC.Execute(ctx, taskusecase.CreateTaskInput{
		WorkDate:    workDate,
		TeamID:      req.TeamID,
		JobTypeID:   req.JobTypeID,
		JobDetailID: req.JobDetailID,
		FeederID:    req.FeederID,
		NumPole:     req.NumPole,
		DeviceCode:  req.DeviceCode,
		Detail:      req.Detail,
		URLsBefore:  req.URLsBefore,
		URLsAfter:   req.URLsAfter,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
	})
	if err != nil {
		// BUG-2 fix: map usecase validation errors to 400 instead of 500
		if isValidationError(err) {
			c.JSON(http.StatusBadRequest, dto.StandardResponse{
				Success: false,
				Error:   &dto.ErrorInfo{Code: "VALIDATION_ERROR", Message: err.Error()},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "INTERNAL_ERROR", Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusCreated, dto.StandardResponse{
		Success: true,
		Data:    convertDomainTaskToResponse(*entity),
	})
}

// Update - PUT /v1/tasks/:id
func (h *TaskHandler) Update(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "INVALID_ID", Message: "Invalid task ID"},
		})
		return
	}

	var req dto.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "VALIDATION_ERROR", Message: err.Error()},
		})
		return
	}

	input := taskusecase.UpdateTaskInput{
		ID:         id,
		FeederID:   req.FeederID,
		NumPole:    req.NumPole,
		DeviceCode: req.DeviceCode,
		Detail:     req.Detail,
		URLsBefore: req.URLsBefore,
		URLsAfter:  req.URLsAfter,
		Latitude:   req.Latitude,
		Longitude:  req.Longitude,
	}
	if req.WorkDate != nil {
		wd, err := time.Parse("2006-01-02", *req.WorkDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, dto.StandardResponse{
				Success: false,
				Error:   &dto.ErrorInfo{Code: "INVALID_DATE", Message: "Invalid work date format. Use YYYY-MM-DD"},
			})
			return
		}
		input.WorkDate = &wd
	}
	if req.TeamID != nil {
		input.TeamID = req.TeamID
	}
	if req.JobTypeID != nil {
		input.JobTypeID = req.JobTypeID
	}
	if req.JobDetailID != nil {
		input.JobDetailID = req.JobDetailID
	}

	entity, err := h.updateUC.Execute(ctx, input)
	if err != nil {
		if errors.Is(err, taskusecase.ErrTaskNotFound) {
			c.JSON(http.StatusNotFound, dto.StandardResponse{
				Success: false,
				Error:   &dto.ErrorInfo{Code: "NOT_FOUND", Message: "Task not found"},
			})
			return
		}
		if errors.Is(err, taskusecase.ErrInvalidTaskID) {
			c.JSON(http.StatusBadRequest, dto.StandardResponse{
				Success: false,
				Error:   &dto.ErrorInfo{Code: "INVALID_ID", Message: "Invalid task ID"},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "INTERNAL_ERROR", Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data:    convertDomainTaskToResponse(*entity),
	})
}

// Delete - DELETE /v1/tasks/:id (Soft Delete)
func (h *TaskHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "INVALID_ID", Message: "Invalid task ID"},
		})
		return
	}

	if err := h.deleteUC.Execute(c.Request.Context(), taskusecase.DeleteTaskInput{ID: id}); err != nil {
		if errors.Is(err, taskusecase.ErrInvalidTaskID) {
			c.JSON(http.StatusBadRequest, dto.StandardResponse{
				Success: false,
				Error:   &dto.ErrorInfo{Code: "INVALID_ID", Message: "Invalid task ID"},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "INTERNAL_ERROR", Message: err.Error()},
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// ListByFilter - GET /v1/tasks/by-filter
func (h *TaskHandler) ListByFilter(c *gin.Context) {
	ctx := c.Request.Context()
	year := c.Query("year")
	month := c.Query("month")

	result, err := h.listByFilterUC.Execute(ctx, taskusecase.ListTasksByFilterInput{
		Year:  year,
		Month: month,
	})
	if err != nil {
		if errors.Is(err, taskusecase.ErrYearMonthRequired) {
			c.JSON(http.StatusBadRequest, dto.StandardResponse{
				Success: false,
				Error:   &dto.ErrorInfo{Code: "VALIDATION_ERROR", Message: "year and month are required"},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "INTERNAL_ERROR", Message: err.Error()},
		})
		return
	}

	// SMELL-1 fix: use shared groupTasksByTeam helper
	response := groupTasksByTeam(result.Tasks)

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data:    response,
	})
}

// ListByTeam - GET /v1/tasks/by-team
func (h *TaskHandler) ListByTeam(c *gin.Context) {
	ctx := c.Request.Context()

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))

	input := taskusecase.ListTasksByTeamInput{
		Page:  page,
		Limit: limit,
	}
	if workDate := c.Query("workDate"); workDate != "" {
		parsedDate, _ := time.Parse("2006-01-02", workDate)
		input.WorkDate = &parsedDate
	}
	if teamID := c.Query("teamId"); teamID != "" {
		id, _ := strconv.ParseInt(teamID, 10, 64)
		input.TeamID = &id
	}
	if jobTypeID := c.Query("jobTypeId"); jobTypeID != "" {
		id, _ := strconv.ParseInt(jobTypeID, 10, 64)
		input.JobTypeID = &id
	}
	if feederID := c.Query("feederId"); feederID != "" {
		id, _ := strconv.ParseInt(feederID, 10, 64)
		input.FeederID = &id
	}

	result, err := h.listByTeamUC.Execute(ctx, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "INTERNAL_ERROR", Message: err.Error()},
		})
		return
	}

	// SMELL-1 fix: use shared groupTasksByTeam helper
	response := groupTasksByTeam(result.Tasks)

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data:    response,
		Meta:    &dto.Meta{Page: result.Page, Limit: result.Limit, Total: result.Total},
	})
}
