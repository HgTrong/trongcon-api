package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	gcv1 "trongcon-api/api/gym_commerce/v1"
	"trongcon-api/internal/entity"
	"trongcon-api/internal/jwtutil"

	"gorm.io/gorm"
)

type checkInStore interface {
	Create(ctx context.Context, row *entity.GymCheckIn) error
	ListRecent(ctx context.Context, limit int) ([]entity.GymCheckIn, error)
	CountToday(ctx context.Context) (int64, int64, error)
	CheckedInOn(ctx context.Context, userID uint, day time.Time) (bool, error)
}

// PremiumFromMembership is implemented by UserSubscriptionService.
type PremiumFromMembership interface {
	GrantPremiumFromMembership(ctx context.Context, userID uint, endDate time.Time) error
}

func (s *gymCommerceService) SetMailer(m MailerService)               { s.mailer = m }
func (s *gymCommerceService) SetPremiumGrant(p PremiumFromMembership) { s.premiumGrant = p }
func (s *gymCommerceService) SetJWTSecret(secret string)              { s.jwtSecret = []byte(secret) }
func (s *gymCommerceService) SetCheckInRepo(r checkInStore)           { s.checkInCreate = r }

func (s *gymCommerceService) ConfigureOps(mailer MailerService, premium PremiumFromMembership, jwtSecret string, checkIns checkInStore) {
	s.mailer = mailer
	s.premiumGrant = premium
	s.jwtSecret = []byte(jwtSecret)
	s.checkInCreate = checkIns
}

func (s *gymCommerceService) notifyEmail(ctx context.Context, key string, data map[string]interface{}, to string) {
	if s.mailer == nil || !s.mailer.Enabled() || to == "" {
		return
	}
	if err := s.mailer.SendByKey(ctx, key, data, []string{to}); err != nil {
		log.Printf("mail %s → %s: %v", key, to, err)
	}
}

func (s *gymCommerceService) userEmail(ctx context.Context, userID uint) (email, name string) {
	if s.userRepo == nil || userID == 0 {
		return "", ""
	}
	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil || u == nil {
		return "", ""
	}
	name = u.Name
	if name == "" {
		name = u.Email
	}
	return u.Email, name
}

// activateGymMembership sets dates/status, handles renew extend, grants Premium, emails.
// Idempotent: only the first pending→active transition sends email / grants premium side-effects.
func (s *gymCommerceService) activateGymMembership(ctx context.Context, m *entity.UserGymMembership) error {
	if m.Status == entity.GymMemStatusActive {
		return nil
	}
	now := time.Now().UTC()
	months := m.DurationMonths
	if months < 1 {
		months = 1
	}

	claimed, end, err := s.membRepo.ActivateWithRenew(ctx, m, months, now)
	if err != nil {
		return err
	}
	m.StartDate = now
	m.EndDate = end
	m.Status = entity.GymMemStatusActive
	if !claimed {
		// Concurrent confirm already activated — do not email again.
		if fresh, e := s.membRepo.GetByID(ctx, m.ID); e == nil && fresh != nil {
			*m = *fresh
		}
		return nil
	}

	if s.premiumGrant != nil {
		if err := s.premiumGrant.GrantPremiumFromMembership(ctx, m.UserID, end); err != nil {
			log.Printf("grant premium from membership %d: %v", m.ID, err)
		}
	}
	email, name := s.userEmail(ctx, m.UserID)
	planName := ""
	if fresh, err := s.membRepo.GetByID(ctx, m.ID); err == nil && fresh != nil {
		*m = *fresh
		if fresh.GymMembershipPlan.ID != 0 {
			planName = fresh.GymMembershipPlan.Name
		}
	}
	s.notifyEmail(ctx, "gym_membership_purchased", map[string]interface{}{
		"UserName":  name,
		"PlanName":  planName,
		"StartDate": m.StartDate.Format("02/01/2006"),
		"EndDate":   m.EndDate.Format("02/01/2006"),
	}, email)
	return nil
}

func (s *gymCommerceService) assertCanBookClass(ctx context.Context, userID uint, sess *entity.ClassSession) error {
	_ = s.membRepo.ExpireEnded(ctx, time.Now().UTC())
	m, err := s.membRepo.GetActiveByUserID(ctx, userID, time.Now().UTC())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("cần thẻ hội viên còn hiệu lực để đặt lớp nhóm")
		}
		return err
	}
	if m.GymMembershipPlan.ID != 0 && !m.GymMembershipPlan.IncludesClasses {
		return fmt.Errorf("gói hội viên của bạn không bao gồm lớp nhóm")
	}
	if m.BranchID != nil && *m.BranchID > 0 {
		// sess is always loaded via classSessionPreload, which preloads GroupClass,
		// so GroupClass.BranchID is already the real value — no fallback needed.
		if branchID := sess.GroupClass.BranchID; branchID > 0 && branchID != *m.BranchID {
			return fmt.Errorf("thẻ hội viên chỉ dùng được tại chi nhánh đã đăng ký")
		}
	}
	return nil
}

func (s *gymCommerceService) IssueCheckInToken(ctx context.Context, userID uint) (*gcv1.CheckInTokenRes, error) {
	if len(s.jwtSecret) == 0 {
		return nil, fmt.Errorf("check-in token is not configured")
	}
	_ = s.membRepo.ExpireEnded(ctx, time.Now().UTC())
	m, err := s.membRepo.GetActiveByUserID(ctx, userID, time.Now().UTC())
	if err != nil {
		return nil, fmt.Errorf("không có thẻ hội viên active để check-in")
	}
	ttl := 60 * time.Second
	token, err := jwtutil.IssueCheckIn(userID, m.ID, nil, jwtutil.PurposeGymCheckIn, s.jwtSecret, ttl)
	if err != nil {
		return nil, err
	}
	return &gcv1.CheckInTokenRes{
		Token:        token,
		ExpiresInSec: int(ttl.Seconds()),
		MembershipID: m.ID,
		PlanName:     m.GymMembershipPlan.Name,
		EndDate:      m.EndDate,
	}, nil
}

func (s *gymCommerceService) VerifyCheckIn(ctx context.Context, staffUserID uint, token string, branchID *uint, note string) (*gcv1.GymCheckInRes, error) {
	if len(s.jwtSecret) == 0 || s.checkInCreate == nil {
		return nil, fmt.Errorf("check-in is not configured")
	}
	claims, err := jwtutil.ParseWithPurpose(token, jwtutil.PurposeGymCheckIn, s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("mã QR không hợp lệ hoặc đã hết hạn")
	}
	if claims.MembershipID == 0 || claims.UserID == 0 {
		return nil, fmt.Errorf("mã QR thiếu thông tin hội viên")
	}
	_ = s.membRepo.ExpireEnded(ctx, time.Now().UTC())
	m, err := s.membRepo.GetByID(ctx, claims.MembershipID)
	if err != nil {
		return nil, notFoundOr(err, "membership not found")
	}
	if m.UserID != claims.UserID {
		return nil, fmt.Errorf("mã QR không khớp hội viên")
	}
	if m.Status != entity.GymMemStatusActive || !m.EndDate.After(time.Now().UTC()) {
		return nil, fmt.Errorf("thẻ hội viên không còn hiệu lực")
	}
	if m.BranchID != nil && branchID != nil && *m.BranchID != *branchID {
		return nil, fmt.Errorf("sai chi nhánh so với thẻ hội viên")
	}
	row := &entity.GymCheckIn{
		UserID:              m.UserID,
		UserGymMembershipID: m.ID,
		BranchID:            branchID,
		CheckedInAt:         time.Now().UTC(),
		VerifiedByUserID:    staffUserID,
		Note:                note,
	}
	if row.BranchID == nil {
		row.BranchID = m.BranchID
	}
	if err := s.checkInCreate.Create(ctx, row); err != nil {
		return nil, err
	}
	email, name := s.userEmail(ctx, m.UserID)
	return &gcv1.GymCheckInRes{
		ID:           row.ID,
		UserID:       row.UserID,
		UserName:     name,
		UserEmail:    email,
		MembershipID: row.UserGymMembershipID,
		PlanName:     m.GymMembershipPlan.Name,
		BranchID:     row.BranchID,
		CheckedInAt:  row.CheckedInAt,
		Note:         row.Note,
	}, nil
}

func (s *gymCommerceService) ListRecentCheckIns(ctx context.Context, limit int) (*gcv1.GymCheckInListRes, error) {
	if s.checkInCreate == nil {
		return &gcv1.GymCheckInListRes{Total: 0, Data: []gcv1.GymCheckInRes{}}, nil
	}
	rows, err := s.checkInCreate.ListRecent(ctx, limit)
	if err != nil {
		return nil, err
	}
	out := make([]gcv1.GymCheckInRes, 0, len(rows))
	for i := range rows {
		email, name := s.userEmail(ctx, rows[i].UserID)
		out = append(out, gcv1.GymCheckInRes{
			ID: rows[i].ID, UserID: rows[i].UserID, UserName: name, UserEmail: email,
			MembershipID: rows[i].UserGymMembershipID, BranchID: rows[i].BranchID,
			CheckedInAt: rows[i].CheckedInAt, Note: rows[i].Note,
		})
	}

	checkedInToday, uniqueToday, err := s.checkInCreate.CountToday(ctx)
	if err != nil {
		return nil, err
	}
	var activeMembers int64
	if s.membRepo != nil {
		_, activeMembers, err = s.membRepo.ListAdmin(ctx, 0, 1, entity.GymMemStatusActive, 0, nil, nil)
		if err != nil {
			return nil, err
		}
	}

	return &gcv1.GymCheckInListRes{
		Total: int64(len(out)),
		Data:  out,
		Stats: gcv1.GymCheckInStatsRes{
			ActiveMembers:      activeMembers,
			CheckedInToday:     checkedInToday,
			UniqueMembersToday: uniqueToday,
		},
	}, nil
}
