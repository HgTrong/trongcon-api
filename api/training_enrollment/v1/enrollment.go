package v1

import "time"

type CreateEnrollmentReq struct {
	RoutineID *uint  `json:"routine_id"`
	WorkoutID *uint  `json:"workout_id"`
	StartDate string `json:"start_date" binding:"required"` // YYYY-MM-DD
	Weeks     int    `json:"weeks" binding:"required,min=1,max=52"`
}

type ListEnrollmentsReq struct {
	Page   int    `form:"page"`
	Limit  int    `form:"limit"`
	Status string `form:"status"`
}

type SlotRes struct {
	ID           uint       `json:"id"`
	WeekIndex    int        `json:"week_index"`
	DayIndex     int        `json:"day_index"`
	WorkoutID    *uint      `json:"workout_id,omitempty"`
	WorkoutTitle string     `json:"workout_title"`
	SessionID    *uint      `json:"session_id,omitempty"`
	SessionStatus string    `json:"session_status,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
}

type EnrollmentRes struct {
	ID        uint      `json:"id"`
	RoutineID *uint     `json:"routine_id,omitempty"`
	Title     string    `json:"title"`
	StartDate time.Time `json:"start_date"`
	Weeks     int       `json:"weeks"`
	Status    string    `json:"status"`
	Slots     []SlotRes `json:"slots"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateEnrollmentRes struct {
	Enrollment EnrollmentRes `json:"enrollment"`
}

type GetEnrollmentRes struct {
	Enrollment EnrollmentRes `json:"enrollment"`
}

type ListEnrollmentsRes struct {
	Total int64           `json:"total"`
	Data  []EnrollmentRes `json:"data"`
}

type WeekExerciseStat struct {
	ExerciseID   uint     `json:"exercise_id"`
	ExerciseName string   `json:"exercise_name"`
	Volume       float64  `json:"volume"`
	BestWeight   *float64 `json:"best_weight,omitempty"`
	BestReps     *int     `json:"best_reps,omitempty"`
	DeltaVolume  *float64 `json:"delta_volume,omitempty"`
}

type WeekCompareRes struct {
	WeekIndex  int                `json:"week_index"`
	TotalVolume float64           `json:"total_volume"`
	Exercises  []WeekExerciseStat `json:"exercises"`
}

type EnrollmentCompareRes struct {
	EnrollmentID uint             `json:"enrollment_id"`
	Weeks        []WeekCompareRes `json:"weeks"`
}
