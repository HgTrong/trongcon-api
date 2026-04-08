package service

import (
	"context"
	"errors"
	"strings"

	equipmentv1 "trongcon-api/api/equipment/v1"
	"trongcon-api/internal/apimap"
	"trongcon-api/internal/entity"
	"trongcon-api/internal/repository"

	"gorm.io/gorm"
)

var ErrEquipmentNotFound = errors.New("equipment not found")

type EquipmentService interface {
	Create(ctx context.Context, req *equipmentv1.CreateReq) (*equipmentv1.CreateRes, error)
	GetByID(ctx context.Context, id uint) (*equipmentv1.GetRes, error)
	Update(ctx context.Context, id uint, req *equipmentv1.UpdateReq) (*equipmentv1.UpdateRes, error)
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, req *equipmentv1.ListReq) (*equipmentv1.ListRes, error)
}

type equipmentService struct {
	repo repository.EquipmentRepository
}

func NewEquipmentService(repo repository.EquipmentRepository) EquipmentService {
	return &equipmentService{repo: repo}
}

func (s *equipmentService) Create(ctx context.Context, req *equipmentv1.CreateReq) (*equipmentv1.CreateRes, error) {
	e := &entity.Equipment{
		Name: req.Name,
		Icon: req.Icon,
	}
	if err := s.repo.Create(ctx, e); err != nil {
		return nil, err
	}
	fresh, err := s.repo.GetByID(ctx, e.ID)
	if err != nil {
		return nil, err
	}
	return &equipmentv1.CreateRes{Equipment: apimap.EquipmentToRes(fresh)}, nil
}

func (s *equipmentService) GetByID(ctx context.Context, id uint) (*equipmentv1.GetRes, error) {
	e, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrEquipmentNotFound
		}
		return nil, err
	}
	return &equipmentv1.GetRes{Equipment: apimap.EquipmentToRes(e)}, nil
}

func (s *equipmentService) Update(ctx context.Context, id uint, req *equipmentv1.UpdateReq) (*equipmentv1.UpdateRes, error) {
	e, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrEquipmentNotFound
		}
		return nil, err
	}
	if req.Name != nil && *req.Name != "" {
		e.Name = *req.Name
	}
	if req.Icon != nil {
		e.Icon = *req.Icon
	}
	if err := s.repo.Update(ctx, e); err != nil {
		return nil, err
	}
	fresh, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &equipmentv1.UpdateRes{Equipment: apimap.EquipmentToRes(fresh)}, nil
}

func (s *equipmentService) Delete(ctx context.Context, id uint) error {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrEquipmentNotFound
		}
		return err
	}
	return s.repo.Delete(ctx, id)
}

func (s *equipmentService) List(ctx context.Context, req *equipmentv1.ListReq) (*equipmentv1.ListRes, error) {
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
	data := make([]equipmentv1.EquipmentRes, 0, len(list))
	for i := range list {
		data = append(data, apimap.EquipmentToRes(&list[i]))
	}
	return &equipmentv1.ListRes{Total: total, Data: data}, nil
}
