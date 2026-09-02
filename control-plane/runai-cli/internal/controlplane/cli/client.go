/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cli

import (
	"context"
	"errors"
	"os"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api/config"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	clierrors "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/errors"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/cli-exposer-client/generated"
)

type Client struct {
	apiClient *openapi.APIClient
}

func NewClient(urlStr, token string) (*Client, error) {
	conf, err := config.CreateExposerConfig(urlStr, token)
	if err != nil {
		return nil, err
	}
	return &Client{apiClient: openapi.NewAPIClient(conf)}, nil
}

func (c *Client) GetInstallerUnix(ctx context.Context) (string, error) {
	request := c.apiClient.ResearcherCommandLineInterfaceAPI.GetInstallerUnix(ctx)
	scriptData, response, err := request.Execute()
	if err != nil {
		err = clierrors.HandleResponseError(response, err)
		return "", err
	}

	if response != nil && response.Body != nil {
		defer response.Body.Close()
	}

	return scriptData, nil
}

func (c *Client) GetBinary(ctx context.Context, operationSystem, arc string) (*os.File, error) {
	request := c.apiClient.ResearcherCommandLineInterfaceAPI.GetBinary(ctx, operationSystem, arc)
	binaryFile, response, err := request.Execute()
	if err != nil {
		err = clierrors.HandleResponseError(response, err)
		return nil, err
	}

	if response != nil && response.Body != nil {
		defer response.Body.Close()
	}
	if binaryFile == nil {
		return nil, errors.New(consts.FailedDownloadBinary)
	}
	return *binaryFile, nil
}

func (c *Client) GetVersion(ctx context.Context) (*openapi.Version, error) {
	request := c.apiClient.ResearcherCommandLineInterfaceAPI.GetVersion(ctx)
	version, response, err := request.Execute()

	if response != nil && response.Body != nil {
		defer response.Body.Close()
	}

	if err != nil {
		err = clierrors.HandleResponseError(response, err)
		return nil, err
	}

	return version, nil
}
