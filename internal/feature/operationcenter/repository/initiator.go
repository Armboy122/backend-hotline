package repository

import (
	"context"
	"errors"

	"backend-hotlines3/internal/feature/operationcenter/entity"
	"backend-hotlines3/internal/feature/operationcenter/mapper"
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
	var rows []models.OperationCenter
	if err := r.db.WithContext(ctx).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]entity.Entity, 0, len(rows))
	for _, m := range rows {
		out = append(out, mapper.FromModel(m))
	}
	return out, nil
}

func (r *repository) GetByID(ctx context.Context, id int64) (*entity.Entity, error) {
	var m models.OperationCenter
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}
	out := mapper.FromModel(m)
	return &out, nil
}

func (r *repository) Create(ctx context.Context, name string) (*entity.Entity, error) {
	m := models.OperationCenter{Name: name}
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return nil, err
	}
	out := mapper.FromModel(m)
	return &out, nil
}

func (r *repository) Update(ctx context.Context, id int64, name string) (*entity.Entity, error) {
	var m models.OperationCenter
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}
	m.Name = name
	if err := r.db.WithContext(ctx).Save(&m).Error; err != nil {
		return nil, err
	}
	out := mapper.FromModel(m)
	return &out, nil
}

func (r *repository) Delete(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Delete(&models.OperationCenter{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return entity.ErrNotFound
	}
	return nil
}
