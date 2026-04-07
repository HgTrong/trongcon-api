package repository

import (
	"context"
	"errors"

	"trongcon-api/internal/entity"

	"gorm.io/gorm"
)

type CategoryRepository interface {
	Create(ctx context.Context, c *entity.Category) error
	GetByID(ctx context.Context, id uint) (*entity.Category, error)
	Update(ctx context.Context, c *entity.Category) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, offset, limit int, order string) ([]entity.Category, int64, error)
}

type categoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) Create(ctx context.Context, c *entity.Category) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *categoryRepository) GetByID(ctx context.Context, id uint) (*entity.Category, error) {
	var c entity.Category
	if err := r.db.WithContext(ctx).First(&c, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	return &c, nil
}

func (r *categoryRepository) Update(ctx context.Context, c *entity.Category) error {
	return r.db.WithContext(ctx).Save(c).Error
}

func (r *categoryRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.Category{}, id).Error
}

func (r *categoryRepository) List(ctx context.Context, offset, limit int, order string) ([]entity.Category, int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&entity.Category{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []entity.Category
	if order == "" {
		order = "id ASC"
	}
	if err := r.db.WithContext(ctx).Order(order).Offset(offset).Limit(limit).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
