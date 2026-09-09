/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package steps

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/config"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/logging"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/model"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/step"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/upgrade"
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

	chartPath, valuesPath, err := upgrade.AirgapChartAndValues(s.upgradeCfg, step.ClusterChartName)
	if err != nil {
		logging.Errorf("[Airgap] Failed to resolve the cluster chart: %v", err)
		s.SetStatus(model.Failed, err)
		return err
	}

	fqdn := s.upgradeCfg.FQDN
	values := append(append([]string{}, s.upgradeCfg.ClusterCarryForwardValues...),
		"controlPlane.url="+fqdn,
		step.ClusterAuthValue(s.installInfo.InstallInfo),
		"cluster.uid="+s.upgradeCfg.ClusterUUID,
		"cluster.url="+fqdn,
		clusterImagePullSecretValue,
	)

	logging.Infof("[Upgrade] Upgrading cluster chart from local path %s", chartPath)
	err = s.helmInstaller.InstallOrUpgradeFromLocalChart(ctx,
		chartPath, step.ClusterReleaseName, step.ClusterNamespace,
		values, []string{valuesPath})
	if err != nil {
		logging.Errorf("[Upgrade] Failed to upgrade cluster chart: %v", err)
		s.SetStatus(model.Failed, err)
		return err
	}
	logging.Infof("[Upgrade] Cluster chart upgraded successfully")
	s.SetStatus(model.Installed, nil)
	return nil
}
