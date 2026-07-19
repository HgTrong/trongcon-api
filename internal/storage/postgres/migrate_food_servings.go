package postgres

import (
	"log"

	"trongcon-api/internal/entity"

	"gorm.io/gorm"
)

// patchFoodEggServings sets egg foods to per-piece servings (legacy DB used 100g).
func patchFoodEggServings(db *gorm.DB) error {
	if !db.Migrator().HasTable(&entity.Food{}) {
		return nil
	}

	type patch struct {
		Name    string
		Serving float64
		Scale   float64
	}
	patches := []patch{
		{Name: "Egg (whole)", Serving: 50, Scale: 0.5},
		{Name: "Egg white", Serving: 33, Scale: 0.33},
	}

	for _, p := range patches {
		var food entity.Food
		if err := db.Where("name = ?", p.Name).First(&food).Error; err != nil {
			continue
		}
		if food.ServingSizeG == p.Serving {
			continue
		}
		if food.ServingSizeG != 100 {
			continue
		}
		updates := map[string]interface{}{
			"serving_size_g": p.Serving,
			"protein":        round2(food.Protein * p.Scale),
			"carb":           round2(food.Carb * p.Scale),
			"fat":            round2(food.Fat * p.Scale),
			"calories":       round2(food.Calories * p.Scale),
		}
		if err := db.Model(&food).Updates(updates).Error; err != nil {
			return err
		}
		log.Printf("migrate: %s → 1 serving = %.0fg", p.Name, p.Serving)
	}
	return nil
}
