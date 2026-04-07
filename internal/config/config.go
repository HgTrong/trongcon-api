package config

import (
	"os"
	"strconv"
	"time"
)

type DbConfig struct {
	Host     string
	User     string
	Password string
	DbName   string
	Port     string
	SSLMode  string
	TimeZone string
}

type S3Config struct {
	Region          string
	Bucket          string
	AccessKeyID     string
	SecretAccessKey string
	Prefix          string
	PublicBaseURL   string
}

type Config struct {
	Port          string
	JWTSecret     string
	JWTExpiration time.Duration
	DB            DbConfig
	S3            S3Config
}

func Load() Config {
	port := getenv("PORT", "8080")
	secret := getenv("JWT_SECRET", "dev-secret-change-in-production")
	expH := getenvInt("JWT_EXPIRE_HOURS", 168)

	return Config{
		Port:          port,
		JWTSecret:     secret,
		JWTExpiration: time.Duration(expH) * time.Hour,
		DB: DbConfig{
			Host:     getenv("DB_HOST", "localhost"),
			User:     getenv("DB_USER", "postgres"),
			Password: getenv("DB_PASSWORD", "postgres"),
			DbName:   getenv("DB_NAME", "trongcon"),
			Port:     getenv("DB_PORT", "5432"),
			SSLMode:  getenv("DB_SSLMODE", "disable"),
			TimeZone: getenv("DB_TIMEZONE", "UTC"),
		},
		S3: S3Config{
			Region:          getenv("AWS_REGION", ""),
			Bucket:          getenv("AWS_S3_BUCKET", ""),
			AccessKeyID:     getenv("AWS_ACCESS_KEY_ID", ""),
			SecretAccessKey: getenv("AWS_SECRET_ACCESS_KEY", ""),
			// strongbody-api upload đang lưu dưới prefix: public/images/<basePath>/...
			Prefix:        getenv("AWS_S3_PREFIX", "public/images"),
			PublicBaseURL:   getenv("AWS_S3_PUBLIC_BASE_URL", ""),
		},
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getenvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return def
}
