/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/application-controller/pkg/controller"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/application-controller/pkg/handler"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/application-controller/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	idmclient "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client"
	scheme_init "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/scheme-init"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/application-controller/pkg/config"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	_ "github.com/dsx-ai-factory/run-ai/runai-common-packages/fips"
)

var scheme = scheme_init.Scheme()

func initLogging(useDebugLogLevel bool) {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if !useDebugLogLevel {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&zap.Options{Development: useDebugLogLevel})))
}

func main() {
	initLogging(config.Config().UseDebugLogLevel)

	log.Info().Msg("Starting Run:AI application-controller")

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress: fmt.Sprintf(":%s", config.Config().MetricsPort),
		},
		LeaderElection:   config.Config().EnableLeaderElection,
		LeaderElectionID: "application-controller.run.ai",
	})
	if err != nil {
		log.Fatal().Msgf("Failed creating manager client, error: %s", err.Error())
	}

	webhook.RegisterAccessRuleWebhook(mgr)

	managerContext := ctrl.SetupSignalHandler()
	applicationController := newApplicationController(managerContext, mgr.GetClient())
	if err = applicationController.SetupWithManager(mgr); err != nil {
		log.Fatal().Msgf("Failed to create application-controller, error: %s", err.Error())
	}
	//+kubebuilder:scaffold:builder

	if err = mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		log.Fatal().Msgf("Unable to set up health check, error: %s", err.Error())
	}
	if err = mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		log.Fatal().Msgf("Unable to set up ready check, error: %s", err.Error())
	}

	go func() {
		<-mgr.Elected()
		// Starting application-handler which is in charge of managing applications according to secrets created in helm installation
		mgrClient := mgr.GetClient()
		mgrCache := mgr.GetCache()
		handler.StartApplicationHandler(managerContext, mgrClient, mgrCache)
	}()

	log.Info().Msg("Starting ApplicationController")
	if err = mgr.Start(managerContext); err != nil {
		log.Fatal().Msgf("Problem running manager, error: %s", err.Error())
	}
}

// newApplicationController builds the controller, enabling authenticated control-plane chart
// downloads when the cluster credentials and endpoint are configured (see cl-creds). When they are
// not configured (feature off, e.g. mid-upgrade) it falls back to the plain controller so unrelated
// applications still reconcile. When they are configured but the initial token fetch fails, it exits
// so the pod restarts and retries once the control plane is reachable — matching the agent and
// cluster-sync startup behaviour, rather than silently disabling authenticated downloads until the
// next restart.
func newApplicationController(ctx context.Context, k8sClient client.Client) *controller.ApplicationController {
	cfg := config.Config()
	if cfg.ClientName == "" || cfg.ClientSecret == "" || cfg.RunaiApiEndPoint == "" {
		log.Warn().Msg("control-plane credentials or endpoint not configured; authenticated chart downloads disabled")
		return controller.NewApplicationController(k8sClient)
	}

	idmClient := idmclient.NewIdmClient(cfg.RunaiApiEndPoint)
	authnManager, err := idmclient.NewAuthnManager(ctx, idmclient.AuthnManagerConfig{
		ClientId:     cfg.ClientName,
		ClientSecret: cfg.ClientSecret,
		IdmClient:    idmClient,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to init control-plane authn manager")
	}

	log.Info().Msg("control-plane authn manager initialized; authenticated chart downloads enabled")
	return controller.NewApplicationControllerWithChartAuth(k8sClient, authnManager.HttpClient())
}
