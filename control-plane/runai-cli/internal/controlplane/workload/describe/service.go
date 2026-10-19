/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package describe

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"

	workloadmanager "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client/generated"
	"github.com/samber/lo"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api/filter"

	internalApi "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/describe/displayer"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/describe/distributedinference"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/describe/inference"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/describe/nim"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
)

const (
	ListLimit                                    = 10
	OffsetDefault                                = 0
	AdditionalFieldsIssues                       = "issues"
	AdditionalFieldsGPUNetworkAccelerationStatus = "GPUNetworkAccelerationStatus"
	LatestCreatedRevisionName                    = "latestCreatedRevisionName"
	LatestReadyRevisionName                      = "latestReadyRevisionName"
	RoleLeader                                   = "Leader"
	RoleWorker                                   = "Worker"
	containerNameLeader                          = "leader"
	containerNameWorker                          = "worker"
)

type IssueGroup string

const (
	ComputeGroup  IssueGroup = "Compute"
	NetworkGroup  IssueGroup = "Network"
	InternalGroup IssueGroup = "Internal"
	StorageGroup  IssueGroup = "Storage"
)

func (ig IssueGroup) String() string {
	return string(ig)
}

type WorkloadClient interface {
	ListSlim(ctx context.Context, limit, offset int32, filters []string) (*openapi.GetWorkloadsSlimResponse, error)
	GetWorkload(ctx context.Context, id string) (*openapi.WorkloadDetailed, error)
	GetTraining(ctx context.Context, id string) (*openapi.Training, error)
	GetWorkspace(ctx context.Context, id string) (*openapi.Workspace, error)
	GetInference(ctx context.Context, id string) (*openapi.Inference, error)
	GetDistributed(ctx context.Context, id string) (*openapi.DistributedWorkload, error)
	GetDistributedInference(ctx context.Context, id string) (*openapi.DistributedInference, error)
	GetNimService(ctx context.Context, id string) (*workloadmanager.NimServiceResponse, error)
}

type WorkloadV2Getter interface {
	Get(ctx context.Context, id string) (*workloadmanager.WorkloadV2GetResponse, error)
}

type PodsClient interface {
	ListByLimit(ctx context.Context, workloadID string, limit int32) ([]openapi.Pod, error)
}

type EventClient interface {
	ListByLimit(ctx context.Context, workloadID string, limit int32) ([]openapi.Event, error)
}

type GPUNetworkAcceleration struct {
	Used bool `json:"used"`
}

type Service struct {
	workloadClient   WorkloadClient
	workloadV2Client WorkloadV2Getter
	podsClient       PodsClient
	eventClient      EventClient
}

func NewService(workloadClient WorkloadClient, workloadV2client WorkloadV2Getter, podsClient PodsClient, eventClient EventClient) *Service {
	return &Service{
		workloadClient:   workloadClient,
		workloadV2Client: workloadV2client,
		podsClient:       podsClient,
		eventClient:      eventClient,
	}
}

func (s *Service) GetWorkload(ctx context.Context, reqWorkload *api.Workload, podLimit int32, eventLimit int32) (*Workload, error) {
	var workload *openapi.WorkloadDetailed
	var err error
	var workloadId string

	if reqWorkload.ID != "" {
		workloadId = reqWorkload.ID
	} else {
		var workloadFromList *openapi.WorkloadSlim
		workloadFromList, err = getWorkloadByFilter(ctx, s.workloadClient, reqWorkload)
		if err != nil {
			return nil, err
		}

		workloadId = workloadFromList.Id
	}

	workload, err = s.workloadClient.GetWorkload(ctx, workloadId)
	if err != nil {
		return nil, err
	}

	var workloadPods []openapi.Pod
	var workloadEvents []openapi.Event
	if s.podsClient != nil {
		workloadPods, err = s.podsClient.ListByLimit(ctx, workload.Id, podLimit)
		if err != nil {
			return nil, fmt.Errorf("failed to get pods for workload %s: %w", workload.Id, err)
		}
	}
	if s.eventClient != nil {
		workloadEvents, err = s.eventClient.ListByLimit(ctx, workload.Id, eventLimit)
		if err != nil {
			return nil, fmt.Errorf("failed to get events for workload %s: %w", workload.Id, err)
		}
	}

	specData, err := s.generateSpec(ctx, workload)
	if err != nil {
		return nil, fmt.Errorf("failed to generate spec for workload %s: %w", workload.Id, err)
	}

	workloadToDisplay, err := s.buildWorkloadFromApiResponses(workload, workloadPods, workloadEvents, specData)

	if err != nil {
		return nil, fmt.Errorf("failed to generate workload %s: %w", workload.Id, err)
	}

	return workloadToDisplay, nil
}

// specWithRuntime holds both the displayer.Spec and extracted Runtime fields
type specWithRuntime struct {
	spec                 *displayer.Spec
	runtimeFields        *displayer.Runtime               // Extracted runtime fields from spec (ImagePullPolicy, Security, etc.)
	storage              *displayer.Storage               // Extracted storage from spec (PVCs, ConfigMaps, Secrets, EmptyDirs)
	environmentVariables *map[string]string               // Effective environment from the typed response
	compute              *displayer.Compute               // Effective requested resources from the typed response
	command              *string                          // Effective command from the typed response
	arguments            *string                          // Effective arguments from the typed response
	useTypedSummary      bool                             // Prefer typed desired state over workload-level projections
	masterConfig         *displayer.WorkloadRuntimeConfig // Full master configuration for distributed workloads
	nimFields            *nim.Fields
}

func (s *Service) generateSpec(ctx context.Context, workload *openapi.WorkloadDetailed) (*specWithRuntime, error) {
	if workload == nil {
		return nil, errors.New("workload is nil")
	}

	workloadTypeLower := strings.ToLower(workload.Type)

	if workloadTypeLower == internalApi.NimService.String() {
		nimService, err := s.workloadClient.GetNimService(ctx, workload.Id)
		if err != nil {
			return nil, err
		}
		if nimService == nil {
			return nil, fmt.Errorf("NIM service response is nil")
		}
		var annotations []openapi.Annotation
		if len(nimService.Spec.Annotations) > 0 {
			annotations = make([]openapi.Annotation, len(nimService.Spec.Annotations))
			for i, ann := range nimService.Spec.Annotations {
				name := ann.GetName()
				value := ""
				if ann.Value.IsSet() {
					value = ann.GetValue()
				}
				annotations[i] = openapi.Annotation{
					Name:  *openapi.NewNullableString(&name),
					Value: *openapi.NewNullableString(&value),
				}
			}
		}
		var labels []openapi.Label
		if len(nimService.Spec.Labels) > 0 {
			labels = make([]openapi.Label, len(nimService.Spec.Labels))
			for i, lbl := range nimService.Spec.Labels {
				name := lbl.GetName()
				value := ""
				if lbl.Value.IsSet() {
					value = lbl.GetValue()
				}
				labels[i] = openapi.Label{
					Name:  *openapi.NewNullableString(&name),
					Value: *openapi.NewNullableString(&value),
				}
			}
		}
		spec := displayer.NewSpecBuilder().
			Labels(lo.ToSlicePtr(labels)).
			Annotations(lo.ToSlicePtr(annotations)).
			SetRunaiCliCommandFromAnnotations(&annotations).
			Build()
		runtimeFields := s.extractRuntimeFromNimSpec(displayer.NewRuntimeBuilder(), &nimService.Spec).Build()
		nimFields := nim.BuildFromSpec(&nimService.Spec)
		if nimFields == nil {
			nimFields = &nim.Fields{}
		}
		metadata := nimService.Metadata
		if metadata.Category.IsSet() {
			category := metadata.GetCategory()
			nimFields.Category = &category
		}
		if metadata.Priority.IsSet() {
			priority := metadata.GetPriority()
			nimFields.PriorityClass = &priority
		}
		if metadata.Preemptibility.IsSet() {
			preemptibility := string(metadata.GetPreemptibility())
			nimFields.Preemptibility = &preemptibility
		}
		return &specWithRuntime{spec: &spec, runtimeFields: &runtimeFields, nimFields: nimFields}, nil
	}

	switch workloadTypeLower {
	case strings.ToLower(internalApi.Training.String()):
		training, err := s.workloadClient.GetTraining(ctx, workload.Id)
		if err != nil {
			return nil, err
		}
		spec := displayer.NewSpecBuilder().
			Labels(lo.ToSlicePtr(training.Spec.Labels)).Annotations(lo.ToSlicePtr(training.Spec.Annotations)).
			SetRunaiCliCommandFromAnnotations(&training.Spec.Annotations).
			ExposedUrls(training.Spec.ExposedUrls).
			Build()
		runtimeFields := s.extractRuntimeFromTrainingSpec(displayer.NewRuntimeBuilder(), training.Spec).Build()
		storage := extractStorageFromSpec(training.Spec.Storage.IsSet(), training.Spec.Storage.Get())
		return buildTypedSpecData(workload, &spec, &runtimeFields, storage, training.Spec), nil
	case strings.ToLower(internalApi.Workspace.String()):
		workspace, err := s.workloadClient.GetWorkspace(ctx, workload.Id)
		if err != nil {
			return nil, err
		}
		spec := displayer.NewSpecBuilder().
			Labels(lo.ToSlicePtr(workspace.Spec.Labels)).Annotations(lo.ToSlicePtr(workspace.Spec.Annotations)).
			SetRunaiCliCommandFromAnnotations(&workspace.Spec.Annotations).
			ExposedUrls(workspace.Spec.ExposedUrls).
			Build()
		runtimeFields := s.extractRuntimeFromWorkspaceSpec(displayer.NewRuntimeBuilder(), workspace.Spec).Build()
		storage := extractStorageFromSpec(workspace.Spec.Storage.IsSet(), workspace.Spec.Storage.Get())
		return buildTypedSpecData(workload, &spec, &runtimeFields, storage, workspace.Spec), nil
	case strings.ToLower(internalApi.Distributed.String()):
		dist, err := s.workloadClient.GetDistributed(ctx, workload.Id)
		if err != nil {
			return nil, err
		}

		specBuilder := displayer.NewSpecBuilder().
			Labels(lo.ToSlicePtr(dist.Spec.Labels)).Annotations(lo.ToSlicePtr(dist.Spec.Annotations)).
			SetRunaiCliCommandFromAnnotations(&dist.Spec.Annotations).
			ExposedUrls(dist.Spec.ExposedUrls)

		var masterConfig *displayer.WorkloadRuntimeConfig
		if dist.HasMasterSpec() {
			masterSpec := dist.MasterSpec.Get()
			specBuilder.MasterLabels(lo.ToSlicePtr(masterSpec.Labels))
			masterConfig = buildMasterConfigFromSpec(masterSpec)
		}

		spec := specBuilder.Build()
		runtimeFields := s.extractRuntimeFromDistributedSpec(displayer.NewRuntimeBuilder(), dist.Spec).Build()
		storage := extractStorageFromSpec(dist.Spec.Storage.IsSet(), dist.Spec.Storage.Get())
		specData := buildTypedSpecData(workload, &spec, &runtimeFields, storage, dist.Spec)
		specData.masterConfig = masterConfig
		return specData, nil
	case strings.ToLower(internalApi.Inference.String()):
		inference, err := s.workloadClient.GetInference(ctx, workload.Id)
		if err != nil {
			return nil, err
		}
		spec := displayer.NewSpecBuilder().
			Labels(lo.ToSlicePtr(inference.Spec.Labels)).Annotations(lo.ToSlicePtr(inference.Spec.Annotations)).
			AutoScalingPtr(inference.Spec.Autoscaling.Get()).
			ServingConfigurationPtr(inference.Spec.ServingConfiguration.Get()).
			SetRunaiCliCommandFromAnnotations(&inference.Spec.Annotations).
			ExposedUrls(inference.Spec.ExposedUrls).
			ServingPort(inference.Spec.ServingPort.Get()).
			Build()
		runtimeFields := s.extractRuntimeFromInferenceSpec(displayer.NewRuntimeBuilder(), inference.Spec).Build()
		storage := extractStorageFromSpec(inference.Spec.Storage.IsSet(), &inferenceStorageAdapter{inference.Spec.Storage.Get()})
		return buildTypedSpecData(workload, &spec, &runtimeFields, storage, inference.Spec), nil
	case strings.ToLower(internalApi.DistributedInference.StringCapitalize()):
		distInference, err := s.workloadClient.GetDistributedInference(ctx, workload.Id)
		if err != nil {
			return nil, err
		}
		var annotations []openapi.Annotation
		if distInference.Spec != nil && distInference.Spec.Leader.IsSet() {
			leader := distInference.Spec.Leader.Get()
			if leader != nil {
				annotations = leader.Annotations
			}
		}
		spec := displayer.NewSpecBuilder().
			SetRunaiCliCommandFromAnnotations(&annotations).
			DistributedInferenceSpec(distInference.Spec).
			Build()
		// DistributedInference doesn't populate runtime fields here (they're in Leader/Worker sections)
		return &specWithRuntime{spec: &spec, runtimeFields: nil}, nil
	default:
		if isWorkloadV2(workload) {
			return s.loadAndParseWorkloadV2ManifestSpec(ctx, workload.Id)
		}
		return nil, fmt.Errorf("unknown workload type %s", workload.Type)
	}
}

func isWorkloadV2(workload *openapi.WorkloadDetailed) bool {
	return workload.SourceApi.IsSet() && *workload.SourceApi.Get() == openapi.SOURCEAPI_WORKLOADS_V2
}

func (s *Service) loadAndParseWorkloadV2ManifestSpec(ctx context.Context, workloadID string) (*specWithRuntime, error) {
	workloadManifest, err := s.getWorkloadV2Manifest(ctx, workloadID)
	if err != nil {
		return nil, err
	}
	return buildWorkloadV2ManifestSpec(workloadManifest), nil
}

func (s *Service) getWorkloadV2Manifest(ctx context.Context, workloadID string) (map[string]interface{}, error) {
	workloadV2, err := s.workloadV2Client.Get(ctx, workloadID)
	if err != nil {
		return nil, err
	}
	workloadDetails := workloadV2.GetWorkload()
	return workloadDetails.GetWorkloadManifest(), nil
}

func buildWorkloadV2ManifestSpec(workloadManifest map[string]interface{}) *specWithRuntime {
	workloadAnnotations := GetAnnotationsAsStructs(workloadManifest)
	spec := displayer.NewSpecBuilder().Labels(GetLabelsAsStructs(workloadManifest)).Annotations(workloadAnnotations).Build()
	return &specWithRuntime{spec: &spec, runtimeFields: nil, storage: nil}
}

func buildTypedSpecData[T interface {
	GetArgsOk() (*string, bool)
	GetCommandOk() (*string, bool)
	GetComputeOk() (*openapi.WorkspaceSpecSpecAllOfCompute, bool)
	GetEnvironmentVariables() []openapi.EnvironmentVariable
	GetNodePools() []string
}](
	workload *openapi.WorkloadDetailed,
	spec *displayer.Spec,
	runtimeFields *displayer.Runtime,
	storage *displayer.Storage,
	typedSpec T,
) *specWithRuntime {
	specData := &specWithRuntime{spec: spec, runtimeFields: runtimeFields, storage: storage}
	if !isWorkloadV2(workload) {
		return specData
	}
	specData.useTypedSummary = true

	environmentVariables := buildEnvironmentVariablesForDisplay(typedSpec.GetEnvironmentVariables())
	specData.environmentVariables = &environmentVariables

	if command, found := typedSpec.GetCommandOk(); found {
		specData.command = command
	}
	if arguments, found := typedSpec.GetArgsOk(); found {
		specData.arguments = arguments
	}

	typedCompute := displayer.NewComputeBuilder().RequestedNodePools(typedSpec.GetNodePools())
	if computeSpec, found := typedSpec.GetComputeOk(); found {
		if computeFromSpec := displayer.BuildComputeFromSpec(computeSpec); computeFromSpec != nil {
			typedCompute.RequestedResources(computeFromSpec.RequestedResources)
		}
	}
	compute := typedCompute.Build()
	specData.compute = &compute

	return specData
}

func buildEnvironmentVariablesForDisplay(environmentVariables []openapi.EnvironmentVariable) map[string]string {
	result := make(map[string]string, len(environmentVariables))
	for _, environmentVariable := range environmentVariables {
		name := environmentVariable.GetName()
		if name == "" || environmentVariable.GetExclude() {
			continue
		}
		result[name] = environmentVariableValueForDisplay(environmentVariable)
	}
	return result
}

func environmentVariableValueForDisplay(environmentVariable openapi.EnvironmentVariable) string {
	if value, found := environmentVariable.GetValueOk(); found && value != nil {
		return *value
	}
	if secret, found := environmentVariable.GetSecretOk(); found && secret != nil {
		return fmt.Sprintf("SECRET:%s,%s", secret.GetName(), secret.GetKey())
	}
	if configMap, found := environmentVariable.GetConfigMapOk(); found && configMap != nil {
		return fmt.Sprintf("CONFIGMAP:%s,%s", configMap.GetName(), configMap.GetKey())
	}
	if podFieldReference, found := environmentVariable.GetPodFieldRefOk(); found && podFieldReference != nil {
		return fmt.Sprintf("PODFIELDREF:%s", podFieldReference.GetPath())
	}
	if userCredential, found := environmentVariable.GetUserCredentialOk(); found && userCredential != nil {
		return fmt.Sprintf("USERCREDENTIAL:%s,%s", userCredential.GetName(), userCredential.GetKey())
	}
	return ""
}

// extractMetadata is a generic function handling the extraction and iteration logic
// T is the target struct (e.g., openapi.Label or openapi.Annotation)
func extractMetadata[T any](
	workloadManifest map[string]interface{},
	subField string,
	factory func(k, v string) T,
) []*T {
	// Get the map using the unstructured helper
	dataMap, found, err := unstructured.NestedStringMap(workloadManifest, "metadata", subField)
	if err != nil || !found {
		return []*T{}
	}
	output := make([]*T, 0, len(dataMap))

	for key, value := range dataMap {
		item := factory(key, value)
		output = append(output, &item)
	}

	return output
}

func GetLabelsAsStructs(workloadManifest map[string]interface{}) []*openapi.Label {
	return extractMetadata(workloadManifest, "labels", func(k, v string) openapi.Label {
		return openapi.Label{
			Name:  *openapi.NewNullableString(&k),
			Value: *openapi.NewNullableString(&v),
		}
	})
}

func GetAnnotationsAsStructs(workloadManifest map[string]interface{}) []*openapi.Annotation {
	return extractMetadata(workloadManifest, "annotations", func(k, v string) openapi.Annotation {
		return openapi.Annotation{
			Name:  *openapi.NewNullableString(&k),
			Value: *openapi.NewNullableString(&v),
		}
	})
}

func (s *Service) buildWorkloadFromApiResponses(workload *openapi.WorkloadDetailed, pods []openapi.Pod, events []openapi.Event, specData *specWithRuntime) (*Workload, error) {
	workloadBuilder := NewWorkloadBuilder()

	spec := specData.spec
	setWorkloadTypeSpecificFields(workloadBuilder, workload, spec, specData)

	setGeneralInformation(workloadBuilder, workload, specData)
	s.setRuntimeInformation(workloadBuilder, workload, specData.runtimeFields)
	setComputeInformation(workloadBuilder, workload, specData.compute)
	setStorage(workloadBuilder, specData.storage)
	workloadBuilder.SetMasterConfig(specData.masterConfig)
	setNetworks(workloadBuilder, workload, spec)
	setNetworkTopology(workloadBuilder, workload)

	err := setAdditionalFields(workloadBuilder, workload.AdditionalFields)
	if err != nil {
		return nil, err
	}

	if pods != nil {
		setPods(workloadBuilder, pods, workload.RunningPods)
	}
	if events != nil {
		setEvents(workloadBuilder, events)
	}

	if spec != nil {
		workloadBuilder.SetSpec(*spec)
	}

	describeWorkload := workloadBuilder.Build()
	return describeWorkload, nil
}

func setIssues(builder *WorkloadBuilder, fields map[string]interface{}) error {
	issues, err := getIssuesFromAdditionalFields(fields)
	if err != nil {
		return err
	}

	flatIssues := flattenIssues(issues)

	slices.SortFunc(flatIssues, func(a, b displayer.Issue) int {
		return cmp.Or(
			cmp.Compare(a.Type, b.Type),
			cmp.Compare(a.Kind, b.Kind),
			cmp.Compare(a.Name, b.Name),
		)
	})

	builder.SetIssues(flatIssues)

	return nil
}

func setGPUNetworkAccelerationStatus(builder *WorkloadBuilder, fields map[string]interface{}) error {
	field, ok := fields[AdditionalFieldsGPUNetworkAccelerationStatus]
	if !ok {
		return nil
	}

	data, ok := field.(string)
	if ok {
		var gpuNetworkAcceleration GPUNetworkAcceleration
		err := json.Unmarshal([]byte(data), &gpuNetworkAcceleration)
		if err != nil {
			return err
		}
		// GPU network acceleration (MNNVL) used
		builder.GPUNetworkAcceleration(&gpuNetworkAcceleration.Used)
	}
	return nil
}

// flattenIssues returns a flat list of issues from the original Issues struct and enriches them with their group/type
func flattenIssues(issues Issues) []displayer.Issue {
	issueGroups := map[IssueGroup][]displayer.Issue{
		ComputeGroup:  issues.Compute,
		NetworkGroup:  issues.Network,
		StorageGroup:  issues.Storage,
		InternalGroup: issues.Internal,
	}

	var flatIssues []displayer.Issue
	for group, issueList := range issueGroups {
		for _, issue := range issueList {
			issue.Type = group.String()
			flatIssues = append(flatIssues, issue)
		}
	}

	return flatIssues
}

func getIssuesFromAdditionalFields(additionalFieldsMap map[string]interface{}) (Issues, error) {
	var issues Issues

	issuesField, ok := additionalFieldsMap[AdditionalFieldsIssues]
	if !ok {
		return issues, nil
	}

	issuesStr := []byte(issuesField.(string))
	err := json.Unmarshal(issuesStr, &issues)
	if err != nil {
		return issues, err
	}

	return issues, nil
}

func setRequestAndLimit(cores openapi.NullableRequestResourceCores, request, limit func(limit float64) *displayer.ComputeResourcesBuilder) {
	if cores.IsSet() {
		resource := cores.Get()
		if resource.Request.IsSet() {
			request(*resource.Request.Get())
		}
		if resource.Limit.IsSet() {
			limit(*resource.Limit.Get())
		}
	}
}

func setRequestAndLimitQuantity(memory openapi.NullableRequestResourceQuantity, request, limit func(requested string) *displayer.ComputeResourcesBuilder) {
	if memory.IsSet() {
		resource := memory.Get()
		if resource.Request.IsSet() {
			request(*resource.Request.Get())
		}
		if resource.Limit.IsSet() {
			limit(*resource.Limit.Get())
		}
	}
}

func setPods(builder *WorkloadBuilder, pods []openapi.Pod, runningPods int32) {
	if pods == nil {
		return
	}
	podsBuilder := displayer.NewPodsBuilder()
	for _, apiPod := range pods {
		podBuilder := buildPod(apiPod)
		podsBuilder.Add(podBuilder.Build())
	}
	podsInfo := displayer.PodsInfo{
		Pods:        podsBuilder.Build(),
		RunningPods: runningPods,
	}
	builder.SetPodsInfo(podsInfo)
}

func buildPod(apiPod openapi.Pod) *displayer.PodBuilder {
	podBuilder := displayer.NewPodBuilder().
		Name(apiPod.Name).
		K8sPhase(apiPod.K8sPhase)

	if apiPod.NodeName.Get() != nil {
		podBuilder.NodeName(*apiPod.NodeName.Get())
	}
	if apiPod.CompletedAt.Get() != nil {
		podBuilder.CompletedAt(*apiPod.CompletedAt.Get())
	}
	if apiPod.CurrentNodePool.Get() != nil {
		podBuilder.CurrentNodePool(*apiPod.CurrentNodePool.Get())
	}
	if len(apiPod.Containers) >= 1 {
		// The first container is the user container, the rest are system containers
		userContainer := buildContainer(apiPod.Containers[0])
		podBuilder.AddContainer(userContainer)
	}

	// Set role based on any container name (leader/worker)
	for _, container := range apiPod.Containers {
		if role := getPodRoleFromContainerName(container.Name); role != "" {
			podBuilder.Role(role)
			break
		}
	}

	return podBuilder
}

func getPodRoleFromContainerName(containerName string) string {
	switch strings.ToLower(containerName) {
	case containerNameLeader:
		return RoleLeader
	case containerNameWorker:
		return RoleWorker
	default:
		return ""
	}
}

func buildContainer(apiContainer openapi.Container) displayer.Container {
	container := displayer.Container{
		Name:  apiContainer.Name,
		Image: apiContainer.Image,
	}
	if apiContainer.StartedAt.Get() != nil {
		container.StartedAt = apiContainer.StartedAt.Get()
	}
	return container
}

func setEvents(builder *WorkloadBuilder, events []openapi.Event) {
	if events == nil {
		return
	}
	eventsBuilder := displayer.NewEventsBuilder()
	for _, apiEvent := range events {
		eventBuilder := buildEvent(apiEvent)
		eventsBuilder.Add(eventBuilder.Build())
	}
	builder.SetEvents(eventsBuilder.Build())
}

func buildEvent(apiEvent openapi.Event) *displayer.EventBuilder {
	eventBuilder := displayer.NewEventBuilder().
		CreatedAt(apiEvent.CreatedAt).
		Message(apiEvent.Message)

	if apiEvent.Id != nil {
		eventBuilder.Id(*apiEvent.Id)
	}
	if apiEvent.Reason != nil {
		eventBuilder.Reason(*apiEvent.Reason)
	}
	if apiEvent.Source != nil {
		eventBuilder.Source(*apiEvent.Source)
	}
	if apiEvent.InvolvedObject != nil {
		involvedObject := buildInvolvedObject(apiEvent.InvolvedObject)
		eventBuilder.InvolvedObject(involvedObject)
	}

	return eventBuilder
}

func buildInvolvedObject(apiObject *openapi.InvolvedObject) displayer.InvolvedObject {
	return displayer.NewInvolvedObjectBuilder().
		Uid(apiObject.Uid).
		Kind(apiObject.Kind).
		Name(apiObject.Name).
		Namespace(apiObject.Namespace).
		Build()
}

func getWorkloadByFilter(ctx context.Context, workloadClient WorkloadClient, reqWorkload *api.Workload) (*openapi.WorkloadSlim, error) {
	fb := filter.NewBuilder().
		WithClusterID(reqWorkload.ClusterID).
		WithProject(reqWorkload.Project)

	if reqWorkload.Name != "" {
		fb.AppendName(reqWorkload.Name)
	}
	if reqWorkload.Type != internalApi.Undefined {
		workloadType := reqWorkload.Type.QueryParam()
		if reqWorkload.Framework == internalApi.NimService {
			workloadType = internalApi.BackendNimServiceWorkloadType
		}
		fb.WithWorkloadType(workloadType)
	}
	if reqWorkload.Framework != internalApi.UndefinedFramework {
		fb.WithWorkloadFramework(reqWorkload.Framework)
	}
	workloadList, err := workloadClient.ListSlim(ctx, ListLimit, OffsetDefault, fb.StringList())
	if err != nil {
		return nil, err
	}
	if len(workloadList.Workloads) == 0 {
		return nil, errors.New("workload not found")
	}
	if len(workloadList.Workloads) > 1 {
		var workloadsFound []api.MinimalWorkloadInfo
		for _, workload := range workloadList.Workloads {
			workloadsFound = append(workloadsFound, api.MinimalWorkloadInfo{
				Name:    workload.Name,
				ID:      workload.Id,
				Project: workload.ProjectName,
				Kind:    workload.GetType(),
			})
		}
		return nil, api.NewMultiWorkloadsError(workloadsFound)
	}
	return &workloadList.Workloads[0], nil
}

func setWorkloadTypeSpecificFields(builder *WorkloadBuilder, workload *openapi.WorkloadDetailed, spec *displayer.Spec, specData *specWithRuntime) {
	workloadType := strings.ToLower(workload.Type)

	switch workloadType {
	case strings.ToLower(internalApi.Inference.String()):
		var autoScaling *openapi.AutoScaling
		var servingConfig *openapi.ServingConfiguration
		if spec != nil {
			autoScaling = &spec.AutoScaling
			servingConfig = &spec.ServingConfiguration
		}
		inference.BuildAndSetFields(&builder.wd.Inference, workload.AdditionalFields, workload.RequestedPods, autoScaling, servingConfig)

	case strings.ToLower(internalApi.DistributedInference.StringCapitalize()):
		if spec != nil {
			distributedinference.BuildAndSetFields(&builder.wd.DistributedInference, spec.DistributedInferenceSpec, workload.AdditionalFields, workload.Urls)
		}
	case internalApi.NimService.String():
		nimFields := &nim.Fields{}
		if specData != nil && specData.nimFields != nil {
			nimFields = specData.nimFields
		}
		if workload.RequestedPods != nil && workload.RequestedPods.Number.IsSet() {
			nimFields.ReplicasRequested = workload.RequestedPods.Number.Get()
		}
		nimFields.ReplicasReady = &workload.RunningPods
		if len(workload.Urls) > 0 {
			nimFields.ServingURL = &workload.Urls[0]
		}
		nim.SetFields(&builder.wd.Nim, nimFields)
	}
}

func setGeneralInformation(
	builder *WorkloadBuilder,
	workload *openapi.WorkloadDetailed,
	specData *specWithRuntime,
) {
	spec := specData.spec
	generalBuilder := displayer.NewGeneralBuilder().
		Name(workload.Name).
		ID(workload.Id).
		Type(workload.Type).
		Phase(PhaseFromApi(workload.Phase)).
		Project(workload.ProjectName).
		Department(workload.DepartmentName).
		CreatedAt(workload.CreatedAt).
		UpdatedAt(workload.UpdatedAt).
		Category(workload.Category)

	if workload.PriorityClassName != "" {
		priority := workload.PriorityClassName
		if workload.Priority.IsSet() && workload.Priority.Get() != nil {
			priority = fmt.Sprintf("%s (%d)", workload.PriorityClassName, *workload.Priority.Get())
		}
		generalBuilder.Priority(priority)
	}

	if workload.Preemptibility.IsSet() && workload.Preemptibility.Get() != nil {
		generalBuilder.Preemptibility(*workload.Preemptibility.Get())
	} else if workload.Preemptible.IsSet() && workload.Preemptible.Get() != nil {
		if *workload.Preemptible.Get() {
			generalBuilder.Preemptibility(string(openapi.PREEMPTIBILITY_PREEMPTIBLE))
		} else {
			generalBuilder.Preemptibility(string(openapi.PREEMPTIBILITY_NON_PREEMPTIBLE))
		}
	}

	if workload.Topology.IsSet() {
		topology := workload.GetTopology()
		if topology.Name != nil && *topology.Name != "" {
			generalBuilder.Topology(*topology.Name)
		}
	}

	if spec != nil && spec.RunaiCliCommand != "" {
		generalBuilder.RunaiCliCommand(spec.RunaiCliCommand)
	}

	if specData.useTypedSummary {
		if specData.command != nil {
			generalBuilder.Command(*specData.command)
		}
		if specData.arguments != nil {
			generalBuilder.Arguments(*specData.arguments)
		}
		generalBuilder.EnvironmentVariables(*specData.environmentVariables)
	} else {
		if workload.Command != nil {
			generalBuilder.Command(*workload.Command)
		}
		if workload.Arguments != nil {
			generalBuilder.Arguments(*workload.Arguments)
		}
		if workload.EnvironmentVariables != nil {
			generalBuilder.EnvironmentVariables(*workload.EnvironmentVariables)
		}
	}

	if workload.PhaseMessage != nil {
		generalBuilder.PhaseMessage(*workload.PhaseMessage)
	}
	if workload.SubmittedBy != nil {
		generalBuilder.SubmittedBy(*workload.SubmittedBy)
	}
	if workload.DistributedFramework != nil {
		generalBuilder.DistributedFramework(*workload.DistributedFramework)
	}

	builder.SetGeneral(generalBuilder.Build())
}

func (s *Service) setRuntimeInformation(builder *WorkloadBuilder, workload *openapi.WorkloadDetailed, runtimeFieldsFromSpec *displayer.Runtime) {
	// Start with fields from the spec (if available)
	runtimeBuilder := displayer.NewRuntimeBuilder()

	if runtimeFieldsFromSpec != nil {
		runtimeBuilder.Set(*runtimeFieldsFromSpec)
	}

	// Prefer fields extracted from the effective spec because workload-level fields can lag behind V2 updates.
	if runtimeFieldsFromSpec == nil || len(runtimeFieldsFromSpec.Images) == 0 {
		runtimeBuilder.Images(workload.Images)
	}

	if workload.Preemptible.Get() != nil {
		runtimeBuilder.Preemptible(*workload.Preemptible.Get())
	}

	builder.SetRuntime(runtimeBuilder.Build())
}

func (s *Service) extractRuntimeFromTrainingSpec(runtimeBuilder *displayer.RuntimeBuilder, spec *openapi.TrainingSpecSpec) *displayer.RuntimeBuilder {
	return extractRuntimeFromSpec(runtimeBuilder, spec, &trainingSpecAdapter{spec: spec})
}

func (s *Service) extractRuntimeFromWorkspaceSpec(runtimeBuilder *displayer.RuntimeBuilder, spec *openapi.WorkspaceSpecSpec) *displayer.RuntimeBuilder {
	return extractRuntimeFromSpec(runtimeBuilder, spec, &workspaceSpecAdapter{spec: spec})
}

func (s *Service) extractRuntimeFromDistributedSpec(runtimeBuilder *displayer.RuntimeBuilder, spec *openapi.DistributedSpecSpec) *displayer.RuntimeBuilder {
	return extractRuntimeFromSpec(runtimeBuilder, spec, &distributedSpecAdapter{spec: spec})
}

func (s *Service) extractRuntimeFromInferenceSpec(runtimeBuilder *displayer.RuntimeBuilder, spec *openapi.InferenceSpecSpec) *displayer.RuntimeBuilder {
	return extractRuntimeFromSpec(runtimeBuilder, spec, &inferenceSpecAdapter{spec: spec})
}

func (s *Service) extractRuntimeFromNimSpec(runtimeBuilder *displayer.RuntimeBuilder, spec *workloadmanager.NimServiceSpec) *displayer.RuntimeBuilder {
	return extractRuntimeFromSpec(runtimeBuilder, spec, &nimSpecAdapter{spec: spec})
}

func buildRequestedResources(workload *openapi.WorkloadDetailed) displayer.ComputeResources {
	computeResourcesBuilder := displayer.NewComputeResourcesBuilder()
	requestedResources := workload.WorkloadRequestedResources.Get()
	if requestedResources == nil {
		return computeResourcesBuilder.Build()
	}

	if requestedResources.HasGpu() {
		setRequestAndLimit(requestedResources.Gpu, computeResourcesBuilder.GPUComputeRequest, computeResourcesBuilder.GPUComputeLimit)
	}
	if requestedResources.HasCpu() {
		setRequestAndLimit(requestedResources.Cpu, computeResourcesBuilder.CPUCoresRequest, computeResourcesBuilder.CPUCoresLimit)
	}
	if requestedResources.HasGpuMemory() {
		setRequestAndLimitQuantity(requestedResources.GpuMemory, computeResourcesBuilder.GPUMemoryRequest, computeResourcesBuilder.GPUMemoryLimit)
	}
	if requestedResources.HasCpuMemory() {
		setRequestAndLimitQuantity(requestedResources.CpuMemory, computeResourcesBuilder.CPUMemoryRequest, computeResourcesBuilder.CPUMemoryLimit)
	}

	return computeResourcesBuilder.Build()
}

func buildAllocatedResources(workload *openapi.WorkloadDetailed) displayer.AllocatedResources {
	allocatedResourcesBuilder := displayer.NewAllocatedResourcesBuilder()
	allocatedResource := workload.AllocatedResources.Get()
	if allocatedResource != nil {
		setAllocatedResourceFloat(allocatedResource.Gpu, allocatedResourcesBuilder.GPUCompute)
		setAllocatedResourceFloat(allocatedResource.Cpu, allocatedResourcesBuilder.CPUCores)
		setAllocatedResourceString(allocatedResource.GpuMemory, allocatedResourcesBuilder.GPUMemory)
		setAllocatedResourceString(allocatedResource.CpuMemory, allocatedResourcesBuilder.CPUMemory)
	}
	return allocatedResourcesBuilder.Build()
}

func setAllocatedResourceFloat(resource openapi.NullableFloat64, setter func(float64) *displayer.AllocatedResourcesBuilder) {
	if resource.IsSet() {
		setter(*resource.Get())
	}
}

func setAllocatedResourceString(resource openapi.NullableString, setter func(string) *displayer.AllocatedResourcesBuilder) {
	if resource.IsSet() {
		setter(*resource.Get())
	}
}

func setComputeInformation(builder *WorkloadBuilder, workload *openapi.WorkloadDetailed, computeFromSpec *displayer.Compute) {
	computeBuilder := displayer.NewComputeBuilder()
	if computeFromSpec != nil {
		computeBuilder.Set(*computeFromSpec).CurrentNodePools(workload.CurrentNodePools)
	} else {
		computeBuilder.
			RequestedNodePools(workload.RequestedNodePools).
			CurrentNodePools(workload.CurrentNodePools)
	}

	if computeFromSpec == nil && workload.WorkloadRequestedResources.IsSet() {
		computeBuilder.RequestedResources(buildRequestedResources(workload))
	}

	if workload.AllocatedResources.IsSet() {
		computeBuilder.AllocatedResources(buildAllocatedResources(workload))
	}

	builder.SetCompute(computeBuilder.Build())
}

func setStorage(builder *WorkloadBuilder, storage *displayer.Storage) {
	builder.SetStorage(storage)
}

func buildMasterConfigFromSpec(masterSpec *openapi.MasterSpec) *displayer.WorkloadRuntimeConfig {
	if masterSpec == nil {
		return nil
	}

	builder := displayer.NewWorkloadRuntimeConfigBuilder()

	if masterSpec.Image.IsSet() {
		builder.Image(masterSpec.GetImage())
	}

	if masterSpec.ImagePullPolicy.IsSet() {
		builder.ImagePullPolicy(string(masterSpec.GetImagePullPolicy()))
	}

	if len(masterSpec.ImagePullSecrets) > 0 {
		builder.ImagePullSecrets(lo.Map(masterSpec.ImagePullSecrets, func(s openapi.ImagePullSecret, _ int) string {
			return s.GetName()
		}))
	}

	if masterSpec.Command.IsSet() {
		builder.Command(masterSpec.GetCommand())
	}

	if masterSpec.Args.IsSet() {
		builder.Arguments(masterSpec.GetArgs())
	}

	if masterSpec.WorkingDir.IsSet() {
		builder.WorkingDir(masterSpec.GetWorkingDir())
	}

	if len(masterSpec.EnvironmentVariables) > 0 {
		builder.EnvironmentVariables(buildEnvironmentVariablesForDisplay(masterSpec.EnvironmentVariables))
	}

	if len(masterSpec.Labels) > 0 {
		builder.Labels(lo.Associate(masterSpec.Labels, func(label openapi.Label) (string, string) {
			return label.GetName(), label.GetValue()
		}))
	}

	//nolint:staticcheck // NodeType is deprecated but still supported for backward compatibility
	if masterSpec.NodeType.IsSet() {
		builder.NodeType(masterSpec.GetNodeType())
	}

	if len(masterSpec.Tolerations) > 0 {
		builder.Tolerations(lo.FilterMap(masterSpec.Tolerations, func(t openapi.Toleration, _ int) (string, bool) {
			s := displayer.FormatToleration(&t)
			return s, s != ""
		}))
	}

	if masterSpec.Security.IsSet() {
		security := masterSpec.GetSecurity()

		if security.RunAsUid.IsSet() {
			builder.RunAsUser(security.GetRunAsUid())
		}

		if security.RunAsGid.IsSet() {
			builder.RunAsGroup(security.GetRunAsGid())
		}

		if len(security.Capabilities) > 0 {
			builder.Capabilities(lo.Map(security.Capabilities, func(cap openapi.Capability, _ int) string {
				return string(cap)
			}))
		}
	}

	if masterSpec.Probes.IsSet() && masterSpec.Probes.Get().Readiness.IsSet() {
		if probeStr := displayer.FormatProbe(masterSpec.Probes.Get().Readiness.Get()); probeStr != "" {
			builder.ReadinessProbe(probeStr)
		}
	}

	if masterSpec.Compute.IsSet() {
		builder.Compute(displayer.BuildComputeFromSpec(masterSpec.Compute.Get()))
	}

	if masterSpec.Storage.IsSet() {
		builder.Storage(extractStorageFromSpec(masterSpec.Storage.IsSet(), masterSpec.Storage.Get()))
	}

	return builder.BuildPtr()
}

// storageFields is an interface for extracting storage fields from different spec storage types
type storageFields interface {
	GetConfigMapVolume() []openapi.ConfigMapInstance
	GetSecretVolume() []openapi.SecretInstance
	GetPvc() []openapi.PvcInstance
	GetEmptyDirVolume() []openapi.EmptyDirInstance
	GetGit() []openapi.GitInstance
	GetHostPath() []openapi.HostPathInstance
	GetS3() []openapi.S3Instance
}

// inferenceStorageAdapter wraps InferenceSpecSpecAllOfStorage to add missing GetS3() method
type inferenceStorageAdapter struct {
	*openapi.InferenceSpecSpecAllOfStorage
}

// GetS3 returns nil since Inference workloads don't support S3 storage
func (a *inferenceStorageAdapter) GetS3() []openapi.S3Instance {
	return nil
}

// extractStorageFromSpec extracts storage from any storage spec that implements storageFields
func extractStorageFromSpec(isSet bool, storage storageFields) *displayer.Storage {
	if !isSet || storage == nil {
		return nil
	}
	return buildStorageFromFields(storage)
}

// buildStorageFromFields builds storage from any type that implements storageFields
func buildStorageFromFields(storage storageFields) *displayer.Storage {
	return displayer.BuildStorageFromWorkloadSpec(
		storage.GetConfigMapVolume(),
		storage.GetSecretVolume(),
		storage.GetPvc(),
		storage.GetEmptyDirVolume(),
		storage.GetGit(),
		storage.GetHostPath(),
		storage.GetS3(),
	)
}

func setNetworks(builder *WorkloadBuilder, workload *openapi.WorkloadDetailed, spec *displayer.Spec) {
	authMap := buildExposedUrlsAuthMap(spec)
	networksBuilder := displayer.NewNetworksBuilder()

	if workload.ExternalConnections != nil {
		for _, c := range workload.ExternalConnections {
			network := buildNetworkWithAuth(c, authMap, spec)
			networksBuilder.Add(network)
		}
	} else if workload.Environments != nil {
		for _, env := range workload.Environments {
			for _, c := range env.Connections {
				network := buildNetworkWithAuth(c, authMap, spec)
				networksBuilder.Add(network)
			}
		}
	} else if len(workload.Urls) != 0 {
		for _, url := range workload.Urls {
			network := displayer.NewNetworkBuilder().
				Url(url).
				Build()
			networksBuilder.Add(network)
		}
	}

	builder.SetNetworks(networksBuilder.Build())
}

func setNetworkTopology(builder *WorkloadBuilder, workload *openapi.WorkloadDetailed) {
	if !workload.Topology.IsSet() {
		return
	}
	topology := workload.GetTopology()
	if topology.Name == nil || *topology.Name == "" {
		return
	}

	nt := &displayer.NetworkTopology{
		Name:               topology.Name,
		RequiredPlacement:  topology.RequiredPlacement,
		PreferredPlacement: topology.PreferredPlacement,
	}

	if topology.ActualPlacement.IsSet() {
		actual := topology.ActualPlacement.Get()
		nt.ActualPlacement = displayer.FormatActualPlacement(actual.Level, actual.Value)
	}

	builder.SetNetworkTopology(nt)
}

func buildExposedUrlsAuthMap(spec *displayer.Spec) map[string]*openapi.ExposedUrl {
	authMap := make(map[string]*openapi.ExposedUrl)
	if spec != nil {
		for i := range spec.ExposedUrls {
			eu := &spec.ExposedUrls[i]
			if eu.ToolName.IsSet() && eu.ToolName.Get() != nil {
				authMap[*eu.ToolName.Get()] = eu
			}
		}
	}
	return authMap
}

func buildNetworkWithAuth(conn openapi.Connection, authMap map[string]*openapi.ExposedUrl, spec *displayer.Spec) displayer.Network {
	networkBuilder := displayer.NewNetworkBuilder().
		Name(conn.Name).
		ToolType(conn.ToolType).
		ConnectionType(conn.ConnectionType)

	if conn.Url != nil {
		networkBuilder.Url(*conn.Url)
	}
	if conn.ContainerPort != nil {
		networkBuilder.Port(*conn.ContainerPort)
	}

	enrichNetworkAuth(networkBuilder, conn, authMap, spec)

	return networkBuilder.Build()
}

// enrichNetworkAuth sets auth from: 1) ExposedUrls, 2) ServingPort, 3) Connection fields
func enrichNetworkAuth(networkBuilder *displayer.NetworkBuilder, conn openapi.Connection, authMap map[string]*openapi.ExposedUrl, spec *displayer.Spec) {
	if eu, ok := authMap[conn.Name]; ok {
		enrichNetworkFromExposedUrl(networkBuilder, eu)
		return
	}

	if conn.ConnectionType == displayer.ConnectionTypeServingPort && spec != nil && spec.ServingPort != nil {
		enrichNetworkFromServingPort(networkBuilder, spec.ServingPort)
		return
	}

	setNetworkAuthFields(networkBuilder, conn.AuthorizationType, conn.AuthorizedUsers, conn.AuthorizedGroups)
}

func enrichNetworkFromExposedUrl(networkBuilder *displayer.NetworkBuilder, eu *openapi.ExposedUrl) {
	var authType *string
	if eu.AuthorizationType.IsSet() && eu.AuthorizationType.Get() != nil {
		val := string(*eu.AuthorizationType.Get())
		authType = &val
	}
	setNetworkAuthFields(networkBuilder, authType, eu.AuthorizedUsers, eu.AuthorizedGroups)
}

func enrichNetworkFromServingPort(networkBuilder *displayer.NetworkBuilder, sp *openapi.ServingPort) {
	var authType *string
	if sp.AuthorizationType.IsSet() && sp.AuthorizationType.Get() != nil {
		val := string(*sp.AuthorizationType.Get())
		authType = &val
	}
	setNetworkAuthFields(networkBuilder, authType, sp.AuthorizedUsers, sp.AuthorizedGroups)
}

func setNetworkAuthFields(networkBuilder *displayer.NetworkBuilder, authType *string, users, groups []string) {
	if authType != nil {
		networkBuilder.AuthorizationType(*authType)
	}
	if len(users) > 0 {
		networkBuilder.AuthorizedUsers(users)
	}
	if len(groups) > 0 {
		networkBuilder.AuthorizedGroups(groups)
	}
}

func setAdditionalFields(builder *WorkloadBuilder, fields map[string]interface{}) error {
	err := setIssues(builder, fields)
	if err != nil {
		return err
	}

	err = setGPUNetworkAccelerationStatus(builder, fields)
	if err != nil {
		return err
	}

	return nil
}
