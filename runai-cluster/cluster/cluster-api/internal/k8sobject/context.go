/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package k8sobject provides middleware and a cached authorizer for the
// workload-scoped k8s-object endpoints (cluster-api /api/v2/workloads/.../object).
package k8sobject

import (
	"context"
)

// ctxKey is a private type for request-context keys so external packages
// can't accidentally collide.
type ctxKey int

const (
	workloadCtxKey ctxKey = iota
)

func WithWorkloadInfo(ctx context.Context, w *AuthorizedWorkloadInfo) context.Context {
	return context.WithValue(ctx, workloadCtxKey, w)
}

// WorkloadInfoFromContext returns ok=false if no workload was stored (or it was nil).
func WorkloadInfoFromContext(ctx context.Context) (*AuthorizedWorkloadInfo, bool) {
	v, ok := ctx.Value(workloadCtxKey).(*AuthorizedWorkloadInfo)
	return v, ok && v != nil
}
