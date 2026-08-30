package nodepool_controller

import (
	"context"
	"errors"
	"fmt"

	kaiv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"
	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/runai-cluster/cluster/nodepool-controller/pkg/common"
	"github.com/run-ai/runai/runai-cluster/cluster/nodepool-controller/pkg/config"
	"github.com/run-ai/runai/runai-cluster/cluster/nodepool-controller/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (npc *NodePoolController) validateNodeStatusWhenChangingNodePool(ctx context.Context, node *corev1.Node, oldNodePoolName string) (schedulable bool, err error) {
	arePodsRunningWithSameNodePoolOnNode, err := npc.arePodsRunningWithSameNodePool(ctx, node.Name, oldNodePoolName)
	if err != nil {
		return false, err
	}

	if !arePodsRunningWithSameNodePoolOnNode {
		err = npc.updateNodeUnschedulableIfNeeded(ctx, node, false)
		if err != nil {
			return false, err
		}
		return true, nil
	}

	log.Warn().Msgf("Node <%v>: cannot change nodepool from <%v> - still has running pods with same nodepool - node unschedulable",
		node.Name, oldNodePoolName)

	// error handling at this point will be useless - need to return false anyway.
	err = npc.updateNodeUnschedulableIfNeeded(ctx, node, true)
	return false, err
}

func (npc *NodePoolController) validateNodeStatusInNodePool(ctx context.Context, node *corev1.Node, nodesNodePoolName string) (schedulable bool, err error) {
	arePodsRunningWithDifferentNodePoolOnNode, err := npc.arePodsRunningWithDifferentNodePool(ctx, node.Name, nodesNodePoolName)
	if err != nil {
		return false, err
	}

	if !arePodsRunningWithDifferentNodePoolOnNode {
		err = npc.updateNodeUnschedulableIfNeeded(ctx, node, false)
		if err != nil {
			return false, err
		}
		return true, nil
	}

	log.Info().Msgf("Found running pods with wrong nodepool (not <%v>) on node <%v> - node unschedulable",
		nodesNodePoolName, node.Name)

	err = npc.updateNodeUnschedulableIfNeeded(ctx, node, true)
	return false, err
}

func (npc *NodePoolController) validateNodeStatusInNodePoolWrapper(ctx context.Context, node *corev1.Node, nodesNodePoolName string) (err error) {
	schedulable, err := npc.validateNodeStatusInNodePool(ctx, node, nodesNodePoolName)
	if err != nil {
		log.Error().Msgf("Failed validating node status <%v> in node pool <%v>, err: %v",
			node.Name, nodesNodePoolName, err.Error())
		return
	}

	if !schedulable {
		unschedulableErr := errors.New(fmt.Sprintf("Node <%v> is unschedulable", node.Name))
		err = utils.AppendErrIfNotNil(err, unschedulableErr)
	}

	return
}

func (npc *NodePoolController) updateNodeUnschedulableIfNeeded(ctx context.Context, node *corev1.Node, unschedulable bool) (err error) {
	if node.Spec.Unschedulable == unschedulable {
		err = npc.validateNodeLabeledWithUnschedulable(ctx, node, unschedulable)
		if err != nil {
			log.Error().Msgf("Failed labeling node <%v> with %s=%t, err: %v",
				node.Name, config.Get().UnschedulableLabelKey, unschedulable, err.Error())
		}
		return nil
	}

	if !unschedulable {
		_, found := node.Labels[config.Get().UnschedulableLabelKey]
		if !found {
			log.Info().Msgf("No need to mark node as Unschedulable:false; " +
				"it was not marked as Unschedulable by us")
			return nil
		}
	}

	err = npc.patchNodeUnschedulable(ctx, node, unschedulable)
	if err != nil {
		log.Error().Msgf("Failed updating node <%v> with unschedulable <%t>, err: %v",
			node.Name, unschedulable, err.Error())
		return err
	}

	log.Info().Msgf("Successfully updated node <%v> with unschedulable <%t>",
		node.Name, unschedulable)
	return nil
}

func (npc *NodePoolController) arePodsRunningWithSameNodePool(ctx context.Context, nodeName, nodePoolName string) (bool, error) {
	return npc.arePodsRunningWithNodePool(ctx, nodeName, nodePoolName, true)
}

func (npc *NodePoolController) arePodsRunningWithDifferentNodePool(ctx context.Context, nodeName, nodePoolName string) (bool, error) {
	return npc.arePodsRunningWithNodePool(ctx, nodeName, nodePoolName, false)
}

func (npc *NodePoolController) arePodsRunningWithNodePool(ctx context.Context, nodeName, nodePoolName string, sameNodePool bool) (bool, error) {
	pods, err := npc.getRunningPodsWithNodePoolOnNode(ctx, nodeName, nodePoolName, sameNodePool)
	if err != nil {
		return true, err
	}

	nodePoolMessage := "nodepool"
	if !sameNodePool {
		nodePoolMessage = "a different nodepool than"
	}

	if len(pods.Items) > 0 {
		pod := pods.Items[0]
		log.Info().Msgf("Found running pod <%v/%v> on node <%v> with %v <%v>",
			pod.Namespace, pod.Name, nodeName, nodePoolMessage, nodePoolName)
		return true, nil
	}

	log.Debug().Msgf("Didn't find any running pods on node <%v> with %v <%v>",
		nodeName, nodePoolMessage, nodePoolName)
	return false, nil
}

func (npc *NodePoolController) getRunningPodsWithNodePoolOnNode(ctx context.Context, nodeName, nodePoolName string, sameNodePool bool) (*corev1.PodList, error) {
	// this field selector filters in only running pods with runai-scheduler,
	// and indexes the nodeName to be matched
	runningWithRunaiSchedulerNodeNameFieldSelector :=
		fields.OneTermEqualSelector(common.PodRunningWithRunaiSchedulerNodeNameField, nodeName)

	var err error
	var matchNodePool *labels.Requirement
	if sameNodePool {
		matchNodePool, err = utils.GetNodePoolLabelSelector(nodePoolName)
	} else {
		matchNodePool, err = utils.GetNotNodePoolLabelSelector(nodePoolName)
	}
	if err != nil {
		return nil, err
	}

	pods, err := npc.listPodsWithSelectors(ctx, runningWithRunaiSchedulerNodeNameFieldSelector, *matchNodePool)
	if err != nil {
		return nil, err
	}

	filteredPods := npc.filterPodsWithIncorrectNodepoolLabel(ctx, pods)

	return filteredPods, nil
}

func (npc *NodePoolController) filterPodsWithIncorrectNodepoolLabel(ctx context.Context, pods *corev1.PodList) *corev1.PodList {
	// Organize pods by pod group name
	podsByGroup := make(map[string][]corev1.Pod)
	for _, pod := range pods.Items {
		podGroupName, found := pod.Annotations[config.PodGroupAnnotationForPod]
		if !found {
			continue
		}
		podsByGroup[podGroupName] = append(podsByGroup[podGroupName], pod)
	}

	// Validate each pod group
	filteredPods := &corev1.PodList{}
	for podGroupName, groupPods := range podsByGroup {
		namespace := groupPods[0].Namespace
		// Get pod group for this group of pods
		podGroup := &kaiv2alpha2.PodGroup{}
		if err := npc.Client.Get(ctx, types.NamespacedName{Name: podGroupName, Namespace: namespace}, podGroup); err != nil {
			log.Error().Msgf("Failed to get pod group <%s>: namespace: <%s>, %v", podGroupName, namespace, err)
			continue
		}
		// Get pod group's node pool label
		nodePoolLabelKey := config.Get().NodePoolNameLabel
		groupNodePool, podGroupLabelFound := podGroup.Labels[nodePoolLabelKey]

		// Validate each pod in the group
		for _, pod := range groupPods {
			podNodePool, podLabelFound := pod.Labels[nodePoolLabelKey]
			// If node pool labels match, add pod to filtered list
			if podLabelFound == podGroupLabelFound && podNodePool == groupNodePool {
				filteredPods.Items = append(filteredPods.Items, pod)
			}
		}
	}
	return filteredPods
}

func (npc *NodePoolController) listPodsWithSelectors(ctx context.Context, fieldSelector fields.Selector, labelRequirements ...labels.Requirement) (
	pods *corev1.PodList, err error) {
	labelSelector := labels.NewSelector()
	labelSelector = labelSelector.Add(labelRequirements...)

	pods = &corev1.PodList{}
	err = npc.Client.List(
		ctx,
		pods,
		&client.ListOptions{LabelSelector: labelSelector, FieldSelector: fieldSelector})
	if err != nil {
		log.Error().Msgf("Failed listing pods with field selector <%v> and label selector <%v>, err: %v",
			fieldSelector.String(), labelSelector.String(), err.Error())
		return nil, err
	}

	return pods, nil
}

func isNodeReady(node *corev1.Node) bool {
	if node.Spec.Unschedulable {
		log.Debug().Msgf("Node <%v> is Unschedulable - not ready", node.Name)
		return false
	}

	for _, condition := range node.Status.Conditions {
		if condition.Status == corev1.ConditionTrue && condition.Type == corev1.NodeReady {
			log.Debug().Msgf("Node <%v> has 'Ready' condition - ready", node.Name)
			return true
		}
	}
	log.Debug().Msgf("Node <%v> doesn't have 'Ready' condition - not ready", node.Name)
	return false
}
