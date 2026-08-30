package helmclient

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRetryOnTransientAPIError_Success(t *testing.T) {
	calls := 0
	err := retryOnTransientAPIError(context.Background(), 3, 1*time.Millisecond, "test", func() error {
		calls++
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, 1, calls)
}

func TestRetryOnTransientAPIError_TransientThenSuccess(t *testing.T) {
	calls := 0
	err := retryOnTransientAPIError(context.Background(), 3, 1*time.Millisecond, "test", func() error {
		calls++
		if calls < 3 {
			return errors.New("etcdserver: request timed out")
		}
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, 3, calls)
}

func TestRetryOnTransientAPIError_NonTransientStopsImmediately(t *testing.T) {
	calls := 0
	terminalErr := errors.New("forbidden")
	err := retryOnTransientAPIError(context.Background(), 3, 1*time.Millisecond, "test", func() error {
		calls++
		return terminalErr
	})
	assert.Equal(t, 1, calls)
	assert.Same(t, terminalErr, err)
}

func TestRetryOnTransientAPIError_AllAttemptsExhausted(t *testing.T) {
	calls := 0
	transient := errors.New("etcdserver: request timed out")
	err := retryOnTransientAPIError(context.Background(), 3, 1*time.Millisecond, "test", func() error {
		calls++
		return transient
	})
	assert.Equal(t, 3, calls)
	assert.Same(t, transient, err)
}

func TestRetryOnTransientAPIError_ContextCancelDuringBackoff(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	go func() {
		time.Sleep(2 * time.Second)
		cancel()
	}()
	err := retryOnTransientAPIError(ctx, 3, 1*time.Hour, "test", func() error {
		calls++
		return errors.New("etcdserver: request timed out")
	})
	assert.Equal(t, 1, calls)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestRetryOnTransientAPIError_DeadlineExceededDuringBackoff(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	calls := 0
	err := retryOnTransientAPIError(ctx, 3, 1*time.Hour, "test", func() error {
		calls++
		return errors.New("etcdserver: request timed out")
	})
	assert.Equal(t, 1, calls)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestRetryOnTransientAPIError_ZeroAttemptsDefaultsToOne(t *testing.T) {
	calls := 0
	err := retryOnTransientAPIError(context.Background(), 0, 1*time.Millisecond, "test", func() error {
		calls++
		return errors.New("etcdserver: request timed out")
	})
	assert.Equal(t, 1, calls)
	assert.Error(t, err)
}
