package apimap

import (
	musclev1 "trongcon-api/api/muscle/v1"
	"trongcon-api/internal/entity"
)

func MuscleToRes(m *entity.Muscle) musclev1.MuscleRes {
	return musclev1.MuscleRes{
		ID:        m.ID,
		Name:      m.Name,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}
