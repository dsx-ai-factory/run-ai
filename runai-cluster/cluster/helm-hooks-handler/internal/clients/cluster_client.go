package clients

import (
	"context"

	clusterclient "github.com/run-ai/runai/control-plane/cluster-service/pkg/client"
	idmclient "github.com/run-ai/runai/control-plane/identity-manager/pkg/client"
	openapi "github.com/run-ai/runai/runai-cluster/common/clients/control-plane/generated"
)

const defaultControlPlaneClientID = "runai-internal.installer"

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
