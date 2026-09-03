/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dbmodel

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbtables"
)

// ModelFullRec is a full model record, including the asset meta and the model spec.
// Used only for gorm Scan from joins — never JSON-marshalled — so duplicate json
// tags between embedded api.AssetMeta and api.ModelAssetSpec are harmless.
type ModelFullRec struct {
	api.AssetMeta              //nolint:govet // duplicate json tag with ModelAssetSpec; struct is not marshalled
	api.ModelAssetSpec         //nolint:govet // duplicate json tag with AssetMeta; struct is not marshalled
	api.WorkloadSupportedTypes `gorm:"embedded"`
	ComputeAssetId             string
	EnvironmentAssetId         string
}

func (p ModelFullRec) GetApiAssetMeta() api.AssetMeta {
	return p.AssetMeta
}

func (p ModelFullRec) GetApiWorkloadSupportedTypes() *api.WorkloadSupportedTypes {
	return &p.WorkloadSupportedTypes
}

type ModelAssetRec struct {
	AssetId            string
	Title              string
	Description        string
	Thumbnail          string
	HuggingfaceId      string
	ComputeAssetId     string
	EnvironmentAssetId string
	License            string
}

func (ModelAssetRec) TableName() string {
	return dbtables.TableNameModel
}

func (m ModelAssetRec) GetAssetId() string {
	return m.AssetId
}

type ModelAssetUpdate struct {
	api.ModelAssetSpec `gorm:"embedded"`
}

func (ModelAssetUpdate) TableName() string {
	return dbtables.TableNameModel
}
