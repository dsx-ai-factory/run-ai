/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pkg

import (
	"fmt"
	"log"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/metrics_exporter/pkg/reconcilers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/scheduling/v2alpha2"
	kaiv2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2"
	kubeaischedv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	v1 "k8s.io/api/core/v1"
	k8sschedulingv1 "k8s.io/api/scheduling/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const podSchedulerNameField = "spec.schedulerName"

var scheme = runtime.NewScheme()

// CreateControllerManager initializes the manager with a filtered cache so the exporter
// only watches cluster resources relevant to Run:ai metrics.
func CreateControllerManager(metricsListenPort string, enableLeaderElection bool) (manager.Manager, error) {
	utilruntime.Must(v1.AddToScheme(scheme))
	utilruntime.Must(v2alpha2.AddToScheme(scheme))
	utilruntime.Must(k8sschedulingv1.AddToScheme(scheme))
	utilruntime.Must(kubeaischedv2alpha2.AddToScheme(scheme))
	utilruntime.Must(kaiv2.AddToScheme(scheme))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Cache:  cacheOptions(),
		Metrics: metricsserver.Options{
			BindAddress: fmt.Sprintf(":%s", metricsListenPort),
		},
		LeaderElection:   enableLeaderElection,
		LeaderElectionID: "metrics-exporter.run.ai",
	})

	return mgr, err
}

// cacheOptions restricts the Pod informer at the API server level to avoid caching
// non-Run:ai pods in memory on large clusters.
func cacheOptions() cache.Options {
	return cache.Options{
		ByObject: map[client.Object]cache.ByObject{
			&v1.Pod{}: {
				Field: fields.OneTermEqualSelector(podSchedulerNameField, constants.RunaiScheduler),
			},
		},
	}
}

// SetUpReconcilers registers each reconciler with the manager before starting its cache.
func SetUpReconcilers(mgr ctrl.Manager, reconcilers []reconcilers.Reconciler) error {
	for _, r := range reconcilers {
		err := r.SetupWithManager(mgr)
		if err != nil {
			log.Printf("Failed to setup reconciler with manager, error: %v", err)
			return err
		}
	}
	return nil
}
