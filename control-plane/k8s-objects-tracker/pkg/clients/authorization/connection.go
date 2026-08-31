/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package authorization

import (
	"fmt"
	"net/http"
	"net/url"

	authorizationClient "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/pkg/config"
	runai_config "github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/logging"
)

type Client struct{}

var (
	clientConn *authorizationClient.APIClient
	clientImpl ClientInterface
)

func InitClient() *authorizationClient.APIClient {
	var err *api.Error

	if clientConn == nil {
		clientConn, err = newAuthorizationConnection()
		if err != nil {
			logging.Fatalf("failed to connect to authorization service: %v", err)
		}

		clientImpl = Client{}
	}

	return clientConn
}

func SetMock(mockClient ClientInterface) {
	clientImpl = mockClient
	clientConn = &authorizationClient.APIClient{}
}

func Get() ClientInterface {
	InitClient()
	return clientImpl
}

func newAuthorizationConnection() (*authorizationClient.APIClient, *api.Error) {
	servicesConfig := runai_config.Get[*config.ServicesConfig]()
	authorizationUrlStr := servicesConfig.AuthorizationUrl

	authorizationUrl, err := url.Parse(authorizationUrlStr)
	if err != nil {
		return nil, api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to connect to authorization service on '%s'", authorizationUrlStr), err)
	}

	authorizationConfiguration := authorizationClient.NewConfiguration()
	authorizationConfiguration.Scheme = authorizationUrl.Scheme
	authorizationConfiguration.Host = authorizationUrl.Host
	conf := runai_config.Get[*config.K8sObjectsTrackerConfig]()
	if conf.LogLevel == "debug" {
		authorizationConfiguration.Debug = true
	}

	apiClient := authorizationClient.NewAPIClient(authorizationConfiguration)

	return apiClient, nil
}
