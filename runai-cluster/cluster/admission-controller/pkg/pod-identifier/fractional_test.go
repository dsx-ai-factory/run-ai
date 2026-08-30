package pod_identifier

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"

	"github.com/run-ai/runai/runai-cluster/cluster/admission-controller/pkg/types"
)

func TestIsFractionalPodFractionAnnotation(t *testing.T) {
	annotations := map[string]string{
		types.RunaiGpuFraction: "0.1",
	}
	containers := []v1.Container{}
	assert.True(t, IsFractionalPod(annotations, containers))
}

func TestIsFractionalPodMemoryAnnotation(t *testing.T) {
	annotations := map[string]string{
		types.RunaiGpuMemory: "0.1",
	}
	containers := []v1.Container{}
	assert.True(t, IsFractionalPod(annotations, containers))
}

func TestIsFractionalPodEnvVar(t *testing.T) {
	annotations := map[string]string{}
	containers := []v1.Container{
		{
			Env: []v1.EnvVar{
				{
					Name:  types.RunaiGpuFracEnvVar,
					Value: "some-value",
				},
			},
		},
	}
	assert.True(t, IsFractionalPod(annotations, containers))
}

func TestNotFractionalPod(t *testing.T) {
	annotations := map[string]string{}
	containers := []v1.Container{}
	assert.False(t, IsFractionalPod(annotations, containers))
}

func TestHasGpuFractionalAnnotationWithFraction(t *testing.T) {
	annotations := map[string]string{
		types.RunaiGpuFraction: "0.1",
	}
	assert.True(t, HasGpuFractionalAnnotation(annotations))
}

func TestHasGpuFractionalAnnotationWithGpuMemory(t *testing.T) {
	annotations := map[string]string{
		types.RunaiGpuMemory: "0.1",
	}
	assert.True(t, HasGpuFractionalAnnotation(annotations))
}

func TestHasGpuFractionalAnnotationNoAnnotations(t *testing.T) {
	annotations := map[string]string{}
	assert.False(t, HasGpuFractionalAnnotation(annotations))
}

func TestHasFractionalEnvVar(t *testing.T) {
	containers := []v1.Container{
		{
			Env: []v1.EnvVar{
				{
					Name:  types.RunaiGpuFracEnvVar,
					Value: "some-value",
				},
			},
		},
	}
	exists, value := HasFractionalEnvVar(containers)
	assert.True(t, exists)
	assert.Equal(t, "some-value", value)
}

func TestHasFractionalEnvVarDoesNotExist(t *testing.T) {
	containers := []v1.Container{
		{
			Env: []v1.EnvVar{
				{
					Name:  "another-var",
					Value: "some-value",
				},
			},
		},
	}
	exists, _ := HasFractionalEnvVar(containers)
	assert.False(t, exists)
}
