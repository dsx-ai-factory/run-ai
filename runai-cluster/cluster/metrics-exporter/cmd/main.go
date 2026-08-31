/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"flag"
	"log"

	"go.uber.org/zap/zapcore"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/run-ai/runai/runai-cluster/cluster/metrics_exporter/pkg"
	"github.com/run-ai/runai/runai-cluster/cluster/metrics_exporter/pkg/metrics"
	"github.com/run-ai/runai/runai-cluster/cluster/metrics_exporter/pkg/reconcilers"
	"github.com/run-ai/runai/runai-cluster/cluster/metrics_exporter/pkg/utils"

	_ "github.com/run-ai/runai/runai-common-packages/fips"
)

const defaultListenPort = "8080"

func main() {
	var listenPort string
	var newWorkloadMetrics bool
	var enableLeaderElection bool

	flag.StringVar(&listenPort, "listen-port", defaultListenPort,
		"The port to listen on for HTTP requests")
	flag.BoolVar(&newWorkloadMetrics, "new-workload-metrics", false,
		"Does the system export new workload metrics or not")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	klog.InitFlags(flag.CommandLine)
	flag.Parse()

	// leaving the flag there for compatibility - but overriding with "true" always.
	// and not using it anyway - assuming true always.
	newWorkloadMetrics = true

	initLogger()

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Metrics exporter started")

	metrics.InitMetrics()

	mgr, err := pkg.CreateControllerManager(listenPort, enableLeaderElection)
	if err != nil {
		log.Fatalf("Unable to create controller manager, error: %v", err)
	}

	if err = utils.SetIndexersManager(mgr); err != nil {
		log.Fatalf("Unable to create indexers for manager client, error: %v", err)
	}

	client := mgr.GetClient()

	objectReconcilers := []reconcilers.Reconciler{
		reconcilers.NewPodReconciler(client),
		reconcilers.NewPodGroupReconciler(client),
	}

	err = pkg.SetUpReconcilers(mgr, objectReconcilers)

	if err != nil {
		log.Fatalf("Failed to setup reconcilers with manager, error: %v", err)
	}

	if err = mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		log.Fatalf("Failed to start manager, error: %v", err)
	}
}

func initLogger() {
	logOptions := zap.Options{
		Development: true,
		TimeEncoder: zapcore.ISO8601TimeEncoder,
	}
	logOptions.BindFlags(flag.CommandLine)
	flag.Parse()
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&logOptions)))
}
