package gorm

import (
	"context"

	taskdomain "backend-hotlines3/internal/domain/task"
	"backend-hotlines3/internal/models"
	"backend-hotlines3/internal/port/outbound/repository"

	"gorm.io/gorm"
)

type TaskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) List(ctx context.Context, query repository.TaskListQuery) ([]taskdomain.Entity, int64, error) {
	offset := (query.Page - 1) * query.Limit

	dbQuery := r.db.WithContext(ctx).Model(&models.TaskDaily{}).Scopes(models.TaskNotDeleted)

	if query.Filter.WorkDate != nil {
		dbQuery = dbQuery.Where("WorkDate = ?", query.Filter.WorkDate)
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
		Order("WorkDate DESC, CreatedAt DESC").
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
