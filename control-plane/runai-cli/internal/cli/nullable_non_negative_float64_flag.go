/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cli

import (
	"fmt"
	"strconv"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"

	"k8s.io/utils/ptr"
)

const (
	typeNullableNonNegativeFloat64 = "positiveFloat"
)

type NullableNonNegativeFloat64Flag struct {
	value *openapi.NullableFloat64
}

func NewNullableNonNegativeFloat64Flag(value *openapi.NullableFloat64) *NullableNonNegativeFloat64Flag {
	return &NullableNonNegativeFloat64Flag{
		value: value,
	}
}

func (ni *NullableNonNegativeFloat64Flag) IsSet() bool {
	return ni.value.IsSet()
}

func (ni *NullableNonNegativeFloat64Flag) String() string {
	if !ni.value.IsSet() {
		return ""
	}
	return strconv.FormatFloat(*ni.value.Get(), 'f', -1, 64)
}

func (ni *NullableNonNegativeFloat64Flag) Value() *float64 {
	return ni.value.Get()
}

func (ni *NullableNonNegativeFloat64Flag) Set(value string) error {
	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return err
	}
	if f < 0 {
		return fmt.Errorf(consts.InvalidNegativeValue)
	}
	ni.value.Set(ptr.To(f))
	return nil
}

func (ni *NullableNonNegativeFloat64Flag) Type() string {
	return typeNullableNonNegativeFloat64
}
