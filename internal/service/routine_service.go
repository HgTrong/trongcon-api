package service

import (
	"context"
	"errors"
	"strings"

	routinev1 "trongcon-api/api/routine/v1"
	"trongcon-api/internal/entity"
	"trongcon-api/internal/repository"

	"gorm.io/gorm"
)

var ErrRoutineNotFound = errors.New("routine not found")

type RoutineService interface {
	Create(ctx context.Context, req *routinev1.CreateReq) (*routinev1.CreateRes, error)
	GetByID(ctx context.Context, id uint) (*routinev1.GetRes, error)
	Update(ctx context.Context, id uint, req *routinev1.UpdateReq) (*routinev1.UpdateRes, error)
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, req *routinev1.ListReq) (*routinev1.ListRes, error)
}

type routineService struct {
	repo        repository.RoutineRepository
	workoutRepo repository.WorkoutRepository
	userRepo    repository.UserRepository
}

func NewRoutineService(repo repository.RoutineRepository, workoutRepo repository.WorkoutRepository, userRepo repository.UserRepository) RoutineService {
	return &routineService{repo: repo, workoutRepo: workoutRepo, userRepo: userRepo}
}

func buildRoutineWorkouts(ctx context.Context, inputs []routinev1.RoutineItemInput, workoutRepo repository.WorkoutRepository) ([]entity.RoutineWorkout, error) {
	items := make([]entity.RoutineWorkout, 0, len(inputs))
	for _, in := range inputs {
		if in.WorkoutID == 0 {
			continue
		}
		w, err := workoutRepo.GetByID(ctx, in.WorkoutID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrWorkoutNotFound
			}
			return nil, err
		}
		items = append(items, entity.RoutineWorkout{
			WorkoutID:    w.ID,
			WorkoutTitle: w.Title,
		})
	}
	return items, nil
}

func toRoutineWorkoutRes(rw *entity.RoutineWorkout) routinev1.RoutineWorkoutRes {
	res := routinev1.RoutineWorkoutRes{
		ID:           rw.ID,
		WorkoutID:    rw.WorkoutID,
		WorkoutTitle: rw.WorkoutTitle,
		SortOrder:    rw.SortOrder,
		CreatedAt:    rw.CreatedAt,
		UpdatedAt:    rw.UpdatedAt,
		Items:        make([]routinev1.WorkoutItemRes, 0),
	}
	if rw.Workout.ID > 0 {
		res.Difficulty = rw.Workout.Difficulty
		for _, it := range rw.Workout.Items {
			res.Items = append(res.Items, routinev1.WorkoutItemRes{
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
	}
	return res
}

func toRoutineRes(rt *entity.Routine) routinev1.RoutineRes {
	res := routinev1.RoutineRes{
		ID:          rt.ID,
		Title:       rt.Title,
		Description: rt.Description,
		Difficulty:  rt.Difficulty,
		UserID:      rt.UserID,
		IsPublic:    rt.IsPublic,
		CreatedAt:   rt.CreatedAt,
		UpdatedAt:   rt.UpdatedAt,
		Items:       make([]routinev1.RoutineWorkoutRes, 0, len(rt.Items)),
	}
	if rt.User.ID > 0 {
		res.UserEmail = rt.User.Email
	}
	for i := range rt.Items {
		wr := toRoutineWorkoutRes(&rt.Items[i])
		res.Items = append(res.Items, wr)
		res.WorkoutCount++
		res.ExerciseCount += wr.ExerciseCount
		res.TotalSets += wr.TotalSets
	}
	return res
}

func (s *routineService) Create(ctx context.Context, req *routinev1.CreateReq) (*routinev1.CreateRes, error) {
	if _, err := s.userRepo.GetByID(ctx, req.UserID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	rt := &entity.Routine{
		Title:       strings.TrimSpace(req.Title),
		Description: strings.TrimSpace(req.Description),
		Difficulty:  req.Difficulty,
		UserID:      req.UserID,
		IsPublic:    req.IsPublic,
	}
	if err := s.repo.Create(ctx, rt); err != nil {
		return nil, err
	}
	items, err := buildRoutineWorkouts(ctx, req.Items, s.workoutRepo)
	if err != nil {
		return nil, err
	}
	if err := s.repo.ReplaceItems(ctx, rt.ID, items); err != nil {
		return nil, err
	}
	fresh, err := s.repo.GetByID(ctx, rt.ID)
	if err != nil {
		return nil, err
	}
	return &routinev1.CreateRes{Routine: toRoutineRes(fresh)}, nil
}

func (s *routineService) GetByID(ctx context.Context, id uint) (*routinev1.GetRes, error) {
	rt, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRoutineNotFound
		}
		return nil, err
	}
	return &routinev1.GetRes{Routine: toRoutineRes(rt)}, nil
}

func (s *routineService) Update(ctx context.Context, id uint, req *routinev1.UpdateReq) (*routinev1.UpdateRes, error) {
	rt, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRoutineNotFound
		}
		return nil, err
	}
	if req.Title != nil && strings.TrimSpace(*req.Title) != "" {
		rt.Title = strings.TrimSpace(*req.Title)
	}
	if req.Description != nil {
		rt.Description = strings.TrimSpace(*req.Description)
	}
	if req.Difficulty != nil {
		rt.Difficulty = *req.Difficulty
	}
	if req.UserID != nil {
		if _, err := s.userRepo.GetByID(ctx, *req.UserID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrUserNotFound
			}
			return nil, err
		}
		rt.UserID = *req.UserID
	}
	if req.IsPublic != nil {
		rt.IsPublic = *req.IsPublic
	}
	if err := s.repo.Update(ctx, rt); err != nil {
		return nil, err
	}
	if req.Items != nil {
		items, err := buildRoutineWorkouts(ctx, *req.Items, s.workoutRepo)
		if err != nil {
			return nil, err
		}
		if err := s.repo.ReplaceItems(ctx, rt.ID, items); err != nil {
			return nil, err
		}
	}
	fresh, err := s.repo.GetByID(ctx, rt.ID)
	if err != nil {
		return nil, err
	}
	return &routinev1.UpdateRes{Routine: toRoutineRes(fresh)}, nil
}

func (s *routineService) Delete(ctx context.Context, id uint) error {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRoutineNotFound
		}
		return err
	}
	return s.repo.Delete(ctx, id)
}

func (s *routineService) List(ctx context.Context, req *routinev1.ListReq) (*routinev1.ListRes, error) {
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

	list, total, err := s.repo.List(
		ctx,
		offset,
		limit,
		order,
		strings.TrimSpace(req.Q),
		strings.TrimSpace(req.Difficulty),
		req.UserID,
		parseBoolPointer(req.IsPublic),
	)
	if err != nil {
		return nil, err
	}
	data := make([]routinev1.RoutineRes, 0, len(list))
	for i := range list {
		data = append(data, toRoutineRes(&list[i]))
	}
	return &routinev1.ListRes{Total: total, Data: data}, nil
}
