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
	"trongcon-api/internal/money"
	"trongcon-api/internal/repository"

	"github.com/stripe/stripe-go/v82"
	"gorm.io/gorm"
)

var ErrPremiumRequired = errors.New("premium_required")

type UserSubscriptionService interface {
	CheckoutVNPay(ctx context.Context, userID uint, planID uint, clientIP string) (*subv1.CheckoutRes, error)
	ConfirmVNPay(ctx context.Context, userID uint, params map[string]string) (*subv1.CaptureRes, error)
	CheckoutStripe(ctx context.Context, userID uint, planID uint) (*subv1.CheckoutRes, error)
	ConfirmStripe(ctx context.Context, userID uint, sessionID string) (*subv1.CaptureRes, error)
	ActivateFromStripeSession(ctx context.Context, sess *stripe.CheckoutSession) (*entity.UserSubscription, error)
	StartPremiumTrial(ctx context.Context, userID uint) (*subv1.CaptureRes, error)
	Me(ctx context.Context, userID uint) (*subv1.MeRes, error)
	ListAdmin(ctx context.Context, req *subv1.ListAdminReq) (*subv1.ListAdminRes, error)
	IsPremium(ctx context.Context, userID uint) (bool, error)
	SyncUserAccountType(ctx context.Context, userID uint) error
	// GrantPremiumFromMembership unlocks Premium for the membership window (dinhuong.md).
	GrantPremiumFromMembership(ctx context.Context, userID uint, endDate time.Time) error
}

type userSubscriptionService struct {
	subRepo     repository.UserSubscriptionRepository
	planRepo    repository.SubscriptionPlanRepository
	historyRepo repository.PaymentHistoryRepository
	userRepo    repository.UserRepository
	vnpay       VNPayService
	stripe      StripeService
}

func NewUserSubscriptionService(
	subRepo repository.UserSubscriptionRepository,
	planRepo repository.SubscriptionPlanRepository,
	historyRepo repository.PaymentHistoryRepository,
	userRepo repository.UserRepository,
	vnpay VNPayService,
	stripeSvc StripeService,
) UserSubscriptionService {
	return &userSubscriptionService{
		subRepo: subRepo, planRepo: planRepo, historyRepo: historyRepo, userRepo: userRepo, vnpay: vnpay, stripe: stripeSvc,
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

func (s *userSubscriptionService) CheckoutVNPay(ctx context.Context, userID uint, planID uint, clientIP string) (*subv1.CheckoutRes, error) {
	if s.vnpay == nil || !s.vnpay.Enabled() {
		return nil, fmt.Errorf("vnpay is not configured")
	}
	if u, err := s.userRepo.GetByID(ctx, userID); err == nil {
		for _, r := range u.Roles {
			if r.Name == entity.RolePT {
				return nil, fmt.Errorf("trainers do not need Premium")
			}
		}
	}
	plan, err := s.loadPlan(ctx, planID)
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
		PaymentProvider:    entity.PaymentProviderVNPay,
		Currency:           money.Normalize(plan.Currency),
	}
	if err := s.subRepo.Create(ctx, sub); err != nil {
		return nil, err
	}

	txnRef := NewVNPayTxnRef(sub.ID)
	amountVND := s.vnpay.AmountVND(plan.Price, plan.Currency)
	pay, err := s.vnpay.CreatePaymentURL(txnRef, amountVND, "TrongCon Premium - "+plan.PlanName, clientIP, "")
	if err != nil {
		return nil, err
	}
	sub.VnpTxnRef = pay.TxnRef
	if err := s.subRepo.Update(ctx, sub); err != nil {
		return nil, err
	}
	return &subv1.CheckoutRes{
		SubscriptionID: sub.ID,
		OrderID:        pay.TxnRef,
		ApproveURL:     pay.PaymentURL,
		CheckoutURL:    pay.PaymentURL,
		Provider:       entity.PaymentProviderVNPay,
	}, nil
}

func (s *userSubscriptionService) ConfirmVNPay(ctx context.Context, userID uint, params map[string]string) (*subv1.CaptureRes, error) {
	if s.vnpay == nil || !s.vnpay.Enabled() {
		return nil, fmt.Errorf("vnpay is not configured")
	}
	if len(params) == 0 {
		return nil, fmt.Errorf("vnpay params required")
	}

	verified, err := s.vnpay.VerifyReturn(params)
	if err != nil {
		return nil, err
	}
	if !verified.Valid {
		return nil, fmt.Errorf("invalid vnpay signature")
	}

	sub, err := s.subRepo.GetByVnpTxnRef(ctx, verified.TxnRef)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("subscription order not found")
		}
		return nil, err
	}
	if sub.UserID != userID {
		return nil, fmt.Errorf("subscription does not belong to user")
	}
	if sub.Status == entity.SubStatusActive {
		return &subv1.CaptureRes{Subscription: toSubRes(sub)}, nil
	}
	if !verified.Success {
		return nil, fmt.Errorf("vnpay payment not successful: %s", verified.Message)
	}

	sub.VnpTransactionNo = verified.TransactionNo
	if err := s.activateSubscription(ctx, sub, entity.PaymentProviderVNPay, verified.TxnRef, verified.TransactionNo); err != nil {
		return nil, err
	}
	sub, _ = s.subRepo.GetByID(ctx, sub.ID)
	return &subv1.CaptureRes{Subscription: toSubRes(sub)}, nil
}

func (s *userSubscriptionService) CheckoutStripe(ctx context.Context, userID uint, planID uint) (*subv1.CheckoutRes, error) {
	if s.stripe == nil || !s.stripe.Enabled() {
		return nil, fmt.Errorf("stripe is not configured")
	}
	if u, err := s.userRepo.GetByID(ctx, userID); err == nil {
		for _, r := range u.Roles {
			if r.Name == entity.RolePT {
				return nil, fmt.Errorf("trainers do not need Premium")
			}
		}
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
		Currency:           money.Normalize(plan.Currency),
	}
	if err := s.subRepo.Create(ctx, sub); err != nil {
		return nil, err
	}

	sess, err := s.stripe.CreateCheckoutSession(
		plan.PlanName,
		StripeAmountCents(plan.Price),
		money.DefaultCurrency,
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
	if sub.IsTrial {
		if sub.StartDate.IsZero() {
			sub.StartDate = now
		}
		if sub.EndDate.IsZero() || !sub.EndDate.After(sub.StartDate) {
			sub.EndDate = sub.StartDate.AddDate(0, 0, entity.PremiumTrialDays)
		}
	} else {
		months := sub.DurationMonths
		if months < 1 {
			months = 1
		}
		sub.StartDate = now
		sub.EndDate = now.AddDate(0, months, 0)
	}
	sub.Status = entity.SubStatusActive
	if err := s.subRepo.Update(ctx, sub); err != nil {
		return err
	}
	if !sub.IsTrial {
		_ = s.recordPaymentHistory(ctx, sub, method, paymentIntentID, transactionID)
	}
	return s.SyncUserAccountType(ctx, sub.UserID)
}

func (s *userSubscriptionService) StartPremiumTrial(ctx context.Context, userID uint) (*subv1.CaptureRes, error) {
	if u, err := s.userRepo.GetByID(ctx, userID); err == nil {
		for _, r := range u.Roles {
			if r.Name == entity.RolePT {
				return nil, fmt.Errorf("trainers do not need Premium")
			}
		}
	}
	used, err := s.subRepo.HasUsedTrial(ctx, userID)
	if err != nil {
		return nil, err
	}
	if used {
		return nil, fmt.Errorf("premium trial already used")
	}
	now := time.Now().UTC()
	_ = s.subRepo.ExpireEnded(ctx, now)
	active, err := s.subRepo.HasActive(ctx, userID, now)
	if err != nil {
		return nil, err
	}
	if active {
		return nil, fmt.Errorf("you already have an active Premium subscription")
	}

	// Attach trial to the cheapest active premium plan (catalog reference only).
	activeTrue := true
	plans, _, err := s.planRepo.List(ctx, 0, 1, "", entity.PlanKindPremium, "price", "ASC", &activeTrue)
	if err != nil {
		return nil, err
	}
	if len(plans) == 0 {
		return nil, fmt.Errorf("no premium plan available for trial")
	}
	plan := plans[0]

	sub := &entity.UserSubscription{
		UserID:             userID,
		SubscriptionPlanID: plan.ID,
		StartDate:          now,
		EndDate:            now.AddDate(0, 0, entity.PremiumTrialDays),
		DurationMonths:     0,
		OriginalPrice:      plan.Price,
		FinalPrice:         0,
		Status:             entity.SubStatusActive,
		PaymentProvider:    entity.PaymentProviderTrial,
		Currency:           money.Normalize(plan.Currency),
		IsTrial:            true,
	}
	if err := s.subRepo.Create(ctx, sub); err != nil {
		return nil, err
	}
	if err := s.SyncUserAccountType(ctx, userID); err != nil {
		return nil, err
	}
	fresh, err := s.subRepo.GetByID(ctx, sub.ID)
	if err != nil {
		return nil, err
	}
	return &subv1.CaptureRes{Subscription: toSubRes(fresh)}, nil
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
	usedTrial, err := s.subRepo.HasUsedTrial(ctx, userID)
	if err != nil {
		return nil, err
	}
	hasActive := active != nil
	trialAvailable := !usedTrial && !hasActive
	if u, err := s.userRepo.GetByID(ctx, userID); err == nil {
		for _, r := range u.Roles {
			if r.Name == entity.RolePT || r.Name == entity.RoleSuper {
				trialAvailable = false
				break
			}
		}
	}

	out := &subv1.MeRes{
		Recent:         make([]subv1.SubscriptionRes, 0, len(recent)),
		TrialAvailable: trialAvailable,
		TrialDays:      entity.PremiumTrialDays,
	}
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
	_ = s.subRepo.ExpireEnded(ctx, time.Now().UTC())
	page, limit := pageLimit(req.Page, req.Limit)
	rows, total, err := s.subRepo.ListAdmin(ctx, (page-1)*limit, limit, req.Status, req.UserID, req.Q, req.OrderBy, req.OrderDir)
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
	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return false, err
	}
	// Trainers / super never need a paid Premium subscription.
	for _, r := range u.Roles {
		if r.Name == entity.RolePT || r.Name == entity.RoleSuper {
			return true, nil
		}
	}
	_ = s.subRepo.ExpireEnded(ctx, time.Now().UTC())
	ok, err := s.subRepo.HasActive(ctx, userID, time.Now().UTC())
	if err != nil {
		return false, err
	}
	if ok {
		return true, nil
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

// GrantPremiumFromMembership creates/extends complimentary Premium covering the gym pass window.
func (s *userSubscriptionService) GrantPremiumFromMembership(ctx context.Context, userID uint, endDate time.Time) error {
	if userID == 0 || endDate.IsZero() {
		return nil
	}
	if u, err := s.userRepo.GetByID(ctx, userID); err == nil {
		for _, r := range u.Roles {
			if r.Name == entity.RolePT {
				return nil
			}
		}
	}
	now := time.Now().UTC()
	_ = s.subRepo.ExpireEnded(ctx, now)
	if !endDate.After(now) {
		return s.SyncUserAccountType(ctx, userID)
	}
	activeTrue := true
	plans, _, err := s.planRepo.List(ctx, 0, 1, "", entity.PlanKindPremium, "price", "ASC", &activeTrue)
	if err != nil {
		return err
	}
	if len(plans) == 0 {
		return nil
	}
	plan := plans[0]
	if existing, err := s.subRepo.GetActiveByUserID(ctx, userID, now); err == nil && existing != nil {
		if existing.EndDate.Before(endDate) {
			existing.EndDate = endDate
			if err := s.subRepo.Update(ctx, existing); err != nil {
				return err
			}
		}
		return s.SyncUserAccountType(ctx, userID)
	}
	sub := &entity.UserSubscription{
		UserID:             userID,
		SubscriptionPlanID: plan.ID,
		StartDate:          now,
		EndDate:            endDate,
		DurationMonths:     0,
		OriginalPrice:      plan.Price,
		FinalPrice:         0,
		Status:             entity.SubStatusActive,
		PaymentProvider:    entity.PaymentProviderMembership,
		Currency:           money.Normalize(plan.Currency),
		IsTrial:            false,
	}
	if err := s.subRepo.Create(ctx, sub); err != nil {
		return err
	}
	return s.SyncUserAccountType(ctx, userID)
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
	userEmail, userName := "", ""
	if s.User.ID != 0 {
		userEmail = s.User.Email
		userName = strings.TrimSpace(s.User.Name)
		if userName == "" {
			userName = strings.TrimSpace(strings.TrimSpace(s.User.FirstName) + " " + strings.TrimSpace(s.User.LastName))
		}
	}
	daysLeft := 0
	if !s.EndDate.IsZero() {
		daysLeft = int(time.Until(s.EndDate.UTC()).Hours() / 24)
		if s.Status != entity.SubStatusActive {
			if daysLeft > 0 {
				daysLeft = 0
			}
		}
		if daysLeft < 0 {
			daysLeft = 0
		}
	}
	return subv1.SubscriptionRes{
		ID:                      s.ID,
		UserID:                  s.UserID,
		UserEmail:               userEmail,
		UserName:                userName,
		SubscriptionPlanID:      s.SubscriptionPlanID,
		PlanName:                planName,
		StartDate:               s.StartDate,
		EndDate:                 s.EndDate,
		DurationMonths:          s.DurationMonths,
		DaysLeft:                daysLeft,
		OriginalPrice:           s.OriginalPrice,
		FinalPrice:              s.FinalPrice,
		Status:                  s.Status,
		PaymentProvider:         s.PaymentProvider,
		PayPalOrderID:           s.PayPalOrderID,
		PayPalCaptureID:         s.PayPalCaptureID,
		StripeCheckoutSessionID: s.StripeCheckoutSessionID,
		VnpTxnRef:               s.VnpTxnRef,
		VnpTransactionNo:        s.VnpTransactionNo,
		Currency:                s.Currency,
		IsTrial:                 s.IsTrial,
		CreatedAt:               s.CreatedAt,
	}
}
