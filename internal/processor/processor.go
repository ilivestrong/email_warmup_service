package processor

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/ilivestrong/email_warmup_service/internal/config"
	"github.com/ilivestrong/email_warmup_service/internal/providers"
	"github.com/ilivestrong/email_warmup_service/internal/queue"
	"github.com/ilivestrong/email_warmup_service/internal/quota"
	"github.com/ilivestrong/email_warmup_service/internal/resolver"
	"github.com/ilivestrong/email_warmup_service/internal/validator"
)

type Processor struct {
	qs            quota.Store
	v             *validator.Validator
	pf            *providers.Factory
	qc            queue.Client
	rp            config.RetryPolicy
	emailResolver resolver.Resolver
	log           *slog.Logger
}

func New(qs quota.Store, v *validator.Validator, pf *providers.Factory, er resolver.Resolver, qc queue.Client, rp config.RetryPolicy, log *slog.Logger) *Processor {
	return &Processor{qs, v, pf, qc, rp, er, log}
}

func (p *Processor) Start(ctx context.Context) error { return p.qc.Consume(ctx, p.handle) }

func (p *Processor) handle(ctx context.Context, ev *queue.SendEmailEvent) error {

	l := p.log.With(
		slog.String("tenant_id", ev.TenantID),
		slog.String("event_id", uuid.New().String()),
		slog.String("to", ev.ToAddress),
		slog.String("subject", ev.Subject),
	)

	l.Info("EVENT_RECEIVED")

	if !p.v.IsValid(ev.ToAddress) {
		fmt.Printf("[%s] is invalid, skipping...\n", ev.ToAddress)
		return nil
	}
	l.Info("VALIDATION_PASSED")

	now := time.Now().UTC()
	r, _ := p.qs.GetRemainingQuota(ctx, ev.TenantID, now)
	if r <= 0 {
		p.qs.ResetQuota(ctx, ev.TenantID, now.Format("2006-01-02"), 100)
		r = 100
		l.Info("QUOTA_RESET", slog.Int("reset_to", 100))
	}
	l.Info("QUOTA_CHECKED")

	fromAddr, err := p.emailResolver.Resolve(ctx, ev.TenantID)
	if err != nil {
		l.Error("ADDRESS_RESOLVE_FAILED", slog.Any("error", err))
		return err
	}
	prov, err := p.pf.Get(ev.TenantID)
	if err != nil {
		l.Error("PROVIDER_SELECT_FAILED", slog.Any("error", err))
		return err
	}

	delivered := false
	delay := p.rp.InitialDelay
	for i := 0; i <= p.rp.MaxRetries; i++ {
		l.Info("SEND_ATTEMPT", slog.Int("attempt", i+1))
		if err := prov.Send(ctx, fromAddr, ev.ToAddress, ev.Subject, ev.Body); err == nil {
			delivered = true
			l.Info("SEND_SUCCESS", slog.Int("attempt", i+1))
			break
		} else {
			fmt.Println("error while sending email: ", err)
		}
		l.Warn("SEND_FAIL", slog.Int("attempt", i+1), slog.Any("error", err))
		time.Sleep(delay)
		delay *= 2
	}

	bounced, _ := prov.CheckBounce(ctx, ev.ToAddress, ev.Subject, ev.Body)
	opened, _ := prov.CheckOpen(ctx, ev.ToAddress, ev.Subject, ev.Body)
	spam, _ := prov.CheckSpam(ctx, ev.ToAddress, ev.Subject, ev.Body)
	l.Info("STATUS_RECONCILED",
		slog.Bool("delivered", delivered),
		slog.Bool("bounced", bounced),
		slog.Bool("opened", opened),
		slog.Bool("spam", spam),
	)

	score := calcScore(delivered, bounced, opened, spam)
	_ = p.qs.SaveScore(ctx, ev.TenantID, now.Format("2006-01-02"), score)
	l.Info("SCORE_SAVED", slog.Int("score", score))

	_, rem, _ := p.qs.DeductQuota(ctx, ev.TenantID, now.Format("2006-01-02"))
	log.Printf("sent, rem %d", rem)
	l.Info("QUOTA_DEDUCTED", slog.Int("remaining", rem))

	l.Info("DONE")
	return nil
}

func calcScore(delivered, bounced, opened, spam bool) int {
	score := 0
	if delivered {
		score += 2
	} else if bounced {
		score -= 1
	}
	if opened {
		score += 1
	}
	if spam {
		score -= 2
	}
	return score
}
