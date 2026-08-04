package entity

type MealPlan struct {
	BaseEntity
	Title       string         `json:"title" gorm:"type:varchar(200);not null;index"`
	Description string         `json:"description" gorm:"type:text"`
	UserID      uint           `json:"user_id" gorm:"not null;index"`
	User        User           `json:"user,omitempty" gorm:"foreignKey:UserID"`
	IsPublic    bool           `json:"is_public" gorm:"not null;default:false;index"`
	Views       int64          `json:"views" gorm:"not null;default:0"`
	Meals       []MealPlanMeal `json:"meals,omitempty" gorm:"foreignKey:MealPlanID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

type MealPlanMeal struct {
	BaseEntity
	MealPlanID uint           `json:"meal_plan_id" gorm:"not null;index"`
	Name       string         `json:"name" gorm:"type:varchar(100);not null"`
	SortOrder  int            `json:"sort_order" gorm:"not null;default:0"`
	Items      []MealPlanItem `json:"items,omitempty" gorm:"foreignKey:MealPlanMealID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

type MealPlanItem struct {
	BaseEntity
	MealPlanMealID uint    `json:"meal_plan_meal_id" gorm:"not null;index"`
	FoodID         uint    `json:"food_id" gorm:"not null;index"`
	FoodName       string  `json:"food_name" gorm:"type:varchar(200);not null"`
	Quantity       float64 `json:"quantity" gorm:"type:numeric(10,2);not null;default:1"`
	ServingSizeG   float64 `json:"serving_size_g" gorm:"type:numeric(10,2);not null;default:100"`
	Protein        float64 `json:"protein" gorm:"type:numeric(10,2);not null;default:0"`
	Carb           float64 `json:"carb" gorm:"type:numeric(10,2);not null;default:0"`
	Fat            float64 `json:"fat" gorm:"type:numeric(10,2);not null;default:0"`
	Calories       float64 `json:"calories" gorm:"type:numeric(10,2);not null;default:0"`
}
