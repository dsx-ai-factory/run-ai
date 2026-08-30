package pod

import (
	corev1 "k8s.io/api/core/v1"
)

// AddNodeAffinityToPodSpec adds matchExpressionsToAdd to the pod spec's required
// node affinity. Each requirement is OR'd against the others (one per node-selector
// term), while expressions sharing a term are AND'd. An existing expression on the
// same key is replaced.
func AddNodeAffinityToPodSpec(podSpec *corev1.PodSpec, matchExpressionsToAdd []corev1.NodeSelectorRequirement) {
	if podSpec.Affinity == nil {
		podSpec.Affinity = &corev1.Affinity{}
	}
	if podSpec.Affinity.NodeAffinity == nil {
		podSpec.Affinity.NodeAffinity = &corev1.NodeAffinity{}
	}
	if podSpec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
		podSpec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution = &corev1.NodeSelector{}
	}
	selector := podSpec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution
	if len(selector.NodeSelectorTerms) == 0 {
		selector.NodeSelectorTerms = []corev1.NodeSelectorTerm{}
		for _, matchExpressionToAdd := range matchExpressionsToAdd {
			selector.NodeSelectorTerms = append(selector.NodeSelectorTerms, corev1.NodeSelectorTerm{
				MatchExpressions: []corev1.NodeSelectorRequirement{matchExpressionToAdd},
			})
		}
		return
	}

	updatedTerms := make([]corev1.NodeSelectorTerm, 0)
	for _, matchExpressionToAdd := range matchExpressionsToAdd {
		for _, term := range selector.NodeSelectorTerms {
			updatedExpressions := make([]corev1.NodeSelectorRequirement, 0)
			for _, expr := range term.MatchExpressions {
				if expr.Key != matchExpressionToAdd.Key {
					updatedExpressions = append(updatedExpressions, expr)
				}
			}
			updatedExpressions = append(updatedExpressions, matchExpressionToAdd)
			term.MatchExpressions = updatedExpressions
			updatedTerms = append(updatedTerms, term)
		}
	}
	if len(updatedTerms) > 0 {
		selector.NodeSelectorTerms = updatedTerms
	}
}
