package v1

import "time"

type CreateReq struct {
	Name        string `json:"name" binding:"required,min=1,max=255"`
	Key         string `json:"key" binding:"required,min=1,max=255"`
	Subject     string `json:"subject" binding:"required,min=1,max=500"`
	Body        string `json:"body" binding:"required"`
	Description string `json:"description"`
	IsActive    *bool  `json:"is_active"`
}

type UpdateReq struct {
	Name        *string `json:"name" binding:"omitempty,min=1,max=255"`
	Key         *string `json:"key" binding:"omitempty,min=1,max=255"`
	Subject     *string `json:"subject" binding:"omitempty,min=1,max=500"`
	Body        *string `json:"body"`
	Description *string `json:"description"`
	IsActive    *bool   `json:"is_active"`
}

type ListReq struct {
	Page     int    `form:"page"`
	Limit    int    `form:"limit"`
	Q        string `form:"q"`
	IsActive *bool  `form:"is_active"`
	OrderBy  string `form:"order_by"`
	OrderDir string `form:"order_dir"`
}

type PreviewReq struct {
	Key     string                 `json:"key"`
	Subject string                 `json:"subject"`
	Body    string                 `json:"body"`
	Data    map[string]interface{} `json:"data"`
}

type PreviewRes struct {
	Subject string `json:"subject"`
	HTML    string `json:"html"`
}

type TestSendReq struct {
	To   string                 `json:"to" binding:"required,email"`
	Data map[string]interface{} `json:"data"`
}

type TestSendRes struct {
	Status string `json:"status"`
}

type DeleteRes struct {
	Status string `json:"status"`
}

type ListRes struct {
	Total int64              `json:"total"`
	Data  []EmailTemplateRes `json:"data"`
}

type GetRes struct {
	Template EmailTemplateRes `json:"template"`
}

type CreateRes struct {
	Template EmailTemplateRes `json:"template"`
}

type UpdateRes struct {
	Template EmailTemplateRes `json:"template"`
}

type EmailTemplateRes struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Key         string    `json:"key"`
	Subject     string    `json:"subject"`
	Body        string    `json:"body"`
	Description string    `json:"description,omitempty"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
