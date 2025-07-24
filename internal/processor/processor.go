package processor

import (
	"context"
	"fmt"
	"log"
	"time"

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
}

func New(qs quota.Store, v *validator.Validator, pf *providers.Factory, er resolver.Resolver, qc queue.Client, rp config.RetryPolicy) *Processor {
	return &Processor{qs, v, pf, qc, rp, er}
}

func (p *Processor) Start(ctx context.Context) error { return p.qc.Consume(ctx, p.handle) }

func (p *Processor) handle(ctx context.Context, ev *queue.SendEmailEvent) error {
	now := time.Now().UTC()
	r, _ := p.qs.GetRemainingQuota(ctx, ev.TenantID, now)
	if r <= 0 {
		p.qs.ResetQuota(ctx, ev.TenantID, now.Format("2006-01-02"), 100)
		r = 100
	}
	if !p.v.IsValid(ev.ToAddress) {
		fmt.Printf("[%s] is invalid, skipping...\n", ev.ToAddress)
		return nil
	}
	fromAddr, err := p.emailResolver.Resolve(ctx, ev.TenantID)
	if err != nil {
		return err
	}
	prov, err := p.pf.Get(ev.TenantID)
	if err != nil {
		fmt.Println("p.pf.Get(ev.TenantID) error", err)
		return err
	}
	delivered := false
	opened := false
	bounced := false
	spam := false

	delay := p.rp.InitialDelay
	for i := 0; i <= p.rp.MaxRetries; i++ {
		if err := prov.Send(ctx, fromAddr, ev.ToAddress, ev.Subject, ev.Body); err == nil {
			delivered = true
			break
		} else {
			fmt.Println("error while sending email: ", err)
		}
		time.Sleep(delay)
		delay *= 2
	}

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
	_ = p.qs.SaveScore(ctx, ev.TenantID, now.Format("2006-01-02"), score)

	_, rem, _ := p.qs.DeductQuota(ctx, ev.TenantID, now.Format("2006-01-02"))
	log.Printf("sent, rem %d", rem)
	return nil
}
