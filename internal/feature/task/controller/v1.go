package controller

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	taskdto "backend-hotlines3/internal/feature/task/dto"
	taskentity "backend-hotlines3/internal/feature/task/entity"
	"backend-hotlines3/internal/feature/task/mapper"
	taskrepository "backend-hotlines3/internal/feature/task/repository"
	taskservice "backend-hotlines3/internal/feature/task/service"

	"github.com/gin-gonic/gin"
)

type V1ControllerInterface interface {
	List(ctx *gin.Context)
	ListByTeam(ctx *gin.Context)
	ListByFilter(ctx *gin.Context)
	GetByID(ctx *gin.Context)
	Create(ctx *gin.Context)
	Update(ctx *gin.Context)
	Delete(ctx *gin.Context)
}

func isValidationError(err error) bool {
	return errors.Is(err, taskservice.ErrWorkDateRequired) ||
		errors.Is(err, taskservice.ErrTeamIDRequired) ||
		errors.Is(err, taskservice.ErrJobTypeIDRequired) ||
		errors.Is(err, taskservice.ErrJobDetailIDRequired) ||
		errors.Is(err, taskservice.ErrInvalidTaskID) ||
		errors.Is(err, taskservice.ErrYearMonthRequired)
}

func actorFromContext(ctx *gin.Context) (taskentity.Actor, bool) {
	uidVal, ok := ctx.Get("user_id")
	if !ok {
		return taskentity.Actor{}, false
	}
	roleVal, _ := ctx.Get("role")
	uid, ok := uidVal.(uint)
	if !ok {
		return taskentity.Actor{}, false
	}
	role, _ := roleVal.(string)
	var teamID *int64
	if rawTeamID, exists := ctx.Get("team_id"); exists {
		switch v := rawTeamID.(type) {
		case int64:
			teamID = &v
		case int:
			id := int64(v)
			teamID = &id
		case uint:
			id := int64(v)
			teamID = &id
		case float64:
			id := int64(v)
			teamID = &id
		}
	}
	return taskentity.Actor{UserID: int64(uid), Role: role, TeamID: teamID}, true
}

func requireActor(ctx *gin.Context) (taskentity.Actor, bool) {
	actor, ok := actorFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, taskdto.StandardResponse{
			Success: false,
			Error:   &taskdto.ErrorInfo{Code: "UNAUTHORIZED", Message: "Unauthorized"},
		})
		return taskentity.Actor{}, false
	}
	return actor, true
}

func mapTaskErrorStatus(err error) int {
	switch {
	case errors.Is(err, taskservice.ErrTaskNotFound):
		return http.StatusNotFound
	case errors.Is(err, taskservice.ErrForbidden):
		return http.StatusForbidden
	case isValidationError(err):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func (c *controller) List(ctx *gin.Context) {
	actor, ok := requireActor(ctx)
	if !ok {
		return
	}

	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "50"))

	filter := taskrepository.ListFilter{}
	if workDate := ctx.Query("workDate"); workDate != "" {
		parsedDate, _ := time.Parse("2006-01-02", workDate)
		filter.WorkDate = &parsedDate
	}
	if teamID := ctx.Query("teamId"); teamID != "" {
		id, _ := strconv.ParseInt(teamID, 10, 64)
		filter.TeamID = &id
	}
	if jobTypeID := ctx.Query("jobTypeId"); jobTypeID != "" {
		id, _ := strconv.ParseInt(jobTypeID, 10, 64)
		filter.JobTypeID = &id
	}
	if feederID := ctx.Query("feederId"); feederID != "" {
		id, _ := strconv.ParseInt(feederID, 10, 64)
		filter.FeederID = &id
	}

	result, err := c.service.List(ctx.Request.Context(), actor, taskservice.ListTasksInput{
		Page:   page,
		Limit:  limit,
		Filter: filter,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, taskdto.StandardResponse{
			Success: false,
			Error:   &taskdto.ErrorInfo{Code: "INTERNAL_ERROR", Message: err.Error()},
		})
		return
	}

	response := make([]taskdto.TaskResponse, 0, len(result.Tasks))
	for _, task := range result.Tasks {
		response = append(response, mapper.ToResponse(task))
	}

	ctx.JSON(http.StatusOK, taskdto.StandardResponse{
		Success: true,
		Data:    response,
		Meta:    &taskdto.Meta{Page: result.Page, Limit: result.Limit, Total: result.Total},
	})
}

func (c *controller) GetByID(ctx *gin.Context) {
	actor, ok := requireActor(ctx)
	if !ok {
		return
	}

	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, taskdto.StandardResponse{
			Success: false,
			Error:   &taskdto.ErrorInfo{Code: "INVALID_ID", Message: "Invalid task ID"},
		})
		return
	}

	task, err := c.service.GetByID(ctx.Request.Context(), actor, id)
	if err != nil {
		if errors.Is(err, taskservice.ErrTaskNotFound) {
			ctx.JSON(http.StatusNotFound, taskdto.StandardResponse{
				Success: false,
				Error:   &taskdto.ErrorInfo{Code: "NOT_FOUND", Message: "Task not found"},
			})
			return
		}
		if errors.Is(err, taskservice.ErrInvalidTaskID) {
			ctx.JSON(http.StatusBadRequest, taskdto.StandardResponse{
				Success: false,
				Error:   &taskdto.ErrorInfo{Code: "INVALID_ID", Message: "Invalid task ID"},
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, taskdto.StandardResponse{
			Success: false,
			Error:   &taskdto.ErrorInfo{Code: "INTERNAL_ERROR", Message: err.Error()},
		})
		return
	}

	ctx.JSON(http.StatusOK, taskdto.StandardResponse{Success: true, Data: mapper.ToResponse(*task)})
}

func (c *controller) Create(ctx *gin.Context) {
	actor, ok := requireActor(ctx)
	if !ok {
		return
	}

	var req taskdto.CreateTaskRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, taskdto.StandardResponse{
			Success: false,
			Error:   &taskdto.ErrorInfo{Code: "VALIDATION_ERROR", Message: err.Error()},
		})
		return
	}

	workDate, err := time.Parse("2006-01-02", req.WorkDate)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, taskdto.StandardResponse{
			Success: false,
			Error:   &taskdto.ErrorInfo{Code: "INVALID_DATE", Message: "Invalid work date format. Use YYYY-MM-DD"},
		})
		return
	}

	task, err := c.service.Create(ctx.Request.Context(), actor, taskservice.CreateTaskInput{
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
		if isValidationError(err) {
			ctx.JSON(http.StatusBadRequest, taskdto.StandardResponse{
				Success: false,
				Error:   &taskdto.ErrorInfo{Code: "VALIDATION_ERROR", Message: err.Error()},
			})
			return
		}
		if errors.Is(err, taskservice.ErrForbidden) {
			ctx.JSON(http.StatusForbidden, taskdto.StandardResponse{
				Success: false,
				Error:   &taskdto.ErrorInfo{Code: "FORBIDDEN", Message: err.Error()},
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, taskdto.StandardResponse{
			Success: false,
			Error:   &taskdto.ErrorInfo{Code: "INTERNAL_ERROR", Message: err.Error()},
		})
		return
	}

	ctx.JSON(http.StatusCreated, taskdto.StandardResponse{Success: true, Data: mapper.ToResponse(*task)})
}

func (c *controller) Update(ctx *gin.Context) {
	actor, ok := requireActor(ctx)
	if !ok {
		return
	}

	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, taskdto.StandardResponse{
			Success: false,
			Error:   &taskdto.ErrorInfo{Code: "INVALID_ID", Message: "Invalid task ID"},
		})
		return
	}

	var req taskdto.UpdateTaskRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, taskdto.StandardResponse{
			Success: false,
			Error:   &taskdto.ErrorInfo{Code: "VALIDATION_ERROR", Message: err.Error()},
		})
		return
	}

	input := taskservice.UpdateTaskInput{
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
			ctx.JSON(http.StatusBadRequest, taskdto.StandardResponse{
				Success: false,
				Error:   &taskdto.ErrorInfo{Code: "INVALID_DATE", Message: "Invalid work date format. Use YYYY-MM-DD"},
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

	task, err := c.service.Update(ctx.Request.Context(), actor, input)
	if err != nil {
		if errors.Is(err, taskservice.ErrTaskNotFound) {
			ctx.JSON(http.StatusNotFound, taskdto.StandardResponse{
				Success: false,
				Error:   &taskdto.ErrorInfo{Code: "NOT_FOUND", Message: "Task not found"},
			})
			return
		}
		if errors.Is(err, taskservice.ErrInvalidTaskID) {
			ctx.JSON(http.StatusBadRequest, taskdto.StandardResponse{
				Success: false,
				Error:   &taskdto.ErrorInfo{Code: "INVALID_ID", Message: "Invalid task ID"},
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, taskdto.StandardResponse{
			Success: false,
			Error:   &taskdto.ErrorInfo{Code: "INTERNAL_ERROR", Message: err.Error()},
		})
		return
	}

	ctx.JSON(http.StatusOK, taskdto.StandardResponse{Success: true, Data: mapper.ToResponse(*task)})
}

func (c *controller) Delete(ctx *gin.Context) {
	actor, ok := requireActor(ctx)
	if !ok {
		return
	}

	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, taskdto.StandardResponse{
			Success: false,
			Error:   &taskdto.ErrorInfo{Code: "INVALID_ID", Message: "Invalid task ID"},
		})
		return
	}

	if err := c.service.Delete(ctx.Request.Context(), actor, id); err != nil {
		if errors.Is(err, taskservice.ErrInvalidTaskID) {
			ctx.JSON(http.StatusBadRequest, taskdto.StandardResponse{
				Success: false,
				Error:   &taskdto.ErrorInfo{Code: "INVALID_ID", Message: "Invalid task ID"},
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, taskdto.StandardResponse{
			Success: false,
			Error:   &taskdto.ErrorInfo{Code: "INTERNAL_ERROR", Message: err.Error()},
		})
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (c *controller) ListByFilter(ctx *gin.Context) {
	actor, ok := requireActor(ctx)
	if !ok {
		return
	}

	input := taskservice.ListTasksByFilterInput{
		Year:  ctx.Query("year"),
		Month: ctx.Query("month"),
	}
	if teamID := ctx.Query("teamId"); teamID != "" {
		id, _ := strconv.ParseInt(teamID, 10, 64)
		input.TeamID = &id
	}

	result, err := c.service.ListByFilter(ctx.Request.Context(), actor, input)
	if err != nil {
		if errors.Is(err, taskservice.ErrYearMonthRequired) {
			ctx.JSON(http.StatusBadRequest, taskdto.StandardResponse{
				Success: false,
				Error:   &taskdto.ErrorInfo{Code: "VALIDATION_ERROR", Message: "year and month are required"},
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, taskdto.StandardResponse{
			Success: false,
			Error:   &taskdto.ErrorInfo{Code: "INTERNAL_ERROR", Message: err.Error()},
		})
		return
	}

	ctx.JSON(http.StatusOK, taskdto.StandardResponse{Success: true, Data: mapper.GroupByTeam(result.Tasks)})
}

func (c *controller) ListByTeam(ctx *gin.Context) {
	actor, ok := requireActor(ctx)
	if !ok {
		return
	}

	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "100"))

	input := taskservice.ListTasksByTeamInput{Page: page, Limit: limit}
	if workDate := ctx.Query("workDate"); workDate != "" {
		parsedDate, _ := time.Parse("2006-01-02", workDate)
		input.WorkDate = &parsedDate
	}
	if teamID := ctx.Query("teamId"); teamID != "" {
		id, _ := strconv.ParseInt(teamID, 10, 64)
		input.TeamID = &id
	}
	if jobTypeID := ctx.Query("jobTypeId"); jobTypeID != "" {
		id, _ := strconv.ParseInt(jobTypeID, 10, 64)
		input.JobTypeID = &id
	}
	if feederID := ctx.Query("feederId"); feederID != "" {
		id, _ := strconv.ParseInt(feederID, 10, 64)
		input.FeederID = &id
	}

	result, err := c.service.ListByTeam(ctx.Request.Context(), actor, input)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, taskdto.StandardResponse{
			Success: false,
			Error:   &taskdto.ErrorInfo{Code: "INTERNAL_ERROR", Message: err.Error()},
		})
		return
	}

	ctx.JSON(http.StatusOK, taskdto.StandardResponse{
		Success: true,
		Data:    mapper.GroupByTeam(result.Tasks),
		Meta:    &taskdto.Meta{Page: result.Page, Limit: result.Limit, Total: result.Total},
	})
}
