/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package constants

type OperationName string

const (
	UpdateOperation OperationName = "update"
	SyncOperation   OperationName = "sync"
	DeleteOperation OperationName = "delete"
	AddOperation    OperationName = "add"
	UpsertOperation OperationName = "upsert"
)
