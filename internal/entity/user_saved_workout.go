package entity

// UserSavedWorkout bookmarks a catalog workout for a member.
type UserSavedWorkout struct {
	BaseEntity
	UserID    uint    `json:"user_id" gorm:"not null;index;uniqueIndex:idx_user_saved_workout"`
	WorkoutID uint    `json:"workout_id" gorm:"not null;index;uniqueIndex:idx_user_saved_workout"`
	User      User    `json:"-" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Workout   Workout `json:"workout,omitempty" gorm:"foreignKey:WorkoutID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
