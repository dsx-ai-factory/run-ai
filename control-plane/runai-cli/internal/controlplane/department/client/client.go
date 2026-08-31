/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"context"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/api/config"

	instancecfg "github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"go.uber.org/zap"

	clierrors "github.com/run-ai/runai/control-plane/runai-cli/internal/errors"
	openapi "github.com/run-ai/runai/runai-clients/org-unit-service-client/generated"
)

const logMsg = "GetDepartmentList"

type DepartmentClient struct {
	apiClient *openapi.APIClient
}

func NewDepartmentClient(urlStr, token string) (*DepartmentClient, error) {
	conf, err := config.CreateDepartmentConfig(urlStr, token)
	if err != nil {
		return nil, err
	}
	return &DepartmentClient{apiClient: openapi.NewAPIClient(conf)}, nil
}

func (opc *DepartmentClient) GetDepartmentList(ctx context.Context, filters []string, limit, offset int32) (*openapi.GetDepartments200Response, error) {
	request := opc.apiClient.DepartmentsAPI.GetDepartments(ctx)
	request = request.Verbosity(openapi.VERBOSITY_VERBOSE)
	request = request.FilterBy(filters)
	request = request.Limit(limit)
	request = request.Offset(offset)
	departmentList, response, err := request.Execute()
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

	return departmentList, nil
}
