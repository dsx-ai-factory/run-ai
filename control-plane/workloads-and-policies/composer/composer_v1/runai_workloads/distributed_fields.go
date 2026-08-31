/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runai_workloads

import (
	"fmt"
	"net/http"

	"github.com/run-ai/runai/control-plane/workloads-and-policies/composer"
	workloadsApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

// DistributedFieldsComposer takes care for distributed-specific fields.
// it is created and run per replica type, and so it stores the replica type as data member.
// This composer is used by two other composers:
// 1) DistributedWorkloadComposer - for composing the distributed specific fields of distributed workload
// 2) NotificationsComposer - for composing the workloads creation related distributed fields (framework type, for example)
type DistributedFieldsComposer struct {
	fc          *composer.FieldsComposer
	framework   *workloadsApi.DistributedFramework
	replicaType *workloadsApi.ReplicaType
}

func NewDistributedFieldsComposer(fc *composer.FieldsComposer, framework *workloadsApi.DistributedFramework, replicaType *workloadsApi.ReplicaType) *DistributedFieldsComposer {
	return &DistributedFieldsComposer{
		fc:          fc,
		replicaType: replicaType,
		framework:   framework,
	}
}

// ComposeJobType works out the JobType field, which holds a value that is synonymous to the framework type.
func (df *DistributedFieldsComposer) ComposeJobType(src *string,
	target any, setter composer.FnStringSetter, cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if !composer.IsMainNode(df.replicaType) {
			return nil
		}

		if src == nil {
			if !df.fc.Options.IgnoreMissingValues {
				return workloadsApi.NewError(http.StatusBadRequest, "mandatory distributedFramework is not provided")
			}
			return nil
		}

		frameworkWithJobSuffix := fmt.Sprintf("%sJob", *src)

		return setter(&frameworkWithJobSuffix, target, cookie)
	}
}

func (df *DistributedFieldsComposer) ComposeCleanPodPolicy(src *string,
	target any, setter composer.FnStringSetter, cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src == nil || !composer.IsMainNode(df.replicaType) {
			return nil
		}

		return setter(src, target, cookie)
	}
}

func (df *DistributedFieldsComposer) ComposeMpiSlotsPerWorker(src *int32,
	target any, setter composer.FnIntSetter[int32], cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src == nil || !composer.IsMainNode(df.replicaType) {
			return nil
		}

		err := df.frameworkSpecific(src, "slotsPerWorker", workloadsApi.DISTRIBUTEDFRAMEWORK_MPI)
		if err != nil {
			return err
		}

		return setter(src, target, cookie)
	}
}

func (df *DistributedFieldsComposer) ComposeMpiLauncherCreationPolicy(src *workloadsApi.MpiLauncherCreationPolicy,
	target any, setter composer.FnStringSetter, cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src == nil || !composer.IsMainNode(df.replicaType) {
			return nil
		}

		err := df.frameworkSpecific(src, "mpiLauncherCreationPolicy", workloadsApi.DISTRIBUTEDFRAMEWORK_MPI)
		if err != nil {
			return err
		}

		if !df.fc.Options.LauncherCreationPolicySupported {
			return workloadsApi.NewError(http.StatusBadRequest, "mpiLauncherCreationPolicy is not supported in this cluster")
		}

		return setter(composer.EnumToString(src), target, cookie)
	}
}

func (df *DistributedFieldsComposer) ComposeMpiSshAuthMountPath(src *string,
	target any, setter composer.FnStringSetter, cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src == nil || !composer.IsMainNode(df.replicaType) {
			return nil
		}

		err := df.frameworkSpecific(src, "sshAuthMountPath", workloadsApi.DISTRIBUTEDFRAMEWORK_MPI)
		if err != nil {
			return err
		}

		return setter(src, target, cookie)
	}
}

func (df *DistributedFieldsComposer) ComposePytorchMinReplicas(src *int32,
	target any, setter composer.FnIntSetter[int32], cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src == nil || !composer.IsMainNode(df.replicaType) {
			return nil
		}

		err := df.frameworkSpecific(src, "minReplicas", workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH)
		if err != nil {
			return err
		}

		return setter(src, target, cookie)
	}
}

func (df *DistributedFieldsComposer) ComposePytorchMaxReplicas(src *int32, minReplicas *int32,
	target any, setter composer.FnIntSetter[int32], cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src == nil || !composer.IsMainNode(df.replicaType) {
			return nil
		}

		if minReplicas != nil && *src < *minReplicas {
			return workloadsApi.NewError(http.StatusBadRequest, "maxReplicas should be no less than minReplicas")
		}

		err := df.frameworkSpecific(src, "maxReplicas", workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH)
		if err != nil {
			return err
		}

		return setter(src, target, cookie)
	}
}

func (df *DistributedFieldsComposer) ComposeNumWorkers(src *int32,
	target any, setter composer.FnIntSetter[int32], cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if !composer.IsMainNode(df.replicaType) {
			return nil
		}

		return setter(src, target, cookie)
	}
}

func (df *DistributedFieldsComposer) ComposeBackoffLimit(src *int64,
	target any, setter composer.FnIntSetter[int64], cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src == nil || !composer.IsMainNode(df.replicaType) {
			return nil
		}

		return setter(src, target, cookie)
	}
}

func (df *DistributedFieldsComposer) frameworkSpecific(src any, field string, expectedFramework workloadsApi.DistributedFramework) *workloadsApi.Error {
	if df.framework == nil || *df.framework == expectedFramework {
		return nil
	}

	if src != nil {
		return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("%s is not supported by %s framework", field, *df.framework))
	}

	return nil
}
