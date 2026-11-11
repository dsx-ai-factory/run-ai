import localVarRequest from 'request';

export * from './aIApplication';
export * from './aIApplicationAllocatedResources';
export * from './aIApplicationAuth';
export * from './aIApplicationChartUploadResponse';
export * from './aIApplicationCreationRequest';
export * from './aIApplicationForSync';
export * from './aIApplicationRequestResources';
export * from './aIApplicationResource';
export * from './aIApplicationSpec';
export * from './aIApplicationSpecChartRef';
export * from './aIApplicationSyncAction';
export * from './aIApplicationSyncActionInfo';
export * from './aIApplicationSyncInfo';
export * from './aIApplicationSyncItem';
export * from './aIApplicationSyncResource';
export * from './aIApplicationSyncResponse';
export * from './aIApplicationVerbosity';
export * from './aIApplicationsForSync';
export * from './aIApplicationsSortFilterFields';
export * from './action';
export * from './actionsSupport';
export * from './activationReplicasField';
export * from './aiApplicationEndpoint';
export * from './aiApplicationEndpointAllOf';
export * from './aiApplicationResourceConnection';
export * from './aiApplicationResourceConnectionAllOf';
export * from './allocatedResources';
export * from './ancestors';
export * from './annotation';
export * from './annotationsDefaults';
export * from './applicationSubmissionSetting';
export * from './arbitraryRules';
export * from './arrayRules';
export * from './assetKind';
export * from './assetSubmissionIds';
export * from './assetsMapping';
export * from './authorizationType';
export * from './autoScaling';
export * from './autoScalingCommonFields';
export * from './autoScalingField';
export * from './autoScalingMetricField';
export * from './autoScalingMetricFields';
export * from './autoScalingMetricOptions';
export * from './autoScalingMetricOptionsOptionsInner';
export * from './autoScalingMetricRules';
export * from './autoScalingRules';
export * from './autoScalingRulesInitializationTimeoutSeconds';
export * from './basePriority';
export * from './batchResponse';
export * from './batchResponseFailureItem';
export * from './booleanRules';
export * from './capability';
export * from './catalog';
export * from './catalogParams';
export * from './catalogParamsOrderByInner';
export * from './catalogResourcesInner';
export * from './claimInfo';
export * from './claimInfoRules';
export * from './claimInfoUpdatable';
export * from './cleanPodPolicy';
export * from './cleanPodPolicyOptions';
export * from './cleanPodPolicyOptionsOptionsInner';
export * from './cleanPodPolicyRules';
export * from './clusterPriority';
export * from './clusterPriorityAllOf';
export * from './clusterSyncReport';
export * from './clusterSyncReportUpdate';
export * from './complianceInfoReason';
export * from './complianceIssues';
export * from './complianceIssuesComplianceIssuesInner';
export * from './complianceIssuesV2';
export * from './complianceIssuesV2ComplianceIssuesInner';
export * from './concurrencyHardLimitField';
export * from './condition';
export * from './configMap';
export * from './configMapInstance';
export * from './configMapRules';
export * from './configMapsDefaults';
export * from './configMapsRules';
export * from './connection';
export * from './container';
export * from './countResponse';
export * from './countWorkloadElements200Response';
export * from './countWorkloads200Response';
export * from './dataVolume';
export * from './dataVolumeInstance';
export * from './dataVolumeRules';
export * from './dataVolumesDefaults';
export * from './dataVolumesRules';
export * from './datasource';
export * from './defaultFromRule';
export * from './distributedCreationRequest';
export * from './distributedFramework';
export * from './distributedFrameworkOptions';
export * from './distributedFrameworkOptionsOptionsInner';
export * from './distributedFrameworkRules';
export * from './distributedInference';
export * from './distributedInferenceAssets';
export * from './distributedInferenceAssetsAssets';
export * from './distributedInferenceCommonDefaultsV2';
export * from './distributedInferenceCommonRulesV2';
export * from './distributedInferenceCommonRulesV2AllOf';
export * from './distributedInferenceCommonSpec';
export * from './distributedInferenceCommonSpecAllOf';
export * from './distributedInferenceCreationRequest';
export * from './distributedInferenceLeaderDefaultsFields';
export * from './distributedInferenceLeaderDefaultsFieldsLeader';
export * from './distributedInferenceLeaderRulesFields';
export * from './distributedInferenceLeaderRulesFieldsLeader';
export * from './distributedInferenceLeaderSpecFields';
export * from './distributedInferenceLeaderSpecFieldsLeader';
export * from './distributedInferenceLeaderTemplateSpecFields';
export * from './distributedInferenceLeaderTemplateSpecFieldsLeader';
export * from './distributedInferenceLeaderWorkerDefaultsV2';
export * from './distributedInferenceLeaderWorkerDefaultsV2Security';
export * from './distributedInferenceLeaderWorkerDefaultsV2Storage';
export * from './distributedInferenceLeaderWorkerRulesV2';
export * from './distributedInferenceLeaderWorkerRulesV2Security';
export * from './distributedInferenceLeaderWorkerRulesV2SecurityPrivileged';
export * from './distributedInferenceLeaderWorkerRulesV2Storage';
export * from './distributedInferenceLeaderWorkerSpec';
export * from './distributedInferenceLeaderWorkerSpecCompute';
export * from './distributedInferenceLeaderWorkerSpecSecurity';
export * from './distributedInferenceLeaderWorkerSpecStorage';
export * from './distributedInferenceLeaderWorkerTemplateSpec';
export * from './distributedInferenceReplicasField';
export * from './distributedInferenceReplicasFieldRules';
export * from './distributedInferenceRestartPolicy';
export * from './distributedInferenceRestartPolicyOptions';
export * from './distributedInferenceRestartPolicyOptionsOptionsInner';
export * from './distributedInferenceRestartPolicyRules';
export * from './distributedInferenceServingPort';
export * from './distributedInferenceServingPortAccess';
export * from './distributedInferenceServingPortAccessAuthorizationTypeEnum';
export * from './distributedInferenceServingPortAuthorizationTypeOptions';
export * from './distributedInferenceServingPortAuthorizationTypeOptionsOptionsInner';
export * from './distributedInferenceServingPortAuthorizationTypeRules';
export * from './distributedInferenceServingPortContainerAndProtocol';
export * from './distributedInferenceServingPortProtocol';
export * from './distributedInferenceServingPortRules';
export * from './distributedInferenceServingPortRulesProperties';
export * from './distributedInferenceSpec';
export * from './distributedInferenceSpecSpec';
export * from './distributedInferenceStartupPolicy';
export * from './distributedInferenceStartupPolicyField';
export * from './distributedInferenceStartupPolicyFieldRules';
export * from './distributedInferenceStartupPolicyFieldRulesStartupPolicy';
export * from './distributedInferenceStartupPolicyOptions';
export * from './distributedInferenceTemplate';
export * from './distributedInferenceTemplateAllOf';
export * from './distributedInferenceTemplateAllOf1';
export * from './distributedInferenceTemplateCreationRequest';
export * from './distributedInferenceTemplatePatchRequest';
export * from './distributedInferenceTemplateSpec';
export * from './distributedInferenceTemplateUpdateRequest';
export * from './distributedInferenceTemplatesListResponse';
export * from './distributedInferenceTemplatesListResponseItem';
export * from './distributedInferenceWorkerDefaultsFields';
export * from './distributedInferenceWorkerDefaultsFieldsWorker';
export * from './distributedInferenceWorkerRulesFields';
export * from './distributedInferenceWorkerRulesFieldsWorker';
export * from './distributedInferenceWorkerSpecFields';
export * from './distributedInferenceWorkerSpecFieldsWorker';
export * from './distributedInferenceWorkerTemplateSpecFields';
export * from './distributedInferenceWorkerTemplateSpecFieldsWorker';
export * from './distributedInferenceWorkersField';
export * from './distributedInferenceWorkersFieldRules';
export * from './distributedSpec';
export * from './distributedSpecSpec';
export * from './distributedSpecSpecAllOf';
export * from './distributedTemplate';
export * from './distributedTemplateAllOf';
export * from './distributedTemplateCreationRequest';
export * from './distributedTemplateCreationRequestAllOf';
export * from './distributedTemplateCreationRequestAllOf1';
export * from './distributedTemplatePatchRequest';
export * from './distributedTemplateSpec';
export * from './distributedTemplateSpecAllOf';
export * from './distributedTemplateUpdateRequest';
export * from './distributedTemplatesListResponse';
export * from './distributedTemplatesListResponseItem';
export * from './distributedWorkload';
export * from './elementMetricType';
export * from './emptyDir';
export * from './emptyDirInstance';
export * from './emptyDirRules';
export * from './emptyDirsDefaults';
export * from './emptyDirsRules';
export * from './enableHybridSubmissionSetting';
export * from './enablePoliciesSetting';
export * from './enableTemplatesV2Setting';
export * from './endpoint';
export * from './endpointSync';
export * from './endpointSyncAllOf';
export * from './endpointSyncAllOfSourceRef';
export * from './environment';
export * from './environmentVariable';
export * from './environmentVariableConfigMap';
export * from './environmentVariablePodFieldReference';
export * from './environmentVariableSecret';
export * from './environmentVariableUserCredential';
export * from './environmentVariablesDefaults';
export * from './event';
export * from './eventComputedFields';
export * from './eventSyncFields';
export * from './eventsSyncRequest';
export * from './excludeField';
export * from './exposedUrl';
export * from './exposedUrlRules';
export * from './exposedUrlsDefaults';
export * from './exposedUrlsRules';
export * from './extendedResource';
export * from './extendedResourceRules';
export * from './extendedResourcesDefaults';
export * from './extendedResourcesRules';
export * from './fieldsToClear';
export * from './getApplicationSubmission';
export * from './getAvailablePrioritiesAndCategories200Response';
export * from './getEnableHybridSubmission';
export * from './getEnablePolicies';
export * from './getEnableSettingsValue';
export * from './getEnableTemplatesV2';
export * from './getGpuOverProvisioning';
export * from './getInferenceWorkloadRevisions200Response';
export * from './getRevisionPods200Response';
export * from './getWorkloadClusterPriorities200Response';
export * from './getWorkloadEndpoints200Response';
export * from './getWorkloadEvents200Response';
export * from './getWorkloadHistory200Response';
export * from './getWorkloadPriorities200Response';
export * from './getWorkloadSubmissionV2Enabled';
export * from './getWorkloads200Response';
export * from './getWorkloadsSetting';
export * from './gitAuthenticationMethod';
export * from './gitCommon';
export * from './gitInstance';
export * from './gitInstanceAllOf';
export * from './gitPassword';
export * from './gitRules';
export * from './gitSecretRef';
export * from './gitsDefaults';
export * from './gitsRules';
export * from './gpuOverProvisioningSetting';
export * from './gpuRequestOptions';
export * from './gpuRequestOptionsOptionsInner';
export * from './gpuRequestRules';
export * from './gpuRequestType';
export * from './histogramSeries';
export * from './histogramValue';
export * from './historyRecord';
export * from './historyRecordMeta';
export * from './historyRecordSpec';
export * from './historyRecordType';
export * from './hostPath';
export * from './hostPathInstance';
export * from './hostPathMountPropagation';
export * from './hostPathMountPropagationOptions';
export * from './hostPathMountPropagationOptionsOptionsInner';
export * from './hostPathMountPropagationRules';
export * from './hostPathRules';
export * from './hostPathsDefaults';
export * from './hostPathsRules';
export * from './httpResponse';
export * from './imagePullPolicy';
export * from './imagePullPolicyOptions';
export * from './imagePullPolicyOptionsOptionsInner';
export * from './imagePullPolicyRules';
export * from './imagePullSecret';
export * from './imagePullSecretsDefaults';
export * from './imagePullSecretsRules';
export * from './imagePullSecretsRulesAttributes';
export * from './inference';
export * from './inferenceCreationRequest';
export * from './inferenceFields';
export * from './inferenceFieldsRules';
export * from './inferencePodMetricType';
export * from './inferenceSpec';
export * from './inferenceSpecSpec';
export * from './inferenceSpecSpecAllOf';
export * from './inferenceSpecSpecAllOfSecurity';
export * from './inferenceSpecSpecAllOfStorage';
export * from './inferenceTemplate';
export * from './inferenceTemplateAllOf';
export * from './inferenceTemplateAllOf1';
export * from './inferenceTemplateCreationRequest';
export * from './inferenceTemplatePatchRequest';
export * from './inferenceTemplateSpec';
export * from './inferenceTemplateSpecAllOf';
export * from './inferenceTemplateUpdateRequest';
export * from './inferenceTemplatesListResponse';
export * from './inferenceTemplatesListResponseItem';
export * from './inferenceUpdateRequest';
export * from './inferenceUpdateSpec';
export * from './inferenceUpdateSpecAutoscaling';
export * from './inferenceUpdateSpecServingConfiguration';
export * from './inferenceUpdateSpecSpec';
export * from './inferenceUpdateSpecSpecAllOf';
export * from './inferenceWorkloadMetricType';
export * from './inferencesMaxReplicasField';
export * from './inferencesMinReplicasField';
export * from './initialReplicasField';
export * from './initializationTimeoutField';
export * from './instancesRules';
export * from './integerRules';
export * from './integerRulesOptional';
export * from './internalEnrichOrgUnitRankingRequest';
export * from './internalEnrichOrgUnitRankingResponse';
export * from './internalPodsSyncRequest';
export * from './internalSyncPods207Response';
export * from './internalSyncWorkloads207Response';
export * from './internalWorkloadsSyncRequest';
export * from './involvedObject';
export * from './itemRules';
export * from './label';
export * from './labelsDefaults';
export * from './listApplications200Response';
export * from './listCategories200Response';
export * from './listPods200Response';
export * from './listWorkloadTypes200Response';
export * from './masterSpec';
export * from './masterSpecSecurity';
export * from './masterSpecStorage';
export * from './masterTemplateSpec';
export * from './matchExpression';
export * from './matchExpressionOperator';
export * from './measurementResponse';
export * from './measurementResponseValuesInner';
export * from './metricThresholdField';
export * from './metricThresholdPercentageField';
export * from './metricsCompositeResponse';
export * from './metricsResponse';
export * from './modelError';
export * from './mpiLauncherCreationPolicy';
export * from './mpiLauncherCreationPolicyOptions';
export * from './mpiLauncherCreationPolicyOptionsOptionsInner';
export * from './mpiLauncherCreationPolicyRules';
export * from './nfs';
export * from './nfsInstance';
export * from './nfsRules';
export * from './nfssDefaults';
export * from './nfssRules';
export * from './nodeAffinityRequired';
export * from './nodeSelectorTerm';
export * from './numberRules';
export * from './orgType';
export * from './orgUnitRankingItem';
export * from './paginationInfo';
export * from './pendingSchedulingMessage';
export * from './phase';
export * from './phaseReason';
export * from './phaseUpdate';
export * from './pod';
export * from './podAffinity';
export * from './podAffinityRules';
export * from './podAffinityType';
export * from './podAffinityTypeOptions';
export * from './podAffinityTypeOptionsOptionsInner';
export * from './podAffinityTypeRules';
export * from './podComputedFields';
export * from './podElement';
export * from './podGroupsSyncFields';
export * from './podGroupsSyncRequest';
export * from './podMetricType';
export * from './podNodeAffinity';
export * from './podNodeAffinityPreferred';
export * from './podNodeAffinityRequired';
export * from './podPreferredNodeSelectorTerm';
export * from './podRequestResources';
export * from './podSortFilterFields';
export * from './podSyncFields';
export * from './podToleration';
export * from './podVerbosity';
export * from './podVerbosityFields';
export * from './podsSyncRequest';
export * from './policyRuleEnum';
export * from './port';
export * from './portRules';
export * from './portServiceType';
export * from './portServiceTypeOptions';
export * from './portServiceTypeOptionsOptionsInner';
export * from './portServiceTypeRules';
export * from './portsDefaults';
export * from './portsRules';
export * from './preemptibility';
export * from './preemptibilityOptions';
export * from './preemptibilityOptionsOptionsInner';
export * from './preemptibilityRules';
export * from './prioritiesForStatus';
export * from './prioritiesForStatusPrioritiesInner';
export * from './prioritiesForStatusPrioritiesInnerAllOf';
export * from './prioritiesForStatusPrioritiesInnerAllOfStatus';
export * from './priorityForCreate';
export * from './priorityStatus';
export * from './probe';
export * from './probeHandler';
export * from './probeHandlerHttpGet';
export * from './probeHandlerRules';
export * from './probeHandlerRulesHttpGet';
export * from './probeHandlerScheme';
export * from './probeRules';
export * from './probes';
export * from './probesRules';
export * from './pvc';
export * from './pvcAccessModes';
export * from './pvcAccessModesRules';
export * from './pvcAddedAttrValue';
export * from './pvcFieldsNonUpdatable';
export * from './pvcFieldsUpdatable';
export * from './pvcInstance';
export * from './pvcRules';
export * from './pvcVolumeMode';
export * from './pvcsDefaults';
export * from './pvcsRules';
export * from './quantityRules';
export * from './queueResources';
export * from './relatedUrl';
export * from './relatedUrlRules';
export * from './relatedUrlsDefaults';
export * from './relatedUrlsRules';
export * from './replicaType';
export * from './requestInfo';
export * from './requestResourceCores';
export * from './requestResourceQuantity';
export * from './requestedPods';
export * from './resourceRef';
export * from './resourceRefKind';
export * from './responseInfo';
export * from './restartPolicy';
export * from './revision';
export * from './revisionSyncFields';
export * from './revisionsSyncRequest';
export * from './runnableInfo';
export * from './s3AccessKey';
export * from './s3Common';
export * from './s3Instance';
export * from './s3Rules';
export * from './s3sDefaults';
export * from './s3sRules';
export * from './scaleDownDelayField';
export * from './scaleToZeroRetentionField';
export * from './schedulingCondition';
export * from './schedulingConditionQueueDetails';
export * from './scopeType';
export * from './seccompProfileType';
export * from './seccompProfileTypeOptions';
export * from './seccompProfileTypeOptionsOptionsInner';
export * from './seccompProfileTypeRules';
export * from './secret';
export * from './secretFieldsNonUpdatable';
export * from './secretFieldsUpdatable';
export * from './secretInstance';
export * from './secretRules';
export * from './secretsDefaults';
export * from './secretsRules';
export * from './servingConfiguration';
export * from './servingConfigurationField';
export * from './servingConfigurationRules';
export * from './servingPort';
export * from './servingPortAccess';
export * from './servingPortAccessAuthorizationTypeEnum';
export * from './servingPortContainerAndProtocol';
export * from './servingPortProtocol';
export * from './source';
export * from './sourceApi';
export * from './sourceOfRule';
export * from './storageInstanceName';
export * from './stringOption';
export * from './stringRules';
export * from './stringRulesCommon';
export * from './stringRulesOptions';
export * from './submissionError';
export * from './submissionErrorV2';
export * from './submitWithAssetIds';
export * from './submitWithTemplateId';
export * from './supersetDefaults';
export * from './supersetDefaultsAllOf';
export * from './supersetDefaultsAllOfCompute';
export * from './supersetDefaultsAllOfSecurity';
export * from './supersetDefaultsAllOfStorage';
export * from './supersetRules';
export * from './supersetRulesAllOf';
export * from './supersetRulesAllOfCompute';
export * from './supersetRulesAllOfSecurity';
export * from './supersetRulesAllOfStorage';
export * from './supersetServingPort';
export * from './supersetServingPortRules';
export * from './supersetSpec';
export * from './supersetSpecAllOf';
export * from './supersetTemplate';
export * from './supersetTemplateAllOf';
export * from './supersetWorkload';
export * from './supersetWorkloadAllOf';
export * from './syncJobsBackwardsRequest';
export * from './syncPodGroups207Response';
export * from './syncPodsBackwards200Response';
export * from './syncPriority';
export * from './syncPriorityAllOf';
export * from './syncRevisions207Response';
export * from './systemInfo';
export * from './telemetryResponse';
export * from './telemetryResponseValuesInner';
export * from './telemetryResponseValuesInnerGroupsInner';
export * from './templateAssets';
export * from './templateAssetsIds';
export * from './templateAssetsIdsOptional';
export * from './templateComplianceInfo';
export * from './templateCreationMeta';
export * from './templateCreationMetaFields';
export * from './templateMasterSection';
export * from './templateMeta';
export * from './templateMetaFields';
export * from './templatePatchMeta';
export * from './templatePatchMetaFields';
export * from './templateUpdateMeta';
export * from './templateUpdateMetaFields';
export * from './templatesListResponse';
export * from './toleration';
export * from './tolerationEffect';
export * from './tolerationEffectOptions';
export * from './tolerationEffectOptionsOptionsInner';
export * from './tolerationEffectRules';
export * from './tolerationOperator';
export * from './tolerationOperatorOptions';
export * from './tolerationOperatorOptionsOptionsInner';
export * from './tolerationOperatorRules';
export * from './tolerationRules';
export * from './tolerationsDefaults';
export * from './tolerationsRules';
export * from './training';
export * from './trainingCreationRequest';
export * from './trainingSpec';
export * from './trainingSpecSpec';
export * from './trainingSpecSpecAllOf';
export * from './trainingTemplate';
export * from './trainingTemplateAllOf';
export * from './trainingTemplateAllOf1';
export * from './trainingTemplateCreationRequest';
export * from './trainingTemplatePatchRequest';
export * from './trainingTemplateSpec';
export * from './trainingTemplateSpecAllOf';
export * from './trainingTemplateUpdateRequest';
export * from './trainingTemplatesListResponse';
export * from './trainingTemplatesListResponseItem';
export * from './uidGidSource';
export * from './uidGidSourceOptions';
export * from './uidGidSourceOptionsOptionsInner';
export * from './uidGidSourceRules';
export * from './unifiedPolicyInfo';
export * from './unifiedPolicyInfoPerReplicaType';
export * from './unifiedPolicyInfoPerReplicaTypeDefaults';
export * from './unifiedPolicyInfoPerReplicaTypeRules';
export * from './unifiedProjectInfo';
export * from './unifiedSpec';
export * from './unifiedSpecMaster';
export * from './unifiedSpecWorker';
export * from './unifiedWorkload';
export * from './unifiedWorkloadRequest';
export * from './updateNodePodsRequest';
export * from './updateRequest';
export * from './updateSpec';
export * from './updateSpecSpec';
export * from './updateSpecSpecAllOf';
export * from './updateWorkloadsSettingByKeyRequest';
export * from './workload';
export * from './workloadAllocatedResources';
export * from './workloadBatch';
export * from './workloadCategory';
export * from './workloadChildrenIdsInner';
export * from './workloadCreationFields';
export * from './workloadCreationMeta';
export * from './workloadDesiredPhase';
export * from './workloadDetailed';
export * from './workloadDetailedAllOf';
export * from './workloadElement';
export * from './workloadElementKubernetesRef';
export * from './workloadElementKubernetesRefGvk';
export * from './workloadElementResponse';
export * from './workloadElementResponseAllOf';
export * from './workloadElementTopology';
export * from './workloadElementsResponse';
export * from './workloadKubernetesRef';
export * from './workloadMeta';
export * from './workloadMetricType';
export * from './workloadNetworkStatus';
export * from './workloadNetworkStatusServicesInner';
export * from './workloadNetworkStatusServicesInnerLoadBalancer';
export * from './workloadNetworkStatusServicesInnerLoadBalancerIngressInner';
export * from './workloadNetworkStatusServicesInnerPortsInner';
export * from './workloadNetworkStatusServicesInnerUrlsInner';
export * from './workloadNodeAffinity';
export * from './workloadPatchFields';
export * from './workloadPodsSyncRequest';
export * from './workloadPriority';
export * from './workloadPriorityAllOf';
export * from './workloadRequestResources';
export * from './workloadSortFilterFields';
export * from './workloadSubmissionV2EnabledSetting';
export * from './workloadSyncFields';
export * from './workloadSyncFieldsTopology';
export * from './workloadTelemetryType';
export * from './workloadTopology';
export * from './workloadTopologyActualPlacement';
export * from './workloadType';
export * from './workloadTypeAdditionalCreateFields';
export * from './workloadTypeAdditionalResponseFields';
export * from './workloadTypeConfig';
export * from './workloadTypeConfigAllOf';
export * from './workloadTypeCreateFields';
export * from './workloadTypeCreateFieldsAllOf';
export * from './workloadTypeCreateUpdateInnerFields';
export * from './workloadTypeForSync';
export * from './workloadTypeForSyncAllOf';
export * from './workloadTypeKartasWithVersionInner';
export * from './workloadTypePreemptibility';
export * from './workloadTypeResourceInterfacesWithVersionInner';
export * from './workloadTypeStatus';
export * from './workloadTypeStatusAllOf';
export * from './workloadTypeStatusFields';
export * from './workloadTypeUpdateFields';
export * from './workloadTypeUpdateFieldsAllOf';
export * from './workloadTypesForSync';
export * from './workloadsExtendedResource';
export * from './workloadsSettingsGetResponse';
export * from './workloadsSettingsKeyEnum';
export * from './workloadsSyncInfo';
export * from './workloadsSyncRequest';
export * from './workloadsSyncResponse';
export * from './workspace';
export * from './workspaceCreationRequest';
export * from './workspaceSpec';
export * from './workspaceSpecSpec';
export * from './workspaceSpecSpecAllOf';
export * from './workspaceSpecSpecAllOfCompute';
export * from './workspaceSpecSpecAllOfSecurity';
export * from './workspaceSpecSpecAllOfStorage';
export * from './workspaceTemplate';
export * from './workspaceTemplateAllOf';
export * from './workspaceTemplateAllOf1';
export * from './workspaceTemplateCreationRequest';
export * from './workspaceTemplatePatchRequest';
export * from './workspaceTemplateSpec';
export * from './workspaceTemplateSpecAllOf';
export * from './workspaceTemplateUpdateRequest';
export * from './workspaceTemplatesListResponse';
export * from './workspaceTemplatesListResponseItem';

import * as fs from 'fs';

export interface RequestDetailedFile {
    value: Buffer;
    options?: {
        filename?: string;
        contentType?: string;
    }
}

export type RequestFile = string | Buffer | fs.ReadStream | RequestDetailedFile;


import { AIApplication } from './aIApplication';
import { AIApplicationAllocatedResources } from './aIApplicationAllocatedResources';
import { AIApplicationAuth } from './aIApplicationAuth';
import { AIApplicationChartUploadResponse } from './aIApplicationChartUploadResponse';
import { AIApplicationCreationRequest } from './aIApplicationCreationRequest';
import { AIApplicationForSync } from './aIApplicationForSync';
import { AIApplicationRequestResources } from './aIApplicationRequestResources';
import { AIApplicationResource } from './aIApplicationResource';
import { AIApplicationSpec } from './aIApplicationSpec';
import { AIApplicationSpecChartRef } from './aIApplicationSpecChartRef';
import { AIApplicationSyncAction } from './aIApplicationSyncAction';
import { AIApplicationSyncActionInfo } from './aIApplicationSyncActionInfo';
import { AIApplicationSyncInfo } from './aIApplicationSyncInfo';
import { AIApplicationSyncItem } from './aIApplicationSyncItem';
import { AIApplicationSyncResource } from './aIApplicationSyncResource';
import { AIApplicationSyncResponse } from './aIApplicationSyncResponse';
import { AIApplicationVerbosity } from './aIApplicationVerbosity';
import { AIApplicationsForSync } from './aIApplicationsForSync';
import { AIApplicationsSortFilterFields } from './aIApplicationsSortFilterFields';
import { Action } from './action';
import { ActionsSupport } from './actionsSupport';
import { ActivationReplicasField } from './activationReplicasField';
import { AiApplicationEndpoint } from './aiApplicationEndpoint';
import { AiApplicationEndpointAllOf } from './aiApplicationEndpointAllOf';
import { AiApplicationResourceConnection } from './aiApplicationResourceConnection';
import { AiApplicationResourceConnectionAllOf } from './aiApplicationResourceConnectionAllOf';
import { AllocatedResources } from './allocatedResources';
import { Ancestors } from './ancestors';
import { Annotation } from './annotation';
import { AnnotationsDefaults } from './annotationsDefaults';
import { ApplicationSubmissionSetting } from './applicationSubmissionSetting';
import { ArbitraryRules } from './arbitraryRules';
import { ArrayRules } from './arrayRules';
import { AssetKind } from './assetKind';
import { AssetSubmissionIds } from './assetSubmissionIds';
import { AssetsMapping } from './assetsMapping';
import { AuthorizationType } from './authorizationType';
import { AutoScaling } from './autoScaling';
import { AutoScalingCommonFields } from './autoScalingCommonFields';
import { AutoScalingField } from './autoScalingField';
import { AutoScalingMetricField } from './autoScalingMetricField';
import { AutoScalingMetricFields } from './autoScalingMetricFields';
import { AutoScalingMetricOptions } from './autoScalingMetricOptions';
import { AutoScalingMetricOptionsOptionsInner } from './autoScalingMetricOptionsOptionsInner';
import { AutoScalingMetricRules } from './autoScalingMetricRules';
import { AutoScalingRules } from './autoScalingRules';
import { AutoScalingRulesInitializationTimeoutSeconds } from './autoScalingRulesInitializationTimeoutSeconds';
import { BasePriority } from './basePriority';
import { BatchResponse } from './batchResponse';
import { BatchResponseFailureItem } from './batchResponseFailureItem';
import { BooleanRules } from './booleanRules';
import { Capability } from './capability';
import { Catalog } from './catalog';
import { CatalogParams } from './catalogParams';
import { CatalogParamsOrderByInner } from './catalogParamsOrderByInner';
import { CatalogResourcesInner } from './catalogResourcesInner';
import { ClaimInfo } from './claimInfo';
import { ClaimInfoRules } from './claimInfoRules';
import { ClaimInfoUpdatable } from './claimInfoUpdatable';
import { CleanPodPolicy } from './cleanPodPolicy';
import { CleanPodPolicyOptions } from './cleanPodPolicyOptions';
import { CleanPodPolicyOptionsOptionsInner } from './cleanPodPolicyOptionsOptionsInner';
import { CleanPodPolicyRules } from './cleanPodPolicyRules';
import { ClusterPriority } from './clusterPriority';
import { ClusterPriorityAllOf } from './clusterPriorityAllOf';
import { ClusterSyncReport } from './clusterSyncReport';
import { ClusterSyncReportUpdate } from './clusterSyncReportUpdate';
import { ComplianceInfoReason } from './complianceInfoReason';
import { ComplianceIssues } from './complianceIssues';
import { ComplianceIssuesComplianceIssuesInner } from './complianceIssuesComplianceIssuesInner';
import { ComplianceIssuesV2 } from './complianceIssuesV2';
import { ComplianceIssuesV2ComplianceIssuesInner } from './complianceIssuesV2ComplianceIssuesInner';
import { ConcurrencyHardLimitField } from './concurrencyHardLimitField';
import { Condition } from './condition';
import { ConfigMap } from './configMap';
import { ConfigMapInstance } from './configMapInstance';
import { ConfigMapRules } from './configMapRules';
import { ConfigMapsDefaults } from './configMapsDefaults';
import { ConfigMapsRules } from './configMapsRules';
import { Connection } from './connection';
import { Container } from './container';
import { CountResponse } from './countResponse';
import { CountWorkloadElements200Response } from './countWorkloadElements200Response';
import { CountWorkloads200Response } from './countWorkloads200Response';
import { DataVolume } from './dataVolume';
import { DataVolumeInstance } from './dataVolumeInstance';
import { DataVolumeRules } from './dataVolumeRules';
import { DataVolumesDefaults } from './dataVolumesDefaults';
import { DataVolumesRules } from './dataVolumesRules';
import { Datasource } from './datasource';
import { DefaultFromRule } from './defaultFromRule';
import { DistributedCreationRequest } from './distributedCreationRequest';
import { DistributedFramework } from './distributedFramework';
import { DistributedFrameworkOptions } from './distributedFrameworkOptions';
import { DistributedFrameworkOptionsOptionsInner } from './distributedFrameworkOptionsOptionsInner';
import { DistributedFrameworkRules } from './distributedFrameworkRules';
import { DistributedInference } from './distributedInference';
import { DistributedInferenceAssets } from './distributedInferenceAssets';
import { DistributedInferenceAssetsAssets } from './distributedInferenceAssetsAssets';
import { DistributedInferenceCommonDefaultsV2 } from './distributedInferenceCommonDefaultsV2';
import { DistributedInferenceCommonRulesV2 } from './distributedInferenceCommonRulesV2';
import { DistributedInferenceCommonRulesV2AllOf } from './distributedInferenceCommonRulesV2AllOf';
import { DistributedInferenceCommonSpec } from './distributedInferenceCommonSpec';
import { DistributedInferenceCommonSpecAllOf } from './distributedInferenceCommonSpecAllOf';
import { DistributedInferenceCreationRequest } from './distributedInferenceCreationRequest';
import { DistributedInferenceLeaderDefaultsFields } from './distributedInferenceLeaderDefaultsFields';
import { DistributedInferenceLeaderDefaultsFieldsLeader } from './distributedInferenceLeaderDefaultsFieldsLeader';
import { DistributedInferenceLeaderRulesFields } from './distributedInferenceLeaderRulesFields';
import { DistributedInferenceLeaderRulesFieldsLeader } from './distributedInferenceLeaderRulesFieldsLeader';
import { DistributedInferenceLeaderSpecFields } from './distributedInferenceLeaderSpecFields';
import { DistributedInferenceLeaderSpecFieldsLeader } from './distributedInferenceLeaderSpecFieldsLeader';
import { DistributedInferenceLeaderTemplateSpecFields } from './distributedInferenceLeaderTemplateSpecFields';
import { DistributedInferenceLeaderTemplateSpecFieldsLeader } from './distributedInferenceLeaderTemplateSpecFieldsLeader';
import { DistributedInferenceLeaderWorkerDefaultsV2 } from './distributedInferenceLeaderWorkerDefaultsV2';
import { DistributedInferenceLeaderWorkerDefaultsV2Security } from './distributedInferenceLeaderWorkerDefaultsV2Security';
import { DistributedInferenceLeaderWorkerDefaultsV2Storage } from './distributedInferenceLeaderWorkerDefaultsV2Storage';
import { DistributedInferenceLeaderWorkerRulesV2 } from './distributedInferenceLeaderWorkerRulesV2';
import { DistributedInferenceLeaderWorkerRulesV2Security } from './distributedInferenceLeaderWorkerRulesV2Security';
import { DistributedInferenceLeaderWorkerRulesV2SecurityPrivileged } from './distributedInferenceLeaderWorkerRulesV2SecurityPrivileged';
import { DistributedInferenceLeaderWorkerRulesV2Storage } from './distributedInferenceLeaderWorkerRulesV2Storage';
import { DistributedInferenceLeaderWorkerSpec } from './distributedInferenceLeaderWorkerSpec';
import { DistributedInferenceLeaderWorkerSpecCompute } from './distributedInferenceLeaderWorkerSpecCompute';
import { DistributedInferenceLeaderWorkerSpecSecurity } from './distributedInferenceLeaderWorkerSpecSecurity';
import { DistributedInferenceLeaderWorkerSpecStorage } from './distributedInferenceLeaderWorkerSpecStorage';
import { DistributedInferenceLeaderWorkerTemplateSpec } from './distributedInferenceLeaderWorkerTemplateSpec';
import { DistributedInferenceReplicasField } from './distributedInferenceReplicasField';
import { DistributedInferenceReplicasFieldRules } from './distributedInferenceReplicasFieldRules';
import { DistributedInferenceRestartPolicy } from './distributedInferenceRestartPolicy';
import { DistributedInferenceRestartPolicyOptions } from './distributedInferenceRestartPolicyOptions';
import { DistributedInferenceRestartPolicyOptionsOptionsInner } from './distributedInferenceRestartPolicyOptionsOptionsInner';
import { DistributedInferenceRestartPolicyRules } from './distributedInferenceRestartPolicyRules';
import { DistributedInferenceServingPort } from './distributedInferenceServingPort';
import { DistributedInferenceServingPortAccess } from './distributedInferenceServingPortAccess';
import { DistributedInferenceServingPortAccessAuthorizationTypeEnum } from './distributedInferenceServingPortAccessAuthorizationTypeEnum';
import { DistributedInferenceServingPortAuthorizationTypeOptions } from './distributedInferenceServingPortAuthorizationTypeOptions';
import { DistributedInferenceServingPortAuthorizationTypeOptionsOptionsInner } from './distributedInferenceServingPortAuthorizationTypeOptionsOptionsInner';
import { DistributedInferenceServingPortAuthorizationTypeRules } from './distributedInferenceServingPortAuthorizationTypeRules';
import { DistributedInferenceServingPortContainerAndProtocol } from './distributedInferenceServingPortContainerAndProtocol';
import { DistributedInferenceServingPortProtocol } from './distributedInferenceServingPortProtocol';
import { DistributedInferenceServingPortRules } from './distributedInferenceServingPortRules';
import { DistributedInferenceServingPortRulesProperties } from './distributedInferenceServingPortRulesProperties';
import { DistributedInferenceSpec } from './distributedInferenceSpec';
import { DistributedInferenceSpecSpec } from './distributedInferenceSpecSpec';
import { DistributedInferenceStartupPolicy } from './distributedInferenceStartupPolicy';
import { DistributedInferenceStartupPolicyField } from './distributedInferenceStartupPolicyField';
import { DistributedInferenceStartupPolicyFieldRules } from './distributedInferenceStartupPolicyFieldRules';
import { DistributedInferenceStartupPolicyFieldRulesStartupPolicy } from './distributedInferenceStartupPolicyFieldRulesStartupPolicy';
import { DistributedInferenceStartupPolicyOptions } from './distributedInferenceStartupPolicyOptions';
import { DistributedInferenceTemplate } from './distributedInferenceTemplate';
import { DistributedInferenceTemplateAllOf } from './distributedInferenceTemplateAllOf';
import { DistributedInferenceTemplateAllOf1 } from './distributedInferenceTemplateAllOf1';
import { DistributedInferenceTemplateCreationRequest } from './distributedInferenceTemplateCreationRequest';
import { DistributedInferenceTemplatePatchRequest } from './distributedInferenceTemplatePatchRequest';
import { DistributedInferenceTemplateSpec } from './distributedInferenceTemplateSpec';
import { DistributedInferenceTemplateUpdateRequest } from './distributedInferenceTemplateUpdateRequest';
import { DistributedInferenceTemplatesListResponse } from './distributedInferenceTemplatesListResponse';
import { DistributedInferenceTemplatesListResponseItem } from './distributedInferenceTemplatesListResponseItem';
import { DistributedInferenceWorkerDefaultsFields } from './distributedInferenceWorkerDefaultsFields';
import { DistributedInferenceWorkerDefaultsFieldsWorker } from './distributedInferenceWorkerDefaultsFieldsWorker';
import { DistributedInferenceWorkerRulesFields } from './distributedInferenceWorkerRulesFields';
import { DistributedInferenceWorkerRulesFieldsWorker } from './distributedInferenceWorkerRulesFieldsWorker';
import { DistributedInferenceWorkerSpecFields } from './distributedInferenceWorkerSpecFields';
import { DistributedInferenceWorkerSpecFieldsWorker } from './distributedInferenceWorkerSpecFieldsWorker';
import { DistributedInferenceWorkerTemplateSpecFields } from './distributedInferenceWorkerTemplateSpecFields';
import { DistributedInferenceWorkerTemplateSpecFieldsWorker } from './distributedInferenceWorkerTemplateSpecFieldsWorker';
import { DistributedInferenceWorkersField } from './distributedInferenceWorkersField';
import { DistributedInferenceWorkersFieldRules } from './distributedInferenceWorkersFieldRules';
import { DistributedSpec } from './distributedSpec';
import { DistributedSpecSpec } from './distributedSpecSpec';
import { DistributedSpecSpecAllOf } from './distributedSpecSpecAllOf';
import { DistributedTemplate } from './distributedTemplate';
import { DistributedTemplateAllOf } from './distributedTemplateAllOf';
import { DistributedTemplateCreationRequest } from './distributedTemplateCreationRequest';
import { DistributedTemplateCreationRequestAllOf } from './distributedTemplateCreationRequestAllOf';
import { DistributedTemplateCreationRequestAllOf1 } from './distributedTemplateCreationRequestAllOf1';
import { DistributedTemplatePatchRequest } from './distributedTemplatePatchRequest';
import { DistributedTemplateSpec } from './distributedTemplateSpec';
import { DistributedTemplateSpecAllOf } from './distributedTemplateSpecAllOf';
import { DistributedTemplateUpdateRequest } from './distributedTemplateUpdateRequest';
import { DistributedTemplatesListResponse } from './distributedTemplatesListResponse';
import { DistributedTemplatesListResponseItem } from './distributedTemplatesListResponseItem';
import { DistributedWorkload } from './distributedWorkload';
import { ElementMetricType } from './elementMetricType';
import { EmptyDir } from './emptyDir';
import { EmptyDirInstance } from './emptyDirInstance';
import { EmptyDirRules } from './emptyDirRules';
import { EmptyDirsDefaults } from './emptyDirsDefaults';
import { EmptyDirsRules } from './emptyDirsRules';
import { EnableHybridSubmissionSetting } from './enableHybridSubmissionSetting';
import { EnablePoliciesSetting } from './enablePoliciesSetting';
import { EnableTemplatesV2Setting } from './enableTemplatesV2Setting';
import { Endpoint } from './endpoint';
import { EndpointSync } from './endpointSync';
import { EndpointSyncAllOf } from './endpointSyncAllOf';
import { EndpointSyncAllOfSourceRef } from './endpointSyncAllOfSourceRef';
import { Environment } from './environment';
import { EnvironmentVariable } from './environmentVariable';
import { EnvironmentVariableConfigMap } from './environmentVariableConfigMap';
import { EnvironmentVariablePodFieldReference } from './environmentVariablePodFieldReference';
import { EnvironmentVariableSecret } from './environmentVariableSecret';
import { EnvironmentVariableUserCredential } from './environmentVariableUserCredential';
import { EnvironmentVariablesDefaults } from './environmentVariablesDefaults';
import { Event } from './event';
import { EventComputedFields } from './eventComputedFields';
import { EventSyncFields } from './eventSyncFields';
import { EventsSyncRequest } from './eventsSyncRequest';
import { ExcludeField } from './excludeField';
import { ExposedUrl } from './exposedUrl';
import { ExposedUrlRules } from './exposedUrlRules';
import { ExposedUrlsDefaults } from './exposedUrlsDefaults';
import { ExposedUrlsRules } from './exposedUrlsRules';
import { ExtendedResource } from './extendedResource';
import { ExtendedResourceRules } from './extendedResourceRules';
import { ExtendedResourcesDefaults } from './extendedResourcesDefaults';
import { ExtendedResourcesRules } from './extendedResourcesRules';
import { FieldsToClear } from './fieldsToClear';
import { GetApplicationSubmission } from './getApplicationSubmission';
import { GetAvailablePrioritiesAndCategories200Response } from './getAvailablePrioritiesAndCategories200Response';
import { GetEnableHybridSubmission } from './getEnableHybridSubmission';
import { GetEnablePolicies } from './getEnablePolicies';
import { GetEnableSettingsValue } from './getEnableSettingsValue';
import { GetEnableTemplatesV2 } from './getEnableTemplatesV2';
import { GetGpuOverProvisioning } from './getGpuOverProvisioning';
import { GetInferenceWorkloadRevisions200Response } from './getInferenceWorkloadRevisions200Response';
import { GetRevisionPods200Response } from './getRevisionPods200Response';
import { GetWorkloadClusterPriorities200Response } from './getWorkloadClusterPriorities200Response';
import { GetWorkloadEndpoints200Response } from './getWorkloadEndpoints200Response';
import { GetWorkloadEvents200Response } from './getWorkloadEvents200Response';
import { GetWorkloadHistory200Response } from './getWorkloadHistory200Response';
import { GetWorkloadPriorities200Response } from './getWorkloadPriorities200Response';
import { GetWorkloadSubmissionV2Enabled } from './getWorkloadSubmissionV2Enabled';
import { GetWorkloads200Response } from './getWorkloads200Response';
import { GetWorkloadsSetting } from './getWorkloadsSetting';
import { GitAuthenticationMethod } from './gitAuthenticationMethod';
import { GitCommon } from './gitCommon';
import { GitInstance } from './gitInstance';
import { GitInstanceAllOf } from './gitInstanceAllOf';
import { GitPassword } from './gitPassword';
import { GitRules } from './gitRules';
import { GitSecretRef } from './gitSecretRef';
import { GitsDefaults } from './gitsDefaults';
import { GitsRules } from './gitsRules';
import { GpuOverProvisioningSetting } from './gpuOverProvisioningSetting';
import { GpuRequestOptions } from './gpuRequestOptions';
import { GpuRequestOptionsOptionsInner } from './gpuRequestOptionsOptionsInner';
import { GpuRequestRules } from './gpuRequestRules';
import { GpuRequestType } from './gpuRequestType';
import { HistogramSeries } from './histogramSeries';
import { HistogramValue } from './histogramValue';
import { HistoryRecord } from './historyRecord';
import { HistoryRecordMeta } from './historyRecordMeta';
import { HistoryRecordSpec } from './historyRecordSpec';
import { HistoryRecordType } from './historyRecordType';
import { HostPath } from './hostPath';
import { HostPathInstance } from './hostPathInstance';
import { HostPathMountPropagation } from './hostPathMountPropagation';
import { HostPathMountPropagationOptions } from './hostPathMountPropagationOptions';
import { HostPathMountPropagationOptionsOptionsInner } from './hostPathMountPropagationOptionsOptionsInner';
import { HostPathMountPropagationRules } from './hostPathMountPropagationRules';
import { HostPathRules } from './hostPathRules';
import { HostPathsDefaults } from './hostPathsDefaults';
import { HostPathsRules } from './hostPathsRules';
import { HttpResponse } from './httpResponse';
import { ImagePullPolicy } from './imagePullPolicy';
import { ImagePullPolicyOptions } from './imagePullPolicyOptions';
import { ImagePullPolicyOptionsOptionsInner } from './imagePullPolicyOptionsOptionsInner';
import { ImagePullPolicyRules } from './imagePullPolicyRules';
import { ImagePullSecret } from './imagePullSecret';
import { ImagePullSecretsDefaults } from './imagePullSecretsDefaults';
import { ImagePullSecretsRules } from './imagePullSecretsRules';
import { ImagePullSecretsRulesAttributes } from './imagePullSecretsRulesAttributes';
import { Inference } from './inference';
import { InferenceCreationRequest } from './inferenceCreationRequest';
import { InferenceFields } from './inferenceFields';
import { InferenceFieldsRules } from './inferenceFieldsRules';
import { InferencePodMetricType } from './inferencePodMetricType';
import { InferenceSpec } from './inferenceSpec';
import { InferenceSpecSpec } from './inferenceSpecSpec';
import { InferenceSpecSpecAllOf } from './inferenceSpecSpecAllOf';
import { InferenceSpecSpecAllOfSecurity } from './inferenceSpecSpecAllOfSecurity';
import { InferenceSpecSpecAllOfStorage } from './inferenceSpecSpecAllOfStorage';
import { InferenceTemplate } from './inferenceTemplate';
import { InferenceTemplateAllOf } from './inferenceTemplateAllOf';
import { InferenceTemplateAllOf1 } from './inferenceTemplateAllOf1';
import { InferenceTemplateCreationRequest } from './inferenceTemplateCreationRequest';
import { InferenceTemplatePatchRequest } from './inferenceTemplatePatchRequest';
import { InferenceTemplateSpec } from './inferenceTemplateSpec';
import { InferenceTemplateSpecAllOf } from './inferenceTemplateSpecAllOf';
import { InferenceTemplateUpdateRequest } from './inferenceTemplateUpdateRequest';
import { InferenceTemplatesListResponse } from './inferenceTemplatesListResponse';
import { InferenceTemplatesListResponseItem } from './inferenceTemplatesListResponseItem';
import { InferenceUpdateRequest } from './inferenceUpdateRequest';
import { InferenceUpdateSpec } from './inferenceUpdateSpec';
import { InferenceUpdateSpecAutoscaling } from './inferenceUpdateSpecAutoscaling';
import { InferenceUpdateSpecServingConfiguration } from './inferenceUpdateSpecServingConfiguration';
import { InferenceUpdateSpecSpec } from './inferenceUpdateSpecSpec';
import { InferenceUpdateSpecSpecAllOf } from './inferenceUpdateSpecSpecAllOf';
import { InferenceWorkloadMetricType } from './inferenceWorkloadMetricType';
import { InferencesMaxReplicasField } from './inferencesMaxReplicasField';
import { InferencesMinReplicasField } from './inferencesMinReplicasField';
import { InitialReplicasField } from './initialReplicasField';
import { InitializationTimeoutField } from './initializationTimeoutField';
import { InstancesRules } from './instancesRules';
import { IntegerRules } from './integerRules';
import { IntegerRulesOptional } from './integerRulesOptional';
import { InternalEnrichOrgUnitRankingRequest } from './internalEnrichOrgUnitRankingRequest';
import { InternalEnrichOrgUnitRankingResponse } from './internalEnrichOrgUnitRankingResponse';
import { InternalPodsSyncRequest } from './internalPodsSyncRequest';
import { InternalSyncPods207Response } from './internalSyncPods207Response';
import { InternalSyncWorkloads207Response } from './internalSyncWorkloads207Response';
import { InternalWorkloadsSyncRequest } from './internalWorkloadsSyncRequest';
import { InvolvedObject } from './involvedObject';
import { ItemRules } from './itemRules';
import { Label } from './label';
import { LabelsDefaults } from './labelsDefaults';
import { ListApplications200Response } from './listApplications200Response';
import { ListCategories200Response } from './listCategories200Response';
import { ListPods200Response } from './listPods200Response';
import { ListWorkloadTypes200Response } from './listWorkloadTypes200Response';
import { MasterSpec } from './masterSpec';
import { MasterSpecSecurity } from './masterSpecSecurity';
import { MasterSpecStorage } from './masterSpecStorage';
import { MasterTemplateSpec } from './masterTemplateSpec';
import { MatchExpression } from './matchExpression';
import { MatchExpressionOperator } from './matchExpressionOperator';
import { MeasurementResponse } from './measurementResponse';
import { MeasurementResponseValuesInner } from './measurementResponseValuesInner';
import { MetricThresholdField } from './metricThresholdField';
import { MetricThresholdPercentageField } from './metricThresholdPercentageField';
import { MetricsCompositeResponse } from './metricsCompositeResponse';
import { MetricsResponse } from './metricsResponse';
import { ModelError } from './modelError';
import { MpiLauncherCreationPolicy } from './mpiLauncherCreationPolicy';
import { MpiLauncherCreationPolicyOptions } from './mpiLauncherCreationPolicyOptions';
import { MpiLauncherCreationPolicyOptionsOptionsInner } from './mpiLauncherCreationPolicyOptionsOptionsInner';
import { MpiLauncherCreationPolicyRules } from './mpiLauncherCreationPolicyRules';
import { Nfs } from './nfs';
import { NfsInstance } from './nfsInstance';
import { NfsRules } from './nfsRules';
import { NfssDefaults } from './nfssDefaults';
import { NfssRules } from './nfssRules';
import { NodeAffinityRequired } from './nodeAffinityRequired';
import { NodeSelectorTerm } from './nodeSelectorTerm';
import { NumberRules } from './numberRules';
import { OrgType } from './orgType';
import { OrgUnitRankingItem } from './orgUnitRankingItem';
import { PaginationInfo } from './paginationInfo';
import { PendingSchedulingMessage } from './pendingSchedulingMessage';
import { Phase } from './phase';
import { PhaseReason } from './phaseReason';
import { PhaseUpdate } from './phaseUpdate';
import { Pod } from './pod';
import { PodAffinity } from './podAffinity';
import { PodAffinityRules } from './podAffinityRules';
import { PodAffinityType } from './podAffinityType';
import { PodAffinityTypeOptions } from './podAffinityTypeOptions';
import { PodAffinityTypeOptionsOptionsInner } from './podAffinityTypeOptionsOptionsInner';
import { PodAffinityTypeRules } from './podAffinityTypeRules';
import { PodComputedFields } from './podComputedFields';
import { PodElement } from './podElement';
import { PodGroupsSyncFields } from './podGroupsSyncFields';
import { PodGroupsSyncRequest } from './podGroupsSyncRequest';
import { PodMetricType } from './podMetricType';
import { PodNodeAffinity } from './podNodeAffinity';
import { PodNodeAffinityPreferred } from './podNodeAffinityPreferred';
import { PodNodeAffinityRequired } from './podNodeAffinityRequired';
import { PodPreferredNodeSelectorTerm } from './podPreferredNodeSelectorTerm';
import { PodRequestResources } from './podRequestResources';
import { PodSortFilterFields } from './podSortFilterFields';
import { PodSyncFields } from './podSyncFields';
import { PodToleration } from './podToleration';
import { PodVerbosity } from './podVerbosity';
import { PodVerbosityFields } from './podVerbosityFields';
import { PodsSyncRequest } from './podsSyncRequest';
import { PolicyRuleEnum } from './policyRuleEnum';
import { Port } from './port';
import { PortRules } from './portRules';
import { PortServiceType } from './portServiceType';
import { PortServiceTypeOptions } from './portServiceTypeOptions';
import { PortServiceTypeOptionsOptionsInner } from './portServiceTypeOptionsOptionsInner';
import { PortServiceTypeRules } from './portServiceTypeRules';
import { PortsDefaults } from './portsDefaults';
import { PortsRules } from './portsRules';
import { Preemptibility } from './preemptibility';
import { PreemptibilityOptions } from './preemptibilityOptions';
import { PreemptibilityOptionsOptionsInner } from './preemptibilityOptionsOptionsInner';
import { PreemptibilityRules } from './preemptibilityRules';
import { PrioritiesForStatus } from './prioritiesForStatus';
import { PrioritiesForStatusPrioritiesInner } from './prioritiesForStatusPrioritiesInner';
import { PrioritiesForStatusPrioritiesInnerAllOf } from './prioritiesForStatusPrioritiesInnerAllOf';
import { PrioritiesForStatusPrioritiesInnerAllOfStatus } from './prioritiesForStatusPrioritiesInnerAllOfStatus';
import { PriorityForCreate } from './priorityForCreate';
import { PriorityStatus } from './priorityStatus';
import { Probe } from './probe';
import { ProbeHandler } from './probeHandler';
import { ProbeHandlerHttpGet } from './probeHandlerHttpGet';
import { ProbeHandlerRules } from './probeHandlerRules';
import { ProbeHandlerRulesHttpGet } from './probeHandlerRulesHttpGet';
import { ProbeHandlerScheme } from './probeHandlerScheme';
import { ProbeRules } from './probeRules';
import { Probes } from './probes';
import { ProbesRules } from './probesRules';
import { Pvc } from './pvc';
import { PvcAccessModes } from './pvcAccessModes';
import { PvcAccessModesRules } from './pvcAccessModesRules';
import { PvcAddedAttrValue } from './pvcAddedAttrValue';
import { PvcFieldsNonUpdatable } from './pvcFieldsNonUpdatable';
import { PvcFieldsUpdatable } from './pvcFieldsUpdatable';
import { PvcInstance } from './pvcInstance';
import { PvcRules } from './pvcRules';
import { PvcVolumeMode } from './pvcVolumeMode';
import { PvcsDefaults } from './pvcsDefaults';
import { PvcsRules } from './pvcsRules';
import { QuantityRules } from './quantityRules';
import { QueueResources } from './queueResources';
import { RelatedUrl } from './relatedUrl';
import { RelatedUrlRules } from './relatedUrlRules';
import { RelatedUrlsDefaults } from './relatedUrlsDefaults';
import { RelatedUrlsRules } from './relatedUrlsRules';
import { ReplicaType } from './replicaType';
import { RequestInfo } from './requestInfo';
import { RequestResourceCores } from './requestResourceCores';
import { RequestResourceQuantity } from './requestResourceQuantity';
import { RequestedPods } from './requestedPods';
import { ResourceRef } from './resourceRef';
import { ResourceRefKind } from './resourceRefKind';
import { ResponseInfo } from './responseInfo';
import { RestartPolicy } from './restartPolicy';
import { Revision } from './revision';
import { RevisionSyncFields } from './revisionSyncFields';
import { RevisionsSyncRequest } from './revisionsSyncRequest';
import { RunnableInfo } from './runnableInfo';
import { S3AccessKey } from './s3AccessKey';
import { S3Common } from './s3Common';
import { S3Instance } from './s3Instance';
import { S3Rules } from './s3Rules';
import { S3sDefaults } from './s3sDefaults';
import { S3sRules } from './s3sRules';
import { ScaleDownDelayField } from './scaleDownDelayField';
import { ScaleToZeroRetentionField } from './scaleToZeroRetentionField';
import { SchedulingCondition } from './schedulingCondition';
import { SchedulingConditionQueueDetails } from './schedulingConditionQueueDetails';
import { ScopeType } from './scopeType';
import { SeccompProfileType } from './seccompProfileType';
import { SeccompProfileTypeOptions } from './seccompProfileTypeOptions';
import { SeccompProfileTypeOptionsOptionsInner } from './seccompProfileTypeOptionsOptionsInner';
import { SeccompProfileTypeRules } from './seccompProfileTypeRules';
import { Secret } from './secret';
import { SecretFieldsNonUpdatable } from './secretFieldsNonUpdatable';
import { SecretFieldsUpdatable } from './secretFieldsUpdatable';
import { SecretInstance } from './secretInstance';
import { SecretRules } from './secretRules';
import { SecretsDefaults } from './secretsDefaults';
import { SecretsRules } from './secretsRules';
import { ServingConfiguration } from './servingConfiguration';
import { ServingConfigurationField } from './servingConfigurationField';
import { ServingConfigurationRules } from './servingConfigurationRules';
import { ServingPort } from './servingPort';
import { ServingPortAccess } from './servingPortAccess';
import { ServingPortAccessAuthorizationTypeEnum } from './servingPortAccessAuthorizationTypeEnum';
import { ServingPortContainerAndProtocol } from './servingPortContainerAndProtocol';
import { ServingPortProtocol } from './servingPortProtocol';
import { Source } from './source';
import { SourceApi } from './sourceApi';
import { SourceOfRule } from './sourceOfRule';
import { StorageInstanceName } from './storageInstanceName';
import { StringOption } from './stringOption';
import { StringRules } from './stringRules';
import { StringRulesCommon } from './stringRulesCommon';
import { StringRulesOptions } from './stringRulesOptions';
import { SubmissionError } from './submissionError';
import { SubmissionErrorV2 } from './submissionErrorV2';
import { SubmitWithAssetIds } from './submitWithAssetIds';
import { SubmitWithTemplateId } from './submitWithTemplateId';
import { SupersetDefaults } from './supersetDefaults';
import { SupersetDefaultsAllOf } from './supersetDefaultsAllOf';
import { SupersetDefaultsAllOfCompute } from './supersetDefaultsAllOfCompute';
import { SupersetDefaultsAllOfSecurity } from './supersetDefaultsAllOfSecurity';
import { SupersetDefaultsAllOfStorage } from './supersetDefaultsAllOfStorage';
import { SupersetRules } from './supersetRules';
import { SupersetRulesAllOf } from './supersetRulesAllOf';
import { SupersetRulesAllOfCompute } from './supersetRulesAllOfCompute';
import { SupersetRulesAllOfSecurity } from './supersetRulesAllOfSecurity';
import { SupersetRulesAllOfStorage } from './supersetRulesAllOfStorage';
import { SupersetServingPort } from './supersetServingPort';
import { SupersetServingPortRules } from './supersetServingPortRules';
import { SupersetSpec } from './supersetSpec';
import { SupersetSpecAllOf } from './supersetSpecAllOf';
import { SupersetTemplate } from './supersetTemplate';
import { SupersetTemplateAllOf } from './supersetTemplateAllOf';
import { SupersetWorkload } from './supersetWorkload';
import { SupersetWorkloadAllOf } from './supersetWorkloadAllOf';
import { SyncJobsBackwardsRequest } from './syncJobsBackwardsRequest';
import { SyncPodGroups207Response } from './syncPodGroups207Response';
import { SyncPodsBackwards200Response } from './syncPodsBackwards200Response';
import { SyncPriority } from './syncPriority';
import { SyncPriorityAllOf } from './syncPriorityAllOf';
import { SyncRevisions207Response } from './syncRevisions207Response';
import { SystemInfo } from './systemInfo';
import { TelemetryResponse } from './telemetryResponse';
import { TelemetryResponseValuesInner } from './telemetryResponseValuesInner';
import { TelemetryResponseValuesInnerGroupsInner } from './telemetryResponseValuesInnerGroupsInner';
import { TemplateAssets } from './templateAssets';
import { TemplateAssetsIds } from './templateAssetsIds';
import { TemplateAssetsIdsOptional } from './templateAssetsIdsOptional';
import { TemplateComplianceInfo } from './templateComplianceInfo';
import { TemplateCreationMeta } from './templateCreationMeta';
import { TemplateCreationMetaFields } from './templateCreationMetaFields';
import { TemplateMasterSection } from './templateMasterSection';
import { TemplateMeta } from './templateMeta';
import { TemplateMetaFields } from './templateMetaFields';
import { TemplatePatchMeta } from './templatePatchMeta';
import { TemplatePatchMetaFields } from './templatePatchMetaFields';
import { TemplateUpdateMeta } from './templateUpdateMeta';
import { TemplateUpdateMetaFields } from './templateUpdateMetaFields';
import { TemplatesListResponse } from './templatesListResponse';
import { Toleration } from './toleration';
import { TolerationEffect } from './tolerationEffect';
import { TolerationEffectOptions } from './tolerationEffectOptions';
import { TolerationEffectOptionsOptionsInner } from './tolerationEffectOptionsOptionsInner';
import { TolerationEffectRules } from './tolerationEffectRules';
import { TolerationOperator } from './tolerationOperator';
import { TolerationOperatorOptions } from './tolerationOperatorOptions';
import { TolerationOperatorOptionsOptionsInner } from './tolerationOperatorOptionsOptionsInner';
import { TolerationOperatorRules } from './tolerationOperatorRules';
import { TolerationRules } from './tolerationRules';
import { TolerationsDefaults } from './tolerationsDefaults';
import { TolerationsRules } from './tolerationsRules';
import { Training } from './training';
import { TrainingCreationRequest } from './trainingCreationRequest';
import { TrainingSpec } from './trainingSpec';
import { TrainingSpecSpec } from './trainingSpecSpec';
import { TrainingSpecSpecAllOf } from './trainingSpecSpecAllOf';
import { TrainingTemplate } from './trainingTemplate';
import { TrainingTemplateAllOf } from './trainingTemplateAllOf';
import { TrainingTemplateAllOf1 } from './trainingTemplateAllOf1';
import { TrainingTemplateCreationRequest } from './trainingTemplateCreationRequest';
import { TrainingTemplatePatchRequest } from './trainingTemplatePatchRequest';
import { TrainingTemplateSpec } from './trainingTemplateSpec';
import { TrainingTemplateSpecAllOf } from './trainingTemplateSpecAllOf';
import { TrainingTemplateUpdateRequest } from './trainingTemplateUpdateRequest';
import { TrainingTemplatesListResponse } from './trainingTemplatesListResponse';
import { TrainingTemplatesListResponseItem } from './trainingTemplatesListResponseItem';
import { UidGidSource } from './uidGidSource';
import { UidGidSourceOptions } from './uidGidSourceOptions';
import { UidGidSourceOptionsOptionsInner } from './uidGidSourceOptionsOptionsInner';
import { UidGidSourceRules } from './uidGidSourceRules';
import { UnifiedPolicyInfo } from './unifiedPolicyInfo';
import { UnifiedPolicyInfoPerReplicaType } from './unifiedPolicyInfoPerReplicaType';
import { UnifiedPolicyInfoPerReplicaTypeDefaults } from './unifiedPolicyInfoPerReplicaTypeDefaults';
import { UnifiedPolicyInfoPerReplicaTypeRules } from './unifiedPolicyInfoPerReplicaTypeRules';
import { UnifiedProjectInfo } from './unifiedProjectInfo';
import { UnifiedSpec } from './unifiedSpec';
import { UnifiedSpecMaster } from './unifiedSpecMaster';
import { UnifiedSpecWorker } from './unifiedSpecWorker';
import { UnifiedWorkload } from './unifiedWorkload';
import { UnifiedWorkloadRequest } from './unifiedWorkloadRequest';
import { UpdateNodePodsRequest } from './updateNodePodsRequest';
import { UpdateRequest } from './updateRequest';
import { UpdateSpec } from './updateSpec';
import { UpdateSpecSpec } from './updateSpecSpec';
import { UpdateSpecSpecAllOf } from './updateSpecSpecAllOf';
import { UpdateWorkloadsSettingByKeyRequest } from './updateWorkloadsSettingByKeyRequest';
import { Workload } from './workload';
import { WorkloadAllocatedResources } from './workloadAllocatedResources';
import { WorkloadBatch } from './workloadBatch';
import { WorkloadCategory } from './workloadCategory';
import { WorkloadChildrenIdsInner } from './workloadChildrenIdsInner';
import { WorkloadCreationFields } from './workloadCreationFields';
import { WorkloadCreationMeta } from './workloadCreationMeta';
import { WorkloadDesiredPhase } from './workloadDesiredPhase';
import { WorkloadDetailed } from './workloadDetailed';
import { WorkloadDetailedAllOf } from './workloadDetailedAllOf';
import { WorkloadElement } from './workloadElement';
import { WorkloadElementKubernetesRef } from './workloadElementKubernetesRef';
import { WorkloadElementKubernetesRefGvk } from './workloadElementKubernetesRefGvk';
import { WorkloadElementResponse } from './workloadElementResponse';
import { WorkloadElementResponseAllOf } from './workloadElementResponseAllOf';
import { WorkloadElementTopology } from './workloadElementTopology';
import { WorkloadElementsResponse } from './workloadElementsResponse';
import { WorkloadKubernetesRef } from './workloadKubernetesRef';
import { WorkloadMeta } from './workloadMeta';
import { WorkloadMetricType } from './workloadMetricType';
import { WorkloadNetworkStatus } from './workloadNetworkStatus';
import { WorkloadNetworkStatusServicesInner } from './workloadNetworkStatusServicesInner';
import { WorkloadNetworkStatusServicesInnerLoadBalancer } from './workloadNetworkStatusServicesInnerLoadBalancer';
import { WorkloadNetworkStatusServicesInnerLoadBalancerIngressInner } from './workloadNetworkStatusServicesInnerLoadBalancerIngressInner';
import { WorkloadNetworkStatusServicesInnerPortsInner } from './workloadNetworkStatusServicesInnerPortsInner';
import { WorkloadNetworkStatusServicesInnerUrlsInner } from './workloadNetworkStatusServicesInnerUrlsInner';
import { WorkloadNodeAffinity } from './workloadNodeAffinity';
import { WorkloadPatchFields } from './workloadPatchFields';
import { WorkloadPodsSyncRequest } from './workloadPodsSyncRequest';
import { WorkloadPriority } from './workloadPriority';
import { WorkloadPriorityAllOf } from './workloadPriorityAllOf';
import { WorkloadRequestResources } from './workloadRequestResources';
import { WorkloadSortFilterFields } from './workloadSortFilterFields';
import { WorkloadSubmissionV2EnabledSetting } from './workloadSubmissionV2EnabledSetting';
import { WorkloadSyncFields } from './workloadSyncFields';
import { WorkloadSyncFieldsTopology } from './workloadSyncFieldsTopology';
import { WorkloadTelemetryType } from './workloadTelemetryType';
import { WorkloadTopology } from './workloadTopology';
import { WorkloadTopologyActualPlacement } from './workloadTopologyActualPlacement';
import { WorkloadType } from './workloadType';
import { WorkloadTypeAdditionalCreateFields } from './workloadTypeAdditionalCreateFields';
import { WorkloadTypeAdditionalResponseFields } from './workloadTypeAdditionalResponseFields';
import { WorkloadTypeConfig } from './workloadTypeConfig';
import { WorkloadTypeConfigAllOf } from './workloadTypeConfigAllOf';
import { WorkloadTypeCreateFields } from './workloadTypeCreateFields';
import { WorkloadTypeCreateFieldsAllOf } from './workloadTypeCreateFieldsAllOf';
import { WorkloadTypeCreateUpdateInnerFields } from './workloadTypeCreateUpdateInnerFields';
import { WorkloadTypeForSync } from './workloadTypeForSync';
import { WorkloadTypeForSyncAllOf } from './workloadTypeForSyncAllOf';
import { WorkloadTypeKartasWithVersionInner } from './workloadTypeKartasWithVersionInner';
import { WorkloadTypePreemptibility } from './workloadTypePreemptibility';
import { WorkloadTypeResourceInterfacesWithVersionInner } from './workloadTypeResourceInterfacesWithVersionInner';
import { WorkloadTypeStatus } from './workloadTypeStatus';
import { WorkloadTypeStatusAllOf } from './workloadTypeStatusAllOf';
import { WorkloadTypeStatusFields } from './workloadTypeStatusFields';
import { WorkloadTypeUpdateFields } from './workloadTypeUpdateFields';
import { WorkloadTypeUpdateFieldsAllOf } from './workloadTypeUpdateFieldsAllOf';
import { WorkloadTypesForSync } from './workloadTypesForSync';
import { WorkloadsExtendedResource } from './workloadsExtendedResource';
import { WorkloadsSettingsGetResponse } from './workloadsSettingsGetResponse';
import { WorkloadsSettingsKeyEnum } from './workloadsSettingsKeyEnum';
import { WorkloadsSyncInfo } from './workloadsSyncInfo';
import { WorkloadsSyncRequest } from './workloadsSyncRequest';
import { WorkloadsSyncResponse } from './workloadsSyncResponse';
import { Workspace } from './workspace';
import { WorkspaceCreationRequest } from './workspaceCreationRequest';
import { WorkspaceSpec } from './workspaceSpec';
import { WorkspaceSpecSpec } from './workspaceSpecSpec';
import { WorkspaceSpecSpecAllOf } from './workspaceSpecSpecAllOf';
import { WorkspaceSpecSpecAllOfCompute } from './workspaceSpecSpecAllOfCompute';
import { WorkspaceSpecSpecAllOfSecurity } from './workspaceSpecSpecAllOfSecurity';
import { WorkspaceSpecSpecAllOfStorage } from './workspaceSpecSpecAllOfStorage';
import { WorkspaceTemplate } from './workspaceTemplate';
import { WorkspaceTemplateAllOf } from './workspaceTemplateAllOf';
import { WorkspaceTemplateAllOf1 } from './workspaceTemplateAllOf1';
import { WorkspaceTemplateCreationRequest } from './workspaceTemplateCreationRequest';
import { WorkspaceTemplatePatchRequest } from './workspaceTemplatePatchRequest';
import { WorkspaceTemplateSpec } from './workspaceTemplateSpec';
import { WorkspaceTemplateSpecAllOf } from './workspaceTemplateSpecAllOf';
import { WorkspaceTemplateUpdateRequest } from './workspaceTemplateUpdateRequest';
import { WorkspaceTemplatesListResponse } from './workspaceTemplatesListResponse';
import { WorkspaceTemplatesListResponseItem } from './workspaceTemplatesListResponseItem';

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
        "AIApplicationSyncAction": AIApplicationSyncAction,
        "AIApplicationVerbosity": AIApplicationVerbosity,
        "AIApplicationsSortFilterFields": AIApplicationsSortFilterFields,
        "Action": Action,
        "AssetKind": AssetKind,
        "AuthorizationType": AuthorizationType,
        "Capability": Capability,
        "CleanPodPolicy": CleanPodPolicy,
        "Connection.AuthorizationTypeEnum": Connection.AuthorizationTypeEnum,
        "DistributedFramework": DistributedFramework,
        "DistributedInferenceRestartPolicy": DistributedInferenceRestartPolicy,
        "DistributedInferenceServingPortAccessAuthorizationTypeEnum": DistributedInferenceServingPortAccessAuthorizationTypeEnum,
        "DistributedInferenceServingPortProtocol": DistributedInferenceServingPortProtocol,
        "DistributedInferenceStartupPolicy": DistributedInferenceStartupPolicy,
        "ElementMetricType": ElementMetricType,
        "GitAuthenticationMethod": GitAuthenticationMethod,
        "GpuRequestType": GpuRequestType,
        "HistoryRecordType": HistoryRecordType,
        "HostPathMountPropagation": HostPathMountPropagation,
        "ImagePullPolicy": ImagePullPolicy,
        "InferencePodMetricType": InferencePodMetricType,
        "InferenceWorkloadMetricType": InferenceWorkloadMetricType,
        "InternalEnrichOrgUnitRankingRequest.OrgUnitTypeEnum": InternalEnrichOrgUnitRankingRequest.OrgUnitTypeEnum,
        "MatchExpressionOperator": MatchExpressionOperator,
        "MpiLauncherCreationPolicy": MpiLauncherCreationPolicy,
        "OrgType": OrgType,
        "Phase": Phase,
        "PhaseReason": PhaseReason,
        "PodAffinityType": PodAffinityType,
        "PodMetricType": PodMetricType,
        "PodSortFilterFields": PodSortFilterFields,
        "PodVerbosity": PodVerbosity,
        "PolicyRuleEnum": PolicyRuleEnum,
        "PortServiceType": PortServiceType,
        "Preemptibility": Preemptibility,
        "ProbeHandlerScheme": ProbeHandlerScheme,
        "PvcVolumeMode": PvcVolumeMode,
        "ReplicaType": ReplicaType,
        "ResourceRefKind": ResourceRefKind,
        "RestartPolicy": RestartPolicy,
        "ScopeType": ScopeType,
        "SeccompProfileType": SeccompProfileType,
        "ServingPortAccessAuthorizationTypeEnum": ServingPortAccessAuthorizationTypeEnum,
        "ServingPortProtocol": ServingPortProtocol,
        "Source": Source,
        "SourceApi": SourceApi,
        "TolerationEffect": TolerationEffect,
        "TolerationOperator": TolerationOperator,
        "UidGidSource": UidGidSource,
        "WorkloadDesiredPhase": WorkloadDesiredPhase,
        "WorkloadMetricType": WorkloadMetricType,
        "WorkloadSortFilterFields": WorkloadSortFilterFields,
        "WorkloadTelemetryType": WorkloadTelemetryType,
        "WorkloadType": WorkloadType,
        "WorkloadTypePreemptibility": WorkloadTypePreemptibility,
        "WorkloadsSettingsKeyEnum": WorkloadsSettingsKeyEnum,
}

let typeMap: {[index: string]: any} = {
    "AIApplication": AIApplication,
    "AIApplicationAllocatedResources": AIApplicationAllocatedResources,
    "AIApplicationAuth": AIApplicationAuth,
    "AIApplicationChartUploadResponse": AIApplicationChartUploadResponse,
    "AIApplicationCreationRequest": AIApplicationCreationRequest,
    "AIApplicationForSync": AIApplicationForSync,
    "AIApplicationRequestResources": AIApplicationRequestResources,
    "AIApplicationResource": AIApplicationResource,
    "AIApplicationSpec": AIApplicationSpec,
    "AIApplicationSpecChartRef": AIApplicationSpecChartRef,
    "AIApplicationSyncActionInfo": AIApplicationSyncActionInfo,
    "AIApplicationSyncInfo": AIApplicationSyncInfo,
    "AIApplicationSyncItem": AIApplicationSyncItem,
    "AIApplicationSyncResource": AIApplicationSyncResource,
    "AIApplicationSyncResponse": AIApplicationSyncResponse,
    "AIApplicationsForSync": AIApplicationsForSync,
    "ActionsSupport": ActionsSupport,
    "ActivationReplicasField": ActivationReplicasField,
    "AiApplicationEndpoint": AiApplicationEndpoint,
    "AiApplicationEndpointAllOf": AiApplicationEndpointAllOf,
    "AiApplicationResourceConnection": AiApplicationResourceConnection,
    "AiApplicationResourceConnectionAllOf": AiApplicationResourceConnectionAllOf,
    "AllocatedResources": AllocatedResources,
    "Ancestors": Ancestors,
    "Annotation": Annotation,
    "AnnotationsDefaults": AnnotationsDefaults,
    "ApplicationSubmissionSetting": ApplicationSubmissionSetting,
    "ArbitraryRules": ArbitraryRules,
    "ArrayRules": ArrayRules,
    "AssetSubmissionIds": AssetSubmissionIds,
    "AssetsMapping": AssetsMapping,
    "AutoScaling": AutoScaling,
    "AutoScalingCommonFields": AutoScalingCommonFields,
    "AutoScalingField": AutoScalingField,
    "AutoScalingMetricField": AutoScalingMetricField,
    "AutoScalingMetricFields": AutoScalingMetricFields,
    "AutoScalingMetricOptions": AutoScalingMetricOptions,
    "AutoScalingMetricOptionsOptionsInner": AutoScalingMetricOptionsOptionsInner,
    "AutoScalingMetricRules": AutoScalingMetricRules,
    "AutoScalingRules": AutoScalingRules,
    "AutoScalingRulesInitializationTimeoutSeconds": AutoScalingRulesInitializationTimeoutSeconds,
    "BasePriority": BasePriority,
    "BatchResponse": BatchResponse,
    "BatchResponseFailureItem": BatchResponseFailureItem,
    "BooleanRules": BooleanRules,
    "Catalog": Catalog,
    "CatalogParams": CatalogParams,
    "CatalogParamsOrderByInner": CatalogParamsOrderByInner,
    "CatalogResourcesInner": CatalogResourcesInner,
    "ClaimInfo": ClaimInfo,
    "ClaimInfoRules": ClaimInfoRules,
    "ClaimInfoUpdatable": ClaimInfoUpdatable,
    "CleanPodPolicyOptions": CleanPodPolicyOptions,
    "CleanPodPolicyOptionsOptionsInner": CleanPodPolicyOptionsOptionsInner,
    "CleanPodPolicyRules": CleanPodPolicyRules,
    "ClusterPriority": ClusterPriority,
    "ClusterPriorityAllOf": ClusterPriorityAllOf,
    "ClusterSyncReport": ClusterSyncReport,
    "ClusterSyncReportUpdate": ClusterSyncReportUpdate,
    "ComplianceInfoReason": ComplianceInfoReason,
    "ComplianceIssues": ComplianceIssues,
    "ComplianceIssuesComplianceIssuesInner": ComplianceIssuesComplianceIssuesInner,
    "ComplianceIssuesV2": ComplianceIssuesV2,
    "ComplianceIssuesV2ComplianceIssuesInner": ComplianceIssuesV2ComplianceIssuesInner,
    "ConcurrencyHardLimitField": ConcurrencyHardLimitField,
    "Condition": Condition,
    "ConfigMap": ConfigMap,
    "ConfigMapInstance": ConfigMapInstance,
    "ConfigMapRules": ConfigMapRules,
    "ConfigMapsDefaults": ConfigMapsDefaults,
    "ConfigMapsRules": ConfigMapsRules,
    "Connection": Connection,
    "Container": Container,
    "CountResponse": CountResponse,
    "CountWorkloadElements200Response": CountWorkloadElements200Response,
    "CountWorkloads200Response": CountWorkloads200Response,
    "DataVolume": DataVolume,
    "DataVolumeInstance": DataVolumeInstance,
    "DataVolumeRules": DataVolumeRules,
    "DataVolumesDefaults": DataVolumesDefaults,
    "DataVolumesRules": DataVolumesRules,
    "Datasource": Datasource,
    "DefaultFromRule": DefaultFromRule,
    "DistributedCreationRequest": DistributedCreationRequest,
    "DistributedFrameworkOptions": DistributedFrameworkOptions,
    "DistributedFrameworkOptionsOptionsInner": DistributedFrameworkOptionsOptionsInner,
    "DistributedFrameworkRules": DistributedFrameworkRules,
    "DistributedInference": DistributedInference,
    "DistributedInferenceAssets": DistributedInferenceAssets,
    "DistributedInferenceAssetsAssets": DistributedInferenceAssetsAssets,
    "DistributedInferenceCommonDefaultsV2": DistributedInferenceCommonDefaultsV2,
    "DistributedInferenceCommonRulesV2": DistributedInferenceCommonRulesV2,
    "DistributedInferenceCommonRulesV2AllOf": DistributedInferenceCommonRulesV2AllOf,
    "DistributedInferenceCommonSpec": DistributedInferenceCommonSpec,
    "DistributedInferenceCommonSpecAllOf": DistributedInferenceCommonSpecAllOf,
    "DistributedInferenceCreationRequest": DistributedInferenceCreationRequest,
    "DistributedInferenceLeaderDefaultsFields": DistributedInferenceLeaderDefaultsFields,
    "DistributedInferenceLeaderDefaultsFieldsLeader": DistributedInferenceLeaderDefaultsFieldsLeader,
    "DistributedInferenceLeaderRulesFields": DistributedInferenceLeaderRulesFields,
    "DistributedInferenceLeaderRulesFieldsLeader": DistributedInferenceLeaderRulesFieldsLeader,
    "DistributedInferenceLeaderSpecFields": DistributedInferenceLeaderSpecFields,
    "DistributedInferenceLeaderSpecFieldsLeader": DistributedInferenceLeaderSpecFieldsLeader,
    "DistributedInferenceLeaderTemplateSpecFields": DistributedInferenceLeaderTemplateSpecFields,
    "DistributedInferenceLeaderTemplateSpecFieldsLeader": DistributedInferenceLeaderTemplateSpecFieldsLeader,
    "DistributedInferenceLeaderWorkerDefaultsV2": DistributedInferenceLeaderWorkerDefaultsV2,
    "DistributedInferenceLeaderWorkerDefaultsV2Security": DistributedInferenceLeaderWorkerDefaultsV2Security,
    "DistributedInferenceLeaderWorkerDefaultsV2Storage": DistributedInferenceLeaderWorkerDefaultsV2Storage,
    "DistributedInferenceLeaderWorkerRulesV2": DistributedInferenceLeaderWorkerRulesV2,
    "DistributedInferenceLeaderWorkerRulesV2Security": DistributedInferenceLeaderWorkerRulesV2Security,
    "DistributedInferenceLeaderWorkerRulesV2SecurityPrivileged": DistributedInferenceLeaderWorkerRulesV2SecurityPrivileged,
    "DistributedInferenceLeaderWorkerRulesV2Storage": DistributedInferenceLeaderWorkerRulesV2Storage,
    "DistributedInferenceLeaderWorkerSpec": DistributedInferenceLeaderWorkerSpec,
    "DistributedInferenceLeaderWorkerSpecCompute": DistributedInferenceLeaderWorkerSpecCompute,
    "DistributedInferenceLeaderWorkerSpecSecurity": DistributedInferenceLeaderWorkerSpecSecurity,
    "DistributedInferenceLeaderWorkerSpecStorage": DistributedInferenceLeaderWorkerSpecStorage,
    "DistributedInferenceLeaderWorkerTemplateSpec": DistributedInferenceLeaderWorkerTemplateSpec,
    "DistributedInferenceReplicasField": DistributedInferenceReplicasField,
    "DistributedInferenceReplicasFieldRules": DistributedInferenceReplicasFieldRules,
    "DistributedInferenceRestartPolicyOptions": DistributedInferenceRestartPolicyOptions,
    "DistributedInferenceRestartPolicyOptionsOptionsInner": DistributedInferenceRestartPolicyOptionsOptionsInner,
    "DistributedInferenceRestartPolicyRules": DistributedInferenceRestartPolicyRules,
    "DistributedInferenceServingPort": DistributedInferenceServingPort,
    "DistributedInferenceServingPortAccess": DistributedInferenceServingPortAccess,
    "DistributedInferenceServingPortAuthorizationTypeOptions": DistributedInferenceServingPortAuthorizationTypeOptions,
    "DistributedInferenceServingPortAuthorizationTypeOptionsOptionsInner": DistributedInferenceServingPortAuthorizationTypeOptionsOptionsInner,
    "DistributedInferenceServingPortAuthorizationTypeRules": DistributedInferenceServingPortAuthorizationTypeRules,
    "DistributedInferenceServingPortContainerAndProtocol": DistributedInferenceServingPortContainerAndProtocol,
    "DistributedInferenceServingPortRules": DistributedInferenceServingPortRules,
    "DistributedInferenceServingPortRulesProperties": DistributedInferenceServingPortRulesProperties,
    "DistributedInferenceSpec": DistributedInferenceSpec,
    "DistributedInferenceSpecSpec": DistributedInferenceSpecSpec,
    "DistributedInferenceStartupPolicyField": DistributedInferenceStartupPolicyField,
    "DistributedInferenceStartupPolicyFieldRules": DistributedInferenceStartupPolicyFieldRules,
    "DistributedInferenceStartupPolicyFieldRulesStartupPolicy": DistributedInferenceStartupPolicyFieldRulesStartupPolicy,
    "DistributedInferenceStartupPolicyOptions": DistributedInferenceStartupPolicyOptions,
    "DistributedInferenceTemplate": DistributedInferenceTemplate,
    "DistributedInferenceTemplateAllOf": DistributedInferenceTemplateAllOf,
    "DistributedInferenceTemplateAllOf1": DistributedInferenceTemplateAllOf1,
    "DistributedInferenceTemplateCreationRequest": DistributedInferenceTemplateCreationRequest,
    "DistributedInferenceTemplatePatchRequest": DistributedInferenceTemplatePatchRequest,
    "DistributedInferenceTemplateSpec": DistributedInferenceTemplateSpec,
    "DistributedInferenceTemplateUpdateRequest": DistributedInferenceTemplateUpdateRequest,
    "DistributedInferenceTemplatesListResponse": DistributedInferenceTemplatesListResponse,
    "DistributedInferenceTemplatesListResponseItem": DistributedInferenceTemplatesListResponseItem,
    "DistributedInferenceWorkerDefaultsFields": DistributedInferenceWorkerDefaultsFields,
    "DistributedInferenceWorkerDefaultsFieldsWorker": DistributedInferenceWorkerDefaultsFieldsWorker,
    "DistributedInferenceWorkerRulesFields": DistributedInferenceWorkerRulesFields,
    "DistributedInferenceWorkerRulesFieldsWorker": DistributedInferenceWorkerRulesFieldsWorker,
    "DistributedInferenceWorkerSpecFields": DistributedInferenceWorkerSpecFields,
    "DistributedInferenceWorkerSpecFieldsWorker": DistributedInferenceWorkerSpecFieldsWorker,
    "DistributedInferenceWorkerTemplateSpecFields": DistributedInferenceWorkerTemplateSpecFields,
    "DistributedInferenceWorkerTemplateSpecFieldsWorker": DistributedInferenceWorkerTemplateSpecFieldsWorker,
    "DistributedInferenceWorkersField": DistributedInferenceWorkersField,
    "DistributedInferenceWorkersFieldRules": DistributedInferenceWorkersFieldRules,
    "DistributedSpec": DistributedSpec,
    "DistributedSpecSpec": DistributedSpecSpec,
    "DistributedSpecSpecAllOf": DistributedSpecSpecAllOf,
    "DistributedTemplate": DistributedTemplate,
    "DistributedTemplateAllOf": DistributedTemplateAllOf,
    "DistributedTemplateCreationRequest": DistributedTemplateCreationRequest,
    "DistributedTemplateCreationRequestAllOf": DistributedTemplateCreationRequestAllOf,
    "DistributedTemplateCreationRequestAllOf1": DistributedTemplateCreationRequestAllOf1,
    "DistributedTemplatePatchRequest": DistributedTemplatePatchRequest,
    "DistributedTemplateSpec": DistributedTemplateSpec,
    "DistributedTemplateSpecAllOf": DistributedTemplateSpecAllOf,
    "DistributedTemplateUpdateRequest": DistributedTemplateUpdateRequest,
    "DistributedTemplatesListResponse": DistributedTemplatesListResponse,
    "DistributedTemplatesListResponseItem": DistributedTemplatesListResponseItem,
    "DistributedWorkload": DistributedWorkload,
    "EmptyDir": EmptyDir,
    "EmptyDirInstance": EmptyDirInstance,
    "EmptyDirRules": EmptyDirRules,
    "EmptyDirsDefaults": EmptyDirsDefaults,
    "EmptyDirsRules": EmptyDirsRules,
    "EnableHybridSubmissionSetting": EnableHybridSubmissionSetting,
    "EnablePoliciesSetting": EnablePoliciesSetting,
    "EnableTemplatesV2Setting": EnableTemplatesV2Setting,
    "Endpoint": Endpoint,
    "EndpointSync": EndpointSync,
    "EndpointSyncAllOf": EndpointSyncAllOf,
    "EndpointSyncAllOfSourceRef": EndpointSyncAllOfSourceRef,
    "Environment": Environment,
    "EnvironmentVariable": EnvironmentVariable,
    "EnvironmentVariableConfigMap": EnvironmentVariableConfigMap,
    "EnvironmentVariablePodFieldReference": EnvironmentVariablePodFieldReference,
    "EnvironmentVariableSecret": EnvironmentVariableSecret,
    "EnvironmentVariableUserCredential": EnvironmentVariableUserCredential,
    "EnvironmentVariablesDefaults": EnvironmentVariablesDefaults,
    "Event": Event,
    "EventComputedFields": EventComputedFields,
    "EventSyncFields": EventSyncFields,
    "EventsSyncRequest": EventsSyncRequest,
    "ExcludeField": ExcludeField,
    "ExposedUrl": ExposedUrl,
    "ExposedUrlRules": ExposedUrlRules,
    "ExposedUrlsDefaults": ExposedUrlsDefaults,
    "ExposedUrlsRules": ExposedUrlsRules,
    "ExtendedResource": ExtendedResource,
    "ExtendedResourceRules": ExtendedResourceRules,
    "ExtendedResourcesDefaults": ExtendedResourcesDefaults,
    "ExtendedResourcesRules": ExtendedResourcesRules,
    "FieldsToClear": FieldsToClear,
    "GetApplicationSubmission": GetApplicationSubmission,
    "GetAvailablePrioritiesAndCategories200Response": GetAvailablePrioritiesAndCategories200Response,
    "GetEnableHybridSubmission": GetEnableHybridSubmission,
    "GetEnablePolicies": GetEnablePolicies,
    "GetEnableSettingsValue": GetEnableSettingsValue,
    "GetEnableTemplatesV2": GetEnableTemplatesV2,
    "GetGpuOverProvisioning": GetGpuOverProvisioning,
    "GetInferenceWorkloadRevisions200Response": GetInferenceWorkloadRevisions200Response,
    "GetRevisionPods200Response": GetRevisionPods200Response,
    "GetWorkloadClusterPriorities200Response": GetWorkloadClusterPriorities200Response,
    "GetWorkloadEndpoints200Response": GetWorkloadEndpoints200Response,
    "GetWorkloadEvents200Response": GetWorkloadEvents200Response,
    "GetWorkloadHistory200Response": GetWorkloadHistory200Response,
    "GetWorkloadPriorities200Response": GetWorkloadPriorities200Response,
    "GetWorkloadSubmissionV2Enabled": GetWorkloadSubmissionV2Enabled,
    "GetWorkloads200Response": GetWorkloads200Response,
    "GetWorkloadsSetting": GetWorkloadsSetting,
    "GitCommon": GitCommon,
    "GitInstance": GitInstance,
    "GitInstanceAllOf": GitInstanceAllOf,
    "GitPassword": GitPassword,
    "GitRules": GitRules,
    "GitSecretRef": GitSecretRef,
    "GitsDefaults": GitsDefaults,
    "GitsRules": GitsRules,
    "GpuOverProvisioningSetting": GpuOverProvisioningSetting,
    "GpuRequestOptions": GpuRequestOptions,
    "GpuRequestOptionsOptionsInner": GpuRequestOptionsOptionsInner,
    "GpuRequestRules": GpuRequestRules,
    "HistogramSeries": HistogramSeries,
    "HistogramValue": HistogramValue,
    "HistoryRecord": HistoryRecord,
    "HistoryRecordMeta": HistoryRecordMeta,
    "HistoryRecordSpec": HistoryRecordSpec,
    "HostPath": HostPath,
    "HostPathInstance": HostPathInstance,
    "HostPathMountPropagationOptions": HostPathMountPropagationOptions,
    "HostPathMountPropagationOptionsOptionsInner": HostPathMountPropagationOptionsOptionsInner,
    "HostPathMountPropagationRules": HostPathMountPropagationRules,
    "HostPathRules": HostPathRules,
    "HostPathsDefaults": HostPathsDefaults,
    "HostPathsRules": HostPathsRules,
    "HttpResponse": HttpResponse,
    "ImagePullPolicyOptions": ImagePullPolicyOptions,
    "ImagePullPolicyOptionsOptionsInner": ImagePullPolicyOptionsOptionsInner,
    "ImagePullPolicyRules": ImagePullPolicyRules,
    "ImagePullSecret": ImagePullSecret,
    "ImagePullSecretsDefaults": ImagePullSecretsDefaults,
    "ImagePullSecretsRules": ImagePullSecretsRules,
    "ImagePullSecretsRulesAttributes": ImagePullSecretsRulesAttributes,
    "Inference": Inference,
    "InferenceCreationRequest": InferenceCreationRequest,
    "InferenceFields": InferenceFields,
    "InferenceFieldsRules": InferenceFieldsRules,
    "InferenceSpec": InferenceSpec,
    "InferenceSpecSpec": InferenceSpecSpec,
    "InferenceSpecSpecAllOf": InferenceSpecSpecAllOf,
    "InferenceSpecSpecAllOfSecurity": InferenceSpecSpecAllOfSecurity,
    "InferenceSpecSpecAllOfStorage": InferenceSpecSpecAllOfStorage,
    "InferenceTemplate": InferenceTemplate,
    "InferenceTemplateAllOf": InferenceTemplateAllOf,
    "InferenceTemplateAllOf1": InferenceTemplateAllOf1,
    "InferenceTemplateCreationRequest": InferenceTemplateCreationRequest,
    "InferenceTemplatePatchRequest": InferenceTemplatePatchRequest,
    "InferenceTemplateSpec": InferenceTemplateSpec,
    "InferenceTemplateSpecAllOf": InferenceTemplateSpecAllOf,
    "InferenceTemplateUpdateRequest": InferenceTemplateUpdateRequest,
    "InferenceTemplatesListResponse": InferenceTemplatesListResponse,
    "InferenceTemplatesListResponseItem": InferenceTemplatesListResponseItem,
    "InferenceUpdateRequest": InferenceUpdateRequest,
    "InferenceUpdateSpec": InferenceUpdateSpec,
    "InferenceUpdateSpecAutoscaling": InferenceUpdateSpecAutoscaling,
    "InferenceUpdateSpecServingConfiguration": InferenceUpdateSpecServingConfiguration,
    "InferenceUpdateSpecSpec": InferenceUpdateSpecSpec,
    "InferenceUpdateSpecSpecAllOf": InferenceUpdateSpecSpecAllOf,
    "InferencesMaxReplicasField": InferencesMaxReplicasField,
    "InferencesMinReplicasField": InferencesMinReplicasField,
    "InitialReplicasField": InitialReplicasField,
    "InitializationTimeoutField": InitializationTimeoutField,
    "InstancesRules": InstancesRules,
    "IntegerRules": IntegerRules,
    "IntegerRulesOptional": IntegerRulesOptional,
    "InternalEnrichOrgUnitRankingRequest": InternalEnrichOrgUnitRankingRequest,
    "InternalEnrichOrgUnitRankingResponse": InternalEnrichOrgUnitRankingResponse,
    "InternalPodsSyncRequest": InternalPodsSyncRequest,
    "InternalSyncPods207Response": InternalSyncPods207Response,
    "InternalSyncWorkloads207Response": InternalSyncWorkloads207Response,
    "InternalWorkloadsSyncRequest": InternalWorkloadsSyncRequest,
    "InvolvedObject": InvolvedObject,
    "ItemRules": ItemRules,
    "Label": Label,
    "LabelsDefaults": LabelsDefaults,
    "ListApplications200Response": ListApplications200Response,
    "ListCategories200Response": ListCategories200Response,
    "ListPods200Response": ListPods200Response,
    "ListWorkloadTypes200Response": ListWorkloadTypes200Response,
    "MasterSpec": MasterSpec,
    "MasterSpecSecurity": MasterSpecSecurity,
    "MasterSpecStorage": MasterSpecStorage,
    "MasterTemplateSpec": MasterTemplateSpec,
    "MatchExpression": MatchExpression,
    "MeasurementResponse": MeasurementResponse,
    "MeasurementResponseValuesInner": MeasurementResponseValuesInner,
    "MetricThresholdField": MetricThresholdField,
    "MetricThresholdPercentageField": MetricThresholdPercentageField,
    "MetricsCompositeResponse": MetricsCompositeResponse,
    "MetricsResponse": MetricsResponse,
    "ModelError": ModelError,
    "MpiLauncherCreationPolicyOptions": MpiLauncherCreationPolicyOptions,
    "MpiLauncherCreationPolicyOptionsOptionsInner": MpiLauncherCreationPolicyOptionsOptionsInner,
    "MpiLauncherCreationPolicyRules": MpiLauncherCreationPolicyRules,
    "Nfs": Nfs,
    "NfsInstance": NfsInstance,
    "NfsRules": NfsRules,
    "NfssDefaults": NfssDefaults,
    "NfssRules": NfssRules,
    "NodeAffinityRequired": NodeAffinityRequired,
    "NodeSelectorTerm": NodeSelectorTerm,
    "NumberRules": NumberRules,
    "OrgUnitRankingItem": OrgUnitRankingItem,
    "PaginationInfo": PaginationInfo,
    "PendingSchedulingMessage": PendingSchedulingMessage,
    "PhaseUpdate": PhaseUpdate,
    "Pod": Pod,
    "PodAffinity": PodAffinity,
    "PodAffinityRules": PodAffinityRules,
    "PodAffinityTypeOptions": PodAffinityTypeOptions,
    "PodAffinityTypeOptionsOptionsInner": PodAffinityTypeOptionsOptionsInner,
    "PodAffinityTypeRules": PodAffinityTypeRules,
    "PodComputedFields": PodComputedFields,
    "PodElement": PodElement,
    "PodGroupsSyncFields": PodGroupsSyncFields,
    "PodGroupsSyncRequest": PodGroupsSyncRequest,
    "PodNodeAffinity": PodNodeAffinity,
    "PodNodeAffinityPreferred": PodNodeAffinityPreferred,
    "PodNodeAffinityRequired": PodNodeAffinityRequired,
    "PodPreferredNodeSelectorTerm": PodPreferredNodeSelectorTerm,
    "PodRequestResources": PodRequestResources,
    "PodSyncFields": PodSyncFields,
    "PodToleration": PodToleration,
    "PodVerbosityFields": PodVerbosityFields,
    "PodsSyncRequest": PodsSyncRequest,
    "Port": Port,
    "PortRules": PortRules,
    "PortServiceTypeOptions": PortServiceTypeOptions,
    "PortServiceTypeOptionsOptionsInner": PortServiceTypeOptionsOptionsInner,
    "PortServiceTypeRules": PortServiceTypeRules,
    "PortsDefaults": PortsDefaults,
    "PortsRules": PortsRules,
    "PreemptibilityOptions": PreemptibilityOptions,
    "PreemptibilityOptionsOptionsInner": PreemptibilityOptionsOptionsInner,
    "PreemptibilityRules": PreemptibilityRules,
    "PrioritiesForStatus": PrioritiesForStatus,
    "PrioritiesForStatusPrioritiesInner": PrioritiesForStatusPrioritiesInner,
    "PrioritiesForStatusPrioritiesInnerAllOf": PrioritiesForStatusPrioritiesInnerAllOf,
    "PrioritiesForStatusPrioritiesInnerAllOfStatus": PrioritiesForStatusPrioritiesInnerAllOfStatus,
    "PriorityForCreate": PriorityForCreate,
    "PriorityStatus": PriorityStatus,
    "Probe": Probe,
    "ProbeHandler": ProbeHandler,
    "ProbeHandlerHttpGet": ProbeHandlerHttpGet,
    "ProbeHandlerRules": ProbeHandlerRules,
    "ProbeHandlerRulesHttpGet": ProbeHandlerRulesHttpGet,
    "ProbeRules": ProbeRules,
    "Probes": Probes,
    "ProbesRules": ProbesRules,
    "Pvc": Pvc,
    "PvcAccessModes": PvcAccessModes,
    "PvcAccessModesRules": PvcAccessModesRules,
    "PvcAddedAttrValue": PvcAddedAttrValue,
    "PvcFieldsNonUpdatable": PvcFieldsNonUpdatable,
    "PvcFieldsUpdatable": PvcFieldsUpdatable,
    "PvcInstance": PvcInstance,
    "PvcRules": PvcRules,
    "PvcsDefaults": PvcsDefaults,
    "PvcsRules": PvcsRules,
    "QuantityRules": QuantityRules,
    "QueueResources": QueueResources,
    "RelatedUrl": RelatedUrl,
    "RelatedUrlRules": RelatedUrlRules,
    "RelatedUrlsDefaults": RelatedUrlsDefaults,
    "RelatedUrlsRules": RelatedUrlsRules,
    "RequestInfo": RequestInfo,
    "RequestResourceCores": RequestResourceCores,
    "RequestResourceQuantity": RequestResourceQuantity,
    "RequestedPods": RequestedPods,
    "ResourceRef": ResourceRef,
    "ResponseInfo": ResponseInfo,
    "Revision": Revision,
    "RevisionSyncFields": RevisionSyncFields,
    "RevisionsSyncRequest": RevisionsSyncRequest,
    "RunnableInfo": RunnableInfo,
    "S3AccessKey": S3AccessKey,
    "S3Common": S3Common,
    "S3Instance": S3Instance,
    "S3Rules": S3Rules,
    "S3sDefaults": S3sDefaults,
    "S3sRules": S3sRules,
    "ScaleDownDelayField": ScaleDownDelayField,
    "ScaleToZeroRetentionField": ScaleToZeroRetentionField,
    "SchedulingCondition": SchedulingCondition,
    "SchedulingConditionQueueDetails": SchedulingConditionQueueDetails,
    "SeccompProfileTypeOptions": SeccompProfileTypeOptions,
    "SeccompProfileTypeOptionsOptionsInner": SeccompProfileTypeOptionsOptionsInner,
    "SeccompProfileTypeRules": SeccompProfileTypeRules,
    "Secret": Secret,
    "SecretFieldsNonUpdatable": SecretFieldsNonUpdatable,
    "SecretFieldsUpdatable": SecretFieldsUpdatable,
    "SecretInstance": SecretInstance,
    "SecretRules": SecretRules,
    "SecretsDefaults": SecretsDefaults,
    "SecretsRules": SecretsRules,
    "ServingConfiguration": ServingConfiguration,
    "ServingConfigurationField": ServingConfigurationField,
    "ServingConfigurationRules": ServingConfigurationRules,
    "ServingPort": ServingPort,
    "ServingPortAccess": ServingPortAccess,
    "ServingPortContainerAndProtocol": ServingPortContainerAndProtocol,
    "SourceOfRule": SourceOfRule,
    "StorageInstanceName": StorageInstanceName,
    "StringOption": StringOption,
    "StringRules": StringRules,
    "StringRulesCommon": StringRulesCommon,
    "StringRulesOptions": StringRulesOptions,
    "SubmissionError": SubmissionError,
    "SubmissionErrorV2": SubmissionErrorV2,
    "SubmitWithAssetIds": SubmitWithAssetIds,
    "SubmitWithTemplateId": SubmitWithTemplateId,
    "SupersetDefaults": SupersetDefaults,
    "SupersetDefaultsAllOf": SupersetDefaultsAllOf,
    "SupersetDefaultsAllOfCompute": SupersetDefaultsAllOfCompute,
    "SupersetDefaultsAllOfSecurity": SupersetDefaultsAllOfSecurity,
    "SupersetDefaultsAllOfStorage": SupersetDefaultsAllOfStorage,
    "SupersetRules": SupersetRules,
    "SupersetRulesAllOf": SupersetRulesAllOf,
    "SupersetRulesAllOfCompute": SupersetRulesAllOfCompute,
    "SupersetRulesAllOfSecurity": SupersetRulesAllOfSecurity,
    "SupersetRulesAllOfStorage": SupersetRulesAllOfStorage,
    "SupersetServingPort": SupersetServingPort,
    "SupersetServingPortRules": SupersetServingPortRules,
    "SupersetSpec": SupersetSpec,
    "SupersetSpecAllOf": SupersetSpecAllOf,
    "SupersetTemplate": SupersetTemplate,
    "SupersetTemplateAllOf": SupersetTemplateAllOf,
    "SupersetWorkload": SupersetWorkload,
    "SupersetWorkloadAllOf": SupersetWorkloadAllOf,
    "SyncJobsBackwardsRequest": SyncJobsBackwardsRequest,
    "SyncPodGroups207Response": SyncPodGroups207Response,
    "SyncPodsBackwards200Response": SyncPodsBackwards200Response,
    "SyncPriority": SyncPriority,
    "SyncPriorityAllOf": SyncPriorityAllOf,
    "SyncRevisions207Response": SyncRevisions207Response,
    "SystemInfo": SystemInfo,
    "TelemetryResponse": TelemetryResponse,
    "TelemetryResponseValuesInner": TelemetryResponseValuesInner,
    "TelemetryResponseValuesInnerGroupsInner": TelemetryResponseValuesInnerGroupsInner,
    "TemplateAssets": TemplateAssets,
    "TemplateAssetsIds": TemplateAssetsIds,
    "TemplateAssetsIdsOptional": TemplateAssetsIdsOptional,
    "TemplateComplianceInfo": TemplateComplianceInfo,
    "TemplateCreationMeta": TemplateCreationMeta,
    "TemplateCreationMetaFields": TemplateCreationMetaFields,
    "TemplateMasterSection": TemplateMasterSection,
    "TemplateMeta": TemplateMeta,
    "TemplateMetaFields": TemplateMetaFields,
    "TemplatePatchMeta": TemplatePatchMeta,
    "TemplatePatchMetaFields": TemplatePatchMetaFields,
    "TemplateUpdateMeta": TemplateUpdateMeta,
    "TemplateUpdateMetaFields": TemplateUpdateMetaFields,
    "TemplatesListResponse": TemplatesListResponse,
    "Toleration": Toleration,
    "TolerationEffectOptions": TolerationEffectOptions,
    "TolerationEffectOptionsOptionsInner": TolerationEffectOptionsOptionsInner,
    "TolerationEffectRules": TolerationEffectRules,
    "TolerationOperatorOptions": TolerationOperatorOptions,
    "TolerationOperatorOptionsOptionsInner": TolerationOperatorOptionsOptionsInner,
    "TolerationOperatorRules": TolerationOperatorRules,
    "TolerationRules": TolerationRules,
    "TolerationsDefaults": TolerationsDefaults,
    "TolerationsRules": TolerationsRules,
    "Training": Training,
    "TrainingCreationRequest": TrainingCreationRequest,
    "TrainingSpec": TrainingSpec,
    "TrainingSpecSpec": TrainingSpecSpec,
    "TrainingSpecSpecAllOf": TrainingSpecSpecAllOf,
    "TrainingTemplate": TrainingTemplate,
    "TrainingTemplateAllOf": TrainingTemplateAllOf,
    "TrainingTemplateAllOf1": TrainingTemplateAllOf1,
    "TrainingTemplateCreationRequest": TrainingTemplateCreationRequest,
    "TrainingTemplatePatchRequest": TrainingTemplatePatchRequest,
    "TrainingTemplateSpec": TrainingTemplateSpec,
    "TrainingTemplateSpecAllOf": TrainingTemplateSpecAllOf,
    "TrainingTemplateUpdateRequest": TrainingTemplateUpdateRequest,
    "TrainingTemplatesListResponse": TrainingTemplatesListResponse,
    "TrainingTemplatesListResponseItem": TrainingTemplatesListResponseItem,
    "UidGidSourceOptions": UidGidSourceOptions,
    "UidGidSourceOptionsOptionsInner": UidGidSourceOptionsOptionsInner,
    "UidGidSourceRules": UidGidSourceRules,
    "UnifiedPolicyInfo": UnifiedPolicyInfo,
    "UnifiedPolicyInfoPerReplicaType": UnifiedPolicyInfoPerReplicaType,
    "UnifiedPolicyInfoPerReplicaTypeDefaults": UnifiedPolicyInfoPerReplicaTypeDefaults,
    "UnifiedPolicyInfoPerReplicaTypeRules": UnifiedPolicyInfoPerReplicaTypeRules,
    "UnifiedProjectInfo": UnifiedProjectInfo,
    "UnifiedSpec": UnifiedSpec,
    "UnifiedSpecMaster": UnifiedSpecMaster,
    "UnifiedSpecWorker": UnifiedSpecWorker,
    "UnifiedWorkload": UnifiedWorkload,
    "UnifiedWorkloadRequest": UnifiedWorkloadRequest,
    "UpdateNodePodsRequest": UpdateNodePodsRequest,
    "UpdateRequest": UpdateRequest,
    "UpdateSpec": UpdateSpec,
    "UpdateSpecSpec": UpdateSpecSpec,
    "UpdateSpecSpecAllOf": UpdateSpecSpecAllOf,
    "UpdateWorkloadsSettingByKeyRequest": UpdateWorkloadsSettingByKeyRequest,
    "Workload": Workload,
    "WorkloadAllocatedResources": WorkloadAllocatedResources,
    "WorkloadBatch": WorkloadBatch,
    "WorkloadCategory": WorkloadCategory,
    "WorkloadChildrenIdsInner": WorkloadChildrenIdsInner,
    "WorkloadCreationFields": WorkloadCreationFields,
    "WorkloadCreationMeta": WorkloadCreationMeta,
    "WorkloadDetailed": WorkloadDetailed,
    "WorkloadDetailedAllOf": WorkloadDetailedAllOf,
    "WorkloadElement": WorkloadElement,
    "WorkloadElementKubernetesRef": WorkloadElementKubernetesRef,
    "WorkloadElementKubernetesRefGvk": WorkloadElementKubernetesRefGvk,
    "WorkloadElementResponse": WorkloadElementResponse,
    "WorkloadElementResponseAllOf": WorkloadElementResponseAllOf,
    "WorkloadElementTopology": WorkloadElementTopology,
    "WorkloadElementsResponse": WorkloadElementsResponse,
    "WorkloadKubernetesRef": WorkloadKubernetesRef,
    "WorkloadMeta": WorkloadMeta,
    "WorkloadNetworkStatus": WorkloadNetworkStatus,
    "WorkloadNetworkStatusServicesInner": WorkloadNetworkStatusServicesInner,
    "WorkloadNetworkStatusServicesInnerLoadBalancer": WorkloadNetworkStatusServicesInnerLoadBalancer,
    "WorkloadNetworkStatusServicesInnerLoadBalancerIngressInner": WorkloadNetworkStatusServicesInnerLoadBalancerIngressInner,
    "WorkloadNetworkStatusServicesInnerPortsInner": WorkloadNetworkStatusServicesInnerPortsInner,
    "WorkloadNetworkStatusServicesInnerUrlsInner": WorkloadNetworkStatusServicesInnerUrlsInner,
    "WorkloadNodeAffinity": WorkloadNodeAffinity,
    "WorkloadPatchFields": WorkloadPatchFields,
    "WorkloadPodsSyncRequest": WorkloadPodsSyncRequest,
    "WorkloadPriority": WorkloadPriority,
    "WorkloadPriorityAllOf": WorkloadPriorityAllOf,
    "WorkloadRequestResources": WorkloadRequestResources,
    "WorkloadSubmissionV2EnabledSetting": WorkloadSubmissionV2EnabledSetting,
    "WorkloadSyncFields": WorkloadSyncFields,
    "WorkloadSyncFieldsTopology": WorkloadSyncFieldsTopology,
    "WorkloadTopology": WorkloadTopology,
    "WorkloadTopologyActualPlacement": WorkloadTopologyActualPlacement,
    "WorkloadTypeAdditionalCreateFields": WorkloadTypeAdditionalCreateFields,
    "WorkloadTypeAdditionalResponseFields": WorkloadTypeAdditionalResponseFields,
    "WorkloadTypeConfig": WorkloadTypeConfig,
    "WorkloadTypeConfigAllOf": WorkloadTypeConfigAllOf,
    "WorkloadTypeCreateFields": WorkloadTypeCreateFields,
    "WorkloadTypeCreateFieldsAllOf": WorkloadTypeCreateFieldsAllOf,
    "WorkloadTypeCreateUpdateInnerFields": WorkloadTypeCreateUpdateInnerFields,
    "WorkloadTypeForSync": WorkloadTypeForSync,
    "WorkloadTypeForSyncAllOf": WorkloadTypeForSyncAllOf,
    "WorkloadTypeKartasWithVersionInner": WorkloadTypeKartasWithVersionInner,
    "WorkloadTypeResourceInterfacesWithVersionInner": WorkloadTypeResourceInterfacesWithVersionInner,
    "WorkloadTypeStatus": WorkloadTypeStatus,
    "WorkloadTypeStatusAllOf": WorkloadTypeStatusAllOf,
    "WorkloadTypeStatusFields": WorkloadTypeStatusFields,
    "WorkloadTypeUpdateFields": WorkloadTypeUpdateFields,
    "WorkloadTypeUpdateFieldsAllOf": WorkloadTypeUpdateFieldsAllOf,
    "WorkloadTypesForSync": WorkloadTypesForSync,
    "WorkloadsExtendedResource": WorkloadsExtendedResource,
    "WorkloadsSettingsGetResponse": WorkloadsSettingsGetResponse,
    "WorkloadsSyncInfo": WorkloadsSyncInfo,
    "WorkloadsSyncRequest": WorkloadsSyncRequest,
    "WorkloadsSyncResponse": WorkloadsSyncResponse,
    "Workspace": Workspace,
    "WorkspaceCreationRequest": WorkspaceCreationRequest,
    "WorkspaceSpec": WorkspaceSpec,
    "WorkspaceSpecSpec": WorkspaceSpecSpec,
    "WorkspaceSpecSpecAllOf": WorkspaceSpecSpecAllOf,
    "WorkspaceSpecSpecAllOfCompute": WorkspaceSpecSpecAllOfCompute,
    "WorkspaceSpecSpecAllOfSecurity": WorkspaceSpecSpecAllOfSecurity,
    "WorkspaceSpecSpecAllOfStorage": WorkspaceSpecSpecAllOfStorage,
    "WorkspaceTemplate": WorkspaceTemplate,
    "WorkspaceTemplateAllOf": WorkspaceTemplateAllOf,
    "WorkspaceTemplateAllOf1": WorkspaceTemplateAllOf1,
    "WorkspaceTemplateCreationRequest": WorkspaceTemplateCreationRequest,
    "WorkspaceTemplatePatchRequest": WorkspaceTemplatePatchRequest,
    "WorkspaceTemplateSpec": WorkspaceTemplateSpec,
    "WorkspaceTemplateSpecAllOf": WorkspaceTemplateSpecAllOf,
    "WorkspaceTemplateUpdateRequest": WorkspaceTemplateUpdateRequest,
    "WorkspaceTemplatesListResponse": WorkspaceTemplatesListResponse,
    "WorkspaceTemplatesListResponseItem": WorkspaceTemplatesListResponseItem,
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
