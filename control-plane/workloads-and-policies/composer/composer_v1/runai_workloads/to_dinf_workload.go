/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runai_workloads

import (
	"context"
	"net/http"

	"github.com/run-ai/runai/control-plane/workloads-and-policies/composer"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/composer/notifications"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/itemized"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	trackerApi "github.com/run-ai/runai/runai-clients/k8s-objects-tracker-client/generated"
	workloadsApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	dp "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferencepolicy/v1alpha1"
	dinfwv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferenceworkload/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
)

type DinfWorkloadComposer struct {
	// we have one composer for each spec. they share the same fields composer (fc)
	fc                *composer.FieldsComposer
	workerJobComposer JobSpecComposer
	leaderJobComposer JobSpecComposer

	// we also have a notifications composer, with its own fields composer
	NotificationsComposer *notifications.NotificationsComposer
}

func NewDinfWorkloadComposer(
	workloadName string,
	requester string,
	notificationsComposer *notifications.NotificationsComposer,
	options composer.FieldsComposerOptions,
) *DinfWorkloadComposer {
	fc := composer.FieldsComposer{
		WorkloadType:      string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE),
		WorkloadName:      workloadName,
		SubmittedBy:       requester,
		Options:           options,
		ItemizedNumbering: make(itemized.ItemizedNumbering),
		OauthPortNumber:   composer.Oauth2ContainerDefaultPort,
	}

	result := DinfWorkloadComposer{
		fc: &fc,

		workerJobComposer: JobSpecComposer{
			fc:          &fc,
			ReplicaType: convutils.Ref(workloadsApi.REPLICATYPE_WORKER),
		},

		leaderJobComposer: JobSpecComposer{
			fc:          &fc,
			ReplicaType: convutils.Ref(workloadsApi.REPLICATYPE_LEADER),
		},

		NotificationsComposer: notificationsComposer,
	}

	return &result
}

func (dwc *DinfWorkloadComposer) WithStorageClassConfiguration(config map[string]trackerApi.StorageClassConfiguration) *DinfWorkloadComposer {
	dwc.fc.WithStorageClassConfiguration(config)
	return dwc
}

func (dwc *DinfWorkloadComposer) Compose(ctx context.Context,
	src *workloadsApi.SupersetSpec, target *dinfwv1.DistributedInferenceWorkload,
) *workloadsApi.Error {
	if src == nil {
		return nil
	}

	if !convutils.HasValue(src.Leader) {
		return workloadsApi.NewError(http.StatusBadRequest, "leader spec is required for distributed inference workload")
	}

	err := dwc.ComposeDinfSpecFields(ctx, src, &target.Spec)()
	if err != nil {
		return err
	}

	var leaderAsSuperset workloadsApi.SupersetSpec
	var leaderJobSpec jobs.JobSpec

	copyErr := convutils.CopySpecs(&leaderAsSuperset, src.Leader.Get())
	if copyErr != nil {
		return workloadsApi.NewError(http.StatusInternalServerError, copyErr.Error())
	}

	// we need the category in both leader spec and worker spec, so copy it from the spec
	leaderAsSuperset.Category.Set(src.Category.Get())
	leaderAsSuperset.ServingPort.Set(src.ServingPort.Get())

	err = dwc.leaderJobComposer.Compose(ctx, &leaderAsSuperset, &leaderJobSpec)()
	if err != nil {
		return workloadsApi.NewError(err.Code, "Leader spec: "+err.Message)
	} else {
		copyErr := convutils.CopySpecs(&target.LeaderSpec, leaderJobSpec)
		if copyErr != nil {
			return workloadsApi.NewError(http.StatusInternalServerError, copyErr.Error())
		}
	}

	if dwc.NotificationsComposer != nil {
		leaderAsSuperset.NodePools = src.NodePools
		leaderAsSuperset.Workers.Set(src.Workers.Get())
		leaderAsSuperset.Replicas.Set(src.Replicas.Get())

		err := dwc.NotificationsComposer.Compose(ctx, dwc.leaderJobComposer.ReplicaType, nil, &leaderAsSuperset)()
		if err != nil {
			return err
		}
	}

	if convutils.HasValue(src.Worker) {
		var workerJobSpec jobs.JobSpec
		var workerAsSuperset workloadsApi.SupersetSpec

		copyErr := convutils.CopySpecs(&workerAsSuperset, src.Worker.Get())
		if copyErr != nil {
			return workloadsApi.NewError(http.StatusInternalServerError, copyErr.Error())
		}

		// copy elements from the common spec to the worker spec
		workerAsSuperset.Category.Set(src.Category.Get())

		err = dwc.workerJobComposer.Compose(ctx, &workerAsSuperset, &workerJobSpec)()
		if err != nil {
			return workloadsApi.NewError(err.Code, "Worker spec: "+err.Message)
		} else {
			target.WorkerSpec = &dp.DistributedInferenceEntitySpec{}

			copyErr := convutils.CopySpecs(target.WorkerSpec, workerJobSpec)
			if copyErr != nil {
				return workloadsApi.NewError(http.StatusInternalServerError, copyErr.Error())
			}
		}

		if dwc.NotificationsComposer != nil {
			err = dwc.NotificationsComposer.Compose(ctx, dwc.workerJobComposer.ReplicaType, nil, &workerAsSuperset)()
			if err != nil {
				return err
			}
		}
	}

	//
	// validation that server port is given
	//
	if dwc.NotificationsComposer != nil && dwc.NotificationsComposer.Output != nil {
		foundServingPort := false

		for _, env := range dwc.NotificationsComposer.Output.Environments {
			for _, conn := range env.Connections {
				if conn.ConnectionType == composer.PortServiceType_SERVING_PORT {
					foundServingPort = true
					break
				}
			}
		}

		if !foundServingPort {
			return workloadsApi.NewError(http.StatusBadRequest, "servingPort is required for distributed inference workloads")
		}
	}

	return nil
}

func (dwc *DinfWorkloadComposer) ComposeDinfSpecFields(ctx context.Context, src *workloadsApi.SupersetSpec,
	target *dp.DistributedInferenceSpec,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src == nil {
			return nil
		}

		var portAuthInfo composer.PortAuthInfo

		runner := &composer.StepRunner{}

		runner.Batch(
			dwc.fc.CopyString(convutils.Ref(dwc.fc.SubmittedBy), &target.UserName, stringToStringField, nil),
			dwc.ComposeNumWorkers(src.Workers.Get(), src.Worker.Get(), &target.Workers, intToIntField, nil),
			dwc.fc.CopyInt(src.Replicas.Get(), &target.Replicas, intToIntField, nil),
			dwc.fc.ComposeRestartPolicy(validRestartPoliciesDinf, src.RestartPolicy.Get(), &target.RestartPolicy, stringToStringField, nil),
			dwc.fc.CopyString(composer.EnumToString(src.StartupPolicy.Get()), &target.StartupPolicy, nonEmptyStringToStringField, nil),
			dwc.fc.CopyArray(src.NodePools, &target.NodePools, arrayToStringField, nil),
			dwc.fc.ComposePriorityClass(ctx, src.PriorityClass.Get(), &target.PriorityClassName, stringToStringField, nil, nil, nil),
			dwc.fc.ComposePreemptibility(ctx, src.Preemptibility.Get(), &target.Preemptibility, stringToStringField, nil),
			dwc.fc.ComposeServingPort(nil, src.ServingPort.Get(), &target.ServingPorts, servingPortToServingPortField, nil, &portAuthInfo),
		)

		return runner.Error()
	}
}

func (dwc *DinfWorkloadComposer) ComposeNumWorkers(src *int32, workerSpec *workloadsApi.DistributedInferenceLeaderWorkerSpec, target any,
	setter composer.FnIntSetter[int32], cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if !dwc.fc.Options.IgnoreMissingValues {
			numWorkers := convutils.Deref(src, 0)
			if numWorkers == 0 {
				if workerSpec != nil {
					return workloadsApi.NewError(http.StatusBadRequest, "worker spec requires workers field to be set to a positive value")
				}
			} else {
				if workerSpec == nil {
					return workloadsApi.NewError(http.StatusBadRequest, "worker spec is required when workers is set to a positive value")
				}
			}
		}

		err := setter(src, target, cookie)
		if err != nil {
			return err
		}

		return nil
	}
}

func (dwc *DinfWorkloadComposer) Requirements() composer.WorkloadRequirements {
	return dwc.workerJobComposer.WorkloadRequirements.Or(dwc.leaderJobComposer.WorkloadRequirements)
}
