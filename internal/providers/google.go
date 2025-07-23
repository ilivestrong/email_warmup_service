package providers

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/ilivestrong/email_warmup_service/internal/config"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// Ensure Provider interface is implemented
var _ Provider = (*GoogleProvider)(nil)

type GoogleProvider struct {
	service *gmail.Service
	sender  string
}

func NewGoogleProvider(cfg config.GoogleOAuthConfig) Provider {
	ctx := context.Background()
	oauthConfig, err := google.ConfigFromJSON([]byte(cfg.GoogleCredentialsJSON), gmail.GmailSendScope, gmail.GmailReadonlyScope)
	if err != nil {
		fmt.Println("----------- ", cfg.GoogleCredentialsJSON)
		panic(fmt.Sprintf("failed to parse OAuth config: %v", err))
	}
	token := &oauth2.Token{
		AccessToken:  cfg.GoogleAccessToken,
		RefreshToken: cfg.GoogleRefreshToken,
		Expiry:       time.Now().Add(time.Hour),
	}

	client := oauthConfig.Client(ctx, token)
	service, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		panic(fmt.Sprintf("failed to create Gmail service: %v", err))
	}

	return &GoogleProvider{
		service: service,
		sender:  cfg.GoogleEmailSender,
	}
}

func (g *GoogleProvider) Send(ctx context.Context, from, to, subject, body string) error {
	fmt.Println(from, to)
	msg := []byte(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", from, to, subject, body))
	raw := base64.URLEncoding.EncodeToString(msg)
	_, err := g.service.Users.Messages.Send(from, &gmail.Message{Raw: raw}).Do()
	return err
}

func (g *GoogleProvider) CheckDelivery(ctx context.Context, to, subj, body string) (bool, error) {
	query := fmt.Sprintf("to:%s subject:%q", to, subj)
	resp, err := g.service.Users.Messages.List("me").Q(query).Do()
	if err != nil {
		return false, err
	}
	return len(resp.Messages) > 0, nil
}

func (g *GoogleProvider) CheckBounce(ctx context.Context, to, subj, body string) (bool, error) {
	query := fmt.Sprintf("from:mailer-daemon@googlemail.com to:%s subject:bounced", to)
	resp, err := g.service.Users.Messages.List("me").Q(query).Do()
	if err != nil {
		return false, err
	}
	return len(resp.Messages) > 0, nil
}

func (g *GoogleProvider) CheckOpen(ctx context.Context, to, subj, body string) (bool, error) {
	query := fmt.Sprintf("to:%s subject:%q label:UNREAD", to, subj)
	resp, err := g.service.Users.Messages.List("me").Q(query).Do()
	if err != nil {
		return false, err
	}
	// If not unread, we assume it's opened
	return len(resp.Messages) == 0, nil
}

func (g *GoogleProvider) CheckSpam(ctx context.Context, to, subj, body string) (bool, error) {
	query := fmt.Sprintf("to:%s subject:%q in:spam", to, subj)
	resp, err := g.service.Users.Messages.List("me").Q(query).Do()
	if err != nil {
		return false, err
	}
	return len(resp.Messages) > 0, nil
}
