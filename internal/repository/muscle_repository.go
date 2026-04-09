package repository

import (
	"context"
	"errors"

	"trongcon-api/internal/entity"

	"gorm.io/gorm"
)

type MuscleRepository interface {
	Create(ctx context.Context, m *entity.Muscle) error
	GetByID(ctx context.Context, id uint) (*entity.Muscle, error)
	Update(ctx context.Context, m *entity.Muscle) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, offset, limit int, order string) ([]entity.Muscle, int64, error)
}

type muscleRepository struct {
	db *gorm.DB
}

func NewMuscleRepository(db *gorm.DB) MuscleRepository {
	return &muscleRepository{db: db}
}

func (r *muscleRepository) Create(ctx context.Context, m *entity.Muscle) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *muscleRepository) GetByID(ctx context.Context, id uint) (*entity.Muscle, error) {
	var m entity.Muscle
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	return &m, nil
}

func (r *muscleRepository) Update(ctx context.Context, m *entity.Muscle) error {
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *muscleRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.Muscle{}, id).Error
}

func (r *muscleRepository) List(ctx context.Context, offset, limit int, order string) ([]entity.Muscle, int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&entity.Muscle{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []entity.Muscle
	if order == "" {
		order = "id DESC"
	}
	if err := r.db.WithContext(ctx).Order(order).Offset(offset).Limit(limit).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
