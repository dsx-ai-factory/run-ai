/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"sync"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
)

var once sync.Once

func registerConfigurations() {
	if err := config.Register[*Configuration](); err != nil {
		logging.Fatalf("failed to register configuration %+v", err)
	}
	if err := config.Register[*Service](); err != nil {
		logging.Fatalf("failed to register service configuration %+v", err)
	}
	if err := config.Register[*BootConfig](); err != nil {
		logging.Fatalf("failed to register boot configuration %+v", err)
	}
	if err := config.Register[*Backend](); err != nil {
		logging.Fatalf("failed to register backend configuration %+v", err)
	}
}

func InitConfig() {
	logging.Infof("initializing configuration")
	once.Do(func() {
		registerConfigurations()
	})
}
