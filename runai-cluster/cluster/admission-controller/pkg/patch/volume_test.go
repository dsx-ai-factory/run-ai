package patch

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

func TestAddHostPathDirectoryOrCreateVolume(t *testing.T) {
	podSpec := &v1.PodSpec{}
	AddHostPathDirectoryOrCreateVolume(podSpec, "vol", "path")
	assert.Equal(t, len(podSpec.Volumes), 1)
	assert.Equal(t, podSpec.Volumes[0].Name, "vol")
	assert.Equal(t, podSpec.Volumes[0].HostPath.Path, "path")
	assert.Equal(t, *podSpec.Volumes[0].HostPath.Type, v1.HostPathDirectoryOrCreate)
}

func TestAddHostPathDirectoryVolume(t *testing.T) {
	podSpec := &v1.PodSpec{}
	AddHostPathDirectoryVolume(podSpec, "vol", "path")
	assert.Equal(t, len(podSpec.Volumes), 1)
	assert.Equal(t, podSpec.Volumes[0].Name, "vol")
	assert.Equal(t, podSpec.Volumes[0].HostPath.Path, "path")
	assert.Equal(t, *podSpec.Volumes[0].HostPath.Type, v1.HostPathDirectory)
}

func TestAddMultipleVolumes(t *testing.T) {
	podSpec := &v1.PodSpec{}
	AddHostPathDirectoryVolume(podSpec, "vol1", "path1")
	AddHostPathDirectoryOrCreateVolume(podSpec, "vol2", "path2")
	assert.Equal(t, len(podSpec.Volumes), 2)

	for _, vol := range podSpec.Volumes {
		switch vol.Name {
		case "vol1":
			assert.Equal(t, vol.HostPath.Path, "path1")
			assert.Equal(t, *vol.HostPath.Type, v1.HostPathDirectory)
		case "vol2":
			assert.Equal(t, vol.HostPath.Path, "path2")
			assert.Equal(t, *vol.HostPath.Type, v1.HostPathDirectoryOrCreate)
		default:
			assert.Fail(t, "unexpected volume")
		}
	}
}

func TestOverrideVolume(t *testing.T) {
	podSpec := &v1.PodSpec{
		Volumes: []v1.Volume{
			{
				Name: "vol-name",
				VolumeSource: v1.VolumeSource{
					HostPath: &v1.HostPathVolumeSource{
						Path: "path",
					},
				},
			},
		},
	}

	AddHostPathDirectoryVolume(podSpec, "vol-name", "path2")
	assert.Equal(t, len(podSpec.Volumes), 1)
	assert.Equal(t, podSpec.Volumes[0].Name, "vol-name")
	assert.Equal(t, podSpec.Volumes[0].HostPath.Path, "path2")
	assert.Equal(t, *podSpec.Volumes[0].HostPath.Type, v1.HostPathDirectory)
}

func TestAddConfigMapVolume(t *testing.T) {
	podSpec := &v1.PodSpec{}
	AddConfigMapVolume(podSpec, "vol", "conf")
	assert.Equal(t, len(podSpec.Volumes), 1)
	assert.Equal(t, podSpec.Volumes[0].Name, "vol")
	assert.Equal(t, podSpec.Volumes[0].ConfigMap.LocalObjectReference.Name, "conf")
}

func TestOverrideConfigMapVolume(t *testing.T) {
	podSpec := &v1.PodSpec{
		Volumes: []v1.Volume{
			{
				Name: "name",
				VolumeSource: v1.VolumeSource{
					ConfigMap: &v1.ConfigMapVolumeSource{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "conf1",
						},
					},
				},
			},
		},
	}

	AddConfigMapVolume(podSpec, "name", "conf2")
	assert.Equal(t, len(podSpec.Volumes), 1)
	assert.Equal(t, podSpec.Volumes[0].Name, "name")
	assert.Equal(t, podSpec.Volumes[0].ConfigMap.LocalObjectReference.Name, "conf2")
}

func TestAddVolumeMountToContainer(t *testing.T) {
	container := &v1.Container{}
	AddVolumeMountToContainer(container, "name", "mount", "path", true)
	assert.Equal(t, len(container.VolumeMounts), 1)
	assert.Equal(t, container.VolumeMounts[0].Name, "name")
	assert.Equal(t, container.VolumeMounts[0].MountPath, "mount")
	assert.Equal(t, container.VolumeMounts[0].SubPath, "path")
	assert.Equal(t, container.VolumeMounts[0].ReadOnly, true)
}

func TestOverrideVolumeMountInContainer(t *testing.T) {
	container := &v1.Container{
		VolumeMounts: []v1.VolumeMount{
			{
				Name:      "name",
				MountPath: "path1",
				SubPath:   "subPath1",
				ReadOnly:  false,
			},
		},
	}
	AddVolumeMountToContainer(container, "name", "path1", "subPath2", true)
	assert.Equal(t, len(container.VolumeMounts), 1)
	assert.Equal(t, container.VolumeMounts[0].Name, "name")
	assert.Equal(t, container.VolumeMounts[0].MountPath, "path1")
	assert.Equal(t, container.VolumeMounts[0].SubPath, "subPath2")
	assert.Equal(t, container.VolumeMounts[0].ReadOnly, true)
}
