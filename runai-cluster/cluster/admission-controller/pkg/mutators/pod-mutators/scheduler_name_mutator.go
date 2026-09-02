/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pod_mutators

import (
	"k8s.io/api/admission/v1beta1"
	v1 "k8s.io/api/core/v1"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/admission-controller/pkg/types"
)

func (pm *PodMutator) mutateSchedulerName(pod *v1.Pod, _ *v1beta1.AdmissionRequest) error {
	pod.Spec.SchedulerName = types.RunaiScheduler
	return nil
}
