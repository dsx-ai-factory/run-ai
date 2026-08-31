/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"sync"

	runaiconfig "github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/logging"
)

var once sync.Once

func InitConfig() {
	once.Do(func() {
		if err := runaiconfig.Register[*Config](); err != nil {
			logging.Fatalf("failed to register gateway-hub configuration: %v", err)
		}
	})
}
