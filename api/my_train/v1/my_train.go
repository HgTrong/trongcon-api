package v1

import "time"

type WorkoutItemInput struct {
	ExerciseID uint   `json:"exercise_id" binding:"required"`
	Sets       int    `json:"sets"`
	Reps       string `json:"reps"`
}

type CreateMyWorkoutReq struct {
	Title      string             `json:"title" binding:"required,min=1,max=200"`
	Difficulty string             `json:"difficulty" binding:"required,oneof=novice intermediate advanced"`
	Goal       string             `json:"goal" binding:"required,oneof=gain_muscle gain_strength lose_weight"`
	ImageURL   string             `json:"image_url"`
	IsPublic   bool               `json:"is_public"`
	Items      []WorkoutItemInput `json:"items"`
}

type CloneCatalogReq struct {
	WorkoutID uint `json:"workout_id" binding:"required"`
}

type UpdateMyWorkoutReq struct {
	Title      *string             `json:"title" binding:"omitempty,min=1,max=200"`
	Difficulty *string             `json:"difficulty" binding:"omitempty,oneof=novice intermediate advanced"`
	Goal       *string             `json:"goal" binding:"omitempty,oneof=gain_muscle gain_strength lose_weight"`
	ImageURL   *string             `json:"image_url"`
	IsPublic   *bool               `json:"is_public"`
	Items      *[]WorkoutItemInput `json:"items"`
}

type ListMyWorkoutsReq struct {
	Page  int    `form:"page"`
	Limit int    `form:"limit"`
	Q     string `form:"q"`
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
	ID            uint             `json:"id"`
	Title         string           `json:"title"`
	Difficulty    string           `json:"difficulty"`
	Goal          string           `json:"goal"`
	ImageURL      string           `json:"image_url,omitempty"`
	UserID        uint             `json:"user_id"`
	OwnerUserID   *uint            `json:"owner_user_id,omitempty"`
	IsPublic      bool             `json:"is_public"`
	Views         int64            `json:"views"`
	Items         []WorkoutItemRes `json:"items"`
	ExerciseCount int              `json:"exercise_count"`
	TotalSets     int              `json:"total_sets"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
}

type CreateRes struct {
	Workout WorkoutRes `json:"workout"`
}

type UpdateRes struct {
	Workout WorkoutRes `json:"workout"`
}

type GetRes struct {
	Workout WorkoutRes `json:"workout"`
}

type ListRes struct {
	Total int64        `json:"total"`
	Data  []WorkoutRes `json:"data"`
}

type DeleteRes struct {
	Status string `json:"status"`
}

type CreateRoutineReq struct {
	Title       string             `json:"title" binding:"required,min=1,max=200"`
	Description string             `json:"description"`
	ImageURL    string             `json:"image_url"`
	Difficulty  string             `json:"difficulty" binding:"required,oneof=novice intermediate advanced"`
	IsPublic    bool               `json:"is_public"`
	Items       []RoutineItemInput `json:"items"`
}

type RoutineItemInput struct {
	WorkoutID uint `json:"workout_id" binding:"required"`
}

type UpdateRoutineReq struct {
	Title       *string             `json:"title" binding:"omitempty,min=1,max=200"`
	Description *string             `json:"description"`
	ImageURL    *string             `json:"image_url"`
	Difficulty  *string             `json:"difficulty" binding:"omitempty,oneof=novice intermediate advanced"`
	IsPublic    *bool               `json:"is_public"`
	Items       *[]RoutineItemInput `json:"items"`
}

type ListMyRoutinesReq struct {
	Page  int    `form:"page"`
	Limit int    `form:"limit"`
	Q     string `form:"q"`
}
