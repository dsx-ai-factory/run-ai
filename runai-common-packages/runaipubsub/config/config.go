/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
)

func InitConfig(onConfigChange func()) error {
	if onConfigChange != nil {
		config.OnConfigChange(onConfigChange)
	}

	var err error
	if err = config.SetKbSecretsFolder(""); err != nil {
		return fmt.Errorf("failed to set secrets folder: %w", err)
	}
	if err = config.Register[*Pubsub](); err != nil {
		return fmt.Errorf("failed to register pubsub config: %w", err)
	}

	return nil
}
