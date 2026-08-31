/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"context"
	"os"

	"go.uber.org/zap/zapcore"

	"github.com/go-logr/zerologr"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/common/dra"
	scheme2 "github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/cmd/scheme"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/cachedclient"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/config"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/externalworkloads"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/kartacontroller"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/utils"
	scheme_init "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/scheme-init"
	controllerCache "sigs.k8s.io/controller-runtime/pkg/cache"

	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	_ "github.com/run-ai/runai/runai-common-packages/fips"
)

var (
	scheme = scheme_init.Scheme()
)

func initLogging(useDebugLogLevel bool) {
	if !useDebugLogLevel {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	zapLoggerOpts := &zap.Options{
		Development:     useDebugLogLevel,
		StacktraceLevel: zapcore.LevelEnabler(zapcore.FatalLevel),
		TimeEncoder:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")}
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(zapLoggerOpts)))
}

func main() {
	initLogging(config.Config().UseDebugLogLevel)
	additionalRiGVKs, unregisteredRiGVKs := scheme2.InitScheme(scheme)

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:           scheme,
		LeaderElection:   config.Config().EnableLeaderElection,
		LeaderElectionID: "external-workload-int.run.ai",
		Cache: controllerCache.Options{
			ByObject: map[client.Object]controllerCache.ByObject{
				&corev1.ConfigMap{}: {
					Namespaces: map[string]controllerCache.Config{
						constants.NamespaceRunAi: {},
					},
				},
			},
		},
	})
	if err != nil {
		log.Error().Msgf("Failed to configure manager: %v", err)
		os.Exit(1)
	}
	ctrl.SetLogger(zerologr.New(&log.Logger))
	managerContext, cancel := context.WithCancel(ctrl.SetupSignalHandler())

	go func() {
		<-mgr.Elected()
		log.Info().Msg("Leader elected")
		runServiceLogic(managerContext, cancel, mgr, additionalRiGVKs, unregisteredRiGVKs)
	}()

	if err = mgr.Start(managerContext); err != nil {
		log.Error().Msgf("Manager stopped: [%v]", err)
		os.Exit(1)
	}
}

func filterCrd(crd *apiextensionsv1.CustomResourceDefinition, allowedGvks map[schema.GroupVersionKind]bool) bool {
	gvk, err := utils.GetGvkFromCrd(crd)
	if err != nil {
		log.Warn().Msgf("%s", err.Error())
		return false
	}
	_, found := allowedGvks[*gvk]
	return found
}

func runServiceLogic(managerContext context.Context, cancel context.CancelFunc, mgr ctrl.Manager, additionalRiGVKs, unregisteredRiGVKs []schema.GroupVersionKind) {
	log.Info().Msg("Starting Run:AI external-workload-integrator")
	clientConfig := ctrl.GetConfigOrDie()
	clientConfig.QPS = float32(config.Config().K8sClientConfigQPS)
	clientConfig.Burst = config.Config().K8sClientConfigBurst

	writerClient, err := client.New(clientConfig, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		log.Fatal().Msgf("Failed creating writerClient client, error: %s", err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		log.Fatal().Msgf("Failed creating kubernetes client, error: %s", err.Error())
	}

	cachedClient := cachedclient.NewCachedClient(managerContext, writerClient, clientConfig, scheme)
	draParser := dra.NewDRAParser(cachedClient, kubeClient)

	setKartaController(managerContext, mgr, additionalRiGVKs, cachedClient)

	_ = externalworkloads.StartExternalWorkloadIntegrator(managerContext, managerContext.Done(),
		cachedClient, draParser, config.Config().AllowNonSupportedTypes, additionalRiGVKs)

	unexistingGvks := cachedClient.GetUnexistingGvks()
	for _, gvk := range unregisteredRiGVKs {
		unexistingGvks[gvk] = true
	}

	crdInformer := cachedClient.SetCrdInformer(managerContext, managerContext.Done())
	_, err = crdInformer.AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			crd, ok := obj.(*apiextensionsv1.CustomResourceDefinition)
			if !ok {
				return false
			}
			return filterCrd(crd, unexistingGvks)
		},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				gvk, innerErr := utils.GetGvkFromCrd(obj.(*apiextensionsv1.CustomResourceDefinition))
				if innerErr != nil {
					log.Warn().Msgf("%s", innerErr.Error())
					return
				}

				log.Info().Msgf("Detected new type in api server: <%s>, exiting...", gvk.String())
				cancel()
			},
		},
	})
	if err != nil {
		log.Fatal().Msgf("Failed to add event handler for crds, error: %s", err.Error())
	} else {
		log.Info().Msgf("Added event handler for crds")
	}

	<-managerContext.Done()
}

func setKartaController(managerContext context.Context, mgr ctrl.Manager, additionalGVKs []schema.GroupVersionKind, cachedClient cachedclient.CachedClient) {
	kartaController := kartacontroller.NewKartaController(
		mgr.GetClient(),
		&kartacontroller.KartaControlParams{
			RBACUpdaterEnabled:       config.Config().WorkloadTypeRBACHandlerEnabled,
			ClusterRoleReadName:      config.Config().ClusterRoleReadNameForRBAC,
			ClusterRoleCrudName:      config.Config().ClusterRoleCrudNameForRBAC,
			RegisteredAdditionalGVKs: additionalGVKs,
			InformersApi:             cachedClient,
			CachedAdditionalGvksApi:  cachedClient,
		})
	err := kartaController.SetupWithManager(managerContext, mgr)
	if err != nil {
		log.Error().Msgf("Error starting KartaController: %v", err)
		os.Exit(1)
	}
}
