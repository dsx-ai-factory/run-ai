/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	//"context"

	"flag"

	scheme_init "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/scheme-init"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/shared-objects-controller/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/shared-objects-controller/pkg/controller"
	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	_ "github.com/dsx-ai-factory/run-ai/runai-common-packages/fips"
)

var (
	scheme  = scheme_init.Scheme()
	options = &Options{}
)

type Options struct {
	EnableLeaderElection bool
}

func init() {
	flag.BoolVar(&options.EnableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.Parse()
	initLogging()

	// Register the kai.resources Project/Department types so the controller can watch them.
	if err := kaiorg.AddToScheme(scheme); err != nil {
		log.Fatal().Msgf("Failed to add kai.resources to scheme, error: %s", err.Error())
	}
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
	log.Info().Msg("Starting Run:AI shared-objects-controller")

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:           scheme,
		LeaderElection:   options.EnableLeaderElection,
		LeaderElectionID: "shared-objects-controller.run.ai",
	})
	if err != nil {
		log.Fatal().Msgf("Failed creating manager client, error: %s", err.Error())
	}

	managerContext := ctrl.SetupSignalHandler()
	datavolumeController := controller.NewDataVolumeController(mgr.GetClient())
	if err = datavolumeController.SetupWithManager(managerContext, mgr); err != nil {
		log.Fatal().Msgf("Failed to create datavolume-controller, error: %s", err.Error())
	}

	projectHandlerConfig := config.Config().ProjectHandler
	if projectHandlerConfig.CreateServiceMonitors || projectHandlerConfig.ClusterWideSecrets || projectHandlerConfig.ClusterWidePvcs || projectHandlerConfig.ClusterWideConfigMaps {
		projectController, projectControllerErr := controller.NewProjectController(managerContext, mgr.GetClient())
		if projectControllerErr != nil {
			log.Fatal().Msgf("Failed to create project-controller, error: %v", projectControllerErr)
		}
		if err = projectController.SetupWithManager(mgr); err != nil {
			log.Fatal().Msgf("Failed to setup project-controller, error: %s", err.Error())
		}
	}
	//+kubebuilder:scaffold:builder

	if err = mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		log.Fatal().Msgf("Unable to set up health check, error: %s", err.Error())
	}
	if err = mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		log.Fatal().Msgf("Unable to set up ready check, error: %s", err.Error())
	}

	log.Info().Msg("Starting SharedObjectsController")
	if err = mgr.Start(managerContext); err != nil {
		log.Fatal().Msgf("Problem running manager, error: %s", err.Error())
	}
}
