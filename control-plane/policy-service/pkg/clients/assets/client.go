/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package assets

import (
	"context"
	"io"
	"net/http"

	assetsService "github.com/run-ai/runai/runai-clients/assets-service-client"
	assetsServiceApi "github.com/run-ai/runai/runai-clients/assets-service-client/generated"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

type Client struct {
	apiClient assetsService.IAssetsClient
}

func (c Client) ListDatasourcesByIds(ctx context.Context, ids []string) (*assetsServiceApi.DatasourceListResponse, *assetsServiceApi.Error) {
	resp, r, err := c.apiClient.ListDatasourceAssets(ctx, ids)
	if r != nil {
		if r.StatusCode == http.StatusOK {
			return resp, nil
		}
		data, _ := io.ReadAll(r.Body)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to list assets: %s", data)
		return nil, assetsServiceApi.NewError(http.StatusServiceUnavailable, "failed to list imposedAssets")
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to request list of assets: %v", err)
	return nil, assetsServiceApi.NewError(http.StatusServiceUnavailable, "failed to process list of imposedAssets")
}
