/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package parsers

import (
	"context"
	"encoding/json"
	"strings"

	kaiv2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2"
	kubeaischedv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/util"
	clusterconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/extensions"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/getters"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/resourceutils/unschedulable"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	field_types "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	workloadTypeConfig "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/workload-type-config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/common"
	dinfwv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferenceworkload/v1alpha1"
	diwv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	infwv2a1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"
	iwv2a1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	twv2a1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/priority"
	topowner "github.com/dsx-ai-factory/run-ai/runai-cluster/common/top-owner"
	cu "github.com/dsx-ai-factory/run-ai/runai-cluster/common/utils"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	knv1 "knative.dev/serving/pkg/apis/serving/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	WorkspaceDisplayType   = "Workspace"
	TrainingDisplayType    = "Training"
	InteractiveDisplayType = "Interactive"
	InferenceDisplayType   = "Inference"
	DistributedDisplayType = "Distributed"

	NodePoolsSeparator = " "

	// additional fields
	additionalIssues                    = "issues"
	additionalLatestReadyRevisionName   = "latestReadyRevisionName"
	additionalLatestCreatedRevisionName = "latestCreatedRevisionName"
	additionalCliCommand                = "cliCommand"
	additionalGpuAccelerationStatus     = "GPUNetworkAccelerationStatus"

	True  = "True"
	False = "False"
)

type GPUNetworkAcceleration struct {
	Used bool `json:"used"`
}

// WorkloadFieldsForParser is used instead of converting to "JobWorkload", which is rather
// an expensive operation. Use this struct to get the fields we need for parsing.
type WorkloadFieldsForParser struct {
	Kind string
	metav1.TypeMeta
	metav1.ObjectMeta
	field_groups.ResourceAllocationFields
	field_groups.RunaiSchedulerFields
	field_groups.SchedulingFields
	WorkloadFields  field_groups.WorkloadFields
	ContainerFields field_groups.ContainerFields
	DebugStatus     common.WorkloadDebugStatus
	CommonStatus    *common.WorkloadCommonStatus  // only for non-inference workloads
	InferenceFields *field_groups.InferenceFields // optional - only for inference workloads
	InferenceStatus *common.InferenceStatus       // optional - only for inference workloads
}

func (p *WorkloadParser) ParseTrainingWorkloadToServerScheme(
	workload *twv2a1.TrainingWorkload) (*openapi.WorkloadSyncFields, []corev1.Pod, error) {
	workloadFieldsForParser := &WorkloadFieldsForParser{
		Kind: TrainingDisplayType,
		TypeMeta: metav1.TypeMeta{
			Kind:       "TrainingWorkload",
			APIVersion: twv2a1.GroupVersion.String(),
		},
		ObjectMeta:               workload.ObjectMeta,
		ResourceAllocationFields: workload.Spec.ResourceAllocationFields,
		RunaiSchedulerFields:     workload.Spec.RunaiSchedulerFields,
		SchedulingFields:         workload.Spec.SchedulingFields,
		WorkloadFields:           workload.Spec.WorkloadFields,
		ContainerFields:          workload.Spec.ContainerFields,
		DebugStatus:              workload.Status.WorkloadDebugStatus,
		CommonStatus:             &workload.Status.WorkloadCommonStatus,
	}
	result, workloadPods, err := p.parseJobWorkloadToServerScheme(workloadFieldsForParser)
	if err != nil {
		return nil, nil, err
	}

	result.RequestedPods = parseTrainRequestedPods(workload)
	result.ActionsSupport.Suspend = ptr.To(true)

	status := workload.Status.DeepCopyToJobStatus()
	p.setRunaiWorkloadStatus(&status, result)
	workload.Status.DeepCopyToTrainingStatus(status)

	return result, workloadPods, nil
}

func (p *WorkloadParser) ParseInteractiveWorkloadToServerScheme(
	workload *iwv2a1.InteractiveWorkload) (*openapi.WorkloadSyncFields, []corev1.Pod, error) {
	workloadFieldsForParser := &WorkloadFieldsForParser{
		Kind: InteractiveDisplayType,
		TypeMeta: metav1.TypeMeta{
			Kind:       "InteractiveWorkload",
			APIVersion: iwv2a1.GroupVersion.String(),
		},
		ObjectMeta:               workload.ObjectMeta,
		ResourceAllocationFields: workload.Spec.ResourceAllocationFields,
		RunaiSchedulerFields:     workload.Spec.RunaiSchedulerFields,
		SchedulingFields:         workload.Spec.SchedulingFields,
		WorkloadFields:           workload.Spec.WorkloadFields,
		ContainerFields:          workload.Spec.ContainerFields,
		DebugStatus:              workload.Status.WorkloadDebugStatus,
		CommonStatus:             &workload.Status.WorkloadCommonStatus,
	}
	result, workloadPods, err := p.parseJobWorkloadToServerScheme(workloadFieldsForParser)
	if err != nil {
		return nil, nil, err
	}

	result.RequestedPods = &openapi.RequestedPods{
		Number: getApiInt32Value(1),
	}
	result.ActionsSupport.Suspend = ptr.To(true)

	jobStatus := workload.Status.DeepCopyToJobStatus()
	p.setRunaiWorkloadStatus(&jobStatus, result)
	workload.Status.DeepCopyToInteractiveStatus(jobStatus)

	return result, workloadPods, nil
}

func (p *WorkloadParser) ParseInferenceWorkloadToServerScheme(
	workload *infwv2a1.InferenceWorkload) (*openapi.WorkloadSyncFields, []corev1.Pod, error) {
	workloadFieldsForParser := &WorkloadFieldsForParser{
		Kind: InferenceDisplayType,
		TypeMeta: metav1.TypeMeta{
			Kind:       "InferenceWorkload",
			APIVersion: infwv2a1.GroupVersion.String(),
		},
		ObjectMeta:               workload.ObjectMeta,
		ResourceAllocationFields: workload.Spec.ResourceAllocationFields,
		RunaiSchedulerFields:     workload.Spec.RunaiSchedulerFields,
		SchedulingFields:         workload.Spec.SchedulingFields,
		WorkloadFields:           workload.Spec.WorkloadFields,
		ContainerFields:          workload.Spec.ContainerFields,
		DebugStatus:              workload.Status.WorkloadDebugStatus,
		InferenceFields:          &workload.Spec.InferenceFields,
		InferenceStatus:          &workload.Status.InferenceStatus,
	}
	result, workloadPods, err := p.parseJobWorkloadToServerScheme(workloadFieldsForParser)
	if err != nil {
		return nil, nil, err
	}

	result.RequestedPods = parseInferenceRequestedPods(workload)
	result.ActionsSupport.Suspend = ptr.To(false)

	p.setRunaiInferenceWorkloadStatus(&workload.Status, result)

	return result, workloadPods, nil
}

func (p *WorkloadParser) ParseDistributedWorkloadToServerScheme(
	workload *diwv2.DistributedWorkload, clusterFrameworkExtensions extensions.Group) (*openapi.WorkloadSyncFields, []corev1.Pod, error) {
	workloadFieldsForParser := &WorkloadFieldsForParser{
		Kind: DistributedDisplayType,
		TypeMeta: metav1.TypeMeta{
			Kind:       "DistributedWorkload",
			APIVersion: diwv2.GroupVersion.String(),
		},
		ObjectMeta:               workload.ObjectMeta,
		ResourceAllocationFields: workload.Spec.ResourceAllocationFields,
		RunaiSchedulerFields:     workload.Spec.RunaiSchedulerFields,
		SchedulingFields:         workload.Spec.SchedulingFields,
		WorkloadFields:           workload.Spec.WorkloadFields,
		ContainerFields:          workload.Spec.ContainerFields,
		DebugStatus:              workload.Status.WorkloadDebugStatus,
		CommonStatus:             &workload.Status.WorkloadCommonStatus,
	}
	result, workloadPods, err := p.parseJobWorkloadToServerScheme(workloadFieldsForParser)
	if err != nil {
		return nil, nil, err
	}

	setDistributedNumOfRequestedPods(result, workload)
	setRecalculatedRequestedResourcesIfNeeded(result, workload)
	setDistributedNodeAffinities(result, workload)

	result.DistributedFramework = parseDistributedFramework(workload)

	suspendable, err := isDistributedFrameworkSuspendable(clusterFrameworkExtensions, workload.Spec.JobType)
	if err != nil {
		zap.S().Errorf("Failed to check if workload <%s/%s> is suspendable, defaulting to false. error: %s",
			workload.Namespace, workload.Name, err.Error())
		suspendable = false
	}
	result.ActionsSupport.Suspend = ptr.To(suspendable)

	jobStatus := &jobs.JobStatus{
		WorkloadCommonStatus: workload.Status.WorkloadCommonStatus,
		WorkloadDebugStatus:  workload.Status.WorkloadDebugStatus,
		WorkloadStatus:       workload.Status.WorkloadStatus,
		NetworkStatus:        workload.Status.NetworkStatus,
		SyncStatus:           workload.Status.SyncStatus,
	}
	p.setRunaiWorkloadStatus(jobStatus, result)

	return result, workloadPods, nil
}

func (p *WorkloadParser) ParseDistributedInferenceWorkloadToServerScheme(
	workload *dinfwv1.DistributedInferenceWorkload) (*openapi.WorkloadSyncFields, []corev1.Pod, error) {

	// Aggregate leader and worker specs for distributed inference
	aggregatedContainerFields := p.aggregateDistributedInferenceContainerFields(workload)
	aggregatedResourceAllocationFields := p.aggregateDistributedInferenceResourceAllocationFields(workload)
	aggregatedStorageFields := p.aggregateDistributedInferenceStorageFields(workload)

	// NodePools are taken from the general spec, the other fields are taken from the leader spec
	runaiSchedulerFields := field_groups.RunaiSchedulerFields{
		NodePools:                 workload.Spec.NodePools,
		NodeType:                  workload.LeaderSpec.NodeType,
		PodAffinity:               workload.LeaderSpec.PodAffinity,
		PodAffinityTopology:       workload.LeaderSpec.PodAffinityTopology,
		PodAffinitySchedulingRule: workload.LeaderSpec.PodAffinitySchedulingRule,
	}

	workloadFieldsForParser := &WorkloadFieldsForParser{
		Kind: "DistributedInference",
		TypeMeta: metav1.TypeMeta{
			Kind:       "DistributedInferenceWorkload",
			APIVersion: dinfwv1.GroupVersion.String(),
		},
		ObjectMeta:               workload.ObjectMeta,
		ResourceAllocationFields: aggregatedResourceAllocationFields,
		RunaiSchedulerFields:     runaiSchedulerFields,
		SchedulingFields:         workload.Spec.SchedulingFields,
		WorkloadFields: field_groups.WorkloadFields{
			IdentityFields:  workload.Spec.IdentityFields,
			ContainerFields: aggregatedContainerFields,
			StorageFields:   aggregatedStorageFields,
		},
		ContainerFields: aggregatedContainerFields,
		DebugStatus:     workload.Status.WorkloadDebugStatus,
	}
	result, pods, err := p.parseJobWorkloadToServerScheme(workloadFieldsForParser)
	if err != nil {
		return nil, nil, err
	}

	// Override the images with aggregated images from both leader and worker specs
	result.Images = p.aggregateDistributedInferenceImages(workload)

	// Set distributed inference specific fields
	setDistributedInferenceNumOfRequestedPods(result, workload)

	// Add distributed inference specific fields to additional fields
	mergeDistributedInferenceAdditionalFields(workload, result)

	// Set node affinities from leader and worker specs
	setDistributedInferenceNodeAffinities(workload, result)

	// Set status
	p.setDistributedInferenceWorkloadStatus(&workload.Status, result)

	return result, pods, nil
}

func (p *WorkloadParser) parseJobWorkloadToServerScheme(
	workload *WorkloadFieldsForParser) (*openapi.WorkloadSyncFields, []corev1.Pod, error) {
	orgData := &openapi.WorkloadSyncFields{}
	project, err := p.setOrgData(workload.Namespace, orgData)
	if err != nil {
		return nil, nil, err
	}

	requestedResources := parseWorkloadRequestedResources(&workload.ResourceAllocationFields)

	workloadId := topowner.GetRunaiWorkloadId(workload.UID, workload.Annotations)
	source := getWorkloadSource(workload.Annotations, &workload.ContainerFields)

	podsOfWorkload, err := getters.GetPodsOfWorkload(context.TODO(), workloadId, workload.Namespace, p.cachedReader)
	if err != nil {
		zap.S().Errorf("Failed getting pods of workload <%s/%s>, error: %s",
			workload.Namespace, workload.Name, err.Error())
		return nil, nil, err
	}
	podGroupsOfWorkload, err := getters.GetPodGroupsOfPods(context.TODO(), workload.Namespace, podsOfWorkload, p.cachedReader)
	if err != nil {
		zap.S().Errorf("Failed getting podgroups of workload <%s/%s>, error: %s",
			workload.Namespace, workload.Name, err.Error())
		return nil, nil, err
	}
	priorityClassName, priorityValue, preemptibility := p.getPriorityClassAndPreemptibilityOfRunaiWorkload(
		workload.Labels,
		workload.PriorityClassName,
		workload.Preemptibility,
		workload.Kind,
		workload.GroupVersionKind().Group,
		podGroupsOfWorkload,
	)
	schedulingConditions := p.getSchedulingConditions(podGroupsOfWorkload)

	childrenIds, err := p.parseRunaiWorkloadChildIds(workload, podsOfWorkload, podGroupsOfWorkload)
	if err != nil {
		return nil, nil, err
	}

	guaranteedRuntimeTimestamp := p.getGuaranteedRuntimeTimestamp(podGroupsOfWorkload, &podsOfWorkload)

	workloadSyncFields := openapi.WorkloadSyncFields{
		Type:                 workload.Kind,
		Group:                ptr.To(workload.GroupVersionKind().Group),
		Name:                 workload.Name,
		Id:                   workloadId,
		SubmittedBy:          strPointerIfNotEmpty(workload.WorkloadFields.UserName.Deref()),
		ClusterId:            config.Config().ClusterUuid,
		ProjectName:          orgData.ProjectName,
		ProjectId:            orgData.ProjectId,
		DepartmentName:       orgData.DepartmentName,
		DepartmentId:         orgData.DepartmentId,
		Namespace:            workload.Namespace,
		PriorityClassName:    priorityClassName,
		Preemptibility:       *openapi.NewNullableString(&preemptibility),
		CreatedAt:            workload.CreationTimestamp.Time,
		RequestedResources:   requestedResources,
		Source:               source,
		RequestedNodePools:   parseRequestedNodePools(workload.NodePools.Deref(), project),
		Images:               getApiImages(workload.ContainerFields.Image.Deref()),
		ChildrenIds:          childrenIds,
		ActionsSupport:       &openapi.ActionsSupport{Delete: ptr.To(true)},
		AdditionalFields:     getRunaiWorkloadAdditionalFields(workload),
		FullySaasWorkload:    cu.Ref(isControlPlaneManagedLabeled(workload.Labels)),
		EnvironmentVariables: getEnvironmentVariables(workload.WorkloadFields.Environment.Deref()),
		Command:              strPointerIfNotEmpty(workload.WorkloadFields.Command.Deref()),
		Arguments:            strPointerIfNotEmpty(workload.WorkloadFields.Arguments.Deref()),
		SchedulingConditions: schedulingConditions,
	}

	if priorityValue != nil {
		workloadSyncFields.SetPriority(*priorityValue)
	}
	if guaranteedRuntimeTimestamp != nil {
		workloadSyncFields.SetGuaranteedRuntimeEndsAt(*guaranteedRuntimeTimestamp)
	}
	runaiWorkloadKind := workload.Kind + "Workload"
	category, err := p.getWorkloadCategory(context.TODO(), runaiWorkloadKind, workload.GroupVersionKind().Group, workload.Labels)
	if err != nil {
		zap.S().Errorf("failed workload category for type %s, error: %s",
			workload.Kind, err.Error())
	}
	if category != "" {
		workloadSyncFields.Category = ptr.To(category)
	}

	fillTopology(workload.Annotations, &workloadSyncFields)
	fillNodeAffinities(&workload.NodeAffinityFields, &workloadSyncFields)

	// Native workloads don't go through the EWI, so their storage-ref UIDs are resolved
	// here. This is the same resolution the EWI now does for external workloads
	// (resolveStorageRefUIDs in commonparsers).
	nativeRefs := nativeStorageRefs(workload.Name, workload.Namespace, workload.WorkloadFields.StorageFields, workload.ContainerFields.ImagePullSecrets)
	if resolved := p.resolveStorageRefs(context.TODO(), workload.Namespace, nativeRefs); len(resolved) > 0 {
		workloadSyncFields.ResourceRefs = resolved
	}

	return &workloadSyncFields, podsOfWorkload, nil
}

func (p *WorkloadParser) setOrgData(namespaceName string, orgData *openapi.WorkloadSyncFields) (*util.ProjectData, error) {
	project, err := util.GetProjectFromNamespaceName(namespaceName, p.cachedReader)
	if err != nil {
		return nil, err
	}

	projectId, err := util.GetProjectId(project)
	if err != nil {
		return nil, err
	}

	departmentId, err := util.GetDepartmentId(project)
	if err != nil {
		return nil, err
	}

	orgData.ProjectName = project.Name
	orgData.ProjectId = projectId
	orgData.DepartmentName = project.DepartmentName
	orgData.DepartmentId = departmentId
	return project, nil
}

func parseWorkloadRequestedResources(
	resourceAllocationFields *field_groups.ResourceAllocationFields) openapi.NullableWorkloadRequestResources {
	result := openapi.WorkloadRequestResources{}
	nullableResult := openapi.NullableWorkloadRequestResources{}

	if resourceAllocationFields == nil {
		return nullableResult
	}

	if resourceAllocationFields.Gpu != nil || resourceAllocationFields.GpuLimit != nil {
		result.GpuRequestType = getGpuRequestType(openapi.GPUREQUESTTYPE_PORTION)
		// For native IW/TW, spec.gpu holds the per-device fraction and spec.gpuDevices
		// holds the device count; the total workload GPU request is fraction * devices.
		// For distributed workloads, `unifyGpuFields` nils GpuDevices upstream so this
		// path reduces to just returning the Gpu value.
		numDevices := int32(1)
		if resourceAllocationFields.GpuDevices != nil {
			numDevices = int32(resourceAllocationFields.GpuDevices.Deref())
		}
		result.Gpu.Set(&openapi.RequestResourceCores{
			Request: scaleGpuFractionByDevices(resourceAllocationFields.Gpu, numDevices),
			Limit:   scaleGpuFractionByDevices(resourceAllocationFields.GpuLimit, numDevices),
		})
	} else if resourceAllocationFields.GpuDevices != nil {
		result.GpuRequestType = getGpuRequestType(openapi.GPUREQUESTTYPE_PORTION)
		result.Gpu.Set(&openapi.RequestResourceCores{
			Request: getApiFloat64IntField(resourceAllocationFields.GpuDevices),
			Limit:   getApiFloat64NumberField(resourceAllocationFields.GpuLimit),
		})
	}

	gpuMemory := resourceAllocationFields.GpuMemory.Deref()
	if gpuMemory != "" {
		result.GpuRequestType = getGpuRequestType(openapi.GPUREQUESTTYPE_MEMORY)

		gpuMemoryLimit := resourceAllocationFields.GpuMemoryLimit.Deref()
		if gpuMemoryLimit == "" {
			gpuMemoryLimit = gpuMemory
		}

		result.GpuMemory.Set(&openapi.RequestResourceQuantity{
			Request: getApiString(gpuMemory),
			Limit:   getApiString(gpuMemoryLimit),
		})
	}

	if resourceAllocationFields.Cpu != nil || resourceAllocationFields.CpuLimit != nil {
		result.Cpu.Set(&openapi.RequestResourceCores{
			Request: getApiFloat64NumberField(resourceAllocationFields.Cpu),
			Limit:   getApiFloat64NumberField(resourceAllocationFields.CpuLimit),
		})
	}

	memory := resourceAllocationFields.Memory.Deref()
	memoryLimit := resourceAllocationFields.MemoryLimit.Deref()
	if memory != "" || memoryLimit != "" {
		result.CpuMemory.Set(&openapi.RequestResourceQuantity{
			Request: getApiString(memory),
			Limit:   getApiString(memoryLimit),
		})
	}

	if resourceAllocationFields.ExtendedResources != nil {
		extendedResources := resourceAllocationFields.ExtendedResources.Deref()
		result.ExtendedResources = []openapi.WorkloadsExtendedResource{}
		for _, extendedResource := range extendedResources {
			if extendedResource.Value != nil && extendedResource.Value.Name != nil && extendedResource.Value.Quantity != nil {
				result.ExtendedResources = append(result.ExtendedResources, openapi.WorkloadsExtendedResource{
					Resource: *openapi.NewNullableString(extendedResource.Value.Name),
					Quantity: *openapi.NewNullableString(extendedResource.Value.Quantity),
				})
			}
		}
	}

	nullableResult.Set(&result)
	return nullableResult
}

func fillTopology(annotations map[string]string, workload *openapi.WorkloadSyncFields) {
	if annotations == nil || workload == nil {
		return
	}

	topologyName, foundName := annotations[clusterconstants.TopologyAnnotation]
	topologyPreferred, foundPreferred := annotations[clusterconstants.TopologyPreferredPlacementAnnotation]
	topologyRequired, foundRequired := annotations[clusterconstants.TopologyRequiredPlacementAnnotation]

	if topologyName == "" && topologyPreferred == "" && topologyRequired == "" {
		return
	}

	topology := openapi.WorkloadSyncFieldsTopology{}
	if foundName && topologyName != "" {
		topology.SetName(topologyName)
	}
	if foundPreferred && topologyPreferred != "" {
		topology.SetPreferredPlacement(topologyPreferred)
	}
	if foundRequired && topologyRequired != "" {
		topology.SetRequiredPlacement(topologyRequired)
	}
	workload.SetTopology(topology)
}

func (p *WorkloadParser) parseRunaiWorkloadChildIds(workload *WorkloadFieldsForParser,
	podsOfWorkload []corev1.Pod, podGroupsOfWorkload []*kubeaischedv2alpha2.PodGroup) ([]openapi.WorkloadChildrenIdsInner, error) {
	// add self as a child - because sometimes the original UID is overridden by control-plane workload ID
	result := []openapi.WorkloadChildrenIdsInner{{Id: string(workload.UID), Type: ptr.To(workload.Kind)}}

	for _, owner := range workload.DebugStatus.OwnedResources {
		if owner.Uid == "" || owner.Kind.Kind == "" {
			continue
		}
		result = append(result, openapi.WorkloadChildrenIdsInner{
			Id:   string(owner.Uid),
			Type: ptr.To(owner.Kind.Kind),
		})
	}

	result = append(result, p.getPodsChildren(podsOfWorkload)...)
	result = append(result, p.getPodGroupsChildren(podGroupsOfWorkload)...)

	if workload.Kind == InferenceDisplayType {
		inferenceChildren, err := p.getInferenceWorkloadChildren(workload.Name, workload.Namespace)
		if err != nil {
			return nil, err
		}

		result = append(result, inferenceChildren...)
	}

	return result, nil
}

func (p *WorkloadParser) getPodsChildren(podsOfWorkload []corev1.Pod) []openapi.WorkloadChildrenIdsInner {
	result := []openapi.WorkloadChildrenIdsInner{}

	for _, pod := range podsOfWorkload {
		result = append(result, openapi.WorkloadChildrenIdsInner{
			Id:   string(pod.UID),
			Type: ptr.To("Pod"),
		})
	}

	return result
}

func (p *WorkloadParser) getPodGroupsChildren(podGroupsOfWorkload []*kubeaischedv2alpha2.PodGroup) []openapi.WorkloadChildrenIdsInner {
	result := []openapi.WorkloadChildrenIdsInner{}

	for _, podGroup := range podGroupsOfWorkload {
		result = append(result, openapi.WorkloadChildrenIdsInner{
			Id:   string(podGroup.UID),
			Type: ptr.To("PodGroup"),
		})
	}

	return result
}

// getInferenceWorkloadChildren returns the inference-specific children to add to the
// children id list. it gets the knative configuration and revision(s).
// the parent knative service resource is owned by the inference workload and is thus
// assumed to already be part of the children via the workload's ownedResources
func (p *WorkloadParser) getInferenceWorkloadChildren(workloadName string, workloadNamespace string) ([]openapi.WorkloadChildrenIdsInner, error) {
	result := []openapi.WorkloadChildrenIdsInner{}

	// configuration
	configuration := &knv1.Configuration{}
	err := p.cachedReader.Get(context.TODO(), client.ObjectKey{Namespace: workloadNamespace, Name: workloadName}, configuration)
	if client.IgnoreNotFound(err) != nil {
		return nil, err
	}

	if configuration != nil && configuration.GetName() != "" {
		result = append(result, openapi.WorkloadChildrenIdsInner{
			Id:   string(configuration.UID),
			Type: ptr.To("Configuration"),
		})
	}

	// revisions
	revisions := &knv1.RevisionList{}
	labels := map[string]string{clusterconstants.LabelRelease: workloadName}
	err = p.cachedReader.List(context.TODO(), revisions, client.InNamespace(workloadNamespace), client.MatchingLabels(labels))
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			return nil, err
		}
		return result, nil
	}

	for _, revision := range revisions.Items {
		if revision.GetName() != "" {
			result = append(result, openapi.WorkloadChildrenIdsInner{
				Id:   string(revision.UID),
				Type: ptr.To("Revision"),
			})
		}
	}

	return result, nil
}

func (p *WorkloadParser) getPriorityClassAndPreemptibilityOfRunaiWorkload(
	workloadLabels map[string]string,
	priorityClassNameField *field_types.StringField,
	preemptibilityField *field_types.StringField,
	runaiWorkloadKind string,
	workloadGroup string,
	podGroupsOfWorkload []*kubeaischedv2alpha2.PodGroup) (string, *int32, string) {
	workloadKind := runaiWorkloadKind + "Workload"
	priorityClassName, priorityValue, workloadTypeConfigOfKind := p.getPriorityClassOfWorkload(workloadLabels, priorityClassNameField.Deref(), workloadKind, workloadGroup, podGroupsOfWorkload)
	preemptibility := preemptibilityField.Deref()
	if preemptibility == "" {
		preemptibility = p.getPreemptibilityOfWorkload(workloadLabels, workloadKind, workloadGroup, podGroupsOfWorkload, workloadTypeConfigOfKind)
	}
	return priorityClassName, priorityValue, preemptibility
}

func (p *WorkloadParser) getPriorityClassOfWorkload(
	workloadLabels map[string]string,
	specPriorityClassName string,
	workloadKind string,
	workloadGroup string,
	podGroupsOfWorkload []*kubeaischedv2alpha2.PodGroup) (string, *int32, *workloadTypeConfig.WorkloadTypeConfig) {
	if specPriorityClassName != "" {
		return specPriorityClassName, p.getPriorityClassValueOrNil(specPriorityClassName), nil
	}
	if priorityClassName, found := workloadLabels[clusterconstants.LabelPriorityClassName]; found && priorityClassName != "" {
		return priorityClassName, p.getPriorityClassValueOrNil(priorityClassName), nil
	}

	priorityClassNamePtr, priorityValuePtr := p.getPriorityClassFromPodGroups(podGroupsOfWorkload)
	if priorityClassNamePtr != nil && priorityValuePtr != nil {
		return *priorityClassNamePtr, priorityValuePtr, nil
	}

	typeConfig, err := workloadTypeConfig.GetWorkloadConfigByGroupKind(context.TODO(), p.cachedReader, workloadGroup, workloadKind)
	if err != nil {
		zap.S().Errorf("Failed to get workload type config for <%s>, error: %s", workloadKind, err.Error())
		return priority.TrainPriorityClassName, ptr.To(int32(priority.TrainPriorityValue)), nil
	}

	return typeConfig.PriorityName, p.getPriorityClassValueOrNil(typeConfig.PriorityName), typeConfig
}

func (p *WorkloadParser) getPreemptibilityOfWorkload(
	workloadLabels map[string]string,
	workloadKind string,
	workloadGroup string,
	podGroupsOfWorkload []*kubeaischedv2alpha2.PodGroup,
	workloadTypeConfigOfKind *workloadTypeConfig.WorkloadTypeConfig) string {
	if preemptibility, found := workloadLabels[clusterconstants.LabelPreemptibilityName]; found && preemptibility != "" {
		return preemptibility
	}

	preemptibility := p.getPreemptibilityFromPodGroups(podGroupsOfWorkload)
	if preemptibility != "" {
		return preemptibility
	}

	if workloadTypeConfigOfKind != nil {
		return workloadTypeConfigOfKind.Preemptibility
	}

	typeConfig, err := workloadTypeConfig.GetWorkloadConfigByGroupKind(context.TODO(), p.cachedReader, workloadGroup, workloadKind)
	if err != nil {
		zap.S().Errorf("Failed to get workload type config for <%s>, error: %s", workloadKind, err.Error())
		return ""
	}

	return typeConfig.Preemptibility
}

func (p *WorkloadParser) getSchedulingConditions(podGroupsOfWorkload []*kubeaischedv2alpha2.PodGroup) []openapi.SchedulingCondition {
	// sort first, so the calculated "first" is always the same
	sortPodGroupsByName(podGroupsOfWorkload)

	firstUnscheduleablePodGroup := getFirstUnscheduleablePodGroup(podGroupsOfWorkload)
	if firstUnscheduleablePodGroup == nil {
		return nil
	}

	result := []openapi.SchedulingCondition{}
	for _, schedulingCondition := range firstUnscheduleablePodGroup.Status.SchedulingConditions {
		for i := range schedulingCondition.Reasons {
			unschedulableExplanation := schedulingCondition.Reasons[i]
			reason := p.getApiPhaseReason(string(unschedulableExplanation.Reason))
			if reason == nil {
				continue
			}

			apiSchedulingCondition := openapi.SchedulingCondition{
				NodePool: schedulingCondition.NodePool,
				Reason:   *reason,
			}
			p.parseUnschedulableExplanationDetails(firstUnscheduleablePodGroup.Spec.Queue, &unschedulableExplanation, &apiSchedulingCondition)
			result = append(result, apiSchedulingCondition)
		}
	}

	return result
}

func getFirstUnscheduleablePodGroup(podGroupsOfWorkload []*kubeaischedv2alpha2.PodGroup) *kubeaischedv2alpha2.PodGroup {
	var firstUnscheduleablePodGroup *kubeaischedv2alpha2.PodGroup
	for _, podGroup := range podGroupsOfWorkload {
		if podGroup == nil {
			continue
		}

		if IsPodGroupUnscheduleable(podGroup) {
			firstUnscheduleablePodGroup = podGroup
			break
		}

		return nil
	}
	if firstUnscheduleablePodGroup == nil {
		return nil
	}
	return firstUnscheduleablePodGroup
}

func (p *WorkloadParser) setRunaiWorkloadStatus(status *jobs.JobStatus, result *openapi.WorkloadSyncFields) {
	result.Phase = p.getApiPhase(status.WorkloadCommonStatus.Phase)
	result.PhaseMessage = strPointerIfNotEmpty(status.WorkloadCommonStatus.PhaseMessage)

	conditions := []openapi.Condition{}
	for _, condition := range status.Conditions {
		if condition.Type == "" || condition.Status == "" {
			continue
		}

		conditions = append(conditions, openapi.Condition{
			Type:               string(condition.Type),
			Status:             string(condition.Status),
			Message:            strPointerIfNotEmpty(condition.Message),
			Reason:             strPointerIfNotEmpty(condition.Reason),
			LastTransitionTime: getApiTime(condition.LastTransitionTime),
		})
	}
	result.Conditions = conditions

	result.Urls = status.WorkloadCommonStatus.ExternalUrls

	setNetworkStatus(status.NetworkStatus, result)
}

func (p *WorkloadParser) setRunaiInferenceWorkloadStatus(status *infwv2a1.InferenceWorkloadStatus, result *openapi.WorkloadSyncFields) {
	phase := getPhaseFromInferencePhase(status.Phase)
	result.Phase = p.getApiPhase(phase)
	result.PhaseMessage = strPointerIfNotEmpty(status.PhaseMessage)

	result.Urls = status.ExternalUrls

	setNetworkStatus(status.NetworkStatus, result)
}

func setNetworkStatus(networkStatus common.NetworkStatus, result *openapi.WorkloadSyncFields) {
	if networkStatus.ServiceStatuses == nil {
		return
	}

	services := make([]openapi.WorkloadNetworkStatusServicesInner, 0, len(networkStatus.ServiceStatuses))
	for _, service := range networkStatus.ServiceStatuses {
		services = append(services, openapi.WorkloadNetworkStatusServicesInner{
			Name:         strPointerIfNotEmpty(service.Name),
			Type:         strPointerIfNotEmpty(string(service.Type)),
			Ports:        convertPorts(service.Ports),
			LoadBalancer: convertLoadBalancer(service.LoadBalancer),
		})
	}

	result.NetworkStatus = &openapi.WorkloadNetworkStatus{
		Services: services,
	}
}

func convertPorts(ports []common.ExposePort) []openapi.WorkloadNetworkStatusServicesInnerPortsInner {
	if len(ports) == 0 {
		return nil
	}

	result := make([]openapi.WorkloadNetworkStatusServicesInnerPortsInner, 0, len(ports))
	for _, port := range ports {
		result = append(result, openapi.WorkloadNetworkStatusServicesInnerPortsInner{
			Name:       strPointerIfNotEmpty(port.Name),
			TargetPort: strPointerIfNotEmpty(port.TargetPort.String()),
			Port:       int32PointerIfNotEmpty(port.Port),
			Protocol:   strPointerIfNotEmpty(string(port.Protocol)),
		})
	}
	return result
}

func convertLoadBalancer(loadBalancerStatus *common.LoadBalancerStatus) *openapi.WorkloadNetworkStatusServicesInnerLoadBalancer {
	if loadBalancerStatus == nil || len(loadBalancerStatus.Ingress) == 0 {
		return nil
	}

	ingress := make([]openapi.WorkloadNetworkStatusServicesInnerLoadBalancerIngressInner, 0, len(loadBalancerStatus.Ingress))
	for _, i := range loadBalancerStatus.Ingress {
		ingress = append(ingress, openapi.WorkloadNetworkStatusServicesInnerLoadBalancerIngressInner{
			Ip:       strPointerIfNotEmpty(i.IP),
			Hostname: strPointerIfNotEmpty(i.Hostname),
		})
	}

	return &openapi.WorkloadNetworkStatusServicesInnerLoadBalancer{
		Ingress: ingress,
	}
}

func (p *WorkloadParser) getApiPhase(statusPhase string) openapi.Phase {
	if statusPhase == "" {
		return openapi.PHASE_CREATING
	}
	phase := p.toControlPlanePhase(statusPhase)
	if _, found := p.allowedWorkloadPhases[phase]; !found {
		zap.S().Warnf("Failed to parse status %s of workload.",
			statusPhase)
		phase = openapi.PHASE_UNKNOWN
	}
	return phase
}

func (p *WorkloadParser) toControlPlanePhase(statusPhase string) openapi.Phase {
	switch statusPhase {
	case common.PhaseSuspending.String():
		return openapi.PHASE_STOPPING
	case common.PhaseSuspended.String():
		return openapi.PHASE_STOPPED
	default:
		return openapi.Phase(statusPhase)
	}
}

func (p *WorkloadParser) getApiPhaseReason(phaseReason string) *openapi.PhaseReason {
	if phaseReason == "" {
		return nil
	}

	apiPhaseReason := openapi.PhaseReason(phaseReason)
	if _, found := p.allowedPhaseReasons[apiPhaseReason]; found {
		return &apiPhaseReason
	}
	return nil
}

func getPhaseFromInferencePhase(phase string) string {
	// map the inference phases to the general ones, as the API should
	// be generic and not use inference-specific values

	switch phase {
	case "Ready", "Progressing":
		return "Running"
	}

	// all the other values exist in the general API enum
	return phase
}

func getWorkloadSource(workloadAnnotations map[string]string, containerFields *field_groups.ContainerFields) openapi.Source {
	// CLI V2
	if containerFields != nil {
		_, found := getCliCommandAnnotation(workloadAnnotations, containerFields, constants.WorkloadCliV2Command)
		if found {
			return openapi.SOURCE_CLI
		}
	}

	workloadId, found := workloadAnnotations[commonconstants.ControlPlaneWorkloadIdAnnotationForWorkload]
	if found && workloadId != "" {
		return openapi.SOURCE_CONTROL_PLANE
	}

	return openapi.SOURCE_OTHER
}

func hasTrueWorkloadLabel(labels *field_types.DictionaryField, label string) bool {
	if labels == nil {
		return false
	}

	item, ok := labels.Deref()[label]
	if !ok || item.Value == nil {
		return false
	}

	return ptr.Deref(item.Value, "") == "true"
}

func getCliCommand(workloadAnnotations map[string]string, containerFields field_groups.ContainerFields) string {
	cliCommand, found := getCliCommandAnnotation(workloadAnnotations, &containerFields, constants.WorkloadCliV2Command)
	if found {
		return cliCommand
	}

	return ""
}

func getCliCommandAnnotation(workloadAnnotations map[string]string, containerFields *field_groups.ContainerFields,
	annotationName string) (string, bool) {
	cliCommand, found := workloadAnnotations[annotationName]
	if found && cliCommand != "" {
		return cliCommand, true
	}

	annotationValues := containerFields.Annotations.Deref()
	cliCommandItemAndRules, found := annotationValues[annotationName]
	if !found || cliCommandItemAndRules.Value == nil {
		return "", found
	}
	return *cliCommandItemAndRules.Value, true
}

func getRunaiWorkloadAdditionalFields(workload *WorkloadFieldsForParser) *string {
	additionalFields := map[string]string{}

	nimWorkload := hasTrueWorkloadLabel(workload.WorkloadFields.Labels, clusterconstants.LabelNIMWorkload)
	if nimWorkload {
		additionalFields[clusterconstants.AdditionalNIMWorkload] = "true"
	}
	if hasTrueWorkloadLabel(workload.WorkloadFields.Labels, clusterconstants.LabelVLLMWorkload) {
		additionalFields[clusterconstants.AdditionalVLLMWorkload] = "true"
	}

	// add the CLI command
	cliCommand := getCliCommand(workload.Annotations, workload.ContainerFields)
	if cliCommand != "" {
		additionalFields[additionalCliCommand] = cliCommand
	}

	// for inference, add the revision names
	if workload.InferenceStatus != nil {
		if workload.InferenceStatus.LatestCreatedRevisionName != nil {
			additionalFields[additionalLatestCreatedRevisionName] = *workload.InferenceStatus.LatestCreatedRevisionName
		}

		if workload.InferenceStatus.LatestReadyRevisionName != nil {
			additionalFields[additionalLatestReadyRevisionName] = *workload.InferenceStatus.LatestReadyRevisionName
		}
	}

	if workload.CommonStatus != nil {
		if workload.CommonStatus.Issues != nil {
			issuesStr, err := getWorkloadIssues(*workload.CommonStatus.Issues)
			if err != nil {
				zap.S().Errorf("Failed to marshal workload issues. error: %s", err.Error())
			} else {
				additionalFields[additionalIssues] = issuesStr
			}
		}

		// added for distributed only
		if workload.Kind == DistributedDisplayType {
			gpuAccelerationStatus, err := getWorkloadGpuAccelerationStatus(workload.CommonStatus.Conditions)
			if err != nil {
				zap.S().Errorf("Failed to marshal workload gpuAccelerationStatus. error: %s", err.Error())
			} else {
				additionalFields[additionalGpuAccelerationStatus] = gpuAccelerationStatus
			}
		}
	}

	return getWorkloadAdditionalFields(additionalFields)
}

func getWorkloadIssues(issues common.Issues) (string, error) {
	issuesStr, err := json.Marshal(issues)
	if err != nil {
		return "", err
	}

	return string(issuesStr), nil
}

func getWorkloadGpuAccelerationStatus(conditions []common.WorkloadCondition) (string, error) {
	var gpuAcceleration GPUNetworkAcceleration
	for _, condition := range conditions {
		if condition.Type == common.GPUNetworkAccelerated && condition.Status == True {
			gpuAcceleration.Used = true
		}
	}
	data, err := json.Marshal(gpuAcceleration)
	return string(data), err
}

func parseTrainRequestedPods(workload *twv2a1.TrainingWorkload) *openapi.RequestedPods {
	var number, parallelism, completions *int32 = nil, nil, nil

	if workload.Spec.Parallelism != nil && workload.Spec.Parallelism.Value != nil {
		parallelism = ptr.To(int32(*workload.Spec.Parallelism.Value))
	}
	if workload.Spec.Completions != nil && workload.Spec.Completions.Value != nil {
		completions = ptr.To(int32(*workload.Spec.Completions.Value))
	}

	if parallelism == nil && completions == nil {
		number = ptr.To(int32(1))
	}

	result := &openapi.RequestedPods{
		Number:      getApiInt32(number),
		Parallelism: getApiInt32(parallelism),
		Completions: getApiInt32(completions),
	}

	return result
}

func parseInferenceRequestedPods(workload *infwv2a1.InferenceWorkload) *openapi.RequestedPods {
	var maxPods *int32
	var minPods *int32

	if workload.Spec.MinScale != nil {
		minPods = ptr.To(workload.Spec.MinScale.Deref32())
	} else {
		// the default is 0 if KPA is used, 1 otherwise.
		// KPA is used if the metric is set to concurrency (which is also the default) or throughput

		minPods = ptr.To(int32(0))

		if workload.Spec.Metric != nil {
			if workload.Spec.Metric.Deref() != composer.AutoScalingMetric_THROUGHPUT &&
				workload.Spec.Metric.Deref() != composer.AutoScalingMetric_CONCURRENCY {
				minPods = ptr.To(int32(1))
			}
		}
	}

	if workload.Spec.MaxScale != nil {
		maxPods = ptr.To(workload.Spec.MaxScale.Deref32())
	}

	result := &openapi.RequestedPods{
		//Number: getApiInt32(numPods), // TODO not setting Number - make sure it's OK
		Min: getApiInt32(minPods),
		Max: getApiInt32(maxPods),
	}

	return result
}

func setDistributedNumOfRequestedPods(result *openapi.WorkloadSyncFields, workload *diwv2.DistributedWorkload) {
	numberOfRequestedPods := workload.Spec.Workers.Deref()

	if !workload.Spec.NoMaster.Deref() {
		numberOfRequestedPods += 1
	}

	result.RequestedPods = &openapi.RequestedPods{
		Number: getApiInt32Value(int32(numberOfRequestedPods)),
	}
}

func setRecalculatedRequestedResourcesIfNeeded(result *openapi.WorkloadSyncFields, workload *diwv2.DistributedWorkload) {
	allRequestedResourcesForWorkload := (&workload.Spec.ResourceAllocationFields).DeepCopy()
	resourcesChanged := false

	numberOfRequestedPods := workload.Spec.Workers.Deref()
	if numberOfRequestedPods != 1 {
		allRequestedResourcesForWorkload = MultiplyResourceAllocationFields(allRequestedResourcesForWorkload, numberOfRequestedPods)
		resourcesChanged = true
	}

	if !workload.Spec.NoMaster.Deref() {
		if workload.MasterSpec != nil {
			allRequestedResourcesForWorkload = AddResourceAllocationFields(allRequestedResourcesForWorkload, &workload.MasterSpec.ResourceAllocationFields)
			resourcesChanged = true
		}
	}

	if resourcesChanged {
		requestedResources := parseWorkloadRequestedResources(allRequestedResourcesForWorkload)
		result.RequestedResources = requestedResources
	}
}

func parseRequestedNodePools(workloadNodePools string, project *util.ProjectData) []string {
	if workloadNodePools != "" {
		return strings.Split(workloadNodePools, NodePoolsSeparator)
	}

	if project != nil && project.DefaultNodePools != nil && len(project.DefaultNodePools) > 0 {
		return project.DefaultNodePools
	}

	return []string{v1alpha1.DefaultNodePoolName}
}

func parseDistributedFramework(workload *diwv2.DistributedWorkload) *string {
	return strPointerIfNotEmpty(workload.Spec.JobType)
}

func getEnvironmentVariables(envvars map[string]field_types.DictionaryItemAndRules) *map[string]string {
	result := map[string]string{}

	for _, key := range field_types.SortedDictionaryMapKeys(envvars) {
		value := envvars[key]
		if !ptr.Deref(value.Deleted, false) {
			result[key] = ptr.Deref(value.Value, "")
		}
	}

	if len(result) == 0 {
		return nil
	}
	return &result
}

func isControlPlaneManagedLabeled(labels map[string]string) bool {
	_, ok := labels[commonconstants.ControlPlaneManagedLabel]
	return ok
}

func (p *WorkloadParser) parseUnschedulableExplanationDetails(podGroupSpecQueue string, explanation *kubeaischedv2alpha2.UnschedulableExplanation,
	result *openapi.SchedulingCondition) {
	if explanation == nil || explanation.Details == nil {
		return
	}

	if explanation.Details.QueueDetails == nil {
		return
	}

	result.QueueDetails = p.parseQueueDetails(podGroupSpecQueue, explanation)
}

func (p *WorkloadParser) parseQueueDetails(podGroupSpecQueue string, explanation *kubeaischedv2alpha2.UnschedulableExplanation) openapi.NullableSchedulingConditionQueueDetails {
	details := explanation.Details.QueueDetails
	result := openapi.SchedulingConditionQueueDetails{
		OrgType:                                  p.getOrgTypeFromQueueName(podGroupSpecQueue, details.Name),
		ExceedingResources:                       parseResourcesForSchedulingCondition(getExceedingResources(explanation)),
		QueueRequestedResources:                  parseResourcesForSchedulingCondition(details.QueueRequestedResources),
		QueueDeservedResources:                   parseResourcesForSchedulingCondition(details.QueueDeservedResources),
		QueueAllocatedResources:                  parseResourcesForSchedulingCondition(details.QueueAllocatedResources),
		QueueAllocatedNonPreemptibleResources:    parseResourcesForSchedulingCondition(details.QueueAllocatedNonPreemptibleResources),
		QueueResourceLimits:                      parseResourcesForSchedulingCondition(details.QueueResourceLimits),
		PodGroupRequestedResources:               parseResourcesForSchedulingCondition(details.PodGroupRequestedResources),
		PodGroupRequestedNonPreemptibleResources: parseResourcesForSchedulingCondition(details.PodGroupRequestedNonPreemptibleResources),
	}

	nullableResult := openapi.NullableSchedulingConditionQueueDetails{}
	nullableResult.Set(&result)
	return nullableResult
}

func getExceedingResources(explanation *kubeaischedv2alpha2.UnschedulableExplanation) corev1.ResourceList {
	exceedingResources, err := unschedulable.ParseExceedingResources(*explanation)
	if err != nil {
		zap.S().Errorf("Failed parsing exceeding resources, error: %s", err.Error())
		return corev1.ResourceList{}
	}
	return exceedingResources
}

func (p *WorkloadParser) getOrgTypeFromQueueName(podGroupSpecQueue string, detailsQueueName string) openapi.OrgType {
	if podGroupSpecQueue == detailsQueueName {
		// if the queue is the same as the queue set for the podgroup - it means it's the project's queue
		return openapi.ORGTYPE_PROJECT
	}

	queue := &kaiv2.Queue{}
	err := p.cachedReader.Get(context.TODO(), types.NamespacedName{Name: detailsQueueName}, queue)
	if err != nil {
		zap.S().Errorf("Failed getting queue <%s>, error: %s", detailsQueueName, err.Error())
		return openapi.ORGTYPE_PROJECT
	}

	if _, found := queue.Labels[clusterconstants.RunaiProjectLabel]; found {
		return openapi.ORGTYPE_PROJECT
	}

	return openapi.ORGTYPE_DEPARTMENT
}

func getNumWorkers(workload *dinfwv1.DistributedInferenceWorkload) int32 {
	workers := int32(0) // Default to 0 workers
	if workload.Spec.Workers != nil && workload.Spec.Workers.Value != nil {
		workers = int32(*workload.Spec.Workers.Value)
	}
	return workers
}

func getNumReplicas(workload *dinfwv1.DistributedInferenceWorkload) int32 {
	replicas := int32(1) // Default to 1 replica
	if workload.Spec.Replicas != nil && workload.Spec.Replicas.Value != nil {
		replicas = int32(*workload.Spec.Replicas.Value)
	}
	return replicas
}

func setDistributedInferenceNumOfRequestedPods(result *openapi.WorkloadSyncFields, workload *dinfwv1.DistributedInferenceWorkload) {
	// Calculate total pods: 1 leader + workers * replicas
	workers := getNumWorkers(workload)
	replicas := getNumReplicas(workload)

	totalPods := (workers + 1) * replicas // (1 leader + workers) * replicas

	result.RequestedPods = &openapi.RequestedPods{
		Number: getApiInt32Value(totalPods),
	}
}

//nolint:unused
func _addGpuResourcesToRequestedResources(
	resourceAllocationFields *field_groups.ResourceAllocationFields,
	requestedResources *openapi.WorkloadRequestResources,
	multiplier int32) {

	if resourceAllocationFields.Gpu == nil && resourceAllocationFields.GpuLimit == nil {
		return
	}

	if !requestedResources.Gpu.IsSet() {
		requestedResources.Gpu.Set(&openapi.RequestResourceCores{})
	}

	gpuCores := requestedResources.Gpu.Get()

	// Get GPU resources from the allocation fields
	gpuRequest := getApiFloat64NumberField(resourceAllocationFields.Gpu)
	gpuLimit := getApiFloat64NumberField(resourceAllocationFields.GpuLimit)

	// Add GPU request
	if gpuRequest.IsSet() {
		gpuTotal := *gpuRequest.Get() * float64(multiplier)
		if gpuCores.Request.IsSet() {
			gpuTotal += *gpuCores.Request.Get()
		}
		gpuCores.Request = _getApiFloat64Value(gpuTotal)
	}

	// Add GPU limit
	if gpuLimit.IsSet() {
		gpuLimitTotal := *gpuLimit.Get() * float64(multiplier)
		if gpuCores.Limit.IsSet() {
			gpuLimitTotal += *gpuCores.Limit.Get()
		}
		gpuCores.Limit = _getApiFloat64Value(gpuLimitTotal)
	}

	requestedResources.Gpu.Set(gpuCores)
}

func mergeDistributedInferenceAdditionalFields(workload *dinfwv1.DistributedInferenceWorkload, result *openapi.WorkloadSyncFields) {
	// Get existing additional fields
	var existingFields map[string]interface{}
	if result.AdditionalFields != nil && *result.AdditionalFields != "" {
		if err := json.Unmarshal([]byte(*result.AdditionalFields), &existingFields); err != nil {
			zap.S().Errorf("Failed to unmarshal existing additional fields: %s", err.Error())
			existingFields = make(map[string]interface{})
		}
	} else {
		existingFields = make(map[string]interface{})
	}

	// Get distributed inference fields
	diFields := getDistributedInferenceFields(workload)

	// Merge the fields
	for key, value := range diFields {
		existingFields[key] = value
	}

	// Marshal back to JSON and set the result
	jsonBytes, err := json.Marshal(existingFields)
	if err != nil {
		zap.S().Errorf("Failed to marshal merged additional fields: %s", err.Error())
		return
	}

	additionalFieldsStr := string(jsonBytes)
	result.AdditionalFields = &additionalFieldsStr
}

func getDistributedInferenceFields(workload *dinfwv1.DistributedInferenceWorkload) map[string]interface{} {
	// Build the distributed inference fields with proper nil checks
	diFields := map[string]interface{}{}

	// Add workers field if it exists
	if workload.Spec.Workers != nil && workload.Spec.Workers.Value != nil {
		diFields["workers"] = *workload.Spec.Workers.Value
	}

	// Add specReplicas field if it exists
	if workload.Spec.Replicas != nil && workload.Spec.Replicas.Value != nil {
		diFields["specReplicas"] = *workload.Spec.Replicas.Value
	}

	// Add statusReplicas field if it exists
	if workload.Status.Replicas > 0 {
		diFields["statusReplicas"] = workload.Status.Replicas
	}

	// Add readyReplicas field if status exists
	if workload.Status.ReadyReplicas > 0 {
		diFields["readyReplicas"] = workload.Status.ReadyReplicas
	}

	// Add servingPorts field if it exists - but remove the .value wrapper
	if workload.Spec.ServingPorts != nil && workload.Spec.ServingPorts.Value != nil {
		servingPorts := make(map[string]interface{})
		for name, portItem := range *workload.Spec.ServingPorts.Value {
			if portItem.Value != nil {
				portValue := map[string]interface{}{}
				if portItem.Value.Port != nil {
					portValue["port"] = *portItem.Value.Port
				}
				if portItem.Value.Protocol != nil {
					portValue["protocol"] = *portItem.Value.Protocol
				}
				if portItem.Value.ExposeExternally != nil {
					portValue["exposeExternally"] = *portItem.Value.ExposeExternally
				}
				if portItem.Value.AuthorizationType != nil {
					portValue["authorizationType"] = *portItem.Value.AuthorizationType
				}
				servingPorts[name] = portValue
			}
		}
		diFields["servingPorts"] = servingPorts
	}

	// Add RestartPolicy field if it exists
	if workload.Spec.RestartPolicy != nil && workload.Spec.RestartPolicy.Value != nil {
		diFields["restartPolicy"] = *workload.Spec.RestartPolicy.Value
	}

	// Add StartupPolicy field if it exists
	if workload.Spec.StartupPolicy != nil && workload.Spec.StartupPolicy.Value != nil {
		diFields["startupPolicy"] = *workload.Spec.StartupPolicy.Value
	}

	return diFields
}

func (p *WorkloadParser) setDistributedInferenceWorkloadStatus(status *dinfwv1.DistributedInferenceWorkloadStatus, result *openapi.WorkloadSyncFields) {
	// Set phase
	phase := getPhaseFromInferencePhase(status.Phase)
	result.Phase = p.getApiPhase(phase)
	result.PhaseMessage = strPointerIfNotEmpty(status.PhaseMessage)

	// Set k8sPhase - use the phase as k8sPhase for now
	result.SetK8sPhase(status.Phase)

	// Set urls from DistributedInferenceStatus
	var urls []string
	urls = append(urls, status.ServingUrls...)
	urls = append(urls, status.ExternalUrls...)
	if len(urls) > 0 {
		result.Urls = urls
	}

	// Set network status
	setNetworkStatus(status.NetworkStatus, result)
}

func (p *WorkloadParser) aggregateDistributedInferenceContainerFields(workload *dinfwv1.DistributedInferenceWorkload) field_groups.ContainerFields {
	// Start with leader container fields
	leaderContainerFields := workload.LeaderSpec.ContainerFields
	aggregatedContainerFields := *leaderContainerFields.DeepCopy()

	// Aggregate worker container fields if they exist
	if workload.WorkerSpec == nil {
		return aggregatedContainerFields
	}

	workerContainerFields := workload.WorkerSpec.ContainerFields

	// For container fields, we typically want to prefer leader over worker for most fields
	// but we can aggregate some fields like environment variables

	// Aggregate environment variables
	if workerContainerFields.Environment != nil && leaderContainerFields.Environment != nil {
		// Merge environment variables - worker values can override leader values
		leaderEnv := leaderContainerFields.Environment.Deref()
		workerEnv := workerContainerFields.Environment.Deref()

		mergedEnv := make(map[string]field_types.DictionaryItemAndRules)
		// Start with worker environment
		for k, v := range workerEnv {
			mergedEnv[k] = v
		}

		// Override/add with leader environment
		for k, v := range leaderEnv {
			mergedEnv[k] = v
		}

		aggregatedContainerFields.Environment = &field_types.DictionaryField{
			Value: &mergedEnv,
		}
	} else if workerContainerFields.Environment != nil {
		aggregatedContainerFields.Environment = workerContainerFields.Environment
	}

	// Aggregate labels
	if workerContainerFields.Labels != nil && leaderContainerFields.Labels != nil {
		leaderLabels := leaderContainerFields.Labels.Deref()
		workerLabels := workerContainerFields.Labels.Deref()

		mergedLabels := make(map[string]field_types.DictionaryItemAndRules)
		for k, v := range leaderLabels {
			mergedLabels[k] = v
		}
		for k, v := range workerLabels {
			mergedLabels[k] = v
		}

		aggregatedContainerFields.Labels = &field_types.DictionaryField{
			Value: &mergedLabels,
		}
	} else if workerContainerFields.Labels != nil {
		aggregatedContainerFields.Labels = workerContainerFields.Labels
	}

	// Aggregate annotations
	if workerContainerFields.Annotations != nil && leaderContainerFields.Annotations != nil {
		leaderAnnotations := leaderContainerFields.Annotations.Deref()
		workerAnnotations := workerContainerFields.Annotations.Deref()

		mergedAnnotations := make(map[string]field_types.DictionaryItemAndRules)
		for k, v := range leaderAnnotations {
			mergedAnnotations[k] = v
		}
		for k, v := range workerAnnotations {
			mergedAnnotations[k] = v
		}

		aggregatedContainerFields.Annotations = &field_types.DictionaryField{
			Value: &mergedAnnotations,
		}
	} else if workerContainerFields.Annotations != nil {
		aggregatedContainerFields.Annotations = workerContainerFields.Annotations
	}

	return aggregatedContainerFields
}

// aggregateDistributedInferenceImages aggregates images from both leader and worker specs
func (p *WorkloadParser) aggregateDistributedInferenceImages(workload *dinfwv1.DistributedInferenceWorkload) []string {
	images := []string{}

	// Add leader image if it exists
	if workload.LeaderSpec.Image != nil && workload.LeaderSpec.Image.Value != nil {
		leaderImage := *workload.LeaderSpec.Image.Value
		if leaderImage != "" {
			images = append(images, leaderImage)
		}
	}

	// Add worker image if it exists and is different from leader
	if workload.WorkerSpec != nil && workload.WorkerSpec.Image != nil && workload.WorkerSpec.Image.Value != nil {
		workerImage := *workload.WorkerSpec.Image.Value
		if workerImage != "" {
			// Only add if it's different from leader image
			if len(images) == 0 || images[0] != workerImage {
				images = append(images, workerImage)
			}
		}
	}

	return images
}

func (p *WorkloadParser) aggregateDistributedInferenceResourceAllocationFields(workload *dinfwv1.DistributedInferenceWorkload) field_groups.ResourceAllocationFields {
	numReplicas := getNumReplicas(workload)
	if numReplicas == 0 {
		return field_groups.ResourceAllocationFields{}
	}

	// Start with leader resource allocation fields
	leaderResourceAllocationFields := workload.LeaderSpec.ResourceAllocationFields.DeepCopy()

	// Calculate resources for one replica (leader + all workers)
	oneReplicaResources := *leaderResourceAllocationFields

	if workload.WorkerSpec != nil {
		workerResourceAllocationFields := &workload.WorkerSpec.ResourceAllocationFields

		// Get number of workers per replica
		numWorkers := getNumWorkers(workload)

		// Multiply worker resources by number of workers
		workerResourcesForOneReplica := MultiplyResourceAllocationFields(workerResourceAllocationFields, int(numWorkers))

		// Add leader + all workers for one replica
		oneReplicaResources = *AddResourceAllocationFields(leaderResourceAllocationFields, workerResourcesForOneReplica)
	}

	// Multiply by number of replicas to get total resources
	if numReplicas > 1 {
		return *MultiplyResourceAllocationFields(&oneReplicaResources, int(numReplicas))
	}

	return oneReplicaResources
}

func (p *WorkloadParser) aggregateDistributedInferenceStorageFields(workload *dinfwv1.DistributedInferenceWorkload) field_groups.StorageFields {
	// Start with leader storage fields
	leaderStorageFields := workload.LeaderSpec.StorageFields.DeepCopy()
	aggregatedStorageFields := *leaderStorageFields

	// Aggregate worker storage fields if they exist
	if workload.WorkerSpec != nil {
		workerStorageFields := workload.WorkerSpec.StorageFields

		// For storage fields, we typically want to combine volumes from both leader and worker

		// Aggregate volumes
		if workerStorageFields.Volumes != nil && leaderStorageFields.Volumes != nil {
			leaderVolumes := leaderStorageFields.Volumes.Deref()
			workerVolumes := workerStorageFields.Volumes.Deref()

			mergedVolumes := make(map[string]field_types.VolumeItemAndRules)
			// Start with leader volumes
			for k, v := range leaderVolumes {
				mergedVolumes[k] = v
			}
			// Add worker volumes (worker volumes can have different names)
			for k, v := range workerVolumes {
				mergedVolumes[k] = v
			}

			aggregatedStorageFields.Volumes = &field_types.VolumesField{
				Value: &mergedVolumes,
			}
		} else if workerStorageFields.Volumes != nil {
			aggregatedStorageFields.Volumes = workerStorageFields.Volumes
		}
	}

	return aggregatedStorageFields
}

func fillNodeAffinities(nodeAffinityFields *field_types.NodeAffinityFields, result *openapi.WorkloadSyncFields) {
	affinity := nodeAffinityFieldsToApi(nodeAffinityFields)
	result.NodeAffinities = []openapi.WorkloadNodeAffinity{
		{SpecName: "worker", NodeAffinity: affinity},
	}
}

func setDistributedNodeAffinities(result *openapi.WorkloadSyncFields, workload *diwv2.DistributedWorkload) {
	var affinities []openapi.WorkloadNodeAffinity

	workerAffinity := nodeAffinityFieldsToApi(&workload.Spec.NodeAffinityFields)
	affinities = append(affinities, openapi.WorkloadNodeAffinity{
		SpecName: "worker", NodeAffinity: workerAffinity,
	})

	var masterAffinity *openapi.PodNodeAffinity = nil
	if workload.MasterSpec != nil {
		masterAffinity = nodeAffinityFieldsToApi(&workload.MasterSpec.NodeAffinityFields)
	}
	affinities = append(affinities, openapi.WorkloadNodeAffinity{
		SpecName: "master", NodeAffinity: masterAffinity,
	})

	result.NodeAffinities = affinities
}

func setDistributedInferenceNodeAffinities(workload *dinfwv1.DistributedInferenceWorkload, result *openapi.WorkloadSyncFields) {
	var affinities []openapi.WorkloadNodeAffinity

	leaderAffinity := nodeAffinityFieldsToApi(&workload.LeaderSpec.NodeAffinityFields)
	affinities = append(affinities, openapi.WorkloadNodeAffinity{
		SpecName: "leader", NodeAffinity: leaderAffinity,
	})

	var workerAffinity *openapi.PodNodeAffinity = nil
	if workload.WorkerSpec != nil {
		workerAffinity = nodeAffinityFieldsToApi(&workload.WorkerSpec.NodeAffinityFields)
	}
	affinities = append(affinities, openapi.WorkloadNodeAffinity{
		SpecName: "worker", NodeAffinity: workerAffinity,
	})

	result.NodeAffinities = affinities
}

// nodeAffinityFieldsToApi converts CRD NodeAffinityFields to the API PodNodeAffinity model.
// Reuses NodeSelectorTermModelToApi for the actual term conversion.
func nodeAffinityFieldsToApi(fields *field_types.NodeAffinityFields) *openapi.PodNodeAffinity {
	if fields == nil || fields.NodeAffinityRequired == nil {
		return nil
	}
	required := &openapi.PodNodeAffinityRequired{}
	for _, term := range fields.NodeAffinityRequired.NodeSelectorTerms {
		apiTerm := NodeSelectorTermModelToApi(corev1.NodeSelectorTerm{
			MatchExpressions: term.MatchExpressions,
		})
		required.NodeSelectorTerms = append(required.NodeSelectorTerms, apiTerm)
	}
	result := &openapi.PodNodeAffinity{}
	result.Required = *openapi.NewNullablePodNodeAffinityRequired(required)
	return result
}
