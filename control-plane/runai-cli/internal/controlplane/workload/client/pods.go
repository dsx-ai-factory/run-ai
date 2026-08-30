package client

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/api/config"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api/httputils"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/options"
	clierrors "github.com/run-ai/runai/control-plane/runai-cli/internal/errors"
	"k8s.io/utils/ptr"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"

	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

const (
	NoLimit                   = -1
	defaultPageSizePods int32 = 100
)

type BackendWorkloadPodsClient struct {
	apiClient *openapi.APIClient
}

func NewBackendWorkloadPodsClient(urlStr, token string) (*BackendWorkloadPodsClient, error) {
	conf, err := config.CreateWorkloadServiceConfig(urlStr, token)
	if err != nil {
		return nil, err
	}
	return &BackendWorkloadPodsClient{apiClient: openapi.NewAPIClient(conf)}, nil
}

func (bsc *BackendWorkloadPodsClient) List(ctx context.Context, workloadID string,
	options options.Paginationer) (*openapi.ListPods200Response, error) {
	request := bsc.apiClient.PodsAPI.GetWorkloadPods(ctx, workloadID)
	if options != nil {
		request = request.Limit(options.GetLimit())
		request = request.Offset(options.GetOffset())
	}

	workloadList, response, err := request.Execute()
	if err != nil {
		return nil, clierrors.HandleResponseError(response, err)
	}

	if response == nil {
		return nil, fmt.Errorf("response is nil")
	}

	defer func() {
		if response.Body != nil {
			_ = response.Body.Close()
		}
	}()

	if !httputils.CheckResponseStatusCodeOK(response) {
		return nil, fmt.Errorf(consts.ErrorWorkloadPodsListAPI, response.Status)
	}

	return workloadList, nil
}

func (bsc *BackendWorkloadPodsClient) ListAll(ctx context.Context, workloadID string) ([]openapi.Pod, error) {
	var allPods []openapi.Pod
	var offset int32
	pageSize := defaultPageSizePods
	for {
		pagination := options.Pagination{
			Limit:  ptr.To(pageSize),
			Offset: offset,
		}

		pods, err := bsc.List(ctx, workloadID, &pagination)
		if err != nil {
			return nil, err
		}
		allPods = append(allPods, pods.Pods...)
		if !pods.HasNext() {
			break
		}
		offset = pods.GetNext()
	}
	return allPods, nil
}

func (bsc *BackendWorkloadPodsClient) ListByLimit(ctx context.Context, workloadID string, limit int32) ([]openapi.Pod, error) {
	if limit <= NoLimit {
		return bsc.ListAll(ctx, workloadID)
	}
	var allPods []openapi.Pod
	var offset int32
	pageSize := defaultPageSizePods
	if limit < defaultPageSizePods {
		pageSize = limit
	}

	for {
		pagination := options.Pagination{

			Limit:  ptr.To(pageSize),
			Offset: offset,
		}

		pods, err := bsc.List(ctx, workloadID, &pagination)
		if err != nil {
			return nil, err
		}
		allPods = append(allPods, pods.Pods...)
		if !pods.HasNext() {
			break
		}
		offset = pods.GetNext()
		lenAllPods := int32(len(allPods)) //nolint:gosec // limit is set max to 100
		if limit-lenAllPods <= defaultPageSizePods {
			pageSize = limit - lenAllPods
		}

		if pageSize <= 0 {
			break
		}
	}

	return allPods, nil
}
