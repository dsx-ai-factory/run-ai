/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package notifications

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/gpucompat"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	legacyconnections "github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer/composer_v1/connections"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/itemized"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
)

const (
	runaiWorkloadGroup               = "run.ai"
	runaiWorkloadVersionV2Alpha1     = "v2alpha1"
	runaiDistributedInferenceVersion = "v1alpha1"
)

type NotificationsComposer struct {
	fc     *composer.FieldsComposer
	Output *workloadsApi.WorkloadCreationFields

	AssetsMapping           []workloadsApi.AssetsMapping
	connectionsPreProcessor *composer.ConnectionsPreProcessor
}

func NewNotificationsComposer(
	workloadType string,
	workloadName string,
	workloadId string,
	requester string,
	clusterId string,
	projectId string,
	departmentId string,
	createdAt time.Time,
	assetsMapping []workloadsApi.AssetsMapping,
	options composer.FieldsComposerOptions,
) *NotificationsComposer {
	group, version := nativeWorkloadGroupVersion(workloadType)
	result := NotificationsComposer{
		fc: &composer.FieldsComposer{
			WorkloadType:      workloadType,
			SubmittedBy:       requester,
			Options:           options,
			ItemizedNumbering: make(itemized.ItemizedNumbering),
			OauthPortNumber:   composer.Oauth2ContainerDefaultPort,
		},
		AssetsMapping: assetsMapping,
		Output: &workloadsApi.WorkloadCreationFields{
			Type:         workloadType,
			Group:        group,
			Version:      version,
			Name:         workloadName,
			Id:           workloadId,
			ClusterId:    clusterId,
			ProjectId:    projectId,
			DepartmentId: convutils.Ref(departmentId),
			CreatedAt:    createdAt,
		},
	}

	if workloadType != string(workloadsApi.WORKLOADTYPE_DISTRIBUTED) && workloadType != string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE) {
		result.Output.RequestedResources.Set(workloadsApi.NewWorkloadRequestResources())
	}

	return &result
}

func nativeWorkloadGroupVersion(workloadType string) (*string, *string) {
	group := runaiWorkloadGroup
	version := runaiWorkloadVersionV2Alpha1
	switch workloadType {
	case string(workloadsApi.WORKLOADTYPE_WORKSPACE),
		string(workloadsApi.WORKLOADTYPE_TRAINING),
		string(workloadsApi.WORKLOADTYPE_DISTRIBUTED),
		string(workloadsApi.WORKLOADTYPE_INFERENCE),
		constants.KindInteractiveWorkload,
		constants.KindTrainingWorkload,
		constants.KindDistributedWorkload,
		constants.KindInferenceWorkload:
		return convutils.Ref(group), convutils.Ref(version)
	case string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE),
		constants.KindDistributedInferenceWorkload:
		return convutils.Ref(group), convutils.Ref(runaiDistributedInferenceVersion)
	default:
		return nil, nil
	}
}

// WithConnectionsPreProcessor wires the connections pre-processor. WorkloadRunner
// IW/TW composers pass "job" so port and URL connections can be correlated with
// the networking resources they create. Kubeflow composers pass "kf"; multi-
// worker networking is rejected before resource names require replica indexes.
// V1 (RunaiJob) callers pass "" to keep using the native workload initials
// (iw/tw/dw/infw).
func (nsw *NotificationsComposer) WithConnectionsPreProcessor(projectName string, resourceNamePrefix string) *NotificationsComposer {
	nsw.connectionsPreProcessor = &composer.ConnectionsPreProcessor{
		WorkloadName:                  nsw.Output.Name,
		ProjectName:                   projectName,
		ResourceNamePrefix:            resourceNamePrefix,
		WorkloadV2NetworkingSupported: nsw.fc.Options.WorkloadV2NetworkingSupported,
	}
	if resourceNamePrefix == "" {
		nsw.connectionsPreProcessor.PortResourceNames = func(ports *field_types.PortsField) map[string]composer.ServiceAndPortName {
			return legacyconnections.PortResourceNames(workloadsApi.WorkloadType(nsw.Output.Type), nsw.Output.Name, ports)
		}
	}

	return nsw
}

func (nsw *NotificationsComposer) Compose(ctx context.Context, replicaType *workloadsApi.ReplicaType,
	assetsInfo *composer.ComplimentaryAssetInfo, src *workloadsApi.SupersetSpec,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		runner := &composer.StepRunner{}

		if nsw.fc.WorkloadType == string(workloadsApi.WORKLOADTYPE_DISTRIBUTED) {
			err := nsw.ComposeDistributed(ctx, replicaType, src)
			if err != nil {
				return err
			}
		}

		var portAuthInfo composer.PortAuthInfo

		runner.Batch(
			nsw.ComposeRequestedPods(replicaType, src, &nsw.Output.RequestedPods, requestedPodsSetter, nil),
			nsw.fc.ComposeFullySaasInd(&nsw.Output.FullySaasWorkload, composer.BoolToBool, nil),
			nsw.fc.CopyString(convutils.Ref(nsw.fc.SubmittedBy), &nsw.Output.SubmittedBy, composer.StringToString, nil),
			nsw.fc.ComposeImage(src.Image.Get(), &nsw.Output.Images, composer.AddStringToArray, nil),
			nsw.ComposeEnvironment(&nsw.Output.Environments, replicaType, assetsInfo),
			nsw.ComposeExposedUrls(ctx, replicaType, src.ExposedUrls, &nsw.Output.Environments, src.DistributedFramework.Get()),
			nsw.ComposePorts(ctx, replicaType, src.Ports, &nsw.Output.Environments, src.DistributedFramework.Get()),
			nsw.ComposeRelatedUrls(ctx, src.RelatedUrls, &nsw.Output.ExternalConnections),
			nsw.fc.ComposeServingPort(replicaType, src.ServingPort.Get(), &nsw.Output.Environments, servingPortToConnection, nil, &portAuthInfo),
			nsw.fc.CopyArray(src.NodePools, &nsw.Output.RequestedNodePools, composer.AddStringsAoArray, nil),
			nsw.fc.CopyString(src.Category.Get(), &nsw.Output.Category, composer.StringToString, nil),
			nsw.fc.ComposePreemptibility(ctx, src.Preemptibility.Get(), &nsw.Output.Preemptibility, composer.StringToString, nil),
			nsw.ComposeCompute(ctx, src.Compute.Get()),
			nsw.ComposeStorage(assetsInfo, &nsw.Output.Datasources),
		)

		return runner.Error()
	}
}

func (nsw *NotificationsComposer) ComposeRequestedPods(_ *workloadsApi.ReplicaType, src *workloadsApi.SupersetSpec,
	target any, setter composer.FnRequestedPodsSetter, cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src == nil {
			return workloadsApi.NewError(http.StatusBadRequest, "cannot determine requested number of pods: spec is not provided")
		}

		var result workloadsApi.RequestedPods

		switch nsw.fc.WorkloadType {
		case string(workloadsApi.WORKLOADTYPE_DISTRIBUTED_INFERENCE):
			workersAndLeaderCount := 1 + convutils.Deref(src.Workers.Get(), 0) // leader is always present

			totalCount := workersAndLeaderCount * convutils.Deref(src.Replicas.Get(), 1)

			result.Number.Set(convutils.Ref(totalCount))

		case string(workloadsApi.WORKLOADTYPE_TRAINING):
			result.Number.Set(convutils.Ref(int32(1)))

			if convutils.HasValue(src.Parallelism) {
				result.Parallelism.Set(src.Parallelism.Get())
			}

			if convutils.HasValue(src.Completions) {
				result.Completions.Set(src.Parallelism.Get())
			}

		case string(workloadsApi.WORKLOADTYPE_DISTRIBUTED):
			result.Number.Set(convutils.Ref(int32(1) + convutils.Deref(src.NumWorkers.Get(), 0)))

		default:
			result.Number.Set(convutils.Ref(int32(1)))
		}

		err := setter(&result, target, cookie)
		if err != nil {
			return err
		}

		return nil
	}
}

func (nsw *NotificationsComposer) ComposeDistributed(_ context.Context, replicaType *workloadsApi.ReplicaType, src *workloadsApi.SupersetSpec) *workloadsApi.Error {
	return nsw.ComposeJobType(composer.EnumToString(src.DistributedFramework.Get()), replicaType, &nsw.Output.DistributedFramework, composer.StringToString, nil)()
}

func (nsw *NotificationsComposer) ComposeJobType(src *string, replicaType *workloadsApi.ReplicaType,
	target any, setter composer.FnStringSetter, cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if !composer.IsMainNode(replicaType) {
			return nil
		}

		if src == nil {
			if !nsw.fc.Options.IgnoreMissingValues {
				return workloadsApi.NewError(http.StatusBadRequest, "mandatory distributedFramework is not provided")
			}
			return nil
		}

		frameworkWithJobSuffix := fmt.Sprintf("%sJob", *src)

		return setter(&frameworkWithJobSuffix, target, cookie)
	}
}

func (nsw *NotificationsComposer) ComposeCompute(ctx context.Context,
	src *workloadsApi.WorkspaceSpecSpecAllOfCompute,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src == nil {
			return nil
		}

		canonical, err := gpucompat.NormalizeLegacyWorkloadCompute(src)
		if err != nil {
			return workloadsApi.NewError(400, err.Error())
		}
		src = canonical

		if !nsw.Output.RequestedResources.IsSet() {
			return nil
		}

		var sizingInfo composer.GpuSizingInfo

		runner := &composer.StepRunner{}

		runner.Batch(
			nsw.fc.ComposeCpuCoreRequest(src.CpuCoreRequest.Get(), &nsw.Output.RequestedResources.Get().Cpu, nonZeroFloatToResourcesCore, nil),
			nsw.fc.ComposeCpuMemoryRequest(src.CpuMemoryRequest.Get(), &nsw.Output.RequestedResources.Get().CpuMemory, nonEmptyStringToResourcesQuantity, nil),
			nsw.fc.ComposeGpuRequestsAndLimits(ctx, src.GpuDevicesRequest.Get(), src.GpuRequestType.Get(),
				src.GpuMemoryPortionRequest.Get(), src.GpuMemoryPortionLimit.Get(),
				src.GpuMemoryRequest.Get(), src.GpuMemoryLimit.Get(), gpuRequestedResourcesSetter, &nsw.Output.RequestedResources, nil, &sizingInfo),
		)

		return runner.Error()
	}
}

func (nsw *NotificationsComposer) ComposeEnvironment(target *[]workloadsApi.Environment, replicaType *workloadsApi.ReplicaType,
	assetsInfo *composer.ComplimentaryAssetInfo,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		//  Create an environment, which represents the pod in which this replica runs.
		//  In workspace or training, there is one replica and so there will be one instance
		//  of environment. In distributed, there may be two replicas, and we se tup one
		//  environment for each
		environment := workloadsApi.Environment{
			Name:        convutils.Deref((*string)(replicaType), ""),
			ReplicaType: *workloadsApi.NewNullableReplicaType(replicaType),
			Connections: make([]workloadsApi.Connection, 0),
		}

		if assetsInfo != nil {
			environment.Name = assetsInfo.Environment.Name
			environment.Id = assetsInfo.Environment.Id
		}

		if *target == nil {
			*target = []workloadsApi.Environment{}
		}

		*target = append(*target, environment)

		return nil
	}
}

func (nsw *NotificationsComposer) ComposeExposedUrls(_ context.Context, replicaType *workloadsApi.ReplicaType, urls []workloadsApi.ExposedUrl,
	target *[]workloadsApi.Environment, framework *workloadsApi.DistributedFramework,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if len(urls) == 0 {
			return nil
		}

		environment := EnvironmentByReplica(target, replicaType)
		if environment == nil {
			return nil
		}

		// for the connections pre-processor we need the urls in cluster format
		exposedUrls := make(map[string]field_types.ExposedUrlItemAndRules)

		connections := make(map[string]*workloadsApi.Connection)

		for _, entry := range urls {
			if convutils.BoolIsTrue(entry.Exclude.Get()) {
				continue
			}

			urlItemId := nsw.fc.ItemizedNumbering.Compose(itemized.ExposedUrlPrefix, entry.Name.Get(), replicaType, false)

			workloadApiConnection := workloadsApi.Connection{
				Name:            convutils.Deref(entry.ToolName.Get(), urlItemId),
				ToolType:        convutils.Deref(entry.ToolType.Get(), composer.CustomToolType),
				ConnectionType:  string(openapi.INTERNALCONNECTIONTYPE_EXTERNAL_URL),
				ContainerPort:   entry.Container.Get(),
				CorrelationName: convutils.Ref(urlItemId),
			}

			authType := convutils.Deref(entry.AuthorizationType.Get(), "")

			if authType == "" || authType == workloadsApi.AUTHORIZATIONTYPE_AUTHORIZED_GROUPS {
				workloadApiConnection.AuthorizedGroups = entry.AuthorizedGroups
			}

			if authType == "" || authType == workloadsApi.AUTHORIZATIONTYPE_AUTHORIZED_USERS {
				workloadApiConnection.AuthorizedUsers = entry.AuthorizedUsers
			}

			// we need a mapping from nick to custom url for the connections pre-processor
			if convutils.HasValue(entry.Url) {
				workloadApiConnection.Url = entry.Url.Get()

				exposedUrls[urlItemId] = field_types.ExposedUrlItemAndRules{
					Value: &field_types.ExposedUrlItemValue{
						Url:       entry.Url.Get(),
						CustomUrl: convutils.Ref(true),
					},
				}
			} else {
				exposedUrls[urlItemId] = field_types.ExposedUrlItemAndRules{
					Value: &field_types.ExposedUrlItemValue{
						CustomUrl: convutils.Ref(false),
					},
				}
			}

			connections[urlItemId] = &workloadApiConnection
		}

		//
		//   add it sorted back to the environment
		//
		environment.Connections = make([]workloadsApi.Connection, 0)

		for _, itemId := range composer.SortedMapKeys(connections) {
			environment.Connections = append(environment.Connections, *connections[itemId])
		}

		if nsw.connectionsPreProcessor != nil {
			nsw.connectionsPreProcessor.HandleExposedUrls(environment.Connections, field_types.NewExposedUrlField(exposedUrls),
				replicaType, framework)
		}

		return nil
	}
}

func (nsw *NotificationsComposer) ComposePorts(_ context.Context, replicaType *workloadsApi.ReplicaType, ports []workloadsApi.Port,
	target *[]workloadsApi.Environment, framework *workloadsApi.DistributedFramework,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if len(ports) == 0 {
			return nil
		}

		environment := EnvironmentByReplica(target, replicaType)
		if environment == nil {
			return nil
		}

		exposedPorts := make(map[string]field_types.PortItemAndRules)

		for _, entry := range ports {
			portItemId := nsw.fc.ItemizedNumbering.Compose(itemized.PortPrefix, entry.Name.Get(), replicaType, false)

			if convutils.BoolIsTrue(entry.Exclude.Get()) {
				continue
			}

			if convutils.HasValue(entry.ServiceType) {
				serviceType := string(convutils.Deref(entry.ServiceType.Get(), ""))

				containerPort := int(convutils.Deref(entry.Container.Get(), 0))
				externalPort := int(convutils.Deref(entry.External.Get(), 0))

				exposedPorts[portItemId] = field_types.PortItemAndRules{
					Value: &field_types.PortItemValue{
						ContainerPort: convutils.Ref(containerPort),
						ServiceType:   convutils.Ref(serviceType),
						HostPort:      convutils.Ref(externalPort),
					},
				}

				// cluster-ip and node-port both reported as node-port (backward compatibility)
				// and now we also have load balancer as well
				if serviceType == string(openapi.PORTSERVICETYPE_CLUSTER_IP) {
					serviceType = string(openapi.INTERNALCONNECTIONTYPE_NODE_PORT)
				}

				if serviceType != "" {
					name := convutils.Deref(entry.ToolName.Get(), portItemId)

					environment.Connections = append(environment.Connections, workloadsApi.Connection{
						Name:            name,
						ToolType:        convutils.Deref(entry.ToolType.Get(), ""),
						ConnectionType:  serviceType,
						Url:             nil,
						ContainerPort:   entry.Container.Get(),
						CorrelationName: convutils.Ref(portItemId),
					})
				}
			}

		}

		if nsw.connectionsPreProcessor != nil {
			nsw.connectionsPreProcessor.HandlePorts(environment.Connections, field_types.NewPortsField(exposedPorts),
				replicaType, framework)
		}

		return nil
	}
}

func (nsw *NotificationsComposer) ComposeRelatedUrls(_ context.Context, urls []workloadsApi.RelatedUrl,
	target *[]workloadsApi.Connection,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if len(urls) == 0 {
			return nil
		}

		if *target == nil {
			*target = []workloadsApi.Connection{}
		}

		for _, entry := range urls {
			if convutils.BoolIsTrue(entry.Exclude.Get()) {
				continue
			}

			*target = append(*target, workloadsApi.Connection{
				Name:     entry.GetName(),
				ToolType: convutils.Deref(entry.Type.Get(), composer.CustomToolType),
				Url:      convutils.Ref(entry.GetUrl()),
			})
		}

		return nil
	}
}

func (nsw *NotificationsComposer) ComposeStorage(assetsInfo *composer.ComplimentaryAssetInfo,
	datasources *[]workloadsApi.Datasource,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if assetsInfo == nil || len(assetsInfo.Datasources) == 0 {
			return nil
		}

		if *datasources == nil {
			*datasources = []workloadsApi.Datasource{}
		}

		knownAssets := datastructs.NewSet[workloadsApi.Datasource]()

		for _, datasource := range *datasources {
			knownAssets.Add(datasource)
		}

		for _, dsPerKind := range assetsInfo.Datasources {
			for _, datasource := range dsPerKind {
				if !knownAssets.Has(datasource) {
					*datasources = append(*datasources, datasource)
				}
			}
		}

		return nil
	}
}

func EnvironmentByReplica(environments *[]workloadsApi.Environment, replicaType *workloadsApi.ReplicaType) *workloadsApi.Environment {
	if *environments == nil {
		return nil
	}

	for idx, env := range *environments {
		if convutils.Deref(env.ReplicaType.Get(), "") == convutils.Deref(replicaType, "") {
			return &((*environments)[idx])
		}
	}

	return nil
}
