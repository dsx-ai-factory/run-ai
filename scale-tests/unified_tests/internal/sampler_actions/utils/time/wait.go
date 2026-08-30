package time

import (
	"context"
	"fmt"
	"time"
)

type PredFn func() bool

func WaitFor(ctx context.Context, pred PredFn, timeout, interval time.Duration) error {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctxWithTimeout.Done():
			return fmt.Errorf("condition did not become true within the timeout")
		case <-ticker.C:
			if pred() {
				return nil
			}
		}
	}
}
