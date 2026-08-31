/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cli

import (
	openapi "github.com/run-ai/runai/runai-clients/workloads-manager-client/generated"
)

const (
	typeNullablePreemptibilityWorkloadV2 = "preemptibility"
)

type NullablePreemptibilityFlagWorkloadV2 struct {
	value *openapi.NullablePreemptibility
}

func NewNullablePreemptibilityFlagWorkloadV2(value *openapi.NullablePreemptibility) *NullablePreemptibilityFlagWorkloadV2 {
	return &NullablePreemptibilityFlagWorkloadV2{value: value}
}

func (pf *NullablePreemptibilityFlagWorkloadV2) String() string {
	if !pf.IsSet() {
		return ""
	}
	return string(*pf.value.Get())
}

func (pf *NullablePreemptibilityFlagWorkloadV2) IsSet() bool {
	return pf.value.IsSet()
}

func (pf *NullablePreemptibilityFlagWorkloadV2) Set(value string) error {
	preemptibility, err := openapi.NewPreemptibilityFromValue(value)
	if err != nil {
		return err
	}
	pf.value.Set(preemptibility)
	return nil
}

func (pf *NullablePreemptibilityFlagWorkloadV2) Value() *openapi.NullablePreemptibility {
	return pf.value
}

// Type for CLI help value
func (pf *NullablePreemptibilityFlagWorkloadV2) Type() string {
	return typeNullablePreemptibilityWorkloadV2
}
