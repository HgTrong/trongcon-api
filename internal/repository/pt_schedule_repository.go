package repository

import (
	"context"
	"time"

	"trongcon-api/internal/entity"

	"gorm.io/gorm"
)

type PTWorkingHoursRepository interface {
	ListByTrainer(ctx context.Context, trainerProfileID uint) ([]entity.PTWorkingHours, error)
	ListAllByTrainer(ctx context.Context, trainerProfileID uint) ([]entity.PTWorkingHours, error)
	ReplaceForTrainer(ctx context.Context, trainerProfileID uint, rows []entity.PTWorkingHours) error
}

type ptWorkingHoursRepository struct{ db *gorm.DB }

func NewPTWorkingHoursRepository(db *gorm.DB) PTWorkingHoursRepository {
	return &ptWorkingHoursRepository{db: db}
}

func (r *ptWorkingHoursRepository) ListByTrainer(ctx context.Context, trainerProfileID uint) ([]entity.PTWorkingHours, error) {
	var rows []entity.PTWorkingHours
	err := r.db.WithContext(ctx).
		Where("trainer_profile_id = ? AND is_active = ?", trainerProfileID, true).
		Order("weekday ASC, start_minute ASC").Find(&rows).Error
	return rows, err
}

func (r *ptWorkingHoursRepository) ListAllByTrainer(ctx context.Context, trainerProfileID uint) ([]entity.PTWorkingHours, error) {
	var rows []entity.PTWorkingHours
	err := r.db.WithContext(ctx).
		Where("trainer_profile_id = ?", trainerProfileID).
		Order("weekday ASC, start_minute ASC").Find(&rows).Error
	return rows, err
}

func (r *ptWorkingHoursRepository) ReplaceForTrainer(ctx context.Context, trainerProfileID uint, rows []entity.PTWorkingHours) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("trainer_profile_id = ?", trainerProfileID).Delete(&entity.PTWorkingHours{}).Error; err != nil {
			return err
		}
		for i := range rows {
			rows[i].ID = 0
			rows[i].TrainerProfileID = trainerProfileID
			if err := tx.Create(&rows[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

type PTBlockedSlotRepository interface {
	Create(ctx context.Context, b *entity.PTBlockedSlot) error
	Delete(ctx context.Context, id, trainerProfileID uint) error
	ListInRange(ctx context.Context, trainerProfileID uint, from, to time.Time) ([]entity.PTBlockedSlot, error)
}

type ptBlockedSlotRepository struct{ db *gorm.DB }

func NewPTBlockedSlotRepository(db *gorm.DB) PTBlockedSlotRepository {
	return &ptBlockedSlotRepository{db: db}
}

func (r *ptBlockedSlotRepository) Create(ctx context.Context, b *entity.PTBlockedSlot) error {
	return r.db.WithContext(ctx).Create(b).Error
}

func (r *ptBlockedSlotRepository) Delete(ctx context.Context, id, trainerProfileID uint) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND trainer_profile_id = ?", id, trainerProfileID).
		Delete(&entity.PTBlockedSlot{}).Error
}

func (r *ptBlockedSlotRepository) ListInRange(ctx context.Context, trainerProfileID uint, from, to time.Time) ([]entity.PTBlockedSlot, error) {
	var rows []entity.PTBlockedSlot
	err := r.db.WithContext(ctx).
		Where("trainer_profile_id = ? AND starts_at < ? AND ends_at > ?", trainerProfileID, to, from).
		Order("starts_at ASC").Find(&rows).Error
	return rows, err
}
