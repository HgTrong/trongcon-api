package postgres

import (
	"fmt"
	"log"
	"time"

	"trongcon-api/internal/entity"

	"gorm.io/gorm"
)

var legacyMealSlotNames = map[string]string{
	"breakfast": "Meal 1",
	"lunch":     "Meal 2",
	"dinner":    "Meal 3",
	"snack":     "Meal 4",
}

// migrateFoodLogMeals moves legacy meal_slot entries to food_log_meals + meal_id.
func migrateFoodLogMeals(db *gorm.DB) error {
	if !db.Migrator().HasTable(&entity.FoodLogEntry{}) {
		return nil
	}
	if !tableHasColumn(db, "food_log_entries", "meal_slot") {
		return nil
	}
	if !db.Migrator().HasColumn(&entity.FoodLogEntry{}, "meal_id") {
		if err := db.Exec(`ALTER TABLE food_log_entries ADD COLUMN meal_id bigint`).Error; err != nil {
			return err
		}
	}

	type legacyRow struct {
		ID       uint
		UserID   uint
		LogDate  string
		MealSlot string
	}

	var rows []legacyRow
	if err := db.Raw(`
		SELECT id, user_id, log_date::text AS log_date, meal_slot
		FROM food_log_entries
		WHERE meal_id IS NULL OR meal_id = 0
	`).Scan(&rows).Error; err != nil {
		return err
	}
	if len(rows) == 0 {
		return dropLegacyMealSlot(db)
	}

	mealCache := make(map[string]uint)

	for _, row := range rows {
		key := fmt.Sprintf("%d|%s|%s", row.UserID, row.LogDate, row.MealSlot)
		mealID, ok := mealCache[key]
		if !ok {
			name := legacyMealSlotNames[row.MealSlot]
			if name == "" {
				name = row.MealSlot
			}

			var existing entity.FoodLogMeal
			err := db.Where("user_id = ? AND log_date = ? AND name = ?", row.UserID, row.LogDate, name).
				First(&existing).Error
			if err == nil {
				mealID = existing.ID
			} else if err == gorm.ErrRecordNotFound {
				var count int64
				db.Model(&entity.FoodLogMeal{}).
					Where("user_id = ? AND log_date = ?", row.UserID, row.LogDate).
					Count(&count)

				logDate, _ := time.Parse("2006-01-02", row.LogDate)
				meal := entity.FoodLogMeal{
					UserID:    row.UserID,
					LogDate:   logDate,
					Name:      name,
					SortOrder: int(count),
				}
				if err := db.Create(&meal).Error; err != nil {
					return err
				}
				mealID = meal.ID
			} else {
				return err
			}
			mealCache[key] = mealID
		}

		if err := db.Exec(`UPDATE food_log_entries SET meal_id = ? WHERE id = ?`, mealID, row.ID).Error; err != nil {
			return err
		}
	}

	log.Printf("migrate: backfilled meal_id for %d food log entries", len(rows))
	return dropLegacyMealSlot(db)
}

func dropLegacyMealSlot(db *gorm.DB) error {
	if tableHasColumn(db, "food_log_entries", "meal_slot") {
		if err := db.Exec(`ALTER TABLE food_log_entries DROP COLUMN meal_slot`).Error; err != nil {
			return err
		}
		log.Println("migrate: dropped food_log_entries.meal_slot")
	}
	return nil
}
