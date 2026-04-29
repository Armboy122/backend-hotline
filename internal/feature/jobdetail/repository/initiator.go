package repository

import (
	"context"
	"errors"
	"time"

	"backend-hotlines3/internal/feature/jobdetail/entity"
	"backend-hotlines3/internal/feature/jobdetail/mapper"
	"backend-hotlines3/internal/models"

	"gorm.io/gorm"
)

type Repository interface {
	List(ctx context.Context) ([]entity.Entity, error)
	GetByID(ctx context.Context, id int64) (*entity.Entity, error)
	Create(ctx context.Context, name string, jobTypeID *int64) (*entity.Entity, error)
	Update(ctx context.Context, id int64, name string, jobTypeID *int64) (*entity.Entity, error)
	Delete(ctx context.Context, id int64) error
	Restore(ctx context.Context, id int64) (*entity.Entity, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) List(ctx context.Context) ([]entity.Entity, error) {
	var rows []models.JobDetail
	if err := r.db.WithContext(ctx).Scopes(models.JobDetailNotDeleted).Find(&rows).Error; err != nil {
		return nil, err
	}

	ids := make([]int64, 0, len(rows))
	for _, m := range rows {
		ids = append(ids, m.ID)
	}
	countMap := models.CountTasksBy(r.db.WithContext(ctx), models.TaskCol.JobDetailID, ids)

	out := make([]entity.Entity, 0, len(rows))
	for _, m := range rows {
		out = append(out, mapper.FromModel(m, countMap[m.ID]))
	}
	return out, nil
}

func (r *repository) GetByID(ctx context.Context, id int64) (*entity.Entity, error) {
	var m models.JobDetail
	if err := r.db.WithContext(ctx).Scopes(models.JobDetailNotDeleted).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}
	count := models.CountTasksFor(r.db.WithContext(ctx), models.TaskCol.JobDetailID, id)
	out := mapper.FromModel(m, count)
	return &out, nil
}

func (r *repository) Create(ctx context.Context, name string, jobTypeID *int64) (*entity.Entity, error) {
	now := time.Now()
	m := models.JobDetail{
		Name:      name,
		JobTypeID: jobTypeID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return nil, err
	}
	out := mapper.FromModel(m, 0)
	return &out, nil
}

func (r *repository) Update(ctx context.Context, id int64, name string, jobTypeID *int64) (*entity.Entity, error) {
	var m models.JobDetail
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}
	if name != "" {
		m.Name = name
	}
	if jobTypeID != nil {
		m.JobTypeID = jobTypeID
	}
	m.UpdatedAt = time.Now()
	if err := r.db.WithContext(ctx).Save(&m).Error; err != nil {
		return nil, err
	}
	count := models.CountTasksFor(r.db.WithContext(ctx), models.TaskCol.JobDetailID, id)
	out := mapper.FromModel(m, count)
	return &out, nil
}

func (r *repository) Delete(ctx context.Context, id int64) error {
	var m models.JobDetail
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return entity.ErrNotFound
		}
		return err
	}
	return r.db.WithContext(ctx).Model(&m).Update("deleted_at", time.Now()).Error
}

func (r *repository) Restore(ctx context.Context, id int64) (*entity.Entity, error) {
	var m models.JobDetail
	if err := r.db.WithContext(ctx).Unscoped().First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}
	if m.DeletedAt == nil {
		return nil, entity.ErrNotDeleted
	}
	m.DeletedAt = nil
	m.UpdatedAt = time.Now()
	if err := r.db.WithContext(ctx).Save(&m).Error; err != nil {
		return nil, err
	}
	count := models.CountTasksFor(r.db.WithContext(ctx), models.TaskCol.JobDetailID, id)
	out := mapper.FromModel(m, count)
	return &out, nil
}
