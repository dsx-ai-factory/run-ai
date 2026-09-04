/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package authorization_client

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/helpers"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2/options"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
)

type Config struct {
	Url        string `mapstructure:"authorization_url"`
	IsExternal bool
}

func (c *Config) SetDefaults() {
	c.Url = "http://runai-backend-authorization:8080"
}

var (
	mockClient *IAuthorizationClient
)

func Init(c *Config) {
	if c == nil {
		config.Register[*Config]()
	} else {
		config.Set(c)
	}
	cache.RegisterSharedStore[helpers.Ancestors](helpers.Field, options.MaxRecordSizeKB(1), options.MaxStoreSizeMB(200))
}

func SetMockAuthorizationClient(mc IAuthorizationClient) {
	mockClient = &mc
}
