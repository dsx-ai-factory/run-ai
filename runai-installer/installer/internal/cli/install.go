/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cli

import (
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/config"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/logging"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator"
	"github.com/spf13/cobra"
)

func newInstallCmd() *cobra.Command {
	var (
		executionTimeout time.Duration
		gpuType          config.GPUType
		env              config.RunaiEnvType
		clusterEnv       config.RunaiEnvType
		cpVersion        string
		clusterVersion   string
		preReq           config.ArrayFlags
		configValues     config.MapFlags = make(config.MapFlags)
		installFormat    config.RunaiInstallationsOptions
		skipControlPlane bool
		skipCluster      bool
	)

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install Run:ai on the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			var clusterEnvPtr *config.RunaiEnvType
			if cmd.Flags().Changed("cluster-env") {
				clusterEnvPtr = &clusterEnv
			}
			cfg := &config.InstallConfig{
				DevMode:            globals.devMode,
				LogsDir:            globals.logsDir,
				StepTimeout:        globals.stepTimeout,
				ExecutionTimeout:   executionTimeout,
				GPUType:            gpuType,
				Env:                env,
				ClusterEnv:         clusterEnvPtr,
				CpVersion:          cpVersion,
				ClusterVersion:     clusterVersion,
				Mode:               config.ModeCLI,
				PreReq:             preReq,
				ConfigValues:       configValues,
				RunaiInstallFormat: installFormat,
				SkipControlPlane:   skipControlPlane,
				SkipCluster:        skipCluster,
			}

			if err := logging.InitFromDir(cfg.LogsDir); err != nil {
				return err
			}
			logBanner("Starting")
			logConfig(cfg)

			k8sClients, err := initK8sClient("")
			if err != nil {
				return err
			}

			runaiData, err := initRunaiData(cfg, k8sClients)
			if err != nil {
				return err
			}

			deps := orchestrator.ProductionInstallDeps(k8sClients, cfg.CpVersion)
			if err := orchestrator.RunInstall(runaiData, deps); err != nil {
				return err
			}

			logBanner("Finished")
			return nil
		},
	}

	cmd.Flags().DurationVar(&executionTimeout, "execution-timeout", 50*time.Minute, "total timeout")
	cmd.Flags().Var(&gpuType, "gpu", "GPU type: real or fake")
	cmd.Flags().Var(&env, "env", "environment: prod or test")
	cmd.Flags().Var(&clusterEnv, "cluster-env", "cluster chart environment (defaults to --env)")
	cmd.Flags().StringVar(&cpVersion, "cp-version", ">0.0.0-0", "CP version")
	cmd.Flags().StringVar(&clusterVersion, "cluster-version", ">0.0.0-0", "Cluster version")
	cmd.Flags().Var(&preReq, "pre-req", "pre-requisite steps")
	cmd.Flags().Var(&configValues, "config-values", "key=value config pairs")
	cmd.Flags().Var(&installFormat, "runai-install-format", "cp-only, self-hosted, sh-with-researcher-setup, airgap")
	cmd.Flags().BoolVar(&skipControlPlane, "skip-cp", false, "skip Control Plane installation")
	cmd.Flags().BoolVar(&skipCluster, "skip-cluster", false, "skip RunAI Cluster installation")

	return cmd
}
