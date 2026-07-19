package postgres

import (
	"fmt"
	"log"

	"trongcon-api/internal/entity"
	"trongcon-api/internal/pkg/slug"

	"gorm.io/gorm"
)

type seedMuscle struct {
	Name   string
	Region string
}

// Danh sách muscles theo MuscleWiki.
var seedMusclesList = []seedMuscle{
	{Name: "Biceps", Region: "arms"},
	{Name: "Long Head Bicep", Region: "arms"},
	{Name: "Short Head Bicep", Region: "arms"},
	{Name: "Traps (mid-back)", Region: "back"},
	{Name: "Lower back", Region: "back"},
	{Name: "Abdominals", Region: "core"},
	{Name: "Lower Abdominals", Region: "core"},
	{Name: "Upper Abdominals", Region: "core"},
	{Name: "Calves", Region: "legs"},
	{Name: "Tibialis", Region: "legs"},
	{Name: "Soleus", Region: "legs"},
	{Name: "Gastrocnemius", Region: "legs"},
	{Name: "Forearms", Region: "arms"},
	{Name: "Wrist Extensors", Region: "arms"},
	{Name: "Wrist Flexors", Region: "arms"},
	{Name: "Glutes", Region: "legs"},
	{Name: "Gluteus Medius", Region: "legs"},
	{Name: "Gluteus Maximus", Region: "legs"},
	{Name: "Hamstrings", Region: "legs"},
	{Name: "Medial Hamstrings", Region: "legs"},
	{Name: "Lateral Hamstrings", Region: "legs"},
	{Name: "Lats", Region: "back"},
	{Name: "Shoulders", Region: "shoulders"},
	{Name: "Lateral Deltoid", Region: "shoulders"},
	{Name: "Anterior Deltoid", Region: "shoulders"},
	{Name: "Posterior Deltoid", Region: "shoulders"},
	{Name: "Triceps", Region: "arms"},
	{Name: "Long Head Tricep", Region: "arms"},
	{Name: "Lateral Head Triceps", Region: "arms"},
	{Name: "Medial Head Triceps", Region: "arms"},
	{Name: "Traps", Region: "back"},
	{Name: "Upper Traps", Region: "back"},
	{Name: "Lower Traps", Region: "back"},
	{Name: "Quads", Region: "legs"},
	{Name: "Inner Thigh", Region: "legs"},
	{Name: "Inner Quadriceps", Region: "legs"},
	{Name: "Outer Quadricep", Region: "legs"},
	{Name: "Rectus Femoris", Region: "legs"},
	{Name: "Chest", Region: "chest"},
	{Name: "Upper Pectoralis", Region: "chest"},
	{Name: "Mid and Lower Chest", Region: "chest"},
	{Name: "Obliques", Region: "core"},
	{Name: "Hands", Region: "arms"},
	{Name: "Feet", Region: "legs"},
	{Name: "Neck", Region: "other"},
	{Name: "Groin", Region: "legs"},
}

func seedMuscles(db *gorm.DB) error {
	created := 0
	for _, m := range seedMusclesList {
		var n int64
		if err := db.Model(&entity.Muscle{}).Where("name = ?", m.Name).Count(&n).Error; err != nil {
			return err
		}
		if n > 0 {
			continue
		}
		slugVal, err := allocateMuscleSlug(db, slug.FromTitle(m.Name), 0)
		if err != nil {
			return err
		}
		if err := db.Create(&entity.Muscle{
			Name:   m.Name,
			Slug:   slugVal,
			Region: m.Region,
		}).Error; err != nil {
			return err
		}
		created++
	}
	if created > 0 {
		log.Printf("seed: added %d muscles (MuscleWiki list)", created)
	}
	return nil
}

func allocateMuscleSlug(db *gorm.DB, base string, excludeID uint) (string, error) {
	if base == "" {
		base = "muscle"
	}
	for n := 0; n < 10000; n++ {
		candidate := base
		if n > 0 {
			candidate = fmt.Sprintf("%s-%d", base, n)
		}
		var count int64
		q := db.Unscoped().Model(&entity.Muscle{}).Where("slug = ?", candidate)
		if excludeID > 0 {
			q = q.Where("id <> ?", excludeID)
		}
		if err := q.Count(&count).Error; err != nil {
			return "", err
		}
		if count == 0 {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("could not allocate unique slug for %q", base)
}
