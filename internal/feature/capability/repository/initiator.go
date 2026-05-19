package repository

import (
	"context"
	"time"

	"backend-hotlines3/internal/models"

	"gorm.io/gorm"
)

type Repository interface {
	ListCodesByUserID(ctx context.Context, userID uint) ([]string, error)
	ReplaceCodes(ctx context.Context, userID uint, codes []string, grantedBy *uint) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) ListCodesByUserID(ctx context.Context, userID uint) ([]string, error) {
	var rows []models.UserCapability
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Order("code ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]string, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.Code)
	}
	return out, nil
}

func (r *repository) ReplaceCodes(ctx context.Context, userID uint, codes []string, grantedBy *uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		if err := tx.Model(&models.UserCapability{}).
			Where("user_id = ? AND revoked_at IS NULL", userID).
			Update("revoked_at", now).Error; err != nil {
			return err
		}
		for _, code := range codes {
			row := models.UserCapability{
				UserID:          userID,
				Code:            code,
				GrantedByUserID: grantedBy,
				CreatedAt:       now,
			}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
