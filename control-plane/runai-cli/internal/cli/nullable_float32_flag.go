/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cli

import (
	"fmt"
	"strconv"

	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"

	"k8s.io/utils/ptr"
)

const (
	typeNullableFloat2 = "float32"
)

type NullableFloat32Flag struct {
	value *openapi.NullableFloat32
}

func NewNullableFloat32Flag(value *openapi.NullableFloat32) *NullableFloat32Flag {
	return &NullableFloat32Flag{
		value: value,
	}
}

func (nf *NullableFloat32Flag) String() string {
	if !nf.value.IsSet() {
		return ""
	}
	return fmt.Sprintf("%f", *nf.value.Get())
}

func (nf *NullableFloat32Flag) IsSet() bool {
	return nf.value.IsSet()
}

func (nf *NullableFloat32Flag) Value() *float32 {
	if !nf.IsSet() {
		return nil
	}
	return nf.value.Get()
}

func (nf *NullableFloat32Flag) Set(value string) error {
	i, err := strconv.ParseFloat(value, 32)
	if err != nil {
		return err
	}

	nf.value.Set(ptr.To(float32(i)))
	return nil
}

func (nf *NullableFloat32Flag) Type() string {
	return typeNullableFloat2
}
