package nodepool_controller

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
	"golang.org/x/exp/slices"

	"github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/nodepool-controller/pkg/common"
	"github.com/run-ai/runai/runai-cluster/cluster/nodepool-controller/pkg/config"
	"github.com/run-ai/runai/runai-cluster/cluster/nodepool-controller/pkg/utils"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const MNNVLLabel = v1alpha1.DefaultGPUNetworkAccelerationLabelKey

func (npc *NodePoolController) updateNodePoolStatusIfNeeded(ctx context.Context,
	nodePoolWithUpdatedStatus *v1alpha1.NodePool, nodePool *v1alpha1.NodePool) error {
	if !isStatusEqual(&nodePoolWithUpdatedStatus.Status, &nodePool.Status) {
		return npc.updateNodePoolStatus(ctx, nodePoolWithUpdatedStatus)
	}
	return nil
}

func (npc *NodePoolController) updateNodePoolStatus(ctx context.Context, nodePool *v1alpha1.NodePool) error {
	patchBytes, err := updateNodePoolStatusPatchBytes(nodePool.Status)
	if err != nil {
		log.Error().Msgf("Failed to json.Marshal patch - update status of nodepool <%v>, err: %v", nodePool.Name, err)
		return err
	}

	patch := client.RawPatch(types.MergePatchType, patchBytes)
	err = npc.Client.Status().Patch(ctx, nodePool, patch)
	if err != nil {
		log.Error().Msgf("Failed updating nodepool <%v> with status <%v/%v>, err: %v",
			nodePool.Name, nodePool.Status.Phase, nodePool.Status.Message, err.Error())
		return err
	}

	log.Info().Msgf("Updated nodepool <%v> with status <%v/%v>",
		nodePool.Name, nodePool.Status.Phase, nodePool.Status.Message)
	return nil
}

// param of alteredNodes (nodePoolNodes) - nodes could be altered during previous operations and
// the cache is sometimes not updated fast enough
func (npc *NodePoolController) calculateNodePoolStatus(ctx context.Context,
	nodePool *v1alpha1.NodePool, nodePoolNodes []*corev1.Node) error {
	schedulerReady, err := npc.getNodePoolStatusByScheduler(ctx, nodePool)
	if err != nil {
		log.Error().Msgf("Failed getting scheduler status for nodepool <%v>, err: %v", nodePool.Name, err)
		nodePool.Status.Phase = v1alpha1.NodePoolUnschedulable
		nodePool.Status.Message = common.SchedulerNotReadyMessage
		npc.getNodesStatusForNodePool(nodePool, nodePoolNodes)
		return nil
	}

	if !schedulerReady {
		npc.getNodesStatusForNodePool(nodePool, nodePoolNodes)
		return nil
	}

	err = npc.getNodePoolStatusByNodes(ctx, nodePool, nodePoolNodes)
	if err != nil {
		log.Error().Msgf("Failed getting status by nodes for nodepool <%v>, err: %v", nodePool.Name, err)
		return err
	}
	return nil
}

func hasMNNVLLabel(labels map[string]string, labelKey string) bool {
	for key, _ := range labels {
		if key == labelKey {
			return true
		}
	}
	return false
}

func GetMNNVLLabelOtDefault(gpuNetworkAccelerationLabelKey string) string {
	labelKey := MNNVLLabel
	if gpuNetworkAccelerationLabelKey != "" {
		labelKey = gpuNetworkAccelerationLabelKey
	}
	return labelKey
}

// Empty means detection is not configured (annotation absent) and is disabled.
func getGPUNetworkAccelerationDetection(nodePool *v1alpha1.NodePool) v1alpha1.GPUNetworkAccelerationDetection {
	return v1alpha1.GPUNetworkAccelerationDetection(nodePool.Annotations[v1alpha1.AnnotationGPUNetworkAccelerationDetection])
}

func getGPUNetworkAccelerationLabelKey(nodePool *v1alpha1.NodePool) string {
	return GetMNNVLLabelOtDefault(nodePool.Annotations[v1alpha1.AnnotationGPUNetworkAccelerationLabelKey])
}

func calculateGPUNetworkAccelerationDetected(nodePool *v1alpha1.NodePool, nodes []*corev1.Node) bool {
	labelKey := getGPUNetworkAccelerationLabelKey(nodePool)
	switch getGPUNetworkAccelerationDetection(nodePool) {
	case v1alpha1.UseGPUNetworkAcceleration:
		return true
	case v1alpha1.AutoGPUNetworkAccelerationDetection:
		for _, node := range nodes {
			if hasMNNVLLabel(node.Labels, labelKey) {
				return true
			}
		}
		return false
	default: // DontUse or unset
		return false
	}
}

// param of alteredNodes (nodePoolNodes) - nodes could be altered during previous operations and
// the cache is sometimes not updated fast enough
func (npc *NodePoolController) getNodePoolStatusByNodes(ctx context.Context,
	nodePool *v1alpha1.NodePool, nodePoolNodes []*corev1.Node) error {
	_, missingNrtPrerequisiteItems, nodesInNodePool, atLeastOneNodeReady,
		atLeastOneNodeUnschedulable := getNodesSituation(nodePoolNodes)

	nodePool.Status.Phase = nodePoolPhaseForNodeSituation(
		atLeastOneNodeReady,
		atLeastOneNodeUnschedulable,
		len(missingNrtPrerequisiteItems) > 0,
	)

	nodePool.Status.Message = joinNodePoolStatusMessages(
		npc.getFullStatusMessageForNodePool(ctx, nodePool, nodePoolNodes),
		nrtNodePoolPrereqMessage(missingNrtPrerequisiteItems),
	)
	setNodePoolNrtHealthCondition(nodePool, missingNrtPrerequisiteItems)

	mismatchNodes, mismatchErr := npc.getTopologyMismatchNodes(ctx, nodePool, nodePoolNodes)
	if mismatchErr == nil {
		// If multiple nodes have topology mismatches, add condition to nodepool
		npc.addOrUpdateNodePoolTopologyMismatchCondition(nodePool, mismatchNodes)

		npc.updateNodesTopologyMismatch(mismatchNodes, nodesInNodePool)
	}
	nodePool.Status.Nodes = nodesInNodePool
	return mismatchErr
}

// Updates annotations that reflect controller-computed state.
func (npc *NodePoolController) updateAnnotationStatusFields(ctx context.Context,
	nodePool *v1alpha1.NodePool, nodePoolNodes []*corev1.Node) error {
	if getGPUNetworkAccelerationDetection(nodePool) == "" {
		return nil
	}

	// Writes the detection result to the -detected annotation, only when detection is configured.
	detected := strconv.FormatBool(calculateGPUNetworkAccelerationDetected(nodePool, nodePoolNodes))
	if current, found := nodePool.Annotations[v1alpha1.AnnotationGPUNetworkAccelerationDetected]; found && current == detected {
		return nil
	}

	return npc.UpdateAnnotations(ctx, nodePool,
		map[string]any{v1alpha1.AnnotationGPUNetworkAccelerationDetected: detected})
}

func (npc *NodePoolController) updateNodesTopologyMismatch(mismatchNodes []*corev1.Node, nodesInNodePool []v1alpha1.NodeInNodePool) {
	nodeNames := map[string]bool{}
	for _, node := range mismatchNodes {
		nodeNames[node.Name] = true
	}
	for i := range nodesInNodePool {
		node := nodesInNodePool[i]
		_, exist := nodeNames[node.Name]
		node.TopologyMismatch = exist
		nodesInNodePool[i] = node
	}
}

// param of alteredNodes (nodePoolNodes) - nodes could be altered during previous operations and
// the cache is sometimes not updated fast enough.
func (npc *NodePoolController) getFullStatusMessageForNodePool(ctx context.Context,
	nodePool *v1alpha1.NodePool, nodePoolNodes []*corev1.Node) string {
	nodesAssignedToNodePoolWaitingForDrain, nodesUnschedulableByUs, nodesNotReady, err := npc.getNodesFullSituationForFullPhaseMessage(ctx, nodePool, nodePoolNodes)
	if err != nil {
		log.Error().Msgf("Failed getting nodes situation for nodepool <%s>, not setting phase message", nodePool.Name)
		return ""
	}

	phaseMessageParts := []string{}

	if len(nodesAssignedToNodePoolWaitingForDrain) > 0 {
		nodeNamesStr := strings.Join(nodesAssignedToNodePoolWaitingForDrain, ", ")
		phaseMessageParts = append(phaseMessageParts, fmt.Sprintf(common.NodesAssignedToThisNodePoolWaitingForDrainMessage, nodeNamesStr))
	}
	if len(nodesUnschedulableByUs) > 0 {
		nodeNamesStr := strings.Join(nodesUnschedulableByUs, ", ")
		phaseMessageParts = append(phaseMessageParts, fmt.Sprintf(common.NodesAssignedToDifferentNodePoolWaitingForDrainMessage, nodeNamesStr))
	}
	if len(nodesNotReady) > 0 {
		nodeNamesStr := strings.Join(nodesNotReady, ", ")
		phaseMessageParts = append(phaseMessageParts, fmt.Sprintf(common.NodesNotReadyMessage, nodeNamesStr))
	}

	result := strings.Join(phaseMessageParts, "\n\n")
	return result
}

// param of alteredNodes (nodePoolNodes) - nodes could be altered during previous operations and
// the cache is sometimes not updated fast enough
func (npc *NodePoolController) getNodePoolDeletionStatus(ctx context.Context,
	nodePool *v1alpha1.NodePool, nodePoolNodes []*corev1.Node) error {
	unschedulableNodes, _, nodesInNodePool, _, atLeastOneNodeUnschedulable := getNodesSituation(nodePoolNodes)

	referencingProjects, err := npc.getProjectsReferencingNodePool(ctx, nodePool.Name)
	if err != nil {
		return err
	}

	nodePoolPhase := v1alpha1.NodePoolEmpty
	if atLeastOneNodeUnschedulable || len(referencingProjects) > 0 {
		// we need all nodes ready and no project referencing the nodepool to declare it
		// as ready for deletion
		nodePoolPhase = v1alpha1.NodePoolDeleting
	}

	npc.addOrUpdateProjectsReferenceCondition(nodePool, referencingProjects)

	statusMessage := []string{}
	if drainingMessage := getNodePoolDeletingMessage(unschedulableNodes); drainingMessage != "" {
		statusMessage = append(statusMessage, drainingMessage)
	}

	if projectsMessage := getProjectsReferenceMessage(referencingProjects); projectsMessage != "" {
		statusMessage = append(statusMessage, projectsMessage)
	}

	nodePool.Status.Phase = nodePoolPhase
	nodePool.Status.Message = strings.Join(statusMessage, "\n")
	nodePool.Status.Nodes = nodesInNodePool
	return nil
}

func (npc *NodePoolController) getNodesStatusForNodePool(nodePool *v1alpha1.NodePool, nodePoolNodes []*corev1.Node) {
	_, _, nodesInNodePool, _, _ := getNodesSituation(nodePoolNodes)
	nodePool.Status.Nodes = nodesInNodePool
}

func (npc *NodePoolController) updateNodePoolStatusOnDeletion(ctx context.Context,
	nodePool *v1alpha1.NodePool, nodePoolNodes []*corev1.Node) (phase v1alpha1.NodePoolPhase, err error) {
	nodePoolCopy := nodePool.DeepCopy()
	innerErr := npc.getNodePoolDeletionStatus(ctx, nodePool, nodePoolNodes)
	err = utils.AppendErrIfNotNil(err, innerErr)

	phase = nodePool.Status.Phase
	innerErr = npc.updateNodePoolStatusIfNeeded(ctx, nodePool, nodePoolCopy)
	err = utils.AppendErrIfNotNil(err, innerErr)

	return
}

// param of alteredNodes (nodePoolNodes) - nodes could be altered during previous operations and
// the cache is sometimes not updated fast enough
func (npc *NodePoolController) updateNodePoolStatusOnDeletionError(ctx context.Context,
	nodePool *v1alpha1.NodePool, nodePoolNodes []*corev1.Node) error {
	nodePoolWithUpdatedStatus := nodePool.DeepCopy()
	nodePoolWithUpdatedStatus.Status.Phase = v1alpha1.NodePoolDeleting
	nodePoolWithUpdatedStatus.Status.Message = common.ErrorDeletingMessage
	npc.getNodesStatusForNodePool(nodePool, nodePoolNodes)

	return npc.updateNodePoolStatusIfNeeded(ctx, nodePoolWithUpdatedStatus, nodePool)
}

// param of alteredNodes (nodePoolNodes) - nodes could be altered during previous operations and
// the cache is sometimes not updated fast enough.
//
// in order to construct a full phase message for the nodepool, this function returns:
//   - nodes that are assigned to the nodepool and waiting for drain - not yet migrated to the nodepool.
//   - also, out of the nodes in the nodepool: which are unschedulable by us, and which are not ready/unschedulable not by us.
func (npc *NodePoolController) getNodesFullSituationForFullPhaseMessage(ctx context.Context,
	nodePool *v1alpha1.NodePool, nodePoolNodes []*corev1.Node) ([]string, []string, []string, error) {
	allNodesAssignedToNodePoolWaitingForDrain, err := npc.getNodesAssignedToNodePoolWaitingForDrain(ctx, nodePool)
	if err != nil {
		log.Error().Msgf("Failed listing nodes assigned to nodepool <%s> waiting for drain, error: %s", nodePool.Name, err.Error())
		return nil, nil, nil, err
	}

	nodesUnschedulableByUs := []string{}
	nodesNotReady := []string{}
	nodePoolNodeNames := []string{}
	for _, node := range nodePoolNodes {
		_, ok := node.Labels[config.Get().UnschedulableLabelKey]
		if ok && node.Spec.Unschedulable {
			nodesUnschedulableByUs = append(nodesUnschedulableByUs, node.Name)
		} else if !isNodeReady(node) {
			nodesNotReady = append(nodesNotReady, node.Name)
		}

		nodePoolNodeNames = append(nodePoolNodeNames, node.Name)
	}

	// validate all nodes of allNodesAssignedToNodePoolWaitingForDrain don't appear in nodePoolNodes -
	// since nodePoolNodes are the nodes that are part of the nodepool and that were previously altered;
	// while allNodesAssignedToNodePoolWaitingForDrain are listed from cache that is not always updated fast enough.
	nodesAssignedToNodePoolWaitingForDrain := utils.Difference(allNodesAssignedToNodePoolWaitingForDrain, nodePoolNodeNames)

	slices.Sort(nodesAssignedToNodePoolWaitingForDrain)
	slices.Sort(nodesUnschedulableByUs)
	slices.Sort(nodesNotReady)

	return nodesAssignedToNodePoolWaitingForDrain, nodesUnschedulableByUs, nodesNotReady, nil
}

func getNodesSituation(nodes []*corev1.Node) (
	unschedulableNodes, missingNrtPrerequisiteItems []string, nodesInNodePool []v1alpha1.NodeInNodePool,
	atLeastOneNodeReady, atLeastOneNodeUnschedulable bool) {
	atLeastOneNodeReady = false
	atLeastOneNodeUnschedulable = false
	missingNrtItems := map[string]bool{}

	for i := range nodes {
		node := nodes[i]
		nodeStatus := v1alpha1.NodeReady
		nrtHealthCondition := findNodeCondition(node, NrtHealthyConditionType)
		missingNrtPrerequisite := nrtHealthCondition != nil && nrtHealthCondition.Status == corev1.ConditionTrue
		if missingNrtPrerequisite {
			if missingItem := nrtNodePoolMissingItemForReason(nrtHealthCondition.Reason); missingItem != "" {
				missingNrtItems[missingItem] = true
			}
		}

		if isNodeReady(node) {
			atLeastOneNodeReady = true
			if missingNrtPrerequisite {
				nodeStatus = v1alpha1.NodeMissingNrtHealthyPrerequisite
			}
		} else {
			unschedulableNodes = append(unschedulableNodes, node.Name)
			nodeStatus = v1alpha1.NodeUnschedulable
			atLeastOneNodeUnschedulable = true
		}

		nodeInNodePool := v1alpha1.NodeInNodePool{
			Name:   node.Name,
			Status: nodeStatus,
		}

		nodesInNodePool = append(nodesInNodePool, nodeInNodePool)
	}

	sortNodesInNodePool(nodesInNodePool)
	slices.Sort(unschedulableNodes)

	for missingItem := range missingNrtItems {
		missingNrtPrerequisiteItems = append(missingNrtPrerequisiteItems, missingItem)
	}
	slices.Sort(missingNrtPrerequisiteItems)

	return unschedulableNodes, missingNrtPrerequisiteItems, nodesInNodePool, atLeastOneNodeReady,
		atLeastOneNodeUnschedulable
}

func joinNodePoolStatusMessages(messages ...string) string {
	nonEmptyMessages := make([]string, 0, len(messages))
	for _, message := range messages {
		if message != "" {
			nonEmptyMessages = append(nonEmptyMessages, message)
		}
	}
	return strings.Join(nonEmptyMessages, "\n")
}

func nodePoolPhaseForNodeSituation(atLeastOneNodeReady, atLeastOneNodeUnschedulable, missingNrtPrerequisite bool) v1alpha1.NodePoolPhase {
	nodePoolPhase := v1alpha1.NodePoolEmpty
	if atLeastOneNodeReady {
		// we need at least one ready node to declare the nodepool as ready
		nodePoolPhase = v1alpha1.NodePoolReady
	} else if atLeastOneNodeUnschedulable {
		nodePoolPhase = v1alpha1.NodePoolUnschedulable
	}

	if missingNrtPrerequisite && nodePoolPhase != v1alpha1.NodePoolUnschedulable {
		return v1alpha1.NodePoolMissingPrerequisites
	}
	return nodePoolPhase
}

func getNodePoolUnschedulableMessage(unschedulableNodes []string) string {
	return getNodePoolStatusMessage(unschedulableNodes, common.UnschedulableNodesMessage)
}

func getNodePoolDeletingMessage(unschedulableNodes []string) string {
	return getNodePoolStatusMessage(unschedulableNodes, common.NodesDrainedMessage)
}

func getNodePoolStatusMessage(unschedulableNodes []string, messagePrefix string) string {
	if len(unschedulableNodes) == 0 {
		return ""
	}

	unschedulableNodesStr := strings.Join(unschedulableNodes, ", ")
	return fmt.Sprintf("%v: %v", messagePrefix, unschedulableNodesStr)
}

func updateNodePoolStatusPatchBytes(nodePoolStatus v1alpha1.NodePoolStatus) ([]byte, error) {
	status := map[string]any{
		"phase":      nodePoolStatus.Phase,
		"message":    nodePoolStatus.Message,
		"nodes":      nodePoolStatus.Nodes,
		"conditions": nodePoolStatus.Conditions,
	}
	statusMap := map[string]any{"status": status}
	return json.Marshal(statusMap)
}

func sortNodesInNodePool(nodes []v1alpha1.NodeInNodePool) {
	slices.SortFunc(nodes, func(i, j v1alpha1.NodeInNodePool) int {
		if i.Name < j.Name {
			return -1
		}
		if i.Name > j.Name {
			return 1
		}
		return 0
	})
}

func isStatusEqual(left, right *v1alpha1.NodePoolStatus) bool {
	if left.Phase != right.Phase ||
		left.Message != right.Message ||
		len(left.Nodes) != len(right.Nodes) ||
		len(left.Conditions) != len(right.Conditions) {
		return false
	}

	conditionsEqual := isNodePoolConditionsEqual(left.Conditions, right.Conditions)
	if !conditionsEqual {
		return false
	}

	nodesInNodepoolEqual := isNodeInNodepoolEqual(left.Nodes, right.Nodes)
	if !nodesInNodepoolEqual {
		return false
	}

	return true
}

// isNodeInNodepoolEqual compares two arrays of NodeInNodePool structs.
// It returns true if any NodeInNodePool status changes.
func isNodeInNodepoolEqual(left []v1alpha1.NodeInNodePool, right []v1alpha1.NodeInNodePool) bool {
	if len(left) != len(right) {
		return false
	}
	leftMap := make(map[string]v1alpha1.NodeInNodePool)
	for _, nodeStatus := range left {
		leftMap[nodeStatus.Name] = nodeStatus
	}

	for _, rightNodeStatus := range right {
		if leftNodeStatus, ok := leftMap[rightNodeStatus.Name]; !ok ||
			leftNodeStatus.Status != rightNodeStatus.Status ||
			leftNodeStatus.TopologyMismatch != rightNodeStatus.TopologyMismatch {
			return false
		}
	}
	return true
}

// isNodePoolConditionsEqual compares two arrays of NodePoolCondition structs.
// Conditions are compared by type, status, reason and message; timestamps are ignored
// since conditions are rebuilt with a fresh LastTransitionTime on every reconcile.
func isNodePoolConditionsEqual(left []v1alpha1.NodePoolCondition, right []v1alpha1.NodePoolCondition) bool {
	if len(left) != len(right) {
		return false
	}
	leftMap := make(map[v1alpha1.NodePoolConditionType]v1alpha1.NodePoolCondition)
	for _, condition := range left {
		leftMap[condition.Type] = condition
	}

	for _, rightCondition := range right {
		if leftCondition, ok := leftMap[rightCondition.Type]; !ok ||
			leftCondition.Status != rightCondition.Status ||
			leftCondition.Reason != rightCondition.Reason ||
			leftCondition.Message != rightCondition.Message {
			return false
		}
	}
	return true
}
