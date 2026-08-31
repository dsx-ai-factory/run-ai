/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package migration

const (
	NIMMetricsServiceMigrationName = "nim-metrics-service-creation"
	NIMMetricsMigratedAnnotation   = "run.ai/nim-metrics-migrated"
	ChecksumSHA256MigrationName    = "checksum-sha256"
)

var (
	NetworkCleanerMigrationName              = "network-cleaner"
	TopOwnerMigrationName                    = "top-owner"
	ExternalWorkloadByOriginalUIDIndexerName = "externalWorkloadByOriginalUIDIndexer"

	LabelInferenceWorkload string = "run.ai/inferenceworkload"

	paginationLimit int64 = 500
)

func setLimit(l int64) {
	paginationLimit = l
}
