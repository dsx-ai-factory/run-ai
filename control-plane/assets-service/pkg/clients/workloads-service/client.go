/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloads_service

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"strings"

	workloadsServiceClient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

// Client -> for the "real" client we need an APIClient to the workloads service
type Client struct {
	apiClient *workloadsServiceClient.APIClient
}

const downstreamErrorResponseUnavailable = "downstream error response unavailable"

const (
	huggingFaceTokenEnvironmentVariable = "HF_TOKEN"
	ngcAPIKeyEnvironmentVariable        = "NGC_API_KEY"
)

func (c Client) CreateWorkspace(ctx context.Context, workspace *workloadsServiceClient.WorkspaceCreationRequest) (*workloadsServiceClient.Workspace, *workloadsServiceClient.SubmissionErrorV2) {
	resp, r, err := c.apiClient.WorkspacesAPI.CreateWorkspace(ctx).
		WorkspaceCreationRequest(*workspace).Execute()

	if r != nil {
		if r.StatusCode == http.StatusAccepted {
			return resp, nil
		}
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to create workspace: downstream status %d", r.StatusCode)
		return nil, submissionErrorFromDownstreamResponse(r.StatusCode, r.Body, workspace, "workspace")
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed requesting workload service to create workspace (%T)", err)
	return nil, workloadsServiceClient.NewSubmissionErrorV2(http.StatusServiceUnavailable, "failed to request workspace creation")
}

func (c Client) CreateTraining(ctx context.Context, training *workloadsServiceClient.TrainingCreationRequest) (*workloadsServiceClient.Training, *workloadsServiceClient.SubmissionErrorV2) {
	resp, r, err := c.apiClient.TrainingsAPI.CreateTraining(ctx).
		TrainingCreationRequest(*training).Execute()

	if r != nil {
		if r.StatusCode == http.StatusAccepted {
			return resp, nil
		}
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to create training: downstream status %d", r.StatusCode)
		return nil, submissionErrorFromDownstreamResponse(r.StatusCode, r.Body, training, "training")
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed requesting workload service to create training (%T)", err)
	return nil, workloadsServiceClient.NewSubmissionErrorV2(http.StatusServiceUnavailable, "failed to request training creation")
}

func (c Client) CreateDistributed(ctx context.Context, distributed *workloadsServiceClient.DistributedCreationRequest) (*workloadsServiceClient.DistributedWorkload, *workloadsServiceClient.SubmissionErrorV2) {
	resp, r, err := c.apiClient.DistributedAPI.CreateDistributed(ctx).
		DistributedCreationRequest(*distributed).Execute()

	if r != nil {
		if r.StatusCode == http.StatusAccepted {
			return resp, nil
		}

		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to create distributed workload: downstream status %d", r.StatusCode)
		return nil, submissionErrorFromDownstreamResponse(r.StatusCode, r.Body, distributed, "distributed workload")
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed requesting workload service to create distributed workload (%T)", err)
	return nil, workloadsServiceClient.NewSubmissionErrorV2(http.StatusServiceUnavailable, "failed to request distributed workload creation")
}

func (c Client) CreateInference(ctx context.Context, inference *workloadsServiceClient.InferenceCreationRequest) (*workloadsServiceClient.Inference, *workloadsServiceClient.Error) {
	resp, r, err := c.apiClient.InferencesAPI.CreateInference(ctx).
		InferenceCreationRequest(*inference).Execute()

	if r != nil {
		if r.StatusCode == http.StatusAccepted {
			return resp, nil
		}
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to create inference: downstream status %d", r.StatusCode)
		data, readErr := readAndRedactSensitiveDownstreamBody(r.Body, inference)
		if readErr != nil {
			return nil, workloadsServiceClient.NewError(int32(r.StatusCode), downstreamErrorResponseUnavailable)
		}

		var wsError *workloadsServiceClient.Error

		if unmarshalErr := json.Unmarshal(data, &wsError); unmarshalErr != nil {
			return nil, workloadsServiceClient.NewError(int32(r.StatusCode), fmt.Sprintf("failed to create inference: %s", string(data)))
		}
		if wsError == nil {
			return nil, workloadsServiceClient.NewError(int32(r.StatusCode), "failed to create inference: invalid downstream error response")
		}
		return nil, workloadsServiceClient.NewError(int32(r.StatusCode), fmt.Sprintf("failed to create inference: %s", wsError.Message))
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed requesting workload service to create inference (%T)", err)
	return nil, workloadsServiceClient.NewError(http.StatusServiceUnavailable, "failed to request inference creation")
}

func submissionErrorFromDownstreamResponse(
	statusCode int,
	body io.Reader,
	request interface{},
	operation string,
) *workloadsServiceClient.SubmissionErrorV2 {
	data, readErr := readAndRedactSensitiveDownstreamBody(body, request)
	if readErr != nil {
		return workloadsServiceClient.NewSubmissionErrorV2(int32(statusCode), downstreamErrorResponseUnavailable)
	}

	var responseError *workloadsServiceClient.SubmissionErrorV2
	if unmarshalErr := json.Unmarshal(data, &responseError); unmarshalErr != nil {
		return workloadsServiceClient.NewSubmissionErrorV2(
			int32(statusCode),
			fmt.Sprintf("failed to create %s: %s", operation, string(data)),
		)
	}
	if responseError == nil {
		return workloadsServiceClient.NewSubmissionErrorV2(
			int32(statusCode),
			fmt.Sprintf("failed to create %s: invalid downstream error response", operation),
		)
	}
	return responseError
}

// Factory -> the "real" client factory. during testing, it is replaced with factory which
// generated mock clients

// readAndRedactSensitiveDownstreamBody fails closed instead of processing a partial error body.
func readAndRedactSensitiveDownstreamBody(body io.Reader, request interface{}) ([]byte, error) {
	data, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}
	return redactSensitiveValuesFromDownstreamBody(data, request), nil
}

// redactSensitiveValuesFromDownstreamBody removes submitted token values from downstream errors.
func redactSensitiveValuesFromDownstreamBody(body []byte, request interface{}) []byte {
	requestBytes, err := json.Marshal(request)
	if err != nil {
		return []byte(`{"message":"downstream error body omitted because the request could not be inspected"}`)
	}
	var requestBody interface{}
	if err := json.Unmarshal(requestBytes, &requestBody); err != nil {
		return []byte(`{"message":"downstream error body omitted because the request could not be inspected"}`)
	}

	values := collectSensitiveEnvironmentVariableValues(requestBody)
	var structuredBody interface{}
	if err := json.Unmarshal(body, &structuredBody); err == nil {
		redactedBody, marshalErr := json.Marshal(redactSensitiveStringsInDownstreamBody(structuredBody, values))
		if marshalErr != nil {
			return []byte(`{"message":"downstream error body omitted because it could not be sanitized"}`)
		}
		return redactedBody
	}

	return []byte(redactSensitiveStringInDownstreamBody(string(body), values))
}

func redactSensitiveStringsInDownstreamBody(value interface{}, sensitiveValues []string) interface{} {
	switch typed := value.(type) {
	case string:
		return redactSensitiveStringInDownstreamBody(typed, sensitiveValues)
	case map[string]interface{}:
		for key, child := range typed {
			typed[key] = redactSensitiveStringsInDownstreamBody(child, sensitiveValues)
		}
		return typed
	case []interface{}:
		for index, child := range typed {
			typed[index] = redactSensitiveStringsInDownstreamBody(child, sensitiveValues)
		}
		return typed
	default:
		return typed
	}
}

func redactSensitiveStringInDownstreamBody(value string, sensitiveValues []string) string {
	redacted := value
	for _, sensitiveValue := range sensitiveValues {
		for _, representation := range sensitiveValueRepresentations(sensitiveValue) {
			if representation != "" {
				redacted = strings.ReplaceAll(redacted, representation, "[REDACTED]")
			}
		}
	}
	return redacted
}

func sensitiveValueRepresentations(value string) []string {
	representations := []string{value, html.EscapeString(value)}
	if jsonValue, err := json.Marshal(value); err == nil && len(jsonValue) >= 2 {
		representations = append(representations, string(jsonValue[1:len(jsonValue)-1]))
	}
	return representations
}

func collectSensitiveEnvironmentVariableValues(value interface{}) []string {
	var result []string
	switch typed := value.(type) {
	case map[string]interface{}:
		if name, ok := typed["name"].(string); ok && isSensitivePlaintextEnvironmentVariable(name) {
			if literal, ok := typed["value"].(string); ok {
				result = append(result, literal)
			}
		}
		for _, child := range typed {
			result = append(result, collectSensitiveEnvironmentVariableValues(child)...)
		}
	case []interface{}:
		for _, child := range typed {
			result = append(result, collectSensitiveEnvironmentVariableValues(child)...)
		}
	}
	return result
}

func isSensitivePlaintextEnvironmentVariable(name string) bool {
	return name == huggingFaceTokenEnvironmentVariable || name == ngcAPIKeyEnvironmentVariable
}

// Factory creates the workloads-service client used in production.
type Factory struct{}

// Client returns a workload service client
func (f Factory) Client() ClientInterface {
	return &Client{
		apiClient: workloadsServiceClient.NewAPIClient(serviceConfig),
	}
}

// InternalCreateWorkload notifies the workload service on a new workload that was created
func (c Client) InternalCreateWorkload(ctx context.Context, notification *workloadsServiceClient.WorkloadCreationFields) (*workloadsServiceClient.Workload, *workloadsServiceClient.Error) {
	resp, r, err := c.apiClient.WorkloadsAPI.InternalCreateWorkload(ctx).
		WorkloadCreationFields(*notification).
		Execute()
	if r != nil {
		if r.StatusCode == http.StatusCreated {
			return resp, nil
		}
		data, _ := io.ReadAll(r.Body)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to notify on workloads service about the creation of workload, status %d creation md %+v, data: %s", r.StatusCode, *notification, data)
		return nil, workloadsServiceClient.NewError(int32(r.StatusCode), fmt.Sprintf("failed to notify on workloads service about the creation of workload %+v", *notification))
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to notify on workloads service about the creation of workload, creation md %+v: %v", *notification, err)
	return nil, workloadsServiceClient.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to notify on workloads service about the creation of workload, creation md %+v: %v", *notification, err))
}

// InternalGetResourceClaimTemplateWorkloads lists the workloads in the request's scope whose pods
// reference the resource claim template with the request's name
func (c Client) InternalGetResourceClaimTemplateWorkloads(ctx context.Context,
	request *workloadsServiceClient.InternalResourceClaimTemplateWorkloadsRequest, offset int32, limit int32,
) (*workloadsServiceClient.ResourceClaimTemplateWorkloadsResponse, *workloadsServiceClient.Error) {
	resp, r, err := c.apiClient.WorkloadsAPI.InternalGetResourceClaimTemplateWorkloads(ctx).
		InternalResourceClaimTemplateWorkloadsRequest(*request).
		Offset(offset).
		Limit(limit).
		Execute()
	if err == nil && r != nil && r.StatusCode == http.StatusOK {
		return resp, nil
	}

	if r != nil && r.StatusCode >= http.StatusBadRequest {
		data, _ := io.ReadAll(r.Body)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to list the workloads of resource claim template %s in %s %s, status %d, data: %s",
			request.Name, request.ScopeType, request.ScopeId, r.StatusCode, data)
		return nil, workloadsServiceClient.NewError(int32(r.StatusCode), fmt.Sprintf("failed to list the workloads of resource claim template %s", request.Name))
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed requesting workloads service to list the workloads of resource claim template %s in %s %s: %v",
		request.Name, request.ScopeType, request.ScopeId, err)
	return nil, workloadsServiceClient.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to request the workloads of resource claim template %s", request.Name))
}
