package pod_mutators

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/run-ai/runai/runai-cluster/cluster/admission-controller/pkg/types"
)

func TestMutateGKEFractionPod(t *testing.T) {
	pod := createPod(true)
	mutator := NewPodMutator(nil, nil, nil)
	err := mutator.mutateFractionTolerationPod(pod, nil)
	assert.Nil(t, err)

	assert.Equal(t, len(pod.Spec.Tolerations), 1)
	assert.Equal(t, pod.Spec.Tolerations[0].Key, types.GpuResource)
	assert.Equal(t, pod.Spec.Tolerations[0].Operator, v1.TolerationOpExists)
	assert.Equal(t, pod.Spec.Tolerations[0].Effect, v1.TaintEffectNoSchedule)
	assert.Equal(t, pod.Spec.Tolerations[0].Value, "")

}

func TestMutateGKENonFraction(t *testing.T) {
	pod := createPod(false)
	mutator := NewPodMutator(nil, nil, nil)
	err := mutator.mutateFractionTolerationPod(pod, nil)
	assert.Nil(t, err)
	assert.Zero(t, len(pod.Spec.Volumes))
	assert.Zero(t, len(pod.Spec.Tolerations))
}

func createPod(isFraction bool) *v1.Pod {
	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "my-pod",
			Annotations: map[string]string{},
			OwnerReferences: []metav1.OwnerReference{
				{
					Name: "job-1",
					UID:  "12345",
				},
			},
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Resources: v1.ResourceRequirements{
						Limits: v1.ResourceList{types.GpuResource: resource.MustParse("1")},
					},
					Name: "my-container",
				},
			},
		},
	}

	if isFraction {
		pod.ObjectMeta.Annotations[types.RunaiGpuFraction] = "0.1"
	}
	return &pod
}
