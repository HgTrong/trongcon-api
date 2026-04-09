package service

import (
	"context"
	"errors"
	"strings"

	musclev1 "trongcon-api/api/muscle/v1"
	"trongcon-api/internal/apimap"
	"trongcon-api/internal/entity"
	"trongcon-api/internal/repository"

	"gorm.io/gorm"
)

var ErrMuscleNotFound = errors.New("muscle not found")

type MuscleService interface {
	Create(ctx context.Context, req *musclev1.CreateReq) (*musclev1.CreateRes, error)
	GetByID(ctx context.Context, id uint) (*musclev1.GetRes, error)
	Update(ctx context.Context, id uint, req *musclev1.UpdateReq) (*musclev1.UpdateRes, error)
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, req *musclev1.ListReq) (*musclev1.ListRes, error)
}

type muscleService struct {
	repo repository.MuscleRepository
}

func NewMuscleService(repo repository.MuscleRepository) MuscleService {
	return &muscleService{repo: repo}
}

func (s *muscleService) Create(ctx context.Context, req *musclev1.CreateReq) (*musclev1.CreateRes, error) {
	m := &entity.Muscle{Name: req.Name}
	if err := s.repo.Create(ctx, m); err != nil {
		return nil, err
	}
	fresh, err := s.repo.GetByID(ctx, m.ID)
	if err != nil {
		return nil, err
	}
	return &musclev1.CreateRes{Muscle: apimap.MuscleToRes(fresh)}, nil
}

func (s *muscleService) GetByID(ctx context.Context, id uint) (*musclev1.GetRes, error) {
	m, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMuscleNotFound
		}
		return nil, err
	}
	return &musclev1.GetRes{Muscle: apimap.MuscleToRes(m)}, nil
}

func (s *muscleService) Update(ctx context.Context, id uint, req *musclev1.UpdateReq) (*musclev1.UpdateRes, error) {
	m, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMuscleNotFound
		}
		return nil, err
	}
	if req.Name != nil && *req.Name != "" {
		m.Name = *req.Name
	}
	if err := s.repo.Update(ctx, m); err != nil {
		return nil, err
	}
	fresh, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &musclev1.UpdateRes{Muscle: apimap.MuscleToRes(fresh)}, nil
}

func (s *muscleService) Delete(ctx context.Context, id uint) error {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrMuscleNotFound
		}
		return err
	}
	return s.repo.Delete(ctx, id)
}

func (s *muscleService) List(ctx context.Context, req *musclev1.ListReq) (*musclev1.ListRes, error) {
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
	data := make([]musclev1.MuscleRes, 0, len(list))
	for i := range list {
		data = append(data, apimap.MuscleToRes(&list[i]))
	}
	return &musclev1.ListRes{Total: total, Data: data}, nil
}
