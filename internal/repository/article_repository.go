package repository

import (
	"context"
	"errors"
	"strings"

	"trongcon-api/internal/entity"

	"gorm.io/gorm"
)

type ArticleRepository interface {
	Create(ctx context.Context, a *entity.Article) error
	GetByID(ctx context.Context, id uint) (*entity.Article, error)
	Update(ctx context.Context, a *entity.Article) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, offset, limit int, order string, categoryID *uint, featured *bool, q string, activeCategoryOnly bool) ([]entity.Article, int64, error)
	SlugExists(ctx context.Context, slug string, excludeID uint) (bool, error)
	GetBySlug(ctx context.Context, slug string) (*entity.Article, error)
	IncrementViews(ctx context.Context, id uint) (int64, error)
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

func (r *articleRepository) GetBySlug(ctx context.Context, slug string) (*entity.Article, error) {
	var a entity.Article
	if err := r.db.WithContext(ctx).Preload("User").Preload("Category").Where("slug = ?", slug).First(&a).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *articleRepository) Update(ctx context.Context, a *entity.Article) error {
	// Omit BelongsTo associations so a changed UserID/CategoryID is not
	// overwritten by the previously preloaded User/Category primary keys.
	return r.db.WithContext(ctx).Omit("User", "Category").Save(a).Error
}

func (r *articleRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.Article{}, id).Error
}

func (r *articleRepository) List(ctx context.Context, offset, limit int, order string, categoryID *uint, featured *bool, q string, activeCategoryOnly bool) ([]entity.Article, int64, error) {
	query := r.db.WithContext(ctx).Model(&entity.Article{})
	if activeCategoryOnly {
		query = query.Joins("JOIN categories ON categories.id = articles.category_id AND categories.status = ?", "active")
	}
	if categoryID != nil && *categoryID > 0 {
		query = query.Where("articles.category_id = ?", *categoryID)
	}
	if featured != nil {
		query = query.Where("articles.featured = ?", *featured)
	}
	if term := strings.TrimSpace(q); term != "" {
		like := "%" + term + "%"
		query = query.Where(
			"articles.title ILIKE ? OR articles.subtitle ILIKE ? OR articles.slug ILIKE ?",
			like, like, like,
		)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if order == "" {
		order = "articles.id DESC"
	} else if !strings.Contains(order, ".") {
		order = "articles." + order
	}

	var list []entity.Article
	if err := query.Preload("User").Preload("Category").Order(order).Offset(offset).Limit(limit).Find(&list).Error; err != nil {
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

func (r *articleRepository) IncrementViews(ctx context.Context, id uint) (int64, error) {
	res := r.db.WithContext(ctx).Model(&entity.Article{}).Where("id = ?", id).
		UpdateColumn("views", gorm.Expr("views + 1"))
	if res.Error != nil {
		return 0, res.Error
	}
	var views int64
	if err := r.db.WithContext(ctx).Model(&entity.Article{}).Where("id = ?", id).Select("views").Scan(&views).Error; err != nil {
		return 0, err
	}
	return views, nil
}
