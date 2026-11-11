import localVarRequest from 'request';

export * from './assignedResources';
export * from './assignedResourcesCpu';
export * from './assignedResourcesGpu';
export * from './assignedResourcesMemory';
export * from './department';
export * from './departmentAccessControl';
export * from './departmentAllOf';
export * from './departmentCommonFields';
export * from './departmentCommonFieldsResources';
export * from './departmentCreateRequest';
export * from './departmentCreateRequestAllOf';
export * from './departmentUpdateRequest';
export * from './departmentUpdateRequestAllOf';
export * from './jobsNodeAffinity';
export * from './jobsNodeAffinityInteractive';
export * from './jobsNodeAffinityTrain';
export * from './modelError';
export * from './nodeAffinity';
export * from './nodePoolAssignedResources';
export * from './nodePoolAssignedResourcesAllOf';
export * from './nodePoolAssignedResourcesCreate';
export * from './nodePoolAssignedResourcesCreateCpu';
export * from './nodePoolAssignedResourcesCreateGpu';
export * from './nodePoolAssignedResourcesCreateMemory';
export * from './nodePoolAssignedResourcesCreateNodePool';
export * from './nodePoolQuotaStatus';
export * from './nodePoolResources';
export * from './nodeTypeForProject';
export * from './project';
export * from './projectAllOf';
export * from './projectCommonFields';
export * from './projectCommonFieldsNodeAffinity';
export * from './projectCommonFieldsPermissions';
export * from './projectCommonFieldsResources';
export * from './projectCreateRequest';
export * from './projectCreateRequestAllOf';
export * from './projectStatus';
export * from './projectStatusRequest';
export * from './projectUpdateRequest';
export * from './projectUpdateRequestAllOf';
export * from './quotaStatusResourceList';
export * from './resource';
export * from './resourcePermissions';

import * as fs from 'fs';

export interface RequestDetailedFile {
    value: Buffer;
    options?: {
        filename?: string;
        contentType?: string;
    }
}

export type RequestFile = string | Buffer | fs.ReadStream | RequestDetailedFile;


import { AssignedResources } from './assignedResources';
import { AssignedResourcesCpu } from './assignedResourcesCpu';
import { AssignedResourcesGpu } from './assignedResourcesGpu';
import { AssignedResourcesMemory } from './assignedResourcesMemory';
import { Department } from './department';
import { DepartmentAccessControl } from './departmentAccessControl';
import { DepartmentAllOf } from './departmentAllOf';
import { DepartmentCommonFields } from './departmentCommonFields';
import { DepartmentCommonFieldsResources } from './departmentCommonFieldsResources';
import { DepartmentCreateRequest } from './departmentCreateRequest';
import { DepartmentCreateRequestAllOf } from './departmentCreateRequestAllOf';
import { DepartmentUpdateRequest } from './departmentUpdateRequest';
import { DepartmentUpdateRequestAllOf } from './departmentUpdateRequestAllOf';
import { JobsNodeAffinity } from './jobsNodeAffinity';
import { JobsNodeAffinityInteractive } from './jobsNodeAffinityInteractive';
import { JobsNodeAffinityTrain } from './jobsNodeAffinityTrain';
import { ModelError } from './modelError';
import { NodeAffinity } from './nodeAffinity';
import { NodePoolAssignedResources } from './nodePoolAssignedResources';
import { NodePoolAssignedResourcesAllOf } from './nodePoolAssignedResourcesAllOf';
import { NodePoolAssignedResourcesCreate } from './nodePoolAssignedResourcesCreate';
import { NodePoolAssignedResourcesCreateCpu } from './nodePoolAssignedResourcesCreateCpu';
import { NodePoolAssignedResourcesCreateGpu } from './nodePoolAssignedResourcesCreateGpu';
import { NodePoolAssignedResourcesCreateMemory } from './nodePoolAssignedResourcesCreateMemory';
import { NodePoolAssignedResourcesCreateNodePool } from './nodePoolAssignedResourcesCreateNodePool';
import { NodePoolQuotaStatus } from './nodePoolQuotaStatus';
import { NodePoolResources } from './nodePoolResources';
import { NodeTypeForProject } from './nodeTypeForProject';
import { Project } from './project';
import { ProjectAllOf } from './projectAllOf';
import { ProjectCommonFields } from './projectCommonFields';
import { ProjectCommonFieldsNodeAffinity } from './projectCommonFieldsNodeAffinity';
import { ProjectCommonFieldsPermissions } from './projectCommonFieldsPermissions';
import { ProjectCommonFieldsResources } from './projectCommonFieldsResources';
import { ProjectCreateRequest } from './projectCreateRequest';
import { ProjectCreateRequestAllOf } from './projectCreateRequestAllOf';
import { ProjectStatus } from './projectStatus';
import { ProjectStatusRequest } from './projectStatusRequest';
import { ProjectUpdateRequest } from './projectUpdateRequest';
import { ProjectUpdateRequestAllOf } from './projectUpdateRequestAllOf';
import { QuotaStatusResourceList } from './quotaStatusResourceList';
import { Resource } from './resource';
import { ResourcePermissions } from './resourcePermissions';

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
        "JobsNodeAffinityInteractive.AffinityTypeEnum": JobsNodeAffinityInteractive.AffinityTypeEnum,
        "JobsNodeAffinityTrain.AffinityTypeEnum": JobsNodeAffinityTrain.AffinityTypeEnum,
        "NodeAffinity.AffinityTypeEnum": NodeAffinity.AffinityTypeEnum,
}

let typeMap: {[index: string]: any} = {
    "AssignedResources": AssignedResources,
    "AssignedResourcesCpu": AssignedResourcesCpu,
    "AssignedResourcesGpu": AssignedResourcesGpu,
    "AssignedResourcesMemory": AssignedResourcesMemory,
    "Department": Department,
    "DepartmentAccessControl": DepartmentAccessControl,
    "DepartmentAllOf": DepartmentAllOf,
    "DepartmentCommonFields": DepartmentCommonFields,
    "DepartmentCommonFieldsResources": DepartmentCommonFieldsResources,
    "DepartmentCreateRequest": DepartmentCreateRequest,
    "DepartmentCreateRequestAllOf": DepartmentCreateRequestAllOf,
    "DepartmentUpdateRequest": DepartmentUpdateRequest,
    "DepartmentUpdateRequestAllOf": DepartmentUpdateRequestAllOf,
    "JobsNodeAffinity": JobsNodeAffinity,
    "JobsNodeAffinityInteractive": JobsNodeAffinityInteractive,
    "JobsNodeAffinityTrain": JobsNodeAffinityTrain,
    "ModelError": ModelError,
    "NodeAffinity": NodeAffinity,
    "NodePoolAssignedResources": NodePoolAssignedResources,
    "NodePoolAssignedResourcesAllOf": NodePoolAssignedResourcesAllOf,
    "NodePoolAssignedResourcesCreate": NodePoolAssignedResourcesCreate,
    "NodePoolAssignedResourcesCreateCpu": NodePoolAssignedResourcesCreateCpu,
    "NodePoolAssignedResourcesCreateGpu": NodePoolAssignedResourcesCreateGpu,
    "NodePoolAssignedResourcesCreateMemory": NodePoolAssignedResourcesCreateMemory,
    "NodePoolAssignedResourcesCreateNodePool": NodePoolAssignedResourcesCreateNodePool,
    "NodePoolQuotaStatus": NodePoolQuotaStatus,
    "NodePoolResources": NodePoolResources,
    "NodeTypeForProject": NodeTypeForProject,
    "Project": Project,
    "ProjectAllOf": ProjectAllOf,
    "ProjectCommonFields": ProjectCommonFields,
    "ProjectCommonFieldsNodeAffinity": ProjectCommonFieldsNodeAffinity,
    "ProjectCommonFieldsPermissions": ProjectCommonFieldsPermissions,
    "ProjectCommonFieldsResources": ProjectCommonFieldsResources,
    "ProjectCreateRequest": ProjectCreateRequest,
    "ProjectCreateRequestAllOf": ProjectCreateRequestAllOf,
    "ProjectStatus": ProjectStatus,
    "ProjectStatusRequest": ProjectStatusRequest,
    "ProjectUpdateRequest": ProjectUpdateRequest,
    "ProjectUpdateRequestAllOf": ProjectUpdateRequestAllOf,
    "QuotaStatusResourceList": QuotaStatusResourceList,
    "Resource": Resource,
    "ResourcePermissions": ResourcePermissions,
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
