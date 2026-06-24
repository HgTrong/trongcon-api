package service

import (
	"context"
	"errors"
	"strings"

	foodv1 "trongcon-api/api/food/v1"
	"trongcon-api/internal/apimap"
	"trongcon-api/internal/entity"
	"trongcon-api/internal/repository"

	"gorm.io/gorm"
)

var ErrFoodNotFound = errors.New("food not found")

type FoodService interface {
	Create(ctx context.Context, req *foodv1.CreateReq) (*foodv1.CreateRes, error)
	GetByID(ctx context.Context, id uint) (*foodv1.GetRes, error)
	Update(ctx context.Context, id uint, req *foodv1.UpdateReq) (*foodv1.UpdateRes, error)
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, req *foodv1.ListReq) (*foodv1.ListRes, error)
}

type foodService struct {
	repo repository.FoodRepository
}

func NewFoodService(repo repository.FoodRepository) FoodService {
	return &foodService{repo: repo}
}

func withDefaultServing(v float64) float64 {
	if v <= 0 {
		return 100
	}
	return v
}

func normalizeMacro(v float64) float64 {
	if v < 0 {
		return 0
	}
	return v
}

func estimateCalories(protein, carb, fat float64) float64 {
	return protein*4 + carb*4 + fat*9
}

func (s *foodService) Create(ctx context.Context, req *foodv1.CreateReq) (*foodv1.CreateRes, error) {
	protein := normalizeMacro(req.Protein)
	carb := normalizeMacro(req.Carb)
	fat := normalizeMacro(req.Fat)
	calories := req.Calories
	if calories <= 0 {
		calories = estimateCalories(protein, carb, fat)
	}
	f := &entity.Food{
		Name:         req.Name,
		Protein:      protein,
		Carb:         carb,
		Fat:          fat,
		Calories:     normalizeMacro(calories),
		ServingSizeG: withDefaultServing(req.ServingSizeG),
	}
	if err := s.repo.Create(ctx, f); err != nil {
		return nil, err
	}
	fresh, err := s.repo.GetByID(ctx, f.ID)
	if err != nil {
		return nil, err
	}
	return &foodv1.CreateRes{Food: apimap.FoodToRes(fresh)}, nil
}

func (s *foodService) GetByID(ctx context.Context, id uint) (*foodv1.GetRes, error) {
	f, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFoodNotFound
		}
		return nil, err
	}
	return &foodv1.GetRes{Food: apimap.FoodToRes(f)}, nil
}

func (s *foodService) Update(ctx context.Context, id uint, req *foodv1.UpdateReq) (*foodv1.UpdateRes, error) {
	f, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFoodNotFound
		}
		return nil, err
	}
	if req.Name != nil && *req.Name != "" {
		f.Name = *req.Name
	}
	if req.Protein != nil {
		f.Protein = normalizeMacro(*req.Protein)
	}
	if req.Carb != nil {
		f.Carb = normalizeMacro(*req.Carb)
	}
	if req.Fat != nil {
		f.Fat = normalizeMacro(*req.Fat)
	}
	if req.Calories != nil {
		f.Calories = normalizeMacro(*req.Calories)
	} else if req.Protein != nil || req.Carb != nil || req.Fat != nil {
		f.Calories = estimateCalories(f.Protein, f.Carb, f.Fat)
	}
	if req.ServingSizeG != nil {
		f.ServingSizeG = withDefaultServing(*req.ServingSizeG)
	}
	if err := s.repo.Update(ctx, f); err != nil {
		return nil, err
	}
	fresh, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &foodv1.UpdateRes{Food: apimap.FoodToRes(fresh)}, nil
}

func (s *foodService) Delete(ctx context.Context, id uint) error {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrFoodNotFound
		}
		return err
	}
	return s.repo.Delete(ctx, id)
}

func (s *foodService) List(ctx context.Context, req *foodv1.ListReq) (*foodv1.ListRes, error) {
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
	case "id", "name", "protein", "carb", "fat", "calories", "created_at":
	default:
		orderBy = "id"
	}
	dir := strings.ToUpper(strings.TrimSpace(req.OrderDir))
	if dir != "ASC" && dir != "DESC" {
		dir = "DESC"
	}
	order := orderBy + " " + dir

	list, total, err := s.repo.List(ctx, offset, limit, order, strings.TrimSpace(req.Q))
	if err != nil {
		return nil, err
	}
	data := make([]foodv1.FoodRes, 0, len(list))
	for i := range list {
		data = append(data, apimap.FoodToRes(&list[i]))
	}
	return &foodv1.ListRes{Total: total, Data: data}, nil
}
