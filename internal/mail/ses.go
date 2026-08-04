package mail

import (
	"context"
	"fmt"
	"log"
	"net/smtp"
	"strings"
	"time"

	"github.com/jordan-wright/email"
)

// SMTPConfig mirrors AWS SES SMTP settings (SMTP_* / SES_* env).
type SMTPConfig struct {
	Enabled  bool
	Name     string
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

// Sender sends HTML email over SES SMTP.
type Sender struct {
	cfg SMTPConfig
}

func NewSender(cfg SMTPConfig) *Sender {
	if strings.TrimSpace(cfg.Host) == "" {
		cfg.Host = "email-smtp.ap-southeast-1.amazonaws.com"
	}
	if strings.TrimSpace(cfg.Port) == "" {
		cfg.Port = "587"
	}
	return &Sender{cfg: cfg}
}

func (s *Sender) Enabled() bool {
	return s != nil && s.cfg.Enabled &&
		strings.TrimSpace(s.cfg.From) != "" &&
		strings.TrimSpace(s.cfg.Username) != "" &&
		strings.TrimSpace(s.cfg.Password) != ""
}

func (s *Sender) From() string {
	if s == nil {
		return ""
	}
	return s.cfg.From
}

func formatFrom(name, addr string) string {
	name = strings.TrimSpace(name)
	addr = strings.TrimSpace(addr)
	if name == "" {
		return addr
	}
	// Keep ASCII-safe display name for SES SMTP.
	name = strings.ReplaceAll(name, `"`, "")
	return fmt.Sprintf("%s <%s>", name, addr)
}

func sanitizeSubject(subject string) string {
	subject = strings.ReplaceAll(subject, "\r", " ")
	subject = strings.ReplaceAll(subject, "\n", " ")
	return strings.TrimSpace(subject)
}

func (s *Sender) Send(ctx context.Context, subject, html string, to []string) error {
	if s == nil || !s.Enabled() {
		return fmt.Errorf("smtp is disabled or incomplete")
	}
	if len(to) == 0 {
		return fmt.Errorf("recipient is required")
	}
	e := email.NewEmail()
	e.From = formatFrom(s.cfg.Name, s.cfg.From)
	e.To = to
	e.Subject = sanitizeSubject(subject)
	e.HTML = []byte(html)

	addr := fmt.Sprintf("%s:%s", s.cfg.Host, s.cfg.Port)
	auth := smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)

	log.Printf("SES SMTP: sending From=%s To=%v Subject=%s", s.cfg.From, to, e.Subject)

	timeout := 15 * time.Second
	if deadline, ok := ctx.Deadline(); ok {
		if d := time.Until(deadline); d > 0 && d < timeout {
			timeout = d
		}
	}
	sendCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- e.Send(addr, auth)
	}()

	select {
	case <-sendCtx.Done():
		err := fmt.Errorf("ses smtp send timeout: %w", sendCtx.Err())
		log.Printf("SES SMTP: %v", err)
		return err
	case err := <-errCh:
		if err != nil {
			log.Printf("SES SMTP: failed: %v", err)
			return err
		}
		log.Printf("SES SMTP: sent to %v", to)
		return nil
	}
}
