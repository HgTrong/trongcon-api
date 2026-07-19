package entity

import "time"

// UserNutritionGoal stores daily macro targets for a member.
type UserNutritionGoal struct {
	BaseEntity
	UserID        uint    `json:"user_id" gorm:"not null;uniqueIndex"`
	DailyCalories float64 `json:"daily_calories" gorm:"type:numeric(10,2);not null;default:2200"`
	DailyProteinG float64 `json:"daily_protein_g" gorm:"type:numeric(10,2);not null;default:165"`
	DailyCarbG    float64 `json:"daily_carb_g" gorm:"type:numeric(10,2);not null;default:220"`
	DailyFatG     float64 `json:"daily_fat_g" gorm:"type:numeric(10,2);not null;default:73"`
}

// FoodLogMeal is a user-defined meal bucket for one day (Meal 1, Meal 2, pre-workout, etc.).
type FoodLogMeal struct {
	BaseEntity
	UserID    uint      `json:"user_id" gorm:"not null;index:idx_food_log_meal_user_date"`
	LogDate   time.Time `json:"log_date" gorm:"type:date;not null;index:idx_food_log_meal_user_date"`
	Name      string    `json:"name" gorm:"type:varchar(100);not null"`
	SortOrder int       `json:"sort_order" gorm:"not null;default:0"`
}

// FoodLogEntry is one logged food line for a user on a given day.
type FoodLogEntry struct {
	BaseEntity
	UserID       uint      `json:"user_id" gorm:"not null;index:idx_food_log_user_date"`
	LogDate      time.Time `json:"log_date" gorm:"type:date;not null;index:idx_food_log_user_date"`
	MealID       uint      `json:"meal_id" gorm:"not null;index"`
	FoodID       uint      `json:"food_id" gorm:"not null;index"`
	FoodName     string    `json:"food_name" gorm:"type:varchar(200);not null"`
	Quantity     float64   `json:"quantity" gorm:"type:numeric(10,2);not null;default:1"`
	ServingSizeG float64   `json:"serving_size_g" gorm:"type:numeric(10,2);not null;default:100"`
	Protein      float64   `json:"protein" gorm:"type:numeric(10,2);not null;default:0"`
	Carb         float64   `json:"carb" gorm:"type:numeric(10,2);not null;default:0"`
	Fat          float64   `json:"fat" gorm:"type:numeric(10,2);not null;default:0"`
	Calories     float64   `json:"calories" gorm:"type:numeric(10,2);not null;default:0"`
}
