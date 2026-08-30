package assigner

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	kaiv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"
	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/runai-cluster/cluster/pod-group-assigner/pkg/config"
	"github.com/run-ai/runai/runai-cluster/cluster/pod-group-assigner/pkg/controllers/common"
	"github.com/run-ai/runai/runai-cluster/cluster/pod-group-assigner/pkg/controllers/podgroup/assignment_params"
	"github.com/run-ai/runai/runai-cluster/cluster/pod-group-assigner/pkg/controllers/podgroup/requested_nodepools_converter"
	"github.com/run-ai/runai/runai-cluster/cluster/pod-group-assigner/pkg/controllers/utils"
	"github.com/run-ai/runai/runai-cluster/common/node-pool-utils/converter"
	nodepoolutils "github.com/run-ai/runai/runai-cluster/common/node-pool-utils/utils"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type PodGroupAssigner struct {
	Client                      client.Client
	podGroupsNodePoolList       map[types.UID][]string
	requestedNodePoolsConverter *requested_nodepools_converter.RequestedNodePoolsConverter
}

type FullNodePoolAssignmentParams struct {
	Queue string
	assignment_params.NodePoolAssignmentParams
	NetworkTopologyConstraint    kaiv2alpha2.TopologyConstraint
	NetworkTopologySystemSourced bool
}

func NewPodGroupAssigner(cachedClient client.Client) *PodGroupAssigner {
	identifiers := converter.NodePoolIdentifiers{
		NodePoolAssignmentLabelKey: config.Config().NodePoolLabelKey,
		DefaultNodepoolName:        config.Config().DefaultNodepoolName,
		UnexistingNodepoolSentinel: config.Config().UnexistingNodepoolSentinel,
		AnnotationNodepoolsKey:     config.Config().AnnotationNodepoolsKey,
	}
	requestedNodePoolsConverter := requested_nodepools_converter.NewRequestedNodePoolsConverter(cachedClient, identifiers)

	return &PodGroupAssigner{
		Client:                      cachedClient,
		requestedNodePoolsConverter: requestedNodePoolsConverter,
		podGroupsNodePoolList:       make(map[types.UID][]string),
	}
}

func (pga *PodGroupAssigner) Run(ctx context.Context, podGroup *kaiv2alpha2.PodGroup) error {
	if podGroup.Labels == nil {
		podGroup.Labels = map[string]string{}
	}

	isUpForAssignment, lastSchedulingCondition, err := pga.isEligibleForNodePoolAssignment(ctx, podGroup)
	if err != nil {
		return err
	}

	if !isUpForAssignment {
		return nil
	}

	podGroupPods, err := pga.getPodGroupPods(ctx, podGroup.ObjectMeta)
	if err != nil {
		log.Ctx(ctx).Error().Msgf("Failed getting pods for pod group <%s>, err: <%s>",
			getPodGroupNamespacedName(podGroup.ObjectMeta), err.Error())

		return err
	}

	if podGroupPods == nil || len(podGroupPods.Items) == 0 {
		err = fmt.Errorf("didn't find any pods for pod group <%s>", getPodGroupNamespacedName(podGroup.ObjectMeta))
		return err
	}

	if !isEligibleForAssignmentByPods(ctx, podGroup.ObjectMeta, podGroupPods) {
		return nil
	}

	return pga.handleNodePoolAssignment(ctx, podGroup, lastSchedulingCondition, podGroupPods)
}

func (pga *PodGroupAssigner) isEligibleForNodePoolAssignment(ctx context.Context, podGroup *kaiv2alpha2.PodGroup) (
	isUpForAssignment bool, lastSchedulingCondition *kaiv2alpha2.SchedulingCondition, err error) {
	if common.GetSchedulingBackoffValue(podGroup.Spec.SchedulingBackoff) == common.NoSchedulingBackoff {
		log.Ctx(ctx).Debug().Msgf("Pod Group <%s> is not eligible for node pool assignment; scheduling backoff is infinite",
			getPodGroupNamespacedName(podGroup.ObjectMeta))

		return false, nil, nil
	}

	lastSchedulingCondition = common.GetLastSchedulingCondition(podGroup)
	if lastSchedulingCondition == nil {
		return true, nil, nil
	}

	currentNodePoolName := nodepoolutils.GetNodePoolNameFromLabels(podGroup.Labels, config.Config().NodePoolLabelKey, config.Config().DefaultNodepoolName)
	if lastSchedulingCondition.NodePool != currentNodePoolName {
		isNodePoolAvailable, err := pga.isNodePoolAvailableForSchedulingPodGroup(ctx, currentNodePoolName, podGroup.Name, podGroup.Namespace)
		if err != nil || !isNodePoolAvailable {
			return true, lastSchedulingCondition, err
		}

		log.Ctx(ctx).Debug().Msgf("Pod Group <%s> is not eligible for node pool assignment; current node pool scheduler in progress",
			getPodGroupNamespacedName(podGroup.ObjectMeta))

		return false, lastSchedulingCondition, nil
	}

	// if the last scheduling condition was placed by the current node pool -
	// it means the scheduler is done with the pod group, so we should take it and try to re-assign - so it is up for assignment
	return true, lastSchedulingCondition, nil
}

func (pga *PodGroupAssigner) handleNodePoolAssignment(
	ctx context.Context,
	podGroup *kaiv2alpha2.PodGroup,
	lastSchedulingCondition *kaiv2alpha2.SchedulingCondition,
	podGroupPods *corev1.PodList) error {
	log.Ctx(ctx).Info().Msgf("Handling pod group assignment <%s>",
		getPodGroupNamespacedName(podGroup.ObjectMeta))

	requestedNodePools, err := pga.getRequestedNodePoolsForPodGroup(ctx, podGroup, podGroupPods)
	if err != nil {
		log.Ctx(ctx).Error().Msgf("Failed assigning pod group <%s> to a node pool, error: %s",
			getPodGroupNamespacedName(podGroup.ObjectMeta), err.Error())

		return err
	}

	nodePoolAssignmentParams, err := assignment_params.GetNodePoolAssignmentParams(ctx, pga.Client, podGroup.Spec.MarkUnschedulable, lastSchedulingCondition, requestedNodePools)
	if err != nil {
		log.Ctx(ctx).Error().Msgf("Failed assigning pod group <%s> to a node pool, error: %s",
			getPodGroupNamespacedName(podGroup.ObjectMeta), err.Error())

		return err
	}

	return pga.assignToNodePool(ctx, podGroup, nodePoolAssignmentParams, podGroupPods)
}

func (pga *PodGroupAssigner) assignToNodePool(
	ctx context.Context,
	podGroup *kaiv2alpha2.PodGroup,
	nodePoolAssignmentParams *assignment_params.NodePoolAssignmentParams,
	podGroupPods *corev1.PodList) error {
	podGroupQueue, err := pga.calculatePodGroupQueueNameOfNodePool(ctx, podGroup, nodePoolAssignmentParams.NodePoolName)
	if err != nil {
		log.Ctx(ctx).Error().Msgf("Failed assigning pod group <%s> to: <%+v> - failed getting queue name for pod group, error: %s",
			getPodGroupNamespacedName(podGroup.ObjectMeta), nodePoolAssignmentParams, err.Error())

		return err
	}

	nextNodePoolAssignmentParams := &FullNodePoolAssignmentParams{
		NodePoolAssignmentParams: *nodePoolAssignmentParams,
		Queue:                    podGroupQueue,
	}

	log.Ctx(ctx).Debug().Msgf("Assigning pod group <%s> to: <%+v>",
		getPodGroupNamespacedName(podGroup.ObjectMeta), nextNodePoolAssignmentParams)

	err = pga.updatePodGroupWithAssignment(ctx, podGroup, nextNodePoolAssignmentParams, podGroupPods)
	if err != nil {
		log.Ctx(ctx).Error().Msgf("Failed assigning pod group <%s> to: <%+v> - failed updating pod group object, error: %s",
			getPodGroupNamespacedName(podGroup.ObjectMeta), nextNodePoolAssignmentParams, err.Error())

		return err
	}

	return nil
}

func (pga *PodGroupAssigner) updatePodGroupWithAssignment(
	ctx context.Context,
	podGroup *kaiv2alpha2.PodGroup,
	next *FullNodePoolAssignmentParams,
	podGroupPods *corev1.PodList) error {
	if next.RemoveUnschedulableOnNodePoolSchedulingCondition {
		err := pga.removeUnschedulableOnNodePoolSchedulingConditionFromStatus(ctx, podGroup)
		if err != nil {
			log.Ctx(ctx).Error().Msgf("Failed assigning pod group <%s> to: <%+v> - failed removing UnschedulableOnNodePool scheduling condition, error: %s",
				getPodGroupNamespacedName(podGroup.ObjectMeta), next, err.Error())

			return err
		}
		// continue - maybe other assignment params need to be updated (for example, MarkUnschedulable)
	}

	previous := &FullNodePoolAssignmentParams{
		NodePoolAssignmentParams: assignment_params.NodePoolAssignmentParams{
			NodePoolName:      nodepoolutils.GetNodePoolNameFromLabels(podGroup.Labels, config.Config().NodePoolLabelKey, config.Config().DefaultNodepoolName),
			MarkUnschedulable: common.GetMarkUnschedulableValue(podGroup.Spec.MarkUnschedulable),
			SchedulingBackoff: common.GetSchedulingBackoffValue(podGroup.Spec.SchedulingBackoff),
			RemoveUnschedulableOnNodePoolSchedulingCondition: next.RemoveUnschedulableOnNodePoolSchedulingCondition,
		},
		Queue:                        podGroup.Spec.Queue,
		NetworkTopologyConstraint:    podGroup.Spec.TopologyConstraint,
		NetworkTopologySystemSourced: podGroup.Annotations[TopologySourceAnnotationKey] == TopologySourceSystem,
	}

	desiredTopology, topologySystemSourced, err := pga.desiredNetworkTopology(ctx, podGroup, next.NodePoolName)
	if err != nil {
		log.Ctx(ctx).Error().Msgf("Failed assigning pod group <%s> to: <%+v> - failed resolving default network topology, error: %s",
			getPodGroupNamespacedName(podGroup.ObjectMeta), next, err.Error())

		return err
	}
	next.NetworkTopologyConstraint = desiredTopology
	next.NetworkTopologySystemSourced = topologySystemSourced

	if reflect.DeepEqual(previous, next) {
		log.Ctx(ctx).Debug().Msgf("Pod group <%s> is already assigned to node pool <%s> correctly",
			getPodGroupNamespacedName(podGroup.ObjectMeta), next.NodePoolName)

		return nil
	}

	originalPodGroup := podGroup.DeepCopy()
	pga.mutatePodGroupObjForAssignment(podGroup, next)

	err = pga.Client.Patch(ctx, podGroup, client.MergeFrom(originalPodGroup))
	if err == nil {
		log.Ctx(ctx).Info().Msgf("Successfully assigned pod group <%s> to: <%+v>",
			getPodGroupNamespacedName(podGroup.ObjectMeta), next)

		return nil
	}

	return err
}

func (pga *PodGroupAssigner) mutatePodGroupObjForAssignment(podGroup *kaiv2alpha2.PodGroup, next *FullNodePoolAssignmentParams) {
	podGroup.Spec.MarkUnschedulable = ptr.To(next.MarkUnschedulable)
	podGroup.Spec.SchedulingBackoff = ptr.To(next.SchedulingBackoff)
	podGroup.Spec.Queue = next.Queue
	podGroup.Labels[config.Config().QueueLabelKey] = next.Queue

	common.UpdateLabelsWithNodePoolAssignment(podGroup.Labels, next.NodePoolName, config.Config().NodePoolLabelKey)
	applyTopologyToPodGroup(podGroup, next.NetworkTopologyConstraint, next.NetworkTopologySystemSourced)
}

func (pga *PodGroupAssigner) removeUnschedulableOnNodePoolSchedulingConditionFromStatus(ctx context.Context, podGroup *kaiv2alpha2.PodGroup) error {
	log.Ctx(ctx).Debug().Msgf("Removing UnschedulableOnNodePool scheduling condition from pod group <%s>",
		getPodGroupNamespacedName(podGroup.ObjectMeta))

	newStatus := podGroup.Status.DeepCopy()
	newStatus.SchedulingConditions = []kaiv2alpha2.SchedulingCondition{}

	for _, schedulingCondition := range podGroup.Status.SchedulingConditions {
		if schedulingCondition.Type != kaiv2alpha2.UnschedulableOnNodePool {
			newStatus.SchedulingConditions = append(newStatus.SchedulingConditions, schedulingCondition)
		}
	}

	return pga.patchPodGroupSchedulingConditionsStatus(ctx, podGroup, *newStatus)
}

func (pga *PodGroupAssigner) patchPodGroupSchedulingConditionsStatus(ctx context.Context, podGroup *kaiv2alpha2.PodGroup, status kaiv2alpha2.PodGroupStatus) error {
	patchBytes, err := updateSchedulingConditionsPatchBytes(status.SchedulingConditions)
	if err != nil {
		log.Error().Msgf("Failed to json.Marshal patch - update SchedulingConditions status of pod group <%s>, error: %s",
			getPodGroupNamespacedName(podGroup.ObjectMeta), err.Error())

		return err
	}

	patch := client.RawPatch(types.MergePatchType, patchBytes)

	err = pga.Client.Status().Patch(ctx, podGroup, patch)
	if err != nil {
		log.Error().Msgf("Failed updating pod group <%s> with SchedulingConditions status <%v>, error: %s",
			getPodGroupNamespacedName(podGroup.ObjectMeta), status.SchedulingConditions, err.Error())

		return err
	}

	return nil
}

func (pga *PodGroupAssigner) getRequestedNodePoolsForPodGroup(ctx context.Context, podGroup *kaiv2alpha2.PodGroup, podGroupPods *corev1.PodList) ([]string, error) {
	if requestedNodePools, found := pga.podGroupsNodePoolList[podGroup.UID]; found {
		log.Ctx(ctx).Debug().Msgf("For pod group <%s>, got node pool list from cache: <%v>",
			getPodGroupNamespacedName(podGroup.ObjectMeta), requestedNodePools)

		return requestedNodePools, nil
	}

	requestedNodePools, err := pga.requestedNodePoolsConverter.GetRequestedNodePoolsForPodGroup(ctx, podGroup.ObjectMeta, podGroupPods)
	if err != nil {
		return []string{}, err
	}

	pga.podGroupsNodePoolList[podGroup.UID] = requestedNodePools

	return requestedNodePools, nil
}

func (pga *PodGroupAssigner) getPodGroupPods(ctx context.Context, podGroupMeta metav1.ObjectMeta) (
	*corev1.PodList, error) {
	podsByPodGroupFieldSelector :=
		fields.OneTermEqualSelector(common.PodByPodGroupIndexerName, podGroupMeta.Name)

	pods := &corev1.PodList{}

	err := pga.Client.List(
		ctx,
		pods,
		&client.ListOptions{FieldSelector: podsByPodGroupFieldSelector})
	if err != nil {
		log.Ctx(ctx).Error().Msgf("Failed listing pods with field selector <%s>, err: <%s>",
			podsByPodGroupFieldSelector.String(), err.Error())

		return nil, err
	}

	return pods, nil
}

func (pga *PodGroupAssigner) isNodePoolAvailableForSchedulingPodGroup(ctx context.Context, nodePoolName, podGroupName, podGroupNamespace string) (bool, error) {
	isAvailableForScheduling, phase, err := utils.IsNodePoolAvailableForScheduling(ctx, pga.Client, nodePoolName)
	if err != nil {
		return false, err
	}

	if !isAvailableForScheduling {
		log.Ctx(ctx).Warn().Msgf("Warning: node pool <%s> is in phase <%v>; pod group <%s/%s> is unschedulable on this node pool",
			nodePoolName, phase, podGroupNamespace, podGroupName)
	}

	return isAvailableForScheduling, nil
}

func (pga *PodGroupAssigner) calculatePodGroupQueueNameOfNodePool(ctx context.Context, podGroup *kaiv2alpha2.PodGroup, nodePoolName string) (string, error) {
	var err error

	projectName, found := podGroup.Labels[config.Config().ProjectLabelKey]
	if !found {
		projectName, err = pga.getProjectOfNamespace(ctx, podGroup.Namespace)
		if err != nil {
			err = fmt.Errorf("%s; for podgroup <%s>", err.Error(), getPodGroupNamespacedName(podGroup.ObjectMeta))
			return "", err
		}
	}

	return utils.GetQueueNameOfNodePool(ctx, pga.Client, projectName, nodePoolName)
}

func (pga *PodGroupAssigner) getProjectOfNamespace(ctx context.Context, namespace string) (string, error) {
	return utils.GetProjectNameOfNamespace(ctx, pga.Client, namespace)
}

// isEligibleForAssignmentByPods
// we need that in case of a Node Pool suddenly becoming Unschedulable,
// but all pods in a pod group are running - we don't want to re-assign the pod group..
func isEligibleForAssignmentByPods(ctx context.Context, podGroupMeta metav1.ObjectMeta, podGroupPods *corev1.PodList) bool {
	for _, pod := range podGroupPods.Items {
		if !isPodBoundToNode(pod.Status.Phase) {
			return true
		}
	}

	log.Ctx(ctx).Info().Msgf("Pod Group <%s> is not eligible for node pool assignment; all pods in pod group are Running",
		getPodGroupNamespacedName(podGroupMeta))

	return false
}

func isPodBoundToNode(podPhase corev1.PodPhase) bool {
	return podPhase == corev1.PodRunning || podPhase == corev1.PodSucceeded || podPhase == corev1.PodFailed
}

func getPodGroupNamespacedName(podGroupMeta metav1.ObjectMeta) string {
	return fmt.Sprintf("%s/%s", podGroupMeta.Namespace, podGroupMeta.Name)
}

func updateSchedulingConditionsPatchBytes(schedulingConditions []kaiv2alpha2.SchedulingCondition) ([]byte, error) {
	return json.Marshal(map[string]interface{}{"status": map[string]interface{}{
		"schedulingConditions": schedulingConditions}})
}
