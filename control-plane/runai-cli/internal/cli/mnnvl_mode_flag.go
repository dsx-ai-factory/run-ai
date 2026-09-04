/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cli

import (
	mnnvltypes "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/utils/mnnvl"
)

const (
	typeMnnvlMode = "mnnvl-mode"
)

type MnnvlModeFlag struct {
	value *string
}

func NewMnnvlModeFlag(value *string) *MnnvlModeFlag {
	return &MnnvlModeFlag{value: value}
}

func (f *MnnvlModeFlag) String() string {
	if f.value == nil || *f.value == "" {
		return ""
	}
	return *f.value
}

func (f *MnnvlModeFlag) Set(value string) error {
	mode, err := mnnvltypes.GetMode(value)
	if err != nil {
		return err
	}
	*f.value = string(mode)
	return nil
}

func (f *MnnvlModeFlag) Type() string {
	return typeMnnvlMode
}

func (f *MnnvlModeFlag) IsSet() bool {
	return f.value != nil && *f.value != ""
}
