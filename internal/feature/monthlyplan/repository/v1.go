package repository

import (
	"context"
	"fmt"
	"time"

	"backend-hotlines3/internal/feature/monthlyplan/entity"
	"backend-hotlines3/internal/models"

	"gorm.io/gorm"
)

// repository implements MonthlyPlan persistence via GORM.
type repository struct {
	db *gorm.DB
}

// ── Period ───────────────────────────────────────────────────────────────────

func (r *repository) FindPeriod(ctx context.Context, q PeriodFindQuery) (*entity.Entity, error) {
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

func (r *repository) GetPeriodByID(ctx context.Context, id int64) (*entity.Entity, error) {
	var mp models.MonthlyPlan
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&mp).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get period by id: %w", err)
	}
	e := toPeriodEntity(mp)
	return &e, nil
}

func (r *repository) FindOrCreatePeriod(ctx context.Context, year, month int) (*entity.Entity, error) {
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

func (r *repository) FindOrCreatePeriodsForYear(ctx context.Context, year int) ([]entity.Entity, error) {
	var existing []models.MonthlyPlan
	if err := r.db.WithContext(ctx).
		Where("year = ?", year).
		Find(&existing).Error; err != nil {
		return nil, fmt.Errorf("find periods for year: %w", err)
	}

	byMonth := make(map[int]models.MonthlyPlan, len(existing))
	for _, period := range existing {
		byMonth[period.Month] = period
	}

	missing := make([]models.MonthlyPlan, 0, 12-len(byMonth))
	for month := 1; month <= 12; month++ {
		if _, ok := byMonth[month]; !ok {
			missing = append(missing, models.MonthlyPlan{Year: year, Month: month})
		}
	}
	if len(missing) > 0 {
		if err := r.db.WithContext(ctx).Create(&missing).Error; err != nil {
			return nil, fmt.Errorf("create missing periods for year: %w", err)
		}
	}

	var periods []models.MonthlyPlan
	if err := r.db.WithContext(ctx).
		Where("year = ?", year).
		Order("month ASC").
		Find(&periods).Error; err != nil {
		return nil, fmt.Errorf("load periods for year: %w", err)
	}

	result := make([]entity.Entity, 0, 12)
	for _, period := range periods {
		if period.Month < 1 || period.Month > 12 {
			continue
		}
		result = append(result, toPeriodEntity(period))
	}
	return result, nil
}

// ── Settings ─────────────────────────────────────────────────────────────────

func (r *repository) GetOrCreateSettings(ctx context.Context) (*entity.SettingsEntity, error) {
	var s models.MonthlyPlanSetting
	err := r.db.WithContext(ctx).FirstOrCreate(&s).Error
	if err != nil {
		return nil, fmt.Errorf("get or create settings: %w", err)
	}
	e := toSettingsEntity(s)
	return &e, nil
}

func (r *repository) UpdateSettings(ctx context.Context, s *entity.SettingsEntity) error {
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

func (r *repository) ListFiles(ctx context.Context, q PlanFileListQuery) ([]entity.PlanFileEntity, error) {
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

	result := make([]entity.PlanFileEntity, len(files))
	for i, f := range files {
		result[i] = toPlanFileEntity(f)
	}
	return result, nil
}

func (r *repository) ListFilesByPlanIDs(ctx context.Context, q PlanFileBatchListQuery) (map[int64][]entity.PlanFileEntity, error) {
	result := make(map[int64][]entity.PlanFileEntity, len(q.MonthlyPlanIDs))
	if len(q.MonthlyPlanIDs) == 0 {
		return result, nil
	}

	query := r.db.WithContext(ctx).
		Preload("Team").
		Preload("UploadedBy").
		Where(PlanFileColMonthlyPlanID()+" IN ?", q.MonthlyPlanIDs).
		Scopes(models.PlanFileNotDeleted)
	if q.TeamID != 0 {
		query = query.Scopes(models.PlanFileByTeam(q.TeamID))
	}

	var files []models.PlanFile
	if err := query.Order(PlanFileColMonthlyPlanID() + " ASC, " + PlanFileColCreatedAt() + " DESC").Find(&files).Error; err != nil {
		return nil, fmt.Errorf("list files by plan ids: %w", err)
	}
	for _, f := range files {
		result[f.MonthlyPlanID] = append(result[f.MonthlyPlanID], toPlanFileEntity(f))
	}
	return result, nil
}

func (r *repository) GetFileByID(ctx context.Context, id int64) (*entity.PlanFileEntity, error) {
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

func (r *repository) CreateFile(ctx context.Context, input PlanFileCreateInput) (*entity.PlanFileEntity, error) {
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
		WorkStartDate: input.WorkStartDate,
		WorkEndDate:   input.WorkEndDate,
		Destination:   input.Destination,
		Remarks:       input.Remarks,
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

func (r *repository) SoftDeleteFile(ctx context.Context, input PlanFileSoftDeleteInput) error {
	result := r.db.WithContext(ctx).
		Model(&models.PlanFile{}).
		Where("id = ? AND \"isDeleted\" = false", input.ID).
		Updates(map[string]interface{}{
			"\"isDeleted\"": true,
			"\"deletedAt\"": time.Now(),
		})
	if result.Error != nil {
		return fmt.Errorf("soft delete file: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("soft delete file: no rows affected (id=%d)", input.ID)
	}
	return nil
}

func (r *repository) RestoreFile(ctx context.Context, input PlanFileRestoreInput) (*entity.PlanFileEntity, error) {
	result := r.db.WithContext(ctx).
		Model(&models.PlanFile{}).
		Where("id = ? AND \"isDeleted\" = true", input.ID).
		Updates(map[string]interface{}{
			"\"isDeleted\"": false,
			"\"deletedAt\"": nil,
		})
	if result.Error != nil {
		return nil, fmt.Errorf("restore file: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("restore file: no rows affected (id=%d)", input.ID)
	}
	return r.GetFileByID(ctx, input.ID)
}

func (r *repository) HardDeleteFile(ctx context.Context, input PlanFileHardDeleteInput) error {
	if err := r.db.WithContext(ctx).Delete(&models.PlanFile{}, input.ID).Error; err != nil {
		return fmt.Errorf("hard delete file: %w", err)
	}
	return nil
}

// ── File Size Log ────────────────────────────────────────────────────────────

func (r *repository) CreateFileSizeLog(ctx context.Context, input FileSizeLogCreateInput) error {
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

func (r *repository) CountFilesByTeam(ctx context.Context, q SubmissionQuery) ([]TeamCountRow, error) {
	var rows []TeamCountRow
	err := r.db.WithContext(ctx).
		Model(&models.PlanFile{}).
		Select("\"teamId\" as team_id, COUNT(*) as count").
		Where(PlanFileColMonthlyPlanID()+" = ?", q.PlanID).
		Where(PlanFileColIsDeleted() + " = false").
		Where("\"teamId\" IS NOT NULL").
		Group("\"teamId\"").
		Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("count files by team: %w", err)
	}
	return rows, nil
}

func (r *repository) ListTeams(ctx context.Context) ([]TeamInfo, error) {
	var teams []models.Team
	if err := r.db.WithContext(ctx).Find(&teams).Error; err != nil {
		return nil, fmt.Errorf("list teams: %w", err)
	}
	result := make([]TeamInfo, len(teams))
	for i, t := range teams {
		result[i] = TeamInfo{ID: t.ID, Name: t.Name}
	}
	return result, nil
}

// ── Mappers ──────────────────────────────────────────────────────────────────

func toPeriodEntity(mp models.MonthlyPlan) entity.Entity {
	return entity.Entity{
		ID:        mp.ID,
		Year:      mp.Year,
		Month:     mp.Month,
		IsLocked:  mp.IsLocked,
		CreatedAt: mp.CreatedAt,
	}
}

func toSettingsEntity(s models.MonthlyPlanSetting) entity.SettingsEntity {
	return entity.SettingsEntity{
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

func toPlanFileEntity(f models.PlanFile) entity.PlanFileEntity {
	e := entity.PlanFileEntity{
		ID:            f.ID,
		MonthlyPlanID: f.MonthlyPlanID,
		TeamID:        f.TeamID,
		UploadedByID:  f.UploadedByID,
		FileKey:       f.FileKey,
		FileURL:       f.FileURL,
		FileName:      f.FileName,
		FileSizeBytes: f.FileSizeBytes,
		Description:   f.Description,
		WorkStartDate: f.WorkStartDate,
		WorkEndDate:   f.WorkEndDate,
		Destination:   f.Destination,
		Remarks:       f.Remarks,
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
