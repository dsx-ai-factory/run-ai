/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package authorization

import (
	"context"
	"fmt"
	"net/http"

	authClient "github.com/run-ai/runai/control-plane/authorization/pkg/client"
	authApi "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

type Client struct {
	apiClient authClient.IAuthorizationClient
}

func (cl Client) CheckPermittedScopes(ctx context.Context, resourceType authApi.ResourceType, action authApi.Action) (*authApi.PermittedScopes, *authApi.Error) {
	requestInfo := fmt.Sprintf("%s permitted scopes of %s", action, resourceType)

	resp, statusCode, err := cl.apiClient.GetPermittedScopes(ctx, resourceType, &action)
	if statusCode != http.StatusOK {
		if statusCode == http.StatusNotFound {
			return nil, authApi.NewError(http.StatusNotFound, fmt.Sprintf("%s does not exist", requestInfo))
		}
		return nil, authApi.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to receive %s: %v", requestInfo, err))
	}

	switch action {
	case authApi.ACTION_READ:
		return &resp.Read, nil
	case authApi.ACTION_UPDATE:
		return &resp.Update, nil
	case authApi.ACTION_CREATE:
		return &resp.Create, nil
	case authApi.ACTION_DELETE:
		return &resp.Delete, nil
	default:
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("unsupported rbac action %s", action)
		return nil, authApi.NewError(http.StatusInternalServerError, fmt.Sprintf("unsupported action %s", action))
	}
}
