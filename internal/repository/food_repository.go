package repository

import (
	"context"
	"errors"

	"trongcon-api/internal/entity"

	"gorm.io/gorm"
)

type FoodRepository interface {
	Create(ctx context.Context, f *entity.Food) error
	GetByID(ctx context.Context, id uint) (*entity.Food, error)
	Update(ctx context.Context, f *entity.Food) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, offset, limit int, order, q string) ([]entity.Food, int64, error)
}

type foodRepository struct {
	db *gorm.DB
}

func NewFoodRepository(db *gorm.DB) FoodRepository {
	return &foodRepository{db: db}
}

func (r *foodRepository) Create(ctx context.Context, f *entity.Food) error {
	return r.db.WithContext(ctx).Create(f).Error
}

func (r *foodRepository) GetByID(ctx context.Context, id uint) (*entity.Food, error) {
	var f entity.Food
	if err := r.db.WithContext(ctx).First(&f, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	return &f, nil
}

func (r *foodRepository) Update(ctx context.Context, f *entity.Food) error {
	return r.db.WithContext(ctx).Save(f).Error
}

func (r *foodRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.Food{}, id).Error
}

func (r *foodRepository) List(ctx context.Context, offset, limit int, order, q string) ([]entity.Food, int64, error) {
	query := r.db.WithContext(ctx).Model(&entity.Food{})
	if q != "" {
		query = query.Where("name ILIKE ?", "%"+q+"%")
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []entity.Food
	if order == "" {
		order = "id DESC"
	}
	if err := query.Order(order).Offset(offset).Limit(limit).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
