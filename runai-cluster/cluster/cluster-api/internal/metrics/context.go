/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics

import "context"

type keyContext string

const (
	contextMetricsKey keyContext = "metrics"
)

func FromContext(ctx context.Context) *EndpointMeta {
	if ctx == nil {
		return nil
	}

	raw := ctx.Value(contextMetricsKey)
	if raw == nil {
		return nil
	}

	info, ok := raw.(*EndpointMeta)
	if !ok {
		return nil
	}

	return info
}

func SetIntoContext(ctx context.Context, info *EndpointMeta) context.Context {
	return context.WithValue(ctx, contextMetricsKey, info)
}
