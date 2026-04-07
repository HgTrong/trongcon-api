package entity

// Category: icon/image là URL (sau khi upload S3 hoặc nhập tay).
type Category struct {
	BaseEntity
	Name   string `json:"name" gorm:"type:varchar(200);not null;index"`
	Icon   string `json:"icon" gorm:"type:varchar(512)"`
	Image  string `json:"image" gorm:"type:varchar(512)"`
	Status string `json:"status" gorm:"type:varchar(32);not null;default:active;index"`
	Type   string `json:"type" gorm:"type:varchar(64);not null;index"`
}
