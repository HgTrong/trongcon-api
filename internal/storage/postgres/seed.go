package postgres

import (
	"log"

	"trongcon-api/internal/entity"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	seedSuperEmail    = "trong520466@gmail.com"
	seedSuperName     = "HgTrong"
	seedSuperPassword = "123456"
)

func seed(db *gorm.DB) {
	if err := seedRoles(db); err != nil {
		log.Fatalf("seed roles: %v", err)
	}
	if err := seedSuperUser(db); err != nil {
		log.Fatalf("seed super user: %v", err)
	}
}

func seedRoles(db *gorm.DB) error {
	for _, name := range []string{entity.RoleUser, entity.RoleSuper} {
		var r entity.Role
		if err := db.Where("name = ?", name).FirstOrCreate(&r, entity.Role{Name: name}).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedSuperUser(db *gorm.DB) error {
	var n int64
	if err := db.Model(&entity.User{}).Where("email = ?", seedSuperEmail).Count(&n).Error; err != nil {
		return err
	}
	if n > 0 {
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(seedSuperPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	var roleSuper entity.Role
	if err := db.Where("name = ?", entity.RoleSuper).First(&roleSuper).Error; err != nil {
		return err
	}

	u := &entity.User{
		Email:        seedSuperEmail,
		Name:         seedSuperName,
		FirstName:    "Hg",
		LastName:     "Trong",
		Language:     "en",
		AccountType:  entity.AccountFree,
		PasswordHash: string(hash),
	}
	if err := db.Create(u).Error; err != nil {
		return err
	}
	if err := db.Model(u).Association("Roles").Append(&roleSuper); err != nil {
		return err
	}
	log.Printf("seed: đã tạo tài khoản super %s", seedSuperEmail)
	return nil
}
