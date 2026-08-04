package postgres

import (
	"log"
	"sync"
	"time"

	"trongcon-api/internal/config"
	"trongcon-api/internal/entity"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database struct {
	Connection *gorm.DB
}

var dbInstance *Database
var dbOnce sync.Once

func GetDatabaseConnection() *Database {
	dbOnce.Do(func() {
		dbInstance = &Database{
			Connection: dbConnect(),
		}
	})
	return dbInstance
}

func dbConnect() *gorm.DB {
	dsn := getDSN()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to get sql.DB: %v", err)
	}
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)

	if err := autoMigrate(db); err != nil {
		log.Fatalf("auto migrate: %v", err)
	}
	seed(db)

	return db
}

func autoMigrate(db *gorm.DB) error {
	if err := patchLegacySchema(db); err != nil {
		return err
	}
	if err := db.AutoMigrate(
		&entity.Role{},
		&entity.User{},
		&entity.Category{},
		&entity.Article{},
		&entity.Equipment{},
		&entity.Exercise{},
		&entity.ExerciseStep{},
		&entity.ExerciseMuscle{},
		&entity.Food{},
		&entity.Muscle{},
		&entity.MealPlan{},
		&entity.MealPlanMeal{},
		&entity.MealPlanItem{},
		&entity.Routine{},
		&entity.RoutineWorkout{},
		&entity.Workout{},
		&entity.WorkoutItem{},
		&entity.UserSavedWorkout{},
		&entity.UserNutritionGoal{},
		&entity.FoodLogMeal{},
		&entity.FoodLogEntry{},
		&entity.WorkoutSession{},
		&entity.WorkoutSessionItem{},
		&entity.WorkoutSetLog{},
		&entity.TrainingEnrollment{},
		&entity.EnrollmentSlot{},
		&entity.AiChatThread{},
		&entity.AiChatMessage{},
		&entity.GymBranch{},
		&entity.TrainerProfile{},
		&entity.FAQ{},
		&entity.SubscriptionPlan{},
		&entity.UserSubscription{},
		&entity.PaymentHistory{},
		&entity.GymMembershipPlan{},
		&entity.UserGymMembership{},
		&entity.GroupClass{},
		&entity.ClassSession{},
		&entity.ClassBooking{},
		&entity.PTPackage{},
		&entity.UserPTPackage{},
		&entity.RevenueShareSetting{},
		&entity.PTEarning{},
		&entity.PTSessionLog{},
		&entity.PTSessionOffer{},
		&entity.PTWorkingHours{},
		&entity.PTBlockedSlot{},
		&entity.PTPackageChatMessage{},
		&entity.PTPackageChatRead{},
		&entity.PTContentStat{},
		&entity.PTAttribution{},
		&entity.PTReview{},
		&entity.GymCheckIn{},
		&entity.EmailTemplate{},
		&entity.EmailOTP{},
	); err != nil {
		return err
	}
	if err := seedTransactionalEmailTemplates(db); err != nil {
		return err
	}
	if err := migrateFoodLogMeals(db); err != nil {
		return err
	}
	if err := migrateMealPlanMeals(db); err != nil {
		return err
	}
	if err := backfillWorkoutUserID(db); err != nil {
		return err
	}
	return patchFoodEggServings(db)
}

// backfillWorkoutUserID copies owner_user_id → user_id for personal workouts
// created before the explicit poster field existed.
func backfillWorkoutUserID(db *gorm.DB) error {
	if !db.Migrator().HasTable(&entity.Workout{}) || !db.Migrator().HasColumn(&entity.Workout{}, "user_id") {
		return nil
	}
	res := db.Exec(`
		UPDATE workouts
		SET user_id = owner_user_id
		WHERE (user_id IS NULL OR user_id = 0)
		  AND owner_user_id IS NOT NULL
		  AND owner_user_id > 0
	`)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected > 0 {
		log.Printf("migrate: backfilled user_id on %d workout(s) from owner_user_id", res.RowsAffected)
	}
	return nil
}

func getDSN() string {
	cfg := config.Load().DB
	return "host=" + cfg.Host +
		" user=" + cfg.User +
		" password=" + cfg.Password +
		" dbname=" + cfg.DbName +
		" port=" + cfg.Port +
		" sslmode=" + cfg.SSLMode +
		" TimeZone=" + cfg.TimeZone
}

func (d *Database) Close() error {
	sqlDB, err := d.Connection.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
