package repository

import (
	"context"
	"time"

	"trongcon-api/internal/entity"

	"gorm.io/gorm"
)

type GymCheckInRepository interface {
	Create(ctx context.Context, row *entity.GymCheckIn) error
	ListRecent(ctx context.Context, limit int) ([]entity.GymCheckIn, error)
	ListByUser(ctx context.Context, userID uint, limit int) ([]entity.GymCheckIn, error)
	// CountToday returns (total check-ins today, distinct members checked in today).
	CountToday(ctx context.Context) (int64, int64, error)
	// CheckedInOn reports whether userID has a check-in on the calendar day of `day`.
	CheckedInOn(ctx context.Context, userID uint, day time.Time) (bool, error)
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

func (r *gymCheckInRepository) CountToday(ctx context.Context) (int64, int64, error) {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	var total int64
	if err := r.db.WithContext(ctx).Model(&entity.GymCheckIn{}).
		Where("checked_in_at >= ?", start).Count(&total).Error; err != nil {
		return 0, 0, err
	}

	var uniqueUsers int64
	if err := r.db.WithContext(ctx).Model(&entity.GymCheckIn{}).
		Where("checked_in_at >= ?", start).
		Distinct("user_id").Count(&uniqueUsers).Error; err != nil {
		return 0, 0, err
	}

	return total, uniqueUsers, nil
}

func (r *gymCheckInRepository) CheckedInOn(ctx context.Context, userID uint, day time.Time) (bool, error) {
	start := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, day.Location())
	end := start.Add(24 * time.Hour)

	var n int64
	err := r.db.WithContext(ctx).Model(&entity.GymCheckIn{}).
		Where("user_id = ? AND checked_in_at >= ? AND checked_in_at < ?", userID, start, end).
		Count(&n).Error
	return n > 0, err
}
