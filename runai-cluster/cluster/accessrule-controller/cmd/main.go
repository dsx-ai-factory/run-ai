/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"flag"
	"fmt"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/accessrule-controller/internal/config"
	"os"
	"strings"

	"go.uber.org/zap/zapcore"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/accessrule-controller/pkg/controllers/accessrule"
	accessrulevalidator "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/accessrule-controller/pkg/validators/accessrule"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/accessrule-controller/pkg/version"
	scheme_init "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/scheme-init"

	rbacv1 "k8s.io/api/rbac/v1"
	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	//+kubebuilder:scaffold:imports

	_ "github.com/dsx-ai-factory/run-ai/runai-common-packages/fips"
)

var (
	scheme  = scheme_init.Scheme()
	options = &config.Options{}
)

func init() {
	var err error
	options, err = config.SetOptions()
	if err != nil {
		panic(fmt.Sprintf("unable to parse arguments: %v", err))
	}
	initLogging(options.UseDebugLogLevel)

	_ = rbacv1.AddToScheme(scheme)
}

func initLogging(useDebugLogLevel bool) {
	if !useDebugLogLevel {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	// controller log needs to be set up for the underlying webhook code, for example
	opts := zap.Options{
		Development: useDebugLogLevel,
		TimeEncoder: zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))
}

func main() {
	log.Info().Msg("Run:AI Access Rule Controller")

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress: "0", // disable metrics
		},
		LeaderElection:   options.EnableLeaderElection,
		LeaderElectionID: "access-rule.run.ai",
	})
	if err != nil {
		log.Error().Msgf("unable to start manager, error: %s", err.Error())
		os.Exit(1)
	}

	registerAccessRuleWebhook(mgr)

	managerContext := ctrl.SetupSignalHandler()
	accessRuleController := accessrule.NewAccessRuleController(mgr.GetClient())
	if err = accessRuleController.SetupWithManager(managerContext, mgr); err != nil {
		log.Error().Msgf("unable to create accessrule-controller, error: %s", err.Error())
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err = mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		log.Error().Msgf("unable to set up health check, error: %s", err.Error())
		os.Exit(1)
	}
	if err = mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		log.Error().Msgf("unable to set up ready check, error: %s", err.Error())
		os.Exit(1)
	}

	printVersion()
	log.Info().Msg("Starting AccessRuleController")
	if err = mgr.Start(managerContext); err != nil {
		log.Error().Msgf("problem running manager, error: %s", err.Error())
		os.Exit(1)
	}
}

func registerAccessRuleWebhook(mgr manager.Manager) {
	accessRuleHandler := accessrulevalidator.NewAccessRuleValidator(mgr.GetClient(), options.AllowedServiceAccounts)
	webhookServer := webhook.NewServer(webhook.Options{
		Port:    options.WebhookPort,
		CertDir: options.CertDir,
	})
	webhookServer.Register("/validate-access-rule", &webhook.Admission{Handler: accessRuleHandler})
	mgr.Add(webhookServer)
}

func printVersion() {
	versionInfo := version.Version()
	log.Info().Msgf("Version: %s", versionInfo.Version)
	if strings.HasSuffix(versionInfo.Version, "-DEVELOPMENT") {
		return
	}
	if versionInfo.BuildDate != "" {
		log.Info().Msgf("BuildDate: %s", versionInfo.BuildDate)
	}
	if versionInfo.GitCommit != "" {
		log.Info().Msgf("Commit Hash: %s", versionInfo.GitCommit)
	}
	if versionInfo.GitTag != "" {
		log.Info().Msgf("Tag: %s", versionInfo.GitTag)
	}
	log.Info().Msgf("Go Version: %s", versionInfo.GoVersion)
	log.Info().Msgf("Platform: %s", versionInfo.Platform)
}
