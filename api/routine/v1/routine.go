package v1

import "time"

type RoutineItemInput struct {
	WorkoutID uint `json:"workout_id" binding:"required"`
}

type CreateReq struct {
	Title       string             `json:"title" binding:"required,min=1,max=200"`
	Description string             `json:"description"`
	Difficulty  string             `json:"difficulty" binding:"required,oneof=novice intermediate advanced"`
	UserID      uint               `json:"user_id" binding:"required"`
	IsPublic    bool               `json:"is_public"`
	Items       []RoutineItemInput `json:"items"`
}

type CreateRes struct {
	Routine RoutineRes `json:"routine"`
}

type UpdateReq struct {
	Title       *string             `json:"title" binding:"omitempty,min=1,max=200"`
	Description *string             `json:"description"`
	Difficulty  *string             `json:"difficulty" binding:"omitempty,oneof=novice intermediate advanced"`
	UserID      *uint               `json:"user_id"`
	IsPublic    *bool               `json:"is_public"`
	Items       *[]RoutineItemInput `json:"items"`
}

type UpdateRes struct {
	Routine RoutineRes `json:"routine"`
}

type GetRes struct {
	Routine RoutineRes `json:"routine"`
}

type ListReq struct {
	Page       int    `form:"page"`
	Limit      int    `form:"limit"`
	Q          string `form:"q"`
	UserID     *uint  `form:"user_id"`
	IsPublic   string `form:"is_public"`
	Difficulty string `form:"difficulty"`
	OrderBy    string `form:"order_by"`
	OrderDir   string `form:"order_dir"`
}

type ListRes struct {
	Total int64        `json:"total"`
	Data  []RoutineRes `json:"data"`
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

type RoutineWorkoutRes struct {
	ID            uint             `json:"id"`
	WorkoutID     uint             `json:"workout_id"`
	WorkoutTitle  string           `json:"workout_title"`
	Difficulty    string           `json:"difficulty,omitempty"`
	SortOrder     int              `json:"sort_order"`
	ExerciseCount int              `json:"exercise_count"`
	TotalSets     int              `json:"total_sets"`
	Items         []WorkoutItemRes `json:"items,omitempty"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
}

type RoutineRes struct {
	ID            uint                `json:"id"`
	Title         string              `json:"title"`
	Description   string              `json:"description"`
	Difficulty    string              `json:"difficulty"`
	UserID        uint                `json:"user_id"`
	UserEmail     string              `json:"user_email,omitempty"`
	IsPublic      bool                `json:"is_public"`
	Items         []RoutineWorkoutRes `json:"items"`
	WorkoutCount  int                 `json:"workout_count"`
	ExerciseCount int                 `json:"exercise_count"`
	TotalSets     int                 `json:"total_sets"`
	CreatedAt     time.Time           `json:"created_at"`
	UpdatedAt     time.Time           `json:"updated_at"`
}
