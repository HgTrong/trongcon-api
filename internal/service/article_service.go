package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	articlev1 "trongcon-api/api/article/v1"
	"trongcon-api/internal/apimap"
	"trongcon-api/internal/entity"
	"trongcon-api/internal/pkg/slug"
	"trongcon-api/internal/repository"

	"gorm.io/gorm"
)

var ErrArticleNotFound = errors.New("article not found")

type ArticleService interface {
	Create(ctx context.Context, req *articlev1.CreateReq) (*articlev1.CreateRes, error)
	GetByID(ctx context.Context, id uint) (*articlev1.GetRes, error)
	Update(ctx context.Context, id uint, req *articlev1.UpdateReq) (*articlev1.UpdateRes, error)
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, req *articlev1.ListReq) (*articlev1.ListRes, error)
}

type articleService struct {
	articleRepo  repository.ArticleRepository
	categoryRepo repository.CategoryRepository
	userRepo     repository.UserRepository
}

func NewArticleService(
	articleRepo repository.ArticleRepository,
	categoryRepo repository.CategoryRepository,
	userRepo repository.UserRepository,
) ArticleService {
	return &articleService{
		articleRepo:  articleRepo,
		categoryRepo: categoryRepo,
		userRepo:     userRepo,
	}
}

func (s *articleService) allocateUniqueSlug(ctx context.Context, base string, excludeID uint) (string, error) {
	if base == "" {
		base = "article"
	}
	for n := 0; n < 10000; n++ {
		candidate := base
		if n > 0 {
			candidate = fmt.Sprintf("%s-%d", base, n)
		}
		exists, err := s.articleRepo.SlugExists(ctx, candidate, excludeID)
		if err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
	}
	return "", errors.New("could not allocate unique slug")
}

func (s *articleService) Create(ctx context.Context, req *articlev1.CreateReq) (*articlev1.CreateRes, error) {
	if _, err := s.categoryRepo.GetByID(ctx, req.CategoryID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCategoryNotFound
		}
		return nil, err
	}
	if _, err := s.userRepo.GetByID(ctx, req.UserID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	slugBase := slug.FromTitle(req.Title)
	slugVal, err := s.allocateUniqueSlug(ctx, slugBase, 0)
	if err != nil {
		return nil, err
	}
	a := &entity.Article{
		Title:      req.Title,
		Subtitle:   req.Subtitle,
		Slug:       slugVal,
		Thumbnail:  req.Thumbnail,
		Video:      req.Video,
		Content:    req.Content,
		UserID:     req.UserID,
		CategoryID: req.CategoryID,
	}
	if err := s.articleRepo.Create(ctx, a); err != nil {
		return nil, err
	}
	fresh, err := s.articleRepo.GetByID(ctx, a.ID)
	if err != nil {
		return nil, err
	}
	return &articlev1.CreateRes{Article: apimap.ArticleToDetail(fresh)}, nil
}

func (s *articleService) GetByID(ctx context.Context, id uint) (*articlev1.GetRes, error) {
	a, err := s.articleRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrArticleNotFound
		}
		return nil, err
	}
	return &articlev1.GetRes{Article: apimap.ArticleToDetail(a)}, nil
}

func (s *articleService) Update(ctx context.Context, id uint, req *articlev1.UpdateReq) (*articlev1.UpdateRes, error) {
	a, err := s.articleRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrArticleNotFound
		}
		return nil, err
	}
	if a.Slug == "" {
		slugVal, err := s.allocateUniqueSlug(ctx, slug.FromTitle(a.Title), a.ID)
		if err != nil {
			return nil, err
		}
		a.Slug = slugVal
	}
	oldTitle := a.Title
	if req.Title != nil && *req.Title != "" {
		a.Title = *req.Title
	}
	if req.Title != nil && *req.Title != "" && *req.Title != oldTitle {
		slugBase := slug.FromTitle(a.Title)
		slugVal, err := s.allocateUniqueSlug(ctx, slugBase, a.ID)
		if err != nil {
			return nil, err
		}
		a.Slug = slugVal
	}
	if req.Subtitle != nil {
		a.Subtitle = *req.Subtitle
	}
	if req.Thumbnail != nil {
		a.Thumbnail = *req.Thumbnail
	}
	if req.Video != nil {
		a.Video = *req.Video
	}
	if req.Content != nil {
		a.Content = *req.Content
	}
	if req.UserID != nil {
		if _, err := s.userRepo.GetByID(ctx, *req.UserID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrUserNotFound
			}
			return nil, err
		}
		a.UserID = *req.UserID
	}
	if req.CategoryID != nil {
		if _, err := s.categoryRepo.GetByID(ctx, *req.CategoryID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrCategoryNotFound
			}
			return nil, err
		}
		a.CategoryID = *req.CategoryID
	}
	if err := s.articleRepo.Update(ctx, a); err != nil {
		return nil, err
	}
	fresh, err := s.articleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &articlev1.UpdateRes{Article: apimap.ArticleToDetail(fresh)}, nil
}

func (s *articleService) Delete(ctx context.Context, id uint) error {
	if _, err := s.articleRepo.GetByID(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrArticleNotFound
		}
		return err
	}
	return s.articleRepo.Delete(ctx, id)
}

func (s *articleService) List(ctx context.Context, req *articlev1.ListReq) (*articlev1.ListRes, error) {
	page, limit := req.Page, req.Limit
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit

	orderBy := strings.ToLower(strings.TrimSpace(req.OrderBy))
	if orderBy == "" {
		orderBy = "id"
	}
	switch orderBy {
	case "id", "title", "created_at":
	default:
		orderBy = "id"
	}
	dir := strings.ToUpper(strings.TrimSpace(req.OrderDir))
	if dir != "ASC" && dir != "DESC" {
		dir = "DESC"
	}
	order := orderBy + " " + dir

	list, total, err := s.articleRepo.List(ctx, offset, limit, order)
	if err != nil {
		return nil, err
	}
	data := make([]articlev1.ArticleListRes, 0, len(list))
	for i := range list {
		data = append(data, apimap.ArticleToList(&list[i]))
	}
	return &articlev1.ListRes{Total: total, Data: data}, nil
}
