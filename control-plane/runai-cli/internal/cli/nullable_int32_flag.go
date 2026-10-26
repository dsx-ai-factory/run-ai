/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cli

import (
	"strconv"

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
	i, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		return err
	}

	ni.value.Set(ptr.To(int32(i)))
	return nil
}

func (ni *NullableInt32Flag) Type() string {
	return typeNullableInt32
}
