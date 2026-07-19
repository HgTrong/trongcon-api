package repository

import (
	"context"
	"errors"

	"trongcon-api/internal/entity"

	"gorm.io/gorm"
)

type SavedWorkoutRepository interface {
	Save(ctx context.Context, userID, workoutID uint) (*entity.UserSavedWorkout, error)
	Unsave(ctx context.Context, userID, workoutID uint) error
	IsSaved(ctx context.Context, userID, workoutID uint) (bool, error)
	ListWorkoutIDs(ctx context.Context, userID uint) ([]uint, error)
	List(ctx context.Context, userID uint, offset, limit int) ([]entity.Workout, int64, error)
}

type savedWorkoutRepository struct {
	db *gorm.DB
}

func NewSavedWorkoutRepository(db *gorm.DB) SavedWorkoutRepository {
	return &savedWorkoutRepository{db: db}
}

func (r *savedWorkoutRepository) Save(ctx context.Context, userID, workoutID uint) (*entity.UserSavedWorkout, error) {
	var existing entity.UserSavedWorkout
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND workout_id = ?", userID, workoutID).
		First(&existing).Error
	if err == nil {
		return &existing, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	row := &entity.UserSavedWorkout{UserID: userID, WorkoutID: workoutID}
	if err := r.db.WithContext(ctx).Create(row).Error; err != nil {
		return nil, err
	}
	return row, nil
}

func (r *savedWorkoutRepository) Unsave(ctx context.Context, userID, workoutID uint) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND workout_id = ?", userID, workoutID).
		Delete(&entity.UserSavedWorkout{}).Error
}

func (r *savedWorkoutRepository) IsSaved(ctx context.Context, userID, workoutID uint) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entity.UserSavedWorkout{}).
		Where("user_id = ? AND workout_id = ?", userID, workoutID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *savedWorkoutRepository) ListWorkoutIDs(ctx context.Context, userID uint) ([]uint, error) {
	var ids []uint
	if err := r.db.WithContext(ctx).Model(&entity.UserSavedWorkout{}).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Pluck("workout_id", &ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}

func (r *savedWorkoutRepository) List(ctx context.Context, userID uint, offset, limit int) ([]entity.Workout, int64, error) {
	query := r.db.WithContext(ctx).Model(&entity.UserSavedWorkout{}).Where("user_id = ?", userID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var saved []entity.UserSavedWorkout
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).
		Preload("Workout").
		Preload("Workout.Items", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("sort_order ASC, id ASC")
		}).
		Find(&saved).Error; err != nil {
		return nil, 0, err
	}

	workouts := make([]entity.Workout, 0, len(saved))
	for _, row := range saved {
		if row.Workout.ID > 0 {
			workouts = append(workouts, row.Workout)
		}
	}
	return workouts, total, nil
}
