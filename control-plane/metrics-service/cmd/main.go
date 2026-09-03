/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	audit "github.com/dsx-ai-factory/run-ai/control-plane/audit-service/pkg/client"
	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/clusters"
	metrics_config "github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/connections"
	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/database"
	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/datasources"
	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/metrics"
	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/metricstores"
	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/providers"
	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/reports"
	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/routers"
	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/services"
	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/tenants"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	runaimetrics "github.com/dsx-ai-factory/run-ai/runai-common-packages/metrics"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaihttp"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/s3"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/tracing"

	_ "github.com/dsx-ai-factory/run-ai/runai-common-packages/fips"
)

const (
	envConfPath = "CONF_PATH"
)

func main() {
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)
	ctx, cancel := context.WithCancel(context.Background())
	logging.Init(nil)
	if err := metrics_config.Init(os.Getenv(envConfPath), ""); err != nil {
		logging.Fatalf("failed to init configuration: %v", err)
	}

	runaidb.SetLogger(logging.Actual())
	if err := runaidb.Init(nil); err != nil {
		logging.Fatalf("failed to init database: %v", err)
	}

	_, err := runaimetrics.Init(nil)
	if err != nil {
		logging.Errorf("failed to initialize metrics: %v", err)
	}
	_, err = tracing.Init(nil)
	if err != nil {
		logging.Errorf("failed to initialize tracing: %v", err)
	}
	err = audit.Init(nil)
	if err != nil {
		logging.Errorf("failed to initialize audit: %v", err)
	}

	conf := config.Get[*metrics_config.ServiceConfiguration]()

	database.InitCleanerIfEnabled(ctx)
	logging.Init(&logging.Config{LogLevel: conf.LogLevel})
	s3.Init(nil)

	if err := cache.Init(nil); err != nil {
		logging.Errorf("failed to initialize cache: %v", err)
	}

	runaipubsub.SetLogger(logging.Actual())
	if _, err := runaipubsub.InitAsync(nil); err != nil {
		logging.Fatalf("failed to initialize pubsub: %v", err)
	}

	metricConfigProcessor := metrics.NewMetricConfigurationProcessor()
	processedMetricsConfigurations, err := metricConfigProcessor.Process()
	if err != nil {
		logging.Fatalf("failed to process metric configurations: %v", err)
	}

	clustersClient := clusters.InitClient(conf.UseMocks)
	queryVersionPicker := metrics.NewQueryVersionPicker()
	metricQueryPicker := metrics.NewMetricQueryPicker(processedMetricsConfigurations)
	injectorChain := buildInjectorChain()
	reportsDao := dao.NewPostgresReportsDao()
	metricStoresDao := dao.NewMetricStoresDao()
	tenantsClient := tenants.InitClient(conf.UseMocks)

	metricStoreManager := metricstores.NewManager(metricStoresDao, tenantsClient, providers.NewDatasourceProvider())
	metricStoreManager.SubscribeToTenantDeleteEvent()

	connectionManager := connections.NewManager(metricStoreManager)
	datasourceClient := datasources.NewClient(connectionManager)

	metricQueryBuilder := metrics.NewPromMetricQueryBuilder(clustersClient, queryVersionPicker, metricQueryPicker, injectorChain)

	processor, deleter, err := startProcessor(ctx, metricQueryBuilder, datasourceClient, clustersClient)
	if err != nil {
		logging.Errorf("failed to start processor: %v", err)
	} else {
		defer processor.WaitStop()
		defer deleter.WaitStop()
	}

	metricsApiController := api.NewMetricsAPIController(services.NewMetricsApiService(metricQueryBuilder, datasourceClient, tenantsClient))
	reportsApiController := api.NewReportsAPIController(services.NewReportsApiService(metricQueryBuilder, reportsDao))
	metricStoresApiController := api.NewMetricStoresAPIController(services.NewMetricStoresAPIService(metricStoreManager))
	healthApiController := api.NewHealthAPIController(services.NewHealthApiService())
	router := routers.CreateRouter(healthApiController, metricsApiController, reportsApiController, metricStoresApiController)

	go runaihttp.ServeHTTP(
		router,
		nil,
	)

	<-signalCh
	cancel()
	runaidb.Close()
}

func startProcessor(
	ctx context.Context, metricQueryBuilder *metrics.PromMetricQueryBuilder,
	datasourceClient *datasources.Client, clustersClient *clusters.Client,
) (*reports.Processor, *reports.Deleter, error) {
	collector, err := reports.NewCollector(metricQueryBuilder, datasourceClient, clustersClient)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create reports collector: %v", err)
	}

	s3Client, err := reports.NewS3Client()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create s3 client: %v", err)
	}

	database.InitReportsCronJob(ctx)

	processor := reports.NewProcessor(collector)
	logging.Infof("starting reports processor, interval is %d sec", config.Get[*reports.ReportsConfiguration]().RunPeriodSec)
	go processor.Start(ctx)

	deleter := reports.NewDeleter(s3Client)
	logging.Infof("starting reports deleter, interval is %d sec", config.Get[*reports.ReportsConfiguration]().RunPeriodSec)
	go deleter.Start(ctx)

	return processor, deleter, nil
}

func buildInjectorChain() metrics.QueryInjector {
	return metrics.NewSimpleFilterInjector(metrics.NewAggregationRangeInjector(metrics.NewQuantileInjector(nil)))
}
