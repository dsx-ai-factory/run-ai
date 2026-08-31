/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"github.com/stretchr/testify/require"

	openapi "github.com/run-ai/runai/control-plane/tenants-manager/pkg/client/generated"
)

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

var (
	host       = getenv("TM_HOST", "localhost")
	port       = getenv("TM_PORT", "8080")
	dbHost     = getenv("DB_HOST", "localhost")
	dbPort     = getenv("DB_PORT", "5432")
	dbUser     = getenv("DB_USER", "runai")
	dbPassword = getenv("DB_PASSWORD", "password")
	dbName     = getenv("DB_NAME", "runai")
	dbScheme   = getenv("DB_SCHEMA", "tenants_manager")
	dbSslMode  = getenv("DB_SSL_MODE", "disable")
	natsURL    = getenv("NATS_URL", "nats://localhost:4222")
	natsUser   = getenv("NATS_USER", "user")
	natsPass   = getenv("NATS_PASSWORD", "abc123")
	hostUrl    = fmt.Sprintf("http://%s:%s", host, port)

	TestUserName = "test@run.ai"

	OpenapiClient   *openapi.APIClient
	TestStepsClient *TestSteps
)

func newTestContext() context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, TestUserName)
	ctx = context.WithValue(ctx, runaictx.CfIsManagementTenant, "true")
	return ctx
}

func initClients() {
	OpenapiClient = openapi.NewAPIClient(&openapi.Configuration{
		Servers: openapi.ServerConfigurations{{
			URL: hostUrl,
		}},
	})
	TestStepsClient = NewTestSteps(OpenapiClient)
	logging.Init(nil)
}

func TestMain(m *testing.M) {
	flag.Parse()
	if testing.Short() {
		os.Exit(0)
	}
	initClients()

	code := m.Run()

	os.Exit(code)
}

func TestHealth(t *testing.T) {
	t.Parallel()
	health, err := OpenapiClient.HealthAPI.Health(context.Background()).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, health.StatusCode)
}
