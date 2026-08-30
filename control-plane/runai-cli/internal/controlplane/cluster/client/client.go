package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/api/config"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api/filter"

	clierrors "github.com/run-ai/runai/control-plane/runai-cli/internal/errors"

	openapi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
)

type BackendClusterClient struct {
	apiClient *openapi.APIClient
}

func NewBackendClusterClient(urlStr, token string) (*BackendClusterClient, error) {
	conf, err := config.CreateClusterConfig(urlStr, token)
	if err != nil {
		return nil, err
	}
	return &BackendClusterClient{apiClient: openapi.NewAPIClient(conf)}, nil
}

func (bpc *BackendClusterClient) GetClusterList(ctx context.Context) ([]openapi.ClustersMinimalData, error) {
	request := bpc.apiClient.ClustersAPI.GetClustersMinimalData(ctx)
	clusterListResponse, response, err := request.Execute()
	err = clierrors.HandleResponseError(response, err)
	if err != nil {
		return nil, err
	}

	if response != nil && response.Body != nil {
		defer response.Body.Close()
	}

	if !checkResponseStatusCodeOK(response) {
		return nil, fmt.Errorf("cluster list API call failed, status code error: %s", response.Status)
	}

	return clusterListResponse.Clusters, nil
}

func (bpc *BackendClusterClient) GetClusterByID(ctx context.Context, uuid string) (*openapi.ClustersMinimalData, error) {
	request := bpc.apiClient.ClustersAPI.GetClusterMinimalDataById(ctx, uuid)
	clustersMinimalData, response, err := request.Execute()

	err = clierrors.HandleResponseError(response, err)
	if err != nil {
		return nil, err
	}

	if response != nil && response.Body != nil {
		defer response.Body.Close()
	}

	if !checkResponseStatusCodeOK(response) {
		return nil, fmt.Errorf("cluster get by ID API call for ID: %s, failed, status code error: %s", uuid, response.Status)
	}

	return clustersMinimalData, nil
}

func (bpc *BackendClusterClient) GetClusterByName(ctx context.Context, name string) (*openapi.ClustersMinimalData, error) {
	filterBy := filter.NewBuilder().AppendName(name).StringList()
	request := bpc.apiClient.ClustersAPI.GetClustersMinimalData(ctx).
		FilterBy(filterBy)
	clustersMinimalData, response, err := request.Execute()

	err = clierrors.HandleResponseError(response, err)
	if err != nil {
		return nil, err
	}

	if response != nil && response.Body != nil {
		defer response.Body.Close()
	}

	if !checkResponseStatusCodeOK(response) {
		return nil, fmt.Errorf("clusters list API call failed, status code error: %s", response.Status)
	}

	if len(clustersMinimalData.Clusters) > 0 {
		return &clustersMinimalData.Clusters[0], nil
	}

	return nil, fmt.Errorf("failed to find cluster name: %s", name)
}

func checkResponseStatusCodeOK(response *http.Response) bool {
	if response == nil {
		return false
	}

	return response.StatusCode >= http.StatusOK && response.StatusCode < http.StatusIMUsed
}
