package v1

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"backend-hotlines3/internal/config"
	"backend-hotlines3/internal/dto"
	"backend-hotlines3/internal/models"
	"backend-hotlines3/pkg/s3"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MonthlyPlanHandler handles all monthly plan file management endpoints.
type MonthlyPlanHandler struct {
	db       *gorm.DB
	r2Client *s3.R2Client
}

// NewMonthlyPlanHandler creates a new MonthlyPlanHandler.
// r2Client may be nil when R2 is not configured (upload endpoints will return error).
func NewMonthlyPlanHandler(db *gorm.DB, cfg *config.Config) (*MonthlyPlanHandler, error) {
	r2Client, err := s3.NewR2Client(s3.R2Config{
		AccountID:       cfg.Cloudflare.R2.AccountID,
		AccessKeyID:     cfg.Cloudflare.R2.AccessKeyID,
		SecretAccessKey: cfg.Cloudflare.R2.SecretAccessKey,
		BucketName:      cfg.Cloudflare.R2.BucketName,
		PublicURL:       cfg.Cloudflare.R2.PublicURL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to init R2 client for monthly plan handler: %w", err)
	}
	return &MonthlyPlanHandler{db: db, r2Client: r2Client}, nil
}

// ─── Helpers ───────────────────────────────────────────────────────────────

func convertPlanToResponse(p *models.MonthlyPlan) dto.MonthlyPlanResponse {
	return dto.MonthlyPlanResponse{
		ID:        p.ID,
		Year:      p.Year,
		Month:     p.Month,
		IsLocked:  p.IsLocked,
		CreatedAt: p.CreatedAt.Format(time.RFC3339),
	}
}

func convertPlanFileToResponse(f *models.PlanFile) dto.PlanFileResponse {
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
	if f.Team != nil {
		resp.Team = &dto.TeamNested{ID: f.Team.ID, Name: f.Team.Name}
	}
	if f.UploadedBy != nil {
		resp.UploadedBy = &dto.PlanFileUploaderNested{
			ID:       f.UploadedBy.ID,
			Username: f.UploadedBy.Username,
		}
	}
	return resp
}

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

// parseYearMonth extracts and validates :year/:month from URL params.
func parseYearMonth(c *gin.Context) (year, month int, ok bool) {
	y, err1 := strconv.Atoi(c.Param("year"))
	m, err2 := strconv.Atoi(c.Param("month"))
	if err1 != nil || err2 != nil || y < 2020 || y > 2100 || m < 1 || m > 12 {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INVALID_PERIOD",
				Message: "year และ month ไม่ถูกต้อง",
			},
		})
		return 0, 0, false
	}
	return y, m, true
}

// getOrFindPeriod finds an existing MonthlyPlan record; returns 404 if absent.
func (h *MonthlyPlanHandler) findPeriod(ctx context.Context, year, month int) (*models.MonthlyPlan, error) {
	var plan models.MonthlyPlan
	err := h.db.WithContext(ctx).
		Where(`"year" = ? AND "month" = ?`, year, month).
		First(&plan).Error
	return &plan, err
}

// getOrCreateSettings reads the single settings row from DB.
// If no row exists (first run), it creates one with sensible defaults.
func (h *MonthlyPlanHandler) getOrCreateSettings(ctx context.Context) (*models.MonthlyPlanSetting, error) {
	var s models.MonthlyPlanSetting
	if err := h.db.WithContext(ctx).First(&s).Error; err != nil {
		// Create default row
		s = models.MonthlyPlanSetting{
			LockDay:                 23,
			AutoCreateDay:           1,
			AutoCreateTarget:        "next_month",
			AllowedFileTypes:        models.StringArray{"application/pdf"},
			ReminderStartDay:        20,
			AdminCanUploadAfterLock: true,
			UpdatedAt:               time.Now(),
		}
		if err := h.db.WithContext(ctx).Create(&s).Error; err != nil {
			return nil, fmt.Errorf("failed to create default settings: %w", err)
		}
	}
	return &s, nil
}

// saveSettings persists an updated MonthlyPlanSetting row.
func (h *MonthlyPlanHandler) saveSettings(ctx context.Context, s *models.MonthlyPlanSetting) error {
	s.UpdatedAt = time.Now()
	return h.db.WithContext(ctx).Save(s).Error
}

// ─── GET /v1/monthly-plans/:year/:month ────────────────────────────────────
// GetOrCreatePeriod returns the period, creating it if it doesn't exist yet.

func (h *MonthlyPlanHandler) GetOrCreatePeriod(c *gin.Context) {
	ctx := c.Request.Context()
	year, month, ok := parseYearMonth(c)
	if !ok {
		return
	}

	var plan models.MonthlyPlan
	result := h.db.WithContext(ctx).
		Where(`"year" = ? AND "month" = ?`, year, month).
		FirstOrCreate(&plan, models.MonthlyPlan{Year: year, Month: month})

	if result.Error != nil {
		log.Printf("GetOrCreatePeriod error: %v", result.Error)
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "INTERNAL_ERROR", Message: result.Error.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data:    convertPlanToResponse(&plan),
	})
}

// ─── GET /v1/monthly-plans/:year/:month/files ──────────────────────────────
// ListFiles returns all files for the period, grouped by team.
// Query params: search (partial filename), teamId (filter by team)

func (h *MonthlyPlanHandler) ListFiles(c *gin.Context) {
	ctx := c.Request.Context()
	year, month, ok := parseYearMonth(c)
	if !ok {
		return
	}

	plan, err := h.findPeriod(ctx, year, month)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "NOT_FOUND", Message: "ไม่พบข้อมูลเดือนนี้"},
		})
		return
	}

	search := strings.TrimSpace(c.Query("search"))

	var filterTeamID int64
	if rawTeam := c.Query("teamId"); rawTeam != "" {
		filterTeamID, _ = strconv.ParseInt(rawTeam, 10, 64)
	}

	query := h.db.WithContext(ctx).
		Model(&models.PlanFile{}).
		Scopes(models.PlanFileByPlan(plan.ID)).
		Scopes(models.PlanFileNotDeleted).
		Scopes(models.PlanFileByTeam(filterTeamID)).
		Preload("Team").
		Preload("UploadedBy").
		Order(`"isMasterPlan" DESC, "createdAt" ASC`)

	if search != "" {
		query = query.Where(`"fileName" ILIKE ?`, "%"+search+"%")
	}

	var files []models.PlanFile
	if err := query.Find(&files).Error; err != nil {
		log.Printf("ListFiles error: %v", err)
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "INTERNAL_ERROR", Message: err.Error()},
		})
		return
	}

	resp := make([]dto.PlanFileResponse, 0, len(files))
	for i := range files {
		resp = append(resp, convertPlanFileToResponse(&files[i]))
	}

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data:    resp,
		Meta:    &dto.Meta{Total: int64(len(resp))},
	})
}

// ─── POST /v1/monthly-plans/:year/:month/files/presign ─────────────────────
// PresignUpload generates a presigned PUT URL for uploading a PDF to R2.

func (h *MonthlyPlanHandler) PresignUpload(c *gin.Context) {
	ctx := c.Request.Context()
	year, month, ok := parseYearMonth(c)
	if !ok {
		return
	}

	userID, role, authOk := getAuthInfo(c)
	if !authOk {
		c.JSON(http.StatusUnauthorized, dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "UNAUTHORIZED", Message: "กรุณาเข้าสู่ระบบ"},
		})
		return
	}

	// Load settings to check allowed file types and lock rules
	settings, err := h.getOrCreateSettings(ctx)
	if err != nil {
		log.Printf("PresignUpload: failed to load settings: %v", err)
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "SETTINGS_ERROR", Message: "ไม่สามารถโหลด settings ได้"},
		})
		return
	}

	// Check lock status (non-admin blocked after lock)
	plan, err := h.findPeriod(ctx, year, month)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "NOT_FOUND", Message: "ไม่พบข้อมูลเดือนนี้"},
		})
		return
	}

	if plan.IsLocked && role != "admin" {
		c.JSON(http.StatusForbidden, dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "PERIOD_LOCKED", Message: "เลย deadline แล้ว ไม่สามารถอัพโหลดได้"},
		})
		return
	}

	var req dto.PresignPlanFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "VALIDATION_ERROR", Message: err.Error()},
		})
		return
	}

	// Validate file type
	allowed := false
	for _, ft := range settings.AllowedFileTypes {
		if req.FileType == ft {
			allowed = true
			break
		}
	}
	if !allowed {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INVALID_FILE_TYPE",
				Message: fmt.Sprintf("รองรับเฉพาะ %v เท่านั้น", settings.AllowedFileTypes),
			},
		})
		return
	}

	_ = userID // used for permission context, actual upload is to R2 directly

	// Build unique object key: plans/{year}/{month}/{uuid}.pdf
	fileKey := fmt.Sprintf("plans/%d/%d/%s.pdf", year, month, uuid.New().String())

	ctxTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	result, err := h.r2Client.GeneratePresignedURL(ctxTimeout, fileKey, req.FileType, 15*time.Minute)
	if err != nil {
		log.Printf("PresignUpload: failed to generate URL for key %s: %v", fileKey, err)
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "UPLOAD_ERROR", Message: "ไม่สามารถสร้าง upload URL ได้"},
		})
		return
	}

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data: dto.PresignedURLResponse{
			UploadURL: result.UploadURL,
			FileURL:   result.FileURL,
			FileKey:   result.FileKey,
		},
	})
}

// ─── POST /v1/monthly-plans/:year/:month/files ─────────────────────────────
// ConfirmUpload saves the file metadata after the client has PUT the file to R2.

func (h *MonthlyPlanHandler) ConfirmUpload(c *gin.Context) {
	ctx := c.Request.Context()
	year, month, ok := parseYearMonth(c)
	if !ok {
		return
	}

	userID, role, authOk := getAuthInfo(c)
	if !authOk {
		c.JSON(http.StatusUnauthorized, dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "UNAUTHORIZED", Message: "กรุณาเข้าสู่ระบบ"},
		})
		return
	}

	plan, err := h.findPeriod(ctx, year, month)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "NOT_FOUND", Message: "ไม่พบข้อมูลเดือนนี้"},
		})
		return
	}

	if plan.IsLocked && role != "admin" {
		c.JSON(http.StatusForbidden, dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "PERIOD_LOCKED", Message: "เลย deadline แล้ว ไม่สามารถอัพโหลดได้"},
		})
		return
	}

	var req dto.ConfirmPlanFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "VALIDATION_ERROR", Message: err.Error()},
		})
		return
	}

	// เฉพาะ admin เท่านั้นที่อัพโหลด master plan ได้
	if req.IsMasterPlan && role != "admin" {
		c.JSON(http.StatusForbidden, dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "FORBIDDEN", Message: "เฉพาะ Admin เท่านั้นที่อัพโหลดแผนรวมได้"},
		})
		return
	}

	// ─── Determine effective teamId ───────────────────────────────────────────
	// Priority:
	//   1. isMasterPlan = true  → teamId = nil (master plan ไม่สังกัดทีม)
	//   2. req.TeamID ระบุมา + role = admin  → ใช้ teamId ที่ระบุ (upload แทนทีม)
	//   3. req.TeamID ระบุมา + role ≠ admin  → 403 Forbidden
	//   4. req.TeamID ไม่ถูกส่งมา  → ใช้ teamId ของ user ที่ login (พฤติกรรมเดิม)
	var effectiveTeamID *int64

	if req.IsMasterPlan {
		// master plan ไม่มีทีม — teamId = nil
		effectiveTeamID = nil
	} else if req.TeamID != nil {
		// admin ส่ง teamId มาเพื่ออัพโหลดแทนทีมนั้น
		if role != "admin" {
			c.JSON(http.StatusForbidden, dto.StandardResponse{
				Success: false,
				Error:   &dto.ErrorInfo{Code: "FORBIDDEN", Message: "เฉพาะ Admin เท่านั้นที่อัพโหลดแทนทีมอื่นได้"},
			})
			return
		}
		effectiveTeamID = req.TeamID
	} else {
		// ใช้ teamId ของ user ที่ login (พฤติกรรมเดิม)
		var user models.User
		if err := h.db.WithContext(ctx).First(&user, userID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, dto.StandardResponse{
				Success: false,
				Error:   &dto.ErrorInfo{Code: "INTERNAL_ERROR", Message: "ไม่พบข้อมูลผู้ใช้"},
			})
			return
		}
		effectiveTeamID = user.TeamID
	}

	now := time.Now()

	// ─── Create new PlanFile ──────────────────────────────────────────────────
	// หมายเหตุ: isMasterPlan = true สามารถมีได้หลาย records ต่อเดือน (no soft-delete of old ones)
	// fileName ถูก save จาก req.FileName ตรงๆ เสมอ
	planFile := models.PlanFile{
		MonthlyPlanID: plan.ID,
		TeamID:        effectiveTeamID,
		UploadedByID:  int64(userID),
		FileKey:       req.FileKey,
		FileURL:       req.FileURL,
		FileName:      req.FileName,
		FileSizeBytes: req.FileSizeBytes,
		Description:   req.Description,
		IsMasterPlan:  req.IsMasterPlan,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := h.db.WithContext(ctx).Create(&planFile).Error; err != nil {
		log.Printf("ConfirmUpload: create PlanFile error: %v", err)
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "INTERNAL_ERROR", Message: err.Error()},
		})
		return
	}

	// Write file size log (F-13)
	sizeLog := models.FileSizeLog{
		PlanFileID:    planFile.ID,
		UploadedByID:  int64(userID),
		FileSizeBytes: req.FileSizeBytes,
		CreatedAt:     now,
	}
	if err := h.db.WithContext(ctx).Create(&sizeLog).Error; err != nil {
		log.Printf("ConfirmUpload: failed to create FileSizeLog: %v", err)
		// Non-fatal — don't fail the request
	}

	// Reload with relations
	h.db.WithContext(ctx).Preload("Team").Preload("UploadedBy").First(&planFile, planFile.ID)

	c.JSON(http.StatusCreated, dto.StandardResponse{
		Success: true,
		Data:    convertPlanFileToResponse(&planFile),
	})
}

// ─── DELETE /v1/monthly-plans/files/:id ────────────────────────────────────
// SoftDeleteFile marks a file as deleted. Own team members or admin only.

func (h *MonthlyPlanHandler) SoftDeleteFile(c *gin.Context) {
	ctx := c.Request.Context()
	fileID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "INVALID_ID", Message: "ID ไม่ถูกต้อง"},
		})
		return
	}

	userID, role, authOk := getAuthInfo(c)
	if !authOk {
		c.JSON(http.StatusUnauthorized, dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "UNAUTHORIZED", Message: "กรุณาเข้าสู่ระบบ"},
		})
		return
	}

	var file models.PlanFile
	if err := h.db.WithContext(ctx).Preload("Team").First(&file, fileID).Error; err != nil {
		c.JSON(http.StatusNotFound, dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "NOT_FOUND", Message: "ไม่พบไฟล์"},
		})
		return
	}

	// Permission: must be admin OR the file belongs to user's team
	if role != "admin" {
		var user models.User
		h.db.WithContext(ctx).First(&user, userID)
		if user.TeamID == nil || file.TeamID == nil || *user.TeamID != *file.TeamID {
			c.JSON(http.StatusForbidden, dto.StandardResponse{
				Success: false,
				Error:   &dto.ErrorInfo{Code: "FORBIDDEN", Message: "ไม่มีสิทธิ์ลบไฟล์ของทีมอื่น"},
			})
			return
		}
	}

	now := time.Now()
	if err := h.db.WithContext(ctx).Model(&file).Updates(map[string]interface{}{
		"isDeleted": true,
		"deletedAt": now,
		"updatedAt": now,
	}).Error; err != nil {
		log.Printf("SoftDeleteFile error: %v", err)
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "INTERNAL_ERROR", Message: err.Error()},
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// ─── DELETE /v1/monthly-plans/files/:id/permanent ──────────────────────────
// HardDeleteFile permanently removes the file record and deletes from R2. Admin only.

func (h *MonthlyPlanHandler) HardDeleteFile(c *gin.Context) {
	ctx := c.Request.Context()
	fileID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "INVALID_ID", Message: "ID ไม่ถูกต้อง"},
		})
		return
	}

	var file models.PlanFile
	if err := h.db.WithContext(ctx).First(&file, fileID).Error; err != nil {
		c.JSON(http.StatusNotFound, dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "NOT_FOUND", Message: "ไม่พบไฟล์"},
		})
		return
	}

	// Delete related size logs first (FK constraint)
	h.db.WithContext(ctx).Where(`"planFileId" = ?`, file.ID).Delete(&models.FileSizeLog{})

	// Delete from DB
	if err := h.db.WithContext(ctx).Delete(&file).Error; err != nil {
		log.Printf("HardDeleteFile DB error: %v", err)
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "INTERNAL_ERROR", Message: err.Error()},
		})
		return
	}

	// Delete from R2 (best-effort)
	ctxR2, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := h.r2Client.DeleteObject(ctxR2, file.FileKey); err != nil {
		log.Printf("HardDeleteFile: R2 delete warning for key %s: %v", file.FileKey, err)
	}

	c.Status(http.StatusNoContent)
}

// ─── POST /v1/monthly-plans/files/:id/restore ──────────────────────────────
// RestoreFile un-deletes a soft-deleted file. Admin only.

func (h *MonthlyPlanHandler) RestoreFile(c *gin.Context) {
	ctx := c.Request.Context()
	fileID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "INVALID_ID", Message: "ID ไม่ถูกต้อง"},
		})
		return
	}

	var file models.PlanFile
	if err := h.db.WithContext(ctx).First(&file, fileID).Error; err != nil {
		c.JSON(http.StatusNotFound, dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "NOT_FOUND", Message: "ไม่พบไฟล์"},
		})
		return
	}

	if !file.IsDeleted {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "NOT_DELETED", Message: "ไฟล์นี้ไม่ได้ถูกลบ"},
		})
		return
	}

	now := time.Now()
	if err := h.db.WithContext(ctx).Model(&file).Updates(map[string]interface{}{
		"isDeleted": false,
		"deletedAt": nil,
		"updatedAt": now,
	}).Error; err != nil {
		log.Printf("RestoreFile error: %v", err)
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "INTERNAL_ERROR", Message: err.Error()},
		})
		return
	}

	h.db.WithContext(ctx).Preload("Team").Preload("UploadedBy").First(&file, file.ID)

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data:    convertPlanFileToResponse(&file),
	})
}

// ─── GET /v1/monthly-plans/files/:id/download ──────────────────────────────
// GetDownloadURL returns a presigned GET URL for a file.
// Permission:
//   - Master plan → everyone
//   - Team file → own team or admin only (other teams get 403)

func (h *MonthlyPlanHandler) GetDownloadURL(c *gin.Context) {
	ctx := c.Request.Context()
	fileID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "INVALID_ID", Message: "ID ไม่ถูกต้อง"},
		})
		return
	}

	userID, role, authOk := getAuthInfo(c)
	if !authOk {
		c.JSON(http.StatusUnauthorized, dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "UNAUTHORIZED", Message: "กรุณาเข้าสู่ระบบ"},
		})
		return
	}

	var file models.PlanFile
	if err := h.db.WithContext(ctx).First(&file, fileID).Error; err != nil {
		c.JSON(http.StatusNotFound, dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "NOT_FOUND", Message: "ไม่พบไฟล์"},
		})
		return
	}

	// Permission check for non-master plan files
	if !file.IsMasterPlan && role != "admin" {
		var user models.User
		h.db.WithContext(ctx).First(&user, userID)
		if user.TeamID == nil || file.TeamID == nil || *user.TeamID != *file.TeamID {
			c.JSON(http.StatusForbidden, dto.StandardResponse{
				Success: false,
				Error:   &dto.ErrorInfo{Code: "FORBIDDEN", Message: "ไม่มีสิทธิ์ดาวน์โหลดไฟล์ของทีมอื่น"},
			})
			return
		}
	}

	// Generate presigned GET URL (valid 60 minutes)
	ctxTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	presignResult, err := h.r2Client.GeneratePresignedGetURL(ctxTimeout, file.FileKey, 60*time.Minute)
	if err != nil {
		log.Printf("GetDownloadURL: presign error for %s: %v", file.FileKey, err)
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "DOWNLOAD_ERROR", Message: "ไม่สามารถสร้าง download URL ได้"},
		})
		return
	}

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data:    map[string]string{"downloadUrl": presignResult},
	})
}

// ─── GET /v1/monthly-plans/:year/:month/status ─────────────────────────────
// GetSubmissionStatus returns 3-state submission status for every team.

func (h *MonthlyPlanHandler) GetSubmissionStatus(c *gin.Context) {
	ctx := c.Request.Context()
	year, month, ok := parseYearMonth(c)
	if !ok {
		return
	}

	settings, err := h.getOrCreateSettings(ctx)
	if err != nil {
		log.Printf("GetSubmissionStatus: failed to load settings: %v", err)
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "SETTINGS_ERROR", Message: "ไม่สามารถโหลด settings ได้"},
		})
		return
	}

	plan, err := h.findPeriod(ctx, year, month)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "NOT_FOUND", Message: "ไม่พบข้อมูลเดือนนี้"},
		})
		return
	}

	// Deadline: day lockDay of the SAME year/month (plan belongs to the period month)
	deadline := time.Date(year, time.Month(month), settings.LockDay, 23, 59, 59, 0, time.UTC)
	now := time.Now().UTC()
	pastDeadline := now.After(deadline)

	// Fetch all teams
	var teams []models.Team
	if err := h.db.WithContext(ctx).Find(&teams).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "INTERNAL_ERROR", Message: err.Error()},
		})
		return
	}

	// Count non-deleted, non-master plan files per team for this period
	type teamCount struct {
		TeamID int64
		Count  int
	}
	var counts []teamCount
	h.db.WithContext(ctx).
		Model(&models.PlanFile{}).
		Select(`"teamId" as "TeamID", count(*) as "Count"`).
		Where(`"monthlyPlanId" = ? AND "isDeleted" = false AND "isMasterPlan" = false`, plan.ID).
		Group(`"teamId"`).
		Scan(&counts)

	countMap := make(map[int64]int)
	for _, tc := range counts {
		countMap[tc.TeamID] = tc.Count
	}

	teamStatuses := make([]dto.TeamSubmissionStatus, 0, len(teams))
	for _, t := range teams {
		cnt := countMap[t.ID]
		var status string
		switch {
		case cnt > 0:
			status = "submitted"
		case pastDeadline || plan.IsLocked:
			status = "missed"
		default:
			status = "pending"
		}
		teamStatuses = append(teamStatuses, dto.TeamSubmissionStatus{
			Team:      dto.TeamNested{ID: t.ID, Name: t.Name},
			Status:    status,
			FileCount: cnt,
		})
	}

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data: dto.SubmissionStatusResponse{
			Period:   convertPlanToResponse(plan),
			Deadline: deadline.Format("2006-01-02"),
			Teams:    teamStatuses,
		},
	})
}

// GetSettings returns the current DB settings. Admin only.
func (h *MonthlyPlanHandler) GetSettings(c *gin.Context) {
	ctx := c.Request.Context()
	s, err := h.getOrCreateSettings(ctx)
	if err != nil {
		log.Printf("GetSettings error: %v", err)
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "INTERNAL_ERROR", Message: "ไม่สามารถโหลด settings ได้"},
		})
		return
	}

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data: dto.MonthlyPlanSettingsResponse{
			LockDay:                 s.LockDay,
			AutoCreateDay:           s.AutoCreateDay,
			AutoCreateTarget:        s.AutoCreateTarget,
			AllowedFileTypes:        []string(s.AllowedFileTypes),
			MaxFileSizeMB:           s.MaxFileSizeMB,
			ReminderStartDay:        s.ReminderStartDay,
			AdminCanUploadAfterLock: s.AdminCanUploadAfterLock,
		},
	})
}

// UpdateSettings patches the settings row in DB. Admin only.
func (h *MonthlyPlanHandler) UpdateSettings(c *gin.Context) {
	ctx := c.Request.Context()
	var req dto.UpdateMonthlyPlanSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "VALIDATION_ERROR", Message: err.Error()},
		})
		return
	}

	s, err := h.getOrCreateSettings(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "SETTINGS_ERROR", Message: "ไม่สามารถโหลด settings ได้"},
		})
		return
	}

	// Patch only provided fields
	if req.LockDay != nil {
		s.LockDay = *req.LockDay
	}
	if req.AutoCreateDay != nil {
		s.AutoCreateDay = *req.AutoCreateDay
	}
	if req.AutoCreateTarget != nil {
		s.AutoCreateTarget = *req.AutoCreateTarget
	}
	if req.AllowedFileTypes != nil {
		s.AllowedFileTypes = models.StringArray(req.AllowedFileTypes)
	}
	if req.MaxFileSizeMB != nil {
		s.MaxFileSizeMB = req.MaxFileSizeMB
	}
	if req.ReminderStartDay != nil {
		s.ReminderStartDay = *req.ReminderStartDay
	}
	if req.AdminCanUploadAfterLock != nil {
		s.AdminCanUploadAfterLock = *req.AdminCanUploadAfterLock
	}

	if err := h.saveSettings(ctx, s); err != nil {
		log.Printf("UpdateSettings: save error: %v", err)
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error:   &dto.ErrorInfo{Code: "INTERNAL_ERROR", Message: "ไม่สามารถบันทึก settings ได้"},
		})
		return
	}

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data: dto.MonthlyPlanSettingsResponse{
			LockDay:                 s.LockDay,
			AutoCreateDay:           s.AutoCreateDay,
			AutoCreateTarget:        s.AutoCreateTarget,
			AllowedFileTypes:        []string(s.AllowedFileTypes),
			MaxFileSizeMB:           s.MaxFileSizeMB,
			ReminderStartDay:        s.ReminderStartDay,
			AdminCanUploadAfterLock: s.AdminCanUploadAfterLock,
		},
	})
}
