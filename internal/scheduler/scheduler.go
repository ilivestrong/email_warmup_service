package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/ilivestrong/email_warmup_service/internal/config"
	"github.com/ilivestrong/email_warmup_service/internal/providers"
	"github.com/ilivestrong/email_warmup_service/internal/quota"
)

type Scheduler struct {
	cfg     *config.Config
	store   quota.Store
	factory *providers.Factory
}

func NewScheduler(cfg *config.Config, store quota.Store, factory *providers.Factory) *Scheduler {
	return &Scheduler{cfg: cfg, store: store, factory: factory}
}

func (s *Scheduler) StartDaily(ctx context.Context) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("daily scheduler stopped")
			return
		case <-ticker.C:
			s.runDailyScoreCheck(ctx)
		}
	}
}

func (s *Scheduler) runDailyScoreCheck(ctx context.Context) {
	for tenantID := range s.cfg.ProviderMap {
		date := time.Now().Add(-24 * time.Hour).Format("2006-01-02")
		scores, err := s.store.GetScores(ctx, tenantID, date)
		if err != nil {
			log.Printf("no scores for tenant %s: %v", tenantID, err)
			continue
		}

		total := 0
		for _, score := range scores {
			total += score
		}
		avg := float64(total) / float64(len(scores))

		if avg >= 0.8 {
			log.Printf("scaling quota for %s due to good score %.2f", tenantID, avg)
			if err := s.store.IncreaseQuota(ctx, tenantID, date); err != nil {
				log.Printf("failed to increase quota: %v", err)
			}
		}
	}
}
