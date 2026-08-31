/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"fmt"

	"github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/logging"
)

const AuthorizationDBSchema = "authorization"

func Init() error {
	logging.Debugf("initializing config")
	if err := registerConfigs(); err != nil {
		return err
	}

	return nil
}

func registerConfigs() error {
	if err := config.Register[*General](); err != nil {
		return fmt.Errorf("failed to register general config: %w", err)
	}

	if err := config.Register[*Server](); err != nil {
		return fmt.Errorf("failed to register server config: %w", err)
	}

	if err := config.Register[*SyncJob](); err != nil {
		return fmt.Errorf("failed to register sync job config: %w", err)
	}

	return nil
}
