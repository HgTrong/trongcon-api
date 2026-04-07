package v1

import "time"

type CreateReq struct {
	Title      string `json:"title" binding:"required,min=1,max=500"`
	Subtitle   string `json:"subtitle" binding:"omitempty,max=1000"`
	Thumbnail  string `json:"thumbnail"`
	Video      string `json:"video"`
	Content    string `json:"content"`
	UserID     uint   `json:"user_id" binding:"required"`
	CategoryID uint   `json:"category_id" binding:"required"`
}

type CreateRes struct {
	Article ArticleDetailRes `json:"article"`
}

type UpdateReq struct {
	Title      *string `json:"title" binding:"omitempty,min=1,max=500"`
	Subtitle   *string `json:"subtitle" binding:"omitempty,max=1000"`
	Thumbnail  *string `json:"thumbnail"`
	Video      *string `json:"video"`
	Content    *string `json:"content"`
	UserID     *uint   `json:"user_id"`
	CategoryID *uint   `json:"category_id"`
}

type UpdateRes struct {
	Article ArticleDetailRes `json:"article"`
}

type GetRes struct {
	Article ArticleDetailRes `json:"article"`
}

type ListReq struct {
	Page     int    `form:"page"`
	Limit    int    `form:"limit"`
	OrderBy  string `form:"order_by"`
	OrderDir string `form:"order_dir"`
}

type ListRes struct {
	Total int64            `json:"total"`
	Data  []ArticleListRes `json:"data"`
}

type ArticleListRes struct {
	ID           uint      `json:"id"`
	Title        string    `json:"title"`
	Subtitle     string    `json:"subtitle"`
	Slug         string    `json:"slug"`
	Thumbnail    string    `json:"thumbnail"`
	Video        string    `json:"video"`
	UserID       uint      `json:"user_id"`
	UserEmail    string    `json:"user_email,omitempty"`
	CategoryID   uint      `json:"category_id"`
	CategoryName string    `json:"category_name,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type DeleteRes struct {
	Status string `json:"status"`
}

type ArticleDetailRes struct {
	ID           uint      `json:"id"`
	Title        string    `json:"title"`
	Subtitle     string    `json:"subtitle"`
	Slug         string    `json:"slug"`
	Thumbnail    string    `json:"thumbnail"`
	Video        string    `json:"video"`
	Content      string    `json:"content"`
	UserID       uint      `json:"user_id"`
	UserEmail    string    `json:"user_email,omitempty"`
	CategoryID   uint      `json:"category_id"`
	CategoryName string    `json:"category_name,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
