package services

// Numeric and quantity step/range checks mirror
// workloads-and-policies/gvk_compliance/policy_action_validate.go so policy
// rules stay consistent with the runtime range/quantityRange enforcement
// checks.

import (
	"fmt"
	"math"

	"github.com/run-ai/runai/runai-application-packages/validation"
	"k8s.io/apimachinery/pkg/api/resource"
)

const policyRuleStepMaxAllowedDiff = 0.0001

func policyRuleOnStep(value, step, minAnchor float64) bool {
	if step == 0 {
		return true
	}
	steps := (value - minAnchor) / step
	diff := math.Abs(steps - math.Round(steps))
	return diff <= policyRuleStepMaxAllowedDiff
}

// coerceNumericElementForRangeValidator converts whole-number scalars that often
// appear in policy JSON / []interface{} literals (int, int64, json float) into
// float64 so validation.Min / Max / ToFloat accept them consistently.
func coerceNumericElementForRangeValidator(element any) any {
	switch v := element.(type) {
	case int:
		return float64(v)
	case int8:
		return float64(v)
	case int16:
		return float64(v)
	case int32:
		return float64(v)
	case int64:
		return float64(v)
	case uint:
		return float64(v)
	case uint8:
		return float64(v)
	case uint16:
		return float64(v)
	case uint32:
		return float64(v)
	case uint64:
		return float64(v)
	case float32:
		return float64(v)
	default:
		return element
	}
}

func validatePolicyNumericElementAgainstRange(element any, minVal, maxVal, stepVal *float32) error {
	element = coerceNumericElementForRangeValidator(element)
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
		if !policyRuleOnStep(numeric, float64(*stepVal), anchor) {
			return fmt.Errorf("value %v must be a multiple of %v from %v", numeric, *stepVal, anchor)
		}
	}
	return nil
}

func validatePolicyQuantityRangeBounds(minQty, maxQty, stepQty *resource.Quantity) error {
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

func validatePolicyQuantityValueAgainstRange(q resource.Quantity, minQty, maxQty, stepQty *resource.Quantity) error {
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
		if !policyRuleOnStep(q.AsApproximateFloat64(), stepFloat, anchor) {
			return fmt.Errorf("value %s must be a multiple of %s from %s", q.String(), stepQty.String(), policyQuantityAnchorString(minQty))
		}
	}
	return nil
}

func policyQuantityAnchorString(minQty *resource.Quantity) string {
	if minQty == nil {
		return "0"
	}
	return minQty.String()
}
