package queue

import (
	"context"
	"errors"
	"strings"

	"github.com/ilivestrong/email_warmup_service/internal/queue/events"
	"github.com/ilivestrong/email_warmup_service/internal/queue/rmq"
)

type (
	SendEmailEvent        = events.SendEmailEvent
	SendEmailEventHandler = events.SendEmailEventHandler

	Client interface {
		Publish(ctx context.Context, event *SendEmailEvent) error
		Consume(ctx context.Context, handler SendEmailEventHandler) error
	}
)

func NewClient(url string) (Client, error) {
	if strings.HasPrefix(url, "amqp://") ||
		strings.HasPrefix(url, "amqps://") {
		return rmq.New(url)
	}
	return nil, errors.New("unsupported queue URL scheme")
}
