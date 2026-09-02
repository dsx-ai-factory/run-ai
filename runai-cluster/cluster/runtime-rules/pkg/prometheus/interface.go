/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package prometheus

import (
	"context"
	"time"
)

type Querier interface {
	GetWorkloadIdleDuration(ctx context.Context, workloadName, workloadID string) (*time.Duration, error)
}
