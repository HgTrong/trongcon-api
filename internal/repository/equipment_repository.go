package repository

import (
	"context"
	"errors"

	"trongcon-api/internal/entity"

	"gorm.io/gorm"
)

type EquipmentRepository interface {
	Create(ctx context.Context, e *entity.Equipment) error
	GetByID(ctx context.Context, id uint) (*entity.Equipment, error)
	Update(ctx context.Context, e *entity.Equipment) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, offset, limit int, order string) ([]entity.Equipment, int64, error)
}

type equipmentRepository struct {
	db *gorm.DB
}

func NewEquipmentRepository(db *gorm.DB) EquipmentRepository {
	return &equipmentRepository{db: db}
}

func (r *equipmentRepository) Create(ctx context.Context, e *entity.Equipment) error {
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *equipmentRepository) GetByID(ctx context.Context, id uint) (*entity.Equipment, error) {
	var e entity.Equipment
	if err := r.db.WithContext(ctx).First(&e, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	return &e, nil
}

func (r *equipmentRepository) Update(ctx context.Context, e *entity.Equipment) error {
	return r.db.WithContext(ctx).Save(e).Error
}

func (r *equipmentRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.Equipment{}, id).Error
}

func (r *equipmentRepository) List(ctx context.Context, offset, limit int, order string) ([]entity.Equipment, int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&entity.Equipment{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []entity.Equipment
	if order == "" {
		order = "id DESC"
	}
	if err := r.db.WithContext(ctx).Order(order).Offset(offset).Limit(limit).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
