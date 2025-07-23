package events

import "context"

type SendEmailEvent struct {
	ToAddress string `json:"toAddress"`
	TenantID  string `json:"tenantId"`
	Subject   string `json:"subject"`
	Body      string `json:"body"`
}

type SendEmailEventHandler func(ctx context.Context, event *SendEmailEvent) error
