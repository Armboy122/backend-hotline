package repository

import (
	"context"
	"errors"

	"backend-hotlines3/internal/feature/feeder/entity"
	"backend-hotlines3/internal/feature/feeder/mapper"
	"backend-hotlines3/internal/models"

	"gorm.io/gorm"
)

type Repository interface {
	List(ctx context.Context) ([]entity.Entity, error)
	GetByID(ctx context.Context, id int64) (*entity.Entity, error)
	Create(ctx context.Context, code string, stationID int64) (*entity.Entity, error)
	Update(ctx context.Context, id int64, code string, stationID int64) (*entity.Entity, error)
	Delete(ctx context.Context, id int64) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) List(ctx context.Context) ([]entity.Entity, error) {
	type feederRow struct {
		ID              int64  `gorm:"column:id"`
		Code            string `gorm:"column:code"`
		StationID       int64  `gorm:"column:stationId"`
		StationName     string `gorm:"column:station_name"`
		StationCodeName string `gorm:"column:station_code_name"`
		OpCenterID      int64  `gorm:"column:op_center_id"`
		OpCenterName    string `gorm:"column:op_center_name"`
	}

	var rows []feederRow
	if err := r.db.WithContext(ctx).Table(`"Feeder"`).
		Select(`"Feeder"."id", "Feeder"."code", "Feeder"."stationId", "Station"."name" as station_name, "Station"."codeName" as station_code_name, "OperationCenter"."id" as op_center_id, "OperationCenter"."name" as op_center_name`).
		Joins(`LEFT JOIN "Station" ON "Station"."id" = "Feeder"."stationId"`).
		Joins(`LEFT JOIN "OperationCenter" ON "OperationCenter"."id" = "Station"."operationId"`).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	ids := make([]int64, 0, len(rows))
	for _, f := range rows {
		ids = append(ids, f.ID)
	}
	countMap := models.CountTasksBy(r.db.WithContext(ctx), models.TaskCol.FeederID, ids)

	out := make([]entity.Entity, 0, len(rows))
	for _, f := range rows {
		e := entity.Entity{
			ID:        f.ID,
			Code:      f.Code,
			StationID: f.StationID,
			Tasks:     countMap[f.ID],
		}
		if f.StationID != 0 {
			nested := &entity.NestedStation{
				ID:       f.StationID,
				Name:     f.StationName,
				CodeName: f.StationCodeName,
			}
			if f.OpCenterID != 0 {
				nested.OperationCenter = &entity.NestedOperationCenter{
					ID:   f.OpCenterID,
					Name: f.OpCenterName,
				}
			}
			e.Station = nested
		}
		out = append(out, e)
	}
	return out, nil
}

func (r *repository) GetByID(ctx context.Context, id int64) (*entity.Entity, error) {
	var m models.Feeder
	if err := r.db.WithContext(ctx).Preload("Station.OperationCenter").First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}
	count := models.CountTasksFor(r.db.WithContext(ctx), models.TaskCol.FeederID, id)
	out := mapper.FromModel(m, count)
	return &out, nil
}

func (r *repository) Create(ctx context.Context, code string, stationID int64) (*entity.Entity, error) {
	m := models.Feeder{Code: code, StationID: stationID}
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return nil, err
	}
	r.db.WithContext(ctx).Preload("Station.OperationCenter").First(&m, m.ID)
	out := mapper.FromModel(m, 0)
	return &out, nil
}

func (r *repository) Update(ctx context.Context, id int64, code string, stationID int64) (*entity.Entity, error) {
	var m models.Feeder
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}
	if code != "" {
		m.Code = code
	}
	if stationID != 0 {
		m.StationID = stationID
	}
	if err := r.db.WithContext(ctx).Save(&m).Error; err != nil {
		return nil, err
	}
	r.db.WithContext(ctx).Preload("Station.OperationCenter").First(&m, m.ID)
	count := models.CountTasksFor(r.db.WithContext(ctx), models.TaskCol.FeederID, id)
	out := mapper.FromModel(m, count)
	return &out, nil
}

func (r *repository) Delete(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Delete(&models.Feeder{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return entity.ErrNotFound
	}
	return nil
}
