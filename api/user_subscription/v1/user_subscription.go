package v1

import "time"

type CheckoutReq struct {
	PlanID uint `json:"plan_id" binding:"required"`
}

type CheckoutRes struct {
	SubscriptionID uint   `json:"subscription_id"`
	OrderID        string `json:"order_id,omitempty"`
	ApproveURL     string `json:"approve_url,omitempty"`
	SessionID      string `json:"session_id,omitempty"`
	CheckoutURL    string `json:"checkout_url,omitempty"`
	Provider       string `json:"provider"`
}

type CaptureReq struct {
	Token   string `json:"token"`    // PayPal return token / order id
	OrderID string `json:"order_id"` // alias for mock/dev
}

type ConfirmStripeReq struct {
	SessionID string `json:"session_id" binding:"required"`
}

type CaptureRes struct {
	Subscription SubscriptionRes `json:"subscription"`
}

type MeRes struct {
	Active *SubscriptionRes  `json:"active"`
	Recent []SubscriptionRes `json:"recent"`
}

type ListAdminReq struct {
	Page     int    `form:"page"`
	Limit    int    `form:"limit"`
	Status   string `form:"status"`
	UserID   uint   `form:"user_id"`
	OrderBy  string `form:"order_by"`
	OrderDir string `form:"order_dir"`
}

type ListAdminRes struct {
	Total int64              `json:"total"`
	Data  []SubscriptionRes  `json:"data"`
}

type SubscriptionRes struct {
	ID                 uint      `json:"id"`
	UserID             uint      `json:"user_id"`
	SubscriptionPlanID uint      `json:"subscription_plan_id"`
	PlanName           string    `json:"plan_name,omitempty"`
	StartDate          time.Time `json:"start_date"`
	EndDate            time.Time `json:"end_date"`
	DurationMonths     int       `json:"duration_months"`
	OriginalPrice      float64   `json:"original_price"`
	FinalPrice         float64   `json:"final_price"`
	Status             string    `json:"status"`
	PaymentProvider    string    `json:"payment_provider"`
	PayPalOrderID      string    `json:"paypal_order_id,omitempty"`
	PayPalCaptureID    string    `json:"paypal_capture_id,omitempty"`
	StripeCheckoutSessionID string `json:"stripe_checkout_session_id,omitempty"`
	Currency           string    `json:"currency"`
	CreatedAt          time.Time `json:"created_at"`
}
