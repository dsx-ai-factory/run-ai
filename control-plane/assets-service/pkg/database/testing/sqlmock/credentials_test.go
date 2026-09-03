/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package sqlmock

import (
	"sort"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/stretchr/testify/require"
)

func metaRec(kind api.AssetKind) dbmodel.AssetsContainedInWorkload {
	return dbmodel.AssetsContainedInWorkload{
		AssetMetaRec: dbmodel.AssetMetaRec{AssetMeta: dbmodel.AssetMeta{Kind: kind}},
	}
}

func envVarWithCredential(assetId string) api.EnvironmentVariableOfAsset {
	return api.EnvironmentVariableOfAsset{
		Credential: &api.EnvironmentVariableCredential{AssetId: convutils.Ref(assetId)},
	}
}

func TestIdentifyCredentialAssets(t *testing.T) {
	tests := []struct {
		name     string
		assets   []dbmodel.AssetsContainedInWorkload
		expected []string
	}{
		{
			name:     "empty input returns empty slice",
			assets:   []dbmodel.AssetsContainedInWorkload{},
			expected: []string{},
		},
		{
			name: "assets without credentials return empty slice",
			assets: []dbmodel.AssetsContainedInWorkload{
				metaRec(api.ASSETKIND_COMPUTE),
				metaRec(api.ASSETKIND_HOST_PATH),
				metaRec(api.ASSETKIND_NFS),
				metaRec(api.ASSETKIND_PVC),
				func() dbmodel.AssetsContainedInWorkload {
					a := metaRec(api.ASSETKIND_S3)
					a.AssetsS3 = &dbmodel.S3AssetRec{}
					return a
				}(),
				func() dbmodel.AssetsContainedInWorkload {
					a := metaRec(api.ASSETKIND_GIT)
					a.AssetsGit = &dbmodel.GitAssetRec{}
					return a
				}(),
				func() dbmodel.AssetsContainedInWorkload {
					a := metaRec(api.ASSETKIND_SECRET_VOLUME)
					a.AssetsSecretVolume = &dbmodel.SecretAssetRec{}
					return a
				}(),
			},
			expected: []string{},
		},
		{
			name: "mix of credential and non-credential assets across kinds",
			assets: []dbmodel.AssetsContainedInWorkload{
				metaRec(api.ASSETKIND_COMPUTE),
				metaRec(api.ASSETKIND_PVC),
				func() dbmodel.AssetsContainedInWorkload {
					a := metaRec(api.ASSETKIND_S3)
					a.AssetsS3 = &dbmodel.S3AssetRec{
						S3AssetSpec: api.S3AssetSpec{AccessKeyAssetId: convutils.Ref("s3-key-1")},
					}
					return a
				}(),
				func() dbmodel.AssetsContainedInWorkload {
					a := metaRec(api.ASSETKIND_GIT)
					a.AssetsGit = &dbmodel.GitAssetRec{
						GitAssetSpec: api.GitAssetSpec{
							CredentialId:    convutils.Ref("git-cred-1"),
							PasswordAssetId: convutils.Ref("git-pass-1"),
						},
					}
					return a
				}(),
				func() dbmodel.AssetsContainedInWorkload {
					a := metaRec(api.ASSETKIND_SECRET_VOLUME)
					a.AssetsSecretVolume = &dbmodel.SecretAssetRec{
						SecretAssetSpec: api.SecretAssetSpec{CredentialAssetId: convutils.Ref("secret-cred-1")},
					}
					return a
				}(),
				func() dbmodel.AssetsContainedInWorkload {
					a := metaRec(api.ASSETKIND_ENVIRONMENT)
					a.AssetsEnvironment = &dbmodel.EnvironmentAssetRec{
						EnvironmentAssetSpec: api.EnvironmentAssetSpec{
							EnvironmentVariables: &[]api.EnvironmentVariableOfAsset{
								envVarWithCredential("env-cred-1"),
								{},
								envVarWithCredential("env-cred-2"),
							},
						},
					}
					return a
				}(),
			},
			expected: []string{
				"env-cred-1", "env-cred-2", "git-cred-1", "git-pass-1", "s3-key-1", "secret-cred-1",
			},
		},
		{
			name: "duplicate credential ids are de-duplicated",
			assets: []dbmodel.AssetsContainedInWorkload{
				func() dbmodel.AssetsContainedInWorkload {
					a := metaRec(api.ASSETKIND_S3)
					a.AssetsS3 = &dbmodel.S3AssetRec{
						S3AssetSpec: api.S3AssetSpec{AccessKeyAssetId: convutils.Ref("dup-cred")},
					}
					return a
				}(),
				func() dbmodel.AssetsContainedInWorkload {
					a := metaRec(api.ASSETKIND_SECRET_VOLUME)
					a.AssetsSecretVolume = &dbmodel.SecretAssetRec{
						SecretAssetSpec: api.SecretAssetSpec{CredentialAssetId: convutils.Ref("dup-cred")},
					}
					return a
				}(),
			},
			expected: []string{"dup-cred"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := dao.IdentifyCredentialAssets(tc.assets)
			sort.Strings(got)
			require.Equal(t, tc.expected, got)
		})
	}
}

func TestGetCredentialIdsFromEnvironmentVariables(t *testing.T) {
	tests := []struct {
		name     string
		input    *[]api.EnvironmentVariableOfAsset
		expected []string
	}{
		{
			name:     "nil input returns nil",
			input:    nil,
			expected: nil,
		},
		{
			name:     "empty slice returns nil",
			input:    &[]api.EnvironmentVariableOfAsset{},
			expected: nil,
		},
		{
			name: "variables without credentials return nil",
			input: &[]api.EnvironmentVariableOfAsset{
				{},
				{},
			},
			expected: nil,
		},
		{
			name: "credential ids preserve order and duplicates",
			input: &[]api.EnvironmentVariableOfAsset{
				envVarWithCredential("cred-1"),
				{},
				envVarWithCredential("cred-2"),
				envVarWithCredential("cred-1"),
			},
			expected: []string{"cred-1", "cred-2", "cred-1"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := dao.GetCredentialIdsFromEnvironmentVariables(tc.input)
			require.Equal(t, tc.expected, got)
		})
	}
}
