package controller

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"backend-hotlines3/internal/feature/monthlyplan/dto"
	mpdomain "backend-hotlines3/internal/feature/monthlyplan/entity"
	"backend-hotlines3/internal/feature/monthlyplan/mapper"
	"backend-hotlines3/internal/feature/monthlyplan/service"

	"github.com/gin-gonic/gin"
)

// Controller handles MonthlyPlan routes.
// ─── Helpers ───────────────────────────────────────────────────────────────

// getAuthInfo extracts user_id (uint) and role from the gin context.
func getAuthInfo(c *gin.Context) (userID uint, role string, ok bool) {
	uidVal, exists := c.Get("user_id")
	if !exists {
		return 0, "", false
	}
	roleVal, _ := c.Get("role")
	uid, _ := uidVal.(uint)
	r, _ := roleVal.(string)
	return uid, r, true
}

// actorFromContext builds a domain Actor from gin context.
func actorFromContext(c *gin.Context) (mpdomain.Actor, bool) {
	uid, role, ok := getAuthInfo(c)
	if !ok {
		return mpdomain.Actor{}, false
	}
	var teamID *int64
	if tidVal, exists := c.Get("team_id"); exists {
		if tid, err := toInt64(tidVal); err == nil && tid > 0 {
			teamID = &tid
		}
	}
	return mpdomain.Actor{
		UserID:       int64(uid),
		Role:         role,
		TeamID:       teamID,
		Capabilities: capabilitiesFromContext(c),
	}, true
}

func capabilitiesFromContext(c *gin.Context) []string {
	raw, exists := c.Get("capabilities")
	if !exists {
		return nil
	}
	caps, ok := raw.([]string)
	if !ok {
		return nil
	}
	return append([]string{}, caps...)
}

func parseOptionalDate(raw *string) (*time.Time, error) {
	if raw == nil || *raw == "" {
		return nil, nil
	}
	parsed, err := time.Parse("2006-01-02", *raw)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func toInt64(v interface{}) (int64, error) {
	switch val := v.(type) {
	case int:
		return int64(val), nil
	case int64:
		return val, nil
	case uint:
		return int64(val), nil
	case uint64:
		return int64(val), nil
	case float64:
		return int64(val), nil
	default:
		return 0, fmt.Errorf("cannot convert %T to int64", v)
	}
}

// errResp is a helper to build error responses.
func errResp(code, msg string) *dto.ErrorInfo {
	return &dto.ErrorInfo{Code: code, Message: msg}
}

// mapMPError maps domain errors to HTTP status codes.
func mapMPError(err error) int {
	switch {
	case errors.Is(err, mpdomain.ErrPeriodNotFound),
		errors.Is(err, mpdomain.ErrFileNotFound):
		return http.StatusNotFound
	case errors.Is(err, mpdomain.ErrInvalidPeriod),
		errors.Is(err, mpdomain.ErrInvalidFileID),
		errors.Is(err, mpdomain.ErrInvalidFileType),
		errors.Is(err, mpdomain.ErrFileNotDeleted):
		return http.StatusBadRequest
	case errors.Is(err, mpdomain.ErrPeriodLocked),
		errors.Is(err, mpdomain.ErrForbiddenAction):
		return http.StatusForbidden
	case errors.Is(err, mpdomain.ErrSettingsLoadFail),
		errors.Is(err, mpdomain.ErrSettingsSaveFail):
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// ─── Period Endpoints ──────────────────────────────────────────────────────

func (c *Controller) GetYearOverview(ctx *gin.Context) {
	year, err := strconv.Atoi(ctx.Param("year"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("BAD_REQUEST", "Invalid year")})
		return
	}
	actor, ok := actorFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.StandardResponse{Success: false, Error: errResp("UNAUTHORIZED", "Unauthorized")})
		return
	}
	overview, err := c.service.GetYearOverview(ctx.Request.Context(), actor, year)
	if err != nil {
		ctx.JSON(mapMPError(err), dto.StandardResponse{Success: false, Error: errResp("ERROR", err.Error())})
		return
	}
	months := make([]dto.MonthlyPlanOverviewMonthResponse, len(overview.Months))
	for i, m := range overview.Months {
		files := make([]dto.PlanFileResponse, len(m.Files))
		for j, f := range m.Files {
			files[j] = mapper.ToPlanFileResponse(f)
		}
		months[i] = dto.MonthlyPlanOverviewMonthResponse{
			Period: dto.MonthlyPlanResponse{
				ID:        m.Period.ID,
				Year:      m.Period.Year,
				Month:     m.Period.Month,
				IsLocked:  m.IsLocked,
				CreatedAt: m.Period.CreatedAt.Format(time.RFC3339),
			},
			Month:    m.Period.Month,
			Deadline: m.Deadline,
			IsLocked: m.IsLocked,
			Status:   m.Status,
			Actions:  dto.MonthlyPlanActionResponse{CanUpload: m.CanUpload},
			Files:    files,
		}
	}
	ctx.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data:    dto.MonthlyPlanYearOverviewResponse{Year: overview.Year, Months: months},
	})
}

func (c *Controller) GetOrCreatePeriod(ctx *gin.Context) {
	year, err := strconv.Atoi(ctx.Param("year"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("BAD_REQUEST", "Invalid year")})
		return
	}
	month, err := strconv.Atoi(ctx.Param("month"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("BAD_REQUEST", "Invalid month")})
		return
	}

	period, err := c.service.EnsurePeriod(ctx.Request.Context(), year, month)
	if err != nil {
		ctx.JSON(mapMPError(err), dto.StandardResponse{Success: false, Error: errResp("ERROR", err.Error())})
		return
	}

	ctx.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data: dto.MonthlyPlanResponse{
			ID:        period.ID,
			Year:      period.Year,
			Month:     period.Month,
			IsLocked:  period.IsLocked,
			CreatedAt: period.CreatedAt.Format(time.RFC3339),
		},
	})
}

// ─── File Endpoints ────────────────────────────────────────────────────────

func (c *Controller) ListFiles(ctx *gin.Context) {
	year, err := strconv.Atoi(ctx.Param("year"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("BAD_REQUEST", "Invalid year")})
		return
	}
	month, err := strconv.Atoi(ctx.Param("month"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("BAD_REQUEST", "Invalid month")})
		return
	}

	actor, ok := actorFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.StandardResponse{Success: false, Error: errResp("UNAUTHORIZED", "Unauthorized")})
		return
	}

	period, err := c.service.EnsurePeriod(ctx.Request.Context(), year, month)
	if err != nil {
		ctx.JSON(mapMPError(err), dto.StandardResponse{Success: false, Error: errResp("ERROR", err.Error())})
		return
	}

	var filterTeamID int64
	if rawTeam := ctx.Query("teamId"); rawTeam != "" {
		filterTeamID, _ = strconv.ParseInt(rawTeam, 10, 64)
	}
	search := ctx.Query("search")

	files, err := c.service.ListFiles(ctx.Request.Context(), actor, period.ID, filterTeamID, search)
	if err != nil {
		ctx.JSON(mapMPError(err), dto.StandardResponse{Success: false, Error: errResp("ERROR", err.Error())})
		return
	}

	respFiles := make([]dto.PlanFileResponse, len(files))
	for i, f := range files {
		respFiles[i] = mapper.ToPlanFileResponse(f)
	}

	ctx.JSON(http.StatusOK, dto.StandardResponse{Success: true, Data: respFiles})
}

func (c *Controller) PresignUpload(ctx *gin.Context) {
	year, err := strconv.Atoi(ctx.Param("year"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("BAD_REQUEST", "Invalid year")})
		return
	}
	month, err := strconv.Atoi(ctx.Param("month"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("BAD_REQUEST", "Invalid month")})
		return
	}

	actor, ok := actorFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.StandardResponse{Success: false, Error: errResp("UNAUTHORIZED", "Unauthorized")})
		return
	}

	period, err := c.service.EnsurePeriod(ctx.Request.Context(), year, month)
	if err != nil {
		ctx.JSON(mapMPError(err), dto.StandardResponse{Success: false, Error: errResp("ERROR", err.Error())})
		return
	}
	canUpload, _, err := c.service.CanUploadForPeriod(ctx.Request.Context(), actor, year, month)
	if err != nil {
		ctx.JSON(mapMPError(err), dto.StandardResponse{Success: false, Error: errResp("ERROR", err.Error())})
		return
	}
	if !canUpload {
		ctx.JSON(mapMPError(mpdomain.ErrPeriodLocked), dto.StandardResponse{Success: false, Error: errResp("ERROR", mpdomain.ErrPeriodLocked.Error())})
		return
	}

	var req dto.PresignPlanFileRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("BAD_REQUEST", "fileName and fileType are required")})
		return
	}

	isMaster := ctx.Query("isMasterPlan") == "true"
	var teamOverride *int64
	if rawTeam := ctx.Query("teamId"); rawTeam != "" {
		if tid, e := strconv.ParseInt(rawTeam, 10, 64); e == nil {
			teamOverride = &tid
		}
	}

	out, err := c.service.PresignUpload(ctx.Request.Context(), actor, period.ID, req.FileName, req.FileType, isMaster, teamOverride)
	if err != nil {
		ctx.JSON(mapMPError(err), dto.StandardResponse{Success: false, Error: errResp("ERROR", err.Error())})
		return
	}

	ctx.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data:    dto.PresignedURLResponse{UploadURL: out.UploadURL, FileURL: out.FileURL, FileKey: out.FileKey},
	})
}

func (c *Controller) ConfirmUpload(ctx *gin.Context) {
	year, err := strconv.Atoi(ctx.Param("year"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("BAD_REQUEST", "Invalid year")})
		return
	}
	month, err := strconv.Atoi(ctx.Param("month"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("BAD_REQUEST", "Invalid month")})
		return
	}

	actor, ok := actorFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.StandardResponse{Success: false, Error: errResp("UNAUTHORIZED", "Unauthorized")})
		return
	}

	period, err := c.service.EnsurePeriod(ctx.Request.Context(), year, month)
	if err != nil {
		ctx.JSON(mapMPError(err), dto.StandardResponse{Success: false, Error: errResp("ERROR", err.Error())})
		return
	}
	canUpload, _, err := c.service.CanUploadForPeriod(ctx.Request.Context(), actor, year, month)
	if err != nil {
		ctx.JSON(mapMPError(err), dto.StandardResponse{Success: false, Error: errResp("ERROR", err.Error())})
		return
	}
	if !canUpload {
		ctx.JSON(mapMPError(mpdomain.ErrPeriodLocked), dto.StandardResponse{Success: false, Error: errResp("ERROR", mpdomain.ErrPeriodLocked.Error())})
		return
	}

	var req dto.ConfirmPlanFileRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("BAD_REQUEST", "Invalid request body")})
		return
	}

	workStartDate, err := parseOptionalDate(req.WorkStartDate)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("BAD_REQUEST", "workStartDate must be YYYY-MM-DD")})
		return
	}
	workEndDate, err := parseOptionalDate(req.WorkEndDate)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("BAD_REQUEST", "workEndDate must be YYYY-MM-DD")})
		return
	}

	file, err := c.service.ConfirmUpload(ctx.Request.Context(), actor, service.PlanFileCreateInput{
		MonthlyPlanID: period.ID,
		TeamID:        req.TeamID,
		UploadedByID:  actor.UserID,
		FileKey:       req.FileKey,
		FileURL:       req.FileURL,
		FileName:      req.FileName,
		FileSizeBytes: req.FileSizeBytes,
		Description:   req.Description,
		WorkStartDate: workStartDate,
		WorkEndDate:   workEndDate,
		Destination:   req.Destination,
		Remarks:       req.Remarks,
		IsMasterPlan:  req.IsMasterPlan,
	})
	if err != nil {
		ctx.JSON(mapMPError(err), dto.StandardResponse{Success: false, Error: errResp("ERROR", err.Error())})
		return
	}

	ctx.JSON(http.StatusCreated, dto.StandardResponse{Success: true, Data: mapper.ToPlanFileResponse(*file)})
}

func (c *Controller) SoftDeleteFile(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("BAD_REQUEST", "Invalid file ID")})
		return
	}

	actor, ok := actorFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.StandardResponse{Success: false, Error: errResp("UNAUTHORIZED", "Unauthorized")})
		return
	}

	if err := c.service.SoftDeleteFile(ctx.Request.Context(), actor, id); err != nil {
		ctx.JSON(mapMPError(err), dto.StandardResponse{Success: false, Error: errResp("ERROR", err.Error())})
		return
	}

	ctx.JSON(http.StatusOK, dto.StandardResponse{Success: true, Data: "File deleted"})
}

func (c *Controller) HardDeleteFile(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("BAD_REQUEST", "Invalid file ID")})
		return
	}

	actor, ok := actorFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.StandardResponse{Success: false, Error: errResp("UNAUTHORIZED", "Unauthorized")})
		return
	}

	if err := c.service.HardDeleteFile(ctx.Request.Context(), actor, id); err != nil {
		ctx.JSON(mapMPError(err), dto.StandardResponse{Success: false, Error: errResp("ERROR", err.Error())})
		return
	}

	ctx.JSON(http.StatusOK, dto.StandardResponse{Success: true, Data: "File permanently deleted"})
}

func (c *Controller) RestoreFile(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("BAD_REQUEST", "Invalid file ID")})
		return
	}

	actor, ok := actorFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.StandardResponse{Success: false, Error: errResp("UNAUTHORIZED", "Unauthorized")})
		return
	}

	file, err := c.service.RestoreFile(ctx.Request.Context(), actor, id)
	if err != nil {
		ctx.JSON(mapMPError(err), dto.StandardResponse{Success: false, Error: errResp("ERROR", err.Error())})
		return
	}

	ctx.JSON(http.StatusOK, dto.StandardResponse{Success: true, Data: mapper.ToPlanFileResponse(*file)})
}

func (c *Controller) GetDownloadURL(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("BAD_REQUEST", "Invalid file ID")})
		return
	}

	actor, ok := actorFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.StandardResponse{Success: false, Error: errResp("UNAUTHORIZED", "Unauthorized")})
		return
	}

	if !actor.CanDownloadFile() {
		ctx.JSON(http.StatusForbidden, dto.StandardResponse{Success: false, Error: errResp("FORBIDDEN", "Viewers cannot download files")})
		return
	}

	file, err := c.service.GetFile(ctx.Request.Context(), actor, id)
	if err != nil {
		ctx.JSON(mapMPError(err), dto.StandardResponse{Success: false, Error: errResp("ERROR", err.Error())})
		return
	}

	downloadURL, err := c.service.PresignDownload(ctx.Request.Context(), file.FileKey)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.StandardResponse{Success: false, Error: errResp("ERROR", "Failed to generate download URL")})
		return
	}

	ctx.JSON(http.StatusOK, dto.StandardResponse{Success: true, Data: downloadURL})
}

func (c *Controller) GetSubmissionStatus(ctx *gin.Context) {
	year, err := strconv.Atoi(ctx.Param("year"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("BAD_REQUEST", "Invalid year")})
		return
	}
	month, err := strconv.Atoi(ctx.Param("month"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("BAD_REQUEST", "Invalid month")})
		return
	}

	actor, ok := actorFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.StandardResponse{Success: false, Error: errResp("UNAUTHORIZED", "Unauthorized")})
		return
	}

	period, err := c.service.EnsurePeriod(ctx.Request.Context(), year, month)
	if err != nil {
		ctx.JSON(mapMPError(err), dto.StandardResponse{Success: false, Error: errResp("ERROR", err.Error())})
		return
	}

	out, err := c.service.GetSubmissionStatus(ctx.Request.Context(), actor, period.ID)
	if err != nil {
		ctx.JSON(mapMPError(err), dto.StandardResponse{Success: false, Error: errResp("ERROR", err.Error())})
		return
	}

	teams := make([]dto.TeamSubmissionStatus, len(out.Teams))
	for i, t := range out.Teams {
		teams[i] = dto.TeamSubmissionStatus{
			Team:      dto.TeamNested{ID: t.TeamID, Name: t.TeamName},
			Status:    t.Status,
			FileCount: t.FileCount,
		}
	}

	ctx.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data: dto.SubmissionStatusResponse{
			Period: dto.MonthlyPlanResponse{
				ID:        period.ID,
				Year:      period.Year,
				Month:     period.Month,
				IsLocked:  period.IsLocked,
				CreatedAt: period.CreatedAt.Format(time.RFC3339),
			},
			Deadline: out.Deadline,
			Teams:    teams,
		},
	})
}

// ─── Settings Endpoints ────────────────────────────────────────────────────

func (c *Controller) GetSettings(ctx *gin.Context) {
	settings, err := c.service.GetSettings(ctx.Request.Context())
	if err != nil {
		ctx.JSON(mapMPError(err), dto.StandardResponse{Success: false, Error: errResp("ERROR", err.Error())})
		return
	}

	ctx.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data: dto.MonthlyPlanSettingsResponse{
			LockDay:                 settings.LockDay,
			AutoCreateDay:           settings.AutoCreateDay,
			AutoCreateTarget:        settings.AutoCreateTarget,
			AllowedFileTypes:        settings.AllowedFileTypes,
			MaxFileSizeMB:           settings.MaxFileSizeMB,
			ReminderStartDay:        settings.ReminderStartDay,
			AdminCanUploadAfterLock: settings.AdminCanUploadAfterLock,
		},
	})
}

func (c *Controller) UpdateSettings(ctx *gin.Context) {
	actor, ok := actorFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.StandardResponse{Success: false, Error: errResp("UNAUTHORIZED", "Unauthorized")})
		return
	}

	var req dto.UpdateMonthlyPlanSettingsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("BAD_REQUEST", "Invalid request body")})
		return
	}

	settings, err := c.service.UpdateSettings(ctx.Request.Context(), actor, service.SettingsPatch{
		LockDay:                 req.LockDay,
		AutoCreateDay:           req.AutoCreateDay,
		AutoCreateTarget:        req.AutoCreateTarget,
		AllowedFileTypes:        req.AllowedFileTypes,
		MaxFileSizeMB:           req.MaxFileSizeMB,
		ReminderStartDay:        req.ReminderStartDay,
		AdminCanUploadAfterLock: req.AdminCanUploadAfterLock,
	})
	if err != nil {
		ctx.JSON(mapMPError(err), dto.StandardResponse{Success: false, Error: errResp("ERROR", err.Error())})
		return
	}

	ctx.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data: dto.MonthlyPlanSettingsResponse{
			LockDay:                 settings.LockDay,
			AutoCreateDay:           settings.AutoCreateDay,
			AutoCreateTarget:        settings.AutoCreateTarget,
			AllowedFileTypes:        settings.AllowedFileTypes,
			MaxFileSizeMB:           settings.MaxFileSizeMB,
			ReminderStartDay:        settings.ReminderStartDay,
			AdminCanUploadAfterLock: settings.AdminCanUploadAfterLock,
		},
	})
}
