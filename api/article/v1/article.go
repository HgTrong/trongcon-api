package v1

import (
	"time"

	authorv1 "trongcon-api/api/author/v1"
)

type CreateReq struct {
	Title      string `json:"title" binding:"required,min=1,max=500"`
	Subtitle   string `json:"subtitle" binding:"omitempty,max=1000"`
	Thumbnail  string `json:"thumbnail"`
	Video      string `json:"video"`
	Content    string `json:"content"`
	UserID     uint   `json:"user_id" binding:"required"`
	CategoryID uint   `json:"category_id" binding:"required"`
	Featured   bool   `json:"featured"`
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
	Featured   *bool   `json:"featured"`
}

type UpdateRes struct {
	Article ArticleDetailRes `json:"article"`
}

type GetRes struct {
	Article ArticleDetailRes `json:"article"`
}

type ListReq struct {
	Page       int    `form:"page"`
	Limit      int    `form:"limit"`
	CategoryID *uint  `form:"category_id"`
	Featured   *bool  `form:"featured"`
	Q          string `form:"q"`
	OrderBy    string `form:"order_by"`
	OrderDir   string `form:"order_dir"`
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
	Featured            bool      `json:"featured"`
	AuthorName            string    `json:"author_name,omitempty"`
	AuthorProfilePicture  string    `json:"author_profile_picture,omitempty"`
	Author                *authorv1.AuthorRes `json:"author,omitempty"`
	Views                 int64     `json:"views"`
	CreatedAt             time.Time `json:"created_at"`
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
	Featured            bool      `json:"featured"`
	AuthorName            string    `json:"author_name,omitempty"`
	AuthorProfilePicture  string    `json:"author_profile_picture,omitempty"`
	Author                *authorv1.AuthorRes `json:"author,omitempty"`
	Views                 int64     `json:"views"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
