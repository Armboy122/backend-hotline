package gorm

import (
	"context"
	"fmt"
	"time"

	mpdomain "backend-hotlines3/internal/domain/monthlyplan"
	"backend-hotlines3/internal/models"
	"backend-hotlines3/internal/port/outbound/repository"

	"gorm.io/gorm"
)

// MonthlyPlanRepository implements repository.MonthlyPlanRepository via GORM.
type MonthlyPlanRepository struct {
	db *gorm.DB
}

func NewMonthlyPlanRepository(db *gorm.DB) *MonthlyPlanRepository {
	return &MonthlyPlanRepository{db: db}
}

// ── Period ───────────────────────────────────────────────────────────────────

func (r *MonthlyPlanRepository) FindPeriod(ctx context.Context, q repository.PeriodFindQuery) (*mpdomain.Entity, error) {
	var mp models.MonthlyPlan
	err := r.db.WithContext(ctx).
		Where("year = ? AND month = ?", q.Year, q.Month).
		First(&mp).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("find period: %w", err)
	}
	e := toPeriodEntity(mp)
	return &e, nil
}

func (r *MonthlyPlanRepository) FindOrCreatePeriod(ctx context.Context, year, month int) (*mpdomain.Entity, error) {
	var mp models.MonthlyPlan
	err := r.db.WithContext(ctx).
		Where("year = ? AND month = ?", year, month).
		Attrs(models.MonthlyPlan{Year: year, Month: month}).
		FirstOrCreate(&mp).Error
	if err != nil {
		return nil, fmt.Errorf("find or create period: %w", err)
	}
	e := toPeriodEntity(mp)
	return &e, nil
}

// ── Settings ─────────────────────────────────────────────────────────────────

func (r *MonthlyPlanRepository) GetOrCreateSettings(ctx context.Context) (*mpdomain.SettingsEntity, error) {
	var s models.MonthlyPlanSetting
	err := r.db.WithContext(ctx).FirstOrCreate(&s).Error
	if err != nil {
		return nil, fmt.Errorf("get or create settings: %w", err)
	}
	e := toSettingsEntity(s)
	return &e, nil
}

func (r *MonthlyPlanRepository) UpdateSettings(ctx context.Context, s *mpdomain.SettingsEntity) error {
	var model models.MonthlyPlanSetting
	if err := r.db.WithContext(ctx).First(&model).Error; err != nil {
		return fmt.Errorf("find settings for update: %w", err)
	}

	updates := map[string]interface{}{}
	if s.LockDay != 0 {
		updates[models.PlanFileCol.CreatedAt] = nil // placeholder, real fields below
	}
	// Build explicit update map
	updates = map[string]interface{}{
		"lockDay":                 s.LockDay,
		"autoCreateDay":           s.AutoCreateDay,
		"autoCreateTarget":        s.AutoCreateTarget,
		"allowedFileTypes":        s.AllowedFileTypes,
		"maxFileSizeMB":           s.MaxFileSizeMB,
		"reminderStartDay":        s.ReminderStartDay,
		"adminCanUploadAfterLock": s.AdminCanUploadAfterLock,
		"updatedAt":               time.Now(),
	}

	if err := r.db.WithContext(ctx).Model(&model).Updates(updates).Error; err != nil {
		return fmt.Errorf("update settings: %w", err)
	}
	return nil
}

// ── Plan Files ───────────────────────────────────────────────────────────────

func (r *MonthlyPlanRepository) ListFiles(ctx context.Context, q repository.PlanFileListQuery) ([]mpdomain.PlanFileEntity, error) {
	query := r.db.WithContext(ctx).
		Preload("Team").
		Preload("UploadedBy").
		Where(PlanFileColMonthlyPlanID()+" = ?", q.MonthlyPlanID).
		Scopes(models.PlanFileNotDeleted)

	if q.TeamID != 0 {
		query = query.Scopes(models.PlanFileByTeam(q.TeamID))
	}
	if q.Search != "" {
		query = query.Where("\"fileName\" ILIKE ?", "%"+q.Search+"%")
	}

	var files []models.PlanFile
	if err := query.Order(PlanFileColCreatedAt() + " DESC").Find(&files).Error; err != nil {
		return nil, fmt.Errorf("list files: %w", err)
	}

	result := make([]mpdomain.PlanFileEntity, len(files))
	for i, f := range files {
		result[i] = toPlanFileEntity(f)
	}
	return result, nil
}

func (r *MonthlyPlanRepository) GetFileByID(ctx context.Context, id int64) (*mpdomain.PlanFileEntity, error) {
	var f models.PlanFile
	err := r.db.WithContext(ctx).
		Preload("Team").
		Preload("UploadedBy").
		Where("id = ?", id).
		First(&f).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get file: %w", err)
	}
	e := toPlanFileEntity(f)
	return &e, nil
}

func (r *MonthlyPlanRepository) CreateFile(ctx context.Context, input repository.PlanFileCreateInput) (*mpdomain.PlanFileEntity, error) {
	now := time.Now()
	f := models.PlanFile{
		MonthlyPlanID: input.MonthlyPlanID,
		TeamID:        input.TeamID,
		UploadedByID:  input.UploadedByID,
		FileKey:       input.FileKey,
		FileURL:       input.FileURL,
		FileName:      input.FileName,
		FileSizeBytes: input.FileSizeBytes,
		Description:   input.Description,
		IsMasterPlan:  input.IsMasterPlan,
		IsDeleted:     false,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := r.db.WithContext(ctx).Create(&f).Error; err != nil {
		return nil, fmt.Errorf("create file: %w", err)
	}
	e := toPlanFileEntity(f)
	return &e, nil
}

func (r *MonthlyPlanRepository) SoftDeleteFile(ctx context.Context, input repository.PlanFileSoftDeleteInput) error {
	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&models.PlanFile{}).
		Where("id = ? AND \"isDeleted\" = false", input.ID).
		Updates(map[string]interface{}{
			"\"isDeleted\"": true,
			"\"deletedAt\"": now,
			"\"updatedAt\"": now,
		})
	if result.Error != nil {
		return fmt.Errorf("soft delete file: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("soft delete file: no rows affected (id=%d)", input.ID)
	}
	return nil
}

func (r *MonthlyPlanRepository) RestoreFile(ctx context.Context, input repository.PlanFileRestoreInput) (*mpdomain.PlanFileEntity, error) {
	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&models.PlanFile{}).
		Where("id = ? AND \"isDeleted\" = true", input.ID).
		Updates(map[string]interface{}{
			"\"isDeleted\"": false,
			"\"deletedAt\"": nil,
			"\"updatedAt\"": now,
		})
	if result.Error != nil {
		return nil, fmt.Errorf("restore file: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("restore file: no rows affected (id=%d)", input.ID)
	}
	return r.GetFileByID(ctx, input.ID)
}

func (r *MonthlyPlanRepository) HardDeleteFile(ctx context.Context, input repository.PlanFileHardDeleteInput) error {
	if err := r.db.WithContext(ctx).Delete(&models.PlanFile{}, input.ID).Error; err != nil {
		return fmt.Errorf("hard delete file: %w", err)
	}
	return nil
}

// ── File Size Log ────────────────────────────────────────────────────────────

func (r *MonthlyPlanRepository) CreateFileSizeLog(ctx context.Context, input repository.FileSizeLogCreateInput) error {
	log := models.FileSizeLog{
		PlanFileID:    input.PlanFileID,
		UploadedByID:  input.UploadedByID,
		FileSizeBytes: input.FileSizeBytes,
	}
	if err := r.db.WithContext(ctx).Create(&log).Error; err != nil {
		return fmt.Errorf("create file size log: %w", err)
	}
	return nil
}

// ── Submission Status ────────────────────────────────────────────────────────

func (r *MonthlyPlanRepository) CountFilesByTeam(ctx context.Context, q repository.SubmissionQuery) ([]repository.TeamCountRow, error) {
	var rows []repository.TeamCountRow
	err := r.db.WithContext(ctx).
		Model(&models.PlanFile{}).
		Select("\"teamId\" as team_id, COUNT(*) as count").
		Where(PlanFileColMonthlyPlanID()+" = ?", q.PlanID).
		Where(PlanFileColIsDeleted()+" = false").
		Where("\"teamId\" IS NOT NULL").
		Group("\"teamId\"").
		Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("count files by team: %w", err)
	}
	return rows, nil
}

func (r *MonthlyPlanRepository) ListTeams(ctx context.Context) ([]repository.TeamInfo, error) {
	var teams []models.Team
	if err := r.db.WithContext(ctx).Find(&teams).Error; err != nil {
		return nil, fmt.Errorf("list teams: %w", err)
	}
	result := make([]repository.TeamInfo, len(teams))
	for i, t := range teams {
		result[i] = repository.TeamInfo{ID: t.ID, Name: t.Name}
	}
	return result, nil
}

// ── Mappers ──────────────────────────────────────────────────────────────────

func toPeriodEntity(mp models.MonthlyPlan) mpdomain.Entity {
	return mpdomain.Entity{
		ID:        mp.ID,
		Year:      mp.Year,
		Month:     mp.Month,
		IsLocked:  mp.IsLocked,
		CreatedAt: mp.CreatedAt,
	}
}

func toSettingsEntity(s models.MonthlyPlanSetting) mpdomain.SettingsEntity {
	return mpdomain.SettingsEntity{
		ID:                      s.ID,
		LockDay:                 s.LockDay,
		AutoCreateDay:           s.AutoCreateDay,
		AutoCreateTarget:        s.AutoCreateTarget,
		AllowedFileTypes:        []string(s.AllowedFileTypes),
		MaxFileSizeMB:           s.MaxFileSizeMB,
		ReminderStartDay:        s.ReminderStartDay,
		AdminCanUploadAfterLock: s.AdminCanUploadAfterLock,
		UpdatedAt:               s.UpdatedAt,
	}
}

func toPlanFileEntity(f models.PlanFile) mpdomain.PlanFileEntity {
	e := mpdomain.PlanFileEntity{
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
		DeletedAt:     f.DeletedAt,
		CreatedAt:     f.CreatedAt,
		UpdatedAt:     f.UpdatedAt,
	}
	if f.Team != nil {
		e.TeamName = &f.Team.Name
	}
	if f.UploadedBy != nil {
		e.UploaderName = &f.UploadedBy.Username
	}
	return e
}

// Column name helpers (avoid import cycle with models.PlanFileCol where possible)
func PlanFileColMonthlyPlanID() string { return `"monthlyPlanId"` }
func PlanFileColIsDeleted() string     { return `"isDeleted"` }
func PlanFileColCreatedAt() string     { return `"createdAt"` }
