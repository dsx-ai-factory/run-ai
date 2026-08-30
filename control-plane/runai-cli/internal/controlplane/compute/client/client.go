package client

import (
	"context"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/api/config"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api/httputils"
	openapi "github.com/run-ai/runai/runai-clients/assets-service-client/generated"
)

type Client struct {
	apiClient *openapi.APIClient
}

func NewClient(urlStr, token string) (*Client, error) {
	conf, err := config.CreateAssetsServiceConfig(urlStr, token)
	if err != nil {
		return nil, err
	}
	return &Client{apiClient: openapi.NewAPIClient(conf)}, nil
}

func (c *Client) List(ctx context.Context, projectID *int32, clusterID *string) (*openapi.ComputeListResponse, error) {
	request := c.apiClient.ComputeAPI.ListComputeAssets(ctx)

	if projectID != nil {
		request = request.ProjectId(*projectID)
	}
	if clusterID != nil {
		request = request.ClusterId(*clusterID)
	}

	// Always include descendants
	request = request.IncludeDescendants(true)

	computes, response, err := request.Execute() //nolint:bodyclose // handle in handler
	return httputils.HandleResponse(computes, response, err)
}
