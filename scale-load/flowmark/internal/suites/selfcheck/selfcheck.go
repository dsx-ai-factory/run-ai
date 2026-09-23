/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package selfcheck

import (
	"context"
	"time"

	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/core"
)

type Suite struct{}

func (Suite) Name() string {
	return "selfcheck"
}

func (Suite) Run(_ context.Context, _ core.Env, cfg core.Config) (core.Section, error) {
	start := time.Now()
	time.Sleep(100 * time.Millisecond)

	return core.Section{
		Category: core.CategoryScale,
		Metrics: map[string]any{
			"selfcheck_duration_seconds": time.Since(start).Seconds(),
			"knobs_received":             cfg.Knobs,
		},
	}, nil
}
