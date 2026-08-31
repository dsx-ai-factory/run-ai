/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package clients

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	idmclient "github.com/run-ai/runai/control-plane/identity-manager/pkg/client"
	openapi "github.com/run-ai/runai/control-plane/identity-manager/pkg/client/generated"
	"k8s.io/utils/ptr"
)

func getAuthenticationToken(cpDetails *ControlPlaneDetails) (*openapi.TokenResponse, error) {
	clientId := getClientIDOrDefault(cpDetails.ClientId)
	idmExternalClient := idmclient.NewIdmExternalClient(cpDetails.URL)
	tokenResponse, err := idmExternalClient.ClientCredentialsLogin(context.Background(), clientId, cpDetails.ClientSecret)
	if err != nil {
		return nil, err
	}
	return tokenResponse, nil
}

func getOrRefreshToken(cpDetails *ControlPlaneDetails, token *openapi.TokenResponse) (*openapi.TokenResponse, error) {
	if token == nil {
		return getAuthenticationToken(cpDetails)
	}
	isExpired, err := IsTokenExpired(token)
	if err != nil {
		return getAuthenticationToken(cpDetails)
	}

	if *isExpired {
		return getAuthenticationToken(cpDetails)
	}

	return token, nil
}

func IsTokenExpired(token *openapi.TokenResponse) (*bool, error) {
	claims := jwt.MapClaims{}
	_, _, err := jwt.NewParser().ParseUnverified(*token.AccessToken, claims)
	if err != nil {
		return nil, err
	}
	epoch := claims["exp"].(float64)
	timeObj := time.Unix(int64(epoch), 0)
	return ptr.To(timeObj.Before(time.Now())), nil
}

func getClientIDOrDefault(clientID string) string {
	if clientID == "" {
		return defaultControlPlaneClientID
	}

	return clientID
}
