package v1

import "time"

type CreateReq struct {
	Code           string   `json:"code"`
	PlanName       string   `json:"plan_name" binding:"required,min=1"`
	Title          string   `json:"title"`
	Description    []string `json:"description"`
	Price          float64  `json:"price" binding:"required,min=0"`
	Currency       string   `json:"currency"`
	DurationMonths int      `json:"duration_months" binding:"required,min=1"`
	IsActive       *bool    `json:"is_active"`
	SortOrder      int      `json:"sort_order"`
	Kind           string   `json:"kind"`
}

type UpdateReq struct {
	Code           *string  `json:"code"`
	PlanName       *string  `json:"plan_name"`
	Title          *string  `json:"title"`
	Description    []string `json:"description"`
	Price          *float64 `json:"price"`
	Currency       *string  `json:"currency"`
	DurationMonths *int     `json:"duration_months"`
	IsActive       *bool    `json:"is_active"`
	SortOrder      *int     `json:"sort_order"`
	Kind           *string  `json:"kind"`
}

type ListReq struct {
	Page     int    `form:"page"`
	Limit    int    `form:"limit"`
	Q        string `form:"q"`
	Kind     string `form:"kind"`
	Active   *bool  `form:"active"`
	OrderBy  string `form:"order_by"`
	OrderDir string `form:"order_dir"`
}

type PlanRes struct {
	ID             uint      `json:"id"`
	Code           string    `json:"code"`
	PlanName       string    `json:"plan_name"`
	Title          string    `json:"title"`
	Description    []string  `json:"description"`
	Price          float64   `json:"price"`
	Currency       string    `json:"currency"`
	DurationMonths int       `json:"duration_months"`
	IsActive       bool      `json:"is_active"`
	SortOrder      int       `json:"sort_order"`
	Kind           string    `json:"kind"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type GetRes struct {
	Plan PlanRes `json:"plan"`
}

type ListRes struct {
	Total int64     `json:"total"`
	Data  []PlanRes `json:"data"`
}

type DeleteRes struct {
	Status string `json:"status"`
}
