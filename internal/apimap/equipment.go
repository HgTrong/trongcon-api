package apimap

import (
	equipmentv1 "trongcon-api/api/equipment/v1"
	"trongcon-api/internal/entity"
)

func EquipmentToRes(e *entity.Equipment) equipmentv1.EquipmentRes {
	return equipmentv1.EquipmentRes{
		ID:        e.ID,
		Name:      e.Name,
		Icon:      e.Icon,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}
}
