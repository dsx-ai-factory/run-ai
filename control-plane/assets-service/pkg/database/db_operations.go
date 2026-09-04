/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package database

const (
	OperationList   = "list"
	OperationGet    = "get"
	OperationCreate = "create"
	OperationUpdate = "update"
	OperationDelete = "delete"
	OperationCount  = "count"
	OperationUpsert = "upsert"

	OperationCommitCreations = "commit/creations"
	OperationCommitUpdates   = "commit/updates"
	OperationCommitDeletions = "commit/deletions"
)
