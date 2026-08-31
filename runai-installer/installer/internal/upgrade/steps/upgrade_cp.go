/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package steps

import (
	"context"

	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	"github.com/run-ai/runai/runai-installer/installer/internal/config"
	"github.com/run-ai/runai/runai-installer/installer/internal/step"
	"github.com/run-ai/runai/runai-installer/installer/internal/logging"
)

const cpImagePullSecretValue = "global.imagePullSecrets[0].name=runai-reg-creds"

type UpgradeCPChart struct {
	*step.RunAIInstallStepImpl
	helmInstaller step.HelmInstaller
	upgradeCfg    *config.UpgradeConfig
}

func NewUpgradeCPChart(upgradeCfg *config.UpgradeConfig, helmInstaller step.HelmInstaller) *UpgradeCPChart {
	s := &UpgradeCPChart{
		RunAIInstallStepImpl: step.NewRunAIInstallStepImpl("Upgrade Control Plane", upgradeCfg.StepTimeout),
		helmInstaller:        helmInstaller,
		upgradeCfg:           upgradeCfg,
	}
	return s
}

func (s *UpgradeCPChart) Execute(ctx context.Context, runaiData *model.RunAIInstallData) error {
	if err := s.CheckContextCancellation(ctx); err != nil {
		return err
	}
	s.SetStatus(model.Installing, nil)

	version := s.upgradeCfg.ControlPlaneVersion
	values := append(append([]string{}, s.upgradeCfg.CPCarryForwardValues...),
		"global.domain="+s.upgradeCfg.FQDN,
		cpImagePullSecretValue,
	)

	repo := step.CPChartRepo(s.upgradeCfg.Env)
	logging.Infof("[Upgrade] Upgrading CP chart to version %s from %s", version, repo)
	err := s.helmInstaller.InstallOrUpgradeRelease(ctx,
		repo, version, step.CPReleaseName, step.CPChartName, step.CPNamespace, false, values)
	if err != nil {
		s.SetStatus(model.Failed, err)
		return err
	}
	logging.Infof("[Upgrade] CP chart upgraded successfully")
	s.SetStatus(model.Installed, nil)
	return nil
}
