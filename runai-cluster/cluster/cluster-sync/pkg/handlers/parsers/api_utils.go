/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package parsers

import (
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	flds "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func getApiFloat64NumberField(numberField *flds.NumberField) openapi.NullableFloat64 {
	result := openapi.NullableFloat64{}

	if numberField == nil {
		return result
	}

	result.Set(ptr.To(numberField.Deref()))
	return result
}

func getApiFloat64IntField(numberField *flds.IntField) openapi.NullableFloat64 {
	result := openapi.NullableFloat64{}

	if numberField == nil {
		return result
	}

	result.Set(ptr.To(float64(numberField.Deref())))
	return result
}

// scaleGpuFractionByDevices returns the API value for a per-device GPU fraction
// scaled by the number of GPU devices. When both a fractional value in (0, 1] and
// a device count > 1 are provided, the total workload GPU request is fraction *
// devices. Non-fractional values (nil, <=0, >1) and single-device workloads are
// returned unchanged.
func scaleGpuFractionByDevices(fraction *flds.NumberField, numDevices int32) openapi.NullableFloat64 {
	result := getApiFloat64NumberField(fraction)
	if !result.IsSet() || numDevices <= 1 {
		return result
	}
	value := fraction.Deref()
	if value <= 0 || value > 1 {
		return result
	}
	total := value * float64(numDevices)
	result.Set(&total)
	return result
}

func getApiString(str string) openapi.NullableString {
	result := openapi.NullableString{}
	if str == "" {
		return result
	}
	result.Set(&str)
	return result
}

func getApiTime(time metav1.Time) openapi.NullableTime {
	result := openapi.NullableTime{}
	if !time.IsZero() {
		result.Set(&time.Time)
	}
	return result
}

func strPointerIfNotEmpty(str string) *string {
	if str == "" {
		return nil
	}
	return &str
}

func int32PointerIfNotEmpty(num int32) *int32 {
	if num == 0 {
		return nil
	}
	return &num
}

func getApiFloat64Quantity(quantity *resource.Quantity) openapi.NullableFloat64 {
	result := openapi.NullableFloat64{}

	if quantity == nil {
		return result
	}

	floatQuantity := quantity.AsApproximateFloat64()
	result.Set(&floatQuantity)
	return result
}

func getApiStringQuantity(quantity *resource.Quantity) openapi.NullableString {
	result := openapi.NullableString{}

	if quantity == nil {
		return result
	}

	quantityStr := quantity.String()
	result.Set(&quantityStr)
	return result
}

func getApiInt32(num *int32) openapi.NullableInt32 {
	if num == nil {
		return openapi.NullableInt32{}
	}
	return getApiInt32Value(*num)
}

func getApiInt32Value(num int32) openapi.NullableInt32 {
	result := openapi.NullableInt32{}
	result.Set(&num)
	return result
}

func getApiInt32FromInt64(num *int64) openapi.NullableInt32 {
	if num == nil {
		return openapi.NullableInt32{}
	}
	return getApiInt32Value(int32(*num))
}

//nolint:unused
func _getApiFloat64Value(num float64) openapi.NullableFloat64 {
	result := openapi.NullableFloat64{}
	result.Set(&num)
	return result
}
