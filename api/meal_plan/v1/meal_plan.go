package v1

import (
	"time"

	authorv1 "trongcon-api/api/author/v1"
)

type MealPlanItemInput struct {
	FoodID   uint    `json:"food_id" binding:"required"`
	Quantity float64 `json:"quantity"`
}

type MealPlanMealInput struct {
	Name  string              `json:"name" binding:"required,min=1,max=100"`
	Items []MealPlanItemInput `json:"items"`
}

type CreateReq struct {
	Title       string              `json:"title" binding:"required,min=1,max=200"`
	Description string              `json:"description"`
	UserID      uint                `json:"user_id" binding:"required"`
	IsPublic    bool                `json:"is_public"`
	Meals       []MealPlanMealInput `json:"meals"`
}

type CreateRes struct {
	MealPlan MealPlanRes `json:"meal_plan"`
}

type UpdateReq struct {
	Title       *string              `json:"title" binding:"omitempty,min=1,max=200"`
	Description *string              `json:"description"`
	UserID      *uint                `json:"user_id"`
	IsPublic    *bool                `json:"is_public"`
	Meals       *[]MealPlanMealInput `json:"meals"`
}

type UpdateRes struct {
	MealPlan MealPlanRes `json:"meal_plan"`
}

type GetRes struct {
	MealPlan MealPlanRes `json:"meal_plan"`
}

type ListReq struct {
	Page     int    `form:"page"`
	Limit    int    `form:"limit"`
	Q        string `form:"q"`
	UserID   *uint  `form:"user_id"`
	IsPublic string `form:"is_public"`
	OrderBy  string `form:"order_by"`
	OrderDir string `form:"order_dir"`
}

type ListRes struct {
	Total int64         `json:"total"`
	Data  []MealPlanRes `json:"data"`
}

type DeleteRes struct {
	Status string `json:"status"`
}

type MacroTotalsRes struct {
	Calories float64 `json:"calories"`
	ProteinG float64 `json:"protein_g"`
	CarbG    float64 `json:"carb_g"`
	FatG     float64 `json:"fat_g"`
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

type MealPlanMealRes struct {
	ID        uint              `json:"id"`
	Name      string            `json:"name"`
	SortOrder int               `json:"sort_order"`
	Items     []MealPlanItemRes `json:"items"`
	Totals    MacroTotalsRes    `json:"totals"`
}

type MealPlanRes struct {
	ID            uint              `json:"id"`
	Title         string            `json:"title"`
	Description   string            `json:"description"`
	UserID        uint              `json:"user_id"`
	UserEmail     string            `json:"user_email,omitempty"`
	IsPublic      bool              `json:"is_public"`
	Author        *authorv1.AuthorRes `json:"author,omitempty"`
	Meals         []MealPlanMealRes `json:"meals"`
	MealCount     int               `json:"meal_count"`
	TotalProtein  float64           `json:"total_protein"`
	TotalCarb     float64           `json:"total_carb"`
	TotalFat      float64           `json:"total_fat"`
	TotalCalories float64           `json:"total_calories"`
	Views         int64             `json:"views"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}
