package service

import (
	"context"
	"errors"
	"strings"

	workoutv1 "trongcon-api/api/workout/v1"
	"trongcon-api/internal/entity"
	"trongcon-api/internal/repository"

	"gorm.io/gorm"
)

var ErrWorkoutNotFound = errors.New("workout not found")

type WorkoutService interface {
	Create(ctx context.Context, req *workoutv1.CreateReq) (*workoutv1.CreateRes, error)
	GetByID(ctx context.Context, id uint) (*workoutv1.GetRes, error)
	Update(ctx context.Context, id uint, req *workoutv1.UpdateReq) (*workoutv1.UpdateRes, error)
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, req *workoutv1.ListReq) (*workoutv1.ListRes, error)
}

type workoutService struct {
	repo         repository.WorkoutRepository
	exerciseRepo repository.ExerciseRepository
}

func NewWorkoutService(repo repository.WorkoutRepository, exerciseRepo repository.ExerciseRepository) WorkoutService {
	return &workoutService{repo: repo, exerciseRepo: exerciseRepo}
}

func normalizeSets(v int) int {
	if v <= 0 {
		return 3
	}
	return v
}

func normalizeReps(v string) string {
	s := strings.TrimSpace(v)
	if s == "" {
		return "10"
	}
	return s
}

func buildWorkoutItemsFromInput(ctx context.Context, inputs []workoutv1.WorkoutItemInput, exerciseRepo repository.ExerciseRepository) ([]entity.WorkoutItem, error) {
	items := make([]entity.WorkoutItem, 0, len(inputs))
	for i, in := range inputs {
		ex, err := exerciseRepo.GetByID(ctx, in.ExerciseID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrExerciseNotFound
			}
			return nil, err
		}
		items = append(items, entity.WorkoutItem{
			ExerciseID:   ex.ID,
			ExerciseName: ex.Name,
			SortOrder:    i + 1,
			Sets:         normalizeSets(in.Sets),
			Reps:         normalizeReps(in.Reps),
		})
	}
	return items, nil
}

func toWorkoutAPIRes(w *entity.Workout) workoutv1.WorkoutRes {
	res := workoutv1.WorkoutRes{
		ID:         w.ID,
		Title:      w.Title,
		Difficulty: w.Difficulty,
		CreatedAt:  w.CreatedAt,
		UpdatedAt:  w.UpdatedAt,
		Items:      make([]workoutv1.WorkoutItemRes, 0, len(w.Items)),
	}
	for _, it := range w.Items {
		res.Items = append(res.Items, workoutv1.WorkoutItemRes{
			ID:           it.ID,
			ExerciseID:   it.ExerciseID,
			ExerciseName: it.ExerciseName,
			SortOrder:    it.SortOrder,
			Sets:         it.Sets,
			Reps:         it.Reps,
			CreatedAt:    it.CreatedAt,
			UpdatedAt:    it.UpdatedAt,
		})
		res.ExerciseCount++
		res.TotalSets += it.Sets
	}
	return res
}

func (s *workoutService) Create(ctx context.Context, req *workoutv1.CreateReq) (*workoutv1.CreateRes, error) {
	items, err := buildWorkoutItemsFromInput(ctx, req.Items, s.exerciseRepo)
	if err != nil {
		return nil, err
	}

	w := &entity.Workout{
		Title:      strings.TrimSpace(req.Title),
		Difficulty: req.Difficulty,
		Items:      items,
	}
	if err := s.repo.Create(ctx, w); err != nil {
		return nil, err
	}

	fresh, err := s.repo.GetByID(ctx, w.ID)
	if err != nil {
		return nil, err
	}
	return &workoutv1.CreateRes{Workout: toWorkoutAPIRes(fresh)}, nil
}

func (s *workoutService) GetByID(ctx context.Context, id uint) (*workoutv1.GetRes, error) {
	w, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWorkoutNotFound
		}
		return nil, err
	}
	return &workoutv1.GetRes{Workout: toWorkoutAPIRes(w)}, nil
}

func (s *workoutService) Update(ctx context.Context, id uint, req *workoutv1.UpdateReq) (*workoutv1.UpdateRes, error) {
	w, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWorkoutNotFound
		}
		return nil, err
	}

	if req.Title != nil && strings.TrimSpace(*req.Title) != "" {
		w.Title = strings.TrimSpace(*req.Title)
	}
	if req.Difficulty != nil {
		w.Difficulty = *req.Difficulty
	}
	if err := s.repo.Update(ctx, w); err != nil {
		return nil, err
	}

	if req.Items != nil {
		items, err := buildWorkoutItemsFromInput(ctx, *req.Items, s.exerciseRepo)
		if err != nil {
			return nil, err
		}
		if err := s.repo.ReplaceItems(ctx, w.ID, items); err != nil {
			return nil, err
		}
	}

	fresh, err := s.repo.GetByID(ctx, w.ID)
	if err != nil {
		return nil, err
	}
	return &workoutv1.UpdateRes{Workout: toWorkoutAPIRes(fresh)}, nil
}

func (s *workoutService) Delete(ctx context.Context, id uint) error {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrWorkoutNotFound
		}
		return err
	}
	return s.repo.Delete(ctx, id)
}

func (s *workoutService) List(ctx context.Context, req *workoutv1.ListReq) (*workoutv1.ListRes, error) {
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
	case "id", "title", "created_at", "difficulty":
	default:
		orderBy = "id"
	}
	dir := strings.ToUpper(strings.TrimSpace(req.OrderDir))
	if dir != "ASC" && dir != "DESC" {
		dir = "DESC"
	}
	order := orderBy + " " + dir

	list, total, err := s.repo.List(ctx, offset, limit, order, strings.TrimSpace(req.Q), strings.TrimSpace(req.Difficulty))
	if err != nil {
		return nil, err
	}
	data := make([]workoutv1.WorkoutRes, 0, len(list))
	for i := range list {
		data = append(data, toWorkoutAPIRes(&list[i]))
	}
	return &workoutv1.ListRes{Total: total, Data: data}, nil
}
