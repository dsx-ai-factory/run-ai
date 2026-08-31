/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dbmodel

import (
	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbtables"
)

type WorkloadSupportedTypes struct {
	AssetId       string  `json:"asset_id" gorm:"not null;"`
	Workspace     *bool   `json:"workspace" gorm:""`
	Training      *bool   `json:"training" gorm:""`
	Inference     *bool   `json:"inference" gorm:""`
	Distributed   *bool   `json:"distributed" gorm:""`
	DistFramework *string `json:"dist_framework" gorm:""`
}

type WorkloadSupportedTypesRec struct {
	WorkloadSupportedTypes `gorm:"embedded"`
}

// ToApiWorkloadSupportedTypes convert WorkloadSupportedTypesRec to api.WorkloadSupportedTypes
func (w *WorkloadSupportedTypes) ToApiWorkloadSupportedTypes() *api.WorkloadSupportedTypes {
	return &api.WorkloadSupportedTypes{
		Workspace:     w.Workspace,
		Training:      w.Training,
		Inference:     w.Inference,
		Distributed:   w.Distributed,
		DistFramework: w.DistFramework,
	}
}

func (WorkloadSupportedTypesRec) TableName() string {
	return dbtables.TableNameWorkloadSupportedTypes
}

func (WorkloadSupportedTypesRec) AssetIdColumn() string {
	return dbtables.WorkloadSupportedTypesColumnAssetId
}
