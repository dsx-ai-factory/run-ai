package services

import (
	"github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/models"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	"k8s.io/utils/ptr"
)

func NodeAffinityModelToApi(affinity models.NodeAffinity) api.PodNodeAffinity {
	result := api.PodNodeAffinity{}

	if affinity.Required != nil {
		result.Required = &api.PodNodeAffinityRequired{
			NodeSelectorTerms: convutils.ConvertSliceTypes(affinity.Required.NodeSelectorTerms, NodeSelectorTermModelToApi),
		}
	}

	if affinity.Preferred != nil {
		result.Preferred = &api.PodNodeAffinityPreferred{
			NodeSelectorTerms: convutils.ConvertSliceTypes(affinity.Preferred.NodeSelectorTerms, PreferredNodeSelectorTermModelToApi),
		}
	}

	return result
}

func NodeAffinityApiToModel(affinity api.PodNodeAffinity) models.NodeAffinity {
	result := models.NodeAffinity{}

	if affinity.Required != nil {
		result.Required = &models.NodeSelector{
			NodeSelectorTerms: convutils.ConvertSliceTypes(affinity.Required.NodeSelectorTerms, NodeSelectorTermApiToModel),
		}
	}

	if affinity.Preferred != nil {
		result.Preferred = &models.PreferredNodeSelector{
			NodeSelectorTerms: convutils.ConvertSliceTypes(affinity.Preferred.NodeSelectorTerms, PreferredNodeSelectorTermApiToModel),
		}
	}

	return result
}

func NodeSelectorTermModelToApi(term models.NodeSelectorTerm) api.NodeSelectorTerm {
	return api.NodeSelectorTerm{
		MatchExpressions: convutils.ConvertSliceTypes(term.MatchExpressions, matchExpressionModelToApi),
	}
}

func NodeSelectorTermApiToModel(term api.NodeSelectorTerm) models.NodeSelectorTerm {
	return models.NodeSelectorTerm{
		MatchExpressions: convutils.ConvertSliceTypes(term.MatchExpressions, matchExpressionApiToModel),
	}
}

func PreferredNodeSelectorTermModelToApi(term models.PreferredNodeSelectorTerm) api.PodPreferredNodeSelectorTerm {
	return api.PodPreferredNodeSelectorTerm{
		Weight:     float32(term.Weight),
		Preference: ptr.To(NodeSelectorTermModelToApi(term.Preference)),
	}
}

func PreferredNodeSelectorTermApiToModel(term api.PodPreferredNodeSelectorTerm) models.PreferredNodeSelectorTerm {
	var preference models.NodeSelectorTerm
	if term.Preference != nil {
		preference = NodeSelectorTermApiToModel(*term.Preference)
	}
	return models.PreferredNodeSelectorTerm{
		Weight:     int32(term.Weight),
		Preference: preference,
	}
}

func matchExpressionModelToApi(expr models.NodeSelectorRequirement) api.MatchExpression {
	return api.MatchExpression{
		Key:      expr.Key,
		Operator: api.MatchExpressionOperator(expr.Operator),
		Values:   expr.Values,
	}
}

func matchExpressionApiToModel(expr api.MatchExpression) models.NodeSelectorRequirement {
	return models.NodeSelectorRequirement{
		Key:      expr.Key,
		Operator: string(expr.Operator),
		Values:   expr.Values,
	}
}

func workloadNodeAffinityApiToModel(affinity api.WorkloadNodeAffinity) models.WorkloadNodeAffinity {
	return models.WorkloadNodeAffinity{
		SpecName:     affinity.SpecName,
		NodeAffinity: NodeAffinityApiToModel(affinity.NodeAffinity),
	}
}

func workloadNodeAffinityModelToApi(affinity models.WorkloadNodeAffinity) api.WorkloadNodeAffinity {
	return api.WorkloadNodeAffinity{
		SpecName:     affinity.SpecName,
		NodeAffinity: NodeAffinityModelToApi(affinity.NodeAffinity),
	}
}
