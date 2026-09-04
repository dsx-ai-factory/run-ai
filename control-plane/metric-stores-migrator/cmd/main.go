/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"context"
	"github.com/dsx-ai-factory/run-ai/control-plane/metric-stores-migrator/internal"
	metrics_client "github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/pkg/client"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"os"
	"os/signal"

	_ "github.com/dsx-ai-factory/run-ai/runai-common-packages/fips"
)

func init() {
	logging.Init(nil)
}

func main() {
	logging.Infof("starting to migrate metric stores from tenants manager service")
	signalCh := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(signalCh, os.Interrupt)
	ctx, cancel := context.WithCancel(context.WithValue(context.Background(), runaictx.CfTenantId, "1"))
	err := config.Register[*internal.TenantsConfiguration]()
	if err != nil {
		logging.Errorf("failed to register tenants-manager configuration %v", err)
		os.Exit(1)
	}

	tc := internal.NewTenantsClient()
	mc := metrics_client.GetMetricStoresClient()
	go func() {
		internal.NewMigrator(tc, mc).Migrate(ctx)
		done <- true
	}()

	defer func() {
		logging.Infof("metric stors migration completed")
		os.Exit(0)
	}()

	select {
	case <-signalCh:
		cancel()
		<-done
	case <-done:
	}
}
