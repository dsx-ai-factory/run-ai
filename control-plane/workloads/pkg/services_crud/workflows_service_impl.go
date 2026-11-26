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
	"slices"

	wfv1 "github.com/argoproj/argo-workflows/v4/pkg/apis/workflow/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer/composer_v2/workflows"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/gpucompat"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/topology"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/dao_crud"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models_crud"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/services_crud/validators"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils/gojsonq"
	workloadMgrApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client/generated"
	logging "github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

const workflowTypeGroup = "argoproj.io"

func createWorkflow(ctx context.Context, params *createWorkloadParams, spec *workloadsApi.WorkflowSpecSpec) *workloadsApi.SubmissionErrorV2 {
	if err := validateWorkflowComputeProfiles(params.Spec); err != nil {
		return workloadsApi.NewSubmissionErrorV2(http.StatusBadRequest, err.Error())
	}

	target, targetErr := resolveSubmissionTarget(ctx, params.Response.ProjectId, params.Response.ClusterId)
	if targetErr != nil {
		return targetErr
	}

	if identityErr := assignWorkloadIdentity(ctx, params, target); identityErr != nil {
		return identityErr
	}

	if fieldsErr := applyWorkflowFields(ctx, target, params.Response.ClusterId, spec); fieldsErr != nil {
		return fieldsErr
	}

	references := workflowSecretReferences(spec)
	if imagePullSecretErr := preValidateImagePullSecrets(ctx, params.Response.ClusterId, params.Response.ProjectId, references...); imagePullSecretErr != nil {
		return imagePullSecretErr
	}

	assetsMapping, credentialsProcessErr := processCredentials(ctx, params.Response.ClusterId, references...)
	if credentialsProcessErr != nil {
		return credentialsProcessErr
	}
	params.Response.AssetsMapping = assetsMapping

	if securityErr := setTokenUIDGIDForWorkflowTasks(ctx, spec); securityErr != nil {
		return securityErr
	}

	if nodePoolsErr := validateWorkflowNodePools(ctx, params.Response.ClusterId, spec); nodePoolsErr != nil {
		return nodePoolsErr
	}

	workflow := &wfv1.Workflow{}
	workflowComposer := workflows.NewWorkflowComposer(params.WorkloadType, params.Response.Name, target.project.Name, target.namespace, params.Response.CreatedBy)
	if composerErr := workflowComposer.Compose(spec, workflow)(); composerErr != nil {
		return workloadsApi.NewSubmissionErrorV2(composerErr.Code, composerErr.Message)
	}

	manifest, manifestErr := gojsonq.LoadJson(workflow)
	if manifestErr != nil {
		return workloadsApi.NewSubmissionErrorV2(http.StatusInternalServerError, manifestErr.Error())
	}
	// The Argo type serializes an unset status with null timestamps, which the Workflow CRD schema rejects on create.
	delete(manifest, "status")

	submission := &Submission{
		Request: workloadMgrApi.WorkloadV2CreateRequest{
			Metadata: buildWorkflowV2Metadata(params, spec),
			Manifest: manifest,
		},
	}
	if submitErr := sendSubmissionToWorkloadMgr(ctx, params, submission); submitErr != nil {
		return submitErr
	}

	persistWorkflowSnapshot(ctx, params, spec, target.namespace)

	return nil
}

func buildWorkflowV2Metadata(params *createWorkloadParams, spec *workloadsApi.WorkflowSpecSpec) workloadMgrApi.WorkloadV2MetadataCreateParams {
	metadata := workloadMgrApi.WorkloadV2MetadataCreateParams{
		Name:      &params.Response.Name,
		ProjectId: &params.Response.ProjectId,
	}
	if category := spec.Category.Get(); category != nil {
		metadata.Category.Set(category)
	}
	if priorityClass := spec.PriorityClass.Get(); priorityClass != nil {
		metadata.Priority.Set(priorityClass)
	}
	if preemptibility := spec.Preemptibility.Get(); preemptibility != nil {
		convertedPreemptibility := workloadMgrApi.Preemptibility(*preemptibility)
		metadata.Preemptibility.Set(&convertedPreemptibility)
	}
	return metadata
}

// persistWorkflowSnapshot stores the workflow spec for a later get. Workloads-manager has already accepted the
// workflow at this point, so a failure is logged and does not fail the submission.
func persistWorkflowSnapshot(ctx context.Context, params *createWorkloadParams, spec *workloadsApi.WorkflowSpecSpec, namespace string) {
	snapshot := &models_crud.WorkflowSnapshot{
		ID:            params.WorkloadId,
		CreatedAt:     params.Response.CreatedAt,
		CreatedBy:     params.Response.CreatedBy,
		UpdatedAt:     params.Response.CreatedAt,
		UpdatedBy:     params.Response.CreatedBy,
		TenantId:      params.TenantId,
		ClusterId:     params.Response.ClusterId,
		DepartmentId:  convutils.Deref(params.Response.DepartmentId, ""),
		ProjectId:     params.Response.ProjectId,
		WorkloadType:  params.WorkloadType,
		RequestedName: params.Response.RequestedName,
		Name:          params.Response.Name,
		Namespace:     namespace,
		Spec:          spec,
		AssetsMapping: params.Response.AssetsMapping,
	}
	if dbErr := dao_crud.DbUpsertWorkflowSnapshot(ctx, snapshot); dbErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to persist workflow snapshot for workflow %s: %v", snapshot.ID, dbErr.Message)
	}
}

func validateWorkflowComputeProfiles(spec map[string]interface{}) error {
	profiles, _ := spec["compute"].([]interface{})
	for index, profile := range profiles {
		if profileFields, isMap := profile.(map[string]interface{}); isMap {
			if err := gpucompat.ValidateComputeMemoryRequestInput(profileFields, fmt.Sprintf("compute[%d]", index)); err != nil {
				return err
			}
		}
	}
	return nil
}

func workflowSecretReferences(spec *workloadsApi.WorkflowSpecSpec) []secretReferences {
	var references []secretReferences
	for _, group := range spec.Groups {
		for _, task := range group.Tasks {
			references = append(references, secretReferences{environmentVariables: task.EnvironmentVariables, imagePullSecrets: task.ImagePullSecrets})
		}
	}
	return references
}

func setTokenUIDGIDForWorkflowTasks(ctx context.Context, spec *workloadsApi.WorkflowSpecSpec) *workloadsApi.SubmissionErrorV2 {
	for groupIndex := range spec.Groups {
		for taskIndex := range spec.Groups[groupIndex].Tasks {
			task := &spec.Groups[groupIndex].Tasks[taskIndex]
			if task.Security.Get() == nil {
				continue
			}
			security, setTokenErr := setTokenUIDGIDForSecurity(ctx, *task.Security.Get())
			if setTokenErr != nil {
				return workloadsApi.NewSubmissionErrorV2(setTokenErr.Code, setTokenErr.Message)
			}
			if security != nil {
				task.SetSecurity(*security)
			}
		}
	}
	return nil
}

func validateWorkflowNodePools(ctx context.Context, clusterId string, spec *workloadsApi.WorkflowSpecSpec) *workloadsApi.SubmissionErrorV2 {
	actualNodePools, nodePoolsFetchErr := getClusterNodePools(ctx, clusterId)
	if nodePoolsFetchErr != nil {
		return nodePoolsFetchErr
	}

	nodePools := slices.Clone(spec.NodePools)
	for _, group := range spec.Groups {
		nodePools = append(nodePools, group.NodePools...)
	}
	if _, checkErr := topology.CheckNodePoolsNames(ctx, actualNodePools, nodePools); checkErr != nil {
		return NewSubmissionErrorWithDetails(http.StatusBadRequest, checkErr.Message, checkErr.Details)
	}
	return nil
}

// applyWorkflowFields defaults and validates the workflow-level fields. The category, priority and preemptibility
// defaults come from the Argo Workflow workload type config, so admins can change them per tenant.
func applyWorkflowFields(ctx context.Context, target *submissionTarget, clusterId string, spec *workloadsApi.WorkflowSpecSpec) *workloadsApi.SubmissionErrorV2 {
	if categoryErr := validators.ValidateWorkloadCategory(ctx, spec.Category); categoryErr != nil {
		return categoryErr
	}

	if !convutils.HasValue(spec.Category) || !convutils.HasValue(spec.PriorityClass) {
		typeConfig, typeConfigErr := dao.DbGetWorkloadTypeConfigByNameGroup(ctx, target.tenantId, Workflow, workflowTypeGroup)
		if typeConfigErr != nil {
			return ToSubmissionError(typeConfigErr)
		}

		if !convutils.HasValue(spec.Category) && typeConfig.CategoryName != "" {
			spec.SetCategory(typeConfig.CategoryName)
		}

		if !convutils.HasValue(spec.PriorityClass) && typeConfig.PriorityName != "" {
			spec.SetPriorityClass(typeConfig.PriorityName)
			if !convutils.HasValue(spec.Preemptibility) {
				if preemptibility, valueErr := workloadsApi.NewPreemptibilityFromValue(typeConfig.Preemptibility); valueErr == nil {
					spec.SetPreemptibility(*preemptibility)
				}
			}
		}
	}

	priorityValues, priorityErr := validators.ValidateCustomPriorityClass(ctx, clusterId, spec.PriorityClass)
	if priorityErr != nil {
		return priorityErr
	}

	if !convutils.HasValue(spec.Preemptibility) {
		if preemptibility, found := preemptibilityOfPriority(spec.PriorityClass, priorityValues); found {
			spec.SetPreemptibility(preemptibility)
		}
	}

	if len(spec.NodePools) == 0 {
		spec.NodePools = getDefaultNodepools(target.project)
	}

	return nil
}

func loadWorkflowSpec(spec map[string]interface{}) (*workloadsApi.WorkflowSpecSpec, *workloadsApi.SubmissionErrorV2) {
	var workflowSpec workloadsApi.WorkflowSpecSpec
	if err := json.Unmarshal(gojsonq.ToBytes(spec), &workflowSpec); err != nil {
		return nil, workloadsApi.NewSubmissionErrorV2(http.StatusBadRequest, fmt.Sprintf("invalid workflow spec: %v", err))
	}
	return &workflowSpec, nil
}

func newWorkflowResponse(meta *workloadsApi.SupersetWorkload, spec *workloadsApi.WorkflowSpecSpec) workloadsApi.Workflow {
	return workloadsApi.Workflow{
		Name:          meta.Name,
		RequestedName: meta.RequestedName,
		WorkloadId:    meta.WorkloadId,
		ProjectId:     meta.ProjectId,
		DepartmentId:  meta.DepartmentId,
		ClusterId:     meta.ClusterId,
		CreatedBy:     meta.CreatedBy,
		CreatedAt:     meta.CreatedAt,
		DeletedAt:     meta.DeletedAt,
		DesiredPhase:  meta.DesiredPhase,
		ActualPhase:   meta.ActualPhase,
		Spec:          spec,
	}
}
