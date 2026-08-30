package converter

import (
	"context"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/runai-cluster/common/node-pool-utils/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type NodeAffinityConversionError struct {
	error
}

func NewNodeAffinityConversionError(originalErr error) error {
	return NodeAffinityConversionError{originalErr}
}

// GetRequestedNodePools returns the list of node pools requested by the object.
// The order of sources for the node pools list:
// 1. node pools annotation from pods
// 2. node affinity
// 3. from the object's labels
// 4. from the project
// If none of the sources provides a value, the default nodepool is returned.
func GetRequestedNodePools(ctx context.Context, objectMeta metav1.ObjectMeta, kind string,
	identifiers NodePoolIdentifiers, sources NodePoolsSources) ([]string, error) {
	// 1. annotation
	if sources.PodAnnotations != nil {
		nodePoolsAnnotation, found := sources.PodAnnotations[identifiers.AnnotationNodepoolsKey]
		if found && nodePoolsAnnotation != "" {
			log.Debug().Msgf("For %s <%s/%s>, got node pool list from node pools annotation: <%v>",
				kind, objectMeta.Namespace, objectMeta.Name, nodePoolsAnnotation)
			return strings.Split(nodePoolsAnnotation, " "), nil
		}
	}

	// 2. node affinity
	if sources.AffinitySource.HasValue() {
		requestedNodePoolsFromAffinity, err := convertNodeAffinityToNodePoolNames(ctx, sources.AffinitySource.Affinity, identifiers, sources.AffinitySource.ReaderClient)
		if err != nil {
			log.Error().Msgf("Failed converting NodeAffinity to node pool names, err: <%s>", err.Error())
			if !sources.AffinitySource.IgnoreConversionErrors {
				// wrap with typed error so it will be identifiable
				return []string{}, NewNodeAffinityConversionError(err)
			}
		}

		if len(requestedNodePoolsFromAffinity) > 0 {
			log.Debug().Msgf("For %s <%s/%s>, got node pool list from node affinity: <%v>",
				kind, objectMeta.Namespace, objectMeta.Name, requestedNodePoolsFromAffinity)
			return requestedNodePoolsFromAffinity, nil
		}
	}

	// 3. object's labels
	// see if there's a node pool label on the object itself
	nodePoolName, found := objectMeta.Labels[identifiers.NodePoolAssignmentLabelKey]
	if found && nodePoolName != identifiers.UnexistingNodepoolSentinel {
		log.Debug().Msgf("For %s <%s/%s>, got node pool list from node pool label: <[%s]>",
			kind, objectMeta.Namespace, objectMeta.Name, nodePoolName)
		return []string{nodePoolName}, nil
	}

	// 4. project
	if sources.Project != nil && sources.Project.Spec.DefaultNodePools != nil && len(sources.Project.Spec.DefaultNodePools) > 0 {
		log.Debug().Msgf("For %s <%s/%s>, got node pool list from project default nodepools: <%v>",
			kind, objectMeta.Namespace, objectMeta.Name, sources.Project.Spec.DefaultNodePools)
		return sources.Project.Spec.DefaultNodePools, nil
	}

	// default
	log.Debug().Msgf("For %s <%s/%s>, using default-node-pool as list: <[%s]>",
		kind, objectMeta.Namespace, objectMeta.Name, identifiers.DefaultNodepoolName)
	return []string{identifiers.DefaultNodepoolName}, nil
}

func convertNodeAffinityToNodePoolNames(
	ctx context.Context, affinity *corev1.Affinity,
	identifiers NodePoolIdentifiers, readerClient client.Reader) ([]string, error) {
	if affinity == nil {
		log.Ctx(ctx).Debug().Msg("Affinity is nil")
		return []string{}, nil
	}

	nodeAffinity := affinity.NodeAffinity
	if nodeAffinity == nil {
		log.Ctx(ctx).Debug().Msg("NodeAffinity is nil")
		return []string{}, nil
	}
	nodeSelector := nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution
	if nodeSelector == nil {
		log.Ctx(ctx).Debug().Msg("NodeSelector is nil")
		return []string{}, nil
	}
	nodeSelectorTerms := nodeSelector.NodeSelectorTerms
	if len(nodeSelectorTerms) == 0 {
		log.Ctx(ctx).Debug().Msg("NodeSelectorTerms is empty")
		return []string{}, nil
	}

	allNodePools, err := utils.GetAllNodePools(ctx, readerClient)
	if err != nil {
		return []string{}, err
	}
	nodePoolKeyToValueToName, deletingNodePools := utils.GetNodePoolsMap(allNodePools.Items)

	result := []string{}
	for _, nodeSelectorTerm := range nodeSelectorTerms {
		foundNodePoolsInCurrentNodeSelectorTerm := false
		for _, matchExpression := range nodeSelectorTerm.MatchExpressions {
			if foundNodePoolsInCurrentNodeSelectorTerm {
				// can't have multiple node pool conditions in same nodeSelectorTerm (but in different matchExpression)
				log.Ctx(ctx).Debug().Msgf("Found node pools in current NodeSelectorTerm - continuing to next one")
				break
			}
			result, foundNodePoolsInCurrentNodeSelectorTerm = getNodePoolsForMatchExpression(
				ctx, identifiers, matchExpression, nodePoolKeyToValueToName, deletingNodePools, result)
		}
	}

	// removing duplicates from the resulting node pool options list
	return utils.RemoveDuplicates(result), nil
}

func getNodePoolsForMatchExpression(
	ctx context.Context,
	identifiers NodePoolIdentifiers,
	matchExpression corev1.NodeSelectorRequirement,
	nodePoolKeyToValueToName map[string]map[string]string,
	deletingNodePools map[string]string,
	result []string) ([]string, bool) {
	// we only support operator "In"
	if matchExpression.Operator != corev1.NodeSelectorOpIn {
		if matchExpression.Operator == corev1.NodeSelectorOpDoesNotExist {
			// we only support the special case of the assignment label key with DoesNotExist
			// — to support the default node pool (which is represented by absence of the label).
			if matchExpression.Key == identifiers.NodePoolAssignmentLabelKey {
				result = append(result, identifiers.DefaultNodepoolName)
				return result, true
			}
		}
		log.Ctx(ctx).Debug().Msgf("MatchExpression's operator <%v> is not supported", matchExpression.Operator)
		return result, false
	}

	nodePoolsWithKeyMap, found := nodePoolKeyToValueToName[matchExpression.Key]
	if !found {
		log.Ctx(ctx).Debug().Msgf("MatchExpression's key <%s> is not found in node pool map", matchExpression.Key)
		return result, false
	}

	foundNodePools := false
	for _, matchExpressionValue := range matchExpression.Values {
		nodePoolName, npnFound := nodePoolsWithKeyMap[matchExpressionValue]
		if !npnFound {
			log.Ctx(ctx).Debug().Msgf("MatchExpression's key and value <%s/%s> are not found in node pool map",
				matchExpression.Key, matchExpressionValue)
			continue
		}

		if nodePoolPhase, nppFound := deletingNodePools[nodePoolName]; nppFound {
			log.Ctx(ctx).Warn().Msgf("Warning: MatchExpression <%v> matched nodepool <%s>,"+
				" but it is not ready for Scheduling, its phase is: <%v>",
				matchExpression, nodePoolName, nodePoolPhase)
			continue
		}

		log.Ctx(ctx).Debug().Msgf("MatchExpression <%v> matched nodepool <%s>!",
			matchExpression, nodePoolName)
		result = append(result, nodePoolName)
		foundNodePools = true
	}
	return result, foundNodePools
}
