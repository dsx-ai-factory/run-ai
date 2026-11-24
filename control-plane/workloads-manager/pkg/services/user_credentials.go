/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/internal/generated/pkg/api"
	ourConfig "github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
	assetsClient "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client"
	assetsApi "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
)

const credentialSecretIDPrefixLength = 10

type credentialAssetsClient interface {
	ListCredentialsAssets(ctx context.Context, privateOnly bool) (*assetsApi.CredentialsListResponse, *assetsApi.Error)
}

func (w *WorkloadsV2Impl) validateUserCredentials(ctx context.Context, sources []api.WorkloadV2UpdateConfigurationUserCredentialSource) *api.SubmissionErrorV2 {
	client := w.credentialsClient
	if client == nil {
		// The service is constructed before configuration is loaded.
		client = assetsClient.NewAssetsClient(config.Get[*ourConfig.Services]().AssetsServiceUrl)
	}
	credentials, err := client.ListCredentialsAssets(ctx, true)
	if err != nil {
		return NewSubmissionError(http.StatusInternalServerError, err.Message, nil)
	}

	allowedSecretNames := datastructs.NewSet[string]()
	for _, credential := range credentials.Entries {
		secretNames, err := buildCredentialSecretNames(credential.Meta)
		if err != nil {
			return NewSubmissionError(http.StatusInternalServerError, "invalid user credential", err)
		}
		allowedSecretNames.Add(secretNames...)
	}
	// Reject the workload if any requested Secret does not belong to the caller.
	// Use the same 404 for missing and unauthorized credentials to avoid revealing their existence.
	for _, source := range sources {
		if !allowedSecretNames.Has(source.Ref) {
			return NewSubmissionError(http.StatusNotFound, "user credential not found", nil)
		}
	}
	return nil
}

// buildCredentialSecretNames matches Kubernetes Secret names created by the agent
// from database asset metadata. Workload credential refs use these names.
func buildCredentialSecretNames(credential assetsApi.AssetMeta) ([]string, error) {
	// Cluster Secret names include the asset ID prefix to distinguish credentials with the same name.
	assetIDPrefix := credential.Id[:credentialSecretIDPrefixLength]
	baseSecretName := credential.Name + "-" + assetIDPrefix

	switch credential.Kind {
	case assetsApi.ASSETKIND_GENERIC_SECRET, assetsApi.ASSETKIND_DOCKER_REGISTRY:
		return []string{baseSecretName}, nil
	case assetsApi.ASSETKIND_NGC_API_KEY:
		// The agent expands one NGC asset into a Secret for application access and another for image pulls.
		return []string{
			composer.NgcApiKeyGenericSecretPrefix + "-" + baseSecretName,
			composer.NgcApiKeyDockerRegistryPrefix + "-" + baseSecretName,
		}, nil
	default:
		return nil, fmt.Errorf("credential kind %q does not support private user credentials", credential.Kind)
	}
}
