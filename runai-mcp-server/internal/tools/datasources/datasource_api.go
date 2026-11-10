/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package datasources

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	assets "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/common"
)

type datasourceAdapter struct {
	validateCreate func(map[string]any) error
	validateUpdate func(map[string]any) error
	create         func(context.Context, *runaiclient.Client, map[string]any) (*http.Response, error)
	update         func(context.Context, *runaiclient.Client, string, map[string]any) (*http.Response, error)
	get            func(context.Context, *runaiclient.Client, string) (*http.Response, error)
	delete         func(context.Context, *runaiclient.Client, string) (*http.Response, error)
}

var datasourceAdapters = map[assets.AssetKind]datasourceAdapter{
	assets.ASSETKIND_NFS:           nfsDatasourceAdapter(),
	assets.ASSETKIND_PVC:           pvcDatasourceAdapter(),
	assets.ASSETKIND_S3:            s3DatasourceAdapter(),
	assets.ASSETKIND_GIT:           gitDatasourceAdapter(),
	assets.ASSETKIND_HOST_PATH:     hostPathDatasourceAdapter(),
	assets.ASSETKIND_CONFIG_MAP:    configMapDatasourceAdapter(),
	assets.ASSETKIND_SECRET_VOLUME: secretVolumeDatasourceAdapter(),
}

// The generated NFS create method incorrectly returns S3Asset; preserve its raw body.
func nfsDatasourceAdapter() datasourceAdapter {
	return datasourceAdapterFor(
		func(ctx context.Context, c *runaiclient.Client, r assets.NFSCreationRequest) (*http.Response, error) {
			return assetResponse(c.Assets.NFSAPI.CreateNfsAsset(ctx).NFSCreationRequest(r).Execute())
		},
		func(ctx context.Context, c *runaiclient.Client, id string, r assets.NFSUpdateRequest) (*http.Response, error) {
			return assetResponse(c.Assets.NFSAPI.UpdateNfsAssetById(ctx, id).NFSUpdateRequest(r).Execute())
		},
		func(ctx context.Context, c *runaiclient.Client, id string) (*http.Response, error) {
			return assetResponse(c.Assets.NFSAPI.GetNfsAssetById(ctx, id).Execute())
		},
		func(ctx context.Context, c *runaiclient.Client, id string) (*http.Response, error) {
			return assetResponse(c.Assets.NFSAPI.DeleteNfsAssetById(ctx, id).Execute())
		},
	)
}

func pvcDatasourceAdapter() datasourceAdapter {
	return datasourceAdapterFor(
		func(ctx context.Context, c *runaiclient.Client, r assets.PVCCreationRequest) (*http.Response, error) {
			return assetResponse(c.Assets.PVCAPI.CreatePvcAsset(ctx).PVCCreationRequest(r).Execute())
		},
		func(ctx context.Context, c *runaiclient.Client, id string, r assets.PVCUpdateRequest) (*http.Response, error) {
			return assetResponse(c.Assets.PVCAPI.UpdatePvcAssetById(ctx, id).PVCUpdateRequest(r).Execute())
		},
		func(ctx context.Context, c *runaiclient.Client, id string) (*http.Response, error) {
			return assetResponse(c.Assets.PVCAPI.GetPvcAssetById(ctx, id).Execute())
		},
		func(ctx context.Context, c *runaiclient.Client, id string) (*http.Response, error) {
			return assetResponse(c.Assets.PVCAPI.DeletePvcAssetById(ctx, id).Execute())
		},
	)
}

func s3DatasourceAdapter() datasourceAdapter {
	return datasourceAdapterFor(
		func(ctx context.Context, c *runaiclient.Client, r assets.S3CreationRequest) (*http.Response, error) {
			return assetResponse(c.Assets.S3API.CreateS3Asset(ctx).S3CreationRequest(r).Execute())
		},
		func(ctx context.Context, c *runaiclient.Client, id string, r assets.S3UpdateRequest) (*http.Response, error) {
			return assetResponse(c.Assets.S3API.UpdateS3AssetById(ctx, id).S3UpdateRequest(r).Execute())
		},
		func(ctx context.Context, c *runaiclient.Client, id string) (*http.Response, error) {
			return assetResponse(c.Assets.S3API.GetS3AssetById(ctx, id).Execute())
		},
		func(ctx context.Context, c *runaiclient.Client, id string) (*http.Response, error) {
			return assetResponse(c.Assets.S3API.DeleteS3AssetById(ctx, id).Execute())
		},
	)
}

func gitDatasourceAdapter() datasourceAdapter {
	return datasourceAdapterFor(
		func(ctx context.Context, c *runaiclient.Client, r assets.GitCreationRequest) (*http.Response, error) {
			return assetResponse(c.Assets.GitAPI.CreateGitAsset(ctx).GitCreationRequest(r).Execute())
		},
		func(ctx context.Context, c *runaiclient.Client, id string, r assets.GitUpdateRequest) (*http.Response, error) {
			return assetResponse(c.Assets.GitAPI.UpdateGitAssetById(ctx, id).GitUpdateRequest(r).Execute())
		},
		func(ctx context.Context, c *runaiclient.Client, id string) (*http.Response, error) {
			return assetResponse(c.Assets.GitAPI.GetGitAssetById(ctx, id).Execute())
		},
		func(ctx context.Context, c *runaiclient.Client, id string) (*http.Response, error) {
			return assetResponse(c.Assets.GitAPI.DeleteGitAssetById(ctx, id).Execute())
		},
	)
}

func hostPathDatasourceAdapter() datasourceAdapter {
	return datasourceAdapterFor(
		func(ctx context.Context, c *runaiclient.Client, r assets.HostPathCreationRequest) (*http.Response, error) {
			return assetResponse(c.Assets.HostPathAPI.CreateHostPath(ctx).HostPathCreationRequest(r).Execute())
		},
		func(ctx context.Context, c *runaiclient.Client, id string, r assets.HostPathUpdateRequest) (*http.Response, error) {
			return assetResponse(c.Assets.HostPathAPI.UpdateHostPathById(ctx, id).HostPathUpdateRequest(r).Execute())
		},
		func(ctx context.Context, c *runaiclient.Client, id string) (*http.Response, error) {
			return assetResponse(c.Assets.HostPathAPI.GetHostPathById(ctx, id).Execute())
		},
		func(ctx context.Context, c *runaiclient.Client, id string) (*http.Response, error) {
			return assetResponse(c.Assets.HostPathAPI.DeleteHostPathById(ctx, id).Execute())
		},
	)
}

func configMapDatasourceAdapter() datasourceAdapter {
	return datasourceAdapterFor(
		func(ctx context.Context, c *runaiclient.Client, r assets.ConfigMapCreationRequest) (*http.Response, error) {
			return assetResponse(c.Assets.ConfigMapAPI.CreateConfigMapAsset(ctx).ConfigMapCreationRequest(r).Execute())
		},
		func(ctx context.Context, c *runaiclient.Client, id string, r assets.ConfigMapUpdateRequest) (*http.Response, error) {
			return assetResponse(c.Assets.ConfigMapAPI.UpdateConfigMapAssetById(ctx, id).ConfigMapUpdateRequest(r).Execute())
		},
		func(ctx context.Context, c *runaiclient.Client, id string) (*http.Response, error) {
			return assetResponse(c.Assets.ConfigMapAPI.GetConfigMapAssetById(ctx, id).Execute())
		},
		func(ctx context.Context, c *runaiclient.Client, id string) (*http.Response, error) {
			return assetResponse(c.Assets.ConfigMapAPI.DeleteConfigMapAssetById(ctx, id).Execute())
		},
	)
}

func secretVolumeDatasourceAdapter() datasourceAdapter {
	return datasourceAdapterFor(
		func(ctx context.Context, c *runaiclient.Client, r assets.SecretAssetCreationRequest) (*http.Response, error) {
			return assetResponse(c.Assets.SecretAPI.CreateSecretAsset(ctx).SecretAssetCreationRequest(r).Execute())
		},
		func(ctx context.Context, c *runaiclient.Client, id string, r assets.SecretAssetUpdateRequest) (*http.Response, error) {
			return assetResponse(c.Assets.SecretAPI.UpdateSecretAssetById(ctx, id).SecretAssetUpdateRequest(r).Execute())
		},
		func(ctx context.Context, c *runaiclient.Client, id string) (*http.Response, error) {
			return assetResponse(c.Assets.SecretAPI.GetSecretAssetById(ctx, id).Execute())
		},
		func(ctx context.Context, c *runaiclient.Client, id string) (*http.Response, error) {
			return assetResponse(c.Assets.SecretAPI.DeleteSecretAssetById(ctx, id).Execute())
		},
	)
}

func datasourceAdapterFor[C, U any](
	create func(context.Context, *runaiclient.Client, C) (*http.Response, error),
	update func(context.Context, *runaiclient.Client, string, U) (*http.Response, error),
	get, deleteCall func(context.Context, *runaiclient.Client, string) (*http.Response, error),
) datasourceAdapter {
	return datasourceAdapter{
		validateCreate: func(raw map[string]any) error { var r C; return common.DecodeTypedRequest(raw, &r) },
		validateUpdate: func(raw map[string]any) error { var r U; return common.DecodeTypedRequest(raw, &r) },
		create: func(ctx context.Context, c *runaiclient.Client, raw map[string]any) (*http.Response, error) {
			var r C
			if err := common.DecodeTypedRequest(raw, &r); err != nil {
				return nil, err
			}
			return create(ctx, c, r)
		},
		update: func(ctx context.Context, c *runaiclient.Client, id string, raw map[string]any) (*http.Response, error) {
			var r U
			if err := common.DecodeTypedRequest(raw, &r); err != nil {
				return nil, err
			}
			return update(ctx, c, id, r)
		},
		get: get, delete: deleteCall,
	}
}

func assetResponse[T any](_ T, response *http.Response, err error) (*http.Response, error) {
	// The generated client formats Title/Detail, but the assets API uses Message.
	var apiErr *assets.GenericOpenAPIError
	if errors.As(err, &apiErr) {
		if model, ok := apiErr.Model().(assets.Error); ok && model.Message != "" {
			err = fmt.Errorf("%s: %w", model.Message, err)
		}
	}
	return response, err
}
