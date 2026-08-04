package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	gcv1 "trongcon-api/api/gym_commerce/v1"
	"trongcon-api/internal/entity"
	"trongcon-api/internal/money"
	"trongcon-api/internal/repository"

	"github.com/stripe/stripe-go/v82"
	"gorm.io/gorm"
)

// GymCommerceService covers membership plans, group classes, PT packages,
// revenue share and PT earnings — admin, user and public surfaces.
type GymCommerceService interface {
	// Admin — membership plans
	CreatePlan(ctx context.Context, req *gcv1.MembershipPlanReq) (*gcv1.MembershipPlanRes, error)
	UpdatePlan(ctx context.Context, id uint, req *gcv1.MembershipPlanReq) (*gcv1.MembershipPlanRes, error)
	DeletePlan(ctx context.Context, id uint) error
	ListPlansAdmin(ctx context.Context, page, limit int, q string, active *bool) (*gcv1.ListRes, error)
	GetPlan(ctx context.Context, id uint) (*gcv1.MembershipPlanRes, error)
	SetPlanHighlight(ctx context.Context, id uint, highlighted bool) (*gcv1.MembershipPlanRes, error)

	// Admin — user memberships
	ListUserGymMembershipsAdmin(ctx context.Context, page, limit int, status string, userID uint) (*gcv1.ListRes, error)
	AdminActivateMembership(ctx context.Context, membershipID uint) (*gcv1.GymMembershipRes, error)
	AdminCancelMembership(ctx context.Context, membershipID uint) (*gcv1.GymMembershipRes, error)
	AdminSetPTEarningPaidOut(ctx context.Context, earningID uint, paid bool) error

	// Admin — group classes
	CreateGroupClass(ctx context.Context, req *gcv1.GroupClassReq) (*gcv1.GroupClassRes, error)
	UpdateGroupClass(ctx context.Context, id uint, req *gcv1.GroupClassReq) (*gcv1.GroupClassRes, error)
	DeleteGroupClass(ctx context.Context, id uint) error
	ListGroupClasses(ctx context.Context, page, limit int, q string, branchID *uint) (*gcv1.ListRes, error)

	// Admin — class sessions
	CreateClassSession(ctx context.Context, req *gcv1.ClassSessionReq) (*gcv1.ClassSessionRes, error)
	DeleteClassSession(ctx context.Context, id uint) error
	ListClassSessionsAdmin(ctx context.Context, page, limit int, groupClassID *uint) (*gcv1.ListRes, error)

	// Admin — revenue share
	GetRevenueShare(ctx context.Context) (*gcv1.RevenueShareRes, error)
	UpdateRevenueShare(ctx context.Context, req *gcv1.RevenueShareReq) (*gcv1.RevenueShareRes, error)

	// Admin — PT earnings / packages catalog / sold packages
	ListPTEarningsAdmin(ctx context.Context, page, limit int, trainerProfileID uint) (*gcv1.EarningsSummaryRes, error)
	ListPTPackagesAdmin(ctx context.Context, page, limit int, q string, trainerProfileID uint) (*gcv1.ListRes, error)
	ListUserPTPackagesAdmin(ctx context.Context, page, limit int, status string, trainerProfileID, userID uint) (*gcv1.ListRes, error)
	ListPTSessionsAdmin(ctx context.Context, userPTPackageID uint) (*gcv1.ListRes, error)

	// User — gym membership checkout
	CheckoutMembershipVNPay(ctx context.Context, userID, planID uint, clientIP string) (*gcv1.CheckoutRes, error)
	ConfirmMembershipVNPay(ctx context.Context, userID uint, params map[string]string) (*gcv1.GymMembershipRes, error)
	CheckoutMembershipStripe(ctx context.Context, userID, planID uint) (*gcv1.CheckoutRes, error)
	ConfirmMembershipStripe(ctx context.Context, userID uint, sessionID string) (*gcv1.GymMembershipRes, error)
	MyMembership(ctx context.Context, userID uint) (*gcv1.MembershipMeRes, error)
	IssueCheckInToken(ctx context.Context, userID uint) (*gcv1.CheckInTokenRes, error)
	VerifyCheckIn(ctx context.Context, staffUserID uint, token string, branchID *uint, note string) (*gcv1.GymCheckInRes, error)
	ListRecentCheckIns(ctx context.Context, limit int) (*gcv1.ListRes, error)

	// User — group classes
	ListUpcomingClassSessions(ctx context.Context, page, limit int) (*gcv1.ListRes, error)
	BookClassSession(ctx context.Context, userID, sessionID uint) (*gcv1.ClassBookingRes, error)
	CancelClassBooking(ctx context.Context, userID, bookingID uint) error
	MyClassBookings(ctx context.Context, userID uint) (*gcv1.ListRes, error)

	// User — PT packages (as a trainer selling packages)
	CreatePTPackage(ctx context.Context, trainerUserID uint, req *gcv1.PTPackageReq) (*gcv1.PTPackageRes, error)
	UpdatePTPackage(ctx context.Context, trainerUserID, id uint, req *gcv1.PTPackageReq) (*gcv1.PTPackageRes, error)
	DeletePTPackage(ctx context.Context, trainerUserID, id uint) error
	ListMyPTPackages(ctx context.Context, trainerUserID uint, page, limit int) (*gcv1.ListRes, error)
	ListSoldPTPackages(ctx context.Context, trainerUserID uint, page, limit int, status string) (*gcv1.ListRes, error)

	// User — PT packages (as a buyer)
	ListPurchasedPTPackages(ctx context.Context, userID uint) (*gcv1.ListRes, error)
	GetUserPTPackage(ctx context.Context, requesterUserID, userPTPackageID uint) (*gcv1.UserPTPackageRes, error)
	CheckoutPTPackageVNPay(ctx context.Context, userID, packageID uint, clientIP string) (*gcv1.CheckoutRes, error)
	ConfirmPTPackageVNPay(ctx context.Context, userID uint, params map[string]string) (*gcv1.UserPTPackageRes, error)
	ActivateMembershipFromStripeWebhook(ctx context.Context, sess *stripe.CheckoutSession) error
	ActivatePTPackageFromStripeWebhook(ctx context.Context, sess *stripe.CheckoutSession) error
	CheckoutPTPackageStripe(ctx context.Context, userID, packageID uint) (*gcv1.CheckoutRes, error)
	ConfirmPTPackageStripe(ctx context.Context, userID uint, sessionID string) (*gcv1.UserPTPackageRes, error)
	LogPTSession(ctx context.Context, trainerUserID, userPTPackageID uint, req *gcv1.LogPTSessionReq) (*gcv1.PTSessionLogRes, error)
	ListPTSessions(ctx context.Context, requesterUserID, userPTPackageID uint) (*gcv1.ListRes, error)
	ListChatMessages(ctx context.Context, requesterUserID, userPTPackageID, afterID uint, limit int) (*gcv1.ChatMessagesRes, error)
	SendChatMessage(ctx context.Context, requesterUserID, userPTPackageID uint, req *gcv1.SendChatMessageReq) (*gcv1.ChatMessageRes, error)
	CreateSessionOffer(ctx context.Context, requesterUserID, userPTPackageID uint, req *gcv1.CreateSessionOfferReq) (*gcv1.SessionOfferRes, error)
	ListSessionOffers(ctx context.Context, requesterUserID, userPTPackageID uint) (*gcv1.ListRes, error)
	AcceptSessionOffer(ctx context.Context, requesterUserID, userPTPackageID, offerID uint) (*gcv1.SessionOfferRes, error)
	DeclineSessionOffer(ctx context.Context, requesterUserID, userPTPackageID, offerID uint) (*gcv1.SessionOfferRes, error)
	CancelSessionOffer(ctx context.Context, requesterUserID, userPTPackageID, offerID uint) (*gcv1.SessionOfferRes, error)
	CompleteSessionOffer(ctx context.Context, requesterUserID, userPTPackageID, offerID uint, req *gcv1.CompleteSessionOfferReq) (*gcv1.SessionOfferRes, error)
	ConfirmSessionOffer(ctx context.Context, requesterUserID, userPTPackageID, offerID uint) (*gcv1.SessionOfferRes, error)
	RejectSessionOfferProof(ctx context.Context, requesterUserID, userPTPackageID, offerID uint) (*gcv1.SessionOfferRes, error)
	MarkSessionNoShow(ctx context.Context, requesterUserID, userPTPackageID, offerID uint) (*gcv1.SessionOfferRes, error)

	// User / public — PT calendar booking
	GetMyBookingSettings(ctx context.Context, trainerUserID uint) (*gcv1.BookingSettingsRes, error)
	UpdateMyBookingSettings(ctx context.Context, trainerUserID uint, req *gcv1.BookingSettingsReq) (*gcv1.BookingSettingsRes, error)
	GetMyWorkingHours(ctx context.Context, trainerUserID uint) (*gcv1.ListRes, error)
	SetMyWorkingHours(ctx context.Context, trainerUserID uint, req *gcv1.SetWorkingHoursReq) (*gcv1.ListRes, error)
	ListMyBlockedSlots(ctx context.Context, trainerUserID uint, from, to time.Time) (*gcv1.ListRes, error)
	BlockMySlot(ctx context.Context, trainerUserID uint, req *gcv1.BlockSlotReq) (*gcv1.BlockedSlotRes, error)
	UnblockMySlot(ctx context.Context, trainerUserID, blockID uint) error
	ListAvailableSlots(ctx context.Context, trainerProfileID uint, from, to time.Time) (*gcv1.ListRes, error)
	BookSlot(ctx context.Context, studentUserID uint, req *gcv1.BookSlotReq) (*gcv1.SessionOfferRes, error)
	RescheduleSessionOffer(ctx context.Context, requesterUserID, userPTPackageID, offerID uint, newStartsAt time.Time) (*gcv1.SessionOfferRes, error)
	AutoConfirmExpiredSessionProofs(ctx context.Context, olderThan time.Duration, limit int) (int, error)
	ExpireStalePendingOffers(ctx context.Context, olderThan time.Duration) (int, error)
	RunExpiryHousekeeping(ctx context.Context) error

	ConfigureOps(mailer MailerService, premium PremiumFromMembership, jwtSecret string, checkIns checkInStore)

	// Admin — PT ops
	AdminTrainerOpsOverview(ctx context.Context, trainerProfileID uint) (*gcv1.TrainerOpsOverviewRes, error)
	AdminListTrainerClients(ctx context.Context, trainerProfileID uint, page, limit int, status string) (*gcv1.ListRes, error)
	AdminListTrainerBookings(ctx context.Context, trainerProfileID uint, from, to time.Time) (*gcv1.ListRes, error)
	AdminContentFunnel(ctx context.Context, trainerProfileID uint) (*gcv1.ContentFunnelRes, error)
	AdminTrainerQuality(ctx context.Context, trainerProfileID uint) (*gcv1.TrainerQualityRes, error)
	AdminTrainerCalendar(ctx context.Context, trainerProfileID uint, from, to time.Time) (*gcv1.TrainerCalendarRes, error)
	CreatePTReview(ctx context.Context, studentUserID uint, req *gcv1.CreatePTReviewReq) (*gcv1.PTReviewRes, error)
	ListPTReviews(ctx context.Context, trainerProfileID uint, page, limit int) (*gcv1.ListRes, error)
	TouchPTFunnel(ctx context.Context, viewerUserID uint, req *gcv1.TouchPTFunnelReq) error

	// User — PT earnings
	MyPTEarnings(ctx context.Context, trainerUserID uint, page, limit int) (*gcv1.EarningsSummaryRes, error)

	// Public
	ListPublicPlans(ctx context.Context, page, limit int) (*gcv1.ListRes, error)
	ListHighlightedPlans(ctx context.Context, page, limit int) (*gcv1.ListRes, error)
	ListPublicUpcomingSessions(ctx context.Context, page, limit int) (*gcv1.ListRes, error)
	ListPublicPTPackagesByTrainer(ctx context.Context, trainerProfileID uint, page, limit int) (*gcv1.ListRes, error)
}

type gymCommerceService struct {
	planRepo       repository.GymMembershipPlanRepository
	membRepo       repository.UserGymMembershipRepository
	classRepo      repository.GroupClassRepository
	sessionRepo    repository.ClassSessionRepository
	bookingRepo    repository.ClassBookingRepository
	ptPkgRepo      repository.PTPackageRepository
	userPtPkgRepo  repository.UserPTPackageRepository
	sessionLogRepo repository.PTSessionLogRepository
	offerRepo      repository.PTSessionOfferRepository
	chatRepo       repository.PTPackageChatRepository
	hoursRepo      repository.PTWorkingHoursRepository
	blockedRepo    repository.PTBlockedSlotRepository
	statRepo       repository.PTContentStatRepository
	attrRepo       repository.PTAttributionRepository
	reviewRepo     repository.PTReviewRepository
	growth         PTGrowthTracker
	revShareRepo   repository.RevenueShareSettingRepository
	earningRepo    repository.PTEarningRepository
	trainerRepo     repository.TrainerProfileRepository
	userRepo       repository.UserRepository
	vnpay           VNPayService
	stripe          StripeService
	mailer          MailerService
	premiumGrant    PremiumFromMembership
	jwtSecret       []byte
	checkInCreate   checkInStore
	vnpayCfgRet     struct {
		Membership string
		Package    string
	}
	stripeCfgRet struct {
		MembershipSuccess string
		MembershipCancel  string
		PackageSuccess    string
		PackageCancel     string
	}
}

// NewGymCommerceService wires all repositories needed for the gym commerce module.
func NewGymCommerceService(
	planRepo repository.GymMembershipPlanRepository,
	membRepo repository.UserGymMembershipRepository,
	classRepo repository.GroupClassRepository,
	sessionRepo repository.ClassSessionRepository,
	bookingRepo repository.ClassBookingRepository,
	ptPkgRepo repository.PTPackageRepository,
	userPtPkgRepo repository.UserPTPackageRepository,
	sessionLogRepo repository.PTSessionLogRepository,
	offerRepo repository.PTSessionOfferRepository,
	chatRepo repository.PTPackageChatRepository,
	hoursRepo repository.PTWorkingHoursRepository,
	blockedRepo repository.PTBlockedSlotRepository,
	statRepo repository.PTContentStatRepository,
	attrRepo repository.PTAttributionRepository,
	reviewRepo repository.PTReviewRepository,
	growth PTGrowthTracker,
	revShareRepo repository.RevenueShareSettingRepository,
	earningRepo repository.PTEarningRepository,
	trainerRepo repository.TrainerProfileRepository,
	userRepo repository.UserRepository,
	vnpay VNPayService,
	stripe StripeService,
	membershipReturnURL string,
	packageReturnURL string,
	membershipStripeSuccessURL string,
	membershipStripeCancelURL string,
	packageStripeSuccessURL string,
	packageStripeCancelURL string,
) GymCommerceService {
	s := &gymCommerceService{
		planRepo: planRepo, membRepo: membRepo, classRepo: classRepo, sessionRepo: sessionRepo,
		bookingRepo: bookingRepo, ptPkgRepo: ptPkgRepo, userPtPkgRepo: userPtPkgRepo,
		sessionLogRepo: sessionLogRepo, offerRepo: offerRepo, chatRepo: chatRepo,
		hoursRepo: hoursRepo, blockedRepo: blockedRepo,
		statRepo: statRepo, attrRepo: attrRepo, reviewRepo: reviewRepo, growth: growth,
		revShareRepo: revShareRepo, earningRepo: earningRepo,
		trainerRepo: trainerRepo, userRepo: userRepo, vnpay: vnpay, stripe: stripe,
	}
	s.vnpayCfgRet.Membership = membershipReturnURL
	s.vnpayCfgRet.Package = packageReturnURL
	s.stripeCfgRet.MembershipSuccess = membershipStripeSuccessURL
	s.stripeCfgRet.MembershipCancel = membershipStripeCancelURL
	s.stripeCfgRet.PackageSuccess = packageStripeSuccessURL
	s.stripeCfgRet.PackageCancel = packageStripeCancelURL
	return s
}

// ============================== Membership Plans (admin) ==============================

func (s *gymCommerceService) CreatePlan(ctx context.Context, req *gcv1.MembershipPlanReq) (*gcv1.MembershipPlanRes, error) {
	currency := money.Normalize(req.Currency)
	includesClasses := true
	if req.IncludesClasses != nil {
		includesClasses = *req.IncludesClasses
	}
	active := true
	if req.IsActive != nil {
		active = *req.IsActive
	}
	highlighted := false
	if req.IsHighlighted != nil {
		highlighted = *req.IsHighlighted
	}
	p := &entity.GymMembershipPlan{
		Code:            strings.TrimSpace(req.Code),
		Name:            strings.TrimSpace(req.Name),
		Description:     req.Description,
		Price:           req.Price,
		Currency:        currency,
		DurationMonths:  req.DurationMonths,
		BranchID:        req.BranchID,
		IncludesClasses: includesClasses,
		IsHighlighted:   highlighted,
		IsActive:        active,
		SortOrder:       req.SortOrder,
	}
	if err := s.planRepo.Create(ctx, p); err != nil {
		return nil, err
	}
	fresh, err := s.planRepo.GetByID(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	res := toMembershipPlanRes(fresh)
	return &res, nil
}

func (s *gymCommerceService) UpdatePlan(ctx context.Context, id uint, req *gcv1.MembershipPlanReq) (*gcv1.MembershipPlanRes, error) {
	p, err := s.planRepo.GetByID(ctx, id)
	if err != nil {
		return nil, notFoundOr(err, "plan not found")
	}
	p.Code = strings.TrimSpace(req.Code)
	p.Name = strings.TrimSpace(req.Name)
	p.Description = req.Description
	p.Price = req.Price
	if strings.TrimSpace(req.Currency) != "" {
		p.Currency = money.Normalize(req.Currency)
	}
	p.DurationMonths = req.DurationMonths
	p.BranchID = req.BranchID
	if req.IncludesClasses != nil {
		p.IncludesClasses = *req.IncludesClasses
	}
	if req.IsHighlighted != nil {
		p.IsHighlighted = *req.IsHighlighted
	}
	if req.IsActive != nil {
		p.IsActive = *req.IsActive
	}
	p.SortOrder = req.SortOrder
	if err := s.planRepo.Update(ctx, p); err != nil {
		return nil, err
	}
	fresh, err := s.planRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	res := toMembershipPlanRes(fresh)
	return &res, nil
}

func (s *gymCommerceService) DeletePlan(ctx context.Context, id uint) error {
	return s.planRepo.Delete(ctx, id)
}

func (s *gymCommerceService) GetPlan(ctx context.Context, id uint) (*gcv1.MembershipPlanRes, error) {
	p, err := s.planRepo.GetByID(ctx, id)
	if err != nil {
		return nil, notFoundOr(err, "plan not found")
	}
	res := toMembershipPlanRes(p)
	return &res, nil
}

func (s *gymCommerceService) ListPlansAdmin(ctx context.Context, page, limit int, q string, active *bool) (*gcv1.ListRes, error) {
	page, limit = pageLimit(page, limit)
	rows, total, err := s.planRepo.List(ctx, (page-1)*limit, limit, q, active, nil, nil)
	if err != nil {
		return nil, err
	}
	out := make([]gcv1.MembershipPlanRes, 0, len(rows))
	for i := range rows {
		out = append(out, toMembershipPlanRes(&rows[i]))
	}
	return &gcv1.ListRes{Total: total, Data: out}, nil
}

func (s *gymCommerceService) ListPublicPlans(ctx context.Context, page, limit int) (*gcv1.ListRes, error) {
	active := true
	return s.ListPlansAdmin(ctx, page, limit, "", &active)
}

func (s *gymCommerceService) ListHighlightedPlans(ctx context.Context, page, limit int) (*gcv1.ListRes, error) {
	page, limit = pageLimit(page, limit)
	active := true
	highlighted := true
	rows, total, err := s.planRepo.List(ctx, (page-1)*limit, limit, "", &active, nil, &highlighted)
	if err != nil {
		return nil, err
	}
	out := make([]gcv1.MembershipPlanRes, 0, len(rows))
	for i := range rows {
		out = append(out, toMembershipPlanRes(&rows[i]))
	}
	return &gcv1.ListRes{Total: total, Data: out}, nil
}

func (s *gymCommerceService) SetPlanHighlight(ctx context.Context, id uint, highlighted bool) (*gcv1.MembershipPlanRes, error) {
	p, err := s.planRepo.GetByID(ctx, id)
	if err != nil {
		return nil, notFoundOr(err, "plan not found")
	}
	p.IsHighlighted = highlighted
	if err := s.planRepo.Update(ctx, p); err != nil {
		return nil, err
	}
	fresh, err := s.planRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	res := toMembershipPlanRes(fresh)
	return &res, nil
}

func toMembershipPlanRes(p *entity.GymMembershipPlan) gcv1.MembershipPlanRes {
	branchName := ""
	if p.Branch != nil {
		branchName = p.Branch.Name
	}
	return gcv1.MembershipPlanRes{
		ID:              p.ID,
		Code:            p.Code,
		Name:            p.Name,
		Description:     p.Description,
		Price:           p.Price,
		Currency:        p.Currency,
		DurationMonths:  p.DurationMonths,
		BranchID:        p.BranchID,
		BranchName:      branchName,
		IncludesClasses: p.IncludesClasses,
		IsHighlighted:   p.IsHighlighted,
		IsActive:        p.IsActive,
		SortOrder:       p.SortOrder,
	}
}

// ============================== User gym memberships ==============================

func (s *gymCommerceService) ListUserGymMembershipsAdmin(ctx context.Context, page, limit int, status string, userID uint) (*gcv1.ListRes, error) {
	_ = s.membRepo.ExpireEnded(ctx, time.Now().UTC())
	page, limit = pageLimit(page, limit)
	rows, total, err := s.membRepo.ListAdmin(ctx, (page-1)*limit, limit, status, userID)
	if err != nil {
		return nil, err
	}
	out := make([]gcv1.GymMembershipRes, 0, len(rows))
	for i := range rows {
		out = append(out, toGymMembershipRes(&rows[i]))
	}
	return &gcv1.ListRes{Total: total, Data: out}, nil
}

// AdminActivateMembership lets front-desk/CSKH staff manually flip a "pending"
// order to active — e.g. the member paid cash at the counter, or the payment
// gateway confirm step failed even though the money was actually captured.
func (s *gymCommerceService) AdminActivateMembership(ctx context.Context, membershipID uint) (*gcv1.GymMembershipRes, error) {
	m, err := s.membRepo.GetByID(ctx, membershipID)
	if err != nil {
		return nil, notFoundOr(err, "membership not found")
	}
	if m.Status == entity.GymMemStatusActive {
		res := toGymMembershipRes(m)
		return &res, nil
	}
	if m.Status != entity.GymMemStatusPending {
		return nil, fmt.Errorf("only a pending membership can be activated")
	}
	if err := s.activateGymMembership(ctx, m); err != nil {
		return nil, err
	}
	fresh, err := s.membRepo.GetByID(ctx, m.ID)
	if err != nil {
		return nil, err
	}
	res := toGymMembershipRes(fresh)
	return &res, nil
}

// AdminCancelMembership lets staff cancel a membership — member requests early
// termination/refund, or a "pending" order that will never be paid and should stop
// cluttering the active list.
func (s *gymCommerceService) AdminCancelMembership(ctx context.Context, membershipID uint) (*gcv1.GymMembershipRes, error) {
	m, err := s.membRepo.GetByID(ctx, membershipID)
	if err != nil {
		return nil, notFoundOr(err, "membership not found")
	}
	if m.Status == entity.GymMemStatusCanceled || m.Status == entity.GymMemStatusExpired {
		res := toGymMembershipRes(m)
		return &res, nil
	}
	m.Status = entity.GymMemStatusCanceled
	if err := s.membRepo.Update(ctx, m); err != nil {
		return nil, err
	}
	res := toGymMembershipRes(m)
	return &res, nil
}

// AdminSetPTEarningPaidOut lets staff record that a PT has (or hasn't) been paid
// out for one ledger row, e.g. during a weekly/monthly payroll run.
func (s *gymCommerceService) AdminSetPTEarningPaidOut(ctx context.Context, earningID uint, paid bool) error {
	if _, err := s.earningRepo.GetByID(ctx, earningID); err != nil {
		return notFoundOr(err, "pt earning not found")
	}
	return s.earningRepo.MarkPaidOut(ctx, earningID, paid)
}

func (s *gymCommerceService) CheckoutMembershipVNPay(ctx context.Context, userID, planID uint, clientIP string) (*gcv1.CheckoutRes, error) {
	if s.vnpay == nil || !s.vnpay.Enabled() {
		return nil, fmt.Errorf("vnpay is not configured")
	}
	plan, err := s.planRepo.GetByID(ctx, planID)
	if err != nil {
		return nil, notFoundOr(err, "plan not found")
	}
	if !plan.IsActive {
		return nil, fmt.Errorf("plan is not available")
	}

	now := time.Now().UTC()
	m := &entity.UserGymMembership{
		UserID:              userID,
		GymMembershipPlanID: plan.ID,
		BranchID:             plan.BranchID,
		StartDate:            now,
		EndDate:              now,
		DurationMonths:       plan.DurationMonths,
		Price:                plan.Price,
		Currency: money.Normalize(plan.Currency),
		Status:               entity.GymMemStatusPending,
		PaymentProvider:      entity.PaymentProviderVNPay,
	}
	if err := s.membRepo.Create(ctx, m); err != nil {
		return nil, err
	}

	txnRef := NewVNPayTxnRefWithPrefix("GM", m.ID)
	amountVND := s.vnpay.AmountVND(plan.Price, plan.Currency)
	pay, err := s.vnpay.CreatePaymentURL(txnRef, amountVND, "TrongCon Gym Membership - "+plan.Name, clientIP, s.vnpayCfgRet.Membership)
	if err != nil {
		return nil, err
	}
	m.VnpTxnRef = pay.TxnRef
	if err := s.membRepo.Update(ctx, m); err != nil {
		return nil, err
	}
	return &gcv1.CheckoutRes{
		ID:          m.ID,
		OrderID:     pay.TxnRef,
		CheckoutURL: pay.PaymentURL,
		ApproveURL:  pay.PaymentURL,
		Provider:    entity.PaymentProviderVNPay,
	}, nil
}

func (s *gymCommerceService) ConfirmMembershipVNPay(ctx context.Context, userID uint, params map[string]string) (*gcv1.GymMembershipRes, error) {
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
	m, err := s.membRepo.GetByVnpTxnRef(ctx, verified.TxnRef)
	if err != nil {
		return nil, notFoundOr(err, "membership order not found")
	}
	if m.UserID != userID {
		return nil, fmt.Errorf("membership does not belong to user")
	}
	if m.Status == entity.GymMemStatusActive {
		res := toGymMembershipRes(m)
		return &res, nil
	}
	if !verified.Success {
		return nil, fmt.Errorf("vnpay payment not successful: %s", verified.Message)
	}

	m.VnpTransactionNo = verified.TransactionNo
	if err := s.activateGymMembership(ctx, m); err != nil {
		return nil, err
	}
	fresh, err := s.membRepo.GetByID(ctx, m.ID)
	if err != nil {
		return nil, err
	}
	res := toGymMembershipRes(fresh)
	return &res, nil
}

func (s *gymCommerceService) CheckoutMembershipStripe(ctx context.Context, userID, planID uint) (*gcv1.CheckoutRes, error) {
	if s.stripe == nil || !s.stripe.Enabled() {
		return nil, fmt.Errorf("stripe is not configured")
	}
	plan, err := s.planRepo.GetByID(ctx, planID)
	if err != nil {
		return nil, notFoundOr(err, "plan not found")
	}
	if !plan.IsActive {
		return nil, fmt.Errorf("plan is not available")
	}
	email := ""
	if u, err := s.userRepo.GetByID(ctx, userID); err == nil && u != nil {
		email = u.Email
	}

	now := time.Now().UTC()
	m := &entity.UserGymMembership{
		UserID:              userID,
		GymMembershipPlanID: plan.ID,
		BranchID:            plan.BranchID,
		StartDate:           now,
		EndDate:             now,
		DurationMonths:      plan.DurationMonths,
		Price:               plan.Price,
		Currency: money.Normalize(plan.Currency),
		Status:              entity.GymMemStatusPending,
		PaymentProvider:     entity.PaymentProviderStripe,
	}
	if err := s.membRepo.Create(ctx, m); err != nil {
		return nil, err
	}

	sess, err := s.stripe.CreateCheckout(StripeCheckoutOpts{
		PlanName:      "Gym Membership - " + plan.Name,
		AmountCents:   StripeAmountCents(plan.Price),
		Currency:      money.Normalize(plan.Currency),
		UserID:        userID,
		PlanID:        plan.ID,
		RecordID:      m.ID,
		CustomerEmail: email,
		SuccessURL:    s.stripeCfgRet.MembershipSuccess,
		CancelURL:     s.stripeCfgRet.MembershipCancel,
		MetaType:      "gym_membership",
		RecordMetaKey: "membership_id",
	})
	if err != nil {
		return nil, err
	}
	m.StripeCheckoutSessionID = sess.SessionID
	if err := s.membRepo.Update(ctx, m); err != nil {
		return nil, err
	}
	return &gcv1.CheckoutRes{
		ID:          m.ID,
		SessionID:   sess.SessionID,
		CheckoutURL: sess.CheckoutURL,
		ApproveURL:  sess.CheckoutURL,
		Provider:    entity.PaymentProviderStripe,
	}, nil
}

// findMembershipBySession resolves the pending UserGymMembership behind a Stripe
// checkout session, by session ID first and by the "membership_id" metadata as
// fallback (metadata survives even if our StripeCheckoutSessionID write failed).
func (s *gymCommerceService) findMembershipBySession(ctx context.Context, sess *stripe.CheckoutSession) (*entity.UserGymMembership, error) {
	if sess == nil {
		return nil, fmt.Errorf("nil stripe session")
	}
	if sess.PaymentStatus != stripe.CheckoutSessionPaymentStatusPaid &&
		sess.Status != stripe.CheckoutSessionStatusComplete {
		return nil, fmt.Errorf("stripe session not paid (status=%s payment=%s)", sess.Status, sess.PaymentStatus)
	}
	m, err := s.membRepo.GetByStripeCheckoutSessionID(ctx, sess.ID)
	if err != nil {
		if metaID := strings.TrimSpace(sess.Metadata["membership_id"]); metaID != "" {
			if id64, e := strconv.ParseUint(metaID, 10, 64); e == nil {
				m, err = s.membRepo.GetByID(ctx, uint(id64))
			}
		}
		if err != nil {
			return nil, notFoundOr(err, "membership for stripe session not found")
		}
		if m.StripeCheckoutSessionID == "" {
			m.StripeCheckoutSessionID = sess.ID
		}
	}
	return m, nil
}

// activateMembershipFromSession is shared by the browser-driven confirm endpoint
// and the Stripe webhook: both just need to resolve+activate, they differ only in
// whether a userID ownership check runs first.
func (s *gymCommerceService) activateMembershipFromSession(ctx context.Context, m *entity.UserGymMembership, sess *stripe.CheckoutSession) (*entity.UserGymMembership, error) {
	if m.Status == entity.GymMemStatusActive {
		return m, nil
	}
	pi := ""
	if sess.PaymentIntent != nil {
		pi = sess.PaymentIntent.ID
	}
	m.PaymentProvider = entity.PaymentProviderStripe
	m.StripeCheckoutSessionID = sess.ID
	m.StripePaymentIntentID = pi
	if err := s.activateGymMembership(ctx, m); err != nil {
		return nil, err
	}
	return s.membRepo.GetByID(ctx, m.ID)
}

func (s *gymCommerceService) ConfirmMembershipStripe(ctx context.Context, userID uint, sessionID string) (*gcv1.GymMembershipRes, error) {
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
	m, err := s.findMembershipBySession(ctx, sess)
	if err != nil {
		return nil, err
	}
	if m.UserID != userID {
		return nil, fmt.Errorf("membership does not belong to user")
	}
	fresh, err := s.activateMembershipFromSession(ctx, m, sess)
	if err != nil {
		return nil, err
	}
	res := toGymMembershipRes(fresh)
	return &res, nil
}

// ActivateMembershipFromStripeWebhook activates a gym membership purchase from a
// verified Stripe webhook event. Unlike ConfirmMembershipStripe (browser return),
// this has no userID to check — the webhook payload itself, signed by Stripe, is
// the trusted source of truth, and it is the only reliable path if the user closes
// the tab right after paying before the browser can call the confirm endpoint.
func (s *gymCommerceService) ActivateMembershipFromStripeWebhook(ctx context.Context, sess *stripe.CheckoutSession) error {
	m, err := s.findMembershipBySession(ctx, sess)
	if err != nil {
		return err
	}
	_, err = s.activateMembershipFromSession(ctx, m, sess)
	return err
}

func (s *gymCommerceService) MyMembership(ctx context.Context, userID uint) (*gcv1.MembershipMeRes, error) {
	_ = s.membRepo.ExpireEnded(ctx, time.Now().UTC())
	now := time.Now().UTC()
	active, err := s.membRepo.GetActiveByUserID(ctx, userID, now)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	recent, err := s.membRepo.ListByUserID(ctx, userID, 10)
	if err != nil {
		return nil, err
	}
	out := &gcv1.MembershipMeRes{Recent: make([]gcv1.GymMembershipRes, 0, len(recent))}
	if active != nil {
		r := toGymMembershipRes(active)
		out.Active = &r
	}
	for i := range recent {
		out.Recent = append(out.Recent, toGymMembershipRes(&recent[i]))
	}
	return out, nil
}

func toGymMembershipRes(m *entity.UserGymMembership) gcv1.GymMembershipRes {
	planName := ""
	includesClasses := true
	if m.GymMembershipPlan.ID != 0 {
		planName = m.GymMembershipPlan.Name
		includesClasses = m.GymMembershipPlan.IncludesClasses
	}
	userEmail, userName := "", ""
	if m.User.ID != 0 {
		userEmail = m.User.Email
		userName = strings.TrimSpace(m.User.Name)
	}
	return gcv1.GymMembershipRes{
		ID:                  m.ID,
		UserID:              m.UserID,
		UserEmail:           userEmail,
		UserName:            userName,
		GymMembershipPlanID: m.GymMembershipPlanID,
		PlanName:            planName,
		IncludesClasses:     includesClasses,
		BranchID:            m.BranchID,
		StartDate:           m.StartDate,
		EndDate:             m.EndDate,
		DurationMonths:      m.DurationMonths,
		Price:               m.Price,
		Currency:            m.Currency,
		Status:              m.Status,
		PaymentProvider:     m.PaymentProvider,
		VnpTxnRef:           m.VnpTxnRef,
		CreatedAt:           m.CreatedAt,
	}
}

// ============================== Group classes (admin) ==============================

func (s *gymCommerceService) CreateGroupClass(ctx context.Context, req *gcv1.GroupClassReq) (*gcv1.GroupClassRes, error) {
	durationMin := req.DurationMin
	if durationMin <= 0 {
		durationMin = 60
	}
	capacity := req.Capacity
	if capacity <= 0 {
		capacity = 20
	}
	active := true
	if req.IsActive != nil {
		active = *req.IsActive
	}
	g := &entity.GroupClass{
		BranchID:    req.BranchID,
		Name:        strings.TrimSpace(req.Name),
		Category:    strings.TrimSpace(req.Category),
		Description: req.Description,
		DurationMin: durationMin,
		Capacity:    capacity,
		TrainerID:   req.TrainerID,
		IsActive:    active,
	}
	if err := s.classRepo.Create(ctx, g); err != nil {
		return nil, err
	}
	fresh, err := s.classRepo.GetByID(ctx, g.ID)
	if err != nil {
		return nil, err
	}
	res := toGroupClassRes(fresh)
	return &res, nil
}

func (s *gymCommerceService) UpdateGroupClass(ctx context.Context, id uint, req *gcv1.GroupClassReq) (*gcv1.GroupClassRes, error) {
	g, err := s.classRepo.GetByID(ctx, id)
	if err != nil {
		return nil, notFoundOr(err, "group class not found")
	}
	g.BranchID = req.BranchID
	g.Name = strings.TrimSpace(req.Name)
	g.Category = strings.TrimSpace(req.Category)
	g.Description = req.Description
	if req.DurationMin > 0 {
		g.DurationMin = req.DurationMin
	}
	if req.Capacity > 0 {
		g.Capacity = req.Capacity
	}
	g.TrainerID = req.TrainerID
	if req.IsActive != nil {
		g.IsActive = *req.IsActive
	}
	if err := s.classRepo.Update(ctx, g); err != nil {
		return nil, err
	}
	fresh, err := s.classRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	res := toGroupClassRes(fresh)
	return &res, nil
}

func (s *gymCommerceService) DeleteGroupClass(ctx context.Context, id uint) error {
	return s.classRepo.Delete(ctx, id)
}

func (s *gymCommerceService) ListGroupClasses(ctx context.Context, page, limit int, q string, branchID *uint) (*gcv1.ListRes, error) {
	page, limit = pageLimit(page, limit)
	rows, total, err := s.classRepo.List(ctx, (page-1)*limit, limit, q, branchID, nil)
	if err != nil {
		return nil, err
	}
	out := make([]gcv1.GroupClassRes, 0, len(rows))
	for i := range rows {
		out = append(out, toGroupClassRes(&rows[i]))
	}
	return &gcv1.ListRes{Total: total, Data: out}, nil
}

func toGroupClassRes(g *entity.GroupClass) gcv1.GroupClassRes {
	branchName := ""
	if g.Branch.ID != 0 {
		branchName = g.Branch.Name
	}
	return gcv1.GroupClassRes{
		ID:          g.ID,
		BranchID:    g.BranchID,
		BranchName:  branchName,
		Name:        g.Name,
		Category:    g.Category,
		Description: g.Description,
		DurationMin: g.DurationMin,
		Capacity:    g.Capacity,
		TrainerID:   g.TrainerID,
		IsActive:    g.IsActive,
	}
}

// ============================== Class sessions ==============================

func (s *gymCommerceService) CreateClassSession(ctx context.Context, req *gcv1.ClassSessionReq) (*gcv1.ClassSessionRes, error) {
	if !req.EndsAt.After(req.StartsAt) {
		return nil, fmt.Errorf("ends_at must be after starts_at")
	}
	gc, err := s.classRepo.GetByID(ctx, req.GroupClassID)
	if err != nil {
		return nil, notFoundOr(err, "group class not found")
	}
	capacity := req.Capacity
	if capacity <= 0 {
		capacity = gc.Capacity
	}
	sess := &entity.ClassSession{
		GroupClassID: req.GroupClassID,
		StartsAt:     req.StartsAt,
		EndsAt:       req.EndsAt,
		Capacity:     capacity,
	}
	if err := s.sessionRepo.Create(ctx, sess); err != nil {
		return nil, err
	}
	fresh, err := s.sessionRepo.GetByID(ctx, sess.ID)
	if err != nil {
		return nil, err
	}
	res := toClassSessionRes(fresh)
	return &res, nil
}

func (s *gymCommerceService) DeleteClassSession(ctx context.Context, id uint) error {
	return s.sessionRepo.Delete(ctx, id)
}

func (s *gymCommerceService) ListClassSessionsAdmin(ctx context.Context, page, limit int, groupClassID *uint) (*gcv1.ListRes, error) {
	page, limit = pageLimit(page, limit)
	rows, total, err := s.sessionRepo.List(ctx, (page-1)*limit, limit, groupClassID, nil)
	if err != nil {
		return nil, err
	}
	out := make([]gcv1.ClassSessionRes, 0, len(rows))
	for i := range rows {
		out = append(out, toClassSessionRes(&rows[i]))
	}
	return &gcv1.ListRes{Total: total, Data: out}, nil
}

func (s *gymCommerceService) ListUpcomingClassSessions(ctx context.Context, page, limit int) (*gcv1.ListRes, error) {
	page, limit = pageLimit(page, limit)
	rows, total, err := s.sessionRepo.ListUpcoming(ctx, (page-1)*limit, limit, nil)
	if err != nil {
		return nil, err
	}
	out := make([]gcv1.ClassSessionRes, 0, len(rows))
	for i := range rows {
		out = append(out, toClassSessionRes(&rows[i]))
	}
	return &gcv1.ListRes{Total: total, Data: out}, nil
}

func (s *gymCommerceService) ListPublicUpcomingSessions(ctx context.Context, page, limit int) (*gcv1.ListRes, error) {
	return s.ListUpcomingClassSessions(ctx, page, limit)
}

func toClassSessionRes(s *entity.ClassSession) gcv1.ClassSessionRes {
	className, category := "", ""
	var branchID uint
	branchName := ""
	if s.GroupClass.ID != 0 {
		className = s.GroupClass.Name
		category = s.GroupClass.Category
		branchID = s.GroupClass.BranchID
		if s.GroupClass.Branch.ID != 0 {
			branchName = s.GroupClass.Branch.Name
		}
	}
	return gcv1.ClassSessionRes{
		ID:           s.ID,
		GroupClassID: s.GroupClassID,
		ClassName:    className,
		Category:     category,
		BranchID:     branchID,
		BranchName:   branchName,
		StartsAt:     s.StartsAt,
		EndsAt:       s.EndsAt,
		Capacity:     s.Capacity,
		BookedCount:  s.BookedCount,
		IsCanceled:   s.IsCanceled,
	}
}

// ============================== Class bookings (user) ==============================

func (s *gymCommerceService) BookClassSession(ctx context.Context, userID, sessionID uint) (*gcv1.ClassBookingRes, error) {
	sess, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return nil, notFoundOr(err, "class session not found")
	}
	if sess.IsCanceled {
		return nil, fmt.Errorf("class session is canceled")
	}
	if sess.StartsAt.Before(time.Now().UTC()) {
		return nil, fmt.Errorf("class session already started")
	}
	if sess.BookedCount >= sess.Capacity {
		return nil, fmt.Errorf("class session is full")
	}
	if err := s.assertCanBookClass(ctx, userID, sess); err != nil {
		return nil, err
	}
	if existing, err := s.bookingRepo.GetActiveByUserAndSession(ctx, userID, sessionID); err == nil && existing != nil {
		return nil, fmt.Errorf("already booked")
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	b := &entity.ClassBooking{
		UserID:         userID,
		ClassSessionID: sessionID,
		Status:         entity.ClassBookingBooked,
	}
	if err := s.bookingRepo.Create(ctx, b); err != nil {
		return nil, err
	}
	if err := s.sessionRepo.IncrementBooked(ctx, sessionID, 1); err != nil {
		return nil, err
	}
	fresh, err := s.bookingRepo.GetByID(ctx, b.ID)
	if err != nil {
		return nil, err
	}
	res := toClassBookingRes(fresh)
	return &res, nil
}

func (s *gymCommerceService) CancelClassBooking(ctx context.Context, userID, bookingID uint) error {
	b, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return notFoundOr(err, "booking not found")
	}
	if b.UserID != userID {
		return fmt.Errorf("booking does not belong to user")
	}
	if b.Status == entity.ClassBookingCanceled {
		return nil
	}
	b.Status = entity.ClassBookingCanceled
	if err := s.bookingRepo.Update(ctx, b); err != nil {
		return err
	}
	return s.sessionRepo.IncrementBooked(ctx, b.ClassSessionID, -1)
}

func (s *gymCommerceService) MyClassBookings(ctx context.Context, userID uint) (*gcv1.ListRes, error) {
	rows, err := s.bookingRepo.ListByUserID(ctx, userID, 50)
	if err != nil {
		return nil, err
	}
	out := make([]gcv1.ClassBookingRes, 0, len(rows))
	for i := range rows {
		out = append(out, toClassBookingRes(&rows[i]))
	}
	return &gcv1.ListRes{Total: int64(len(out)), Data: out}, nil
}

func toClassBookingRes(b *entity.ClassBooking) gcv1.ClassBookingRes {
	className := ""
	var startsAt time.Time
	if b.ClassSession.ID != 0 {
		startsAt = b.ClassSession.StartsAt
		if b.ClassSession.GroupClass.ID != 0 {
			className = b.ClassSession.GroupClass.Name
		}
	}
	return gcv1.ClassBookingRes{
		ID:             b.ID,
		ClassSessionID: b.ClassSessionID,
		Status:         b.Status,
		ClassName:      className,
		StartsAt:       startsAt,
		CreatedAt:      b.CreatedAt,
	}
}

// ============================== PT packages (owned by trainer) ==============================

func (s *gymCommerceService) trainerProfileForUser(ctx context.Context, userID uint) (*entity.TrainerProfile, error) {
	t, err := s.trainerRepo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("trainer profile required")
		}
		return nil, err
	}
	return t, nil
}

func (s *gymCommerceService) CreatePTPackage(ctx context.Context, trainerUserID uint, req *gcv1.PTPackageReq) (*gcv1.PTPackageRes, error) {
	t, err := s.trainerProfileForUser(ctx, trainerUserID)
	if err != nil {
		return nil, err
	}
	currency := money.Normalize(req.Currency)
	validDays := req.ValidDays
	if validDays <= 0 {
		validDays = 90
	}
	isPublic := false
	if req.IsPublic != nil {
		isPublic = *req.IsPublic
	}
	active := true
	if req.IsActive != nil {
		active = *req.IsActive
	}
	p := &entity.PTPackage{
		TrainerProfileID: t.ID,
		Title:            strings.TrimSpace(req.Title),
		Description:      req.Description,
		SessionCount:     req.SessionCount,
		Price:            req.Price,
		Currency:         currency,
		ValidDays:        validDays,
		IsPublic:         isPublic,
		IsActive:         active,
	}
	if err := s.ptPkgRepo.Create(ctx, p); err != nil {
		return nil, err
	}
	fresh, err := s.ptPkgRepo.GetByID(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	res := toPTPackageRes(fresh)
	return &res, nil
}

func (s *gymCommerceService) UpdatePTPackage(ctx context.Context, trainerUserID, id uint, req *gcv1.PTPackageReq) (*gcv1.PTPackageRes, error) {
	t, err := s.trainerProfileForUser(ctx, trainerUserID)
	if err != nil {
		return nil, err
	}
	p, err := s.ptPkgRepo.GetByID(ctx, id)
	if err != nil {
		return nil, notFoundOr(err, "pt package not found")
	}
	if p.TrainerProfileID != t.ID {
		return nil, fmt.Errorf("package does not belong to trainer")
	}
	p.Title = strings.TrimSpace(req.Title)
	p.Description = req.Description
	if req.SessionCount > 0 {
		p.SessionCount = req.SessionCount
	}
	p.Price = req.Price
	if strings.TrimSpace(req.Currency) != "" {
		p.Currency = money.Normalize(req.Currency)
	}
	if req.ValidDays > 0 {
		p.ValidDays = req.ValidDays
	}
	if req.IsPublic != nil {
		p.IsPublic = *req.IsPublic
	}
	if req.IsActive != nil {
		p.IsActive = *req.IsActive
	}
	if err := s.ptPkgRepo.Update(ctx, p); err != nil {
		return nil, err
	}
	fresh, err := s.ptPkgRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	res := toPTPackageRes(fresh)
	return &res, nil
}

func (s *gymCommerceService) DeletePTPackage(ctx context.Context, trainerUserID, id uint) error {
	t, err := s.trainerProfileForUser(ctx, trainerUserID)
	if err != nil {
		return err
	}
	p, err := s.ptPkgRepo.GetByID(ctx, id)
	if err != nil {
		return notFoundOr(err, "pt package not found")
	}
	if p.TrainerProfileID != t.ID {
		return fmt.Errorf("package does not belong to trainer")
	}
	return s.ptPkgRepo.Delete(ctx, id)
}

func (s *gymCommerceService) ListMyPTPackages(ctx context.Context, trainerUserID uint, page, limit int) (*gcv1.ListRes, error) {
	t, err := s.trainerProfileForUser(ctx, trainerUserID)
	if err != nil {
		return nil, err
	}
	page, limit = pageLimit(page, limit)
	rows, total, err := s.ptPkgRepo.ListByTrainer(ctx, (page-1)*limit, limit, t.ID, nil)
	if err != nil {
		return nil, err
	}
	out := make([]gcv1.PTPackageRes, 0, len(rows))
	for i := range rows {
		out = append(out, toPTPackageRes(&rows[i]))
	}
	return &gcv1.ListRes{Total: total, Data: out}, nil
}

func (s *gymCommerceService) ListPTPackagesAdmin(ctx context.Context, page, limit int, q string, trainerProfileID uint) (*gcv1.ListRes, error) {
	page, limit = pageLimit(page, limit)
	rows, total, err := s.ptPkgRepo.ListAdmin(ctx, (page-1)*limit, limit, q, trainerProfileID)
	if err != nil {
		return nil, err
	}
	out := make([]gcv1.PTPackageRes, 0, len(rows))
	for i := range rows {
		out = append(out, toPTPackageRes(&rows[i]))
	}
	return &gcv1.ListRes{Total: total, Data: out}, nil
}

func (s *gymCommerceService) ListPublicPTPackagesByTrainer(ctx context.Context, trainerProfileID uint, page, limit int) (*gcv1.ListRes, error) {
	page, limit = pageLimit(page, limit)
	rows, total, err := s.ptPkgRepo.ListPublicByTrainer(ctx, (page-1)*limit, limit, trainerProfileID)
	if err != nil {
		return nil, err
	}
	out := make([]gcv1.PTPackageRes, 0, len(rows))
	for i := range rows {
		out = append(out, toPTPackageRes(&rows[i]))
	}
	return &gcv1.ListRes{Total: total, Data: out}, nil
}

func toPTPackageRes(p *entity.PTPackage) gcv1.PTPackageRes {
	trainerName := ""
	if p.Trainer.ID != 0 {
		trainerName = p.Trainer.DisplayName
	}
	return gcv1.PTPackageRes{
		ID:               p.ID,
		TrainerProfileID: p.TrainerProfileID,
		TrainerName:      trainerName,
		Title:            p.Title,
		Description:      p.Description,
		SessionCount:     p.SessionCount,
		Price:            p.Price,
		Currency:         p.Currency,
		ValidDays:        p.ValidDays,
		IsPublic:         p.IsPublic,
		IsActive:         p.IsActive,
	}
}

// ============================== PT packages (purchased by user) ==============================

func (s *gymCommerceService) ListPurchasedPTPackages(ctx context.Context, userID uint) (*gcv1.ListRes, error) {
	_ = s.userPtPkgRepo.ExpireEnded(ctx, time.Now().UTC())
	rows, err := s.userPtPkgRepo.ListByUserID(ctx, userID, 50)
	if err != nil {
		return nil, err
	}
	out := make([]gcv1.UserPTPackageRes, 0, len(rows))
	for i := range rows {
		out = append(out, toUserPTPackageRes(&rows[i]))
	}
	if err := s.enrichPackageInbox(ctx, userID, out); err != nil {
		return nil, err
	}
	return &gcv1.ListRes{Total: int64(len(out)), Data: out}, nil
}

func (s *gymCommerceService) CheckoutPTPackageVNPay(ctx context.Context, userID, packageID uint, clientIP string) (*gcv1.CheckoutRes, error) {
	if s.vnpay == nil || !s.vnpay.Enabled() {
		return nil, fmt.Errorf("vnpay is not configured")
	}
	pkg, err := s.ptPkgRepo.GetByID(ctx, packageID)
	if err != nil {
		return nil, notFoundOr(err, "pt package not found")
	}
	if !pkg.IsActive {
		return nil, fmt.Errorf("pt package is not available")
	}
	if err := s.assertCanPurchasePTPackage(ctx, userID, pkg.TrainerProfileID); err != nil {
		return nil, err
	}

	up := &entity.UserPTPackage{
		UserID:           userID,
		PTPackageID:      pkg.ID,
		TrainerProfileID: pkg.TrainerProfileID,
		SessionTotal:     pkg.SessionCount,
		Price:            pkg.Price,
		Currency: money.Normalize(pkg.Currency),
		Status:           entity.PTPkgStatusPending,
		PaymentProvider:  entity.PaymentProviderVNPay,
	}
	if err := s.userPtPkgRepo.Create(ctx, up); err != nil {
		return nil, err
	}

	txnRef := NewVNPayTxnRefWithPrefix("PP", up.ID)
	amountVND := s.vnpay.AmountVND(pkg.Price, pkg.Currency)
	pay, err := s.vnpay.CreatePaymentURL(txnRef, amountVND, "TrongCon PT Package - "+pkg.Title, clientIP, s.vnpayCfgRet.Package)
	if err != nil {
		return nil, err
	}
	up.VnpTxnRef = pay.TxnRef
	if err := s.userPtPkgRepo.Update(ctx, up); err != nil {
		return nil, err
	}
	return &gcv1.CheckoutRes{
		ID:          up.ID,
		OrderID:     pay.TxnRef,
		CheckoutURL: pay.PaymentURL,
		ApproveURL:  pay.PaymentURL,
		Provider:    entity.PaymentProviderVNPay,
	}, nil
}

func (s *gymCommerceService) ConfirmPTPackageVNPay(ctx context.Context, userID uint, params map[string]string) (*gcv1.UserPTPackageRes, error) {
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
	up, err := s.userPtPkgRepo.GetByVnpTxnRef(ctx, verified.TxnRef)
	if err != nil {
		return nil, notFoundOr(err, "pt package order not found")
	}
	if up.UserID != userID {
		return nil, fmt.Errorf("package does not belong to user")
	}
	if up.Status == entity.PTPkgStatusActive {
		res := toUserPTPackageRes(up)
		return &res, nil
	}
	if !verified.Success {
		return nil, fmt.Errorf("vnpay payment not successful: %s", verified.Message)
	}

	now := time.Now().UTC()
	validDays := 90
	if pkg, err := s.ptPkgRepo.GetByID(ctx, up.PTPackageID); err == nil {
		validDays = pkg.ValidDays
	}
	up.StartsAt = now
	up.ExpiresAt = now.AddDate(0, 0, validDays)
	up.Status = entity.PTPkgStatusActive
	up.VnpTransactionNo = verified.TransactionNo
	claimed, err := s.userPtPkgRepo.ActivateFromPending(ctx, up)
	if err != nil {
		return nil, err
	}
	if !claimed {
		fresh, err := s.userPtPkgRepo.GetByID(ctx, up.ID)
		if err != nil {
			return nil, err
		}
		res := toUserPTPackageRes(fresh)
		return &res, nil
	}
	// No earning is booked here: PT commission is recorded per session as it
	// is actually taught (see recordPTSessionEarning), not at purchase time.
	if s.growth != nil {
		s.growth.TrackBooking(ctx, up.TrainerProfileID, up.UserID, "", 0, "")
	}
	fresh, err := s.userPtPkgRepo.GetByID(ctx, up.ID)
	if err != nil {
		return nil, err
	}
	s.notifyPTPackagePurchased(ctx, fresh)
	res := toUserPTPackageRes(fresh)
	return &res, nil
}

func (s *gymCommerceService) CheckoutPTPackageStripe(ctx context.Context, userID, packageID uint) (*gcv1.CheckoutRes, error) {
	if s.stripe == nil || !s.stripe.Enabled() {
		return nil, fmt.Errorf("stripe is not configured")
	}
	pkg, err := s.ptPkgRepo.GetByID(ctx, packageID)
	if err != nil {
		return nil, notFoundOr(err, "pt package not found")
	}
	if !pkg.IsActive {
		return nil, fmt.Errorf("pt package is not available")
	}
	if err := s.assertCanPurchasePTPackage(ctx, userID, pkg.TrainerProfileID); err != nil {
		return nil, err
	}
	email := ""
	if u, err := s.userRepo.GetByID(ctx, userID); err == nil && u != nil {
		email = u.Email
	}

	up := &entity.UserPTPackage{
		UserID:           userID,
		PTPackageID:      pkg.ID,
		TrainerProfileID: pkg.TrainerProfileID,
		SessionTotal:     pkg.SessionCount,
		Price:            pkg.Price,
		Currency: money.Normalize(pkg.Currency),
		Status:           entity.PTPkgStatusPending,
		PaymentProvider:  entity.PaymentProviderStripe,
	}
	if err := s.userPtPkgRepo.Create(ctx, up); err != nil {
		return nil, err
	}

	sess, err := s.stripe.CreateCheckout(StripeCheckoutOpts{
		PlanName:      "PT Package - " + pkg.Title,
		AmountCents:   StripeAmountCents(pkg.Price),
		Currency:      money.Normalize(pkg.Currency),
		UserID:        userID,
		PlanID:        pkg.ID,
		RecordID:      up.ID,
		CustomerEmail: email,
		SuccessURL:    s.stripeCfgRet.PackageSuccess,
		CancelURL:     s.stripeCfgRet.PackageCancel,
		MetaType:      "pt_package",
		RecordMetaKey: "user_pt_package_id",
	})
	if err != nil {
		return nil, err
	}
	up.StripeCheckoutSessionID = sess.SessionID
	if err := s.userPtPkgRepo.Update(ctx, up); err != nil {
		return nil, err
	}
	return &gcv1.CheckoutRes{
		ID:          up.ID,
		SessionID:   sess.SessionID,
		CheckoutURL: sess.CheckoutURL,
		ApproveURL:  sess.CheckoutURL,
		Provider:    entity.PaymentProviderStripe,
	}, nil
}

// findPTPackageBySession resolves the pending UserPTPackage behind a Stripe
// checkout session, by session ID first and by the "user_pt_package_id" metadata
// as fallback.
func (s *gymCommerceService) findPTPackageBySession(ctx context.Context, sess *stripe.CheckoutSession) (*entity.UserPTPackage, error) {
	if sess == nil {
		return nil, fmt.Errorf("nil stripe session")
	}
	if sess.PaymentStatus != stripe.CheckoutSessionPaymentStatusPaid &&
		sess.Status != stripe.CheckoutSessionStatusComplete {
		return nil, fmt.Errorf("stripe session not paid (status=%s payment=%s)", sess.Status, sess.PaymentStatus)
	}
	up, err := s.userPtPkgRepo.GetByStripeCheckoutSessionID(ctx, sess.ID)
	if err != nil {
		if metaID := strings.TrimSpace(sess.Metadata["user_pt_package_id"]); metaID != "" {
			if id64, e := strconv.ParseUint(metaID, 10, 64); e == nil {
				up, err = s.userPtPkgRepo.GetByID(ctx, uint(id64))
			}
		}
		if err != nil {
			return nil, notFoundOr(err, "pt package order for stripe session not found")
		}
		if up.StripeCheckoutSessionID == "" {
			up.StripeCheckoutSessionID = sess.ID
		}
	}
	return up, nil
}

// activatePTPackageFromSession is shared by the browser-driven confirm endpoint
// and the Stripe webhook.
func (s *gymCommerceService) activatePTPackageFromSession(ctx context.Context, up *entity.UserPTPackage, sess *stripe.CheckoutSession) (*entity.UserPTPackage, error) {
	if up.Status == entity.PTPkgStatusActive {
		return up, nil
	}
	pi := ""
	if sess.PaymentIntent != nil {
		pi = sess.PaymentIntent.ID
	}
	now := time.Now().UTC()
	validDays := 90
	if pkg, err := s.ptPkgRepo.GetByID(ctx, up.PTPackageID); err == nil {
		validDays = pkg.ValidDays
	}
	up.StartsAt = now
	up.ExpiresAt = now.AddDate(0, 0, validDays)
	up.Status = entity.PTPkgStatusActive
	up.PaymentProvider = entity.PaymentProviderStripe
	up.StripeCheckoutSessionID = sess.ID
	up.StripePaymentIntentID = pi
	claimed, err := s.userPtPkgRepo.ActivateFromPending(ctx, up)
	if err != nil {
		return nil, err
	}
	if !claimed {
		return s.userPtPkgRepo.GetByID(ctx, up.ID)
	}
	// No earning is booked here: PT commission is recorded per session as it
	// is actually taught (see recordPTSessionEarning), not at purchase time.
	if s.growth != nil {
		s.growth.TrackBooking(ctx, up.TrainerProfileID, up.UserID, "", 0, "")
	}
	fresh, err := s.userPtPkgRepo.GetByID(ctx, up.ID)
	if err != nil {
		return nil, err
	}
	s.notifyPTPackagePurchased(ctx, fresh)
	return fresh, nil
}

func (s *gymCommerceService) ConfirmPTPackageStripe(ctx context.Context, userID uint, sessionID string) (*gcv1.UserPTPackageRes, error) {
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
	up, err := s.findPTPackageBySession(ctx, sess)
	if err != nil {
		return nil, err
	}
	if up.UserID != userID {
		return nil, fmt.Errorf("package does not belong to user")
	}
	fresh, err := s.activatePTPackageFromSession(ctx, up, sess)
	if err != nil {
		return nil, err
	}
	res := toUserPTPackageRes(fresh)
	return &res, nil
}

// ActivatePTPackageFromStripeWebhook mirrors ActivateMembershipFromStripeWebhook
// for PT packages — see that comment for why the webhook skips the userID check.
func (s *gymCommerceService) ActivatePTPackageFromStripeWebhook(ctx context.Context, sess *stripe.CheckoutSession) error {
	up, err := s.findPTPackageBySession(ctx, sess)
	if err != nil {
		return err
	}
	_, err = s.activatePTPackageFromSession(ctx, up, sess)
	return err
}

// recordPTSessionEarning books the PT/gym revenue split for exactly one
// delivered (or no-show) session, not for the whole package at purchase time.
// A PT is only paid for sessions actually taught — unused sessions in an
// expired/canceled package never generate an earning row.
func (s *gymCommerceService) recordPTSessionEarning(ctx context.Context, up *entity.UserPTPackage, sessionIndex int, sessionOfferID uint, note string) error {
	share, err := s.revShareRepo.GetSingleton(ctx)
	if err != nil {
		return err
	}
	sessionTotal := up.SessionTotal
	if sessionTotal <= 0 {
		sessionTotal = 1
	}
	gross := up.Price / float64(sessionTotal)
	ptAmount := gross * share.PTPercent / 100
	gymAmount := gross - ptAmount
	e := &entity.PTEarning{
		TrainerProfileID: up.TrainerProfileID,
		UserPTPackageID:  up.ID,
		SessionIndex:     sessionIndex,
		GrossAmount:      gross,
		PTPercent:        share.PTPercent,
		PTAmount:         ptAmount,
		GymAmount:        gymAmount,
		Currency:         up.Currency,
		Note:             note,
	}
	if sessionOfferID != 0 {
		e.SessionOfferID = &sessionOfferID
	}
	return s.earningRepo.Create(ctx, e)
}

func (s *gymCommerceService) LogPTSession(ctx context.Context, trainerUserID, userPTPackageID uint, req *gcv1.LogPTSessionReq) (*gcv1.PTSessionLogRes, error) {
	_, _, _, _ = ctx, trainerUserID, userPTPackageID, req
	return nil, fmt.Errorf("use session offers: propose a time in chat, accept it, then complete with photo proof")
}

func (s *gymCommerceService) ListPTSessions(ctx context.Context, requesterUserID, userPTPackageID uint) (*gcv1.ListRes, error) {
	up, err := s.userPtPkgRepo.GetByID(ctx, userPTPackageID)
	if err != nil {
		return nil, notFoundOr(err, "user pt package not found")
	}
	if err := s.assertCanViewUserPTPackage(ctx, requesterUserID, up); err != nil {
		return nil, err
	}
	rows, err := s.sessionLogRepo.ListByUserPTPackageID(ctx, userPTPackageID)
	if err != nil {
		return nil, err
	}
	out := make([]gcv1.PTSessionLogRes, 0, len(rows))
	for i := range rows {
		out = append(out, toPTSessionLogRes(&rows[i]))
	}
	return &gcv1.ListRes{Total: int64(len(out)), Data: out}, nil
}

func (s *gymCommerceService) ListPTSessionsAdmin(ctx context.Context, userPTPackageID uint) (*gcv1.ListRes, error) {
	if _, err := s.userPtPkgRepo.GetByID(ctx, userPTPackageID); err != nil {
		return nil, notFoundOr(err, "user pt package not found")
	}
	rows, err := s.sessionLogRepo.ListByUserPTPackageID(ctx, userPTPackageID)
	if err != nil {
		return nil, err
	}
	out := make([]gcv1.PTSessionLogRes, 0, len(rows))
	for i := range rows {
		out = append(out, toPTSessionLogRes(&rows[i]))
	}
	return &gcv1.ListRes{Total: int64(len(out)), Data: out}, nil
}

func (s *gymCommerceService) ListSoldPTPackages(ctx context.Context, trainerUserID uint, page, limit int, status string) (*gcv1.ListRes, error) {
	t, err := s.trainerProfileForUser(ctx, trainerUserID)
	if err != nil {
		return nil, err
	}
	_ = s.userPtPkgRepo.ExpireEnded(ctx, time.Now().UTC())
	page, limit = pageLimit(page, limit)
	rows, total, err := s.userPtPkgRepo.ListByTrainerProfileID(ctx, t.ID, (page-1)*limit, limit, status)
	if err != nil {
		return nil, err
	}
	out := make([]gcv1.UserPTPackageRes, 0, len(rows))
	for i := range rows {
		out = append(out, toUserPTPackageRes(&rows[i]))
	}
	if err := s.enrichPackageInbox(ctx, trainerUserID, out); err != nil {
		return nil, err
	}
	return &gcv1.ListRes{Total: total, Data: out}, nil
}

func (s *gymCommerceService) ListUserPTPackagesAdmin(ctx context.Context, page, limit int, status string, trainerProfileID, userID uint) (*gcv1.ListRes, error) {
	_ = s.userPtPkgRepo.ExpireEnded(ctx, time.Now().UTC())
	page, limit = pageLimit(page, limit)
	rows, total, err := s.userPtPkgRepo.ListAdmin(ctx, (page-1)*limit, limit, status, trainerProfileID, userID)
	if err != nil {
		return nil, err
	}
	out := make([]gcv1.UserPTPackageRes, 0, len(rows))
	for i := range rows {
		out = append(out, toUserPTPackageRes(&rows[i]))
	}
	return &gcv1.ListRes{Total: total, Data: out}, nil
}

func (s *gymCommerceService) GetUserPTPackage(ctx context.Context, requesterUserID, userPTPackageID uint) (*gcv1.UserPTPackageRes, error) {
	up, err := s.userPtPkgRepo.GetByID(ctx, userPTPackageID)
	if err != nil {
		return nil, notFoundOr(err, "user pt package not found")
	}
	if err := s.assertCanViewUserPTPackage(ctx, requesterUserID, up); err != nil {
		return nil, err
	}
	out := []gcv1.UserPTPackageRes{toUserPTPackageRes(up)}
	if err := s.enrichPackageInbox(ctx, requesterUserID, out); err != nil {
		return nil, err
	}
	return &out[0], nil
}

func (s *gymCommerceService) enrichPackageInbox(ctx context.Context, viewerUserID uint, out []gcv1.UserPTPackageRes) error {
	if len(out) == 0 {
		return nil
	}
	ids := make([]uint, 0, len(out))
	for i := range out {
		ids = append(ids, out[i].ID)
	}
	unread, err := s.chatRepo.CountUnreadByPackages(ctx, viewerUserID, ids)
	if err != nil {
		return err
	}
	latest, err := s.chatRepo.GetLatestByPackages(ctx, ids)
	if err != nil {
		return err
	}
	pending, err := s.offerRepo.CountPendingByPackages(ctx, ids)
	if err != nil {
		return err
	}
	for i := range out {
		out[i].UnreadCount = unread[out[i].ID]
		out[i].PendingOffers = pending[out[i].ID]
		if msg, ok := latest[out[i].ID]; ok {
			body := strings.TrimSpace(msg.Body)
			if len(body) > 120 {
				body = body[:117] + "…"
			}
			out[i].LastMessage = body
			t := msg.CreatedAt
			out[i].LastMessageAt = &t
		}
	}
	return nil
}

func (s *gymCommerceService) assertCanViewUserPTPackage(ctx context.Context, requesterUserID uint, up *entity.UserPTPackage) error {
	if up.UserID == requesterUserID {
		return nil
	}
	t, err := s.trainerProfileForUser(ctx, requesterUserID)
	if err == nil && t.ID == up.TrainerProfileID {
		return nil
	}
	return fmt.Errorf("unauthorized")
}

func (s *gymCommerceService) ListChatMessages(ctx context.Context, requesterUserID, userPTPackageID, afterID uint, limit int) (*gcv1.ChatMessagesRes, error) {
	up, err := s.userPtPkgRepo.GetByID(ctx, userPTPackageID)
	if err != nil {
		return nil, notFoundOr(err, "user pt package not found")
	}
	if err := s.assertCanViewUserPTPackage(ctx, requesterUserID, up); err != nil {
		return nil, err
	}
	rows, err := s.chatRepo.ListByPackage(ctx, userPTPackageID, afterID, limit)
	if err != nil {
		return nil, err
	}
	total, err := s.chatRepo.CountByPackage(ctx, userPTPackageID)
	if err != nil {
		return nil, err
	}
	offerIDs := make([]uint, 0)
	for i := range rows {
		if rows[i].SessionOfferID != nil && *rows[i].SessionOfferID > 0 {
			offerIDs = append(offerIDs, *rows[i].SessionOfferID)
		}
	}
	offerMap := map[uint]entity.PTSessionOffer{}
	if len(offerIDs) > 0 {
		offers, err := s.offerRepo.GetByIDs(ctx, offerIDs)
		if err != nil {
			return nil, err
		}
		for i := range offers {
			offerMap[offers[i].ID] = offers[i]
		}
	}
	out := make([]gcv1.ChatMessageRes, 0, len(rows))
	for i := range rows {
		out = append(out, s.toChatMessageRes(ctx, &rows[i], offerMap))
	}
	if maxID, err := s.chatRepo.GetMaxMessageID(ctx, userPTPackageID); err == nil && maxID > 0 {
		_ = s.chatRepo.UpsertRead(ctx, requesterUserID, userPTPackageID, maxID)
	}
	return &gcv1.ChatMessagesRes{
		Total:     total,
		Data:      out,
		CanSend:   up.Status == entity.PTPkgStatusActive,
		PackageID: up.ID,
	}, nil
}

func (s *gymCommerceService) SendChatMessage(ctx context.Context, requesterUserID, userPTPackageID uint, req *gcv1.SendChatMessageReq) (*gcv1.ChatMessageRes, error) {
	body := strings.TrimSpace(req.Body)
	if body == "" {
		return nil, fmt.Errorf("message body is required")
	}
	if len(body) > 4000 {
		return nil, fmt.Errorf("message too long (max 4000 characters)")
	}
	up, err := s.userPtPkgRepo.GetByID(ctx, userPTPackageID)
	if err != nil {
		return nil, notFoundOr(err, "user pt package not found")
	}
	if err := s.assertCanViewUserPTPackage(ctx, requesterUserID, up); err != nil {
		return nil, err
	}
	if up.Status != entity.PTPkgStatusActive {
		return nil, fmt.Errorf("chat is only available while the package is active")
	}
	msg := &entity.PTPackageChatMessage{
		UserPTPackageID: up.ID,
		SenderUserID:    requesterUserID,
		Body:            body,
		MessageType:     entity.ChatMsgTypeText,
	}
	if err := s.chatRepo.Create(ctx, msg); err != nil {
		return nil, err
	}
	res := s.toChatMessageRes(ctx, msg, nil)
	return &res, nil
}

func (s *gymCommerceService) CreateSessionOffer(ctx context.Context, requesterUserID, userPTPackageID uint, req *gcv1.CreateSessionOfferReq) (*gcv1.SessionOfferRes, error) {
	up, err := s.userPtPkgRepo.GetByID(ctx, userPTPackageID)
	if err != nil {
		return nil, notFoundOr(err, "user pt package not found")
	}
	if err := s.assertCanViewUserPTPackage(ctx, requesterUserID, up); err != nil {
		return nil, err
	}
	if up.Status != entity.PTPkgStatusActive {
		return nil, fmt.Errorf("package is not active")
	}
	if err := s.assertSessionsAvailable(ctx, up); err != nil {
		return nil, err
	}
	startsAt := req.StartsAt.UTC()
	if startsAt.IsZero() {
		return nil, fmt.Errorf("starts_at is required")
	}
	var endsAt *time.Time
	if t, err := s.trainerRepo.GetByID(ctx, up.TrainerProfileID); err == nil && t != nil {
		e := startsAt.Add(s.sessionDuration(t))
		endsAt = &e
	}
	offer := &entity.PTSessionOffer{
		UserPTPackageID:  up.ID,
		TrainerProfileID: up.TrainerProfileID,
		StudentUserID:    up.UserID,
		StartsAt:         startsAt,
		EndsAt:           endsAt,
		Note:             strings.TrimSpace(req.Note),
		ProposedByUserID: requesterUserID,
		Status:           entity.SessionOfferPending,
	}
	if err := s.offerRepo.Create(ctx, offer); err != nil {
		return nil, err
	}
	offerID := offer.ID
	body := fmt.Sprintf("Proposed session: %s", startsAt.Format(time.RFC3339))
	if offer.Note != "" {
		body += "\n" + offer.Note
	}
	msg := &entity.PTPackageChatMessage{
		UserPTPackageID: up.ID,
		SenderUserID:    requesterUserID,
		Body:            body,
		MessageType:     entity.ChatMsgTypeSessionOffer,
		SessionOfferID:  &offerID,
	}
	if err := s.chatRepo.Create(ctx, msg); err != nil {
		return nil, err
	}
	// Notify the other party.
	otherID := up.UserID
	if requesterUserID == up.UserID {
		if t, err := s.trainerRepo.GetByID(ctx, up.TrainerProfileID); err == nil && t != nil {
			otherID = t.UserID
		}
	}
	fromEmail, fromName := s.userEmail(ctx, requesterUserID)
	_ = fromEmail
	toEmail, toName := s.userEmail(ctx, otherID)
	s.notifyEmail(ctx, "pt_session_proposed", map[string]interface{}{
		"UserName":  toName,
		"FromName":  fromName,
		"StartsAt":  startsAt.In(vnLocation()).Format("15:04 02/01/2006"),
	}, toEmail)
	res := toSessionOfferRes(offer)
	return &res, nil
}

func (s *gymCommerceService) ListSessionOffers(ctx context.Context, requesterUserID, userPTPackageID uint) (*gcv1.ListRes, error) {
	up, err := s.userPtPkgRepo.GetByID(ctx, userPTPackageID)
	if err != nil {
		return nil, notFoundOr(err, "user pt package not found")
	}
	if err := s.assertCanViewUserPTPackage(ctx, requesterUserID, up); err != nil {
		return nil, err
	}
	rows, err := s.offerRepo.ListByUserPTPackageID(ctx, userPTPackageID)
	if err != nil {
		return nil, err
	}
	out := make([]gcv1.SessionOfferRes, 0, len(rows))
	for i := range rows {
		out = append(out, toSessionOfferRes(&rows[i]))
	}
	return &gcv1.ListRes{Total: int64(len(out)), Data: out}, nil
}

func (s *gymCommerceService) AcceptSessionOffer(ctx context.Context, requesterUserID, userPTPackageID, offerID uint) (*gcv1.SessionOfferRes, error) {
	offer, up, err := s.getPackageOffer(ctx, requesterUserID, userPTPackageID, offerID)
	if err != nil {
		return nil, err
	}
	if offer.Status != entity.SessionOfferPending {
		return nil, fmt.Errorf("offer is not pending")
	}
	if offer.ProposedByUserID == requesterUserID {
		return nil, fmt.Errorf("proposer already accepted; wait for the other party")
	}
	if up.SessionUsed >= up.SessionTotal {
		return nil, fmt.Errorf("no sessions left")
	}
	if t, err := s.trainerRepo.GetByID(ctx, up.TrainerProfileID); err == nil && t != nil {
		dur := s.sessionDuration(t)
		end := offer.StartsAt.Add(dur)
		if offer.EndsAt != nil {
			end = *offer.EndsAt
		}
		busy, bErr := s.offerRepo.ListBusyInRange(ctx, up.TrainerProfileID, offer.StartsAt.Add(-time.Minute), end.Add(time.Minute))
		if bErr == nil {
			for i := range busy {
				if busy[i].ID == offer.ID {
					continue
				}
				if rangesOverlap(offer.StartsAt.UTC(), end.UTC(), busy[i].StartsAt.UTC(), offerEnd(&busy[i], dur)) {
					return nil, fmt.Errorf("time conflicts with another booking")
				}
			}
		}
		if err := s.assertStudentNotBusy(ctx, up.UserID, offer.ID, offer.StartsAt, end); err != nil {
			return nil, err
		}
	}
	now := time.Now().UTC()
	offer.Status = entity.SessionOfferScheduled
	offer.AcceptedByUserID = requesterUserID
	offer.AcceptedAt = &now
	if offer.EndsAt == nil {
		if t, err := s.trainerRepo.GetByID(ctx, up.TrainerProfileID); err == nil && t != nil {
			e := offer.StartsAt.Add(s.sessionDuration(t))
			offer.EndsAt = &e
		}
	}
	if err := s.offerRepo.Update(ctx, offer); err != nil {
		return nil, err
	}
	res := toSessionOfferRes(offer)
	return &res, nil
}

func (s *gymCommerceService) DeclineSessionOffer(ctx context.Context, requesterUserID, userPTPackageID, offerID uint) (*gcv1.SessionOfferRes, error) {
	offer, _, err := s.getPackageOffer(ctx, requesterUserID, userPTPackageID, offerID)
	if err != nil {
		return nil, err
	}
	if offer.Status != entity.SessionOfferPending {
		return nil, fmt.Errorf("offer is not pending")
	}
	if offer.ProposedByUserID == requesterUserID {
		return nil, fmt.Errorf("use cancel instead of decline for your own offer")
	}
	offer.Status = entity.SessionOfferDeclined
	if err := s.offerRepo.Update(ctx, offer); err != nil {
		return nil, err
	}
	res := toSessionOfferRes(offer)
	return &res, nil
}

func (s *gymCommerceService) CancelSessionOffer(ctx context.Context, requesterUserID, userPTPackageID, offerID uint) (*gcv1.SessionOfferRes, error) {
	offer, _, err := s.getPackageOffer(ctx, requesterUserID, userPTPackageID, offerID)
	if err != nil {
		return nil, err
	}
	switch offer.Status {
	case entity.SessionOfferPending:
		// A pending offer is only the proposer's to retract; the other party
		// should decline it instead (DeclineSessionOffer), not cancel it.
		if offer.ProposedByUserID != requesterUserID {
			return nil, fmt.Errorf("only the proposer can cancel a pending offer; use decline instead")
		}
	case entity.SessionOfferScheduled:
		// Once both sides accepted, either the student or the trainer may
		// cancel — getPackageOffer already proved requesterUserID is one of them.
	default:
		return nil, fmt.Errorf("offer cannot be cancelled")
	}
	offer.Status = entity.SessionOfferCancelled
	if err := s.offerRepo.Update(ctx, offer); err != nil {
		return nil, err
	}
	res := toSessionOfferRes(offer)
	return &res, nil
}

// RescheduleSessionOffer moves an already-agreed (scheduled) session to a new
// time, re-running the same trainer/student conflict checks as booking a new
// slot — without this, changing a time meant cancel + re-propose + re-accept,
// losing the offer's history and needing the other party to agree all over again.
func (s *gymCommerceService) RescheduleSessionOffer(ctx context.Context, requesterUserID, userPTPackageID, offerID uint, newStartsAt time.Time) (*gcv1.SessionOfferRes, error) {
	offer, up, err := s.getPackageOffer(ctx, requesterUserID, userPTPackageID, offerID)
	if err != nil {
		return nil, err
	}
	if offer.Status != entity.SessionOfferScheduled {
		return nil, fmt.Errorf("only a scheduled session can be rescheduled")
	}
	newStartsAt = newStartsAt.UTC()
	if !newStartsAt.After(time.Now().UTC()) {
		return nil, fmt.Errorf("new time must be in the future")
	}
	t, err := s.trainerRepo.GetByID(ctx, up.TrainerProfileID)
	if err != nil || t == nil {
		return nil, fmt.Errorf("trainer not found")
	}
	dur := s.sessionDuration(t)
	newEnd := newStartsAt.Add(dur)
	busy, err := s.offerRepo.ListBusyInRange(ctx, up.TrainerProfileID, newStartsAt.Add(-time.Minute), newEnd.Add(time.Minute))
	if err == nil {
		for i := range busy {
			if busy[i].ID == offer.ID {
				continue
			}
			if rangesOverlap(newStartsAt, newEnd, busy[i].StartsAt.UTC(), offerEnd(&busy[i], dur)) {
				return nil, fmt.Errorf("new time conflicts with another booking")
			}
		}
	}
	if err := s.assertStudentNotBusy(ctx, up.UserID, offer.ID, newStartsAt, newEnd); err != nil {
		return nil, err
	}
	offer.StartsAt = newStartsAt
	offer.EndsAt = &newEnd
	if err := s.offerRepo.Update(ctx, offer); err != nil {
		return nil, err
	}
	res := toSessionOfferRes(offer)
	return &res, nil
}

func (s *gymCommerceService) CompleteSessionOffer(ctx context.Context, requesterUserID, userPTPackageID, offerID uint, req *gcv1.CompleteSessionOfferReq) (*gcv1.SessionOfferRes, error) {
	proof := strings.TrimSpace(req.ProofImageURL)
	if proof == "" {
		return nil, fmt.Errorf("proof_image_url is required")
	}
	offer, up, err := s.getPackageOffer(ctx, requesterUserID, userPTPackageID, offerID)
	if err != nil {
		return nil, err
	}
	t, err := s.trainerProfileForUser(ctx, requesterUserID)
	if err != nil || t.ID != up.TrainerProfileID {
		return nil, fmt.Errorf("only the package trainer can complete a session")
	}
	if offer.Status != entity.SessionOfferScheduled {
		return nil, fmt.Errorf("offer must be scheduled before completion")
	}
	if up.Status != entity.PTPkgStatusActive {
		return nil, fmt.Errorf("package is not active")
	}
	if up.SessionUsed >= up.SessionTotal {
		return nil, fmt.Errorf("no sessions left")
	}
	now := time.Now().UTC()
	if now.Before(offer.StartsAt) {
		return nil, fmt.Errorf("cannot complete a session before its scheduled start time")
	}

	note := strings.TrimSpace(req.Note)
	if note != "" {
		offer.Note = note
	}
	// PT submits proof — student must confirm before the session is counted.
	offer.Status = entity.SessionOfferAwaitingConfirmation
	offer.ProofImageURL = proof
	offer.ProofSubmittedAt = &now
	offer.CompletedByUserID = requesterUserID
	offer.CompletedAt = nil
	offer.SessionIndex = 0
	offer.ConfirmedByUserID = 0
	offer.ConfirmedAt = nil
	if err := s.offerRepo.Update(ctx, offer); err != nil {
		return nil, err
	}
	res := toSessionOfferRes(offer)
	return &res, nil
}

func (s *gymCommerceService) ConfirmSessionOffer(ctx context.Context, requesterUserID, userPTPackageID, offerID uint) (*gcv1.SessionOfferRes, error) {
	offer, up, err := s.getPackageOffer(ctx, requesterUserID, userPTPackageID, offerID)
	if err != nil {
		return nil, err
	}
	if up.UserID != requesterUserID {
		return nil, fmt.Errorf("only the student can confirm the session")
	}
	return s.finalizeSessionConfirmation(ctx, offer, requesterUserID)
}

func (s *gymCommerceService) RejectSessionOfferProof(ctx context.Context, requesterUserID, userPTPackageID, offerID uint) (*gcv1.SessionOfferRes, error) {
	offer, up, err := s.getPackageOffer(ctx, requesterUserID, userPTPackageID, offerID)
	if err != nil {
		return nil, err
	}
	if up.UserID != requesterUserID {
		return nil, fmt.Errorf("only the student can reject session proof")
	}
	if offer.Status != entity.SessionOfferAwaitingConfirmation {
		return nil, fmt.Errorf("offer is not awaiting confirmation")
	}
	// Back to scheduled so PT can upload a new proof photo.
	offer.Status = entity.SessionOfferScheduled
	offer.ProofImageURL = ""
	offer.ProofSubmittedAt = nil
	offer.CompletedByUserID = 0
	offer.CompletedAt = nil
	offer.SessionIndex = 0
	offer.ConfirmedByUserID = 0
	offer.ConfirmedAt = nil
	if err := s.offerRepo.Update(ctx, offer); err != nil {
		return nil, err
	}
	res := toSessionOfferRes(offer)
	return &res, nil
}

func (s *gymCommerceService) getPackageOffer(ctx context.Context, requesterUserID, userPTPackageID, offerID uint) (*entity.PTSessionOffer, *entity.UserPTPackage, error) {
	up, err := s.userPtPkgRepo.GetByID(ctx, userPTPackageID)
	if err != nil {
		return nil, nil, notFoundOr(err, "user pt package not found")
	}
	if err := s.assertCanViewUserPTPackage(ctx, requesterUserID, up); err != nil {
		return nil, nil, err
	}
	offer, err := s.offerRepo.GetByID(ctx, offerID)
	if err != nil {
		return nil, nil, notFoundOr(err, "session offer not found")
	}
	if offer.UserPTPackageID != up.ID {
		return nil, nil, fmt.Errorf("offer does not belong to package")
	}
	return offer, up, nil
}

func (s *gymCommerceService) assertSessionsAvailable(ctx context.Context, up *entity.UserPTPackage) error {
	if up.SessionUsed >= up.SessionTotal {
		return fmt.Errorf("no sessions left")
	}
	open, err := s.offerRepo.CountOpenByPackage(ctx, up.ID)
	if err != nil {
		return err
	}
	if int(up.SessionUsed)+int(open) >= up.SessionTotal {
		return fmt.Errorf("no open session slots left (pending/scheduled offers count toward the package)")
	}
	return nil
}

func (s *gymCommerceService) toChatMessageRes(ctx context.Context, m *entity.PTPackageChatMessage, offerMap map[uint]entity.PTSessionOffer) gcv1.ChatMessageRes {
	name, avatar := "", ""
	if u, err := s.userRepo.GetByID(ctx, m.SenderUserID); err == nil && u != nil {
		name = strings.TrimSpace(u.Name)
		if name == "" {
			name = strings.TrimSpace(u.FirstName + " " + u.LastName)
		}
		if name == "" {
			name = u.Email
		}
		avatar = strings.TrimSpace(u.ProfilePicture)
	}
	if s.trainerRepo != nil {
		if t, err := s.trainerRepo.GetByUserID(ctx, m.SenderUserID); err == nil && t != nil {
			if dn := strings.TrimSpace(t.DisplayName); dn != "" && (name == "" || strings.Contains(name, "@")) {
				name = dn
			}
		}
	}
	msgType := strings.TrimSpace(m.MessageType)
	if msgType == "" {
		msgType = entity.ChatMsgTypeText
	}
	res := gcv1.ChatMessageRes{
		ID:              m.ID,
		UserPTPackageID: m.UserPTPackageID,
		SenderUserID:    m.SenderUserID,
		SenderName:      name,
		SenderAvatarURL: avatar,
		Body:            m.Body,
		MessageType:     msgType,
		SessionOfferID:  m.SessionOfferID,
		CreatedAt:       m.CreatedAt,
	}
	if m.SessionOfferID != nil && offerMap != nil {
		if o, ok := offerMap[*m.SessionOfferID]; ok {
			or := toSessionOfferRes(&o)
			res.SessionOffer = &or
		}
	}
	return res
}

func toSessionOfferRes(o *entity.PTSessionOffer) gcv1.SessionOfferRes {
	return gcv1.SessionOfferRes{
		ID:                o.ID,
		UserPTPackageID:   o.UserPTPackageID,
		TrainerProfileID:  o.TrainerProfileID,
		StudentUserID:     o.StudentUserID,
		StartsAt:          o.StartsAt,
		Note:              o.Note,
		ProposedByUserID:  o.ProposedByUserID,
		Status:            o.Status,
		AcceptedByUserID:  o.AcceptedByUserID,
		AcceptedAt:        o.AcceptedAt,
		CompletedAt:       o.CompletedAt,
		ProofImageURL:     o.ProofImageURL,
		SessionIndex:      o.SessionIndex,
		CompletedByUserID: o.CompletedByUserID,
		ConfirmedByUserID: o.ConfirmedByUserID,
		ConfirmedAt:       o.ConfirmedAt,
		ProofSubmittedAt:  o.ProofSubmittedAt,
		EndsAt:            o.EndsAt,
		BookedViaSlot:     o.BookedViaSlot,
		CreatedAt:         o.CreatedAt,
	}
}

func toPTSessionLogRes(l *entity.PTSessionLog) gcv1.PTSessionLogRes {
	return gcv1.PTSessionLogRes{
		ID:               l.ID,
		UserPTPackageID:  l.UserPTPackageID,
		TrainerProfileID: l.TrainerProfileID,
		UserID:           l.UserID,
		SessionIndex:     l.SessionIndex,
		TaughtAt:         l.TaughtAt,
		Note:             l.Note,
		ProofImageURL:    l.ProofImageURL,
		CreatedByUserID:  l.CreatedByUserID,
		CreatedAt:        l.CreatedAt,
	}
}

func toUserPTPackageRes(p *entity.UserPTPackage) gcv1.UserPTPackageRes {
	packageTitle, trainerName, studentName, studentEmail := "", "", "", ""
	if p.PTPackage.ID != 0 {
		packageTitle = p.PTPackage.Title
		if p.PTPackage.Trainer.ID != 0 {
			trainerName = p.PTPackage.Trainer.DisplayName
		}
	}
	if p.User.ID != 0 {
		studentEmail = p.User.Email
		studentName = strings.TrimSpace(p.User.Name)
		if studentName == "" {
			studentName = strings.TrimSpace(strings.TrimSpace(p.User.FirstName + " " + p.User.LastName))
		}
		if studentName == "" {
			studentName = studentEmail
		}
	}
	return gcv1.UserPTPackageRes{
		ID:               p.ID,
		UserID:           p.UserID,
		StudentName:      studentName,
		StudentEmail:     studentEmail,
		PTPackageID:      p.PTPackageID,
		PackageTitle:     packageTitle,
		TrainerProfileID: p.TrainerProfileID,
		TrainerName:      trainerName,
		SessionTotal:     p.SessionTotal,
		SessionUsed:      p.SessionUsed,
		SessionLeft:      p.SessionTotal - p.SessionUsed,
		Price:            p.Price,
		Currency:         p.Currency,
		Status:           p.Status,
		StartsAt:         p.StartsAt,
		ExpiresAt:        p.ExpiresAt,
		PaymentProvider:  p.PaymentProvider,
		VnpTxnRef:        p.VnpTxnRef,
		VnpTransactionNo: p.VnpTransactionNo,
	}
}

// ============================== Revenue share ==============================

func (s *gymCommerceService) GetRevenueShare(ctx context.Context) (*gcv1.RevenueShareRes, error) {
	share, err := s.revShareRepo.GetSingleton(ctx)
	if err != nil {
		return nil, err
	}
	return toRevenueShareRes(share), nil
}

func (s *gymCommerceService) UpdateRevenueShare(ctx context.Context, req *gcv1.RevenueShareReq) (*gcv1.RevenueShareRes, error) {
	share, err := s.revShareRepo.GetSingleton(ctx)
	if err != nil {
		return nil, err
	}
	share.PTPercent = req.PTPercent
	share.GymPercent = req.GymPercent
	share.Notes = req.Notes
	if err := s.revShareRepo.Update(ctx, share); err != nil {
		return nil, err
	}
	return toRevenueShareRes(share), nil
}

func toRevenueShareRes(s *entity.RevenueShareSetting) *gcv1.RevenueShareRes {
	return &gcv1.RevenueShareRes{
		ID:         s.ID,
		PTPercent:  s.PTPercent,
		GymPercent: s.GymPercent,
		Notes:      s.Notes,
	}
}

// ============================== PT earnings ==============================

func (s *gymCommerceService) ListPTEarningsAdmin(ctx context.Context, page, limit int, trainerProfileID uint) (*gcv1.EarningsSummaryRes, error) {
	page, limit = pageLimit(page, limit)
	rows, total, err := s.earningRepo.ListAdmin(ctx, (page-1)*limit, limit, trainerProfileID)
	if err != nil {
		return nil, err
	}
	sum, err := s.earningRepo.SumPTAmount(ctx, trainerProfileID)
	if err != nil {
		return nil, err
	}
	out := make([]gcv1.PTEarningRes, 0, len(rows))
	currency := money.DefaultCurrency
	for i := range rows {
		out = append(out, toPTEarningRes(&rows[i]))
		if rows[i].Currency != "" {
			currency = money.Normalize(rows[i].Currency)
		}
	}
	return &gcv1.EarningsSummaryRes{TotalPTAmount: sum, Currency: currency, Data: out, Total: total}, nil
}

func (s *gymCommerceService) MyPTEarnings(ctx context.Context, trainerUserID uint, page, limit int) (*gcv1.EarningsSummaryRes, error) {
	t, err := s.trainerProfileForUser(ctx, trainerUserID)
	if err != nil {
		return nil, err
	}
	return s.ListPTEarningsAdmin(ctx, page, limit, t.ID)
}

func toPTEarningRes(e *entity.PTEarning) gcv1.PTEarningRes {
	trainerName, packageTitle, studentName, studentEmail := "", "", "", ""
	if e.Trainer.ID != 0 {
		trainerName = e.Trainer.DisplayName
	}
	if e.UserPTPackage.ID != 0 {
		if e.UserPTPackage.PTPackage.ID != 0 {
			packageTitle = e.UserPTPackage.PTPackage.Title
		}
		if e.UserPTPackage.User.ID != 0 {
			studentEmail = e.UserPTPackage.User.Email
			studentName = strings.TrimSpace(e.UserPTPackage.User.Name)
			if studentName == "" {
				studentName = strings.TrimSpace(e.UserPTPackage.User.FirstName + " " + e.UserPTPackage.User.LastName)
			}
			if studentName == "" {
				studentName = studentEmail
			}
		}
	}
	return gcv1.PTEarningRes{
		ID:               e.ID,
		TrainerProfileID: e.TrainerProfileID,
		TrainerName:      trainerName,
		UserPTPackageID:  e.UserPTPackageID,
		PackageTitle:     packageTitle,
		StudentName:      studentName,
		StudentEmail:     studentEmail,
		GrossAmount:      e.GrossAmount,
		PTPercent:        e.PTPercent,
		PTAmount:         e.PTAmount,
		GymAmount:        e.GymAmount,
		Currency:         e.Currency,
		Note:             e.Note,
		PaidOut:          e.PaidOut,
		CreatedAt:        e.CreatedAt,
	}
}

// ============================== helpers ==============================

func notFoundOr(err error, msg string) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New(msg)
	}
	return err
}
