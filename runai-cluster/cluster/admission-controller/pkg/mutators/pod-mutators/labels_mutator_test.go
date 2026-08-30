package pod_mutators

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/run-ai/runai/runai-cluster/cluster/admission-controller/pkg/kube-client/mock"
)

func TestMutateProjectNameLabel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	namespace := v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"runai/queue": testProjectName,
			},
		},
	}

	m := mock.NewMockClient(ctrl)
	m.EXPECT().GetNamespace(testNamespaceName).Return(&namespace, nil)

	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNamespaceName,
		},
	}

	mutator := createMutator(testNamespaceName, false)
	mutator.client = m
	err := mutator.mutateProjectNameLabel(&pod, nil)
	assert.Nil(t, err)
	assert.Equal(t, testProjectName, pod.Labels["project"])
}

func TestMutateProjectNameLabelProjectExists(t *testing.T) {
	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNamespaceName,
			Labels: map[string]string{
				"project": testProjectName,
			},
		},
	}

	mutator := createMutator(testNamespaceName, false)
	err := mutator.mutateProjectNameLabel(&pod, nil)
	assert.Nil(t, err)
	assert.Equal(t, testProjectName, pod.Labels["project"])
}

func TestMutateProjectNameLabelNoProject(t *testing.T) {
	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNamespaceName,
		},
	}

	mutator := createMutator(testNamespaceName, false)
	err := mutator.mutateProjectNameLabel(&pod, nil)
	assert.Nil(t, err)
	assert.Nil(t, pod.Labels)
}
