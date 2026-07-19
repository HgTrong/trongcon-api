package service

import (
	"fmt"
	"strings"

	"trongcon-api/internal/config"

	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/checkout/session"
)

type StripeCheckoutResult struct {
	SessionID   string
	CheckoutURL string
}

type StripeService interface {
	CreateCheckoutSession(planName string, amountCents int64, currency string, userID, planID, subID uint, customerEmail string) (*StripeCheckoutResult, error)
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
	if !s.Enabled() {
		return nil, fmt.Errorf("stripe is not configured")
	}
	currency = strings.ToLower(strings.TrimSpace(currency))
	if currency == "" {
		currency = "usd"
	}
	successURL := s.cfg.SuccessURL
	if successURL == "" {
		successURL = "http://localhost:3001/premium/success?session_id={CHECKOUT_SESSION_ID}"
	}
	cancelURL := s.cfg.CancelURL
	if cancelURL == "" {
		cancelURL = "http://localhost:3001/premium/cancel"
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
					Name: stripe.String(planName),
				},
				UnitAmount: stripe.Int64(amountCents),
			},
		}},
	}
	params.AddMetadata("type", "user_subscription")
	params.AddMetadata("user_id", fmt.Sprintf("%d", userID))
	params.AddMetadata("plan_id", fmt.Sprintf("%d", planID))
	params.AddMetadata("subscription_id", fmt.Sprintf("%d", subID))
	if email := strings.TrimSpace(customerEmail); email != "" {
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

func StripeAmountCents(price float64) int64 {
	return int64(price*100 + 0.5)
}
