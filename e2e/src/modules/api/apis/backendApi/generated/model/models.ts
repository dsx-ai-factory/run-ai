import localVarRequest from 'request';

export * from './app';
export * from './appAllOf';
export * from './authEntity';
export * from './authEntityType';
export * from './cluster';
export * from './clusterCurrent';
export * from './clusterCurrentProjectResourcesInner';
export * from './clusterCurrentProjectResourcesInnerAllOf';
export * from './clusterCurrentResourcesInner';
export * from './clusterCurrentResourcesInnerAllOf';
export * from './clusterMetadata';
export * from './createSweep200Response';
export * from './createSweepRequest';
export * from './department';
export * from './departmentCurrent';
export * from './departmentCurrentProjectResourcesInner';
export * from './departmentCurrentProjectResourcesInnerAllOf';
export * from './departmentMetadata';
export * from './departments';
export * from './departmentsDataInner';
export * from './departmentsDataInnerAllOf';
export * from './getAncestors200Response';
export * from './getOrgTree200Response';
export * from './getScopeName200Response';
export * from './group';
export * from './groupCreationRequest';
export * from './groupCreationResponse';
export * from './job';
export * from './jobAdvancedData';
export * from './jobCurrent';
export * from './jobMetadata';
export * from './jobResourceData';
export * from './jobResourcesData';
export * from './jobResourcesDataAdvanced';
export * from './jobResourcesDataOvertime';
export * from './jobTimeRange';
export * from './logo200Response';
export * from './modelError';
export * from './orgUnit';
export * from './postOAUTHSSOConfigurations200Response';
export * from './postOAUTHSSOConfigurationsRequest';
export * from './postSAMLSSOConfigurations200Response';
export * from './postSAMLSSOConfigurationsRequest';
export * from './project';
export * from './projectCurrent';
export * from './projectMetadata';
export * from './projects';
export * from './projectsDataInner';
export * from './resourceData';
export * from './resourcesData';
export * from './resourcesDataOvertime';
export * from './resourcesDataOvertimeAllOf';
export * from './role';
export * from './scopeType';
export * from './secret';
export * from './secretSecret';
export * from './tenant';
export * from './tenantConfig';
export * from './tenantSetting';
export * from './tenantSettingCreationRequest';
export * from './tenantSettingCreationResponse';
export * from './timeRange';
export * from './updateGroupByNameRequest';
export * from './updateGroupByNameRequestAllOf';
export * from './user';
export * from './userAllOf';
export * from './userCreationRequest';
export * from './userCreationResponse';

import * as fs from 'fs';

export interface RequestDetailedFile {
    value: Buffer;
    options?: {
        filename?: string;
        contentType?: string;
    }
}

export type RequestFile = string | Buffer | fs.ReadStream | RequestDetailedFile;


import { App } from './app';
import { AppAllOf } from './appAllOf';
import { AuthEntity } from './authEntity';
import { AuthEntityType } from './authEntityType';
import { Cluster } from './cluster';
import { ClusterCurrent } from './clusterCurrent';
import { ClusterCurrentProjectResourcesInner } from './clusterCurrentProjectResourcesInner';
import { ClusterCurrentProjectResourcesInnerAllOf } from './clusterCurrentProjectResourcesInnerAllOf';
import { ClusterCurrentResourcesInner } from './clusterCurrentResourcesInner';
import { ClusterCurrentResourcesInnerAllOf } from './clusterCurrentResourcesInnerAllOf';
import { ClusterMetadata } from './clusterMetadata';
import { CreateSweep200Response } from './createSweep200Response';
import { CreateSweepRequest } from './createSweepRequest';
import { Department } from './department';
import { DepartmentCurrent } from './departmentCurrent';
import { DepartmentCurrentProjectResourcesInner } from './departmentCurrentProjectResourcesInner';
import { DepartmentCurrentProjectResourcesInnerAllOf } from './departmentCurrentProjectResourcesInnerAllOf';
import { DepartmentMetadata } from './departmentMetadata';
import { Departments } from './departments';
import { DepartmentsDataInner } from './departmentsDataInner';
import { DepartmentsDataInnerAllOf } from './departmentsDataInnerAllOf';
import { GetAncestors200Response } from './getAncestors200Response';
import { GetOrgTree200Response } from './getOrgTree200Response';
import { GetScopeName200Response } from './getScopeName200Response';
import { Group } from './group';
import { GroupCreationRequest } from './groupCreationRequest';
import { GroupCreationResponse } from './groupCreationResponse';
import { Job } from './job';
import { JobAdvancedData } from './jobAdvancedData';
import { JobCurrent } from './jobCurrent';
import { JobMetadata } from './jobMetadata';
import { JobResourceData } from './jobResourceData';
import { JobResourcesData } from './jobResourcesData';
import { JobResourcesDataAdvanced } from './jobResourcesDataAdvanced';
import { JobResourcesDataOvertime } from './jobResourcesDataOvertime';
import { JobTimeRange } from './jobTimeRange';
import { Logo200Response } from './logo200Response';
import { ModelError } from './modelError';
import { OrgUnit } from './orgUnit';
import { PostOAUTHSSOConfigurations200Response } from './postOAUTHSSOConfigurations200Response';
import { PostOAUTHSSOConfigurationsRequest } from './postOAUTHSSOConfigurationsRequest';
import { PostSAMLSSOConfigurations200Response } from './postSAMLSSOConfigurations200Response';
import { PostSAMLSSOConfigurationsRequest } from './postSAMLSSOConfigurationsRequest';
import { Project } from './project';
import { ProjectCurrent } from './projectCurrent';
import { ProjectMetadata } from './projectMetadata';
import { Projects } from './projects';
import { ProjectsDataInner } from './projectsDataInner';
import { ResourceData } from './resourceData';
import { ResourcesData } from './resourcesData';
import { ResourcesDataOvertime } from './resourcesDataOvertime';
import { ResourcesDataOvertimeAllOf } from './resourcesDataOvertimeAllOf';
import { Role } from './role';
import { ScopeType } from './scopeType';
import { Secret } from './secret';
import { SecretSecret } from './secretSecret';
import { Tenant } from './tenant';
import { TenantConfig } from './tenantConfig';
import { TenantSetting } from './tenantSetting';
import { TenantSettingCreationRequest } from './tenantSettingCreationRequest';
import { TenantSettingCreationResponse } from './tenantSettingCreationResponse';
import { TimeRange } from './timeRange';
import { UpdateGroupByNameRequest } from './updateGroupByNameRequest';
import { UpdateGroupByNameRequestAllOf } from './updateGroupByNameRequestAllOf';
import { User } from './user';
import { UserAllOf } from './userAllOf';
import { UserCreationRequest } from './userCreationRequest';
import { UserCreationResponse } from './userCreationResponse';

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
        "AuthEntityType": AuthEntityType,
        "Role": Role,
        "ScopeType": ScopeType,
        "SecretSecret.TypeEnum": SecretSecret.TypeEnum,
}

let typeMap: {[index: string]: any} = {
    "App": App,
    "AppAllOf": AppAllOf,
    "AuthEntity": AuthEntity,
    "Cluster": Cluster,
    "ClusterCurrent": ClusterCurrent,
    "ClusterCurrentProjectResourcesInner": ClusterCurrentProjectResourcesInner,
    "ClusterCurrentProjectResourcesInnerAllOf": ClusterCurrentProjectResourcesInnerAllOf,
    "ClusterCurrentResourcesInner": ClusterCurrentResourcesInner,
    "ClusterCurrentResourcesInnerAllOf": ClusterCurrentResourcesInnerAllOf,
    "ClusterMetadata": ClusterMetadata,
    "CreateSweep200Response": CreateSweep200Response,
    "CreateSweepRequest": CreateSweepRequest,
    "Department": Department,
    "DepartmentCurrent": DepartmentCurrent,
    "DepartmentCurrentProjectResourcesInner": DepartmentCurrentProjectResourcesInner,
    "DepartmentCurrentProjectResourcesInnerAllOf": DepartmentCurrentProjectResourcesInnerAllOf,
    "DepartmentMetadata": DepartmentMetadata,
    "Departments": Departments,
    "DepartmentsDataInner": DepartmentsDataInner,
    "DepartmentsDataInnerAllOf": DepartmentsDataInnerAllOf,
    "GetAncestors200Response": GetAncestors200Response,
    "GetOrgTree200Response": GetOrgTree200Response,
    "GetScopeName200Response": GetScopeName200Response,
    "Group": Group,
    "GroupCreationRequest": GroupCreationRequest,
    "GroupCreationResponse": GroupCreationResponse,
    "Job": Job,
    "JobAdvancedData": JobAdvancedData,
    "JobCurrent": JobCurrent,
    "JobMetadata": JobMetadata,
    "JobResourceData": JobResourceData,
    "JobResourcesData": JobResourcesData,
    "JobResourcesDataAdvanced": JobResourcesDataAdvanced,
    "JobResourcesDataOvertime": JobResourcesDataOvertime,
    "JobTimeRange": JobTimeRange,
    "Logo200Response": Logo200Response,
    "ModelError": ModelError,
    "OrgUnit": OrgUnit,
    "PostOAUTHSSOConfigurations200Response": PostOAUTHSSOConfigurations200Response,
    "PostOAUTHSSOConfigurationsRequest": PostOAUTHSSOConfigurationsRequest,
    "PostSAMLSSOConfigurations200Response": PostSAMLSSOConfigurations200Response,
    "PostSAMLSSOConfigurationsRequest": PostSAMLSSOConfigurationsRequest,
    "Project": Project,
    "ProjectCurrent": ProjectCurrent,
    "ProjectMetadata": ProjectMetadata,
    "Projects": Projects,
    "ProjectsDataInner": ProjectsDataInner,
    "ResourceData": ResourceData,
    "ResourcesData": ResourcesData,
    "ResourcesDataOvertime": ResourcesDataOvertime,
    "ResourcesDataOvertimeAllOf": ResourcesDataOvertimeAllOf,
    "Secret": Secret,
    "SecretSecret": SecretSecret,
    "Tenant": Tenant,
    "TenantConfig": TenantConfig,
    "TenantSetting": TenantSetting,
    "TenantSettingCreationRequest": TenantSettingCreationRequest,
    "TenantSettingCreationResponse": TenantSettingCreationResponse,
    "TimeRange": TimeRange,
    "UpdateGroupByNameRequest": UpdateGroupByNameRequest,
    "UpdateGroupByNameRequestAllOf": UpdateGroupByNameRequestAllOf,
    "User": User,
    "UserAllOf": UserAllOf,
    "UserCreationRequest": UserCreationRequest,
    "UserCreationResponse": UserCreationResponse,
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
