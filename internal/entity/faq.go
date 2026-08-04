package entity

// FAQ is a simple Q&A item for the marketing site.
type FAQ struct {
	BaseEntity
	Question  string `json:"question" gorm:"type:text;not null"`
	Answer    string `json:"answer" gorm:"type:text;not null"`
	SortOrder int    `json:"sort_order" gorm:"not null;default:0;index"`
	IsActive  bool   `json:"is_active" gorm:"not null;default:true;index"`
}

func (FAQ) TableName() string { return "faqs" }
