package service

import (
	"context"
	"errors"
	"strings"

	mealplanv1 "trongcon-api/api/meal_plan/v1"
	"trongcon-api/internal/entity"
	"trongcon-api/internal/repository"

	"gorm.io/gorm"
)

var ErrMealPlanNotFound = errors.New("meal plan not found")

type MealPlanService interface {
	Create(ctx context.Context, req *mealplanv1.CreateReq) (*mealplanv1.CreateRes, error)
	GetByID(ctx context.Context, id uint) (*mealplanv1.GetRes, error)
	Update(ctx context.Context, id uint, req *mealplanv1.UpdateReq) (*mealplanv1.UpdateRes, error)
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, req *mealplanv1.ListReq) (*mealplanv1.ListRes, error)
}

type mealPlanService struct {
	repo     repository.MealPlanRepository
	foodRepo repository.FoodRepository
	userRepo repository.UserRepository
}

func NewMealPlanService(repo repository.MealPlanRepository, foodRepo repository.FoodRepository, userRepo repository.UserRepository) MealPlanService {
	return &mealPlanService{repo: repo, foodRepo: foodRepo, userRepo: userRepo}
}

func normalizeQty(v float64) float64 {
	if v <= 0 {
		return 1
	}
	return v
}

func mergeMealPlanInputs(inputs []mealplanv1.MealPlanItemInput) []mealplanv1.MealPlanItemInput {
	qtyByFood := make(map[uint]float64)
	order := make([]uint, 0, len(inputs))
	for _, in := range inputs {
		if in.FoodID == 0 {
			continue
		}
		if _, seen := qtyByFood[in.FoodID]; !seen {
			order = append(order, in.FoodID)
		}
		qtyByFood[in.FoodID] += normalizeQty(in.Quantity)
	}
	out := make([]mealplanv1.MealPlanItemInput, 0, len(order))
	for _, foodID := range order {
		out = append(out, mealplanv1.MealPlanItemInput{
			FoodID:   foodID,
			Quantity: qtyByFood[foodID],
		})
	}
	return out
}

func buildItemsSnapshot(ctx context.Context, inputs []mealplanv1.MealPlanItemInput, foodRepo repository.FoodRepository) ([]entity.MealPlanItem, error) {
	inputs = mergeMealPlanInputs(inputs)
	items := make([]entity.MealPlanItem, 0, len(inputs))
	for _, in := range inputs {
		food, err := foodRepo.GetByID(ctx, in.FoodID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrFoodNotFound
			}
			return nil, err
		}
		qty := normalizeQty(in.Quantity)
		items = append(items, entity.MealPlanItem{
			FoodID:       food.ID,
			FoodName:     food.Name,
			Quantity:     qty,
			ServingSizeG: food.ServingSizeG * qty,
			Protein:      food.Protein * qty,
			Carb:         food.Carb * qty,
			Fat:          food.Fat * qty,
			Calories:     food.Calories * qty,
		})
	}
	return items, nil
}

func toMealPlanRes(mp *entity.MealPlan) mealplanv1.MealPlanRes {
	res := mealplanv1.MealPlanRes{
		ID:          mp.ID,
		Title:       mp.Title,
		Description: mp.Description,
		UserID:      mp.UserID,
		IsPublic:    mp.IsPublic,
		CreatedAt:   mp.CreatedAt,
		UpdatedAt:   mp.UpdatedAt,
		Items:       make([]mealplanv1.MealPlanItemRes, 0, len(mp.Items)),
	}
	if mp.User.ID > 0 {
		res.UserEmail = mp.User.Email
	}
	for _, it := range mp.Items {
		res.Items = append(res.Items, mealplanv1.MealPlanItemRes{
			ID:           it.ID,
			FoodID:       it.FoodID,
			FoodName:     it.FoodName,
			Quantity:     it.Quantity,
			ServingSizeG: it.ServingSizeG,
			Protein:      it.Protein,
			Carb:         it.Carb,
			Fat:          it.Fat,
			Calories:     it.Calories,
			CreatedAt:    it.CreatedAt,
			UpdatedAt:    it.UpdatedAt,
		})
		res.TotalProtein += it.Protein
		res.TotalCarb += it.Carb
		res.TotalFat += it.Fat
		res.TotalCalories += it.Calories
	}
	return res
}

func parseBoolPointer(s string) *bool {
	v := strings.ToLower(strings.TrimSpace(s))
	switch v {
	case "":
		return nil
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

func (s *mealPlanService) Create(ctx context.Context, req *mealplanv1.CreateReq) (*mealplanv1.CreateRes, error) {
	if _, err := s.userRepo.GetByID(ctx, req.UserID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	mp := &entity.MealPlan{
		Title:       req.Title,
		Description: req.Description,
		UserID:      req.UserID,
		IsPublic:    req.IsPublic,
	}
	if err := s.repo.Create(ctx, mp); err != nil {
		return nil, err
	}
	items, err := buildItemsSnapshot(ctx, req.Items, s.foodRepo)
	if err != nil {
		return nil, err
	}
	if err := s.repo.ReplaceItems(ctx, mp.ID, items); err != nil {
		return nil, err
	}
	fresh, err := s.repo.GetByID(ctx, mp.ID)
	if err != nil {
		return nil, err
	}
	return &mealplanv1.CreateRes{MealPlan: toMealPlanRes(fresh)}, nil
}

func (s *mealPlanService) GetByID(ctx context.Context, id uint) (*mealplanv1.GetRes, error) {
	mp, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMealPlanNotFound
		}
		return nil, err
	}
	return &mealplanv1.GetRes{MealPlan: toMealPlanRes(mp)}, nil
}

func (s *mealPlanService) Update(ctx context.Context, id uint, req *mealplanv1.UpdateReq) (*mealplanv1.UpdateRes, error) {
	mp, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMealPlanNotFound
		}
		return nil, err
	}
	if req.Title != nil && strings.TrimSpace(*req.Title) != "" {
		mp.Title = *req.Title
	}
	if req.Description != nil {
		mp.Description = *req.Description
	}
	if req.UserID != nil {
		if _, err := s.userRepo.GetByID(ctx, *req.UserID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrUserNotFound
			}
			return nil, err
		}
		mp.UserID = *req.UserID
	}
	if req.IsPublic != nil {
		mp.IsPublic = *req.IsPublic
	}
	if err := s.repo.Update(ctx, mp); err != nil {
		return nil, err
	}
	if req.Items != nil {
		items, err := buildItemsSnapshot(ctx, *req.Items, s.foodRepo)
		if err != nil {
			return nil, err
		}
		if err := s.repo.ReplaceItems(ctx, mp.ID, items); err != nil {
			return nil, err
		}
	}
	fresh, err := s.repo.GetByID(ctx, mp.ID)
	if err != nil {
		return nil, err
	}
	return &mealplanv1.UpdateRes{MealPlan: toMealPlanRes(fresh)}, nil
}

func (s *mealPlanService) Delete(ctx context.Context, id uint) error {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrMealPlanNotFound
		}
		return err
	}
	return s.repo.Delete(ctx, id)
}

func (s *mealPlanService) List(ctx context.Context, req *mealplanv1.ListReq) (*mealplanv1.ListRes, error) {
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

	list, total, err := s.repo.List(
		ctx,
		offset,
		limit,
		order,
		strings.TrimSpace(req.Q),
		req.UserID,
		parseBoolPointer(req.IsPublic),
	)
	if err != nil {
		return nil, err
	}
	data := make([]mealplanv1.MealPlanRes, 0, len(list))
	for i := range list {
		data = append(data, toMealPlanRes(&list[i]))
	}
	return &mealplanv1.ListRes{Total: total, Data: data}, nil
}
