package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	subv1 "trongcon-api/api/user_subscription/v1"
	"trongcon-api/internal/entity"
	"trongcon-api/internal/repository"

	"github.com/stripe/stripe-go/v82"
	"gorm.io/gorm"
)

var ErrPremiumRequired = errors.New("premium_required")

type UserSubscriptionService interface {
	CheckoutPayPal(ctx context.Context, userID uint, planID uint) (*subv1.CheckoutRes, error)
	CapturePayPal(ctx context.Context, userID uint, token, orderID string) (*subv1.CaptureRes, error)
	CheckoutStripe(ctx context.Context, userID uint, planID uint) (*subv1.CheckoutRes, error)
	ConfirmStripe(ctx context.Context, userID uint, sessionID string) (*subv1.CaptureRes, error)
	ActivateFromStripeSession(ctx context.Context, sess *stripe.CheckoutSession) (*entity.UserSubscription, error)
	Me(ctx context.Context, userID uint) (*subv1.MeRes, error)
	ListAdmin(ctx context.Context, req *subv1.ListAdminReq) (*subv1.ListAdminRes, error)
	IsPremium(ctx context.Context, userID uint) (bool, error)
	SyncUserAccountType(ctx context.Context, userID uint) error
}

type userSubscriptionService struct {
	subRepo     repository.UserSubscriptionRepository
	planRepo    repository.SubscriptionPlanRepository
	historyRepo repository.PaymentHistoryRepository
	userRepo    repository.UserRepository
	paypal      PayPalService
	stripe      StripeService
}

func NewUserSubscriptionService(
	subRepo repository.UserSubscriptionRepository,
	planRepo repository.SubscriptionPlanRepository,
	historyRepo repository.PaymentHistoryRepository,
	userRepo repository.UserRepository,
	paypal PayPalService,
	stripeSvc StripeService,
) UserSubscriptionService {
	return &userSubscriptionService{
		subRepo: subRepo, planRepo: planRepo, historyRepo: historyRepo, userRepo: userRepo, paypal: paypal, stripe: stripeSvc,
	}
}

func (s *userSubscriptionService) loadPlan(ctx context.Context, planID uint) (*entity.SubscriptionPlan, error) {
	plan, err := s.planRepo.GetByID(ctx, planID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("plan not found")
		}
		return nil, err
	}
	if !plan.IsActive || plan.Kind != entity.PlanKindPremium {
		return nil, fmt.Errorf("plan is not available")
	}
	return plan, nil
}

func (s *userSubscriptionService) CheckoutPayPal(ctx context.Context, userID uint, planID uint) (*subv1.CheckoutRes, error) {
	plan, err := s.loadPlan(ctx, planID)
	if err != nil {
		return nil, err
	}

	order, err := s.paypal.CreateOrder(ctx, FormatPayPalAmount(plan.Price), plan.Currency, userID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	sub := &entity.UserSubscription{
		UserID:             userID,
		SubscriptionPlanID: plan.ID,
		StartDate:          now,
		EndDate:            now,
		DurationMonths:     plan.DurationMonths,
		OriginalPrice:      plan.Price,
		FinalPrice:         plan.Price,
		Status:             entity.SubStatusPending,
		PaymentProvider:    entity.PaymentProviderPayPal,
		PayPalOrderID:      order.OrderID,
		Currency:           plan.Currency,
	}
	if err := s.subRepo.Create(ctx, sub); err != nil {
		return nil, err
	}
	return &subv1.CheckoutRes{
		SubscriptionID: sub.ID,
		OrderID:        order.OrderID,
		ApproveURL:     order.ApproveURL,
		Provider:       entity.PaymentProviderPayPal,
	}, nil
}

func (s *userSubscriptionService) CapturePayPal(ctx context.Context, userID uint, token, orderID string) (*subv1.CaptureRes, error) {
	ref := strings.TrimSpace(token)
	if ref == "" {
		ref = strings.TrimSpace(orderID)
	}
	if ref == "" {
		return nil, fmt.Errorf("token or order_id required")
	}

	sub, err := s.subRepo.GetByPayPalOrderID(ctx, ref)
	if err != nil {
		if orderID != "" && orderID != ref {
			sub, err = s.subRepo.GetByPayPalOrderID(ctx, orderID)
		}
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("subscription order not found")
			}
			return nil, err
		}
	}
	if sub.UserID != userID {
		return nil, fmt.Errorf("subscription does not belong to user")
	}
	if sub.Status == entity.SubStatusActive {
		return &subv1.CaptureRes{Subscription: toSubRes(sub)}, nil
	}

	capRes, err := s.paypal.CaptureOrder(ctx, ref, userID)
	if err != nil {
		return nil, err
	}
	if capRes.CaptureID != "" {
		sub.PayPalCaptureID = capRes.CaptureID
	}
	if capRes.OrderID != "" && sub.PayPalOrderID == "" {
		sub.PayPalOrderID = capRes.OrderID
	}
	if err := s.activateSubscription(ctx, sub, entity.PaymentProviderPayPal, sub.PayPalOrderID, sub.PayPalCaptureID); err != nil {
		return nil, err
	}
	sub, _ = s.subRepo.GetByID(ctx, sub.ID)
	return &subv1.CaptureRes{Subscription: toSubRes(sub)}, nil
}

func (s *userSubscriptionService) CheckoutStripe(ctx context.Context, userID uint, planID uint) (*subv1.CheckoutRes, error) {
	if s.stripe == nil || !s.stripe.Enabled() {
		return nil, fmt.Errorf("stripe is not configured")
	}
	plan, err := s.loadPlan(ctx, planID)
	if err != nil {
		return nil, err
	}
	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	sub := &entity.UserSubscription{
		UserID:             userID,
		SubscriptionPlanID: plan.ID,
		StartDate:          now,
		EndDate:            now,
		DurationMonths:     plan.DurationMonths,
		OriginalPrice:      plan.Price,
		FinalPrice:         plan.Price,
		Status:             entity.SubStatusPending,
		PaymentProvider:    entity.PaymentProviderStripe,
		Currency:           plan.Currency,
	}
	if err := s.subRepo.Create(ctx, sub); err != nil {
		return nil, err
	}

	sess, err := s.stripe.CreateCheckoutSession(
		plan.PlanName,
		StripeAmountCents(plan.Price),
		plan.Currency,
		userID,
		plan.ID,
		sub.ID,
		u.Email,
	)
	if err != nil {
		return nil, err
	}
	sub.StripeCheckoutSessionID = sess.SessionID
	if err := s.subRepo.Update(ctx, sub); err != nil {
		return nil, err
	}
	return &subv1.CheckoutRes{
		SubscriptionID: sub.ID,
		SessionID:      sess.SessionID,
		CheckoutURL:    sess.CheckoutURL,
		Provider:       entity.PaymentProviderStripe,
	}, nil
}

func (s *userSubscriptionService) ConfirmStripe(ctx context.Context, userID uint, sessionID string) (*subv1.CaptureRes, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return nil, fmt.Errorf("session_id required")
	}
	if s.stripe == nil || !s.stripe.Enabled() {
		return nil, fmt.Errorf("stripe is not configured")
	}
	sess, err := s.stripe.GetCheckoutSession(sessionID)
	if err != nil {
		return nil, err
	}
	sub, err := s.ActivateFromStripeSession(ctx, sess)
	if err != nil {
		return nil, err
	}
	if sub.UserID != userID {
		return nil, fmt.Errorf("subscription does not belong to user")
	}
	return &subv1.CaptureRes{Subscription: toSubRes(sub)}, nil
}

func (s *userSubscriptionService) ActivateFromStripeSession(ctx context.Context, sess *stripe.CheckoutSession) (*entity.UserSubscription, error) {
	if sess == nil {
		return nil, fmt.Errorf("nil stripe session")
	}
	if sess.PaymentStatus != stripe.CheckoutSessionPaymentStatusPaid &&
		sess.Status != stripe.CheckoutSessionStatusComplete {
		return nil, fmt.Errorf("stripe session not paid (status=%s payment=%s)", sess.Status, sess.PaymentStatus)
	}

	sub, err := s.subRepo.GetByStripeCheckoutSessionID(ctx, sess.ID)
	if err != nil {
		if metaID := strings.TrimSpace(sess.Metadata["subscription_id"]); metaID != "" {
			if id64, e := strconv.ParseUint(metaID, 10, 64); e == nil {
				sub, err = s.subRepo.GetByID(ctx, uint(id64))
			}
		}
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("subscription for stripe session not found")
			}
			return nil, err
		}
		if sub.StripeCheckoutSessionID == "" {
			sub.StripeCheckoutSessionID = sess.ID
		}
	}

	if sub.Status == entity.SubStatusActive {
		return sub, nil
	}

	pi := ""
	if sess.PaymentIntent != nil {
		pi = sess.PaymentIntent.ID
	}
	sub.PaymentProvider = entity.PaymentProviderStripe
	sub.StripeCheckoutSessionID = sess.ID
	sub.StripePaymentIntentID = pi
	if err := s.activateSubscription(ctx, sub, entity.PaymentProviderStripe, pi, sess.ID); err != nil {
		return nil, err
	}
	return s.subRepo.GetByID(ctx, sub.ID)
}

func (s *userSubscriptionService) activateSubscription(ctx context.Context, sub *entity.UserSubscription, method, paymentIntentID, transactionID string) error {
	now := time.Now().UTC()
	months := sub.DurationMonths
	if months < 1 {
		months = 1
	}
	sub.StartDate = now
	sub.EndDate = now.AddDate(0, months, 0)
	sub.Status = entity.SubStatusActive
	if err := s.subRepo.Update(ctx, sub); err != nil {
		return err
	}
	_ = s.recordPaymentHistory(ctx, sub, method, paymentIntentID, transactionID)
	return s.SyncUserAccountType(ctx, sub.UserID)
}

func (s *userSubscriptionService) Me(ctx context.Context, userID uint) (*subv1.MeRes, error) {
	_ = s.subRepo.ExpireEnded(ctx, time.Now().UTC())
	_ = s.SyncUserAccountType(ctx, userID)

	now := time.Now().UTC()
	active, err := s.subRepo.GetActiveByUserID(ctx, userID, now)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	recent, err := s.subRepo.ListByUserID(ctx, userID, 10)
	if err != nil {
		return nil, err
	}
	out := &subv1.MeRes{Recent: make([]subv1.SubscriptionRes, 0, len(recent))}
	if active != nil {
		r := toSubRes(active)
		out.Active = &r
	}
	for i := range recent {
		out.Recent = append(out.Recent, toSubRes(&recent[i]))
	}
	return out, nil
}

func (s *userSubscriptionService) ListAdmin(ctx context.Context, req *subv1.ListAdminReq) (*subv1.ListAdminRes, error) {
	page, limit := pageLimit(req.Page, req.Limit)
	rows, total, err := s.subRepo.ListAdmin(ctx, (page-1)*limit, limit, req.Status, req.UserID, req.OrderBy, req.OrderDir)
	if err != nil {
		return nil, err
	}
	out := make([]subv1.SubscriptionRes, 0, len(rows))
	for i := range rows {
		out = append(out, toSubRes(&rows[i]))
	}
	return &subv1.ListAdminRes{Total: total, Data: out}, nil
}

func (s *userSubscriptionService) IsPremium(ctx context.Context, userID uint) (bool, error) {
	_ = s.subRepo.ExpireEnded(ctx, time.Now().UTC())
	ok, err := s.subRepo.HasActive(ctx, userID, time.Now().UTC())
	if err != nil {
		return false, err
	}
	if ok {
		return true, nil
	}
	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return false, err
	}
	return u.AccountType == entity.AccountPremium, nil
}

func (s *userSubscriptionService) SyncUserAccountType(ctx context.Context, userID uint) error {
	now := time.Now().UTC()
	_ = s.subRepo.ExpireEnded(ctx, now)
	ok, err := s.subRepo.HasActive(ctx, userID, now)
	if err != nil {
		return err
	}
	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	want := entity.AccountFree
	if ok {
		want = entity.AccountPremium
	}
	if u.AccountType != want {
		u.AccountType = want
		return s.userRepo.Update(ctx, u)
	}
	return nil
}

func (s *userSubscriptionService) recordPaymentHistory(ctx context.Context, sub *entity.UserSubscription, method, paymentIntentID, transactionID string) error {
	if s.historyRepo == nil || sub == nil {
		return nil
	}
	if paymentIntentID == "" {
		paymentIntentID = transactionID
	}
	if transactionID == "" {
		transactionID = paymentIntentID
	}
	if paymentIntentID != "" {
		if existing, err := s.historyRepo.GetByPaymentIntentID(ctx, paymentIntentID); err == nil && existing != nil {
			return nil
		} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
	}
	sid := sub.ID
	h := &entity.PaymentHistory{
		UserID:             sub.UserID,
		UserSubscriptionID: &sid,
		TransactionID:      transactionID,
		PaymentIntentID:    paymentIntentID,
		Price:              sub.OriginalPrice,
		Amount:             sub.FinalPrice,
		Currency:           sub.Currency,
		PaymentMethod:      method,
		PaymentType:        entity.PHTypeUserSubscription,
		Status:             entity.PHStatusSucceeded,
	}
	return s.historyRepo.Create(ctx, h)
}

func toSubRes(s *entity.UserSubscription) subv1.SubscriptionRes {
	if s == nil {
		return subv1.SubscriptionRes{}
	}
	planName := ""
	if s.SubscriptionPlan.ID != 0 {
		planName = s.SubscriptionPlan.PlanName
	}
	return subv1.SubscriptionRes{
		ID:                      s.ID,
		UserID:                  s.UserID,
		SubscriptionPlanID:      s.SubscriptionPlanID,
		PlanName:                planName,
		StartDate:               s.StartDate,
		EndDate:                 s.EndDate,
		DurationMonths:          s.DurationMonths,
		OriginalPrice:           s.OriginalPrice,
		FinalPrice:              s.FinalPrice,
		Status:                  s.Status,
		PaymentProvider:         s.PaymentProvider,
		PayPalOrderID:           s.PayPalOrderID,
		PayPalCaptureID:         s.PayPalCaptureID,
		StripeCheckoutSessionID: s.StripeCheckoutSessionID,
		Currency:                s.Currency,
		CreatedAt:               s.CreatedAt,
	}
}
