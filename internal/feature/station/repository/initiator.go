package repository

import (
	"context"
	"errors"

	"backend-hotlines3/internal/feature/station/entity"
	"backend-hotlines3/internal/feature/station/mapper"
	"backend-hotlines3/internal/models"

	"gorm.io/gorm"
)

type Repository interface {
	List(ctx context.Context) ([]entity.Entity, error)
	GetByID(ctx context.Context, id int64) (*entity.Entity, error)
	Create(ctx context.Context, input entity.CreateInput) (*entity.Entity, error)
	Update(ctx context.Context, input entity.UpdateInput) (*entity.Entity, error)
	Delete(ctx context.Context, id int64) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) List(ctx context.Context) ([]entity.Entity, error) {
	var modelsOut []models.Station
	if err := r.db.WithContext(ctx).Preload("OperationCenter").Find(&modelsOut).Error; err != nil {
		return nil, err
	}
	out := make([]entity.Entity, 0, len(modelsOut))
	for _, item := range modelsOut {
		out = append(out, mapper.FromModel(item))
	}
	return out, nil
}

func (r *repository) GetByID(ctx context.Context, id int64) (*entity.Entity, error) {
	var model models.Station
	if err := r.db.WithContext(ctx).Preload("OperationCenter").First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}
	out := mapper.FromModel(model)
	return &out, nil
}

func (r *repository) Create(ctx context.Context, input entity.CreateInput) (*entity.Entity, error) {
	model := models.Station{
		Name:        input.Name,
		CodeName:    input.CodeName,
		OperationID: input.OperationID,
	}
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).Preload("OperationCenter").First(&model, model.ID).Error; err != nil {
		return nil, err
	}
	out := mapper.FromModel(model)
	return &out, nil
}

func (r *repository) Update(ctx context.Context, input entity.UpdateInput) (*entity.Entity, error) {
	var model models.Station
	if err := r.db.WithContext(ctx).First(&model, input.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}
	if input.Name != "" {
		model.Name = input.Name
	}
	if input.CodeName != "" {
		model.CodeName = input.CodeName
	}
	if input.OperationID != 0 {
		model.OperationID = input.OperationID
	}
	if err := r.db.WithContext(ctx).Save(&model).Error; err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).Preload("OperationCenter").First(&model, model.ID).Error; err != nil {
		return nil, err
	}
	out := mapper.FromModel(model)
	return &out, nil
}

func (r *repository) Delete(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Delete(&models.Station{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return entity.ErrNotFound
	}
	return nil
}
