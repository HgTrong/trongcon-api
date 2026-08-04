package repository

import (
	"context"
	"strings"

	"trongcon-api/internal/entity"

	"gorm.io/gorm"
)

type FAQRepository interface {
	Create(ctx context.Context, f *entity.FAQ) error
	GetByID(ctx context.Context, id uint) (*entity.FAQ, error)
	Update(ctx context.Context, f *entity.FAQ) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, offset, limit int, q string, active *bool, order string) ([]entity.FAQ, int64, error)
}

type faqRepository struct {
	db *gorm.DB
}

func NewFAQRepository(db *gorm.DB) FAQRepository {
	return &faqRepository{db: db}
}

func (r *faqRepository) Create(ctx context.Context, f *entity.FAQ) error {
	return r.db.WithContext(ctx).Create(f).Error
}

func (r *faqRepository) GetByID(ctx context.Context, id uint) (*entity.FAQ, error) {
	var f entity.FAQ
	if err := r.db.WithContext(ctx).First(&f, id).Error; err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *faqRepository) Update(ctx context.Context, f *entity.FAQ) error {
	return r.db.WithContext(ctx).Save(f).Error
}

func (r *faqRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.FAQ{}, id).Error
}

func (r *faqRepository) List(ctx context.Context, offset, limit int, q string, active *bool, order string) ([]entity.FAQ, int64, error) {
	tx := r.db.WithContext(ctx).Model(&entity.FAQ{})
	q = strings.TrimSpace(q)
	if q != "" {
		like := "%" + q + "%"
		tx = tx.Where("question ILIKE ? OR answer ILIKE ?", like, like)
	}
	if active != nil {
		tx = tx.Where("is_active = ?", *active)
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if strings.TrimSpace(order) == "" {
		order = "sort_order ASC, id ASC"
	}
	var rows []entity.FAQ
	err := tx.Order(order).Offset(offset).Limit(limit).Find(&rows).Error
	return rows, total, err
}
