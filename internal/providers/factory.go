package providers

import (
	"context"
	"errors"

	"github.com/ilivestrong/email_warmup_service/internal/config"
)

type Provider interface {
	Send(ctx context.Context, to, subj, body string) error
	CheckDelivery(ctx context.Context, to, subj, body string) (bool, error)
	CheckBounce(ctx context.Context, to, subj, body string) (bool, error)
	CheckOpen(ctx context.Context, to, subj, body string) (bool, error)
	CheckSpam(ctx context.Context, to, subj, body string) (bool, error)
}

type Factory struct {
	cfg               map[string]string
	smtpConfig        config.SMTPConfig
	googleOAuthConfig config.GoogleOAuthConfig
}

func NewFactory(m map[string]string, smtpCfg config.SMTPConfig, googleOAuthCfg config.GoogleOAuthConfig) *Factory {
	return &Factory{m, smtpCfg, googleOAuthCfg}
}

func (f *Factory) Get(tenantID string) (Provider, error) {
	t, ok := f.cfg[tenantID]
	if !ok {
		return nil, errors.New("no provider for tenant: " + tenantID)
	}
	switch t {
	case "smtp":
		return NewSMTPProvider(f.smtpConfig), nil
	case "google":
		return NewGoogleProvider(f.googleOAuthConfig), nil

	case "outlook":
		// return NewOutlookProvider()
		return nil, nil
	}
	return nil, errors.New("unknown provider: " + t)
}
