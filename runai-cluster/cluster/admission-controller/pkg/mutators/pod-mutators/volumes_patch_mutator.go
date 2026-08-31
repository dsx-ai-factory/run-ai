/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pod_mutators

import (
	"encoding/json"
	"log"

	"k8s.io/api/admission/v1beta1"
	v1 "k8s.io/api/core/v1"

	"github.com/run-ai/runai/runai-cluster/cluster/admission-controller/pkg/patch"
	"github.com/run-ai/runai/runai-cluster/cluster/admission-controller/pkg/utils"
)

const (
	volumesPatchAnnotationKey      = "run.ai/patch-volumes"
	volumeMountsPatchAnnotationKey = "run.ai/patch-volume-mounts"
)

func (pm *PodMutator) patchVolumesAndMounts(pod *v1.Pod, _ *v1beta1.AdmissionRequest) error {
	volumesAnnotationValue, ok := pod.Annotations[volumesPatchAnnotationKey]
	if !ok {
		return nil
	}
	log.Println("Found volume patch annotation")

	mountsAnnotationValue, ok := pod.Annotations[volumeMountsPatchAnnotationKey]
	if !ok {
		return nil
	}
	log.Println("Found volume mount patch annotation")

	volumes, err := parseVolumesAnnotation(volumesAnnotationValue)
	if err != nil {
		log.Printf("Failed to parse volume patch annotation: %v\n", err)
		return err
	}

	volumeMounts, err := parseVolumeMountsAnnotation(mountsAnnotationValue)
	if err != nil {
		log.Printf("Failed to parse volume mount patch annotation: %v\n", err)
		return err
	}

	containerIndex := utils.GetRunaiContainerIndex(pod.Labels, pod.Spec.Containers)
	container := &pod.Spec.Containers[containerIndex]

	for _, volume := range volumes {
		patch.AddVolumeToPodSpec(&pod.Spec, volume)
	}
	for _, volumeMount := range volumeMounts {
		patch.AddVolumeMountToContainer(container, volumeMount.Name, volumeMount.MountPath,
			volumeMount.SubPath, volumeMount.ReadOnly)
	}

	return nil
}

func parseVolumesAnnotation(annotationValue string) ([]v1.Volume, error) {
	var volumes []v1.Volume
	err := json.Unmarshal([]byte(annotationValue), &volumes)
	if err != nil {
		return nil, err
	}
	return volumes, nil
}

func parseVolumeMountsAnnotation(annotationValue string) ([]v1.VolumeMount, error) {
	var volumeMounts []v1.VolumeMount
	err := json.Unmarshal([]byte(annotationValue), &volumeMounts)
	if err != nil {
		return nil, err
	}

	return volumeMounts, nil
}
