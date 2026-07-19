package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	authorv1 "trongcon-api/api/author/v1"
	gymv1 "trongcon-api/api/gym/v1"
	mealplanv1 "trongcon-api/api/meal_plan/v1"
	routinev1 "trongcon-api/api/routine/v1"
	workoutv1 "trongcon-api/api/workout/v1"
	"trongcon-api/internal/entity"
	"trongcon-api/internal/pkg/slug"
	"trongcon-api/internal/repository"

	"gorm.io/gorm"
)

var ErrBranchNotFound = errors.New("branch not found")
var ErrTrainerNotFound = errors.New("trainer profile not found")
var ErrTrainerExists = errors.New("user already has a trainer profile")

type GymService interface {
	CreateBranch(ctx context.Context, req *gymv1.CreateBranchReq) (*gymv1.CreateBranchRes, error)
	GetBranch(ctx context.Context, id uint) (*gymv1.GetBranchRes, error)
	GetBranchPublic(ctx context.Context, id uint) (*gymv1.GetBranchRes, error)
	UpdateBranch(ctx context.Context, id uint, req *gymv1.UpdateBranchReq) (*gymv1.UpdateBranchRes, error)
	DeleteBranch(ctx context.Context, id uint) error
	ListBranches(ctx context.Context, req *gymv1.ListBranchReq) (*gymv1.ListBranchRes, error)
	ListBranchesPublic(ctx context.Context, req *gymv1.ListBranchReq) (*gymv1.ListBranchRes, error)

	CreateTrainer(ctx context.Context, req *gymv1.CreateTrainerReq) (*gymv1.CreateTrainerRes, error)
	GetTrainer(ctx context.Context, id uint) (*gymv1.GetTrainerRes, error)
	GetTrainerPublic(ctx context.Context, id uint) (*gymv1.GetTrainerRes, error)
	UpdateTrainer(ctx context.Context, id uint, req *gymv1.UpdateTrainerReq) (*gymv1.UpdateTrainerRes, error)
	DeleteTrainer(ctx context.Context, id uint) error
	ListTrainers(ctx context.Context, req *gymv1.ListTrainerReq) (*gymv1.ListTrainerRes, error)
	ListTrainersPublic(ctx context.Context, req *gymv1.ListTrainerReq) (*gymv1.ListTrainerRes, error)

	ListTrainerWorkouts(ctx context.Context, trainerID uint) (*workoutv1.ListRes, error)
	ListTrainerRoutines(ctx context.Context, trainerID uint) (*routinev1.ListRes, error)
	ListTrainerMealPlans(ctx context.Context, trainerID uint) (*mealplanv1.ListRes, error)
}

type gymService struct {
	branchRepo   repository.GymBranchRepository
	trainerRepo  repository.TrainerProfileRepository
	userRepo     repository.UserRepository
	roleRepo     repository.RoleRepository
	workoutRepo  repository.WorkoutRepository
	routineRepo  repository.RoutineRepository
	mealPlanRepo repository.MealPlanRepository
}

func NewGymService(
	branchRepo repository.GymBranchRepository,
	trainerRepo repository.TrainerProfileRepository,
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
	workoutRepo repository.WorkoutRepository,
	routineRepo repository.RoutineRepository,
	mealPlanRepo repository.MealPlanRepository,
) GymService {
	return &gymService{
		branchRepo:   branchRepo,
		trainerRepo:  trainerRepo,
		userRepo:     userRepo,
		roleRepo:     roleRepo,
		workoutRepo:  workoutRepo,
		routineRepo:  routineRepo,
		mealPlanRepo: mealPlanRepo,
	}
}

// trainerAuthor builds the author badge attached to a trainer's public content.
func trainerAuthor(t *entity.TrainerProfile) *authorv1.AuthorRes {
	return &authorv1.AuthorRes{
		TrainerID:   t.ID,
		UserID:      t.UserID,
		DisplayName: t.DisplayName,
		AvatarURL:   t.User.ProfilePicture,
		Title:       t.Title,
	}
}

func (s *gymService) publicTrainer(ctx context.Context, trainerID uint) (*entity.TrainerProfile, error) {
	t, err := s.trainerRepo.GetByID(ctx, trainerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTrainerNotFound
		}
		return nil, err
	}
	if !t.IsPublic {
		return nil, ErrTrainerNotFound
	}
	return t, nil
}

func (s *gymService) ListTrainerWorkouts(ctx context.Context, trainerID uint) (*workoutv1.ListRes, error) {
	t, err := s.publicTrainer(ctx, trainerID)
	if err != nil {
		return nil, err
	}
	author := trainerAuthor(t)
	list, total, err := s.workoutRepo.ListPublicByOwner(ctx, t.UserID, 0, 100, "id DESC")
	if err != nil {
		return nil, err
	}
	data := make([]workoutv1.WorkoutRes, 0, len(list))
	for i := range list {
		res := toWorkoutAPIRes(&list[i])
		res.Author = author
		data = append(data, res)
	}
	return &workoutv1.ListRes{Total: total, Data: data}, nil
}

func (s *gymService) ListTrainerRoutines(ctx context.Context, trainerID uint) (*routinev1.ListRes, error) {
	t, err := s.publicTrainer(ctx, trainerID)
	if err != nil {
		return nil, err
	}
	author := trainerAuthor(t)
	uid := t.UserID
	pub := true
	list, total, err := s.routineRepo.List(ctx, 0, 100, "id DESC", "", "", &uid, &pub)
	if err != nil {
		return nil, err
	}
	data := make([]routinev1.RoutineRes, 0, len(list))
	for i := range list {
		res := toRoutineRes(&list[i])
		res.Author = author
		data = append(data, res)
	}
	return &routinev1.ListRes{Total: total, Data: data}, nil
}

func (s *gymService) ListTrainerMealPlans(ctx context.Context, trainerID uint) (*mealplanv1.ListRes, error) {
	t, err := s.publicTrainer(ctx, trainerID)
	if err != nil {
		return nil, err
	}
	author := trainerAuthor(t)
	uid := t.UserID
	pub := true
	list, total, err := s.mealPlanRepo.List(ctx, 0, 100, "id DESC", "", &uid, &pub)
	if err != nil {
		return nil, err
	}
	data := make([]mealplanv1.MealPlanRes, 0, len(list))
	for i := range list {
		res := toMealPlanRes(&list[i])
		res.Author = author
		data = append(data, res)
	}
	return &mealplanv1.ListRes{Total: total, Data: data}, nil
}

func toBranchRes(b *entity.GymBranch) gymv1.BranchRes {
	return gymv1.BranchRes{
		ID: b.ID, Name: b.Name, Slug: b.Slug, Address: b.Address, City: b.City,
		Phone: b.Phone, Email: b.Email, Hours: b.Hours, Description: b.Description,
		ImageURL: b.ImageURL, IsActive: b.IsActive, SortOrder: b.SortOrder,
		CreatedAt: b.CreatedAt, UpdatedAt: b.UpdatedAt,
	}
}

func toTrainerRes(t *entity.TrainerProfile) gymv1.TrainerRes {
	res := gymv1.TrainerRes{
		ID: t.ID, UserID: t.UserID, BranchID: t.BranchID,
		DisplayName: t.DisplayName, Title: t.Title, Bio: t.Bio,
		Specialties: t.Specialties, Certifications: t.Certifications,
		YearsExperience: t.YearsExperience, IsPublic: t.IsPublic,
		CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt,
	}
	if t.User.ID > 0 {
		res.Email = t.User.Email
		res.AvatarURL = t.User.ProfilePicture
	}
	if t.Branch != nil {
		res.BranchName = t.Branch.Name
		res.BranchSlug = t.Branch.Slug
	}
	return res
}

func parseActivePtr(s string) *bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
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

func (s *gymService) allocateBranchSlug(ctx context.Context, base string, excludeID uint) (string, error) {
	if base == "" {
		base = "branch"
	}
	for n := 0; n < 10000; n++ {
		candidate := base
		if n > 0 {
			candidate = fmt.Sprintf("%s-%d", base, n)
		}
		exists, err := s.branchRepo.SlugExists(ctx, candidate, excludeID)
		if err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
	}
	return "", errors.New("could not allocate unique slug")
}

func (s *gymService) CreateBranch(ctx context.Context, req *gymv1.CreateBranchReq) (*gymv1.CreateBranchRes, error) {
	slugVal, err := s.allocateBranchSlug(ctx, slug.FromTitle(req.Name), 0)
	if err != nil {
		return nil, err
	}
	active := true
	if req.IsActive != nil {
		active = *req.IsActive
	}
	b := &entity.GymBranch{
		Name: strings.TrimSpace(req.Name), Slug: slugVal,
		Address: strings.TrimSpace(req.Address), City: strings.TrimSpace(req.City),
		Phone: strings.TrimSpace(req.Phone), Email: strings.TrimSpace(req.Email),
		Hours: strings.TrimSpace(req.Hours), Description: strings.TrimSpace(req.Description),
		ImageURL: strings.TrimSpace(req.ImageURL), IsActive: active, SortOrder: req.SortOrder,
	}
	if err := s.branchRepo.Create(ctx, b); err != nil {
		return nil, err
	}
	fresh, err := s.branchRepo.GetByID(ctx, b.ID)
	if err != nil {
		return nil, err
	}
	return &gymv1.CreateBranchRes{Branch: toBranchRes(fresh)}, nil
}

func (s *gymService) GetBranch(ctx context.Context, id uint) (*gymv1.GetBranchRes, error) {
	b, err := s.branchRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrBranchNotFound
		}
		return nil, err
	}
	return &gymv1.GetBranchRes{Branch: toBranchRes(b)}, nil
}

func (s *gymService) GetBranchPublic(ctx context.Context, id uint) (*gymv1.GetBranchRes, error) {
	res, err := s.GetBranch(ctx, id)
	if err != nil {
		return nil, err
	}
	if !res.Branch.IsActive {
		return nil, ErrBranchNotFound
	}
	return res, nil
}

func (s *gymService) UpdateBranch(ctx context.Context, id uint, req *gymv1.UpdateBranchReq) (*gymv1.UpdateBranchRes, error) {
	b, err := s.branchRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrBranchNotFound
		}
		return nil, err
	}
	if req.Name != nil && strings.TrimSpace(*req.Name) != "" {
		b.Name = strings.TrimSpace(*req.Name)
		slugVal, err := s.allocateBranchSlug(ctx, slug.FromTitle(b.Name), b.ID)
		if err != nil {
			return nil, err
		}
		b.Slug = slugVal
	}
	if req.Address != nil {
		b.Address = strings.TrimSpace(*req.Address)
	}
	if req.City != nil {
		b.City = strings.TrimSpace(*req.City)
	}
	if req.Phone != nil {
		b.Phone = strings.TrimSpace(*req.Phone)
	}
	if req.Email != nil {
		b.Email = strings.TrimSpace(*req.Email)
	}
	if req.Hours != nil {
		b.Hours = strings.TrimSpace(*req.Hours)
	}
	if req.Description != nil {
		b.Description = strings.TrimSpace(*req.Description)
	}
	if req.ImageURL != nil {
		b.ImageURL = strings.TrimSpace(*req.ImageURL)
	}
	if req.IsActive != nil {
		b.IsActive = *req.IsActive
	}
	if req.SortOrder != nil {
		b.SortOrder = *req.SortOrder
	}
	if err := s.branchRepo.Update(ctx, b); err != nil {
		return nil, err
	}
	fresh, err := s.branchRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &gymv1.UpdateBranchRes{Branch: toBranchRes(fresh)}, nil
}

func (s *gymService) DeleteBranch(ctx context.Context, id uint) error {
	if _, err := s.branchRepo.GetByID(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrBranchNotFound
		}
		return err
	}
	return s.branchRepo.Delete(ctx, id)
}

func (s *gymService) listBranches(ctx context.Context, req *gymv1.ListBranchReq, forceActive *bool) (*gymv1.ListBranchRes, error) {
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
	orderBy := strings.ToLower(strings.TrimSpace(req.OrderBy))
	if orderBy == "" {
		orderBy = "sort_order"
	}
	switch orderBy {
	case "id", "name", "city", "sort_order", "created_at":
	default:
		orderBy = "sort_order"
	}
	dir := strings.ToUpper(strings.TrimSpace(req.OrderDir))
	if dir != "ASC" && dir != "DESC" {
		dir = "ASC"
	}
	order := orderBy + " " + dir
	if orderBy == "sort_order" {
		order = "sort_order ASC, id ASC"
	}
	active := forceActive
	if active == nil {
		active = parseActivePtr(req.Active)
	}
	list, total, err := s.branchRepo.List(ctx, (page-1)*limit, limit, order, req.Q, req.City, active)
	if err != nil {
		return nil, err
	}
	data := make([]gymv1.BranchRes, 0, len(list))
	for i := range list {
		data = append(data, toBranchRes(&list[i]))
	}
	return &gymv1.ListBranchRes{Total: total, Data: data}, nil
}

func (s *gymService) ListBranches(ctx context.Context, req *gymv1.ListBranchReq) (*gymv1.ListBranchRes, error) {
	return s.listBranches(ctx, req, nil)
}

func (s *gymService) ListBranchesPublic(ctx context.Context, req *gymv1.ListBranchReq) (*gymv1.ListBranchRes, error) {
	active := true
	return s.listBranches(ctx, req, &active)
}

func (s *gymService) ensurePTRole(ctx context.Context, user *entity.User) error {
	for _, r := range user.Roles {
		if r.Name == entity.RolePT {
			return nil
		}
	}
	role, err := s.roleRepo.GetByName(ctx, entity.RolePT)
	if err != nil {
		return err
	}
	return s.userRepo.AppendRole(ctx, user, role)
}

func (s *gymService) CreateTrainer(ctx context.Context, req *gymv1.CreateTrainerReq) (*gymv1.CreateTrainerRes, error) {
	user, err := s.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	if _, err := s.trainerRepo.GetByUserID(ctx, req.UserID); err == nil {
		return nil, ErrTrainerExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if req.BranchID != nil && *req.BranchID > 0 {
		if _, err := s.branchRepo.GetByID(ctx, *req.BranchID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrBranchNotFound
			}
			return nil, err
		}
	}
	if err := s.ensurePTRole(ctx, user); err != nil {
		return nil, err
	}
	isPublic := true
	if req.IsPublic != nil {
		isPublic = *req.IsPublic
	}
	display := strings.TrimSpace(req.DisplayName)
	if display == "" {
		display = strings.TrimSpace(user.Name)
	}
	if display == "" {
		display = user.Email
	}
	t := &entity.TrainerProfile{
		UserID: req.UserID, BranchID: req.BranchID,
		DisplayName: display, Title: strings.TrimSpace(req.Title),
		Bio: strings.TrimSpace(req.Bio), Specialties: strings.TrimSpace(req.Specialties),
		Certifications: strings.TrimSpace(req.Certifications),
		YearsExperience: req.YearsExperience, IsPublic: isPublic,
	}
	if t.BranchID != nil && *t.BranchID == 0 {
		t.BranchID = nil
	}
	if err := s.trainerRepo.Create(ctx, t); err != nil {
		return nil, err
	}
	fresh, err := s.trainerRepo.GetByID(ctx, t.ID)
	if err != nil {
		return nil, err
	}
	return &gymv1.CreateTrainerRes{Trainer: toTrainerRes(fresh)}, nil
}

func (s *gymService) GetTrainer(ctx context.Context, id uint) (*gymv1.GetTrainerRes, error) {
	t, err := s.trainerRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTrainerNotFound
		}
		return nil, err
	}
	return &gymv1.GetTrainerRes{Trainer: toTrainerRes(t)}, nil
}

func (s *gymService) GetTrainerPublic(ctx context.Context, id uint) (*gymv1.GetTrainerRes, error) {
	res, err := s.GetTrainer(ctx, id)
	if err != nil {
		return nil, err
	}
	if !res.Trainer.IsPublic {
		return nil, ErrTrainerNotFound
	}
	return res, nil
}

func (s *gymService) UpdateTrainer(ctx context.Context, id uint, req *gymv1.UpdateTrainerReq) (*gymv1.UpdateTrainerRes, error) {
	t, err := s.trainerRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTrainerNotFound
		}
		return nil, err
	}
	if req.ClearBranch {
		t.BranchID = nil
	} else if req.BranchID != nil {
		if *req.BranchID == 0 {
			t.BranchID = nil
		} else {
			if _, err := s.branchRepo.GetByID(ctx, *req.BranchID); err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil, ErrBranchNotFound
				}
				return nil, err
			}
			t.BranchID = req.BranchID
		}
	}
	if req.DisplayName != nil && strings.TrimSpace(*req.DisplayName) != "" {
		t.DisplayName = strings.TrimSpace(*req.DisplayName)
	}
	if req.Title != nil {
		t.Title = strings.TrimSpace(*req.Title)
	}
	if req.Bio != nil {
		t.Bio = strings.TrimSpace(*req.Bio)
	}
	if req.Specialties != nil {
		t.Specialties = strings.TrimSpace(*req.Specialties)
	}
	if req.Certifications != nil {
		t.Certifications = strings.TrimSpace(*req.Certifications)
	}
	if req.YearsExperience != nil {
		t.YearsExperience = *req.YearsExperience
	}
	if req.IsPublic != nil {
		t.IsPublic = *req.IsPublic
	}
	if err := s.trainerRepo.Update(ctx, t); err != nil {
		return nil, err
	}
	fresh, err := s.trainerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &gymv1.UpdateTrainerRes{Trainer: toTrainerRes(fresh)}, nil
}

func (s *gymService) DeleteTrainer(ctx context.Context, id uint) error {
	if _, err := s.trainerRepo.GetByID(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTrainerNotFound
		}
		return err
	}
	return s.trainerRepo.Delete(ctx, id)
}

func (s *gymService) listTrainers(ctx context.Context, req *gymv1.ListTrainerReq, forcePublic *bool) (*gymv1.ListTrainerRes, error) {
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
	orderBy := strings.ToLower(strings.TrimSpace(req.OrderBy))
	if orderBy == "" {
		orderBy = "id"
	}
	switch orderBy {
	case "id", "display_name", "years_experience", "created_at":
	default:
		orderBy = "id"
	}
	dir := strings.ToUpper(strings.TrimSpace(req.OrderDir))
	if dir != "ASC" && dir != "DESC" {
		dir = "DESC"
	}
	order := orderBy + " " + dir
	pub := forcePublic
	if pub == nil {
		pub = parseActivePtr(req.Public)
	}
	list, total, err := s.trainerRepo.List(ctx, (page-1)*limit, limit, order, req.Q, req.BranchID, pub)
	if err != nil {
		return nil, err
	}
	data := make([]gymv1.TrainerRes, 0, len(list))
	for i := range list {
		data = append(data, toTrainerRes(&list[i]))
	}
	return &gymv1.ListTrainerRes{Total: total, Data: data}, nil
}

func (s *gymService) ListTrainers(ctx context.Context, req *gymv1.ListTrainerReq) (*gymv1.ListTrainerRes, error) {
	return s.listTrainers(ctx, req, nil)
}

func (s *gymService) ListTrainersPublic(ctx context.Context, req *gymv1.ListTrainerReq) (*gymv1.ListTrainerRes, error) {
	pub := true
	return s.listTrainers(ctx, req, &pub)
}
