/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	"math"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/common"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/dao"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/clusterservice"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/services_utils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/dbmodels"
	telemetry_utils "github.com/dsx-ai-factory/run-ai/runai-application-packages/telemetry"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	"k8s.io/utils/ptr"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

const (
	CsvAccept                   = "text/csv"
	PriorityPreemptionThreshold = 100

	RunaiCrdWorkloadTypeGroup = "run.ai"
	// HistogramPlusInfBucketKey is the canonical key used for the +Inf bucket in Prometheus histograms
	HistogramPlusInfBucketKey = "+Inf"
)

func NoContentResponse() api.ImplResponse {
	return api.ImplResponse{
		Code: http.StatusNoContent,
	}
}

func NewErrorResponse(apiError *api.Error) api.ImplResponse {
	apiError.Details = ""
	return api.ImplResponse{
		Code: int(apiError.Code),
		Body: *apiError,
	}
}

func ErrorToApiError(err error, errorCode int) api.ImplResponse {
	return api.ImplResponse{
		Code: errorCode,
		Body: err.Error(),
	}
}

func historyRecordToApiFromFields(createdAt time.Time, eventID *uuid.UUID, typeStr *string, clusterID uuid.UUID, message, reason, source *string,
	involvedObjectUid, involvedObjectKind, involvedObjectName, involvedObjectNamespace *string, phase, phaseMessage *string,
) api.HistoryRecord {
	var recordType api.HistoryRecordType
	var spec api.HistoryRecordSpec

	if eventID != nil {
		recordType = api.HISTORYRECORDTYPE_EVENT
		spec = api.HistoryRecordSpec{
			Event: &api.Event{
				CreatedAt: createdAt,
				Id:        eventID.String(),
				Type:      *typeStr,
				ClusterId: clusterID.String(),
				Message:   *message,
				Reason:    *reason,
				Source:    *source,
				InvolvedObject: api.InvolvedObject{
					Uid:       *involvedObjectUid,
					Kind:      *involvedObjectKind,
					Name:      *involvedObjectName,
					Namespace: *involvedObjectNamespace,
				},
			},
		}
	} else {
		recordType = api.HISTORYRECORDTYPE_PHASE_UPDATE
		spec = api.HistoryRecordSpec{
			PhaseUpdate: &api.PhaseUpdate{
				Phase:        api.Phase(*phase),
				PhaseMessage: *phaseMessage,
			},
		}
	}

	return api.HistoryRecord{
		Meta: api.HistoryRecordMeta{
			CreationTimestamp: createdAt,
			Type:              recordType,
		},
		Spec: spec,
	}
}

func HistoryRecordModelToApi(record models.WorkloadHistoryRecord) api.HistoryRecord {
	return historyRecordToApiFromFields(
		record.CreatedAt,
		record.EventID,
		record.Type,
		record.ClusterID,
		record.Message,
		record.Reason,
		record.Source,
		record.InvolvedObjectUid,
		record.InvolvedObjectKind,
		record.InvolvedObjectName,
		record.InvolvedObjectNamespace,
		record.Phase,
		record.PhaseMessage,
	)
}

func RevisionHistoryRecordModelToApi(record models.RevisionHistoryRecord) api.HistoryRecord {
	return historyRecordToApiFromFields(
		record.CreatedAt,
		record.EventID,
		record.Type,
		record.ClusterID,
		record.Message,
		record.Reason,
		record.Source,
		record.InvolvedObjectUid,
		record.InvolvedObjectKind,
		record.InvolvedObjectName,
		record.InvolvedObjectNamespace,
		record.Phase,
		record.PhaseMessage,
	)
}

func EventApiToModel(tenantId uint, event api.EventSyncFields) (models.Event, error) {
	id, err := uuid.Parse(event.Id)
	if err != nil {
		return models.Event{}, err
	}

	clusterId, err := uuid.Parse(event.ClusterId)
	if err != nil {
		return models.Event{}, err
	}

	err = uuid.Validate(event.InvolvedObject.Uid)
	if err != nil {
		return models.Event{}, err
	}

	return models.Event{
		TenantID:                tenantId,
		ID:                      id,
		ClusterID:               clusterId,
		Type:                    event.Type,
		Message:                 event.Message,
		Reason:                  event.Reason,
		CreatedAt:               event.CreatedAt,
		Source:                  event.Source,
		InvolvedObjectUid:       event.InvolvedObject.Uid,
		InvolvedObjectKind:      event.InvolvedObject.Kind,
		InvolvedObjectName:      event.InvolvedObject.Name,
		InvolvedObjectNamespace: event.InvolvedObject.Namespace,
	}, nil
}

func EventViewModelToApi(event models.WorkloadEventView) api.Event {
	return api.Event{
		CreatedAt: event.CreatedAt,
		Id:        event.EventID.String(),
		Type:      event.Type,
		ClusterId: event.ClusterID.String(),
		Message:   event.Message,
		Reason:    event.Reason,
		Source:    event.Source,
		InvolvedObject: api.InvolvedObject{
			Uid:       event.InvolvedObjectUid,
			Kind:      event.InvolvedObjectKind,
			Name:      event.InvolvedObjectName,
			Namespace: event.InvolvedObjectNamespace,
		},
	}
}

func RevisionEventViewModelToApi(event models.RevisionEventView) api.Event {
	return api.Event{
		CreatedAt: event.CreatedAt,
		Id:        event.EventID.String(),
		Type:      event.Type,
		ClusterId: event.ClusterID.String(),
		Message:   event.Message,
		Reason:    event.Reason,
		Source:    event.Source,
		InvolvedObject: api.InvolvedObject{
			Uid:       event.InvolvedObjectUid,
			Kind:      event.InvolvedObjectKind,
			Name:      event.InvolvedObjectName,
			Namespace: event.InvolvedObjectNamespace,
		},
	}
}

func WorkloadModelToApi(ctx context.Context, workload models.WorkloadView) api.Workload {
	actionsSupport := api.ActionsSupport{
		Delete:  workload.SupportDelete,
		Suspend: workload.SupportSuspend,
	}

	if workload.SourceApi != nil && *workload.SourceApi == api.SOURCEAPI_WORKLOADS_V2 {
		actionsSupport.Delete = true
	}

	err := services_utils.CheckWorkloadOwnershipProtection(ctx, workload.ClusterID.String(), workload.DepartmentID, workload.SubmittedBy)
	if err != nil {
		actionsSupport = api.ActionsSupport{
			Delete:  false,
			Suspend: false,
		}
	}

	apiWorkload := api.Workload{
		Type:                 workload.Type,
		Name:                 workload.Name,
		ExternalWorkloadName: workload.ExternalWorkloadName,
		Source:               api.Source(workload.Source),
		SourceApi:            workload.SourceApi,
		Id:                   workload.ID.String(),
		Priority:             workload.Priority,
		PriorityClassName:    workload.PriorityClassName,
		Preemptible:          workload.Preemptible,
		Preemptibility:       workload.Preemptibility,
		CompletedAt:          getTimeOrNil(workload.CompletedAt.Time),
		SubmittedBy:          workload.SubmittedBy,
		ClusterId:            workload.ClusterID.String(),
		ProjectName:          workload.ProjectName,
		ProjectId:            workload.ProjectID,
		DepartmentName:       workload.DepartmentName,
		DepartmentId:         workload.DepartmentID,
		Namespace:            workload.Namespace,
		CreatedAt:            workload.CreatedAt,
		WorkloadRequestedResources: getWorkloadRequestedResources(
			workload.RequestedGPULimit,
			workload.RequestedGPURequest,
			workload.RequestedGPUMemoryLimit,
			workload.RequestedGPUMemoryRequest,
			workload.RequestedCPULimit,
			workload.RequestedCPURequest,
			workload.RequestedCPUMemoryLimit,
			workload.RequestedCPUMemoryRequest,
		),
		PodsRequestedResources: getPodsRequestedResources(
			workload.PodsRequestedGPULimit,
			workload.PodsRequestedGPURequest,
			workload.PodsRequestedGPUMemoryLimit,
			workload.PodsRequestedGPUMemoryRequest,
			workload.PodsRequestedCPULimit,
			workload.PodsRequestedCPURequest,
			workload.PodsRequestedCPUMemoryLimit,
			workload.PodsRequestedCPUMemoryRequest,
		),
		AllocatedResources: &api.WorkloadAllocatedResources{
			Gpu:       workload.AllocatedGPU,
			GpuMemory: sumQuantities(workload.AllocatedGPUMemory),
			Cpu:       workload.AllocatedCPU,
			CpuMemory: sumQuantities(workload.AllocatedCPUMemory),
		},
		Conditions:           convutils.ConvertSliceTypes(workload.Conditions, conditionModelToApi),
		Datasources:          convutils.ConvertSliceTypes(workload.Datasources, dataSourceModelToApi),
		Environments:         convutils.ConvertSliceTypes(workload.Environments, environmentModelToApi),
		ExternalConnections:  convutils.ConvertSliceTypes(workload.ExternalConnections, connectionModelToApi),
		DistributedFramework: workload.DistributedFramework,
		ActionsSupport:       actionsSupport,
		Phase:                workload.Phase,
		PhaseMessage:         workload.PhaseMessage,
		K8sPhase:             workload.K8sPhase,
		RequestedPods: api.RequestedPods{
			Number:      workload.RequestedPodsNumber,
			Min:         workload.RequestedPodsMin,
			Max:         workload.RequestedPodsMax,
			Parallelism: workload.RequestedPodsParallelism,
			Completions: workload.RequestedPodsCompletions,
		},
		RequestedNodePools:      workload.RequestedNodepools,
		CurrentNodePools:        workload.CurrentNodepools,
		Images:                  workload.Images,
		Urls:                    workload.Urls,
		AdditionalFields:        workload.AdditionalFields,
		FullySaasWorkload:       workload.FullySaasWorkload,
		TenantId:                int32(workload.TenantID),
		RunningPods:             workload.NumberOfRunningPods,
		PhaseUpdatedAt:          workload.PhaseUpdatedAt,
		K8sPhaseUpdatedAt:       workload.K8sPhaseUpdatedAt,
		UpdatedAt:               workload.UpdatedAt,
		DeletedAt:               dbmodels.GormDeletedAtToNullableTime(workload.DeletedAt),
		Command:                 workload.Command,
		Arguments:               workload.Arguments,
		EnvironmentVariables:    workload.EnvironmentVariables,
		IdleAllocatedGpus:       workload.IdleAllocatedGpus,
		IdleSince:               workload.IdleSince,
		TotalPendingTimeSeconds: workload.TotalPendingTimeSeconds,
		TotalRunningTimeSeconds: workload.TotalRunningTimeSeconds,
		Category:                workload.Category,
		GuaranteedRuntimeEndsAt: workload.GuaranteedRuntimeEndsAt,
		NodeAffinities:          convutils.ConvertSliceTypes(workload.NodeAffinities, workloadNodeAffinityModelToApi),
		KubernetesRef:           buildKubernetesRef(workload.K8sObjectGroup, workload.K8sObjectVersion),
		// NOTE to dev: if you add fields here, remember to add them in "WorkloadModelToWorkloadDetailedApi" as well!
	}
	if workload.ApplicationId != nil && workload.ApplicationName != nil {
		apiWorkload.AiApplicationId = *workload.ApplicationId
		apiWorkload.AiApplicationName = *workload.ApplicationName
	}

	if workload.Topology != nil {
		apiWorkload.Topology = &api.WorkloadTopology{
			Name:               workload.Topology.Name,
			PreferredPlacement: workload.Topology.PreferredPlacement,
			RequiredPlacement:  workload.Topology.RequiredPlacement,
		}
	}

	if len(apiWorkload.Environments) > 0 {
		_, err := patchConnections(ctx, apiWorkload.Environments, workload.Urls, workload.NetworkStatus)
		if err != nil {
			// if the workload is not running - makes sense there are no urls yet, no need to log
			if workload.Phase == api.PHASE_RUNNING {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf(`in workload <%s> could not match connections to urls, err: %v`, workload.Name, err)
			}
		}

		if common.WorkloadType(workload.Type).IsNative() {
			apiWorkload.Endpoints = endpointsFromConnections(apiWorkload.Environments)
		}
	}

	// if preemptible is not set, we determine it based on priority
	if apiWorkload.Preemptible == nil && workload.Priority != nil {
		// this is the scheduler way of deciding which workload is preemptible - also specified in the workload view
		preemptible := *(workload.Priority) < PriorityPreemptionThreshold
		apiWorkload.Preemptible = &preemptible
	}
	if workload.PhaseReason != nil && workload.Phase == api.PHASE_PENDING {
		phaseReason := api.PhaseReason(*workload.PhaseReason)
		if phaseReason.IsValid() {
			apiWorkload.PhaseReason = &phaseReason
		}
	}

	if len(workload.Endpoints) > 0 && !common.WorkloadType(workload.Type).IsNative() {
		apiEndpoints := ModelEndpointToApi(workload.Endpoints)
		apiWorkload.Endpoints = apiEndpoints
	}
	return apiWorkload
}

func ModelEndpointToApi(endpoints models.EndpointArray) []api.Endpoint {
	apiEndpoints := []api.Endpoint{}
	for _, endpoint := range endpoints {
		apiEndpoints = append(apiEndpoints, api.Endpoint{
			Host:       endpoint.Host,
			Port:       endpoint.Port,
			Path:       endpoint.Path,
			TlsEnabled: endpoint.TlsEnabled,
			Protocol:   endpoint.Protocol,
			Url:        endpoint.Url,
		})
	}
	return apiEndpoints
}

func WorkloadModelToWorkloadDetailedApi(ctx context.Context, workload models.WorkloadView) api.WorkloadDetailed {
	apiWorkload := WorkloadModelToApi(ctx, workload)
	var pendingSchedulingMessages *[]api.PendingSchedulingMessage
	if workload.Phase == api.PHASE_PENDING {
		pendingSchedulingMessages = getPendingSchedulingMessages(workload.SchedulingConditions)
	}

	topology := apiWorkload.Topology
	if topologyName := resolveTopologyNameForWorkload(ctx, workload); topologyName != "" {
		if topology == nil {
			topology = &api.WorkloadTopology{}
		}
		if topology.Name == "" {
			topology.Name = topologyName
		}
		topology.ActualPlacement = getActualTopologyPlacementForWorkload(ctx, workload.TenantID, workload.ID, workload.ClusterID.String(), topologyName)
	}

	return api.WorkloadDetailed{
		TenantId:                   apiWorkload.TenantId,
		RunningPods:                apiWorkload.RunningPods,
		PhaseUpdatedAt:             apiWorkload.PhaseUpdatedAt,
		K8sPhaseUpdatedAt:          apiWorkload.K8sPhaseUpdatedAt,
		UpdatedAt:                  apiWorkload.UpdatedAt,
		Source:                     apiWorkload.Source,
		SourceApi:                  apiWorkload.SourceApi,
		DeletedAt:                  apiWorkload.DeletedAt,
		Type:                       apiWorkload.Type,
		Name:                       apiWorkload.Name,
		ExternalWorkloadName:       apiWorkload.ExternalWorkloadName,
		Id:                         apiWorkload.Id,
		Priority:                   apiWorkload.Priority,
		PriorityClassName:          apiWorkload.PriorityClassName,
		SubmittedBy:                apiWorkload.SubmittedBy,
		ClusterId:                  apiWorkload.ClusterId,
		ProjectName:                apiWorkload.ProjectName,
		ProjectId:                  apiWorkload.ProjectId,
		DepartmentName:             apiWorkload.DepartmentName,
		DepartmentId:               apiWorkload.DepartmentId,
		Namespace:                  apiWorkload.Namespace,
		CreatedAt:                  apiWorkload.CreatedAt,
		WorkloadRequestedResources: apiWorkload.WorkloadRequestedResources,
		PodsRequestedResources:     apiWorkload.PodsRequestedResources,
		AllocatedResources:         apiWorkload.AllocatedResources,
		ActionsSupport:             apiWorkload.ActionsSupport,
		Phase:                      apiWorkload.Phase,
		Conditions:                 apiWorkload.Conditions,
		PhaseMessage:               apiWorkload.PhaseMessage,
		K8sPhase:                   apiWorkload.K8sPhase,
		RequestedPods:              apiWorkload.RequestedPods,
		RequestedNodePools:         apiWorkload.RequestedNodePools,
		CurrentNodePools:           apiWorkload.CurrentNodePools,
		CompletedAt:                apiWorkload.CompletedAt,
		Images:                     apiWorkload.Images,
		ChildrenIds:                apiWorkload.ChildrenIds,
		Urls:                       apiWorkload.Urls,
		Datasources:                apiWorkload.Datasources,
		Environments:               apiWorkload.Environments,
		ExternalConnections:        apiWorkload.ExternalConnections,
		DistributedFramework:       apiWorkload.DistributedFramework,
		AdditionalFields:           apiWorkload.AdditionalFields,
		Preemptible:                apiWorkload.Preemptible,
		Preemptibility:             apiWorkload.Preemptibility,
		FullySaasWorkload:          apiWorkload.FullySaasWorkload,
		EnvironmentVariables:       apiWorkload.EnvironmentVariables,
		Command:                    apiWorkload.Command,
		Arguments:                  apiWorkload.Arguments,
		PhaseReason:                apiWorkload.PhaseReason,
		IdleAllocatedGpus:          apiWorkload.IdleAllocatedGpus,
		IdleSince:                  apiWorkload.IdleSince,
		TotalPendingTimeSeconds:    apiWorkload.TotalPendingTimeSeconds,
		TotalRunningTimeSeconds:    apiWorkload.TotalRunningTimeSeconds,
		Endpoints:                  apiWorkload.Endpoints,
		NodeAffinities:             apiWorkload.NodeAffinities,

		// fields that are added only in WorkloadDetailed:
		PendingSchedulingMessages: pendingSchedulingMessages,
		Category:                  apiWorkload.Category,
		GuaranteedRuntimeEndsAt:   apiWorkload.GuaranteedRuntimeEndsAt,
		AiApplicationId:           apiWorkload.AiApplicationId,
		AiApplicationName:         apiWorkload.AiApplicationName,
		Topology:                  topology,
		KubernetesRef:             apiWorkload.KubernetesRef,
	}
}

// Native workload parsers forward Group (used elsewhere for category lookup)
// but never Version, so a populated Group alone signals a native workload —
// for which kubernetesRef must stay null.
func buildKubernetesRef(group, version *string) *api.WorkloadKubernetesRef {
	if group == nil || version == nil || *group == "" || *version == "" {
		return nil
	}
	return &api.WorkloadKubernetesRef{
		Group:   *group,
		Version: *version,
	}
}

func TelemetryGroupsToApi(telemetryType string, timestamp time.Time, groups []telemetry_utils.TelemetryGroupValue, groupBy []string) *api.TelemetryResponse {
	telemetryGroups := &[]api.TelemetryResponseValuesInner{}
	for _, group := range groups {
		telemetryGroupApi := TelemetryGroupToApi(group, groupBy)
		*telemetryGroups = append(*telemetryGroups, *telemetryGroupApi)
	}
	telemetryResponse := api.TelemetryResponse{
		Type:      telemetryType,
		Timestamp: timestamp,
		Values:    *telemetryGroups,
	}
	return &telemetryResponse
}

func keyValueGroupToApi(fieldName string, apiGroups []api.TelemetryResponseValuesInnerGroupsInner, group1Value string, group2Value string) []api.TelemetryResponseValuesInnerGroupsInner {
	switch fieldName {
	case ClusterId, DepartmentId, ProjectId, Bucket:
		return append(apiGroups, api.TelemetryResponseValuesInnerGroupsInner{
			Key:   fieldName,
			Value: group1Value,
			Name:  group2Value,
		})
	case CurrentNodepools:
		value := group1Value
		if strings.HasPrefix(group1Value, "{") && strings.HasSuffix(group1Value, "}") {
			value = group1Value[1 : len(group1Value)-1]
		}
		return append(apiGroups, api.TelemetryResponseValuesInnerGroupsInner{
			Key:   fieldName,
			Value: value,
		})
	default:
		return append(apiGroups, api.TelemetryResponseValuesInnerGroupsInner{
			Key:   fieldName,
			Value: group1Value,
		})
	}
}

func TelemetryGroupToApi(group telemetry_utils.TelemetryGroupValue, groupBy []string) *api.TelemetryResponseValuesInner {
	var apiGroups []api.TelemetryResponseValuesInnerGroupsInner
	for i, fieldName := range groupBy {
		if i == 0 {
			apiGroups = keyValueGroupToApi(fieldName, apiGroups, group.Group1MainValue, group.Group1SecondaryValue)
		}
		if i == 1 {
			apiGroups = keyValueGroupToApi(fieldName, apiGroups, group.Group2MainValue, group.Group2SecondaryValue)
		}
	}
	telemetryGroup := api.TelemetryResponseValuesInner{
		Value:  group.Data,
		Groups: apiGroups,
	}
	return &telemetryGroup
}

func CreationRequestToInternalCreatedWorkloadModel(workload api.WorkloadCreationFields) models.InternalCreatedWorkload {
	model := models.InternalCreatedWorkload{
		Datasources:         convutils.ConvertSliceTypes(workload.Datasources, dataSourceApiToModel),
		Environments:        convutils.ConvertSliceTypes(workload.Environments, environmentApiToModel),
		ExternalConnections: convutils.ConvertSliceTypes(workload.ExternalConnections, connectionApiToModel),
		SourceApi:           workload.SourceApi,
		Preemptibility:      workload.Preemptibility,
	}

	return model
}

func dataSourceApiToModel(ds api.Datasource) models.Datasource {
	return models.Datasource{
		Type: ds.Type,
		Name: ds.Name,
		Id:   ds.Id,
	}
}

func dataSourceModelToApi(ds models.Datasource) api.Datasource {
	return api.Datasource{
		Type: ds.Type,
		Name: ds.Name,
		Id:   ds.Id,
	}
}

func environmentApiToModel(env api.Environment) models.Environment {
	return models.Environment{
		Connections: convutils.ConvertSliceTypes(env.Connections, connectionApiToModel),
		Name:        env.Name,
		Id:          env.Id,
		ReplicaType: getStringOrNil(env.ReplicaType),
	}
}

func connectionApiToModel(connection api.Connection) models.Connection {
	return models.Connection{
		Name:                connection.Name,
		Type:                connection.ToolType,
		Url:                 connection.Url,
		AuthorizationType:   connection.AuthorizationType,
		AuthorizedUsers:     connection.AuthorizedUsers,
		AuthorizedGroups:    connection.AuthorizedGroups,
		ConnectionType:      connection.ConnectionType,
		ContainerPort:       connection.ContainerPort,
		CorrelationName:     connection.CorrelationName,
		ExposedUrlSubDomain: connection.ExposedUrlSubDomain,
		ExposedUrlPath:      connection.ExposedUrlPath,
		PortResourceName:    connection.PortResourceName,
		ServiceResourceName: connection.ServiceResourceName,
	}
}

func workloadResourceRefApiToModel(ref api.ResourceRef) models.WorkloadResourceRef {
	return models.WorkloadResourceRef{
		UID:  ref.Uid,
		Name: ref.Name,
		Kind: string(ref.Kind),
	}
}

func endpointsFromConnections(environments []api.Environment) []api.Endpoint {
	var endpoints []api.Endpoint
	for _, env := range environments {
		for _, conn := range env.Connections {
			if conn.Url == "" {
				continue
			}
			endpoints = append(endpoints, connectionToEndpoint(conn))
		}
	}
	return endpoints
}

func connectionToEndpoint(conn api.Connection) api.Endpoint {
	ep := api.Endpoint{Url: conn.Url}

	if conn.Url == "" {
		return ep
	}

	parsed, err := url.Parse(conn.Url)
	if err == nil && parsed.Scheme != "" && parsed.Opaque == "" {
		port, _ := strconv.ParseInt(parsed.Port(), 10, 32)
		ep.Host = parsed.Hostname()
		ep.Port = int32(port)
		ep.Path = parsed.Path
		ep.Protocol = strings.ToUpper(parsed.Scheme)
		ep.TlsEnabled = ptr.To(strings.EqualFold(parsed.Scheme, "https"))
	} else {
		host, portStr, splitErr := net.SplitHostPort(conn.Url)
		port, _ := strconv.ParseInt(portStr, 10, 32)
		if splitErr == nil {
			ep.Host = host
			ep.Port = int32(port)
		}
	}

	if ep.Port == 0 && conn.ContainerPort > 0 {
		ep.Port = conn.ContainerPort
	}
	return ep
}

func connectionModelToApi(connection models.Connection) api.Connection {
	return api.Connection{
		Name:                connection.Name,
		ToolType:            connection.Type,
		Url:                 connection.Url,
		AuthorizationType:   connection.AuthorizationType,
		AuthorizedUsers:     connection.AuthorizedUsers,
		AuthorizedGroups:    connection.AuthorizedGroups,
		ConnectionType:      connection.ConnectionType,
		ContainerPort:       connection.ContainerPort,
		CorrelationName:     connection.CorrelationName,
		ExposedUrlSubDomain: connection.ExposedUrlSubDomain,
		ExposedUrlPath:      connection.ExposedUrlPath,
		PortResourceName:    connection.PortResourceName,
		ServiceResourceName: connection.ServiceResourceName,
	}
}

func getStringOrNil[T ~string](stringLikePtr *T) *string {
	if stringLikePtr == nil {
		return nil
	}

	return ptr.To(string(*stringLikePtr))
}

func strPointerIfNotEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func environmentModelToApi(env models.Environment) api.Environment {
	return api.Environment{
		Connections: convutils.ConvertSliceTypes(env.Connections, connectionModelToApi),
		Name:        env.Name,
		Id:          env.Id,
		ReplicaType: getReplicaTypeOrNil(env.ReplicaType),
	}
}

func getReplicaTypeOrNil(replicaType *string) *api.ReplicaType {
	if replicaType == nil {
		return nil
	}

	return (*api.ReplicaType)(replicaType)
}

func SyncRequestToWorkloadsModel(ctx context.Context, tenantId uint, projectId string, workload api.WorkloadSyncFields) (models.Workload, []models.WorkloadChild, []models.WorkloadElement, error) {
	workloadChildren := []models.WorkloadChild{}
	workloadElements := []models.WorkloadElement{}

	if projectId != workload.ProjectId {
		return models.Workload{}, workloadChildren, workloadElements, fmt.Errorf("workload <%s> is in project <%s> but request is for project <%s>", workload.Name, workload.ProjectId, projectId)
	}

	id, err := uuid.Parse(workload.Id)
	if err != nil {
		return models.Workload{}, workloadChildren, workloadElements, err
	}

	clusterId, err := uuid.Parse(workload.ClusterId)
	if err != nil {
		return models.Workload{}, workloadChildren, workloadElements, err
	}

	additionalFields := &models.AdditionalFields{}
	if workload.AdditionalFields != "" {
		err = additionalFields.Scan([]byte(workload.AdditionalFields))
		if err != nil {
			return models.Workload{}, workloadChildren, workloadElements, err
		}
	}

	if workload.Priority == nil {
		priority, exists := (*additionalFields)["priority"]
		if exists {
			priorityNum, ok := priority.(float64)
			if !ok {
				logging.Warnf("failed to sync workload priority, cannot get priority from additional data %s", priority)
			} else {
				i := int32(priorityNum)
				workload.Priority = &i
			}
		}
	}

	// store network status in the DB only if it has a meaningful data, otherwise it's just a waste of
	// space and time to convert to empty json.
	var networkStatus *api.WorkloadNetworkStatus
	if len(workload.NetworkStatus.Services) != 0 {
		networkStatus = &workload.NetworkStatus
	}

	schedulingConditions, phaseReason := getSchedulingConditionsFromRequest(workload.SchedulingConditions)
	workloadType := common.NewWorkloadType(workload.Type)

	model := models.Workload{
		ID:                       id,
		UpdatedAt:                time.Now(),
		SubmittedBy:              workload.SubmittedBy,
		Name:                     workload.Name,
		ExternalWorkloadName:     workload.ExternalWorkloadName,
		Source:                   string(workload.Source),
		Type:                     workloadType.String(),
		PriorityClassName:        workload.PriorityClassName,
		Priority:                 workload.Priority,
		Preemptibility:           workload.Preemptibility,
		TenantID:                 tenantId,
		ClusterID:                clusterId,
		ProjectName:              workload.ProjectName,
		ProjectID:                workload.ProjectId,
		DepartmentName:           workload.DepartmentName,
		DepartmentID:             workload.DepartmentId,
		Namespace:                workload.Namespace,
		Phase:                    workload.Phase,
		PhaseMessage:             workload.PhaseMessage,
		K8sPhase:                 workload.K8sPhase,
		RequestedPodsNumber:      workload.RequestedPods.Number,
		RequestedPodsMin:         workload.RequestedPods.Min,
		RequestedPodsMax:         workload.RequestedPods.Max,
		RequestedPodsParallelism: workload.RequestedPods.Parallelism,
		RequestedPodsCompletions: workload.RequestedPods.Completions,
		RequestedNodepools:       pq.StringArray(workload.RequestedNodePools),
		SupportDelete:            workload.ActionsSupport.Delete,
		SupportSuspend:           workload.ActionsSupport.Suspend,
		Images:                   pq.StringArray(workload.Images),
		DistributedFramework:     workload.DistributedFramework,
		Urls:                     make([]string, 0),
		NetworkStatus:            networkStatus,
		Conditions:               convutils.ConvertSliceTypes(workload.Conditions, conditionApiToModel),
		AdditionalFields:         *additionalFields,
		FullySaasWorkload:        workload.FullySaasWorkload,
		Command:                  workload.Command,
		Arguments:                workload.Arguments,
		EnvironmentVariables:     workload.EnvironmentVariables,
		SchedulingConditions:     schedulingConditions,
		PhaseReason:              phaseReason,
		GuaranteedRuntimeEndsAt:  workload.GuaranteedRuntimeEndsAt,
		// temp
		RequestedGPULimit:         nil,
		RequestedGPURequest:       nil,
		RequestedGPUMemoryLimit:   nil,
		RequestedGPUMemoryRequest: nil,
		RequestedCPULimit:         nil,
		RequestedCPURequest:       nil,
		RequestedCPUMemoryLimit:   nil,
		RequestedCPUMemoryRequest: nil,
		RequestedMigDevice:        nil,
		ApplicationId:             workload.ApplicationId,
		ApplicationName:           workload.ApplicationName,
		K8sObjectGroup:            strPointerIfNotEmpty(workload.Group),
		K8sObjectVersion:          strPointerIfNotEmpty(workload.Version),
	}

	return updateOptionalParams(ctx, tenantId, workload, model, workloadType, workloadChildren, workloadElements, id, clusterId)
}

func updateOptionalParams(ctx context.Context, tenantId uint, workload api.WorkloadSyncFields, model models.Workload,
	workloadType common.WorkloadType, workloadChildren []models.WorkloadChild, workloadElements []models.WorkloadElement,
	id uuid.UUID, clusterId uuid.UUID,
) (models.Workload, []models.WorkloadChild, []models.WorkloadElement, error) {
	if workload.Urls != nil {
		model.Urls = workload.Urls
	}
	// is external workload
	if !workloadType.IsNative() {
		connections, err := createConnectionsFromUrls(workload.Urls, workload.Type)
		if err == nil {
			model.ExternalConnections = connections
		}
	}

	// CreatedAt is one time set in the model, in case the does not exist, set it.
	if workload.CreatedAt.IsZero() {
		model.CreatedAt = time.Now().Truncate(time.Microsecond)
	} else {
		model.CreatedAt = workload.CreatedAt
	}

	category := workload.Category
	if category != "" {
		model.CategoryID = getCategoryIdByName(ctx, tenantId, category)
	}
	if model.CategoryID == nil {
		model.CategoryID, _, _ = GetCategoryIdByWorkloadTypeAndGroup(ctx, tenantId, workloadType, workload.Group)
	}

	if workload.Topology != nil {
		model.Topology = &models.Topology{
			Name:               workload.Topology.Name,
			PreferredPlacement: workload.Topology.PreferredPlacement,
			RequiredPlacement:  workload.Topology.RequiredPlacement,
		}
	}

	if workload.StructureElementsTotalCount != nil {
		count := int(*workload.StructureElementsTotalCount)
		model.StructureElementsTotalCount = &count
	}

	if len(workload.Endpoints) > 0 {
		endpoints := make(models.EndpointArray, len(workload.Endpoints))
		for i, ep := range workload.Endpoints {
			endpoints[i] = models.Endpoint{
				Host:       ep.Host,
				Port:       ep.Port,
				Path:       ep.Path,
				TlsEnabled: ep.TlsEnabled,
				Protocol:   ep.Protocol,
				Url:        ep.Url,
				SourceRef: models.EndpointSourceRef{
					Kind: ep.SourceRef.Kind,
					Name: ep.SourceRef.Name,
				},
			}
		}
		model.Endpoints = endpoints
	}

	if len(workload.NodeAffinities) > 0 {
		model.NodeAffinities = convutils.ConvertSliceTypes(workload.NodeAffinities, workloadNodeAffinityApiToModel)
	}

	if len(workload.ResourceRefs) > 0 {
		model.ResourceRefs = convutils.ConvertSliceTypes(workload.ResourceRefs, workloadResourceRefApiToModel)
	}

	err := setWorkloadResources(workload, &model)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Got an error while parsing <%s> requested resources, err: %v, requested resource: %+v",
			workload.Name, err, workload.RequestedResources)
		return models.Workload{}, workloadChildren, workloadElements, err
	}

	for _, child := range workload.ChildrenIds {
		childId, err := uuid.Parse(child.Id)
		if err != nil {
			return models.Workload{}, workloadChildren, workloadElements, err
		}

		workloadChildren = append(workloadChildren, models.WorkloadChild{
			WorkloadID: id,
			ChildType:  child.Type,
			ChildID:    childId,
		})
	}

	workloadChildren = append(workloadChildren, models.WorkloadChild{
		WorkloadID: id,
		ChildType:  workload.Type,
		ChildID:    id,
	})

	if workload.StructureElements != nil && len(*workload.StructureElements) > 0 {
		workloadElements = SyncWorkloadElementsToModels(id, tenantId, clusterId, *workload.StructureElements)
	}

	return model, workloadChildren, workloadElements, nil
}

func GetCategoryIdByWorkloadTypeAndGroup(ctx context.Context, tenantId uint, workloadType common.WorkloadType, workloadGroup string) (*uuid.UUID, string, *api.Error) {
	var categoryId *uuid.UUID
	var categoryName string

	if workloadType.IsNative() {
		policiesWorkloadTypeConfig, policyErr := GetWorkloadConfigOfTenant(ctx)
		if policyErr == nil && policiesWorkloadTypeConfig != nil {
			workloadKind := workloadType.ToKind()
			conf := (*policiesWorkloadTypeConfig)[workloadKind]
			if conf != nil && conf.CategoryName != nil {
				categoryName = *conf.CategoryName
				categoryId = getCategoryIdByName(ctx, tenantId, *conf.CategoryName)
			}
		}
	} else {
		externalWorkloadTypeConfig, err := dao.DbGetWorkloadTypeConfigByNameGroup(ctx, tenantId, workloadType.String(), workloadGroup)
		if err != nil {
			return nil, "", err
		}
		categoryId = ptr.To(externalWorkloadTypeConfig.CategoryId)
	}
	return categoryId, categoryName, nil
}

func getCategoryIdByName(ctx context.Context, tenantId uint, categoryName string) *uuid.UUID {
	if categoryName == "" {
		return nil
	}
	// to improve performance - the map of ids by names of categories should be cached
	categoryModel, categoryErr := dao.DbGetCategoryByName(ctx, tenantId, categoryName)
	if categoryErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).
			Errorf("Got an error while fetching category <%s>, requested resource: %v", categoryName, categoryErr)
		return nil
	}
	if categoryModel == nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).
			Errorf("Did not find category with name <%s>, for tenant: %d", categoryName, tenantId)
		return nil
	}
	categoryId, err := uuid.Parse(categoryModel.Id)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).
			Errorf("Failed to parse category ID %s, err: %v", categoryModel.Id, err)
		return nil
	}
	return &categoryId
}

func conditionApiToModel(condition api.Condition) models.Condition {
	return models.Condition{
		Type:               condition.Type,
		Status:             condition.Status,
		Message:            condition.Message,
		Reason:             condition.Reason,
		LastTransitionTime: condition.LastTransitionTime,
	}
}

func conditionModelToApi(condition models.Condition) api.Condition {
	return api.Condition{
		Type:               condition.Type,
		Status:             condition.Status,
		Message:            condition.Message,
		Reason:             condition.Reason,
		LastTransitionTime: condition.LastTransitionTime,
	}
}

func parseFromToUuids(fromStr string, toStr string) (*uuid.UUID, *uuid.UUID, *api.Error) {
	var from *uuid.UUID = nil
	if fromStr != "" {
		fromField, err := uuid.Parse(fromStr)
		if err != nil {
			return nil, nil, api.NewError(http.StatusBadRequest, fmt.Sprintf("From field must be a valid uuid. <%s>", fromStr), err)
		}
		from = &fromField
	}

	var to *uuid.UUID = nil
	if toStr != "" {
		toField, err := uuid.Parse(toStr)
		if err != nil {
			return nil, nil, api.NewError(http.StatusBadRequest, fmt.Sprintf("To field must be a valid uuid. <%s>", toStr), err)
		}
		to = &toField
	}

	return from, to, nil
}

func CreationRequestToPodModel(ctx context.Context, tenantId uint, projectId string, pod api.PodSyncFields) (models.Pod, []models.Container, []models.PodElement, models.WorkloadChild, *models.RevisionChild, error) {
	if pod.ProjectId != "" && projectId != pod.ProjectId {
		return models.Pod{}, nil, nil, models.WorkloadChild{}, nil, fmt.Errorf("pod <%s> is in project <%s> but request is for project <%s>", pod.Name, pod.ProjectId, projectId)
	}

	id, err := uuid.Parse(pod.Id)
	if err != nil {
		return models.Pod{}, nil, nil, models.WorkloadChild{}, nil, err
	}

	workloadId, err := uuid.Parse(pod.WorkloadId)
	if err != nil {
		return models.Pod{}, nil, nil, models.WorkloadChild{}, nil, err
	}
	if workloadId == uuid.Nil {
		return models.Pod{}, nil, nil, models.WorkloadChild{}, nil, fmt.Errorf("got pod <%s> sync request with empty workload id in cluster <%s> and project <%s>", pod.Name, pod.ClusterId, pod.ProjectId)
	}

	clusterId, err := uuid.Parse(pod.ClusterId)
	if err != nil {
		return models.Pod{}, nil, nil, models.WorkloadChild{}, nil, err
	}

	if pod.AllocatedResources == nil {
		pod.AllocatedResources = &api.AllocatedResources{}
	}

	model := models.Pod{
		ID:                id,
		CreatedAt:         pod.CreatedAt,
		UpdatedAt:         time.Now(),
		CompletedAt:       getTimeOrNil(convutils.Deref(pod.CompletedAt, time.Time{})),
		WorkloadID:        workloadId,
		RevisionID:        pod.RevisionId,
		Name:              pod.Name,
		PriorityClassName: pod.PriorityClassName,
		TenantID:          tenantId,
		ClusterID:         clusterId,
		// In clusters 2.9 we are not syncing the project id of pods and fetching it will have bad performance, so we will put a zero value. Anyway it has a FK to workload which does store the correct project id.
		ProjectID:          convutils.ValueOrDefault(pod.ProjectId, "0"),
		NodeName:           pod.NodeName,
		CurrentNodepool:    pod.CurrentNodePool,
		RequestedNodepools: pq.StringArray(pod.RequestedNodePools),
		AllocatedGPU:       pod.AllocatedResources.Gpu,
		AllocatedGPUMemory: pod.AllocatedResources.GpuMemory,
		AllocatedCPU:       pod.AllocatedResources.Cpu,
		AllocatedCPUMemory: pod.AllocatedResources.CpuMemory,
		Tolerations:        convutils.ConvertSliceTypes(pod.Tolerations, tolerationApiToModel),
		Phase:              pod.K8sPhase,
		Ip:                 pod.Ip,
		Task:               pod.Task,
		NodeAffinity:       NodeAffinityApiToModel(pod.NodeAffinity),
		// DRA usage reported as-is by cluster-sync. Managed/ownership
		// resolution against the k8s-object store happens at API read time,
		// not here — this ingest path performs no cross-service lookup.
		ResourceClaims: podResourceClaimsApiToModel(pod.ResourceClaims),
		// temp
		RequestedGPULimit:         nil,
		RequestedGPURequest:       nil,
		RequestedGPUMemoryLimit:   nil,
		RequestedGPUMemoryRequest: nil,
		RequestedCPULimit:         nil,
		RequestedCPURequest:       nil,
		RequestedCPUMemoryLimit:   nil,
		RequestedCPUMemoryRequest: nil,
		RequestedMigDevice:        nil,
	}

	if pod.NodeName != nil && *pod.NodeName != "" {
		clusterCapabilityChecker := checker.NewClusterCapabilitiesChecker(clusterId.String())
		topologyAddressClusterSupport, _ := clusterCapabilityChecker.ValidateClusterCapability(ctx, checker.MinVersionTopologyAddress)

		if topologyAddressClusterSupport {
			networkTopologyLabelValues, err := clusterservice.Get().GetNodeNetworkTopologyLabelValues(ctx, pod.ClusterId, *pod.NodeName)
			if err != nil {
				// Log the error but don't fail the pod sync
				logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to get topology labels for projectId <%s> pod <%s> on node <%s> in cluster <%s>. err: %v", projectId, pod.Name, *pod.NodeName, pod.ClusterId, err)
			} else if networkTopologyLabelValues != nil {
				labels := pq.StringArray(networkTopologyLabelValues)
				model.NetworkTopologyLabelValues = &labels
			}
		}
	}

	model, err = setPodResources(pod.RequestedResources, model)
	if err != nil {
		return models.Pod{}, nil, nil, models.WorkloadChild{}, nil, err
	}

	containers := []models.Container{}
	for _, container := range pod.Containers {
		containers = append(containers, models.Container{
			StartedAt: container.StartedAt,
			ClusterID: clusterId,
			PodID:     id,
			Name:      container.Name,
			Image:     container.Image,
		})
	}

	var podElements []models.PodElement
	if pod.Elements != nil {
		for _, elem := range *pod.Elements {
			podElements = append(podElements, models.PodElement{
				PodID:     id,
				ElementID: elem.ElementId,
				Level:     int(elem.HierarchyLevel),
			})
		}
		sort.Slice(podElements, func(i, j int) bool {
			return podElements[i].Level < podElements[j].Level
		})
	}

	workloadChild := models.WorkloadChild{
		WorkloadID: workloadId,
		ChildID:    id,
		ChildType:  "Pod",
	}

	revisionChild, err := getRevisionChild(pod, id)
	if err != nil {
		// we don't return an error in case of failure, in order to let the pod sync continue, just without the revision child
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Got an error while creating revision child for pod <%s>, err: %v", pod.Name, err)
	}

	return model, containers, podElements, workloadChild, revisionChild, nil
}

func getRevisionChild(pod api.PodSyncFields, podId uuid.UUID) (*models.RevisionChild, error) {
	if pod.RevisionId == nil || *pod.RevisionId == "" {
		return nil, nil
	}

	revisionId, err := uuid.Parse(*pod.RevisionId)
	if err != nil {
		return nil, err
	}

	return &models.RevisionChild{
		RevisionID: revisionId,
		ChildID:    podId,
		ChildType:  "Pod",
	}, nil
}

func CreationRequestToRevisionModel(ctx context.Context, tenantId uint, projectId string, revision api.RevisionSyncFields) (models.Revision, []models.RevisionChild, error) {
	revisionChildren := []models.RevisionChild{}

	if projectId != revision.ProjectId {
		return models.Revision{}, revisionChildren, fmt.Errorf("revision <%s> is in project <%s> but request is for project <%s>", revision.Name, revision.ProjectId, projectId)
	}

	id, err := uuid.Parse(revision.Id)
	if err != nil {
		return models.Revision{}, revisionChildren, err
	}

	workloadId, err := uuid.Parse(revision.WorkloadId)
	if err != nil {
		return models.Revision{}, revisionChildren, err
	}

	clusterId, err := uuid.Parse(revision.ClusterId)
	if err != nil {
		return models.Revision{}, revisionChildren, err
	}

	additionalFields := &models.AdditionalFields{}
	if revision.AdditionalFields != "" {
		err = additionalFields.Scan([]byte(revision.AdditionalFields))
		if err != nil {
			return models.Revision{}, revisionChildren, err
		}
	}

	model := models.Revision{
		ID:                       id,
		CreatedAt:                revision.CreatedAt,
		Name:                     revision.Name,
		Type:                     revision.Type,
		WorkloadID:               workloadId,
		TenantID:                 tenantId,
		ClusterID:                clusterId,
		DepartmentID:             revision.DepartmentId,
		ProjectID:                revision.ProjectId,
		RequestedPodsNumber:      revision.RequestedPods.Number,
		RequestedPodsMin:         revision.RequestedPods.Min,
		RequestedPodsMax:         revision.RequestedPods.Max,
		RequestedPodsParallelism: revision.RequestedPods.Parallelism,
		RequestedPodsCompletions: revision.RequestedPods.Completions,
		RequestedNodepools:       pq.StringArray(revision.RequestedNodePools),
		Images:                   pq.StringArray(revision.Images),
		Command:                  revision.Command,
		Arguments:                revision.Arguments,
		EnvironmentVariables:     revision.EnvironmentVariables,
		Conditions:               convutils.ConvertSliceTypes(revision.Conditions, conditionApiToModel),
		Phase:                    revision.Phase,
		PhaseMessage:             revision.PhaseMessage,
		PhaseUpdatedAt:           revision.PhaseUpdatedAt,
		AdditionalFields:         *additionalFields,
		// temp
		RequestedGPULimit:         nil,
		RequestedGPURequest:       nil,
		RequestedGPUMemoryLimit:   nil,
		RequestedGPUMemoryRequest: nil,
		RequestedCPULimit:         nil,
		RequestedCPURequest:       nil,
		RequestedCPUMemoryLimit:   nil,
		RequestedCPUMemoryRequest: nil,
		RequestedMigDevice:        nil,
	}

	err = setRevisionResources(revision, &model)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Got an error while parsing <%s> requested resources, err: %v, requested resource: %+v", revision.Name, err, revision.RequestedResources)
		return models.Revision{}, revisionChildren, err
	}

	for _, child := range revision.ChildrenIds {
		childId, err := uuid.Parse(child.Id)
		if err != nil {
			return models.Revision{}, revisionChildren, err
		}

		revisionChildren = append(revisionChildren, models.RevisionChild{
			RevisionID: id,
			ChildType:  child.Type,
			ChildID:    childId,
		})
	}

	revisionChildren = append(revisionChildren, models.RevisionChild{
		RevisionID: id,
		ChildType:  revision.Type,
		ChildID:    id,
	})

	return model, revisionChildren, nil
}

func tolerationApiToModel(toleration api.PodToleration) models.Toleration {
	return models.Toleration{
		Key:               toleration.Key,
		Operator:          toleration.Operator,
		ValueField:        toleration.Value,
		Effect:            toleration.Effect,
		TolerationSeconds: toleration.TolerationSeconds,
	}
}

// podResourceClaimsApiToModel is a straight pass-through from the API wire
// shape into the jsonb-backed model slice. Nil pointer → nil model slice;
// empty-but-non-nil → empty model slice, matching the Toleration pattern
// (no empty→nil collapse).
func podResourceClaimsApiToModel(claims *[]api.PodResourceClaim) models.PodResourceClaimArray {
	if claims == nil {
		return nil
	}
	return convutils.ConvertSliceTypes(*claims, podResourceClaimApiToModel)
}

func podResourceClaimApiToModel(claim api.PodResourceClaim) models.PodResourceClaim {
	return models.PodResourceClaim{
		Name:                       claim.Name,
		ResourceClaimTemplateName:  claim.ResourceClaimTemplateName,
		ResourceClaimName:          claim.ResourceClaimName,
		AllocatedResourceClaimName: claim.AllocatedResourceClaimName,
	}
}

func PodModelToApi(pod models.Pod, projectId string) api.Pod {
	apiPod := api.Pod{
		Name:               pod.Name,
		PriorityClassName:  pod.PriorityClassName,
		Id:                 pod.ID.String(),
		WorkloadId:         pod.WorkloadID.String(),
		RevisionId:         getStringOrNil(pod.RevisionID),
		ClusterId:          pod.ClusterID.String(),
		ProjectId:          projectId,
		NodeName:           pod.NodeName,
		CreatedAt:          pod.CreatedAt,
		CompletedAt:        getTimeOrNil(convutils.Deref(pod.CompletedAt, time.Time{})),
		Containers:         convutils.ConvertSliceTypes(pod.Containers, ContainerModelToApi),
		Tolerations:        convutils.ConvertSliceTypes(pod.Tolerations, TolerationModelToApi),
		CurrentNodePool:    pod.CurrentNodepool,
		RequestedNodePools: pod.RequestedNodepools,
		RequestedResources: getPodRequestedResources(
			pod.RequestedGPULimit,
			pod.RequestedGPURequest,
			pod.RequestedGPUMemoryLimit,
			pod.RequestedGPUMemoryRequest,
			pod.RequestedCPULimit,
			pod.RequestedCPURequest,
			pod.RequestedCPUMemoryLimit,
			pod.RequestedCPUMemoryRequest,
			pod.RequestedMigDevice,
		),
		AllocatedResources: &api.AllocatedResources{
			Gpu:       pod.AllocatedGPU,
			GpuMemory: pod.AllocatedGPUMemory,
			Cpu:       pod.AllocatedCPU,
			CpuMemory: pod.AllocatedCPUMemory,
		},
		K8sPhase:          pod.Phase,
		TenantId:          int32(pod.TenantID),
		K8sPhaseUpdatedAt: pod.PhaseUpdatedAt,
		UpdatedAt:         pod.UpdatedAt,
		DeletedAt:         dbmodels.GormDeletedAtToNullableTime(pod.DeletedAt),
		Ip:                pod.Ip,
		Task:              pod.Task,
		NodeAffinity:      NodeAffinityModelToApi(pod.NodeAffinity),
		Elements:          podElementsModelToApi(pod.Elements),
		ResourceClaims:    podResourceClaimsModelToApi(pod.ResourceClaims),
	}

	return apiPod
}

func TolerationModelToApi(toleration models.Toleration) api.PodToleration {
	return api.PodToleration{
		Key:               toleration.Key,
		Operator:          toleration.Operator,
		Value:             toleration.ValueField,
		Effect:            toleration.Effect,
		TolerationSeconds: toleration.TolerationSeconds,
	}
}

func podElementsModelToApi(elements []*models.PodElement) *[]api.PodElement {
	if len(elements) == 0 {
		return nil
	}
	result := make([]api.PodElement, 0, len(elements))
	for _, e := range elements {
		result = append(result, api.PodElement{
			ElementId:      e.ElementID,
			HierarchyLevel: int32(e.Level),
		})
	}
	return &result
}

// podResourceClaimsModelToApi mirrors podResourceClaimsApiToModel for the
// read side. nil model → nil API (field omitted); the API's *[]-shape is a
// nullable schema, so a pod with no DRA usage omits the field entirely rather
// than serializing as an empty array.
func podResourceClaimsModelToApi(claims models.PodResourceClaimArray) *[]api.PodResourceClaim {
	if claims == nil {
		return nil
	}
	result := convutils.ConvertSliceTypes(claims, podResourceClaimModelToApi)
	return &result
}

func podResourceClaimModelToApi(claim models.PodResourceClaim) api.PodResourceClaim {
	return api.PodResourceClaim{
		Name:                       claim.Name,
		ResourceClaimTemplateName:  claim.ResourceClaimTemplateName,
		ResourceClaimName:          claim.ResourceClaimName,
		AllocatedResourceClaimName: claim.AllocatedResourceClaimName,
	}
}

func ContainerModelToApi(container *models.Container) api.Container {
	if container == nil {
		return api.Container{}
	}

	return api.Container{
		Name:      container.Name,
		Image:     container.Image,
		StartedAt: container.StartedAt,
	}
}

func getTimeOrNil(value time.Time) *time.Time {
	if api.IsZeroValue(value) {
		return nil
	}
	return &value
}

func setDefaultLimit(limit int32) int32 {
	if limit == 0 {
		limit = 50
	}
	limit = int32(math.Min(float64(limit), 500))
	limit = int32(math.Max(float64(limit), 1))

	return limit
}

func GetTelemetryCsv(telemetryGroupsApi *api.TelemetryResponse) [][]string {
	var csvData [][]string
	numberOfValues := len(telemetryGroupsApi.Values)
	if numberOfValues == 0 {
		return csvData
	}
	titles := []string{"value"}
	for _, group := range telemetryGroupsApi.Values[0].Groups {
		groupKey := group.Key
		titles = append(titles,
			fmt.Sprintf("%s.value", groupKey),
			fmt.Sprintf("%s.name", groupKey))
	}
	csvData = append(csvData, titles)
	for _, telemetryValue := range telemetryGroupsApi.Values {
		values := []string{telemetryValue.Value}
		for _, group := range telemetryValue.Groups {
			values = append(values, group.Value, group.Name)
		}
		csvData = append(csvData, values)
	}
	return csvData
}

func getTimeMeasurementsMap(response *api.MetricsResponse) map[time.Time]map[string]string {
	timestampMap := make(map[time.Time]map[string]string)

	for _, m := range response.Measurements {
		if m.Values != nil {
			for _, v := range *m.Values {
				if _, ok := timestampMap[*v.Timestamp]; !ok {
					timestampMap[*v.Timestamp] = make(map[string]string)
				}
				timestampMap[*v.Timestamp][m.Type] = v.Value
			}
		}
	}

	return timestampMap
}

func getOrderedTimestamps(timeToMeasurements map[time.Time]map[string]string) []time.Time {
	var orderedTimestamps []time.Time
	for k := range timeToMeasurements {
		orderedTimestamps = append(orderedTimestamps, k)
	}

	sort.Slice(orderedTimestamps, func(i, j int) bool {
		return orderedTimestamps[i].Before(orderedTimestamps[j])
	})

	return orderedTimestamps
}

func GetMetricsCsv(response *api.MetricsResponse, metricsType []string) [][]string {
	var csvData [][]string

	timeToMeasurements := getTimeMeasurementsMap(response)
	orderedTimestamps := getOrderedTimestamps(timeToMeasurements)

	titles := []string{"Timestamp"}
	titles = append(titles, metricsType...)

	csvData = append(csvData, titles)

	for _, timestamp := range orderedTimestamps {
		row := []string{}
		row = append(row, timestamp.String())
		timestampVal := timeToMeasurements[timestamp]
		for _, metricType := range metricsType {
			metricValue := timestampVal[metricType]
			row = append(row, metricValue)
		}
		csvData = append(csvData, row)
	}

	return csvData
}

// GetMetricsCsvComposite builds CSV including both measurements and histogram series.
// Column order follows metricsType; histogram metrics expand into multiple columns per bin/percentile.
//
// Example (NIM):
//   - Inputs:
//     metricsType = [
//     "NIM_TIME_TO_FIRST_TOKEN_SECONDS",
//     "NIM_TIME_TO_FIRST_TOKEN_SECONDS_PERCENTILES",
//     ]
//     Histogram buckets at t=1711058503000: {le=0.1:1, le=+Inf:5}
//     Percentiles at t=1711158503000: {p50:0.08, p90:0.20}
//   - Output CSV header:
//     [
//     "Timestamp",
//     "NIM_TIME_TO_FIRST_TOKEN_SECONDS[0.1]",
//     "NIM_TIME_TO_FIRST_TOKEN_SECONDS[+Inf]",
//     "NIM_TIME_TO_FIRST_TOKEN_SECONDS_PERCENTILES[p50]",
//     "NIM_TIME_TO_FIRST_TOKEN_SECONDS_PERCENTILES[p90]",
//     ]
//   - Example rows:
//     ["2024-03-22 10:01:43 +0000 UTC","1","5","","" ]   // bucket timestamp values
//     ["2024-03-22 10:15:03 +0000 UTC","","","0.08","0.20"] // percentile timestamp values
//
// Notes:
// - Bucket keys are sorted numerically ascending with "+Inf" last.
// - Percentile keys use pXX format (e.g., p50, p90, p99).
// - Missing values for a column at a given timestamp are left empty.
func GetMetricsCsvComposite(response *api.MetricsCompositeResponse, metricsType []string) [][]string {
	// Build a unified timestamp->columnName->value map
	timeToValues := map[time.Time]map[string]string{}
	// measurements
	for _, m := range response.Measurements {
		if m.Values == nil {
			continue
		}
		for _, v := range *m.Values {
			if _, ok := timeToValues[*v.Timestamp]; !ok {
				timeToValues[*v.Timestamp] = map[string]string{}
			}
			timeToValues[*v.Timestamp][m.Type] = v.Value
		}
	}
	// histogram
	histogramKeysByType := map[string][]string{}
	if response.Histogram != nil {
		tmpSets := map[string]map[string]struct{}{}
		for _, series := range *response.Histogram {
			if _, ok := tmpSets[series.Type]; !ok {
				tmpSets[series.Type] = map[string]struct{}{}
			}
			for _, v := range series.Values {
				if _, ok := timeToValues[*v.Timestamp]; !ok {
					timeToValues[*v.Timestamp] = map[string]string{}
				}
				for k, val := range v.Data {
					col := series.Type + "[" + k + "]"
					timeToValues[*v.Timestamp][col] = val
					tmpSets[series.Type][k] = struct{}{}
				}
			}
		}
		for t, set := range tmpSets {
			keys := make([]string, 0, len(set))
			for k := range set {
				keys = append(keys, k)
			}
			sort.Slice(keys, func(i, j int) bool {
				if keys[i] == HistogramPlusInfBucketKey {
					return false
				}
				if keys[j] == HistogramPlusInfBucketKey {
					return true
				}
				fi, errI := strconv.ParseFloat(keys[i], 64)
				fj, errJ := strconv.ParseFloat(keys[j], 64)
				if errI == nil && errJ == nil {
					return fi < fj
				}
				return keys[i] < keys[j]
			})
			histogramKeysByType[t] = keys
		}
	}

	// order timestamps
	orderedTimestamps := make([]time.Time, 0, len(timeToValues))
	for ts := range timeToValues {
		orderedTimestamps = append(orderedTimestamps, ts)
	}
	sort.Slice(orderedTimestamps, func(i, j int) bool { return orderedTimestamps[i].Before(orderedTimestamps[j]) })

	// Build header
	titles := []string{"Timestamp"}
	expandedColumns := []string{}
	for _, t := range metricsType {
		if keys, ok := histogramKeysByType[t]; ok {
			for _, k := range keys {
				col := t + "[" + k + "]"
				titles = append(titles, col)
				expandedColumns = append(expandedColumns, col)
			}
		} else {
			titles = append(titles, t)
			expandedColumns = append(expandedColumns, t)
		}
	}

	// Build rows
	var csvData [][]string
	csvData = append(csvData, titles)
	for _, ts := range orderedTimestamps {
		row := []string{ts.String()}
		values := timeToValues[ts]
		for _, col := range expandedColumns {
			row = append(row, values[col])
		}
		csvData = append(csvData, row)
	}
	return csvData
}

func RevisionModelToApi(revision *models.RevisionView) api.Revision {
	apiRevision := api.Revision{
		Type:         revision.Type,
		Name:         revision.Name,
		Id:           revision.ID.String(),
		WorkloadId:   revision.WorkloadID.String(),
		TenantId:     int32(revision.TenantID),
		ClusterId:    revision.ClusterID.String(),
		ProjectId:    revision.ProjectID,
		DepartmentId: revision.DepartmentID,
		CreatedAt:    revision.CreatedAt,
		DeletedAt:    dbmodels.GormDeletedAtToNullableTime(revision.DeletedAt),
		RevisionRequestedResources: getWorkloadRequestedResources(
			revision.RequestedGPULimit,
			revision.RequestedGPURequest,
			revision.RequestedGPUMemoryLimit,
			revision.RequestedGPUMemoryRequest,
			revision.RequestedCPULimit,
			revision.RequestedCPURequest,
			revision.RequestedCPUMemoryLimit,
			revision.RequestedCPUMemoryRequest,
		),
		PodsRequestedResources: getPodsRequestedResources(
			revision.PodsRequestedGPULimit,
			revision.PodsRequestedGPURequest,
			revision.PodsRequestedGPUMemoryLimit,
			revision.PodsRequestedGPUMemoryRequest,
			revision.PodsRequestedCPULimit,
			revision.PodsRequestedCPURequest,
			revision.PodsRequestedCPUMemoryLimit,
			revision.PodsRequestedCPUMemoryRequest,
		),
		RequestedPods: api.RequestedPods{
			Number:      &revision.NumberOfPods,
			Min:         revision.RequestedPodsMin,
			Max:         revision.RequestedPodsMax,
			Parallelism: revision.RequestedPodsParallelism,
			Completions: revision.RequestedPodsCompletions,
		},
		AllocatedResources: &api.WorkloadAllocatedResources{
			Gpu:       revision.AllocatedGPU,
			GpuMemory: sumQuantities(revision.AllocatedGPUMemory),
			Cpu:       revision.AllocatedCPU,
			CpuMemory: sumQuantities(revision.AllocatedCPUMemory),
		},
		RequestedNodePools:   revision.RequestedNodepools,
		CurrentNodePools:     revision.CurrentNodepools,
		RunningPods:          revision.NumberOfRunningPods,
		Images:               revision.Images,
		EnvironmentVariables: revision.EnvironmentVariables,
		Command:              revision.Command,
		Arguments:            revision.Arguments,
		Conditions:           convutils.ConvertSliceTypes(revision.Conditions, conditionModelToApi),
		Phase:                revision.Phase,
		PhaseMessage:         revision.PhaseMessage,
		PhaseUpdatedAt:       revision.PhaseUpdatedAt,
		AdditionalFields:     revision.AdditionalFields,
	}

	return apiRevision
}

func validateRevisionSupported(ctx context.Context, clusterId uuid.UUID) *api.Error {
	clusterCapabilityChecker := checker.NewClusterCapabilitiesChecker(clusterId.String())
	clusterSupport, err := clusterCapabilityChecker.ValidateClusterCapability(ctx, checker.MinVersionRevisions)
	if err != nil {
		return api.NewError(http.StatusServiceUnavailable, err.Error(), nil)
	}

	if !clusterSupport {
		return api.NewError(http.StatusBadRequest, "Revisions APIs not supported in the cluster. Required minimum cluster version: 2.21.0", nil)
	}

	return nil
}

// ActualPlacement is deliberately not stored; it is computed per request.
func syncElementTopologyToModel(topology *api.WorkloadElementTopology) *models.Topology {
	if topology == nil {
		return nil
	}
	if topology.Name == "" && topology.PreferredPlacement == "" && topology.RequiredPlacement == "" {
		return nil
	}
	return &models.Topology{
		Name:               topology.Name,
		PreferredPlacement: topology.PreferredPlacement,
		RequiredPlacement:  topology.RequiredPlacement,
	}
}

func elementTopologyToApi(topology *models.Topology) *api.WorkloadElementTopology {
	if topology == nil {
		return nil
	}
	return &api.WorkloadElementTopology{
		Name:               topology.Name,
		PreferredPlacement: topology.PreferredPlacement,
		RequiredPlacement:  topology.RequiredPlacement,
	}
}

// SyncWorkloadElementsToModels converts API WorkloadElement objects to DB models.
func SyncWorkloadElementsToModels(workloadID uuid.UUID, tenantID uint, clusterID uuid.UUID, elements []api.WorkloadElement) []models.WorkloadElement {
	result := make([]models.WorkloadElement, 0, len(elements))
	for _, e := range elements {
		m := models.WorkloadElement{
			ID:          e.Id,
			WorkloadID:  workloadID,
			TenantID:    tenantID,
			ClusterID:   clusterID,
			Name:        e.Name,
			Level:       int(e.HierarchyLevel),
			ParentID:    e.ParentId,
			InstanceKey: e.InstanceKey,
			ReplicaKey:  e.ReplicaKey,
			Topology:    syncElementTopologyToModel(e.Topology),
		}
		if e.KubernetesRef != nil {
			m.K8sRefUID = &e.KubernetesRef.Uid
			m.K8sRefName = &e.KubernetesRef.Name
			m.GVKGroup = &e.KubernetesRef.Gvk.Group
			m.GVKVersion = &e.KubernetesRef.Gvk.Version
			m.GVKKind = &e.KubernetesRef.Gvk.Kind
		}
		result = append(result, m)
	}
	return result
}

func validateWorkloadElementsSupported(ctx context.Context, clusterId uuid.UUID) *api.Error {
	clusterCapabilityChecker := checker.NewClusterCapabilitiesChecker(clusterId.String())
	clusterSupport, err := clusterCapabilityChecker.ValidateClusterCapability(ctx, checker.MinVersionWorkloadsStructureElements)
	if err != nil {
		return api.NewError(http.StatusServiceUnavailable, err.Error(), nil)
	}

	if !clusterSupport {
		return api.NewError(http.StatusBadRequest, "Workload elements API not supported in the cluster. Required minimum cluster version: 2.25.0", nil)
	}

	return nil
}

// modelsToWorkloadElements converts DB WorkloadElement models to API WorkloadElement objects.
func modelsToWorkloadElements(dbElements []models.WorkloadElement) []api.WorkloadElement {
	result := make([]api.WorkloadElement, 0, len(dbElements))
	for _, m := range dbElements {
		e := api.WorkloadElement{
			Id:             m.ID,
			Name:           m.Name,
			HierarchyLevel: int32(m.Level),
			ParentId:       m.ParentID,
			InstanceKey:    m.InstanceKey,
			ReplicaKey:     m.ReplicaKey,
			CreatedAt:      m.CreatedAt,
			Topology:       elementTopologyToApi(m.Topology),
		}
		if m.K8sRefUID != nil {
			e.KubernetesRef = &api.WorkloadElementKubernetesRef{
				Uid:  *m.K8sRefUID,
				Name: derefOrEmpty(m.K8sRefName),
				Gvk: api.WorkloadElementKubernetesRefGvk{
					Group:   derefOrEmpty(m.GVKGroup),
					Version: derefOrEmpty(m.GVKVersion),
					Kind:    derefOrEmpty(m.GVKKind),
				},
			}
		}
		result = append(result, e)
	}
	return result
}

// modelToWorkloadElementResponse converts a DB WorkloadElement model and its pods into an API WorkloadElementResponse.
func modelToWorkloadElementResponse(m models.WorkloadElement, pods []models.Pod) api.WorkloadElementResponse {
	resp := api.WorkloadElementResponse{
		Id:             m.ID,
		Name:           m.Name,
		HierarchyLevel: int32(m.Level),
		ParentId:       m.ParentID,
		InstanceKey:    m.InstanceKey,
		ReplicaKey:     m.ReplicaKey,
		CreatedAt:      m.CreatedAt,
		Topology:       elementTopologyToApi(m.Topology),
	}
	if m.K8sRefUID != nil {
		resp.KubernetesRef = &api.WorkloadElementKubernetesRef{
			Uid:  *m.K8sRefUID,
			Name: derefOrEmpty(m.K8sRefName),
			Gvk: api.WorkloadElementKubernetesRefGvk{
				Group:   derefOrEmpty(m.GVKGroup),
				Version: derefOrEmpty(m.GVKVersion),
				Kind:    derefOrEmpty(m.GVKKind),
			},
		}
	}
	resp.PodsRequestedResources = aggregatePodsRequestedResources(pods)
	resp.AllocatedResources = aggregatePodsAllocatedResources(pods)
	return resp
}

func derefOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
