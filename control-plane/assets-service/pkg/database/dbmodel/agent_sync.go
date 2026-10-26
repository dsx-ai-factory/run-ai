/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dbmodel

import (
	"context"
	"fmt"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/encryption"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils/gojsonq"
	"github.com/jinzhu/copier"
	"net/http"
)

const (
	NGC_API_KEY       = "NGC_API_KEY"
	NGC_REGISTRY_URL  = "nvcr.io"
	NGC_REGISTRY_USER = "$oauthtoken"
)

// AssetsToSync is a combined metadata+spec for assets that we sync to the cluster
type AssetsToSync struct {
	AssetMetaRec `gorm:"embedded"`

	AssetsPvc *PVCAssetRec `gorm:"foreignKey:AssetId"`

	AssetsAccessKey      *AccessKeyRecord      `gorm:"foreignKey:AssetId"`
	AssetsPassword       *PasswordRecord       `gorm:"foreignKey:AssetId"`
	AssetsGenericSecret  *GenericSecretRecord  `gorm:"foreignKey:AssetId"`
	AssetsDockerRegistry *DockerRegistryRecord `gorm:"foreignKey:AssetId"`
	AssetsNgcApiKey      *NgcApiKeyRecord      `gorm:"foreignKey:AssetId"`

	// AssetsSpec holds the row from the generic assets_spec table used by
	// k8s-object asset kinds (resource claim templates today). Preloaded via
	// TablesOfSyncedAssets so the k8s-objects sync endpoint sees the raw
	// Kubernetes spec alongside the metadata.
	AssetsSpec *AssetSpecRec `gorm:"foreignKey:AssetId"`
}

type EmptySpec struct{}

func (ats *AssetsToSync) ToDatasourceUpdate() api.DatasourceUpdate {
	isDeleted := ats.DeletedAt.Valid

	result := api.DatasourceUpdate{
		Meta:        ats.ToAssetMetaSync(),
		UpdateCount: ats.UpdateCount,
		IsDeleted:   convutils.Ref(isDeleted),
	}

	if ats.AssetsPvc != nil {
		// for deletion, we only need the claim name.
		if isDeleted {
			result.Spec = api.DatasourceUpdateSpec{
				Pvc: &api.PvcAssetSpec{
					ClaimName: ats.AssetsPvc.ClaimName,
				},
			}
		} else {
			// for apply, we need the entire spec.
			result.Spec = api.DatasourceUpdateSpec{
				Pvc: &ats.AssetsPvc.PvcAssetSpec,
			}
		}
	}

	return result
}

// ToK8sObjectsUpdate converts the row into a wire-format update for the
// k8s-objects sync endpoint. On delete the spec is left as the zero value —
// the cluster agent identifies the object by name from meta and does not need
// the full body to remove it (unlike PVC, whose claim name is not the asset
// name and therefore has to be echoed back).
func (ats *AssetsToSync) ToK8sObjectsUpdate() (api.K8sObjectsUpdate, *api.Error) {
	isDeleted := ats.DeletedAt.Valid

	result := api.K8sObjectsUpdate{
		Meta:        ats.ToAssetMetaSync(),
		UpdateCount: ats.UpdateCount,
		IsDeleted:   convutils.Ref(isDeleted),
	}

	if isDeleted || ats.AssetsSpec == nil {
		return result, nil
	}

	if err := FromJsonBMap(ats.AssetsSpec.Spec, &result.Spec); err != nil {
		return api.K8sObjectsUpdate{}, api.NewError(http.StatusInternalServerError,
			fmt.Sprintf("failed to decode k8s object spec of %s asset %s: %v", ats.Kind, ats.Id, err), err)
	}

	return result, nil
}

func (ats *AssetsToSync) ToCredentialsUpdate(ctx context.Context, tenantKeys *encryption.TenantEncryptionKeys) (*api.CredentialsUpdate, *api.Error) {
	isDeleted := ats.DeletedAt.Valid

	result := api.CredentialsUpdate{
		Meta:        ats.ToAssetMetaSync(),
		UpdateCount: ats.UpdateCount,
		IsDeleted:   convutils.Ref(isDeleted),
	}

	if isDeleted {
		return &result, nil
	}

	keysToDecrypt := make([]*string, 0)
	var keyId int32

	if ats.AssetsAccessKey != nil {
		result.Spec.AccessKey = &api.AccessKeyCreationSpec{}
		copier.Copy(result.Spec.AccessKey, &ats.AssetsAccessKey.AccessKeySpec)

		keyId = ats.AssetsAccessKey.KeyId

		if result.Spec.AccessKey.SecretAccessKey != nil {
			keysToDecrypt = append(keysToDecrypt, result.Spec.AccessKey.SecretAccessKey)
		}

	} else if ats.AssetsPassword != nil {
		result.Spec.Password = &api.PasswordCreationSpec{}
		copier.Copy(result.Spec.Password, &ats.AssetsPassword.PasswordSpec)

		keyId = ats.AssetsPassword.KeyId

		if result.Spec.Password.Password != nil {
			keysToDecrypt = append(keysToDecrypt, result.Spec.Password.Password)
		}

	} else if ats.AssetsGenericSecret != nil {
		result.Spec.GenericSecret = &api.GenericSecretCreationSpec{}
		copier.Copy(result.Spec.GenericSecret, &ats.AssetsGenericSecret.GenericSecretSpec)

		keyId = ats.AssetsGenericSecret.KeyId

		if result.Spec.GenericSecret.KeyValuePairs != nil {
			for _, key := range *result.Spec.GenericSecret.KeyValuePairs {
				keysToDecrypt = append(keysToDecrypt, key.Key)
				keysToDecrypt = append(keysToDecrypt, key.Value)
			}
		}

	} else if ats.AssetsDockerRegistry != nil {
		result.Spec.DockerRegistry = &api.DockerRegistryCreationSpec{}
		copier.Copy(result.Spec.DockerRegistry, &ats.AssetsDockerRegistry.DockerRegistrySpec)

		keyId = ats.AssetsDockerRegistry.KeyId

		if result.Spec.DockerRegistry.Password != nil {
			keysToDecrypt = append(keysToDecrypt, result.Spec.DockerRegistry.Password)
		}
	} else if ats.AssetsNgcApiKey != nil {
		keyId = ats.AssetsNgcApiKey.KeyId
		// NGC API key asset is a special case, as we need to create both a docker registry and a generic secret
		// from the same underlying secret.
		result.Spec.DockerRegistry = &api.DockerRegistryCreationSpec{
			User:     convutils.Ref(NGC_REGISTRY_USER),
			Url:      convutils.Ref(NGC_REGISTRY_URL),
			Password: &ats.AssetsNgcApiKey.NgcApiKey,
		}

		result.Spec.GenericSecret = &api.GenericSecretCreationSpec{
			KeyValuePairs: &[]api.KeyValuePair{
				{Key: convutils.Ref(NGC_API_KEY), Value: &ats.AssetsNgcApiKey.NgcApiKey},
			},
		}
		// decrypt the NGC API key once for both specs
		keysToDecrypt = append(keysToDecrypt, &ats.AssetsNgcApiKey.NgcApiKey)
	}

	for _, keyToDecrypt := range keysToDecrypt {
		secret, decryptErr := encryption.DecryptOAEPChunks(ctx, tenantKeys, keyId, keyToDecrypt, ats.Name)
		if decryptErr != nil {
			return nil, decryptErr
		}
		*keyToDecrypt = secret
	}

	return &result, nil
}

func (ats *AssetsToSync) SpecAsJsonBmap() (JsonBMapFromString, *api.Error) {
	var specAsJson map[string]interface{}
	var err error

	if ats.DeletedAt.Valid {
		specAsJson, err = gojsonq.LoadJson(EmptySpec{})
	} else if ats.AssetsPvc != nil {
		specAsJson, err = gojsonq.LoadJson(ats.AssetsPvc.PvcAssetSpec)
	} else if ats.AssetsAccessKey != nil {
		specAsJson, err = gojsonq.LoadJson(ats.AssetsAccessKey.AccessKeySpec)
	} else if ats.AssetsPassword != nil {
		specAsJson, err = gojsonq.LoadJson(ats.AssetsPassword.PasswordSpec)
	} else if ats.AssetsDockerRegistry != nil {
		specAsJson, err = gojsonq.LoadJson(ats.AssetsDockerRegistry.DockerRegistrySpec)
	} else if ats.AssetsGenericSecret != nil {
		specAsJson, err = gojsonq.LoadJson(ats.AssetsGenericSecret.GenericSecretSpec)
	} else if ats.AssetsNgcApiKey != nil {
		specAsJson, err = gojsonq.LoadJson(ats.AssetsNgcApiKey.NgcApiKeySpec)
	} else if ats.AssetsSpec != nil {
		// k8s-object assets store their spec as raw JSON in assets_spec.
		// Return the map as-is; no re-encoding needed.
		return ats.AssetsSpec.Spec, nil
	} else {
		return nil, api.NewError(http.StatusInternalServerError, fmt.Sprintf("missing spec for %s asset %s", ats.Kind, ats.Id), nil)
	}

	if err != nil {
		return nil, api.NewError(http.StatusInternalServerError, fmt.Sprintf("failed to parse spec of %s asset %s: %v", ats.Kind, ats.Id, err), err)
	}

	return specAsJson, nil
}
