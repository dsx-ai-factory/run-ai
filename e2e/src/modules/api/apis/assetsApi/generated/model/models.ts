import localVarRequest from 'request';

export * from './accessKey';
export * from './accessKeyCreationRequest';
export * from './accessKeyCreationSpec';
export * from './accessKeyListResponse';
export * from './accessKeySpec';
export * from './accessKeyUpdateRequest';
export * from './activationReplicasField';
export * from './allowOverQuota';
export * from './annotation';
export * from './arbitraryRules';
export * from './arrayRules';
export * from './assetClusterStatusInfo';
export * from './assetClusterStatusIssue';
export * from './assetClusterStatusIssueEnum';
export * from './assetCountResponse';
export * from './assetCreationFields';
export * from './assetCreationRequest';
export * from './assetDatasourceRef';
export * from './assetFlatRequest';
export * from './assetFlatResponse';
export * from './assetIdAndKind';
export * from './assetKind';
export * from './assetListByIdsResponse';
export * from './assetMeta';
export * from './assetMetaForSync';
export * from './assetMetaForSyncAllOf';
export * from './assetMetaForSyncAllOf1';
export * from './assetMetaForSyncAllOf2';
export * from './assetPhaseUpdate';
export * from './assetReadOnlyFields';
export * from './assetRef';
export * from './assetResourcesIdsItem';
export * from './assetResourcesIdsSyncRequest';
export * from './assetSyncStatusEnum';
export * from './assetUpdatableFields';
export * from './assetUpdateRequest';
export * from './assetUsageInfo';
export * from './assetsDefaults';
export * from './assetsIds';
export * from './assetsRef';
export * from './assetsRefCompute';
export * from './assetsRefEnvironment';
export * from './assetsRefOptional';
export * from './assetsRules';
export * from './assetsRulesDatasources';
export * from './assetsUsageRef';
export * from './assetsUsageRefCompute';
export * from './assetsUsageRefEnvironment';
export * from './attributeDefaults';
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
export * from './capability';
export * from './claimInfo';
export * from './claimInfoRules';
export * from './claimInfoUpdatable';
export * from './cleanPodPolicy';
export * from './clusterInfo';
export * from './clusterResourceInfo';
export * from './clusterSyncReport';
export * from './clusterSyncReportUpdate';
export * from './complianceInfo';
export * from './complianceInfoReason';
export * from './complianceIssuesV2';
export * from './complianceIssuesV2ComplianceIssuesInner';
export * from './computeAsset';
export * from './computeAssetSpec';
export * from './computeCreationRequest';
export * from './computeCreationRequestSpec';
export * from './computeListResponse';
export * from './computeUpdateRequest';
export * from './concurrencyHardLimitField';
export * from './configMap';
export * from './configMapAsset';
export * from './configMapAssetSpec';
export * from './configMapCreationRequest';
export * from './configMapInstance';
export * from './configMapListResponse';
export * from './configMapRules';
export * from './configMapUpdateRequest';
export * from './configMapsRules';
export * from './connection';
export * from './credentialAsset';
export * from './credentialKind';
export * from './credentialUpdatableFields';
export * from './credentialUpdateRequest';
export * from './credentialsListResponse';
export * from './credentialsListResponseAssetSpec';
export * from './credentialsListResponseAssetSpecAccessKey';
export * from './credentialsListResponseAssetSpecDockerRegistry';
export * from './credentialsListResponseAssetSpecGenericSecret';
export * from './credentialsListResponseAssetSpecNgcApiKey';
export * from './credentialsListResponseAssetSpecPassword';
export * from './credentialsListResponseEntry';
export * from './credentialsUpdate';
export * from './credentialsUpdateSpec';
export * from './credentialsUpdateSpecAccessKey';
export * from './credentialsUpdateSpecDockerRegistry';
export * from './credentialsUpdateSpecGenericSecret';
export * from './credentialsUpdateSpecPassword';
export * from './credentialsUpdatesResp';
export * from './dataSourceOverrides';
export * from './dataVolume';
export * from './dataVolumeAssetSpec';
export * from './dataVolumeInstance';
export * from './dataVolumeRules';
export * from './dataVolumesRules';
export * from './datasourceDefaults';
export * from './datasourceListResponse';
export * from './datasourceListResponseAssetSpec';
export * from './datasourceListResponseAssetSpecConfigMap';
export * from './datasourceListResponseAssetSpecDataVolume';
export * from './datasourceListResponseAssetSpecGit';
export * from './datasourceListResponseAssetSpecHostPath';
export * from './datasourceListResponseAssetSpecNfs';
export * from './datasourceListResponseAssetSpecPvc';
export * from './datasourceListResponseAssetSpecS3';
export * from './datasourceListResponseAssetSpecSecret';
export * from './datasourceListResponseEntry';
export * from './datasourceRef';
export * from './datasourceUpdate';
export * from './datasourceUpdateSpec';
export * from './datasourceUpdateSpecPvc';
export * from './datasourceUpdatesResp';
export * from './defaultFromRule';
export * from './displayedJob';
export * from './distMaster';
export * from './distributed';
export * from './distributedAllOf';
export * from './distributedData';
export * from './distributedFramework';
export * from './distributedInferenceCommonSpec';
export * from './distributedInferenceCommonSpecAllOf';
export * from './distributedInferenceLeaderSpecFields';
export * from './distributedInferenceLeaderSpecFieldsLeader';
export * from './distributedInferenceLeaderWorkerSpec';
export * from './distributedInferenceLeaderWorkerSpecCompute';
export * from './distributedInferenceLeaderWorkerSpecSecurity';
export * from './distributedInferenceLeaderWorkerSpecStorage';
export * from './distributedInferenceReplicasField';
export * from './distributedInferenceRestartPolicy';
export * from './distributedInferenceServingPort';
export * from './distributedInferenceServingPortAccess';
export * from './distributedInferenceServingPortAccessAuthorizationTypeEnum';
export * from './distributedInferenceServingPortContainerAndProtocol';
export * from './distributedInferenceServingPortProtocol';
export * from './distributedInferenceStartupPolicy';
export * from './distributedInferenceStartupPolicyField';
export * from './distributedInferenceWorkerSpecFields';
export * from './distributedInferenceWorkerSpecFieldsWorker';
export * from './distributedInferenceWorkersField';
export * from './distributedMaster';
export * from './distributedMasterData';
export * from './distributedParams';
export * from './dockerRegistry';
export * from './dockerRegistryCreationRequest';
export * from './dockerRegistryCreationSpec';
export * from './dockerRegistryListResponse';
export * from './dockerRegistrySpec';
export * from './dockerRegistryUpdateRequest';
export * from './emptyDir';
export * from './emptyDirInstance';
export * from './emptyDirRules';
export * from './emptyDirsRules';
export * from './encryptionKeyId';
export * from './environmentAsset';
export * from './environmentAssetRef';
export * from './environmentAssetSpec';
export * from './environmentCreationRequest';
export * from './environmentListResponse';
export * from './environmentUpdateRequest';
export * from './environmentVariable';
export * from './environmentVariableConfigMap';
export * from './environmentVariableCredential';
export * from './environmentVariableOfAsset';
export * from './environmentVariablePodFieldReference';
export * from './environmentVariableSecret';
export * from './environmentVariableUserCredential';
export * from './event';
export * from './eventSyncFields';
export * from './eventsSyncRequest';
export * from './excludeField';
export * from './exposedUrl';
export * from './extendedResource';
export * from './externalToolInfo';
export * from './externalToolType';
export * from './externalUrlInfo';
export * from './genericSecret';
export * from './genericSecretCreationRequest';
export * from './genericSecretCreationSpec';
export * from './genericSecretListResponse';
export * from './genericSecretSpec';
export * from './genericSecretUpdateRequest';
export * from './gitAsset';
export * from './gitAssetSpec';
export * from './gitAuthenticationMethod';
export * from './gitCommon';
export * from './gitCreationRequest';
export * from './gitCredentialId';
export * from './gitInstance';
export * from './gitInstanceAllOf';
export * from './gitListResponse';
export * from './gitPassword';
export * from './gitPasswordAsset';
export * from './gitRules';
export * from './gitSecretRef';
export * from './gitUpdateRequest';
export * from './gitsRules';
export * from './gpuRequestType';
export * from './historyRecord';
export * from './historyRecordMeta';
export * from './historyRecordSpec';
export * from './historyRecordType';
export * from './hostPath';
export * from './hostPathAsset';
export * from './hostPathCreationRequest';
export * from './hostPathInstance';
export * from './hostPathListResponse';
export * from './hostPathMountPropagation';
export * from './hostPathMountPropagationOptions';
export * from './hostPathMountPropagationOptionsOptionsInner';
export * from './hostPathMountPropagationRules';
export * from './hostPathRules';
export * from './hostPathSpec';
export * from './hostPathUpdateRequest';
export * from './hostPathsRules';
export * from './httpResponse';
export * from './huggingFaceModelInfo';
export * from './huggingFaceModelInfoRequest';
export * from './imagePullPolicy';
export * from './imagePullSecret';
export * from './imagePullSecretsRules';
export * from './imagePullSecretsRulesAttributes';
export * from './inference';
export * from './inferenceCreationRequestV2';
export * from './inferenceFields';
export * from './inferenceFieldsRules';
export * from './inferenceSpecificRunParams';
export * from './inferenceV2';
export * from './inferencesMaxReplicasField';
export * from './inferencesMinReplicasField';
export * from './infoDistMaster';
export * from './infoDistributed';
export * from './infoDistributedAllOf';
export * from './initialReplicasField';
export * from './initializationTimeoutField';
export * from './instancesRules';
export * from './integerRules';
export * from './integerRulesOptional';
export * from './internalConnectionType';
export * from './internalToolInfo';
export * from './internalToolType';
export * from './involvedObject';
export * from './itemRules';
export * from './keyValuePair';
export * from './label';
export * from './masterSpec';
export * from './masterSpecSecurity';
export * from './masterSpecStorage';
export * from './matchExpression';
export * from './matchExpressionOperator';
export * from './metricThresholdField';
export * from './metricThresholdPercentageField';
export * from './modelAsset';
export * from './modelAssetSpec';
export * from './modelCreationRequest';
export * from './modelError';
export * from './modelInference';
export * from './modelInferenceCreationRequest';
export * from './modelInferenceCreationRequestAssets';
export * from './modelInferenceCreationRequestV2';
export * from './modelInferenceCreationRequestV2Assets';
export * from './modelInferenceV2';
export * from './modelListResponse';
export * from './modelUpdateRequest';
export * from './mpiLauncherCreationPolicy';
export * from './nFSAsset';
export * from './nFSAssetSpec';
export * from './nFSCreationRequest';
export * from './nFSListResponse';
export * from './nFSUpdateRequest';
export * from './nfs';
export * from './nfsInstance';
export * from './nfsRules';
export * from './nfssRules';
export * from './ngcApiKey';
export * from './ngcApiKeyCreationRequest';
export * from './ngcApiKeyCreationSpec';
export * from './ngcApiKeyInternal';
export * from './ngcApiKeyListResponse';
export * from './ngcApiKeySpec';
export * from './ngcApiKeyUpdateRequest';
export * from './nodeAffinityRequired';
export * from './nodePortInfo';
export * from './nodeSelectorTerm';
export * from './nonOverridableSpecFields';
export * from './nonOverridableSpecFieldsAllOf';
export * from './nonOverridableSpecFieldsAllOf1';
export * from './numberRules';
export * from './optionalAssetIds';
export * from './overridableSpecFields';
export * from './pVCAsset';
export * from './pVCAssetSpec';
export * from './pVCCreationRequest';
export * from './pVCHistoryResponse';
export * from './pVCListResponse';
export * from './pVCUpdateRequest';
export * from './password';
export * from './passwordCreationRequest';
export * from './passwordCreationSpec';
export * from './passwordListResponse';
export * from './passwordSpec';
export * from './passwordUpdateRequest';
export * from './phase';
export * from './podAffinity';
export * from './podAffinityRules';
export * from './podAffinityType';
export * from './podAffinityTypeOptions';
export * from './podAffinityTypeOptionsOptionsInner';
export * from './podAffinityTypeRules';
export * from './policyInfo';
export * from './policyInfoAllOf';
export * from './policyInfoAllOf1';
export * from './policyRuleEnum';
export * from './port';
export * from './portServiceType';
export * from './preemptibility';
export * from './preemptibilityOptions';
export * from './preemptibilityOptionsOptionsInner';
export * from './preemptibilityRules';
export * from './probe';
export * from './probeHandler';
export * from './probeHandlerHttpGet';
export * from './probeHandlerScheme';
export * from './probes';
export * from './pvc';
export * from './pvcAccessModes';
export * from './pvcAccessModesRules';
export * from './pvcAddedAttrValue';
export * from './pvcFieldsNonUpdatable';
export * from './pvcFieldsUpdatable';
export * from './pvcInstance';
export * from './pvcRules';
export * from './pvcVolumeMode';
export * from './pvcsRules';
export * from './registry';
export * from './registryCreationRequest';
export * from './registryListResponse';
export * from './registrySpec';
export * from './registryUpdateRequest';
export * from './relatedUrl';
export * from './repositories';
export * from './repositoryTags';
export * from './requestInfo';
export * from './responseInfo';
export * from './restartPolicy';
export * from './restartPolicyOptions';
export * from './restartPolicyOptionsOptionsInner';
export * from './restartPolicyRule';
export * from './rules';
export * from './s3AccessKey';
export * from './s3AccessKeyAsset';
export * from './s3Asset';
export * from './s3AssetSpec';
export * from './s3Common';
export * from './s3CreationRequest';
export * from './s3Instance';
export * from './s3ListResponse';
export * from './s3Rules';
export * from './s3UpdateRequest';
export * from './s3sRules';
export * from './scaleDownDelayField';
export * from './scaleToZeroRetentionField';
export * from './scope';
export * from './scopeType';
export * from './seccompProfileType';
export * from './secret';
export * from './secretAsset';
export * from './secretAssetCreationRequest';
export * from './secretAssetListResponse';
export * from './secretAssetSpec';
export * from './secretAssetSpecAllOf';
export * from './secretAssetUpdateRequest';
export * from './secretFieldsNonUpdatable';
export * from './secretFieldsUpdatable';
export * from './secretInstance';
export * from './secretRules';
export * from './secretsRules';
export * from './servingConfiguration';
export * from './servingConfigurationField';
export * from './servingConfigurationRules';
export * from './servingPortAccess';
export * from './servingPortAccessAuthorizationTypeEnum';
export * from './servingPortInfo';
export * from './servingPortProtocol';
export * from './sourceOfRule';
export * from './specificRunAutoScaling';
export * from './specificRunAutoScalingAutoScaling';
export * from './specificRunConnectionInfo';
export * from './specificRunConnections';
export * from './specificRunCreationFields';
export * from './specificRunInfoFields';
export * from './specificRunMetricFields';
export * from './specificRunParamAttrDefaults';
export * from './specificRunParams';
export * from './specificRunParamsAllOf';
export * from './specificRunParamsRules';
export * from './specificRunParamsRulesAllOf';
export * from './specificRunServingPortAccess';
export * from './specificRunServingPortAccessServingPortAccess';
export * from './storageInstanceName';
export * from './stringOption';
export * from './stringRules';
export * from './stringRulesCommon';
export * from './stringRulesOptions';
export * from './submissionErrorV2';
export * from './supersetServingPort';
export * from './supersetSpec';
export * from './supersetSpecAllOf';
export * from './supersetSpecAllOfCompute';
export * from './supersetSpecAllOfSecurity';
export * from './supersetSpecAllOfStorage';
export * from './systemInfo';
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
export * from './tolerationsRules';
export * from './toolType';
export * from './training';
export * from './trainingCreationRequestV2';
export * from './trainingListResponse';
export * from './trainingSpecificRunParams';
export * from './trainingV2';
export * from './uidGidSource';
export * from './usageTimesInfo';
export * from './validateSyncRequest';
export * from './validateSyncResponse';
export * from './workloadCreationMeta';
export * from './workloadCreationRequest';
export * from './workloadDesiredPhase';
export * from './workloadMeta';
export * from './workloadMetaV1';
export * from './workloadRefAndStatus';
export * from './workloadState';
export * from './workloadStatusUpdate';
export * from './workloadSupportedTypes';
export * from './workloadTemplate';
export * from './workloadTemplateCreationRequest';
export * from './workloadTemplateListResponse';
export * from './workloadTemplateUpdateRequest';
export * from './workloadType';
export * from './workspace';
export * from './workspaceCreationRequestV2';
export * from './workspaceListResponse';
export * from './workspaceSpecificRunParams';
export * from './workspaceSpecificRunParamsAllOf';
export * from './workspaceV2';

import * as fs from 'fs';

export interface RequestDetailedFile {
    value: Buffer;
    options?: {
        filename?: string;
        contentType?: string;
    }
}

export type RequestFile = string | Buffer | fs.ReadStream | RequestDetailedFile;


import { AccessKey } from './accessKey';
import { AccessKeyCreationRequest } from './accessKeyCreationRequest';
import { AccessKeyCreationSpec } from './accessKeyCreationSpec';
import { AccessKeyListResponse } from './accessKeyListResponse';
import { AccessKeySpec } from './accessKeySpec';
import { AccessKeyUpdateRequest } from './accessKeyUpdateRequest';
import { ActivationReplicasField } from './activationReplicasField';
import { AllowOverQuota } from './allowOverQuota';
import { Annotation } from './annotation';
import { ArbitraryRules } from './arbitraryRules';
import { ArrayRules } from './arrayRules';
import { AssetClusterStatusInfo } from './assetClusterStatusInfo';
import { AssetClusterStatusIssue } from './assetClusterStatusIssue';
import { AssetClusterStatusIssueEnum } from './assetClusterStatusIssueEnum';
import { AssetCountResponse } from './assetCountResponse';
import { AssetCreationFields } from './assetCreationFields';
import { AssetCreationRequest } from './assetCreationRequest';
import { AssetDatasourceRef } from './assetDatasourceRef';
import { AssetFlatRequest } from './assetFlatRequest';
import { AssetFlatResponse } from './assetFlatResponse';
import { AssetIdAndKind } from './assetIdAndKind';
import { AssetKind } from './assetKind';
import { AssetListByIdsResponse } from './assetListByIdsResponse';
import { AssetMeta } from './assetMeta';
import { AssetMetaForSync } from './assetMetaForSync';
import { AssetMetaForSyncAllOf } from './assetMetaForSyncAllOf';
import { AssetMetaForSyncAllOf1 } from './assetMetaForSyncAllOf1';
import { AssetMetaForSyncAllOf2 } from './assetMetaForSyncAllOf2';
import { AssetPhaseUpdate } from './assetPhaseUpdate';
import { AssetReadOnlyFields } from './assetReadOnlyFields';
import { AssetRef } from './assetRef';
import { AssetResourcesIdsItem } from './assetResourcesIdsItem';
import { AssetResourcesIdsSyncRequest } from './assetResourcesIdsSyncRequest';
import { AssetSyncStatusEnum } from './assetSyncStatusEnum';
import { AssetUpdatableFields } from './assetUpdatableFields';
import { AssetUpdateRequest } from './assetUpdateRequest';
import { AssetUsageInfo } from './assetUsageInfo';
import { AssetsDefaults } from './assetsDefaults';
import { AssetsIds } from './assetsIds';
import { AssetsRef } from './assetsRef';
import { AssetsRefCompute } from './assetsRefCompute';
import { AssetsRefEnvironment } from './assetsRefEnvironment';
import { AssetsRefOptional } from './assetsRefOptional';
import { AssetsRules } from './assetsRules';
import { AssetsRulesDatasources } from './assetsRulesDatasources';
import { AssetsUsageRef } from './assetsUsageRef';
import { AssetsUsageRefCompute } from './assetsUsageRefCompute';
import { AssetsUsageRefEnvironment } from './assetsUsageRefEnvironment';
import { AttributeDefaults } from './attributeDefaults';
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
import { Capability } from './capability';
import { ClaimInfo } from './claimInfo';
import { ClaimInfoRules } from './claimInfoRules';
import { ClaimInfoUpdatable } from './claimInfoUpdatable';
import { CleanPodPolicy } from './cleanPodPolicy';
import { ClusterInfo } from './clusterInfo';
import { ClusterResourceInfo } from './clusterResourceInfo';
import { ClusterSyncReport } from './clusterSyncReport';
import { ClusterSyncReportUpdate } from './clusterSyncReportUpdate';
import { ComplianceInfo } from './complianceInfo';
import { ComplianceInfoReason } from './complianceInfoReason';
import { ComplianceIssuesV2 } from './complianceIssuesV2';
import { ComplianceIssuesV2ComplianceIssuesInner } from './complianceIssuesV2ComplianceIssuesInner';
import { ComputeAsset } from './computeAsset';
import { ComputeAssetSpec } from './computeAssetSpec';
import { ComputeCreationRequest } from './computeCreationRequest';
import { ComputeCreationRequestSpec } from './computeCreationRequestSpec';
import { ComputeListResponse } from './computeListResponse';
import { ComputeUpdateRequest } from './computeUpdateRequest';
import { ConcurrencyHardLimitField } from './concurrencyHardLimitField';
import { ConfigMap } from './configMap';
import { ConfigMapAsset } from './configMapAsset';
import { ConfigMapAssetSpec } from './configMapAssetSpec';
import { ConfigMapCreationRequest } from './configMapCreationRequest';
import { ConfigMapInstance } from './configMapInstance';
import { ConfigMapListResponse } from './configMapListResponse';
import { ConfigMapRules } from './configMapRules';
import { ConfigMapUpdateRequest } from './configMapUpdateRequest';
import { ConfigMapsRules } from './configMapsRules';
import { Connection } from './connection';
import { CredentialAsset } from './credentialAsset';
import { CredentialKind } from './credentialKind';
import { CredentialUpdatableFields } from './credentialUpdatableFields';
import { CredentialUpdateRequest } from './credentialUpdateRequest';
import { CredentialsListResponse } from './credentialsListResponse';
import { CredentialsListResponseAssetSpec } from './credentialsListResponseAssetSpec';
import { CredentialsListResponseAssetSpecAccessKey } from './credentialsListResponseAssetSpecAccessKey';
import { CredentialsListResponseAssetSpecDockerRegistry } from './credentialsListResponseAssetSpecDockerRegistry';
import { CredentialsListResponseAssetSpecGenericSecret } from './credentialsListResponseAssetSpecGenericSecret';
import { CredentialsListResponseAssetSpecNgcApiKey } from './credentialsListResponseAssetSpecNgcApiKey';
import { CredentialsListResponseAssetSpecPassword } from './credentialsListResponseAssetSpecPassword';
import { CredentialsListResponseEntry } from './credentialsListResponseEntry';
import { CredentialsUpdate } from './credentialsUpdate';
import { CredentialsUpdateSpec } from './credentialsUpdateSpec';
import { CredentialsUpdateSpecAccessKey } from './credentialsUpdateSpecAccessKey';
import { CredentialsUpdateSpecDockerRegistry } from './credentialsUpdateSpecDockerRegistry';
import { CredentialsUpdateSpecGenericSecret } from './credentialsUpdateSpecGenericSecret';
import { CredentialsUpdateSpecPassword } from './credentialsUpdateSpecPassword';
import { CredentialsUpdatesResp } from './credentialsUpdatesResp';
import { DataSourceOverrides } from './dataSourceOverrides';
import { DataVolume } from './dataVolume';
import { DataVolumeAssetSpec } from './dataVolumeAssetSpec';
import { DataVolumeInstance } from './dataVolumeInstance';
import { DataVolumeRules } from './dataVolumeRules';
import { DataVolumesRules } from './dataVolumesRules';
import { DatasourceDefaults } from './datasourceDefaults';
import { DatasourceListResponse } from './datasourceListResponse';
import { DatasourceListResponseAssetSpec } from './datasourceListResponseAssetSpec';
import { DatasourceListResponseAssetSpecConfigMap } from './datasourceListResponseAssetSpecConfigMap';
import { DatasourceListResponseAssetSpecDataVolume } from './datasourceListResponseAssetSpecDataVolume';
import { DatasourceListResponseAssetSpecGit } from './datasourceListResponseAssetSpecGit';
import { DatasourceListResponseAssetSpecHostPath } from './datasourceListResponseAssetSpecHostPath';
import { DatasourceListResponseAssetSpecNfs } from './datasourceListResponseAssetSpecNfs';
import { DatasourceListResponseAssetSpecPvc } from './datasourceListResponseAssetSpecPvc';
import { DatasourceListResponseAssetSpecS3 } from './datasourceListResponseAssetSpecS3';
import { DatasourceListResponseAssetSpecSecret } from './datasourceListResponseAssetSpecSecret';
import { DatasourceListResponseEntry } from './datasourceListResponseEntry';
import { DatasourceRef } from './datasourceRef';
import { DatasourceUpdate } from './datasourceUpdate';
import { DatasourceUpdateSpec } from './datasourceUpdateSpec';
import { DatasourceUpdateSpecPvc } from './datasourceUpdateSpecPvc';
import { DatasourceUpdatesResp } from './datasourceUpdatesResp';
import { DefaultFromRule } from './defaultFromRule';
import { DisplayedJob } from './displayedJob';
import { DistMaster } from './distMaster';
import { Distributed } from './distributed';
import { DistributedAllOf } from './distributedAllOf';
import { DistributedData } from './distributedData';
import { DistributedFramework } from './distributedFramework';
import { DistributedInferenceCommonSpec } from './distributedInferenceCommonSpec';
import { DistributedInferenceCommonSpecAllOf } from './distributedInferenceCommonSpecAllOf';
import { DistributedInferenceLeaderSpecFields } from './distributedInferenceLeaderSpecFields';
import { DistributedInferenceLeaderSpecFieldsLeader } from './distributedInferenceLeaderSpecFieldsLeader';
import { DistributedInferenceLeaderWorkerSpec } from './distributedInferenceLeaderWorkerSpec';
import { DistributedInferenceLeaderWorkerSpecCompute } from './distributedInferenceLeaderWorkerSpecCompute';
import { DistributedInferenceLeaderWorkerSpecSecurity } from './distributedInferenceLeaderWorkerSpecSecurity';
import { DistributedInferenceLeaderWorkerSpecStorage } from './distributedInferenceLeaderWorkerSpecStorage';
import { DistributedInferenceReplicasField } from './distributedInferenceReplicasField';
import { DistributedInferenceRestartPolicy } from './distributedInferenceRestartPolicy';
import { DistributedInferenceServingPort } from './distributedInferenceServingPort';
import { DistributedInferenceServingPortAccess } from './distributedInferenceServingPortAccess';
import { DistributedInferenceServingPortAccessAuthorizationTypeEnum } from './distributedInferenceServingPortAccessAuthorizationTypeEnum';
import { DistributedInferenceServingPortContainerAndProtocol } from './distributedInferenceServingPortContainerAndProtocol';
import { DistributedInferenceServingPortProtocol } from './distributedInferenceServingPortProtocol';
import { DistributedInferenceStartupPolicy } from './distributedInferenceStartupPolicy';
import { DistributedInferenceStartupPolicyField } from './distributedInferenceStartupPolicyField';
import { DistributedInferenceWorkerSpecFields } from './distributedInferenceWorkerSpecFields';
import { DistributedInferenceWorkerSpecFieldsWorker } from './distributedInferenceWorkerSpecFieldsWorker';
import { DistributedInferenceWorkersField } from './distributedInferenceWorkersField';
import { DistributedMaster } from './distributedMaster';
import { DistributedMasterData } from './distributedMasterData';
import { DistributedParams } from './distributedParams';
import { DockerRegistry } from './dockerRegistry';
import { DockerRegistryCreationRequest } from './dockerRegistryCreationRequest';
import { DockerRegistryCreationSpec } from './dockerRegistryCreationSpec';
import { DockerRegistryListResponse } from './dockerRegistryListResponse';
import { DockerRegistrySpec } from './dockerRegistrySpec';
import { DockerRegistryUpdateRequest } from './dockerRegistryUpdateRequest';
import { EmptyDir } from './emptyDir';
import { EmptyDirInstance } from './emptyDirInstance';
import { EmptyDirRules } from './emptyDirRules';
import { EmptyDirsRules } from './emptyDirsRules';
import { EncryptionKeyId } from './encryptionKeyId';
import { EnvironmentAsset } from './environmentAsset';
import { EnvironmentAssetRef } from './environmentAssetRef';
import { EnvironmentAssetSpec } from './environmentAssetSpec';
import { EnvironmentCreationRequest } from './environmentCreationRequest';
import { EnvironmentListResponse } from './environmentListResponse';
import { EnvironmentUpdateRequest } from './environmentUpdateRequest';
import { EnvironmentVariable } from './environmentVariable';
import { EnvironmentVariableConfigMap } from './environmentVariableConfigMap';
import { EnvironmentVariableCredential } from './environmentVariableCredential';
import { EnvironmentVariableOfAsset } from './environmentVariableOfAsset';
import { EnvironmentVariablePodFieldReference } from './environmentVariablePodFieldReference';
import { EnvironmentVariableSecret } from './environmentVariableSecret';
import { EnvironmentVariableUserCredential } from './environmentVariableUserCredential';
import { Event } from './event';
import { EventSyncFields } from './eventSyncFields';
import { EventsSyncRequest } from './eventsSyncRequest';
import { ExcludeField } from './excludeField';
import { ExposedUrl } from './exposedUrl';
import { ExtendedResource } from './extendedResource';
import { ExternalToolInfo } from './externalToolInfo';
import { ExternalToolType } from './externalToolType';
import { ExternalUrlInfo } from './externalUrlInfo';
import { GenericSecret } from './genericSecret';
import { GenericSecretCreationRequest } from './genericSecretCreationRequest';
import { GenericSecretCreationSpec } from './genericSecretCreationSpec';
import { GenericSecretListResponse } from './genericSecretListResponse';
import { GenericSecretSpec } from './genericSecretSpec';
import { GenericSecretUpdateRequest } from './genericSecretUpdateRequest';
import { GitAsset } from './gitAsset';
import { GitAssetSpec } from './gitAssetSpec';
import { GitAuthenticationMethod } from './gitAuthenticationMethod';
import { GitCommon } from './gitCommon';
import { GitCreationRequest } from './gitCreationRequest';
import { GitCredentialId } from './gitCredentialId';
import { GitInstance } from './gitInstance';
import { GitInstanceAllOf } from './gitInstanceAllOf';
import { GitListResponse } from './gitListResponse';
import { GitPassword } from './gitPassword';
import { GitPasswordAsset } from './gitPasswordAsset';
import { GitRules } from './gitRules';
import { GitSecretRef } from './gitSecretRef';
import { GitUpdateRequest } from './gitUpdateRequest';
import { GitsRules } from './gitsRules';
import { GpuRequestType } from './gpuRequestType';
import { HistoryRecord } from './historyRecord';
import { HistoryRecordMeta } from './historyRecordMeta';
import { HistoryRecordSpec } from './historyRecordSpec';
import { HistoryRecordType } from './historyRecordType';
import { HostPath } from './hostPath';
import { HostPathAsset } from './hostPathAsset';
import { HostPathCreationRequest } from './hostPathCreationRequest';
import { HostPathInstance } from './hostPathInstance';
import { HostPathListResponse } from './hostPathListResponse';
import { HostPathMountPropagation } from './hostPathMountPropagation';
import { HostPathMountPropagationOptions } from './hostPathMountPropagationOptions';
import { HostPathMountPropagationOptionsOptionsInner } from './hostPathMountPropagationOptionsOptionsInner';
import { HostPathMountPropagationRules } from './hostPathMountPropagationRules';
import { HostPathRules } from './hostPathRules';
import { HostPathSpec } from './hostPathSpec';
import { HostPathUpdateRequest } from './hostPathUpdateRequest';
import { HostPathsRules } from './hostPathsRules';
import { HttpResponse } from './httpResponse';
import { HuggingFaceModelInfo } from './huggingFaceModelInfo';
import { HuggingFaceModelInfoRequest } from './huggingFaceModelInfoRequest';
import { ImagePullPolicy } from './imagePullPolicy';
import { ImagePullSecret } from './imagePullSecret';
import { ImagePullSecretsRules } from './imagePullSecretsRules';
import { ImagePullSecretsRulesAttributes } from './imagePullSecretsRulesAttributes';
import { Inference } from './inference';
import { InferenceCreationRequestV2 } from './inferenceCreationRequestV2';
import { InferenceFields } from './inferenceFields';
import { InferenceFieldsRules } from './inferenceFieldsRules';
import { InferenceSpecificRunParams } from './inferenceSpecificRunParams';
import { InferenceV2 } from './inferenceV2';
import { InferencesMaxReplicasField } from './inferencesMaxReplicasField';
import { InferencesMinReplicasField } from './inferencesMinReplicasField';
import { InfoDistMaster } from './infoDistMaster';
import { InfoDistributed } from './infoDistributed';
import { InfoDistributedAllOf } from './infoDistributedAllOf';
import { InitialReplicasField } from './initialReplicasField';
import { InitializationTimeoutField } from './initializationTimeoutField';
import { InstancesRules } from './instancesRules';
import { IntegerRules } from './integerRules';
import { IntegerRulesOptional } from './integerRulesOptional';
import { InternalConnectionType } from './internalConnectionType';
import { InternalToolInfo } from './internalToolInfo';
import { InternalToolType } from './internalToolType';
import { InvolvedObject } from './involvedObject';
import { ItemRules } from './itemRules';
import { KeyValuePair } from './keyValuePair';
import { Label } from './label';
import { MasterSpec } from './masterSpec';
import { MasterSpecSecurity } from './masterSpecSecurity';
import { MasterSpecStorage } from './masterSpecStorage';
import { MatchExpression } from './matchExpression';
import { MatchExpressionOperator } from './matchExpressionOperator';
import { MetricThresholdField } from './metricThresholdField';
import { MetricThresholdPercentageField } from './metricThresholdPercentageField';
import { ModelAsset } from './modelAsset';
import { ModelAssetSpec } from './modelAssetSpec';
import { ModelCreationRequest } from './modelCreationRequest';
import { ModelError } from './modelError';
import { ModelInference } from './modelInference';
import { ModelInferenceCreationRequest } from './modelInferenceCreationRequest';
import { ModelInferenceCreationRequestAssets } from './modelInferenceCreationRequestAssets';
import { ModelInferenceCreationRequestV2 } from './modelInferenceCreationRequestV2';
import { ModelInferenceCreationRequestV2Assets } from './modelInferenceCreationRequestV2Assets';
import { ModelInferenceV2 } from './modelInferenceV2';
import { ModelListResponse } from './modelListResponse';
import { ModelUpdateRequest } from './modelUpdateRequest';
import { MpiLauncherCreationPolicy } from './mpiLauncherCreationPolicy';
import { NFSAsset } from './nFSAsset';
import { NFSAssetSpec } from './nFSAssetSpec';
import { NFSCreationRequest } from './nFSCreationRequest';
import { NFSListResponse } from './nFSListResponse';
import { NFSUpdateRequest } from './nFSUpdateRequest';
import { Nfs } from './nfs';
import { NfsInstance } from './nfsInstance';
import { NfsRules } from './nfsRules';
import { NfssRules } from './nfssRules';
import { NgcApiKey } from './ngcApiKey';
import { NgcApiKeyCreationRequest } from './ngcApiKeyCreationRequest';
import { NgcApiKeyCreationSpec } from './ngcApiKeyCreationSpec';
import { NgcApiKeyInternal } from './ngcApiKeyInternal';
import { NgcApiKeyListResponse } from './ngcApiKeyListResponse';
import { NgcApiKeySpec } from './ngcApiKeySpec';
import { NgcApiKeyUpdateRequest } from './ngcApiKeyUpdateRequest';
import { NodeAffinityRequired } from './nodeAffinityRequired';
import { NodePortInfo } from './nodePortInfo';
import { NodeSelectorTerm } from './nodeSelectorTerm';
import { NonOverridableSpecFields } from './nonOverridableSpecFields';
import { NonOverridableSpecFieldsAllOf } from './nonOverridableSpecFieldsAllOf';
import { NonOverridableSpecFieldsAllOf1 } from './nonOverridableSpecFieldsAllOf1';
import { NumberRules } from './numberRules';
import { OptionalAssetIds } from './optionalAssetIds';
import { OverridableSpecFields } from './overridableSpecFields';
import { PVCAsset } from './pVCAsset';
import { PVCAssetSpec } from './pVCAssetSpec';
import { PVCCreationRequest } from './pVCCreationRequest';
import { PVCHistoryResponse } from './pVCHistoryResponse';
import { PVCListResponse } from './pVCListResponse';
import { PVCUpdateRequest } from './pVCUpdateRequest';
import { Password } from './password';
import { PasswordCreationRequest } from './passwordCreationRequest';
import { PasswordCreationSpec } from './passwordCreationSpec';
import { PasswordListResponse } from './passwordListResponse';
import { PasswordSpec } from './passwordSpec';
import { PasswordUpdateRequest } from './passwordUpdateRequest';
import { Phase } from './phase';
import { PodAffinity } from './podAffinity';
import { PodAffinityRules } from './podAffinityRules';
import { PodAffinityType } from './podAffinityType';
import { PodAffinityTypeOptions } from './podAffinityTypeOptions';
import { PodAffinityTypeOptionsOptionsInner } from './podAffinityTypeOptionsOptionsInner';
import { PodAffinityTypeRules } from './podAffinityTypeRules';
import { PolicyInfo } from './policyInfo';
import { PolicyInfoAllOf } from './policyInfoAllOf';
import { PolicyInfoAllOf1 } from './policyInfoAllOf1';
import { PolicyRuleEnum } from './policyRuleEnum';
import { Port } from './port';
import { PortServiceType } from './portServiceType';
import { Preemptibility } from './preemptibility';
import { PreemptibilityOptions } from './preemptibilityOptions';
import { PreemptibilityOptionsOptionsInner } from './preemptibilityOptionsOptionsInner';
import { PreemptibilityRules } from './preemptibilityRules';
import { Probe } from './probe';
import { ProbeHandler } from './probeHandler';
import { ProbeHandlerHttpGet } from './probeHandlerHttpGet';
import { ProbeHandlerScheme } from './probeHandlerScheme';
import { Probes } from './probes';
import { Pvc } from './pvc';
import { PvcAccessModes } from './pvcAccessModes';
import { PvcAccessModesRules } from './pvcAccessModesRules';
import { PvcAddedAttrValue } from './pvcAddedAttrValue';
import { PvcFieldsNonUpdatable } from './pvcFieldsNonUpdatable';
import { PvcFieldsUpdatable } from './pvcFieldsUpdatable';
import { PvcInstance } from './pvcInstance';
import { PvcRules } from './pvcRules';
import { PvcVolumeMode } from './pvcVolumeMode';
import { PvcsRules } from './pvcsRules';
import { Registry } from './registry';
import { RegistryCreationRequest } from './registryCreationRequest';
import { RegistryListResponse } from './registryListResponse';
import { RegistrySpec } from './registrySpec';
import { RegistryUpdateRequest } from './registryUpdateRequest';
import { RelatedUrl } from './relatedUrl';
import { Repositories } from './repositories';
import { RepositoryTags } from './repositoryTags';
import { RequestInfo } from './requestInfo';
import { ResponseInfo } from './responseInfo';
import { RestartPolicy } from './restartPolicy';
import { RestartPolicyOptions } from './restartPolicyOptions';
import { RestartPolicyOptionsOptionsInner } from './restartPolicyOptionsOptionsInner';
import { RestartPolicyRule } from './restartPolicyRule';
import { Rules } from './rules';
import { S3AccessKey } from './s3AccessKey';
import { S3AccessKeyAsset } from './s3AccessKeyAsset';
import { S3Asset } from './s3Asset';
import { S3AssetSpec } from './s3AssetSpec';
import { S3Common } from './s3Common';
import { S3CreationRequest } from './s3CreationRequest';
import { S3Instance } from './s3Instance';
import { S3ListResponse } from './s3ListResponse';
import { S3Rules } from './s3Rules';
import { S3UpdateRequest } from './s3UpdateRequest';
import { S3sRules } from './s3sRules';
import { ScaleDownDelayField } from './scaleDownDelayField';
import { ScaleToZeroRetentionField } from './scaleToZeroRetentionField';
import { Scope } from './scope';
import { ScopeType } from './scopeType';
import { SeccompProfileType } from './seccompProfileType';
import { Secret } from './secret';
import { SecretAsset } from './secretAsset';
import { SecretAssetCreationRequest } from './secretAssetCreationRequest';
import { SecretAssetListResponse } from './secretAssetListResponse';
import { SecretAssetSpec } from './secretAssetSpec';
import { SecretAssetSpecAllOf } from './secretAssetSpecAllOf';
import { SecretAssetUpdateRequest } from './secretAssetUpdateRequest';
import { SecretFieldsNonUpdatable } from './secretFieldsNonUpdatable';
import { SecretFieldsUpdatable } from './secretFieldsUpdatable';
import { SecretInstance } from './secretInstance';
import { SecretRules } from './secretRules';
import { SecretsRules } from './secretsRules';
import { ServingConfiguration } from './servingConfiguration';
import { ServingConfigurationField } from './servingConfigurationField';
import { ServingConfigurationRules } from './servingConfigurationRules';
import { ServingPortAccess } from './servingPortAccess';
import { ServingPortAccessAuthorizationTypeEnum } from './servingPortAccessAuthorizationTypeEnum';
import { ServingPortInfo } from './servingPortInfo';
import { ServingPortProtocol } from './servingPortProtocol';
import { SourceOfRule } from './sourceOfRule';
import { SpecificRunAutoScaling } from './specificRunAutoScaling';
import { SpecificRunAutoScalingAutoScaling } from './specificRunAutoScalingAutoScaling';
import { SpecificRunConnectionInfo } from './specificRunConnectionInfo';
import { SpecificRunConnections } from './specificRunConnections';
import { SpecificRunCreationFields } from './specificRunCreationFields';
import { SpecificRunInfoFields } from './specificRunInfoFields';
import { SpecificRunMetricFields } from './specificRunMetricFields';
import { SpecificRunParamAttrDefaults } from './specificRunParamAttrDefaults';
import { SpecificRunParams } from './specificRunParams';
import { SpecificRunParamsAllOf } from './specificRunParamsAllOf';
import { SpecificRunParamsRules } from './specificRunParamsRules';
import { SpecificRunParamsRulesAllOf } from './specificRunParamsRulesAllOf';
import { SpecificRunServingPortAccess } from './specificRunServingPortAccess';
import { SpecificRunServingPortAccessServingPortAccess } from './specificRunServingPortAccessServingPortAccess';
import { StorageInstanceName } from './storageInstanceName';
import { StringOption } from './stringOption';
import { StringRules } from './stringRules';
import { StringRulesCommon } from './stringRulesCommon';
import { StringRulesOptions } from './stringRulesOptions';
import { SubmissionErrorV2 } from './submissionErrorV2';
import { SupersetServingPort } from './supersetServingPort';
import { SupersetSpec } from './supersetSpec';
import { SupersetSpecAllOf } from './supersetSpecAllOf';
import { SupersetSpecAllOfCompute } from './supersetSpecAllOfCompute';
import { SupersetSpecAllOfSecurity } from './supersetSpecAllOfSecurity';
import { SupersetSpecAllOfStorage } from './supersetSpecAllOfStorage';
import { SystemInfo } from './systemInfo';
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
import { TolerationsRules } from './tolerationsRules';
import { ToolType } from './toolType';
import { Training } from './training';
import { TrainingCreationRequestV2 } from './trainingCreationRequestV2';
import { TrainingListResponse } from './trainingListResponse';
import { TrainingSpecificRunParams } from './trainingSpecificRunParams';
import { TrainingV2 } from './trainingV2';
import { UidGidSource } from './uidGidSource';
import { UsageTimesInfo } from './usageTimesInfo';
import { ValidateSyncRequest } from './validateSyncRequest';
import { ValidateSyncResponse } from './validateSyncResponse';
import { WorkloadCreationMeta } from './workloadCreationMeta';
import { WorkloadCreationRequest } from './workloadCreationRequest';
import { WorkloadDesiredPhase } from './workloadDesiredPhase';
import { WorkloadMeta } from './workloadMeta';
import { WorkloadMetaV1 } from './workloadMetaV1';
import { WorkloadRefAndStatus } from './workloadRefAndStatus';
import { WorkloadState } from './workloadState';
import { WorkloadStatusUpdate } from './workloadStatusUpdate';
import { WorkloadSupportedTypes } from './workloadSupportedTypes';
import { WorkloadTemplate } from './workloadTemplate';
import { WorkloadTemplateCreationRequest } from './workloadTemplateCreationRequest';
import { WorkloadTemplateListResponse } from './workloadTemplateListResponse';
import { WorkloadTemplateUpdateRequest } from './workloadTemplateUpdateRequest';
import { WorkloadType } from './workloadType';
import { Workspace } from './workspace';
import { WorkspaceCreationRequestV2 } from './workspaceCreationRequestV2';
import { WorkspaceListResponse } from './workspaceListResponse';
import { WorkspaceSpecificRunParams } from './workspaceSpecificRunParams';
import { WorkspaceSpecificRunParamsAllOf } from './workspaceSpecificRunParamsAllOf';
import { WorkspaceV2 } from './workspaceV2';

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
        "AssetClusterStatusIssueEnum": AssetClusterStatusIssueEnum,
        "AssetKind": AssetKind,
        "AssetSyncStatusEnum": AssetSyncStatusEnum,
        "AuthorizationType": AuthorizationType,
        "Capability": Capability,
        "CleanPodPolicy": CleanPodPolicy,
        "CredentialKind": CredentialKind,
        "DistributedFramework": DistributedFramework,
        "DistributedInferenceRestartPolicy": DistributedInferenceRestartPolicy,
        "DistributedInferenceServingPortAccessAuthorizationTypeEnum": DistributedInferenceServingPortAccessAuthorizationTypeEnum,
        "DistributedInferenceServingPortProtocol": DistributedInferenceServingPortProtocol,
        "DistributedInferenceStartupPolicy": DistributedInferenceStartupPolicy,
        "ExternalToolType": ExternalToolType,
        "GitAuthenticationMethod": GitAuthenticationMethod,
        "GpuRequestType": GpuRequestType,
        "HistoryRecordType": HistoryRecordType,
        "HostPathMountPropagation": HostPathMountPropagation,
        "ImagePullPolicy": ImagePullPolicy,
        "InternalConnectionType": InternalConnectionType,
        "InternalToolType": InternalToolType,
        "MatchExpressionOperator": MatchExpressionOperator,
        "MpiLauncherCreationPolicy": MpiLauncherCreationPolicy,
        "Phase": Phase,
        "PodAffinityType": PodAffinityType,
        "PolicyRuleEnum": PolicyRuleEnum,
        "PortServiceType": PortServiceType,
        "Preemptibility": Preemptibility,
        "ProbeHandlerScheme": ProbeHandlerScheme,
        "PvcVolumeMode": PvcVolumeMode,
        "RestartPolicy": RestartPolicy,
        "Scope": Scope,
        "ScopeType": ScopeType,
        "SeccompProfileType": SeccompProfileType,
        "ServingPortAccessAuthorizationTypeEnum": ServingPortAccessAuthorizationTypeEnum,
        "ServingPortInfo.ProtocolEnum": ServingPortInfo.ProtocolEnum,
        "ServingPortProtocol": ServingPortProtocol,
        "TolerationEffect": TolerationEffect,
        "TolerationOperator": TolerationOperator,
        "ToolType": ToolType,
        "UidGidSource": UidGidSource,
        "WorkloadDesiredPhase": WorkloadDesiredPhase,
        "WorkloadState": WorkloadState,
        "WorkloadSupportedTypes.DistFrameworkEnum": WorkloadSupportedTypes.DistFrameworkEnum,
        "WorkloadType": WorkloadType,
}

let typeMap: {[index: string]: any} = {
    "AccessKey": AccessKey,
    "AccessKeyCreationRequest": AccessKeyCreationRequest,
    "AccessKeyCreationSpec": AccessKeyCreationSpec,
    "AccessKeyListResponse": AccessKeyListResponse,
    "AccessKeySpec": AccessKeySpec,
    "AccessKeyUpdateRequest": AccessKeyUpdateRequest,
    "ActivationReplicasField": ActivationReplicasField,
    "AllowOverQuota": AllowOverQuota,
    "Annotation": Annotation,
    "ArbitraryRules": ArbitraryRules,
    "ArrayRules": ArrayRules,
    "AssetClusterStatusInfo": AssetClusterStatusInfo,
    "AssetClusterStatusIssue": AssetClusterStatusIssue,
    "AssetCountResponse": AssetCountResponse,
    "AssetCreationFields": AssetCreationFields,
    "AssetCreationRequest": AssetCreationRequest,
    "AssetDatasourceRef": AssetDatasourceRef,
    "AssetFlatRequest": AssetFlatRequest,
    "AssetFlatResponse": AssetFlatResponse,
    "AssetIdAndKind": AssetIdAndKind,
    "AssetListByIdsResponse": AssetListByIdsResponse,
    "AssetMeta": AssetMeta,
    "AssetMetaForSync": AssetMetaForSync,
    "AssetMetaForSyncAllOf": AssetMetaForSyncAllOf,
    "AssetMetaForSyncAllOf1": AssetMetaForSyncAllOf1,
    "AssetMetaForSyncAllOf2": AssetMetaForSyncAllOf2,
    "AssetPhaseUpdate": AssetPhaseUpdate,
    "AssetReadOnlyFields": AssetReadOnlyFields,
    "AssetRef": AssetRef,
    "AssetResourcesIdsItem": AssetResourcesIdsItem,
    "AssetResourcesIdsSyncRequest": AssetResourcesIdsSyncRequest,
    "AssetUpdatableFields": AssetUpdatableFields,
    "AssetUpdateRequest": AssetUpdateRequest,
    "AssetUsageInfo": AssetUsageInfo,
    "AssetsDefaults": AssetsDefaults,
    "AssetsIds": AssetsIds,
    "AssetsRef": AssetsRef,
    "AssetsRefCompute": AssetsRefCompute,
    "AssetsRefEnvironment": AssetsRefEnvironment,
    "AssetsRefOptional": AssetsRefOptional,
    "AssetsRules": AssetsRules,
    "AssetsRulesDatasources": AssetsRulesDatasources,
    "AssetsUsageRef": AssetsUsageRef,
    "AssetsUsageRefCompute": AssetsUsageRefCompute,
    "AssetsUsageRefEnvironment": AssetsUsageRefEnvironment,
    "AttributeDefaults": AttributeDefaults,
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
    "ClaimInfo": ClaimInfo,
    "ClaimInfoRules": ClaimInfoRules,
    "ClaimInfoUpdatable": ClaimInfoUpdatable,
    "ClusterInfo": ClusterInfo,
    "ClusterResourceInfo": ClusterResourceInfo,
    "ClusterSyncReport": ClusterSyncReport,
    "ClusterSyncReportUpdate": ClusterSyncReportUpdate,
    "ComplianceInfo": ComplianceInfo,
    "ComplianceInfoReason": ComplianceInfoReason,
    "ComplianceIssuesV2": ComplianceIssuesV2,
    "ComplianceIssuesV2ComplianceIssuesInner": ComplianceIssuesV2ComplianceIssuesInner,
    "ComputeAsset": ComputeAsset,
    "ComputeAssetSpec": ComputeAssetSpec,
    "ComputeCreationRequest": ComputeCreationRequest,
    "ComputeCreationRequestSpec": ComputeCreationRequestSpec,
    "ComputeListResponse": ComputeListResponse,
    "ComputeUpdateRequest": ComputeUpdateRequest,
    "ConcurrencyHardLimitField": ConcurrencyHardLimitField,
    "ConfigMap": ConfigMap,
    "ConfigMapAsset": ConfigMapAsset,
    "ConfigMapAssetSpec": ConfigMapAssetSpec,
    "ConfigMapCreationRequest": ConfigMapCreationRequest,
    "ConfigMapInstance": ConfigMapInstance,
    "ConfigMapListResponse": ConfigMapListResponse,
    "ConfigMapRules": ConfigMapRules,
    "ConfigMapUpdateRequest": ConfigMapUpdateRequest,
    "ConfigMapsRules": ConfigMapsRules,
    "Connection": Connection,
    "CredentialAsset": CredentialAsset,
    "CredentialUpdatableFields": CredentialUpdatableFields,
    "CredentialUpdateRequest": CredentialUpdateRequest,
    "CredentialsListResponse": CredentialsListResponse,
    "CredentialsListResponseAssetSpec": CredentialsListResponseAssetSpec,
    "CredentialsListResponseAssetSpecAccessKey": CredentialsListResponseAssetSpecAccessKey,
    "CredentialsListResponseAssetSpecDockerRegistry": CredentialsListResponseAssetSpecDockerRegistry,
    "CredentialsListResponseAssetSpecGenericSecret": CredentialsListResponseAssetSpecGenericSecret,
    "CredentialsListResponseAssetSpecNgcApiKey": CredentialsListResponseAssetSpecNgcApiKey,
    "CredentialsListResponseAssetSpecPassword": CredentialsListResponseAssetSpecPassword,
    "CredentialsListResponseEntry": CredentialsListResponseEntry,
    "CredentialsUpdate": CredentialsUpdate,
    "CredentialsUpdateSpec": CredentialsUpdateSpec,
    "CredentialsUpdateSpecAccessKey": CredentialsUpdateSpecAccessKey,
    "CredentialsUpdateSpecDockerRegistry": CredentialsUpdateSpecDockerRegistry,
    "CredentialsUpdateSpecGenericSecret": CredentialsUpdateSpecGenericSecret,
    "CredentialsUpdateSpecPassword": CredentialsUpdateSpecPassword,
    "CredentialsUpdatesResp": CredentialsUpdatesResp,
    "DataSourceOverrides": DataSourceOverrides,
    "DataVolume": DataVolume,
    "DataVolumeAssetSpec": DataVolumeAssetSpec,
    "DataVolumeInstance": DataVolumeInstance,
    "DataVolumeRules": DataVolumeRules,
    "DataVolumesRules": DataVolumesRules,
    "DatasourceDefaults": DatasourceDefaults,
    "DatasourceListResponse": DatasourceListResponse,
    "DatasourceListResponseAssetSpec": DatasourceListResponseAssetSpec,
    "DatasourceListResponseAssetSpecConfigMap": DatasourceListResponseAssetSpecConfigMap,
    "DatasourceListResponseAssetSpecDataVolume": DatasourceListResponseAssetSpecDataVolume,
    "DatasourceListResponseAssetSpecGit": DatasourceListResponseAssetSpecGit,
    "DatasourceListResponseAssetSpecHostPath": DatasourceListResponseAssetSpecHostPath,
    "DatasourceListResponseAssetSpecNfs": DatasourceListResponseAssetSpecNfs,
    "DatasourceListResponseAssetSpecPvc": DatasourceListResponseAssetSpecPvc,
    "DatasourceListResponseAssetSpecS3": DatasourceListResponseAssetSpecS3,
    "DatasourceListResponseAssetSpecSecret": DatasourceListResponseAssetSpecSecret,
    "DatasourceListResponseEntry": DatasourceListResponseEntry,
    "DatasourceRef": DatasourceRef,
    "DatasourceUpdate": DatasourceUpdate,
    "DatasourceUpdateSpec": DatasourceUpdateSpec,
    "DatasourceUpdateSpecPvc": DatasourceUpdateSpecPvc,
    "DatasourceUpdatesResp": DatasourceUpdatesResp,
    "DefaultFromRule": DefaultFromRule,
    "DisplayedJob": DisplayedJob,
    "DistMaster": DistMaster,
    "Distributed": Distributed,
    "DistributedAllOf": DistributedAllOf,
    "DistributedData": DistributedData,
    "DistributedInferenceCommonSpec": DistributedInferenceCommonSpec,
    "DistributedInferenceCommonSpecAllOf": DistributedInferenceCommonSpecAllOf,
    "DistributedInferenceLeaderSpecFields": DistributedInferenceLeaderSpecFields,
    "DistributedInferenceLeaderSpecFieldsLeader": DistributedInferenceLeaderSpecFieldsLeader,
    "DistributedInferenceLeaderWorkerSpec": DistributedInferenceLeaderWorkerSpec,
    "DistributedInferenceLeaderWorkerSpecCompute": DistributedInferenceLeaderWorkerSpecCompute,
    "DistributedInferenceLeaderWorkerSpecSecurity": DistributedInferenceLeaderWorkerSpecSecurity,
    "DistributedInferenceLeaderWorkerSpecStorage": DistributedInferenceLeaderWorkerSpecStorage,
    "DistributedInferenceReplicasField": DistributedInferenceReplicasField,
    "DistributedInferenceServingPort": DistributedInferenceServingPort,
    "DistributedInferenceServingPortAccess": DistributedInferenceServingPortAccess,
    "DistributedInferenceServingPortContainerAndProtocol": DistributedInferenceServingPortContainerAndProtocol,
    "DistributedInferenceStartupPolicyField": DistributedInferenceStartupPolicyField,
    "DistributedInferenceWorkerSpecFields": DistributedInferenceWorkerSpecFields,
    "DistributedInferenceWorkerSpecFieldsWorker": DistributedInferenceWorkerSpecFieldsWorker,
    "DistributedInferenceWorkersField": DistributedInferenceWorkersField,
    "DistributedMaster": DistributedMaster,
    "DistributedMasterData": DistributedMasterData,
    "DistributedParams": DistributedParams,
    "DockerRegistry": DockerRegistry,
    "DockerRegistryCreationRequest": DockerRegistryCreationRequest,
    "DockerRegistryCreationSpec": DockerRegistryCreationSpec,
    "DockerRegistryListResponse": DockerRegistryListResponse,
    "DockerRegistrySpec": DockerRegistrySpec,
    "DockerRegistryUpdateRequest": DockerRegistryUpdateRequest,
    "EmptyDir": EmptyDir,
    "EmptyDirInstance": EmptyDirInstance,
    "EmptyDirRules": EmptyDirRules,
    "EmptyDirsRules": EmptyDirsRules,
    "EncryptionKeyId": EncryptionKeyId,
    "EnvironmentAsset": EnvironmentAsset,
    "EnvironmentAssetRef": EnvironmentAssetRef,
    "EnvironmentAssetSpec": EnvironmentAssetSpec,
    "EnvironmentCreationRequest": EnvironmentCreationRequest,
    "EnvironmentListResponse": EnvironmentListResponse,
    "EnvironmentUpdateRequest": EnvironmentUpdateRequest,
    "EnvironmentVariable": EnvironmentVariable,
    "EnvironmentVariableConfigMap": EnvironmentVariableConfigMap,
    "EnvironmentVariableCredential": EnvironmentVariableCredential,
    "EnvironmentVariableOfAsset": EnvironmentVariableOfAsset,
    "EnvironmentVariablePodFieldReference": EnvironmentVariablePodFieldReference,
    "EnvironmentVariableSecret": EnvironmentVariableSecret,
    "EnvironmentVariableUserCredential": EnvironmentVariableUserCredential,
    "Event": Event,
    "EventSyncFields": EventSyncFields,
    "EventsSyncRequest": EventsSyncRequest,
    "ExcludeField": ExcludeField,
    "ExposedUrl": ExposedUrl,
    "ExtendedResource": ExtendedResource,
    "ExternalToolInfo": ExternalToolInfo,
    "ExternalUrlInfo": ExternalUrlInfo,
    "GenericSecret": GenericSecret,
    "GenericSecretCreationRequest": GenericSecretCreationRequest,
    "GenericSecretCreationSpec": GenericSecretCreationSpec,
    "GenericSecretListResponse": GenericSecretListResponse,
    "GenericSecretSpec": GenericSecretSpec,
    "GenericSecretUpdateRequest": GenericSecretUpdateRequest,
    "GitAsset": GitAsset,
    "GitAssetSpec": GitAssetSpec,
    "GitCommon": GitCommon,
    "GitCreationRequest": GitCreationRequest,
    "GitCredentialId": GitCredentialId,
    "GitInstance": GitInstance,
    "GitInstanceAllOf": GitInstanceAllOf,
    "GitListResponse": GitListResponse,
    "GitPassword": GitPassword,
    "GitPasswordAsset": GitPasswordAsset,
    "GitRules": GitRules,
    "GitSecretRef": GitSecretRef,
    "GitUpdateRequest": GitUpdateRequest,
    "GitsRules": GitsRules,
    "HistoryRecord": HistoryRecord,
    "HistoryRecordMeta": HistoryRecordMeta,
    "HistoryRecordSpec": HistoryRecordSpec,
    "HostPath": HostPath,
    "HostPathAsset": HostPathAsset,
    "HostPathCreationRequest": HostPathCreationRequest,
    "HostPathInstance": HostPathInstance,
    "HostPathListResponse": HostPathListResponse,
    "HostPathMountPropagationOptions": HostPathMountPropagationOptions,
    "HostPathMountPropagationOptionsOptionsInner": HostPathMountPropagationOptionsOptionsInner,
    "HostPathMountPropagationRules": HostPathMountPropagationRules,
    "HostPathRules": HostPathRules,
    "HostPathSpec": HostPathSpec,
    "HostPathUpdateRequest": HostPathUpdateRequest,
    "HostPathsRules": HostPathsRules,
    "HttpResponse": HttpResponse,
    "HuggingFaceModelInfo": HuggingFaceModelInfo,
    "HuggingFaceModelInfoRequest": HuggingFaceModelInfoRequest,
    "ImagePullSecret": ImagePullSecret,
    "ImagePullSecretsRules": ImagePullSecretsRules,
    "ImagePullSecretsRulesAttributes": ImagePullSecretsRulesAttributes,
    "Inference": Inference,
    "InferenceCreationRequestV2": InferenceCreationRequestV2,
    "InferenceFields": InferenceFields,
    "InferenceFieldsRules": InferenceFieldsRules,
    "InferenceSpecificRunParams": InferenceSpecificRunParams,
    "InferenceV2": InferenceV2,
    "InferencesMaxReplicasField": InferencesMaxReplicasField,
    "InferencesMinReplicasField": InferencesMinReplicasField,
    "InfoDistMaster": InfoDistMaster,
    "InfoDistributed": InfoDistributed,
    "InfoDistributedAllOf": InfoDistributedAllOf,
    "InitialReplicasField": InitialReplicasField,
    "InitializationTimeoutField": InitializationTimeoutField,
    "InstancesRules": InstancesRules,
    "IntegerRules": IntegerRules,
    "IntegerRulesOptional": IntegerRulesOptional,
    "InternalToolInfo": InternalToolInfo,
    "InvolvedObject": InvolvedObject,
    "ItemRules": ItemRules,
    "KeyValuePair": KeyValuePair,
    "Label": Label,
    "MasterSpec": MasterSpec,
    "MasterSpecSecurity": MasterSpecSecurity,
    "MasterSpecStorage": MasterSpecStorage,
    "MatchExpression": MatchExpression,
    "MetricThresholdField": MetricThresholdField,
    "MetricThresholdPercentageField": MetricThresholdPercentageField,
    "ModelAsset": ModelAsset,
    "ModelAssetSpec": ModelAssetSpec,
    "ModelCreationRequest": ModelCreationRequest,
    "ModelError": ModelError,
    "ModelInference": ModelInference,
    "ModelInferenceCreationRequest": ModelInferenceCreationRequest,
    "ModelInferenceCreationRequestAssets": ModelInferenceCreationRequestAssets,
    "ModelInferenceCreationRequestV2": ModelInferenceCreationRequestV2,
    "ModelInferenceCreationRequestV2Assets": ModelInferenceCreationRequestV2Assets,
    "ModelInferenceV2": ModelInferenceV2,
    "ModelListResponse": ModelListResponse,
    "ModelUpdateRequest": ModelUpdateRequest,
    "NFSAsset": NFSAsset,
    "NFSAssetSpec": NFSAssetSpec,
    "NFSCreationRequest": NFSCreationRequest,
    "NFSListResponse": NFSListResponse,
    "NFSUpdateRequest": NFSUpdateRequest,
    "Nfs": Nfs,
    "NfsInstance": NfsInstance,
    "NfsRules": NfsRules,
    "NfssRules": NfssRules,
    "NgcApiKey": NgcApiKey,
    "NgcApiKeyCreationRequest": NgcApiKeyCreationRequest,
    "NgcApiKeyCreationSpec": NgcApiKeyCreationSpec,
    "NgcApiKeyInternal": NgcApiKeyInternal,
    "NgcApiKeyListResponse": NgcApiKeyListResponse,
    "NgcApiKeySpec": NgcApiKeySpec,
    "NgcApiKeyUpdateRequest": NgcApiKeyUpdateRequest,
    "NodeAffinityRequired": NodeAffinityRequired,
    "NodePortInfo": NodePortInfo,
    "NodeSelectorTerm": NodeSelectorTerm,
    "NonOverridableSpecFields": NonOverridableSpecFields,
    "NonOverridableSpecFieldsAllOf": NonOverridableSpecFieldsAllOf,
    "NonOverridableSpecFieldsAllOf1": NonOverridableSpecFieldsAllOf1,
    "NumberRules": NumberRules,
    "OptionalAssetIds": OptionalAssetIds,
    "OverridableSpecFields": OverridableSpecFields,
    "PVCAsset": PVCAsset,
    "PVCAssetSpec": PVCAssetSpec,
    "PVCCreationRequest": PVCCreationRequest,
    "PVCHistoryResponse": PVCHistoryResponse,
    "PVCListResponse": PVCListResponse,
    "PVCUpdateRequest": PVCUpdateRequest,
    "Password": Password,
    "PasswordCreationRequest": PasswordCreationRequest,
    "PasswordCreationSpec": PasswordCreationSpec,
    "PasswordListResponse": PasswordListResponse,
    "PasswordSpec": PasswordSpec,
    "PasswordUpdateRequest": PasswordUpdateRequest,
    "PodAffinity": PodAffinity,
    "PodAffinityRules": PodAffinityRules,
    "PodAffinityTypeOptions": PodAffinityTypeOptions,
    "PodAffinityTypeOptionsOptionsInner": PodAffinityTypeOptionsOptionsInner,
    "PodAffinityTypeRules": PodAffinityTypeRules,
    "PolicyInfo": PolicyInfo,
    "PolicyInfoAllOf": PolicyInfoAllOf,
    "PolicyInfoAllOf1": PolicyInfoAllOf1,
    "Port": Port,
    "PreemptibilityOptions": PreemptibilityOptions,
    "PreemptibilityOptionsOptionsInner": PreemptibilityOptionsOptionsInner,
    "PreemptibilityRules": PreemptibilityRules,
    "Probe": Probe,
    "ProbeHandler": ProbeHandler,
    "ProbeHandlerHttpGet": ProbeHandlerHttpGet,
    "Probes": Probes,
    "Pvc": Pvc,
    "PvcAccessModes": PvcAccessModes,
    "PvcAccessModesRules": PvcAccessModesRules,
    "PvcAddedAttrValue": PvcAddedAttrValue,
    "PvcFieldsNonUpdatable": PvcFieldsNonUpdatable,
    "PvcFieldsUpdatable": PvcFieldsUpdatable,
    "PvcInstance": PvcInstance,
    "PvcRules": PvcRules,
    "PvcsRules": PvcsRules,
    "Registry": Registry,
    "RegistryCreationRequest": RegistryCreationRequest,
    "RegistryListResponse": RegistryListResponse,
    "RegistrySpec": RegistrySpec,
    "RegistryUpdateRequest": RegistryUpdateRequest,
    "RelatedUrl": RelatedUrl,
    "Repositories": Repositories,
    "RepositoryTags": RepositoryTags,
    "RequestInfo": RequestInfo,
    "ResponseInfo": ResponseInfo,
    "RestartPolicyOptions": RestartPolicyOptions,
    "RestartPolicyOptionsOptionsInner": RestartPolicyOptionsOptionsInner,
    "RestartPolicyRule": RestartPolicyRule,
    "Rules": Rules,
    "S3AccessKey": S3AccessKey,
    "S3AccessKeyAsset": S3AccessKeyAsset,
    "S3Asset": S3Asset,
    "S3AssetSpec": S3AssetSpec,
    "S3Common": S3Common,
    "S3CreationRequest": S3CreationRequest,
    "S3Instance": S3Instance,
    "S3ListResponse": S3ListResponse,
    "S3Rules": S3Rules,
    "S3UpdateRequest": S3UpdateRequest,
    "S3sRules": S3sRules,
    "ScaleDownDelayField": ScaleDownDelayField,
    "ScaleToZeroRetentionField": ScaleToZeroRetentionField,
    "Secret": Secret,
    "SecretAsset": SecretAsset,
    "SecretAssetCreationRequest": SecretAssetCreationRequest,
    "SecretAssetListResponse": SecretAssetListResponse,
    "SecretAssetSpec": SecretAssetSpec,
    "SecretAssetSpecAllOf": SecretAssetSpecAllOf,
    "SecretAssetUpdateRequest": SecretAssetUpdateRequest,
    "SecretFieldsNonUpdatable": SecretFieldsNonUpdatable,
    "SecretFieldsUpdatable": SecretFieldsUpdatable,
    "SecretInstance": SecretInstance,
    "SecretRules": SecretRules,
    "SecretsRules": SecretsRules,
    "ServingConfiguration": ServingConfiguration,
    "ServingConfigurationField": ServingConfigurationField,
    "ServingConfigurationRules": ServingConfigurationRules,
    "ServingPortAccess": ServingPortAccess,
    "ServingPortInfo": ServingPortInfo,
    "SourceOfRule": SourceOfRule,
    "SpecificRunAutoScaling": SpecificRunAutoScaling,
    "SpecificRunAutoScalingAutoScaling": SpecificRunAutoScalingAutoScaling,
    "SpecificRunConnectionInfo": SpecificRunConnectionInfo,
    "SpecificRunConnections": SpecificRunConnections,
    "SpecificRunCreationFields": SpecificRunCreationFields,
    "SpecificRunInfoFields": SpecificRunInfoFields,
    "SpecificRunMetricFields": SpecificRunMetricFields,
    "SpecificRunParamAttrDefaults": SpecificRunParamAttrDefaults,
    "SpecificRunParams": SpecificRunParams,
    "SpecificRunParamsAllOf": SpecificRunParamsAllOf,
    "SpecificRunParamsRules": SpecificRunParamsRules,
    "SpecificRunParamsRulesAllOf": SpecificRunParamsRulesAllOf,
    "SpecificRunServingPortAccess": SpecificRunServingPortAccess,
    "SpecificRunServingPortAccessServingPortAccess": SpecificRunServingPortAccessServingPortAccess,
    "StorageInstanceName": StorageInstanceName,
    "StringOption": StringOption,
    "StringRules": StringRules,
    "StringRulesCommon": StringRulesCommon,
    "StringRulesOptions": StringRulesOptions,
    "SubmissionErrorV2": SubmissionErrorV2,
    "SupersetServingPort": SupersetServingPort,
    "SupersetSpec": SupersetSpec,
    "SupersetSpecAllOf": SupersetSpecAllOf,
    "SupersetSpecAllOfCompute": SupersetSpecAllOfCompute,
    "SupersetSpecAllOfSecurity": SupersetSpecAllOfSecurity,
    "SupersetSpecAllOfStorage": SupersetSpecAllOfStorage,
    "SystemInfo": SystemInfo,
    "Toleration": Toleration,
    "TolerationEffectOptions": TolerationEffectOptions,
    "TolerationEffectOptionsOptionsInner": TolerationEffectOptionsOptionsInner,
    "TolerationEffectRules": TolerationEffectRules,
    "TolerationOperatorOptions": TolerationOperatorOptions,
    "TolerationOperatorOptionsOptionsInner": TolerationOperatorOptionsOptionsInner,
    "TolerationOperatorRules": TolerationOperatorRules,
    "TolerationRules": TolerationRules,
    "TolerationsRules": TolerationsRules,
    "Training": Training,
    "TrainingCreationRequestV2": TrainingCreationRequestV2,
    "TrainingListResponse": TrainingListResponse,
    "TrainingSpecificRunParams": TrainingSpecificRunParams,
    "TrainingV2": TrainingV2,
    "UsageTimesInfo": UsageTimesInfo,
    "ValidateSyncRequest": ValidateSyncRequest,
    "ValidateSyncResponse": ValidateSyncResponse,
    "WorkloadCreationMeta": WorkloadCreationMeta,
    "WorkloadCreationRequest": WorkloadCreationRequest,
    "WorkloadMeta": WorkloadMeta,
    "WorkloadMetaV1": WorkloadMetaV1,
    "WorkloadRefAndStatus": WorkloadRefAndStatus,
    "WorkloadStatusUpdate": WorkloadStatusUpdate,
    "WorkloadSupportedTypes": WorkloadSupportedTypes,
    "WorkloadTemplate": WorkloadTemplate,
    "WorkloadTemplateCreationRequest": WorkloadTemplateCreationRequest,
    "WorkloadTemplateListResponse": WorkloadTemplateListResponse,
    "WorkloadTemplateUpdateRequest": WorkloadTemplateUpdateRequest,
    "Workspace": Workspace,
    "WorkspaceCreationRequestV2": WorkspaceCreationRequestV2,
    "WorkspaceListResponse": WorkspaceListResponse,
    "WorkspaceSpecificRunParams": WorkspaceSpecificRunParams,
    "WorkspaceSpecificRunParamsAllOf": WorkspaceSpecificRunParamsAllOf,
    "WorkspaceV2": WorkspaceV2,
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
