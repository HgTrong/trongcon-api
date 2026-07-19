package v1

import "time"

type CreateBranchReq struct {
	Name        string `json:"name" binding:"required,min=1,max=200"`
	Address     string `json:"address" binding:"required,min=1,max=500"`
	City        string `json:"city"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	Hours       string `json:"hours"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
	IsActive    *bool  `json:"is_active"`
	SortOrder   int    `json:"sort_order"`
}

type UpdateBranchReq struct {
	Name        *string `json:"name" binding:"omitempty,min=1,max=200"`
	Address     *string `json:"address" binding:"omitempty,min=1,max=500"`
	City        *string `json:"city"`
	Phone       *string `json:"phone"`
	Email       *string `json:"email"`
	Hours       *string `json:"hours"`
	Description *string `json:"description"`
	ImageURL    *string `json:"image_url"`
	IsActive    *bool   `json:"is_active"`
	SortOrder   *int    `json:"sort_order"`
}

type ListBranchReq struct {
	Page     int    `form:"page"`
	Limit    int    `form:"limit"`
	Q        string `form:"q"`
	City     string `form:"city"`
	Active   string `form:"active"` // true|false|empty
	OrderBy  string `form:"order_by"`
	OrderDir string `form:"order_dir"`
}

type BranchRes struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Address     string    `json:"address"`
	City        string    `json:"city"`
	Phone       string    `json:"phone"`
	Email       string    `json:"email"`
	Hours       string    `json:"hours"`
	Description string    `json:"description"`
	ImageURL    string    `json:"image_url"`
	IsActive    bool      `json:"is_active"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateBranchRes struct {
	Branch BranchRes `json:"branch"`
}

type UpdateBranchRes struct {
	Branch BranchRes `json:"branch"`
}

type GetBranchRes struct {
	Branch BranchRes `json:"branch"`
}

type ListBranchRes struct {
	Total int64       `json:"total"`
	Data  []BranchRes `json:"data"`
}

type DeleteRes struct {
	Status string `json:"status"`
}

type CreateTrainerReq struct {
	UserID          uint   `json:"user_id" binding:"required"`
	BranchID        *uint  `json:"branch_id"`
	DisplayName     string `json:"display_name" binding:"required,min=1,max=200"`
	Title           string `json:"title"`
	Bio             string `json:"bio"`
	Specialties     string `json:"specialties"`
	Certifications  string `json:"certifications"`
	YearsExperience int    `json:"years_experience"`
	IsPublic        *bool  `json:"is_public"`
}

type UpdateTrainerReq struct {
	BranchID        *uint   `json:"branch_id"`
	ClearBranch     bool    `json:"clear_branch"`
	DisplayName     *string `json:"display_name" binding:"omitempty,min=1,max=200"`
	Title           *string `json:"title"`
	Bio             *string `json:"bio"`
	Specialties     *string `json:"specialties"`
	Certifications  *string `json:"certifications"`
	YearsExperience *int    `json:"years_experience"`
	IsPublic        *bool   `json:"is_public"`
}

type ListTrainerReq struct {
	Page     int    `form:"page"`
	Limit    int    `form:"limit"`
	Q        string `form:"q"`
	BranchID *uint  `form:"branch_id"`
	Public   string `form:"public"` // true|false|empty — public list forces true
	OrderBy  string `form:"order_by"`
	OrderDir string `form:"order_dir"`
}

type TrainerRes struct {
	ID              uint       `json:"id"`
	UserID          uint       `json:"user_id"`
	Email           string     `json:"email,omitempty"`
	AvatarURL       string     `json:"avatar_url,omitempty"`
	BranchID        *uint      `json:"branch_id,omitempty"`
	BranchName      string     `json:"branch_name,omitempty"`
	BranchSlug      string     `json:"branch_slug,omitempty"`
	DisplayName     string     `json:"display_name"`
	Title           string     `json:"title"`
	Bio             string     `json:"bio"`
	Specialties     string     `json:"specialties"`
	Certifications  string     `json:"certifications"`
	YearsExperience int        `json:"years_experience"`
	IsPublic        bool       `json:"is_public"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type CreateTrainerRes struct {
	Trainer TrainerRes `json:"trainer"`
}

type UpdateTrainerRes struct {
	Trainer TrainerRes `json:"trainer"`
}

type GetTrainerRes struct {
	Trainer TrainerRes `json:"trainer"`
}

type ListTrainerRes struct {
	Total int64        `json:"total"`
	Data  []TrainerRes `json:"data"`
}
