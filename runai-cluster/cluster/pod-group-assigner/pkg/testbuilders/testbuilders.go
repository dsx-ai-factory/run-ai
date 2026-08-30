// Package testbuilders provides object constructors shared by the
// pod-group-assigner fake-client-based test suites. Living in a non-test
// package lets the controller suite and the converter unit tests import the
// same builders.
package testbuilders

import (
	kaitopologyv1alpha1 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1alpha1"
	commonconstants "github.com/kai-scheduler/KAI-scheduler/pkg/common/constants"
	"github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BuildPodWithNodeAffinity returns a Pod annotated for the given pod group,
// with the supplied required node-affinity expressions. Each inner slice
// becomes one NodeSelectorTerm. Pass nil/empty to skip affinity entirely.
func BuildPodWithNodeAffinity(podName, podGroupName, namespace string, matchExpressionss [][]corev1.NodeSelectorRequirement) *corev1.Pod {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: namespace,
			Annotations: map[string]string{
				commonconstants.PodGroupAnnotationForPod: podGroupName,
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{Name: "container-1", Image: "ubuntu"}},
		},
	}
	if len(matchExpressionss) == 0 {
		return pod
	}

	pod.Spec.Affinity = &corev1.Affinity{
		NodeAffinity: &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{},
		},
	}
	for _, matchExpressions := range matchExpressionss {
		pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms =
			append(pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms,
				corev1.NodeSelectorTerm{MatchExpressions: matchExpressions})
	}
	return pod
}

// BuildNodePool returns a kai.resources NodePool with the given identity and a status
// phase. Used by both envtest setup and unit tests.
func BuildNodePool(name, labelKey, labelValue string, phase v1alpha1.NodePoolPhase) *v1alpha1.NodePool {
	return &v1alpha1.NodePool{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: v1alpha1.NodePoolSpec{
			LabelKey:   labelKey,
			LabelValue: labelValue,
		},
		Status: v1alpha1.NodePoolStatus{Phase: phase},
	}
}

// BuildTopology returns a kai.scheduler Topology CR whose levels are ordered
// highest-to-lowest, mirroring the real CR. The pod-group-assigner stamps the
// lowest (last) level as the preferred placement.
func BuildTopology(name string, levelNodeLabels ...string) *kaitopologyv1alpha1.Topology {
	levels := make([]kaitopologyv1alpha1.TopologyLevel, 0, len(levelNodeLabels))
	for _, nodeLabel := range levelNodeLabels {
		levels = append(levels, kaitopologyv1alpha1.TopologyLevel{NodeLabel: nodeLabel})
	}

	return &kaitopologyv1alpha1.Topology{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec:       kaitopologyv1alpha1.TopologySpec{Levels: levels},
	}
}
