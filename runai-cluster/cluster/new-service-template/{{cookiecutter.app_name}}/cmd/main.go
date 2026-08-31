/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	scheme_init "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/scheme-init"
	"github.com/run-ai/runai/runai-cluster/cluster/{{cookiecutter.app_name}}/pkg/cachedclient"
	"github.com/run-ai/runai/runai-cluster/cluster/{{cookiecutter.app_name}}/pkg/config"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	scheme = scheme_init.Scheme()
)

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
	initLogging()

	log.Info().Msg("Starting Run:AI {{cookiecutter.app_name}}")

	//TODO: choose the client you want to work with
	///////////////////// 1. cachedclient
	clientConfig := ctrl.GetConfigOrDie()
	clientConfig.QPS = float32(config.Config().K8sClientConfigQPS)
	clientConfig.Burst = config.Config().K8sClientConfigBurst

	writerClient, err := client.New(clientConfig, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		log.Fatal().Msgf("Failed creating writerClient client, error: %s", err.Error())
	}

	stopper := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		close(stopper)
		cancel()
	}()

	cachedClient := cachedclient.NewCachedClient(ctx, writerClient, clientConfig, scheme)

	{{cookiecutter.__app_name_camel_case_lower}} := pkg.app_name.New{{cookiecutter.__app_name_camel_case_upper}}(cachedClient)
	{{cookiecutter.__app_name_camel_case_lower}}.Start(ctx, stopper)
	<-stopper

	//TODO: either the above cachedclient, or this one - manager client.
	//TODO: if you work with manager client, make sure to delete the "pkg/cachedclient" dir.
	///////////////////// 2. manager client
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
	})
	if err != nil {
		log.Fatal().Msgf("Failed creating manager client, error: %s", err.Error())
	}

	managerContext := ctrl.SetupSignalHandler()
	{{cookiecutter.__app_name_camel_case_lower}} := pkg.app_name.New{{cookiecutter.__app_name_camel_case_upper}}(mgr.GetClient())
	if err = {{cookiecutter.__app_name_camel_case_lower}}.SetupWithManager(managerContext, mgr); err != nil {
		log.Fatal().Msgf("Failed to create {{cookiecutter.app_name}}, error: %s", err.Error())
	}
	//+kubebuilder:scaffold:builder

	if err = mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		log.Fatal().Msgf("Unable to set up health check, error: %s", err.Error())
	}
	if err = mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		log.Fatal().Msgf("Unable to set up ready check, error: %s", err.Error())
	}

	log.Info().Msg("Starting {{cookiecutter.__app_name_camel_case_upper}}")
	if err = mgr.Start(managerContext); err != nil {
		log.Fatal().Msgf("Problem running manager, error: %s", err.Error())
	}
}
