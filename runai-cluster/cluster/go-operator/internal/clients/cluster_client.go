/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package clients

import (
	"context"
	"errors"

	clusterClient "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	idm_client_api "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client/generated"
)

const defaultControlPlaneClientID = "runai-internal.installer"

type ControlPlaneDetails struct {
	URL          string
	ClientId     string
	ClientSecret string
	ClusterUUID  string
}

type ClusterServiceClient interface {
	GetClusterInternalConfig(ctx context.Context) (*openapi.GetInternalClusterConfigResponse, error)
	PutRemoteConfigStatus(ctx context.Context, putRequest *openapi.PutClusterConfigStatusRequest) error
	SetCpDetails(cpDetails *ControlPlaneDetails) (ClusterServiceClient, error)
}

type ClusterServiceClientImpl struct {
	token     *idm_client_api.TokenResponse
	cpDetails *ControlPlaneDetails
}

func NewClusterServiceClient() ClusterServiceClient {
	return &ClusterServiceClientImpl{}
}

func (c *ClusterServiceClientImpl) SetCpDetails(cpDetails *ControlPlaneDetails) (ClusterServiceClient, error) {
	if cpDetails == nil || cpDetails.ClientId == "" || cpDetails.ClientSecret == "" || cpDetails.URL == "" {
		return nil, errors.New("missing control plane details needed to authenticate")
	}
	c.cpDetails = cpDetails
	return c, nil
}

func (c *ClusterServiceClientImpl) authenticatedClusterClient() (clusterClient.IClusterExternalClient, error) {
	if c.cpDetails == nil || c.cpDetails.ClientId == "" || c.cpDetails.ClientSecret == "" || c.cpDetails.URL == "" {
		return nil, errors.New("missing control plane details needed to authenticate")
	}
	token, err := getOrRefreshToken(c.cpDetails, c.token)
	if err != nil {
		return nil, err
	}
	c.token = token
	return clusterClient.NewClusterExternalClient(c.cpDetails.URL, *c.token.AccessToken), nil

}

func (c *ClusterServiceClientImpl) GetClusterInternalConfig(ctx context.Context) (*openapi.GetInternalClusterConfigResponse, error) {
	client, err := c.authenticatedClusterClient()
	if err != nil {
		return nil, err
	}
	return client.GetClusterInternalConfig(ctx, c.cpDetails.ClusterUUID)
}

func (c *ClusterServiceClientImpl) PutRemoteConfigStatus(ctx context.Context, putRequest *openapi.PutClusterConfigStatusRequest) error {
	client, err := c.authenticatedClusterClient()
	if err != nil {
		return err
	}
	return client.PutClusterConfigStatus(ctx, c.cpDetails.ClusterUUID, putRequest)
}
