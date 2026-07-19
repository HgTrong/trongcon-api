package entity

import "time"

const (
	EnrollmentStatusActive    = "active"
	EnrollmentStatusCompleted = "completed"
	EnrollmentStatusCancelled = "cancelled"
)

// TrainingEnrollment applies a routine for N weeks.
type TrainingEnrollment struct {
	BaseEntity
	UserID    uint      `json:"user_id" gorm:"not null;index"`
	RoutineID *uint     `json:"routine_id" gorm:"index"`
	Title     string    `json:"title" gorm:"type:varchar(200);not null"`
	StartDate time.Time `json:"start_date" gorm:"type:date;not null;index"`
	Weeks     int       `json:"weeks" gorm:"not null"`
	Status    string    `json:"status" gorm:"type:varchar(20);not null;default:active;index"`
	Slots     []EnrollmentSlot `json:"slots,omitempty" gorm:"foreignKey:EnrollmentID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

// EnrollmentSlot is one scheduled day within an enrollment week.
type EnrollmentSlot struct {
	BaseEntity
	EnrollmentID uint   `json:"enrollment_id" gorm:"not null;index:idx_enrollment_slot_week"`
	WeekIndex    int    `json:"week_index" gorm:"not null;index:idx_enrollment_slot_week"` // 1..N
	DayIndex     int    `json:"day_index" gorm:"not null"`                                  // 0..len-1
	WorkoutID    *uint  `json:"workout_id" gorm:"index"`
	WorkoutTitle string `json:"workout_title" gorm:"type:varchar(200);not null"`
	SessionID    *uint  `json:"session_id" gorm:"index"`
}
