package config

import (
	"os"
	"strconv"
	"strings"
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
	SecretKey              string
	WebhookSecret          string
	SuccessURL             string
	CancelURL              string
	MembershipSuccessURL   string
	MembershipCancelURL    string
	PackageSuccessURL      string
	PackageCancelURL       string
}

type VNPayConfig struct {
	TmnCode             string
	SecretKey           string
	PaymentURL          string // sandbox gateway
	ReturnURL           string // FE success page (Premium)
	MembershipReturnURL string // FE success page (gym membership)
	PackageReturnURL    string // FE success page (PT package)
	USDToVND            float64
}

// SMTPConfig is AWS SES SMTP (compatible with strongbody SES_* / SMTP_*).
type SMTPConfig struct {
	Enabled  bool
	Name     string
	Host     string
	Port     string
	Username string
	Password string
	From     string
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
	VNPay         VNPayConfig
	SMTP          SMTPConfig
}

func Load() Config {
	port := getenv("PORT", "8080")
	secret := getenv("JWT_SECRET", "dev-secret-change-in-production")
	expH := getenvInt("JWT_EXPIRE_HOURS", 168)

	cfg := Config{
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
			SecretKey:            getenv("STRIPE_SECRET_KEY", ""),
			WebhookSecret:        getenv("STRIPE_WEBHOOK_SECRET", ""),
			SuccessURL:           getenv("STRIPE_SUCCESS_URL", "http://localhost:3001/premium/success?session_id={CHECKOUT_SESSION_ID}"),
			CancelURL:            getenv("STRIPE_CANCEL_URL", "http://localhost:3001/premium/cancel"),
			MembershipSuccessURL: getenv("STRIPE_MEMBERSHIP_SUCCESS_URL", ""),
			MembershipCancelURL:  getenv("STRIPE_MEMBERSHIP_CANCEL_URL", ""),
			PackageSuccessURL:    getenv("STRIPE_PACKAGE_SUCCESS_URL", ""),
			PackageCancelURL:     getenv("STRIPE_PACKAGE_CANCEL_URL", ""),
		},
		VNPay: loadVNPayConfig(),
		SMTP: SMTPConfig{
			Enabled:  getenvBool("SMTP_ENABLED", false) || getenvBool("SES_ENABLED", false),
			Name:     getenv("SMTP_NAME", getenv("SES_NAME_EMAIL", "TrongCon")),
			Host:     getenv("SMTP_HOST", getenv("SES_SMTP_HOST", "email-smtp.ap-southeast-1.amazonaws.com")),
			Port:     getenv("SMTP_PORT", getenv("SES_SMTP_PORT", "587")),
			Username: getenv("SMTP_USERNAME", getenv("SES_SMTP_USERNAME", "")),
			Password: getenv("SMTP_PASSWORD", getenv("SES_SMTP_PASSWORD", "")),
			From:     getenv("SMTP_FROM", getenv("SES_FROM_EMAIL", "")),
		},
	}
	applyStripeFEDefaults(&cfg)
	return cfg
}

func loadVNPayConfig() VNPayConfig {
	paymentURL := getenv("VNPAY_PAYMENT_URL", "")
	legacyReturn := getenv("VNPAY_RETURN_URL", "")
	// User may have put the sandbox gateway URL in VNPAY_RETURN_URL by mistake.
	if paymentURL == "" && strings.Contains(legacyReturn, "vnpayment.vn") {
		paymentURL = legacyReturn
	}
	if paymentURL == "" {
		paymentURL = "https://sandbox.vnpayment.vn/paymentv2/vpcpay.html"
	}

	fe := strings.TrimRight(getenv("URL_FE_APP", "http://localhost:3001"), "/")
	if fe == "http://localhost:3000" || fe == "https://localhost:3000" {
		fe = "http://localhost:3001"
	}

	success := getenv("VNPAY_RETURN_URL_SUCCESS", "")
	returnURL := success
	if returnURL == "" {
		returnURL = fe + "/premium/success"
	}
	if success == "" && legacyReturn != "" && !strings.Contains(legacyReturn, "vnpayment.vn") {
		returnURL = legacyReturn
	}

	membershipReturnURL := getenv("VNPAY_RETURN_URL_MEMBERSHIP", "")
	if membershipReturnURL == "" {
		membershipReturnURL = fe + "/membership/success"
	}

	packageReturnURL := getenv("VNPAY_RETURN_URL_PT_PACKAGE", "")
	if packageReturnURL == "" {
		packageReturnURL = fe + "/packages/success"
	}

	cfg := VNPayConfig{
		TmnCode:             getenv("VNPAY_TMN_CODE", ""),
		SecretKey:           getenv("VNPAY_SECRET_KEY", ""),
		PaymentURL:          paymentURL,
		ReturnURL:           returnURL,
		MembershipReturnURL: membershipReturnURL,
		PackageReturnURL:    packageReturnURL,
		USDToVND:            float64(getenvInt("VNPAY_USD_VND_RATE", 25000)),
	}
	return cfg
}

// applyStripeFEDefaults fills membership/package Stripe return URLs from FE base when empty.
func applyStripeFEDefaults(cfg *Config) {
	fe := strings.TrimRight(getenv("URL_FE_APP", "http://localhost:3001"), "/")
	if fe == "http://localhost:3000" || fe == "https://localhost:3000" {
		fe = "http://localhost:3001"
	}
	if strings.TrimSpace(cfg.Stripe.MembershipSuccessURL) == "" {
		cfg.Stripe.MembershipSuccessURL = fe + "/membership/success?session_id={CHECKOUT_SESSION_ID}"
	}
	if strings.TrimSpace(cfg.Stripe.MembershipCancelURL) == "" {
		cfg.Stripe.MembershipCancelURL = fe + "/membership"
	}
	if strings.TrimSpace(cfg.Stripe.PackageSuccessURL) == "" {
		cfg.Stripe.PackageSuccessURL = fe + "/packages/success?session_id={CHECKOUT_SESSION_ID}"
	}
	if strings.TrimSpace(cfg.Stripe.PackageCancelURL) == "" {
		cfg.Stripe.PackageCancelURL = fe + "/trainers"
	}
}

func getenv(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

func getenvBool(key string, def bool) bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if v == "" {
		return def
	}
	switch v {
	case "1", "true", "yes", "on", "y":
		return true
	case "0", "false", "no", "off", "n":
		return false
	default:
		return def
	}
}

func getenvInt(key string, def int) int {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return def
}
