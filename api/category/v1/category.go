package v1

import "time"

type CreateReq struct {
	Name   string `json:"name" binding:"required,min=1,max=200"`
	Icon   string `json:"icon"`
	Image  string `json:"image"`
	Status string `json:"status" binding:"omitempty,oneof=active inactive"`
	Type   string `json:"type"`
}

type CreateRes struct {
	Category CategoryRes `json:"category"`
}

type UpdateReq struct {
	Name   *string `json:"name" binding:"omitempty,min=1,max=200"`
	Icon   *string `json:"icon"`
	Image  *string `json:"image"`
	Status *string `json:"status" binding:"omitempty,oneof=active inactive"`
	Type   *string `json:"type"`
}

type UpdateRes struct {
	Category CategoryRes `json:"category"`
}

type GetRes struct {
	Category CategoryRes `json:"category"`
}

type ListReq struct {
	Page     int    `form:"page"`
	Limit    int    `form:"limit"`
	OrderBy  string `form:"order_by"`
	OrderDir string `form:"order_dir"`
}

type ListRes struct {
	Total int64         `json:"total"`
	Data  []CategoryRes `json:"data"`
}

type DeleteRes struct {
	Status string `json:"status"`
}

type CategoryRes struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Icon      string    `json:"icon"`
	Image     string    `json:"image"`
	Status    string    `json:"status"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
