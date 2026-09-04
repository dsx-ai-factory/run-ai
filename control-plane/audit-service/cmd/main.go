/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/dsx-ai-factory/run-ai/control-plane/audit-service/internal/auth"
	"github.com/dsx-ai-factory/run-ai/control-plane/audit-service/internal/ingest"
	"github.com/dsx-ai-factory/run-ai/control-plane/audit-service/internal/pubsub"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub"

	audit_config "github.com/dsx-ai-factory/run-ai/control-plane/audit-service/internal/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/audit-service/internal/database"
	"github.com/dsx-ai-factory/run-ai/control-plane/audit-service/internal/router"
	"github.com/dsx-ai-factory/run-ai/control-plane/audit-service/internal/services"
	"k8s.io/utils/env"

	authorization_client "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/metrics"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaihttp"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/tracing"

	_ "github.com/dsx-ai-factory/run-ai/runai-common-packages/fips"
)

var messageConsumer ingest.MessageConsumer = nil

func init() {
	logging.Init(nil)
	_, err := metrics.Init(nil)
	if err != nil {
		logging.Errorf("failed to initialize metrics: %v", err)
	}
	_, err = tracing.Init(nil)
	if err != nil {
		logging.Errorf("failed to initialize tracing: %v", err)
	}
	audit_config.Init()

	err = initPubsub()
	if err != nil {
		logging.Fatalf("Failed to initialize pubsub: error: %v", err)
	}

	runaidb.SetLogger(logging.Actual())
	err = runaidb.Init(nil)
	if err != nil {
		logging.Fatalf("failed to initialize database: %v", err)
	}
	authorization_client.Init(nil)

	if b, _ := env.GetBool("USE_MOCKS", false); b {
		authorization_client.SetMockAuthorizationClient(auth.MockService{})
	}
}

func initPubsub() error {
	errorCallback := func(err error) {
		logging.Errorf("Error in pubsub: %v", err)
	}

	runaipubsub.SetLogger(logging.Actual())
	err := runaipubsub.Init(errorCallback)
	if err != nil {
		return err
	}
	messageConsumer = ingest.NewPubSubConsumer()

	return nil
}

func main() {
	cancelableCtx, cancel := context.WithCancel(context.Background())
	logging.Infof(">>>>>>>>>>> Starting audit service <<<<<<<<<<<<")
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)
	auditDao := database.NewAuditLogsDao()
	authClient := authorization_client.NewAuthorizationClient()
	logsBatcher := ingest.NewBatcher(cancelableCtx, auditDao.Store)
	internalAuditService := services.NewInternalApiAuditLogsService(logsBatcher)
	auditService := services.NewApiAuditLogsService(auditDao, authClient)
	ingestor, err := ingest.NewIngestor(logsBatcher, messageConsumer)
	if err != nil {
		logging.Fatalf(">>>>>>>>>>> Failed to start audit logs consumption <<<<<<<<<<<<\n %v", err)
	}

	subscriber := pubsub.NewSubscriber(auditDao)
	if err := subscriber.SubscribeTenantDelete(cancelableCtx); err != nil {
		logging.Errorf("Failed to subscribe to tenant delete event: %v", err)
	}

	database.InitCleaner(cancelableCtx)

	runaihttp.ServeHTTP(
		router.CreateRouter(auditService, internalAuditService),
		nil,
	)

	<-signalCh
	cancel()
	logging.Infof("received stop signal")

	err = ingestor.Stop()
	if err != nil {
		logging.Errorf("failed to stop logs consumer %v", err)
		os.Exit(1)
	}

	os.Exit(0)
}
