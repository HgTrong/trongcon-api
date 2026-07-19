package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	mealplanv1 "trongcon-api/api/meal_plan/v1"
	"trongcon-api/internal/entity"
	"trongcon-api/internal/repository"

	"gorm.io/gorm"
)

var ErrMealPlanNotFound = errors.New("meal plan not found")
var ErrForbiddenMealPlan = errors.New("meal plan not owned by user")

const defaultMealPlanMeals = 3

type MealPlanService interface {
	Create(ctx context.Context, req *mealplanv1.CreateReq) (*mealplanv1.CreateRes, error)
	GetByID(ctx context.Context, id uint) (*mealplanv1.GetRes, error)
	GetByIDPublic(ctx context.Context, id uint) (*mealplanv1.GetRes, error)
	Update(ctx context.Context, id uint, req *mealplanv1.UpdateReq) (*mealplanv1.UpdateRes, error)
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, req *mealplanv1.ListReq) (*mealplanv1.ListRes, error)
	ListPublic(ctx context.Context, req *mealplanv1.ListReq) (*mealplanv1.ListRes, error)

	CreateForUser(ctx context.Context, userID uint, title, description string, isPublic bool, meals []mealplanv1.MealPlanMealInput) (*mealplanv1.CreateRes, error)
	ListForUser(ctx context.Context, userID uint, page, limit int, q string) (*mealplanv1.ListRes, error)
	GetForUser(ctx context.Context, userID, id uint) (*mealplanv1.GetRes, error)
	UpdateForUser(ctx context.Context, userID, id uint, title, description *string, isPublic *bool, meals *[]mealplanv1.MealPlanMealInput) (*mealplanv1.UpdateRes, error)
	DeleteForUser(ctx context.Context, userID, id uint) error
	HydratePreview(ctx context.Context, title, description string, meals []mealplanv1.MealPlanMealInput) (*mealplanv1.MealPlanRes, error)
}

type mealPlanService struct {
	repo        repository.MealPlanRepository
	foodRepo    repository.FoodRepository
	userRepo    repository.UserRepository
	trainerRepo repository.TrainerProfileRepository
}

func NewMealPlanService(repo repository.MealPlanRepository, foodRepo repository.FoodRepository, userRepo repository.UserRepository, trainerRepo repository.TrainerProfileRepository) MealPlanService {
	return &mealPlanService{repo: repo, foodRepo: foodRepo, userRepo: userRepo, trainerRepo: trainerRepo}
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

func buildMealItems(ctx context.Context, inputs []mealplanv1.MealPlanItemInput, foodRepo repository.FoodRepository) ([]entity.MealPlanItem, error) {
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

func defaultMealInputs() []mealplanv1.MealPlanMealInput {
	out := make([]mealplanv1.MealPlanMealInput, 0, defaultMealPlanMeals)
	for i := 1; i <= defaultMealPlanMeals; i++ {
		out = append(out, mealplanv1.MealPlanMealInput{
			Name:  fmt.Sprintf("Meal %d", i),
			Items: []mealplanv1.MealPlanItemInput{},
		})
	}
	return out
}

func normalizeMealInputs(meals []mealplanv1.MealPlanMealInput) []mealplanv1.MealPlanMealInput {
	if len(meals) == 0 {
		return defaultMealInputs()
	}
	out := make([]mealplanv1.MealPlanMealInput, 0, len(meals))
	for i, m := range meals {
		name := strings.TrimSpace(m.Name)
		if name == "" {
			name = fmt.Sprintf("Meal %d", i+1)
		}
		out = append(out, mealplanv1.MealPlanMealInput{
			Name:  name,
			Items: m.Items,
		})
	}
	return out
}

func buildMealsSnapshot(ctx context.Context, inputs []mealplanv1.MealPlanMealInput, foodRepo repository.FoodRepository) ([]entity.MealPlanMeal, error) {
	inputs = normalizeMealInputs(inputs)
	meals := make([]entity.MealPlanMeal, 0, len(inputs))
	for i, in := range inputs {
		items, err := buildMealItems(ctx, in.Items, foodRepo)
		if err != nil {
			return nil, err
		}
		meals = append(meals, entity.MealPlanMeal{
			Name:      in.Name,
			SortOrder: i,
			Items:     items,
		})
	}
	return meals, nil
}

func sumItemTotals(items []entity.MealPlanItem) mealplanv1.MacroTotalsRes {
	var t mealplanv1.MacroTotalsRes
	for _, it := range items {
		t.Calories += it.Calories
		t.ProteinG += it.Protein
		t.CarbG += it.Carb
		t.FatG += it.Fat
	}
	return t
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
		Meals:       make([]mealplanv1.MealPlanMealRes, 0, len(mp.Meals)),
		MealCount:   len(mp.Meals),
	}
	if mp.User.ID > 0 {
		res.UserEmail = mp.User.Email
	}
	for _, meal := range mp.Meals {
		totals := sumItemTotals(meal.Items)
		mealRes := mealplanv1.MealPlanMealRes{
			ID:        meal.ID,
			Name:      meal.Name,
			SortOrder: meal.SortOrder,
			Totals:    totals,
			Items:     make([]mealplanv1.MealPlanItemRes, 0, len(meal.Items)),
		}
		for _, it := range meal.Items {
			mealRes.Items = append(mealRes.Items, mealplanv1.MealPlanItemRes{
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
		}
		res.Meals = append(res.Meals, mealRes)
		res.TotalProtein += totals.ProteinG
		res.TotalCarb += totals.CarbG
		res.TotalFat += totals.FatG
		res.TotalCalories += totals.Calories
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
	meals, err := buildMealsSnapshot(ctx, req.Meals, s.foodRepo)
	if err != nil {
		return nil, err
	}
	if err := s.repo.ReplaceMeals(ctx, mp.ID, meals); err != nil {
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

func (s *mealPlanService) GetByIDPublic(ctx context.Context, id uint) (*mealplanv1.GetRes, error) {
	mp, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMealPlanNotFound
		}
		return nil, err
	}
	if !mp.IsPublic {
		return nil, ErrMealPlanNotFound
	}
	res := toMealPlanRes(mp)
	res.Author = authorForUserID(ctx, s.trainerRepo, s.userRepo, mp.UserID)
	return &mealplanv1.GetRes{MealPlan: res}, nil
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
	if req.Meals != nil {
		meals, err := buildMealsSnapshot(ctx, *req.Meals, s.foodRepo)
		if err != nil {
			return nil, err
		}
		if err := s.repo.ReplaceMeals(ctx, mp.ID, meals); err != nil {
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
		res := toMealPlanRes(&list[i])
		if list[i].IsPublic {
			res.Author = authorForUserID(ctx, s.trainerRepo, s.userRepo, list[i].UserID)
		}
		data = append(data, res)
	}
	return &mealplanv1.ListRes{Total: total, Data: data}, nil
}

func (s *mealPlanService) ListPublic(ctx context.Context, req *mealplanv1.ListReq) (*mealplanv1.ListRes, error) {
	if req == nil {
		req = &mealplanv1.ListReq{}
	}
	req.IsPublic = "true"
	return s.List(ctx, req)
}

func (s *mealPlanService) assertOwner(ctx context.Context, userID, id uint) (*entity.MealPlan, error) {
	mp, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMealPlanNotFound
		}
		return nil, err
	}
	if mp.UserID != userID {
		return nil, ErrForbiddenMealPlan
	}
	return mp, nil
}

func (s *mealPlanService) CreateForUser(ctx context.Context, userID uint, title, description string, isPublic bool, meals []mealplanv1.MealPlanMealInput) (*mealplanv1.CreateRes, error) {
	return s.Create(ctx, &mealplanv1.CreateReq{
		Title:       title,
		Description: description,
		UserID:      userID,
		IsPublic:    isPublic,
		Meals:       meals,
	})
}

func (s *mealPlanService) ListForUser(ctx context.Context, userID uint, page, limit int, q string) (*mealplanv1.ListRes, error) {
	uid := userID
	return s.List(ctx, &mealplanv1.ListReq{
		Page:   page,
		Limit:  limit,
		Q:      q,
		UserID: &uid,
	})
}

func (s *mealPlanService) GetForUser(ctx context.Context, userID, id uint) (*mealplanv1.GetRes, error) {
	mp, err := s.assertOwner(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	return &mealplanv1.GetRes{MealPlan: toMealPlanRes(mp)}, nil
}

func (s *mealPlanService) UpdateForUser(ctx context.Context, userID, id uint, title, description *string, isPublic *bool, meals *[]mealplanv1.MealPlanMealInput) (*mealplanv1.UpdateRes, error) {
	mp, err := s.assertOwner(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if title != nil && strings.TrimSpace(*title) != "" {
		mp.Title = strings.TrimSpace(*title)
	}
	if description != nil {
		mp.Description = *description
	}
	if isPublic != nil {
		mp.IsPublic = *isPublic
	}
	if err := s.repo.Update(ctx, mp); err != nil {
		return nil, err
	}
	if meals != nil {
		snap, err := buildMealsSnapshot(ctx, *meals, s.foodRepo)
		if err != nil {
			return nil, err
		}
		if err := s.repo.ReplaceMeals(ctx, mp.ID, snap); err != nil {
			return nil, err
		}
	}
	fresh, err := s.repo.GetByID(ctx, mp.ID)
	if err != nil {
		return nil, err
	}
	return &mealplanv1.UpdateRes{MealPlan: toMealPlanRes(fresh)}, nil
}

func (s *mealPlanService) DeleteForUser(ctx context.Context, userID, id uint) error {
	if _, err := s.assertOwner(ctx, userID, id); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

func (s *mealPlanService) HydratePreview(ctx context.Context, title, description string, meals []mealplanv1.MealPlanMealInput) (*mealplanv1.MealPlanRes, error) {
	snap, err := buildMealsSnapshot(ctx, meals, s.foodRepo)
	if err != nil {
		return nil, err
	}
	mp := &entity.MealPlan{
		Title:       title,
		Description: description,
		IsPublic:    false,
		Meals:       snap,
	}
	res := toMealPlanRes(mp)
	return &res, nil
}
