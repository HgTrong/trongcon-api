package service

import (
	"context"
	"errors"
	"strings"
	"time"

	sessionv1 "trongcon-api/api/workout_session/v1"
	"trongcon-api/internal/entity"
	"trongcon-api/internal/repository"

	"gorm.io/gorm"
)

var ErrSessionNotFound = errors.New("workout session not found")
var ErrSetNotFound = errors.New("set not found")

type WorkoutSessionService interface {
	Create(ctx context.Context, userID uint, req *sessionv1.CreateSessionReq) (*sessionv1.CreateSessionRes, error)
	CreateFromEnrollmentSlot(ctx context.Context, userID uint, slot *entity.EnrollmentSlot, workout *entity.Workout) (*entity.WorkoutSession, error)
	Get(ctx context.Context, userID, id uint) (*sessionv1.GetSessionRes, error)
	List(ctx context.Context, userID uint, req *sessionv1.ListSessionsReq) (*sessionv1.ListSessionsRes, error)
	UpdateSet(ctx context.Context, userID, setID uint, req *sessionv1.UpdateSetReq) (*sessionv1.SetLogRes, error)
	AddItem(ctx context.Context, userID, sessionID uint, req *sessionv1.AddSessionItemReq) (*sessionv1.SessionItemRes, error)
	Complete(ctx context.Context, userID, id uint) (*sessionv1.GetSessionRes, error)
	Abandon(ctx context.Context, userID, id uint) (*sessionv1.GetSessionRes, error)
	ExercisePerformance(ctx context.Context, userID, exerciseID uint, limit int) (*sessionv1.ExercisePerformanceRes, error)
}

type workoutSessionService struct {
	repo         repository.WorkoutSessionRepository
	workoutRepo  repository.WorkoutRepository
	exerciseRepo repository.ExerciseRepository
}

func NewWorkoutSessionService(
	repo repository.WorkoutSessionRepository,
	workoutRepo repository.WorkoutRepository,
	exerciseRepo repository.ExerciseRepository,
) WorkoutSessionService {
	return &workoutSessionService{repo: repo, workoutRepo: workoutRepo, exerciseRepo: exerciseRepo}
}

func seedSets(n int) []entity.WorkoutSetLog {
	if n <= 0 {
		n = 3
	}
	sets := make([]entity.WorkoutSetLog, n)
	for i := 0; i < n; i++ {
		sets[i] = entity.WorkoutSetLog{SetNumber: i + 1}
	}
	return sets
}

func toSetRes(s entity.WorkoutSetLog) sessionv1.SetLogRes {
	return sessionv1.SetLogRes{
		ID:          s.ID,
		SetNumber:   s.SetNumber,
		WeightKg:    s.WeightKg,
		Reps:        s.Reps,
		Completed:   s.Completed,
		CompletedAt: s.CompletedAt,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
	}
}

func (s *workoutSessionService) attachPrevious(ctx context.Context, userID uint, session *entity.WorkoutSession, items []sessionv1.SessionItemRes) []sessionv1.SessionItemRes {
	for i := range items {
		prev, err := s.repo.LastCompletedItemForExercise(ctx, userID, items[i].ExerciseID, session.ID)
		if err != nil || prev == nil {
			continue
		}
		items[i].Previous = make([]sessionv1.SetLogRes, 0, len(prev.Sets))
		for _, set := range prev.Sets {
			items[i].Previous = append(items[i].Previous, toSetRes(set))
		}
	}
	return items
}

func toSessionRes(sess *entity.WorkoutSession) sessionv1.SessionRes {
	res := sessionv1.SessionRes{
		ID:               sess.ID,
		WorkoutID:        sess.WorkoutID,
		EnrollmentSlotID: sess.EnrollmentSlotID,
		Title:            sess.Title,
		Source:           sess.Source,
		Status:           sess.Status,
		StartedAt:        sess.StartedAt,
		CompletedAt:      sess.CompletedAt,
		Notes:            sess.Notes,
		CreatedAt:        sess.CreatedAt,
		UpdatedAt:        sess.UpdatedAt,
		Items:            make([]sessionv1.SessionItemRes, 0, len(sess.Items)),
	}
	for _, it := range sess.Items {
		item := sessionv1.SessionItemRes{
			ID:           it.ID,
			ExerciseID:   it.ExerciseID,
			ExerciseName: it.ExerciseName,
			SortOrder:    it.SortOrder,
			TargetSets:   it.TargetSets,
			TargetReps:   it.TargetReps,
			Sets:         make([]sessionv1.SetLogRes, 0, len(it.Sets)),
		}
		for _, set := range it.Sets {
			item.Sets = append(item.Sets, toSetRes(set))
		}
		res.Items = append(res.Items, item)
	}
	return res
}

func (s *workoutSessionService) buildItemsFromWorkout(w *entity.Workout) []entity.WorkoutSessionItem {
	items := make([]entity.WorkoutSessionItem, 0, len(w.Items))
	for _, it := range w.Items {
		wid := it.ID
		items = append(items, entity.WorkoutSessionItem{
			WorkoutItemID: &wid,
			ExerciseID:    it.ExerciseID,
			ExerciseName:  it.ExerciseName,
			TargetSets:    it.Sets,
			TargetReps:    it.Reps,
			Sets:          seedSets(it.Sets),
		})
	}
	return items
}

func (s *workoutSessionService) Create(ctx context.Context, userID uint, req *sessionv1.CreateSessionReq) (*sessionv1.CreateSessionRes, error) {
	now := time.Now()
	sess := &entity.WorkoutSession{
		UserID:    userID,
		StartedAt: now,
		Status:    entity.SessionStatusInProgress,
		Notes:     strings.TrimSpace(req.Notes),
	}

	if req.WorkoutID != nil && *req.WorkoutID > 0 {
		w, err := s.workoutRepo.GetByID(ctx, *req.WorkoutID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrWorkoutNotFound
			}
			return nil, err
		}
		if w.OwnerUserID != nil && *w.OwnerUserID != userID {
			return nil, ErrForbiddenWorkout
		}
		wid := w.ID
		sess.WorkoutID = &wid
		sess.Title = w.Title
		if w.OwnerUserID == nil {
			sess.Source = entity.SessionSourceCatalog
		} else {
			sess.Source = entity.SessionSourcePersonal
		}
		if title := strings.TrimSpace(req.Title); title != "" {
			sess.Title = title
		}
		sess.Items = s.buildItemsFromWorkout(w)
	} else {
		sess.Source = entity.SessionSourceFreestyle
		sess.Title = strings.TrimSpace(req.Title)
		if sess.Title == "" {
			sess.Title = "Freestyle session"
		}
		items := make([]entity.WorkoutSessionItem, 0, len(req.Items))
		for _, in := range req.Items {
			ex, err := s.exerciseRepo.GetByID(ctx, in.ExerciseID)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil, ErrExerciseNotFound
				}
				return nil, err
			}
			nSets := normalizeSets(in.Sets)
			items = append(items, entity.WorkoutSessionItem{
				ExerciseID:   ex.ID,
				ExerciseName: ex.Name,
				TargetSets:   nSets,
				TargetReps:   normalizeReps(in.Reps),
				Sets:         seedSets(nSets),
			})
		}
		sess.Items = items
	}

	if err := s.repo.Create(ctx, sess); err != nil {
		return nil, err
	}
	fresh, err := s.repo.GetByID(ctx, userID, sess.ID)
	if err != nil {
		return nil, err
	}
	res := toSessionRes(fresh)
	res.Items = s.attachPrevious(ctx, userID, fresh, res.Items)
	return &sessionv1.CreateSessionRes{Session: res}, nil
}

func (s *workoutSessionService) CreateFromEnrollmentSlot(ctx context.Context, userID uint, slot *entity.EnrollmentSlot, workout *entity.Workout) (*entity.WorkoutSession, error) {
	slotID := slot.ID
	sess := &entity.WorkoutSession{
		UserID:           userID,
		EnrollmentSlotID: &slotID,
		Title:            slot.WorkoutTitle,
		Source:           entity.SessionSourceEnrollment,
		Status:           entity.SessionStatusInProgress,
		StartedAt:        time.Now(),
	}
	if workout != nil {
		wid := workout.ID
		sess.WorkoutID = &wid
		sess.Items = s.buildItemsFromWorkout(workout)
	}
	if err := s.repo.Create(ctx, sess); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, userID, sess.ID)
}

func (s *workoutSessionService) Get(ctx context.Context, userID, id uint) (*sessionv1.GetSessionRes, error) {
	sess, err := s.repo.GetByID(ctx, userID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}
	res := toSessionRes(sess)
	res.Items = s.attachPrevious(ctx, userID, sess, res.Items)
	return &sessionv1.GetSessionRes{Session: res}, nil
}

func (s *workoutSessionService) List(ctx context.Context, userID uint, req *sessionv1.ListSessionsReq) (*sessionv1.ListSessionsRes, error) {
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
	list, total, err := s.repo.List(ctx, userID, offset, limit, strings.TrimSpace(req.Status))
	if err != nil {
		return nil, err
	}
	data := make([]sessionv1.SessionRes, 0, len(list))
	for i := range list {
		data = append(data, toSessionRes(&list[i]))
	}
	return &sessionv1.ListSessionsRes{Total: total, Data: data}, nil
}

func (s *workoutSessionService) UpdateSet(ctx context.Context, userID, setID uint, req *sessionv1.UpdateSetReq) (*sessionv1.SetLogRes, error) {
	set, _, err := s.repo.GetSet(ctx, userID, setID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSetNotFound
		}
		return nil, err
	}
	updates := map[string]interface{}{}
	if req.WeightKg != nil {
		updates["weight_kg"] = *req.WeightKg
	}
	if req.Reps != nil {
		updates["reps"] = *req.Reps
	}
	if req.Completed != nil {
		updates["completed"] = *req.Completed
		if *req.Completed {
			updates["completed_at"] = time.Now()
		} else {
			updates["completed_at"] = nil
		}
	}
	if len(updates) > 0 {
		if err := s.repo.UpdateSetFields(ctx, set.ID, updates); err != nil {
			return nil, err
		}
	}
	fresh, _, err := s.repo.GetSet(ctx, userID, setID)
	if err != nil {
		return nil, err
	}
	res := toSetRes(*fresh)
	return &res, nil
}

func (s *workoutSessionService) AddItem(ctx context.Context, userID, sessionID uint, req *sessionv1.AddSessionItemReq) (*sessionv1.SessionItemRes, error) {
	sess, err := s.repo.GetByID(ctx, userID, sessionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}
	if sess.Status != entity.SessionStatusInProgress {
		return nil, errors.New("session is not in progress")
	}
	ex, err := s.exerciseRepo.GetByID(ctx, req.ExerciseID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrExerciseNotFound
		}
		return nil, err
	}
	nSets := normalizeSets(req.Sets)
	item := &entity.WorkoutSessionItem{
		SessionID:    sessionID,
		ExerciseID:   ex.ID,
		ExerciseName: ex.Name,
		SortOrder:    len(sess.Items) + 1,
		TargetSets:   nSets,
		TargetReps:   normalizeReps(req.Reps),
		Sets:         seedSets(nSets),
	}
	if err := s.repo.AddItem(ctx, item); err != nil {
		return nil, err
	}
	fresh, err := s.repo.GetByID(ctx, userID, sessionID)
	if err != nil {
		return nil, err
	}
	for _, it := range fresh.Items {
		if it.ID == item.ID {
			res := sessionv1.SessionItemRes{
				ID:           it.ID,
				ExerciseID:   it.ExerciseID,
				ExerciseName: it.ExerciseName,
				SortOrder:    it.SortOrder,
				TargetSets:   it.TargetSets,
				TargetReps:   it.TargetReps,
				Sets:         make([]sessionv1.SetLogRes, 0, len(it.Sets)),
			}
			for _, set := range it.Sets {
				res.Sets = append(res.Sets, toSetRes(set))
			}
			prev, err := s.repo.LastCompletedItemForExercise(ctx, userID, it.ExerciseID, sessionID)
			if err == nil && prev != nil {
				res.Previous = make([]sessionv1.SetLogRes, 0, len(prev.Sets))
				for _, set := range prev.Sets {
					res.Previous = append(res.Previous, toSetRes(set))
				}
			}
			return &res, nil
		}
	}
	return nil, ErrSessionNotFound
}

func (s *workoutSessionService) Complete(ctx context.Context, userID, id uint) (*sessionv1.GetSessionRes, error) {
	sess, err := s.repo.GetByID(ctx, userID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}
	// Auto-check sets that already have weight + reps (Finish without tapping ✓).
	now := time.Now()
	for i := range sess.Items {
		for j := range sess.Items[i].Sets {
			set := sess.Items[i].Sets[j]
			if set.WeightKg != nil && set.Reps != nil && !set.Completed {
				if err := s.repo.UpdateSetFields(ctx, set.ID, map[string]interface{}{
					"completed":    true,
					"completed_at": now,
				}); err != nil {
					return nil, err
				}
			}
		}
	}
	// Column-only update — never Save() a preloaded session (can wipe set kg/reps).
	if err := s.repo.UpdateSessionFields(ctx, sess.ID, map[string]interface{}{
		"status":       entity.SessionStatusCompleted,
		"completed_at": now,
	}); err != nil {
		return nil, err
	}
	return s.Get(ctx, userID, id)
}

func (s *workoutSessionService) Abandon(ctx context.Context, userID, id uint) (*sessionv1.GetSessionRes, error) {
	sess, err := s.repo.GetByID(ctx, userID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}
	if err := s.repo.UpdateSessionFields(ctx, sess.ID, map[string]interface{}{
		"status": entity.SessionStatusAbandoned,
	}); err != nil {
		return nil, err
	}
	return s.Get(ctx, userID, id)
}

func computeVolumeAndBest(sets []entity.WorkoutSetLog) (volume float64, bestW *float64, bestR *int) {
	for _, set := range sets {
		// Count logged work even if ✓ wasn't tapped (kg×reps present).
		if set.WeightKg == nil || set.Reps == nil {
			continue
		}
		volume += (*set.WeightKg) * float64(*set.Reps)
		if bestW == nil || *set.WeightKg > *bestW || (*set.WeightKg == *bestW && bestR != nil && *set.Reps > *bestR) {
			w := *set.WeightKg
			r := *set.Reps
			bestW = &w
			bestR = &r
		}
	}
	return
}

func (s *workoutSessionService) ExercisePerformance(ctx context.Context, userID, exerciseID uint, limit int) (*sessionv1.ExercisePerformanceRes, error) {
	ex, err := s.exerciseRepo.GetByID(ctx, exerciseID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrExerciseNotFound
		}
		return nil, err
	}
	items, err := s.repo.ListExerciseHistory(ctx, userID, exerciseID, limit)
	if err != nil {
		return nil, err
	}
	data := make([]sessionv1.PerformanceSessionRes, 0, len(items))
	for _, it := range items {
		vol, bw, br := computeVolumeAndBest(it.Sets)
		performed := it.CreatedAt
		sets := make([]sessionv1.SetLogRes, 0, len(it.Sets))
		for _, set := range it.Sets {
			sets = append(sets, toSetRes(set))
		}
		data = append(data, sessionv1.PerformanceSessionRes{
			SessionID:   it.SessionID,
			Title:       ex.Name,
			PerformedAt: performed,
			Volume:      vol,
			BestWeight:  bw,
			BestReps:    br,
			Sets:        sets,
		})
	}
	return &sessionv1.ExercisePerformanceRes{
		ExerciseID:   ex.ID,
		ExerciseName: ex.Name,
		Data:         data,
	}, nil
}
