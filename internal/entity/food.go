package entity

// Food: macro lưu dạng field riêng để query/sort/tính tổng thuận tiện.
type Food struct {
	BaseEntity
	Name           string  `json:"name" gorm:"type:varchar(200);not null;index"`
	Protein        float64 `json:"protein" gorm:"type:numeric(10,2);not null;default:0"` // gram
	Carb           float64 `json:"carb" gorm:"type:numeric(10,2);not null;default:0"`       // gram
	Fat            float64 `json:"fat" gorm:"type:numeric(10,2);not null;default:0"`         // gram
	Calories       float64 `json:"calories" gorm:"type:numeric(10,2);not null;default:0"`    // kcal
	ServingSizeG   float64 `json:"serving_size_g" gorm:"type:numeric(10,2);not null;default:100"`
}
