/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package steps

import (
	"context"
	"fmt"
	"time"

	cluster "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/config"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/logging"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/model"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/step"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/upgrade"
)

const upgradeInfoRetryInterval = 15 * time.Second

type GetUpgradeInstallInfo struct {
	*step.RunAIInstallStepImpl
	upgradeCfg    *config.UpgradeConfig
	clusterLookup upgrade.ClusterLookup
	InstallInfo   *cluster.ClusterInstallationInfoResponse
}

func NewGetUpgradeInstallInfo(upgradeCfg *config.UpgradeConfig, clusterLookup upgrade.ClusterLookup) *GetUpgradeInstallInfo {
	return &GetUpgradeInstallInfo{
		RunAIInstallStepImpl: step.NewRunAIInstallStepImpl("Get cluster upgrade info", upgradeCfg.StepTimeout),
		upgradeCfg:           upgradeCfg,
		clusterLookup:        clusterLookup,
	}
}

func (s *GetUpgradeInstallInfo) Execute(ctx context.Context, runaiData *model.RunAIInstallData) error {
	if err := s.CheckContextCancellation(ctx); err != nil {
		return err
	}
	s.SetStatus(model.Installing, nil)

	uuid := s.upgradeCfg.ClusterUUID
	for {
		installInfo, statusCode, err := s.clusterLookup.GetClusterInstallInfoByUUID(
			ctx, uuid, s.upgradeCfg.ClusterVersion, s.upgradeCfg.FQDN)
		if err == nil && statusCode >= 200 && statusCode < 300 {
			s.InstallInfo = installInfo
			logging.Infof("[Upgrade] Retrieved install info for cluster %s", uuid)
			s.SetStatus(model.Installed, nil)
			return nil
		}

		elapsed := s.GetElapsedTime()
		logging.Infof("[Upgrade] CP not yet reachable for cluster %s (elapsed: %v): status=%d err=%v; retrying in %v",
			uuid, elapsed, statusCode, err, upgradeInfoRetryInterval)

		select {
		case <-ctx.Done():
			wrapped := fmt.Errorf("timed out waiting for CP to become reachable for cluster %s: %w", uuid, ctx.Err())
			s.SetStatus(model.Timeout, wrapped)
			return wrapped
		case <-time.After(upgradeInfoRetryInterval):
		}
	}
}

type UpgradeClusterChart struct {
	*step.RunAIInstallStepImpl
	helmInstaller step.HelmInstaller
	upgradeCfg    *config.UpgradeConfig
	installInfo   *GetUpgradeInstallInfo
}

func NewUpgradeClusterChart(
	upgradeCfg *config.UpgradeConfig,
	installInfo *GetUpgradeInstallInfo,
	helmInstaller step.HelmInstaller,
) *UpgradeClusterChart {
	s := &UpgradeClusterChart{
		RunAIInstallStepImpl: step.NewRunAIInstallStepImpl("Upgrade cluster chart", upgradeCfg.StepTimeout),
		helmInstaller:        helmInstaller,
		upgradeCfg:           upgradeCfg,
		installInfo:          installInfo,
	}
	return s
}

func (s *UpgradeClusterChart) Execute(ctx context.Context, runaiData *model.RunAIInstallData) error {
	if err := s.CheckContextCancellation(ctx); err != nil {
		return err
	}
	s.SetStatus(model.Installing, nil)

	fqdn := s.upgradeCfg.FQDN
	values := append(append([]string{}, s.upgradeCfg.ClusterCarryForwardValues...),
		"controlPlane.url="+fqdn,
		step.ClusterAuthValue(s.installInfo.InstallInfo),
		"cluster.uid="+s.upgradeCfg.ClusterUUID,
		"cluster.url="+fqdn,
	)

	version := s.upgradeCfg.ClusterVersion
	repo := step.ClusterChartRepo(s.upgradeCfg.EffectiveClusterEnv())
	if repo == "" {
		err := step.MissingChartRepoError(step.ClusterChartName)
		s.SetStatus(model.Failed, err)
		return err
	}
	logging.Infof("[Upgrade] Upgrading cluster chart to version %s from %s", version, repo)
	err := s.helmInstaller.InstallOrUpgradeRelease(ctx,
		repo, version, step.ClusterReleaseName, step.ClusterChartName, step.ClusterNamespace, false, values)
	if err != nil {
		s.SetStatus(model.Failed, err)
		return err
	}
	logging.Infof("[Upgrade] Cluster chart upgraded successfully")
	s.SetStatus(model.Installed, nil)
	return nil
}
