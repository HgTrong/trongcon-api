package entity

// Exercise: bài tập chi tiết (MuscleWiki-style).
type Exercise struct {
	BaseEntity
	Name        string         `json:"name" gorm:"type:varchar(200);not null;index"`
	Slug        string         `json:"slug" gorm:"type:varchar(220);uniqueIndex"`
	Summary     string         `json:"summary" gorm:"type:varchar(500)"`
	Difficulty  string         `json:"difficulty" gorm:"type:varchar(32);not null;index"`
	Force       string         `json:"force" gorm:"type:varchar(32);not null;index"`
	Grips       string         `json:"grips" gorm:"type:varchar(32);not null"`
	Mechanic    string         `json:"mechanic" gorm:"type:varchar(32);not null;index"`
	DemoVideo1   string         `json:"demo_video_1" gorm:"column:demo_video_1;type:varchar(512)"`
	DemoVideo2   string         `json:"demo_video_2" gorm:"column:demo_video_2;type:varchar(512)"`
	VideoURL    string         `json:"video_url" gorm:"type:varchar(512)"`
	Thumbnail   string         `json:"thumbnail" gorm:"type:varchar(512)"`
	Content     string         `json:"content" gorm:"type:text"`
	Status      string         `json:"status" gorm:"type:varchar(20);not null;default:active;index"`
	Views       int            `json:"views" gorm:"not null;default:0;index"`
	EquipmentID *uint          `json:"equipment_id" gorm:"index"`
	Equipment   *Equipment     `json:"equipment,omitempty" gorm:"foreignKey:EquipmentID"`
	Steps       []ExerciseStep `json:"steps,omitempty" gorm:"foreignKey:ExerciseID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Muscles     []ExerciseMuscle `json:"muscles,omitempty" gorm:"foreignKey:ExerciseID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

type ExerciseStep struct {
	BaseEntity
	ExerciseID uint   `json:"exercise_id" gorm:"not null;index"`
	SortOrder  int    `json:"sort_order" gorm:"not null;default:0"`
	Content    string `json:"content" gorm:"type:text;not null"`
}

type ExerciseMuscle struct {
	ExerciseID uint   `json:"exercise_id" gorm:"primaryKey"`
	MuscleID   uint   `json:"muscle_id" gorm:"primaryKey"`
	Role       string `json:"role" gorm:"type:varchar(20);not null;index"`
	Muscle     Muscle `json:"muscle,omitempty" gorm:"foreignKey:MuscleID"`
}
