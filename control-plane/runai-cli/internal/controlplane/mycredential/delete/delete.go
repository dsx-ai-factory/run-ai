/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package delete

import (
	"fmt"
	"strings"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/mycredential/client"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/mycredential/list"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/transaction"
)

func Delete(txn *transaction.Txn, cpUrl, token, typeFlag, name string) error {
	credType, err := list.CredentialTypeFromFlag(typeFlag)
	if err != nil {
		return err
	}
	if credType == "" {
		return fmt.Errorf("--type is required for delete")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("SECRET_NAME is required and cannot be empty")
	}

	apiClient, err := client.NewClient(cpUrl, token)
	if err != nil {
		return err
	}
	ctx := txn.GetCtx()

	assetId, err := apiClient.GetPrivateCredentialAssetId(ctx, name, string(credType))
	if err != nil {
		return err
	}

	switch credType {
	case list.GenericSecret:
		return apiClient.DeleteGenericSecretById(ctx, assetId)
	case list.DockerRegistry:
		return apiClient.DeleteDockerRegistryById(ctx, assetId)
	case list.NgcApiKey:
		return apiClient.DeleteNgcApiKeyById(ctx, assetId)
	default:
		return fmt.Errorf("unsupported credential type: %s", credType)
	}
}
