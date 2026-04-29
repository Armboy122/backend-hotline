package repository

import (
	"context"
	"errors"
	"time"

	"backend-hotlines3/internal/feature/task/entity"
	"backend-hotlines3/internal/feature/task/mapper"
	"backend-hotlines3/internal/models"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type repository struct {
	db *gorm.DB
}

func (r *repository) List(ctx context.Context, query ListQuery) ([]entity.Task, int64, error) {
	page := query.Page
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * query.Limit

	dbQuery := r.db.WithContext(ctx).Model(&models.TaskDaily{}).Scopes(models.TaskNotDeleted)
	dbQuery = applyListFilter(dbQuery, query.Filter)

	var total int64
	dbQuery.Count(&total)

	var tasks []models.TaskDaily
	if err := preloadTaskRelations(dbQuery).
		Order(models.TaskCol.WorkDate + " DESC, \"createdat\" DESC").
		Offset(offset).
		Limit(query.Limit).
		Find(&tasks).Error; err != nil {
		return nil, 0, err
	}

	return mapTasks(tasks), total, nil
}

func (r *repository) GetByID(ctx context.Context, query GetQuery) (*entity.Task, error) {
	var task models.TaskDaily
	if err := preloadTaskRelations(r.db.WithContext(ctx).Scopes(models.TaskNotDeleted)).
		First(&task, query.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	out := mapper.FromModel(task)
	return &out, nil
}

func (r *repository) Create(ctx context.Context, input CreateInput) (*entity.Task, error) {
	now := time.Now()
	task := models.TaskDaily{
		WorkDate:    input.WorkDate,
		TeamID:      input.TeamID,
		JobTypeID:   input.JobTypeID,
		JobDetailID: input.JobDetailID,
		FeederID:    input.FeederID,
		NumPole:     input.NumPole,
		DeviceCode:  input.DeviceCode,
		Detail:      input.Detail,
		URLsBefore:  models.StringArray(input.URLsBefore),
		URLsAfter:   models.StringArray(input.URLsAfter),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if input.Latitude != nil && input.Longitude != nil {
		lat := decimal.NewFromFloat(*input.Latitude)
		lng := decimal.NewFromFloat(*input.Longitude)
		task.Latitude = &lat
		task.Longitude = &lng
	}
	if err := r.db.WithContext(ctx).Create(&task).Error; err != nil {
		return nil, err
	}
	return r.reload(ctx, task.ID)
}

func (r *repository) Update(ctx context.Context, input UpdateInput) (*entity.Task, error) {
	var task models.TaskDaily
	if err := r.db.WithContext(ctx).Scopes(models.TaskNotDeleted).First(&task, input.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	updates := map[string]interface{}{}
	if input.WorkDate != nil {
		updates[models.TaskCol.WorkDate] = *input.WorkDate
	}
	if input.TeamID != nil {
		updates[models.TaskCol.TeamID] = *input.TeamID
	}
	if input.JobTypeID != nil {
		updates[models.TaskCol.JobTypeID] = *input.JobTypeID
	}
	if input.JobDetailID != nil {
		updates[models.TaskCol.JobDetailID] = *input.JobDetailID
	}
	if input.FeederID != nil {
		updates[models.TaskCol.FeederID] = *input.FeederID
	}
	if input.NumPole != nil {
		updates["numPole"] = *input.NumPole
	}
	if input.DeviceCode != nil {
		updates["deviceCode"] = *input.DeviceCode
	}
	if input.Detail != nil {
		updates["detail"] = *input.Detail
	}
	if input.URLsBefore != nil {
		updates["urlsBefore"] = models.StringArray(input.URLsBefore)
	}
	if input.URLsAfter != nil {
		updates["urlsAfter"] = models.StringArray(input.URLsAfter)
	}
	if input.Latitude != nil && input.Longitude != nil {
		updates["latitude"] = decimal.NewFromFloat(*input.Latitude)
		updates["longitude"] = decimal.NewFromFloat(*input.Longitude)
	}
	updates["updatedat"] = time.Now()

	if err := r.db.WithContext(ctx).Model(&task).Updates(updates).Error; err != nil {
		return nil, err
	}
	return r.reload(ctx, task.ID)
}

func (r *repository) SoftDelete(ctx context.Context, cmd DeleteCommand) error {
	var task models.TaskDaily
	if err := r.db.WithContext(ctx).Scopes(models.TaskNotDeleted).First(&task, cmd.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	now := time.Now()
	return r.db.WithContext(ctx).Model(&task).Update("deletedat", now).Error
}

func (r *repository) ListByTeam(ctx context.Context, query ByTeamQuery) ([]entity.Task, int64, error) {
	page := query.Page
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * query.Limit

	dbQuery := r.db.WithContext(ctx).Model(&models.TaskDaily{}).Scopes(models.TaskNotDeleted)
	dbQuery = applyListFilter(dbQuery, ListFilter{
		WorkDate:  query.WorkDate,
		TeamID:    query.TeamID,
		JobTypeID: query.JobTypeID,
		FeederID:  query.FeederID,
	})

	var total int64
	dbQuery.Count(&total)

	var tasks []models.TaskDaily
	if err := preloadTaskRelations(dbQuery).
		Order(models.TaskCol.WorkDate + " DESC, \"createdat\" DESC").
		Offset(offset).
		Limit(query.Limit).
		Find(&tasks).Error; err != nil {
		return nil, 0, err
	}

	return mapTasks(tasks), total, nil
}

func (r *repository) ListByFilter(ctx context.Context, query ByFilterQuery) ([]entity.Task, error) {
	dbQuery := r.db.WithContext(ctx).Model(&models.TaskDaily{}).Scopes(models.TaskNotDeleted)
	dbQuery = dbQuery.
		Where("EXTRACT(YEAR FROM "+models.TaskCol.WorkDate+") = ?", query.Year).
		Where("EXTRACT(MONTH FROM "+models.TaskCol.WorkDate+") = ?", query.Month)

	if query.TeamID != nil {
		dbQuery = dbQuery.Where(models.TaskCol.TeamID+" = ?", *query.TeamID)
	}

	var tasks []models.TaskDaily
	if err := preloadTaskRelations(dbQuery).
		Order(models.TaskCol.WorkDate + " DESC, \"createdat\" DESC").
		Find(&tasks).Error; err != nil {
		return nil, err
	}

	return mapTasks(tasks), nil
}

func (r *repository) reload(ctx context.Context, id int64) (*entity.Task, error) {
	var task models.TaskDaily
	if err := preloadTaskRelations(r.db.WithContext(ctx)).
		First(&task, id).Error; err != nil {
		return nil, err
	}
	out := mapper.FromModel(task)
	return &out, nil
}

func applyListFilter(db *gorm.DB, filter ListFilter) *gorm.DB {
	if filter.WorkDate != nil {
		db = db.Where(models.TaskCol.WorkDate+" = ?", filter.WorkDate)
	}
	if filter.TeamID != nil {
		db = db.Where(models.TaskCol.TeamID+" = ?", *filter.TeamID)
	}
	if filter.JobTypeID != nil {
		db = db.Where(models.TaskCol.JobTypeID+" = ?", *filter.JobTypeID)
	}
	if filter.FeederID != nil {
		db = db.Where(models.TaskCol.FeederID+" = ?", *filter.FeederID)
	}
	return db
}

func preloadTaskRelations(db *gorm.DB) *gorm.DB {
	return db.
		Preload("Team").
		Preload("JobType").
		Preload("JobDetail").
		Preload("Feeder.Station.OperationCenter")
}

func mapTasks(tasks []models.TaskDaily) []entity.Task {
	out := make([]entity.Task, 0, len(tasks))
	for _, task := range tasks {
		out = append(out, mapper.FromModel(task))
	}
	return out
}
