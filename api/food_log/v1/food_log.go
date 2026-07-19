package v1

import "time"

type MacroTotals struct {
	Calories float64 `json:"calories"`
	ProteinG float64 `json:"protein_g"`
	CarbG    float64 `json:"carb_g"`
	FatG     float64 `json:"fat_g"`
}

type NutritionGoalRes struct {
	DailyCalories float64 `json:"daily_calories"`
	DailyProteinG float64 `json:"daily_protein_g"`
	DailyCarbG    float64 `json:"daily_carb_g"`
	DailyFatG     float64 `json:"daily_fat_g"`
}

type UpdateGoalsReq struct {
	DailyCalories float64 `json:"daily_calories" binding:"required,gt=0"`
	DailyProteinG float64 `json:"daily_protein_g" binding:"gte=0"`
	DailyCarbG    float64 `json:"daily_carb_g" binding:"gte=0"`
	DailyFatG     float64 `json:"daily_fat_g" binding:"gte=0"`
}

type FoodLogEntryRes struct {
	ID           uint      `json:"id"`
	FoodID       uint      `json:"food_id"`
	FoodName     string    `json:"food_name"`
	MealID       uint      `json:"meal_id"`
	Quantity     float64   `json:"quantity"`
	ServingSizeG float64   `json:"serving_size_g"`
	Protein      float64   `json:"protein"`
	Carb         float64   `json:"carb"`
	Fat          float64   `json:"fat"`
	Calories     float64   `json:"calories"`
	CreatedAt    time.Time `json:"created_at"`
}

type MealRes struct {
	ID        uint              `json:"id"`
	Name      string            `json:"name"`
	SortOrder int               `json:"sort_order"`
	Entries   []FoodLogEntryRes `json:"entries"`
	Totals    MacroTotals       `json:"totals"`
}

type NutritionHintRes struct {
	Type    string `json:"type"`
	Macro   string `json:"macro"`
	Message string `json:"message"`
}

type DayLogRes struct {
	Date        string             `json:"date"`
	Goals       NutritionGoalRes   `json:"goals"`
	Totals      MacroTotals        `json:"totals"`
	Remaining   MacroTotals        `json:"remaining"`
	Meals       []MealRes          `json:"meals"`
	Suggestions []NutritionHintRes `json:"suggestions"`
}

type GetDayReq struct {
	Date string `form:"date" binding:"required"`
}

type CreateEntryReq struct {
	FoodID   uint    `json:"food_id" binding:"required"`
	MealID   uint    `json:"meal_id" binding:"required"`
	Quantity float64 `json:"quantity" binding:"required,gt=0"`
	Date     string  `json:"date" binding:"required"`
}

type UpdateEntryReq struct {
	MealID   *uint    `json:"meal_id"`
	Quantity *float64 `json:"quantity" binding:"omitempty,gt=0"`
}

type CreateMealReq struct {
	Date string `json:"date" binding:"required"`
	Name string `json:"name" binding:"omitempty,max=100"`
}

type UpdateMealReq struct {
	Name string `json:"name" binding:"required,min=1,max=100"`
}

type MealOnlyRes struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	SortOrder int    `json:"sort_order"`
}

type RecentFoodRes struct {
	FoodID       uint      `json:"food_id"`
	FoodName     string    `json:"food_name"`
	ServingSizeG float64   `json:"serving_size_g"`
	Protein      float64   `json:"protein"`
	Carb         float64   `json:"carb"`
	Fat          float64   `json:"fat"`
	Calories     float64   `json:"calories"`
	LastLoggedAt time.Time `json:"last_logged_at"`
}

type RecentRes struct {
	Data []RecentFoodRes `json:"data"`
}

type DeleteRes struct {
	Status string `json:"status"`
}

type SaveFromCaloriesReq struct {
	DailyCalories float64 `json:"daily_calories" binding:"required,gt=0"`
	Preset        string  `json:"preset" binding:"omitempty,oneof=balanced low_carb high_protein ketogenic"`
}

type FoodLogStreakRes struct {
	Current      int  `json:"current"`
	Longest      int  `json:"longest"`
	LoggedToday  bool `json:"logged_today"`
	DaysThisWeek int  `json:"days_this_week"`
}

type TodayProgressRes struct {
	Calories     float64 `json:"calories"`
	GoalCalories float64 `json:"goal_calories"`
	ProteinG     float64 `json:"protein_g"`
	GoalProteinG float64 `json:"goal_protein_g"`
	Pct          int     `json:"pct"`
}

type MemberStatsRes struct {
	Goals      NutritionGoalRes `json:"goals"`
	GoalsSaved bool             `json:"goals_saved"`
	Streak     FoodLogStreakRes `json:"streak"`
	Today      TodayProgressRes `json:"today"`
}
