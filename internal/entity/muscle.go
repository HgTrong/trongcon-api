package entity

type Muscle struct {
	BaseEntity
	Name   string `json:"name" gorm:"type:varchar(200);not null;index"`
	Slug   string `json:"slug" gorm:"type:varchar(220);uniqueIndex"`
	Region string `json:"region" gorm:"type:varchar(32);not null;default:other;index"`
}
