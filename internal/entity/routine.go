package entity

type Routine struct {
	BaseEntity
	Title       string           `json:"title" gorm:"type:varchar(200);not null;index"`
	Description string           `json:"description" gorm:"type:text"`
	ImageURL    string           `json:"image_url" gorm:"type:varchar(512)"`
	Difficulty  string           `json:"difficulty" gorm:"type:varchar(32);not null;default:novice;index"`
	UserID      uint             `json:"user_id" gorm:"not null;index"`
	User        User             `json:"user,omitempty" gorm:"foreignKey:UserID"`
	IsPublic    bool             `json:"is_public" gorm:"not null;default:false;index"`
	Views       int64            `json:"views" gorm:"not null;default:0"`
	Items       []RoutineWorkout `json:"items,omitempty" gorm:"foreignKey:RoutineID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

type RoutineWorkout struct {
	BaseEntity
	RoutineID   uint   `json:"routine_id" gorm:"not null;index"`
	WorkoutID   uint   `json:"workout_id" gorm:"not null;index"`
	Workout     Workout `json:"workout,omitempty" gorm:"foreignKey:WorkoutID"`
	WorkoutTitle string `json:"workout_title" gorm:"type:varchar(200);not null"`
	SortOrder   int    `json:"sort_order" gorm:"not null;default:0"`
}

type Workout struct {
	BaseEntity
	Title        string        `json:"title" gorm:"type:varchar(200);not null;index"`
	Difficulty   string        `json:"difficulty" gorm:"type:varchar(32);not null;default:novice;index"`
	Goal         string        `json:"goal" gorm:"type:varchar(32);not null;default:gain_muscle;index"`
	ImageURL     string        `json:"image_url" gorm:"type:varchar(512)"`
	UserID       uint          `json:"user_id" gorm:"index"` // poster / author (0 = unset legacy)
	User         User          `json:"user,omitempty" gorm:"foreignKey:UserID"`
	OwnerUserID  *uint         `json:"owner_user_id,omitempty" gorm:"index"` // nil = catalog; set = personal copy
	IsPublic     bool          `json:"is_public" gorm:"not null;default:false;index"`
	Views        int64         `json:"views" gorm:"not null;default:0"`
	Items        []WorkoutItem `json:"items,omitempty" gorm:"foreignKey:WorkoutID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

type WorkoutItem struct {
	BaseEntity
	WorkoutID    uint   `json:"workout_id" gorm:"not null;index"`
	ExerciseID   uint   `json:"exercise_id" gorm:"not null;index"`
	ExerciseName string `json:"exercise_name" gorm:"type:varchar(200);not null"`
	SortOrder    int    `json:"sort_order" gorm:"not null;default:0"`
	Sets         int    `json:"sets" gorm:"not null;default:3"`
	Reps         string `json:"reps" gorm:"type:varchar(32);not null;default:'10'"`
}
