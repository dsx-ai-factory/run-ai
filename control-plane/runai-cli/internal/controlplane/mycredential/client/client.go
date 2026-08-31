/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/api/config"
	clierrors "github.com/run-ai/runai/control-plane/runai-cli/internal/errors"
	assetsgenerated "github.com/run-ai/runai/runai-clients/assets-service-client/generated"
)

type KeyValue struct {
	Key   string
	Value string
}

type MyCredentialEntry struct {
	Type        string `json:"type,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

const (
	kindGenericSecret  = "genericSecret"
	kindDockerRegistry = "dockerRegistry"
	kindNgcApiKey      = "ngcApiKey"
)

type Client struct {
	apiClient *assetsgenerated.APIClient
}

func NewClient(domain, token string) (*Client, error) {
	assetsConfig, err := config.CreateAssetsServiceConfig(domain, token)
	if err != nil {
		return nil, fmt.Errorf("failed to create assets service config: %w", err)
	}

	return &Client{
		apiClient: assetsgenerated.NewAPIClient(assetsConfig),
	}, nil
}

func (c *Client) ListUserCredentials(ctx context.Context, typeFilter string) ([]MyCredentialEntry, error) {
	req := c.apiClient.CredentialsAPI.ListCredentialsAssets(ctx).PrivateOnly(true)
	if typeFilter != "" {
		req = req.Type_(typeFilter)
	}
	resp, httpResp, err := req.Execute()
	defer func() {
		if httpResp != nil && httpResp.Body != nil {
			httpResp.Body.Close()
		}
	}()
	if err != nil {
		return nil, clierrors.HandleResponseError(httpResp, err)
	}
	if resp == nil || len(resp.Entries) == 0 {
		return []MyCredentialEntry{}, nil
	}
	entries := make([]MyCredentialEntry, len(resp.Entries))
	for i, e := range resp.Entries {
		meta := e.Meta
		entries[i] = MyCredentialEntry{
			Type:        string(meta.Kind),
			Name:        meta.Name,
			Description: meta.GetDescription(),
		}
	}
	return entries, nil
}

func (c *Client) GetPrivateCredentialAssetId(ctx context.Context, name, typeFilter string) (string, error) {
	req := c.apiClient.CredentialsAPI.ListCredentialsAssets(ctx).PrivateOnly(true).Name(name).Type_(typeFilter)
	resp, httpResp, err := req.Execute()
	defer func() {
		if httpResp != nil && httpResp.Body != nil {
			httpResp.Body.Close()
		}
	}()
	if err != nil {
		return "", clierrors.HandleResponseError(httpResp, err)
	}
	if resp == nil || len(resp.Entries) == 0 {
		return "", fmt.Errorf("personal credential '%s' of type %s not found", name, typeFilter)
	}
	return resp.Entries[0].Meta.GetId(), nil
}

func (c *Client) DeleteGenericSecretById(ctx context.Context, assetId string) error {
	return c.deleteCredentialById(ctx, assetId, kindGenericSecret)
}

func (c *Client) DeleteDockerRegistryById(ctx context.Context, assetId string) error {
	return c.deleteCredentialById(ctx, assetId, kindDockerRegistry)
}

func (c *Client) DeleteNgcApiKeyById(ctx context.Context, assetId string) error {
	return c.deleteCredentialById(ctx, assetId, kindNgcApiKey)
}

func (c *Client) deleteCredentialById(ctx context.Context, assetId string, kind string) error {
	var httpResp *http.Response
	var err error
	switch kind {
	case kindGenericSecret:
		httpResp, err = c.apiClient.CredentialsAPI.DeleteGenericSecret(ctx, assetId).Execute()
	case kindDockerRegistry:
		_, httpResp, err = c.apiClient.CredentialsAPI.DeleteDockerRegistry(ctx, assetId).Execute()
	case kindNgcApiKey:
		httpResp, err = c.apiClient.CredentialsAPI.DeleteNgcApiKey(ctx, assetId).Execute()
	default:
		return fmt.Errorf("unsupported credential kind: %s", kind)
	}
	defer func() {
		if httpResp != nil && httpResp.Body != nil {
			httpResp.Body.Close()
		}
	}()
	return clierrors.HandleResponseError(httpResp, err)
}

func (c *Client) CreateGenericSecret(ctx context.Context, name, description string, keyValuePairs []KeyValue) (*assetsgenerated.GenericSecret, error) {
	return createCredential(ctx, name, description, func(meta assetsgenerated.AssetCreationRequest) (*assetsgenerated.GenericSecret, *http.Response, error) {
		spec := assetsgenerated.GenericSecretCreationSpec{}
		if len(keyValuePairs) > 0 {
			pairs := make([]assetsgenerated.KeyValuePair, len(keyValuePairs))
			for i, kv := range keyValuePairs {
				pairs[i] = assetsgenerated.KeyValuePair{}
				pairs[i].SetKey(kv.Key)
				pairs[i].SetValue(kv.Value)
			}
			spec.SetKeyValuePairs(pairs)
		}
		req := assetsgenerated.GenericSecretCreationRequest{Meta: meta, Spec: spec}
		return c.apiClient.CredentialsAPI.CreateGenericSecret(ctx).GenericSecretCreationRequest(req).Execute()
	})
}

func (c *Client) CreateDockerRegistry(ctx context.Context, name, description string, user, password, url string) (*assetsgenerated.DockerRegistry, error) {
	return createCredential(ctx, name, description, func(meta assetsgenerated.AssetCreationRequest) (*assetsgenerated.DockerRegistry, *http.Response, error) {
		spec := assetsgenerated.DockerRegistryCreationSpec{}
		spec.SetUser(user)
		spec.SetPassword(password)
		spec.SetUrl(url)
		req := assetsgenerated.DockerRegistryCreationRequest{Meta: meta, Spec: spec}
		return c.apiClient.CredentialsAPI.CreateDockerRegistry(ctx).DockerRegistryCreationRequest(req).Execute()
	})
}

func (c *Client) CreateNgcApiKey(ctx context.Context, name, description string, ngcApiKey string) (*assetsgenerated.NgcApiKey, error) {
	return createCredential(ctx, name, description, func(meta assetsgenerated.AssetCreationRequest) (*assetsgenerated.NgcApiKey, *http.Response, error) {
		spec := assetsgenerated.NgcApiKeyCreationSpec{}
		spec.SetNgcApiKey(ngcApiKey)
		req := assetsgenerated.NgcApiKeyCreationRequest{Meta: meta, Spec: spec}
		return c.apiClient.CredentialsAPI.CreateNgcApiKey(ctx).NgcApiKeyCreationRequest(req).Execute()
	})
}

func createCredential[T any](ctx context.Context, name, description string, doCreate func(assetsgenerated.AssetCreationRequest) (T, *http.Response, error)) (T, error) {
	meta := newPrivateMeta(name, description)
	resp, httpResp, err := doCreate(meta)
	return withCredentialResp(resp, httpResp, err)
}

func withCredentialResp[T any](resp T, httpResp *http.Response, err error) (T, error) {
	defer func() {
		if httpResp != nil && httpResp.Body != nil {
			httpResp.Body.Close()
		}
	}()
	if err != nil {
		var zero T
		return zero, clierrors.HandleResponseError(httpResp, err)
	}
	return resp, nil
}

func newPrivateMeta(name, description string) assetsgenerated.AssetCreationRequest {
	meta := assetsgenerated.AssetCreationRequest{}
	meta.Name = name
	meta.Scope = assetsgenerated.SCOPE_TENANT
	meta.SetPrivate(true)
	if description != "" {
		meta.SetDescription(description)
	}
	return meta
}
