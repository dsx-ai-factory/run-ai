/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"context"
	"fmt"
	"net/http"

	clusterApi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer/composer_v2"
	dinfToLWS "github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer/composer_v2/dinf_to_lws"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/clusterservice"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models_crud"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils/gojsonq"
	workloadsApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	workloadMgrApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
)

// AnyWorkloadSubmitterDinf handles submission decisions for Distributed Inference workloads (LeaderWorkerSet).
type AnyWorkloadSubmitterDinf struct{}

// CanCreate checks if this submitter can handle workload creation.
func (s *AnyWorkloadSubmitterDinf) DetermineSubmissionMethods(ctx context.Context, wc *workloadContext) SubmissionMethods {
	return determineSubmissionMethods(ctx, wc.tenantId, wc.workloadType, wc.clusterCapabilities)
}

// CanUpdate checks if this submitter can handle workload updates for a specific workload.
// For distributed inference workloads, checks workload-manager ownership and validates the LWS GVK.
// Returns the workload from workload-manager so the Update function can use it.
func (s *AnyWorkloadSubmitterDinf) CanUpdate(ctx context.Context, origSubmissionMethod SubmissionMethods, wc *workloadContext, _ string) *workloadsApi.SubmissionErrorV2 {
	return canUpdateForSubmissionMethod(ctx, origSubmissionMethod, wc)
}

func (s *AnyWorkloadSubmitterDinf) Create(ctx context.Context, wc *workloadContext) (*Submission, *workloadsApi.SubmissionErrorV2) {
	if wc == nil || wc.create == nil || wc.create.createParams == nil {
		return nil, workloadsApi.NewSubmissionErrorV2(http.StatusInternalServerError, "distributed inference create context is incomplete")
	}

	spec := wc.distributedInferenceSpec
	if spec == nil {
		var err error
		spec, err = projectDistributedInferenceSpec(wc.spec)
		if err != nil {
			return nil, workloadsApi.NewSubmissionErrorV2(http.StatusInternalServerError, fmt.Sprintf("failed to project distributed inference spec: %v", err))
		}
	}

	options := wc.options
	if err := composer_v2.SetClusterDependentOptions(ctx, &options, wc.clusterCapabilities); err != nil {
		return nil, workloadsApi.NewSubmissionErrorV2(http.StatusServiceUnavailable, err.Error())
	}

	params := wc.create.createParams
	composer := dinfToLWS.NewDistributedInferenceToLWSComposer(
		wc.workloadName,
		wc.create.project.Name,
		wc.create.namespace,
		wc.createdBy,
		wc.create.assetsMapping,
		wc.create.assetsMapping,
		options,
		wc.create.actualNodePools,
	).WithWorkloadID(params.Response.WorkloadId)

	clusterStorageConfig, configErr := buildClusterStorageConfig(ctx, params.Response.ClusterId)
	if configErr != nil {
		return nil, configErr
	}
	composer.WithStorageClassConfiguration(wc.storageClassConfigs).WithClusterStorageConfig(clusterStorageConfig)

	target := dinfToLWS.NewMinimalLeaderWorkerSetTarget()
	if composeErr := composer.Compose(ctx, spec, target)(); composeErr != nil {
		return nil, workloadsApi.NewSubmissionErrorV2(composeErr.Code, composeErr.Message)
	}
	manifest, err := gojsonq.LoadJson(target.LeaderWorkerSet)
	if err != nil {
		return nil, workloadsApi.NewSubmissionErrorV2(http.StatusInternalServerError, err.Error())
	}
	if metadataErr := composeDinfConnectionMetadata(ctx, wc, spec); metadataErr != nil {
		return nil, metadataErr
	}

	return &Submission{
		Request: workloadMgrApi.WorkloadV2CreateRequest{
			Metadata: buildDinfWorkloadV2Metadata(&params.Response.Name, &params.Response.ProjectId, &target.Config, spec),
			Manifest: manifest,
		},
		Requirements:             composer.WorkloadRequirements,
		ConnectionMetadata:       newWorkloadsServiceConnectionMetadata(wc.create.notificationsComposer.Output),
		ConnectionCreationFields: wc.create.notificationsComposer.Output,
	}, nil
}

func composeDinfConnectionMetadata(
	ctx context.Context,
	wc *workloadContext,
	spec *workloadsApi.DistributedInferenceSpecSpec,
) *workloadsApi.SubmissionErrorV2 {
	if wc.create.notificationsComposer == nil {
		return workloadsApi.NewSubmissionErrorV2(http.StatusInternalServerError, "workload notification composer was not initialized")
	}

	leaderSpec, ok := spec.GetLeaderOk()
	if !ok || leaderSpec == nil {
		return workloadsApi.NewSubmissionErrorV2(http.StatusBadRequest, "leader spec is required for distributed inference workload")
	}

	var leaderAsSuperset workloadsApi.SupersetSpec
	if err := convutils.CopySpecs(&leaderAsSuperset, leaderSpec); err != nil {
		return workloadsApi.NewSubmissionErrorV2(http.StatusInternalServerError, err.Error())
	}
	leaderAsSuperset.Category.Set(spec.Category.Get())
	if servingPort := spec.ServingPort.Get(); servingPort != nil {
		var servingPortAsSuperset workloadsApi.SupersetServingPort
		if err := convutils.CopySpecs(&servingPortAsSuperset, servingPort); err != nil {
			return workloadsApi.NewSubmissionErrorV2(http.StatusInternalServerError, err.Error())
		}
		leaderAsSuperset.ServingPort.Set(&servingPortAsSuperset)
	}
	leaderAsSuperset.NodePools = spec.NodePools
	leaderAsSuperset.Workers.Set(spec.Workers.Get())
	leaderAsSuperset.Replicas.Set(spec.Replicas.Get())

	leaderReplicaType := workloadsApi.REPLICATYPE_LEADER
	if composerErr := wc.create.notificationsComposer.Compose(ctx, &leaderReplicaType, nil, &leaderAsSuperset)(); composerErr != nil {
		return workloadsApi.NewSubmissionErrorV2(composerErr.Code, composerErr.Message)
	}

	if spec.GetWorkers() == 0 {
		return nil
	}

	workerSpec, ok := spec.GetWorkerOk()
	if !ok || workerSpec == nil {
		// V2 LWS composition copies the leader pod template when workers are requested
		// without a distinct worker spec. Mirror that effective topology in metadata.
		workerSpec = leaderSpec
	}

	var workerAsSuperset workloadsApi.SupersetSpec
	if err := convutils.CopySpecs(&workerAsSuperset, workerSpec); err != nil {
		return workloadsApi.NewSubmissionErrorV2(http.StatusInternalServerError, err.Error())
	}
	workerAsSuperset.Category.Set(spec.Category.Get())

	workerReplicaType := workloadsApi.REPLICATYPE_WORKER
	if composerErr := wc.create.notificationsComposer.Compose(ctx, &workerReplicaType, nil, &workerAsSuperset)(); composerErr != nil {
		return workloadsApi.NewSubmissionErrorV2(composerErr.Code, composerErr.Message)
	}

	return nil
}

func (s *AnyWorkloadSubmitterDinf) Update(ctx context.Context, wc *workloadContext, workloadId string,
	existingWorkload *workloadMgrApi.WorkloadV2GetResponse, mergedWorkload *models_crud.Workload,
) (*Update, *workloadsApi.SubmissionErrorV2) {
	if mergedWorkload == nil || existingWorkload == nil {
		return nil, workloadsApi.NewSubmissionErrorV2(http.StatusInternalServerError, "distributed inference update context is incomplete")
	}

	spec, err := projectDistributedInferenceSpec(mergedWorkload.Spec)
	if err != nil {
		return nil, workloadsApi.NewSubmissionErrorV2(http.StatusInternalServerError, fmt.Sprintf("failed to project distributed inference spec: %v", err))
	}

	clusterCapabilities := wc.clusterCapabilities
	if clusterCapabilities == nil {
		clusterCapabilities = checker.NewClusterCapabilitiesChecker(mergedWorkload.ClusterId)
	}
	options := wc.options
	if err := composer_v2.SetClusterDependentOptions(ctx, &options, clusterCapabilities); err != nil {
		return nil, workloadsApi.NewSubmissionErrorV2(http.StatusServiceUnavailable, err.Error())
	}

	nodePoolsResponse, nodePoolsErr := clusterservice.Get().GetNodePoolsInternalByCluster(ctx, mergedWorkload.ClusterId)
	if nodePoolsErr != nil {
		return nil, workloadsApi.NewSubmissionErrorV2(http.StatusInternalServerError, nodePoolsErr.Error())
	}
	var actualNodePools []clusterApi.Nodepool
	if nodePoolsResponse != nil {
		actualNodePools = nodePoolsResponse.Nodepools
	}

	dinfComposer := dinfToLWS.NewDistributedInferenceToLWSComposer(
		mergedWorkload.Name,
		existingWorkload.Workload.Metadata.ProjectName,
		mergedWorkload.Namespace,
		mergedWorkload.CreatedBy,
		mergedWorkload.AssetsMapping,
		mergedWorkload.AssetsMapping,
		options,
		actualNodePools,
	)
	existingLWSName, ok := manifestName(existingWorkload.Workload.WorkloadManifest)
	if !ok {
		return nil, workloadsApi.NewSubmissionErrorV2(http.StatusInternalServerError, "existing LeaderWorkerSet manifest has no name")
	}
	dinfComposer.WithLeaderWorkerSetName(existingLWSName)
	clusterStorageConfig, configErr := buildClusterStorageConfig(ctx, mergedWorkload.ClusterId)
	if configErr != nil {
		return nil, configErr
	}
	dinfComposer.WithStorageClassConfiguration(wc.storageClassConfigs).WithClusterStorageConfig(clusterStorageConfig)

	target := dinfToLWS.NewMinimalLeaderWorkerSetTarget()
	if composeErr := dinfComposer.Compose(ctx, spec, target)(); composeErr != nil {
		return nil, workloadsApi.NewSubmissionErrorV2(composeErr.Code, composeErr.Message)
	}
	manifest, err := gojsonq.LoadJson(target.LeaderWorkerSet)
	if err != nil {
		return nil, workloadsApi.NewSubmissionErrorV2(http.StatusInternalServerError, err.Error())
	}

	return &Update{
		Request: workloadMgrApi.WorkloadV2UpdateRequest{
			Metadata: &workloadMgrApi.WorkloadV2MetadataUpdateParams{
				Configuration: *workloadMgrApi.NewNullableWorkloadV2Configuration(&target.Config),
			},
			Manifest: manifest,
		},
		Requirements: dinfComposer.WorkloadRequirements,
	}, nil
}

func manifestName(manifest map[string]interface{}) (string, bool) {
	metadata, ok := manifest["metadata"].(map[string]interface{})
	if !ok {
		return "", false
	}
	name, ok := metadata["name"].(string)
	return name, ok && name != ""
}

func projectDistributedInferenceSpec(src *workloadsApi.SupersetSpec) (*workloadsApi.DistributedInferenceSpecSpec, error) {
	target := &workloadsApi.DistributedInferenceSpecSpec{}
	if err := convutils.CopySpecs(target, src); err != nil {
		return nil, err
	}
	return target, nil
}

func buildDinfWorkloadV2Metadata(
	name, projectID *string,
	configuration *workloadMgrApi.WorkloadV2Configuration,
	spec *workloadsApi.DistributedInferenceSpecSpec,
) workloadMgrApi.WorkloadV2MetadataCreateParams {
	metadata := workloadMgrApi.WorkloadV2MetadataCreateParams{
		Name:          name,
		ProjectId:     projectID,
		Configuration: configuration,
	}
	if spec == nil {
		return metadata
	}
	if category := spec.Category.Get(); category != nil {
		metadata.Category.Set(category)
	}
	if priority := spec.PriorityClass.Get(); priority != nil {
		metadata.Priority.Set(priority)
	}
	if preemptibility := spec.Preemptibility.Get(); preemptibility != nil {
		converted := workloadMgrApi.Preemptibility(*preemptibility)
		metadata.Preemptibility.Set(&converted)
	}
	return metadata
}
