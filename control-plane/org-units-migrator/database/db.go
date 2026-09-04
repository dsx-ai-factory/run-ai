/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package database

import (
	service_config "github.com/dsx-ai-factory/run-ai/control-plane/org-units-migrator/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
)

var (
	sourceClient orm.DB
	globalClient orm.DB
	targetClient orm.DB
)

func Init() {
	runaidb.SetLogger(logging.Actual())
	if err := runaidb.Init(nil); err != nil {
		logging.Fatalf("failed to init database: %v", err)
	}

	var err error
	if sourceClient, err = runaidb.NewDB(config.Get[*service_config.ConfigData]().SchemaSource); err != nil {
		logging.Fatalf("failed to open source schema connection: %v", err)
	}
	if globalClient, err = runaidb.NewDB(""); err != nil {
		logging.Fatalf("failed to open global connection: %v", err)
	}
	targetClient = runaidb.DB()
}
