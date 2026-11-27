import localVarRequest from 'request';

export * from './agentSyncDataDepricated';
export * from './cluster';
export * from './clusterControlPlaneConfig';
export * from './clusterControlPlaneConfigAgent';
export * from './clusterControlPlaneConfigAgentClient';
export * from './clusterControlPlaneConfigMetricsRemoteWrite';
export * from './clusterControlPlaneConfigMetricsRemoteWriteCredentials';
export * from './clusterControlPlaneConfigResearcherService';
export * from './clusterControlPlaneConfigResearcherServiceOauthProxy';
export * from './clusterCustomConfig';
export * from './clusterInfo';
export * from './clusterInfoDependencies';
export * from './container';
export * from './createCluster';
export * from './dependencyStatus';
export * from './deployment';
export * from './deploymentCondition';
export * from './deploymentsSync';
export * from './displayedJob';
export * from './event';
export * from './eventsSyncData';
export * from './groups';
export * from './idsResponse';
export * from './modelError';
export * from './nodePool';
export * from './nodePoolAllOf';
export * from './nodePoolCreateRequest';
export * from './nodePoolLabelsRequest';
export * from './nodePoolResponse';
export * from './nodePoolStatus';
export * from './nodePoolStatusRequest';
export * from './owner';
export * from './pVC';
export * from './placementStrategy';
export * from './pod';
export * from './podGroup';
export * from './podGroupsSync';
export * from './podTemplate';
export * from './podsSync';
export * from './secret';
export * from './service';
export * from './servicesSync';
export * from './statusChange';
export * from './storageClass';
export * from './updateClusterLivenessRequest';
export * from './updateNodePoolRequest';
export * from './user';
export * from './users';
export * from './v1K8sClustersUuidDelete200Response';
export * from './workloadEvent';
export * from './workloadHistoryMeta';
export * from './workloadHistoryResponse';
export * from './workloadHistoryResponseEntriesInner';
export * from './workloadHistorySpec';
export * from './workloadStatus';

import * as fs from 'fs';

export interface RequestDetailedFile {
    value: Buffer;
    options?: {
        filename?: string;
        contentType?: string;
    }
}

export type RequestFile = string | Buffer | fs.ReadStream | RequestDetailedFile;


import { AgentSyncDataDepricated } from './agentSyncDataDepricated';
import { Cluster } from './cluster';
import { ClusterControlPlaneConfig } from './clusterControlPlaneConfig';
import { ClusterControlPlaneConfigAgent } from './clusterControlPlaneConfigAgent';
import { ClusterControlPlaneConfigAgentClient } from './clusterControlPlaneConfigAgentClient';
import { ClusterControlPlaneConfigMetricsRemoteWrite } from './clusterControlPlaneConfigMetricsRemoteWrite';
import { ClusterControlPlaneConfigMetricsRemoteWriteCredentials } from './clusterControlPlaneConfigMetricsRemoteWriteCredentials';
import { ClusterControlPlaneConfigResearcherService } from './clusterControlPlaneConfigResearcherService';
import { ClusterControlPlaneConfigResearcherServiceOauthProxy } from './clusterControlPlaneConfigResearcherServiceOauthProxy';
import { ClusterCustomConfig } from './clusterCustomConfig';
import { ClusterInfo } from './clusterInfo';
import { ClusterInfoDependencies } from './clusterInfoDependencies';
import { Container } from './container';
import { CreateCluster } from './createCluster';
import { DependencyStatus } from './dependencyStatus';
import { Deployment } from './deployment';
import { DeploymentCondition } from './deploymentCondition';
import { DeploymentsSync } from './deploymentsSync';
import { DisplayedJob } from './displayedJob';
import { Event } from './event';
import { EventsSyncData } from './eventsSyncData';
import { Groups } from './groups';
import { IdsResponse } from './idsResponse';
import { ModelError } from './modelError';
import { NodePool } from './nodePool';
import { NodePoolAllOf } from './nodePoolAllOf';
import { NodePoolCreateRequest } from './nodePoolCreateRequest';
import { NodePoolLabelsRequest } from './nodePoolLabelsRequest';
import { NodePoolResponse } from './nodePoolResponse';
import { NodePoolStatus } from './nodePoolStatus';
import { NodePoolStatusRequest } from './nodePoolStatusRequest';
import { Owner } from './owner';
import { PVC } from './pVC';
import { PlacementStrategy } from './placementStrategy';
import { Pod } from './pod';
import { PodGroup } from './podGroup';
import { PodGroupsSync } from './podGroupsSync';
import { PodTemplate } from './podTemplate';
import { PodsSync } from './podsSync';
import { Secret } from './secret';
import { Service } from './service';
import { ServicesSync } from './servicesSync';
import { StatusChange } from './statusChange';
import { StorageClass } from './storageClass';
import { UpdateClusterLivenessRequest } from './updateClusterLivenessRequest';
import { UpdateNodePoolRequest } from './updateNodePoolRequest';
import { User } from './user';
import { Users } from './users';
import { V1K8sClustersUuidDelete200Response } from './v1K8sClustersUuidDelete200Response';
import { WorkloadEvent } from './workloadEvent';
import { WorkloadHistoryMeta } from './workloadHistoryMeta';
import { WorkloadHistoryResponse } from './workloadHistoryResponse';
import { WorkloadHistoryResponseEntriesInner } from './workloadHistoryResponseEntriesInner';
import { WorkloadHistorySpec } from './workloadHistorySpec';
import { WorkloadStatus } from './workloadStatus';

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
        "DeploymentCondition.TypeEnum": DeploymentCondition.TypeEnum,
        "DeploymentCondition.StatusEnum": DeploymentCondition.StatusEnum,
        "NodePool.StatusEnum": NodePool.StatusEnum,
        "NodePoolAllOf.StatusEnum": NodePoolAllOf.StatusEnum,
        "NodePoolResponse.StatusEnum": NodePoolResponse.StatusEnum,
        "NodePoolStatusRequest.StatusEnum": NodePoolStatusRequest.StatusEnum,
        "PlacementStrategy.CpuEnum": PlacementStrategy.CpuEnum,
        "PlacementStrategy.GpuEnum": PlacementStrategy.GpuEnum,
        "PodGroup.WorkloadTypeEnum": PodGroup.WorkloadTypeEnum,
        "PodGroup.PodGroupPhaseEnum": PodGroup.PodGroupPhaseEnum,
        "WorkloadStatus.StatusEnum": WorkloadStatus.StatusEnum,
}

let typeMap: {[index: string]: any} = {
    "AgentSyncDataDepricated": AgentSyncDataDepricated,
    "Cluster": Cluster,
    "ClusterControlPlaneConfig": ClusterControlPlaneConfig,
    "ClusterControlPlaneConfigAgent": ClusterControlPlaneConfigAgent,
    "ClusterControlPlaneConfigAgentClient": ClusterControlPlaneConfigAgentClient,
    "ClusterControlPlaneConfigMetricsRemoteWrite": ClusterControlPlaneConfigMetricsRemoteWrite,
    "ClusterControlPlaneConfigMetricsRemoteWriteCredentials": ClusterControlPlaneConfigMetricsRemoteWriteCredentials,
    "ClusterControlPlaneConfigResearcherService": ClusterControlPlaneConfigResearcherService,
    "ClusterControlPlaneConfigResearcherServiceOauthProxy": ClusterControlPlaneConfigResearcherServiceOauthProxy,
    "ClusterCustomConfig": ClusterCustomConfig,
    "ClusterInfo": ClusterInfo,
    "ClusterInfoDependencies": ClusterInfoDependencies,
    "Container": Container,
    "CreateCluster": CreateCluster,
    "DependencyStatus": DependencyStatus,
    "Deployment": Deployment,
    "DeploymentCondition": DeploymentCondition,
    "DeploymentsSync": DeploymentsSync,
    "DisplayedJob": DisplayedJob,
    "Event": Event,
    "EventsSyncData": EventsSyncData,
    "Groups": Groups,
    "IdsResponse": IdsResponse,
    "ModelError": ModelError,
    "NodePool": NodePool,
    "NodePoolAllOf": NodePoolAllOf,
    "NodePoolCreateRequest": NodePoolCreateRequest,
    "NodePoolLabelsRequest": NodePoolLabelsRequest,
    "NodePoolResponse": NodePoolResponse,
    "NodePoolStatus": NodePoolStatus,
    "NodePoolStatusRequest": NodePoolStatusRequest,
    "Owner": Owner,
    "PVC": PVC,
    "PlacementStrategy": PlacementStrategy,
    "Pod": Pod,
    "PodGroup": PodGroup,
    "PodGroupsSync": PodGroupsSync,
    "PodTemplate": PodTemplate,
    "PodsSync": PodsSync,
    "Secret": Secret,
    "Service": Service,
    "ServicesSync": ServicesSync,
    "StatusChange": StatusChange,
    "StorageClass": StorageClass,
    "UpdateClusterLivenessRequest": UpdateClusterLivenessRequest,
    "UpdateNodePoolRequest": UpdateNodePoolRequest,
    "User": User,
    "Users": Users,
    "V1K8sClustersUuidDelete200Response": V1K8sClustersUuidDelete200Response,
    "WorkloadEvent": WorkloadEvent,
    "WorkloadHistoryMeta": WorkloadHistoryMeta,
    "WorkloadHistoryResponse": WorkloadHistoryResponse,
    "WorkloadHistoryResponseEntriesInner": WorkloadHistoryResponseEntriesInner,
    "WorkloadHistorySpec": WorkloadHistorySpec,
    "WorkloadStatus": WorkloadStatus,
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
