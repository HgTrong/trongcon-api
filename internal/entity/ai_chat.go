package entity

type AiChatThread struct {
	BaseEntity
	UserID uint   `json:"user_id" gorm:"not null;index"`
	Title  string `json:"title" gorm:"type:varchar(200);not null;default:''"`
}

type AiChatMessage struct {
	BaseEntity
	ThreadID uint   `json:"thread_id" gorm:"not null;index"`
	Role     string `json:"role" gorm:"type:varchar(20);not null"` // user | assistant | system
	Content  string `json:"content" gorm:"type:text;not null"`
}
