package postgres

import (
	"log"

	"trongcon-api/internal/entity"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// patchLegacySchema xử lý DB cũ trước AutoMigrate (tránh ADD NOT NULL khi đã có row NULL).
func patchLegacySchema(db *gorm.DB) error {
	if err := patchRolesLegacy(db); err != nil {
		return err
	}
	if err := patchUsersPasswordHash(db); err != nil {
		return err
	}
	if err := patchMusclesSlugRegion(db); err != nil {
		return err
	}
	if err := patchExerciseDemoVideoColumns(db); err != nil {
		return err
	}
	return nil
}

func tableHasColumn(db *gorm.DB, table, column string) bool {
	var n int64
	db.Raw(`
		SELECT COUNT(*) FROM information_schema.columns
		WHERE table_schema = CURRENT_SCHEMA() AND table_name = ? AND column_name = ?
	`, table, column).Scan(&n)
	return n > 0
}

// patchRolesLegacy — DB cũ (vd. strongbody) có cột key/description/app_id mà model mới không dùng hết.
func patchRolesLegacy(db *gorm.DB) error {
	if !db.Migrator().HasTable(&entity.Role{}) {
		return nil
	}
	if tableHasColumn(db, "roles", "key") {
		if err := db.Exec(`UPDATE roles SET key = name WHERE key IS NULL OR key = ''`).Error; err != nil {
			return err
		}
	}
	if tableHasColumn(db, "roles", "description") {
		if err := db.Exec(`UPDATE roles SET description = COALESCE(NULLIF(description, ''), name, '') WHERE description IS NULL`).Error; err != nil {
			return err
		}
	}
	if tableHasColumn(db, "roles", "app_id") {
		if err := db.Exec(`UPDATE roles SET app_id = 1 WHERE app_id IS NULL`).Error; err != nil {
			return err
		}
	}
	return nil
}

func patchUsersPasswordHash(db *gorm.DB) error {
	if !db.Migrator().HasTable(&entity.User{}) {
		return nil
	}
	if !db.Migrator().HasColumn(&entity.User{}, "password_hash") {
		if err := db.Exec(`ALTER TABLE users ADD COLUMN password_hash varchar(255)`).Error; err != nil {
			return err
		}
	}

	var n int64
	if err := db.Model(&entity.User{}).Where("password_hash IS NULL OR password_hash = ''").Count(&n).Error; err != nil {
		return err
	}
	if n == 0 {
		return nil
	}

	// Hash không dùng được để login — user cũ cần đổi mật khẩu qua admin hoặc tạo lại tài khoản.
	hash, err := bcrypt.GenerateFromPassword([]byte("__legacy_user_reset_required__"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if err := db.Exec(
		`UPDATE users SET password_hash = ? WHERE password_hash IS NULL OR password_hash = ''`,
		string(hash),
	).Error; err != nil {
		return err
	}
	log.Printf("migrate: backfilled password_hash for %d user(s) — reset password if login fails", n)
	return nil
}

func patchMusclesSlugRegion(db *gorm.DB) error {
	if !db.Migrator().HasTable(&entity.Muscle{}) {
		return nil
	}
	if !db.Migrator().HasColumn(&entity.Muscle{}, "slug") {
		if err := db.Exec(`ALTER TABLE muscles ADD COLUMN slug varchar(220)`).Error; err != nil {
			return err
		}
	}
	if !db.Migrator().HasColumn(&entity.Muscle{}, "region") {
		if err := db.Exec(`ALTER TABLE muscles ADD COLUMN region varchar(32) DEFAULT 'other'`).Error; err != nil {
			return err
		}
	}
	return nil
}

func patchExerciseDemoVideoColumns(db *gorm.DB) error {
	if !db.Migrator().HasTable(&entity.Exercise{}) {
		return nil
	}
	if tableHasColumn(db, "exercises", "demo_gif_1") && !tableHasColumn(db, "exercises", "demo_video_1") {
		if err := db.Exec(`ALTER TABLE exercises RENAME COLUMN demo_gif_1 TO demo_video_1`).Error; err != nil {
			return err
		}
		log.Println("migrate: renamed exercises.demo_gif_1 → demo_video_1")
	}
	if tableHasColumn(db, "exercises", "demo_gif_2") && !tableHasColumn(db, "exercises", "demo_video_2") {
		if err := db.Exec(`ALTER TABLE exercises RENAME COLUMN demo_gif_2 TO demo_video_2`).Error; err != nil {
			return err
		}
		log.Println("migrate: renamed exercises.demo_gif_2 → demo_video_2")
	}
	return nil
}
