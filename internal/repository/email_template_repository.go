package repository

import (
	"context"
	"errors"
	"strings"

	"trongcon-api/internal/entity"

	"gorm.io/gorm"
)

type EmailTemplateRepository interface {
	Create(ctx context.Context, t *entity.EmailTemplate) error
	GetByID(ctx context.Context, id uint) (*entity.EmailTemplate, error)
	GetByKey(ctx context.Context, key string) (*entity.EmailTemplate, error)
	Update(ctx context.Context, t *entity.EmailTemplate) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, offset, limit int, order, q string, isActive *bool) ([]entity.EmailTemplate, int64, error)
}

type emailTemplateRepository struct {
	db *gorm.DB
}

func NewEmailTemplateRepository(db *gorm.DB) EmailTemplateRepository {
	return &emailTemplateRepository{db: db}
}

func (r *emailTemplateRepository) Create(ctx context.Context, t *entity.EmailTemplate) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *emailTemplateRepository) GetByID(ctx context.Context, id uint) (*entity.EmailTemplate, error) {
	var t entity.EmailTemplate
	if err := r.db.WithContext(ctx).First(&t, id).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *emailTemplateRepository) GetByKey(ctx context.Context, key string) (*entity.EmailTemplate, error) {
	var t entity.EmailTemplate
	if err := r.db.WithContext(ctx).Where("key = ?", strings.TrimSpace(key)).First(&t).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	return &t, nil
}

func (r *emailTemplateRepository) Update(ctx context.Context, t *entity.EmailTemplate) error {
	return r.db.WithContext(ctx).Save(t).Error
}

func (r *emailTemplateRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.EmailTemplate{}, id).Error
}

func (r *emailTemplateRepository) List(ctx context.Context, offset, limit int, order, q string, isActive *bool) ([]entity.EmailTemplate, int64, error) {
	query := r.db.WithContext(ctx).Model(&entity.EmailTemplate{})
	if q = strings.TrimSpace(q); q != "" {
		like := "%" + q + "%"
		query = query.Where("name ILIKE ? OR key ILIKE ? OR subject ILIKE ?", like, like, like)
	}
	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if strings.TrimSpace(order) == "" {
		order = "id DESC"
	}
	var list []entity.EmailTemplate
	if err := query.Order(order).Offset(offset).Limit(limit).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
