import localVarRequest from 'request';

export * from './agentNetworkingConfiguration';
export * from './agentNetworkingService';
export * from './agentServiceCreate';
export * from './annotation';
export * from './authorization';
export * from './authorizationType';
export * from './batchResponse';
export * from './batchResponseFailureItem';
export * from './claimInfo';
export * from './clusterSyncReport';
export * from './clusterSyncReportUpdate';
export * from './complianceIssues';
export * from './complianceIssuesComplianceIssuesInner';
export * from './complianceIssuesV2';
export * from './complianceIssuesV2ComplianceIssuesInner';
export * from './desiredPhase';
export * from './environmentVariable';
export * from './environmentVariableConfigMap';
export * from './environmentVariablePodFieldReference';
export * from './environmentVariableSecret';
export * from './environmentVariableUserCredential';
export * from './externalAccess';
export * from './gVK';
export * from './getYamlSubmission';
export * from './getYamlSubmissionValue';
export * from './gpuRequestType';
export * from './imagePullPolicy';
export * from './imagePullSecret';
export * from './label';
export * from './mnnvlMode';
export * from './modelError';
export * from './nIMServiceMetadataCreateParams';
export * from './networkingEndpoint';
export * from './networkingService';
export * from './nimCache';
export * from './nimServiceCreateRequest';
export * from './nimServiceMultiNode';
export * from './nimServicePvcFields';
export * from './nimServiceRemovableFields';
export * from './nimServiceResponse';
export * from './nimServiceServingPort';
export * from './nimServiceServingPortAuthorizationType';
export * from './nimServiceSpec';
export * from './nimServiceSpecAllOf';
export * from './nimServiceSpecAllOfAutoscaling';
export * from './nimServiceSpecAllOfCompute';
export * from './nimServiceSpecAllOfModelStore';
export * from './nimServiceSpecAllOfSecurity';
export * from './nimServiceUpdateRequest';
export * from './policyRuleEnum';
export * from './preemptibility';
export * from './probe';
export * from './probeHandler';
export * from './probeHandlerHttpGet';
export * from './probeHandlerScheme';
export * from './probes';
export * from './pvcAccessModes';
export * from './pvcAddedAttrValue';
export * from './pvcVolumeMode';
export * from './requestInfo';
export * from './responseInfo';
export * from './runState';
export * from './servicePort';
export * from './serviceProtocol';
export * from './serviceRef';
export * from './serviceSpec';
export * from './serviceType';
export * from './servingPortExposedProtocol';
export * from './servingPortServiceType';
export * from './submissionError';
export * from './submissionErrorV2';
export * from './systemInfo';
export * from './toleration';
export * from './tolerationEffect';
export * from './tolerationOperator';
export * from './workloadRunStateRequest';
export * from './workloadV2';
export * from './workloadV2Configuration';
export * from './workloadV2CreateRequest';
export * from './workloadV2GetResponse';
export * from './workloadV2Metadata';
export * from './workloadV2MetadataAutoFill';
export * from './workloadV2MetadataCreateParams';
export * from './workloadV2MetadataResponse';
export * from './workloadV2MetadataUpdateParams';
export * from './workloadV2SyncInfo';
export * from './workloadV2SyncStatus';
export * from './workloadV2Update';
export * from './workloadV2UpdateConfiguration';
export * from './workloadV2UpdateConfigurationMnnvl';
export * from './workloadV2UpdateConfigurationPvc';
export * from './workloadV2UpdateConfigurationPvcAccessMode';
export * from './workloadV2UpdateConfigurationPvcMetadata';
export * from './workloadV2UpdateConfigurationPvcResources';
export * from './workloadV2UpdateConfigurationPvcResourcesRequests';
export * from './workloadV2UpdateConfigurationPvcSpec';
export * from './workloadV2UpdateConfigurationPvcVolumeMode';
export * from './workloadV2UpdateConfigurationUserCredentialSource';
export * from './workloadV2UpdateConfigurationUserCredentials';
export * from './workloadV2UpdateRequest';
export * from './workloadsSettingsGetResponse';
export * from './workloadsSettingsKeyEnum';
export * from './workloadsV2Batch';
export * from './workloadsV2BatchAction';
export * from './workloadsV2Updates';
export * from './workloadv2Authorization';
export * from './workloadv2AuthorizationAllOf';
export * from './yamlSubmissionSetting';

import * as fs from 'fs';

export interface RequestDetailedFile {
    value: Buffer;
    options?: {
        filename?: string;
        contentType?: string;
    }
}

export type RequestFile = string | Buffer | fs.ReadStream | RequestDetailedFile;


import { AgentNetworkingConfiguration } from './agentNetworkingConfiguration';
import { AgentNetworkingService } from './agentNetworkingService';
import { AgentServiceCreate } from './agentServiceCreate';
import { Annotation } from './annotation';
import { Authorization } from './authorization';
import { AuthorizationType } from './authorizationType';
import { BatchResponse } from './batchResponse';
import { BatchResponseFailureItem } from './batchResponseFailureItem';
import { ClaimInfo } from './claimInfo';
import { ClusterSyncReport } from './clusterSyncReport';
import { ClusterSyncReportUpdate } from './clusterSyncReportUpdate';
import { ComplianceIssues } from './complianceIssues';
import { ComplianceIssuesComplianceIssuesInner } from './complianceIssuesComplianceIssuesInner';
import { ComplianceIssuesV2 } from './complianceIssuesV2';
import { ComplianceIssuesV2ComplianceIssuesInner } from './complianceIssuesV2ComplianceIssuesInner';
import { DesiredPhase } from './desiredPhase';
import { EnvironmentVariable } from './environmentVariable';
import { EnvironmentVariableConfigMap } from './environmentVariableConfigMap';
import { EnvironmentVariablePodFieldReference } from './environmentVariablePodFieldReference';
import { EnvironmentVariableSecret } from './environmentVariableSecret';
import { EnvironmentVariableUserCredential } from './environmentVariableUserCredential';
import { ExternalAccess } from './externalAccess';
import { GVK } from './gVK';
import { GetYamlSubmission } from './getYamlSubmission';
import { GetYamlSubmissionValue } from './getYamlSubmissionValue';
import { GpuRequestType } from './gpuRequestType';
import { ImagePullPolicy } from './imagePullPolicy';
import { ImagePullSecret } from './imagePullSecret';
import { Label } from './label';
import { MnnvlMode } from './mnnvlMode';
import { ModelError } from './modelError';
import { NIMServiceMetadataCreateParams } from './nIMServiceMetadataCreateParams';
import { NetworkingEndpoint } from './networkingEndpoint';
import { NetworkingService } from './networkingService';
import { NimCache } from './nimCache';
import { NimServiceCreateRequest } from './nimServiceCreateRequest';
import { NimServiceMultiNode } from './nimServiceMultiNode';
import { NimServicePvcFields } from './nimServicePvcFields';
import { NimServiceRemovableFields } from './nimServiceRemovableFields';
import { NimServiceResponse } from './nimServiceResponse';
import { NimServiceServingPort } from './nimServiceServingPort';
import { NimServiceServingPortAuthorizationType } from './nimServiceServingPortAuthorizationType';
import { NimServiceSpec } from './nimServiceSpec';
import { NimServiceSpecAllOf } from './nimServiceSpecAllOf';
import { NimServiceSpecAllOfAutoscaling } from './nimServiceSpecAllOfAutoscaling';
import { NimServiceSpecAllOfCompute } from './nimServiceSpecAllOfCompute';
import { NimServiceSpecAllOfModelStore } from './nimServiceSpecAllOfModelStore';
import { NimServiceSpecAllOfSecurity } from './nimServiceSpecAllOfSecurity';
import { NimServiceUpdateRequest } from './nimServiceUpdateRequest';
import { PolicyRuleEnum } from './policyRuleEnum';
import { Preemptibility } from './preemptibility';
import { Probe } from './probe';
import { ProbeHandler } from './probeHandler';
import { ProbeHandlerHttpGet } from './probeHandlerHttpGet';
import { ProbeHandlerScheme } from './probeHandlerScheme';
import { Probes } from './probes';
import { PvcAccessModes } from './pvcAccessModes';
import { PvcAddedAttrValue } from './pvcAddedAttrValue';
import { PvcVolumeMode } from './pvcVolumeMode';
import { RequestInfo } from './requestInfo';
import { ResponseInfo } from './responseInfo';
import { RunState } from './runState';
import { ServicePort } from './servicePort';
import { ServiceProtocol } from './serviceProtocol';
import { ServiceRef } from './serviceRef';
import { ServiceSpec } from './serviceSpec';
import { ServiceType } from './serviceType';
import { ServingPortExposedProtocol } from './servingPortExposedProtocol';
import { ServingPortServiceType } from './servingPortServiceType';
import { SubmissionError } from './submissionError';
import { SubmissionErrorV2 } from './submissionErrorV2';
import { SystemInfo } from './systemInfo';
import { Toleration } from './toleration';
import { TolerationEffect } from './tolerationEffect';
import { TolerationOperator } from './tolerationOperator';
import { WorkloadRunStateRequest } from './workloadRunStateRequest';
import { WorkloadV2 } from './workloadV2';
import { WorkloadV2Configuration } from './workloadV2Configuration';
import { WorkloadV2CreateRequest } from './workloadV2CreateRequest';
import { WorkloadV2GetResponse } from './workloadV2GetResponse';
import { WorkloadV2Metadata } from './workloadV2Metadata';
import { WorkloadV2MetadataAutoFill } from './workloadV2MetadataAutoFill';
import { WorkloadV2MetadataCreateParams } from './workloadV2MetadataCreateParams';
import { WorkloadV2MetadataResponse } from './workloadV2MetadataResponse';
import { WorkloadV2MetadataUpdateParams } from './workloadV2MetadataUpdateParams';
import { WorkloadV2SyncInfo } from './workloadV2SyncInfo';
import { WorkloadV2SyncStatus } from './workloadV2SyncStatus';
import { WorkloadV2Update } from './workloadV2Update';
import { WorkloadV2UpdateConfiguration } from './workloadV2UpdateConfiguration';
import { WorkloadV2UpdateConfigurationMnnvl } from './workloadV2UpdateConfigurationMnnvl';
import { WorkloadV2UpdateConfigurationPvc } from './workloadV2UpdateConfigurationPvc';
import { WorkloadV2UpdateConfigurationPvcAccessMode } from './workloadV2UpdateConfigurationPvcAccessMode';
import { WorkloadV2UpdateConfigurationPvcMetadata } from './workloadV2UpdateConfigurationPvcMetadata';
import { WorkloadV2UpdateConfigurationPvcResources } from './workloadV2UpdateConfigurationPvcResources';
import { WorkloadV2UpdateConfigurationPvcResourcesRequests } from './workloadV2UpdateConfigurationPvcResourcesRequests';
import { WorkloadV2UpdateConfigurationPvcSpec } from './workloadV2UpdateConfigurationPvcSpec';
import { WorkloadV2UpdateConfigurationPvcVolumeMode } from './workloadV2UpdateConfigurationPvcVolumeMode';
import { WorkloadV2UpdateConfigurationUserCredentialSource } from './workloadV2UpdateConfigurationUserCredentialSource';
import { WorkloadV2UpdateConfigurationUserCredentials } from './workloadV2UpdateConfigurationUserCredentials';
import { WorkloadV2UpdateRequest } from './workloadV2UpdateRequest';
import { WorkloadsSettingsGetResponse } from './workloadsSettingsGetResponse';
import { WorkloadsSettingsKeyEnum } from './workloadsSettingsKeyEnum';
import { WorkloadsV2Batch } from './workloadsV2Batch';
import { WorkloadsV2BatchAction } from './workloadsV2BatchAction';
import { WorkloadsV2Updates } from './workloadsV2Updates';
import { Workloadv2Authorization } from './workloadv2Authorization';
import { Workloadv2AuthorizationAllOf } from './workloadv2AuthorizationAllOf';
import { YamlSubmissionSetting } from './yamlSubmissionSetting';

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
        "AuthorizationType": AuthorizationType,
        "DesiredPhase": DesiredPhase,
        "GpuRequestType": GpuRequestType,
        "ImagePullPolicy": ImagePullPolicy,
        "MnnvlMode": MnnvlMode,
        "NimServiceRemovableFields": NimServiceRemovableFields,
        "NimServiceServingPortAuthorizationType": NimServiceServingPortAuthorizationType,
        "PolicyRuleEnum": PolicyRuleEnum,
        "Preemptibility": Preemptibility,
        "ProbeHandlerScheme": ProbeHandlerScheme,
        "PvcVolumeMode": PvcVolumeMode,
        "RunState": RunState,
        "ServiceProtocol": ServiceProtocol,
        "ServiceType": ServiceType,
        "ServingPortExposedProtocol": ServingPortExposedProtocol,
        "ServingPortServiceType": ServingPortServiceType,
        "TolerationEffect": TolerationEffect,
        "TolerationOperator": TolerationOperator,
        "WorkloadV2SyncStatus": WorkloadV2SyncStatus,
        "WorkloadV2UpdateConfigurationPvcAccessMode": WorkloadV2UpdateConfigurationPvcAccessMode,
        "WorkloadV2UpdateConfigurationPvcVolumeMode": WorkloadV2UpdateConfigurationPvcVolumeMode,
        "WorkloadsSettingsKeyEnum": WorkloadsSettingsKeyEnum,
        "WorkloadsV2BatchAction": WorkloadsV2BatchAction,
}

let typeMap: {[index: string]: any} = {
    "AgentNetworkingConfiguration": AgentNetworkingConfiguration,
    "AgentNetworkingService": AgentNetworkingService,
    "AgentServiceCreate": AgentServiceCreate,
    "Annotation": Annotation,
    "Authorization": Authorization,
    "BatchResponse": BatchResponse,
    "BatchResponseFailureItem": BatchResponseFailureItem,
    "ClaimInfo": ClaimInfo,
    "ClusterSyncReport": ClusterSyncReport,
    "ClusterSyncReportUpdate": ClusterSyncReportUpdate,
    "ComplianceIssues": ComplianceIssues,
    "ComplianceIssuesComplianceIssuesInner": ComplianceIssuesComplianceIssuesInner,
    "ComplianceIssuesV2": ComplianceIssuesV2,
    "ComplianceIssuesV2ComplianceIssuesInner": ComplianceIssuesV2ComplianceIssuesInner,
    "EnvironmentVariable": EnvironmentVariable,
    "EnvironmentVariableConfigMap": EnvironmentVariableConfigMap,
    "EnvironmentVariablePodFieldReference": EnvironmentVariablePodFieldReference,
    "EnvironmentVariableSecret": EnvironmentVariableSecret,
    "EnvironmentVariableUserCredential": EnvironmentVariableUserCredential,
    "ExternalAccess": ExternalAccess,
    "GVK": GVK,
    "GetYamlSubmission": GetYamlSubmission,
    "GetYamlSubmissionValue": GetYamlSubmissionValue,
    "ImagePullSecret": ImagePullSecret,
    "Label": Label,
    "ModelError": ModelError,
    "NIMServiceMetadataCreateParams": NIMServiceMetadataCreateParams,
    "NetworkingEndpoint": NetworkingEndpoint,
    "NetworkingService": NetworkingService,
    "NimCache": NimCache,
    "NimServiceCreateRequest": NimServiceCreateRequest,
    "NimServiceMultiNode": NimServiceMultiNode,
    "NimServicePvcFields": NimServicePvcFields,
    "NimServiceResponse": NimServiceResponse,
    "NimServiceServingPort": NimServiceServingPort,
    "NimServiceSpec": NimServiceSpec,
    "NimServiceSpecAllOf": NimServiceSpecAllOf,
    "NimServiceSpecAllOfAutoscaling": NimServiceSpecAllOfAutoscaling,
    "NimServiceSpecAllOfCompute": NimServiceSpecAllOfCompute,
    "NimServiceSpecAllOfModelStore": NimServiceSpecAllOfModelStore,
    "NimServiceSpecAllOfSecurity": NimServiceSpecAllOfSecurity,
    "NimServiceUpdateRequest": NimServiceUpdateRequest,
    "Probe": Probe,
    "ProbeHandler": ProbeHandler,
    "ProbeHandlerHttpGet": ProbeHandlerHttpGet,
    "Probes": Probes,
    "PvcAccessModes": PvcAccessModes,
    "PvcAddedAttrValue": PvcAddedAttrValue,
    "RequestInfo": RequestInfo,
    "ResponseInfo": ResponseInfo,
    "ServicePort": ServicePort,
    "ServiceRef": ServiceRef,
    "ServiceSpec": ServiceSpec,
    "SubmissionError": SubmissionError,
    "SubmissionErrorV2": SubmissionErrorV2,
    "SystemInfo": SystemInfo,
    "Toleration": Toleration,
    "WorkloadRunStateRequest": WorkloadRunStateRequest,
    "WorkloadV2": WorkloadV2,
    "WorkloadV2Configuration": WorkloadV2Configuration,
    "WorkloadV2CreateRequest": WorkloadV2CreateRequest,
    "WorkloadV2GetResponse": WorkloadV2GetResponse,
    "WorkloadV2Metadata": WorkloadV2Metadata,
    "WorkloadV2MetadataAutoFill": WorkloadV2MetadataAutoFill,
    "WorkloadV2MetadataCreateParams": WorkloadV2MetadataCreateParams,
    "WorkloadV2MetadataResponse": WorkloadV2MetadataResponse,
    "WorkloadV2MetadataUpdateParams": WorkloadV2MetadataUpdateParams,
    "WorkloadV2SyncInfo": WorkloadV2SyncInfo,
    "WorkloadV2Update": WorkloadV2Update,
    "WorkloadV2UpdateConfiguration": WorkloadV2UpdateConfiguration,
    "WorkloadV2UpdateConfigurationMnnvl": WorkloadV2UpdateConfigurationMnnvl,
    "WorkloadV2UpdateConfigurationPvc": WorkloadV2UpdateConfigurationPvc,
    "WorkloadV2UpdateConfigurationPvcMetadata": WorkloadV2UpdateConfigurationPvcMetadata,
    "WorkloadV2UpdateConfigurationPvcResources": WorkloadV2UpdateConfigurationPvcResources,
    "WorkloadV2UpdateConfigurationPvcResourcesRequests": WorkloadV2UpdateConfigurationPvcResourcesRequests,
    "WorkloadV2UpdateConfigurationPvcSpec": WorkloadV2UpdateConfigurationPvcSpec,
    "WorkloadV2UpdateConfigurationUserCredentialSource": WorkloadV2UpdateConfigurationUserCredentialSource,
    "WorkloadV2UpdateConfigurationUserCredentials": WorkloadV2UpdateConfigurationUserCredentials,
    "WorkloadV2UpdateRequest": WorkloadV2UpdateRequest,
    "WorkloadsSettingsGetResponse": WorkloadsSettingsGetResponse,
    "WorkloadsV2Batch": WorkloadsV2Batch,
    "WorkloadsV2Updates": WorkloadsV2Updates,
    "Workloadv2Authorization": Workloadv2Authorization,
    "Workloadv2AuthorizationAllOf": Workloadv2AuthorizationAllOf,
    "YamlSubmissionSetting": YamlSubmissionSetting,
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
