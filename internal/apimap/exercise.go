package apimap

import (
	"sort"

	exercisev1 "trongcon-api/api/exercise/v1"
	"trongcon-api/internal/entity"
)

func ExerciseToRes(ex *entity.Exercise) exercisev1.ExerciseRes {
	res := exercisev1.ExerciseRes{
		ID:          ex.ID,
		Name:        ex.Name,
		Slug:        ex.Slug,
		Summary:     ex.Summary,
		Difficulty:  ex.Difficulty,
		Force:       ex.Force,
		Grips:       ex.Grips,
		Mechanic:    ex.Mechanic,
		DemoGif1:    ex.DemoGif1,
		DemoGif2:    ex.DemoGif2,
		VideoURL:    ex.VideoURL,
		Thumbnail:   ex.Thumbnail,
		Content:     ex.Content,
		Status:      ex.Status,
		EquipmentID: ex.EquipmentID,
		CreatedAt:   ex.CreatedAt,
		UpdatedAt:   ex.UpdatedAt,
		Steps:       make([]exercisev1.ExerciseStepRes, 0, len(ex.Steps)),
		Muscles:     make([]exercisev1.ExerciseMuscleRes, 0, len(ex.Muscles)),
	}
	if ex.Equipment != nil && ex.Equipment.ID > 0 {
		res.EquipmentName = ex.Equipment.Name
	}
	steps := append([]entity.ExerciseStep(nil), ex.Steps...)
	sort.Slice(steps, func(i, j int) bool {
		if steps[i].SortOrder == steps[j].SortOrder {
			return steps[i].ID < steps[j].ID
		}
		return steps[i].SortOrder < steps[j].SortOrder
	})
	for _, st := range steps {
		res.Steps = append(res.Steps, exercisev1.ExerciseStepRes{
			ID:        st.ID,
			SortOrder: st.SortOrder,
			Content:   st.Content,
			CreatedAt: st.CreatedAt,
			UpdatedAt: st.UpdatedAt,
		})
	}
	for _, m := range ex.Muscles {
		name := ""
		if m.Muscle.ID > 0 {
			name = m.Muscle.Name
		}
		res.Muscles = append(res.Muscles, exercisev1.ExerciseMuscleRes{
			MuscleID:   m.MuscleID,
			MuscleName: name,
			Role:       m.Role,
		})
		switch m.Role {
		case "primary":
			res.PrimaryMuscleIDs = append(res.PrimaryMuscleIDs, m.MuscleID)
		case "secondary":
			res.SecondaryMuscleIDs = append(res.SecondaryMuscleIDs, m.MuscleID)
		case "tertiary":
			res.TertiaryMuscleIDs = append(res.TertiaryMuscleIDs, m.MuscleID)
		}
	}
	return res
}
