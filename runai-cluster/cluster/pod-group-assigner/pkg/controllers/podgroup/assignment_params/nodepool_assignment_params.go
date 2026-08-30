package assignment_params

import (
	"context"
	"fmt"

	kaiv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"
	"github.com/rs/zerolog/log"

	"github.com/run-ai/runai/runai-cluster/cluster/pod-group-assigner/pkg/controllers/common"
	"github.com/run-ai/runai/runai-cluster/cluster/pod-group-assigner/pkg/controllers/utils"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type NodePoolAssignmentParams struct {
	NodePoolName                                     string
	MarkUnschedulable                                bool
	SchedulingBackoff                                int32
	RemoveUnschedulableOnNodePoolSchedulingCondition bool
}

func GetNodePoolAssignmentParams(
	ctx context.Context,
	client client.Client,
	currentMarkUnschedulableForPodGroup *bool,
	lastSchedulingCondition *kaiv2alpha2.SchedulingCondition,
	requestedNodePools []string) (*NodePoolAssignmentParams, error) {
	nodePoolAssignmentParams, err := calculateNodePoolAssignmentParams(ctx, client, currentMarkUnschedulableForPodGroup, lastSchedulingCondition, requestedNodePools)
	if err != nil {
		return nil, err
	}

	if len(requestedNodePools) > 1 && lastSchedulingCondition != nil && nodePoolAssignmentParams.NodePoolName == lastSchedulingCondition.NodePool {
		// re-set the scheduling conditions instead of changing the scheduling backoff, since other nodepools could become available in the future
		nodePoolAssignmentParams.RemoveUnschedulableOnNodePoolSchedulingCondition = true
		log.Ctx(ctx).Warn().Msgf("Next assignment is to the same node pool as current (%s), removing UnschedulableOnNodePool SchedulingCondition",
			nodePoolAssignmentParams.NodePoolName)
	}

	return nodePoolAssignmentParams, nil
}

func calculateNodePoolAssignmentParams(
	ctx context.Context,
	client client.Client,
	currentMarkUnschedulableForPodGroup *bool,
	lastSchedulingCondition *kaiv2alpha2.SchedulingCondition,
	requestedNodePools []string) (*NodePoolAssignmentParams, error) {
	nodePoolAssignmentParams := getInitialNodePoolAssignmentParams(currentMarkUnschedulableForPodGroup, lastSchedulingCondition, requestedNodePools)

	isNodePoolAvailable, err := isNodePoolAvailableForAssignment(ctx, client, nodePoolAssignmentParams.NodePoolName)
	if err != nil {
		return nil, err
	}

	if isNodePoolAvailable {
		return nodePoolAssignmentParams, nil
	}

	// try to get the next ready-or-empty node pool from the options list
	return findNextNodePoolAvailableForScheduling(ctx, client, requestedNodePools, nodePoolAssignmentParams)
}

func findNextNodePoolAvailableForScheduling(
	ctx context.Context,
	client client.Client,
	requestedNodePools []string,
	nodePoolAssignmentParams *NodePoolAssignmentParams) (*NodePoolAssignmentParams, error) {
	previousNodePoolIndex := utils.IndexOfItemInList(requestedNodePools, nodePoolAssignmentParams.NodePoolName)

	for index := 1; index < len(requestedNodePools); index++ {
		nextNodePoolIndex := (previousNodePoolIndex + index) % len(requestedNodePools)
		nextNodePoolName := requestedNodePools[nextNodePoolIndex]
		log.Ctx(ctx).Debug().Msgf("Trying next node pool: <%s>, from the options list: <%v>",
			nextNodePoolName, requestedNodePools)

		isNodePoolAvailable, err := isNodePoolAvailableForAssignment(ctx, client, nextNodePoolName)
		if err != nil {
			return nil, err
		}

		if !isNodePoolAvailable {
			continue
		}

		if !nodePoolAssignmentParams.MarkUnschedulable && (nextNodePoolIndex == len(requestedNodePools)-1 || nextNodePoolIndex < previousNodePoolIndex) {
			// if originally markUnschedulable was 'false', but now it's a different node pool from the list,
			// and it is the last scheduling option - so now markUnschedulable should be 'true'
			nodePoolAssignmentParams.MarkUnschedulable = true
		}

		nodePoolAssignmentParams.NodePoolName = nextNodePoolName

		return nodePoolAssignmentParams, nil
	}

	err := fmt.Errorf("failed finding node pool for assigning - all node pools from options are non available: <%v>", requestedNodePools)
	log.Ctx(ctx).Error().Msgf("%s", err.Error())

	return nil, err
}

func getInitialNodePoolAssignmentParams(
	currentMarkUnschedulableForPodGroup *bool,
	lastSchedulingCondition *kaiv2alpha2.SchedulingCondition,
	requestedNodePools []string) *NodePoolAssignmentParams {
	schedulingBackoff := getSchedulingBackoff(requestedNodePools)

	if lastSchedulingCondition == nil {
		// this is either:
		// the first NodePool assignment for this PodGroup;
		// or only one nodepool from the list is available and the podgroup is "staying" in one nodepool, but without SchedulingBackoff -1.
		markUnschedulable := len(requestedNodePools) == 1
		if !markUnschedulable && currentMarkUnschedulableForPodGroup != nil && *currentMarkUnschedulableForPodGroup {
			markUnschedulable = true
		}

		return &NodePoolAssignmentParams{
			NodePoolName:      requestedNodePools[0],
			MarkUnschedulable: markUnschedulable,
			SchedulingBackoff: schedulingBackoff,
		}
	}

	currentMarkUnschedulableForPodGroupVal := common.GetMarkUnschedulableValue(currentMarkUnschedulableForPodGroup)

	previousNodePool := lastSchedulingCondition.NodePool
	previousNodePoolIndex := utils.IndexOfItemInList(requestedNodePools, previousNodePool)

	nextNodePoolIndex := (previousNodePoolIndex + 1) % len(requestedNodePools) // round-robin
	markUnschedulable := currentMarkUnschedulableForPodGroupVal || nextNodePoolIndex == len(requestedNodePools)-1

	return &NodePoolAssignmentParams{
		NodePoolName:      requestedNodePools[nextNodePoolIndex],
		MarkUnschedulable: markUnschedulable,
		SchedulingBackoff: schedulingBackoff,
	}
}

func getSchedulingBackoff(requestedNodePools []string) int32 {
	if len(requestedNodePools) == 1 {
		return common.NoSchedulingBackoff
	}

	return common.SingleSchedulingBackoff
}

func isNodePoolAvailableForAssignment(ctx context.Context, client client.Client, nodePoolName string) (bool, error) {
	isAvailableForScheduling, phase, err := utils.IsNodePoolAvailableForScheduling(ctx, client, nodePoolName)
	if err != nil {
		return false, err
	}

	if !isAvailableForScheduling {
		log.Ctx(ctx).Warn().Msgf("Warning: node pool <%s> is in phase <%v>; will not try to assign to it",
			nodePoolName, phase)
	}

	return isAvailableForScheduling, nil
}
