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
	if err := seedCatalog(db); err != nil {
		log.Fatalf("seed catalog: %v", err)
	}
	if err := seedDemoBranches(db); err != nil {
		log.Printf("seed branches: %v", err)
	}
}

func seedDemoBranches(db *gorm.DB) error {
	var n int64
	if err := db.Model(&entity.GymBranch{}).Count(&n).Error; err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	demos := []entity.GymBranch{
		{
			Name: "TrongCon Central", Slug: "trongcon-central",
			Address: "101 Vo Thi Sau, Dist. 3", City: "Ho Chi Minh",
			Phone: "0912 345 678", Hours: "Mon–Fri 5:30–21:30 · Sat–Sun 5:30–20:30",
			Description: "Flagship club — full floor + coaching desk.",
			IsActive: true, SortOrder: 1,
		},
		{
			Name: "TrongCon West", Slug: "trongcon-west",
			Address: "112 Tran Phu", City: "Ha Dong",
			Phone: "0912 345 679", Hours: "Mon–Sun 6:00–22:00",
			Description: "West-side club with strength focus.",
			IsActive: true, SortOrder: 2,
		},
		{
			Name: "TrongCon East", Slug: "trongcon-east",
			Address: "360 Giai Phong", City: "Thanh Xuan",
			Phone: "0912 345 680", Hours: "Mon–Fri 5:00–22:00",
			Description: "East club — classes + PT studio.",
			IsActive: true, SortOrder: 3,
		},
	}
	for i := range demos {
		if err := db.Create(&demos[i]).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedRoles(db *gorm.DB) error {
	for _, name := range []string{entity.RoleUser, entity.RoleSuper, entity.RolePT} {
		if err := ensureRole(db, name); err != nil {
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
	log.Printf("seed: created super account %s", seedSuperEmail)
	return nil
}
