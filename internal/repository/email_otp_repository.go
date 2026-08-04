package repository

import (
	"context"
	"strings"
	"time"

	"trongcon-api/internal/entity"

	"gorm.io/gorm"
)

type EmailOTPRepository interface {
	Create(ctx context.Context, row *entity.EmailOTP) error
	FindValid(ctx context.Context, email, purpose, otp string) (*entity.EmailOTP, error)
	MarkUsed(ctx context.Context, id uint) error
	InvalidateOpen(ctx context.Context, email, purpose string) error
}

type emailOTPRepository struct {
	db *gorm.DB
}

func NewEmailOTPRepository(db *gorm.DB) EmailOTPRepository {
	return &emailOTPRepository{db: db}
}

func (r *emailOTPRepository) Create(ctx context.Context, row *entity.EmailOTP) error {
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *emailOTPRepository) FindValid(ctx context.Context, email, purpose, otp string) (*entity.EmailOTP, error) {
	var row entity.EmailOTP
	err := r.db.WithContext(ctx).
		Where("email = ? AND purpose = ? AND otp = ? AND used = ? AND expire_at > ?",
			strings.ToLower(strings.TrimSpace(email)),
			purpose,
			strings.TrimSpace(otp),
			false,
			time.Now().UTC(),
		).
		Order("id DESC").
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *emailOTPRepository) MarkUsed(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Model(&entity.EmailOTP{}).
		Where("id = ?", id).
		Update("used", true).Error
}

func (r *emailOTPRepository) InvalidateOpen(ctx context.Context, email, purpose string) error {
	return r.db.WithContext(ctx).Model(&entity.EmailOTP{}).
		Where("email = ? AND purpose = ? AND used = ?", strings.ToLower(strings.TrimSpace(email)), purpose, false).
		Update("used", true).Error
}
