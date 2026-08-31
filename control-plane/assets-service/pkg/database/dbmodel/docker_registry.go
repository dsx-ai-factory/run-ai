/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dbmodel

import (
	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbtables"
)

type DockerRegistryRecord struct {
	AssetId                string
	api.DockerRegistrySpec `gorm:"embedded"`
}

func (DockerRegistryRecord) TableName() string {
	return dbtables.TableNameAssetsDockerRegistry
}

type DockerRegistryUpdate struct {
	api.DockerRegistrySpec `gorm:"embedded"`
}

func (DockerRegistryUpdate) TableName() string {
	return dbtables.TableNameAssetsDockerRegistry
}

type DockerRegistryFullRecord struct {
	api.AssetMeta              `gorm:"embedded"`
	PrevName                   *string
	api.DockerRegistrySpec     `gorm:"embedded"`
	api.WorkloadSupportedTypes `gorm:"embedded"`
}

func (p DockerRegistryFullRecord) GetApiAssetMeta() api.AssetMeta {
	return p.AssetMeta
}

func (p DockerRegistryFullRecord) GetApiWorkloadSupportedTypes() *api.WorkloadSupportedTypes {
	return &p.WorkloadSupportedTypes
}
