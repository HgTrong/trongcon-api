package service

import (
	"context"
	"errors"
	"strings"

	faqv1 "trongcon-api/api/faq/v1"
	"trongcon-api/internal/entity"
	"trongcon-api/internal/repository"

	"gorm.io/gorm"
)

var ErrFAQNotFound = errors.New("faq not found")

type FAQService interface {
	Create(ctx context.Context, req *faqv1.CreateReq) (*faqv1.CreateRes, error)
	GetByID(ctx context.Context, id uint) (*faqv1.GetRes, error)
	GetByIDPublic(ctx context.Context, id uint) (*faqv1.GetRes, error)
	Update(ctx context.Context, id uint, req *faqv1.UpdateReq) (*faqv1.UpdateRes, error)
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, req *faqv1.ListReq) (*faqv1.ListRes, error)
	ListPublic(ctx context.Context, req *faqv1.ListReq) (*faqv1.ListRes, error)
}

type faqService struct {
	repo repository.FAQRepository
}

func NewFAQService(repo repository.FAQRepository) FAQService {
	return &faqService{repo: repo}
}

func toFAQRes(f *entity.FAQ) faqv1.FAQRes {
	return faqv1.FAQRes{
		ID: f.ID, Question: f.Question, Answer: f.Answer,
		SortOrder: f.SortOrder, IsActive: f.IsActive,
		CreatedAt: f.CreatedAt, UpdatedAt: f.UpdatedAt,
	}
}

func parseFAQActive(s string) *bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "true", "yes":
		b := true
		return &b
	case "0", "false", "no":
		b := false
		return &b
	default:
		return nil
	}
}

func (s *faqService) Create(ctx context.Context, req *faqv1.CreateReq) (*faqv1.CreateRes, error) {
	active := true
	if req.IsActive != nil {
		active = *req.IsActive
	}
	f := &entity.FAQ{
		Question:  strings.TrimSpace(req.Question),
		Answer:    strings.TrimSpace(req.Answer),
		SortOrder: req.SortOrder,
		IsActive:  active,
	}
	if err := s.repo.Create(ctx, f); err != nil {
		return nil, err
	}
	fresh, err := s.repo.GetByID(ctx, f.ID)
	if err != nil {
		return nil, err
	}
	return &faqv1.CreateRes{FAQ: toFAQRes(fresh)}, nil
}

func (s *faqService) GetByID(ctx context.Context, id uint) (*faqv1.GetRes, error) {
	f, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFAQNotFound
		}
		return nil, err
	}
	return &faqv1.GetRes{FAQ: toFAQRes(f)}, nil
}

func (s *faqService) GetByIDPublic(ctx context.Context, id uint) (*faqv1.GetRes, error) {
	res, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !res.FAQ.IsActive {
		return nil, ErrFAQNotFound
	}
	return res, nil
}

func (s *faqService) Update(ctx context.Context, id uint, req *faqv1.UpdateReq) (*faqv1.UpdateRes, error) {
	f, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFAQNotFound
		}
		return nil, err
	}
	if req.Question != nil {
		f.Question = strings.TrimSpace(*req.Question)
	}
	if req.Answer != nil {
		f.Answer = strings.TrimSpace(*req.Answer)
	}
	if req.SortOrder != nil {
		f.SortOrder = *req.SortOrder
	}
	if req.IsActive != nil {
		f.IsActive = *req.IsActive
	}
	if err := s.repo.Update(ctx, f); err != nil {
		return nil, err
	}
	fresh, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &faqv1.UpdateRes{FAQ: toFAQRes(fresh)}, nil
}

func (s *faqService) Delete(ctx context.Context, id uint) error {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrFAQNotFound
		}
		return err
	}
	return s.repo.Delete(ctx, id)
}

func (s *faqService) list(ctx context.Context, req *faqv1.ListReq, forceActive *bool) (*faqv1.ListRes, error) {
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
	active := forceActive
	if active == nil {
		active = parseFAQActive(req.Active)
	}
	order := "sort_order ASC, id ASC"
	if strings.TrimSpace(req.OrderBy) != "" {
		dir := strings.ToUpper(strings.TrimSpace(req.OrderDir))
		if dir != "ASC" && dir != "DESC" {
			dir = "ASC"
		}
		col := strings.ToLower(strings.TrimSpace(req.OrderBy))
		switch col {
		case "id", "sort_order", "created_at", "updated_at":
			order = col + " " + dir
		}
	}
	rows, total, err := s.repo.List(ctx, (page-1)*limit, limit, req.Q, active, order)
	if err != nil {
		return nil, err
	}
	out := make([]faqv1.FAQRes, 0, len(rows))
	for i := range rows {
		out = append(out, toFAQRes(&rows[i]))
	}
	return &faqv1.ListRes{Total: total, Data: out}, nil
}

func (s *faqService) List(ctx context.Context, req *faqv1.ListReq) (*faqv1.ListRes, error) {
	return s.list(ctx, req, nil)
}

func (s *faqService) ListPublic(ctx context.Context, req *faqv1.ListReq) (*faqv1.ListRes, error) {
	active := true
	return s.list(ctx, req, &active)
}
