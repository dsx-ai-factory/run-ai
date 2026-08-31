/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pvc

import (
	"context"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/api/config"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api/httputils"

	openapi "github.com/run-ai/runai/runai-clients/k8s-objects-tracker-client/generated"
)

const (
	OriginAllStr    = "all"
	ScopeProjectStr = "project"
)

type ObjTrackerApiClient struct {
	apiClient *openapi.APIClient
}

func NewObjTrackerApiClient(urlStr, token string) (*ObjTrackerApiClient, error) {
	conf, err := config.CreateObjTrackerApiConfig(urlStr, token)
	if err != nil {
		return nil, err
	}
	return &ObjTrackerApiClient{apiClient: openapi.NewAPIClient(conf)}, nil
}

func (kot *ObjTrackerApiClient) List(ctx context.Context, projectId string) (*openapi.PVCListResponse, error) {
	request := kot.apiClient.PVCsAPI.ListPvcs(ctx)
	request = request.Origin(OriginAllStr)
	request = request.IncludeAncestors(true)
	request = request.ScopeType(ScopeProjectStr)
	request = request.ScopeId(projectId)

	pvcList, response, err := request.Execute() //nolint:bodyclose // handle in handler
	return httputils.HandleResponse(pvcList, response, err)
}
