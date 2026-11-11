import localVarRequest from 'request';

export * from './accessModeCustomization';
export * from './claimSizeCustomization';
export * from './clusterObjectIssue';
export * from './clusterObjectIssueType';
export * from './clusterObjectStatus';
export * from './clusterObjectsStatus';
export * from './clusterResourceType';
export * from './complianceIssues';
export * from './complianceIssuesComplianceIssuesInner';
export * from './configMap';
export * from './configMapListInfo';
export * from './configMaps';
export * from './configMapsListResponse';
export * from './createClusterWidePvcV1Request';
export * from './issue';
export * from './modelError';
export * from './origin';
export * from './pVC';
export * from './pVCCreationRequest';
export * from './pVCListInfo';
export * from './pVCListResponse';
export * from './pVCNames';
export * from './pVCs';
export * from './pvcAccessModes';
export * from './pvcAddedAttribute';
export * from './pvcAddedAttributeFormat';
export * from './pvcAttributesToRemove';
export * from './pvcPhaseOptional';
export * from './pvcSupportedAccessModes';
export * from './pvcVolumeMode';
export * from './secret';
export * from './secretCreationRequest';
export * from './secretListInfo';
export * from './secretListResponse';
export * from './secretType';
export * from './secrets';
export * from './singlePVC';
export * from './singleSecret';
export * from './storageClass';
export * from './storageClassConfiguration';
export * from './storageClassConfigurationCreationRequest';
export * from './storageClassConfigurationListResponse';
export * from './storageClassConfigurationPatchRequest';
export * from './storageClassConfigurationPatchRequestCustomization';
export * from './storageClassConfigurationUpdateRequest';
export * from './storageClassCustomization';
export * from './storageClassPermissions';
export * from './storageClassSpec';
export * from './storageClassV2';
export * from './storageClasses';
export * from './storageClassesV2';
export * from './submissionError';
export * from './v1Error';
export * from './v1GetStorageClasses200Response';
export * from './v1StorageClass';
export * from './volumeModeCustomization';
export * from './volumeModeCustomizationSupportedValues';

import * as fs from 'fs';

export interface RequestDetailedFile {
    value: Buffer;
    options?: {
        filename?: string;
        contentType?: string;
    }
}

export type RequestFile = string | Buffer | fs.ReadStream | RequestDetailedFile;


import { AccessModeCustomization } from './accessModeCustomization';
import { ClaimSizeCustomization } from './claimSizeCustomization';
import { ClusterObjectIssue } from './clusterObjectIssue';
import { ClusterObjectIssueType } from './clusterObjectIssueType';
import { ClusterObjectStatus } from './clusterObjectStatus';
import { ClusterObjectsStatus } from './clusterObjectsStatus';
import { ClusterResourceType } from './clusterResourceType';
import { ComplianceIssues } from './complianceIssues';
import { ComplianceIssuesComplianceIssuesInner } from './complianceIssuesComplianceIssuesInner';
import { ConfigMap } from './configMap';
import { ConfigMapListInfo } from './configMapListInfo';
import { ConfigMaps } from './configMaps';
import { ConfigMapsListResponse } from './configMapsListResponse';
import { CreateClusterWidePvcV1Request } from './createClusterWidePvcV1Request';
import { Issue } from './issue';
import { ModelError } from './modelError';
import { Origin } from './origin';
import { PVC } from './pVC';
import { PVCCreationRequest } from './pVCCreationRequest';
import { PVCListInfo } from './pVCListInfo';
import { PVCListResponse } from './pVCListResponse';
import { PVCNames } from './pVCNames';
import { PVCs } from './pVCs';
import { PvcAccessModes } from './pvcAccessModes';
import { PvcAddedAttribute } from './pvcAddedAttribute';
import { PvcAddedAttributeFormat } from './pvcAddedAttributeFormat';
import { PvcAttributesToRemove } from './pvcAttributesToRemove';
import { PvcPhaseOptional } from './pvcPhaseOptional';
import { PvcSupportedAccessModes } from './pvcSupportedAccessModes';
import { PvcVolumeMode } from './pvcVolumeMode';
import { Secret } from './secret';
import { SecretCreationRequest } from './secretCreationRequest';
import { SecretListInfo } from './secretListInfo';
import { SecretListResponse } from './secretListResponse';
import { SecretType } from './secretType';
import { Secrets } from './secrets';
import { SinglePVC } from './singlePVC';
import { SingleSecret } from './singleSecret';
import { StorageClass } from './storageClass';
import { StorageClassConfiguration } from './storageClassConfiguration';
import { StorageClassConfigurationCreationRequest } from './storageClassConfigurationCreationRequest';
import { StorageClassConfigurationListResponse } from './storageClassConfigurationListResponse';
import { StorageClassConfigurationPatchRequest } from './storageClassConfigurationPatchRequest';
import { StorageClassConfigurationPatchRequestCustomization } from './storageClassConfigurationPatchRequestCustomization';
import { StorageClassConfigurationUpdateRequest } from './storageClassConfigurationUpdateRequest';
import { StorageClassCustomization } from './storageClassCustomization';
import { StorageClassPermissions } from './storageClassPermissions';
import { StorageClassSpec } from './storageClassSpec';
import { StorageClassV2 } from './storageClassV2';
import { StorageClasses } from './storageClasses';
import { StorageClassesV2 } from './storageClassesV2';
import { SubmissionError } from './submissionError';
import { V1Error } from './v1Error';
import { V1GetStorageClasses200Response } from './v1GetStorageClasses200Response';
import { V1StorageClass } from './v1StorageClass';
import { VolumeModeCustomization } from './volumeModeCustomization';
import { VolumeModeCustomizationSupportedValues } from './volumeModeCustomizationSupportedValues';

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
        "ClaimSizeCustomization.SupportedUnitsEnum": ClaimSizeCustomization.SupportedUnitsEnum,
        "ClusterObjectIssueType": ClusterObjectIssueType,
        "ClusterResourceType": ClusterResourceType,
        "Origin": Origin,
        "PvcAddedAttributeFormat": PvcAddedAttributeFormat,
        "PvcPhaseOptional": PvcPhaseOptional,
        "PvcVolumeMode": PvcVolumeMode,
        "SecretType": SecretType,
        "StorageClassConfiguration.ScopeTypeEnum": StorageClassConfiguration.ScopeTypeEnum,
        "StorageClassConfigurationCreationRequest.ScopeTypeEnum": StorageClassConfigurationCreationRequest.ScopeTypeEnum,
}

let typeMap: {[index: string]: any} = {
    "AccessModeCustomization": AccessModeCustomization,
    "ClaimSizeCustomization": ClaimSizeCustomization,
    "ClusterObjectIssue": ClusterObjectIssue,
    "ClusterObjectStatus": ClusterObjectStatus,
    "ClusterObjectsStatus": ClusterObjectsStatus,
    "ComplianceIssues": ComplianceIssues,
    "ComplianceIssuesComplianceIssuesInner": ComplianceIssuesComplianceIssuesInner,
    "ConfigMap": ConfigMap,
    "ConfigMapListInfo": ConfigMapListInfo,
    "ConfigMaps": ConfigMaps,
    "ConfigMapsListResponse": ConfigMapsListResponse,
    "CreateClusterWidePvcV1Request": CreateClusterWidePvcV1Request,
    "Issue": Issue,
    "ModelError": ModelError,
    "PVC": PVC,
    "PVCCreationRequest": PVCCreationRequest,
    "PVCListInfo": PVCListInfo,
    "PVCListResponse": PVCListResponse,
    "PVCNames": PVCNames,
    "PVCs": PVCs,
    "PvcAccessModes": PvcAccessModes,
    "PvcAddedAttribute": PvcAddedAttribute,
    "PvcAttributesToRemove": PvcAttributesToRemove,
    "PvcSupportedAccessModes": PvcSupportedAccessModes,
    "Secret": Secret,
    "SecretCreationRequest": SecretCreationRequest,
    "SecretListInfo": SecretListInfo,
    "SecretListResponse": SecretListResponse,
    "Secrets": Secrets,
    "SinglePVC": SinglePVC,
    "SingleSecret": SingleSecret,
    "StorageClass": StorageClass,
    "StorageClassConfiguration": StorageClassConfiguration,
    "StorageClassConfigurationCreationRequest": StorageClassConfigurationCreationRequest,
    "StorageClassConfigurationListResponse": StorageClassConfigurationListResponse,
    "StorageClassConfigurationPatchRequest": StorageClassConfigurationPatchRequest,
    "StorageClassConfigurationPatchRequestCustomization": StorageClassConfigurationPatchRequestCustomization,
    "StorageClassConfigurationUpdateRequest": StorageClassConfigurationUpdateRequest,
    "StorageClassCustomization": StorageClassCustomization,
    "StorageClassPermissions": StorageClassPermissions,
    "StorageClassSpec": StorageClassSpec,
    "StorageClassV2": StorageClassV2,
    "StorageClasses": StorageClasses,
    "StorageClassesV2": StorageClassesV2,
    "SubmissionError": SubmissionError,
    "V1Error": V1Error,
    "V1GetStorageClasses200Response": V1GetStorageClasses200Response,
    "V1StorageClass": V1StorageClass,
    "VolumeModeCustomization": VolumeModeCustomization,
    "VolumeModeCustomizationSupportedValues": VolumeModeCustomizationSupportedValues,
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
