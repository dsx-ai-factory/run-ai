package client

import (
	"context"
	"fmt"

	"k8s.io/utils/ptr"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/api/config"
	clierrors "github.com/run-ai/runai/control-plane/runai-cli/internal/errors"
	openapi "github.com/run-ai/runai/runai-clients/datavolumes-service-client/generated"
)

type Client struct {
	apiClient *openapi.APIClient
}

func NewClient(domain, token string) (*Client, error) {
	dvConfig, err := config.CreateDatavolumesServiceConfig(domain, token)
	if err != nil {
		return nil, fmt.Errorf("failed to create datavolumes service config: %w", err)
	}

	return &Client{
		apiClient: openapi.NewAPIClient(dvConfig),
	}, nil
}

func (c *Client) List(ctx context.Context, projectId *string) (*openapi.GetDatavolumes200Response, error) {
	request := c.apiClient.DatavolumesAPI.GetDatavolumes(ctx)

	if ptr.Deref(projectId, "") != "" {
		request = request.RequestType(openapi.DATAVOLUMEREQUESTTYPE_USABLE_IN_PROJECT)
		request = request.UsableInProjectId(*projectId)
	} else {
		request = request.RequestType(openapi.DATAVOLUMEREQUESTTYPE_ORIGINATED)
	}

	response, httpResp, err := request.Execute()

	defer func() {
		if httpResp != nil && httpResp.Body != nil {
			httpResp.Body.Close()
		}
	}()

	if err != nil {
		return nil, clierrors.HandleResponseError(httpResp, err)
	}

	return response, nil
}
