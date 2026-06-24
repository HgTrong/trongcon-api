package v1

type CalculateReq struct {
	Gender        string  `json:"gender" binding:"required,oneof=male female"`
	Units         string  `json:"units" binding:"required,oneof=metric imperial"`
	Age           int     `json:"age" binding:"required,min=13,max=99"`
	Height        float64 `json:"height" binding:"required,gt=0"`
	HeightInches  float64 `json:"height_inches"`
	Weight        float64 `json:"weight" binding:"required,gt=0"`
	ActivityLevel string  `json:"activity_level" binding:"required"`
	Goal          string  `json:"goal" binding:"required"`
}

type CalculateRes struct {
	BMR            float64 `json:"bmr"`
	TDEE           float64 `json:"tdee"`
	DailyCalories  float64 `json:"daily_calories"`
	HeightCm       float64 `json:"height_cm"`
	WeightKg       float64 `json:"weight_kg"`
	Formula        string  `json:"formula"`
	ActivityFactor float64 `json:"activity_factor"`
	GoalAdjustment int     `json:"goal_adjustment"`
	Goal           string  `json:"goal"`
	ActivityLevel  string  `json:"activity_level"`
}
