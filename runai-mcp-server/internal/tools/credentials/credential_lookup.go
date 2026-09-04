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
)

type credentialKindAdapter interface {
	userCredentialAllowed() bool
	create(context.Context, *runaiclient.Client, assets.AssetCreationRequest, *credentialSpec) (assets.AssetMeta, *http.Response, error)
	get(context.Context, *runaiclient.Client, string) (assets.AssetMeta, *http.Response, error)
	update(context.Context, *runaiclient.Client, string, string) (assets.AssetMeta, *http.Response, error)
	delete(context.Context, *runaiclient.Client, string) (*http.Response, error)
}

type accessKeyAdapter struct{}
type passwordAdapter struct{}
type dockerRegistryAdapter struct{}
type genericSecretAdapter struct{}
type ngcAPIKeyAdapter struct{}

var credentialKindAdapters = map[string]credentialKindAdapter{
	kindAccessKey:      accessKeyAdapter{},
	kindPassword:       passwordAdapter{},
	kindDockerRegistry: dockerRegistryAdapter{},
	kindGenericSecret:  genericSecretAdapter{},
	kindNGCAPIKey:      ngcAPIKeyAdapter{},
}

func (accessKeyAdapter) userCredentialAllowed() bool      { return false }
func (passwordAdapter) userCredentialAllowed() bool       { return false }
func (dockerRegistryAdapter) userCredentialAllowed() bool { return true }
func (genericSecretAdapter) userCredentialAllowed() bool  { return true }
func (ngcAPIKeyAdapter) userCredentialAllowed() bool      { return true }

func credentialAdapterByType(kind string) (credentialKindAdapter, error) {
	adapter, ok := credentialKindAdapters[kind]
	if !ok {
		return nil, fmt.Errorf(
			"credentialType must be one of: accessKey, password, dockerRegistry, genericSecret, ngcApiKey; got %q: %w",
			kind,
			runaimcp.ErrValidation,
		)
	}
	return adapter, nil
}

func getCredentialMeta(
	ctx context.Context,
	client *runaiclient.Client,
	kind string,
	id string,
) (assets.AssetMeta, error) {
	adapter, err := credentialAdapterByType(kind)
	if err != nil {
		return assets.AssetMeta{}, err
	}
	meta, resp, err := adapter.get(ctx, client, id)
	if err != nil {
		return assets.AssetMeta{}, fmt.Errorf(
			"get %s credential %q: %w", kind, id, runaimcp.MapHTTPError(resp, err))
	}
	return meta, nil
}

func (accessKeyAdapter) get(ctx context.Context, client *runaiclient.Client, id string) (assets.AssetMeta, *http.Response, error) {
	result, resp, err := client.Assets.CredentialsAPI.GetAccessKeyById(ctx, id).Execute()
	if result == nil {
		return assets.AssetMeta{}, resp, err
	}
	return result.Meta, resp, err
}

func (passwordAdapter) get(ctx context.Context, client *runaiclient.Client, id string) (assets.AssetMeta, *http.Response, error) {
	result, resp, err := client.Assets.CredentialsAPI.GetPasswordById(ctx, id).Execute()
	if result == nil {
		return assets.AssetMeta{}, resp, err
	}
	return result.Meta, resp, err
}

func (dockerRegistryAdapter) get(ctx context.Context, client *runaiclient.Client, id string) (assets.AssetMeta, *http.Response, error) {
	result, resp, err := client.Assets.CredentialsAPI.GetDockerRegistryById(ctx, id).Execute()
	if result == nil {
		return assets.AssetMeta{}, resp, err
	}
	return result.Meta, resp, err
}

func (genericSecretAdapter) get(ctx context.Context, client *runaiclient.Client, id string) (assets.AssetMeta, *http.Response, error) {
	result, resp, err := client.Assets.CredentialsAPI.GetGenericSecretById(ctx, id).Execute()
	if result == nil {
		return assets.AssetMeta{}, resp, err
	}
	return result.Meta, resp, err
}

func (ngcAPIKeyAdapter) get(ctx context.Context, client *runaiclient.Client, id string) (assets.AssetMeta, *http.Response, error) {
	result, resp, err := client.Assets.CredentialsAPI.GetNgcApiKeyById(ctx, id).Execute()
	if result == nil {
		return assets.AssetMeta{}, resp, err
	}
	return result.Meta, resp, err
}
