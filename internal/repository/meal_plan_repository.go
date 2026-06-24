package repository

import (
	"context"
	"errors"
	"strings"

	"trongcon-api/internal/entity"

	"gorm.io/gorm"
)

type MealPlanRepository interface {
	Create(ctx context.Context, mp *entity.MealPlan) error
	GetByID(ctx context.Context, id uint) (*entity.MealPlan, error)
	Update(ctx context.Context, mp *entity.MealPlan) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, offset, limit int, order, q string, userID *uint, isPublic *bool) ([]entity.MealPlan, int64, error)
	ReplaceItems(ctx context.Context, mealPlanID uint, items []entity.MealPlanItem) error
}

type mealPlanRepository struct {
	db *gorm.DB
}

func NewMealPlanRepository(db *gorm.DB) MealPlanRepository {
	return &mealPlanRepository{db: db}
}

func (r *mealPlanRepository) Create(ctx context.Context, mp *entity.MealPlan) error {
	return r.db.WithContext(ctx).Create(mp).Error
}

func (r *mealPlanRepository) GetByID(ctx context.Context, id uint) (*entity.MealPlan, error) {
	var mp entity.MealPlan
	if err := r.db.WithContext(ctx).Preload("User").Preload("Items").First(&mp, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	return &mp, nil
}

func (r *mealPlanRepository) Update(ctx context.Context, mp *entity.MealPlan) error {
	return r.db.WithContext(ctx).Save(mp).Error
}

func (r *mealPlanRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.MealPlan{}, id).Error
}

func (r *mealPlanRepository) List(ctx context.Context, offset, limit int, order, q string, userID *uint, isPublic *bool) ([]entity.MealPlan, int64, error) {
	query := r.db.WithContext(ctx).Model(&entity.MealPlan{})
	if q != "" {
		query = query.Where("title ILIKE ?", "%"+q+"%")
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

	var list []entity.MealPlan
	if err := query.Preload("User").Preload("Items").Order(order).Offset(offset).Limit(limit).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *mealPlanRepository) ReplaceItems(ctx context.Context, mealPlanID uint, items []entity.MealPlanItem) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("meal_plan_id = ?", mealPlanID).Delete(&entity.MealPlanItem{}).Error; err != nil {
			return err
		}
		if len(items) == 0 {
			return nil
		}
		for i := range items {
			items[i].MealPlanID = mealPlanID
		}
		return tx.Create(&items).Error
	})
}
