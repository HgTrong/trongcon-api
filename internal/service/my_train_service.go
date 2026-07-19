package service

import (
	"context"
	"errors"
	"strings"

	mytrainv1 "trongcon-api/api/my_train/v1"
	routinev1 "trongcon-api/api/routine/v1"
	"trongcon-api/internal/entity"
	"trongcon-api/internal/repository"

	"gorm.io/gorm"
)

var ErrForbiddenWorkout = errors.New("workout not owned by user")
var ErrForbiddenRoutine = errors.New("routine not owned by user")

type MyTrainService interface {
	CreateWorkout(ctx context.Context, userID uint, req *mytrainv1.CreateMyWorkoutReq) (*mytrainv1.CreateRes, error)
	CloneFromCatalog(ctx context.Context, userID uint, req *mytrainv1.CloneCatalogReq) (*mytrainv1.CreateRes, error)
	GetWorkout(ctx context.Context, userID, id uint) (*mytrainv1.GetRes, error)
	UpdateWorkout(ctx context.Context, userID, id uint, req *mytrainv1.UpdateMyWorkoutReq) (*mytrainv1.UpdateRes, error)
	DeleteWorkout(ctx context.Context, userID, id uint) error
	ListWorkouts(ctx context.Context, userID uint, req *mytrainv1.ListMyWorkoutsReq) (*mytrainv1.ListRes, error)

	CreateRoutine(ctx context.Context, userID uint, req *mytrainv1.CreateRoutineReq) (*routinev1.CreateRes, error)
	GetRoutine(ctx context.Context, userID, id uint) (*routinev1.GetRes, error)
	UpdateRoutine(ctx context.Context, userID, id uint, req *mytrainv1.UpdateRoutineReq) (*routinev1.UpdateRes, error)
	DeleteRoutine(ctx context.Context, userID, id uint) error
	ListRoutines(ctx context.Context, userID uint, req *mytrainv1.ListMyRoutinesReq) (*routinev1.ListRes, error)
}

type myTrainService struct {
	workoutRepo  repository.WorkoutRepository
	exerciseRepo repository.ExerciseRepository
	routineRepo  repository.RoutineRepository
}

func NewMyTrainService(
	workoutRepo repository.WorkoutRepository,
	exerciseRepo repository.ExerciseRepository,
	routineRepo repository.RoutineRepository,
) MyTrainService {
	return &myTrainService{
		workoutRepo:  workoutRepo,
		exerciseRepo: exerciseRepo,
		routineRepo:  routineRepo,
	}
}

func toMyWorkoutRes(w *entity.Workout) mytrainv1.WorkoutRes {
	res := mytrainv1.WorkoutRes{
		ID:          w.ID,
		Title:       w.Title,
		Difficulty:  w.Difficulty,
		Goal:        w.Goal,
		ImageURL:    w.ImageURL,
		UserID:      w.UserID,
		OwnerUserID: w.OwnerUserID,
		IsPublic:    w.IsPublic,
		CreatedAt:   w.CreatedAt,
		UpdatedAt:   w.UpdatedAt,
		Items:       make([]mytrainv1.WorkoutItemRes, 0, len(w.Items)),
	}
	for _, it := range w.Items {
		res.Items = append(res.Items, mytrainv1.WorkoutItemRes{
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

func (s *myTrainService) ownedWorkout(ctx context.Context, userID, id uint) (*entity.Workout, error) {
	w, err := s.workoutRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWorkoutNotFound
		}
		return nil, err
	}
	// Personal copy owned by user, or catalog item they authored (Posted by).
	if w.OwnerUserID != nil && *w.OwnerUserID == userID {
		return w, nil
	}
	if w.OwnerUserID == nil && w.UserID == userID {
		return w, nil
	}
	return nil, ErrForbiddenWorkout
}

func (s *myTrainService) CreateWorkout(ctx context.Context, userID uint, req *mytrainv1.CreateMyWorkoutReq) (*mytrainv1.CreateRes, error) {
	inputs := make([]workoutItemBridge, len(req.Items))
	for i, in := range req.Items {
		inputs[i] = workoutItemBridge{ExerciseID: in.ExerciseID, Sets: in.Sets, Reps: in.Reps}
	}
	items, err := buildMyWorkoutItems(ctx, inputs, s.exerciseRepo)
	if err != nil {
		return nil, err
	}
	owner := userID
	w := &entity.Workout{
		Title:       strings.TrimSpace(req.Title),
		Difficulty:  req.Difficulty,
		Goal:        req.Goal,
		ImageURL:    strings.TrimSpace(req.ImageURL),
		UserID:      userID,
		OwnerUserID: &owner,
		IsPublic:    req.IsPublic,
		Items:       items,
	}
	if err := s.workoutRepo.Create(ctx, w); err != nil {
		return nil, err
	}
	fresh, err := s.workoutRepo.GetByID(ctx, w.ID)
	if err != nil {
		return nil, err
	}
	return &mytrainv1.CreateRes{Workout: toMyWorkoutRes(fresh)}, nil
}

func (s *myTrainService) CloneFromCatalog(ctx context.Context, userID uint, req *mytrainv1.CloneCatalogReq) (*mytrainv1.CreateRes, error) {
	src, err := s.workoutRepo.GetByID(ctx, req.WorkoutID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWorkoutNotFound
		}
		return nil, err
	}
	if src.OwnerUserID != nil {
		return nil, ErrWorkoutNotFound
	}
	items := make([]entity.WorkoutItem, 0, len(src.Items))
	for _, it := range src.Items {
		items = append(items, entity.WorkoutItem{
			ExerciseID:   it.ExerciseID,
			ExerciseName: it.ExerciseName,
			Sets:         it.Sets,
			Reps:         it.Reps,
		})
	}
	owner := userID
	w := &entity.Workout{
		Title:       src.Title,
		Difficulty:  src.Difficulty,
		Goal:        src.Goal,
		ImageURL:    src.ImageURL,
		UserID:      userID,
		OwnerUserID: &owner,
		Items:       items,
	}
	if err := s.workoutRepo.Create(ctx, w); err != nil {
		return nil, err
	}
	fresh, err := s.workoutRepo.GetByID(ctx, w.ID)
	if err != nil {
		return nil, err
	}
	return &mytrainv1.CreateRes{Workout: toMyWorkoutRes(fresh)}, nil
}

func (s *myTrainService) GetWorkout(ctx context.Context, userID, id uint) (*mytrainv1.GetRes, error) {
	w, err := s.ownedWorkout(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	return &mytrainv1.GetRes{Workout: toMyWorkoutRes(w)}, nil
}

func (s *myTrainService) UpdateWorkout(ctx context.Context, userID, id uint, req *mytrainv1.UpdateMyWorkoutReq) (*mytrainv1.UpdateRes, error) {
	w, err := s.ownedWorkout(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if req.Title != nil && strings.TrimSpace(*req.Title) != "" {
		w.Title = strings.TrimSpace(*req.Title)
	}
	if req.Difficulty != nil {
		w.Difficulty = *req.Difficulty
	}
	if req.Goal != nil {
		w.Goal = *req.Goal
	}
	if req.ImageURL != nil {
		w.ImageURL = strings.TrimSpace(*req.ImageURL)
	}
	if req.IsPublic != nil {
		w.IsPublic = *req.IsPublic
	}
	if err := s.workoutRepo.Update(ctx, w); err != nil {
		return nil, err
	}
	if req.Items != nil {
		inputs := make([]workoutItemBridge, len(*req.Items))
		for i, in := range *req.Items {
			inputs[i] = workoutItemBridge{ExerciseID: in.ExerciseID, Sets: in.Sets, Reps: in.Reps}
		}
		items, err := buildMyWorkoutItems(ctx, inputs, s.exerciseRepo)
		if err != nil {
			return nil, err
		}
		if err := s.workoutRepo.ReplaceItems(ctx, w.ID, items); err != nil {
			return nil, err
		}
	}
	fresh, err := s.workoutRepo.GetByID(ctx, w.ID)
	if err != nil {
		return nil, err
	}
	return &mytrainv1.UpdateRes{Workout: toMyWorkoutRes(fresh)}, nil
}

func (s *myTrainService) DeleteWorkout(ctx context.Context, userID, id uint) error {
	if _, err := s.ownedWorkout(ctx, userID, id); err != nil {
		return err
	}
	return s.workoutRepo.Delete(ctx, id)
}

func (s *myTrainService) ListWorkouts(ctx context.Context, userID uint, req *mytrainv1.ListMyWorkoutsReq) (*mytrainv1.ListRes, error) {
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
	list, total, err := s.workoutRepo.ListByOwner(ctx, userID, offset, limit, "id DESC", strings.TrimSpace(req.Q))
	if err != nil {
		return nil, err
	}
	data := make([]mytrainv1.WorkoutRes, 0, len(list))
	for i := range list {
		data = append(data, toMyWorkoutRes(&list[i]))
	}
	return &mytrainv1.ListRes{Total: total, Data: data}, nil
}

type workoutItemBridge struct {
	ExerciseID uint
	Sets       int
	Reps       string
}

func buildMyWorkoutItems(ctx context.Context, inputs []workoutItemBridge, exerciseRepo repository.ExerciseRepository) ([]entity.WorkoutItem, error) {
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

func (s *myTrainService) ownedRoutine(ctx context.Context, userID, id uint) (*entity.Routine, error) {
	rt, err := s.routineRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRoutineNotFound
		}
		return nil, err
	}
	if rt.UserID != userID {
		return nil, ErrForbiddenRoutine
	}
	return rt, nil
}

func (s *myTrainService) assertUsableWorkout(ctx context.Context, userID, workoutID uint) (*entity.Workout, error) {
	w, err := s.workoutRepo.GetByID(ctx, workoutID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWorkoutNotFound
		}
		return nil, err
	}
	if w.OwnerUserID != nil && *w.OwnerUserID != userID {
		return nil, ErrForbiddenWorkout
	}
	return w, nil
}

func (s *myTrainService) CreateRoutine(ctx context.Context, userID uint, req *mytrainv1.CreateRoutineReq) (*routinev1.CreateRes, error) {
	rt := &entity.Routine{
		Title:       strings.TrimSpace(req.Title),
		Description: strings.TrimSpace(req.Description),
		ImageURL:    strings.TrimSpace(req.ImageURL),
		Difficulty:  req.Difficulty,
		UserID:      userID,
		IsPublic:    req.IsPublic,
	}
	if err := s.routineRepo.Create(ctx, rt); err != nil {
		return nil, err
	}
	items := make([]entity.RoutineWorkout, 0, len(req.Items))
	for _, in := range req.Items {
		w, err := s.assertUsableWorkout(ctx, userID, in.WorkoutID)
		if err != nil {
			return nil, err
		}
		items = append(items, entity.RoutineWorkout{
			WorkoutID:    w.ID,
			WorkoutTitle: w.Title,
		})
	}
	if err := s.routineRepo.ReplaceItems(ctx, rt.ID, items); err != nil {
		return nil, err
	}
	fresh, err := s.routineRepo.GetByID(ctx, rt.ID)
	if err != nil {
		return nil, err
	}
	return &routinev1.CreateRes{Routine: toRoutineRes(fresh)}, nil
}

func (s *myTrainService) GetRoutine(ctx context.Context, userID, id uint) (*routinev1.GetRes, error) {
	rt, err := s.ownedRoutine(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	return &routinev1.GetRes{Routine: toRoutineRes(rt)}, nil
}

func (s *myTrainService) UpdateRoutine(ctx context.Context, userID, id uint, req *mytrainv1.UpdateRoutineReq) (*routinev1.UpdateRes, error) {
	rt, err := s.ownedRoutine(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if req.Title != nil && strings.TrimSpace(*req.Title) != "" {
		rt.Title = strings.TrimSpace(*req.Title)
	}
	if req.Description != nil {
		rt.Description = strings.TrimSpace(*req.Description)
	}
	if req.ImageURL != nil {
		rt.ImageURL = strings.TrimSpace(*req.ImageURL)
	}
	if req.Difficulty != nil {
		rt.Difficulty = *req.Difficulty
	}
	if req.IsPublic != nil {
		rt.IsPublic = *req.IsPublic
	}
	if err := s.routineRepo.Update(ctx, rt); err != nil {
		return nil, err
	}
	if req.Items != nil {
		items := make([]entity.RoutineWorkout, 0, len(*req.Items))
		for _, in := range *req.Items {
			w, err := s.assertUsableWorkout(ctx, userID, in.WorkoutID)
			if err != nil {
				return nil, err
			}
			items = append(items, entity.RoutineWorkout{
				WorkoutID:    w.ID,
				WorkoutTitle: w.Title,
			})
		}
		if err := s.routineRepo.ReplaceItems(ctx, rt.ID, items); err != nil {
			return nil, err
		}
	}
	fresh, err := s.routineRepo.GetByID(ctx, rt.ID)
	if err != nil {
		return nil, err
	}
	return &routinev1.UpdateRes{Routine: toRoutineRes(fresh)}, nil
}

func (s *myTrainService) DeleteRoutine(ctx context.Context, userID, id uint) error {
	if _, err := s.ownedRoutine(ctx, userID, id); err != nil {
		return err
	}
	return s.routineRepo.Delete(ctx, id)
}

func (s *myTrainService) ListRoutines(ctx context.Context, userID uint, req *mytrainv1.ListMyRoutinesReq) (*routinev1.ListRes, error) {
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
	uid := userID
	list, total, err := s.routineRepo.List(ctx, offset, limit, "id DESC", strings.TrimSpace(req.Q), "", &uid, nil)
	if err != nil {
		return nil, err
	}
	data := make([]routinev1.RoutineRes, 0, len(list))
	for i := range list {
		data = append(data, toRoutineRes(&list[i]))
	}
	return &routinev1.ListRes{Total: total, Data: data}, nil
}
