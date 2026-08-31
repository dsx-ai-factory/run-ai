/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package install

import (
	"sync"

	"github.com/run-ai/runai/runai-installer/installer/internal/config"
	"github.com/run-ai/runai/runai-installer/installer/internal/logging"
)

var (
	installationOptionsOnce     sync.Once
	installationOptionsInstance map[config.RunaiInstallationsOptions]RunaiInstallationFormat
)

func GetRunaiInstallationOptions() map[config.RunaiInstallationsOptions]RunaiInstallationFormat {
	installationOptionsOnce.Do(func() {
		installationOptionsInstance = map[config.RunaiInstallationsOptions]RunaiInstallationFormat{
			config.CPOnly:                &RunaiInstallationCPOnly{},
			config.SelfHosted:            &RunaiInstallationSelfHosted{},
			config.SHWithResearcherSetup: &RunaiInstallationSHWithResearcherSetup{},
			config.Airgap:                &RunaiInstallationAirgap{},
		}
	})
	return installationOptionsInstance
}

func GetRunaiInstallFormatFromIdentifier(identifier config.RunaiInstallationsOptions) RunaiInstallationFormat {
	if installationOptionsInstance == nil {
		installationOptionsInstance = GetRunaiInstallationOptions()
	}
	value, ok := installationOptionsInstance[identifier]
	if !ok {
		logging.Infof("installation option: %s not found, returning default SelfHosted", identifier)
		return installationOptionsInstance[config.SelfHosted]
	}
	return value
}
