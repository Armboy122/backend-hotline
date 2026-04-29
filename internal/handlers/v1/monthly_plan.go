package v1

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	mpdomain "backend-hotlines3/internal/domain/monthlyplan"
	"backend-hotlines3/internal/dto"

	mpusecase "backend-hotlines3/internal/app/monthlyplan/usecase"

	"github.com/gin-gonic/gin"
)

// MonthlyPlanHandler handles all monthly plan file management endpoints.
// It delegates business logic to use cases and only handles HTTP concerns.
type MonthlyPlanHandler struct {
	getSettingsUC      *mpusecase.GetSettingsUseCase
	updateSettingsUC   *mpusecase.UpdateSettingsUseCase
	ensurePeriodUC     *mpusecase.EnsurePeriodUseCase
	presignUploadUC    *mpusecase.PresignUploadUseCase
	confirmUploadUC    *mpusecase.ConfirmUploadUseCase
	listFilesUC        *mpusecase.ListFilesUseCase
	getFileUC          *mpusecase.GetFileUseCase
	softDeleteFileUC   *mpusecase.SoftDeleteFileUseCase
	restoreFileUC      *mpusecase.RestoreFileUseCase
	hardDeleteFileUC   *mpusecase.HardDeleteFileUseCase
	submissionStatusUC *mpusecase.GetSubmissionStatusUseCase
	downloadURLFunc    func(fileKey string) (string, error)
}

// MonthlyPlanHandlerDeps holds all use case dependencies for the handler.
type MonthlyPlanHandlerDeps struct {
	GetSettingsUC      *mpusecase.GetSettingsUseCase
	UpdateSettingsUC   *mpusecase.UpdateSettingsUseCase
	EnsurePeriodUC     *mpusecase.EnsurePeriodUseCase
	PresignUploadUC    *mpusecase.PresignUploadUseCase
	ConfirmUploadUC    *mpusecase.ConfirmUploadUseCase
	ListFilesUC        *mpusecase.ListFilesUseCase
	GetFileUC          *mpusecase.GetFileUseCase
	SoftDeleteFileUC   *mpusecase.SoftDeleteFileUseCase
	RestoreFileUC      *mpusecase.RestoreFileUseCase
	HardDeleteFileUC   *mpusecase.HardDeleteFileUseCase
	SubmissionStatusUC *mpusecase.GetSubmissionStatusUseCase
	DownloadURLFunc    func(fileKey string) (string, error)
}

// NewMonthlyPlanHandler creates a handler wired to the given use cases.
func NewMonthlyPlanHandler(deps MonthlyPlanHandlerDeps) *MonthlyPlanHandler {
	return &MonthlyPlanHandler{
		getSettingsUC:      deps.GetSettingsUC,
		updateSettingsUC:   deps.UpdateSettingsUC,
		ensurePeriodUC:     deps.EnsurePeriodUC,
		presignUploadUC:    deps.PresignUploadUC,
		confirmUploadUC:    deps.ConfirmUploadUC,
		listFilesUC:        deps.ListFilesUC,
		getFileUC:          deps.GetFileUC,
		softDeleteFileUC:   deps.SoftDeleteFileUC,
		restoreFileUC:      deps.RestoreFileUC,
		hardDeleteFileUC:   deps.HardDeleteFileUC,
		submissionStatusUC: deps.SubmissionStatusUC,
		downloadURLFunc:    deps.DownloadURLFunc,
	}
}

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
		UserID: int64(uid),
		Role:   role,
		TeamID: teamID,
	}, true
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

// mapError maps domain errors to HTTP status codes.
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

func (h *MonthlyPlanHandler) GetOrCreatePeriod(c *gin.Context) {
	year, err := strconv.Atoi(c.Param("year"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("BAD_REQUEST", "Invalid year")})
		return
	}
	month, err := strconv.Atoi(c.Param("month"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("BAD_REQUEST", "Invalid month")})
		return
	}

	out, err := h.ensurePeriodUC.Execute(c.Request.Context(), mpusecase.EnsurePeriodInput{Year: year, Month: month})
	if err != nil {
		c.JSON(mapMPError(err), dto.StandardResponse{Success: false, Error: errResp("ERROR", err.Error())})
		return
	}

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data: dto.MonthlyPlanResponse{
			ID:        out.Period.ID,
			Year:      out.Period.Year,
			Month:     out.Period.Month,
			IsLocked:  out.Period.IsLocked,
			CreatedAt: out.Period.CreatedAt.Format(time.RFC3339),
		},
	})
}

// ─── File Endpoints ────────────────────────────────────────────────────────

func (h *MonthlyPlanHandler) ListFiles(c *gin.Context) {
	year, err := strconv.Atoi(c.Param("year"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("BAD_REQUEST", "Invalid year")})
		return
	}
	month, err := strconv.Atoi(c.Param("month"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("BAD_REQUEST", "Invalid month")})
		return
	}

	actor, ok := actorFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.StandardResponse{Success: false, Error: errResp("UNAUTHORIZED", "Unauthorized")})
		return
	}

	periodOut, err := h.ensurePeriodUC.Execute(c.Request.Context(), mpusecase.EnsurePeriodInput{Year: year, Month: month})
	if err != nil {
		c.JSON(mapMPError(err), dto.StandardResponse{Success: false, Error: errResp("ERROR", err.Error())})
		return
	}

	var filterTeamID int64
	if rawTeam := c.Query("teamId"); rawTeam != "" {
		filterTeamID, _ = strconv.ParseInt(rawTeam, 10, 64)
	}
	search := c.Query("search")

	out, err := h.listFilesUC.Execute(c.Request.Context(), mpusecase.ListFilesInput{
		Actor:  actor,
		PlanID: periodOut.Period.ID,
		TeamID: filterTeamID,
		Search: search,
	})
	if err != nil {
		c.JSON(mapMPError(err), dto.StandardResponse{Success: false, Error: errResp("ERROR", err.Error())})
		return
	}

	files := make([]dto.PlanFileResponse, len(out.Files))
	for i, f := range out.Files {
		files[i] = toPlanFileResponse(f)
	}

	c.JSON(http.StatusOK, dto.StandardResponse{Success: true, Data: files})
}

func (h *MonthlyPlanHandler) PresignUpload(c *gin.Context) {
	year, err := strconv.Atoi(c.Param("year"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("BAD_REQUEST", "Invalid year")})
		return
	}
	month, err := strconv.Atoi(c.Param("month"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("BAD_REQUEST", "Invalid month")})
		return
	}

	actor, ok := actorFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.StandardResponse{Success: false, Error: errResp("UNAUTHORIZED", "Unauthorized")})
		return
	}

	periodOut, err := h.ensurePeriodUC.Execute(c.Request.Context(), mpusecase.EnsurePeriodInput{Year: year, Month: month})
	if err != nil {
		c.JSON(mapMPError(err), dto.StandardResponse{Success: false, Error: errResp("ERROR", err.Error())})
		return
	}

	var req dto.PresignPlanFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("BAD_REQUEST", "fileName and fileType are required")})
		return
	}

	isMaster := c.Query("isMasterPlan") == "true"
	var teamOverride *int64
	if rawTeam := c.Query("teamId"); rawTeam != "" {
		if tid, e := strconv.ParseInt(rawTeam, 10, 64); e == nil {
			teamOverride = &tid
		}
	}

	out, err := h.presignUploadUC.Execute(c.Request.Context(), mpusecase.PresignUploadInput{
		Actor:    actor,
		PlanID:   periodOut.Period.ID,
		FileName: req.FileName,
		FileType: req.FileType,
		IsMaster: isMaster,
		TeamID:   teamOverride,
	})
	if err != nil {
		c.JSON(mapMPError(err), dto.StandardResponse{Success: false, Error: errResp("ERROR", err.Error())})
		return
	}

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data: dto.PresignedURLResponse{
			UploadURL: out.UploadURL,
			FileURL:   out.FileURL,
			FileKey:   out.FileKey,
		},
	})
}

func (h *MonthlyPlanHandler) ConfirmUpload(c *gin.Context) {
	year, err := strconv.Atoi(c.Param("year"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("BAD_REQUEST", "Invalid year")})
		return
	}
	month, err := strconv.Atoi(c.Param("month"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("BAD_REQUEST", "Invalid month")})
		return
	}

	actor, ok := actorFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.StandardResponse{Success: false, Error: errResp("UNAUTHORIZED", "Unauthorized")})
		return
	}

	periodOut, err := h.ensurePeriodUC.Execute(c.Request.Context(), mpusecase.EnsurePeriodInput{Year: year, Month: month})
	if err != nil {
		c.JSON(mapMPError(err), dto.StandardResponse{Success: false, Error: errResp("ERROR", err.Error())})
		return
	}

	var req dto.ConfirmPlanFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("BAD_REQUEST", "Invalid request body")})
		return
	}

	out, err := h.confirmUploadUC.Execute(c.Request.Context(), mpusecase.ConfirmUploadInput{
		Actor:         actor,
		PlanID:        periodOut.Period.ID,
		FileKey:       req.FileKey,
		FileURL:       req.FileURL,
		FileName:      req.FileName,
		FileSizeBytes: req.FileSizeBytes,
		Description:   req.Description,
		IsMasterPlan:  req.IsMasterPlan,
		TeamID:        req.TeamID,
	})
	if err != nil {
		c.JSON(mapMPError(err), dto.StandardResponse{Success: false, Error: errResp("ERROR", err.Error())})
		return
	}

	c.JSON(http.StatusCreated, dto.StandardResponse{Success: true, Data: toPlanFileResponse(out.File)})
}

func (h *MonthlyPlanHandler) SoftDeleteFile(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("BAD_REQUEST", "Invalid file ID")})
		return
	}

	actor, ok := actorFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.StandardResponse{Success: false, Error: errResp("UNAUTHORIZED", "Unauthorized")})
		return
	}

	_, err = h.softDeleteFileUC.Execute(c.Request.Context(), mpusecase.SoftDeleteFileInput{Actor: actor, FileID: id})
	if err != nil {
		c.JSON(mapMPError(err), dto.StandardResponse{Success: false, Error: errResp("ERROR", err.Error())})
		return
	}

	c.JSON(http.StatusOK, dto.StandardResponse{Success: true, Data: "File deleted"})
}

func (h *MonthlyPlanHandler) HardDeleteFile(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("BAD_REQUEST", "Invalid file ID")})
		return
	}

	actor, ok := actorFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.StandardResponse{Success: false, Error: errResp("UNAUTHORIZED", "Unauthorized")})
		return
	}

	_, err = h.hardDeleteFileUC.Execute(c.Request.Context(), mpusecase.HardDeleteFileInput{Actor: actor, FileID: id})
	if err != nil {
		c.JSON(mapMPError(err), dto.StandardResponse{Success: false, Error: errResp("ERROR", err.Error())})
		return
	}

	c.JSON(http.StatusOK, dto.StandardResponse{Success: true, Data: "File permanently deleted"})
}

func (h *MonthlyPlanHandler) RestoreFile(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("BAD_REQUEST", "Invalid file ID")})
		return
	}

	actor, ok := actorFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.StandardResponse{Success: false, Error: errResp("UNAUTHORIZED", "Unauthorized")})
		return
	}

	out, err := h.restoreFileUC.Execute(c.Request.Context(), mpusecase.RestoreFileInput{Actor: actor, FileID: id})
	if err != nil {
		c.JSON(mapMPError(err), dto.StandardResponse{Success: false, Error: errResp("ERROR", err.Error())})
		return
	}

	c.JSON(http.StatusOK, dto.StandardResponse{Success: true, Data: toPlanFileResponse(out.File)})
}

func (h *MonthlyPlanHandler) GetDownloadURL(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("BAD_REQUEST", "Invalid file ID")})
		return
	}

	actor, ok := actorFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.StandardResponse{Success: false, Error: errResp("UNAUTHORIZED", "Unauthorized")})
		return
	}

	out, err := h.getFileUC.Execute(c.Request.Context(), mpusecase.GetFileInput{Actor: actor, FileID: id})
	if err != nil {
		c.JSON(mapMPError(err), dto.StandardResponse{Success: false, Error: errResp("ERROR", err.Error())})
		return
	}

	if h.downloadURLFunc == nil {
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{Success: false, Error: errResp("ERROR", "Download not configured")})
		return
	}

	downloadURL, err := h.downloadURLFunc(out.File.FileKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{Success: false, Error: errResp("ERROR", "Failed to generate download URL")})
		return
	}

	c.JSON(http.StatusOK, dto.StandardResponse{Success: true, Data: downloadURL})
}

func (h *MonthlyPlanHandler) GetSubmissionStatus(c *gin.Context) {
	year, err := strconv.Atoi(c.Param("year"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("BAD_REQUEST", "Invalid year")})
		return
	}
	month, err := strconv.Atoi(c.Param("month"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("BAD_REQUEST", "Invalid month")})
		return
	}

	actor, ok := actorFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.StandardResponse{Success: false, Error: errResp("UNAUTHORIZED", "Unauthorized")})
		return
	}

	periodOut, err := h.ensurePeriodUC.Execute(c.Request.Context(), mpusecase.EnsurePeriodInput{Year: year, Month: month})
	if err != nil {
		c.JSON(mapMPError(err), dto.StandardResponse{Success: false, Error: errResp("ERROR", err.Error())})
		return
	}

	out, err := h.submissionStatusUC.Execute(c.Request.Context(), mpusecase.GetSubmissionStatusInput{Actor: actor, PlanID: periodOut.Period.ID})
	if err != nil {
		c.JSON(mapMPError(err), dto.StandardResponse{Success: false, Error: errResp("ERROR", err.Error())})
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

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data: dto.SubmissionStatusResponse{
			Deadline: out.Deadline,
			Teams:    teams,
		},
	})
}

// ─── Settings Endpoints ────────────────────────────────────────────────────

func (h *MonthlyPlanHandler) GetSettings(c *gin.Context) {
	out, err := h.getSettingsUC.Execute(c.Request.Context(), mpusecase.GetSettingsInput{})
	if err != nil {
		c.JSON(mapMPError(err), dto.StandardResponse{Success: false, Error: errResp("ERROR", err.Error())})
		return
	}

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data: dto.MonthlyPlanSettingsResponse{
			LockDay:                 out.Settings.LockDay,
			AutoCreateDay:           out.Settings.AutoCreateDay,
			AutoCreateTarget:        out.Settings.AutoCreateTarget,
			AllowedFileTypes:        out.Settings.AllowedFileTypes,
			MaxFileSizeMB:           out.Settings.MaxFileSizeMB,
			ReminderStartDay:        out.Settings.ReminderStartDay,
			AdminCanUploadAfterLock: out.Settings.AdminCanUploadAfterLock,
		},
	})
}

func (h *MonthlyPlanHandler) UpdateSettings(c *gin.Context) {
	actor, ok := actorFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.StandardResponse{Success: false, Error: errResp("UNAUTHORIZED", "Unauthorized")})
		return
	}

	var req dto.UpdateMonthlyPlanSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{Success: false, Error: errResp("BAD_REQUEST", "Invalid request body")})
		return
	}

	out, err := h.updateSettingsUC.Execute(c.Request.Context(), mpusecase.UpdateSettingsInput{
		Actor:                   actor,
		LockDay:                 req.LockDay,
		AutoCreateDay:           req.AutoCreateDay,
		AutoCreateTarget:        req.AutoCreateTarget,
		AllowedFileTypes:        req.AllowedFileTypes,
		MaxFileSizeMB:           req.MaxFileSizeMB,
		ReminderStartDay:        req.ReminderStartDay,
		AdminCanUploadAfterLock: req.AdminCanUploadAfterLock,
	})
	if err != nil {
		c.JSON(mapMPError(err), dto.StandardResponse{Success: false, Error: errResp("ERROR", err.Error())})
		return
	}

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data: dto.MonthlyPlanSettingsResponse{
			LockDay:                 out.Settings.LockDay,
			AutoCreateDay:           out.Settings.AutoCreateDay,
			AutoCreateTarget:        out.Settings.AutoCreateTarget,
			AllowedFileTypes:        out.Settings.AllowedFileTypes,
			MaxFileSizeMB:           out.Settings.MaxFileSizeMB,
			ReminderStartDay:        out.Settings.ReminderStartDay,
			AdminCanUploadAfterLock: out.Settings.AdminCanUploadAfterLock,
		},
	})
}

// ─── Entity → DTO mapper ───────────────────────────────────────────────────

func toPlanFileResponse(f mpdomain.PlanFileEntity) dto.PlanFileResponse {
	resp := dto.PlanFileResponse{
		ID:            f.ID,
		MonthlyPlanID: f.MonthlyPlanID,
		TeamID:        f.TeamID,
		UploadedByID:  f.UploadedByID,
		FileKey:       f.FileKey,
		FileURL:       f.FileURL,
		FileName:      f.FileName,
		FileSizeBytes: f.FileSizeBytes,
		Description:   f.Description,
		IsMasterPlan:  f.IsMasterPlan,
		IsDeleted:     f.IsDeleted,
		CreatedAt:     f.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     f.UpdatedAt.Format(time.RFC3339),
	}
	if f.DeletedAt != nil {
		s := f.DeletedAt.Format(time.RFC3339)
		resp.DeletedAt = &s
	}
	if f.TeamName != nil {
		tid := int64(0)
		if f.TeamID != nil {
			tid = *f.TeamID
		}
		resp.Team = &dto.TeamNested{ID: tid, Name: *f.TeamName}
	}
	if f.UploaderName != nil {
		resp.UploadedBy = &dto.PlanFileUploaderNested{
			ID:       uint(f.UploadedByID),
			Username: *f.UploaderName,
		}
	}
	return resp
}
