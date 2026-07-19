package repository

import (
	"context"

	"trongcon-api/internal/entity"

	"gorm.io/gorm"
)

type AiChatRepository interface {
	CreateThread(ctx context.Context, t *entity.AiChatThread) error
	GetThread(ctx context.Context, userID, id uint) (*entity.AiChatThread, error)
	ListThreads(ctx context.Context, userID uint, limit int) ([]entity.AiChatThread, error)
	UpdateThreadTitle(ctx context.Context, id uint, title string) error
	AddMessage(ctx context.Context, m *entity.AiChatMessage) error
	ListMessages(ctx context.Context, threadID uint, limit int) ([]entity.AiChatMessage, error)
}

type aiChatRepository struct {
	db *gorm.DB
}

func NewAiChatRepository(db *gorm.DB) AiChatRepository {
	return &aiChatRepository{db: db}
}

func (r *aiChatRepository) CreateThread(ctx context.Context, t *entity.AiChatThread) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *aiChatRepository) GetThread(ctx context.Context, userID, id uint) (*entity.AiChatThread, error) {
	var t entity.AiChatThread
	if err := r.db.WithContext(ctx).Where("user_id = ? AND id = ?", userID, id).First(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *aiChatRepository) ListThreads(ctx context.Context, userID uint, limit int) ([]entity.AiChatThread, error) {
	if limit <= 0 {
		limit = 20
	}
	var list []entity.AiChatThread
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("updated_at DESC, id DESC").
		Limit(limit).
		Find(&list).Error
	return list, err
}

func (r *aiChatRepository) UpdateThreadTitle(ctx context.Context, id uint, title string) error {
	return r.db.WithContext(ctx).Model(&entity.AiChatThread{}).Where("id = ?", id).Update("title", title).Error
}

func (r *aiChatRepository) AddMessage(ctx context.Context, m *entity.AiChatMessage) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *aiChatRepository) ListMessages(ctx context.Context, threadID uint, limit int) ([]entity.AiChatMessage, error) {
	if limit <= 0 {
		limit = 40
	}
	var list []entity.AiChatMessage
	err := r.db.WithContext(ctx).
		Where("thread_id = ?", threadID).
		Order("id ASC").
		Limit(limit).
		Find(&list).Error
	return list, err
}
