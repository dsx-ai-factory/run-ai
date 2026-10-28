/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"os"

	"github.com/dsx-ai-factory/nvidia-recovery/advice-operator/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/advice-restart-controller/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/advice-restart-controller/pkg/controller"
	scheme_init "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/scheme-init"
	runv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	rcpconfig "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	kubeflowv1 "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	_ "github.com/dsx-ai-factory/run-ai/runai-common-packages/fips"
)

var setupLog = ctrl.Log.WithName("setup")

const controllerName = "advice-restart-controller"

func initLogging() {
	useDebugLogLevel := config.Config().UseDebugLogLevel
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&zap.Options{Development: useDebugLogLevel})))
}

func main() {
	if err := rcpconfig.Register[*config.AdviceRestartControllerConfig](); err != nil {
		setupLog.Error(err, "Unable to register configuration")
		os.Exit(1)
	}

	initLogging()

	setupLog.Info("Starting Run:AI advice-restart-controller")

	scheme, err := initScheme()
	if err != nil {
		setupLog.Error(err, "Failed to create scheme")
		os.Exit(1)
	}

	cfg := config.Config()

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress: cfg.MetricsBindAddress,
		},
	})
	if err != nil {
		setupLog.Error(err, "Failed creating manager client")
		os.Exit(1)
	}

	recorder := mgr.GetEventRecorderFor(controllerName)
	adviceController := controller.NewAdviceController(mgr.GetClient(), recorder, cfg.Mode, cfg.RestartLimit, cfg.MaxRetriesPerAdvice)
	if err = adviceController.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "Failed to create advice-restart-controller")
		os.Exit(1)
	}

	if err = mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "Unable to set up health check")
		os.Exit(1)
	}

	if err = mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "Unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("Starting AdviceRestartController", "mode", cfg.Mode, "restartLimit", cfg.RestartLimit)

	if err = mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "Problem running manager")
		os.Exit(1)
	}
}

func initScheme() (*runtime.Scheme, error) {
	scheme := scheme_init.Scheme()

	if err := v1alpha1.AddToScheme(scheme); err != nil {
		return nil, err
	}

	if err := runv1alpha1.AddToScheme(scheme); err != nil {
		return nil, err
	}

	if err := kubeflowv1.AddToScheme(scheme); err != nil {
		return nil, err
	}

	return scheme, nil
}
