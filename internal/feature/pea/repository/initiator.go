package repository

import (
	"context"
	"errors"

	"backend-hotlines3/internal/feature/pea/entity"
	"backend-hotlines3/internal/feature/pea/mapper"
	"backend-hotlines3/internal/models"

	"gorm.io/gorm"
)

type Repository interface {
	List(ctx context.Context) ([]entity.Entity, error)
	GetByID(ctx context.Context, id int64) (*entity.Entity, error)
	Create(ctx context.Context, shortname, fullname string, operationID int64) (*entity.Entity, error)
	BulkCreate(ctx context.Context, items []entity.Entity) ([]entity.Entity, error)
	Update(ctx context.Context, id int64, shortname, fullname string, operationID int64) (*entity.Entity, error)
	Delete(ctx context.Context, id int64) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) List(ctx context.Context) ([]entity.Entity, error) {
	var rows []models.PEA
	if err := r.db.WithContext(ctx).Preload("OperationCenter").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]entity.Entity, 0, len(rows))
	for _, m := range rows {
		out = append(out, mapper.FromModel(m))
	}
	return out, nil
}

func (r *repository) GetByID(ctx context.Context, id int64) (*entity.Entity, error) {
	var m models.PEA
	if err := r.db.WithContext(ctx).Preload("OperationCenter").First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}
	out := mapper.FromModel(m)
	return &out, nil
}

func (r *repository) Create(ctx context.Context, shortname, fullname string, operationID int64) (*entity.Entity, error) {
	m := models.PEA{Shortname: shortname, Fullname: fullname, OperationID: operationID}
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return nil, err
	}
	r.db.WithContext(ctx).Preload("OperationCenter").First(&m, m.ID)
	out := mapper.FromModel(m)
	return &out, nil
}

func (r *repository) BulkCreate(ctx context.Context, items []entity.Entity) ([]entity.Entity, error) {
	rows := make([]models.PEA, 0, len(items))
	for _, item := range items {
		rows = append(rows, models.PEA{
			Shortname:   item.Shortname,
			Fullname:    item.Fullname,
			OperationID: item.OperationID,
		})
	}
	if err := r.db.WithContext(ctx).Create(&rows).Error; err != nil {
		return nil, err
	}

	ids := make([]int64, 0, len(rows))
	for _, m := range rows {
		ids = append(ids, m.ID)
	}
	r.db.WithContext(ctx).Preload("OperationCenter").Where("id IN ?", ids).Find(&rows)

	out := make([]entity.Entity, 0, len(rows))
	for _, m := range rows {
		out = append(out, mapper.FromModel(m))
	}
	return out, nil
}

func (r *repository) Update(ctx context.Context, id int64, shortname, fullname string, operationID int64) (*entity.Entity, error) {
	var m models.PEA
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}
	if shortname != "" {
		m.Shortname = shortname
	}
	if fullname != "" {
		m.Fullname = fullname
	}
	if operationID != 0 {
		m.OperationID = operationID
	}
	if err := r.db.WithContext(ctx).Save(&m).Error; err != nil {
		return nil, err
	}
	r.db.WithContext(ctx).Preload("OperationCenter").First(&m, m.ID)
	out := mapper.FromModel(m)
	return &out, nil
}

func (r *repository) Delete(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Delete(&models.PEA{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return entity.ErrNotFound
	}
	return nil
}
