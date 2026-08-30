package prometheus

import (
	"context"
	"time"
)

type Querier interface {
	GetWorkloadIdleDuration(ctx context.Context, workloadName, workloadID string) (*time.Duration, error)
}
