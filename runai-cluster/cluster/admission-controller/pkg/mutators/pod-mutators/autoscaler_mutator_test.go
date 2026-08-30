package pod_mutators

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"

	"github.com/run-ai/runai/runai-cluster/common/priority"
)

func TestAutoScalerEvictAnnotationForInteractivePod(t *testing.T) {
	namespaceName := "runai-some-name-1"
	interactivePod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{},
			Labels: map[string]string{
				"priorityClassName": priority.InteractivePriorityClassName,
			},
		},
	}

	mutator := createMutator(namespaceName, false)
	err := mutator.setAutoScalerAnnotations(interactivePod, nil)

	assert.Equal(t, err, nil)
	assert.Equal(t, len(interactivePod.Annotations), 2)

	val, ok := interactivePod.Annotations[autoscalerEvictAnnotationKey]
	assert.Equal(t, ok, true)
	assert.Equal(t, val, autoscalerEvictAnnotationValue)

	val, ok = interactivePod.Annotations[karpenterDoNotDisruptAnnotationKey]
	assert.Equal(t, ok, true)
	assert.Equal(t, val, karpenterDoNotDisruptAnnotationValue)
}

func TestAutoScalerEvictAnnotationForInference(t *testing.T) {
	namespaceName := "runai-some-name-1"
	interactivePod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{},
			Labels: map[string]string{
				"priorityClassName": priority.InferencePriorityClassName,
			},
		},
	}

	mutator := createMutator(namespaceName, false)
	err := mutator.setAutoScalerAnnotations(interactivePod, nil)

	assert.Equal(t, err, nil)
	assert.Equal(t, len(interactivePod.Annotations), 2)

	val, ok := interactivePod.Annotations[autoscalerEvictAnnotationKey]
	assert.Equal(t, ok, true)
	assert.Equal(t, val, autoscalerEvictAnnotationValue)

	val, ok = interactivePod.Annotations[karpenterDoNotDisruptAnnotationKey]
	assert.Equal(t, ok, true)
	assert.Equal(t, val, karpenterDoNotDisruptAnnotationValue)
}

func TestAutoScalerEvictAnnotationForTrainPod(t *testing.T) {
	namespaceName := "runai-some-name-1"
	nonInteractivePod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{},
			Labels: map[string]string{
				"priorityClassName": priority.TrainPriorityClassName,
			},
		},
	}

	mutator := createMutator(namespaceName, false)
	err := mutator.setAutoScalerAnnotations(nonInteractivePod, nil)

	assert.Equal(t, err, nil)
	assert.Equal(t, len(nonInteractivePod.Annotations), 0)
}

func TestAutoScalerEvictAnnotationForInteractivePreemptiblePod(t *testing.T) {
	namespaceName := "runai-some-name-1"
	nonInteractivePod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{},
			Labels: map[string]string{
				"priorityClassName": priority.InteractivePreemptiblePriorityClassName,
			},
		},
	}

	mutator := createMutator(namespaceName, false)
	err := mutator.setAutoScalerAnnotations(nonInteractivePod, nil)

	assert.Equal(t, err, nil)
	assert.Equal(t, len(nonInteractivePod.Annotations), 0)
}

func TestAutoScalerEvictAnnotationForPodWithUnknownPriority(t *testing.T) {
	namespaceName := "runai-some-name-1"
	nonInteractivePod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{},
			Labels: map[string]string{
				"priorityClassName": "unknown",
			},
		},
	}

	mutator := createMutator(namespaceName, false)
	err := mutator.setAutoScalerAnnotations(nonInteractivePod, nil)

	assert.Equal(t, err, nil)
	assert.Equal(t, len(nonInteractivePod.Annotations), 0)
}

func TestAutoScalerEvictAnnotationForPodWithoutPriorityLabel(t *testing.T) {
	namespaceName := "runai-some-name-1"
	nonInteractivePod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{},
			Labels:      map[string]string{},
		},
	}

	mutator := createMutator(namespaceName, false)
	err := mutator.setAutoScalerAnnotations(nonInteractivePod, nil)

	assert.Equal(t, err, nil)
	assert.Equal(t, len(nonInteractivePod.Annotations), 0)
}
