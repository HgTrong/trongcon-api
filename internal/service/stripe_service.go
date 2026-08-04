package service

import (
	"fmt"
	"math"
	"strings"

	"trongcon-api/internal/config"

	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/checkout/session"
)

type StripeCheckoutResult struct {
	SessionID   string
	CheckoutURL string
}

type StripeCheckoutOpts struct {
	PlanName      string
	AmountCents   int64 // Stripe unit_amount (cents for USD; whole VND for vnd)
	Currency      string
	UserID        uint
	PlanID        uint
	RecordID      uint
	CustomerEmail string
	SuccessURL    string
	CancelURL     string
	MetaType      string // user_subscription | gym_membership | pt_package
	RecordMetaKey string // subscription_id | membership_id | user_pt_package_id
}

type StripeService interface {
	CreateCheckoutSession(planName string, amountCents int64, currency string, userID, planID, subID uint, customerEmail string) (*StripeCheckoutResult, error)
	CreateCheckout(opts StripeCheckoutOpts) (*StripeCheckoutResult, error)
	GetCheckoutSession(sessionID string) (*stripe.CheckoutSession, error)
	Enabled() bool
}

type stripeService struct {
	cfg config.StripeConfig
}

func NewStripeService(cfg config.StripeConfig) StripeService {
	if strings.TrimSpace(cfg.SecretKey) != "" {
		stripe.Key = cfg.SecretKey
	}
	return &stripeService{cfg: cfg}
}

func (s *stripeService) Enabled() bool {
	return strings.TrimSpace(s.cfg.SecretKey) != ""
}

func (s *stripeService) CreateCheckoutSession(planName string, amountCents int64, currency string, userID, planID, subID uint, customerEmail string) (*StripeCheckoutResult, error) {
	return s.CreateCheckout(StripeCheckoutOpts{
		PlanName:      planName,
		AmountCents:   amountCents,
		Currency:      currency,
		UserID:        userID,
		PlanID:        planID,
		RecordID:      subID,
		CustomerEmail: customerEmail,
		MetaType:      "user_subscription",
		RecordMetaKey: "subscription_id",
	})
}

func (s *stripeService) CreateCheckout(opts StripeCheckoutOpts) (*StripeCheckoutResult, error) {
	if !s.Enabled() {
		return nil, fmt.Errorf("stripe is not configured")
	}
	currency := normalizeStripeCurrency(opts.Currency)
	successURL := strings.TrimSpace(opts.SuccessURL)
	if successURL == "" {
		successURL = s.cfg.SuccessURL
	}
	if successURL == "" {
		successURL = "http://localhost:3001/premium/success?session_id={CHECKOUT_SESSION_ID}"
	}
	cancelURL := strings.TrimSpace(opts.CancelURL)
	if cancelURL == "" {
		cancelURL = s.cfg.CancelURL
	}
	if cancelURL == "" {
		cancelURL = "http://localhost:3001/premium/cancel"
	}
	metaType := strings.TrimSpace(opts.MetaType)
	if metaType == "" {
		metaType = "user_subscription"
	}
	recordKey := strings.TrimSpace(opts.RecordMetaKey)
	if recordKey == "" {
		recordKey = "subscription_id"
	}

	params := &stripe.CheckoutSessionParams{
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
		LineItems: []*stripe.CheckoutSessionLineItemParams{{
			Quantity: stripe.Int64(1),
			PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
				Currency: stripe.String(currency),
				ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
					Name: stripe.String(opts.PlanName),
				},
				UnitAmount: stripe.Int64(opts.AmountCents),
			},
		}},
	}
	params.AddMetadata("type", metaType)
	params.AddMetadata("user_id", fmt.Sprintf("%d", opts.UserID))
	params.AddMetadata("plan_id", fmt.Sprintf("%d", opts.PlanID))
	params.AddMetadata(recordKey, fmt.Sprintf("%d", opts.RecordID))
	if email := strings.TrimSpace(opts.CustomerEmail); email != "" {
		params.CustomerEmail = stripe.String(email)
	}

	sess, err := session.New(params)
	if err != nil {
		return nil, fmt.Errorf("stripe checkout session: %w", err)
	}
	return &StripeCheckoutResult{SessionID: sess.ID, CheckoutURL: sess.URL}, nil
}

func (s *stripeService) GetCheckoutSession(sessionID string) (*stripe.CheckoutSession, error) {
	if !s.Enabled() {
		return nil, fmt.Errorf("stripe is not configured")
	}
	sess, err := session.Get(sessionID, nil)
	if err != nil {
		return nil, err
	}
	return sess, nil
}

func normalizeStripeCurrency(currency string) string {
	cur := strings.ToLower(strings.TrimSpace(currency))
	if cur == "" || cur == "usd" {
		return "vnd"
	}
	return cur
}

// StripeUnitAmount maps a catalog price to Stripe unit_amount.
// VND (and other zero-decimal currencies) use whole units; others use cents.
func StripeUnitAmount(price float64, currency string) int64 {
	cur := normalizeStripeCurrency(currency)
	switch cur {
	case "vnd", "jpy", "krw":
		return int64(math.Round(price))
	default:
		return int64(price*100 + 0.5)
	}
}

// StripeAmountCents is kept for callers; always treats price as VND whole units.
func StripeAmountCents(price float64) int64 {
	return StripeUnitAmount(price, "vnd")
}
