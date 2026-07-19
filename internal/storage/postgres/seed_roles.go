package postgres

import (
	"errors"
	"strings"
	"time"

	"trongcon-api/internal/entity"

	"gorm.io/gorm"
)

func ensureRole(db *gorm.DB, name string) error {
	var existing entity.Role
	err := db.Where("name = ?", name).First(&existing).Error
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	now := time.Now()
	if tableHasColumn(db, "roles", "key") {
		cols := []string{"name", "key", "created_at", "updated_at"}
		args := []interface{}{name, name, now, now}
		if tableHasColumn(db, "roles", "description") {
			cols = append(cols, "description")
			args = append(args, name)
		}
		if tableHasColumn(db, "roles", "app_id") {
			cols = append(cols, "app_id")
			args = append(args, 1)
		}
		placeholders := make([]string, len(cols))
		for i := range cols {
			placeholders[i] = "?"
		}
		sql := "INSERT INTO roles (" + strings.Join(cols, ", ") + ") VALUES (" + strings.Join(placeholders, ", ") + ")"
		return db.Exec(sql, args...).Error
	}

	return db.Create(&entity.Role{Name: name, Key: name}).Error
}
