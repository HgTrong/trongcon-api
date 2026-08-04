package postgres

import (
	"log"
	"time"

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
	if err := seedRevenueShare(db); err != nil {
		log.Printf("seed revenue share: %v", err)
	}
	if err := seedGroupClasses(db); err != nil {
		log.Printf("seed group classes: %v", err)
	}
}

// seedRevenueShare ensures the singleton PT/gym revenue split row (id=1) exists.
func seedRevenueShare(db *gorm.DB) error {
	var n int64
	if err := db.Model(&entity.RevenueShareSetting{}).Where("id = ?", 1).Count(&n).Error; err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	s := &entity.RevenueShareSetting{
		PTPercent:  40,
		GymPercent: 60,
		Notes:      "Default PT / gym revenue split",
	}
	s.ID = 1
	return db.Create(s).Error
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
			Latitude: ptrFloat(10.7826), Longitude: ptrFloat(106.6912),
			IsActive: true, SortOrder: 1,
		},
		{
			Name: "TrongCon West", Slug: "trongcon-west",
			Address: "112 Tran Phu", City: "Ha Dong",
			Phone: "0912 345 679", Hours: "Mon–Sun 6:00–22:00",
			Description: "West-side club with strength focus.",
			Latitude: ptrFloat(20.9714), Longitude: ptrFloat(105.7788),
			IsActive: true, SortOrder: 2,
		},
		{
			Name: "TrongCon East", Slug: "trongcon-east",
			Address: "360 Giai Phong", City: "Thanh Xuan",
			Phone: "0912 345 680", Hours: "Mon–Fri 5:00–22:00",
			Description: "East club — classes + PT studio.",
			Latitude: ptrFloat(20.9982), Longitude: ptrFloat(105.8415),
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

// seedGroupClasses creates demo group class types + upcoming sessions when empty.
func seedGroupClasses(db *gorm.DB) error {
	var classCount int64
	if err := db.Model(&entity.GroupClass{}).Count(&classCount).Error; err != nil {
		return err
	}

	var branches []entity.GymBranch
	if err := db.Where("is_active = ?", true).Order("sort_order ASC, id ASC").Limit(3).Find(&branches).Error; err != nil {
		return err
	}
	if len(branches) == 0 {
		log.Printf("seed group classes: skipped — no branches yet")
		return nil
	}

	branchID := func(i int) uint {
		if i < 0 {
			i = 0
		}
		if i >= len(branches) {
			i = len(branches) - 1
		}
		return branches[i].ID
	}

	type classDef struct {
		Name, Category, Description string
		DurationMin, Capacity       int
		BranchIdx                   int
	}
	defs := []classDef{
		{
			Name: "Morning Yoga Flow", Category: "yoga",
			Description: "Gentle vinyasa for mobility and breath — all levels welcome.",
			DurationMin: 60, Capacity: 20, BranchIdx: 0,
		},
		{
			Name: "Power HIIT", Category: "hiit",
			Description: "40 minutes of intervals to spike heart rate. Bring a towel.",
			DurationMin: 45, Capacity: 16, BranchIdx: 0,
		},
		{
			Name: "Zumba Party", Category: "zumba",
			Description: "Dance cardio with Latin beats. No experience needed.",
			DurationMin: 55, Capacity: 25, BranchIdx: 1,
		},
		{
			Name: "Evening Stretch & Mobility", Category: "yoga",
			Description: "Cool-down focused class after work — hips, spine, shoulders.",
			DurationMin: 50, Capacity: 18, BranchIdx: 0,
		},
		{
			Name: "Core & Strength Circuit", Category: "hiit",
			Description: "Bodyweight strength blocks with short rest.",
			DurationMin: 45, Capacity: 14, BranchIdx: 2,
		},
	}

	createdClasses := make([]entity.GroupClass, 0)
	if classCount == 0 {
		for _, d := range defs {
			gc := entity.GroupClass{
				BranchID:    branchID(d.BranchIdx),
				Name:        d.Name,
				Category:    d.Category,
				Description: d.Description,
				DurationMin: d.DurationMin,
				Capacity:    d.Capacity,
				IsActive:    true,
			}
			if err := db.Create(&gc).Error; err != nil {
				return err
			}
			createdClasses = append(createdClasses, gc)
		}
		log.Printf("seed: created %d group classes", len(createdClasses))
	} else {
		if err := db.Where("is_active = ?", true).Order("id ASC").Find(&createdClasses).Error; err != nil {
			return err
		}
	}
	if len(createdClasses) == 0 {
		return nil
	}

	var upcoming int64
	now := time.Now().UTC()
	if err := db.Model(&entity.ClassSession{}).
		Where("is_canceled = ? AND starts_at > ?", false, now).
		Count(&upcoming).Error; err != nil {
		return err
	}
	if upcoming > 0 {
		return nil
	}

	// Schedule sessions over the next ~7 days (Vietnam time mornings/evenings).
	loc := time.FixedZone("ICT", 7*3600)
	base := time.Now().In(loc)
	day0 := time.Date(base.Year(), base.Month(), base.Day()+1, 0, 0, 0, 0, loc)

	type slot struct {
		DayOffset int
		Hour      int
		Minute    int
		ClassIdx  int
	}
	slots := []slot{
		{0, 7, 0, 0},
		{0, 18, 30, 1},
		{1, 9, 0, 2},
		{1, 19, 0, 3},
		{2, 7, 30, 0},
		{2, 18, 0, 4},
		{3, 18, 30, 1},
		{4, 9, 0, 2},
		{4, 19, 0, 3},
		{5, 8, 0, 0},
		{6, 10, 0, 2},
		{6, 17, 30, 4},
	}

	n := 0
	for _, sl := range slots {
		ci := sl.ClassIdx % len(createdClasses)
		gc := createdClasses[ci]
		startLocal := day0.AddDate(0, 0, sl.DayOffset).
			Add(time.Duration(sl.Hour)*time.Hour + time.Duration(sl.Minute)*time.Minute)
		endLocal := startLocal.Add(time.Duration(gc.DurationMin) * time.Minute)
		sess := entity.ClassSession{
			GroupClassID: gc.ID,
			StartsAt:     startLocal.UTC(),
			EndsAt:       endLocal.UTC(),
			Capacity:     gc.Capacity,
			BookedCount:  0,
			IsCanceled:   false,
		}
		if err := db.Create(&sess).Error; err != nil {
			return err
		}
		n++
	}
	log.Printf("seed: created %d upcoming class sessions", n)
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

func ptrFloat(v float64) *float64 {
	return &v
}
