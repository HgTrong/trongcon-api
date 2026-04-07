package service

import (
	"context"
	"errors"
	"strings"

	v1 "trongcon-api/api/user/v1"
	"trongcon-api/internal/apimap"
	"trongcon-api/internal/entity"
	"trongcon-api/internal/repository"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrUserNotFound   = errors.New("user not found")
	ErrEmailExists    = errors.New("email already exists")
	ErrInvalidPayload = errors.New("invalid payload")
)

type UserService interface {
	Create(ctx context.Context, req *v1.CreateReq) (*v1.CreateRes, error)
	GetByID(ctx context.Context, id uint) (*v1.GetRes, error)
	Update(ctx context.Context, id uint, req *v1.UpdateReq) (*v1.UpdateRes, error)
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, req *v1.ListUsersReq) (*v1.ListUsersRes, error)
}

type userService struct {
	repo     repository.UserRepository
	roleRepo repository.RoleRepository
}

func NewUserService(repo repository.UserRepository, roleRepo repository.RoleRepository) UserService {
	return &userService{repo: repo, roleRepo: roleRepo}
}

func validGender(g string) bool {
	if g == "" {
		return true
	}
	switch g {
	case "male", "female", "other", "prefer_not_to_say":
		return true
	default:
		return false
	}
}

func validAccountType(a string) bool {
	if a == "" {
		return true
	}
	return a == entity.AccountFree || a == entity.AccountPremium
}

func (s *userService) Create(ctx context.Context, req *v1.CreateReq) (*v1.CreateRes, error) {
	if !validGender(req.Gender) || !validAccountType(req.AccountType) {
		return nil, ErrInvalidPayload
	}

	if _, err := s.repo.GetByEmail(ctx, req.Email); err == nil {
		return nil, ErrEmailExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	at := req.AccountType
	if at == "" {
		at = entity.AccountFree
	}
	lang := req.Language
	if lang == "" {
		lang = "en"
	}
	name := req.Name
	if name == "" {
		name = strings.TrimSpace(req.FirstName + " " + req.LastName)
	}

	u := &entity.User{
		Email:        req.Email,
		Name:         name,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Gender:       req.Gender,
		Language:     lang,
		AccountType:  at,
		PasswordHash: string(hash),
	}
	if err := s.repo.Create(ctx, u); err != nil {
		return nil, err
	}

	roleUser, err := s.roleRepo.GetByName(ctx, entity.RoleUser)
	if err != nil {
		return nil, err
	}
	if err := s.repo.AppendRole(ctx, u, roleUser); err != nil {
		return nil, err
	}

	fresh, err := s.repo.GetByID(ctx, u.ID)
	if err != nil {
		return nil, err
	}
	return &v1.CreateRes{User: apimap.UserToRes(fresh)}, nil
}

func (s *userService) GetByID(ctx context.Context, id uint) (*v1.GetRes, error) {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &v1.GetRes{User: apimap.UserToRes(u)}, nil
}

func (s *userService) Update(ctx context.Context, id uint, req *v1.UpdateReq) (*v1.UpdateRes, error) {
	if !validGender(req.Gender) || !validAccountType(req.AccountType) {
		return nil, ErrInvalidPayload
	}

	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	if req.Email != "" && req.Email != u.Email {
		if existing, err := s.repo.GetByEmail(ctx, req.Email); err == nil && existing.ID != u.ID {
			return nil, ErrEmailExists
		} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		u.Email = req.Email
	}
	if req.Name != "" {
		u.Name = req.Name
	}
	u.FirstName = req.FirstName
	u.LastName = req.LastName
	if req.Gender != "" {
		u.Gender = req.Gender
	}
	if req.Language != "" {
		u.Language = req.Language
	}
	if req.AccountType != "" {
		u.AccountType = req.AccountType
	}

	if err := s.repo.Update(ctx, u); err != nil {
		return nil, err
	}
	fresh, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &v1.UpdateRes{User: apimap.UserToRes(fresh)}, nil
}

func (s *userService) Delete(ctx context.Context, id uint) error {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}
	return s.repo.Delete(ctx, id)
}

func (s *userService) List(ctx context.Context, req *v1.ListUsersReq) (*v1.ListUsersRes, error) {
	page, limit := req.Page, req.Limit
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
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
	case "id", "email", "created_at":
	default:
		orderBy = "id"
	}
	dir := strings.ToUpper(strings.TrimSpace(req.OrderDir))
	if dir != "ASC" && dir != "DESC" {
		dir = "DESC"
	}
	order := orderBy + " " + dir

	list, total, err := s.repo.List(ctx, offset, limit, order)
	if err != nil {
		return nil, err
	}
	data := make([]v1.UserRes, 0, len(list))
	for i := range list {
		data = append(data, apimap.UserToRes(&list[i]))
	}
	return &v1.ListUsersRes{Total: total, Data: data}, nil
}
