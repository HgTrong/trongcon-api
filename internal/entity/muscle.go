package entity

type Muscle struct {
	BaseEntity
	Name string `json:"name" gorm:"type:varchar(200);not null;index"`
}
