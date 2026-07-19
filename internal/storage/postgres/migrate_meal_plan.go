package postgres

import (
	"log"

	"trongcon-api/internal/entity"

	"gorm.io/gorm"
)

// migrateMealPlanMeals moves legacy flat meal_plan_items into meal_plan_meals groups.
func migrateMealPlanMeals(db *gorm.DB) error {
	if !db.Migrator().HasTable(&entity.MealPlanItem{}) {
		return nil
	}
	if !tableHasColumn(db, "meal_plan_items", "meal_plan_id") {
		return nil
	}
	if !db.Migrator().HasColumn(&entity.MealPlanItem{}, "meal_plan_meal_id") {
		if err := db.Exec(`ALTER TABLE meal_plan_items ADD COLUMN meal_plan_meal_id bigint`).Error; err != nil {
			return err
		}
	}

	type legacyItem struct {
		ID         uint
		MealPlanID uint
	}

	var items []legacyItem
	if err := db.Raw(`
		SELECT id, meal_plan_id
		FROM meal_plan_items
		WHERE meal_plan_meal_id IS NULL OR meal_plan_meal_id = 0
		ORDER BY meal_plan_id ASC, id ASC
	`).Scan(&items).Error; err != nil {
		return err
	}

	if len(items) > 0 {
		byPlan := make(map[uint][]uint)
		for _, it := range items {
			byPlan[it.MealPlanID] = append(byPlan[it.MealPlanID], it.ID)
		}
		for planID, itemIDs := range byPlan {
			meal := entity.MealPlanMeal{
				MealPlanID: planID,
				Name:       "Meal 1",
				SortOrder:  0,
			}
			if err := db.Create(&meal).Error; err != nil {
				return err
			}
			if err := db.Exec(`UPDATE meal_plan_items SET meal_plan_meal_id = ? WHERE id IN ?`, meal.ID, itemIDs).Error; err != nil {
				return err
			}
		}
		log.Printf("migrate: grouped %d meal plan items into daily meals", len(items))
	}

	if tableHasColumn(db, "meal_plan_items", "meal_plan_id") {
		if err := db.Exec(`ALTER TABLE meal_plan_items DROP COLUMN meal_plan_id`).Error; err != nil {
			return err
		}
		log.Println("migrate: dropped meal_plan_items.meal_plan_id")
	}
	return nil
}
