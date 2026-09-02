/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package migration

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-migrator/clusters/version_control"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
)

func Migrate(sourceClient orm.DB, targetClient orm.DB) {
	if version_control.IsMigrationDone(targetClient) {
		logging.Infof("Clusters migration done already, finishing.")
		return
	}

	err := copyClusters(sourceClient, targetClient)
	if err != nil {
		logging.Infof("copy clusters failed")
		return
	}

	version_control.MarkMigrationDone(targetClient)
}

func copyClusters(sourceClient orm.DB, targetClient orm.DB) error {
	rows, err := sourceClient.WithContext(context.Background()).Raw(
		"SELECT cluster_uuid, cluster_name, tenant_id, domain, version, last_liveness, created_at, updated_at" +
			" FROM clusters").Rows()
	if err != nil {
		if pgError, ok := errors.AsType[*pgconn.PgError](err); ok && pgError.Code == "42P01" {
			logging.Infof("Tables are not set, this means that migration is not necessary")
			return nil
		}

		logging.Errorf("Failed to query data. err: %+v", err)
		panic(err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			logging.Errorf("Failed to close cursor. err: %+v", err)
		}
	}()

	errCount := 0
	for rows.Next() {
		var row BackendClusterRow
		err := rows.Scan(
			&row.ClusterUUID,
			&row.ClusterName,
			&row.TenantId,
			&row.Domain,
			&row.Version,
			&row.LastLiveness,
			&row.CreatedAt,
			&row.UpdatedAt,
		)
		if err != nil {
			logging.Errorf("Failed to scan data. err: %+v", err)
			errCount += 1
			continue
		}
		logging.Infof("Migrating row: %+v", row)

		CreateV2ClusterModel(targetClient, row)
	}
	if err := rows.Err(); err != nil {
		// Until all teams fix their code, at the moment, preventing unsafe queries can cause catastrophic behavior, so we just log it for now, but in the near future, the entire function will be removed.
		logging.Errorf("Error iterating migration rows: %+v", err)
	}

	logging.Infof("copying cluster succeeded, with %d errors", errCount)
	return nil
}
