/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package clients

import (
	"context"

	clusterclient "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client"
	idmclient "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client"
)

const defaultControlPlaneClientID = "runai-internal.installer"

type ControlPlaneDetails struct {
	URL            string
	ClientId       string
	ClientSecret   string
	BootstrapToken string
}

func NewAuthenticatedClusterClient(cpDetails *ControlPlaneDetails) (clusterclient.IClusterExternalClient, error) {
	token, err := getCPAccessToken(cpDetails)
	if err != nil {
		return nil, err
	}
	return clusterclient.NewClusterExternalClient(cpDetails.URL, token), nil
}

func getCPAccessToken(cpDetails *ControlPlaneDetails) (string, error) {
	if cpDetails.BootstrapToken != "" {
		return cpDetails.BootstrapToken, nil
	}
	idmExternalClient := idmclient.NewIdmExternalClient(cpDetails.URL)
	tokenResponse, err := idmExternalClient.ClientCredentialsLogin(context.Background(), cpDetails.ClientId, cpDetails.ClientSecret)
	if err != nil {
		return "", err
	}
	return *tokenResponse.AccessToken, nil
}
