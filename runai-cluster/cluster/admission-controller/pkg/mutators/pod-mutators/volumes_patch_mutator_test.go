package pod_mutators

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNoAnnotations(t *testing.T) {
	pod := createTestPod(nil)
	mutator := createMutator(testNamespaceName, false)
	err := mutator.patchVolumesAndMounts(pod, nil)
	assert.Nil(t, err)
	assert.Equal(t, len(pod.Spec.Volumes), 0)
	assert.Equal(t, len(pod.Spec.Containers[0].VolumeMounts), 0)
}

func TestNfsVolume(t *testing.T) {
	annotations := map[string]string{
		volumesPatchAnnotationKey:      getNFSVolumeJson(t),
		volumeMountsPatchAnnotationKey: getVolumeMountJson(t),
	}
	pod := createTestPod(annotations)
	mutator := createMutator(testNamespaceName, false)
	err := mutator.patchVolumesAndMounts(pod, nil)
	assert.Nil(t, err)

	assert.Equal(t, len(pod.Spec.Volumes), 1)
	volume := pod.Spec.Volumes[0]
	assert.Equal(t, volume.Name, "host-path-volume-0")
	assert.Equal(t, volume.NFS.Server, "shaibid68a427183f6612e.file.core.windows.net")
	assert.Equal(t, volume.NFS.Path, "/shaibid68a427183f6612e/runai3")

	container := pod.Spec.Containers[0]
	assert.Equal(t, len(container.VolumeMounts), 1)
	volumeMount := container.VolumeMounts[0]
	assert.Equal(t, volumeMount.Name, "host-path-volume-0")
	assert.Equal(t, volumeMount.ReadOnly, true)
	assert.Equal(t, volumeMount.MountPath, "/container")
}

func TestInvalidVolumeJson(t *testing.T) {
	annotations := map[string]string{
		volumesPatchAnnotationKey:      `[invalid json]`,
		volumeMountsPatchAnnotationKey: `[invalid json]`,
	}
	pod := createTestPod(annotations)

	mutator := createMutator(testNamespaceName, false)
	err := mutator.patchVolumesAndMounts(pod, nil)
	assert.NotNil(t, err)

	assert.Equal(t, len(pod.Spec.Volumes), 0)
	assert.Equal(t, len(pod.Spec.Containers[0].VolumeMounts), 0)
}

func createTestPod(annotations map[string]string) *v1.Pod {
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   testNamespaceName,
			Annotations: annotations,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name: "my-container",
				},
			},
		},
	}
}

func getNFSVolumeJson(t *testing.T) string {
	volumes := []v1.Volume{
		{
			Name: "host-path-volume-0",
			VolumeSource: v1.VolumeSource{
				NFS: &v1.NFSVolumeSource{
					Server: "shaibid68a427183f6612e.file.core.windows.net",
					Path:   "/shaibid68a427183f6612e/runai3",
				},
			},
		},
	}

	volumesJson, err := json.Marshal(volumes)
	assert.Nil(t, err)
	return string(volumesJson)
}

func getVolumeMountJson(t *testing.T) string {
	volumeMounts := []v1.VolumeMount{
		{
			Name:      "host-path-volume-0",
			MountPath: "/container",
			ReadOnly:  true,
		},
	}
	volumeMountJson, err := json.Marshal(volumeMounts)
	assert.Nil(t, err)
	return string(volumeMountJson)
}
