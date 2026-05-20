package repository

import (
	"context"
	"errors"
	"time"

	"backend-hotlines3/internal/feature/capability/entity"
	"backend-hotlines3/internal/models"

	"gorm.io/gorm"
)

type Repository interface {
	ListCodesByUserID(ctx context.Context, userID uint) ([]string, error)
	ReplaceCodes(ctx context.Context, userID uint, codes []string, grantedBy *uint) error
	GetTargetUser(ctx context.Context, userID uint) (*entity.TargetUser, error)
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

func (r *repository) GetTargetUser(ctx context.Context, userID uint) (*entity.TargetUser, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Scopes(models.UserNotDeleted).First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrTargetUserNotFound
		}
		return nil, err
	}
	return &entity.TargetUser{ID: user.ID, Role: user.Role, IsActive: user.IsActive}, nil
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
