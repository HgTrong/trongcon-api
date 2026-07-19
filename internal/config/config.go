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

type OpenAIConfig struct {
	APIKey      string
	AssistantID string
	BaseURL     string
	Timeout     int
	Model       string
}

type PayPalConfig struct {
	ClientID     string
	ClientSecret string
	APIBase      string
	ReturnURL    string
	CancelURL    string
	WebhookID    string
	TestMode     string // "mock" skips real PayPal calls
}

type StripeConfig struct {
	SecretKey     string
	WebhookSecret string
	SuccessURL    string
	CancelURL     string
}

type Config struct {
	Port          string
	JWTSecret     string
	JWTExpiration time.Duration
	DB            DbConfig
	S3            S3Config
	OpenAI        OpenAIConfig
	PayPal        PayPalConfig
	Stripe        StripeConfig
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
			Prefix:          getenv("AWS_S3_PREFIX", "public/images"),
			PublicBaseURL:   getenv("AWS_S3_PUBLIC_BASE_URL", ""),
		},
		OpenAI: OpenAIConfig{
			APIKey:      getenv("OPENAI_API_KEY", ""),
			AssistantID: getenv("OPENAI_ASSISTANT_ID", ""),
			BaseURL:     getenv("OPENAI_BASE_URL", "https://api.openai.com/v1"),
			Timeout:     getenvInt("OPENAI_TIMEOUT", 60),
			Model:       getenv("OPENAI_MODEL", "gpt-4o-mini"),
		},
		PayPal: PayPalConfig{
			ClientID:     getenv("PAYPAL_CLIENT_ID", ""),
			ClientSecret: getenv("PAYPAL_CLIENT_SECRET", ""),
			APIBase:      getenv("PAYPAL_API_BASE", "https://api-m.sandbox.paypal.com"),
			ReturnURL:    getenv("PAYPAL_RETURN_URL", "http://localhost:3001/premium/success"),
			CancelURL:    getenv("PAYPAL_CANCEL_URL", "http://localhost:3001/premium/cancel"),
			WebhookID:    getenv("PAYPAL_WEBHOOK_ID", ""),
			TestMode:     getenv("PAYPAL_TEST_MODE", ""),
		},
		Stripe: StripeConfig{
			SecretKey:     getenv("STRIPE_SECRET_KEY", ""),
			WebhookSecret: getenv("STRIPE_WEBHOOK_SECRET", ""),
			SuccessURL:    getenv("STRIPE_SUCCESS_URL", "http://localhost:3001/premium/success?session_id={CHECKOUT_SESSION_ID}"),
			CancelURL:     getenv("STRIPE_CANCEL_URL", "http://localhost:3001/premium/cancel"),
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
