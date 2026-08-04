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
	AdminGetByID(ctx context.Context, id uint) (*workoutv1.GetRes, error)
	Update(ctx context.Context, id uint, req *workoutv1.UpdateReq) (*workoutv1.UpdateRes, error)
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, req *workoutv1.ListReq) (*workoutv1.ListRes, error)
	AdminList(ctx context.Context, req *workoutv1.ListReq) (*workoutv1.ListRes, error)
}

type workoutService struct {
	repo         repository.WorkoutRepository
	exerciseRepo repository.ExerciseRepository
	userRepo     repository.UserRepository
	trainerRepo  repository.TrainerProfileRepository
	growth       PTGrowthTracker
}

func NewWorkoutService(
	repo repository.WorkoutRepository,
	exerciseRepo repository.ExerciseRepository,
	userRepo repository.UserRepository,
	trainerRepo repository.TrainerProfileRepository,
	growth PTGrowthTracker,
) WorkoutService {
	return &workoutService{repo: repo, exerciseRepo: exerciseRepo, userRepo: userRepo, trainerRepo: trainerRepo, growth: growth}
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
		ID:          w.ID,
		Title:       w.Title,
		Difficulty:  w.Difficulty,
		Goal:        w.Goal,
		ImageURL:    w.ImageURL,
		UserID:      w.UserID,
		OwnerUserID: w.OwnerUserID,
		IsPublic:    w.IsPublic,
		Views:       w.Views,
		CreatedAt:   w.CreatedAt,
		UpdatedAt:   w.UpdatedAt,
		Items:       make([]workoutv1.WorkoutItemRes, 0, len(w.Items)),
	}
	if w.User.ID > 0 {
		res.UserEmail = w.User.Email
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

func (s *workoutService) withAuthor(ctx context.Context, w *entity.Workout, res workoutv1.WorkoutRes) workoutv1.WorkoutRes {
	res.Author = authorForUserID(ctx, s.trainerRepo, s.userRepo, workoutAuthorID(w))
	return res
}

func (s *workoutService) Create(ctx context.Context, req *workoutv1.CreateReq) (*workoutv1.CreateRes, error) {
	if _, err := s.userRepo.GetByID(ctx, req.UserID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	items, err := buildWorkoutItemsFromInput(ctx, req.Items, s.exerciseRepo)
	if err != nil {
		return nil, err
	}

	// Admin create = catalog (OwnerUserID nil). UserID is the posted-by author.
	w := &entity.Workout{
		Title:      strings.TrimSpace(req.Title),
		Difficulty: req.Difficulty,
		Goal:       req.Goal,
		ImageURL:   strings.TrimSpace(req.ImageURL),
		UserID:     req.UserID,
		IsPublic:   req.IsPublic,
		Items:      items,
	}
	if err := s.repo.Create(ctx, w); err != nil {
		return nil, err
	}

	fresh, err := s.repo.GetByID(ctx, w.ID)
	if err != nil {
		return nil, err
	}
	return &workoutv1.CreateRes{Workout: s.withAuthor(ctx, fresh, toWorkoutAPIRes(fresh))}, nil
}

func (s *workoutService) GetByID(ctx context.Context, id uint) (*workoutv1.GetRes, error) {
	w, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWorkoutNotFound
		}
		return nil, err
	}
	// Public Get exposes catalog workouts (owner null) or published PT ones.
	if w.OwnerUserID != nil && !w.IsPublic {
		return nil, ErrWorkoutNotFound
	}
	if views, err := s.repo.IncrementViews(ctx, w.ID); err == nil {
		w.Views = views
	}
	if s.growth != nil && w.UserID > 0 {
		s.growth.TrackContentView(ctx, ContentTypeWorkout, w.ID, w.Title, w.UserID, 0)
	}
	return &workoutv1.GetRes{Workout: s.withAuthor(ctx, w, toWorkoutAPIRes(w))}, nil
}

// AdminGetByID returns any workout (including private PT drafts) without bumping views.
func (s *workoutService) AdminGetByID(ctx context.Context, id uint) (*workoutv1.GetRes, error) {
	w, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWorkoutNotFound
		}
		return nil, err
	}
	return &workoutv1.GetRes{Workout: s.withAuthor(ctx, w, toWorkoutAPIRes(w))}, nil
}

func (s *workoutService) Update(ctx context.Context, id uint, req *workoutv1.UpdateReq) (*workoutv1.UpdateRes, error) {
	w, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWorkoutNotFound
		}
		return nil, err
	}
	// Admin may edit catalog rows and PT-owned workouts alike.

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
	if req.UserID != nil {
		if _, err := s.userRepo.GetByID(ctx, *req.UserID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrUserNotFound
			}
			return nil, err
		}
		w.UserID = *req.UserID
	}
	if req.IsPublic != nil {
		w.IsPublic = *req.IsPublic
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
	return &workoutv1.UpdateRes{Workout: s.withAuthor(ctx, fresh, toWorkoutAPIRes(fresh))}, nil
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

func (s *workoutService) listPage(ctx context.Context, req *workoutv1.ListReq, catalogOnly bool) (*workoutv1.ListRes, error) {
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
	case "id", "title", "created_at", "difficulty", "goal":
	default:
		orderBy = "id"
	}
	dir := strings.ToUpper(strings.TrimSpace(req.OrderDir))
	if dir != "ASC" && dir != "DESC" {
		dir = "DESC"
	}
	order := orderBy + " " + dir

	var list []entity.Workout
	var total int64
	var err error
	if catalogOnly {
		list, total, err = s.repo.ListCatalog(ctx, offset, limit, order, strings.TrimSpace(req.Q), strings.TrimSpace(req.Difficulty), strings.TrimSpace(req.Goal))
	} else {
		list, total, err = s.repo.ListAll(ctx, offset, limit, order, strings.TrimSpace(req.Q), strings.TrimSpace(req.Difficulty), strings.TrimSpace(req.Goal))
	}
	if err != nil {
		return nil, err
	}
	data := make([]workoutv1.WorkoutRes, 0, len(list))
	for i := range list {
		data = append(data, s.withAuthor(ctx, &list[i], toWorkoutAPIRes(&list[i])))
	}
	return &workoutv1.ListRes{Total: total, Data: data}, nil
}

func (s *workoutService) List(ctx context.Context, req *workoutv1.ListReq) (*workoutv1.ListRes, error) {
	return s.listPage(ctx, req, true)
}

func (s *workoutService) AdminList(ctx context.Context, req *workoutv1.ListReq) (*workoutv1.ListRes, error) {
	return s.listPage(ctx, req, false)
}
