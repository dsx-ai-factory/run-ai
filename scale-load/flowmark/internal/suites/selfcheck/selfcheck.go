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

func (Suite) Category() core.Category {
	return core.CategoryScale
}

func (Suite) Run(_ context.Context, _ core.Target, scenario core.Scenario) (core.Section, error) {
	start := time.Now()
	time.Sleep(100 * time.Millisecond)

	return core.Section{
		Category: core.CategoryScale,
		Metrics: map[string]any{
			"selfcheck_duration_seconds": time.Since(start).Seconds(),
			"knobs_received":             scenario.Knobs,
		},
	}, nil
}
