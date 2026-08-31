/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cli

import (
	openapi "github.com/run-ai/runai/runai-clients/workloads-manager-client/generated"
)

const (
	typeNullableMnnvlModeWorkloadV2 = "mnnvl"
)

type NullableMnnvlModeFlagWorkloadV2 struct {
	value *openapi.NullableMnnvlMode
}

func NewNullableMnnvlModeFlagWorkloadV2(value *openapi.NullableMnnvlMode) *NullableMnnvlModeFlagWorkloadV2 {
	return &NullableMnnvlModeFlagWorkloadV2{value: value}
}

func (mf *NullableMnnvlModeFlagWorkloadV2) String() string {
	if !mf.IsSet() {
		return ""
	}
	return string(*mf.value.Get())
}

func (mf *NullableMnnvlModeFlagWorkloadV2) IsSet() bool {
	return mf.value.IsSet()
}

func (mf *NullableMnnvlModeFlagWorkloadV2) Set(value string) error {
	mnnvlMode, err := openapi.NewMnnvlModeFromValue(value)
	if err != nil {
		return err
	}
	mf.value.Set(mnnvlMode)
	return nil
}

func (mf *NullableMnnvlModeFlagWorkloadV2) Value() *openapi.NullableMnnvlMode {
	return mf.value
}

// Type for CLI help value
func (mf *NullableMnnvlModeFlagWorkloadV2) Type() string {
	return typeNullableMnnvlModeWorkloadV2
}
