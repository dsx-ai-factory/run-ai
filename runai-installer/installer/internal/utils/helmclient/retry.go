package helmclient

import (
	"context"
	"strings"
	"time"

	"github.com/run-ai/runai/runai-installer/installer/internal/logging"
)

const (
	helmRetryAttempts    = 3
	helmRetryBaseBackoff = 2 * time.Second
)

// transientAPIErrorPatterns are substrings we treat as transient
// kube-apiserver / etcd errors. Conservative on purpose — we'd rather miss a
// retry-able error than retry something deterministic that will keep failing.
// Add new patterns here as we observe them in the field.
var transientAPIErrorPatterns = []string{
	"etcdserver: request timed out",
	"etcdserver: leader changed",
	"the server is currently unable to handle the request",
	"connection reset by peer",
	"i/o timeout",
	"unexpected EOF",
}

// isTransientAPIError reports whether err looks like a transient kube/etcd
// failure that is worth retrying. It matches against the error string because
// helm wraps SDK errors with github.com/pkg/errors, breaking the standard
// errors.Is chain (see also isHelmWaitTimeout).
func isTransientAPIError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	// "context deadline exceeded" is the helm SDK's own timeout firing; retrying
	// would only make wall time worse.
	if strings.Contains(msg, "context deadline exceeded") {
		return false
	}
	for _, pattern := range transientAPIErrorPatterns {
		if strings.Contains(msg, pattern) {
			return true
		}
	}
	return false
}

// retryOnTransientAPIError invokes fn up to attempts times, retrying only when
// fn returns an error that isTransientAPIError recognizes. Backoff doubles
// each iteration starting at baseBackoff. If ctx is cancelled during a
// backoff sleep, the helper returns ctx.Err() immediately. On a non-transient
// error, the helper exits early with that error.
func retryOnTransientAPIError(ctx context.Context, attempts int, baseBackoff time.Duration, label string, fn func() error) error {
	if attempts < 1 {
		attempts = 1
	}
	var lastErr error
	backoff := baseBackoff
	for attempt := 1; attempt <= attempts; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}
		lastErr = err
		if !isTransientAPIError(err) {
			return err
		}
		if attempt == attempts {
			logging.Errorf("[Helm] %s: transient API error on final attempt %d/%d: %v", label, attempt, attempts, err)
			break
		}
		logging.Infof("[Helm] %s: transient API error on attempt %d/%d: %v; retrying in %v", label, attempt, attempts, err, backoff)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
		}
		backoff *= 2
	}
	return lastErr
}
