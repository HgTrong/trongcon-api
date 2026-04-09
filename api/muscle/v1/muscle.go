package v1

import "time"

type CreateReq struct {
	Name string `json:"name" binding:"required,min=1,max=200"`
}

type CreateRes struct {
	Muscle MuscleRes `json:"muscle"`
}

type UpdateReq struct {
	Name *string `json:"name" binding:"omitempty,min=1,max=200"`
}

type UpdateRes struct {
	Muscle MuscleRes `json:"muscle"`
}

type GetRes struct {
	Muscle MuscleRes `json:"muscle"`
}

type ListReq struct {
	Page     int    `form:"page"`
	Limit    int    `form:"limit"`
	OrderBy  string `form:"order_by"`
	OrderDir string `form:"order_dir"`
}

type ListRes struct {
	Total int64       `json:"total"`
	Data  []MuscleRes `json:"data"`
}

type DeleteRes struct {
	Status string `json:"status"`
}

type MuscleRes struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
