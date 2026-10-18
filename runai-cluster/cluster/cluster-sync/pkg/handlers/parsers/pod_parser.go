/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package parsers

import (
	"context"
	"fmt"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"github.com/google/uuid"
	"k8s.io/utils/ptr"

	"go.uber.org/zap"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/cachedclient"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/util"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/resourceutils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/resourceutils/resources"
	topowner "github.com/dsx-ai-factory/run-ai/runai-cluster/common/top-owner"
	kaiv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const elementLabelPrefix = "run.ai/element-"

type PodParser struct {
	cachedClient cachedclient.CachedClient
}

func NewPodParser(cachedClient cachedclient.CachedClient) *PodParser {
	return &PodParser{
		cachedClient: cachedClient,
	}
}

func (p *PodParser) ParsePodToServerScheme(pod *corev1.Pod) (*openapi.PodSyncFields, error) {
	workloadId, err := p.GetPodWorkloadId(pod)
	if err != nil {
		return nil, err
	}

	projectId, err := util.GetProjectIdFromNamespaceName(pod.Namespace, p.cachedClient)
	if err != nil {
		zap.S().Errorf("Failed getting project id for pod <%s/%s>, error: %s",
			pod.Namespace, pod.Name, err.Error())
		return nil, err
	}

	requestedResources := p.parsePodRequestedResources(pod)
	phase := GetPodRunaiK8sPhase(pod)

	return &openapi.PodSyncFields{
		Name:               pod.Name,
		PriorityClassName:  p.getPriorityClassName(pod),
		Id:                 string(pod.UID),
		WorkloadId:         workloadId,
		RevisionId:         *openapi.NewNullableString(determineRevisionId(pod)),
		ClusterId:          config.Config().ClusterUuid,
		ProjectId:          &projectId,
		NodeName:           getApiString(pod.Spec.NodeName),
		CreatedAt:          pod.CreationTimestamp.Time,
		CompletedAt:        getApiTime(calculatePodCompletedAt(pod, phase)),
		Containers:         getContainers(pod),
		CurrentNodePool:    getApiString(getPodCurrentNodePool(pod.Labels, pod.Status.Phase)),
		RequestedNodePools: getPodRequestedNodePools(pod.ObjectMeta, pod.Spec.Affinity, p.cachedClient),
		RequestedResources: requestedResources,
		AllocatedResources: parsePodAllocatedResources(pod, requestedResources),
		Tolerations:        getPodTolerations(pod.Spec.Tolerations),
		K8sPhase:           phase,
		Ip:                 getApiString(pod.Status.PodIP),
		Task:               *openapi.NewNullableString(parsePodTask(pod)),
		NodeAffinity:       getPodNodeAffinity(pod.Spec),
		Elements:           parsePodElements(pod.Labels),
		ResourceClaims:     parsePodResourceClaims(pod),
	}, nil
}

func determineRevisionId(pod *corev1.Pod) (revisionId *string) {
	if value, exists := pod.Labels[constants.LabelRevisionId]; exists && value != "" {
		if _, err := uuid.Parse(value); err == nil {
			revisionId = &value
		} else {
			zap.S().Errorf("Failed to parse revision id <%s> for pod <%s/%s>, error: %s",
				value, pod.Namespace, pod.Name, err.Error())
		}
	}

	return
}

func (p *PodParser) GetPodWorkloadId(pod *corev1.Pod) (string, error) {
	return topowner.GetTopOwnerWorkloadId(pod, p.cachedClient)
}

func (p *PodParser) parsePodRequestedResources(pod *corev1.Pod) openapi.NullablePodRequestResources {
	requests, limits := resources.PodRunaiRequestsAndLimits(pod.Spec.Containers,
		pod.Spec.InitContainers, pod.Annotations)
	return parseRequestedResources(requests, limits)
}

func (p *PodParser) getPriorityClassName(pod *corev1.Pod) string {
	podGroup := p.getPodGroupOfPod(pod)
	if podGroup != nil && podGroup.Spec.PriorityClassName != "" {
		return podGroup.Spec.PriorityClassName
	}

	if pod.Spec.PriorityClassName != "" {
		return pod.Spec.PriorityClassName
	}
	return ""
}

func (p *PodParser) getPodGroupOfPod(pod *corev1.Pod) *kaiv2alpha2.PodGroup {
	podGroupName, found := pod.Labels[constants.PodGroupNameAnnotation]
	if !found {
		return nil
	}

	podGroup := &kaiv2alpha2.PodGroup{}
	err := p.cachedClient.Get(context.Background(), types.NamespacedName{
		Namespace: pod.Namespace,
		Name:      podGroupName,
	}, podGroup)
	if err != nil {
		zap.S().Warnf("Failed getting PodGroup <%s/%s>, error: %s", pod.Namespace, podGroupName, err.Error())
		return nil
	}
	return podGroup
}

func parsePodAllocatedResources(pod *corev1.Pod,
	requestedResources openapi.NullablePodRequestResources) openapi.NullableAllocatedResources {
	if pod.Status.Phase == corev1.PodRunning || (pod.Status.Phase == corev1.PodPending && pod.Spec.NodeName != "") {
		allocatedResources := resourceutils.PodRunaiAllocatedResources(pod.Spec.Containers, pod.Annotations)
		return parseAllocatedResources(allocatedResources, requestedResources)
	}

	return openapi.NullableAllocatedResources{}
}

func getContainers(pod *corev1.Pod) []openapi.Container {
	containers := []openapi.Container{}
	for _, container := range pod.Spec.Containers {
		var startedAt metav1.Time
		containerStatus := findContainerStatus(pod.Status.ContainerStatuses, container.Name)
		if containerStatus != nil && containerStatus.State.Running != nil {
			startedAt = containerStatus.State.Running.StartedAt
		}

		containers = append(containers, openapi.Container{
			Name:      container.Name,
			Image:     container.Image,
			StartedAt: getApiTime(startedAt),
			//TODO: fill WaitReason
		})
	}
	return containers
}

func getPodTolerations(tolerations []corev1.Toleration) []openapi.PodToleration {
	result := make([]openapi.PodToleration, 0, len(tolerations))
	for _, toleration := range tolerations {
		result = append(result, openapi.PodToleration{
			Key:               strPointerIfNotEmpty(toleration.Key),
			Operator:          strPointerIfNotEmpty(string(toleration.Operator)),
			Value:             strPointerIfNotEmpty(toleration.Value),
			Effect:            strPointerIfNotEmpty(string(toleration.Effect)),
			TolerationSeconds: getApiInt32FromInt64(toleration.TolerationSeconds),
		})
	}
	return result
}

func calculatePodCompletedAt(pod *corev1.Pod, calculatedPodPhase string) metav1.Time {
	var completedAt metav1.Time

	if pod.Status.Phase == corev1.PodPending && pod.DeletionTimestamp != nil {
		// Terminating pods might appear as pending on certain events, so if a pod is pending but also has a deletion time we treat it as completed.
		completedAt = *pod.DeletionTimestamp
		zap.S().Infof("Pod's <%s/%s> CompletedAt calculated from DeletionTimestamp: %v",
			pod.Namespace, pod.Name, completedAt)
	} else if pod.Status.ContainerStatuses != nil {
		// Get completed time from the pod's containers terminated state
		for _, cs := range pod.Status.ContainerStatuses {
			if cs.State.Terminated != nil {
				if completedAt.IsZero() || completedAt.Before(&cs.State.Terminated.FinishedAt) {
					completedAt = cs.State.Terminated.FinishedAt
				}
			}
		}
	} else if calculatedPodPhase == constants.Status.Failed || calculatedPodPhase == constants.Status.Deleted ||
		calculatedPodPhase == constants.Status.TimedOut || calculatedPodPhase == constants.Status.Preempted ||
		calculatedPodPhase == constants.Status.Timeout {
		if pod.DeletionTimestamp != nil {
			completedAt = *pod.DeletionTimestamp
		} else {
			// This is a special case for pods that are Evicted (but status is failed) as this event does not contain a pod termination or DeletionTimestamp sections
			// so in order to have a proper completion time (which we must have to properly calculate total times) we put one artificially (and persist it so future
			// calculations for the pod group show the same numbers
			completedAt = metav1.Time{Time: time.Now()}
			zap.S().Infof("Pod <%s/%s> with status <%s> in cache without completion/deletion time, generating one instead: %v",
				pod.Namespace, pod.Name, calculatedPodPhase, completedAt)
		}
	}

	return completedAt
}

func GetPodRunaiK8sPhase(pod *corev1.Pod) string {
	phase := getPodRunaiPhaseFromAnnotations(pod)
	if phase != "" {
		return phase
	}
	return string(pod.Status.Phase)
}

func getPodRunaiPhaseFromAnnotations(pod *corev1.Pod) string {
	return parseStringAnnotation(pod.Annotations, nil, constants.WorkloadCalculatedStatus, pod.Name)
}

func findContainerStatus(containerStatuses []corev1.ContainerStatus, containerName string) *corev1.ContainerStatus {
	for _, containerStatus := range containerStatuses {
		if containerStatus.Name == containerName {
			return &containerStatus
		}
	}
	return nil
}

func getPodCurrentNodePool(labels map[string]string, phase corev1.PodPhase) string {
	if !isPodAssignedToNodeByPhase(phase) {
		return ""
	}
	return getObjNodePool(labels)
}

func getObjNodePool(labels map[string]string) string {
	nodePool, found := labels[v1alpha1.NodePoolNameLabel]
	if found {
		return nodePool
	}
	return v1alpha1.DefaultNodePoolName
}

func isPodAssignedToNodeByPhase(phase corev1.PodPhase) bool {
	return phase == corev1.PodRunning || phase == corev1.PodFailed || phase == corev1.PodSucceeded
}

func getPodNodeAffinity(podSpec corev1.PodSpec) *openapi.PodNodeAffinity {
	if podSpec.Affinity == nil || podSpec.Affinity.NodeAffinity == nil {
		return nil
	}

	affinity := podSpec.Affinity.NodeAffinity
	podNodeAffinity := &openapi.PodNodeAffinity{}

	if affinity.RequiredDuringSchedulingIgnoredDuringExecution != nil {
		podNodeAffinityRequired := &openapi.PodNodeAffinityRequired{}
		for _, term := range affinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms {
			apiTerm := openapi.NodeSelectorTerm{
				MatchExpressions: convutils.ConvertSliceTypes(term.MatchExpressions, matchExpressionModelToApi),
			}
			podNodeAffinityRequired.NodeSelectorTerms = append(podNodeAffinityRequired.NodeSelectorTerms, apiTerm)
		}

		podNodeAffinity.Required = *openapi.NewNullablePodNodeAffinityRequired(podNodeAffinityRequired)
	}
	if affinity.PreferredDuringSchedulingIgnoredDuringExecution != nil {
		podNodeAffinityPreferred := &openapi.PodNodeAffinityPreferred{
			NodeSelectorTerms: convutils.ConvertSliceTypes(affinity.PreferredDuringSchedulingIgnoredDuringExecution, PreferredNodeSelectorTermClusterToApi),
		}

		podNodeAffinity.Preferred = *openapi.NewNullablePodNodeAffinityPreferred(podNodeAffinityPreferred)
	}

	return podNodeAffinity
}

// parsePodTask extracts the task from pod metadata for distributed workloads.
func parsePodTask(pod *corev1.Pod) *string {
	// Check if this pod is part of a LeaderWorkerSet by looking for the worker-index label
	if workerIndex, hasWorkerIndex := pod.Labels[constants.LwsWorkerIndexLabelKey]; hasWorkerIndex {
		return parseLWSPodTask(workerIndex)
	}

	// More distributed workload use-cases to follow here
	return nil
}

// parseLWSPodTask extracts the task role (leader/worker) from pod labels for LeaderWorkerSet workloads.
// It checks for LeaderWorkerSet labels to determine if the pod is part of a distributed workload
// and returns "leader" if worker-index is "0", "worker" otherwise, or nil if not a distributed workload.
func parseLWSPodTask(workerIndex string) *string {
	// Worker index "0" indicates the leader pod, any other index indicates a worker pod
	var task string
	if workerIndex == "0" {
		task = constants.PodTaskLeader
	} else {
		task = constants.PodTaskWorker
	}

	return &task
}

func PreferredNodeSelectorTermClusterToApi(term corev1.PreferredSchedulingTerm) openapi.PodPreferredNodeSelectorTerm {
	return openapi.PodPreferredNodeSelectorTerm{
		Weight:     ptr.To(float32(term.Weight)),
		Preference: *openapi.NewNullableNodeSelectorTerm(ptr.To(NodeSelectorTermModelToApi(term.Preference))),
	}
}

func NodeSelectorTermModelToApi(term corev1.NodeSelectorTerm) openapi.NodeSelectorTerm {
	return openapi.NodeSelectorTerm{
		MatchExpressions: convutils.ConvertSliceTypes(term.MatchExpressions, matchExpressionModelToApi),
	}
}

func matchExpressionModelToApi(expr corev1.NodeSelectorRequirement) openapi.MatchExpression {
	return openapi.MatchExpression{
		Key:      expr.Key,
		Operator: openapi.MatchExpressionOperator(expr.Operator),
		Values:   expr.Values,
	}
}

func parsePodElements(labels map[string]string) []openapi.PodElement {
	var elements []openapi.PodElement
	for level := int32(1); level <= 3; level++ {
		key := fmt.Sprintf("%s%d-id", elementLabelPrefix, level)
		if id, ok := labels[key]; ok && id != "" {
			elements = append(elements, openapi.PodElement{
				ElementId:      id,
				HierarchyLevel: level,
			})
		}
	}
	return elements
}

// parsePodResourceClaims projects the pod's DRA usage into the server-side
// shape, verbatim. One entry per pod.spec.resourceClaims[] member, enriched
// with the actual bound ResourceClaim name:
//   - template-generated entry: name from pod.status.resourceClaimStatuses[]
//     (nil until Kubernetes publishes it);
//   - standalone entry: equals the spec's resourceClaimName (available
//     immediately — k8s does not populate status for standalone entries,
//     per the "If this is unset, then generating a ResourceClaim was not
//     necessary" contract).
//
// Container-level resources.claims[] references are intentionally not
// reflected here: the design reports per-pod, not per-container.
//
// Returns nil for a pod with no resource claims so the field is omitted on
// the wire for pods that do not use DRA.
func parsePodResourceClaims(pod *corev1.Pod) []openapi.PodResourceClaim {
	if len(pod.Spec.ResourceClaims) == 0 {
		return nil
	}

	// Pre-index status entries by pod-local name for O(1) lookup during
	// the spec walk.
	allocatedByName := make(map[string]*string, len(pod.Status.ResourceClaimStatuses))
	for _, status := range pod.Status.ResourceClaimStatuses {
		allocatedByName[status.Name] = status.ResourceClaimName
	}

	result := make([]openapi.PodResourceClaim, 0, len(pod.Spec.ResourceClaims))
	for _, claim := range pod.Spec.ResourceClaims {
		var allocated *string
		switch {
		case claim.ResourceClaimName != nil:
			// Standalone claim: the spec itself names the actual claim.
			allocated = claim.ResourceClaimName
		default:
			// Template-generated: look up the status entry; nil until bound.
			allocated = allocatedByName[claim.Name]
		}

		result = append(result, openapi.PodResourceClaim{
			Name:                       claim.Name,
			ResourceClaimTemplateName:  *openapi.NewNullableString(claim.ResourceClaimTemplateName),
			ResourceClaimName:          *openapi.NewNullableString(claim.ResourceClaimName),
			AllocatedResourceClaimName: *openapi.NewNullableString(allocated),
		})
	}
	return result
}
