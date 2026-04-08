package entity

// Equipment: icon là URL ảnh (upload hoặc nhập tay).
type Equipment struct {
	BaseEntity
	Name string `json:"name" gorm:"type:varchar(200);not null;index"`
	Icon string `json:"icon" gorm:"type:varchar(512)"`
}
