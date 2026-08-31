/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package credentials

import (
	"context"
	"fmt"
	"net/http"

	assets "github.com/run-ai/runai/runai-clients/assets-service-client/generated"

	"github.com/run-ai/runai/runai-mcp-server/internal/runaiclient"
	"github.com/run-ai/runai/runai-mcp-server/internal/runaimcp"
	"github.com/run-ai/runai/runai-mcp-server/internal/tools/common"
)

func updateCredential(
	ctx context.Context,
	client *runaiclient.Client,
	in manageCredentialsInput,
	out manageCredentialsOutput,
	current assets.AssetMeta,
) (manageCredentialsOutput, error) {
	preview := current
	preview.SetDescription(*in.Description)
	out.After = credentialSnapshot(preview)
	out.Diff = common.DiffMaps(out.Before, out.After)

	if err := common.ConfirmWrite(ctx, in.Confirm, "update", "credential", current.Name, current.Id, out.Diff); err != nil {
		return out, err
	}

	adapter, err := credentialAdapterByType(in.CredentialType)
	if err != nil {
		return manageCredentialsOutput{}, err
	}
	updated, resp, err := adapter.update(ctx, client, current.Id, *in.Description)
	if err != nil {
		return manageCredentialsOutput{}, fmt.Errorf(
			"update %s credential %q: %w", in.CredentialType, current.Id, runaimcp.MapHTTPError(resp, err))
	}

	out.After = credentialSnapshot(updated)
	out.Diff = common.DiffMaps(out.Before, out.After)
	out.Applied = true
	return out, nil
}

func (accessKeyAdapter) update(
	ctx context.Context,
	client *runaiclient.Client,
	id string,
	description string,
) (assets.AssetMeta, *http.Response, error) {
	meta := credentialUpdateMeta(description)
	result, resp, err := client.Assets.CredentialsAPI.UpdateAccessKey(ctx, id).
		AccessKeyUpdateRequest(assets.AccessKeyUpdateRequest{Meta: meta}).
		Execute()
	if result == nil {
		return assets.AssetMeta{}, resp, err
	}
	return result.Meta, resp, err
}

func (passwordAdapter) update(
	ctx context.Context,
	client *runaiclient.Client,
	id string,
	description string,
) (assets.AssetMeta, *http.Response, error) {
	meta := credentialUpdateMeta(description)
	result, resp, err := client.Assets.CredentialsAPI.UpdatePassword(ctx, id).
		PasswordUpdateRequest(assets.PasswordUpdateRequest{Meta: meta}).
		Execute()
	if result == nil {
		return assets.AssetMeta{}, resp, err
	}
	return result.Meta, resp, err
}

func (dockerRegistryAdapter) update(
	ctx context.Context,
	client *runaiclient.Client,
	id string,
	description string,
) (assets.AssetMeta, *http.Response, error) {
	meta := credentialUpdateMeta(description)
	result, resp, err := client.Assets.CredentialsAPI.UpdateDockerRegistry(ctx, id).
		DockerRegistryUpdateRequest(assets.DockerRegistryUpdateRequest{Meta: meta}).
		Execute()
	if result == nil {
		return assets.AssetMeta{}, resp, err
	}
	return result.Meta, resp, err
}

func (genericSecretAdapter) update(
	ctx context.Context,
	client *runaiclient.Client,
	id string,
	description string,
) (assets.AssetMeta, *http.Response, error) {
	meta := credentialUpdateMeta(description)
	result, resp, err := client.Assets.CredentialsAPI.UpdateGenericSecret(ctx, id).
		GenericSecretUpdateRequest(assets.GenericSecretUpdateRequest{Meta: meta}).
		Execute()
	if result == nil {
		return assets.AssetMeta{}, resp, err
	}
	return result.Meta, resp, err
}

func (ngcAPIKeyAdapter) update(
	ctx context.Context,
	client *runaiclient.Client,
	id string,
	description string,
) (assets.AssetMeta, *http.Response, error) {
	meta := credentialUpdateMeta(description)
	result, resp, err := client.Assets.CredentialsAPI.UpdateNgcApiKey(ctx, id).
		NgcApiKeyUpdateRequest(assets.NgcApiKeyUpdateRequest{Meta: meta}).
		Execute()
	if result == nil {
		return assets.AssetMeta{}, resp, err
	}
	return result.Meta, resp, err
}

func credentialUpdateMeta(description string) *assets.CredentialUpdateRequest {
	meta := assets.NewCredentialUpdateRequest()
	meta.SetDescription(description)
	return meta
}
