package v1

import "time"

type MealPlanItemInput struct {
	FoodID   uint    `json:"food_id" binding:"required"`
	Quantity float64 `json:"quantity"`
}

type CreateReq struct {
	Title       string              `json:"title" binding:"required,min=1,max=200"`
	Description string              `json:"description"`
	UserID      uint                `json:"user_id" binding:"required"`
	IsPublic    bool                `json:"is_public"`
	Items       []MealPlanItemInput `json:"items"`
}

type CreateRes struct {
	MealPlan MealPlanRes `json:"meal_plan"`
}

type UpdateReq struct {
	Title       *string              `json:"title" binding:"omitempty,min=1,max=200"`
	Description *string              `json:"description"`
	UserID      *uint                `json:"user_id"`
	IsPublic    *bool                `json:"is_public"`
	Items       *[]MealPlanItemInput `json:"items"`
}

type UpdateRes struct {
	MealPlan MealPlanRes `json:"meal_plan"`
}

type GetRes struct {
	MealPlan MealPlanRes `json:"meal_plan"`
}

type ListReq struct {
	Page      int    `form:"page"`
	Limit     int    `form:"limit"`
	Q         string `form:"q"`
	UserID    *uint  `form:"user_id"`
	IsPublic  string `form:"is_public"`
	OrderBy   string `form:"order_by"`
	OrderDir  string `form:"order_dir"`
}

type ListRes struct {
	Total int64         `json:"total"`
	Data  []MealPlanRes `json:"data"`
}

type DeleteRes struct {
	Status string `json:"status"`
}

type MealPlanRes struct {
	ID           uint              `json:"id"`
	Title        string            `json:"title"`
	Description  string            `json:"description"`
	UserID       uint              `json:"user_id"`
	UserEmail    string            `json:"user_email,omitempty"`
	IsPublic     bool              `json:"is_public"`
	Items        []MealPlanItemRes `json:"items"`
	TotalProtein float64           `json:"total_protein"`
	TotalCarb    float64           `json:"total_carb"`
	TotalFat     float64           `json:"total_fat"`
	TotalCalories float64          `json:"total_calories"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

type MealPlanItemRes struct {
	ID           uint      `json:"id"`
	FoodID       uint      `json:"food_id"`
	FoodName     string    `json:"food_name"`
	Quantity     float64   `json:"quantity"`
	ServingSizeG float64   `json:"serving_size_g"`
	Protein      float64   `json:"protein"`
	Carb         float64   `json:"carb"`
	Fat          float64   `json:"fat"`
	Calories     float64   `json:"calories"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
