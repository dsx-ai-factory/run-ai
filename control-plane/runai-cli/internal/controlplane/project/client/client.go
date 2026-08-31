/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"context"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/api/config"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api/httputils"

	instancecfg "github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"go.uber.org/zap"

	clierrors "github.com/run-ai/runai/control-plane/runai-cli/internal/errors"
	openapi "github.com/run-ai/runai/runai-clients/org-unit-service-client/generated"
)

const logMsg = "GetProjectList"

type ProjectClient struct {
	apiClient *openapi.APIClient
}

func NewProjectClient(urlStr, token string) (*ProjectClient, error) {
	conf, err := config.CreateProjectConfig(urlStr, token)
	if err != nil {
		return nil, err
	}
	return &ProjectClient{apiClient: openapi.NewAPIClient(conf)}, nil
}

func (opc *ProjectClient) CreateProject(ctx context.Context, creationRequest *openapi.ProjectCreationRequest) (*openapi.Project, error) {
	request := opc.apiClient.ProjectsAPI.CreateProject(ctx)
	request = request.ProjectCreationRequest(*creationRequest)
	createdProject, response, err := request.Execute()
	logger := instancecfg.GetInstance().GetLogger()

	if err != nil {
		logger.Error(logMsg,
			zap.Error(err))
		return nil, clierrors.HandleResponseError(response, err)
	}

	if response != nil && response.Body != nil {
		logger.Debug(logMsg, zap.Int("code", response.StatusCode),
			zap.String("url", response.Request.URL.String()))
		defer func() {
			if errClose := response.Body.Close(); errClose != nil {
				logger.Error(logMsg,
					zap.Error(errClose))
			}
		}()
	}

	return createdProject, nil
}

func (opc *ProjectClient) List(ctx context.Context, limit, offset int32, filters []string) (*openapi.GetProjects200Response, error) {
	request := opc.apiClient.ProjectsAPI.GetProjects(ctx)
	request = request.FilterBy(filters)
	request = request.Limit(limit)
	request = request.Offset(offset)

	httputils.SetOrderFunc(&request, request.SortOrder)
	err := httputils.SetSortByFunc(&request, request.SortBy, openapi.NewProjectFilterSortFieldsFromValue)
	if err != nil {
		return nil, err
	}

	projectList, response, err := request.Execute()
	logger := instancecfg.GetInstance().GetLogger()

	if err != nil {
		logger.Error(logMsg,
			zap.Error(err))
		return nil, clierrors.HandleResponseError(response, err)
	}

	if response != nil && response.Body != nil {
		logger.Debug(logMsg, zap.Int("code", response.StatusCode),
			zap.String("url", response.Request.URL.String()))
		defer func() {
			if errClose := response.Body.Close(); errClose != nil {
				logger.Error(logMsg,
					zap.Error(errClose))
			}
		}()
	}

	return projectList, nil
}
