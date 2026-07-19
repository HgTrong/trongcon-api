package service

import (
	"context"
	"errors"
	"strings"
	"time"

	enrollv1 "trongcon-api/api/training_enrollment/v1"
	sessionv1 "trongcon-api/api/workout_session/v1"
	"trongcon-api/internal/entity"
	"trongcon-api/internal/repository"

	"gorm.io/gorm"
)

var ErrEnrollmentNotFound = errors.New("training enrollment not found")
var ErrSlotNotFound = errors.New("enrollment slot not found")
var ErrSlotAlreadyStarted = errors.New("slot already has a session")

type TrainingEnrollmentService interface {
	Create(ctx context.Context, userID uint, req *enrollv1.CreateEnrollmentReq) (*enrollv1.CreateEnrollmentRes, error)
	Get(ctx context.Context, userID, id uint) (*enrollv1.GetEnrollmentRes, error)
	List(ctx context.Context, userID uint, req *enrollv1.ListEnrollmentsReq) (*enrollv1.ListEnrollmentsRes, error)
	Cancel(ctx context.Context, userID, id uint) (*enrollv1.GetEnrollmentRes, error)
	StartSlot(ctx context.Context, userID, slotID uint) (*sessionv1.CreateSessionRes, error)
	Compare(ctx context.Context, userID, id uint) (*enrollv1.EnrollmentCompareRes, error)
	ThisWeek(ctx context.Context, userID uint) (*enrollv1.GetEnrollmentRes, error)
}

type trainingEnrollmentService struct {
	repo        repository.TrainingEnrollmentRepository
	routineRepo repository.RoutineRepository
	workoutRepo repository.WorkoutRepository
	sessionRepo repository.WorkoutSessionRepository
	sessionSvc  WorkoutSessionService
}

func NewTrainingEnrollmentService(
	repo repository.TrainingEnrollmentRepository,
	routineRepo repository.RoutineRepository,
	workoutRepo repository.WorkoutRepository,
	sessionRepo repository.WorkoutSessionRepository,
	sessionSvc WorkoutSessionService,
) TrainingEnrollmentService {
	return &trainingEnrollmentService{
		repo:        repo,
		routineRepo: routineRepo,
		workoutRepo: workoutRepo,
		sessionRepo: sessionRepo,
		sessionSvc:  sessionSvc,
	}
}

func parseDateYYYYMMDD(s string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02", strings.TrimSpace(s), time.Local)
}

func toEnrollmentRes(e *entity.TrainingEnrollment, sessions map[uint]*entity.WorkoutSession) enrollv1.EnrollmentRes {
	res := enrollv1.EnrollmentRes{
		ID:        e.ID,
		RoutineID: e.RoutineID,
		Title:     e.Title,
		StartDate: e.StartDate,
		Weeks:     e.Weeks,
		Status:    e.Status,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
		Slots:     make([]enrollv1.SlotRes, 0, len(e.Slots)),
	}
	for _, slot := range e.Slots {
		sr := enrollv1.SlotRes{
			ID:           slot.ID,
			WeekIndex:    slot.WeekIndex,
			DayIndex:     slot.DayIndex,
			WorkoutID:    slot.WorkoutID,
			WorkoutTitle: slot.WorkoutTitle,
			SessionID:    slot.SessionID,
		}
		if slot.SessionID != nil && sessions != nil {
			if sess, ok := sessions[*slot.SessionID]; ok {
				sr.SessionStatus = sess.Status
				sr.CompletedAt = sess.CompletedAt
			}
		}
		res.Slots = append(res.Slots, sr)
	}
	return res
}

func (s *trainingEnrollmentService) loadSessionMap(ctx context.Context, userID uint, e *entity.TrainingEnrollment) map[uint]*entity.WorkoutSession {
	m := map[uint]*entity.WorkoutSession{}
	for _, slot := range e.Slots {
		if slot.SessionID == nil {
			continue
		}
		sess, err := s.sessionRepo.GetByID(ctx, userID, *slot.SessionID)
		if err == nil {
			m[*slot.SessionID] = sess
		}
	}
	return m
}

func (s *trainingEnrollmentService) Create(ctx context.Context, userID uint, req *enrollv1.CreateEnrollmentReq) (*enrollv1.CreateEnrollmentRes, error) {
	start, err := parseDateYYYYMMDD(req.StartDate)
	if err != nil {
		return nil, errors.New("invalid start_date, use YYYY-MM-DD")
	}
	if req.Weeks < 1 || req.Weeks > 52 {
		return nil, errors.New("weeks must be between 1 and 52")
	}

	hasRoutine := req.RoutineID != nil && *req.RoutineID > 0
	hasWorkout := req.WorkoutID != nil && *req.WorkoutID > 0
	if hasRoutine == hasWorkout {
		return nil, errors.New("provide either routine_id or workout_id")
	}

	en := &entity.TrainingEnrollment{
		UserID:    userID,
		StartDate: start,
		Weeks:     req.Weeks,
		Status:    entity.EnrollmentStatusActive,
	}
	var slots []entity.EnrollmentSlot

	if hasRoutine {
		rt, err := s.routineRepo.GetByID(ctx, *req.RoutineID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrRoutineNotFound
			}
			return nil, err
		}
		if rt.UserID != userID && !rt.IsPublic {
			return nil, ErrForbiddenRoutine
		}
		if len(rt.Items) == 0 {
			return nil, errors.New("routine has no workouts")
		}
		rid := rt.ID
		en.RoutineID = &rid
		en.Title = rt.Title
		slots = make([]entity.EnrollmentSlot, 0, req.Weeks*len(rt.Items))
		for week := 1; week <= req.Weeks; week++ {
			for dayIdx, item := range rt.Items {
				wid := item.WorkoutID
				slots = append(slots, entity.EnrollmentSlot{
					WeekIndex:    week,
					DayIndex:     dayIdx,
					WorkoutID:    &wid,
					WorkoutTitle: item.WorkoutTitle,
				})
			}
		}
	} else {
		w, err := s.workoutRepo.GetByID(ctx, *req.WorkoutID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrWorkoutNotFound
			}
			return nil, err
		}
		// Catalog or owned personal workout only
		if w.OwnerUserID != nil && *w.OwnerUserID != userID {
			return nil, ErrForbiddenWorkout
		}
		en.Title = w.Title
		wid := w.ID
		slots = make([]entity.EnrollmentSlot, 0, req.Weeks)
		for week := 1; week <= req.Weeks; week++ {
			slots = append(slots, entity.EnrollmentSlot{
				WeekIndex:    week,
				DayIndex:     0,
				WorkoutID:    &wid,
				WorkoutTitle: w.Title,
			})
		}
	}

	en.Slots = slots
	if err := s.repo.Create(ctx, en); err != nil {
		return nil, err
	}
	fresh, err := s.repo.GetByID(ctx, userID, en.ID)
	if err != nil {
		return nil, err
	}
	return &enrollv1.CreateEnrollmentRes{Enrollment: toEnrollmentRes(fresh, nil)}, nil
}

func (s *trainingEnrollmentService) Get(ctx context.Context, userID, id uint) (*enrollv1.GetEnrollmentRes, error) {
	e, err := s.repo.GetByID(ctx, userID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrEnrollmentNotFound
		}
		return nil, err
	}
	sm := s.loadSessionMap(ctx, userID, e)
	return &enrollv1.GetEnrollmentRes{Enrollment: toEnrollmentRes(e, sm)}, nil
}

func (s *trainingEnrollmentService) List(ctx context.Context, userID uint, req *enrollv1.ListEnrollmentsReq) (*enrollv1.ListEnrollmentsRes, error) {
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
	data := make([]enrollv1.EnrollmentRes, 0, len(list))
	for i := range list {
		sm := s.loadSessionMap(ctx, userID, &list[i])
		data = append(data, toEnrollmentRes(&list[i], sm))
	}
	return &enrollv1.ListEnrollmentsRes{Total: total, Data: data}, nil
}

func (s *trainingEnrollmentService) Cancel(ctx context.Context, userID, id uint) (*enrollv1.GetEnrollmentRes, error) {
	e, err := s.repo.GetByID(ctx, userID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrEnrollmentNotFound
		}
		return nil, err
	}
	e.Status = entity.EnrollmentStatusCancelled
	if err := s.repo.Update(ctx, e); err != nil {
		return nil, err
	}
	return s.Get(ctx, userID, id)
}

func (s *trainingEnrollmentService) StartSlot(ctx context.Context, userID, slotID uint) (*sessionv1.CreateSessionRes, error) {
	slot, en, err := s.repo.GetSlot(ctx, userID, slotID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSlotNotFound
		}
		return nil, err
	}
	if en.Status != entity.EnrollmentStatusActive {
		return nil, errors.New("enrollment is not active")
	}
	if slot.SessionID != nil {
		return nil, ErrSlotAlreadyStarted
	}
	var workout *entity.Workout
	if slot.WorkoutID != nil {
		w, err := s.workoutRepo.GetByID(ctx, *slot.WorkoutID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		if err == nil {
			workout = w
		}
	}
	sess, err := s.sessionSvc.CreateFromEnrollmentSlot(ctx, userID, slot, workout)
	if err != nil {
		return nil, err
	}
	sid := sess.ID
	slot.SessionID = &sid
	if err := s.repo.UpdateSlot(ctx, slot); err != nil {
		return nil, err
	}
	getRes, err := s.sessionSvc.Get(ctx, userID, sess.ID)
	if err != nil {
		return nil, err
	}
	return &sessionv1.CreateSessionRes{Session: getRes.Session}, nil
}

func (s *trainingEnrollmentService) Compare(ctx context.Context, userID, id uint) (*enrollv1.EnrollmentCompareRes, error) {
	e, err := s.repo.GetByID(ctx, userID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrEnrollmentNotFound
		}
		return nil, err
	}
	sessions, err := s.repo.ListCompletedSessionsForEnrollment(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	sessionByID := map[uint]*entity.WorkoutSession{}
	for i := range sessions {
		sessionByID[sessions[i].ID] = &sessions[i]
	}

	// week -> exercise_id -> stats
	type accum struct {
		name   string
		volume float64
		bestW  *float64
		bestR  *int
	}
	byWeek := map[int]map[uint]*accum{}
	for _, slot := range e.Slots {
		if slot.SessionID == nil {
			continue
		}
		sess, ok := sessionByID[*slot.SessionID]
		if !ok {
			continue
		}
		if byWeek[slot.WeekIndex] == nil {
			byWeek[slot.WeekIndex] = map[uint]*accum{}
		}
		for _, item := range sess.Items {
			vol, bw, br := computeVolumeAndBest(item.Sets)
			a := byWeek[slot.WeekIndex][item.ExerciseID]
			if a == nil {
				a = &accum{name: item.ExerciseName}
				byWeek[slot.WeekIndex][item.ExerciseID] = a
			}
			a.volume += vol
			if bw != nil && (a.bestW == nil || *bw > *a.bestW || (*bw == *a.bestW && br != nil && a.bestR != nil && *br > *a.bestR) || (a.bestR == nil && br != nil)) {
				w := *bw
				a.bestW = &w
				if br != nil {
					r := *br
					a.bestR = &r
				}
			}
		}
	}

	weeks := make([]enrollv1.WeekCompareRes, 0, e.Weeks)
	var prev map[uint]*accum
	for week := 1; week <= e.Weeks; week++ {
		cur := byWeek[week]
		if cur == nil {
			cur = map[uint]*accum{}
		}
		wr := enrollv1.WeekCompareRes{
			WeekIndex:  week,
			Exercises:  make([]enrollv1.WeekExerciseStat, 0, len(cur)),
		}
		for eid, a := range cur {
			st := enrollv1.WeekExerciseStat{
				ExerciseID:   eid,
				ExerciseName: a.name,
				Volume:       a.volume,
				BestWeight:   a.bestW,
				BestReps:     a.bestR,
			}
			if prev != nil {
				if p, ok := prev[eid]; ok {
					d := a.volume - p.volume
					st.DeltaVolume = &d
				}
			}
			wr.TotalVolume += a.volume
			wr.Exercises = append(wr.Exercises, st)
		}
		weeks = append(weeks, wr)
		prev = cur
	}

	return &enrollv1.EnrollmentCompareRes{EnrollmentID: e.ID, Weeks: weeks}, nil
}

func (s *trainingEnrollmentService) ThisWeek(ctx context.Context, userID uint) (*enrollv1.GetEnrollmentRes, error) {
	list, _, err := s.repo.List(ctx, userID, 0, 20, entity.EnrollmentStatusActive)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, ErrEnrollmentNotFound
	}
	// Prefer earliest start still covering today
	today := time.Now().Truncate(24 * time.Hour)
	var best *entity.TrainingEnrollment
	for i := range list {
		e := &list[i]
		end := e.StartDate.AddDate(0, 0, e.Weeks*7)
		if !today.Before(e.StartDate) && today.Before(end) {
			best = e
			break
		}
		if best == nil {
			best = e
		}
	}
	if best == nil {
		return nil, ErrEnrollmentNotFound
	}
	return s.Get(ctx, userID, best.ID)
}
