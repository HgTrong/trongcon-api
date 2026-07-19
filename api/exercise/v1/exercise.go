package v1

import "time"

type StepInput struct {
	Content string `json:"content" binding:"required,min=1"`
}

type CreateReq struct {
	Name               string      `json:"name" binding:"required,min=1,max=200"`
	Summary            string      `json:"summary" binding:"max=500"`
	Difficulty         string      `json:"difficulty" binding:"required,oneof=novice intermediate advanced"`
	Force              string      `json:"force" binding:"required,oneof=pull push static"`
	Grips              string      `json:"grips" binding:"required,oneof=overhand underhand neutral mixed"`
	Mechanic           string      `json:"mechanic" binding:"required,oneof=compound isolation"`
	DemoVideo1         string      `json:"demo_video_1"`
	DemoVideo2         string      `json:"demo_video_2"`
	VideoURL           string      `json:"video_url"`
	Thumbnail          string      `json:"thumbnail"`
	Content            string      `json:"content"`
	Status             string      `json:"status" binding:"omitempty,oneof=active inactive"`
	EquipmentID        *uint       `json:"equipment_id"`
	Steps              []StepInput `json:"steps"`
	PrimaryMuscleIDs   []uint      `json:"primary_muscle_ids"`
	SecondaryMuscleIDs []uint      `json:"secondary_muscle_ids"`
	TertiaryMuscleIDs  []uint      `json:"tertiary_muscle_ids"`
}

type CreateRes struct {
	Exercise ExerciseRes `json:"exercise"`
}

type UpdateReq struct {
	Name               *string      `json:"name" binding:"omitempty,min=1,max=200"`
	Summary            *string      `json:"summary" binding:"omitempty,max=500"`
	Difficulty         *string      `json:"difficulty" binding:"omitempty,oneof=novice intermediate advanced"`
	Force              *string      `json:"force" binding:"omitempty,oneof=pull push static"`
	Grips              *string      `json:"grips" binding:"omitempty,oneof=overhand underhand neutral mixed"`
	Mechanic           *string      `json:"mechanic" binding:"omitempty,oneof=compound isolation"`
	DemoVideo1           *string      `json:"demo_video_1"`
	DemoVideo2           *string      `json:"demo_video_2"`
	VideoURL           *string      `json:"video_url"`
	Thumbnail          *string      `json:"thumbnail"`
	Content            *string      `json:"content"`
	Status             *string      `json:"status" binding:"omitempty,oneof=active inactive"`
	EquipmentID        *uint        `json:"equipment_id"`
	Steps              *[]StepInput `json:"steps"`
	PrimaryMuscleIDs   *[]uint      `json:"primary_muscle_ids"`
	SecondaryMuscleIDs *[]uint      `json:"secondary_muscle_ids"`
	TertiaryMuscleIDs  *[]uint      `json:"tertiary_muscle_ids"`
}

type UpdateRes struct {
	Exercise ExerciseRes `json:"exercise"`
}

type GetRes struct {
	Exercise ExerciseRes `json:"exercise"`
}

type ListReq struct {
	Page        int    `form:"page"`
	Limit       int    `form:"limit"`
	Q           string `form:"q"`
	Difficulty  string `form:"difficulty"`
	Force       string `form:"force"`
	Mechanic    string `form:"mechanic"`
	Status      string `form:"status"`
	EquipmentID *uint  `form:"equipment_id"`
	MuscleID    *uint  `form:"muscle_id"`
	OrderBy     string `form:"order_by"`
	OrderDir    string `form:"order_dir"`
}

type ListRes struct {
	Total int64         `json:"total"`
	Data  []ExerciseRes `json:"data"`
}

type DeleteRes struct {
	Status string `json:"status"`
}

type IncrementViewsRes struct {
	Status string `json:"status"`
	Views  int    `json:"views"`
}

type ExerciseStepRes struct {
	ID        uint      `json:"id"`
	SortOrder int       `json:"sort_order"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ExerciseMuscleRes struct {
	MuscleID   uint   `json:"muscle_id"`
	MuscleName string `json:"muscle_name"`
	Role       string `json:"role"`
}

type ExerciseRes struct {
	ID                 uint                `json:"id"`
	Name               string              `json:"name"`
	Slug               string              `json:"slug"`
	Summary            string              `json:"summary"`
	Difficulty         string              `json:"difficulty"`
	Force              string              `json:"force"`
	Grips              string              `json:"grips"`
	Mechanic           string              `json:"mechanic"`
	DemoVideo1           string              `json:"demo_video_1"`
	DemoVideo2           string              `json:"demo_video_2"`
	VideoURL           string              `json:"video_url"`
	Thumbnail          string              `json:"thumbnail"`
	Content            string              `json:"content"`
	Status             string              `json:"status"`
	Views              int                 `json:"views"`
	EquipmentID        *uint               `json:"equipment_id,omitempty"`
	EquipmentName      string              `json:"equipment_name,omitempty"`
	Steps              []ExerciseStepRes   `json:"steps"`
	PrimaryMuscleIDs   []uint              `json:"primary_muscle_ids"`
	SecondaryMuscleIDs []uint              `json:"secondary_muscle_ids"`
	TertiaryMuscleIDs  []uint              `json:"tertiary_muscle_ids"`
	Muscles            []ExerciseMuscleRes `json:"muscles"`
	CreatedAt          time.Time           `json:"created_at"`
	UpdatedAt          time.Time           `json:"updated_at"`
}
