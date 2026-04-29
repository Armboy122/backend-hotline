package repository

import (
	"context"
	"errors"

	"backend-hotlines3/internal/feature/team/entity"
	"backend-hotlines3/internal/feature/team/mapper"
	"backend-hotlines3/internal/models"

	"gorm.io/gorm"
)

type Repository interface {
	List(ctx context.Context) ([]entity.Entity, error)
	GetByID(ctx context.Context, id int64) (*entity.Entity, error)
	Create(ctx context.Context, name string) (*entity.Entity, error)
	Update(ctx context.Context, id int64, name string) (*entity.Entity, error)
	Delete(ctx context.Context, id int64) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) List(ctx context.Context) ([]entity.Entity, error) {
	var modelsOut []models.Team
	if err := r.db.WithContext(ctx).Find(&modelsOut).Error; err != nil {
		return nil, err
	}

	ids := make([]int64, 0, len(modelsOut))
	for _, item := range modelsOut {
		ids = append(ids, item.ID)
	}
	countMap := models.CountTasksBy(r.db.WithContext(ctx), models.TaskCol.TeamID, ids)

	out := make([]entity.Entity, 0, len(modelsOut))
	for _, item := range modelsOut {
		out = append(out, mapper.FromModel(item, countMap[item.ID]))
	}
	return out, nil
}

func (r *repository) GetByID(ctx context.Context, id int64) (*entity.Entity, error) {
	var model models.Team
	if err := r.db.WithContext(ctx).First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}
	count := models.CountTasksFor(r.db.WithContext(ctx), models.TaskCol.TeamID, id)
	out := mapper.FromModel(model, count)
	return &out, nil
}

func (r *repository) Create(ctx context.Context, name string) (*entity.Entity, error) {
	model := models.Team{Name: name}
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return nil, err
	}
	out := mapper.FromModel(model, 0)
	return &out, nil
}

func (r *repository) Update(ctx context.Context, id int64, name string) (*entity.Entity, error) {
	var model models.Team
	if err := r.db.WithContext(ctx).First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}
	model.Name = name
	if err := r.db.WithContext(ctx).Save(&model).Error; err != nil {
		return nil, err
	}
	count := models.CountTasksFor(r.db.WithContext(ctx), models.TaskCol.TeamID, id)
	out := mapper.FromModel(model, count)
	return &out, nil
}

func (r *repository) Delete(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Delete(&models.Team{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return entity.ErrNotFound
	}
	return nil
}
