/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pod_mutators

import (
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/admission-controller/pkg/patch"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/admission-controller/pkg/types"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/admission-controller/pkg/utils"

	"k8s.io/api/admission/v1beta1"
	v1 "k8s.io/api/core/v1"
)

func (pm *PodMutator) setEnvVariablesForPod(pod *v1.Pod, _ *v1beta1.AdmissionRequest) error {
	containerIndex := utils.GetRunaiContainerIndex(pod.Labels, pod.Spec.Containers)
	container := &pod.Spec.Containers[containerIndex]

	if isRunaiPodJob(pod) {
		injectJobEnvVarsFromPod(pod, container)
	} else {
		injectJobEnvVarsFromPodOwner(pod, container)
	}
	injectGpuNumEnvironmentVariable(container)
	injectEnvVars(container)
	return nil
}

func injectJobEnvVarsFromPod(pod *v1.Pod, container *v1.Container) {
	patch.AddEnvVarToContainer(container, v1.EnvVar{Name: "jobUUID",
		ValueFrom: &v1.EnvVarSource{
			FieldRef: &v1.ObjectFieldSelector{
				FieldPath: "metadata.uid"}}}) // for backward compatibility
	patch.AddEnvVarToContainer(container, v1.EnvVar{Name: "JOB_UUID",
		ValueFrom: &v1.EnvVarSource{
			FieldRef: &v1.ObjectFieldSelector{
				FieldPath: "metadata.uid"}}})

	patch.AddEnvVarToContainer(container, v1.EnvVar{Name: "jobName", Value: pod.Name}) // for backward compatibility
	patch.AddEnvVarToContainer(container, v1.EnvVar{Name: "JOB_NAME", Value: pod.Name})
}

func injectJobEnvVarsFromPodOwner(pod *v1.Pod, container *v1.Container) {
	jobId := fmt.Sprintf("%v", pod.OwnerReferences[0].UID)
	jobName := pod.OwnerReferences[0].Name
	patch.AddEnvVarToContainer(container, v1.EnvVar{Name: "jobUUID", Value: jobId}) // for backward compatibility
	patch.AddEnvVarToContainer(container, v1.EnvVar{Name: "JOB_UUID", Value: jobId})
	patch.AddEnvVarToContainer(container, v1.EnvVar{Name: "jobName", Value: jobName}) // for backward compatibility
	patch.AddEnvVarToContainer(container, v1.EnvVar{Name: "JOB_NAME", Value: jobName})
}

func injectGpuNumEnvironmentVariable(container *v1.Container) {
	if wholeRequestedGpus, found := container.Resources.Limits[types.GpuResource]; found {
		value := fmt.Sprintf("%v", wholeRequestedGpus.Value())
		patch.AddEnvVarToContainer(container, v1.EnvVar{Name: "RUNAI_NUM_OF_GPUS", Value: value})
	}
}

func injectEnvVars(container *v1.Container) {
	patch.AddEnvVarToContainer(container, v1.EnvVar{
		Name: "podUUID",
		ValueFrom: &v1.EnvVarSource{
			FieldRef: &v1.ObjectFieldSelector{
				FieldPath: "metadata.uid"}}})
	patch.AddEnvVarToContainer(container, v1.EnvVar{
		Name: "POD_UUID",
		ValueFrom: &v1.EnvVarSource{
			FieldRef: &v1.ObjectFieldSelector{
				FieldPath: "metadata.uid"}}})
	patch.AddEnvVarToContainer(container, v1.EnvVar{
		Name: "NODE_NAME",
		ValueFrom: &v1.EnvVarSource{
			FieldRef: &v1.ObjectFieldSelector{
				FieldPath: "spec.nodeName"}}})
}
