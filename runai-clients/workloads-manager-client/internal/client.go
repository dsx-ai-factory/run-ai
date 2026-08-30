package internal

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"

	openapi "github.com/run-ai/runai/runai-clients/workloads-manager-client/generated"
)

type WorkloadsManagerExternalClient struct {
	WorkloadsManagerClient
}

type WorkloadsManagerClient struct {
	ApiClient *openapi.APIClient
}

func (wc *WorkloadsManagerClient) BatchWorkloads(ctx context.Context, batch openapi.WorkloadsV2Batch) (*openapi.BatchResponse, *openapi.Error) {
	response, httpResponse, err := wc.ApiClient.
		WorkloadsV2API.BatchWorkloads(ctx).
		WorkloadsV2Batch(batch).Execute()

	if httpResponse != nil {
		defer httpResponse.Body.Close()

		if httpResponse.StatusCode == http.StatusOK {
			return response, nil
		}

		// we need the error as Json, not as []byte
		data, _ := io.ReadAll(httpResponse.Body)

		formattedErr := fmt.Sprintf("failed to execute batch: [%d] %s", httpResponse.StatusCode, data)

		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s", formattedErr)
		return nil, openapi.NewError(int32(httpResponse.StatusCode), formattedErr)
	}

	formattedErr := fmt.Sprintf("failed to request workloads manager service to execute batch: %v", err)

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s", formattedErr)
	return nil, openapi.NewError(http.StatusServiceUnavailable, formattedErr)
}

func (wc *WorkloadsManagerClient) CreateWorkloadV2(ctx context.Context, req openapi.WorkloadV2CreateRequest) (*openapi.WorkloadV2, *openapi.Error) {
	response, httpResponse, err := wc.ApiClient.
		WorkloadsV2API.CreateWorkloadInternal(ctx).
		WorkloadV2CreateRequest(req).Execute()

	if httpResponse != nil {
		defer httpResponse.Body.Close()

		if httpResponse.StatusCode == http.StatusAccepted || httpResponse.StatusCode == http.StatusOK {
			return response, nil
		}

		// we need the error as Json, not as []byte
		data, _ := io.ReadAll(httpResponse.Body)

		formattedErr := fmt.Sprintf("failed to create workload: [%d] %s", httpResponse.StatusCode, data)

		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s", formattedErr)
		return nil, openapi.NewError(int32(httpResponse.StatusCode), formattedErr)
	}

	formattedErr := fmt.Sprintf("failed to request workloads manager service to create workload: %v", err)

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s", formattedErr)
	return nil, openapi.NewError(http.StatusServiceUnavailable, formattedErr)
}

func (wc *WorkloadsManagerClient) GetWorkloadV2ById(ctx context.Context, workloadId string) (*openapi.WorkloadV2GetResponse, *openapi.Error) {
	response, httpResponse, err := wc.ApiClient.
		WorkloadsV2API.GetWorkloadV2ById(ctx, workloadId).Execute()

	if httpResponse != nil {
		defer httpResponse.Body.Close()

		if httpResponse.StatusCode == http.StatusOK {
			return response, nil
		}

		// we need the error as Json, not as []byte
		data, _ := io.ReadAll(httpResponse.Body)

		formattedErr := fmt.Sprintf("failed to get workload: [%d] %s", httpResponse.StatusCode, data)

		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s", formattedErr)
		return nil, openapi.NewError(int32(httpResponse.StatusCode), formattedErr)
	}

	formattedErr := fmt.Sprintf("failed to request workloads manager service to get workload: %v", err)

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s", formattedErr)
	return nil, openapi.NewError(http.StatusServiceUnavailable, formattedErr)
}

func (wc *WorkloadsManagerClient) UpdateWorkloadV2(ctx context.Context, workloadId string, req openapi.WorkloadV2UpdateRequest) (*openapi.WorkloadV2, *openapi.Error) {
	response, httpResponse, err := wc.ApiClient.
		WorkloadsV2API.UpdateWorkloadV2(ctx, workloadId).
		WorkloadV2UpdateRequest(req).Execute()

	if httpResponse != nil {
		defer httpResponse.Body.Close()

		if httpResponse.StatusCode == http.StatusAccepted || httpResponse.StatusCode == http.StatusOK {
			return response, nil
		}

		// we need the error as Json, not as []byte
		data, _ := io.ReadAll(httpResponse.Body)

		formattedErr := fmt.Sprintf("failed to update workload: [%d] %s", httpResponse.StatusCode, data)

		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s", formattedErr)
		return nil, openapi.NewError(int32(httpResponse.StatusCode), formattedErr)
	}

	formattedErr := fmt.Sprintf("failed to request workloads manager service to update workload: %v", err)

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s", formattedErr)
	return nil, openapi.NewError(http.StatusServiceUnavailable, formattedErr)
}

func (wc *WorkloadsManagerClient) UpdateWorkloadInternal(ctx context.Context, workloadId string, req openapi.WorkloadV2UpdateRequest) (*openapi.WorkloadV2, *openapi.Error) {
	response, httpResponse, err := wc.ApiClient.
		WorkloadsV2API.UpdateWorkloadInternal(ctx, workloadId).
		WorkloadV2UpdateRequest(req).Execute()

	if httpResponse != nil {
		defer httpResponse.Body.Close()

		if httpResponse.StatusCode == http.StatusAccepted || httpResponse.StatusCode == http.StatusOK {
			return response, nil
		}

		data, _ := io.ReadAll(httpResponse.Body)

		formattedErr := fmt.Sprintf("failed to update internal workload: [%d] %s", httpResponse.StatusCode, data)

		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s", formattedErr)
		return nil, openapi.NewError(int32(httpResponse.StatusCode), formattedErr)
	}

	formattedErr := fmt.Sprintf("failed to request workloads manager service to update internal workload: %v", err)

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s", formattedErr)
	return nil, openapi.NewError(http.StatusServiceUnavailable, formattedErr)
}
