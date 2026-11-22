import localVarRequest from 'request';

export * from './activationReplicasField';
export * from './allowOverride';
export * from './annotation';
export * from './annotationsDefaults';
export * from './arbitraryRules';
export * from './arrayRules';
export * from './assetChangeNotification';
export * from './assetChangeType';
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
export * from './booleanRules';
export * from './breakdown';
export * from './breakdownBreakdown';
export * from './breakdownFieldRef';
export * from './breakdownFieldVariation';
export * from './breakdownFormatter';
export * from './breakdownFormatterArrayKey';
export * from './breakdownItemSelector';
export * from './breakdownSpecSelector';
export * from './capability';
export * from './claimInfo';
export * from './claimInfoRules';
export * from './claimInfoUpdatable';
export * from './cleanPodPolicy';
export * from './cleanPodPolicyOptions';
export * from './cleanPodPolicyOptionsOptionsInner';
export * from './cleanPodPolicyRules';
export * from './clusterSyncReport';
export * from './clusterSyncReportUpdate';
export * from './concurrencyHardLimitField';
export * from './configMap';
export * from './configMapInstance';
export * from './configMapRules';
export * from './configMapsDefaults';
export * from './configMapsRules';
export * from './createGvkPolicyReq';
export * from './dataVolume';
export * from './dataVolumeInstance';
export * from './dataVolumeRules';
export * from './dataVolumesDefaults';
export * from './dataVolumesRules';
export * from './defaultFromOrder';
export * from './defaultFromRule';
export * from './deleteGvkPolicyRulesReq';
export * from './distributedFramework';
export * from './distributedFrameworkOptions';
export * from './distributedFrameworkOptionsOptionsInner';
export * from './distributedFrameworkRules';
export * from './distributedImposedAssets';
export * from './distributedInferenceCommonDefaultsV2';
export * from './distributedInferenceCommonDefaultsV2AllOf';
export * from './distributedInferenceCommonRulesV2';
export * from './distributedInferenceCommonRulesV2AllOf';
export * from './distributedInferenceDefaultsV2';
export * from './distributedInferenceImposedAssets';
export * from './distributedInferenceLeaderDefaults';
export * from './distributedInferenceLeaderDefaultsFields';
export * from './distributedInferenceLeaderDefaultsFieldsLeader';
export * from './distributedInferenceLeaderRules';
export * from './distributedInferenceLeaderRulesFields';
export * from './distributedInferenceLeaderRulesFieldsLeader';
export * from './distributedInferenceLeaderWorkerDefaultsV2';
export * from './distributedInferenceLeaderWorkerDefaultsV2Security';
export * from './distributedInferenceLeaderWorkerDefaultsV2Storage';
export * from './distributedInferenceLeaderWorkerRulesV2';
export * from './distributedInferenceLeaderWorkerRulesV2Security';
export * from './distributedInferenceLeaderWorkerRulesV2SecurityPrivileged';
export * from './distributedInferenceLeaderWorkerRulesV2Storage';
export * from './distributedInferencePolicyChangeRequestV2';
export * from './distributedInferencePolicyDefaultsAndRulesV2';
export * from './distributedInferencePolicyDefaultsAndRulesV2Defaults';
export * from './distributedInferencePolicyDefaultsAndRulesV2Rules';
export * from './distributedInferencePolicyOverwriteRequestV2';
export * from './distributedInferencePolicyV2';
export * from './distributedInferenceReplicasField';
export * from './distributedInferenceReplicasFieldRules';
export * from './distributedInferenceRestartPolicy';
export * from './distributedInferenceRestartPolicyOptions';
export * from './distributedInferenceRestartPolicyOptionsOptionsInner';
export * from './distributedInferenceRestartPolicyRules';
export * from './distributedInferenceRulesV2';
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
export * from './distributedInferenceStartupPolicy';
export * from './distributedInferenceStartupPolicyField';
export * from './distributedInferenceStartupPolicyFieldRules';
export * from './distributedInferenceStartupPolicyFieldRulesStartupPolicy';
export * from './distributedInferenceStartupPolicyOptions';
export * from './distributedInferenceWorkerDefaults';
export * from './distributedInferenceWorkerDefaultsFields';
export * from './distributedInferenceWorkerDefaultsFieldsWorker';
export * from './distributedInferenceWorkerRules';
export * from './distributedInferenceWorkerRulesFields';
export * from './distributedInferenceWorkerRulesFieldsWorker';
export * from './distributedInferenceWorkersField';
export * from './distributedInferenceWorkersFieldRules';
export * from './distributedPolicyChangeRequestV2';
export * from './distributedPolicyDefaultsAndRulesV2';
export * from './distributedPolicyDefaultsV2';
export * from './distributedPolicyDefaultsV2Worker';
export * from './distributedPolicyDefaultsV2WorkerAllOf';
export * from './distributedPolicyOverwriteRequestV2';
export * from './distributedPolicyRulesV2';
export * from './distributedPolicyRulesV2Worker';
export * from './distributedPolicyRulesV2WorkerAllOf';
export * from './distributedPolicyV2';
export * from './effectiveRules';
export * from './effectiveRulesEffectiveRulesInner';
export * from './emptyDir';
export * from './emptyDirInstance';
export * from './emptyDirRules';
export * from './emptyDirsDefaults';
export * from './emptyDirsRules';
export * from './environmentVariable';
export * from './environmentVariableConfigMap';
export * from './environmentVariablePodFieldReference';
export * from './environmentVariableSecret';
export * from './environmentVariableUserCredential';
export * from './environmentVariablesDefaults';
export * from './excludeField';
export * from './exposedUrl';
export * from './exposedUrlRules';
export * from './exposedUrlsDefaults';
export * from './exposedUrlsRules';
export * from './extendedResource';
export * from './extendedResourceRules';
export * from './extendedResourcesDefaults';
export * from './extendedResourcesRules';
export * from './fallbackPolicy';
export * from './fallbackPolicyChangeRequest';
export * from './fallbackPolicyDefaultsAndRules';
export * from './fallbackPolicyDefaultsAndRulesDefaults';
export * from './fallbackPolicyDefaultsAndRulesRules';
export * from './fallbackPolicyOverwriteRequest';
export * from './fieldGroup';
export * from './fieldGroupCreationRequest';
export * from './fieldGroupDeleteFieldsRequest';
export * from './fieldGroupDeleteItemSelectorsRequest';
export * from './fieldGroupField';
export * from './fieldGroupFieldsWithScopeValue';
export * from './fieldGroupListResponse';
export * from './fieldGroupUpdateRequest';
export * from './formatter';
export * from './formatterArrayKey';
export * from './formatterCreationRequest';
export * from './formatterDeleteFieldsRequest';
export * from './formatterField';
export * from './formatterFieldArrayKey';
export * from './formatterListResponse';
export * from './formatterUpdateRequest';
export * from './gVK';
export * from './getGvkPolicyResp';
export * from './getMappingFieldInfoReq';
export * from './getMappingFieldInfoReqWorkloadType';
export * from './getMappingMinimalInfoReq';
export * from './getMappingMinimalInfoReqWorkloadType';
export * from './getMappingMinimalResp';
export * from './gitAuthenticationMethod';
export * from './gitCommon';
export * from './gitInstance';
export * from './gitInstanceAllOf';
export * from './gitPassword';
export * from './gitRules';
export * from './gitSecretRef';
export * from './gitsDefaults';
export * from './gitsRules';
export * from './gpuRequestOptions';
export * from './gpuRequestOptionsOptionsInner';
export * from './gpuRequestRules';
export * from './gpuRequestType';
export * from './hostPath';
export * from './hostPathInstance';
export * from './hostPathMountPropagation';
export * from './hostPathMountPropagationOptions';
export * from './hostPathMountPropagationOptionsOptionsInner';
export * from './hostPathMountPropagationRules';
export * from './hostPathRules';
export * from './hostPathsDefaults';
export * from './hostPathsRules';
export * from './imagePullPolicy';
export * from './imagePullPolicyOptions';
export * from './imagePullPolicyOptionsOptionsInner';
export * from './imagePullPolicyRules';
export * from './imagePullSecret';
export * from './imagePullSecretsDefaults';
export * from './imagePullSecretsRules';
export * from './imagePullSecretsRulesAttributes';
export * from './inferenceFields';
export * from './inferenceFieldsRules';
export * from './inferencePolicyChangeRequestV2';
export * from './inferencePolicyDefaultsAndRulesV2';
export * from './inferencePolicyDefaultsAndRulesV2Defaults';
export * from './inferencePolicyDefaultsAndRulesV2Rules';
export * from './inferencePolicyDefaultsV2';
export * from './inferencePolicyDefaultsV2AllOf';
export * from './inferencePolicyDefaultsV2AllOfSecurity';
export * from './inferencePolicyDefaultsV2AllOfStorage';
export * from './inferencePolicyOverwriteRequestV2';
export * from './inferencePolicyRulesV2';
export * from './inferencePolicyRulesV2AllOf';
export * from './inferencePolicyRulesV2AllOfSecurity';
export * from './inferencePolicyRulesV2AllOfStorage';
export * from './inferencePolicyV2';
export * from './inferencesMaxReplicasField';
export * from './inferencesMinReplicasField';
export * from './initialReplicasField';
export * from './initializationTimeoutField';
export * from './instancesRules';
export * from './integerRules';
export * from './integerRulesOptional';
export * from './itemRules';
export * from './itemSelector';
export * from './itemSelectorsWithScopeValue';
export * from './label';
export * from './labelsDefaults';
export * from './listGvkPoliciesResp';
export * from './mapping';
export * from './mappingCreUpdRequest';
export * from './mappingDeleteSpecSelectorsReq';
export * from './mappingDeleteSpecSelectorsReqWorkloadType';
export * from './mappingDeletionRequest';
export * from './mappingDeletionRequestWorkloadType';
export * from './mappingFieldType';
export * from './mappingListResponse';
export * from './mappingPossibleEnforcement';
export * from './mappingScope';
export * from './mappingScopeOfElement';
export * from './mappingWorkloadType';
export * from './matchExpression';
export * from './matchExpressionOperator';
export * from './metricThresholdField';
export * from './metricThresholdPercentageField';
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
export * from './nimCache';
export * from './nimCacheRules';
export * from './nimServiceDefaults';
export * from './nimServiceDefaultsAllOf';
export * from './nimServiceDefaultsAllOfAutoscaling';
export * from './nimServiceDefaultsAllOfCompute';
export * from './nimServiceDefaultsAllOfModelStore';
export * from './nimServiceDefaultsAllOfSecurity';
export * from './nimServiceMultiNode';
export * from './nimServiceMultiNodeRules';
export * from './nimServicePolicy';
export * from './nimServicePolicyChangeRequest';
export * from './nimServicePolicyDefaultsAndRules';
export * from './nimServicePolicyDefaultsAndRulesDefaults';
export * from './nimServicePolicyDefaultsAndRulesRules';
export * from './nimServicePolicyOverwriteRequest';
export * from './nimServicePvcFields';
export * from './nimServiceRules';
export * from './nimServiceRulesAllOf';
export * from './nimServiceRulesAllOfAutoscaling';
export * from './nimServiceRulesAllOfCompute';
export * from './nimServiceRulesAllOfModelStore';
export * from './nimServiceRulesAllOfSecurity';
export * from './nimServiceServingPort';
export * from './nimServiceServingPortAuthorizationType';
export * from './nimServiceServingPortAuthorizationTypeOptions';
export * from './nimServiceServingPortAuthorizationTypeOptionsOptionsInner';
export * from './nimServiceServingPortAuthorizationTypeRules';
export * from './nimServiceServingPortRules';
export * from './nodeAffinityRequired';
export * from './nodeSelectorTerm';
export * from './numberRules';
export * from './podAffinity';
export * from './podAffinityRules';
export * from './podAffinityType';
export * from './podAffinityTypeOptions';
export * from './podAffinityTypeOptionsOptionsInner';
export * from './podAffinityTypeRules';
export * from './policy';
export * from './policyCreationFields';
export * from './policyCrossType';
export * from './policyDescriptionField';
export * from './policyEnforcementBlocked';
export * from './policyEnforcementDefault';
export * from './policyEnforcementElements';
export * from './policyEnforcementElementsElements';
export * from './policyEnforcementItems';
export * from './policyEnforcementItemsItems';
export * from './policyEnforcementLength';
export * from './policyEnforcementLengthLength';
export * from './policyEnforcementMatch';
export * from './policyEnforcementOptions';
export * from './policyEnforcementQuantityRange';
export * from './policyEnforcementQuantityRangeQuantityRange';
export * from './policyEnforcementRange';
export * from './policyEnforcementRangeRange';
export * from './policyEnforcementRequired';
export * from './policyEnforcementSet';
export * from './policyIdAndOwnership';
export * from './policyListEntry';
export * from './policyListEntryAvailability';
export * from './policyListEntryStatus';
export * from './policyListResponse';
export * from './policyMeta';
export * from './policyPhase';
export * from './policyReadOnlyFields';
export * from './policyRule';
export * from './policyRuleEnforcement';
export * from './policyRuleEnforcements';
export * from './policyRuleEnforcementsEnforcementsInner';
export * from './policyRuleRef';
export * from './policyRuleSource';
export * from './policyRules';
export * from './policyRulesRef';
export * from './policyStatus';
export * from './policySyncResponse';
export * from './policySyncResponsePoliciesInner';
export * from './policySyncResponseRequestInfo';
export * from './policySyncResponseResponseInfo';
export * from './policySyncResponseSystemInfo';
export * from './policySyncStatus';
export * from './policySyncStatusOfCluster';
export * from './policyType';
export * from './policyValidationStatus';
export * from './policyValidationStatusValidation';
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
export * from './relatedUrl';
export * from './relatedUrlRules';
export * from './relatedUrlsDefaults';
export * from './relatedUrlsRules';
export * from './replicaDefaultsV2';
export * from './replicaDefaultsV2Storage';
export * from './replicaRulesV2';
export * from './replicaRulesV2Storage';
export * from './restartPolicy';
export * from './restartPolicyOptions';
export * from './restartPolicyOptionsOptionsInner';
export * from './restartPolicyRule';
export * from './s3AccessKey';
export * from './s3Common';
export * from './s3Instance';
export * from './s3Rules';
export * from './s3sDefaults';
export * from './s3sRules';
export * from './scaleDownDelayField';
export * from './scaleToZeroRetentionField';
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
export * from './servingPortAuthorizationTypeOptions';
export * from './servingPortAuthorizationTypeOptionsOptionsInner';
export * from './servingPortAuthorizationTypeRules';
export * from './servingPortContainerAndProtocol';
export * from './servingPortExposedProtocol';
export * from './servingPortExposedProtocolOptions';
export * from './servingPortExposedProtocolOptionsOptionsInner';
export * from './servingPortExposedProtocolRules';
export * from './servingPortProtocol';
export * from './servingPortProtocolOptions';
export * from './servingPortProtocolOptionsOptionsInner';
export * from './servingPortProtocolRules';
export * from './servingPortRules';
export * from './servingPortServiceType';
export * from './servingPortServiceTypeOptions';
export * from './servingPortServiceTypeOptionsOptionsInner';
export * from './servingPortServiceTypeRules';
export * from './sourceOfRule';
export * from './specSelector';
export * from './specSelectorDependency';
export * from './specSelectorsWithScopeValue';
export * from './storageInstanceName';
export * from './stringOption';
export * from './stringRules';
export * from './stringRulesCommon';
export * from './stringRulesOptions';
export * from './supersetDefaults';
export * from './supersetDefaultsAllOf';
export * from './supersetDefaultsAllOfCompute';
export * from './supersetDefaultsAllOfSecurity';
export * from './supersetDefaultsAllOfStorage';
export * from './supersetPolicy';
export * from './supersetRules';
export * from './supersetRulesAllOf';
export * from './supersetRulesAllOfCompute';
export * from './supersetRulesAllOfSecurity';
export * from './supersetRulesAllOfStorage';
export * from './supersetServingPort';
export * from './supersetServingPortRules';
export * from './syncConfig';
export * from './syncStatusUpdate';
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
export * from './trainingPolicyChangeRequestV2';
export * from './trainingPolicyDefaultsAndRulesV2';
export * from './trainingPolicyDefaultsAndRulesV2Defaults';
export * from './trainingPolicyDefaultsAndRulesV2Rules';
export * from './trainingPolicyDefaultsV2';
export * from './trainingPolicyDefaultsV2AllOf';
export * from './trainingPolicyOverwriteRequestV2';
export * from './trainingPolicyRulesV2';
export * from './trainingPolicyRulesV2AllOf';
export * from './trainingPolicyV2';
export * from './uidGidSource';
export * from './uidGidSourceOptions';
export * from './uidGidSourceOptionsOptionsInner';
export * from './uidGidSourceRules';
export * from './updateGvkPolicyReq';
export * from './workloadTypeAndScope';
export * from './workspacePolicyChangeRequestV2';
export * from './workspacePolicyDefaultsAndRulesV2';
export * from './workspacePolicyDefaultsAndRulesV2Defaults';
export * from './workspacePolicyDefaultsAndRulesV2Rules';
export * from './workspacePolicyDefaultsV2';
export * from './workspacePolicyDefaultsV2AllOf';
export * from './workspacePolicyOverwriteRequestV2';
export * from './workspacePolicyRulesV2';
export * from './workspacePolicyRulesV2AllOf';
export * from './workspacePolicyV2';

import * as fs from 'fs';

export interface RequestDetailedFile {
    value: Buffer;
    options?: {
        filename?: string;
        contentType?: string;
    }
}

export type RequestFile = string | Buffer | fs.ReadStream | RequestDetailedFile;


import { ActivationReplicasField } from './activationReplicasField';
import { AllowOverride } from './allowOverride';
import { Annotation } from './annotation';
import { AnnotationsDefaults } from './annotationsDefaults';
import { ArbitraryRules } from './arbitraryRules';
import { ArrayRules } from './arrayRules';
import { AssetChangeNotification } from './assetChangeNotification';
import { AssetChangeType } from './assetChangeType';
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
import { BooleanRules } from './booleanRules';
import { Breakdown } from './breakdown';
import { BreakdownBreakdown } from './breakdownBreakdown';
import { BreakdownFieldRef } from './breakdownFieldRef';
import { BreakdownFieldVariation } from './breakdownFieldVariation';
import { BreakdownFormatter } from './breakdownFormatter';
import { BreakdownFormatterArrayKey } from './breakdownFormatterArrayKey';
import { BreakdownItemSelector } from './breakdownItemSelector';
import { BreakdownSpecSelector } from './breakdownSpecSelector';
import { Capability } from './capability';
import { ClaimInfo } from './claimInfo';
import { ClaimInfoRules } from './claimInfoRules';
import { ClaimInfoUpdatable } from './claimInfoUpdatable';
import { CleanPodPolicy } from './cleanPodPolicy';
import { CleanPodPolicyOptions } from './cleanPodPolicyOptions';
import { CleanPodPolicyOptionsOptionsInner } from './cleanPodPolicyOptionsOptionsInner';
import { CleanPodPolicyRules } from './cleanPodPolicyRules';
import { ClusterSyncReport } from './clusterSyncReport';
import { ClusterSyncReportUpdate } from './clusterSyncReportUpdate';
import { ConcurrencyHardLimitField } from './concurrencyHardLimitField';
import { ConfigMap } from './configMap';
import { ConfigMapInstance } from './configMapInstance';
import { ConfigMapRules } from './configMapRules';
import { ConfigMapsDefaults } from './configMapsDefaults';
import { ConfigMapsRules } from './configMapsRules';
import { CreateGvkPolicyReq } from './createGvkPolicyReq';
import { DataVolume } from './dataVolume';
import { DataVolumeInstance } from './dataVolumeInstance';
import { DataVolumeRules } from './dataVolumeRules';
import { DataVolumesDefaults } from './dataVolumesDefaults';
import { DataVolumesRules } from './dataVolumesRules';
import { DefaultFromOrder } from './defaultFromOrder';
import { DefaultFromRule } from './defaultFromRule';
import { DeleteGvkPolicyRulesReq } from './deleteGvkPolicyRulesReq';
import { DistributedFramework } from './distributedFramework';
import { DistributedFrameworkOptions } from './distributedFrameworkOptions';
import { DistributedFrameworkOptionsOptionsInner } from './distributedFrameworkOptionsOptionsInner';
import { DistributedFrameworkRules } from './distributedFrameworkRules';
import { DistributedImposedAssets } from './distributedImposedAssets';
import { DistributedInferenceCommonDefaultsV2 } from './distributedInferenceCommonDefaultsV2';
import { DistributedInferenceCommonDefaultsV2AllOf } from './distributedInferenceCommonDefaultsV2AllOf';
import { DistributedInferenceCommonRulesV2 } from './distributedInferenceCommonRulesV2';
import { DistributedInferenceCommonRulesV2AllOf } from './distributedInferenceCommonRulesV2AllOf';
import { DistributedInferenceDefaultsV2 } from './distributedInferenceDefaultsV2';
import { DistributedInferenceImposedAssets } from './distributedInferenceImposedAssets';
import { DistributedInferenceLeaderDefaults } from './distributedInferenceLeaderDefaults';
import { DistributedInferenceLeaderDefaultsFields } from './distributedInferenceLeaderDefaultsFields';
import { DistributedInferenceLeaderDefaultsFieldsLeader } from './distributedInferenceLeaderDefaultsFieldsLeader';
import { DistributedInferenceLeaderRules } from './distributedInferenceLeaderRules';
import { DistributedInferenceLeaderRulesFields } from './distributedInferenceLeaderRulesFields';
import { DistributedInferenceLeaderRulesFieldsLeader } from './distributedInferenceLeaderRulesFieldsLeader';
import { DistributedInferenceLeaderWorkerDefaultsV2 } from './distributedInferenceLeaderWorkerDefaultsV2';
import { DistributedInferenceLeaderWorkerDefaultsV2Security } from './distributedInferenceLeaderWorkerDefaultsV2Security';
import { DistributedInferenceLeaderWorkerDefaultsV2Storage } from './distributedInferenceLeaderWorkerDefaultsV2Storage';
import { DistributedInferenceLeaderWorkerRulesV2 } from './distributedInferenceLeaderWorkerRulesV2';
import { DistributedInferenceLeaderWorkerRulesV2Security } from './distributedInferenceLeaderWorkerRulesV2Security';
import { DistributedInferenceLeaderWorkerRulesV2SecurityPrivileged } from './distributedInferenceLeaderWorkerRulesV2SecurityPrivileged';
import { DistributedInferenceLeaderWorkerRulesV2Storage } from './distributedInferenceLeaderWorkerRulesV2Storage';
import { DistributedInferencePolicyChangeRequestV2 } from './distributedInferencePolicyChangeRequestV2';
import { DistributedInferencePolicyDefaultsAndRulesV2 } from './distributedInferencePolicyDefaultsAndRulesV2';
import { DistributedInferencePolicyDefaultsAndRulesV2Defaults } from './distributedInferencePolicyDefaultsAndRulesV2Defaults';
import { DistributedInferencePolicyDefaultsAndRulesV2Rules } from './distributedInferencePolicyDefaultsAndRulesV2Rules';
import { DistributedInferencePolicyOverwriteRequestV2 } from './distributedInferencePolicyOverwriteRequestV2';
import { DistributedInferencePolicyV2 } from './distributedInferencePolicyV2';
import { DistributedInferenceReplicasField } from './distributedInferenceReplicasField';
import { DistributedInferenceReplicasFieldRules } from './distributedInferenceReplicasFieldRules';
import { DistributedInferenceRestartPolicy } from './distributedInferenceRestartPolicy';
import { DistributedInferenceRestartPolicyOptions } from './distributedInferenceRestartPolicyOptions';
import { DistributedInferenceRestartPolicyOptionsOptionsInner } from './distributedInferenceRestartPolicyOptionsOptionsInner';
import { DistributedInferenceRestartPolicyRules } from './distributedInferenceRestartPolicyRules';
import { DistributedInferenceRulesV2 } from './distributedInferenceRulesV2';
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
import { DistributedInferenceStartupPolicy } from './distributedInferenceStartupPolicy';
import { DistributedInferenceStartupPolicyField } from './distributedInferenceStartupPolicyField';
import { DistributedInferenceStartupPolicyFieldRules } from './distributedInferenceStartupPolicyFieldRules';
import { DistributedInferenceStartupPolicyFieldRulesStartupPolicy } from './distributedInferenceStartupPolicyFieldRulesStartupPolicy';
import { DistributedInferenceStartupPolicyOptions } from './distributedInferenceStartupPolicyOptions';
import { DistributedInferenceWorkerDefaults } from './distributedInferenceWorkerDefaults';
import { DistributedInferenceWorkerDefaultsFields } from './distributedInferenceWorkerDefaultsFields';
import { DistributedInferenceWorkerDefaultsFieldsWorker } from './distributedInferenceWorkerDefaultsFieldsWorker';
import { DistributedInferenceWorkerRules } from './distributedInferenceWorkerRules';
import { DistributedInferenceWorkerRulesFields } from './distributedInferenceWorkerRulesFields';
import { DistributedInferenceWorkerRulesFieldsWorker } from './distributedInferenceWorkerRulesFieldsWorker';
import { DistributedInferenceWorkersField } from './distributedInferenceWorkersField';
import { DistributedInferenceWorkersFieldRules } from './distributedInferenceWorkersFieldRules';
import { DistributedPolicyChangeRequestV2 } from './distributedPolicyChangeRequestV2';
import { DistributedPolicyDefaultsAndRulesV2 } from './distributedPolicyDefaultsAndRulesV2';
import { DistributedPolicyDefaultsV2 } from './distributedPolicyDefaultsV2';
import { DistributedPolicyDefaultsV2Worker } from './distributedPolicyDefaultsV2Worker';
import { DistributedPolicyDefaultsV2WorkerAllOf } from './distributedPolicyDefaultsV2WorkerAllOf';
import { DistributedPolicyOverwriteRequestV2 } from './distributedPolicyOverwriteRequestV2';
import { DistributedPolicyRulesV2 } from './distributedPolicyRulesV2';
import { DistributedPolicyRulesV2Worker } from './distributedPolicyRulesV2Worker';
import { DistributedPolicyRulesV2WorkerAllOf } from './distributedPolicyRulesV2WorkerAllOf';
import { DistributedPolicyV2 } from './distributedPolicyV2';
import { EffectiveRules } from './effectiveRules';
import { EffectiveRulesEffectiveRulesInner } from './effectiveRulesEffectiveRulesInner';
import { EmptyDir } from './emptyDir';
import { EmptyDirInstance } from './emptyDirInstance';
import { EmptyDirRules } from './emptyDirRules';
import { EmptyDirsDefaults } from './emptyDirsDefaults';
import { EmptyDirsRules } from './emptyDirsRules';
import { EnvironmentVariable } from './environmentVariable';
import { EnvironmentVariableConfigMap } from './environmentVariableConfigMap';
import { EnvironmentVariablePodFieldReference } from './environmentVariablePodFieldReference';
import { EnvironmentVariableSecret } from './environmentVariableSecret';
import { EnvironmentVariableUserCredential } from './environmentVariableUserCredential';
import { EnvironmentVariablesDefaults } from './environmentVariablesDefaults';
import { ExcludeField } from './excludeField';
import { ExposedUrl } from './exposedUrl';
import { ExposedUrlRules } from './exposedUrlRules';
import { ExposedUrlsDefaults } from './exposedUrlsDefaults';
import { ExposedUrlsRules } from './exposedUrlsRules';
import { ExtendedResource } from './extendedResource';
import { ExtendedResourceRules } from './extendedResourceRules';
import { ExtendedResourcesDefaults } from './extendedResourcesDefaults';
import { ExtendedResourcesRules } from './extendedResourcesRules';
import { FallbackPolicy } from './fallbackPolicy';
import { FallbackPolicyChangeRequest } from './fallbackPolicyChangeRequest';
import { FallbackPolicyDefaultsAndRules } from './fallbackPolicyDefaultsAndRules';
import { FallbackPolicyDefaultsAndRulesDefaults } from './fallbackPolicyDefaultsAndRulesDefaults';
import { FallbackPolicyDefaultsAndRulesRules } from './fallbackPolicyDefaultsAndRulesRules';
import { FallbackPolicyOverwriteRequest } from './fallbackPolicyOverwriteRequest';
import { FieldGroup } from './fieldGroup';
import { FieldGroupCreationRequest } from './fieldGroupCreationRequest';
import { FieldGroupDeleteFieldsRequest } from './fieldGroupDeleteFieldsRequest';
import { FieldGroupDeleteItemSelectorsRequest } from './fieldGroupDeleteItemSelectorsRequest';
import { FieldGroupField } from './fieldGroupField';
import { FieldGroupFieldsWithScopeValue } from './fieldGroupFieldsWithScopeValue';
import { FieldGroupListResponse } from './fieldGroupListResponse';
import { FieldGroupUpdateRequest } from './fieldGroupUpdateRequest';
import { Formatter } from './formatter';
import { FormatterArrayKey } from './formatterArrayKey';
import { FormatterCreationRequest } from './formatterCreationRequest';
import { FormatterDeleteFieldsRequest } from './formatterDeleteFieldsRequest';
import { FormatterField } from './formatterField';
import { FormatterFieldArrayKey } from './formatterFieldArrayKey';
import { FormatterListResponse } from './formatterListResponse';
import { FormatterUpdateRequest } from './formatterUpdateRequest';
import { GVK } from './gVK';
import { GetGvkPolicyResp } from './getGvkPolicyResp';
import { GetMappingFieldInfoReq } from './getMappingFieldInfoReq';
import { GetMappingFieldInfoReqWorkloadType } from './getMappingFieldInfoReqWorkloadType';
import { GetMappingMinimalInfoReq } from './getMappingMinimalInfoReq';
import { GetMappingMinimalInfoReqWorkloadType } from './getMappingMinimalInfoReqWorkloadType';
import { GetMappingMinimalResp } from './getMappingMinimalResp';
import { GitAuthenticationMethod } from './gitAuthenticationMethod';
import { GitCommon } from './gitCommon';
import { GitInstance } from './gitInstance';
import { GitInstanceAllOf } from './gitInstanceAllOf';
import { GitPassword } from './gitPassword';
import { GitRules } from './gitRules';
import { GitSecretRef } from './gitSecretRef';
import { GitsDefaults } from './gitsDefaults';
import { GitsRules } from './gitsRules';
import { GpuRequestOptions } from './gpuRequestOptions';
import { GpuRequestOptionsOptionsInner } from './gpuRequestOptionsOptionsInner';
import { GpuRequestRules } from './gpuRequestRules';
import { GpuRequestType } from './gpuRequestType';
import { HostPath } from './hostPath';
import { HostPathInstance } from './hostPathInstance';
import { HostPathMountPropagation } from './hostPathMountPropagation';
import { HostPathMountPropagationOptions } from './hostPathMountPropagationOptions';
import { HostPathMountPropagationOptionsOptionsInner } from './hostPathMountPropagationOptionsOptionsInner';
import { HostPathMountPropagationRules } from './hostPathMountPropagationRules';
import { HostPathRules } from './hostPathRules';
import { HostPathsDefaults } from './hostPathsDefaults';
import { HostPathsRules } from './hostPathsRules';
import { ImagePullPolicy } from './imagePullPolicy';
import { ImagePullPolicyOptions } from './imagePullPolicyOptions';
import { ImagePullPolicyOptionsOptionsInner } from './imagePullPolicyOptionsOptionsInner';
import { ImagePullPolicyRules } from './imagePullPolicyRules';
import { ImagePullSecret } from './imagePullSecret';
import { ImagePullSecretsDefaults } from './imagePullSecretsDefaults';
import { ImagePullSecretsRules } from './imagePullSecretsRules';
import { ImagePullSecretsRulesAttributes } from './imagePullSecretsRulesAttributes';
import { InferenceFields } from './inferenceFields';
import { InferenceFieldsRules } from './inferenceFieldsRules';
import { InferencePolicyChangeRequestV2 } from './inferencePolicyChangeRequestV2';
import { InferencePolicyDefaultsAndRulesV2 } from './inferencePolicyDefaultsAndRulesV2';
import { InferencePolicyDefaultsAndRulesV2Defaults } from './inferencePolicyDefaultsAndRulesV2Defaults';
import { InferencePolicyDefaultsAndRulesV2Rules } from './inferencePolicyDefaultsAndRulesV2Rules';
import { InferencePolicyDefaultsV2 } from './inferencePolicyDefaultsV2';
import { InferencePolicyDefaultsV2AllOf } from './inferencePolicyDefaultsV2AllOf';
import { InferencePolicyDefaultsV2AllOfSecurity } from './inferencePolicyDefaultsV2AllOfSecurity';
import { InferencePolicyDefaultsV2AllOfStorage } from './inferencePolicyDefaultsV2AllOfStorage';
import { InferencePolicyOverwriteRequestV2 } from './inferencePolicyOverwriteRequestV2';
import { InferencePolicyRulesV2 } from './inferencePolicyRulesV2';
import { InferencePolicyRulesV2AllOf } from './inferencePolicyRulesV2AllOf';
import { InferencePolicyRulesV2AllOfSecurity } from './inferencePolicyRulesV2AllOfSecurity';
import { InferencePolicyRulesV2AllOfStorage } from './inferencePolicyRulesV2AllOfStorage';
import { InferencePolicyV2 } from './inferencePolicyV2';
import { InferencesMaxReplicasField } from './inferencesMaxReplicasField';
import { InferencesMinReplicasField } from './inferencesMinReplicasField';
import { InitialReplicasField } from './initialReplicasField';
import { InitializationTimeoutField } from './initializationTimeoutField';
import { InstancesRules } from './instancesRules';
import { IntegerRules } from './integerRules';
import { IntegerRulesOptional } from './integerRulesOptional';
import { ItemRules } from './itemRules';
import { ItemSelector } from './itemSelector';
import { ItemSelectorsWithScopeValue } from './itemSelectorsWithScopeValue';
import { Label } from './label';
import { LabelsDefaults } from './labelsDefaults';
import { ListGvkPoliciesResp } from './listGvkPoliciesResp';
import { Mapping } from './mapping';
import { MappingCreUpdRequest } from './mappingCreUpdRequest';
import { MappingDeleteSpecSelectorsReq } from './mappingDeleteSpecSelectorsReq';
import { MappingDeleteSpecSelectorsReqWorkloadType } from './mappingDeleteSpecSelectorsReqWorkloadType';
import { MappingDeletionRequest } from './mappingDeletionRequest';
import { MappingDeletionRequestWorkloadType } from './mappingDeletionRequestWorkloadType';
import { MappingFieldType } from './mappingFieldType';
import { MappingListResponse } from './mappingListResponse';
import { MappingPossibleEnforcement } from './mappingPossibleEnforcement';
import { MappingScope } from './mappingScope';
import { MappingScopeOfElement } from './mappingScopeOfElement';
import { MappingWorkloadType } from './mappingWorkloadType';
import { MatchExpression } from './matchExpression';
import { MatchExpressionOperator } from './matchExpressionOperator';
import { MetricThresholdField } from './metricThresholdField';
import { MetricThresholdPercentageField } from './metricThresholdPercentageField';
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
import { NimCache } from './nimCache';
import { NimCacheRules } from './nimCacheRules';
import { NimServiceDefaults } from './nimServiceDefaults';
import { NimServiceDefaultsAllOf } from './nimServiceDefaultsAllOf';
import { NimServiceDefaultsAllOfAutoscaling } from './nimServiceDefaultsAllOfAutoscaling';
import { NimServiceDefaultsAllOfCompute } from './nimServiceDefaultsAllOfCompute';
import { NimServiceDefaultsAllOfModelStore } from './nimServiceDefaultsAllOfModelStore';
import { NimServiceDefaultsAllOfSecurity } from './nimServiceDefaultsAllOfSecurity';
import { NimServiceMultiNode } from './nimServiceMultiNode';
import { NimServiceMultiNodeRules } from './nimServiceMultiNodeRules';
import { NimServicePolicy } from './nimServicePolicy';
import { NimServicePolicyChangeRequest } from './nimServicePolicyChangeRequest';
import { NimServicePolicyDefaultsAndRules } from './nimServicePolicyDefaultsAndRules';
import { NimServicePolicyDefaultsAndRulesDefaults } from './nimServicePolicyDefaultsAndRulesDefaults';
import { NimServicePolicyDefaultsAndRulesRules } from './nimServicePolicyDefaultsAndRulesRules';
import { NimServicePolicyOverwriteRequest } from './nimServicePolicyOverwriteRequest';
import { NimServicePvcFields } from './nimServicePvcFields';
import { NimServiceRules } from './nimServiceRules';
import { NimServiceRulesAllOf } from './nimServiceRulesAllOf';
import { NimServiceRulesAllOfAutoscaling } from './nimServiceRulesAllOfAutoscaling';
import { NimServiceRulesAllOfCompute } from './nimServiceRulesAllOfCompute';
import { NimServiceRulesAllOfModelStore } from './nimServiceRulesAllOfModelStore';
import { NimServiceRulesAllOfSecurity } from './nimServiceRulesAllOfSecurity';
import { NimServiceServingPort } from './nimServiceServingPort';
import { NimServiceServingPortAuthorizationType } from './nimServiceServingPortAuthorizationType';
import { NimServiceServingPortAuthorizationTypeOptions } from './nimServiceServingPortAuthorizationTypeOptions';
import { NimServiceServingPortAuthorizationTypeOptionsOptionsInner } from './nimServiceServingPortAuthorizationTypeOptionsOptionsInner';
import { NimServiceServingPortAuthorizationTypeRules } from './nimServiceServingPortAuthorizationTypeRules';
import { NimServiceServingPortRules } from './nimServiceServingPortRules';
import { NodeAffinityRequired } from './nodeAffinityRequired';
import { NodeSelectorTerm } from './nodeSelectorTerm';
import { NumberRules } from './numberRules';
import { PodAffinity } from './podAffinity';
import { PodAffinityRules } from './podAffinityRules';
import { PodAffinityType } from './podAffinityType';
import { PodAffinityTypeOptions } from './podAffinityTypeOptions';
import { PodAffinityTypeOptionsOptionsInner } from './podAffinityTypeOptionsOptionsInner';
import { PodAffinityTypeRules } from './podAffinityTypeRules';
import { Policy } from './policy';
import { PolicyCreationFields } from './policyCreationFields';
import { PolicyCrossType } from './policyCrossType';
import { PolicyDescriptionField } from './policyDescriptionField';
import { PolicyEnforcementBlocked } from './policyEnforcementBlocked';
import { PolicyEnforcementDefault } from './policyEnforcementDefault';
import { PolicyEnforcementElements } from './policyEnforcementElements';
import { PolicyEnforcementElementsElements } from './policyEnforcementElementsElements';
import { PolicyEnforcementItems } from './policyEnforcementItems';
import { PolicyEnforcementItemsItems } from './policyEnforcementItemsItems';
import { PolicyEnforcementLength } from './policyEnforcementLength';
import { PolicyEnforcementLengthLength } from './policyEnforcementLengthLength';
import { PolicyEnforcementMatch } from './policyEnforcementMatch';
import { PolicyEnforcementOptions } from './policyEnforcementOptions';
import { PolicyEnforcementQuantityRange } from './policyEnforcementQuantityRange';
import { PolicyEnforcementQuantityRangeQuantityRange } from './policyEnforcementQuantityRangeQuantityRange';
import { PolicyEnforcementRange } from './policyEnforcementRange';
import { PolicyEnforcementRangeRange } from './policyEnforcementRangeRange';
import { PolicyEnforcementRequired } from './policyEnforcementRequired';
import { PolicyEnforcementSet } from './policyEnforcementSet';
import { PolicyIdAndOwnership } from './policyIdAndOwnership';
import { PolicyListEntry } from './policyListEntry';
import { PolicyListEntryAvailability } from './policyListEntryAvailability';
import { PolicyListEntryStatus } from './policyListEntryStatus';
import { PolicyListResponse } from './policyListResponse';
import { PolicyMeta } from './policyMeta';
import { PolicyPhase } from './policyPhase';
import { PolicyReadOnlyFields } from './policyReadOnlyFields';
import { PolicyRule } from './policyRule';
import { PolicyRuleEnforcement } from './policyRuleEnforcement';
import { PolicyRuleEnforcements } from './policyRuleEnforcements';
import { PolicyRuleEnforcementsEnforcementsInner } from './policyRuleEnforcementsEnforcementsInner';
import { PolicyRuleRef } from './policyRuleRef';
import { PolicyRuleSource } from './policyRuleSource';
import { PolicyRules } from './policyRules';
import { PolicyRulesRef } from './policyRulesRef';
import { PolicyStatus } from './policyStatus';
import { PolicySyncResponse } from './policySyncResponse';
import { PolicySyncResponsePoliciesInner } from './policySyncResponsePoliciesInner';
import { PolicySyncResponseRequestInfo } from './policySyncResponseRequestInfo';
import { PolicySyncResponseResponseInfo } from './policySyncResponseResponseInfo';
import { PolicySyncResponseSystemInfo } from './policySyncResponseSystemInfo';
import { PolicySyncStatus } from './policySyncStatus';
import { PolicySyncStatusOfCluster } from './policySyncStatusOfCluster';
import { PolicyType } from './policyType';
import { PolicyValidationStatus } from './policyValidationStatus';
import { PolicyValidationStatusValidation } from './policyValidationStatusValidation';
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
import { RelatedUrl } from './relatedUrl';
import { RelatedUrlRules } from './relatedUrlRules';
import { RelatedUrlsDefaults } from './relatedUrlsDefaults';
import { RelatedUrlsRules } from './relatedUrlsRules';
import { ReplicaDefaultsV2 } from './replicaDefaultsV2';
import { ReplicaDefaultsV2Storage } from './replicaDefaultsV2Storage';
import { ReplicaRulesV2 } from './replicaRulesV2';
import { ReplicaRulesV2Storage } from './replicaRulesV2Storage';
import { RestartPolicy } from './restartPolicy';
import { RestartPolicyOptions } from './restartPolicyOptions';
import { RestartPolicyOptionsOptionsInner } from './restartPolicyOptionsOptionsInner';
import { RestartPolicyRule } from './restartPolicyRule';
import { S3AccessKey } from './s3AccessKey';
import { S3Common } from './s3Common';
import { S3Instance } from './s3Instance';
import { S3Rules } from './s3Rules';
import { S3sDefaults } from './s3sDefaults';
import { S3sRules } from './s3sRules';
import { ScaleDownDelayField } from './scaleDownDelayField';
import { ScaleToZeroRetentionField } from './scaleToZeroRetentionField';
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
import { ServingPortAuthorizationTypeOptions } from './servingPortAuthorizationTypeOptions';
import { ServingPortAuthorizationTypeOptionsOptionsInner } from './servingPortAuthorizationTypeOptionsOptionsInner';
import { ServingPortAuthorizationTypeRules } from './servingPortAuthorizationTypeRules';
import { ServingPortContainerAndProtocol } from './servingPortContainerAndProtocol';
import { ServingPortExposedProtocol } from './servingPortExposedProtocol';
import { ServingPortExposedProtocolOptions } from './servingPortExposedProtocolOptions';
import { ServingPortExposedProtocolOptionsOptionsInner } from './servingPortExposedProtocolOptionsOptionsInner';
import { ServingPortExposedProtocolRules } from './servingPortExposedProtocolRules';
import { ServingPortProtocol } from './servingPortProtocol';
import { ServingPortProtocolOptions } from './servingPortProtocolOptions';
import { ServingPortProtocolOptionsOptionsInner } from './servingPortProtocolOptionsOptionsInner';
import { ServingPortProtocolRules } from './servingPortProtocolRules';
import { ServingPortRules } from './servingPortRules';
import { ServingPortServiceType } from './servingPortServiceType';
import { ServingPortServiceTypeOptions } from './servingPortServiceTypeOptions';
import { ServingPortServiceTypeOptionsOptionsInner } from './servingPortServiceTypeOptionsOptionsInner';
import { ServingPortServiceTypeRules } from './servingPortServiceTypeRules';
import { SourceOfRule } from './sourceOfRule';
import { SpecSelector } from './specSelector';
import { SpecSelectorDependency } from './specSelectorDependency';
import { SpecSelectorsWithScopeValue } from './specSelectorsWithScopeValue';
import { StorageInstanceName } from './storageInstanceName';
import { StringOption } from './stringOption';
import { StringRules } from './stringRules';
import { StringRulesCommon } from './stringRulesCommon';
import { StringRulesOptions } from './stringRulesOptions';
import { SupersetDefaults } from './supersetDefaults';
import { SupersetDefaultsAllOf } from './supersetDefaultsAllOf';
import { SupersetDefaultsAllOfCompute } from './supersetDefaultsAllOfCompute';
import { SupersetDefaultsAllOfSecurity } from './supersetDefaultsAllOfSecurity';
import { SupersetDefaultsAllOfStorage } from './supersetDefaultsAllOfStorage';
import { SupersetPolicy } from './supersetPolicy';
import { SupersetRules } from './supersetRules';
import { SupersetRulesAllOf } from './supersetRulesAllOf';
import { SupersetRulesAllOfCompute } from './supersetRulesAllOfCompute';
import { SupersetRulesAllOfSecurity } from './supersetRulesAllOfSecurity';
import { SupersetRulesAllOfStorage } from './supersetRulesAllOfStorage';
import { SupersetServingPort } from './supersetServingPort';
import { SupersetServingPortRules } from './supersetServingPortRules';
import { SyncConfig } from './syncConfig';
import { SyncStatusUpdate } from './syncStatusUpdate';
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
import { TrainingPolicyChangeRequestV2 } from './trainingPolicyChangeRequestV2';
import { TrainingPolicyDefaultsAndRulesV2 } from './trainingPolicyDefaultsAndRulesV2';
import { TrainingPolicyDefaultsAndRulesV2Defaults } from './trainingPolicyDefaultsAndRulesV2Defaults';
import { TrainingPolicyDefaultsAndRulesV2Rules } from './trainingPolicyDefaultsAndRulesV2Rules';
import { TrainingPolicyDefaultsV2 } from './trainingPolicyDefaultsV2';
import { TrainingPolicyDefaultsV2AllOf } from './trainingPolicyDefaultsV2AllOf';
import { TrainingPolicyOverwriteRequestV2 } from './trainingPolicyOverwriteRequestV2';
import { TrainingPolicyRulesV2 } from './trainingPolicyRulesV2';
import { TrainingPolicyRulesV2AllOf } from './trainingPolicyRulesV2AllOf';
import { TrainingPolicyV2 } from './trainingPolicyV2';
import { UidGidSource } from './uidGidSource';
import { UidGidSourceOptions } from './uidGidSourceOptions';
import { UidGidSourceOptionsOptionsInner } from './uidGidSourceOptionsOptionsInner';
import { UidGidSourceRules } from './uidGidSourceRules';
import { UpdateGvkPolicyReq } from './updateGvkPolicyReq';
import { WorkloadTypeAndScope } from './workloadTypeAndScope';
import { WorkspacePolicyChangeRequestV2 } from './workspacePolicyChangeRequestV2';
import { WorkspacePolicyDefaultsAndRulesV2 } from './workspacePolicyDefaultsAndRulesV2';
import { WorkspacePolicyDefaultsAndRulesV2Defaults } from './workspacePolicyDefaultsAndRulesV2Defaults';
import { WorkspacePolicyDefaultsAndRulesV2Rules } from './workspacePolicyDefaultsAndRulesV2Rules';
import { WorkspacePolicyDefaultsV2 } from './workspacePolicyDefaultsV2';
import { WorkspacePolicyDefaultsV2AllOf } from './workspacePolicyDefaultsV2AllOf';
import { WorkspacePolicyOverwriteRequestV2 } from './workspacePolicyOverwriteRequestV2';
import { WorkspacePolicyRulesV2 } from './workspacePolicyRulesV2';
import { WorkspacePolicyRulesV2AllOf } from './workspacePolicyRulesV2AllOf';
import { WorkspacePolicyV2 } from './workspacePolicyV2';

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
        "AssetChangeType": AssetChangeType,
        "AuthorizationType": AuthorizationType,
        "Capability": Capability,
        "CleanPodPolicy": CleanPodPolicy,
        "DistributedFramework": DistributedFramework,
        "DistributedInferenceRestartPolicy": DistributedInferenceRestartPolicy,
        "DistributedInferenceServingPortAccessAuthorizationTypeEnum": DistributedInferenceServingPortAccessAuthorizationTypeEnum,
        "DistributedInferenceServingPortProtocol": DistributedInferenceServingPortProtocol,
        "DistributedInferenceStartupPolicy": DistributedInferenceStartupPolicy,
        "GitAuthenticationMethod": GitAuthenticationMethod,
        "GpuRequestType": GpuRequestType,
        "HostPathMountPropagation": HostPathMountPropagation,
        "ImagePullPolicy": ImagePullPolicy,
        "MappingFieldType": MappingFieldType,
        "MappingPossibleEnforcement": MappingPossibleEnforcement,
        "MappingScope": MappingScope,
        "MatchExpressionOperator": MatchExpressionOperator,
        "MpiLauncherCreationPolicy": MpiLauncherCreationPolicy,
        "NimServiceServingPortAuthorizationType": NimServiceServingPortAuthorizationType,
        "PodAffinityType": PodAffinityType,
        "PolicyPhase": PolicyPhase,
        "PolicySyncStatusOfCluster.StatusEnum": PolicySyncStatusOfCluster.StatusEnum,
        "PolicyType": PolicyType,
        "PortServiceType": PortServiceType,
        "Preemptibility": Preemptibility,
        "ProbeHandlerScheme": ProbeHandlerScheme,
        "PvcVolumeMode": PvcVolumeMode,
        "RestartPolicy": RestartPolicy,
        "ScopeType": ScopeType,
        "SeccompProfileType": SeccompProfileType,
        "ServingPortAccessAuthorizationTypeEnum": ServingPortAccessAuthorizationTypeEnum,
        "ServingPortExposedProtocol": ServingPortExposedProtocol,
        "ServingPortProtocol": ServingPortProtocol,
        "ServingPortServiceType": ServingPortServiceType,
        "SyncStatusUpdate.ActionEnum": SyncStatusUpdate.ActionEnum,
        "TolerationEffect": TolerationEffect,
        "TolerationOperator": TolerationOperator,
        "UidGidSource": UidGidSource,
}

let typeMap: {[index: string]: any} = {
    "ActivationReplicasField": ActivationReplicasField,
    "AllowOverride": AllowOverride,
    "Annotation": Annotation,
    "AnnotationsDefaults": AnnotationsDefaults,
    "ArbitraryRules": ArbitraryRules,
    "ArrayRules": ArrayRules,
    "AssetChangeNotification": AssetChangeNotification,
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
    "BooleanRules": BooleanRules,
    "Breakdown": Breakdown,
    "BreakdownBreakdown": BreakdownBreakdown,
    "BreakdownFieldRef": BreakdownFieldRef,
    "BreakdownFieldVariation": BreakdownFieldVariation,
    "BreakdownFormatter": BreakdownFormatter,
    "BreakdownFormatterArrayKey": BreakdownFormatterArrayKey,
    "BreakdownItemSelector": BreakdownItemSelector,
    "BreakdownSpecSelector": BreakdownSpecSelector,
    "ClaimInfo": ClaimInfo,
    "ClaimInfoRules": ClaimInfoRules,
    "ClaimInfoUpdatable": ClaimInfoUpdatable,
    "CleanPodPolicyOptions": CleanPodPolicyOptions,
    "CleanPodPolicyOptionsOptionsInner": CleanPodPolicyOptionsOptionsInner,
    "CleanPodPolicyRules": CleanPodPolicyRules,
    "ClusterSyncReport": ClusterSyncReport,
    "ClusterSyncReportUpdate": ClusterSyncReportUpdate,
    "ConcurrencyHardLimitField": ConcurrencyHardLimitField,
    "ConfigMap": ConfigMap,
    "ConfigMapInstance": ConfigMapInstance,
    "ConfigMapRules": ConfigMapRules,
    "ConfigMapsDefaults": ConfigMapsDefaults,
    "ConfigMapsRules": ConfigMapsRules,
    "CreateGvkPolicyReq": CreateGvkPolicyReq,
    "DataVolume": DataVolume,
    "DataVolumeInstance": DataVolumeInstance,
    "DataVolumeRules": DataVolumeRules,
    "DataVolumesDefaults": DataVolumesDefaults,
    "DataVolumesRules": DataVolumesRules,
    "DefaultFromOrder": DefaultFromOrder,
    "DefaultFromRule": DefaultFromRule,
    "DeleteGvkPolicyRulesReq": DeleteGvkPolicyRulesReq,
    "DistributedFrameworkOptions": DistributedFrameworkOptions,
    "DistributedFrameworkOptionsOptionsInner": DistributedFrameworkOptionsOptionsInner,
    "DistributedFrameworkRules": DistributedFrameworkRules,
    "DistributedImposedAssets": DistributedImposedAssets,
    "DistributedInferenceCommonDefaultsV2": DistributedInferenceCommonDefaultsV2,
    "DistributedInferenceCommonDefaultsV2AllOf": DistributedInferenceCommonDefaultsV2AllOf,
    "DistributedInferenceCommonRulesV2": DistributedInferenceCommonRulesV2,
    "DistributedInferenceCommonRulesV2AllOf": DistributedInferenceCommonRulesV2AllOf,
    "DistributedInferenceDefaultsV2": DistributedInferenceDefaultsV2,
    "DistributedInferenceImposedAssets": DistributedInferenceImposedAssets,
    "DistributedInferenceLeaderDefaults": DistributedInferenceLeaderDefaults,
    "DistributedInferenceLeaderDefaultsFields": DistributedInferenceLeaderDefaultsFields,
    "DistributedInferenceLeaderDefaultsFieldsLeader": DistributedInferenceLeaderDefaultsFieldsLeader,
    "DistributedInferenceLeaderRules": DistributedInferenceLeaderRules,
    "DistributedInferenceLeaderRulesFields": DistributedInferenceLeaderRulesFields,
    "DistributedInferenceLeaderRulesFieldsLeader": DistributedInferenceLeaderRulesFieldsLeader,
    "DistributedInferenceLeaderWorkerDefaultsV2": DistributedInferenceLeaderWorkerDefaultsV2,
    "DistributedInferenceLeaderWorkerDefaultsV2Security": DistributedInferenceLeaderWorkerDefaultsV2Security,
    "DistributedInferenceLeaderWorkerDefaultsV2Storage": DistributedInferenceLeaderWorkerDefaultsV2Storage,
    "DistributedInferenceLeaderWorkerRulesV2": DistributedInferenceLeaderWorkerRulesV2,
    "DistributedInferenceLeaderWorkerRulesV2Security": DistributedInferenceLeaderWorkerRulesV2Security,
    "DistributedInferenceLeaderWorkerRulesV2SecurityPrivileged": DistributedInferenceLeaderWorkerRulesV2SecurityPrivileged,
    "DistributedInferenceLeaderWorkerRulesV2Storage": DistributedInferenceLeaderWorkerRulesV2Storage,
    "DistributedInferencePolicyChangeRequestV2": DistributedInferencePolicyChangeRequestV2,
    "DistributedInferencePolicyDefaultsAndRulesV2": DistributedInferencePolicyDefaultsAndRulesV2,
    "DistributedInferencePolicyDefaultsAndRulesV2Defaults": DistributedInferencePolicyDefaultsAndRulesV2Defaults,
    "DistributedInferencePolicyDefaultsAndRulesV2Rules": DistributedInferencePolicyDefaultsAndRulesV2Rules,
    "DistributedInferencePolicyOverwriteRequestV2": DistributedInferencePolicyOverwriteRequestV2,
    "DistributedInferencePolicyV2": DistributedInferencePolicyV2,
    "DistributedInferenceReplicasField": DistributedInferenceReplicasField,
    "DistributedInferenceReplicasFieldRules": DistributedInferenceReplicasFieldRules,
    "DistributedInferenceRestartPolicyOptions": DistributedInferenceRestartPolicyOptions,
    "DistributedInferenceRestartPolicyOptionsOptionsInner": DistributedInferenceRestartPolicyOptionsOptionsInner,
    "DistributedInferenceRestartPolicyRules": DistributedInferenceRestartPolicyRules,
    "DistributedInferenceRulesV2": DistributedInferenceRulesV2,
    "DistributedInferenceServingPort": DistributedInferenceServingPort,
    "DistributedInferenceServingPortAccess": DistributedInferenceServingPortAccess,
    "DistributedInferenceServingPortAuthorizationTypeOptions": DistributedInferenceServingPortAuthorizationTypeOptions,
    "DistributedInferenceServingPortAuthorizationTypeOptionsOptionsInner": DistributedInferenceServingPortAuthorizationTypeOptionsOptionsInner,
    "DistributedInferenceServingPortAuthorizationTypeRules": DistributedInferenceServingPortAuthorizationTypeRules,
    "DistributedInferenceServingPortContainerAndProtocol": DistributedInferenceServingPortContainerAndProtocol,
    "DistributedInferenceServingPortRules": DistributedInferenceServingPortRules,
    "DistributedInferenceServingPortRulesProperties": DistributedInferenceServingPortRulesProperties,
    "DistributedInferenceStartupPolicyField": DistributedInferenceStartupPolicyField,
    "DistributedInferenceStartupPolicyFieldRules": DistributedInferenceStartupPolicyFieldRules,
    "DistributedInferenceStartupPolicyFieldRulesStartupPolicy": DistributedInferenceStartupPolicyFieldRulesStartupPolicy,
    "DistributedInferenceStartupPolicyOptions": DistributedInferenceStartupPolicyOptions,
    "DistributedInferenceWorkerDefaults": DistributedInferenceWorkerDefaults,
    "DistributedInferenceWorkerDefaultsFields": DistributedInferenceWorkerDefaultsFields,
    "DistributedInferenceWorkerDefaultsFieldsWorker": DistributedInferenceWorkerDefaultsFieldsWorker,
    "DistributedInferenceWorkerRules": DistributedInferenceWorkerRules,
    "DistributedInferenceWorkerRulesFields": DistributedInferenceWorkerRulesFields,
    "DistributedInferenceWorkerRulesFieldsWorker": DistributedInferenceWorkerRulesFieldsWorker,
    "DistributedInferenceWorkersField": DistributedInferenceWorkersField,
    "DistributedInferenceWorkersFieldRules": DistributedInferenceWorkersFieldRules,
    "DistributedPolicyChangeRequestV2": DistributedPolicyChangeRequestV2,
    "DistributedPolicyDefaultsAndRulesV2": DistributedPolicyDefaultsAndRulesV2,
    "DistributedPolicyDefaultsV2": DistributedPolicyDefaultsV2,
    "DistributedPolicyDefaultsV2Worker": DistributedPolicyDefaultsV2Worker,
    "DistributedPolicyDefaultsV2WorkerAllOf": DistributedPolicyDefaultsV2WorkerAllOf,
    "DistributedPolicyOverwriteRequestV2": DistributedPolicyOverwriteRequestV2,
    "DistributedPolicyRulesV2": DistributedPolicyRulesV2,
    "DistributedPolicyRulesV2Worker": DistributedPolicyRulesV2Worker,
    "DistributedPolicyRulesV2WorkerAllOf": DistributedPolicyRulesV2WorkerAllOf,
    "DistributedPolicyV2": DistributedPolicyV2,
    "EffectiveRules": EffectiveRules,
    "EffectiveRulesEffectiveRulesInner": EffectiveRulesEffectiveRulesInner,
    "EmptyDir": EmptyDir,
    "EmptyDirInstance": EmptyDirInstance,
    "EmptyDirRules": EmptyDirRules,
    "EmptyDirsDefaults": EmptyDirsDefaults,
    "EmptyDirsRules": EmptyDirsRules,
    "EnvironmentVariable": EnvironmentVariable,
    "EnvironmentVariableConfigMap": EnvironmentVariableConfigMap,
    "EnvironmentVariablePodFieldReference": EnvironmentVariablePodFieldReference,
    "EnvironmentVariableSecret": EnvironmentVariableSecret,
    "EnvironmentVariableUserCredential": EnvironmentVariableUserCredential,
    "EnvironmentVariablesDefaults": EnvironmentVariablesDefaults,
    "ExcludeField": ExcludeField,
    "ExposedUrl": ExposedUrl,
    "ExposedUrlRules": ExposedUrlRules,
    "ExposedUrlsDefaults": ExposedUrlsDefaults,
    "ExposedUrlsRules": ExposedUrlsRules,
    "ExtendedResource": ExtendedResource,
    "ExtendedResourceRules": ExtendedResourceRules,
    "ExtendedResourcesDefaults": ExtendedResourcesDefaults,
    "ExtendedResourcesRules": ExtendedResourcesRules,
    "FallbackPolicy": FallbackPolicy,
    "FallbackPolicyChangeRequest": FallbackPolicyChangeRequest,
    "FallbackPolicyDefaultsAndRules": FallbackPolicyDefaultsAndRules,
    "FallbackPolicyDefaultsAndRulesDefaults": FallbackPolicyDefaultsAndRulesDefaults,
    "FallbackPolicyDefaultsAndRulesRules": FallbackPolicyDefaultsAndRulesRules,
    "FallbackPolicyOverwriteRequest": FallbackPolicyOverwriteRequest,
    "FieldGroup": FieldGroup,
    "FieldGroupCreationRequest": FieldGroupCreationRequest,
    "FieldGroupDeleteFieldsRequest": FieldGroupDeleteFieldsRequest,
    "FieldGroupDeleteItemSelectorsRequest": FieldGroupDeleteItemSelectorsRequest,
    "FieldGroupField": FieldGroupField,
    "FieldGroupFieldsWithScopeValue": FieldGroupFieldsWithScopeValue,
    "FieldGroupListResponse": FieldGroupListResponse,
    "FieldGroupUpdateRequest": FieldGroupUpdateRequest,
    "Formatter": Formatter,
    "FormatterArrayKey": FormatterArrayKey,
    "FormatterCreationRequest": FormatterCreationRequest,
    "FormatterDeleteFieldsRequest": FormatterDeleteFieldsRequest,
    "FormatterField": FormatterField,
    "FormatterFieldArrayKey": FormatterFieldArrayKey,
    "FormatterListResponse": FormatterListResponse,
    "FormatterUpdateRequest": FormatterUpdateRequest,
    "GVK": GVK,
    "GetGvkPolicyResp": GetGvkPolicyResp,
    "GetMappingFieldInfoReq": GetMappingFieldInfoReq,
    "GetMappingFieldInfoReqWorkloadType": GetMappingFieldInfoReqWorkloadType,
    "GetMappingMinimalInfoReq": GetMappingMinimalInfoReq,
    "GetMappingMinimalInfoReqWorkloadType": GetMappingMinimalInfoReqWorkloadType,
    "GetMappingMinimalResp": GetMappingMinimalResp,
    "GitCommon": GitCommon,
    "GitInstance": GitInstance,
    "GitInstanceAllOf": GitInstanceAllOf,
    "GitPassword": GitPassword,
    "GitRules": GitRules,
    "GitSecretRef": GitSecretRef,
    "GitsDefaults": GitsDefaults,
    "GitsRules": GitsRules,
    "GpuRequestOptions": GpuRequestOptions,
    "GpuRequestOptionsOptionsInner": GpuRequestOptionsOptionsInner,
    "GpuRequestRules": GpuRequestRules,
    "HostPath": HostPath,
    "HostPathInstance": HostPathInstance,
    "HostPathMountPropagationOptions": HostPathMountPropagationOptions,
    "HostPathMountPropagationOptionsOptionsInner": HostPathMountPropagationOptionsOptionsInner,
    "HostPathMountPropagationRules": HostPathMountPropagationRules,
    "HostPathRules": HostPathRules,
    "HostPathsDefaults": HostPathsDefaults,
    "HostPathsRules": HostPathsRules,
    "ImagePullPolicyOptions": ImagePullPolicyOptions,
    "ImagePullPolicyOptionsOptionsInner": ImagePullPolicyOptionsOptionsInner,
    "ImagePullPolicyRules": ImagePullPolicyRules,
    "ImagePullSecret": ImagePullSecret,
    "ImagePullSecretsDefaults": ImagePullSecretsDefaults,
    "ImagePullSecretsRules": ImagePullSecretsRules,
    "ImagePullSecretsRulesAttributes": ImagePullSecretsRulesAttributes,
    "InferenceFields": InferenceFields,
    "InferenceFieldsRules": InferenceFieldsRules,
    "InferencePolicyChangeRequestV2": InferencePolicyChangeRequestV2,
    "InferencePolicyDefaultsAndRulesV2": InferencePolicyDefaultsAndRulesV2,
    "InferencePolicyDefaultsAndRulesV2Defaults": InferencePolicyDefaultsAndRulesV2Defaults,
    "InferencePolicyDefaultsAndRulesV2Rules": InferencePolicyDefaultsAndRulesV2Rules,
    "InferencePolicyDefaultsV2": InferencePolicyDefaultsV2,
    "InferencePolicyDefaultsV2AllOf": InferencePolicyDefaultsV2AllOf,
    "InferencePolicyDefaultsV2AllOfSecurity": InferencePolicyDefaultsV2AllOfSecurity,
    "InferencePolicyDefaultsV2AllOfStorage": InferencePolicyDefaultsV2AllOfStorage,
    "InferencePolicyOverwriteRequestV2": InferencePolicyOverwriteRequestV2,
    "InferencePolicyRulesV2": InferencePolicyRulesV2,
    "InferencePolicyRulesV2AllOf": InferencePolicyRulesV2AllOf,
    "InferencePolicyRulesV2AllOfSecurity": InferencePolicyRulesV2AllOfSecurity,
    "InferencePolicyRulesV2AllOfStorage": InferencePolicyRulesV2AllOfStorage,
    "InferencePolicyV2": InferencePolicyV2,
    "InferencesMaxReplicasField": InferencesMaxReplicasField,
    "InferencesMinReplicasField": InferencesMinReplicasField,
    "InitialReplicasField": InitialReplicasField,
    "InitializationTimeoutField": InitializationTimeoutField,
    "InstancesRules": InstancesRules,
    "IntegerRules": IntegerRules,
    "IntegerRulesOptional": IntegerRulesOptional,
    "ItemRules": ItemRules,
    "ItemSelector": ItemSelector,
    "ItemSelectorsWithScopeValue": ItemSelectorsWithScopeValue,
    "Label": Label,
    "LabelsDefaults": LabelsDefaults,
    "ListGvkPoliciesResp": ListGvkPoliciesResp,
    "Mapping": Mapping,
    "MappingCreUpdRequest": MappingCreUpdRequest,
    "MappingDeleteSpecSelectorsReq": MappingDeleteSpecSelectorsReq,
    "MappingDeleteSpecSelectorsReqWorkloadType": MappingDeleteSpecSelectorsReqWorkloadType,
    "MappingDeletionRequest": MappingDeletionRequest,
    "MappingDeletionRequestWorkloadType": MappingDeletionRequestWorkloadType,
    "MappingListResponse": MappingListResponse,
    "MappingScopeOfElement": MappingScopeOfElement,
    "MappingWorkloadType": MappingWorkloadType,
    "MatchExpression": MatchExpression,
    "MetricThresholdField": MetricThresholdField,
    "MetricThresholdPercentageField": MetricThresholdPercentageField,
    "ModelError": ModelError,
    "MpiLauncherCreationPolicyOptions": MpiLauncherCreationPolicyOptions,
    "MpiLauncherCreationPolicyOptionsOptionsInner": MpiLauncherCreationPolicyOptionsOptionsInner,
    "MpiLauncherCreationPolicyRules": MpiLauncherCreationPolicyRules,
    "Nfs": Nfs,
    "NfsInstance": NfsInstance,
    "NfsRules": NfsRules,
    "NfssDefaults": NfssDefaults,
    "NfssRules": NfssRules,
    "NimCache": NimCache,
    "NimCacheRules": NimCacheRules,
    "NimServiceDefaults": NimServiceDefaults,
    "NimServiceDefaultsAllOf": NimServiceDefaultsAllOf,
    "NimServiceDefaultsAllOfAutoscaling": NimServiceDefaultsAllOfAutoscaling,
    "NimServiceDefaultsAllOfCompute": NimServiceDefaultsAllOfCompute,
    "NimServiceDefaultsAllOfModelStore": NimServiceDefaultsAllOfModelStore,
    "NimServiceDefaultsAllOfSecurity": NimServiceDefaultsAllOfSecurity,
    "NimServiceMultiNode": NimServiceMultiNode,
    "NimServiceMultiNodeRules": NimServiceMultiNodeRules,
    "NimServicePolicy": NimServicePolicy,
    "NimServicePolicyChangeRequest": NimServicePolicyChangeRequest,
    "NimServicePolicyDefaultsAndRules": NimServicePolicyDefaultsAndRules,
    "NimServicePolicyDefaultsAndRulesDefaults": NimServicePolicyDefaultsAndRulesDefaults,
    "NimServicePolicyDefaultsAndRulesRules": NimServicePolicyDefaultsAndRulesRules,
    "NimServicePolicyOverwriteRequest": NimServicePolicyOverwriteRequest,
    "NimServicePvcFields": NimServicePvcFields,
    "NimServiceRules": NimServiceRules,
    "NimServiceRulesAllOf": NimServiceRulesAllOf,
    "NimServiceRulesAllOfAutoscaling": NimServiceRulesAllOfAutoscaling,
    "NimServiceRulesAllOfCompute": NimServiceRulesAllOfCompute,
    "NimServiceRulesAllOfModelStore": NimServiceRulesAllOfModelStore,
    "NimServiceRulesAllOfSecurity": NimServiceRulesAllOfSecurity,
    "NimServiceServingPort": NimServiceServingPort,
    "NimServiceServingPortAuthorizationTypeOptions": NimServiceServingPortAuthorizationTypeOptions,
    "NimServiceServingPortAuthorizationTypeOptionsOptionsInner": NimServiceServingPortAuthorizationTypeOptionsOptionsInner,
    "NimServiceServingPortAuthorizationTypeRules": NimServiceServingPortAuthorizationTypeRules,
    "NimServiceServingPortRules": NimServiceServingPortRules,
    "NodeAffinityRequired": NodeAffinityRequired,
    "NodeSelectorTerm": NodeSelectorTerm,
    "NumberRules": NumberRules,
    "PodAffinity": PodAffinity,
    "PodAffinityRules": PodAffinityRules,
    "PodAffinityTypeOptions": PodAffinityTypeOptions,
    "PodAffinityTypeOptionsOptionsInner": PodAffinityTypeOptionsOptionsInner,
    "PodAffinityTypeRules": PodAffinityTypeRules,
    "Policy": Policy,
    "PolicyCreationFields": PolicyCreationFields,
    "PolicyCrossType": PolicyCrossType,
    "PolicyDescriptionField": PolicyDescriptionField,
    "PolicyEnforcementBlocked": PolicyEnforcementBlocked,
    "PolicyEnforcementDefault": PolicyEnforcementDefault,
    "PolicyEnforcementElements": PolicyEnforcementElements,
    "PolicyEnforcementElementsElements": PolicyEnforcementElementsElements,
    "PolicyEnforcementItems": PolicyEnforcementItems,
    "PolicyEnforcementItemsItems": PolicyEnforcementItemsItems,
    "PolicyEnforcementLength": PolicyEnforcementLength,
    "PolicyEnforcementLengthLength": PolicyEnforcementLengthLength,
    "PolicyEnforcementMatch": PolicyEnforcementMatch,
    "PolicyEnforcementOptions": PolicyEnforcementOptions,
    "PolicyEnforcementQuantityRange": PolicyEnforcementQuantityRange,
    "PolicyEnforcementQuantityRangeQuantityRange": PolicyEnforcementQuantityRangeQuantityRange,
    "PolicyEnforcementRange": PolicyEnforcementRange,
    "PolicyEnforcementRangeRange": PolicyEnforcementRangeRange,
    "PolicyEnforcementRequired": PolicyEnforcementRequired,
    "PolicyEnforcementSet": PolicyEnforcementSet,
    "PolicyIdAndOwnership": PolicyIdAndOwnership,
    "PolicyListEntry": PolicyListEntry,
    "PolicyListEntryAvailability": PolicyListEntryAvailability,
    "PolicyListEntryStatus": PolicyListEntryStatus,
    "PolicyListResponse": PolicyListResponse,
    "PolicyMeta": PolicyMeta,
    "PolicyReadOnlyFields": PolicyReadOnlyFields,
    "PolicyRule": PolicyRule,
    "PolicyRuleEnforcement": PolicyRuleEnforcement,
    "PolicyRuleEnforcements": PolicyRuleEnforcements,
    "PolicyRuleEnforcementsEnforcementsInner": PolicyRuleEnforcementsEnforcementsInner,
    "PolicyRuleRef": PolicyRuleRef,
    "PolicyRuleSource": PolicyRuleSource,
    "PolicyRules": PolicyRules,
    "PolicyRulesRef": PolicyRulesRef,
    "PolicyStatus": PolicyStatus,
    "PolicySyncResponse": PolicySyncResponse,
    "PolicySyncResponsePoliciesInner": PolicySyncResponsePoliciesInner,
    "PolicySyncResponseRequestInfo": PolicySyncResponseRequestInfo,
    "PolicySyncResponseResponseInfo": PolicySyncResponseResponseInfo,
    "PolicySyncResponseSystemInfo": PolicySyncResponseSystemInfo,
    "PolicySyncStatus": PolicySyncStatus,
    "PolicySyncStatusOfCluster": PolicySyncStatusOfCluster,
    "PolicyValidationStatus": PolicyValidationStatus,
    "PolicyValidationStatusValidation": PolicyValidationStatusValidation,
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
    "RelatedUrl": RelatedUrl,
    "RelatedUrlRules": RelatedUrlRules,
    "RelatedUrlsDefaults": RelatedUrlsDefaults,
    "RelatedUrlsRules": RelatedUrlsRules,
    "ReplicaDefaultsV2": ReplicaDefaultsV2,
    "ReplicaDefaultsV2Storage": ReplicaDefaultsV2Storage,
    "ReplicaRulesV2": ReplicaRulesV2,
    "ReplicaRulesV2Storage": ReplicaRulesV2Storage,
    "RestartPolicyOptions": RestartPolicyOptions,
    "RestartPolicyOptionsOptionsInner": RestartPolicyOptionsOptionsInner,
    "RestartPolicyRule": RestartPolicyRule,
    "S3AccessKey": S3AccessKey,
    "S3Common": S3Common,
    "S3Instance": S3Instance,
    "S3Rules": S3Rules,
    "S3sDefaults": S3sDefaults,
    "S3sRules": S3sRules,
    "ScaleDownDelayField": ScaleDownDelayField,
    "ScaleToZeroRetentionField": ScaleToZeroRetentionField,
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
    "ServingPortAuthorizationTypeOptions": ServingPortAuthorizationTypeOptions,
    "ServingPortAuthorizationTypeOptionsOptionsInner": ServingPortAuthorizationTypeOptionsOptionsInner,
    "ServingPortAuthorizationTypeRules": ServingPortAuthorizationTypeRules,
    "ServingPortContainerAndProtocol": ServingPortContainerAndProtocol,
    "ServingPortExposedProtocolOptions": ServingPortExposedProtocolOptions,
    "ServingPortExposedProtocolOptionsOptionsInner": ServingPortExposedProtocolOptionsOptionsInner,
    "ServingPortExposedProtocolRules": ServingPortExposedProtocolRules,
    "ServingPortProtocolOptions": ServingPortProtocolOptions,
    "ServingPortProtocolOptionsOptionsInner": ServingPortProtocolOptionsOptionsInner,
    "ServingPortProtocolRules": ServingPortProtocolRules,
    "ServingPortRules": ServingPortRules,
    "ServingPortServiceTypeOptions": ServingPortServiceTypeOptions,
    "ServingPortServiceTypeOptionsOptionsInner": ServingPortServiceTypeOptionsOptionsInner,
    "ServingPortServiceTypeRules": ServingPortServiceTypeRules,
    "SourceOfRule": SourceOfRule,
    "SpecSelector": SpecSelector,
    "SpecSelectorDependency": SpecSelectorDependency,
    "SpecSelectorsWithScopeValue": SpecSelectorsWithScopeValue,
    "StorageInstanceName": StorageInstanceName,
    "StringOption": StringOption,
    "StringRules": StringRules,
    "StringRulesCommon": StringRulesCommon,
    "StringRulesOptions": StringRulesOptions,
    "SupersetDefaults": SupersetDefaults,
    "SupersetDefaultsAllOf": SupersetDefaultsAllOf,
    "SupersetDefaultsAllOfCompute": SupersetDefaultsAllOfCompute,
    "SupersetDefaultsAllOfSecurity": SupersetDefaultsAllOfSecurity,
    "SupersetDefaultsAllOfStorage": SupersetDefaultsAllOfStorage,
    "SupersetPolicy": SupersetPolicy,
    "SupersetRules": SupersetRules,
    "SupersetRulesAllOf": SupersetRulesAllOf,
    "SupersetRulesAllOfCompute": SupersetRulesAllOfCompute,
    "SupersetRulesAllOfSecurity": SupersetRulesAllOfSecurity,
    "SupersetRulesAllOfStorage": SupersetRulesAllOfStorage,
    "SupersetServingPort": SupersetServingPort,
    "SupersetServingPortRules": SupersetServingPortRules,
    "SyncConfig": SyncConfig,
    "SyncStatusUpdate": SyncStatusUpdate,
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
    "TrainingPolicyChangeRequestV2": TrainingPolicyChangeRequestV2,
    "TrainingPolicyDefaultsAndRulesV2": TrainingPolicyDefaultsAndRulesV2,
    "TrainingPolicyDefaultsAndRulesV2Defaults": TrainingPolicyDefaultsAndRulesV2Defaults,
    "TrainingPolicyDefaultsAndRulesV2Rules": TrainingPolicyDefaultsAndRulesV2Rules,
    "TrainingPolicyDefaultsV2": TrainingPolicyDefaultsV2,
    "TrainingPolicyDefaultsV2AllOf": TrainingPolicyDefaultsV2AllOf,
    "TrainingPolicyOverwriteRequestV2": TrainingPolicyOverwriteRequestV2,
    "TrainingPolicyRulesV2": TrainingPolicyRulesV2,
    "TrainingPolicyRulesV2AllOf": TrainingPolicyRulesV2AllOf,
    "TrainingPolicyV2": TrainingPolicyV2,
    "UidGidSourceOptions": UidGidSourceOptions,
    "UidGidSourceOptionsOptionsInner": UidGidSourceOptionsOptionsInner,
    "UidGidSourceRules": UidGidSourceRules,
    "UpdateGvkPolicyReq": UpdateGvkPolicyReq,
    "WorkloadTypeAndScope": WorkloadTypeAndScope,
    "WorkspacePolicyChangeRequestV2": WorkspacePolicyChangeRequestV2,
    "WorkspacePolicyDefaultsAndRulesV2": WorkspacePolicyDefaultsAndRulesV2,
    "WorkspacePolicyDefaultsAndRulesV2Defaults": WorkspacePolicyDefaultsAndRulesV2Defaults,
    "WorkspacePolicyDefaultsAndRulesV2Rules": WorkspacePolicyDefaultsAndRulesV2Rules,
    "WorkspacePolicyDefaultsV2": WorkspacePolicyDefaultsV2,
    "WorkspacePolicyDefaultsV2AllOf": WorkspacePolicyDefaultsV2AllOf,
    "WorkspacePolicyOverwriteRequestV2": WorkspacePolicyOverwriteRequestV2,
    "WorkspacePolicyRulesV2": WorkspacePolicyRulesV2,
    "WorkspacePolicyRulesV2AllOf": WorkspacePolicyRulesV2AllOf,
    "WorkspacePolicyV2": WorkspacePolicyV2,
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
