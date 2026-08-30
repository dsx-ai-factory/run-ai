package affinity

import (
	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

const (
	K8sHostnameLabelName = "kubernetes.io/hostname"
)

func BuildNodeAffinityRequired(nodesToExclude []string) *openapi.NodeAffinityRequired {
	if len(nodesToExclude) == 0 {
		return nil
	}

	excludeExpr := openapi.MatchExpression{
		Key:      K8sHostnameLabelName,
		Operator: openapi.MATCHEXPRESSIONOPERATOR_NOT_IN,
		Values:   nodesToExclude,
	}

	affinityRequired := openapi.NewNodeAffinityRequired()
	affinityRequired.NodeSelectorTerms = []openapi.NodeSelectorTerm{
		{
			MatchExpressions: []openapi.MatchExpression{
				excludeExpr,
			},
		},
	}

	return affinityRequired
}
