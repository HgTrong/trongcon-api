package entity

import "time"

const (
	SessionStatusInProgress = "in_progress"
	SessionStatusCompleted  = "completed"
	SessionStatusAbandoned  = "abandoned"

	SessionSourceCatalog     = "catalog"
	SessionSourcePersonal    = "personal"
	SessionSourceFreestyle   = "freestyle"
	SessionSourceEnrollment  = "enrollment"
)

// WorkoutSession is one performed training bout for a user.
type WorkoutSession struct {
	BaseEntity
	UserID            uint       `json:"user_id" gorm:"not null;index:idx_workout_session_user"`
	WorkoutID         *uint      `json:"workout_id" gorm:"index"`
	EnrollmentSlotID  *uint      `json:"enrollment_slot_id" gorm:"index"`
	Title             string     `json:"title" gorm:"type:varchar(200);not null"`
	Source            string     `json:"source" gorm:"type:varchar(32);not null;default:freestyle;index"`
	Status            string     `json:"status" gorm:"type:varchar(20);not null;default:in_progress;index"`
	StartedAt         time.Time  `json:"started_at" gorm:"not null;index"`
	CompletedAt       *time.Time `json:"completed_at"`
	Notes             string     `json:"notes" gorm:"type:text"`
	Items             []WorkoutSessionItem `json:"items,omitempty" gorm:"foreignKey:SessionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

// WorkoutSessionItem is one exercise block inside a session.
type WorkoutSessionItem struct {
	BaseEntity
	SessionID    uint   `json:"session_id" gorm:"not null;index"`
	WorkoutItemID *uint `json:"workout_item_id" gorm:"index"`
	ExerciseID   uint   `json:"exercise_id" gorm:"not null;index"`
	ExerciseName string `json:"exercise_name" gorm:"type:varchar(200);not null"`
	SortOrder    int    `json:"sort_order" gorm:"not null;default:0"`
	TargetSets   int    `json:"target_sets" gorm:"not null;default:0"`
	TargetReps   string `json:"target_reps" gorm:"type:varchar(32)"`
	Sets         []WorkoutSetLog `json:"sets,omitempty" gorm:"foreignKey:SessionItemID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

// WorkoutSetLog stores actual weight/reps for one set.
type WorkoutSetLog struct {
	BaseEntity
	SessionItemID uint       `json:"session_item_id" gorm:"not null;index"`
	SetNumber     int        `json:"set_number" gorm:"not null"`
	WeightKg      *float64   `json:"weight_kg" gorm:"type:numeric(10,2)"`
	Reps          *int       `json:"reps"`
	Completed     bool       `json:"completed" gorm:"not null;default:false"`
	CompletedAt   *time.Time `json:"completed_at"`
}
