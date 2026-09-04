/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"fmt"

	dinfw "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferenceworkload/v1alpha1"
	infwv2a1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"sigs.k8s.io/controller-runtime/pkg/cache"

	scheme_init "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/scheme-init"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	diwv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	iwv2a1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	twv2a1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-exporter/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-exporter/pkg/controller/node"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-exporter/pkg/controller/pod"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-exporter/pkg/controller/podGroup"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-exporter/pkg/controller/utils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-exporter/pkg/controller/workload"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-exporter/pkg/metrics"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	_ "github.com/dsx-ai-factory/run-ai/runai-common-packages/fips"
)

var (
	scheme = scheme_init.Scheme()
)

type restMappingResolver interface {
	RESTMapping(schema.GroupKind, ...string) (*meta.RESTMapping, error)
}

type legacyReconcilerRegistration struct {
	gvk        schema.GroupVersionKind
	reconciler utils.Reconciler
}

func initSchemeAndLogging() {
	initLogging()

	utilruntime.Must(corev1.AddToScheme(scheme))
	utilruntime.Must(v1alpha1.AddToScheme(scheme))
	utilruntime.Must(diwv2.AddToScheme(scheme))
	utilruntime.Must(iwv2a1.AddToScheme(scheme))
	utilruntime.Must(twv2a1.AddToScheme(scheme))
	utilruntime.Must(infwv2a1.AddToScheme(scheme))
	utilruntime.Must(dinfw.AddToScheme(scheme))
}

func initLogging() {
	useDebugLogLevel := config.Config().UseDebugLogLevel
	if !useDebugLogLevel {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&zap.Options{Development: useDebugLogLevel})))
}

func main() {
	initSchemeAndLogging()

	log.Info().Msg("Starting Run:AI workload-exporter")

	metrics.InitMetrics()

	clientConfig := ctrl.GetConfigOrDie()
	clientConfig.QPS = float32(config.Config().K8sClientConfigQPS)
	clientConfig.Burst = config.Config().K8sClientConfigBurst

	mgr, err := ctrl.NewManager(clientConfig, ctrl.Options{
		Scheme: scheme,
		Cache: cache.Options{
			ByObject: map[client.Object]cache.ByObject{
				&corev1.ConfigMap{}: {
					Namespaces: map[string]cache.Config{
						constants.NamespaceRunAi: {},
					},
				},
			},
		},
		Metrics: metricsserver.Options{
			BindAddress: fmt.Sprintf(":%s", config.Config().MetricsListenPort),
		},
		LeaderElection:   config.Config().EnableLeaderElection,
		LeaderElectionID: "workload-exporter.run.ai",
	})
	if err != nil {
		log.Fatal().Msgf("Failed creating manager client, error: %s", err.Error())
	}

	legacyReconcilers, err := availableLegacyReconcilers(mgr.GetRESTMapper(), []legacyReconcilerRegistration{
		{
			gvk:        twv2a1.GroupVersion.WithKind("TrainingWorkload"),
			reconciler: workload.NewTrainingWorkloadController(mgr.GetClient()),
		},
		{
			gvk:        iwv2a1.GroupVersion.WithKind("InteractiveWorkload"),
			reconciler: workload.NewInteractiveWorkloadController(mgr.GetClient()),
		},
		{
			gvk:        diwv2.GroupVersion.WithKind("DistributedWorkload"),
			reconciler: workload.NewDistributedWorkloadController(mgr.GetClient()),
		},
		{
			gvk:        infwv2a1.GroupVersion.WithKind("InferenceWorkload"),
			reconciler: workload.NewInferenceWorkloadController(mgr.GetClient()),
		},
		{
			gvk:        dinfw.GroupVersion.WithKind("DistributedInferenceWorkload"),
			reconciler: workload.NewDistributedInferenceWorkloadController(mgr.GetClient()),
		},
	})
	if err != nil {
		log.Fatal().Msgf("Failed to resolve legacy workload CRDs, error: %s", err.Error())
	}

	err = setUpReconcilers(mgr, append([]utils.Reconciler{
		node.NewNodeController(mgr.GetClient()),
		pod.NewPodController(mgr.GetClient()),
		podgroup.NewPodGroupController(mgr.GetClient()),
		workload.NewExternalWorkloadController(mgr.GetClient()),
	}, legacyReconcilers...))
	if err != nil {
		log.Fatal().Msgf("Failed to setup reconcilers with manager, error: %s", err.Error())
	}

	if err = mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		log.Fatal().Msgf("Unable to set up health check, error: %s", err.Error())
	}
	if err = mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		log.Fatal().Msgf("Unable to set up ready check, error: %s", err.Error())
	}

	log.Info().Msg("Starting WorkloadExporter")
	if err = mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		log.Fatal().Msgf("Problem running manager, error: %s", err.Error())
	}
}

func setUpReconcilers(mgr ctrl.Manager, reconcilers []utils.Reconciler) error {
	for _, reconciler := range reconcilers {
		err := reconciler.SetupWithManager(mgr)
		if err != nil {
			return err
		}
	}
	return nil
}

func availableLegacyReconcilers(resolver restMappingResolver, candidates []legacyReconcilerRegistration) ([]utils.Reconciler, error) {
	reconcilers := make([]utils.Reconciler, 0, len(candidates))
	for _, candidate := range candidates {
		_, err := resolver.RESTMapping(candidate.gvk.GroupKind(), candidate.gvk.Version)
		if meta.IsNoMatchError(err) {
			log.Info().Str("gvk", candidate.gvk.String()).Msg("Skipping legacy workload controller because API is not served")
			continue
		}
		if err != nil {
			return nil, err
		}
		reconcilers = append(reconcilers, candidate.reconciler)
	}
	return reconcilers, nil
}
