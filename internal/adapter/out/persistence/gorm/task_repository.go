package gorm

import (
	"context"
	"errors"
	"time"

	taskdomain "backend-hotlines3/internal/domain/task"
	"backend-hotlines3/internal/models"
	"backend-hotlines3/internal/port/outbound/repository"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type TaskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) List(ctx context.Context, query repository.TaskListQuery) ([]taskdomain.Entity, int64, error) {
	page := query.Page
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * query.Limit

	dbQuery := r.db.WithContext(ctx).Model(&models.TaskDaily{}).Scopes(models.TaskNotDeleted)

	if query.Filter.WorkDate != nil {
		dbQuery = dbQuery.Where(models.TaskCol.WorkDate+" = ?", query.Filter.WorkDate)
	}
	if query.Filter.TeamID != nil {
		dbQuery = dbQuery.Where(models.TaskCol.TeamID+" = ?", *query.Filter.TeamID)
	}
	if query.Filter.JobTypeID != nil {
		dbQuery = dbQuery.Where(models.TaskCol.JobTypeID+" = ?", *query.Filter.JobTypeID)
	}
	if query.Filter.FeederID != nil {
		dbQuery = dbQuery.Where(models.TaskCol.FeederID+" = ?", *query.Filter.FeederID)
	}

	var total int64
	dbQuery.Count(&total)

	var tasks []models.TaskDaily
	if err := dbQuery.
		Preload("Team").
		Preload("JobType").
		Preload("JobDetail").
		Preload("Feeder.Station.OperationCenter").
		Order(models.TaskCol.WorkDate + " DESC, \"createdat\" DESC").
		Offset(offset).
		Limit(query.Limit).
		Find(&tasks).Error; err != nil {
		return nil, 0, err
	}

	result := make([]taskdomain.Entity, 0, len(tasks))
	for _, item := range tasks {
		result = append(result, toDomainTask(item))
	}

	return result, total, nil
}

func (r *TaskRepository) GetByID(ctx context.Context, query repository.TaskGetQuery) (*taskdomain.Entity, error) {
	var task models.TaskDaily
	if err := r.db.WithContext(ctx).
		Scopes(models.TaskNotDeleted).
		Preload("Team").
		Preload("JobType").
		Preload("JobDetail").
		Preload("Feeder.Station.OperationCenter").
		First(&task, query.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	entity := toDomainTask(task)
	return &entity, nil
}

func (r *TaskRepository) Create(ctx context.Context, input repository.TaskCreateInput) (*taskdomain.Entity, error) {
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
	// Reload with relations
	if err := r.db.WithContext(ctx).
		Preload("Team").
		Preload("JobType").
		Preload("JobDetail").
		Preload("Feeder.Station.OperationCenter").
		First(&task, task.ID).Error; err != nil {
		return nil, err
	}
	entity := toDomainTask(task)
	return &entity, nil
}

func (r *TaskRepository) Update(ctx context.Context, input repository.TaskUpdateInput) (*taskdomain.Entity, error) {
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
	// Reload with relations
	if err := r.db.WithContext(ctx).
		Preload("Team").
		Preload("JobType").
		Preload("JobDetail").
		Preload("Feeder.Station.OperationCenter").
		First(&task, task.ID).Error; err != nil {
		return nil, err
	}
	entity := toDomainTask(task)
	return &entity, nil
}

func (r *TaskRepository) SoftDelete(ctx context.Context, cmd repository.TaskDeleteCommand) error {
	var task models.TaskDaily
	if err := r.db.WithContext(ctx).Scopes(models.TaskNotDeleted).First(&task, cmd.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil // idempotent: already deleted
		}
		return err
	}
	now := time.Now()
	return r.db.WithContext(ctx).Model(&task).Update("deletedat", now).Error
}

func (r *TaskRepository) ListByTeam(ctx context.Context, query repository.TaskByTeamQuery) ([]taskdomain.Entity, int64, error) {
	page := query.Page
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * query.Limit

	dbQuery := r.db.WithContext(ctx).Model(&models.TaskDaily{}).Scopes(models.TaskNotDeleted)

	if query.WorkDate != nil {
		dbQuery = dbQuery.Where(models.TaskCol.WorkDate+" = ?", query.WorkDate)
	}
	if query.TeamID != nil {
		dbQuery = dbQuery.Where(models.TaskCol.TeamID+" = ?", *query.TeamID)
	}
	if query.JobTypeID != nil {
		dbQuery = dbQuery.Where(models.TaskCol.JobTypeID+" = ?", *query.JobTypeID)
	}
	if query.FeederID != nil {
		dbQuery = dbQuery.Where(models.TaskCol.FeederID+" = ?", *query.FeederID)
	}

	var total int64
	dbQuery.Count(&total)

	var tasks []models.TaskDaily
	if err := dbQuery.
		Preload("Team").
		Preload("JobType").
		Preload("JobDetail").
		Preload("Feeder.Station.OperationCenter").
		Order(models.TaskCol.WorkDate + " DESC, \"createdat\" DESC").
		Offset(offset).
		Limit(query.Limit).
		Find(&tasks).Error; err != nil {
		return nil, 0, err
	}

	result := make([]taskdomain.Entity, 0, len(tasks))
	for _, item := range tasks {
		result = append(result, toDomainTask(item))
	}
	return result, total, nil
}

func (r *TaskRepository) ListByFilter(ctx context.Context, query repository.TaskByFilterQuery) ([]taskdomain.Entity, error) {
	dbQuery := r.db.WithContext(ctx).Model(&models.TaskDaily{}).Scopes(models.TaskNotDeleted)

	dbQuery = dbQuery.
		Where("EXTRACT(YEAR FROM "+models.TaskCol.WorkDate+") = ?", query.Year).
		Where("EXTRACT(MONTH FROM "+models.TaskCol.WorkDate+") = ?", query.Month)

	if query.TeamID != nil {
		dbQuery = dbQuery.Where(models.TaskCol.TeamID+" = ?", *query.TeamID)
	}

	var tasks []models.TaskDaily
	if err := dbQuery.
		Preload("Team").
		Preload("JobType").
		Preload("JobDetail").
		Preload("Feeder.Station.OperationCenter").
		Order(models.TaskCol.WorkDate + " DESC, \"createdat\" DESC").
		Find(&tasks).Error; err != nil {
		return nil, err
	}

	result := make([]taskdomain.Entity, 0, len(tasks))
	for _, item := range tasks {
		result = append(result, toDomainTask(item))
	}
	return result, nil
}

func toDomainTask(task models.TaskDaily) taskdomain.Entity {
	entity := taskdomain.Entity{
		ID:          task.ID,
		WorkDate:    task.WorkDate,
		TeamID:      task.TeamID,
		JobTypeID:   task.JobTypeID,
		JobDetailID: task.JobDetailID,
		FeederID:    task.FeederID,
		NumPole:     task.NumPole,
		DeviceCode:  task.DeviceCode,
		Detail:      task.Detail,
		URLsBefore:  []string(task.URLsBefore),
		URLsAfter:   []string(task.URLsAfter),
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
		DeletedAt:   task.DeletedAt,
	}

	if task.Latitude != nil {
		lat, _ := task.Latitude.Float64()
		entity.Latitude = &lat
	}
	if task.Longitude != nil {
		lng, _ := task.Longitude.Float64()
		entity.Longitude = &lng
	}

	if task.Team != nil {
		entity.TeamName = &task.Team.Name
	}
	if task.JobType != nil {
		entity.JobTypeName = &task.JobType.Name
	}
	if task.JobDetail != nil {
		entity.JobDetailName = &task.JobDetail.Name
	}
	if task.Feeder != nil {
		entity.FeederCode = &task.Feeder.Code
		if task.Feeder.Station != nil {
			entity.StationName = &task.Feeder.Station.Name
			if task.Feeder.Station.OperationCenter != nil {
				entity.OperationCenterID = &task.Feeder.Station.OperationCenter.ID
				entity.OperationCenterName = &task.Feeder.Station.OperationCenter.Name
			}
		}
	}

	return entity
}
