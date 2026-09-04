//go:build e2e

/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package e2e_test

import (
	"fmt"
	"os"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/e2e/utils"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/config"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/step"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/upgrade"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/utils/apiclient"
	installerk8s "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/utils/k8sclient"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	defaultInstallVersion     = "2.25.9"
	defaultUpgradeFromVersion = "2.25.7"
	defaultUpgradeVersion     = "2.25.9"
)

func installVersion() string {
	if v := os.Getenv("E2E_INSTALL_VERSION"); v != "" {
		return v
	}
	return defaultInstallVersion
}

func upgradeFromVersion() string {
	if v := os.Getenv("E2E_UPGRADE_FROM_VERSION"); v != "" {
		return v
	}
	return defaultUpgradeFromVersion
}

func upgradeToVersion() string {
	if v := os.Getenv("E2E_UPGRADE_TO_VERSION"); v != "" {
		return v
	}
	return defaultUpgradeVersion
}

func getHelmChartVersion(namespace, releaseName string) (string, error) {
	return utils.GetHelmChartVersion(namespace, releaseName)
}

func upgradeAndVerify(caCertPath string) {
	toVersion := upgradeToVersion()
	By(fmt.Sprintf("Upgrading to version %s", toVersion))
	upgradeCfg := &config.UpgradeConfig{
		DevMode:             true,
		LogsDir:             testEnv.TempDir,
		StepTimeout:         30 * time.Minute,
		ExecutionTimeout:    90 * time.Minute,
		Env:                 utils.RunaiEnv(),
		ControlPlaneVersion: toVersion,
		ClusterVersion:      toVersion,
		UpgradeControlPlane: true,
		UpgradeCluster:      true,
		CACertPath:          caCertPath,
	}

	k8sClients, err := installerk8s.GetK8sClients()
	Expect(err).ToNot(HaveOccurred())
	deps := upgrade.ProductionDeps(k8sClients, upgradeCfg.ControlPlaneVersion)
	err = orchestrator.RunUpgrade(upgradeCfg, deps)
	Expect(err).ToNot(HaveOccurred())

	cpVer, err := getHelmChartVersion(step.CPNamespace, step.CPReleaseName)
	Expect(err).ToNot(HaveOccurred())
	Expect(cpVer).To(Equal(toVersion))

	clusterVer, err := getHelmChartVersion(step.ClusterNamespace, step.ClusterReleaseName)
	Expect(err).ToNot(HaveOccurred())
	Expect(clusterVer).To(Equal(toVersion))
}

var _ = Describe("Upgrade E2E", func() {
	BeforeEach(func() {
		apiclient.ResetClients()
		err := utils.Uninstall(testEnv.TempDir)
		Expect(err).ToNot(HaveOccurred())
	})

	It("should upgrade both CP and cluster", Label("pr"), func() {
		utils.InstallVersion(upgradeFromVersion(), testEnv, nil)
		upgradeAndVerify(testEnv.CAChainPath())
	})
})
