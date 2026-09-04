/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/dbmodels"
)

func (PermissionSet) TableName() string {
	return "permission_sets"
}

type PermissionSet struct {
	dbmodels.DeleteBaseModel
	CreatedBy   string
	UpdatedBy   string
	Name        string
	Description string
	Internal    bool
	Permissions []RolePermission
}
