package v1

import (
	"time"

	authorv1 "trongcon-api/api/author/v1"
)

type WorkoutItemInput struct {
	ExerciseID uint   `json:"exercise_id" binding:"required"`
	Sets       int    `json:"sets"`
	Reps       string `json:"reps"`
}

type CreateReq struct {
	Title      string             `json:"title" binding:"required,min=1,max=200"`
	Difficulty string             `json:"difficulty" binding:"required,oneof=novice intermediate advanced"`
	Goal       string             `json:"goal" binding:"required,oneof=gain_muscle gain_strength lose_weight"`
	ImageURL   string             `json:"image_url"`
	UserID     uint               `json:"user_id" binding:"required"`
	IsPublic   bool               `json:"is_public"`
	Items      []WorkoutItemInput `json:"items"`
}

type CreateRes struct {
	Workout WorkoutRes `json:"workout"`
}

type UpdateReq struct {
	Title      *string             `json:"title" binding:"omitempty,min=1,max=200"`
	Difficulty *string             `json:"difficulty" binding:"omitempty,oneof=novice intermediate advanced"`
	Goal       *string             `json:"goal" binding:"omitempty,oneof=gain_muscle gain_strength lose_weight"`
	ImageURL   *string             `json:"image_url"`
	UserID     *uint               `json:"user_id"`
	IsPublic   *bool               `json:"is_public"`
	Items      *[]WorkoutItemInput `json:"items"`
}

type UpdateRes struct {
	Workout WorkoutRes `json:"workout"`
}

type GetRes struct {
	Workout WorkoutRes `json:"workout"`
}

type ListReq struct {
	Page       int    `form:"page"`
	Limit      int    `form:"limit"`
	Q          string `form:"q"`
	Difficulty string `form:"difficulty"`
	Goal       string `form:"goal"`
	OrderBy    string `form:"order_by"`
	OrderDir   string `form:"order_dir"`
}

type ListRes struct {
	Total int64        `json:"total"`
	Data  []WorkoutRes `json:"data"`
}

type DeleteRes struct {
	Status string `json:"status"`
}

type WorkoutItemRes struct {
	ID           uint      `json:"id"`
	ExerciseID   uint      `json:"exercise_id"`
	ExerciseName string    `json:"exercise_name"`
	SortOrder    int       `json:"sort_order"`
	Sets         int       `json:"sets"`
	Reps         string    `json:"reps"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type WorkoutRes struct {
	ID            uint                `json:"id"`
	Title         string              `json:"title"`
	Difficulty    string              `json:"difficulty"`
	Goal          string              `json:"goal"`
	ImageURL      string              `json:"image_url,omitempty"`
	UserID        uint                `json:"user_id"`
	UserEmail     string              `json:"user_email,omitempty"`
	OwnerUserID   *uint               `json:"owner_user_id,omitempty"`
	IsPublic      bool                `json:"is_public"`
	Views         int64               `json:"views"`
	Author        *authorv1.AuthorRes `json:"author,omitempty"`
	Items         []WorkoutItemRes    `json:"items"`
	ExerciseCount int                 `json:"exercise_count"`
	TotalSets     int                 `json:"total_sets"`
	CreatedAt     time.Time           `json:"created_at"`
	UpdatedAt     time.Time           `json:"updated_at"`
}
