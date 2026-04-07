package service

import (
	"context"
	"errors"
	"strings"
	"time"

	authv1 "trongcon-api/api/auth/v1"
	"trongcon-api/internal/apimap"
	"trongcon-api/internal/entity"
	"trongcon-api/internal/jwtutil"
	"trongcon-api/internal/repository"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrNotSuper           = errors.New("admin access requires super role")
)

type AuthService interface {
	Signup(ctx context.Context, req *authv1.SignupReq) (*authv1.LoginRes, error)
	UserLogin(ctx context.Context, email, password string) (*authv1.LoginRes, error)
	AdminLogin(ctx context.Context, email, password string) (*authv1.LoginRes, error)
}

type authService struct {
	userRepo repository.UserRepository
	roleRepo repository.RoleRepository
	jwtSec   []byte
	jwtExp   time.Duration
}

func NewAuthService(userRepo repository.UserRepository, roleRepo repository.RoleRepository, jwtSecret string, jwtExp time.Duration) AuthService {
	return &authService{
		userRepo: userRepo,
		roleRepo: roleRepo,
		jwtSec:   []byte(jwtSecret),
		jwtExp:   jwtExp,
	}
}

func roleNames(u *entity.User) []string {
	return apimap.RoleNames(u)
}

func hasSuper(names []string) bool {
	for _, n := range names {
		if n == entity.RoleSuper {
			return true
		}
	}
	return false
}

func (s *authService) Signup(ctx context.Context, req *authv1.SignupReq) (*authv1.LoginRes, error) {
	if _, err := s.userRepo.GetByEmail(ctx, req.Email); err == nil {
		return nil, ErrEmailExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	lang := req.Language
	if lang == "" {
		lang = "en"
	}
	name := strings.TrimSpace(req.FirstName + " " + req.LastName)

	u := &entity.User{
		Email:        req.Email,
		Name:         name,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Gender:       req.Gender,
		Language:     lang,
		AccountType:  entity.AccountFree,
		PasswordHash: string(hash),
	}
	if err := s.userRepo.Create(ctx, u); err != nil {
		return nil, err
	}

	roleUser, err := s.roleRepo.GetByName(ctx, entity.RoleUser)
	if err != nil {
		return nil, err
	}
	if err := s.userRepo.AppendRole(ctx, u, roleUser); err != nil {
		return nil, err
	}

	fresh, err := s.userRepo.GetByID(ctx, u.ID)
	if err != nil {
		return nil, err
	}

	return s.issueLoginRes(fresh)
}

func (s *authService) UserLogin(ctx context.Context, email, password string) (*authv1.LoginRes, error) {
	u, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	now := time.Now().UTC()
	if err := s.userRepo.UpdateLastLoginAt(ctx, u.ID, now); err != nil {
		return nil, err
	}

	fresh, err := s.userRepo.GetByID(ctx, u.ID)
	if err != nil {
		return nil, err
	}

	return s.issueLoginRes(fresh)
}

func (s *authService) AdminLogin(ctx context.Context, email, password string) (*authv1.LoginRes, error) {
	u, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}
	if !hasSuper(roleNames(u)) {
		return nil, ErrNotSuper
	}

	now := time.Now().UTC()
	if err := s.userRepo.UpdateLastLoginAt(ctx, u.ID, now); err != nil {
		return nil, err
	}

	fresh, err := s.userRepo.GetByID(ctx, u.ID)
	if err != nil {
		return nil, err
	}

	return s.issueLoginRes(fresh)
}

func (s *authService) issueLoginRes(u *entity.User) (*authv1.LoginRes, error) {
	names := roleNames(u)
	tok, err := jwtutil.Issue(u.ID, names, s.jwtSec, s.jwtExp)
	if err != nil {
		return nil, err
	}
	return &authv1.LoginRes{
		Token: tok,
		User:  apimap.UserToRes(u),
	}, nil
}
