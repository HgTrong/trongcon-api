package repository

import (
	"context"

	"trongcon-api/internal/entity"

	"gorm.io/gorm"
)

type GymCheckInRepository interface {
	Create(ctx context.Context, row *entity.GymCheckIn) error
	ListRecent(ctx context.Context, limit int) ([]entity.GymCheckIn, error)
	ListByUser(ctx context.Context, userID uint, limit int) ([]entity.GymCheckIn, error)
}

type gymCheckInRepository struct{ db *gorm.DB }

func NewGymCheckInRepository(db *gorm.DB) GymCheckInRepository {
	return &gymCheckInRepository{db: db}
}

func (r *gymCheckInRepository) Create(ctx context.Context, row *entity.GymCheckIn) error {
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *gymCheckInRepository) ListRecent(ctx context.Context, limit int) ([]entity.GymCheckIn, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	var rows []entity.GymCheckIn
	err := r.db.WithContext(ctx).Order("checked_in_at DESC").Limit(limit).Find(&rows).Error
	return rows, err
}

func (r *gymCheckInRepository) ListByUser(ctx context.Context, userID uint, limit int) ([]entity.GymCheckIn, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	var rows []entity.GymCheckIn
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("checked_in_at DESC").Limit(limit).Find(&rows).Error
	return rows, err
}
