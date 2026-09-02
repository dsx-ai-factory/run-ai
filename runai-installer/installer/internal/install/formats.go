/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package install

import (
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/config"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/step"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/researcher"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/install/components"
)

type RunaiInstallationFormat interface {
	GetComponents() []step.InstallationComponent
	AllowedInModes() []config.InstallerMode
}

type RunaiInstallationCPOnly struct{}

func (r *RunaiInstallationCPOnly) GetComponents() []step.InstallationComponent {
	return []step.InstallationComponent{
		components.NewControlPlane(),
	}
}

func (r *RunaiInstallationCPOnly) AllowedInModes() []config.InstallerMode {
	return []config.InstallerMode{config.ModeAPI}
}

type RunaiInstallationSelfHosted struct{}

func (r *RunaiInstallationSelfHosted) GetComponents() []step.InstallationComponent {
	return []step.InstallationComponent{
		components.NewControlPlane(),
		components.NewRunaiCluster(),
	}
}

func (r *RunaiInstallationSelfHosted) AllowedInModes() []config.InstallerMode {
	return []config.InstallerMode{config.ModeTUI, config.ModeCLI, config.ModeAPI}
}

type RunaiInstallationSHWithResearcherSetup struct{}

func (r *RunaiInstallationSHWithResearcherSetup) GetComponents() []step.InstallationComponent {
	return []step.InstallationComponent{
		components.NewControlPlane(),
		components.NewRunaiCluster(),
		researcher.NewResearcherSetup(),
	}
}

func (r *RunaiInstallationSHWithResearcherSetup) AllowedInModes() []config.InstallerMode {
	return []config.InstallerMode{config.ModeAPI}
}

type RunaiInstallationAirgap struct{}

func (r *RunaiInstallationAirgap) GetComponents() []step.InstallationComponent {
	return []step.InstallationComponent{
		components.NewAirgapPreparation(),
		components.NewControlPlane(),
		components.NewRunaiCluster(),
	}
}

func (r *RunaiInstallationAirgap) AllowedInModes() []config.InstallerMode {
	return []config.InstallerMode{config.ModeCLI, config.ModeAPI}
}
