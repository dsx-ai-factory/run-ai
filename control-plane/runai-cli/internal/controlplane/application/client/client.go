/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api/httputils"
	clierrors "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/errors"
	instancecfg "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"go.uber.org/zap"
)

const (
	logMsg       = "ListAIApplications"
	deleteLogMsg = "DeleteAIApplication"
	createLogMsg = "CreateAIApplication"
)

type AIApplicationClient struct {
	apiClient *openapi.APIClient
}

func NewAIApplicationClient(urlStr, token string) (*AIApplicationClient, error) {
	conf, err := config.CreateWorkloadServiceConfig(urlStr, token)
	if err != nil {
		return nil, err
	}
	return &AIApplicationClient{apiClient: openapi.NewAPIClient(conf)}, nil
}

func (c *AIApplicationClient) List(ctx context.Context, limit, offset int32, filters []string) (*openapi.ListApplications200Response, error) {
	request := c.apiClient.AIApplicationsAPI.ListApplications(ctx)
	request = request.Verbosity(openapi.AIAPPLICATIONVERBOSITY_FULL)
	request = request.Limit(limit)
	request = request.Offset(offset)

	if len(filters) != 0 {
		request = request.FilterBy(filters)
	}

	httputils.SetOrderFunc(&request, request.SortOrder)
	err := httputils.SetSortByFunc(&request, request.SortBy, openapi.NewAIApplicationsSortFilterFieldsFromValue)
	if err != nil {
		return nil, err
	}

	result, response, err := request.Execute()
	logger := instancecfg.GetInstance().GetLogger()

	if err != nil {
		logger.Error(logMsg, zap.Error(err))
		return nil, clierrors.HandleResponseError(response, err)
	}

	if response != nil {
		logger.Debug(logMsg,
			zap.Int("code", response.StatusCode),
			zap.String("url", response.Request.URL.String()))

		if response.Body != nil {
			defer func() {
				if errClose := response.Body.Close(); errClose != nil {
					logger.Error(logMsg, zap.Error(errClose))
				}
			}()
		}
	}

	return result, nil
}

func (c *AIApplicationClient) Delete(ctx context.Context, aiApplicationId string) error {
	request := c.apiClient.AIApplicationsAPI.DeleteApplication(ctx, aiApplicationId)

	_, response, err := request.Execute()
	logger := instancecfg.GetInstance().GetLogger()

	if err != nil {
		logger.Error(deleteLogMsg, zap.Error(err))
		return clierrors.HandleResponseError(response, err)
	}

	if response != nil {
		logger.Debug(deleteLogMsg,
			zap.Int("code", response.StatusCode),
			zap.String("url", response.Request.URL.String()))

		if response.Body != nil {
			defer func() {
				if errClose := response.Body.Close(); errClose != nil {
					logger.Error(deleteLogMsg, zap.Error(errClose))
				}
			}()
		}
	}

	return nil
}

func (c *AIApplicationClient) Create(ctx context.Context, request openapi.AIApplicationCreationRequest) (*openapi.AIApplication, error) {
	apiRequest := c.apiClient.AIApplicationsAPI.CreateApplication(ctx).AIApplicationCreationRequest(request)

	result, response, err := apiRequest.Execute()
	logger := instancecfg.GetInstance().GetLogger()

	if err != nil {
		logger.Error(createLogMsg, zap.Error(err))
		return nil, clierrors.HandleResponseError(response, err)
	}

	if response != nil {
		logger.Debug(createLogMsg,
			zap.Int("code", response.StatusCode),
			zap.String("url", response.Request.URL.String()))

		if response.Body != nil {
			defer func() {
				if errClose := response.Body.Close(); errClose != nil {
					logger.Error(createLogMsg, zap.Error(errClose))
				}
			}()
		}
	}

	return result, nil
}
