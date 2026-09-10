/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package clients

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"

	clusterclient "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client"
	idmclient "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client"
	openapi "github.com/dsx-ai-factory/run-ai/runai-cluster/common/clients/control-plane/generated"
)

const defaultControlPlaneClientID = "runai-internal.installer"

func ConfigureDefaultHTTPTransport(ca []byte) error {
	roots, err := x509.SystemCertPool()
	if err != nil {
		return err
	}
	if !roots.AppendCertsFromPEM(ca) {
		return fmt.Errorf("failed to parse custom CA certificate")
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12}
	if transport.TLSClientConfig != nil {
		tlsConfig = transport.TLSClientConfig.Clone()
	}
	tlsConfig.RootCAs = roots
	transport.TLSClientConfig = tlsConfig
	http.DefaultTransport = transport
	return nil
}

type ControlPlaneDetails struct {
	URL            string
	ClientId       string
	ClientSecret   string
	BootstrapToken string
}

type ClusterServiceClient struct {
	apiClient *openapi.APIClient
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
