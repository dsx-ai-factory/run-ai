/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloads_service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	workloadsServiceClient "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

// Client -> for the "real" client we need an APIClient to the workloads service
type Client struct {
	apiClient *workloadsServiceClient.APIClient
}

func (c Client) CreateWorkspace(ctx context.Context, workspace *workloadsServiceClient.WorkspaceCreationRequest) (*workloadsServiceClient.Workspace, *workloadsServiceClient.SubmissionErrorV2) {
	resp, r, err := c.apiClient.WorkspacesAPI.CreateWorkspace(ctx).
		WorkspaceCreationRequest(*workspace).Execute()

	if r != nil {
		if r.StatusCode == http.StatusAccepted {
			return resp, nil
		}
		data, _ := io.ReadAll(r.Body)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to create workspace %v: [%d] %s", *workspace, r.StatusCode, string(data))

		var wsError *workloadsServiceClient.SubmissionErrorV2

		err := json.Unmarshal(data, &wsError)
		if err != nil {
			return nil, workloadsServiceClient.NewSubmissionErrorV2(int32(r.StatusCode), fmt.Sprintf("failed to create workspace: %s", string(data)))
		} else {
			return nil, wsError
		}
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed requesting workload service to create workspace %v: %v", *workspace, err)
	return nil, workloadsServiceClient.NewSubmissionErrorV2(http.StatusServiceUnavailable, fmt.Sprintf("failed to request workspace creation: %v", err))
}

func (c Client) CreateTraining(ctx context.Context, training *workloadsServiceClient.TrainingCreationRequest) (*workloadsServiceClient.Training, *workloadsServiceClient.SubmissionErrorV2) {
	resp, r, err := c.apiClient.TrainingsAPI.CreateTraining(ctx).
		TrainingCreationRequest(*training).Execute()

	if r != nil {
		if r.StatusCode == http.StatusAccepted {
			return resp, nil
		}
		data, _ := io.ReadAll(r.Body)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to create training %v: [%d] %s", *training, r.StatusCode, string(data))

		var wsError *workloadsServiceClient.SubmissionErrorV2

		err := json.Unmarshal(data, &wsError)
		if err != nil {
			return nil, workloadsServiceClient.NewSubmissionErrorV2(int32(r.StatusCode), fmt.Sprintf("failed to create training: %s", string(data)))
		} else {
			return nil, wsError
		}
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed requesting workload service to create training %v: %v", *training, err)
	return nil, workloadsServiceClient.NewSubmissionErrorV2(http.StatusServiceUnavailable, fmt.Sprintf("failed to request training creation: %v", err))
}

func (c Client) CreateDistributed(ctx context.Context, distributed *workloadsServiceClient.DistributedCreationRequest) (*workloadsServiceClient.DistributedWorkload, *workloadsServiceClient.SubmissionErrorV2) {
	resp, r, err := c.apiClient.DistributedAPI.CreateDistributed(ctx).
		DistributedCreationRequest(*distributed).Execute()

	if r != nil {
		if r.StatusCode == http.StatusAccepted {
			return resp, nil
		}

		data, _ := io.ReadAll(r.Body)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to create distributed workload %v: [%d] %s", *distributed, r.StatusCode, string(data))

		var wsError *workloadsServiceClient.SubmissionErrorV2

		err := json.Unmarshal(data, &wsError)
		if err != nil {
			return nil, workloadsServiceClient.NewSubmissionErrorV2(int32(r.StatusCode), fmt.Sprintf("failed to create distributed workload: %s", string(data)))
		} else {
			return nil, wsError
		}
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed requesting workload service to create distributed %v: %v", *distributed, err)
	return nil, workloadsServiceClient.NewSubmissionErrorV2(http.StatusServiceUnavailable, fmt.Sprintf("failed to request distributed workload creation: %v", err))
}

func (c Client) CreateInference(ctx context.Context, inference *workloadsServiceClient.InferenceCreationRequest) (*workloadsServiceClient.Inference, *workloadsServiceClient.Error) {
	resp, r, err := c.apiClient.InferencesAPI.CreateInference(ctx).
		InferenceCreationRequest(*inference).Execute()

	if r != nil {
		if r.StatusCode == http.StatusAccepted {
			return resp, nil
		}
		data, _ := io.ReadAll(r.Body)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to create inference %v: [%d] %s", *inference, r.StatusCode, string(data))

		var wsError *workloadsServiceClient.Error

		err := json.Unmarshal(data, &wsError)
		if err != nil {
			return nil, workloadsServiceClient.NewError(int32(r.StatusCode), fmt.Sprintf("failed to create inference: %s", string(data)))
		} else {
			return nil, workloadsServiceClient.NewError(int32(r.StatusCode), fmt.Sprintf("failed to create inference: %s", wsError.Message))
		}
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed requesting workload service to create inference %v: %v", *inference, err)
	return nil, workloadsServiceClient.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to request inference creation: %v", err))
}

// Factory -> the "real" client factory. during testing, it is replaced with factory which
// generated mock clients
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
