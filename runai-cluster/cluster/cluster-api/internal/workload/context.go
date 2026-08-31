/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workload

import (
	"context"
)

type keyContext string

const contextWorkloadKey keyContext = "workload"

func FromContext(ctx context.Context) *Workload {
	raw := ctx.Value(contextWorkloadKey)
	if raw == nil {
		return nil
	}

	workload, ok := raw.(*Workload)
	if !ok {
		return nil
	}

	return workload
}

func SetIntoContext(ctx context.Context, workload *Workload) context.Context {
	return context.WithValue(ctx, contextWorkloadKey, workload)
}
