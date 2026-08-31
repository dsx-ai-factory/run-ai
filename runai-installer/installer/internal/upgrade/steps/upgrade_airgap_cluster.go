/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package steps

import (
	"context"
	"path/filepath"

	"github.com/run-ai/runai/runai-installer/installer/internal/config"
	"github.com/run-ai/runai/runai-installer/installer/internal/logging"
	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	"github.com/run-ai/runai/runai-installer/installer/internal/step"
	"github.com/run-ai/runai/runai-installer/installer/internal/utils/airgap"
)

type UpgradeAirgapClusterChart struct {
	*step.RunAIInstallStepImpl
	helmInstaller step.HelmInstaller
	upgradeCfg    *config.UpgradeConfig
	installInfo   *GetUpgradeInstallInfo
}

func NewUpgradeAirgapClusterChart(
	upgradeCfg *config.UpgradeConfig,
	installInfo *GetUpgradeInstallInfo,
	helmInstaller step.HelmInstaller,
) *UpgradeAirgapClusterChart {
	return &UpgradeAirgapClusterChart{
		RunAIInstallStepImpl: step.NewRunAIInstallStepImpl("Upgrade cluster chart (airgap)", upgradeCfg.StepTimeout),
		helmInstaller:        helmInstaller,
		upgradeCfg:           upgradeCfg,
		installInfo:          installInfo,
	}
}

func (s *UpgradeAirgapClusterChart) Execute(ctx context.Context, _ *model.RunAIInstallData) error {
	if err := s.CheckContextCancellation(ctx); err != nil {
		return err
	}
	s.SetStatus(model.Installing, nil)

	extractedDir := s.upgradeCfg.AirgapExtractedDir
	chartPath := airgap.FindChartFile(extractedDir, step.ClusterChartName)
	customEnvPath := filepath.Join(extractedDir, airgap.CustomEnvFile)

	fqdn := s.upgradeCfg.FQDN
	values := append(append([]string{}, s.upgradeCfg.ClusterCarryForwardValues...),
		"controlPlane.url="+fqdn,
		step.ClusterAuthValue(s.installInfo.InstallInfo),
		"cluster.uid="+s.upgradeCfg.ClusterUUID,
		"cluster.url="+fqdn,
	)

	logging.Infof("[Upgrade] Upgrading cluster chart from local path %s", chartPath)
	err := s.helmInstaller.InstallOrUpgradeFromLocalChart(ctx,
		chartPath, step.ClusterReleaseName, step.ClusterNamespace,
		values, []string{customEnvPath})
	if err != nil {
		logging.Errorf("[Upgrade] Failed to upgrade cluster chart: %v", err)
		s.SetStatus(model.Failed, err)
		return err
	}
	logging.Infof("[Upgrade] Cluster chart upgraded successfully")
	s.SetStatus(model.Installed, nil)
	return nil
}
