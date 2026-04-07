package v1

import "time"

type CreateReq struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=6"`
	Name        string `json:"name"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Gender      string `json:"gender" binding:"omitempty,oneof=male female other prefer_not_to_say"`
	Language    string `json:"language"`
	AccountType string `json:"account_type" binding:"omitempty,oneof=free premium"`
}

type CreateRes struct {
	User UserRes `json:"user"`
}

type UpdateReq struct {
	Email       string `json:"email" binding:"omitempty,email"`
	Name        string `json:"name"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Gender      string `json:"gender"`
	Language    string `json:"language"`
	AccountType string `json:"account_type"`
}

type UpdateRes struct {
	User UserRes `json:"user"`
}

type GetRes struct {
	User UserRes `json:"user"`
}

type ListUsersReq struct {
	Page     int    `form:"page"`
	Limit    int    `form:"limit"`
	OrderBy  string `form:"order_by"`
	OrderDir string `form:"order_dir"`
}

type ListUsersRes struct {
	Total int64     `json:"total"`
	Data  []UserRes `json:"data"`
}

type DeleteRes struct {
	Status string `json:"status"`
}

type UserRes struct {
	ID          uint       `json:"id"`
	Email       string     `json:"email"`
	Name        string     `json:"name"`
	FirstName   string     `json:"first_name"`
	LastName    string     `json:"last_name"`
	Gender      string     `json:"gender"`
	Language    string     `json:"language"`
	MemberSince time.Time  `json:"member_since"`
	LastLogin   *time.Time `json:"last_login,omitempty"`
	AccountType string     `json:"account_type"`
	Roles       []string   `json:"roles,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
