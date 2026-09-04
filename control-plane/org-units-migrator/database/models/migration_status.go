/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

type MigrationStatus string

const (
	BEFORE    MigrationStatus = "Before"
	STARTED   MigrationStatus = "Started"
	COMPLETED MigrationStatus = "Completed"
	FAILED    MigrationStatus = "Failed"
)
