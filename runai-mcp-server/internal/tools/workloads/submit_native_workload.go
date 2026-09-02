/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloads

import (
	"context"
	"fmt"
	"net/http"
	"reflect"

	workloadsapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/common"
)

type nativeWorkloadSubmission struct {
	WorkloadType string         `json:"workloadType" jsonschema:"Run:ai native workload type. One of: Training, Workspace, Distributed, Inference, DistributedInference."`
	Request      map[string]any `json:"request"      jsonschema:"Complete creation request for the selected native workload API. Supports every field in the corresponding Run:ai creation model, including templateId, assets, and the full spec."` //nolint:lll // long jsonschema tool description; cannot be wrapped in a struct tag
}

type nativeSubmission struct {
	name         string
	projectID    string
	clusterID    string
	workloadType string
	create       func(context.Context, *runaiclient.Client) (string, string, error)
}

type nativeCreate[T nativeCreationRequest] func(context.Context, *runaiclient.Client, T) (nativeResponse, *http.Response, error)

type nativeWorkloadAdapter struct {
	decode  func(map[string]any) (nativeCreationRequest, error)
	prepare func(*nativeWorkloadSubmission) (nativeSubmission, error)
}

var nativeWorkloadAdapters = map[string]nativeWorkloadAdapter{
	workloadTypeTraining:    newNativeWorkloadAdapter("training", createTraining),
	workloadTypeWorkspace:   newNativeWorkloadAdapter("workspace", createWorkspace),
	workloadTypeDistributed: newNativeWorkloadAdapter("distributed workload", createDistributed),
	workloadTypeInference:   newNativeWorkloadAdapter("inference", createInference),
	workloadTypeDistribInf:  newNativeWorkloadAdapter("distributed inference", createDistributedInference),
}

func newNativeWorkloadAdapter[T nativeCreationRequest](kind string, create nativeCreate[T]) nativeWorkloadAdapter {
	return nativeWorkloadAdapter{
		decode: func(raw map[string]any) (nativeCreationRequest, error) {
			return decodeNativeRequest[T](raw)
		},
		prepare: func(in *nativeWorkloadSubmission) (nativeSubmission, error) {
			return prepareNativeRequest[T](in, kind, create)
		},
	}
}

func validateNativeSubmissionInput(in submitWorkloadInput) error {
	if in.DryRun {
		return fmt.Errorf("dryRun is supported only when source=yaml")
	}
	if in.Native == nil {
		return fmt.Errorf("native is required when source=native")
	}
	if in.YAML != nil {
		return fmt.Errorf("yaml must not be set when source=native")
	}
	if in.Native.WorkloadType == "" {
		return fmt.Errorf("native.workloadType is required")
	}
	if len(in.Native.Request) == 0 {
		return fmt.Errorf("native.request is required")
	}

	adapter, ok := nativeWorkloadAdapters[in.Native.WorkloadType]
	if !ok {
		return unsupportedNativeWorkloadTypeError()
	}
	req, err := adapter.decode(in.Native.Request)
	if err != nil {
		return err
	}
	if _, err := common.RequiredString("native.request.name", req.GetName()); err != nil {
		return err
	}
	if err := common.ValidateNumericString("native.request.projectId", req.GetProjectId()); err != nil {
		return err
	}
	return common.ValidateUUID("native.request.clusterId", req.GetClusterId())
}

func submitNativeWorkload(ctx context.Context, client *runaiclient.Client, in submitWorkloadInput) (submitWorkloadOutput, error) {
	prepared, err := prepareNativeSubmission(in.Native)
	if err != nil {
		return submitWorkloadOutput{}, err
	}
	if _, err := resolveSubmissionProject(ctx, client, prepared.projectID, prepared.clusterID); err != nil {
		return submitWorkloadOutput{}, err
	}

	out := submitWorkloadOutput{
		Name:         prepared.name,
		WorkloadType: prepared.workloadType,
		ProjectID:    prepared.projectID,
		ClusterID:    prepared.clusterID,
	}
	if err := confirmSubmission(ctx, in.Confirm, out, submitSourceNative); err != nil {
		return out, err
	}

	id, name, err := prepared.create(ctx, client)
	if err != nil {
		return submitWorkloadOutput{}, err
	}
	out.ID = id
	if name != "" {
		out.Name = name
	}
	return out, nil
}

func prepareNativeSubmission(in *nativeWorkloadSubmission) (nativeSubmission, error) {
	if in == nil {
		return nativeSubmission{}, fmt.Errorf("native is required: %w", runaimcp.ErrValidation)
	}
	adapter, ok := nativeWorkloadAdapters[in.WorkloadType]
	if !ok {
		return nativeSubmission{}, unsupportedNativeWorkloadTypeError()
	}
	return adapter.prepare(in)
}

func unsupportedNativeWorkloadTypeError() error {
	return fmt.Errorf("native.workloadType must be one of: %s, %s, %s, %s, %s: %w",
		workloadTypeTraining, workloadTypeWorkspace, workloadTypeDistributed, workloadTypeInference, workloadTypeDistribInf, runaimcp.ErrValidation)
}

func createTraining(ctx context.Context, client *runaiclient.Client, req *workloadsapi.TrainingCreationRequest) (nativeResponse, *http.Response, error) {
	return client.Workloads.TrainingsAPI.CreateTraining(ctx).TrainingCreationRequest(*req).Execute()
}

func createWorkspace(ctx context.Context, client *runaiclient.Client, req *workloadsapi.WorkspaceCreationRequest) (nativeResponse, *http.Response, error) {
	return client.Workloads.WorkspacesAPI.CreateWorkspace(ctx).WorkspaceCreationRequest(*req).Execute()
}

func createDistributed(ctx context.Context, client *runaiclient.Client, req *workloadsapi.DistributedCreationRequest) (nativeResponse, *http.Response, error) {
	return client.Workloads.DistributedAPI.CreateDistributed(ctx).DistributedCreationRequest(*req).Execute()
}

func createInference(ctx context.Context, client *runaiclient.Client, req *workloadsapi.InferenceCreationRequest) (nativeResponse, *http.Response, error) {
	return client.Workloads.InferencesAPI.CreateInference(ctx).InferenceCreationRequest(*req).Execute()
}

func createDistributedInference(ctx context.Context, client *runaiclient.Client, req *workloadsapi.DistributedInferenceCreationRequest) (nativeResponse, *http.Response, error) {
	return client.Workloads.DistributedInferencesAPI.CreateDistributedInference(ctx).DistributedInferenceCreationRequest(*req).Execute()
}

type nativeCreationRequest interface {
	GetName() string
	GetProjectId() string
	GetClusterId() string
}

func prepareNativeRequest[T nativeCreationRequest](in *nativeWorkloadSubmission, kind string, create func(context.Context, *runaiclient.Client, T) (nativeResponse, *http.Response, error)) (nativeSubmission, error) {
	req, err := decodeNativeRequest[T](in.Request)
	if err != nil {
		return nativeSubmission{}, err
	}
	return nativeSubmission{
		name:         req.GetName(),
		projectID:    req.GetProjectId(),
		clusterID:    req.GetClusterId(),
		workloadType: in.WorkloadType,
		create: func(ctx context.Context, client *runaiclient.Client) (string, string, error) {
			created, resp, err := create(ctx, client, req)
			return parseNativeCreationResult(kind, created, resp, err)
		},
	}, nil
}

func decodeNativeRequest[T any](raw map[string]any) (T, error) {
	var out T
	err := common.DecodeTypedRequest(raw, &out)
	if err != nil {
		return out, fmt.Errorf("decode native.request: %v: %w", err, runaimcp.ErrValidation)
	}
	return out, nil
}

type nativeResponse interface {
	GetWorkloadId() string
	GetName() string
}

func parseNativeCreationResult(kind string, created nativeResponse, resp *http.Response, err error) (string, string, error) {
	if err != nil {
		mapped := runaimcp.MapHTTPError(resp, err)
		if resp == nil || resp.StatusCode < http.StatusMultipleChoices {
			return "", "", fmt.Errorf("create %s: request outcome is unknown; inspect workloads before retrying: %w", kind, mapped)
		}
		return "", "", fmt.Errorf("create %s: %w", kind, mapped)
	}
	if isNilNativeResponse(created) {
		return "", "", fmt.Errorf("create %s: nil response; creation outcome is unknown, inspect workloads before retrying: %w", kind, runaimcp.ErrUpstream)
	}
	return created.GetWorkloadId(), created.GetName(), nil
}

func isNilNativeResponse(response nativeResponse) bool {
	if response == nil {
		return true
	}
	value := reflect.ValueOf(response)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}
