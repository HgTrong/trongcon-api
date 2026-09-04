package repository

import (
	"context"

	"trongcon-api/internal/entity"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ContentShareRepository interface {
	Create(ctx context.Context, s *entity.ContentShare) error
	Delete(ctx context.Context, contentType string, contentID, recipientUserID uint) error
	ListRecipients(ctx context.Context, contentType string, contentID uint) ([]entity.ContentShare, error)
	IsSharedWithUser(ctx context.Context, contentType string, contentID, userID uint) (bool, error)
	ListSharedWithUser(ctx context.Context, userID uint, contentType string) ([]entity.ContentShare, error)
}

type contentShareRepository struct{ db *gorm.DB }

func NewContentShareRepository(db *gorm.DB) ContentShareRepository {
	return &contentShareRepository{db: db}
}

// Create is idempotent — re-sharing with the same recipient is a no-op, not an error.
func (r *contentShareRepository) Create(ctx context.Context, s *entity.ContentShare) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(s).Error
}

func (r *contentShareRepository) Delete(ctx context.Context, contentType string, contentID, recipientUserID uint) error {
	return r.db.WithContext(ctx).
		Where("content_type = ? AND content_id = ? AND recipient_user_id = ?", contentType, contentID, recipientUserID).
		Delete(&entity.ContentShare{}).Error
}

func (r *contentShareRepository) ListRecipients(ctx context.Context, contentType string, contentID uint) ([]entity.ContentShare, error) {
	var rows []entity.ContentShare
	err := r.db.WithContext(ctx).Preload("Recipient").
		Where("content_type = ? AND content_id = ?", contentType, contentID).
		Order("id DESC").Find(&rows).Error
	return rows, err
}

func (r *contentShareRepository) IsSharedWithUser(ctx context.Context, contentType string, contentID, userID uint) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&entity.ContentShare{}).
		Where("content_type = ? AND content_id = ? AND recipient_user_id = ?", contentType, contentID, userID).
		Limit(1).Count(&n).Error
	return n > 0, err
}

func (r *contentShareRepository) ListSharedWithUser(ctx context.Context, userID uint, contentType string) ([]entity.ContentShare, error) {
	q := r.db.WithContext(ctx).Preload("SharedBy").Where("recipient_user_id = ?", userID)
	if contentType != "" {
		q = q.Where("content_type = ?", contentType)
	}
	var rows []entity.ContentShare
	err := q.Order("id DESC").Find(&rows).Error
	return rows, err
}
