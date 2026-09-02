/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"

	workloadsServerApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	workloadsMgr "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/workloads_mgr"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/services_integration"
	workloadsApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	workloadMgrApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client/generated"
)

type workloadsServiceConnectionMetadata struct {
	Environments        []workloadsServerApi.Environment `json:"environments,omitempty"`
	ExternalConnections []workloadsServerApi.Connection  `json:"externalConnections,omitempty"`
}

// sendSubmissionToWorkloadMgr sends a prepared manifest to workloads-manager and writes the
// returned workload id and desired phase back into params.Response.
func sendSubmissionToWorkloadMgr(ctx context.Context, params *createWorkloadParams, submission *Submission) *workloadsApi.SubmissionErrorV2 {
	workload, createErr := workloadsMgr.Get().CreateWorkloadV2(ctx, submission.Request)
	if createErr != nil {
		return workloadsApi.NewSubmissionErrorV2(createErr.Code, createErr.Message)
	}

	params.Response.WorkloadId = workload.Metadata.Id
	params.Response.DesiredPhase = workloadsApi.WorkloadDesiredPhase(workload.DesiredPhase)

	parsedId, err := uuid.Parse(params.Response.WorkloadId)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to parse workload manager assigned id: %v", err)
	} else {
		params.WorkloadId = parsedId
	}

	if err := persistWorkloadsServiceConnectionMetadata(ctx, workload, submission.ConnectionMetadata, submission.ConnectionCreationFields); err != nil {
		// Workloads-manager already accepted the workload; keep the create response successful
		// and rely on its creation notification plus later cluster sync updates as fallbacks.
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to persist workloads-service connection metadata for workload %s: %v", params.Response.WorkloadId, err)
	}

	return nil
}

func persistWorkloadsServiceConnectionMetadata(
	ctx context.Context,
	workload *workloadMgrApi.WorkloadV2,
	metadata *WorkloadsServiceConnectionMetadata,
	connectionCreationFields *workloadsApi.WorkloadCreationFields,
) error {
	if metadata == nil || (len(metadata.Environments) == 0 && len(metadata.ExternalConnections) == 0) {
		return nil
	}
	if workload == nil {
		return fmt.Errorf("missing workloads-manager workload response")
	}

	serverMetadata, err := workloadsClientConnectionMetadataToServerMetadata(metadata)
	if err != nil {
		return fmt.Errorf("failed to convert workloads-service metadata: %w", err)
	}

	creationFields, err := workloadsServiceCreationFieldsForConnectionMetadata(workload, connectionCreationFields, serverMetadata)
	if err != nil {
		return fmt.Errorf("failed to prepare workloads-service fallback metadata: %w", err)
	}
	if creationFields == nil {
		notifyErr := services_integration.UpdateWorkloadConnectionMetadata(ctx, workload.Metadata.Id, serverMetadata.Environments, serverMetadata.ExternalConnections)
		if notifyErr != nil {
			return fmt.Errorf("%s", notifyErr.Message)
		}
		return nil
	}

	notifyErr := services_integration.EnsureWorkloadConnectionMetadata(ctx, *creationFields, serverMetadata.Environments, serverMetadata.ExternalConnections)
	if notifyErr != nil {
		return fmt.Errorf("%s", notifyErr.Message)
	}

	return nil
}

func workloadsServiceCreationFieldsForConnectionMetadata(
	workload *workloadMgrApi.WorkloadV2,
	connectionCreationFields *workloadsApi.WorkloadCreationFields,
	metadata workloadsServiceConnectionMetadata,
) (*workloadsServerApi.WorkloadCreationFields, error) {
	if workload == nil || connectionCreationFields == nil {
		return nil, nil
	}

	creationFields, err := workloadsClientFieldToServerFields(*connectionCreationFields)
	if err != nil {
		return nil, err
	}

	workloadMetadata := workload.Metadata
	creationFields.Id = workloadMetadata.Id
	creationFields.Name = workloadMetadata.Name
	creationFields.SubmittedBy = workloadMetadata.CreatedBy
	creationFields.ClusterId = workloadMetadata.ClusterId
	creationFields.ProjectId = workloadMetadata.ProjectId
	creationFields.ProjectName = workloadMetadata.ProjectName
	creationFields.DepartmentId = workloadMetadata.DepartmentId
	creationFields.DepartmentName = workloadMetadata.DepartmentName
	creationFields.CreatedAt = workloadMetadata.CreatedAt
	creationFields.Group = convutils.Deref(workloadMetadata.Gvk.Group.Get(), "")
	creationFields.Version = workloadMetadata.Gvk.Version
	creationFields.Environments = metadata.Environments
	creationFields.ExternalConnections = metadata.ExternalConnections
	creationFields.FullySaasWorkload = true

	if category := workloadMetadata.Category.Get(); category != nil {
		creationFields.Category = *category
	}
	if preemptibility := workloadMetadata.Preemptibility.Get(); preemptibility != nil {
		preemptibilityValue := string(*preemptibility)
		creationFields.Preemptibility = &preemptibilityValue
	}

	sourceApi := workloadsServerApi.SOURCEAPI_WORKLOADS_V2
	creationFields.SourceApi = &sourceApi

	return &creationFields, nil
}

func workloadsClientConnectionMetadataToServerMetadata(metadata *WorkloadsServiceConnectionMetadata) (workloadsServiceConnectionMetadata, error) {
	serverMetadata := workloadsServiceConnectionMetadata{}
	bytes, err := json.Marshal(metadata)
	if err != nil {
		return serverMetadata, err
	}
	if err := json.Unmarshal(bytes, &serverMetadata); err != nil {
		return serverMetadata, err
	}
	return serverMetadata, nil
}

// sendUpdateToWorkloadMgr sends an updated typed-V1 manifest through the same internal API family used to create it.
func sendUpdateToWorkloadMgr(ctx context.Context, workloadId string, update *Update) (*workloadMgrApi.WorkloadV2, *workloadsApi.SubmissionErrorV2) {
	workload, updateErr := workloadsMgr.Get().UpdateWorkloadInternal(ctx, workloadId, update.Request)
	if updateErr != nil {
		return nil, workloadsApi.NewSubmissionErrorV2(updateErr.Code, updateErr.Message)
	}
	if workload == nil {
		return nil, workloadsApi.NewSubmissionErrorV2(http.StatusServiceUnavailable, "workloads-manager returned no workload after a successful update")
	}
	return workload, nil
}
