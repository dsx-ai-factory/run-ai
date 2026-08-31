/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package enforcer

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// most of the enforcing tests are in pkg/api/InteractivePolicy. Here we only add some leftovers
func TestStepRule(t *testing.T) {

	step := StepRule{
		minVal: "16",
		step:   "0.2",
	}

	stepErr := "value must follow 0.2 steps from 16 (16.00, 16.20, etc.)"

	err := step.Validate(16.4)
	assert.NoError(t, err, "16.4 is good")

	err = step.Validate("16.6")
	assert.NoError(t, err, "16.6 is good")

	err = step.Validate("16.5")
	assert.Error(t, err, "16.5 is not good")
	if err != nil {
		assert.Contains(t, err.Error(), stepErr)
	}

	err = step.Validate("hello")
	assert.Error(t, err, "hello is not a float")
	if err != nil {
		assert.Contains(t, err.Error(), "quantities must match")
	}
}
