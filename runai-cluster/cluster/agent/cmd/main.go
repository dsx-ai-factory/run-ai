/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	rcpconfig "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/go-logr/zerologr"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	identitymanagerclient "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg"
	backendclient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/backend-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/validator"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	scheme_init "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/scheme-init"
	dp "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedpolicy/v2alpha1"
	infp "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/inferencepolicy/v2alpha1"
	ip "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/interactivepolicy/v2alpha1"
	tp "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/trainingpolicy/v2alpha1"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"

	dinfw "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferenceworkload/v1alpha1"
	dw "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	infw "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"
	iw "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	tw "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	kaiv1alpha1 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1alpha1"
	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	rbacv1 "k8s.io/api/rbac/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	_ "github.com/dsx-ai-factory/run-ai/runai-common-packages/fips"
)

var (
	scheme = scheme_init.Scheme()
)

func initSchemeAndLogging() {
	initLogging(config.Config().UseDebugLogLevel)

	_ = rbacv1.AddToScheme(scheme)

	_ = tp.AddToScheme(scheme)
	_ = ip.AddToScheme(scheme)
	_ = dp.AddToScheme(scheme)
	_ = infp.AddToScheme(scheme)
	_ = kaiv1alpha1.AddToScheme(scheme)
	_ = kaiorg.AddToScheme(scheme)

	_ = iw.AddToScheme(scheme)
	_ = tw.AddToScheme(scheme)
	_ = dw.AddToScheme(scheme)
	_ = infw.AddToScheme(scheme)
	_ = dinfw.AddToScheme(scheme)
	_ = kartav1alpha1.AddToScheme(scheme)
}

func initLogging(useDebugLogLevel bool) {
	if !useDebugLogLevel {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
}

func main() {
	// --version allows to read CLUSTER_VERSION via `kubectl exec` in scratch images,
	// which do not include a shell.
	versionFlag := flag.Bool("version", false, "print cluster version and exit")
	flag.Parse()
	if *versionFlag {
		fmt.Printf("CLUSTER_VERSION=%s\n", os.Getenv("CLUSTER_VERSION"))
		os.Exit(0)
	}

	err := rcpconfig.SetSecretConfigFile(constants.SecretMountPath)
	if err != nil {
		log.Error().Msgf("Unable to load secret, error: %s", err.Error())
		os.Exit(1)
	}
	err = rcpconfig.Register[*config.AgentConfig]()
	if err != nil {
		log.Error().Msgf("Unable to register configuration, error: %s", err.Error())
		os.Exit(1)
	}
	initSchemeAndLogging()
	log.Info().Msg("Starting Run:AI Agent")
	ctrl.SetLogger(zerologr.New(&log.Logger))
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:           scheme,
		LeaderElection:   config.Config().EnableLeaderElection,
		LeaderElectionID: "agent.run.ai",
	})
	if err != nil {
		log.Error().Msgf("Failed to configure manager: [%v]", err)
		os.Exit(1)
	}

	registerValidatingWebhooks(mgr)

	managerContext := ctrl.SetupSignalHandler()

	go func() {
		<-mgr.Elected()
		log.Info().Msg("Leader elected")
		runAgent(managerContext)
	}()

	if err := mgr.Start(managerContext); err != nil {
		log.Error().Msgf("Manager stopped: %v", err)
		os.Exit(1)
	}
}

// registerValidatingWebhooks wires all validating webhooks hosted by the agent
// onto a single TLS webhook server (cert from init-ca, port from config). The
// server is added as a non-leader-election runnable so every replica serves
// admission traffic, independently of which replica runs the agent sync loop.
//
// Registered paths:
//   - /validate-project      -> projects / departments
func registerValidatingWebhooks(mgr ctrl.Manager) {
	allowedServiceAccounts := splitAndTrim(config.Config().AllowedServiceAccounts)
	projectHandler := validator.NewProjectValidator(mgr.GetClient(), mgr.GetAPIReader(), allowedServiceAccounts)

	webhookServer := webhook.NewServer(webhook.Options{
		Port:    config.Config().WebhookPort,
		CertDir: config.Config().CertDir,
	})
	webhookServer.Register("/validate-project", &webhook.Admission{Handler: projectHandler})

	if err := mgr.Add(webhookServer); err != nil {
		log.Error().Msgf("Unable to add webhook server to manager: %v", err)
		os.Exit(1)
	}
}

// splitAndTrim splits a comma-separated list, trimming whitespace and dropping empty entries.
func splitAndTrim(value string) []string {
	var result []string
	for _, part := range strings.Split(value, ",") {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func initBackendClient(ctx context.Context) *backendclient.BackendClient {
	idmClient := identitymanagerclient.NewIdmClient(config.Config().RunaiApiEndPoint)
	authManager, err := identitymanagerclient.NewAuthnManager(ctx, identitymanagerclient.AuthnManagerConfig{
		ClientId:     config.Config().ClientName,
		ClientSecret: config.Config().ClientSecret,
		IdmClient:    idmClient,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create auth manager")
	}
	return backendclient.NewBackendClient(
		config.Config().RunaiApiEndPoint,
		&backendclient.Options{UseDebugLogLevel: config.Config().UseDebugLogLevel, HttpClient: authManager.HttpClient()})
}

func runAgent(mgrCtx context.Context) {
	runaiBackendClient := initBackendClient(mgrCtx)

	log.Debug().Msg("Created RunaiBackend Client")
	clientConfig := ctrl.GetConfigOrDie()
	clientConfig.QPS = float32(config.Config().K8sClientConfigQPS)
	clientConfig.Burst = config.Config().K8sClientConfigBurst
	cachedClient := common.CreateCache(clientConfig, scheme)

	kubeClient, err := client.New(clientConfig, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		log.Error().Msgf("Unable to create kubeclient, error: %s", err.Error())
		os.Exit(1)
	}

	agent := pkg.NewRunaiAgent(runaiBackendClient, kubeClient, cachedClient, config.Config().AgentSyncIntervalMillis)
	common.StartCache(mgrCtx, cachedClient)
	agent.Start(mgrCtx)

	<-mgrCtx.Done()
}
