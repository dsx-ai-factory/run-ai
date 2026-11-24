/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"flag"
	"fmt"
	"os"
	"testing"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
)

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

var (
	host    = getenv("CS_HOST", "localhost")
	port    = getenv("CS_PORT", "8080")
	hostURL = fmt.Sprintf("http://%s:%s", host, port)

	OpenapiClient *openapi.APIClient
)

func initClients() {
	OpenapiClient = openapi.NewAPIClient(&openapi.Configuration{
		Servers: openapi.ServerConfigurations{{URL: hostURL}},
	})
}

func TestMain(m *testing.M) {
	flag.Parse()
	// The suite needs the docker-compose environment, so a plain `go test -short` must not run it.
	if testing.Short() {
		os.Exit(0)
	}
	initClients()

	os.Exit(m.Run())
}
