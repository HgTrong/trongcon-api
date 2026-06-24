package v1

import "time"

type CreateReq struct {
	Name         string  `json:"name" binding:"required,min=1,max=200"`
	Protein      float64 `json:"protein"`
	Carb         float64 `json:"carb"`
	Fat          float64 `json:"fat"`
	Calories     float64 `json:"calories"`
	ServingSizeG float64 `json:"serving_size_g"`
}

type CreateRes struct {
	Food FoodRes `json:"food"`
}

type UpdateReq struct {
	Name         *string  `json:"name" binding:"omitempty,min=1,max=200"`
	Protein      *float64 `json:"protein"`
	Carb         *float64 `json:"carb"`
	Fat          *float64 `json:"fat"`
	Calories     *float64 `json:"calories"`
	ServingSizeG *float64 `json:"serving_size_g"`
}

type UpdateRes struct {
	Food FoodRes `json:"food"`
}

type GetRes struct {
	Food FoodRes `json:"food"`
}

type ListReq struct {
	Page     int    `form:"page"`
	Limit    int    `form:"limit"`
	Q        string `form:"q"`
	OrderBy  string `form:"order_by"`
	OrderDir string `form:"order_dir"`
}

type ListRes struct {
	Total int64     `json:"total"`
	Data  []FoodRes `json:"data"`
}

type DeleteRes struct {
	Status string `json:"status"`
}

type FoodRes struct {
	ID           uint      `json:"id"`
	Name         string    `json:"name"`
	Protein      float64   `json:"protein"`
	Carb         float64   `json:"carb"`
	Fat          float64   `json:"fat"`
	Calories     float64   `json:"calories"`
	ServingSizeG float64   `json:"serving_size_g"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
