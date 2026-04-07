package entity

// Article: content lưu HTML/rich text (TEXT).
type Article struct {
	BaseEntity
	Title      string   `json:"title" gorm:"type:varchar(500);not null;index"`
	Subtitle   string   `json:"subtitle" gorm:"type:varchar(1000)"`
	Slug       string   `json:"slug" gorm:"type:varchar(220);uniqueIndex"`
	Thumbnail  string   `json:"thumbnail" gorm:"type:varchar(512)"`
	Video      string   `json:"video" gorm:"type:varchar(512)"`
	Content    string   `json:"content" gorm:"type:text;not null"`
	UserID     uint     `json:"user_id" gorm:"not null;index"`
	CategoryID uint     `json:"category_id" gorm:"not null;index"`
	User       User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Category   Category `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
}
