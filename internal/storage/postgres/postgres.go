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
	return db.AutoMigrate(
		&entity.Role{},
		&entity.User{},
		&entity.Category{},
		&entity.Article{},
		&entity.Equipment{},
		&entity.Muscle{},
	)
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
