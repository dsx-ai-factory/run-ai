/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package apiclient

import (
	"context"
	"net/http"
	"sync"

	clientutils "github.com/run-ai/runai/runai-application-packages/client-utils"
	runaiclient "github.com/run-ai/runai/runai-installer/clients/pkg/clients"
	"github.com/run-ai/runai/runai-installer/installer/internal/logging"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// K8sClient interface for Kubernetes operations (allows mocking in tests)
type K8sClient interface {
	List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error
	Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error
	Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error
	Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error
	Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error
}

// Clients holds all API clients needed for installation
type Clients struct {
	ClusterServiceClient *ClusterServiceClient
	OrgUnitServiceClient *OrgUnitServiceClient
	AuthServiceClient    *AuthServiceClient
	runaiClient          *runaiclient.RunAIClient
}

var instance *Clients
var once sync.Once

func GetClientsInstanceWithCredentials(fqdn string, caCertPath string, creds runaiclient.Credentials) (*Clients, error) {
	logging.Infof("[Common] Checking if Run:AI CP is responsive by getting version from %s", fqdn)
	serverURL := clientutils.GetExternalUrl(fqdn)
	if err := runaiclient.CheckHealth(serverURL, caCertPath); err != nil {
		logging.Errorf("[Common] Failed to check health of %s: %v", fqdn, err)
		return nil, err
	}

	var initErr error
	once.Do(func() {
		instance, initErr = initClients(fqdn, caCertPath, creds)
	})

	if initErr != nil {
		once = sync.Once{}
		return nil, initErr
	}

	return instance, nil
}

func initClients(fqdn string, caCertPath string, creds runaiclient.Credentials) (*Clients, error) {
	url := clientutils.GetExternalUrl(fqdn)
	logging.Infof("[Common] Initializing clients for %s", url)

	if caCertPath != "" {
		logging.Infof("[Common] Using custom CA certificate from %s", caCertPath)
	}

	runaiClient, err := runaiclient.NewClientWithCA(fqdn, creds, caCertPath)
	if err != nil {
		logging.Errorf("[Common] Failed to create RunAI client: %v", err)
		return nil, err
	}

	// Create ClusterServiceClient using the shared client
	clusterServiceClient := NewClusterServiceClientFromRunAI(runaiClient)
	logging.Infof("[Common] Cluster Service Client created with token rotation for %s", url)

	// Create OrgUnitServiceClient using the shared client
	orgUnitServiceClient := NewOrgUnitServiceClientFromRunAI(runaiClient)
	logging.Infof("[Common] Orgunit Service Client created with token rotation for %s", url)

	// Create RolesHTTPClient for the AuthServiceClient
	rolesHTTPClient, err := NewRolesHTTPClient(fqdn, creds.Username, creds.Password, caCertPath)
	if err != nil {
		logging.Errorf("[Common] Failed to create RolesHTTPClient: %v", err)
		return nil, err
	}

	authServiceClient := NewAuthServiceClient(runaiClient, rolesHTTPClient)
	logging.Infof("[Common] Auth Service Client created with token rotation for %s", url)

	return &Clients{
		ClusterServiceClient: clusterServiceClient,
		OrgUnitServiceClient: orgUnitServiceClient,
		AuthServiceClient:    authServiceClient,
		runaiClient:          runaiClient,
	}, nil
}

// GetRunAIClient returns the underlying RunAI API client for direct API access
func (c *Clients) GetRunAIClient() *runaiclient.RunAIClient {
	return c.runaiClient
}

// GetHTTPClient returns the HTTP client with authentication for making direct HTTP requests
func (c *Clients) GetHTTPClient() *http.Client {
	if c.runaiClient != nil {
		return c.runaiClient.AuthAPIClient().GetConfig().HTTPClient
	}
	return nil
}

// ResetClients clears the singleton instance (useful for testing)
func ResetClients() {
	instance = nil
	once = sync.Once{}
}
