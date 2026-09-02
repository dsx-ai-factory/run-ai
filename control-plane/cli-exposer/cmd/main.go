/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"flag"
	"os"
	"time"

	cli_config "github.com/dsx-ai-factory/run-ai/control-plane/cli-exposer/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/cli-exposer/pkg/router"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaihttp"

	audit "github.com/dsx-ai-factory/run-ai/control-plane/audit-service/pkg/client"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/metrics"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/tracing"

	_ "github.com/dsx-ai-factory/run-ai/runai-common-packages/fips"
)

func init() {
	logging.Init(nil)
	metrics.Init(nil)
	tracing.Init(nil)
	if err := cli_config.InitEnvironment(); err != nil {
		logging.Fatalf("Failed to init configuration: %v", err)
	}
	audit.Init(nil)
}

func main() {
	logging.Infof("Starting Run:AI cli-exposer")

	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.Parse()

	runaihttp.ServeHTTP(
		router.CreateRouter(),
		nil,
	)

	logging.Infof("Shutting down cli-exposer")
	os.Exit(0)
}
