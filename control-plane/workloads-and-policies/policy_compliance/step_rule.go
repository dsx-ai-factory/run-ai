/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package policy_compliance

import (
	"encoding/json"
	"fmt"
	"math"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/validation"
)

// StepRule is the implementation of the step rule, which is not provided by the validation third party
type StepRule struct {
	minVal string
	step   string
}

// when working with float we cannot depend on equality, but on an allowed
// difference. for example 1.0000001 and 0.999999 are considered equal to 1
const maxAllowedDiff = float64(0.0001)

func Step(fldRules interface{}) StepRule {

	fldRulesMap := convutils.MapMe(fldRules)

	result := StepRule{
		step:   fldRulesMap[string(PolicyStep)].(json.Number).String(),
		minVal: "0.0",
	}
	if fldRulesMap[string(PolicyMin)] != nil {
		result.minVal = fldRulesMap[string(PolicyMin)].(json.Number).String()
	}
	return result
}

func (r StepRule) Validate(value interface{}) error {

	minValue, err := validation.ToFloat(r.minVal)
	if err != nil {
		return err
	}

	step, err := validation.ToFloat(r.step)
	if err != nil {
		return err
	}

	numericValue, err := validation.ToFloat(value)
	if err != nil {
		return err
	}

	steps := (numericValue - minValue) / step
	diff := math.Abs(steps - math.Round(steps))
	if diff > maxAllowedDiff {
		return fmt.Errorf("value must follow %v steps from %v (%.2f, %.2f, etc.)",
			r.step, minValue, minValue, minValue+step)
	}

	return nil
}
