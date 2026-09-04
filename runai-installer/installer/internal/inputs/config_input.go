/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package inputs

import (
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/inputs/validation"
)

type validationFunc func(value string) validation.ValidationResult

type ConfigurationItemType int

type ConfigurationItem struct {
	Label           string
	Placeholder     string
	ValidationFunc  validationFunc
	InputIdentifier InstallInputType
	ConfigType      ConfigurationItemType
}

func NewConfigurationItem(label, placeholder string, inputIdentifier InstallInputType, validationFunc validationFunc) *ConfigurationItem {
	return &ConfigurationItem{
		Label:           label,
		Placeholder:     placeholder,
		ValidationFunc:  validationFunc,
		InputIdentifier: inputIdentifier,
	}
}

// InstallInputFromFlag returns the InstallInputType corresponding to the provided flag string.
// The boolean return value indicates whether the flag was recognized.
func InstallInputFromFlag(flag string) (InstallInputType, bool) {
	t, ok := FlagToInstallInput[flag]
	return t, ok
}
