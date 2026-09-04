package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"
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

const (
	forgotPasswordTemplateKey = "forgot_password"
	forgotOTPTTL              = 15 * time.Minute
	forgotResetTokenTTL       = 15 * time.Minute
	signupOTPTTL              = 15 * time.Minute
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrNotSuper           = errors.New("admin access requires super role")
	ErrInvalidOTP         = errors.New("invalid or expired otp")
	ErrInvalidResetToken  = errors.New("invalid or expired reset token")
)

type AuthService interface {
	RequestSignupOTP(ctx context.Context, req *authv1.SignupRequestOTPReq) (*authv1.SignupRequestOTPRes, error)
	Signup(ctx context.Context, req *authv1.SignupReq) (*authv1.LoginRes, error)
	UserLogin(ctx context.Context, email, password string) (*authv1.LoginRes, error)
	AdminLogin(ctx context.Context, email, password string) (*authv1.LoginRes, error)
	ForgotPassword(ctx context.Context, req *authv1.ForgotPasswordReq) (*authv1.ForgotPasswordRes, error)
	VerifyForgotOTP(ctx context.Context, req *authv1.VerifyForgotOTPReq) (*authv1.VerifyForgotOTPRes, error)
	ResetPassword(ctx context.Context, req *authv1.ResetPasswordReq) (*authv1.ResetPasswordRes, error)
}

type authService struct {
	userRepo repository.UserRepository
	roleRepo repository.RoleRepository
	otpRepo  repository.EmailOTPRepository
	mailer   MailerService
	jwtSec   []byte
	jwtExp   time.Duration
}

func NewAuthService(
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
	otpRepo repository.EmailOTPRepository,
	mailer MailerService,
	jwtSecret string,
	jwtExp time.Duration,
) AuthService {
	return &authService{
		userRepo: userRepo,
		roleRepo: roleRepo,
		otpRepo:  otpRepo,
		mailer:   mailer,
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

func (s *authService) RequestSignupOTP(ctx context.Context, req *authv1.SignupRequestOTPReq) (*authv1.SignupRequestOTPRes, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))

	if s.mailer == nil || !s.mailer.Enabled() {
		return nil, ErrSMTPDisabled
	}

	if _, err := s.userRepo.GetByEmail(ctx, email); err == nil {
		return nil, ErrEmailExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	code, err := generateOTP(6)
	if err != nil {
		return nil, err
	}
	_ = s.otpRepo.InvalidateOpen(ctx, email, entity.EmailOTPPurposeSignup)
	row := &entity.EmailOTP{
		Email:    email,
		OTP:      code,
		Purpose:  entity.EmailOTPPurposeSignup,
		Used:     false,
		ExpireAt: time.Now().UTC().Add(signupOTPTTL),
	}
	if err := s.otpRepo.Create(ctx, row); err != nil {
		return nil, err
	}

	data := map[string]interface{}{
		"UserName":   email,
		"Email":      email,
		"OTPCode":    code,
		"VerifyCode": code,
		"ExpireMins": int(signupOTPTTL.Minutes()),
	}
	if err := s.mailer.SendByKey(ctx, forgotPasswordTemplateKey, data, []string{email}); err != nil {
		log.Printf("signup-otp mail failed for %s: %v", email, err)
		return nil, fmt.Errorf("could not send signup verification email (check template key %q and SMTP): %w", forgotPasswordTemplateKey, err)
	}

	return &authv1.SignupRequestOTPRes{Status: "ok", Message: "Mã xác thực đã được gửi tới email của bạn."}, nil
}

func (s *authService) Signup(ctx context.Context, req *authv1.SignupReq) (*authv1.LoginRes, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if _, err := s.userRepo.GetByEmail(ctx, email); err == nil {
		return nil, ErrEmailExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	otpRow, err := s.otpRepo.FindValid(ctx, email, entity.EmailOTPPurposeSignup, strings.TrimSpace(req.OTP))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidOTP
		}
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
		Email:        email,
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

	_ = s.otpRepo.MarkUsed(ctx, otpRow.ID)

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

func (s *authService) ForgotPassword(ctx context.Context, req *authv1.ForgotPasswordReq) (*authv1.ForgotPasswordRes, error) {
	msg := "If an account exists for that email, we sent a verification code."
	email := strings.ToLower(strings.TrimSpace(req.Email))

	if s.mailer == nil || !s.mailer.Enabled() {
		return nil, ErrSMTPDisabled
	}

	u, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &authv1.ForgotPasswordRes{Status: "ok", Message: msg}, nil
		}
		return nil, err
	}

	code, err := generateOTP(6)
	if err != nil {
		return nil, err
	}
	_ = s.otpRepo.InvalidateOpen(ctx, email, entity.EmailOTPPurposeForgotPassword)
	row := &entity.EmailOTP{
		Email:    email,
		OTP:      code,
		Purpose:  entity.EmailOTPPurposeForgotPassword,
		Used:     false,
		ExpireAt: time.Now().UTC().Add(forgotOTPTTL),
	}
	if err := s.otpRepo.Create(ctx, row); err != nil {
		return nil, err
	}

	displayName := strings.TrimSpace(u.Name)
	if displayName == "" {
		displayName = strings.TrimSpace(u.FirstName + " " + u.LastName)
	}
	if displayName == "" {
		displayName = email
	}
	data := map[string]interface{}{
		"UserName":   displayName,
		"Email":      email,
		"OTPCode":    code,
		"VerifyCode": code,
		"ExpireMins": int(forgotOTPTTL.Minutes()),
	}
	if err := s.mailer.SendByKey(ctx, forgotPasswordTemplateKey, data, []string{email}); err != nil {
		log.Printf("forgot-password mail failed for %s: %v", email, err)
		return nil, fmt.Errorf("could not send reset email (check template key %q and SMTP): %w", forgotPasswordTemplateKey, err)
	}

	return &authv1.ForgotPasswordRes{Status: "ok", Message: msg}, nil
}

func (s *authService) VerifyForgotOTP(ctx context.Context, req *authv1.VerifyForgotOTPReq) (*authv1.VerifyForgotOTPRes, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	otp := strings.TrimSpace(req.OTP)

	row, err := s.otpRepo.FindValid(ctx, email, entity.EmailOTPPurposeForgotPassword, otp)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidOTP
		}
		return nil, err
	}
	u, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidOTP
		}
		return nil, err
	}
	if err := s.otpRepo.MarkUsed(ctx, row.ID); err != nil {
		return nil, err
	}

	tok, err := jwtutil.IssueWithPurpose(u.ID, nil, jwtutil.PurposePasswordReset, s.jwtSec, forgotResetTokenTTL)
	if err != nil {
		return nil, err
	}
	return &authv1.VerifyForgotOTPRes{SecretToken: tok}, nil
}

func (s *authService) ResetPassword(ctx context.Context, req *authv1.ResetPasswordReq) (*authv1.ResetPasswordRes, error) {
	claims, err := jwtutil.ParseWithPurpose(strings.TrimSpace(req.SecretToken), jwtutil.PurposePasswordReset, s.jwtSec)
	if err != nil {
		return nil, ErrInvalidResetToken
	}
	u, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidResetToken
		}
		return nil, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	u.PasswordHash = string(hash)
	if err := s.userRepo.Update(ctx, u); err != nil {
		return nil, err
	}
	return &authv1.ResetPasswordRes{Status: "ok"}, nil
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

func generateOTP(digits int) (string, error) {
	if digits <= 0 {
		digits = 6
	}
	max := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(digits)), nil)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%0*d", digits, n.Int64()), nil
}
