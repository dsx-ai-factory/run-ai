package pod_identifier

import (
	"testing"

	"github.com/run-ai/runai/runai-cluster/common/constants"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/run-ai/runai/runai-cluster/cluster/admission-controller/pkg/types"
)

func TestShouldMutatePod(t *testing.T) {
	tests := map[string]struct {
		pod            *v1.Pod
		namespace      *v1.Namespace
		expectedResult bool
	}{
		"runai scheduler pod": {
			pod: &v1.Pod{
				Spec: v1.PodSpec{
					SchedulerName: types.RunaiScheduler,
				},
			},
			namespace: &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
			},
			expectedResult: true,
		},
		"kubeflow notebook pod": {
			pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						notebookLabelName: "kubeflow-notebook",
					},
				},
			},
			namespace: &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
			},
			expectedResult: true,
		},
		"azure ml job pod": {
			pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						azureMlLabelKey: "true",
					},
				},
			},
			namespace: &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
			},
			expectedResult: true,
		},
		"workflow owner pod": {
			pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion: "v1",
							Kind:       "Workflow",
							Name:       "owner-workflow",
						},
					},
				},
			},
			namespace: &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
			},
			expectedResult: true,
		},
		"other pod": {
			pod: &v1.Pod{},
			namespace: &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
			},
			expectedResult: false,
		},
		"enforced scheduler namespace": {
			pod: &v1.Pod{
				Spec: v1.PodSpec{},
			},
			namespace: &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						constants.EnforceSchedulerAnnotationName: "true",
					},
				},
			},
			expectedResult: true,
		},
		"enforced scheduler namespace (disabled explicitly)": {
			pod: &v1.Pod{
				Spec: v1.PodSpec{},
			},
			namespace: &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						constants.EnforceSchedulerAnnotationName: "false",
					},
				},
			},
			expectedResult: false,
		},
		"enforced scheduler namespace (not set)": {
			pod: &v1.Pod{
				Spec: v1.PodSpec{},
			},
			namespace: &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
			},
			expectedResult: false,
		},
	}

	for testName, testData := range tests {
		testName := testName
		testData := testData
		t.Run(testName, func(t *testing.T) {
			result := ShouldMutatePod(testData.pod, testData.namespace, true)
			if result != testData.expectedResult {
				t.Errorf("Expected %t, got %t", testData.expectedResult, result)
			}
		})
	}
}
