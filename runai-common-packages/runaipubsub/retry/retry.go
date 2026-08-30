package retry

import (
	"context"
	"time"

	"github.com/Rican7/retry"
	"github.com/Rican7/retry/backoff"
	"github.com/Rican7/retry/strategy"
)

func Retry(ctx context.Context, f func(uint) error, attempts int) error {
	exp := backoff.Exponential(time.Second, 2)
	return retry.Retry(
		func(u uint) error {
			if ctx.Err() != nil {
				return nil
			}
			return f(u)
		},
		strategy.Limit(uint(attempts)), //nolint:gosec // G115: value is bounded/non-negative
		strategy.Backoff(func(attempt uint) time.Duration {
			return min(time.Minute, exp(attempt))
		}),
	)
}

func RetryStreamErrors(ctx context.Context, errors chan error, f func(uint) error, attempts int) error {
	if err := Retry(
		ctx,
		func(u uint) error {
			if err := f(u); err != nil {
				tryAppendToChan(errors, err)
				return err
			}
			return nil
		},
		attempts,
	); err != nil {
		tryAppendToChan(errors, err)
		return err
	}
	return nil
}

func tryAppendToChan(errChan chan error, err error) {
	select {
	case errChan <- err:
	default:
	}
}
