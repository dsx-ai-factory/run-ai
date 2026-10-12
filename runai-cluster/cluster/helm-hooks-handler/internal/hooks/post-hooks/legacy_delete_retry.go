/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package post_hooks

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// webhookCallFailurePrefix is how the API server reports a transport-level failure
// calling an admission webhook (k8s.io/apiserver/pkg/util/webhook.ErrCallingWebhook),
// wrapped in an InternalError when the webhook's failurePolicy is Fail.
const webhookCallFailurePrefix = "failed calling webhook"

// Retry settings for legacy object cleanup that fails because the API server could
// not reach the admission webhook guarding them. Package variables so tests can
// shorten them.
var (
	webhookRetryInterval = 5 * time.Second
	webhookRetryTimeout  = 3 * time.Minute
)

// deleteLegacyObject deletes a legacy department or project, retrying while the
// API server cannot reach the admission webhook that validates the delete.
//
// During an upgrade the previous release's validating webhook (for example
// runai-project-validation) can still point at a Service port the new operator no
// longer serves; init-ca re-points it asynchronously. Until then every delete fails
// with an InternalError "failed calling webhook ...". The delete is idempotent, so
// it is retried every webhookRetryInterval for up to webhookRetryTimeout in total.
// The deadline also bounds each delete call, so a request that never answers
// cannot hold the hook past it. Any other error is returned immediately. NotFound
// and NoMatch mean the object is already gone and count as success.
func deleteLegacyObject(ctx context.Context, k8sClient client.Client, obj client.Object, kind string, logger logr.Logger) error {
	logger.Info("Deleting legacy "+kind, "name", obj.GetName())
	return retryLegacyOperation(ctx, obj, kind, "delete", logger, func(ctx context.Context) error {
		return k8sClient.Delete(ctx, obj)
	})
}

// Finalizer patches and deletes can both race with admission webhook startup.
// Bound each operation, including its API calls, with the same retry deadline.
func retryLegacyOperation(ctx context.Context, obj client.Object, kind, operation string, logger logr.Logger, call func(context.Context) error) error {
	ctx, cancel := context.WithTimeout(ctx, webhookRetryTimeout)
	defer cancel()

	var lastWebhookErr error
	for {
		err := call(ctx)
		switch {
		case err == nil || apierrors.IsNotFound(err) || meta.IsNoMatchError(err):
			return nil
		case isWebhookCallFailure(err):
			lastWebhookErr = err
		case ctx.Err() != nil:
			// The API call itself was cut off, by the retry deadline or by the caller.
			return retryStopped(ctx, kind, obj, operation, lastWebhookErr)
		default:
			return fmt.Errorf("failed to %s legacy %s %q: %w", operation, kind, obj.GetName(), err)
		}

		logger.Info("Admission webhook not reachable yet, retrying legacy cleanup", "operation", operation, "kind", kind, "name", obj.GetName(), "error", lastWebhookErr.Error())
		// Waits for the retry interval or the remaining deadline, whichever ends first.
		select {
		case <-ctx.Done():
			return retryStopped(ctx, kind, obj, operation, lastWebhookErr)
		case <-time.After(webhookRetryInterval):
		}
	}
}

// retryStopped builds the error for an operation that ran out of time. On the retry
// deadline it carries the last webhook failure seen, if any, so the log explains
// what the hook was waiting for; on caller cancellation it carries the context error.
func retryStopped(ctx context.Context, kind string, obj client.Object, operation string, lastWebhookErr error) error {
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		if lastWebhookErr != nil {
			return fmt.Errorf("failed to %s legacy %s %q: admission webhook still unreachable after %s: %w", operation, kind, obj.GetName(), webhookRetryTimeout, lastWebhookErr)
		}
		return fmt.Errorf("failed to %s legacy %s %q: no answer from the API server within %s: %w", operation, kind, obj.GetName(), webhookRetryTimeout, ctx.Err())
	}
	return fmt.Errorf("failed to %s legacy %s %q: %w", operation, kind, obj.GetName(), ctx.Err())
}

// isWebhookCallFailure reports whether err is the API server failing to call an
// admission webhook, as opposed to the webhook rejecting the request.
func isWebhookCallFailure(err error) bool {
	return apierrors.IsInternalError(err) && strings.Contains(err.Error(), webhookCallFailurePrefix)
}
