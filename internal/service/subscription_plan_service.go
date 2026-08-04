package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	planv1 "trongcon-api/api/subscription_plan/v1"
	"trongcon-api/internal/entity"
	"trongcon-api/internal/money"
	"trongcon-api/internal/repository"

	"gorm.io/gorm"
)

type SubscriptionPlanService interface {
	Create(ctx context.Context, req *planv1.CreateReq) (*planv1.GetRes, error)
	Update(ctx context.Context, id uint, req *planv1.UpdateReq) (*planv1.GetRes, error)
	Delete(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*planv1.GetRes, error)
	List(ctx context.Context, req *planv1.ListReq) (*planv1.ListRes, error)
	ListPublic(ctx context.Context, req *planv1.ListReq) (*planv1.ListRes, error)
}

type subscriptionPlanService struct {
	repo repository.SubscriptionPlanRepository
}

func NewSubscriptionPlanService(repo repository.SubscriptionPlanRepository) SubscriptionPlanService {
	return &subscriptionPlanService{repo: repo}
}

func (s *subscriptionPlanService) Create(ctx context.Context, req *planv1.CreateReq) (*planv1.GetRes, error) {
	active := true
	if req.IsActive != nil {
		active = *req.IsActive
	}
	currency := money.Normalize(req.Currency)
	kind := strings.TrimSpace(req.Kind)
	if kind == "" {
		kind = entity.PlanKindPremium
	}
	code := strings.TrimSpace(req.Code)
	if code == "" {
		code = slugCode(req.PlanName)
	}
	p := &entity.SubscriptionPlan{
		Code:           code,
		PlanName:       strings.TrimSpace(req.PlanName),
		Title:          strings.TrimSpace(req.Title),
		Description:    encodeBullets(req.Description),
		Price:          req.Price,
		Currency:       currency,
		DurationMonths: req.DurationMonths,
		IsActive:       active,
		SortOrder:      req.SortOrder,
		Kind:           kind,
	}
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	return &planv1.GetRes{Plan: toPlanRes(p)}, nil
}

func (s *subscriptionPlanService) Update(ctx context.Context, id uint, req *planv1.UpdateReq) (*planv1.GetRes, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("plan not found")
		}
		return nil, err
	}
	if req.Code != nil {
		p.Code = strings.TrimSpace(*req.Code)
	}
	if req.PlanName != nil {
		p.PlanName = strings.TrimSpace(*req.PlanName)
	}
	if req.Title != nil {
		p.Title = strings.TrimSpace(*req.Title)
	}
	if req.Description != nil {
		p.Description = encodeBullets(req.Description)
	}
	if req.Price != nil {
		p.Price = *req.Price
	}
	if req.Currency != nil {
		p.Currency = money.Normalize(*req.Currency)
	}
	if req.DurationMonths != nil {
		p.DurationMonths = *req.DurationMonths
	}
	if req.IsActive != nil {
		p.IsActive = *req.IsActive
	}
	if req.SortOrder != nil {
		p.SortOrder = *req.SortOrder
	}
	if req.Kind != nil {
		p.Kind = strings.TrimSpace(*req.Kind)
	}
	if err := s.repo.Update(ctx, p); err != nil {
		return nil, err
	}
	return &planv1.GetRes{Plan: toPlanRes(p)}, nil
}

func (s *subscriptionPlanService) Delete(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}

func (s *subscriptionPlanService) GetByID(ctx context.Context, id uint) (*planv1.GetRes, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("plan not found")
		}
		return nil, err
	}
	return &planv1.GetRes{Plan: toPlanRes(p)}, nil
}

func (s *subscriptionPlanService) List(ctx context.Context, req *planv1.ListReq) (*planv1.ListRes, error) {
	page, limit := pageLimit(req.Page, req.Limit)
	rows, total, err := s.repo.List(ctx, (page-1)*limit, limit, req.Q, req.Kind, req.OrderBy, req.OrderDir, req.Active)
	if err != nil {
		return nil, err
	}
	out := make([]planv1.PlanRes, 0, len(rows))
	for i := range rows {
		out = append(out, toPlanRes(&rows[i]))
	}
	return &planv1.ListRes{Total: total, Data: out}, nil
}

func (s *subscriptionPlanService) ListPublic(ctx context.Context, req *planv1.ListReq) (*planv1.ListRes, error) {
	active := true
	req.Active = &active
	if req.Kind == "" {
		req.Kind = entity.PlanKindPremium
	}
	if req.OrderBy == "" {
		req.OrderBy = "sort_order"
	}
	if req.OrderDir == "" {
		req.OrderDir = "ASC"
	}
	return s.List(ctx, req)
}

func toPlanRes(p *entity.SubscriptionPlan) planv1.PlanRes {
	return planv1.PlanRes{
		ID:             p.ID,
		Code:           p.Code,
		PlanName:       p.PlanName,
		Title:          p.Title,
		Description:    decodeBullets(p.Description),
		Price:          p.Price,
		Currency:       p.Currency,
		DurationMonths: p.DurationMonths,
		IsActive:       p.IsActive,
		SortOrder:      p.SortOrder,
		Kind:           p.Kind,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}
}

func encodeBullets(items []string) string {
	if items == nil {
		items = []string{}
	}
	b, _ := json.Marshal(items)
	return string(b)
}

func decodeBullets(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []string{}
	}
	var out []string
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return []string{raw}
	}
	return out
}

func slugCode(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	var b strings.Builder
	prevDash := false
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			prevDash = false
			continue
		}
		if !prevDash {
			b.WriteByte('-')
			prevDash = true
		}
	}
	s := strings.Trim(b.String(), "-")
	if s == "" {
		s = fmt.Sprintf("plan-%d", time.Now().Unix())
	}
	return s
}

func pageLimit(page, limit int) (int, int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return page, limit
}
