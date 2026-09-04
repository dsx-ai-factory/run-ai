/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/cluster/meta"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/metrics"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2"
	rcpconfig "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/auth/providers/oidc/configuration/wellknown"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/auth/providers/oidc/userinfo/cacheduserinfoendpointprovider"
	cache_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaihttp"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/cluster/stream"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/router"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/setup"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	_ "github.com/dsx-ai-factory/run-ai/runai-common-packages/fips"
)

const sleepTime = 2 * time.Second

func initLogging(clusterConfig *config.ClusterAPI) {
	useDebugLogLevel := clusterConfig.UseDebugLogLevel
	if !useDebugLogLevel {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&zap.Options{Development: useDebugLogLevel})))
}

func main() {
	clusterConfig := config.Config()

	err := clusterConfig.Validate()
	if err != nil {
		log.Fatal().Msgf("Failed to validate configuration, error: %s", err.Error())
		os.Exit(1)
	}

	initLogging(clusterConfig)

	log.Info().Msg("Starting Run:AI cluster-api")

	setStreamConfig(clusterConfig)

	err = rcpconfig.Register[*meta.Features]()
	if err != nil {
		log.Fatal().Msgf("Failed to register configuration, error: %s", err.Error())
		os.Exit(1)
	}

	initPrometheusMetrics(clusterConfig.EnableMetrics, clusterConfig.MetricsPort)

	clientConfig := ctrl.GetConfigOrDie()
	clientConfig.QPS = float32(clusterConfig.K8sClientConfigQPS)
	clientConfig.Burst = clusterConfig.K8sClientConfigBurst
	scheme := setup.Scheme()

	initCache()
	//
	//   initialize the controller service
	//
	mgr, err := ctrl.NewManager(clientConfig, ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress: "0",
		},
	})
	if err != nil {
		log.Fatal().Msgf("Failed to create cache client, error: %s", err.Error())
		os.Exit(1)
	}

	workloadLocator := setup.WorkloadLocator(mgr.GetClient())
	handlers := setup.CreateHandlerList(mgr.GetClient(), clientConfig, config.Config().WriteTimeout)

	oidcConfigProvider, err := wellknown.New(clusterConfig.IssuerURL)
	if err != nil {
		log.Fatal().Msgf("Failed to fetch oidc configuration, error: %s", err.Error())
		os.Exit(1)
	}

	oidcConfiguration, err := oidcConfigProvider.Configuration()
	if err != nil {
		log.Fatal().Msgf("Failed to parse oidc configuration, error: %s", err.Error())
		os.Exit(1)
	}

	jwksFn, err := setup.InitJWKSFunction(oidcConfiguration.JWKSURI)
	if err != nil {
		log.Fatal().Msgf("Failed to create JWKS function, error: %s", err.Error())
		os.Exit(1)
	}

	userInfoProvider := cacheduserinfoendpointprovider.New(oidcConfigProvider, clusterConfig.CacheExpirationPeriod)

	routerContainer, err := setup.RouterContainer(workloadLocator, handlers, jwksFn,
		oidcConfigProvider, userInfoProvider, clusterConfig, mgr)
	if err != nil {
		log.Fatal().Msgf("Failed to build router container, error: %s", err.Error())
		os.Exit(1)
	}
	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	go func() {
		if err := mgr.Start(ctx); err != nil {
			log.Fatal().Msgf("Failed to start controller manager, error: %s", err.Error())
			os.Exit(1)
		}
	}()

	time.Sleep(sleepTime)
	log.Info().Msg(fmt.Sprintf("Starting HTTP server at :%d", config.Config().Port))

	runaihttp.ServeHTTP(
		router.CreateRouter(routerContainer),
		&runaihttp.ServerConfig{
			ServicePort:  config.Config().Port,
			WriteTimeout: time.Second * config.Config().WriteTimeout,
			ReadTimeout:  time.Second * config.Config().ReadTimeout,
			IdleTimeout:  time.Second * config.Config().IdleTimeout,
		},
	)

	log.Info().Msg("Shutting down cluster-api")
	os.Exit(0)
}

func initCache() {
	if err := cache.Init(&cache_config.Cache{Type: cache_config.CTInMemory}); err != nil {
		log.Error().Msgf("Failed to init cache, error: %s", err.Error())
	}
}

func setStreamConfig(clusterConfig *config.ClusterAPI) {
	streamConfig := stream.GetInstance()
	streamConfig.SetIdleTime(clusterConfig.KeepaliveIdleTimeout)
	streamConfig.SetPingTime(clusterConfig.KeepalivePingTimeout)
}

func initPrometheusMetrics(enabled bool, port int) {
	err := metrics.Init(enabled, port)
	if err != nil {
		log.Fatal().Msgf("Failed to initialize metrics, error: %s", err.Error())
		os.Exit(1)
	}

	if enabled {
		log.Info().Msgf("Prometheus metrics enabled on port: %d", port)
	} else {
		log.Info().Msg("Prometheus metrics disabled")
	}
}
