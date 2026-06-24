package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	exercisev1 "trongcon-api/api/exercise/v1"
	"trongcon-api/internal/apimap"
	"trongcon-api/internal/entity"
	"trongcon-api/internal/pkg/slug"
	"trongcon-api/internal/repository"

	"gorm.io/gorm"
)

var ErrExerciseNotFound = errors.New("exercise not found")

type ExerciseService interface {
	Create(ctx context.Context, req *exercisev1.CreateReq) (*exercisev1.CreateRes, error)
	GetByID(ctx context.Context, id uint) (*exercisev1.GetRes, error)
	Update(ctx context.Context, id uint, req *exercisev1.UpdateReq) (*exercisev1.UpdateRes, error)
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, req *exercisev1.ListReq) (*exercisev1.ListRes, error)
}

type exerciseService struct {
	repo          repository.ExerciseRepository
	equipmentRepo repository.EquipmentRepository
	muscleRepo    repository.MuscleRepository
}

func NewExerciseService(
	repo repository.ExerciseRepository,
	equipmentRepo repository.EquipmentRepository,
	muscleRepo repository.MuscleRepository,
) ExerciseService {
	return &exerciseService{repo: repo, equipmentRepo: equipmentRepo, muscleRepo: muscleRepo}
}

func (s *exerciseService) allocateUniqueSlug(ctx context.Context, base string, excludeID uint) (string, error) {
	if base == "" {
		base = "exercise"
	}
	for n := 0; n < 10000; n++ {
		candidate := base
		if n > 0 {
			candidate = fmt.Sprintf("%s-%d", base, n)
		}
		exists, err := s.repo.SlugExists(ctx, candidate, excludeID)
		if err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
	}
	return "", errors.New("could not allocate unique slug")
}

func normalizeExerciseStatus(status string) string {
	if strings.TrimSpace(status) == "" {
		return "active"
	}
	return status
}

func buildExerciseSteps(inputs []exercisev1.StepInput) []entity.ExerciseStep {
	steps := make([]entity.ExerciseStep, 0, len(inputs))
	for i, in := range inputs {
		content := strings.TrimSpace(in.Content)
		if content == "" {
			continue
		}
		steps = append(steps, entity.ExerciseStep{
			SortOrder: i + 1,
			Content:   content,
		})
	}
	return steps
}

func uniqueUints(ids []uint) []uint {
	seen := make(map[uint]struct{}, len(ids))
	out := make([]uint, 0, len(ids))
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func buildExerciseMuscles(primary, secondary, tertiary []uint) []entity.ExerciseMuscle {
	used := make(map[uint]string)
	var muscles []entity.ExerciseMuscle

	add := func(ids []uint, role string) {
		for _, id := range uniqueUints(ids) {
			if prev, ok := used[id]; ok {
				if rolePriority(role) <= rolePriority(prev) {
					continue
				}
				muscles = removeMuscle(muscles, id)
			}
			used[id] = role
			muscles = append(muscles, entity.ExerciseMuscle{MuscleID: id, Role: role})
		}
	}
	add(primary, "primary")
	add(secondary, "secondary")
	add(tertiary, "tertiary")
	return muscles
}

func muscleIDsByRole(muscles []entity.ExerciseMuscle, role string) []uint {
	out := make([]uint, 0)
	for _, m := range muscles {
		if m.Role == role {
			out = append(out, m.MuscleID)
		}
	}
	return out
}

func rolePriority(role string) int {
	switch role {
	case "primary":
		return 1
	case "secondary":
		return 2
	case "tertiary":
		return 3
	default:
		return 99
	}
}

func removeMuscle(list []entity.ExerciseMuscle, muscleID uint) []entity.ExerciseMuscle {
	out := list[:0]
	for _, m := range list {
		if m.MuscleID != muscleID {
			out = append(out, m)
		}
	}
	return out
}

func (s *exerciseService) validateEquipment(ctx context.Context, equipmentID *uint) error {
	if equipmentID == nil || *equipmentID == 0 {
		return nil
	}
	if _, err := s.equipmentRepo.GetByID(ctx, *equipmentID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrEquipmentNotFound
		}
		return err
	}
	return nil
}

func (s *exerciseService) validateMuscleIDs(ctx context.Context, ids []uint) error {
	for _, id := range uniqueUints(ids) {
		if _, err := s.muscleRepo.GetByID(ctx, id); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrMuscleNotFound
			}
			return err
		}
	}
	return nil
}

func (s *exerciseService) validateAllMuscles(ctx context.Context, primary, secondary, tertiary []uint) error {
	all := append(append([]uint{}, primary...), secondary...)
	all = append(all, tertiary...)
	return s.validateMuscleIDs(ctx, all)
}

func (s *exerciseService) Create(ctx context.Context, req *exercisev1.CreateReq) (*exercisev1.CreateRes, error) {
	if err := s.validateEquipment(ctx, req.EquipmentID); err != nil {
		return nil, err
	}
	if err := s.validateAllMuscles(ctx, req.PrimaryMuscleIDs, req.SecondaryMuscleIDs, req.TertiaryMuscleIDs); err != nil {
		return nil, err
	}

	slugVal, err := s.allocateUniqueSlug(ctx, slug.FromTitle(req.Name), 0)
	if err != nil {
		return nil, err
	}

	ex := &entity.Exercise{
		Name:        req.Name,
		Slug:        slugVal,
		Summary:     req.Summary,
		Difficulty:  req.Difficulty,
		Force:       req.Force,
		Grips:       req.Grips,
		Mechanic:    req.Mechanic,
		DemoGif1:    req.DemoGif1,
		DemoGif2:    req.DemoGif2,
		VideoURL:    req.VideoURL,
		Thumbnail:   req.Thumbnail,
		Content:     req.Content,
		Status:      normalizeExerciseStatus(req.Status),
		EquipmentID: req.EquipmentID,
	}
	if err := s.repo.Create(ctx, ex); err != nil {
		return nil, err
	}
	if err := s.repo.ReplaceSteps(ctx, ex.ID, buildExerciseSteps(req.Steps)); err != nil {
		return nil, err
	}
	muscles := buildExerciseMuscles(req.PrimaryMuscleIDs, req.SecondaryMuscleIDs, req.TertiaryMuscleIDs)
	if err := s.repo.ReplaceMuscles(ctx, ex.ID, muscles); err != nil {
		return nil, err
	}
	fresh, err := s.repo.GetByID(ctx, ex.ID)
	if err != nil {
		return nil, err
	}
	return &exercisev1.CreateRes{Exercise: apimap.ExerciseToRes(fresh)}, nil
}

func (s *exerciseService) GetByID(ctx context.Context, id uint) (*exercisev1.GetRes, error) {
	ex, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrExerciseNotFound
		}
		return nil, err
	}
	return &exercisev1.GetRes{Exercise: apimap.ExerciseToRes(ex)}, nil
}

func (s *exerciseService) Update(ctx context.Context, id uint, req *exercisev1.UpdateReq) (*exercisev1.UpdateRes, error) {
	ex, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrExerciseNotFound
		}
		return nil, err
	}
	if req.EquipmentID != nil {
		if err := s.validateEquipment(ctx, req.EquipmentID); err != nil {
			return nil, err
		}
		ex.EquipmentID = req.EquipmentID
	}
	if req.Name != nil && strings.TrimSpace(*req.Name) != "" {
		ex.Name = *req.Name
		slugVal, err := s.allocateUniqueSlug(ctx, slug.FromTitle(*req.Name), ex.ID)
		if err != nil {
			return nil, err
		}
		ex.Slug = slugVal
	}
	if req.Summary != nil {
		ex.Summary = *req.Summary
	}
	if req.Difficulty != nil {
		ex.Difficulty = *req.Difficulty
	}
	if req.Force != nil {
		ex.Force = *req.Force
	}
	if req.Grips != nil {
		ex.Grips = *req.Grips
	}
	if req.Mechanic != nil {
		ex.Mechanic = *req.Mechanic
	}
	if req.DemoGif1 != nil {
		ex.DemoGif1 = *req.DemoGif1
	}
	if req.DemoGif2 != nil {
		ex.DemoGif2 = *req.DemoGif2
	}
	if req.VideoURL != nil {
		ex.VideoURL = *req.VideoURL
	}
	if req.Thumbnail != nil {
		ex.Thumbnail = *req.Thumbnail
	}
	if req.Content != nil {
		ex.Content = *req.Content
	}
	if req.Status != nil {
		ex.Status = normalizeExerciseStatus(*req.Status)
	}
	if err := s.repo.Update(ctx, ex); err != nil {
		return nil, err
	}
	if req.Steps != nil {
		if err := s.repo.ReplaceSteps(ctx, ex.ID, buildExerciseSteps(*req.Steps)); err != nil {
			return nil, err
		}
	}
	if req.PrimaryMuscleIDs != nil || req.SecondaryMuscleIDs != nil || req.TertiaryMuscleIDs != nil {
		primary := muscleIDsByRole(ex.Muscles, "primary")
		secondary := muscleIDsByRole(ex.Muscles, "secondary")
		tertiary := muscleIDsByRole(ex.Muscles, "tertiary")
		if req.PrimaryMuscleIDs != nil {
			primary = *req.PrimaryMuscleIDs
		}
		if req.SecondaryMuscleIDs != nil {
			secondary = *req.SecondaryMuscleIDs
		}
		if req.TertiaryMuscleIDs != nil {
			tertiary = *req.TertiaryMuscleIDs
		}
		if err := s.validateAllMuscles(ctx, primary, secondary, tertiary); err != nil {
			return nil, err
		}
		if err := s.repo.ReplaceMuscles(ctx, ex.ID, buildExerciseMuscles(primary, secondary, tertiary)); err != nil {
			return nil, err
		}
	}
	fresh, err := s.repo.GetByID(ctx, ex.ID)
	if err != nil {
		return nil, err
	}
	return &exercisev1.UpdateRes{Exercise: apimap.ExerciseToRes(fresh)}, nil
}

func (s *exerciseService) Delete(ctx context.Context, id uint) error {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrExerciseNotFound
		}
		return err
	}
	return s.repo.Delete(ctx, id)
}

func (s *exerciseService) List(ctx context.Context, req *exercisev1.ListReq) (*exercisev1.ListRes, error) {
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
	case "id", "name", "created_at", "difficulty":
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
		strings.TrimSpace(req.Force),
		strings.TrimSpace(req.Mechanic),
		strings.TrimSpace(req.Status),
		req.EquipmentID,
		req.MuscleID,
	)
	if err != nil {
		return nil, err
	}
	data := make([]exercisev1.ExerciseRes, 0, len(list))
	for i := range list {
		data = append(data, apimap.ExerciseToRes(&list[i]))
	}
	return &exercisev1.ListRes{Total: total, Data: data}, nil
}
