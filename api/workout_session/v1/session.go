package v1

import "time"

type FreestyleItemInput struct {
	ExerciseID uint   `json:"exercise_id" binding:"required"`
	Sets       int    `json:"sets"`
	Reps       string `json:"reps"`
}

type CreateSessionReq struct {
	WorkoutID *uint                `json:"workout_id"`
	Title     string               `json:"title"`
	Notes     string               `json:"notes"`
	Items     []FreestyleItemInput `json:"items"` // freestyle when no workout_id
}

type UpdateSetReq struct {
	WeightKg  *float64 `json:"weight_kg"`
	Reps      *int     `json:"reps"`
	Completed *bool    `json:"completed"`
}

type AddSessionItemReq struct {
	ExerciseID uint   `json:"exercise_id" binding:"required"`
	Sets       int    `json:"sets"`
	Reps       string `json:"reps"`
}

type ListSessionsReq struct {
	Page   int    `form:"page"`
	Limit  int    `form:"limit"`
	Status string `form:"status"`
}

type SetLogRes struct {
	ID            uint       `json:"id"`
	SetNumber     int        `json:"set_number"`
	WeightKg      *float64   `json:"weight_kg"`
	Reps          *int       `json:"reps"`
	Completed     bool       `json:"completed"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type SessionItemRes struct {
	ID           uint        `json:"id"`
	ExerciseID   uint        `json:"exercise_id"`
	ExerciseName string      `json:"exercise_name"`
	SortOrder    int         `json:"sort_order"`
	TargetSets   int         `json:"target_sets"`
	TargetReps   string      `json:"target_reps"`
	Sets         []SetLogRes `json:"sets"`
	Previous     []SetLogRes `json:"previous,omitempty"`
}

type SessionRes struct {
	ID               uint             `json:"id"`
	WorkoutID        *uint            `json:"workout_id,omitempty"`
	EnrollmentSlotID *uint            `json:"enrollment_slot_id,omitempty"`
	EnrollmentID     *uint            `json:"enrollment_id,omitempty"`
	WeekIndex        *int             `json:"week_index,omitempty"`
	DayIndex         *int             `json:"day_index,omitempty"`
	Title            string           `json:"title"`
	Source           string           `json:"source"`
	Status           string           `json:"status"`
	StartedAt        time.Time        `json:"started_at"`
	CompletedAt      *time.Time       `json:"completed_at,omitempty"`
	Notes            string           `json:"notes"`
	Items            []SessionItemRes `json:"items"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
}

type CreateSessionRes struct {
	Session SessionRes `json:"session"`
}

type GetSessionRes struct {
	Session SessionRes `json:"session"`
}

type ListSessionsRes struct {
	Total int64        `json:"total"`
	Data  []SessionRes `json:"data"`
}

type PerformanceSessionRes struct {
	SessionID   uint        `json:"session_id"`
	Title       string      `json:"title"`
	PerformedAt time.Time   `json:"performed_at"`
	Volume      float64     `json:"volume"`
	BestWeight  *float64    `json:"best_weight,omitempty"`
	BestReps    *int        `json:"best_reps,omitempty"`
	Sets        []SetLogRes `json:"sets"`
}

type ExercisePerformanceRes struct {
	ExerciseID   uint                     `json:"exercise_id"`
	ExerciseName string                   `json:"exercise_name"`
	Data         []PerformanceSessionRes  `json:"data"`
}
