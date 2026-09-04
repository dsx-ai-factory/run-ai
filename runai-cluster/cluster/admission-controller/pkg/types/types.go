/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package types

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	RunaiScheduler              = "runai-scheduler"
	RunaiNamespace              = "runai"
	RunaiGpuFraction            = "gpu-fraction"
	RunaiGpuMemory              = "gpu-memory"
	RunaiGpuLimit               = "runai-gpu-limit"
	RunaiGpuPriority            = "runai-gpu-priority"
	GpuResource                 = "nvidia.com/gpu"
	MpsAnnotation               = "mps"
	RunaiOpportunistic          = "runai-opportunistic"
	RunaiGpuFracEnvVar          = "RUNAI_GPU_FRACTION"
	GpuComputeRequestEnvVarName = "RUNAI_GPU_COMPUTE_REQUEST"
	GpuComputeLimitEnvVarName   = "RUNAI_GPU_COMPUTE_LIMIT"
)

// PodControllerSpec defines the desired state of Notebook
type PodControllerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Template PodControllerTemplateSpec `json:"template,omitempty"`
	Replicas int64                     `json:"replicas,omitempty"`
}

type PodControllerTemplateSpec struct {
	Spec              corev1.PodSpec `json:"spec,omitempty"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
}

type PodControllerCommon struct {
	Spec              PodControllerSpec `json:"spec,omitempty"`
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
}

type MutatingParams struct {
	MinAllocationEnforcementBytes  int
	SharedDirectoryHostPath        string
	MpsDirectoryHostPath           string
	CoreDebugLogsDirectoryHostPath string
	CoreDebugLevel                 string
	CoreDebugFile                  string
	CoreAnalytics                  bool
	IsKubevirtEnabledCluster       bool
	ScheduleWorkflowPods           bool
	SpecificClusterType            string
	OverProvisioningRoute          bool
}

func (pcc *PodControllerCommon) DeepCopy() *PodControllerCommon {
	return &PodControllerCommon{
		Spec: *pcc.Spec.DeepCopy(),
		TypeMeta: metav1.TypeMeta{
			Kind:       pcc.TypeMeta.Kind,
			APIVersion: pcc.TypeMeta.APIVersion,
		},
		ObjectMeta: *pcc.ObjectMeta.DeepCopy(),
	}
}

func (pcs *PodControllerSpec) DeepCopy() *PodControllerSpec {
	return &PodControllerSpec{
		Template: *pcs.Template.DeepCopy(),
		Replicas: pcs.Replicas,
	}
}

func (pcts *PodControllerTemplateSpec) DeepCopy() *PodControllerTemplateSpec {
	return &PodControllerTemplateSpec{
		Spec:       *pcts.Spec.DeepCopy(),
		ObjectMeta: *pcts.ObjectMeta.DeepCopy(),
	}
}
