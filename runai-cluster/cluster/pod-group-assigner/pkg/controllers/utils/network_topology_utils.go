package utils

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"

	kaitopologyv1alpha1 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1alpha1"
	"github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetNodePoolPreferredNetworkTopologyName returns the NetworkTopologyName configured on the named
// NodePool, or an empty string when the NodePool exists but has no topology configured. A failure to
// get the NodePool (including not-found) is returned as an error rather than treated as "no
// topology", so a missing or stale NodePool reference is surfaced instead of silently clearing a
// valid constraint.
func GetNodePoolPreferredNetworkTopologyName(ctx context.Context, k8sClient client.Client, nodePoolName string) (string, error) {
	key := types.NamespacedName{Name: nodePoolName}

	nodePool := &v1alpha1.NodePool{}
	if err := k8sClient.Get(ctx, key, nodePool); err != nil {
		log.Ctx(ctx).Error().Msgf("Failed getting NodePool <%s> object, err: <%s>", nodePoolName, err.Error())
		return "", err
	}

	return nodePool.Spec.PreferredNetworkTopologyName, nil
}

// GetNetworkTopologyLowestLevel returns the node label of the lowest (deepest) level of the named
// kai.scheduler Topology CR. The levels are ordered from the highest level to the lowest, so the
// lowest level is the last element.
func GetNetworkTopologyLowestLevel(ctx context.Context, k8sClient client.Client, networkTopologyName string) (string, error) {
	topology := &kaitopologyv1alpha1.Topology{}
	err := k8sClient.Get(ctx, types.NamespacedName{Name: networkTopologyName}, topology)
	if err != nil {
		log.Ctx(ctx).Error().Msgf("Failed getting Topology <%s> object, err: <%s>", networkTopologyName, err.Error())
		return "", err
	}

	if len(topology.Spec.Levels) == 0 {
		return "", fmt.Errorf("topology <%s> has no levels", networkTopologyName)
	}

	return topology.Spec.Levels[len(topology.Spec.Levels)-1].NodeLabel, nil
}
