/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package steps

import (
	"context"
	"path/filepath"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/config"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/logging"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/model"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/step"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/utils/airgap"
)

type UpgradeAirgapCPChart struct {
	*step.RunAIInstallStepImpl
	helmInstaller step.HelmInstaller
	upgradeCfg    *config.UpgradeConfig
}

func NewUpgradeAirgapCPChart(upgradeCfg *config.UpgradeConfig, helmInstaller step.HelmInstaller) *UpgradeAirgapCPChart {
	return &UpgradeAirgapCPChart{
		RunAIInstallStepImpl: step.NewRunAIInstallStepImpl("Upgrade Control Plane (airgap)", upgradeCfg.StepTimeout),
		helmInstaller:        helmInstaller,
		upgradeCfg:           upgradeCfg,
	}
}

func (s *UpgradeAirgapCPChart) Execute(ctx context.Context, _ *model.RunAIInstallData) error {
	if err := s.CheckContextCancellation(ctx); err != nil {
		return err
	}
	s.SetStatus(model.Installing, nil)

	extractedDir := s.upgradeCfg.AirgapExtractedDir
	chartPath := airgap.FindChartFile(extractedDir, step.CPChartName)
	customEnvPath := filepath.Join(extractedDir, airgap.CustomEnvFile)

	values := append(append([]string{}, s.upgradeCfg.CPCarryForwardValues...),
		"global.domain="+s.upgradeCfg.FQDN,
		cpImagePullSecretValue,
	)

	logging.Infof("[Upgrade] Upgrading CP chart from local path %s", chartPath)
	err := s.helmInstaller.InstallOrUpgradeFromLocalChart(ctx,
		chartPath, step.CPReleaseName, step.CPNamespace,
		values, []string{customEnvPath})
	if err != nil {
		logging.Errorf("[Upgrade] Failed to upgrade CP chart: %v", err)
		s.SetStatus(model.Failed, err)
		return err
	}
	logging.Infof("[Upgrade] CP chart upgraded successfully")
	s.SetStatus(model.Installed, nil)
	return nil
}
