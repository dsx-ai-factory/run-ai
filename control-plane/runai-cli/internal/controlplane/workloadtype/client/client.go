package client

import (
	"context"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/api/config"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api/httputils"

	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

type Client struct {
	apiClient *openapi.APIClient
}

func NewClient(urlStr, token string) (*Client, error) {
	conf, err := config.CreateWorkloadServiceConfig(urlStr, token)
	if err != nil {
		return nil, err
	}
	return &Client{apiClient: openapi.NewAPIClient(conf)}, nil
}

func (c *Client) Get(ctx context.Context, workloadTypeId string) (*openapi.WorkloadTypeConfig, error) {
	result, response, err := c.apiClient.WorkloadPropertiesAPI.GetWorkloadType(ctx, workloadTypeId).Execute()
	return httputils.HandleResponse(result, response, err)
}

func (c *Client) List(ctx context.Context, filterBy []string, externalTypesOnly *bool) (*openapi.ListWorkloadTypes200Response, error) {
	request := c.apiClient.WorkloadPropertiesAPI.ListWorkloadTypes(ctx)

	if len(filterBy) > 0 {
		request = request.FilterBy(filterBy)
	}
	if externalTypesOnly != nil {
		request = request.ExternalTypesOnly(*externalTypesOnly)
	}

	workloadTypes, response, err := request.Execute()
	return httputils.HandleResponse(workloadTypes, response, err)
}
