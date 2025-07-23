package quota

import (
	"context"
	"time"
)

type Store interface {
	DeductQuota(ctx context.Context, tenantID string, date string) (bool, int, error)
	ResetQuota(ctx context.Context, tenantID string, date string, count int) error
	SaveScore(ctx context.Context, tenantID, date string, score int) error
	GetScores(ctx context.Context, tenantID, date string) ([]int, error)
	IncreaseQuota(ctx context.Context, tenantID, date string) error
	GetRemainingQuota(ctx context.Context, tenantID string, date time.Time) (int, error)
}
