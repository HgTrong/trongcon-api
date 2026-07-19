package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"trongcon-api/internal/entity"

	"gorm.io/gorm"
)

type SubscriptionPlanRepository interface {
	Create(ctx context.Context, p *entity.SubscriptionPlan) error
	Update(ctx context.Context, p *entity.SubscriptionPlan) error
	Delete(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*entity.SubscriptionPlan, error)
	GetByCode(ctx context.Context, code string) (*entity.SubscriptionPlan, error)
	List(ctx context.Context, offset, limit int, q, kind, orderBy, orderDir string, activeOnly *bool) ([]entity.SubscriptionPlan, int64, error)
}

type subscriptionPlanRepository struct{ db *gorm.DB }

func NewSubscriptionPlanRepository(db *gorm.DB) SubscriptionPlanRepository {
	return &subscriptionPlanRepository{db: db}
}

func (r *subscriptionPlanRepository) Create(ctx context.Context, p *entity.SubscriptionPlan) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *subscriptionPlanRepository) Update(ctx context.Context, p *entity.SubscriptionPlan) error {
	return r.db.WithContext(ctx).Save(p).Error
}

func (r *subscriptionPlanRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.SubscriptionPlan{}, id).Error
}

func (r *subscriptionPlanRepository) GetByID(ctx context.Context, id uint) (*entity.SubscriptionPlan, error) {
	var p entity.SubscriptionPlan
	if err := r.db.WithContext(ctx).First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *subscriptionPlanRepository) GetByCode(ctx context.Context, code string) (*entity.SubscriptionPlan, error) {
	var p entity.SubscriptionPlan
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *subscriptionPlanRepository) List(ctx context.Context, offset, limit int, q, kind, orderBy, orderDir string, activeOnly *bool) ([]entity.SubscriptionPlan, int64, error) {
	q = strings.TrimSpace(q)
	tx := r.db.WithContext(ctx).Model(&entity.SubscriptionPlan{})
	if q != "" {
		like := "%" + q + "%"
		tx = tx.Where("plan_name ILIKE ? OR title ILIKE ? OR code ILIKE ?", like, like, like)
	}
	if kind != "" {
		tx = tx.Where("kind = ?", kind)
	}
	if activeOnly != nil {
		tx = tx.Where("is_active = ?", *activeOnly)
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	orderBy = sanitizeOrder(orderBy, "sort_order")
	orderDir = sanitizeDir(orderDir)
	var rows []entity.SubscriptionPlan
	err := tx.Order(orderBy + " " + orderDir).Offset(offset).Limit(limit).Find(&rows).Error
	return rows, total, err
}

type UserSubscriptionRepository interface {
	Create(ctx context.Context, s *entity.UserSubscription) error
	Update(ctx context.Context, s *entity.UserSubscription) error
	GetByID(ctx context.Context, id uint) (*entity.UserSubscription, error)
	GetByPayPalOrderID(ctx context.Context, orderID string) (*entity.UserSubscription, error)
	GetByStripeCheckoutSessionID(ctx context.Context, sessionID string) (*entity.UserSubscription, error)
	GetActiveByUserID(ctx context.Context, userID uint, now time.Time) (*entity.UserSubscription, error)
	ListByUserID(ctx context.Context, userID uint, limit int) ([]entity.UserSubscription, error)
	ListAdmin(ctx context.Context, offset, limit int, status string, userID uint, orderBy, orderDir string) ([]entity.UserSubscription, int64, error)
	ExpireEnded(ctx context.Context, now time.Time) error
	HasActive(ctx context.Context, userID uint, now time.Time) (bool, error)
}

type userSubscriptionRepository struct{ db *gorm.DB }

func NewUserSubscriptionRepository(db *gorm.DB) UserSubscriptionRepository {
	return &userSubscriptionRepository{db: db}
}

func (r *userSubscriptionRepository) Create(ctx context.Context, s *entity.UserSubscription) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *userSubscriptionRepository) Update(ctx context.Context, s *entity.UserSubscription) error {
	return r.db.WithContext(ctx).Save(s).Error
}

func (r *userSubscriptionRepository) GetByID(ctx context.Context, id uint) (*entity.UserSubscription, error) {
	var s entity.UserSubscription
	if err := r.db.WithContext(ctx).Preload("SubscriptionPlan").First(&s, id).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *userSubscriptionRepository) GetByPayPalOrderID(ctx context.Context, orderID string) (*entity.UserSubscription, error) {
	var s entity.UserSubscription
	if err := r.db.WithContext(ctx).Preload("SubscriptionPlan").Where("paypal_order_id = ?", orderID).First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *userSubscriptionRepository) GetByStripeCheckoutSessionID(ctx context.Context, sessionID string) (*entity.UserSubscription, error) {
	var s entity.UserSubscription
	if err := r.db.WithContext(ctx).Preload("SubscriptionPlan").Where("stripe_checkout_session_id = ?", sessionID).First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *userSubscriptionRepository) GetActiveByUserID(ctx context.Context, userID uint, now time.Time) (*entity.UserSubscription, error) {
	var s entity.UserSubscription
	err := r.db.WithContext(ctx).Preload("SubscriptionPlan").
		Where("user_id = ? AND status = ? AND end_date > ?", userID, entity.SubStatusActive, now).
		Order("end_date DESC").First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *userSubscriptionRepository) ListByUserID(ctx context.Context, userID uint, limit int) ([]entity.UserSubscription, error) {
	if limit <= 0 {
		limit = 10
	}
	var rows []entity.UserSubscription
	err := r.db.WithContext(ctx).Preload("SubscriptionPlan").
		Where("user_id = ?", userID).Order("id DESC").Limit(limit).Find(&rows).Error
	return rows, err
}

func (r *userSubscriptionRepository) ListAdmin(ctx context.Context, offset, limit int, status string, userID uint, orderBy, orderDir string) ([]entity.UserSubscription, int64, error) {
	tx := r.db.WithContext(ctx).Model(&entity.UserSubscription{})
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
	orderBy = sanitizeOrder(orderBy, "id")
	orderDir = sanitizeDir(orderDir)
	var rows []entity.UserSubscription
	err := tx.Preload("SubscriptionPlan").Order(orderBy + " " + orderDir).Offset(offset).Limit(limit).Find(&rows).Error
	return rows, total, err
}

func (r *userSubscriptionRepository) ExpireEnded(ctx context.Context, now time.Time) error {
	return r.db.WithContext(ctx).Model(&entity.UserSubscription{}).
		Where("status = ? AND end_date <= ?", entity.SubStatusActive, now).
		Update("status", entity.SubStatusExpired).Error
}

func (r *userSubscriptionRepository) HasActive(ctx context.Context, userID uint, now time.Time) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&entity.UserSubscription{}).
		Where("user_id = ? AND status = ? AND end_date > ?", userID, entity.SubStatusActive, now).
		Count(&n).Error
	return n > 0, err
}

type PaymentHistoryRepository interface {
	Create(ctx context.Context, h *entity.PaymentHistory) error
	GetByPaymentIntentID(ctx context.Context, id string) (*entity.PaymentHistory, error)
	GetByTransactionID(ctx context.Context, id string) (*entity.PaymentHistory, error)
	List(ctx context.Context, offset, limit int) ([]entity.PaymentHistory, int64, error)
}

type paymentHistoryRepository struct{ db *gorm.DB }

func NewPaymentHistoryRepository(db *gorm.DB) PaymentHistoryRepository {
	return &paymentHistoryRepository{db: db}
}

func (r *paymentHistoryRepository) Create(ctx context.Context, h *entity.PaymentHistory) error {
	return r.db.WithContext(ctx).Create(h).Error
}

func (r *paymentHistoryRepository) GetByPaymentIntentID(ctx context.Context, id string) (*entity.PaymentHistory, error) {
	var h entity.PaymentHistory
	if err := r.db.WithContext(ctx).Where("payment_intent_id = ?", id).First(&h).Error; err != nil {
		return nil, err
	}
	return &h, nil
}

func (r *paymentHistoryRepository) GetByTransactionID(ctx context.Context, id string) (*entity.PaymentHistory, error) {
	var h entity.PaymentHistory
	if err := r.db.WithContext(ctx).Where("transaction_id = ?", id).First(&h).Error; err != nil {
		return nil, err
	}
	return &h, nil
}

func (r *paymentHistoryRepository) List(ctx context.Context, offset, limit int) ([]entity.PaymentHistory, int64, error) {
	var total int64
	tx := r.db.WithContext(ctx).Model(&entity.PaymentHistory{})
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []entity.PaymentHistory
	err := tx.Order("id DESC").Offset(offset).Limit(limit).Find(&rows).Error
	return rows, total, err
}

func sanitizeOrder(v, def string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "id", "created_at", "updated_at", "price", "sort_order", "plan_name", "end_date", "status":
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return def
	}
}

func sanitizeDir(v string) string {
	if strings.EqualFold(v, "ASC") {
		return "ASC"
	}
	return "DESC"
}

var ErrNotFound = errors.New("not found")
