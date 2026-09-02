/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package testing_utils

import (
	"context"
	"fmt"
	"os"
	"time"

	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
)

const (
	TestingNamespace       = runv1.DefaultNamespace
	TestingRunaiConfigName = "test"

	timeoutUnitEnvVar = "TIMEOUT_UNIT"
	pollingUnitEnvVar = "POLLING_UNIT"
)

var (
	globalAPIContext = context.TODO()
	TimeoutUnit      = 240 * time.Second
	PollingUnit      = 1 * time.Second
)

func init() {
	readDurationEnvVarIfFound(timeoutUnitEnvVar, &TimeoutUnit)
	readDurationEnvVarIfFound(pollingUnitEnvVar, &PollingUnit)
}

func readDurationEnvVarIfFound(envVar string, varToSet *time.Duration) {
	val, defined := os.LookupEnv(envVar)
	if defined {
		parsedVal, err := time.ParseDuration(val)
		if err != nil {
			panic(fmt.Errorf("failed to parse %s env var: %v", envVar, err))
		}
		*varToSet = parsedVal
	}
}
