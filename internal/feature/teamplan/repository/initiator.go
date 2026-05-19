package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"backend-hotlines3/internal/feature/teamplan/entity"
	"backend-hotlines3/internal/models"

	"gorm.io/gorm"
)

type ListQuery struct {
	From   time.Time
	To     time.Time
	TeamID *int64
	Status []string
	Page   int
	Limit  int
}

type GetQuery struct {
	ID int64
}

type CreateInput struct {
	TeamID            int64
	CreatedByUserID   int64
	Title             string
	WorkType          *string
	StartDate         *time.Time
	EndDate           *time.Time
	WorkTime          *string
	LocationText      string
	PEAID             *int64
	OperationCenterID *int64
	FeederID          *int64
	StationID         *int64
	Notes             *string
	Status            string
}

type UpdateInput struct {
	ID                int64
	TeamID            *int64
	Title             *string
	WorkType          *string
	StartDate         *time.Time
	EndDate           *time.Time
	WorkTime          *string
	LocationText      *string
	PEAID             *int64
	OperationCenterID *int64
	FeederID          *int64
	StationID         *int64
	Notes             *string
	Status            *string
}

type DeleteCommand struct {
	ID int64
}

type Repository interface {
	List(ctx context.Context, query ListQuery) ([]entity.TeamPlan, int64, error)
	ListAll(ctx context.Context, query ListQuery) ([]entity.TeamPlan, error)
	GetByID(ctx context.Context, query GetQuery) (*entity.TeamPlan, error)
	Create(ctx context.Context, input CreateInput) (*entity.TeamPlan, error)
	Update(ctx context.Context, input UpdateInput) (*entity.TeamPlan, error)
	SoftDelete(ctx context.Context, cmd DeleteCommand) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) List(ctx context.Context, query ListQuery) ([]entity.TeamPlan, int64, error) {
	base := r.db.WithContext(ctx).Model(&models.TeamPlan{})
	base = base.Where("deleted_at IS NULL")
	base = applyListFilters(base, query)

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := query.Page
	if page < 1 {
		page = 1
	}
	limit := query.Limit
	if limit < 1 {
		limit = 50
	}
	offset := (page - 1) * limit

	var modelsOut []models.TeamPlan
	err := applyListFilters(r.db.WithContext(ctx).Model(&models.TeamPlan{}).Where("deleted_at IS NULL"), query).
		Preload("Team").
		Preload("CreatedBy").
		Preload("PEA").
		Preload("OperationCenter").
		Preload("Feeder").
		Preload("Station").
		Preload("DailyTask").
		Order("start_date ASC, id ASC").
		Offset(offset).
		Limit(limit).
		Find(&modelsOut).Error
	if err != nil {
		return nil, 0, err
	}

	items := make([]entity.TeamPlan, 0, len(modelsOut))
	for _, model := range modelsOut {
		items = append(items, toEntity(model))
	}
	return items, total, nil
}

func (r *repository) ListAll(ctx context.Context, query ListQuery) ([]entity.TeamPlan, error) {
	var modelsOut []models.TeamPlan
	err := applyListFilters(r.db.WithContext(ctx).Model(&models.TeamPlan{}).Where("deleted_at IS NULL"), query).
		Preload("Team").
		Preload("CreatedBy").
		Preload("PEA").
		Preload("OperationCenter").
		Preload("Feeder").
		Preload("Station").
		Preload("DailyTask").
		Order("start_date ASC, id ASC").
		Find(&modelsOut).Error
	if err != nil {
		return nil, err
	}

	items := make([]entity.TeamPlan, 0, len(modelsOut))
	for _, model := range modelsOut {
		items = append(items, toEntity(model))
	}
	return items, nil
}

func (r *repository) GetByID(ctx context.Context, query GetQuery) (*entity.TeamPlan, error) {
	var model models.TeamPlan
	err := r.db.WithContext(ctx).
		Preload("Team").
		Preload("CreatedBy").
		Preload("PEA").
		Preload("OperationCenter").
		Preload("Feeder").
		Preload("Station").
		Preload("DailyTask").
		Where("id = ? AND deleted_at IS NULL", query.ID).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}
	out := toEntity(model)
	return &out, nil
}

func (r *repository) Create(ctx context.Context, input CreateInput) (*entity.TeamPlan, error) {
	status := input.Status
	if status == "" {
		status = entity.StatusPlanned
	}
	model := models.TeamPlan{
		TeamID:            input.TeamID,
		CreatedByUserID:   input.CreatedByUserID,
		Title:             input.Title,
		WorkType:          input.WorkType,
		StartDate:         dateOnlyPtr(input.StartDate),
		EndDate:           input.EndDate,
		WorkTime:          input.WorkTime,
		LocationText:      input.LocationText,
		PEAID:             input.PEAID,
		OperationCenterID: input.OperationCenterID,
		FeederID:          input.FeederID,
		StationID:         input.StationID,
		Notes:             input.Notes,
		Status:            status,
	}
	if model.EndDate != nil {
		end := dateOnly(*model.EndDate)
		model.EndDate = &end
	}
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return nil, err
	}
	out := toEntity(model)
	return &out, nil
}

func (r *repository) Update(ctx context.Context, input UpdateInput) (*entity.TeamPlan, error) {
	var model models.TeamPlan
	if err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", input.ID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}

	if input.TeamID != nil {
		model.TeamID = *input.TeamID
	}
	if input.Title != nil {
		model.Title = strings.TrimSpace(*input.Title)
	}
	if input.WorkType != nil {
		model.WorkType = input.WorkType
	}
	if input.StartDate != nil {
		model.StartDate = dateOnlyPtr(input.StartDate)
	}
	if input.EndDate != nil {
		end := dateOnly(*input.EndDate)
		model.EndDate = &end
	}
	if input.WorkTime != nil {
		model.WorkTime = input.WorkTime
	}
	if input.LocationText != nil {
		model.LocationText = strings.TrimSpace(*input.LocationText)
	}
	if input.PEAID != nil {
		model.PEAID = input.PEAID
	}
	if input.OperationCenterID != nil {
		model.OperationCenterID = input.OperationCenterID
	}
	if input.FeederID != nil {
		model.FeederID = input.FeederID
	}
	if input.StationID != nil {
		model.StationID = input.StationID
	}
	if input.Notes != nil {
		model.Notes = input.Notes
	}
	if input.Status != nil {
		model.Status = strings.TrimSpace(*input.Status)
	}
	if err := r.db.WithContext(ctx).Save(&model).Error; err != nil {
		return nil, err
	}
	out := toEntity(model)
	return &out, nil
}

func (r *repository) SoftDelete(ctx context.Context, cmd DeleteCommand) error {
	result := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", cmd.ID).Delete(&models.TeamPlan{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return entity.ErrNotFound
	}
	return nil
}

func applyListFilters(db *gorm.DB, query ListQuery) *gorm.DB {
	if !query.To.IsZero() || !query.From.IsZero() {
		db = db.Where("start_date IS NOT NULL")
	}
	if !query.To.IsZero() {
		db = db.Where("start_date <= ?", dateOnly(query.To))
	}
	if !query.From.IsZero() {
		db = db.Where("COALESCE(end_date, start_date) >= ?", dateOnly(query.From))
	}
	if query.TeamID != nil {
		db = db.Where("team_id = ?", *query.TeamID)
	}
	if len(query.Status) > 0 {
		db = db.Where("status IN ?", query.Status)
	}
	return db
}

func toEntity(model models.TeamPlan) entity.TeamPlan {
	out := entity.TeamPlan{
		ID:                model.ID,
		TeamID:            model.TeamID,
		CreatedByUserID:   model.CreatedByUserID,
		Title:             model.Title,
		WorkType:          model.WorkType,
		StartDate:         dateOnlyPtr(model.StartDate),
		EndDate:           model.EndDate,
		WorkTime:          model.WorkTime,
		LocationText:      model.LocationText,
		PEAID:             model.PEAID,
		OperationCenterID: model.OperationCenterID,
		FeederID:          model.FeederID,
		StationID:         model.StationID,
		Notes:             model.Notes,
		Status:            model.Status,
		DailyTaskID:       model.DailyTaskID,
		CreatedAt:         model.CreatedAt,
		UpdatedAt:         model.UpdatedAt,
		DeletedAt:         model.DeletedAt,
	}
	if model.Team != nil {
		out.Team = &entity.TeamSummary{ID: model.Team.ID, Name: model.Team.Name}
	}
	if model.CreatedBy != nil {
		out.CreatedBy = &entity.UserSummary{ID: int64(model.CreatedBy.ID), Username: model.CreatedBy.Username, DisplayName: model.CreatedBy.DisplayName}
	}
	return out
}

func expandDateRange(start time.Time, end *time.Time) []time.Time {
	start = dateOnly(start)
	stop := start
	if end != nil && !end.IsZero() {
		stop = dateOnly(*end)
	}
	if stop.Before(start) {
		stop = start
	}
	out := make([]time.Time, 0, int(stop.Sub(start)/(24*time.Hour))+1)
	for cur := start; !cur.After(stop); cur = cur.Add(24 * time.Hour) {
		out = append(out, cur)
	}
	return out
}

func dateOnly(t time.Time) time.Time {
	if t.IsZero() {
		return time.Time{}
	}
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func dateOnlyPtr(t *time.Time) *time.Time {
	if t == nil || t.IsZero() {
		return nil
	}
	out := dateOnly(*t)
	return &out
}
