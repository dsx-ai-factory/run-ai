/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package selfcheck

import (
	"strings"
	"testing"

	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/core"
)

func TestValidateRejectsUnknownKnobs(t *testing.T) {
	if err := (Suite{}).Validate(core.Scenario{Suite: "selfcheck", Knobs: map[string]any{"message": "hi"}}); err != nil {
		t.Fatalf("message is the one knob selfcheck takes: %v", err)
	}
	err := (Suite{}).Validate(core.Scenario{Suite: "selfcheck", Knobs: map[string]any{"mesage": "typo"}})
	if err == nil || !strings.Contains(err.Error(), `unknown knob "mesage"`) {
		t.Fatalf("a typo must be rejected before the run: %v", err)
	}
}
