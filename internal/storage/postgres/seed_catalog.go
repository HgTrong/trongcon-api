package postgres

import (
	"log"

	"trongcon-api/internal/entity"
	"trongcon-api/internal/pkg/slug"

	"gorm.io/gorm"
)

func seedCatalog(db *gorm.DB) error {
	if err := seedMuscles(db); err != nil {
		return err
	}
	if err := seedEquipments(db); err != nil {
		return err
	}
	if err := backfillMuscleSlugRegion(db); err != nil {
		return err
	}
	return nil
}

func backfillMuscleSlugRegion(db *gorm.DB) error {
	var emptySlug []entity.Muscle
	if err := db.Where("slug = '' OR slug IS NULL").Find(&emptySlug).Error; err != nil {
		return err
	}
	for i := range emptySlug {
		slugVal, err := allocateMuscleSlug(db, slug.FromTitle(emptySlug[i].Name), emptySlug[i].ID)
		if err != nil {
			return err
		}
		emptySlug[i].Slug = slugVal
		if emptySlug[i].Region == "" {
			emptySlug[i].Region = regionForMuscleName(emptySlug[i].Name)
		}
		if err := db.Model(&emptySlug[i]).Updates(map[string]interface{}{
			"slug":   emptySlug[i].Slug,
			"region": emptySlug[i].Region,
		}).Error; err != nil {
			return err
		}
	}
	if len(emptySlug) > 0 {
		log.Printf("seed: backfilled slug/region for %d muscles", len(emptySlug))
	}
	return nil
}

func regionForMuscleName(name string) string {
	for _, m := range seedMusclesList {
		if m.Name == name {
			return m.Region
		}
	}
	return "other"
}
