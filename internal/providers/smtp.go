package providers

import (
	"context"
	"fmt"
	"net/smtp"

	"github.com/ilivestrong/email_warmup_service/internal/config"
)

type SMTPProvider struct {
	host, port, username, password, from string
}

func NewSMTPProvider(cfg config.SMTPConfig) *SMTPProvider {
	return &SMTPProvider{
		host:     cfg.Host,
		port:     cfg.Port,
		username: cfg.User,
		password: cfg.Pass,
		from:     cfg.From,
	}
}

func (s *SMTPProvider) Send(ctx context.Context, to, subj, body string) error {
	auth := smtp.PlainAuth("", s.username, s.password, s.host)
	addr := fmt.Sprintf("%s:%s", s.host, s.port)
	msg := []byte(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", s.from, to, subj, body))
	return smtp.SendMail(addr, auth, s.from, []string{to}, msg)
}

func (s *SMTPProvider) CheckDelivery(ctx context.Context, to, subj, body string) (bool, error) {
	return true, nil
}
func (s *SMTPProvider) CheckBounce(ctx context.Context, to, subj, body string) (bool, error) {
	return false, nil
}
func (s *SMTPProvider) CheckOpen(ctx context.Context, to, subj, body string) (bool, error) {
	return true, nil
}
func (s *SMTPProvider) CheckSpam(ctx context.Context, to, subj, body string) (bool, error) {
	return false, nil
}
