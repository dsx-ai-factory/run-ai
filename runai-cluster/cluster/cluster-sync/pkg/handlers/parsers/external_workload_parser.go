/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package parsers

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/config"
	constants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/getters"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	issuepkg "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/issue"
	workloadTypeConfig "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/workload-type-config"
	common "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/common"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/priority"
	topowner "github.com/dsx-ai-factory/run-ai/runai-cluster/common/top-owner"
	cu "github.com/dsx-ai-factory/run-ai/runai-cluster/common/utils"
	kaiv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	nimServiceType                   = "NIMService"
	distributedInferenceWorkloadType = "DistributedInference"
	readyReplicasAdditionalField     = "readyReplicas"

	maxStructureElementsPerWorkload = 150
)

type WorkloadParser struct {
	allowedWorkloadPhases map[openapi.Phase]bool
	allowedPhaseReasons   map[openapi.PhaseReason]bool
	cachedReader          client.Reader
}

func NewWorkloadParser(cachedReader client.Reader) *WorkloadParser {
	allowedWorkloadPhases := map[openapi.Phase]bool{}
	for _, allowedPhase := range openapi.AllowedPhaseEnumValues {
		allowedWorkloadPhases[allowedPhase] = true
	}

	allowedPhaseReasons := map[openapi.PhaseReason]bool{}
	for _, allowedPhaseReason := range openapi.AllowedPhaseReasonEnumValues {
		allowedPhaseReasons[allowedPhaseReason] = true
	}

	return &WorkloadParser{
		allowedWorkloadPhases: allowedWorkloadPhases,
		allowedPhaseReasons:   allowedPhaseReasons,
		cachedReader:          cachedReader,
	}
}

func (p *WorkloadParser) getWorkloadCategory(ctx context.Context, workloadKind, workloadGroup string, labels map[string]string) (string, error) {
	workloadCategory := labels[constants.WorkloadCategoryLabel]

	if err := workloadTypeConfig.ValidateWorkloadCategory(ctx, p.cachedReader, workloadCategory); err == nil {
		return workloadCategory, nil
	}
	cfg, err := workloadTypeConfig.GetWorkloadConfigByGroupKind(ctx, p.cachedReader, workloadGroup, workloadKind)
	if err != nil {
		return "", err
	}

	return cfg.CategoryName, nil
}

func (p *WorkloadParser) getWorkloadCategoryForExternalWorkload(ctx context.Context, workload *v1alpha1.ExternalWorkload) (string, error) {
	workloadType := workload.Spec.Type
	workloadGroup := workload.Spec.Group
	return p.getWorkloadCategory(ctx, workloadType, workloadGroup, workload.GetLabels())
}

func distributedFrameworkFromKind(kind string) string {
	switch kind {
	case constants.KindMpiJob:
		return string(openapi.DISTRIBUTEDFRAMEWORK_MPI)
	case constants.KindPyTorchJob:
		return string(openapi.DISTRIBUTEDFRAMEWORK_PY_TORCH)
	case constants.KindTFJob:
		return string(openapi.DISTRIBUTEDFRAMEWORK_TF)
	case constants.KindXGBoostJob:
		return string(openapi.DISTRIBUTEDFRAMEWORK_XG_BOOST)
	case constants.KindJAXJob:
		return string(openapi.DISTRIBUTEDFRAMEWORK_JAX)
	default:
		return ""
	}
}

func workloadNameForSync(workload *v1alpha1.ExternalWorkload) string {
	if workload.Spec.Type == distributedInferenceWorkloadType &&
		workload.Labels[commonconstants.ConvertedFromNativeLabel] == constants.BoolTrueAsString {
		if name := workload.Labels[constants.LabelWorkloadName]; name != "" {
			return name
		}
	}

	return workload.Spec.OriginalWorkloadName
}

func (p *WorkloadParser) distributedFrameworkForExternalWorkload(workload *v1alpha1.ExternalWorkload, workloadType string) string {
	if workload == nil || workloadType != DistributedDisplayType {
		return ""
	}

	if framework := distributedFrameworkFromKind(workload.Spec.OriginalWorkloadKind); framework != "" {
		return framework
	}

	if framework := distributedFrameworkFromKind(workload.Spec.Type); framework != "" {
		return framework
	}

	for _, ownerReference := range workload.OwnerReferences {
		if framework := distributedFrameworkFromKind(ownerReference.Kind); framework != "" {
			return framework
		}
	}

	return ""
}

func (p *WorkloadParser) ParseExternalWorkloadToServerScheme(
	externalWorkload *v1alpha1.ExternalWorkload,
) (*openapi.WorkloadSyncFields, []corev1.Pod, []*kaiv2alpha2.PodGroup, error) {
	workloadId := topowner.GetRunaiWorkloadId(externalWorkload.Spec.OriginalWorkloadUID, externalWorkload.Annotations)
	source := getWorkloadSource(externalWorkload.Annotations, nil)

	podsOfWorkload, err := getters.GetPodsOfWorkload(context.TODO(),
		workloadId, externalWorkload.Namespace, p.cachedReader)
	if err != nil {
		zap.S().Warnf("Failed to get pods for external workload <%s/%s>, error: %s",
			externalWorkload.Namespace, externalWorkload.Name, err.Error())
		return nil, nil, nil, err
	}
	podGroupsOfWorkload, err := getters.GetPodGroupsOfPods(context.TODO(), externalWorkload.Namespace, podsOfWorkload, p.cachedReader)
	if err != nil {
		zap.S().Warnf("Failed to get pod groups of pod for external workload <%s/%s>, error: %s",
			externalWorkload.Namespace, externalWorkload.Name, err.Error())
		return nil, nil, nil, err
	}

	guaranteedRuntimeTimestamp := p.getGuaranteedRuntimeTimestamp(podGroupsOfWorkload, &podsOfWorkload)

	priorityClassName, priorityValue, preemptibility := p.getPriorityClassAndPreemptibilityForExternalWorkload(externalWorkload, podGroupsOfWorkload)
	schedulingConditions := p.getSchedulingConditions(podGroupsOfWorkload)

	childIds := parseExternalWorkloadChildIds(externalWorkload.Status.ObservedStatus.ChildIds)
	childIds = append(childIds, p.getPodGroupsChildren(podGroupsOfWorkload)...)
	workloadType := externalWorkload.Spec.Type

	workloadSyncFields := openapi.WorkloadSyncFields{
		Type:                 workloadType,
		Group:                strPointerIfNotEmpty(externalWorkload.Spec.Group),
		Version:              strPointerIfNotEmpty(externalWorkload.Spec.Version),
		Name:                 workloadNameForSync(externalWorkload),
		ExternalWorkloadName: &externalWorkload.Name,
		Id:                   workloadId,
		ClusterId:            config.Config().ClusterUuid,
		ProjectName:          externalWorkload.Spec.ProjectName,
		ProjectId:            externalWorkload.Spec.ProjectId,
		DepartmentName:       externalWorkload.Spec.DepartmentName,
		DepartmentId:         externalWorkload.Spec.DepartmentId,
		Namespace:            externalWorkload.Namespace,
		PriorityClassName:    priorityClassName,
		Preemptibility:       *openapi.NewNullableString(&preemptibility),
		CreatedAt:            externalWorkload.Spec.CreationTimestamp.Time,
		RequestedResources:   p.parseExternalWorkloadRequestedResources(externalWorkload.Spec.RequestedResources),
		Phase:                p.getApiPhase(string(externalWorkload.Status.ObservedStatus.Phase)),
		Source:               source,
		Conditions:           parseExternalWorkloadConditions(externalWorkload.Status.ObservedStatus.Conditions),
		PhaseMessage:         strPointerIfNotEmpty(externalWorkload.Status.ObservedStatus.Message),
		K8sPhase:             externalWorkload.Status.ObservedStatus.K8sStatus,
		RequestedPods:        parseRequestedPods(externalWorkload.Spec.RequestedPods),
		RequestedNodePools:   externalWorkload.Spec.RequestedNodePools,
		Images:               getApiImages(externalWorkload.Spec.ImageName),
		ChildrenIds:          childIds,
		Urls:                 externalWorkload.Spec.Urls,
		AdditionalFields:     getExternalWorkloadAdditionalFields(externalWorkload.Spec, externalWorkload.Labels, externalWorkload.Annotations),
		DistributedFramework: strPointerIfNotEmpty(p.distributedFrameworkForExternalWorkload(externalWorkload, workloadType)),
		FullySaasWorkload:    cu.Ref(isControlPlaneManagedLabeled(externalWorkload.Labels)),
		SchedulingConditions: schedulingConditions,
	}
	if priorityValue != nil {
		workloadSyncFields.SetPriority(*priorityValue)
	}
	if guaranteedRuntimeTimestamp != nil {
		workloadSyncFields.SetGuaranteedRuntimeEndsAt(*guaranteedRuntimeTimestamp)
	}
	category, err := p.getWorkloadCategoryForExternalWorkload(context.TODO(), externalWorkload)
	if err != nil {
		zap.S().Errorf("failed workload category for type %s, error: %s",
			externalWorkload.Kind, err.Error())
	}
	if category != "" {
		workloadSyncFields.Category = ptr.To(category)
	}
	applicationId, applicationName := getWorkloadApplication(externalWorkload.Labels)
	workloadSyncFields.ApplicationId = applicationId
	workloadSyncFields.ApplicationName = applicationName

	ewTopology := externalWorkload.Spec.Topology
	if ewTopology.Name != "" || ewTopology.PreferredPlacement != "" || ewTopology.RequiredPlacement != "" {
		topology := openapi.WorkloadSyncFieldsTopology{}
		if ewTopology.Name != "" {
			topology.SetName(ewTopology.Name)
		}
		if ewTopology.PreferredPlacement != "" {
			topology.SetPreferredPlacement(ewTopology.PreferredPlacement)
		}
		if ewTopology.RequiredPlacement != "" {
			topology.SetRequiredPlacement(ewTopology.RequiredPlacement)
		}
		workloadSyncFields.SetTopology(topology)
	}

	if se := externalWorkload.Status.StructureElements; se != nil {
		workloadSyncFields.StructureElements = parseStructureElements(se, externalWorkload.Spec.ComponentTopologies)
		workloadSyncFields.SetStructureElementsTotalCount(int32(len(se.Elements))) //nolint:gosec
	}
	if len(externalWorkload.Status.NetworkStatus.ExternalEndpoints) > 0 {
		endpoints, err := parseExternalEndpoints(externalWorkload.Status.NetworkStatus.ExternalEndpoints)
		if err != nil {
			zap.S().Errorf("failed to parse workload endpoints for ew %s, error: %s",
				externalWorkload.Name, err.Error())
		}
		workloadSyncFields.Endpoints = endpoints
	}

	if len(externalWorkload.Spec.NodeAffinities) > 0 {
		workloadSyncFields.NodeAffinities = externalWorkloadNodeAffinitiesToApi(externalWorkload.Spec.NodeAffinities)
	}

	if len(externalWorkload.Spec.StorageRefs) > 0 {
		// The EWI already resolved each ref to its UID (see resolveStorageRefUIDs in
		// commonparsers), so pass them straight through instead of re-resolving per sync.
		refs := make([]openapi.ResourceRef, 0, len(externalWorkload.Spec.StorageRefs))
		for _, ref := range externalWorkload.Spec.StorageRefs {
			if ref.UID == "" {
				continue
			}
			refs = append(refs, openapi.ResourceRef{
				Uid:  ptr.To(ref.UID),
				Name: ref.Name,
				Kind: openapi.ResourceRefKind(ref.Kind),
			})
		}
		if len(refs) > 0 {
			workloadSyncFields.ResourceRefs = refs
		}
	}

	p.overrideFromWorkloadRunner(externalWorkload, &workloadSyncFields)

	return &workloadSyncFields, podsOfWorkload, podGroupsOfWorkload, nil
}

func parseStructureElements(se *v1alpha1.StructureElements,
	componentTopologies []v1alpha1.ComponentTopologyEntry,
) []openapi.WorkloadElement {
	topologyByComponent := componentTopologiesBySpecName(componentTopologies)
	numElementsToSync := len(se.Elements)
	if numElementsToSync > maxStructureElementsPerWorkload {
		// design decision - if there are too many structure elements, we will only sync the first
		// maxStructureElementsPerWorkload elements, and set the total count to the actual number of elements.
		// this way we can still have some visibility into the structure of the workload without
		// overwhelming the system with too many elements.
		numElementsToSync = maxStructureElementsPerWorkload
	}

	result := make([]openapi.WorkloadElement, 0, numElementsToSync)
	for i := 0; i < numElementsToSync; i++ {
		e := se.Elements[i]
		apiElem := openapi.WorkloadElement{
			Id:             e.ID,
			Name:           e.Name,
			HierarchyLevel: int32(e.Level), //nolint:gosec // level is 1-3
		}
		if e.ParentID != nil {
			apiElem.SetParentId(*e.ParentID)
		}
		if e.InstanceKey != nil {
			apiElem.SetInstanceKey(*e.InstanceKey)
			if topology, found := topologyByComponent[*e.InstanceKey]; found {
				apiElem.SetTopology(elementTopologyToApi(topology))
			}
		}
		if e.ReplicaKey != nil {
			apiElem.SetReplicaKey(*e.ReplicaKey)
		}
		if e.K8sRef != nil {
			uid := string(e.K8sRef.UID)
			apiElem.SetKubernetesRef(openapi.WorkloadElementKubernetesRef{
				Uid:  &uid,
				Name: &e.K8sRef.Name,
				Gvk: &openapi.WorkloadElementKubernetesRefGvk{
					Group:   &e.K8sRef.GVK.Group,
					Version: &e.K8sRef.GVK.Version,
					Kind:    &e.K8sRef.GVK.Kind,
				},
			})
		}
		result = append(result, apiElem)
	}
	return result
}

func componentTopologiesBySpecName(entries []v1alpha1.ComponentTopologyEntry) map[string]v1alpha1.Topology {
	if len(entries) == 0 {
		return nil
	}
	byName := make(map[string]v1alpha1.Topology, len(entries))
	for _, entry := range entries {
		byName[entry.SpecName] = entry.Topology
	}
	return byName
}

func elementTopologyToApi(topology v1alpha1.Topology) openapi.WorkloadElementTopology {
	apiTopology := openapi.WorkloadElementTopology{}
	if topology.Name != "" {
		apiTopology.SetName(topology.Name)
	}
	if topology.PreferredPlacement != "" {
		apiTopology.SetPreferredPlacement(topology.PreferredPlacement)
	}
	if topology.RequiredPlacement != "" {
		apiTopology.SetRequiredPlacement(topology.RequiredPlacement)
	}
	return apiTopology
}

func parseExternalEndpoints(endpoints []v1alpha1.ExternalEndpoint) ([]openapi.EndpointSync, error) {
	syncEndpoints := make([]openapi.EndpointSync, 0, len(endpoints))
	var parseErr error

	for _, endpoint := range endpoints {
		port, err := strconv.ParseInt(endpoint.Port, 10, 32)
		if err != nil {
			if parseErr == nil {
				parseErr = fmt.Errorf("failed parsing external endpoint port %q for %s/%s: %w",
					endpoint.Port, endpoint.SourceRef.Kind, endpoint.SourceRef.Name, err)
			}
			continue
		}
		ep := openapi.EndpointSync{
			Host:     &endpoint.Host,
			Port:     ptr.To(int32(port)),
			Path:     &endpoint.Path,
			Protocol: &endpoint.Protocol,
			Url:      &endpoint.Url,
			SourceRef: &openapi.EndpointSyncAllOfSourceRef{
				Kind: &endpoint.SourceRef.Kind,
				Name: &endpoint.SourceRef.Name,
			},
		}
		if endpoint.TlsEnabled != nil {
			ep.SetTlsEnabled(*endpoint.TlsEnabled)
		}
		syncEndpoints = append(syncEndpoints, ep)
	}
	return syncEndpoints, parseErr
}

func (p *WorkloadParser) parseExternalWorkloadRequestedResources(
	requestedResources corev1.ResourceRequirements) openapi.NullableWorkloadRequestResources {
	podNullableRequestedResources := parseRequestedResources(requestedResources.Requests,
		requestedResources.Limits)
	podRequestedResources := podNullableRequestedResources.Get()
	if podRequestedResources == nil {
		return openapi.NullableWorkloadRequestResources{}
	}

	result := openapi.WorkloadRequestResources{
		Cpu:               podRequestedResources.Cpu,
		CpuMemory:         podRequestedResources.CpuMemory,
		Gpu:               podRequestedResources.Gpu,
		GpuMemory:         podRequestedResources.GpuMemory,
		GpuRequestType:    podRequestedResources.GpuRequestType,
		ExtendedResources: podRequestedResources.ExtendedResources,
	}
	nullableResult := openapi.NullableWorkloadRequestResources{}
	nullableResult.Set(&result)
	return nullableResult
}

func (p *WorkloadParser) getPriorityClassAndPreemptibilityForExternalWorkload(
	externalWorkload *v1alpha1.ExternalWorkload,
	podGroupsOfWorkload []*kaiv2alpha2.PodGroup) (string, *int32, string) {
	priorityClassName, priorityValue, workloadTypeConfigOfKind := p.getPriorityClassOfWorkload(externalWorkload.Labels, externalWorkload.Spec.PriorityClassName, externalWorkload.Spec.Type, externalWorkload.Spec.Group, podGroupsOfWorkload)
	preemptibility := externalWorkload.Spec.Preemptibility
	if preemptibility == "" {
		preemptibility = p.getPreemptibilityOfWorkload(externalWorkload.Labels, externalWorkload.Spec.Type, externalWorkload.Spec.Group, podGroupsOfWorkload, workloadTypeConfigOfKind)
	}
	return priorityClassName, priorityValue, preemptibility
}

func (p *WorkloadParser) getPriorityClassFromPodGroups(podGroupsOfWorkload []*kaiv2alpha2.PodGroup) (*string, *int32) {
	if len(podGroupsOfWorkload) == 0 {
		return nil, nil
	}

	// we assume they all have the same priority..
	priorityClassName := podGroupsOfWorkload[0].Spec.PriorityClassName

	priorityValue, err := p.getPriorityClassFromName(priorityClassName)
	if err != nil {
		return nil, nil
	}

	return &priorityClassName, &priorityValue
}

func (p *WorkloadParser) getPriorityClassFromName(priorityClassName string) (int32, error) {
	priorityClass, err := priority.GetPriorityClass(context.TODO(), priorityClassName, p.cachedReader)
	if err == nil && priorityClass != nil {
		return priorityClass.Value, nil
	}
	zap.S().Errorf("Failed to get priorityClass %s, error: %v", priorityClassName, err)

	priorityValue, innerErr := priority.GetDefaultPriority(context.TODO(), p.cachedReader)
	if innerErr != nil {
		zap.S().Errorf("Failed to get default priority, error: %s", innerErr.Error())
		return 0, nil
	}
	zap.S().Warnf("Couldn't find priorityClass %s, error: %v. Using default priority %d",
		priorityClassName, err, priorityValue)
	return priorityValue, nil
}

func (p *WorkloadParser) getPriorityClassValueOrNil(priorityClassName string) *int32 {
	priorityValue, err := p.getPriorityClassFromName(priorityClassName)
	if err != nil {
		// error logged inside
		return nil
	}

	return &priorityValue
}

func (p *WorkloadParser) getPreemptibilityFromPodGroups(podGroupsOfWorkload []*kaiv2alpha2.PodGroup) string {
	if len(podGroupsOfWorkload) == 0 {
		return ""
	}

	// we assume they all have the same preemptibility..
	return string(podGroupsOfWorkload[0].Spec.Preemptibility)
}

func getGpuRequestType(gpuRequestType openapi.GpuRequestType) openapi.NullableGpuRequestType {
	result := openapi.NullableGpuRequestType{}
	result.Set(&gpuRequestType)
	return result
}

func parseExternalWorkloadConditions(conditions []v1alpha1.ObservedCondition) []openapi.Condition {
	result := []openapi.Condition{}
	for _, condition := range conditions {
		// WorkloadIssue conditions surface as issues, not native conditions.
		if condition.Type == commonconstants.WorkloadIssueConditionType {
			continue
		}
		result = append(result, openapi.Condition{
			Type:               condition.Type,
			Status:             string(condition.Status),
			Message:            strPointerIfNotEmpty(condition.Message),
			Reason:             strPointerIfNotEmpty(condition.Reason),
			LastTransitionTime: getApiTime(condition.LastTransitionTime),
		})
	}
	return result
}

func parseExternalWorkloadChildIds(childIds []v1alpha1.ChildId) []openapi.WorkloadChildrenIdsInner {
	result := []openapi.WorkloadChildrenIdsInner{}
	for _, childId := range childIds {
		result = append(result, openapi.WorkloadChildrenIdsInner{
			Id:   childId.UID,
			Type: ptr.To(childId.Type),
		})
	}
	return result
}

func parseRequestedPods(requestedPods v1alpha1.RequestedPods) *openapi.RequestedPods {
	return &openapi.RequestedPods{
		Number:      getApiInt32(requestedPods.Replicas),
		Min:         getApiInt32(requestedPods.MinReplicas),
		Max:         getApiInt32(requestedPods.MaxReplicas),
		Parallelism: getApiInt32(requestedPods.Parallelism),
		Completions: getApiInt32(requestedPods.Completions),
	}
}

func getExternalWorkloadAdditionalFields(externalWorkloadSpec v1alpha1.ExternalWorkloadSpec, labels,
	annotations map[string]string) *string {
	additionalFields := map[string]any{}

	crdAdditionalFields := externalWorkloadSpec.AdditionalFields
	if crdAdditionalFields != nil {
		additionalFieldsBytes := []byte(*crdAdditionalFields)
		err := json.Unmarshal(additionalFieldsBytes, &additionalFields)
		if err != nil {
			zap.S().Warnf("Failed to unmarshal additionalFields <%s> for workload, err: %s",
				*crdAdditionalFields, err.Error())
			return nil
		}
	}

	addNIMWorkloadFlagIfPresent(labels, externalWorkloadSpec.Type, additionalFields)
	addVLLMWorkloadFlagIfPresent(labels, additionalFields)
	addCliCommandIfPresent(annotations, additionalFields)

	return getWorkloadAdditionalFields(additionalFields)
}

func addCliCommandIfPresent(annotations map[string]string, additionalFields map[string]any) {
	if cliCommand := annotations[commonconstants.WorkloadCliCommandAnnotation]; cliCommand != "" {
		additionalFields[additionalCliCommand] = cliCommand
	}
}

func addNIMWorkloadFlagIfPresent(labels map[string]string, externalWorkloadType string, additionalFields map[string]any) {
	if labels[constants.LabelNIMWorkload] == "true" || externalWorkloadType == nimServiceType {
		additionalFields[constants.AdditionalNIMWorkload] = "true"
	}
}

func addVLLMWorkloadFlagIfPresent(labels map[string]string, additionalFields map[string]any) {
	if labels[constants.LabelVLLMWorkload] == "true" {
		additionalFields[constants.AdditionalVLLMWorkload] = "true"
	}
}

func getWorkloadAdditionalFields[T any](additionalFields map[string]T) *string {
	if len(additionalFields) == 0 {
		return nil
	}

	additionalFieldsBytes, err := json.Marshal(additionalFields)
	if err != nil {
		zap.S().Errorf("Failed to marshal additionalFields <%+v> for workload, err: %s",
			additionalFields, err.Error())
		return nil
	}

	additionalFieldsStr := string(additionalFieldsBytes)
	return &additionalFieldsStr
}

func getApiImages(workloadImage string) []string {
	apiImages := []string{}
	if workloadImage != "" {
		apiImages = append(apiImages, workloadImage)
	}
	return apiImages
}

func getWorkloadApplication(labels map[string]string) (openapi.NullableString, openapi.NullableString) {
	id := openapi.NullableString{}
	name := openapi.NullableString{}
	if labels == nil {
		return id, name
	}
	applicationId, ok := labels[constants.ApplicationIdLabel]
	if ok && applicationId != "" {
		id.Set(&applicationId)
	}
	applicationName, ok := labels[constants.ApplicationNameLabel]
	if ok && applicationName != "" {
		name.Set(&applicationName)
	}
	return id, name
}

// overrideFromWorkloadRunner overrides sync fields with WLR status (phase + issues).
// If new WLR fields are synced here, update the change detection in
// workload_runner_controller.go to trigger re-syncs when those fields change.
func (p *WorkloadParser) overrideFromWorkloadRunner(
	externalWorkload *v1alpha1.ExternalWorkload,
	workloadSyncFields *openapi.WorkloadSyncFields,
) {
	wlrName := externalWorkload.Labels[commonconstants.WorkloadRunnerNameLabel]
	if wlrName == "" {
		return
	}

	issues := &common.Issues{}

	wlr := &dromosv1alpha1.WorkloadRunner{}
	if err := p.cachedReader.Get(context.TODO(),
		client.ObjectKey{Name: wlrName, Namespace: externalWorkload.Namespace}, wlr); err != nil {
		zap.S().Warnf("Failed to get WorkloadRunner <%s/%s> for ExternalWorkload <%s/%s>, error: %s",
			externalWorkload.Namespace, wlrName, externalWorkload.Namespace, externalWorkload.Name, err.Error())
		appendWorkloadIssueConditions(issues, externalWorkload)
		writeIssuesToSyncFields(workloadSyncFields, issues)
		return
	}
	projectProfileCompatibilityFields(wlr, workloadSyncFields)
	appendInferenceEndpointCondition(wlr, workloadSyncFields)
	if externalWorkload.Spec.Type == distributedInferenceWorkloadType {
		projectDistributedInferenceCompatibilityFields(wlr, workloadSyncFields)
	}
	// Keep EWI's phase when ewi-status is set; else fall back to the wlr phase.
	if !ewiStatusSet(externalWorkload) && wlr.Status.Phase != "" {
		workloadSyncFields.Phase = p.getApiPhase(string(wlr.Status.Phase))
		// Admission failures are recorded as critical WLR issues; expose their description to the user.
		for _, issue := range wlr.Status.Issues {
			if issue.IsCritical && issue.Description != "" {
				workloadSyncFields.PhaseMessage = ptr.To(issue.Description)
				break
			}
		}
	}

	// Register the runner as a child so its events join to this workload.
	if wlr.UID != "" {
		workloadSyncFields.ChildrenIds = append(workloadSyncFields.ChildrenIds,
			openapi.WorkloadChildrenIdsInner{
				Id:   string(wlr.UID),
				Type: ptr.To(dromosv1alpha1.WorkloadRunnerKind),
			})
	}

	appendWlrIssues(issues, wlr)
	appendWorkloadIssueConditions(issues, externalWorkload)
	writeIssuesToSyncFields(workloadSyncFields, issues)

	if networkingStatus, err := networkingDetailsFromWorkloadRunner(wlr); err != nil {
		zap.S().Warnf("Failed to parse networking plugin details for WorkloadRunner <%s/%s>, error: %s",
			wlr.Namespace, wlr.Name, err.Error())
	} else if networkingStatus != nil {
		applyNetworkingDetailsToSyncFields(networkingStatus, workloadSyncFields)
	}

	// Override with the external URL from ExternalAccessReady only when EWI fell back to a
	// cluster-local URL (Ingress wasn't in cache yet). Never clobber a user-specified exposedUrl.
	if rawURL := externalUrlFromWorkloadRunner(wlr); rawURL != "" && isInternalClusterURL(workloadSyncFields.Urls) {
		workloadSyncFields.Urls = []string{rawURL}
	}
	suppressDistributedInferenceURLsUntilReady(externalWorkload, workloadSyncFields)
}

func ewiStatusSet(externalWorkload *v1alpha1.ExternalWorkload) bool {
	return externalWorkload.Annotations[commonconstants.EwiStatusAnnotation] == "true"
}

func suppressDistributedInferenceURLsUntilReady(externalWorkload *v1alpha1.ExternalWorkload,
	fields *openapi.WorkloadSyncFields,
) {
	if externalWorkload == nil || fields == nil || externalWorkload.Spec.Type != "DistributedInference" ||
		!ewiStatusSet(externalWorkload) {
		return
	}
	additional := map[string]any{}
	if fields.AdditionalFields != nil {
		if err := json.Unmarshal([]byte(*fields.AdditionalFields), &additional); err != nil {
			fields.Urls = nil
			return
		}
	}
	readyReplicas, ok := additional[readyReplicasAdditionalField].(float64)
	if !ok || readyReplicas <= 0 {
		fields.Urls = nil
	}
}

func appendWlrIssues(issues *common.Issues, wlr *dromosv1alpha1.WorkloadRunner) {
	addIssues := func(wlrIssues []dromosv1alpha1.Issue) {
		for _, issue := range wlrIssues {
			appendIssueToGroup(issues, issue.Group, common.Issue{
				Kind:        issue.Kind,
				Name:        issue.Name,
				Description: issue.Description,
			})
		}
	}
	addIssues(wlr.Status.Issues)
	for _, plugin := range wlr.Status.Plugins {
		if plugin != nil {
			addIssues(plugin.Issues)
		}
	}
}

// appendWorkloadIssueConditions folds EWI's WorkloadIssue conditions into issues, grouped by Reason.
func appendWorkloadIssueConditions(issues *common.Issues, externalWorkload *v1alpha1.ExternalWorkload) {
	for _, condition := range externalWorkload.Status.ObservedStatus.Conditions {
		if condition.Type != commonconstants.WorkloadIssueConditionType {
			continue
		}
		appendIssueToGroup(issues, condition.Reason, common.Issue{
			Kind:        "Pod",
			Name:        externalWorkload.Spec.OriginalWorkloadName,
			Description: condition.Message,
		})
	}
}

func appendIssueToGroup(issues *common.Issues, group string, ci common.Issue) {
	switch group {
	case issuepkg.Network.String():
		issues.Network = append(issues.Network, ci)
	case issuepkg.Storage.String():
		issues.Storage = append(issues.Storage, ci)
	case issuepkg.Compute.String():
		issues.Compute = append(issues.Compute, ci)
	default:
		issues.Internal = append(issues.Internal, ci)
	}
}

func writeIssuesToSyncFields(workloadSyncFields *openapi.WorkloadSyncFields, issues *common.Issues) {
	if len(issues.Network)+len(issues.Storage)+len(issues.Compute)+len(issues.Internal) == 0 {
		return
	}
	workloadSyncFields.AdditionalFields = mergeWlrIssuesIntoAdditionalFields(
		workloadSyncFields.AdditionalFields, issues)
}

func mergeWlrIssuesIntoAdditionalFields(existingAdditionalFields *string, issues *common.Issues) *string {
	additionalFields := map[string]any{}

	if existingAdditionalFields != nil {
		if err := json.Unmarshal([]byte(*existingAdditionalFields), &additionalFields); err != nil {
			zap.S().Warnf("Failed to unmarshal existing additionalFields, err: %s", err.Error())
		}
	}

	issuesStr, err := getWorkloadIssues(*issues)
	if err != nil {
		zap.S().Errorf("Failed to marshal WLR issues, error: %s", err.Error())
		return existingAdditionalFields
	}

	additionalFields[additionalIssues] = issuesStr
	return getWorkloadAdditionalFields(additionalFields)
}

// externalWorkloadNodeAffinitiesToApi converts the EW spec's per-template node-affinity
// entries into the control-plane WorkloadNodeAffinity API shape. Entries whose NodeAffinity
// is nil are skipped so we don't emit empty placeholders.
func externalWorkloadNodeAffinitiesToApi(entries []v1alpha1.NodeAffinityEntry) []openapi.WorkloadNodeAffinity {
	result := make([]openapi.WorkloadNodeAffinity, 0, len(entries))
	for i := range entries {
		if entries[i].NodeAffinity == nil {
			continue
		}
		result = append(result, openapi.WorkloadNodeAffinity{
			SpecName:     entries[i].SpecName,
			NodeAffinity: corev1NodeAffinityToApi(entries[i].NodeAffinity),
		})
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func corev1NodeAffinityToApi(na *corev1.NodeAffinity) *openapi.PodNodeAffinity {
	if na == nil {
		return nil
	}
	result := &openapi.PodNodeAffinity{}
	if na.RequiredDuringSchedulingIgnoredDuringExecution != nil {
		required := &openapi.PodNodeAffinityRequired{}
		for _, term := range na.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms {
			required.NodeSelectorTerms = append(required.NodeSelectorTerms, NodeSelectorTermModelToApi(term))
		}
		result.Required = *openapi.NewNullablePodNodeAffinityRequired(required)
	}
	return result
}

// TestGetExternalWorkloadAdditionalFields is a test helper function that exposes getExternalWorkloadAdditionalFields for testing
func TestGetExternalWorkloadAdditionalFields(externalWorkloadSpec v1alpha1.ExternalWorkloadSpec, labels,
	annotations map[string]string) *string {
	return getExternalWorkloadAdditionalFields(externalWorkloadSpec, labels, annotations)
}

// TestAddNIMWorkloadFlagIfPresent is a test helper function that exposes addNIMWorkloadFlagIfPresent for testing
func TestAddNIMWorkloadFlagIfPresent(labels map[string]string, externalWorkloadType string, additionalFields map[string]any) {
	addNIMWorkloadFlagIfPresent(labels, externalWorkloadType, additionalFields)
}

// TestAddVLLMWorkloadFlagIfPresent exposes addVLLMWorkloadFlagIfPresent for parser tests.
func TestAddVLLMWorkloadFlagIfPresent(labels map[string]string, additionalFields map[string]any) {
	addVLLMWorkloadFlagIfPresent(labels, additionalFields)
}
