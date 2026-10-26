/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"context"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/gpucompat"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer/composer_v1/runai_workloads"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/dao_crud"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/services_integration"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	dinfwv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferenceworkload/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

// composeNative runs the native per-workload-type composer.
// Output of the composer is captured via wc.notificationsComposer; the only
// value returned to the caller is the workload requirements summary.
func composeNative(ctx context.Context, wc *workloadContext) (composer.WorkloadRequirements, *workloadsApi.SubmissionErrorV2) {
	if wc.spec != nil {
		if err := validateNativeGpuMode(wc.spec); err != nil {
			return composer.WorkloadRequirements{}, err
		}
	}
	if wc.masterSpec != nil {
		if err := validateNativeGpuMode(wc.masterSpec); err != nil {
			return composer.WorkloadRequirements{}, err
		}
	}
	switch wc.workloadType {
	case DistributedInference:
		return composeDistributedInferenceNative(ctx, wc)
	case Distributed:
		return composeDistributedNative(ctx, wc)
	case Inference:
		return composeInferenceNative(ctx, wc)
	default:
		return composeInteractiveTrainingNative(ctx, wc)
	}
}

func composeDistributedInferenceNative(ctx context.Context, wc *workloadContext) (composer.WorkloadRequirements, *workloadsApi.SubmissionErrorV2) {
	dInfComposer := runai_workloads.NewDinfWorkloadComposer(
		wc.workloadName,
		wc.createdBy,
		wc.create.notificationsComposer,
		wc.options)

	dInfComposer.WithStorageClassConfiguration(wc.storageClassConfigs)

	if composerErr := dInfComposer.Compose(ctx, gpucompat.LegacyWorkloadSpec(wc.spec), &dinfwv1.DistributedInferenceWorkload{}); composerErr != nil {
		return composer.WorkloadRequirements{}, workloadsApi.NewSubmissionErrorV2(composerErr.Code, composerErr.Message)
	}

	return dInfComposer.Requirements(), nil
}

func composeDistributedNative(ctx context.Context, wc *workloadContext) (composer.WorkloadRequirements, *workloadsApi.SubmissionErrorV2) {
	distributedComposer := runai_workloads.NewDistributedWorkloadComposer(
		wc.workloadName,
		wc.createdBy,
		wc.create.notificationsComposer,
		wc.options)

	distributedComposer.
		WithAssetMapping(wc.create.assetsMapping).
		WithComplimentaryAssetInfo(workloadsApi.REPLICATYPE_WORKER, wc.create.complimentaryAssetInfo)

	if convutils.HasValue(wc.create.createParams.Response.MasterSpec) {
		complimentaryAssetInfoMaster, err := composer.LoadComplimentaryAssetInfo(wc.create.createParams.Response.MasterSpec.Get().Annotations, true)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to load asset info of master spec: %v", err)
		}

		distributedComposer.WithComplimentaryAssetInfo(workloadsApi.REPLICATYPE_MASTER, complimentaryAssetInfoMaster)
	}

	distributedComposer.WithStorageClassConfiguration(wc.storageClassConfigs)

	distributedWorkload := workloadsApi.SupersetWorkload{
		Spec: gpucompat.LegacyWorkloadSpec(wc.spec),
	}
	if wc.masterSpec != nil {
		distributedWorkload.MasterSpec = *workloadsApi.NewNullableMasterSpec(gpucompat.LegacyMasterSpec(wc.masterSpec))
	}

	if composerErr := distributedComposer.Compose(ctx, &distributedWorkload, &jobs.JobWorkload{}); composerErr != nil {
		return composer.WorkloadRequirements{}, workloadsApi.NewSubmissionErrorV2(composerErr.Code, composerErr.Message)
	}

	return distributedComposer.Requirements(), nil
}

func composeInferenceNative(ctx context.Context, wc *workloadContext) (composer.WorkloadRequirements, *workloadsApi.SubmissionErrorV2) {
	jobSpecComposer := runai_workloads.NewSupersetToJobSpecComposer(
		wc.workloadType,
		wc.workloadName,
		wc.createdBy,
		nil,
		nil, // we do not send the assets mapping in control plane, only in agent
		wc.options)

	jobSpecComposer.WithStorageClassConfiguration(wc.storageClassConfigs)

	composerErr := jobSpecComposer.Compose(ctx, gpucompat.LegacyWorkloadSpec(wc.spec), &jobs.JobSpec{})()
	if composerErr == nil {
		composerErr = wc.create.notificationsComposer.Compose(ctx, nil, wc.create.complimentaryAssetInfo, wc.spec)()
	}
	if composerErr != nil {
		return composer.WorkloadRequirements{}, workloadsApi.NewSubmissionErrorV2(composerErr.Code, composerErr.Message)
	}

	return jobSpecComposer.WorkloadRequirements, nil
}

func composeInteractiveTrainingNative(ctx context.Context, wc *workloadContext) (composer.WorkloadRequirements, *workloadsApi.SubmissionErrorV2) {
	jobSpecComposer := runai_workloads.NewSupersetToJobSpecComposer(
		wc.workloadType,
		wc.workloadName,
		wc.createdBy,
		nil,
		nil, // we do not send the assets mapping in control plane, only in agent
		wc.options)

	jobSpecComposer.WithStorageClassConfiguration(wc.storageClassConfigs)

	if composerErr := jobSpecComposer.Compose(ctx, gpucompat.LegacyWorkloadSpec(wc.spec), &jobs.JobSpec{})(); composerErr != nil {
		return composer.WorkloadRequirements{}, workloadsApi.NewSubmissionErrorV2(composerErr.Code, composerErr.Message)
	}

	if composerErr := wc.create.notificationsComposer.Compose(ctx, nil, wc.create.complimentaryAssetInfo, wc.spec)(); composerErr != nil {
		return composer.WorkloadRequirements{}, workloadsApi.NewSubmissionErrorV2(composerErr.Code, composerErr.Message)
	}

	return jobSpecComposer.WorkloadRequirements, nil
}

// persistAndRespondNative writes the workload to the DB and notifies workloads-tracking,
// then updates the response with the actual phase if available.
func persistAndRespondNative(ctx context.Context, wc *workloadContext, assetsMapping []workloadsApi.AssetsMapping) *workloadsApi.SubmissionErrorV2 {
	wc.create.createParams.Response.DesiredPhase = workloadsApi.WORKLOADDESIREDPHASE_RUNNING

	workloadsRec := NewWorkloadFromWorkloadParams(wc.create.createParams, assetsMapping)

	if dbErr := dao_crud.DbUpsertWorkload(ctx, workloadsRec); dbErr != nil {
		return ToSubmissionError(dbErr)
	}

	wc.create.createParams.AcceptedAt = time.Now().UTC()

	if wc.create.notificationsComposer.Output == nil {
		return nil
	}

	workloadCreationFields, err := workloadsClientFieldToServerFields(*wc.create.notificationsComposer.Output)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to parse notification in a workload creation fields: %v", err)
		return nil
	}

	workloadCreationFields.ProjectName = wc.create.project.Name
	workloadCreationFields.DepartmentName = wc.create.project.Parent.Get().GetName()

	if notifyErr := services_integration.WorkloadCreated(ctx, workloadCreationFields); notifyErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to notify workloads-tracking about new workload: %v", notifyErr.Message)
	}

	actualPhase, phaseErr := services_integration.GetPhaseOfWorkload(ctx, &wc.create.createParams.Response.ClusterId, wc.create.createParams.Response.WorkloadId)
	if phaseErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get workload %s phase, err: %v", wc.create.createParams.Response.WorkloadId, phaseErr.Message)
		wc.create.createParams.Response.ActualPhase = convutils.Ref(workloadsApi.Phase(api.PHASE_CREATING))
	} else {
		wc.create.createParams.Response.ActualPhase = convutils.Ref(workloadsApi.Phase(actualPhase))
	}

	return nil
}
