package resolver

import (
	"context"
	"fmt"
)

type Resolver interface {
	Resolve(ctx context.Context, tenantID string) (string, error)
}

type StaticMapResolver struct{ m map[string]string }

func NewStatic(m map[string]string) *StaticMapResolver { return &StaticMapResolver{m: m} }

func (s *StaticMapResolver) Resolve(_ context.Context, tid string) (string, error) {
	addr, ok := s.m[tid]
	if !ok {
		return "", fmt.Errorf("no sender address for tenant %s", tid)
	}
	return addr, nil
}
