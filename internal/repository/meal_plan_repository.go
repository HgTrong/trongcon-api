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
	ReplaceMeals(ctx context.Context, mealPlanID uint, meals []entity.MealPlanMeal) error
	IncrementViews(ctx context.Context, id uint) (int64, error)
}

type mealPlanRepository struct {
	db *gorm.DB
}

func NewMealPlanRepository(db *gorm.DB) MealPlanRepository {
	return &mealPlanRepository{db: db}
}

func mealPreload(db *gorm.DB) *gorm.DB {
	return db.Preload("Meals", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort_order ASC, id ASC")
	}).Preload("Meals.Items", func(db *gorm.DB) *gorm.DB {
		return db.Order("id ASC")
	})
}

func (r *mealPlanRepository) Create(ctx context.Context, mp *entity.MealPlan) error {
	return r.db.WithContext(ctx).Create(mp).Error
}

func (r *mealPlanRepository) GetByID(ctx context.Context, id uint) (*entity.MealPlan, error) {
	var mp entity.MealPlan
	q := mealPreload(r.db.WithContext(ctx).Preload("User"))
	if err := q.First(&mp, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	return &mp, nil
}

func (r *mealPlanRepository) Update(ctx context.Context, mp *entity.MealPlan) error {
	return r.db.WithContext(ctx).Omit("User").Save(mp).Error
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
	qb := mealPreload(query.Preload("User"))
	if err := qb.Order(order).Offset(offset).Limit(limit).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *mealPlanRepository) ReplaceMeals(ctx context.Context, mealPlanID uint, meals []entity.MealPlanMeal) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var mealIDs []uint
		if err := tx.Model(&entity.MealPlanMeal{}).Where("meal_plan_id = ?", mealPlanID).Pluck("id", &mealIDs).Error; err != nil {
			return err
		}
		if len(mealIDs) > 0 {
			if err := tx.Where("meal_plan_meal_id IN ?", mealIDs).Delete(&entity.MealPlanItem{}).Error; err != nil {
				return err
			}
		}
		if err := tx.Where("meal_plan_id = ?", mealPlanID).Delete(&entity.MealPlanMeal{}).Error; err != nil {
			return err
		}
		for i := range meals {
			meals[i].MealPlanID = mealPlanID
			meal := meals[i]
			items := meal.Items
			meal.Items = nil
			if err := tx.Create(&meal).Error; err != nil {
				return err
			}
			if len(items) == 0 {
				continue
			}
			for j := range items {
				items[j].MealPlanMealID = meal.ID
			}
			if err := tx.Create(&items).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *mealPlanRepository) IncrementViews(ctx context.Context, id uint) (int64, error) {
	res := r.db.WithContext(ctx).Model(&entity.MealPlan{}).Where("id = ?", id).
		UpdateColumn("views", gorm.Expr("views + 1"))
	if res.Error != nil {
		return 0, res.Error
	}
	var views int64
	if err := r.db.WithContext(ctx).Model(&entity.MealPlan{}).Where("id = ?", id).Select("views").Scan(&views).Error; err != nil {
		return 0, err
	}
	return views, nil
}
