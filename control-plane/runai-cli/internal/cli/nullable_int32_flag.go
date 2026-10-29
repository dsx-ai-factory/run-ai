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
	typeNullableInt32 = "int32"
)

type NullableInt32Flag struct {
	value *openapi.NullableInt32
}

func NewNullableInt32Flag(value *openapi.NullableInt32) *NullableInt32Flag {
	return &NullableInt32Flag{
		value: value,
	}
}

func (ni *NullableInt32Flag) IsSet() bool {
	return ni.value.IsSet()
}

func (ni *NullableInt32Flag) String() string {
	if !ni.value.IsSet() {
		return ""
	}
	return strconv.Itoa(int(*ni.value.Get()))
}

func (ni *NullableInt32Flag) Value() *int32 {
	return ni.value.Get()
}

func (ni *NullableInt32Flag) Set(value string) error {
	i, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return err
	}
	// Parsing as uint used to reject negatives as a side effect of the base
	// conversion. Keep rejecting them explicitly now that the value is parsed
	// as a signed int32 that cannot overflow.
	if i < 0 {
		return fmt.Errorf(consts.InvalidNegativeValue)
	}

	ni.value.Set(ptr.To(int32(i)))
	return nil
}

func (ni *NullableInt32Flag) Type() string {
	return typeNullableInt32
}
