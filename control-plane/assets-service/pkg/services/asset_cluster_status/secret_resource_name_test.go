/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package asset_cluster_status

import (
	"reflect"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/services/response_patching"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
)

// The credential secret name is the lookup key into the cluster's reported
// objects, and the cluster lowercases the name it builds. These helpers have to
// produce exactly the same string, including for an asset whose name has upper
// case letters — otherwise a healthy asset is reported as not created.
func TestCredentialSecretNamesAreLowercased(t *testing.T) {
	for _, tc := range []struct {
		name string
		kind api.AssetKind
	}{
		{name: "access key", kind: api.ASSETKIND_ACCESS_KEY},
		{name: "password", kind: api.ASSETKIND_PASSWORD},
		{name: "generic secret", kind: api.ASSETKIND_GENERIC_SECRET},
		{name: "docker registry", kind: api.ASSETKIND_DOCKER_REGISTRY},
	} {
		t.Run(tc.name, func(t *testing.T) {
			const assetName = "MyKey"

			want := convutils.Deref(dbmodel.GetSecretName(tc.kind, assetName, nil), "")

			got := getCredentialResourceName(dbmodel.CredentialsAssetSpecs{
				AssetMetaRec: dbmodel.AssetMetaRec{AssetMeta: dbmodel.AssetMeta{Name: assetName, Kind: tc.kind}},
			})
			if got != want {
				t.Fatalf("getCredentialResourceName(%s, %q) = %q, want %q", tc.kind, assetName, got, want)
			}
		})
	}
}

func TestCredentialSecretNameKeepsAnExistingSecretName(t *testing.T) {
	const existing = "MyExistingSecret"

	got := getCredentialResourceName(dbmodel.CredentialsAssetSpecs{
		AssetMetaRec:    dbmodel.AssetMetaRec{AssetMeta: dbmodel.AssetMeta{Name: "MyKey", Kind: api.ASSETKIND_ACCESS_KEY}},
		AssetsAccessKey: &dbmodel.AccessKeyRecord{AccessKeySpec: api.AccessKeySpec{ExistingSecretName: convutils.Ref(existing)}},
	})
	if got != existing {
		t.Fatalf("got %q, want the configured existing secret name %q", got, existing)
	}
}

func TestPatchedSecretNamesAreLowercased(t *testing.T) {
	patchingInfo := response_patching.ResponsePatchingPtrs{
		Meta: &api.AssetMeta{Name: "MyKey"},
		Spec: &api.AccessKeySpec{},
	}

	want := convutils.Deref(dbmodel.GetSecretName(api.ASSETKIND_ACCESS_KEY, "MyKey", nil), "")
	if got := convutils.Deref(getAccessKeyResourceName(patchingInfo), ""); got != want {
		t.Fatalf("getAccessKeyResourceName = %q, want %q", got, want)
	}
}

// An NGC API key is reported as two secrets. The cluster lowercases the shared
// ones, while a private name keeps the asset name as spelled, so the two cases
// must not be normalised the same way.
func TestNGCApiKeyResourceNamesMatchTheController(t *testing.T) {
	const assetName = "MyKey"

	for _, tc := range []struct {
		name    string
		private bool
		want    []string
	}{
		{
			name: "shared",
			want: []string{"genericsecret-ngcgs-mykey", "dockerregistry-ngcdr-mykey"},
		},
		{
			name:    "private",
			private: true,
			want:    []string{"ngcgs-MyKey-0123456789", "ngcdr-MyKey-0123456789"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			patchingInfo := response_patching.ResponsePatchingPtrs{
				Meta: &api.AssetMeta{Name: assetName, Private: convutils.Ref(tc.private), Id: "0123456789abcdef"},
				Spec: &api.NgcApiKeySpec{},
			}

			got := getNGCApiKeyResourcesName(patchingInfo)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("getNGCApiKeyResourcesName(private=%t) = %v, want %v", tc.private, got, tc.want)
			}
		})
	}
}
