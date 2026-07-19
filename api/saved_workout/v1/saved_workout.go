package v1

import (
	"time"

	workoutv1 "trongcon-api/api/workout/v1"
)

type SaveReq struct {
	WorkoutID uint `json:"workout_id" binding:"required"`
}

type SaveRes struct {
	Status    string    `json:"status"`
	WorkoutID uint      `json:"workout_id"`
	SavedAt   time.Time `json:"saved_at"`
}

type DeleteRes struct {
	Status string `json:"status"`
}

type ListReq struct {
	Page  int `form:"page"`
	Limit int `form:"limit"`
}

type ListRes struct {
	Total int64                  `json:"total"`
	Data  []workoutv1.WorkoutRes `json:"data"`
}

type IDsRes struct {
	WorkoutIDs []uint `json:"workout_ids"`
}
