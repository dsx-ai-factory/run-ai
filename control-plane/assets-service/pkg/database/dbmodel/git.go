/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dbmodel

import (
	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbtables"
)

type GitAssetRec struct {
	AssetId          string
	api.GitAssetSpec `gorm:"embedded"`
}

func (GitAssetRec) TableName() string {
	return dbtables.TableNameAssetsGit
}

func (g GitAssetRec) GetAssetId() string {
	return g.AssetId
}

func ReplaceGitSpecIds(specs []GitAssetRec, translationMap map[string]string) {
	for idx := range specs {
		specs[idx].AssetId = translationMap[specs[idx].AssetId]
	}
}

type GitAssetUpdate struct {
	api.GitAssetSpec `gorm:"embedded"`
}

func (GitAssetUpdate) TableName() string {
	return dbtables.TableNameAssetsGit
}

// GitCombinedRec holds result of join assets_metadata and assets_git queries
type GitCombinedRec struct {
	api.AssetMeta              `gorm:"embedded"`
	api.GitAssetSpec           `gorm:"embedded"`
	api.WorkloadSupportedTypes `gorm:"embedded"`
}

func (p GitCombinedRec) GetApiAssetMeta() api.AssetMeta {
	return p.AssetMeta
}

func (p GitCombinedRec) GetApiWorkloadSupportedTypes() *api.WorkloadSupportedTypes {
	return &p.WorkloadSupportedTypes
}

// GetGitCredentialId returns the credential id used by the git asset
func GetGitCredentialId(git api.GitAssetSpec) *string {
	// prioritize CredentialId over PasswordAssetId
	if git.CredentialId != nil {
		return git.CredentialId
	}
	// if passwordAssetId is nil, this will return nil as well
	return git.PasswordAssetId
}
