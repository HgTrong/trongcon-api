package v1

import "time"

type CreateReq struct {
	Name string `json:"name" binding:"required,min=1,max=200"`
	Icon string `json:"icon"`
}

type CreateRes struct {
	Equipment EquipmentRes `json:"equipment"`
}

type UpdateReq struct {
	Name *string `json:"name" binding:"omitempty,min=1,max=200"`
	Icon *string `json:"icon"`
}

type UpdateRes struct {
	Equipment EquipmentRes `json:"equipment"`
}

type GetRes struct {
	Equipment EquipmentRes `json:"equipment"`
}

type ListReq struct {
	Page     int    `form:"page"`
	Limit    int    `form:"limit"`
	OrderBy  string `form:"order_by"`
	OrderDir string `form:"order_dir"`
}

type ListRes struct {
	Total int64          `json:"total"`
	Data  []EquipmentRes `json:"data"`
}

type DeleteRes struct {
	Status string `json:"status"`
}

type EquipmentRes struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Icon      string    `json:"icon"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
