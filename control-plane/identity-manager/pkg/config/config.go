/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"sync"

	authorization_client "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/utils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
)

var once sync.Once

func registerConfigurations() {
	if err := config.Register[*Iam](); err != nil {
		logging.Fatalf("failed to register iam configuration %+v", err)
	}
	if err := config.Register[*TenantRealm](); err != nil {
		logging.Fatalf("failed to register tenant realm configuration %+v", err)
	}
	if err := config.RegisterRoot[*Notify](); err != nil {
		logging.Fatalf("failed to register notification service configuration %+v", err)
	}

	iam := config.Get[*Iam]()
	iam.IgnoredRealms = append(iam.IgnoredRealms, utils.MASTER_REALM, utils.SSO_DEFAULT_REALM)
	if err := config.Set(iam); err != nil {
		logging.Fatalf("failed to set updated iam configuration %+v", err)
	}

	if err := config.Register[*authorization_client.Config](); err != nil {
		logging.Fatalf("failed to register authorization configuration %+v", err)
	}
	if err := config.Register[*Configuration](); err != nil {
		logging.Fatalf("failed to register configuration %+v", err)
	}
}

func InitConfig() {
	logging.Infof("initializing configuration")
	once.Do(func() {
		registerConfigurations()
	})
}
