package repository

import (
	"context"
	"strings"

	"trongcon-api/internal/entity"

	"gorm.io/gorm"
)

type RoutineRepository interface {
	Create(ctx context.Context, r *entity.Routine) error
	GetByID(ctx context.Context, id uint) (*entity.Routine, error)
	Update(ctx context.Context, r *entity.Routine) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, offset, limit int, order, q, difficulty string, userID *uint, isPublic *bool) ([]entity.Routine, int64, error)
	ReplaceItems(ctx context.Context, routineID uint, items []entity.RoutineWorkout) error
}

type routineRepository struct {
	db *gorm.DB
}

func NewRoutineRepository(db *gorm.DB) RoutineRepository {
	return &routineRepository{db: db}
}

func routinePreload(db *gorm.DB) *gorm.DB {
	return db.
		Preload("Items", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("sort_order ASC, id ASC")
		}).
		Preload("Items.Workout").
		Preload("Items.Workout.Items", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("sort_order ASC, id ASC")
		})
}

func (r *routineRepository) Create(ctx context.Context, rt *entity.Routine) error {
	return r.db.WithContext(ctx).Create(rt).Error
}

func (r *routineRepository) GetByID(ctx context.Context, id uint) (*entity.Routine, error) {
	var rt entity.Routine
	if err := routinePreload(r.db.WithContext(ctx).Preload("User")).First(&rt, id).Error; err != nil {
		return nil, err
	}
	return &rt, nil
}

func (r *routineRepository) Update(ctx context.Context, rt *entity.Routine) error {
	return r.db.WithContext(ctx).Save(rt).Error
}

func (r *routineRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.Routine{}, id).Error
}

func (r *routineRepository) List(ctx context.Context, offset, limit int, order, q, difficulty string, userID *uint, isPublic *bool) ([]entity.Routine, int64, error) {
	query := r.db.WithContext(ctx).Model(&entity.Routine{})
	if q != "" {
		query = query.Where("title ILIKE ?", "%"+q+"%")
	}
	if difficulty != "" {
		query = query.Where("difficulty = ?", difficulty)
	}
	if userID != nil && *userID > 0 {
		query = query.Where("user_id = ?", *userID)
	}
	if isPublic != nil {
		query = query.Where("is_public = ?", *isPublic)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if order == "" {
		order = "id DESC"
	}
	order = strings.TrimSpace(order)

	var list []entity.Routine
	if err := routinePreload(query.Preload("User")).Order(order).Offset(offset).Limit(limit).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *routineRepository) ReplaceItems(ctx context.Context, routineID uint, items []entity.RoutineWorkout) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("routine_id = ?", routineID).Delete(&entity.RoutineWorkout{}).Error; err != nil {
			return err
		}
		if len(items) == 0 {
			return nil
		}
		for i := range items {
			items[i].RoutineID = routineID
			items[i].SortOrder = i + 1
		}
		return tx.Create(&items).Error
	})
}
