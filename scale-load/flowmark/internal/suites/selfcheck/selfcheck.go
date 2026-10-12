/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package selfcheck

import (
	"context"
	"fmt"
	"strings"
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

var knownKnobs = []string{"message"}

func (Suite) Knobs() []string {
	return knownKnobs
}

func (Suite) Validate(scenario core.Scenario) error {
	if unknown := scenario.UnknownKnobs(knownKnobs...); len(unknown) > 0 {
		return fmt.Errorf("unknown knob %q; selfcheck takes only %s", unknown[0], strings.Join(knownKnobs, ", "))
	}
	return nil
}

func (Suite) Run(_ context.Context, _ core.Target, scenario core.Scenario) (core.Section, error) {
	start := time.Now()
	time.Sleep(100 * time.Millisecond)

	return core.Section{
		Metrics: map[string]any{
			"selfcheck_duration_seconds": time.Since(start).Seconds(),
			"knobs_received":             scenario.Knobs,
		},
	}, nil
}
