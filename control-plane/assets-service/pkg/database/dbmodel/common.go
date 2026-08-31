/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dbmodel

import (
	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
)

// GetAssetModel returns the model for the given asset kind
func GetAssetModel(kind api.AssetKind) interface{} {
	switch kind {
	case api.ASSETKIND_ACCESS_KEY:
		return &AccessKeyRecord{}
	case api.ASSETKIND_COMPUTE:
		return &ComputeAssetRec{}
	case api.ASSETKIND_CONFIG_MAP:
		return &ConfigMapAssetRec{}
	case api.ASSETKIND_DOCKER_REGISTRY:
		return &DockerRegistryRecord{}
	case api.ASSETKIND_ENVIRONMENT:
		return &EnvironmentAssetRec{}
	case api.ASSETKIND_GENERIC_SECRET:
		return &GenericSecretRecord{}
	case api.ASSETKIND_GIT:
		return &GitAssetRec{}
	case api.ASSETKIND_HOST_PATH:
		return &HostPathAssetRec{}
	case api.ASSETKIND_MODEL:
		return &ModelAssetRec{}
	case api.ASSETKIND_NFS:
		return &NFSAssetRec{}
	case api.ASSETKIND_PASSWORD:
		return &PasswordRecord{}
	case api.ASSETKIND_PVC:
		return &PVCAssetRec{}
	case api.ASSETKIND_REGISTRY:
		return &DockerRegistryRecord{}
	case api.ASSETKIND_SECRET_VOLUME:
		return &SecretAssetRec{}
	case api.ASSETKIND_S3:
		return &S3AssetRec{}
	case api.ASSETKIND_WORKLOAD_TEMPLATE:
		return &TemplateAssetRec{}
	case api.ASSETKIND_NGC_API_KEY:
		return &NgcApiKeyRecord{}
	}
	return nil
}
