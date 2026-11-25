import localVarRequest from 'request';

export * from './ancestors';
export * from './clusterConfigStatus';
export * from './clusterControlPlaneConfig';
export * from './clusterControlPlaneConfigAgent';
export * from './clusterControlPlaneConfigAgentClient';
export * from './clusterControlPlaneConfigMetricsRemoteWrite';
export * from './clusterControlPlaneConfigMetricsRemoteWriteCredentials';
export * from './clusterControlPlaneConfigResearcherService';
export * from './clusterControlPlaneConfigResearcherServiceOauthProxy';
export * from './clusterControlPlaneConfigWorkloads';
export * from './clusterControlPlaneConfigWorkloadsClient';
export * from './clusterCreationRequest';
export * from './clusterDependenciesStatus';
export * from './clusterDependencyStatus';
export * from './clusterDisplayedStatus';
export * from './clusterDisplayedStatusAllOf';
export * from './clusterFullConfig';
export * from './clusterFullConfigAllOf';
export * from './clusterFullConfigData';
export * from './clusterFullConfigDataAllOf';
export * from './clusterHistory';
export * from './clusterInfoSyncRequest';
export * from './clusterInfoSyncRequestDependencies';
export * from './clusterInfoSyncRequestStatus';
export * from './clusterInstallationInfoResponse';
export * from './clusterLivenessSyncRequest';
export * from './clusterNotification';
export * from './clusterReportedConfig';
export * from './clusterReportedCustomCA';
export * from './clusterReportedResourceRequirements';
export * from './clusterReportedStatus';
export * from './clusterReportedStatusConditionsInner';
export * from './clusterReportedStatusOperandsValue';
export * from './clusterReportedStatusPlatform';
export * from './clusterSettingsGetResponse';
export * from './clusterSettingsKeyEnum';
export * from './clusterState';
export * from './clusterSystemConfigData';
export * from './clusterUpdateRequest';
export * from './clusterUserConfig';
export * from './clusterUserConfigAllOf';
export * from './clusterUserConfigData';
export * from './clusterUserConfigDataPartialUpdate';
export * from './clustersCount';
export * from './clustersMinimalData';
export * from './complianceIssues';
export * from './complianceIssuesComplianceIssuesInner';
export * from './condition';
export * from './countNodes200Response';
export * from './countResponse';
export * from './cpuPodsPlacementStrategy';
export * from './defaultNodepoolConfig';
export * from './displayedCluster';
export * from './enableAdvancedGpuMetricsSetting';
export * from './fipsModeEnum';
export * from './gPUResourceOptimization';
export * from './getClusterName200Response';
export * from './getClusterUrl200Response';
export * from './getClustersMinimalData200Response';
export * from './getInternalClusterConfigResponse';
export * from './getNetworkTopologies200Response';
export * from './getNodepools200Response';
export * from './getNodepoolsInternal200Response';
export * from './getNodepoolsMinimal200Response';
export * from './getNodepoolsMinimalInternal200Response';
export * from './getPutPatchClusterConfigResponse';
export * from './getPutPatchClusterConfigResponseAllOf';
export * from './gpuInfo';
export * from './gpuPodsPlacementStrategy';
export * from './groupedMeasurementGroup';
export * from './groupedMeasurementResponse';
export * from './groupedMeasurementResponseValuesInner';
export * from './groupedMetricsResponse';
export * from './httpResponse';
export * from './legacyError';
export * from './legacyErrorInt32';
export * from './legacyNodePool';
export * from './legacyNodePoolAllOf';
export * from './legacyNodePoolCreateRequest';
export * from './legacyNodePoolLabelsRequest';
export * from './legacyNodePoolStatusRequest';
export * from './legacyPlacementStrategy';
export * from './legacyUpdateNodePoolRequest';
export * from './measurementResponse';
export * from './metadata';
export * from './metricsResponse';
export * from './metricsType';
export * from './modelError';
export * from './networkTopology';
export * from './networkTopologyAllOf';
export * from './networkTopologyCreateRequest';
export * from './networkTopologyCreateRequestAllOf';
export * from './networkTopologyGrove';
export * from './networkTopologyGroveCondition';
export * from './networkTopologyLabel';
export * from './networkTopologySyncAction';
export * from './networkTopologySyncFields';
export * from './networkTopologySyncGrove';
export * from './networkTopologySyncGroveCondition';
export * from './networkTopologySyncRequest';
export * from './networkTopologyUpdateFields';
export * from './node';
export * from './nodeAdditionalFields';
export * from './nodeAdditionalReadFields';
export * from './nodeExtended';
export * from './nodeExtendedAllOf';
export * from './nodeFilterSortFields';
export * from './nodeForSync';
export * from './nodeForSyncAllOf';
export * from './nodeInfo';
export * from './nodeInternal';
export * from './nodeMetricType';
export * from './nodeNumaForSync';
export * from './nodePoolSetting';
export * from './nodePoolSettingsGetResponse';
export * from './nodePoolSettingsKeyEnum';
export * from './nodeStatusConditionDetails';
export * from './nodeTaint';
export * from './nodeTelemetryType';
export * from './nodepool';
export * from './nodepoolCalculatedFields';
export * from './nodepoolConditionDetails';
export * from './nodepoolCreateFields';
export * from './nodepoolCreateResponseFields';
export * from './nodepoolDeprecatedFields';
export * from './nodepoolDeprecatedPlacementStrategy';
export * from './nodepoolGPUNetworkAccelerationDetection';
export * from './nodepoolGpuTypeDetails';
export * from './nodepoolMinimal';
export * from './nodepoolMinimalSortFields';
export * from './nodepoolPhase';
export * from './nodepoolSortFilterFields';
export * from './nodepoolSyncFields';
export * from './nodepoolSyncFieldsStatus';
export * from './nodepoolSyncFieldsStatusAllOf';
export * from './nodepoolSyncUpdateFields';
export * from './nodepoolSyncUpdateFieldsPlacementStrategy';
export * from './nodepoolUpdateFields';
export * from './nodes';
export * from './nodesSyncRequest';
export * from './notificationOperation';
export * from './numaInfo';
export * from './numaInfoExtended';
export * from './numaInfoExtendedAllOf';
export * from './numaNodeDetails';
export * from './numaNodesSyncRequest';
export * from './openshiftRouteTerminationPolicy';
export * from './patchClusterConfigRequest';
export * from './placementStrategy';
export * from './putClusterConfigRequest';
export * from './putClusterConfigStatusRequest';
export * from './putClusterConfigStatusRequestStatus';
export * from './schedulingCapabilities';
export * from './schedulingCapabilitiesNullable';
export * from './schedulingConfiguration';
export * from './schedulingConfigurationPlacementStrategy';
export * from './submissionError';
export * from './syncNodepoolsStatus207Response';
export * from './syncNodepoolsStatusRequest';
export * from './telemetryResponse';
export * from './telemetryResponseValuesInner';
export * from './telemetryResponseValuesInnerGroupsInner';
export * from './timeBasedFairShare';
export * from './timeBasedFairShareWindowType';
export * from './topologyManagerPolicy';
export * from './updateClusterSettingByKeyRequest';
export * from './updateNodePoolSettingByKeyRequest';

import * as fs from 'fs';

export interface RequestDetailedFile {
    value: Buffer;
    options?: {
        filename?: string;
        contentType?: string;
    }
}

export type RequestFile = string | Buffer | fs.ReadStream | RequestDetailedFile;


import { Ancestors } from './ancestors';
import { ClusterConfigStatus } from './clusterConfigStatus';
import { ClusterControlPlaneConfig } from './clusterControlPlaneConfig';
import { ClusterControlPlaneConfigAgent } from './clusterControlPlaneConfigAgent';
import { ClusterControlPlaneConfigAgentClient } from './clusterControlPlaneConfigAgentClient';
import { ClusterControlPlaneConfigMetricsRemoteWrite } from './clusterControlPlaneConfigMetricsRemoteWrite';
import { ClusterControlPlaneConfigMetricsRemoteWriteCredentials } from './clusterControlPlaneConfigMetricsRemoteWriteCredentials';
import { ClusterControlPlaneConfigResearcherService } from './clusterControlPlaneConfigResearcherService';
import { ClusterControlPlaneConfigResearcherServiceOauthProxy } from './clusterControlPlaneConfigResearcherServiceOauthProxy';
import { ClusterControlPlaneConfigWorkloads } from './clusterControlPlaneConfigWorkloads';
import { ClusterControlPlaneConfigWorkloadsClient } from './clusterControlPlaneConfigWorkloadsClient';
import { ClusterCreationRequest } from './clusterCreationRequest';
import { ClusterDependenciesStatus } from './clusterDependenciesStatus';
import { ClusterDependencyStatus } from './clusterDependencyStatus';
import { ClusterDisplayedStatus } from './clusterDisplayedStatus';
import { ClusterDisplayedStatusAllOf } from './clusterDisplayedStatusAllOf';
import { ClusterFullConfig } from './clusterFullConfig';
import { ClusterFullConfigAllOf } from './clusterFullConfigAllOf';
import { ClusterFullConfigData } from './clusterFullConfigData';
import { ClusterFullConfigDataAllOf } from './clusterFullConfigDataAllOf';
import { ClusterHistory } from './clusterHistory';
import { ClusterInfoSyncRequest } from './clusterInfoSyncRequest';
import { ClusterInfoSyncRequestDependencies } from './clusterInfoSyncRequestDependencies';
import { ClusterInfoSyncRequestStatus } from './clusterInfoSyncRequestStatus';
import { ClusterInstallationInfoResponse } from './clusterInstallationInfoResponse';
import { ClusterLivenessSyncRequest } from './clusterLivenessSyncRequest';
import { ClusterNotification } from './clusterNotification';
import { ClusterReportedConfig } from './clusterReportedConfig';
import { ClusterReportedCustomCA } from './clusterReportedCustomCA';
import { ClusterReportedResourceRequirements } from './clusterReportedResourceRequirements';
import { ClusterReportedStatus } from './clusterReportedStatus';
import { ClusterReportedStatusConditionsInner } from './clusterReportedStatusConditionsInner';
import { ClusterReportedStatusOperandsValue } from './clusterReportedStatusOperandsValue';
import { ClusterReportedStatusPlatform } from './clusterReportedStatusPlatform';
import { ClusterSettingsGetResponse } from './clusterSettingsGetResponse';
import { ClusterSettingsKeyEnum } from './clusterSettingsKeyEnum';
import { ClusterState } from './clusterState';
import { ClusterSystemConfigData } from './clusterSystemConfigData';
import { ClusterUpdateRequest } from './clusterUpdateRequest';
import { ClusterUserConfig } from './clusterUserConfig';
import { ClusterUserConfigAllOf } from './clusterUserConfigAllOf';
import { ClusterUserConfigData } from './clusterUserConfigData';
import { ClusterUserConfigDataPartialUpdate } from './clusterUserConfigDataPartialUpdate';
import { ClustersCount } from './clustersCount';
import { ClustersMinimalData } from './clustersMinimalData';
import { ComplianceIssues } from './complianceIssues';
import { ComplianceIssuesComplianceIssuesInner } from './complianceIssuesComplianceIssuesInner';
import { Condition } from './condition';
import { CountNodes200Response } from './countNodes200Response';
import { CountResponse } from './countResponse';
import { CpuPodsPlacementStrategy } from './cpuPodsPlacementStrategy';
import { DefaultNodepoolConfig } from './defaultNodepoolConfig';
import { DisplayedCluster } from './displayedCluster';
import { EnableAdvancedGpuMetricsSetting } from './enableAdvancedGpuMetricsSetting';
import { FipsModeEnum } from './fipsModeEnum';
import { GPUResourceOptimization } from './gPUResourceOptimization';
import { GetClusterName200Response } from './getClusterName200Response';
import { GetClusterUrl200Response } from './getClusterUrl200Response';
import { GetClustersMinimalData200Response } from './getClustersMinimalData200Response';
import { GetInternalClusterConfigResponse } from './getInternalClusterConfigResponse';
import { GetNetworkTopologies200Response } from './getNetworkTopologies200Response';
import { GetNodepools200Response } from './getNodepools200Response';
import { GetNodepoolsInternal200Response } from './getNodepoolsInternal200Response';
import { GetNodepoolsMinimal200Response } from './getNodepoolsMinimal200Response';
import { GetNodepoolsMinimalInternal200Response } from './getNodepoolsMinimalInternal200Response';
import { GetPutPatchClusterConfigResponse } from './getPutPatchClusterConfigResponse';
import { GetPutPatchClusterConfigResponseAllOf } from './getPutPatchClusterConfigResponseAllOf';
import { GpuInfo } from './gpuInfo';
import { GpuPodsPlacementStrategy } from './gpuPodsPlacementStrategy';
import { GroupedMeasurementGroup } from './groupedMeasurementGroup';
import { GroupedMeasurementResponse } from './groupedMeasurementResponse';
import { GroupedMeasurementResponseValuesInner } from './groupedMeasurementResponseValuesInner';
import { GroupedMetricsResponse } from './groupedMetricsResponse';
import { HttpResponse } from './httpResponse';
import { LegacyError } from './legacyError';
import { LegacyErrorInt32 } from './legacyErrorInt32';
import { LegacyNodePool } from './legacyNodePool';
import { LegacyNodePoolAllOf } from './legacyNodePoolAllOf';
import { LegacyNodePoolCreateRequest } from './legacyNodePoolCreateRequest';
import { LegacyNodePoolLabelsRequest } from './legacyNodePoolLabelsRequest';
import { LegacyNodePoolStatusRequest } from './legacyNodePoolStatusRequest';
import { LegacyPlacementStrategy } from './legacyPlacementStrategy';
import { LegacyUpdateNodePoolRequest } from './legacyUpdateNodePoolRequest';
import { MeasurementResponse } from './measurementResponse';
import { Metadata } from './metadata';
import { MetricsResponse } from './metricsResponse';
import { MetricsType } from './metricsType';
import { ModelError } from './modelError';
import { NetworkTopology } from './networkTopology';
import { NetworkTopologyAllOf } from './networkTopologyAllOf';
import { NetworkTopologyCreateRequest } from './networkTopologyCreateRequest';
import { NetworkTopologyCreateRequestAllOf } from './networkTopologyCreateRequestAllOf';
import { NetworkTopologyGrove } from './networkTopologyGrove';
import { NetworkTopologyGroveCondition } from './networkTopologyGroveCondition';
import { NetworkTopologyLabel } from './networkTopologyLabel';
import { NetworkTopologySyncAction } from './networkTopologySyncAction';
import { NetworkTopologySyncFields } from './networkTopologySyncFields';
import { NetworkTopologySyncGrove } from './networkTopologySyncGrove';
import { NetworkTopologySyncGroveCondition } from './networkTopologySyncGroveCondition';
import { NetworkTopologySyncRequest } from './networkTopologySyncRequest';
import { NetworkTopologyUpdateFields } from './networkTopologyUpdateFields';
import { Node } from './node';
import { NodeAdditionalFields } from './nodeAdditionalFields';
import { NodeAdditionalReadFields } from './nodeAdditionalReadFields';
import { NodeExtended } from './nodeExtended';
import { NodeExtendedAllOf } from './nodeExtendedAllOf';
import { NodeFilterSortFields } from './nodeFilterSortFields';
import { NodeForSync } from './nodeForSync';
import { NodeForSyncAllOf } from './nodeForSyncAllOf';
import { NodeInfo } from './nodeInfo';
import { NodeInternal } from './nodeInternal';
import { NodeMetricType } from './nodeMetricType';
import { NodeNumaForSync } from './nodeNumaForSync';
import { NodePoolSetting } from './nodePoolSetting';
import { NodePoolSettingsGetResponse } from './nodePoolSettingsGetResponse';
import { NodePoolSettingsKeyEnum } from './nodePoolSettingsKeyEnum';
import { NodeStatusConditionDetails } from './nodeStatusConditionDetails';
import { NodeTaint } from './nodeTaint';
import { NodeTelemetryType } from './nodeTelemetryType';
import { Nodepool } from './nodepool';
import { NodepoolCalculatedFields } from './nodepoolCalculatedFields';
import { NodepoolConditionDetails } from './nodepoolConditionDetails';
import { NodepoolCreateFields } from './nodepoolCreateFields';
import { NodepoolCreateResponseFields } from './nodepoolCreateResponseFields';
import { NodepoolDeprecatedFields } from './nodepoolDeprecatedFields';
import { NodepoolDeprecatedPlacementStrategy } from './nodepoolDeprecatedPlacementStrategy';
import { NodepoolGPUNetworkAccelerationDetection } from './nodepoolGPUNetworkAccelerationDetection';
import { NodepoolGpuTypeDetails } from './nodepoolGpuTypeDetails';
import { NodepoolMinimal } from './nodepoolMinimal';
import { NodepoolMinimalSortFields } from './nodepoolMinimalSortFields';
import { NodepoolPhase } from './nodepoolPhase';
import { NodepoolSortFilterFields } from './nodepoolSortFilterFields';
import { NodepoolSyncFields } from './nodepoolSyncFields';
import { NodepoolSyncFieldsStatus } from './nodepoolSyncFieldsStatus';
import { NodepoolSyncFieldsStatusAllOf } from './nodepoolSyncFieldsStatusAllOf';
import { NodepoolSyncUpdateFields } from './nodepoolSyncUpdateFields';
import { NodepoolSyncUpdateFieldsPlacementStrategy } from './nodepoolSyncUpdateFieldsPlacementStrategy';
import { NodepoolUpdateFields } from './nodepoolUpdateFields';
import { Nodes } from './nodes';
import { NodesSyncRequest } from './nodesSyncRequest';
import { NotificationOperation } from './notificationOperation';
import { NumaInfo } from './numaInfo';
import { NumaInfoExtended } from './numaInfoExtended';
import { NumaInfoExtendedAllOf } from './numaInfoExtendedAllOf';
import { NumaNodeDetails } from './numaNodeDetails';
import { NumaNodesSyncRequest } from './numaNodesSyncRequest';
import { OpenshiftRouteTerminationPolicy } from './openshiftRouteTerminationPolicy';
import { PatchClusterConfigRequest } from './patchClusterConfigRequest';
import { PlacementStrategy } from './placementStrategy';
import { PutClusterConfigRequest } from './putClusterConfigRequest';
import { PutClusterConfigStatusRequest } from './putClusterConfigStatusRequest';
import { PutClusterConfigStatusRequestStatus } from './putClusterConfigStatusRequestStatus';
import { SchedulingCapabilities } from './schedulingCapabilities';
import { SchedulingCapabilitiesNullable } from './schedulingCapabilitiesNullable';
import { SchedulingConfiguration } from './schedulingConfiguration';
import { SchedulingConfigurationPlacementStrategy } from './schedulingConfigurationPlacementStrategy';
import { SubmissionError } from './submissionError';
import { SyncNodepoolsStatus207Response } from './syncNodepoolsStatus207Response';
import { SyncNodepoolsStatusRequest } from './syncNodepoolsStatusRequest';
import { TelemetryResponse } from './telemetryResponse';
import { TelemetryResponseValuesInner } from './telemetryResponseValuesInner';
import { TelemetryResponseValuesInnerGroupsInner } from './telemetryResponseValuesInnerGroupsInner';
import { TimeBasedFairShare } from './timeBasedFairShare';
import { TimeBasedFairShareWindowType } from './timeBasedFairShareWindowType';
import { TopologyManagerPolicy } from './topologyManagerPolicy';
import { UpdateClusterSettingByKeyRequest } from './updateClusterSettingByKeyRequest';
import { UpdateNodePoolSettingByKeyRequest } from './updateNodePoolSettingByKeyRequest';

/* tslint:disable:no-unused-variable */
let primitives = [
                    "string",
                    "boolean",
                    "double",
                    "integer",
                    "long",
                    "float",
                    "number",
                    "any"
                 ];

let enumsMap: {[index: string]: any} = {
        "ClusterConfigStatus.StateEnum": ClusterConfigStatus.StateEnum,
        "ClusterControlPlaneConfig.FipsModeEnum": ClusterControlPlaneConfig.FipsModeEnum,
        "ClusterControlPlaneConfig.InstallationTypeEnum": ClusterControlPlaneConfig.InstallationTypeEnum,
        "ClusterReportedStatusConditionsInner.StatusEnum": ClusterReportedStatusConditionsInner.StatusEnum,
        "ClusterReportedStatusPlatform.TypeEnum": ClusterReportedStatusPlatform.TypeEnum,
        "ClusterSettingsKeyEnum": ClusterSettingsKeyEnum,
        "ClusterState": ClusterState,
        "Condition.StatusEnum": Condition.StatusEnum,
        "FipsModeEnum": FipsModeEnum,
        "LegacyNodePool.StatusEnum": LegacyNodePool.StatusEnum,
        "LegacyNodePoolAllOf.StatusEnum": LegacyNodePoolAllOf.StatusEnum,
        "LegacyNodePoolStatusRequest.StatusEnum": LegacyNodePoolStatusRequest.StatusEnum,
        "LegacyPlacementStrategy.CpuEnum": LegacyPlacementStrategy.CpuEnum,
        "LegacyPlacementStrategy.GpuEnum": LegacyPlacementStrategy.GpuEnum,
        "MetricsType": MetricsType,
        "NetworkTopologySyncAction": NetworkTopologySyncAction,
        "Node.StatusEnum": Node.StatusEnum,
        "NodeExtended.StatusEnum": NodeExtended.StatusEnum,
        "NodeFilterSortFields": NodeFilterSortFields,
        "NodeInfo.StatusEnum": NodeInfo.StatusEnum,
        "NodeMetricType": NodeMetricType,
        "NodePoolSettingsKeyEnum": NodePoolSettingsKeyEnum,
        "NodeTaint.EffectEnum": NodeTaint.EffectEnum,
        "NodeTelemetryType": NodeTelemetryType,
        "NodepoolConditionDetails.StatusEnum": NodepoolConditionDetails.StatusEnum,
        "NodepoolGPUNetworkAccelerationDetection": NodepoolGPUNetworkAccelerationDetection,
        "NodepoolMinimalSortFields": NodepoolMinimalSortFields,
        "NodepoolPhase": NodepoolPhase,
        "NodepoolSortFilterFields": NodepoolSortFilterFields,
        "NotificationOperation": NotificationOperation,
        "OpenshiftRouteTerminationPolicy": OpenshiftRouteTerminationPolicy,
        "PlacementStrategy": PlacementStrategy,
        "TimeBasedFairShareWindowType": TimeBasedFairShareWindowType,
        "TopologyManagerPolicy": TopologyManagerPolicy,
}

let typeMap: {[index: string]: any} = {
    "Ancestors": Ancestors,
    "ClusterConfigStatus": ClusterConfigStatus,
    "ClusterControlPlaneConfig": ClusterControlPlaneConfig,
    "ClusterControlPlaneConfigAgent": ClusterControlPlaneConfigAgent,
    "ClusterControlPlaneConfigAgentClient": ClusterControlPlaneConfigAgentClient,
    "ClusterControlPlaneConfigMetricsRemoteWrite": ClusterControlPlaneConfigMetricsRemoteWrite,
    "ClusterControlPlaneConfigMetricsRemoteWriteCredentials": ClusterControlPlaneConfigMetricsRemoteWriteCredentials,
    "ClusterControlPlaneConfigResearcherService": ClusterControlPlaneConfigResearcherService,
    "ClusterControlPlaneConfigResearcherServiceOauthProxy": ClusterControlPlaneConfigResearcherServiceOauthProxy,
    "ClusterControlPlaneConfigWorkloads": ClusterControlPlaneConfigWorkloads,
    "ClusterControlPlaneConfigWorkloadsClient": ClusterControlPlaneConfigWorkloadsClient,
    "ClusterCreationRequest": ClusterCreationRequest,
    "ClusterDependenciesStatus": ClusterDependenciesStatus,
    "ClusterDependencyStatus": ClusterDependencyStatus,
    "ClusterDisplayedStatus": ClusterDisplayedStatus,
    "ClusterDisplayedStatusAllOf": ClusterDisplayedStatusAllOf,
    "ClusterFullConfig": ClusterFullConfig,
    "ClusterFullConfigAllOf": ClusterFullConfigAllOf,
    "ClusterFullConfigData": ClusterFullConfigData,
    "ClusterFullConfigDataAllOf": ClusterFullConfigDataAllOf,
    "ClusterHistory": ClusterHistory,
    "ClusterInfoSyncRequest": ClusterInfoSyncRequest,
    "ClusterInfoSyncRequestDependencies": ClusterInfoSyncRequestDependencies,
    "ClusterInfoSyncRequestStatus": ClusterInfoSyncRequestStatus,
    "ClusterInstallationInfoResponse": ClusterInstallationInfoResponse,
    "ClusterLivenessSyncRequest": ClusterLivenessSyncRequest,
    "ClusterNotification": ClusterNotification,
    "ClusterReportedConfig": ClusterReportedConfig,
    "ClusterReportedCustomCA": ClusterReportedCustomCA,
    "ClusterReportedResourceRequirements": ClusterReportedResourceRequirements,
    "ClusterReportedStatus": ClusterReportedStatus,
    "ClusterReportedStatusConditionsInner": ClusterReportedStatusConditionsInner,
    "ClusterReportedStatusOperandsValue": ClusterReportedStatusOperandsValue,
    "ClusterReportedStatusPlatform": ClusterReportedStatusPlatform,
    "ClusterSettingsGetResponse": ClusterSettingsGetResponse,
    "ClusterSystemConfigData": ClusterSystemConfigData,
    "ClusterUpdateRequest": ClusterUpdateRequest,
    "ClusterUserConfig": ClusterUserConfig,
    "ClusterUserConfigAllOf": ClusterUserConfigAllOf,
    "ClusterUserConfigData": ClusterUserConfigData,
    "ClusterUserConfigDataPartialUpdate": ClusterUserConfigDataPartialUpdate,
    "ClustersCount": ClustersCount,
    "ClustersMinimalData": ClustersMinimalData,
    "ComplianceIssues": ComplianceIssues,
    "ComplianceIssuesComplianceIssuesInner": ComplianceIssuesComplianceIssuesInner,
    "Condition": Condition,
    "CountNodes200Response": CountNodes200Response,
    "CountResponse": CountResponse,
    "CpuPodsPlacementStrategy": CpuPodsPlacementStrategy,
    "DefaultNodepoolConfig": DefaultNodepoolConfig,
    "DisplayedCluster": DisplayedCluster,
    "EnableAdvancedGpuMetricsSetting": EnableAdvancedGpuMetricsSetting,
    "GPUResourceOptimization": GPUResourceOptimization,
    "GetClusterName200Response": GetClusterName200Response,
    "GetClusterUrl200Response": GetClusterUrl200Response,
    "GetClustersMinimalData200Response": GetClustersMinimalData200Response,
    "GetInternalClusterConfigResponse": GetInternalClusterConfigResponse,
    "GetNetworkTopologies200Response": GetNetworkTopologies200Response,
    "GetNodepools200Response": GetNodepools200Response,
    "GetNodepoolsInternal200Response": GetNodepoolsInternal200Response,
    "GetNodepoolsMinimal200Response": GetNodepoolsMinimal200Response,
    "GetNodepoolsMinimalInternal200Response": GetNodepoolsMinimalInternal200Response,
    "GetPutPatchClusterConfigResponse": GetPutPatchClusterConfigResponse,
    "GetPutPatchClusterConfigResponseAllOf": GetPutPatchClusterConfigResponseAllOf,
    "GpuInfo": GpuInfo,
    "GpuPodsPlacementStrategy": GpuPodsPlacementStrategy,
    "GroupedMeasurementGroup": GroupedMeasurementGroup,
    "GroupedMeasurementResponse": GroupedMeasurementResponse,
    "GroupedMeasurementResponseValuesInner": GroupedMeasurementResponseValuesInner,
    "GroupedMetricsResponse": GroupedMetricsResponse,
    "HttpResponse": HttpResponse,
    "LegacyError": LegacyError,
    "LegacyErrorInt32": LegacyErrorInt32,
    "LegacyNodePool": LegacyNodePool,
    "LegacyNodePoolAllOf": LegacyNodePoolAllOf,
    "LegacyNodePoolCreateRequest": LegacyNodePoolCreateRequest,
    "LegacyNodePoolLabelsRequest": LegacyNodePoolLabelsRequest,
    "LegacyNodePoolStatusRequest": LegacyNodePoolStatusRequest,
    "LegacyPlacementStrategy": LegacyPlacementStrategy,
    "LegacyUpdateNodePoolRequest": LegacyUpdateNodePoolRequest,
    "MeasurementResponse": MeasurementResponse,
    "Metadata": Metadata,
    "MetricsResponse": MetricsResponse,
    "ModelError": ModelError,
    "NetworkTopology": NetworkTopology,
    "NetworkTopologyAllOf": NetworkTopologyAllOf,
    "NetworkTopologyCreateRequest": NetworkTopologyCreateRequest,
    "NetworkTopologyCreateRequestAllOf": NetworkTopologyCreateRequestAllOf,
    "NetworkTopologyGrove": NetworkTopologyGrove,
    "NetworkTopologyGroveCondition": NetworkTopologyGroveCondition,
    "NetworkTopologyLabel": NetworkTopologyLabel,
    "NetworkTopologySyncFields": NetworkTopologySyncFields,
    "NetworkTopologySyncGrove": NetworkTopologySyncGrove,
    "NetworkTopologySyncGroveCondition": NetworkTopologySyncGroveCondition,
    "NetworkTopologySyncRequest": NetworkTopologySyncRequest,
    "NetworkTopologyUpdateFields": NetworkTopologyUpdateFields,
    "Node": Node,
    "NodeAdditionalFields": NodeAdditionalFields,
    "NodeAdditionalReadFields": NodeAdditionalReadFields,
    "NodeExtended": NodeExtended,
    "NodeExtendedAllOf": NodeExtendedAllOf,
    "NodeForSync": NodeForSync,
    "NodeForSyncAllOf": NodeForSyncAllOf,
    "NodeInfo": NodeInfo,
    "NodeInternal": NodeInternal,
    "NodeNumaForSync": NodeNumaForSync,
    "NodePoolSetting": NodePoolSetting,
    "NodePoolSettingsGetResponse": NodePoolSettingsGetResponse,
    "NodeStatusConditionDetails": NodeStatusConditionDetails,
    "NodeTaint": NodeTaint,
    "Nodepool": Nodepool,
    "NodepoolCalculatedFields": NodepoolCalculatedFields,
    "NodepoolConditionDetails": NodepoolConditionDetails,
    "NodepoolCreateFields": NodepoolCreateFields,
    "NodepoolCreateResponseFields": NodepoolCreateResponseFields,
    "NodepoolDeprecatedFields": NodepoolDeprecatedFields,
    "NodepoolDeprecatedPlacementStrategy": NodepoolDeprecatedPlacementStrategy,
    "NodepoolGpuTypeDetails": NodepoolGpuTypeDetails,
    "NodepoolMinimal": NodepoolMinimal,
    "NodepoolSyncFields": NodepoolSyncFields,
    "NodepoolSyncFieldsStatus": NodepoolSyncFieldsStatus,
    "NodepoolSyncFieldsStatusAllOf": NodepoolSyncFieldsStatusAllOf,
    "NodepoolSyncUpdateFields": NodepoolSyncUpdateFields,
    "NodepoolSyncUpdateFieldsPlacementStrategy": NodepoolSyncUpdateFieldsPlacementStrategy,
    "NodepoolUpdateFields": NodepoolUpdateFields,
    "Nodes": Nodes,
    "NodesSyncRequest": NodesSyncRequest,
    "NumaInfo": NumaInfo,
    "NumaInfoExtended": NumaInfoExtended,
    "NumaInfoExtendedAllOf": NumaInfoExtendedAllOf,
    "NumaNodeDetails": NumaNodeDetails,
    "NumaNodesSyncRequest": NumaNodesSyncRequest,
    "PatchClusterConfigRequest": PatchClusterConfigRequest,
    "PutClusterConfigRequest": PutClusterConfigRequest,
    "PutClusterConfigStatusRequest": PutClusterConfigStatusRequest,
    "PutClusterConfigStatusRequestStatus": PutClusterConfigStatusRequestStatus,
    "SchedulingCapabilities": SchedulingCapabilities,
    "SchedulingCapabilitiesNullable": SchedulingCapabilitiesNullable,
    "SchedulingConfiguration": SchedulingConfiguration,
    "SchedulingConfigurationPlacementStrategy": SchedulingConfigurationPlacementStrategy,
    "SubmissionError": SubmissionError,
    "SyncNodepoolsStatus207Response": SyncNodepoolsStatus207Response,
    "SyncNodepoolsStatusRequest": SyncNodepoolsStatusRequest,
    "TelemetryResponse": TelemetryResponse,
    "TelemetryResponseValuesInner": TelemetryResponseValuesInner,
    "TelemetryResponseValuesInnerGroupsInner": TelemetryResponseValuesInnerGroupsInner,
    "TimeBasedFairShare": TimeBasedFairShare,
    "UpdateClusterSettingByKeyRequest": UpdateClusterSettingByKeyRequest,
    "UpdateNodePoolSettingByKeyRequest": UpdateNodePoolSettingByKeyRequest,
}

export class ObjectSerializer {
    public static findCorrectType(data: any, expectedType: string) {
        if (data == undefined) {
            return expectedType;
        } else if (primitives.indexOf(expectedType.toLowerCase()) !== -1) {
            return expectedType;
        } else if (expectedType === "Date") {
            return expectedType;
        } else {
            if (enumsMap[expectedType]) {
                return expectedType;
            }

            if (!typeMap[expectedType]) {
                return expectedType; // w/e we don't know the type
            }

            // Check the discriminator
            let discriminatorProperty = typeMap[expectedType].discriminator;
            if (discriminatorProperty == null) {
                return expectedType; // the type does not have a discriminator. use it.
            } else {
                if (data[discriminatorProperty]) {
                    var discriminatorType = data[discriminatorProperty];
                    if(typeMap[discriminatorType]){
                        return discriminatorType; // use the type given in the discriminator
                    } else {
                        return expectedType; // discriminator did not map to a type
                    }
                } else {
                    return expectedType; // discriminator was not present (or an empty string)
                }
            }
        }
    }

    public static serialize(data: any, type: string) {
        if (data == undefined) {
            return data;
        } else if (primitives.indexOf(type.toLowerCase()) !== -1) {
            return data;
        } else if (type.lastIndexOf("Array<", 0) === 0) { // string.startsWith pre es6
            let subType: string = type.replace("Array<", ""); // Array<Type> => Type>
            subType = subType.substring(0, subType.length - 1); // Type> => Type
            let transformedData: any[] = [];
            for (let index = 0; index < data.length; index++) {
                let datum = data[index];
                transformedData.push(ObjectSerializer.serialize(datum, subType));
            }
            return transformedData;
        } else if (type === "Date") {
            return data.toISOString();
        } else {
            if (enumsMap[type]) {
                return data;
            }
            if (!typeMap[type]) { // in case we dont know the type
                return data;
            }

            // Get the actual type of this object
            type = this.findCorrectType(data, type);

            // get the map for the correct type.
            let attributeTypes = typeMap[type].getAttributeTypeMap();
            let instance: {[index: string]: any} = {};
            for (let index = 0; index < attributeTypes.length; index++) {
                let attributeType = attributeTypes[index];
                instance[attributeType.baseName] = ObjectSerializer.serialize(data[attributeType.name], attributeType.type);
            }
            return instance;
        }
    }

    public static deserialize(data: any, type: string) {
        // polymorphism may change the actual type.
        type = ObjectSerializer.findCorrectType(data, type);
        if (data == undefined) {
            return data;
        } else if (primitives.indexOf(type.toLowerCase()) !== -1) {
            return data;
        } else if (type.lastIndexOf("Array<", 0) === 0) { // string.startsWith pre es6
            let subType: string = type.replace("Array<", ""); // Array<Type> => Type>
            subType = subType.substring(0, subType.length - 1); // Type> => Type
            let transformedData: any[] = [];
            for (let index = 0; index < data.length; index++) {
                let datum = data[index];
                transformedData.push(ObjectSerializer.deserialize(datum, subType));
            }
            return transformedData;
        } else if (type === "Date") {
            return new Date(data);
        } else {
            if (enumsMap[type]) {// is Enum
                return data;
            }

            if (!typeMap[type]) { // dont know the type
                return data;
            }
            let instance = new typeMap[type]();
            let attributeTypes = typeMap[type].getAttributeTypeMap();
            for (let index = 0; index < attributeTypes.length; index++) {
                let attributeType = attributeTypes[index];
                instance[attributeType.name] = ObjectSerializer.deserialize(data[attributeType.baseName], attributeType.type);
            }
            return instance;
        }
    }
}

export interface Authentication {
    /**
    * Apply authentication settings to header and query params.
    */
    applyToRequest(requestOptions: localVarRequest.Options): Promise<void> | void;
}

export class HttpBasicAuth implements Authentication {
    public username: string = '';
    public password: string = '';

    applyToRequest(requestOptions: localVarRequest.Options): void {
        requestOptions.auth = {
            username: this.username, password: this.password
        }
    }
}

export class HttpBearerAuth implements Authentication {
    public accessToken: string | (() => string) = '';

    applyToRequest(requestOptions: localVarRequest.Options): void {
        if (requestOptions && requestOptions.headers) {
            const accessToken = typeof this.accessToken === 'function'
                            ? this.accessToken()
                            : this.accessToken;
            requestOptions.headers["Authorization"] = "Bearer " + accessToken;
        }
    }
}

export class ApiKeyAuth implements Authentication {
    public apiKey: string = '';

    constructor(private location: string, private paramName: string) {
    }

    applyToRequest(requestOptions: localVarRequest.Options): void {
        if (this.location == "query") {
            (<any>requestOptions.qs)[this.paramName] = this.apiKey;
        } else if (this.location == "header" && requestOptions && requestOptions.headers) {
            requestOptions.headers[this.paramName] = this.apiKey;
        } else if (this.location == 'cookie' && requestOptions && requestOptions.headers) {
            if (requestOptions.headers['Cookie']) {
                requestOptions.headers['Cookie'] += '; ' + this.paramName + '=' + encodeURIComponent(this.apiKey);
            }
            else {
                requestOptions.headers['Cookie'] = this.paramName + '=' + encodeURIComponent(this.apiKey);
            }
        }
    }
}

export class OAuth implements Authentication {
    public accessToken: string = '';

    applyToRequest(requestOptions: localVarRequest.Options): void {
        if (requestOptions && requestOptions.headers) {
            requestOptions.headers["Authorization"] = "Bearer " + this.accessToken;
        }
    }
}

export class VoidAuth implements Authentication {
    public username: string = '';
    public password: string = '';

    applyToRequest(_: localVarRequest.Options): void {
        // Do nothing
    }
}

export type Interceptor = (requestOptions: localVarRequest.Options) => (Promise<void> | void);
