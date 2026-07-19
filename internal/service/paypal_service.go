package service

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"trongcon-api/internal/config"

	"github.com/plutov/paypal/v4"
)

type PayPalOrderResult struct {
	OrderID    string
	ApproveURL string
}

type PayPalCaptureResult struct {
	Status    string
	CaptureID string
	OrderID   string
}

type PayPalService interface {
	CreateOrder(ctx context.Context, amount, currency string, userID uint) (*PayPalOrderResult, error)
	CaptureOrder(ctx context.Context, token string, userID uint) (*PayPalCaptureResult, error)
	IsMock() bool
}

type payPalService struct {
	cfg    config.PayPalConfig
	client *paypal.Client
}

func NewPayPalService(cfg config.PayPalConfig) (PayPalService, error) {
	s := &payPalService{cfg: cfg}
	if strings.EqualFold(cfg.TestMode, "mock") {
		log.Printf("[PayPal] MOCK mode enabled")
		return s, nil
	}
	if cfg.ClientID == "" || cfg.ClientSecret == "" {
		log.Printf("[PayPal] credentials missing — falling back to MOCK mode")
		s.cfg.TestMode = "mock"
		return s, nil
	}
	apiBase := cfg.APIBase
	if apiBase == "" {
		apiBase = paypal.APIBaseSandBox
	}
	client, err := paypal.NewClient(cfg.ClientID, cfg.ClientSecret, apiBase)
	if err != nil {
		return nil, fmt.Errorf("paypal client: %w", err)
	}
	if _, err := client.GetAccessToken(context.Background()); err != nil {
		return nil, fmt.Errorf("paypal access token: %w", err)
	}
	s.client = client
	return s, nil
}

func (s *payPalService) IsMock() bool {
	return strings.EqualFold(s.cfg.TestMode, "mock") || s.client == nil
}

func (s *payPalService) CreateOrder(ctx context.Context, amount, currency string, userID uint) (*PayPalOrderResult, error) {
	currency = strings.ToUpper(strings.TrimSpace(currency))
	if currency == "" {
		currency = "USD"
	}
	amount = strings.TrimSpace(amount)
	if amount == "" {
		return nil, fmt.Errorf("amount required")
	}

	if s.IsMock() {
		orderID := fmt.Sprintf("MOCK-ORDER-%d", time.Now().UnixNano())
		returnURL := s.cfg.ReturnURL
		if returnURL == "" {
			returnURL = "http://localhost:3000/premium/success"
		}
		approve := fmt.Sprintf("%s?mock=1&token=%s&order_id=%s", returnURL, orderID, orderID)
		log.Printf("[PayPal MOCK] CreateOrder user=%d amount=%s %s order=%s", userID, amount, currency, orderID)
		return &PayPalOrderResult{OrderID: orderID, ApproveURL: approve}, nil
	}

	units := []paypal.PurchaseUnitRequest{{
		Amount: &paypal.PurchaseUnitAmount{
			Currency: currency,
			Value:    amount,
		},
		Description: "TrongCon Premium",
		CustomID:    fmt.Sprintf("user_%d", userID),
	}}
	order, err := s.client.CreateOrder(ctx, paypal.OrderIntentCapture, units, nil, &paypal.ApplicationContext{
		ReturnURL:  s.cfg.ReturnURL,
		CancelURL:  s.cfg.CancelURL,
		BrandName:  "TrongCon",
		UserAction: paypal.UserActionPayNow,
	})
	if err != nil {
		return nil, fmt.Errorf("create paypal order: %w", err)
	}
	var approve string
	for _, link := range order.Links {
		if link.Rel == "approve" {
			approve = link.Href
			break
		}
	}
	if approve == "" {
		return nil, fmt.Errorf("paypal approve url missing")
	}
	return &PayPalOrderResult{OrderID: order.ID, ApproveURL: approve}, nil
}

func (s *payPalService) CaptureOrder(ctx context.Context, token string, userID uint) (*PayPalCaptureResult, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, fmt.Errorf("token required")
	}

	if s.IsMock() || strings.HasPrefix(token, "MOCK-ORDER-") {
		captureID := fmt.Sprintf("MOCK-CAPTURE-%d", time.Now().UnixNano())
		log.Printf("[PayPal MOCK] CaptureOrder user=%d token=%s capture=%s", userID, token, captureID)
		return &PayPalCaptureResult{Status: "COMPLETED", CaptureID: captureID, OrderID: token}, nil
	}

	capture, err := s.client.CaptureOrder(ctx, token, paypal.CaptureOrderRequest{})
	if err != nil {
		return nil, fmt.Errorf("capture paypal order: %w", err)
	}
	captureID := ""
	if len(capture.PurchaseUnits) > 0 && capture.PurchaseUnits[0].Payments != nil &&
		len(capture.PurchaseUnits[0].Payments.Captures) > 0 {
		captureID = capture.PurchaseUnits[0].Payments.Captures[0].ID
	}
	return &PayPalCaptureResult{
		Status:    string(capture.Status),
		CaptureID: captureID,
		OrderID:   capture.ID,
	}, nil
}

func FormatPayPalAmount(price float64) string {
	return strconv.FormatFloat(price, 'f', 2, 64)
}
