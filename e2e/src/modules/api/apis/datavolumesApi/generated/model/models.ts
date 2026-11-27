import localVarRequest from 'request';

export * from './ancestors';
export * from './baseFields';
export * from './complianceIssues';
export * from './complianceIssuesComplianceIssuesInner';
export * from './condition';
export * from './countDatavolumes200Response';
export * from './datavolume';
export * from './datavolumeAllOf';
export * from './datavolumeCreationFields';
export * from './datavolumeInnerFields';
export * from './datavolumeInternal';
export * from './datavolumeInternalAllOf';
export * from './datavolumeNoSharedScopes';
export * from './datavolumeNoSharedScopesAllOf';
export * from './datavolumePatchFields';
export * from './datavolumePhase';
export * from './datavolumeRequestType';
export * from './datavolumeSortFilterFields';
export * from './datavolumeStatus';
export * from './datavolumeStatusForSync';
export * from './datavolumeStatusForSyncAllOf';
export * from './datavolumeStatusInternal';
export * from './datavolumeStatusInternalAllOf';
export * from './datavolumesSettingsGetResponse';
export * from './datavolumesSettingsKeyEnum';
export * from './datavolumesSyncRequest';
export * from './enableDataVolumesSetting';
export * from './enrichmentDataFields';
export * from './getDatavolumes200Response';
export * from './getDatavolumesInternal200Response';
export * from './getEnableDataVolumes';
export * from './getEnableDataVolumesValue';
export * from './getUsedPvcsFromDatavolumes200Response';
export * from './httpResponse';
export * from './internalDatavolume';
export * from './internalDatavolumeCreationFields';
export * from './internalEnrichmentDataFields';
export * from './modelError';
export * from './scopeType';
export * from './sharedScope';
export * from './sharedScopeAllOf';
export * from './sharedScopeBase';
export * from './sharedScopesPatchRequest';
export * from './submissionError';
export * from './updateDatavolumesSettingByKeyRequest';

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
import { BaseFields } from './baseFields';
import { ComplianceIssues } from './complianceIssues';
import { ComplianceIssuesComplianceIssuesInner } from './complianceIssuesComplianceIssuesInner';
import { Condition } from './condition';
import { CountDatavolumes200Response } from './countDatavolumes200Response';
import { Datavolume } from './datavolume';
import { DatavolumeAllOf } from './datavolumeAllOf';
import { DatavolumeCreationFields } from './datavolumeCreationFields';
import { DatavolumeInnerFields } from './datavolumeInnerFields';
import { DatavolumeInternal } from './datavolumeInternal';
import { DatavolumeInternalAllOf } from './datavolumeInternalAllOf';
import { DatavolumeNoSharedScopes } from './datavolumeNoSharedScopes';
import { DatavolumeNoSharedScopesAllOf } from './datavolumeNoSharedScopesAllOf';
import { DatavolumePatchFields } from './datavolumePatchFields';
import { DatavolumePhase } from './datavolumePhase';
import { DatavolumeRequestType } from './datavolumeRequestType';
import { DatavolumeSortFilterFields } from './datavolumeSortFilterFields';
import { DatavolumeStatus } from './datavolumeStatus';
import { DatavolumeStatusForSync } from './datavolumeStatusForSync';
import { DatavolumeStatusForSyncAllOf } from './datavolumeStatusForSyncAllOf';
import { DatavolumeStatusInternal } from './datavolumeStatusInternal';
import { DatavolumeStatusInternalAllOf } from './datavolumeStatusInternalAllOf';
import { DatavolumesSettingsGetResponse } from './datavolumesSettingsGetResponse';
import { DatavolumesSettingsKeyEnum } from './datavolumesSettingsKeyEnum';
import { DatavolumesSyncRequest } from './datavolumesSyncRequest';
import { EnableDataVolumesSetting } from './enableDataVolumesSetting';
import { EnrichmentDataFields } from './enrichmentDataFields';
import { GetDatavolumes200Response } from './getDatavolumes200Response';
import { GetDatavolumesInternal200Response } from './getDatavolumesInternal200Response';
import { GetEnableDataVolumes } from './getEnableDataVolumes';
import { GetEnableDataVolumesValue } from './getEnableDataVolumesValue';
import { GetUsedPvcsFromDatavolumes200Response } from './getUsedPvcsFromDatavolumes200Response';
import { HttpResponse } from './httpResponse';
import { InternalDatavolume } from './internalDatavolume';
import { InternalDatavolumeCreationFields } from './internalDatavolumeCreationFields';
import { InternalEnrichmentDataFields } from './internalEnrichmentDataFields';
import { ModelError } from './modelError';
import { ScopeType } from './scopeType';
import { SharedScope } from './sharedScope';
import { SharedScopeAllOf } from './sharedScopeAllOf';
import { SharedScopeBase } from './sharedScopeBase';
import { SharedScopesPatchRequest } from './sharedScopesPatchRequest';
import { SubmissionError } from './submissionError';
import { UpdateDatavolumesSettingByKeyRequest } from './updateDatavolumesSettingByKeyRequest';

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
        "DatavolumePhase": DatavolumePhase,
        "DatavolumeRequestType": DatavolumeRequestType,
        "DatavolumeSortFilterFields": DatavolumeSortFilterFields,
        "DatavolumesSettingsKeyEnum": DatavolumesSettingsKeyEnum,
        "ScopeType": ScopeType,
}

let typeMap: {[index: string]: any} = {
    "Ancestors": Ancestors,
    "BaseFields": BaseFields,
    "ComplianceIssues": ComplianceIssues,
    "ComplianceIssuesComplianceIssuesInner": ComplianceIssuesComplianceIssuesInner,
    "Condition": Condition,
    "CountDatavolumes200Response": CountDatavolumes200Response,
    "Datavolume": Datavolume,
    "DatavolumeAllOf": DatavolumeAllOf,
    "DatavolumeCreationFields": DatavolumeCreationFields,
    "DatavolumeInnerFields": DatavolumeInnerFields,
    "DatavolumeInternal": DatavolumeInternal,
    "DatavolumeInternalAllOf": DatavolumeInternalAllOf,
    "DatavolumeNoSharedScopes": DatavolumeNoSharedScopes,
    "DatavolumeNoSharedScopesAllOf": DatavolumeNoSharedScopesAllOf,
    "DatavolumePatchFields": DatavolumePatchFields,
    "DatavolumeStatus": DatavolumeStatus,
    "DatavolumeStatusForSync": DatavolumeStatusForSync,
    "DatavolumeStatusForSyncAllOf": DatavolumeStatusForSyncAllOf,
    "DatavolumeStatusInternal": DatavolumeStatusInternal,
    "DatavolumeStatusInternalAllOf": DatavolumeStatusInternalAllOf,
    "DatavolumesSettingsGetResponse": DatavolumesSettingsGetResponse,
    "DatavolumesSyncRequest": DatavolumesSyncRequest,
    "EnableDataVolumesSetting": EnableDataVolumesSetting,
    "EnrichmentDataFields": EnrichmentDataFields,
    "GetDatavolumes200Response": GetDatavolumes200Response,
    "GetDatavolumesInternal200Response": GetDatavolumesInternal200Response,
    "GetEnableDataVolumes": GetEnableDataVolumes,
    "GetEnableDataVolumesValue": GetEnableDataVolumesValue,
    "GetUsedPvcsFromDatavolumes200Response": GetUsedPvcsFromDatavolumes200Response,
    "HttpResponse": HttpResponse,
    "InternalDatavolume": InternalDatavolume,
    "InternalDatavolumeCreationFields": InternalDatavolumeCreationFields,
    "InternalEnrichmentDataFields": InternalEnrichmentDataFields,
    "ModelError": ModelError,
    "SharedScope": SharedScope,
    "SharedScopeAllOf": SharedScopeAllOf,
    "SharedScopeBase": SharedScopeBase,
    "SharedScopesPatchRequest": SharedScopesPatchRequest,
    "SubmissionError": SubmissionError,
    "UpdateDatavolumesSettingByKeyRequest": UpdateDatavolumesSettingByKeyRequest,
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
