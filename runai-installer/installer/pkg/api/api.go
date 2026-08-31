/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package api

import (
	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	"github.com/run-ai/runai/runai-installer/installer/internal/config"
	"github.com/run-ai/runai/runai-installer/installer/internal/flow"
	"github.com/run-ai/runai/runai-installer/installer/internal/utils/apiclient"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator"
)

// ============================================================================
// Installation Event Types
// ============================================================================

// FlowStatusReportingMsg represents installation flow status events
// Clients can receive these events to track installation progress
type FlowStatusReportingMsg = flow.FlowStatusReportingMsg

// InstallPhase represents the current phase/status of an installation step or flow
type InstallPhase = model.InstallPhase

// Installation phase constants
const (
	NotInstalled InstallPhase = model.NotInstalled
	Installing   InstallPhase = model.Installing
	Failed       InstallPhase = model.Failed
	Installed    InstallPhase = model.Installed
	Timeout      InstallPhase = model.Timeout
)

// ============================================================================
// Installer Types
// ============================================================================

// Installer is the main installer instance type
// Re-exported for external use
type Installer = orchestrator.Installer

// ============================================================================
// Installation Data Types
// ============================================================================

// RunAIInstallData contains all installation configuration and state
type RunAIInstallData = model.RunAIInstallData

// Config is the installer configuration
type Config = config.InstallConfig

type GPUType = config.GPUType
type RunaiEnvType = config.RunaiEnvType
type RunaiInstallationsOptions = config.RunaiInstallationsOptions

const (
	GPUTypeReal           GPUType                   = config.GPUTypeReal
	GPUTypeFake           GPUType                   = config.GPUTypeFake
	CPOnly                RunaiInstallationsOptions = config.CPOnly
	SelfHosted            RunaiInstallationsOptions = config.SelfHosted
	SHWithResearcherSetup RunaiInstallationsOptions = config.SHWithResearcherSetup
	Airgap                RunaiInstallationsOptions = config.Airgap
)

// NewRunAIInstallData creates a new installation data instance
func newRunAIInstallData(opts *config.InstallConfig) (*RunAIInstallData, error) {
	if opts == nil {
		opts = &config.InstallConfig{}
	}
	opts.Mode = config.ModeAPI
	opts.Env = config.EnvProd
	return model.NewRunAIInstallData(opts)
}

// ============================================================================
// Global Installer Instance Management
// ============================================================================

// Init initializes the global installer instance with the provided RunAIInstallData
// This should be called once at application startup
// Returns an error if initialization fails
func Init(opts *config.InstallConfig) error {
	runaiData, err := newRunAIInstallData(opts)
	if err != nil {
		return err
	}
	return orchestrator.Init(runaiData)
}

// Get returns the global installer instance
// Panics if Init() or InitWithK8sClient() hasn't been called first
func Get() *orchestrator.Installer {
	return orchestrator.Get()
}

// IsInitialized returns true if the global installer has been initialized
func IsInitialized() bool {
	return orchestrator.IsInitialized()
}

// Reset clears all singleton instances (useful for testing)
func Reset() {
	orchestrator.Reset()
	apiclient.ResetClients()
}
