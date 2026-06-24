package apimap

import (
	foodv1 "trongcon-api/api/food/v1"
	"trongcon-api/internal/entity"
)

func FoodToRes(f *entity.Food) foodv1.FoodRes {
	return foodv1.FoodRes{
		ID:           f.ID,
		Name:         f.Name,
		Protein:      f.Protein,
		Carb:         f.Carb,
		Fat:          f.Fat,
		Calories:     f.Calories,
		ServingSizeG: f.ServingSizeG,
		CreatedAt:    f.CreatedAt,
		UpdatedAt:    f.UpdatedAt,
	}
}
