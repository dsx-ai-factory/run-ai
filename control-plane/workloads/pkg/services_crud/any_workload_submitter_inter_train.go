/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"context"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer/composer_v2"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer/composer_v2/iw_tw_workloads"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models_crud"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils/gojsonq"
	workloadsApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	workloadMgrApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client/generated"
)

// AnyWorkloadSubmitterInterTrain handles submission decisions for interactive and training workloads.
// It composes workloads into a BatchJob target and decides between workload-runner and native submission.
type AnyWorkloadSubmitterInterTrain struct{}

// CanCreate checks if this submitter can handle workload creation.
// For interactive/training workloads, this is controlled by the feature flag.
func (s *AnyWorkloadSubmitterInterTrain) DetermineSubmissionMethods(ctx context.Context, wc *workloadContext) SubmissionMethods {
	return determineSubmissionMethods(ctx, wc.tenantId, wc.workloadType, wc.clusterCapabilities)
}

// CanUpdate checks if this submitter can handle workload updates for a specific workload.
// Interactive/training workloads do not support updates.
func (s *AnyWorkloadSubmitterInterTrain) CanUpdate(_ context.Context, _ SubmissionMethods, _ *workloadContext, _ string) *workloadsApi.SubmissionErrorV2 {
	return nil
}

// Create handles workload creation for interactive/training workloads.
func (s *AnyWorkloadSubmitterInterTrain) Create(ctx context.Context, wc *workloadContext) (*Submission, *workloadsApi.SubmissionErrorV2) {
	projectName := wc.create.project.Name
	namespace := wc.create.namespace

	options := wc.options
	if err := composer_v2.SetClusterDependentOptions(ctx, &options, wc.clusterCapabilities); err != nil {
		return nil, workloadsApi.NewSubmissionErrorV2(http.StatusServiceUnavailable, err.Error())
	}
	// v1 parity: workspace/training workloads delete the PVCs created for them on terminal phase.
	options.CleanupCreatedPvcsOnCompletion = convutils.Ref(true)

	itTwComposer := iw_tw_workloads.NewInterTrainJobComposer(
		wc.workloadType,
		wc.workloadName,
		projectName,
		namespace,
		wc.createdBy,
		wc.create.assetsMapping,
		options,
		wc.create.actualNodePools,
	)

	clusterStorageConfig, cfgErr := buildClusterStorageConfig(ctx, wc.create.createParams.Response.ClusterId)
	if cfgErr != nil {
		return nil, cfgErr
	}

	itTwComposer.WithStorageClassConfiguration(wc.storageClassConfigs)
	itTwComposer.WithDataVolumeClaimNames(wc.dataVolumeClaimNames)
	itTwComposer.WithClusterStorageConfig(clusterStorageConfig)

	target := iw_tw_workloads.NewMinimalBatchJobTarget()

	if composerErr := itTwComposer.Compose(ctx, wc.spec, target)(); composerErr != nil {
		return nil, workloadsApi.NewSubmissionErrorV2(composerErr.Code, composerErr.Message)
	}

	manifest, err := gojsonq.LoadJson(target.Job)
	if err != nil {
		return nil, workloadsApi.NewSubmissionErrorV2(http.StatusInternalServerError, err.Error())
	}

	if wc.create.notificationsComposer == nil {
		return nil, workloadsApi.NewSubmissionErrorV2(http.StatusInternalServerError, "workload notification composer was not initialized")
	}

	if composerErr := wc.create.notificationsComposer.
		WithConnectionsPreProcessor(projectName, composer.NetworkingResourcePrefixJob).
		Compose(ctx, nil, wc.create.complimentaryAssetInfo, wc.spec)(); composerErr != nil {
		return nil, workloadsApi.NewSubmissionErrorV2(composerErr.Code, composerErr.Message)
	}

	return &Submission{
		Request: workloadMgrApi.WorkloadV2CreateRequest{
			Metadata: buildInterTrainWorkloadV2Metadata(
				&wc.create.createParams.Response.Name,
				&wc.create.createParams.Response.ProjectId,
				&target.Config,
				wc.spec,
				target,
			),
			Manifest: manifest,
		},
		Requirements:             itTwComposer.WorkloadRequirements,
		ConnectionMetadata:       newWorkloadsServiceConnectionMetadata(wc.create.notificationsComposer.Output),
		ConnectionCreationFields: wc.create.notificationsComposer.Output,
	}, nil
}

func (s *AnyWorkloadSubmitterInterTrain) Update(ctx context.Context, wc *workloadContext, workloadId string,
	existingWorkload *workloadMgrApi.WorkloadV2GetResponse, mergedWorkload *models_crud.Workload,
) (*Update, *workloadsApi.SubmissionErrorV2) {
	return nil, nil
}

func buildInterTrainWorkloadV2Metadata(name, projectId *string, configuration *workloadMgrApi.WorkloadV2Configuration,
	spec *workloadsApi.SupersetSpec, target *iw_tw_workloads.BatchJobTarget,
) workloadMgrApi.WorkloadV2MetadataCreateParams {
	metadata := workloadMgrApi.WorkloadV2MetadataCreateParams{
		Name:          name,
		ProjectId:     projectId,
		Configuration: configuration,
	}
	if spec != nil {
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
	}
	if target == nil {
		return metadata
	}

	if !metadata.Category.IsSet() {
		if category := target.Job.Labels[composer.WorkloadCategoryLabel]; category != "" {
			metadata.Category.Set(&category)
		}
	}
	if !metadata.Priority.IsSet() {
		if priorityClass := target.Job.Spec.Template.Spec.PriorityClassName; priorityClass != "" {
			metadata.Priority.Set(&priorityClass)
		}
	}
	if !metadata.Preemptibility.IsSet() {
		if preemptibility := target.Job.Spec.Template.Labels[composer.LabelPreemptibilityName]; preemptibility != "" {
			convertedPreemptibility := workloadMgrApi.Preemptibility(preemptibility)
			metadata.Preemptibility.Set(&convertedPreemptibility)
		}
	}

	return metadata
}
