package postgres

import (
	"log"

	"trongcon-api/internal/entity"

	"gorm.io/gorm"
)

// Danh sách equipment theo MuscleWiki (browse filter).
var seedEquipmentNames = []string{
	"Featured",
	"Dumbbells",
	"Machine",
	"Kettlebells",
	"Cables",
	"Plate",
	"Yoga",
	"Cardio",
	"Recovery",
	"Barbell",
	"Bodyweight",
	"Medicine Ball",
	"Stretches",
	"Band",
	"TRX",
	"Bosu Ball",
	"Smith Machine",
	"Pilates",
}

func seedEquipments(db *gorm.DB) error {
	created := 0
	for _, name := range seedEquipmentNames {
		var n int64
		if err := db.Model(&entity.Equipment{}).Where("name = ?", name).Count(&n).Error; err != nil {
			return err
		}
		if n > 0 {
			continue
		}
		if err := db.Create(&entity.Equipment{Name: name}).Error; err != nil {
			return err
		}
		created++
	}
	if created > 0 {
		log.Printf("seed: added %d equipments (MuscleWiki list)", created)
	}
	return nil
}
