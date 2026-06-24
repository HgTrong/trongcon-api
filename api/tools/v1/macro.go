package v1

type MacroCalculateReq struct {
	Preset        string  `json:"preset" binding:"required"`
	DailyCalories float64 `json:"daily_calories" binding:"required,gt=0"`
	MealsPerDay   int     `json:"meals_per_day" binding:"required,min=1,max=12"`
}

type MacroCalculateRes struct {
	Preset          string  `json:"preset"`
	PresetLabel     string  `json:"preset_label"`
	DailyCalories   float64 `json:"daily_calories"`
	MealsPerDay     int     `json:"meals_per_day"`
	CarbPct         int     `json:"carb_pct"`
	ProteinPct      int     `json:"protein_pct"`
	FatPct          int     `json:"fat_pct"`
	DailyCarbsG     float64 `json:"daily_carbs_g"`
	DailyProteinG   float64 `json:"daily_protein_g"`
	DailyFatG       float64 `json:"daily_fat_g"`
	PerMealCarbsG   float64 `json:"per_meal_carbs_g"`
	PerMealProteinG float64 `json:"per_meal_protein_g"`
	PerMealFatG     float64 `json:"per_meal_fat_g"`
}
