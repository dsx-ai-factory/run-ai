/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package gvk_compliance

import (
	"fmt"
	"math"

	"github.com/run-ai/runai/runai-application-packages/validation"
	"k8s.io/apimachinery/pkg/api/resource"
)

// StepMaxAllowedDiff is the float tolerance for step-boundary checks. Floating
// point arithmetic on quantities like 1.0 / 0.1 may yield 9.9999999 or
// 10.0000001 — values within this distance of a whole-number step count are
// considered on the boundary. Same semantics as RangeEnforcement / QuantityRangeEnforcement.
const StepMaxAllowedDiff = 0.0001

// OnStep reports whether value lies on a step boundary from minAnchor. A step of
// 0 disables the check (treated as valid). Shared with policy-service rule validation.
func OnStep(value, step, minAnchor float64) bool {
	if step == 0 {
		return true
	}
	steps := (value - minAnchor) / step
	diff := math.Abs(steps - math.Round(steps))
	return diff <= StepMaxAllowedDiff
}

// ValidateFloat32RangeBounds checks internal consistency of a numeric range enforcement
// (min <= max, step > 0, step <= max-min when all three set). Used by policy-service
// and mirrors validateFloatRangeBounds / validateRangeEnforcement in policy-service.
func ValidateFloat32RangeBounds(minVal, maxVal, stepVal *float32) error {
	if minVal != nil && maxVal != nil && *minVal > *maxVal {
		return fmt.Errorf("range.min (%v) must be less than or equal to range.max (%v)", *minVal, *maxVal)
	}
	if stepVal != nil && *stepVal <= 0 {
		return fmt.Errorf("range.step must be greater than 0, got %v", *stepVal)
	}
	if minVal != nil && maxVal != nil && stepVal != nil {
		if span := *maxVal - *minVal; *stepVal > span {
			return fmt.Errorf("range.step (%v) must be less than or equal to range.max-range.min (%v)", *stepVal, span)
		}
	}
	return nil
}

// ValidateNumericElementAgainstRange validates one scalar against range bounds
// (same semantics as RangeEnforcement.Exec).
func ValidateNumericElementAgainstRange(element any, minVal, maxVal, stepVal *float32) error {
	if minVal != nil {
		if err := validation.Min(*minVal).Validate(element); err != nil {
			return err
		}
	}
	if maxVal != nil {
		if err := validation.Max(*maxVal).Validate(element); err != nil {
			return err
		}
	}
	if stepVal != nil {
		numeric, err := validation.ToFloat(element)
		if err != nil {
			return err
		}
		anchor := float64(0)
		if minVal != nil {
			anchor = float64(*minVal)
		}
		if !OnStep(numeric, float64(*stepVal), anchor) {
			return fmt.Errorf("value %v must be a multiple of %v from %v", numeric, *stepVal, anchor)
		}
	}
	return nil
}

// ValidateQuantityRangeBounds checks min <= max, step > 0, and step <= max-min when all three set.
func ValidateQuantityRangeBounds(minQty, maxQty, stepQty *resource.Quantity) error {
	if minQty != nil && maxQty != nil && minQty.Cmp(*maxQty) > 0 {
		return fmt.Errorf("quantityRange.min (%s) must be less than or equal to quantityRange.max (%s)", minQty.String(), maxQty.String())
	}
	if stepQty != nil && stepQty.Sign() <= 0 {
		return fmt.Errorf("quantityRange.step must be greater than 0, got %s", stepQty.String())
	}
	if minQty != nil && maxQty != nil && stepQty != nil {
		span := maxQty.DeepCopy()
		span.Sub(*minQty)
		if stepQty.Cmp(span) > 0 {
			return fmt.Errorf("quantityRange.step (%s) must be less than or equal to quantityRange.max-quantityRange.min (%s)", stepQty.String(), span.String())
		}
	}
	return nil
}

// ValidateQuantityValueAgainstRange validates one parsed quantity against bounds
// (same semantics as QuantityRangeEnforcement.Exec).
func ValidateQuantityValueAgainstRange(q resource.Quantity, minQty, maxQty, stepQty *resource.Quantity) error {
	if minQty != nil && q.Cmp(*minQty) < 0 {
		return fmt.Errorf("value %s is below minimum %s", q.String(), minQty.String())
	}
	if maxQty != nil && q.Cmp(*maxQty) > 0 {
		return fmt.Errorf("value %s exceeds maximum %s", q.String(), maxQty.String())
	}
	if stepQty != nil {
		stepFloat := stepQty.AsApproximateFloat64()
		anchor := float64(0)
		if minQty != nil {
			anchor = minQty.AsApproximateFloat64()
		}
		if !OnStep(q.AsApproximateFloat64(), stepFloat, anchor) {
			return fmt.Errorf("value %s must be a multiple of %s from %s", q.String(), stepQty.String(), QuantityAnchorString(minQty))
		}
	}
	return nil
}

// QuantityAnchorString renders the step anchor for error messages (0 when no min).
func QuantityAnchorString(minQty *resource.Quantity) string {
	if minQty == nil {
		return "0"
	}
	return minQty.String()
}

// ParseQuantityString parses val as a Kubernetes resource quantity when val is a string.
func ParseQuantityString(val any) (resource.Quantity, error) {
	s, ok := val.(string)
	if !ok {
		return resource.Quantity{}, fmt.Errorf("expected string quantity, got %T", val)
	}
	return resource.ParseQuantity(s)
}
