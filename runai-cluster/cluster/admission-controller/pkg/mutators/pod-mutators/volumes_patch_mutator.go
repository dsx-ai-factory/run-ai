/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pod_mutators

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"

	"k8s.io/api/admission/v1beta1"
	v1 "k8s.io/api/core/v1"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/admission-controller/pkg/patch"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/admission-controller/pkg/utils"
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

	if err = validatePatchVolumes(volumes); err != nil {
		log.Printf("Rejected volume patch annotation: %v\n", err)
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

// Annotation volumes must not expose the node (e.g. hostPath), so only these sources are allowed.
func validatePatchVolumes(volumes []v1.Volume) error {
	for _, volume := range volumes {
		source := volume.VolumeSource
		source.NFS = nil
		source.PersistentVolumeClaim = nil
		source.ConfigMap = nil
		source.Secret = nil
		source.EmptyDir = nil
		if !reflect.DeepEqual(source, v1.VolumeSource{}) {
			return fmt.Errorf("volume %q in %s annotation has an unsupported source: "+
				"only nfs, persistentVolumeClaim, configMap, secret and emptyDir are allowed",
				volume.Name, volumesPatchAnnotationKey)
		}
	}
	return nil
}

func parseVolumeMountsAnnotation(annotationValue string) ([]v1.VolumeMount, error) {
	var volumeMounts []v1.VolumeMount
	err := json.Unmarshal([]byte(annotationValue), &volumeMounts)
	if err != nil {
		return nil, err
	}

	return volumeMounts, nil
}
