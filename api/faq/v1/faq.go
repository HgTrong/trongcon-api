package v1

import "time"

type CreateReq struct {
	Question  string `json:"question" binding:"required,min=1"`
	Answer    string `json:"answer" binding:"required,min=1"`
	SortOrder int    `json:"sort_order"`
	IsActive  *bool  `json:"is_active"`
}

type UpdateReq struct {
	Question  *string `json:"question" binding:"omitempty,min=1"`
	Answer    *string `json:"answer" binding:"omitempty,min=1"`
	SortOrder *int    `json:"sort_order"`
	IsActive  *bool   `json:"is_active"`
}

type ListReq struct {
	Page     int    `form:"page"`
	Limit    int    `form:"limit"`
	Q        string `form:"q"`
	Active   string `form:"active"` // true|false|empty
	OrderBy  string `form:"order_by"`
	OrderDir string `form:"order_dir"`
}

type FAQRes struct {
	ID        uint      `json:"id"`
	Question  string    `json:"question"`
	Answer    string    `json:"answer"`
	SortOrder int       `json:"sort_order"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateRes struct {
	FAQ FAQRes `json:"faq"`
}

type UpdateRes struct {
	FAQ FAQRes `json:"faq"`
}

type GetRes struct {
	FAQ FAQRes `json:"faq"`
}

type ListRes struct {
	Total int64   `json:"total"`
	Data  []FAQRes `json:"data"`
}

type DeleteRes struct {
	Status string `json:"status"`
}
