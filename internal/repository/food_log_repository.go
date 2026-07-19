package repository

import (
	"context"
	"errors"
	"time"

	"trongcon-api/internal/entity"

	"gorm.io/gorm"
)

type FoodLogRepository interface {
	GetGoals(ctx context.Context, userID uint) (*entity.UserNutritionGoal, error)
	UpsertGoals(ctx context.Context, goal *entity.UserNutritionGoal) error
	ListMealsByDate(ctx context.Context, userID uint, date time.Time) ([]entity.FoodLogMeal, error)
	GetMeal(ctx context.Context, userID, mealID uint) (*entity.FoodLogMeal, error)
	CreateMeal(ctx context.Context, meal *entity.FoodLogMeal) error
	UpdateMeal(ctx context.Context, meal *entity.FoodLogMeal) error
	DeleteMeal(ctx context.Context, userID, mealID uint) error
	CountEntriesInMeal(ctx context.Context, userID, mealID uint) (int64, error)
	LatestMealTemplate(ctx context.Context, userID uint, before time.Time) ([]entity.FoodLogMeal, error)
	ListByDate(ctx context.Context, userID uint, date time.Time) ([]entity.FoodLogEntry, error)
	GetEntry(ctx context.Context, userID, entryID uint) (*entity.FoodLogEntry, error)
	CreateEntry(ctx context.Context, entry *entity.FoodLogEntry) error
	UpdateEntry(ctx context.Context, entry *entity.FoodLogEntry) error
	DeleteEntry(ctx context.Context, userID, entryID uint) error
	ListRecentFoods(ctx context.Context, userID uint, limit int) ([]entity.FoodLogEntry, error)
	ListLoggedDates(ctx context.Context, userID uint, since time.Time) ([]time.Time, error)
}

type foodLogRepository struct {
	db *gorm.DB
}

func NewFoodLogRepository(db *gorm.DB) FoodLogRepository {
	return &foodLogRepository{db: db}
}

func (r *foodLogRepository) GetGoals(ctx context.Context, userID uint) (*entity.UserNutritionGoal, error) {
	var goal entity.UserNutritionGoal
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&goal).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &goal, nil
}

func (r *foodLogRepository) UpsertGoals(ctx context.Context, goal *entity.UserNutritionGoal) error {
	var existing entity.UserNutritionGoal
	err := r.db.WithContext(ctx).Where("user_id = ?", goal.UserID).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return r.db.WithContext(ctx).Create(goal).Error
	}
	if err != nil {
		return err
	}
	existing.DailyCalories = goal.DailyCalories
	existing.DailyProteinG = goal.DailyProteinG
	existing.DailyCarbG = goal.DailyCarbG
	existing.DailyFatG = goal.DailyFatG
	return r.db.WithContext(ctx).Save(&existing).Error
}

func (r *foodLogRepository) ListMealsByDate(ctx context.Context, userID uint, date time.Time) ([]entity.FoodLogMeal, error) {
	var list []entity.FoodLogMeal
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND log_date = ?", userID, date.Format("2006-01-02")).
		Order("sort_order ASC, id ASC").
		Find(&list).Error
	return list, err
}

func (r *foodLogRepository) GetMeal(ctx context.Context, userID, mealID uint) (*entity.FoodLogMeal, error) {
	var meal entity.FoodLogMeal
	if err := r.db.WithContext(ctx).Where("user_id = ? AND id = ?", userID, mealID).First(&meal).Error; err != nil {
		return nil, err
	}
	return &meal, nil
}

func (r *foodLogRepository) CreateMeal(ctx context.Context, meal *entity.FoodLogMeal) error {
	return r.db.WithContext(ctx).Create(meal).Error
}

func (r *foodLogRepository) UpdateMeal(ctx context.Context, meal *entity.FoodLogMeal) error {
	return r.db.WithContext(ctx).Save(meal).Error
}

func (r *foodLogRepository) DeleteMeal(ctx context.Context, userID, mealID uint) error {
	return r.db.WithContext(ctx).Where("user_id = ? AND id = ?", userID, mealID).Delete(&entity.FoodLogMeal{}).Error
}

func (r *foodLogRepository) CountEntriesInMeal(ctx context.Context, userID, mealID uint) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&entity.FoodLogEntry{}).
		Where("user_id = ? AND meal_id = ?", userID, mealID).
		Count(&n).Error
	return n, err
}

func (r *foodLogRepository) LatestMealTemplate(ctx context.Context, userID uint, before time.Time) ([]entity.FoodLogMeal, error) {
	var latestDate *time.Time
	err := r.db.WithContext(ctx).Model(&entity.FoodLogMeal{}).
		Where("user_id = ? AND log_date < ?", userID, before.Format("2006-01-02")).
		Select("MAX(log_date)").
		Scan(&latestDate).Error
	if err != nil || latestDate == nil {
		return nil, err
	}

	var list []entity.FoodLogMeal
	err = r.db.WithContext(ctx).
		Where("user_id = ? AND log_date = ?", userID, latestDate.Format("2006-01-02")).
		Order("sort_order ASC, id ASC").
		Find(&list).Error
	return list, err
}

func (r *foodLogRepository) ListByDate(ctx context.Context, userID uint, date time.Time) ([]entity.FoodLogEntry, error) {
	var list []entity.FoodLogEntry
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND log_date = ?", userID, date.Format("2006-01-02")).
		Order("meal_id ASC, id ASC").
		Find(&list).Error
	return list, err
}

func (r *foodLogRepository) GetEntry(ctx context.Context, userID, entryID uint) (*entity.FoodLogEntry, error) {
	var entry entity.FoodLogEntry
	if err := r.db.WithContext(ctx).Where("user_id = ? AND id = ?", userID, entryID).First(&entry).Error; err != nil {
		return nil, err
	}
	return &entry, nil
}

func (r *foodLogRepository) CreateEntry(ctx context.Context, entry *entity.FoodLogEntry) error {
	return r.db.WithContext(ctx).Create(entry).Error
}

func (r *foodLogRepository) UpdateEntry(ctx context.Context, entry *entity.FoodLogEntry) error {
	return r.db.WithContext(ctx).Save(entry).Error
}

func (r *foodLogRepository) DeleteEntry(ctx context.Context, userID, entryID uint) error {
	return r.db.WithContext(ctx).Where("user_id = ? AND id = ?", userID, entryID).Delete(&entity.FoodLogEntry{}).Error
}

func (r *foodLogRepository) ListRecentFoods(ctx context.Context, userID uint, limit int) ([]entity.FoodLogEntry, error) {
	if limit < 1 {
		limit = 10
	}
	sub := r.db.WithContext(ctx).Model(&entity.FoodLogEntry{}).
		Select("MAX(id) as id").
		Where("user_id = ?", userID).
		Group("food_id")

	var ids []uint
	if err := sub.Pluck("id", &ids).Error; err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return []entity.FoodLogEntry{}, nil
	}

	var list []entity.FoodLogEntry
	err := r.db.WithContext(ctx).
		Where("id IN ?", ids).
		Order("updated_at DESC").
		Limit(limit).
		Find(&list).Error
	return list, err
}

func (r *foodLogRepository) ListLoggedDates(ctx context.Context, userID uint, since time.Time) ([]time.Time, error) {
	var dates []time.Time
	err := r.db.WithContext(ctx).Model(&entity.FoodLogEntry{}).
		Where("user_id = ? AND log_date >= ?", userID, since.Format("2006-01-02")).
		Distinct().
		Order("log_date DESC").
		Pluck("log_date", &dates).Error
	return dates, err
}
