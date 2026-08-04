package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"trongcon-api/internal/entity"

	"gorm.io/gorm"
)

// —— GymMembershipPlan ——

type GymMembershipPlanRepository interface {
	Create(ctx context.Context, p *entity.GymMembershipPlan) error
	Update(ctx context.Context, p *entity.GymMembershipPlan) error
	Delete(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*entity.GymMembershipPlan, error)
	List(ctx context.Context, offset, limit int, q string, activeOnly *bool, branchID *uint, highlightedOnly *bool) ([]entity.GymMembershipPlan, int64, error)
}

type gymMembershipPlanRepository struct{ db *gorm.DB }

func NewGymMembershipPlanRepository(db *gorm.DB) GymMembershipPlanRepository {
	return &gymMembershipPlanRepository{db: db}
}

func (r *gymMembershipPlanRepository) Create(ctx context.Context, p *entity.GymMembershipPlan) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *gymMembershipPlanRepository) Update(ctx context.Context, p *entity.GymMembershipPlan) error {
	return r.db.WithContext(ctx).Omit("Branch").Save(p).Error
}

func (r *gymMembershipPlanRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.GymMembershipPlan{}, id).Error
}

func (r *gymMembershipPlanRepository) GetByID(ctx context.Context, id uint) (*entity.GymMembershipPlan, error) {
	var p entity.GymMembershipPlan
	if err := r.db.WithContext(ctx).Preload("Branch").First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *gymMembershipPlanRepository) List(ctx context.Context, offset, limit int, q string, activeOnly *bool, branchID *uint, highlightedOnly *bool) ([]entity.GymMembershipPlan, int64, error) {
	tx := r.db.WithContext(ctx).Model(&entity.GymMembershipPlan{})
	q = strings.TrimSpace(q)
	if q != "" {
		like := "%" + q + "%"
		tx = tx.Where("name ILIKE ? OR code ILIKE ?", like, like)
	}
	if activeOnly != nil {
		tx = tx.Where("is_active = ?", *activeOnly)
	}
	if highlightedOnly != nil {
		tx = tx.Where("is_highlighted = ?", *highlightedOnly)
	}
	if branchID != nil && *branchID > 0 {
		tx = tx.Where("branch_id = ? OR branch_id IS NULL", *branchID)
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []entity.GymMembershipPlan
	err := tx.Preload("Branch").Order("sort_order ASC, id ASC").Offset(offset).Limit(limit).Find(&rows).Error
	return rows, total, err
}

// —— UserGymMembership ——

type UserGymMembershipRepository interface {
	Create(ctx context.Context, m *entity.UserGymMembership) error
	Update(ctx context.Context, m *entity.UserGymMembership) error
	// ActivateFromPending atomically flips pending → active. Returns false if already activated (or missing).
	ActivateFromPending(ctx context.Context, m *entity.UserGymMembership) (bool, error)
	// ActivateWithRenew serializes activation per user (advisory lock) so two concurrent
	// purchases can't both become "active" at once: it looks up the current active
	// membership, extends/expires it, and flips m from pending→active in one transaction.
	// Returns claimed=false if m was already activated by a prior call (idempotent).
	ActivateWithRenew(ctx context.Context, m *entity.UserGymMembership, months int, now time.Time) (claimed bool, end time.Time, err error)
	GetByID(ctx context.Context, id uint) (*entity.UserGymMembership, error)
	GetByVnpTxnRef(ctx context.Context, txnRef string) (*entity.UserGymMembership, error)
	GetByStripeCheckoutSessionID(ctx context.Context, sessionID string) (*entity.UserGymMembership, error)
	GetActiveByUserID(ctx context.Context, userID uint, now time.Time) (*entity.UserGymMembership, error)
	ListByUserID(ctx context.Context, userID uint, limit int) ([]entity.UserGymMembership, error)
	ListAdmin(ctx context.Context, offset, limit int, status string, userID uint) ([]entity.UserGymMembership, int64, error)
	ExpireEnded(ctx context.Context, now time.Time) error
}

type userGymMembershipRepository struct{ db *gorm.DB }

func NewUserGymMembershipRepository(db *gorm.DB) UserGymMembershipRepository {
	return &userGymMembershipRepository{db: db}
}

func (r *userGymMembershipRepository) Create(ctx context.Context, m *entity.UserGymMembership) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *userGymMembershipRepository) Update(ctx context.Context, m *entity.UserGymMembership) error {
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *userGymMembershipRepository) ActivateFromPending(ctx context.Context, m *entity.UserGymMembership) (bool, error) {
	res := r.db.WithContext(ctx).Model(&entity.UserGymMembership{}).
		Where("id = ? AND status = ?", m.ID, entity.GymMemStatusPending).
		Updates(map[string]interface{}{
			"start_date":                 m.StartDate,
			"end_date":                   m.EndDate,
			"status":                     entity.GymMemStatusActive,
			"vnp_transaction_no":         m.VnpTransactionNo,
			"stripe_checkout_session_id": m.StripeCheckoutSessionID,
			"stripe_payment_intent_id":   m.StripePaymentIntentID,
			"payment_provider":           m.PaymentProvider,
		})
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected == 1, nil
}

func (r *userGymMembershipRepository) ActivateWithRenew(ctx context.Context, m *entity.UserGymMembership, months int, now time.Time) (bool, time.Time, error) {
	var claimed bool
	var end time.Time
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Serializes all activations for this user: a concurrent purchase blocks here
		// until this transaction commits, so two "pending" rows can never both become
		// "active" at once (the race the plain ActivateFromPending update misses).
		if err := tx.Exec("SELECT pg_advisory_xact_lock(?)", int64(m.UserID)).Error; err != nil {
			return err
		}

		end = now.AddDate(0, months, 0)
		var active entity.UserGymMembership
		aerr := tx.Where("user_id = ? AND status = ? AND end_date > ?", m.UserID, entity.GymMemStatusActive, now).
			Order("end_date DESC").First(&active).Error
		switch {
		case aerr == nil && active.ID != m.ID:
			if active.EndDate.After(now) {
				end = active.EndDate.AddDate(0, months, 0)
			}
			active.Status = entity.GymMemStatusExpired
			if err := tx.Save(&active).Error; err != nil {
				return err
			}
		case aerr != nil && !errors.Is(aerr, gorm.ErrRecordNotFound):
			return aerr
		}

		res := tx.Model(&entity.UserGymMembership{}).
			Where("id = ? AND status = ?", m.ID, entity.GymMemStatusPending).
			Updates(map[string]interface{}{
				"start_date":                 now,
				"end_date":                   end,
				"status":                     entity.GymMemStatusActive,
				"vnp_transaction_no":         m.VnpTransactionNo,
				"stripe_checkout_session_id": m.StripeCheckoutSessionID,
				"stripe_payment_intent_id":   m.StripePaymentIntentID,
				"payment_provider":           m.PaymentProvider,
			})
		if res.Error != nil {
			return res.Error
		}
		claimed = res.RowsAffected == 1
		return nil
	})
	return claimed, end, err
}

func (r *userGymMembershipRepository) GetByID(ctx context.Context, id uint) (*entity.UserGymMembership, error) {
	var m entity.UserGymMembership
	if err := r.db.WithContext(ctx).Preload("GymMembershipPlan").Preload("User").First(&m, id).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *userGymMembershipRepository) GetByVnpTxnRef(ctx context.Context, txnRef string) (*entity.UserGymMembership, error) {
	var m entity.UserGymMembership
	if err := r.db.WithContext(ctx).Preload("GymMembershipPlan").Where("vnp_txn_ref = ?", txnRef).First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *userGymMembershipRepository) GetByStripeCheckoutSessionID(ctx context.Context, sessionID string) (*entity.UserGymMembership, error) {
	var m entity.UserGymMembership
	if err := r.db.WithContext(ctx).Preload("GymMembershipPlan").
		Where("stripe_checkout_session_id = ?", sessionID).First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *userGymMembershipRepository) GetActiveByUserID(ctx context.Context, userID uint, now time.Time) (*entity.UserGymMembership, error) {
	var m entity.UserGymMembership
	err := r.db.WithContext(ctx).Preload("GymMembershipPlan").
		Where("user_id = ? AND status = ? AND end_date > ?", userID, entity.GymMemStatusActive, now).
		Order("end_date DESC").First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *userGymMembershipRepository) ListByUserID(ctx context.Context, userID uint, limit int) ([]entity.UserGymMembership, error) {
	if limit <= 0 {
		limit = 10
	}
	var rows []entity.UserGymMembership
	err := r.db.WithContext(ctx).Preload("GymMembershipPlan").
		Where("user_id = ?", userID).Order("id DESC").Limit(limit).Find(&rows).Error
	return rows, err
}

func (r *userGymMembershipRepository) ListAdmin(ctx context.Context, offset, limit int, status string, userID uint) ([]entity.UserGymMembership, int64, error) {
	tx := r.db.WithContext(ctx).Model(&entity.UserGymMembership{})
	if status != "" {
		tx = tx.Where("status = ?", status)
	}
	if userID > 0 {
		tx = tx.Where("user_id = ?", userID)
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []entity.UserGymMembership
	err := tx.Preload("GymMembershipPlan").Preload("User").
		Order("id DESC").Offset(offset).Limit(limit).Find(&rows).Error
	return rows, total, err
}

func (r *userGymMembershipRepository) ExpireEnded(ctx context.Context, now time.Time) error {
	return r.db.WithContext(ctx).Model(&entity.UserGymMembership{}).
		Where("status = ? AND end_date <= ?", entity.GymMemStatusActive, now).
		Update("status", entity.GymMemStatusExpired).Error
}

// —— GroupClass ——

type GroupClassRepository interface {
	Create(ctx context.Context, g *entity.GroupClass) error
	Update(ctx context.Context, g *entity.GroupClass) error
	Delete(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*entity.GroupClass, error)
	List(ctx context.Context, offset, limit int, q string, branchID *uint, activeOnly *bool) ([]entity.GroupClass, int64, error)
}

type groupClassRepository struct{ db *gorm.DB }

func NewGroupClassRepository(db *gorm.DB) GroupClassRepository {
	return &groupClassRepository{db: db}
}

func (r *groupClassRepository) Create(ctx context.Context, g *entity.GroupClass) error {
	return r.db.WithContext(ctx).Create(g).Error
}

func (r *groupClassRepository) Update(ctx context.Context, g *entity.GroupClass) error {
	return r.db.WithContext(ctx).Omit("Branch", "Trainer").Save(g).Error
}

func (r *groupClassRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.GroupClass{}, id).Error
}

func (r *groupClassRepository) GetByID(ctx context.Context, id uint) (*entity.GroupClass, error) {
	var g entity.GroupClass
	if err := r.db.WithContext(ctx).Preload("Branch").Preload("Trainer").First(&g, id).Error; err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *groupClassRepository) List(ctx context.Context, offset, limit int, q string, branchID *uint, activeOnly *bool) ([]entity.GroupClass, int64, error) {
	tx := r.db.WithContext(ctx).Model(&entity.GroupClass{})
	q = strings.TrimSpace(q)
	if q != "" {
		like := "%" + q + "%"
		tx = tx.Where("name ILIKE ? OR category ILIKE ?", like, like)
	}
	if branchID != nil && *branchID > 0 {
		tx = tx.Where("branch_id = ?", *branchID)
	}
	if activeOnly != nil {
		tx = tx.Where("is_active = ?", *activeOnly)
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []entity.GroupClass
	err := tx.Preload("Branch").Preload("Trainer").Order("id DESC").Offset(offset).Limit(limit).Find(&rows).Error
	return rows, total, err
}

// —— ClassSession ——

type ClassSessionRepository interface {
	Create(ctx context.Context, s *entity.ClassSession) error
	Update(ctx context.Context, s *entity.ClassSession) error
	Delete(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*entity.ClassSession, error)
	List(ctx context.Context, offset, limit int, groupClassID *uint, fromTime *time.Time) ([]entity.ClassSession, int64, error)
	ListUpcoming(ctx context.Context, offset, limit int, branchID *uint) ([]entity.ClassSession, int64, error)
	IncrementBooked(ctx context.Context, id uint, delta int) error
}

type classSessionRepository struct{ db *gorm.DB }

func NewClassSessionRepository(db *gorm.DB) ClassSessionRepository {
	return &classSessionRepository{db: db}
}

func classSessionPreload(db *gorm.DB) *gorm.DB {
	return db.Preload("GroupClass").Preload("GroupClass.Branch")
}

func (r *classSessionRepository) Create(ctx context.Context, s *entity.ClassSession) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *classSessionRepository) Update(ctx context.Context, s *entity.ClassSession) error {
	return r.db.WithContext(ctx).Omit("GroupClass").Save(s).Error
}

func (r *classSessionRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.ClassSession{}, id).Error
}

func (r *classSessionRepository) GetByID(ctx context.Context, id uint) (*entity.ClassSession, error) {
	var s entity.ClassSession
	if err := classSessionPreload(r.db.WithContext(ctx)).First(&s, id).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *classSessionRepository) List(ctx context.Context, offset, limit int, groupClassID *uint, fromTime *time.Time) ([]entity.ClassSession, int64, error) {
	tx := r.db.WithContext(ctx).Model(&entity.ClassSession{})
	if groupClassID != nil && *groupClassID > 0 {
		tx = tx.Where("group_class_id = ?", *groupClassID)
	}
	if fromTime != nil {
		tx = tx.Where("starts_at >= ?", *fromTime)
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []entity.ClassSession
	err := classSessionPreload(tx).Order("starts_at ASC").Offset(offset).Limit(limit).Find(&rows).Error
	return rows, total, err
}

func (r *classSessionRepository) ListUpcoming(ctx context.Context, offset, limit int, branchID *uint) ([]entity.ClassSession, int64, error) {
	tx := r.db.WithContext(ctx).Model(&entity.ClassSession{}).
		Where("starts_at >= ? AND is_canceled = ?", time.Now().UTC(), false)
	if branchID != nil && *branchID > 0 {
		tx = tx.Joins("JOIN group_classes ON group_classes.id = class_sessions.group_class_id").
			Where("group_classes.branch_id = ?", *branchID)
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []entity.ClassSession
	err := classSessionPreload(tx).Order("starts_at ASC").Offset(offset).Limit(limit).Find(&rows).Error
	return rows, total, err
}

func (r *classSessionRepository) IncrementBooked(ctx context.Context, id uint, delta int) error {
	return r.db.WithContext(ctx).Model(&entity.ClassSession{}).Where("id = ?", id).
		Update("booked_count", gorm.Expr("booked_count + ?", delta)).Error
}

// —— ClassBooking ——

type ClassBookingRepository interface {
	Create(ctx context.Context, b *entity.ClassBooking) error
	Update(ctx context.Context, b *entity.ClassBooking) error
	GetByID(ctx context.Context, id uint) (*entity.ClassBooking, error)
	GetActiveByUserAndSession(ctx context.Context, userID, sessionID uint) (*entity.ClassBooking, error)
	ListByUserID(ctx context.Context, userID uint, limit int) ([]entity.ClassBooking, error)
}

type classBookingRepository struct{ db *gorm.DB }

func NewClassBookingRepository(db *gorm.DB) ClassBookingRepository {
	return &classBookingRepository{db: db}
}

func (r *classBookingRepository) Create(ctx context.Context, b *entity.ClassBooking) error {
	return r.db.WithContext(ctx).Create(b).Error
}

func (r *classBookingRepository) Update(ctx context.Context, b *entity.ClassBooking) error {
	return r.db.WithContext(ctx).Save(b).Error
}

func (r *classBookingRepository) GetByID(ctx context.Context, id uint) (*entity.ClassBooking, error) {
	var b entity.ClassBooking
	if err := r.db.WithContext(ctx).Preload("ClassSession").Preload("ClassSession.GroupClass").First(&b, id).Error; err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *classBookingRepository) GetActiveByUserAndSession(ctx context.Context, userID, sessionID uint) (*entity.ClassBooking, error) {
	var b entity.ClassBooking
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND class_session_id = ? AND status = ?", userID, sessionID, entity.ClassBookingBooked).
		First(&b).Error
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *classBookingRepository) ListByUserID(ctx context.Context, userID uint, limit int) ([]entity.ClassBooking, error) {
	if limit <= 0 {
		limit = 50
	}
	var rows []entity.ClassBooking
	err := r.db.WithContext(ctx).Preload("ClassSession").Preload("ClassSession.GroupClass").
		Where("user_id = ?", userID).Order("id DESC").Limit(limit).Find(&rows).Error
	return rows, err
}

// —— PTPackage ——

type PTPackageRepository interface {
	Create(ctx context.Context, p *entity.PTPackage) error
	Update(ctx context.Context, p *entity.PTPackage) error
	Delete(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*entity.PTPackage, error)
	ListByTrainer(ctx context.Context, offset, limit int, trainerProfileID uint, activeOnly *bool) ([]entity.PTPackage, int64, error)
	ListPublicByTrainer(ctx context.Context, offset, limit int, trainerProfileID uint) ([]entity.PTPackage, int64, error)
	ListAdmin(ctx context.Context, offset, limit int, q string, trainerProfileID uint) ([]entity.PTPackage, int64, error)
}

type ptPackageRepository struct{ db *gorm.DB }

func NewPTPackageRepository(db *gorm.DB) PTPackageRepository {
	return &ptPackageRepository{db: db}
}

func ptPackagePreload(db *gorm.DB) *gorm.DB {
	return db.Preload("Trainer")
}

func (r *ptPackageRepository) Create(ctx context.Context, p *entity.PTPackage) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *ptPackageRepository) Update(ctx context.Context, p *entity.PTPackage) error {
	return r.db.WithContext(ctx).Omit("Trainer").Save(p).Error
}

func (r *ptPackageRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.PTPackage{}, id).Error
}

func (r *ptPackageRepository) GetByID(ctx context.Context, id uint) (*entity.PTPackage, error) {
	var p entity.PTPackage
	if err := ptPackagePreload(r.db.WithContext(ctx)).First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ptPackageRepository) ListByTrainer(ctx context.Context, offset, limit int, trainerProfileID uint, activeOnly *bool) ([]entity.PTPackage, int64, error) {
	tx := r.db.WithContext(ctx).Model(&entity.PTPackage{}).Where("trainer_profile_id = ?", trainerProfileID)
	if activeOnly != nil {
		tx = tx.Where("is_active = ?", *activeOnly)
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []entity.PTPackage
	err := ptPackagePreload(tx).Order("id DESC").Offset(offset).Limit(limit).Find(&rows).Error
	return rows, total, err
}

func (r *ptPackageRepository) ListPublicByTrainer(ctx context.Context, offset, limit int, trainerProfileID uint) ([]entity.PTPackage, int64, error) {
	tx := r.db.WithContext(ctx).Model(&entity.PTPackage{}).
		Where("trainer_profile_id = ? AND is_public = ? AND is_active = ?", trainerProfileID, true, true)
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []entity.PTPackage
	err := ptPackagePreload(tx).Order("id DESC").Offset(offset).Limit(limit).Find(&rows).Error
	return rows, total, err
}

func (r *ptPackageRepository) ListAdmin(ctx context.Context, offset, limit int, q string, trainerProfileID uint) ([]entity.PTPackage, int64, error) {
	tx := r.db.WithContext(ctx).Model(&entity.PTPackage{})
	q = strings.TrimSpace(q)
	if q != "" {
		like := "%" + q + "%"
		tx = tx.Where("title ILIKE ?", like)
	}
	if trainerProfileID > 0 {
		tx = tx.Where("trainer_profile_id = ?", trainerProfileID)
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []entity.PTPackage
	err := ptPackagePreload(tx).Order("id DESC").Offset(offset).Limit(limit).Find(&rows).Error
	return rows, total, err
}

// —— UserPTPackage ——

type UserPTPackageRepository interface {
	Create(ctx context.Context, p *entity.UserPTPackage) error
	Update(ctx context.Context, p *entity.UserPTPackage) error
	ActivateFromPending(ctx context.Context, p *entity.UserPTPackage) (bool, error)
	GetByID(ctx context.Context, id uint) (*entity.UserPTPackage, error)
	GetByVnpTxnRef(ctx context.Context, txnRef string) (*entity.UserPTPackage, error)
	GetByStripeCheckoutSessionID(ctx context.Context, sessionID string) (*entity.UserPTPackage, error)
	ListByUserID(ctx context.Context, userID uint, limit int) ([]entity.UserPTPackage, error)
	ListByTrainerProfileID(ctx context.Context, trainerProfileID uint, offset, limit int, status string) ([]entity.UserPTPackage, int64, error)
	ListAdmin(ctx context.Context, offset, limit int, status string, trainerProfileID, userID uint) ([]entity.UserPTPackage, int64, error)
	CountActiveClients(ctx context.Context, trainerProfileID uint) (int, error)
	HasActivePackage(ctx context.Context, trainerProfileID, userID uint) (bool, error)
	CountClientsEver(ctx context.Context, trainerProfileID uint) (int, error)
	ExpireEnded(ctx context.Context, now time.Time) error
}

type userPTPackageRepository struct{ db *gorm.DB }

func NewUserPTPackageRepository(db *gorm.DB) UserPTPackageRepository {
	return &userPTPackageRepository{db: db}
}

func userPTPackagePreload(db *gorm.DB) *gorm.DB {
	return db.Preload("User").Preload("PTPackage").Preload("PTPackage.Trainer")
}

func (r *userPTPackageRepository) Create(ctx context.Context, p *entity.UserPTPackage) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *userPTPackageRepository) Update(ctx context.Context, p *entity.UserPTPackage) error {
	return r.db.WithContext(ctx).Save(p).Error
}

func (r *userPTPackageRepository) ActivateFromPending(ctx context.Context, p *entity.UserPTPackage) (bool, error) {
	res := r.db.WithContext(ctx).Model(&entity.UserPTPackage{}).
		Where("id = ? AND status = ?", p.ID, entity.PTPkgStatusPending).
		Updates(map[string]interface{}{
			"starts_at":                  p.StartsAt,
			"expires_at":                 p.ExpiresAt,
			"status":                     entity.PTPkgStatusActive,
			"vnp_transaction_no":         p.VnpTransactionNo,
			"stripe_checkout_session_id": p.StripeCheckoutSessionID,
			"stripe_payment_intent_id":   p.StripePaymentIntentID,
			"payment_provider":           p.PaymentProvider,
		})
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected == 1, nil
}

func (r *userPTPackageRepository) GetByID(ctx context.Context, id uint) (*entity.UserPTPackage, error) {
	var p entity.UserPTPackage
	if err := userPTPackagePreload(r.db.WithContext(ctx)).First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *userPTPackageRepository) GetByVnpTxnRef(ctx context.Context, txnRef string) (*entity.UserPTPackage, error) {
	var p entity.UserPTPackage
	if err := userPTPackagePreload(r.db.WithContext(ctx)).Where("vnp_txn_ref = ?", txnRef).First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *userPTPackageRepository) GetByStripeCheckoutSessionID(ctx context.Context, sessionID string) (*entity.UserPTPackage, error) {
	var p entity.UserPTPackage
	if err := userPTPackagePreload(r.db.WithContext(ctx)).
		Where("stripe_checkout_session_id = ?", sessionID).First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *userPTPackageRepository) ListByUserID(ctx context.Context, userID uint, limit int) ([]entity.UserPTPackage, error) {
	if limit <= 0 {
		limit = 50
	}
	var rows []entity.UserPTPackage
	err := userPTPackagePreload(r.db.WithContext(ctx)).
		Where("user_id = ?", userID).Order("id DESC").Limit(limit).Find(&rows).Error
	return rows, err
}

func (r *userPTPackageRepository) ListByTrainerProfileID(ctx context.Context, trainerProfileID uint, offset, limit int, status string) ([]entity.UserPTPackage, int64, error) {
	base := r.db.WithContext(ctx).Model(&entity.UserPTPackage{}).Where("trainer_profile_id = ?", trainerProfileID)
	if status != "" {
		base = base.Where("status = ?", status)
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	q := r.db.WithContext(ctx).Where("trainer_profile_id = ?", trainerProfileID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var rows []entity.UserPTPackage
	err := userPTPackagePreload(q).Order("id DESC").Offset(offset).Limit(limit).Find(&rows).Error
	return rows, total, err
}

func (r *userPTPackageRepository) CountActiveClients(ctx context.Context, trainerProfileID uint) (int, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&entity.UserPTPackage{}).
		Where("trainer_profile_id = ? AND status = ?", trainerProfileID, entity.PTPkgStatusActive).
		Select("COUNT(DISTINCT user_id)").Scan(&n).Error
	return int(n), err
}

func (r *userPTPackageRepository) HasActivePackage(ctx context.Context, trainerProfileID, userID uint) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&entity.UserPTPackage{}).
		Where("trainer_profile_id = ? AND user_id = ? AND status = ?", trainerProfileID, userID, entity.PTPkgStatusActive).
		Limit(1).Count(&n).Error
	return n > 0, err
}

func (r *userPTPackageRepository) CountClientsEver(ctx context.Context, trainerProfileID uint) (int, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&entity.UserPTPackage{}).
		Where("trainer_profile_id = ? AND status IN ?", trainerProfileID, []string{
			entity.PTPkgStatusActive, entity.PTPkgStatusExpired, entity.PTPkgStatusCanceled,
		}).
		Select("COUNT(DISTINCT user_id)").Scan(&n).Error
	return int(n), err
}

func (r *userPTPackageRepository) ListAdmin(ctx context.Context, offset, limit int, status string, trainerProfileID, userID uint) ([]entity.UserPTPackage, int64, error) {
	base := r.db.WithContext(ctx).Model(&entity.UserPTPackage{})
	if status != "" {
		base = base.Where("status = ?", status)
	}
	if trainerProfileID > 0 {
		base = base.Where("trainer_profile_id = ?", trainerProfileID)
	}
	if userID > 0 {
		base = base.Where("user_id = ?", userID)
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	q := r.db.WithContext(ctx).Model(&entity.UserPTPackage{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if trainerProfileID > 0 {
		q = q.Where("trainer_profile_id = ?", trainerProfileID)
	}
	if userID > 0 {
		q = q.Where("user_id = ?", userID)
	}
	var rows []entity.UserPTPackage
	err := userPTPackagePreload(q).Order("id DESC").Offset(offset).Limit(limit).Find(&rows).Error
	return rows, total, err
}

func (r *userPTPackageRepository) ExpireEnded(ctx context.Context, now time.Time) error {
	return r.db.WithContext(ctx).Model(&entity.UserPTPackage{}).
		Where("status = ? AND expires_at <= ?", entity.PTPkgStatusActive, now).
		Update("status", entity.PTPkgStatusExpired).Error
}

// —— PTSessionLog ——

type PTSessionLogRepository interface {
	Create(ctx context.Context, log *entity.PTSessionLog) error
	ListByUserPTPackageID(ctx context.Context, userPTPackageID uint) ([]entity.PTSessionLog, error)
	CountByUserPTPackageID(ctx context.Context, userPTPackageID uint) (int64, error)
}

type ptSessionLogRepository struct{ db *gorm.DB }

func NewPTSessionLogRepository(db *gorm.DB) PTSessionLogRepository {
	return &ptSessionLogRepository{db: db}
}

func (r *ptSessionLogRepository) Create(ctx context.Context, log *entity.PTSessionLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *ptSessionLogRepository) ListByUserPTPackageID(ctx context.Context, userPTPackageID uint) ([]entity.PTSessionLog, error) {
	var rows []entity.PTSessionLog
	err := r.db.WithContext(ctx).Where("user_pt_package_id = ?", userPTPackageID).
		Order("session_index ASC, id ASC").Find(&rows).Error
	return rows, err
}

func (r *ptSessionLogRepository) CountByUserPTPackageID(ctx context.Context, userPTPackageID uint) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&entity.PTSessionLog{}).
		Where("user_pt_package_id = ?", userPTPackageID).Count(&n).Error
	return n, err
}

// —— PTSessionOffer ——

type PTSessionOfferRepository interface {
	Create(ctx context.Context, o *entity.PTSessionOffer) error
	Update(ctx context.Context, o *entity.PTSessionOffer) error
	GetByID(ctx context.Context, id uint) (*entity.PTSessionOffer, error)
	ListByUserPTPackageID(ctx context.Context, userPTPackageID uint) ([]entity.PTSessionOffer, error)
	GetByIDs(ctx context.Context, ids []uint) ([]entity.PTSessionOffer, error)
	CountOpenByPackage(ctx context.Context, userPTPackageID uint) (int64, error)
	CountPendingByPackages(ctx context.Context, packageIDs []uint) (map[uint]int64, error)
	ListBusyInRange(ctx context.Context, trainerProfileID uint, from, to time.Time) ([]entity.PTSessionOffer, error)
	ListBusyInRangeForStudent(ctx context.Context, studentUserID uint, from, to time.Time) ([]entity.PTSessionOffer, error)
	ListByTrainerInRange(ctx context.Context, trainerProfileID uint, from, to time.Time) ([]entity.PTSessionOffer, error)
	CountStatusesByTrainer(ctx context.Context, trainerProfileID uint) (map[string]int64, error)
	ListAwaitingConfirmationOlderThan(ctx context.Context, before time.Time, limit int) ([]entity.PTSessionOffer, error)
	// ExpireStalePending cancels "pending" offers proposed before `before` and never
	// responded to — without this, a forgotten proposal keeps occupying the trainer's
	// slot / the package's session-credit forever.
	ExpireStalePending(ctx context.Context, before time.Time) (int64, error)
}

type ptSessionOfferRepository struct{ db *gorm.DB }

func NewPTSessionOfferRepository(db *gorm.DB) PTSessionOfferRepository {
	return &ptSessionOfferRepository{db: db}
}

func (r *ptSessionOfferRepository) Create(ctx context.Context, o *entity.PTSessionOffer) error {
	return r.db.WithContext(ctx).Create(o).Error
}

func (r *ptSessionOfferRepository) Update(ctx context.Context, o *entity.PTSessionOffer) error {
	return r.db.WithContext(ctx).Save(o).Error
}

func (r *ptSessionOfferRepository) GetByID(ctx context.Context, id uint) (*entity.PTSessionOffer, error) {
	var o entity.PTSessionOffer
	if err := r.db.WithContext(ctx).First(&o, id).Error; err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *ptSessionOfferRepository) ListByUserPTPackageID(ctx context.Context, userPTPackageID uint) ([]entity.PTSessionOffer, error) {
	var rows []entity.PTSessionOffer
	err := r.db.WithContext(ctx).Where("user_pt_package_id = ?", userPTPackageID).
		Order("starts_at ASC, id ASC").Find(&rows).Error
	return rows, err
}

func (r *ptSessionOfferRepository) GetByIDs(ctx context.Context, ids []uint) ([]entity.PTSessionOffer, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var rows []entity.PTSessionOffer
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&rows).Error
	return rows, err
}

func (r *ptSessionOfferRepository) CountOpenByPackage(ctx context.Context, userPTPackageID uint) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&entity.PTSessionOffer{}).
		Where("user_pt_package_id = ? AND status IN ?", userPTPackageID, []string{
			entity.SessionOfferPending, entity.SessionOfferScheduled, entity.SessionOfferAwaitingConfirmation,
		}).Count(&n).Error
	return n, err
}

func (r *ptSessionOfferRepository) CountPendingByPackages(ctx context.Context, packageIDs []uint) (map[uint]int64, error) {
	out := map[uint]int64{}
	if len(packageIDs) == 0 {
		return out, nil
	}
	type row struct {
		UserPTPackageID uint  `gorm:"column:user_pt_package_id"`
		Cnt             int64 `gorm:"column:cnt"`
	}
	var rows []row
	err := r.db.WithContext(ctx).Model(&entity.PTSessionOffer{}).
		Select("user_pt_package_id, COUNT(*) as cnt").
		Where("user_pt_package_id IN ? AND status IN ?", packageIDs, []string{
			entity.SessionOfferPending, entity.SessionOfferAwaitingConfirmation,
		}).
		Group("user_pt_package_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		out[row.UserPTPackageID] = row.Cnt
	}
	return out, nil
}

func (r *ptSessionOfferRepository) ListBusyInRange(ctx context.Context, trainerProfileID uint, from, to time.Time) ([]entity.PTSessionOffer, error) {
	// Fetch a padded window; callers compute ends_at when nil (session duration).
	padFrom := from.Add(-3 * time.Hour)
	var rows []entity.PTSessionOffer
	err := r.db.WithContext(ctx).
		Where("trainer_profile_id = ? AND status IN ? AND starts_at >= ? AND starts_at < ?",
			trainerProfileID,
			[]string{entity.SessionOfferPending, entity.SessionOfferScheduled, entity.SessionOfferAwaitingConfirmation},
			padFrom, to,
		).
		Order("starts_at ASC").Find(&rows).Error
	return rows, err
}

// ListBusyInRangeForStudent finds a student's open offers in a time window across
// ALL trainers/packages — ListBusyInRange only guards one trainer's calendar, so a
// student could otherwise double-book two different trainers at the same time.
func (r *ptSessionOfferRepository) ListBusyInRangeForStudent(ctx context.Context, studentUserID uint, from, to time.Time) ([]entity.PTSessionOffer, error) {
	padFrom := from.Add(-3 * time.Hour)
	var rows []entity.PTSessionOffer
	err := r.db.WithContext(ctx).
		Where("student_user_id = ? AND status IN ? AND starts_at >= ? AND starts_at < ?",
			studentUserID,
			[]string{entity.SessionOfferPending, entity.SessionOfferScheduled, entity.SessionOfferAwaitingConfirmation},
			padFrom, to,
		).
		Order("starts_at ASC").Find(&rows).Error
	return rows, err
}

func (r *ptSessionOfferRepository) ListByTrainerInRange(ctx context.Context, trainerProfileID uint, from, to time.Time) ([]entity.PTSessionOffer, error) {
	var rows []entity.PTSessionOffer
	err := r.db.WithContext(ctx).
		Where("trainer_profile_id = ? AND starts_at >= ? AND starts_at < ?", trainerProfileID, from, to).
		Order("starts_at ASC").Find(&rows).Error
	return rows, err
}

func (r *ptSessionOfferRepository) CountStatusesByTrainer(ctx context.Context, trainerProfileID uint) (map[string]int64, error) {
	type row struct {
		Status string `gorm:"column:status"`
		Cnt    int64  `gorm:"column:cnt"`
	}
	var rows []row
	err := r.db.WithContext(ctx).Model(&entity.PTSessionOffer{}).
		Select("status, COUNT(*) as cnt").
		Where("trainer_profile_id = ?", trainerProfileID).
		Group("status").Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := map[string]int64{}
	for _, r := range rows {
		out[r.Status] = r.Cnt
	}
	return out, nil
}

func (r *ptSessionOfferRepository) ListAwaitingConfirmationOlderThan(ctx context.Context, before time.Time, limit int) ([]entity.PTSessionOffer, error) {
	if limit <= 0 {
		limit = 50
	}
	var rows []entity.PTSessionOffer
	err := r.db.WithContext(ctx).
		Where("status = ? AND proof_submitted_at IS NOT NULL AND proof_submitted_at <= ?",
			entity.SessionOfferAwaitingConfirmation, before).
		Order("proof_submitted_at ASC").
		Limit(limit).
		Find(&rows).Error
	return rows, err
}

func (r *ptSessionOfferRepository) ExpireStalePending(ctx context.Context, before time.Time) (int64, error) {
	res := r.db.WithContext(ctx).Model(&entity.PTSessionOffer{}).
		Where("status = ? AND created_at <= ?", entity.SessionOfferPending, before).
		Update("status", entity.SessionOfferCancelled)
	return res.RowsAffected, res.Error
}

// —— RevenueShareSetting ——

type RevenueShareSettingRepository interface {
	GetSingleton(ctx context.Context) (*entity.RevenueShareSetting, error)
	Update(ctx context.Context, s *entity.RevenueShareSetting) error
}

type revenueShareSettingRepository struct{ db *gorm.DB }

func NewRevenueShareSettingRepository(db *gorm.DB) RevenueShareSettingRepository {
	return &revenueShareSettingRepository{db: db}
}

func (r *revenueShareSettingRepository) GetSingleton(ctx context.Context) (*entity.RevenueShareSetting, error) {
	var s entity.RevenueShareSetting
	if err := r.db.WithContext(ctx).First(&s, 1).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			s = entity.RevenueShareSetting{PTPercent: 40, GymPercent: 60}
			s.ID = 1
			if err := r.db.WithContext(ctx).Create(&s).Error; err != nil {
				return nil, err
			}
			return &s, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *revenueShareSettingRepository) Update(ctx context.Context, s *entity.RevenueShareSetting) error {
	return r.db.WithContext(ctx).Save(s).Error
}

// —— PTEarning ——

type PTEarningRepository interface {
	Create(ctx context.Context, e *entity.PTEarning) error
	ListAdmin(ctx context.Context, offset, limit int, trainerProfileID uint) ([]entity.PTEarning, int64, error)
	SumPTAmount(ctx context.Context, trainerProfileID uint) (float64, error)
	GetByID(ctx context.Context, id uint) (*entity.PTEarning, error)
	// MarkPaidOut lets admin/CSKH record that a PT has been paid out for this
	// earning row (e.g. weekly/monthly payroll run) — the ledger otherwise has
	// no notion of paid vs. still-owed.
	MarkPaidOut(ctx context.Context, id uint, paid bool) error
}

type ptEarningRepository struct{ db *gorm.DB }

func NewPTEarningRepository(db *gorm.DB) PTEarningRepository {
	return &ptEarningRepository{db: db}
}

func (r *ptEarningRepository) Create(ctx context.Context, e *entity.PTEarning) error {
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *ptEarningRepository) ListAdmin(ctx context.Context, offset, limit int, trainerProfileID uint) ([]entity.PTEarning, int64, error) {
	tx := r.db.WithContext(ctx).Model(&entity.PTEarning{})
	if trainerProfileID > 0 {
		tx = tx.Where("trainer_profile_id = ?", trainerProfileID)
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []entity.PTEarning
	err := tx.Preload("Trainer").
		Preload("UserPTPackage").
		Preload("UserPTPackage.User").
		Preload("UserPTPackage.PTPackage").
		Order("id DESC").Offset(offset).Limit(limit).Find(&rows).Error
	return rows, total, err
}

func (r *ptEarningRepository) GetByID(ctx context.Context, id uint) (*entity.PTEarning, error) {
	var e entity.PTEarning
	if err := r.db.WithContext(ctx).First(&e, id).Error; err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *ptEarningRepository) MarkPaidOut(ctx context.Context, id uint, paid bool) error {
	return r.db.WithContext(ctx).Model(&entity.PTEarning{}).Where("id = ?", id).Update("paid_out", paid).Error
}

func (r *ptEarningRepository) SumPTAmount(ctx context.Context, trainerProfileID uint) (float64, error) {
	tx := r.db.WithContext(ctx).Model(&entity.PTEarning{})
	if trainerProfileID > 0 {
		tx = tx.Where("trainer_profile_id = ?", trainerProfileID)
	}
	var sum float64
	row := tx.Select("COALESCE(SUM(pt_amount), 0)").Row()
	if row == nil {
		return 0, nil
	}
	if err := row.Scan(&sum); err != nil {
		return 0, err
	}
	return sum, nil
}

// —— PTPackageChatMessage ——

type PTPackageChatRepository interface {
	Create(ctx context.Context, m *entity.PTPackageChatMessage) error
	ListByPackage(ctx context.Context, userPTPackageID uint, afterID uint, limit int) ([]entity.PTPackageChatMessage, error)
	CountByPackage(ctx context.Context, userPTPackageID uint) (int64, error)
	GetLatestByPackages(ctx context.Context, packageIDs []uint) (map[uint]entity.PTPackageChatMessage, error)
	CountUnreadByPackages(ctx context.Context, userID uint, packageIDs []uint) (map[uint]int64, error)
	UpsertRead(ctx context.Context, userID, userPTPackageID, lastReadMessageID uint) error
	GetMaxMessageID(ctx context.Context, userPTPackageID uint) (uint, error)
}

type ptPackageChatRepository struct{ db *gorm.DB }

func NewPTPackageChatRepository(db *gorm.DB) PTPackageChatRepository {
	return &ptPackageChatRepository{db: db}
}

func (r *ptPackageChatRepository) Create(ctx context.Context, m *entity.PTPackageChatMessage) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *ptPackageChatRepository) ListByPackage(ctx context.Context, userPTPackageID uint, afterID uint, limit int) ([]entity.PTPackageChatMessage, error) {
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	q := r.db.WithContext(ctx).Where("user_pt_package_id = ?", userPTPackageID)
	if afterID > 0 {
		q = q.Where("id > ?", afterID)
	}
	var rows []entity.PTPackageChatMessage
	err := q.Order("id ASC").Limit(limit).Find(&rows).Error
	return rows, err
}

func (r *ptPackageChatRepository) CountByPackage(ctx context.Context, userPTPackageID uint) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&entity.PTPackageChatMessage{}).
		Where("user_pt_package_id = ?", userPTPackageID).Count(&n).Error
	return n, err
}

func (r *ptPackageChatRepository) GetLatestByPackages(ctx context.Context, packageIDs []uint) (map[uint]entity.PTPackageChatMessage, error) {
	out := map[uint]entity.PTPackageChatMessage{}
	if len(packageIDs) == 0 {
		return out, nil
	}
	var rows []entity.PTPackageChatMessage
	err := r.db.WithContext(ctx).Raw(`
		SELECT DISTINCT ON (user_pt_package_id) *
		FROM pt_package_chat_messages
		WHERE user_pt_package_id IN ?
		ORDER BY user_pt_package_id, id DESC
	`, packageIDs).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for i := range rows {
		out[rows[i].UserPTPackageID] = rows[i]
	}
	return out, nil
}

func (r *ptPackageChatRepository) CountUnreadByPackages(ctx context.Context, userID uint, packageIDs []uint) (map[uint]int64, error) {
	out := map[uint]int64{}
	if len(packageIDs) == 0 {
		return out, nil
	}
	type row struct {
		UserPTPackageID uint  `gorm:"column:user_pt_package_id"`
		Cnt             int64 `gorm:"column:cnt"`
	}
	var rows []row
	err := r.db.WithContext(ctx).Raw(`
		SELECT m.user_pt_package_id, COUNT(*) AS cnt
		FROM pt_package_chat_messages m
		LEFT JOIN pt_package_chat_reads r
			ON r.user_pt_package_id = m.user_pt_package_id AND r.user_id = ?
		WHERE m.user_pt_package_id IN ?
			AND m.sender_user_id <> ?
			AND m.id > COALESCE(r.last_read_message_id, 0)
		GROUP BY m.user_pt_package_id
	`, userID, packageIDs, userID).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		out[row.UserPTPackageID] = row.Cnt
	}
	return out, nil
}

func (r *ptPackageChatRepository) UpsertRead(ctx context.Context, userID, userPTPackageID, lastReadMessageID uint) error {
	var existing entity.PTPackageChatRead
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND user_pt_package_id = ?", userID, userPTPackageID).
		First(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return r.db.WithContext(ctx).Create(&entity.PTPackageChatRead{
				UserID:            userID,
				UserPTPackageID:   userPTPackageID,
				LastReadMessageID: lastReadMessageID,
			}).Error
		}
		return err
	}
	if lastReadMessageID <= existing.LastReadMessageID {
		return nil
	}
	existing.LastReadMessageID = lastReadMessageID
	return r.db.WithContext(ctx).Save(&existing).Error
}

func (r *ptPackageChatRepository) GetMaxMessageID(ctx context.Context, userPTPackageID uint) (uint, error) {
	var id *uint
	err := r.db.WithContext(ctx).Model(&entity.PTPackageChatMessage{}).
		Select("MAX(id)").
		Where("user_pt_package_id = ?", userPTPackageID).
		Scan(&id).Error
	if err != nil {
		return 0, err
	}
	if id == nil {
		return 0, nil
	}
	return *id, nil
}
