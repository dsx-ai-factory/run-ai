/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Run:ai's anyworkload-controller binary: a thin consumer of the dromos library
// that wires the Run:AI / KAI conventions the dromos plugins do not own.
package main

import (
	"fmt"
	"os"
	"time"

	rcpconfig "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/rs/zerolog/log"
	ctrl "sigs.k8s.io/controller-runtime"

	anyworkloadconfig "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/anyworkload-controller/internal/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/anyworkload-controller/internal/mnnvlconv"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/anyworkload-controller/internal/ownership"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/anyworkload-controller/internal/restart"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/anyworkload-controller/plugins/inferenceendpoint"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/anyworkload-controller/plugins/networking"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/anyworkload-controller/plugins/personalcreds"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/anyworkload-controller/plugins/restartrules"
	scheme_init "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/scheme-init"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/app"
	dromosconfig "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/plugins/mnnvl"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/plugins/nodeaffinity"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/plugins/storage"

	_ "github.com/dsx-ai-factory/run-ai/runai-common-packages/fips"
)

func main() {
	if err := run(); err != nil {
		log.Error().Err(err).Msg("anyworkload-controller exited with error")
		os.Exit(1)
	}
}

func run() error {
	if err := rcpconfig.Register[*dromosconfig.DromosConfig](); err != nil {
		return fmt.Errorf("register dromos config: %w", err)
	}
	if err := rcpconfig.Register[*anyworkloadconfig.Config](); err != nil {
		return fmt.Errorf("register anyworkload-controller config: %w", err)
	}
	cfg := rcpconfig.Get[*anyworkloadconfig.Config]()
	networkingIntegration := networking.NewIntegration()

	options := []app.Option{
		app.WithScheme(scheme_init.Scheme()),
		app.WithConfig(rcpconfig.Get[*dromosconfig.DromosConfig]()),
		app.WithPlugins(inferenceendpoint.New, networkingIntegration.Build, personalcreds.New, storage.New, nodeaffinity.New, mnnvl.NewBuilder(
			mnnvl.WithPodAffinityProvider(mnnvlconv.ProvidePodAffinity),
			mnnvl.WithWorkloadLabelStampDisabled(),
		)),
		app.WithManagerSetup(
			networkingIntegration.NewSetup(),
			ownership.NewSetup(),
		),
	}
	if cfg.RestartEnabled {
		log.Info().Msg("restart reconcilers enabled")
		options = append(options,
			app.WithManagerSetup(restart.NewSetup(
				cfg.RestartMaxConcurrentReconciles,
				time.Duration(cfg.RestartFrameworkDeletionTimeoutSec)*time.Second,
			)),
			app.WithPlugins(restartrules.NewBuilder(cfg.RestartAllowedServiceAccounts)),
		)
	}
	return app.RunController(ctrl.SetupSignalHandler(), options...)
}
