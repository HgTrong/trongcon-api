package repository

import (
	"context"
	"errors"
	"strings"

	"trongcon-api/internal/entity"

	"gorm.io/gorm"
)

type MuscleRepository interface {
	Create(ctx context.Context, m *entity.Muscle) error
	GetByID(ctx context.Context, id uint) (*entity.Muscle, error)
	Update(ctx context.Context, m *entity.Muscle) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, offset, limit int, order, region string) ([]entity.Muscle, int64, error)
	SlugExists(ctx context.Context, slug string, excludeID uint) (bool, error)
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

func (r *muscleRepository) List(ctx context.Context, offset, limit int, order, region string) ([]entity.Muscle, int64, error) {
	query := r.db.WithContext(ctx).Model(&entity.Muscle{})
	if region != "" {
		query = query.Where("region = ?", region)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if order == "" {
		order = "id DESC"
	}
	order = strings.TrimSpace(order)

	var list []entity.Muscle
	if err := query.Order(order).Offset(offset).Limit(limit).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *muscleRepository) SlugExists(ctx context.Context, slug string, excludeID uint) (bool, error) {
	var count int64
	q := r.db.WithContext(ctx).Model(&entity.Muscle{}).Where("slug = ?", slug)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	if err := q.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
