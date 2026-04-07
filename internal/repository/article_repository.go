package repository

import (
	"context"
	"errors"

	"trongcon-api/internal/entity"

	"gorm.io/gorm"
)

type ArticleRepository interface {
	Create(ctx context.Context, a *entity.Article) error
	GetByID(ctx context.Context, id uint) (*entity.Article, error)
	Update(ctx context.Context, a *entity.Article) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, offset, limit int, order string) ([]entity.Article, int64, error)
	SlugExists(ctx context.Context, slug string, excludeID uint) (bool, error)
}

type articleRepository struct {
	db *gorm.DB
}

func NewArticleRepository(db *gorm.DB) ArticleRepository {
	return &articleRepository{db: db}
}

func (r *articleRepository) Create(ctx context.Context, a *entity.Article) error {
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *articleRepository) GetByID(ctx context.Context, id uint) (*entity.Article, error) {
	var a entity.Article
	if err := r.db.WithContext(ctx).Preload("User").Preload("Category").First(&a, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	return &a, nil
}

func (r *articleRepository) Update(ctx context.Context, a *entity.Article) error {
	return r.db.WithContext(ctx).Save(a).Error
}

func (r *articleRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.Article{}, id).Error
}

func (r *articleRepository) List(ctx context.Context, offset, limit int, order string) ([]entity.Article, int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&entity.Article{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []entity.Article
	if order == "" {
		order = "id DESC"
	}
	if err := r.db.WithContext(ctx).Preload("User").Preload("Category").Order(order).Offset(offset).Limit(limit).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *articleRepository) SlugExists(ctx context.Context, slug string, excludeID uint) (bool, error) {
	if slug == "" {
		return false, nil
	}
	var n int64
	q := r.db.WithContext(ctx).Model(&entity.Article{}).Where("slug = ?", slug)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	if err := q.Count(&n).Error; err != nil {
		return false, err
	}
	return n > 0, nil
}
