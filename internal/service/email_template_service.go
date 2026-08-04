package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	etv1 "trongcon-api/api/email_template/v1"
	"trongcon-api/internal/entity"
	"trongcon-api/internal/mail"
	"trongcon-api/internal/repository"

	"gorm.io/gorm"
)

var (
	ErrEmailTemplateNotFound = errors.New("email template not found")
	ErrEmailTemplateKeyTaken = errors.New("email template key already exists")
	ErrEmailTemplateInactive = errors.New("email template is inactive")
	ErrSMTPDisabled          = errors.New("smtp is disabled")
)

var keyRe = regexp.MustCompile(`^[a-z0-9]+(?:[_-][a-z0-9]+)*$`)

type MailerService interface {
	Enabled() bool
	SendByKey(ctx context.Context, templateKey string, data map[string]interface{}, to []string) error
	SendRaw(ctx context.Context, subject, html string, to []string) error
}

type EmailTemplateService interface {
	Create(ctx context.Context, req *etv1.CreateReq) (*etv1.CreateRes, error)
	GetByID(ctx context.Context, id uint) (*etv1.GetRes, error)
	Update(ctx context.Context, id uint, req *etv1.UpdateReq) (*etv1.UpdateRes, error)
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, req *etv1.ListReq) (*etv1.ListRes, error)
	Preview(ctx context.Context, req *etv1.PreviewReq) (*etv1.PreviewRes, error)
	TestSend(ctx context.Context, id uint, req *etv1.TestSendReq) (*etv1.TestSendRes, error)
}

type emailTemplateService struct {
	repo   repository.EmailTemplateRepository
	mailer *mail.Sender
}

func NewEmailTemplateService(repo repository.EmailTemplateRepository, mailer *mail.Sender) EmailTemplateService {
	return &emailTemplateService{repo: repo, mailer: mailer}
}

func NewMailerService(repo repository.EmailTemplateRepository, mailer *mail.Sender) MailerService {
	return &emailTemplateService{repo: repo, mailer: mailer}
}

func (s *emailTemplateService) Enabled() bool {
	return s.mailer != nil && s.mailer.Enabled()
}

func (s *emailTemplateService) SendByKey(ctx context.Context, templateKey string, data map[string]interface{}, to []string) error {
	if !s.Enabled() {
		return ErrSMTPDisabled
	}
	tpl, err := s.repo.GetByKey(ctx, templateKey)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrEmailTemplateNotFound
		}
		return err
	}
	if !tpl.IsActive {
		return ErrEmailTemplateInactive
	}
	subject, html, err := renderPair(tpl.Key, tpl.Subject, tpl.Body, data)
	if err != nil {
		return err
	}
	return s.mailer.Send(ctx, subject, html, to)
}

func (s *emailTemplateService) SendRaw(ctx context.Context, subject, html string, to []string) error {
	if !s.Enabled() {
		return ErrSMTPDisabled
	}
	return s.mailer.Send(ctx, subject, html, to)
}

func (s *emailTemplateService) Create(ctx context.Context, req *etv1.CreateReq) (*etv1.CreateRes, error) {
	key, err := normalizeKey(req.Key)
	if err != nil {
		return nil, err
	}
	if _, err := s.repo.GetByKey(ctx, key); err == nil {
		return nil, ErrEmailTemplateKeyTaken
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	active := true
	if req.IsActive != nil {
		active = *req.IsActive
	}
	// Validate templates parse.
	if _, _, err := renderPair(key, req.Subject, req.Body, map[string]interface{}{}); err != nil {
		return nil, err
	}
	row := &entity.EmailTemplate{
		Name:        strings.TrimSpace(req.Name),
		Key:         key,
		Subject:     strings.TrimSpace(req.Subject),
		Body:        req.Body,
		Description: strings.TrimSpace(req.Description),
		IsActive:    active,
	}
	if err := s.repo.Create(ctx, row); err != nil {
		return nil, err
	}
	return &etv1.CreateRes{Template: toEmailTemplateRes(row)}, nil
}

func (s *emailTemplateService) GetByID(ctx context.Context, id uint) (*etv1.GetRes, error) {
	row, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrEmailTemplateNotFound
		}
		return nil, err
	}
	return &etv1.GetRes{Template: toEmailTemplateRes(row)}, nil
}

func (s *emailTemplateService) Update(ctx context.Context, id uint, req *etv1.UpdateReq) (*etv1.UpdateRes, error) {
	row, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrEmailTemplateNotFound
		}
		return nil, err
	}
	if req.Name != nil {
		row.Name = strings.TrimSpace(*req.Name)
	}
	if req.Key != nil {
		key, err := normalizeKey(*req.Key)
		if err != nil {
			return nil, err
		}
		if key != row.Key {
			if other, err := s.repo.GetByKey(ctx, key); err == nil && other.ID != row.ID {
				return nil, ErrEmailTemplateKeyTaken
			} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, err
			}
			row.Key = key
		}
	}
	if req.Subject != nil {
		row.Subject = strings.TrimSpace(*req.Subject)
	}
	if req.Body != nil {
		row.Body = *req.Body
	}
	if req.Description != nil {
		row.Description = strings.TrimSpace(*req.Description)
	}
	if req.IsActive != nil {
		row.IsActive = *req.IsActive
	}
	if _, _, err := renderPair(row.Key, row.Subject, row.Body, map[string]interface{}{}); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, row); err != nil {
		return nil, err
	}
	return &etv1.UpdateRes{Template: toEmailTemplateRes(row)}, nil
}

func (s *emailTemplateService) Delete(ctx context.Context, id uint) error {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrEmailTemplateNotFound
		}
		return err
	}
	return s.repo.Delete(ctx, id)
}

func (s *emailTemplateService) List(ctx context.Context, req *etv1.ListReq) (*etv1.ListRes, error) {
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
	switch orderBy {
	case "id", "name", "key", "created_at", "updated_at":
	default:
		orderBy = "id"
	}
	dir := strings.ToUpper(strings.TrimSpace(req.OrderDir))
	if dir != "ASC" && dir != "DESC" {
		dir = "DESC"
	}
	list, total, err := s.repo.List(ctx, (page-1)*limit, limit, orderBy+" "+dir, req.Q, req.IsActive)
	if err != nil {
		return nil, err
	}
	data := make([]etv1.EmailTemplateRes, 0, len(list))
	for i := range list {
		data = append(data, toEmailTemplateRes(&list[i]))
	}
	return &etv1.ListRes{Total: total, Data: data}, nil
}

func (s *emailTemplateService) Preview(ctx context.Context, req *etv1.PreviewReq) (*etv1.PreviewRes, error) {
	subject := strings.TrimSpace(req.Subject)
	body := strings.TrimSpace(req.Body)
	key := strings.TrimSpace(req.Key)
	if key != "" {
		row, err := s.repo.GetByKey(ctx, key)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrEmailTemplateNotFound
			}
			return nil, err
		}
		if subject == "" {
			subject = row.Subject
		}
		if body == "" {
			body = row.Body
		}
		key = row.Key
	} else {
		key = "preview"
	}
	if body == "" {
		return nil, fmt.Errorf("body or key is required")
	}
	if subject == "" {
		html, err := mail.Render(key+"_body", body, req.Data)
		if err != nil {
			return nil, err
		}
		return &etv1.PreviewRes{Subject: "", HTML: html}, nil
	}
	subj, html, err := renderPair(key, subject, body, req.Data)
	if err != nil {
		return nil, err
	}
	return &etv1.PreviewRes{Subject: subj, HTML: html}, nil
}

func (s *emailTemplateService) TestSend(ctx context.Context, id uint, req *etv1.TestSendReq) (*etv1.TestSendRes, error) {
	row, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrEmailTemplateNotFound
		}
		return nil, err
	}
	data := req.Data
	if data == nil {
		data = map[string]interface{}{
			"UserName": "TrongCon Tester",
			"Email":    req.To,
		}
	}
	if err := s.SendByKey(ctx, row.Key, data, []string{strings.TrimSpace(req.To)}); err != nil {
		if errors.Is(err, ErrEmailTemplateInactive) {
			// Allow test-send even if inactive (admin check).
			subject, html, rerr := renderPair(row.Key, row.Subject, row.Body, data)
			if rerr != nil {
				return nil, rerr
			}
			if !s.Enabled() {
				return nil, ErrSMTPDisabled
			}
			if err := s.mailer.Send(ctx, subject, html, []string{strings.TrimSpace(req.To)}); err != nil {
				return nil, err
			}
			return &etv1.TestSendRes{Status: "ok"}, nil
		}
		return nil, err
	}
	return &etv1.TestSendRes{Status: "ok"}, nil
}

func normalizeKey(key string) (string, error) {
	key = strings.ToLower(strings.TrimSpace(key))
	key = strings.ReplaceAll(key, " ", "_")
	if !keyRe.MatchString(key) {
		return "", fmt.Errorf("key must be lowercase slug (a-z, 0-9, _ or -)")
	}
	return key, nil
}

func renderPair(key, subject, body string, data map[string]interface{}) (string, string, error) {
	subj, err := mail.Render(key+"_subject", subject, data)
	if err != nil {
		return "", "", err
	}
	html, err := mail.Render(key+"_body", body, data)
	if err != nil {
		return "", "", err
	}
	return subj, html, nil
}

func toEmailTemplateRes(t *entity.EmailTemplate) etv1.EmailTemplateRes {
	return etv1.EmailTemplateRes{
		ID:          t.ID,
		Name:        t.Name,
		Key:         t.Key,
		Subject:     t.Subject,
		Body:        t.Body,
		Description: t.Description,
		IsActive:    t.IsActive,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}
