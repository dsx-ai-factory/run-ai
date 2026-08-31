/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package orchestrator

import (
	"fmt"

	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	"github.com/run-ai/runai/runai-installer/installer/internal/step"
	"github.com/run-ai/runai/runai-installer/installer/internal/logging"
)

func RunInstall(runaiData *model.RunAIInstallData, deps InstallDeps) error {
	logging.Infof("Starting CLI installation...")

	installerObj, err := NewInstallerWithDeps(runaiData, deps)
	if err != nil {
		return fmt.Errorf("failed to create installer: %w", err)
	}
	step.SetDefaultStepTimeout(runaiData.Config.StepTimeout)
	eventsChan, err := installerObj.RunInstallationWithEventListner()
	if err != nil {
		return fmt.Errorf("failed to start installation: %w", err)
	}

	var firstFailure string
	for event := range eventsChan {
		switch event.FlowStatus {
		case model.Installing:
			logging.Infof("Installing %s...\n", event.FlowName)
		case model.Installed:
			logging.Infof("Installed %s\n", event.FlowName)
		case model.Failed, model.Timeout:
			logging.Errorf("Failed to install %s\n", event.FlowName)
			if firstFailure == "" {
				firstFailure = event.FlowName
			}
		}
	}

	if firstFailure != "" {
		return fmt.Errorf("installation failed: flow %q did not succeed", firstFailure)
	}
	return nil
}
