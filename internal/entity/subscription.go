package entity

import "time"

const (
	PlanKindPremium = "premium"

	SubStatusPending   = "pending"
	SubStatusActive    = "active"
	SubStatusExpired   = "expired"
	SubStatusCanceled  = "canceled"

	PaymentProviderPayPal = "paypal"
	PaymentProviderStripe = "stripe"

	PHTypeUserSubscription = "user_subscription"
	PHStatusSucceeded      = "succeeded"
	PHStatusFailed         = "failed"
	PHStatusPending        = "pending"
)

// SubscriptionPlan is a sellable digital premium package.
type SubscriptionPlan struct {
	BaseEntity
	Code           string  `json:"code" gorm:"type:varchar(64);uniqueIndex"`
	PlanName       string  `json:"plan_name" gorm:"type:varchar(255);not null"`
	Title          string  `json:"title" gorm:"type:text"`
	Description    string  `json:"description" gorm:"type:text"` // JSON array of bullet strings
	Price          float64 `json:"price" gorm:"type:decimal(10,2);not null;default:0"`
	Currency       string  `json:"currency" gorm:"type:varchar(10);not null;default:'USD'"`
	DurationMonths int     `json:"duration_months" gorm:"not null;default:1"`
	IsActive       bool    `json:"is_active" gorm:"not null;default:true"`
	SortOrder      int     `json:"sort_order" gorm:"not null;default:0"`
	Kind           string  `json:"kind" gorm:"type:varchar(32);not null;default:'premium';index"`
}

func (SubscriptionPlan) TableName() string { return "subscription_plans" }

// UserSubscription tracks a user's purchase / entitlement window.
type UserSubscription struct {
	BaseEntity
	UserID             uint       `json:"user_id" gorm:"not null;index"`
	User               User       `json:"-" gorm:"foreignKey:UserID"`
	SubscriptionPlanID uint       `json:"subscription_plan_id" gorm:"not null;index"`
	SubscriptionPlan   SubscriptionPlan `json:"subscription_plan,omitempty" gorm:"foreignKey:SubscriptionPlanID"`
	StartDate          time.Time  `json:"start_date"`
	EndDate            time.Time  `json:"end_date"`
	DurationMonths     int        `json:"duration_months"`
	OriginalPrice      float64    `json:"original_price" gorm:"type:decimal(10,2);not null;default:0"`
	FinalPrice         float64    `json:"final_price" gorm:"type:decimal(10,2);not null;default:0"`
	Status             string     `json:"status" gorm:"type:varchar(20);not null;default:'pending';index"`
	PaymentProvider    string     `json:"payment_provider" gorm:"type:varchar(20);not null;default:'paypal'"`
	PayPalOrderID      string     `json:"paypal_order_id" gorm:"type:varchar(255);index"`
	PayPalCaptureID    string     `json:"paypal_capture_id" gorm:"type:varchar(255);index"`
	StripeCheckoutSessionID string `json:"stripe_checkout_session_id" gorm:"type:varchar(255);index"`
	StripePaymentIntentID   string `json:"stripe_payment_intent_id" gorm:"type:varchar(255);index"`
	Currency           string     `json:"currency" gorm:"type:varchar(10);default:'USD'"`
	CanceledAt         *time.Time `json:"canceled_at"`
}

func (UserSubscription) TableName() string { return "user_subscriptions" }

// PaymentHistory is a flat ledger row for money stats.
type PaymentHistory struct {
	BaseEntity
	UserID            uint    `json:"user_id" gorm:"not null;index"`
	User              User    `json:"-" gorm:"foreignKey:UserID"`
	UserSubscriptionID *uint  `json:"user_subscription_id" gorm:"index"`
	TransactionID     string  `json:"transaction_id" gorm:"type:varchar(255);index"`
	PaymentIntentID   string  `json:"payment_intent_id" gorm:"type:varchar(255);index"`
	Price             float64 `json:"price" gorm:"type:decimal(10,2);not null;default:0"`
	Amount            float64 `json:"amount" gorm:"type:decimal(10,2);not null;default:0"`
	Currency          string  `json:"currency" gorm:"type:varchar(10);not null;default:'USD'"`
	PaymentMethod     string  `json:"payment_method" gorm:"type:varchar(100)"`
	PaymentType       string  `json:"payment_type" gorm:"type:varchar(100);index"`
	Status            string  `json:"status" gorm:"type:varchar(50);index"`
}

func (PaymentHistory) TableName() string { return "payment_histories" }
