package service

import (
	"context"
	"errors"
	"strings"

	categoryv1 "trongcon-api/api/category/v1"
	"trongcon-api/internal/apimap"
	"trongcon-api/internal/entity"
	"trongcon-api/internal/repository"

	"gorm.io/gorm"
)

var (
	ErrCategoryNotFound = errors.New("category not found")
)

type CategoryService interface {
	Create(ctx context.Context, req *categoryv1.CreateReq) (*categoryv1.CreateRes, error)
	GetByID(ctx context.Context, id uint) (*categoryv1.GetRes, error)
	Update(ctx context.Context, id uint, req *categoryv1.UpdateReq) (*categoryv1.UpdateRes, error)
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, req *categoryv1.ListReq) (*categoryv1.ListRes, error)
}

type categoryService struct {
	repo repository.CategoryRepository
}

func NewCategoryService(repo repository.CategoryRepository) CategoryService {
	return &categoryService{repo: repo}
}

func (s *categoryService) Create(ctx context.Context, req *categoryv1.CreateReq) (*categoryv1.CreateRes, error) {
	status := req.Status
	if status == "" {
		status = "active"
	}
	typ := req.Type
	if typ == "" {
		typ = "article"
	}
	c := &entity.Category{
		Name:   req.Name,
		Icon:   req.Icon,
		Image:  req.Image,
		Status: status,
		Type:   typ,
	}
	if err := s.repo.Create(ctx, c); err != nil {
		return nil, err
	}
	fresh, err := s.repo.GetByID(ctx, c.ID)
	if err != nil {
		return nil, err
	}
	return &categoryv1.CreateRes{Category: apimap.CategoryToRes(fresh)}, nil
}

func (s *categoryService) GetByID(ctx context.Context, id uint) (*categoryv1.GetRes, error) {
	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCategoryNotFound
		}
		return nil, err
	}
	return &categoryv1.GetRes{Category: apimap.CategoryToRes(c)}, nil
}

func (s *categoryService) Update(ctx context.Context, id uint, req *categoryv1.UpdateReq) (*categoryv1.UpdateRes, error) {
	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCategoryNotFound
		}
		return nil, err
	}
	if req.Name != nil && *req.Name != "" {
		c.Name = *req.Name
	}
	if req.Icon != nil {
		c.Icon = *req.Icon
	}
	if req.Image != nil {
		c.Image = *req.Image
	}
	if req.Status != nil && *req.Status != "" {
		c.Status = *req.Status
	}
	if req.Type != nil && *req.Type != "" {
		c.Type = *req.Type
	}
	if err := s.repo.Update(ctx, c); err != nil {
		return nil, err
	}
	fresh, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &categoryv1.UpdateRes{Category: apimap.CategoryToRes(fresh)}, nil
}

func (s *categoryService) Delete(ctx context.Context, id uint) error {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCategoryNotFound
		}
		return err
	}
	return s.repo.Delete(ctx, id)
}

func (s *categoryService) List(ctx context.Context, req *categoryv1.ListReq) (*categoryv1.ListRes, error) {
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
	case "id", "name", "created_at":
	default:
		orderBy = "id"
	}
	dir := strings.ToUpper(strings.TrimSpace(req.OrderDir))
	if dir != "ASC" && dir != "DESC" {
		dir = "DESC"
	}
	order := orderBy + " " + dir

	list, total, err := s.repo.List(ctx, offset, limit, order)
	if err != nil {
		return nil, err
	}
	data := make([]categoryv1.CategoryRes, 0, len(list))
	for i := range list {
		data = append(data, apimap.CategoryToRes(&list[i]))
	}
	return &categoryv1.ListRes{Total: total, Data: data}, nil
}
