package repository

import (
	"context"
	"strings"

	"trongcon-api/internal/entity"

	"gorm.io/gorm"
)

type WorkoutRepository interface {
	Create(ctx context.Context, w *entity.Workout) error
	GetByID(ctx context.Context, id uint) (*entity.Workout, error)
	Update(ctx context.Context, w *entity.Workout) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, offset, limit int, order, q, difficulty string) ([]entity.Workout, int64, error)
	ReplaceItems(ctx context.Context, workoutID uint, items []entity.WorkoutItem) error
}

type workoutRepository struct {
	db *gorm.DB
}

func NewWorkoutRepository(db *gorm.DB) WorkoutRepository {
	return &workoutRepository{db: db}
}

func workoutItemPreload(db *gorm.DB) *gorm.DB {
	return db.Preload("Items", func(tx *gorm.DB) *gorm.DB {
		return tx.Order("sort_order ASC, id ASC")
	})
}

func (r *workoutRepository) Create(ctx context.Context, w *entity.Workout) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		items := w.Items
		w.Items = nil
		if err := tx.Create(w).Error; err != nil {
			return err
		}
		if len(items) == 0 {
			return nil
		}
		for i := range items {
			items[i].WorkoutID = w.ID
			items[i].SortOrder = i + 1
		}
		return tx.Create(&items).Error
	})
}

func (r *workoutRepository) GetByID(ctx context.Context, id uint) (*entity.Workout, error) {
	var w entity.Workout
	if err := workoutItemPreload(r.db.WithContext(ctx)).First(&w, id).Error; err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *workoutRepository) Update(ctx context.Context, w *entity.Workout) error {
	return r.db.WithContext(ctx).Save(w).Error
}

func (r *workoutRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.Workout{}, id).Error
}

func (r *workoutRepository) List(ctx context.Context, offset, limit int, order, q, difficulty string) ([]entity.Workout, int64, error) {
	query := r.db.WithContext(ctx).Model(&entity.Workout{})
	if q != "" {
		query = query.Where("title ILIKE ?", "%"+q+"%")
	}
	if difficulty != "" {
		query = query.Where("difficulty = ?", difficulty)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if order == "" {
		order = "id DESC"
	}
	order = strings.TrimSpace(order)

	var list []entity.Workout
	if err := workoutItemPreload(query).Order(order).Offset(offset).Limit(limit).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *workoutRepository) ReplaceItems(ctx context.Context, workoutID uint, items []entity.WorkoutItem) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("workout_id = ?", workoutID).Delete(&entity.WorkoutItem{}).Error; err != nil {
			return err
		}
		if len(items) == 0 {
			return nil
		}
		for i := range items {
			items[i].WorkoutID = workoutID
			items[i].SortOrder = i + 1
		}
		return tx.Create(&items).Error
	})
}
