package service

import (
	"context"
	"errors"

	savedv1 "trongcon-api/api/saved_workout/v1"
	workoutv1 "trongcon-api/api/workout/v1"
	"trongcon-api/internal/repository"

	"gorm.io/gorm"
)

var ErrSavedWorkoutNotFound = errors.New("saved workout not found")

type SavedWorkoutService interface {
	Save(ctx context.Context, userID, workoutID uint) (*savedv1.SaveRes, error)
	Unsave(ctx context.Context, userID, workoutID uint) error
	ListIDs(ctx context.Context, userID uint) (*savedv1.IDsRes, error)
	List(ctx context.Context, userID uint, req *savedv1.ListReq) (*savedv1.ListRes, error)
}

type savedWorkoutService struct {
	repo        repository.SavedWorkoutRepository
	workoutRepo repository.WorkoutRepository
	growth      PTGrowthTracker
}

func NewSavedWorkoutService(repo repository.SavedWorkoutRepository, workoutRepo repository.WorkoutRepository, growth PTGrowthTracker) SavedWorkoutService {
	return &savedWorkoutService{repo: repo, workoutRepo: workoutRepo, growth: growth}
}

func (s *savedWorkoutService) Save(ctx context.Context, userID, workoutID uint) (*savedv1.SaveRes, error) {
	w, err := s.workoutRepo.GetByID(ctx, workoutID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWorkoutNotFound
		}
		return nil, err
	}

	row, err := s.repo.Save(ctx, userID, workoutID)
	if err != nil {
		return nil, err
	}
	if s.growth != nil && w.UserID > 0 {
		s.growth.TrackContentSave(ctx, ContentTypeWorkout, w.ID, w.Title, w.UserID, userID)
	}
	return &savedv1.SaveRes{
		Status:    "saved",
		WorkoutID: workoutID,
		SavedAt:   row.CreatedAt,
	}, nil
}

func (s *savedWorkoutService) Unsave(ctx context.Context, userID, workoutID uint) error {
	saved, err := s.repo.IsSaved(ctx, userID, workoutID)
	if err != nil {
		return err
	}
	if !saved {
		return ErrSavedWorkoutNotFound
	}
	return s.repo.Unsave(ctx, userID, workoutID)
}

func (s *savedWorkoutService) ListIDs(ctx context.Context, userID uint) (*savedv1.IDsRes, error) {
	ids, err := s.repo.ListWorkoutIDs(ctx, userID)
	if err != nil {
		return nil, err
	}
	if ids == nil {
		ids = []uint{}
	}
	return &savedv1.IDsRes{WorkoutIDs: ids}, nil
}

func (s *savedWorkoutService) List(ctx context.Context, userID uint, req *savedv1.ListReq) (*savedv1.ListRes, error) {
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

	list, total, err := s.repo.List(ctx, userID, offset, limit)
	if err != nil {
		return nil, err
	}
	data := make([]workoutv1.WorkoutRes, 0, len(list))
	for i := range list {
		data = append(data, toWorkoutAPIRes(&list[i]))
	}
	return &savedv1.ListRes{Total: total, Data: data}, nil
}
