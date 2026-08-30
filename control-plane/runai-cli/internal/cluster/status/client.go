package status

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api/config"

	instancecfg "github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"go.uber.org/zap"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/cluster/apis"
	clusterurl "github.com/run-ai/runai/control-plane/runai-cli/internal/cluster/url"
	clierrors "github.com/run-ai/runai/control-plane/runai-cli/internal/errors"
	openapi "github.com/run-ai/runai/runai-clients/cluster-api-client/generated"
)

const logMsg = "GetClusterStatus"

type ClusterApiClient struct {
	apiClient *openapi.APIClient
}

func NewClusterApiClient(urlStr, token string) (*ClusterApiClient, error) {
	conf, err := config.CreateClusterApiConfig(urlStr, token)
	if err != nil {
		return nil, err
	}
	return &ClusterApiClient{apiClient: openapi.NewAPIClient(conf)}, nil
}

func (c *ClusterApiClient) GetWorkloadStatus(ctx context.Context, workload *apis.Workload) (*openapi.WorkloadStatus, error) {
	workloadType := workload.Type.ClusterApiType()
	workloadFramework := workload.Framework.ClusterApiFramework()
	// External routing: NimService uses the external/external path natively, and any
	// workload resolved via workloads-manager reaches us with UndefinedFramework so it
	// hits the same path (see internal/controlplane/workload/resolve and the parallel
	// override in internal/cluster/url/url.go).
	if workload.Framework == api.NimService || workload.Framework == api.UndefinedFramework {
		workloadType = clusterurl.External
		workloadFramework = clusterurl.External
	}
	request := c.apiClient.ClusterAPIAPI.GetWorkloadStatus(ctx, workload.Project, workloadType, workload.Name, workloadFramework)
	status, response, err := request.Execute()
	if err != nil {
		return nil, clierrors.HandleResponseError(response, err)
	}

	if response != nil && response.Body != nil {
		defer response.Body.Close()
	}

	return status, nil
}

func (c *ClusterApiClient) GetClusterStatus(ctx context.Context) ([]byte, error) {
	logger := instancecfg.GetInstance().GetLogger()
	request := c.apiClient.StatusAPI.Status(ctx)
	response, err := request.Execute()
	if err != nil {
		return nil, clierrors.HandleResponseError(response, err)
	}

	if response == nil {
		return nil, fmt.Errorf("response is nil")
	}

	if response.Body == nil {
		return nil, fmt.Errorf("response body is nil")
	}

	logger.Debug(logMsg, zap.Int("code", response.StatusCode),
		zap.String("url", response.Request.URL.String()))

	defer func() {
		if errClose := response.Body.Close(); errClose != nil {
			logger.Error(logMsg,
				zap.Error(errClose))
		}
	}()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cluster response status is not 200 OK, status: %d", response.StatusCode)
	}

	// TODO once we resolve OpenAPI generation and move to common - add Metadata as part of the response.
	payload, err := io.ReadAll(response.Body)
	if err != nil {
		logger.Debug("failed to get cluster metadata", zap.Error(err))
	}

	return payload, nil
}
