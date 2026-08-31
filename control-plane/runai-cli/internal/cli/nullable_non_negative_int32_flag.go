/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cli

import (
	"fmt"
	"strconv"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"

	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"

	"k8s.io/utils/ptr"
)

const (
	typeNullableNonNegativeInt32 = "positiveInt"
)

type NullableNonNegativeInt32Flag struct {
	value *openapi.NullableInt32
}

func NewNullableNonNegativeInt32Flag(value *openapi.NullableInt32) *NullableNonNegativeInt32Flag {
	return &NullableNonNegativeInt32Flag{
		value: value,
	}
}

func (ni *NullableNonNegativeInt32Flag) IsSet() bool {
	return ni.value.IsSet()
}

func (ni *NullableNonNegativeInt32Flag) String() string {
	if !ni.value.IsSet() {
		return ""
	}
	return strconv.Itoa(int(*ni.value.Get()))
}

func (ni *NullableNonNegativeInt32Flag) Value() *int32 {
	return ni.value.Get()
}

func (ni *NullableNonNegativeInt32Flag) Set(value string) error {
	i, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return err
	}
	if i < 0 {
		return fmt.Errorf(consts.InvalidNegativeValue)
	}
	ni.value.Set(ptr.To(int32(i)))
	return nil
}

func (ni *NullableNonNegativeInt32Flag) Type() string {
	return typeNullableNonNegativeInt32
}
