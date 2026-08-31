/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package postgresdao

import (
	"context"

	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm"
)

const (
	SchemeName = "postgres"
)

func Init() orm.DB {
	logging.Debugf("initializing database")
	return DB()
}

func DB() orm.DB {
	return runaidb.DB()
}

func DBwithCtx(ctx context.Context) orm.DB {
	return DB().WithContext(ctx)
}
