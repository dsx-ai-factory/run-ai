package patch

import (
	v1 "k8s.io/api/core/v1"
)

func AddHostPathDirectoryOrCreateVolume(podSpec *v1.PodSpec, volumeName string, path string) {
	addHostPathVolumeInner(podSpec, volumeName, path, v1.HostPathDirectoryOrCreate)
}

func AddHostPathDirectoryVolume(podSpec *v1.PodSpec, volumeName string, path string) {
	addHostPathVolumeInner(podSpec, volumeName, path, v1.HostPathDirectory)
}

func addHostPathVolumeInner(podSpec *v1.PodSpec, volumeName string, path string, pathType v1.HostPathType) {
	volume := v1.Volume{
		Name: volumeName,
		VolumeSource: v1.VolumeSource{
			HostPath: &v1.HostPathVolumeSource{
				Path: path,
				Type: &pathType,
			},
		},
	}
	AddVolumeToPodSpec(podSpec, volume)
}

func AddConfigMapVolume(podSpec *v1.PodSpec, volumeName string, configMapName string) {
	if podSpec.Volumes == nil {
		podSpec.Volumes = make([]v1.Volume, 0)
	}

	updatedVolumes := make([]v1.Volume, 0)
	for _, volume := range podSpec.Volumes {
		if volume.Name != volumeName {
			updatedVolumes = append(updatedVolumes, volume)
		}
	}
	podSpec.Volumes = updatedVolumes

	volume := v1.Volume{
		Name: volumeName,
		VolumeSource: v1.VolumeSource{
			ConfigMap: &v1.ConfigMapVolumeSource{
				LocalObjectReference: v1.LocalObjectReference{
					Name: configMapName,
				},
			},
		},
	}
	podSpec.Volumes = append(podSpec.Volumes, volume)
}

func AddVolumeToPodSpec(podSpec *v1.PodSpec, volume v1.Volume) {
	if podSpec.Volumes == nil {
		podSpec.Volumes = make([]v1.Volume, 0)
	}

	updatedVolumes := make([]v1.Volume, 0)
	for _, existingVolume := range podSpec.Volumes {
		if existingVolume.Name != volume.Name {
			updatedVolumes = append(updatedVolumes, existingVolume)
		}
	}

	updatedVolumes = append(updatedVolumes, volume)
	podSpec.Volumes = updatedVolumes
}

func AddVolumeMountToContainer(container *v1.Container, volumeName, mountPath, subPath string, readOnly bool) {
	if container.VolumeMounts == nil {
		container.VolumeMounts = make([]v1.VolumeMount, 0)
	}

	updatedMounts := make([]v1.VolumeMount, 0)
	for _, mount := range container.VolumeMounts {
		if mount.MountPath != mountPath {
			updatedMounts = append(updatedMounts, mount)
		}
	}
	container.VolumeMounts = updatedMounts

	volumeMount := v1.VolumeMount{
		Name:      volumeName,
		MountPath: mountPath,
		SubPath:   subPath,
		ReadOnly:  readOnly,
	}
	container.VolumeMounts = append(container.VolumeMounts, volumeMount)
}
