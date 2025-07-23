package quota

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

type redisStore struct {
	rdb *redis.Client
}

func NewStore(redisURL string) (Store, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}
	return &redisStore{rdb: redis.NewClient(opts)}, nil
}

func (r *redisStore) key(t string, d time.Time) string {
	return fmt.Sprintf("quota:%s:%s", t, d.Format("2006-01-02"))
}

func (r *redisStore) GetRemainingQuota(ctx context.Context, t string, d time.Time) (int, error) {
	v, err := r.rdb.Get(ctx, r.key(t, d)).Result()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	q, err := strconv.Atoi(v)
	if err != nil {
		return 0, err
	}
	return q, nil
}

func (r *redisStore) DeductQuota(ctx context.Context, tenantID string, date string) (bool, int, error) {
	key := fmt.Sprintf("quota:%s:%s", tenantID, date)
	res := r.rdb.Decr(ctx, key)
	val, err := res.Result()
	if err != nil {
		return false, 0, err
	}
	return val >= 0, int(val), nil
}

func (r *redisStore) ResetQuota(ctx context.Context, tenantID string, date string, count int) error {
	key := fmt.Sprintf("quota:%s:%s", tenantID, date)
	return r.rdb.Set(ctx, key, count, 24*time.Hour).Err()
}

func (r *redisStore) SaveScore(ctx context.Context, tenantID, date string, score int) error {
	key := fmt.Sprintf("score:%s:%s", tenantID, date)
	return r.rdb.RPush(ctx, key, score).Err()
}

func (r *redisStore) GetScores(ctx context.Context, tenantID, date string) ([]int, error) {
	key := fmt.Sprintf("score:%s:%s", tenantID, date)
	vals, err := r.rdb.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, err
	}
	var out []int
	for _, v := range vals {
		val, err := strconv.Atoi(v)
		if err == nil {
			out = append(out, val)
		}
	}
	return out, nil
}

func (r *redisStore) IncreaseQuota(ctx context.Context, tenantID, date string) error {
	today := time.Now().Format("2006-01-02")
	key := fmt.Sprintf("quota:%s:%s", tenantID, today)
	cur, err := r.rdb.Get(ctx, key).Int()
	if err != nil && err != redis.Nil {
		return err
	}
	if cur == 0 {
		cur = 5 // Default quota if not set yet
	}
	newQuota := int(float64(cur) * 1.5)
	return r.rdb.Set(ctx, key, newQuota, 24*time.Hour).Err()
}
