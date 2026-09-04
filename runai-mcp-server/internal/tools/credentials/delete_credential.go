/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package credentials

import (
	"context"
	"fmt"
	"net/http"

	assets "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/common"
)

func deleteCredential(
	ctx context.Context,
	client *runaiclient.Client,
	in manageCredentialsInput,
	out manageCredentialsOutput,
	current assets.AssetMeta,
) (manageCredentialsOutput, error) {
	out.Diff = common.DiffMaps(out.Before, nil)

	details := common.WithConfirmWarning(
		out.Before,
		"deleting a credential is permanent and may fail while a workload or asset uses it",
	)
	if err := common.ConfirmWrite(ctx, in.Confirm, "delete", "credential", current.Name, current.Id, details); err != nil {
		return out, err
	}

	adapter, err := credentialAdapterByType(in.CredentialType)
	if err != nil {
		return manageCredentialsOutput{}, err
	}
	resp, err := adapter.delete(ctx, client, current.Id)
	if err != nil {
		mapped := runaimcp.MapHTTPError(resp, err)
		if in.IgnoreMissing && common.IsNotFound(mapped) {
			out.Warnings = append(out.Warnings, ignoredMissingCredentialWarning(in))
			return out, nil
		}
		return manageCredentialsOutput{}, fmt.Errorf(
			"delete %s credential %q: %w", in.CredentialType, current.Id, mapped)
	}
	out.Applied = true
	return out, nil
}

func (accessKeyAdapter) delete(ctx context.Context, client *runaiclient.Client, id string) (*http.Response, error) {
	_, resp, err := client.Assets.CredentialsAPI.DeleteAccessKey(ctx, id).Execute()
	return resp, err
}

func (passwordAdapter) delete(ctx context.Context, client *runaiclient.Client, id string) (*http.Response, error) {
	_, resp, err := client.Assets.CredentialsAPI.DeletePassword(ctx, id).Execute()
	return resp, err
}

func (dockerRegistryAdapter) delete(ctx context.Context, client *runaiclient.Client, id string) (*http.Response, error) {
	_, resp, err := client.Assets.CredentialsAPI.DeleteDockerRegistry(ctx, id).Execute()
	return resp, err
}

func (genericSecretAdapter) delete(ctx context.Context, client *runaiclient.Client, id string) (*http.Response, error) {
	return client.Assets.CredentialsAPI.DeleteGenericSecret(ctx, id).Execute()
}

func (ngcAPIKeyAdapter) delete(ctx context.Context, client *runaiclient.Client, id string) (*http.Response, error) {
	return client.Assets.CredentialsAPI.DeleteNgcApiKey(ctx, id).Execute()
}
