import localVarRequest from 'request';

export * from './complianceIssues';
export * from './complianceIssuesComplianceIssuesInner';
export * from './deleteTenantById200Response';
export * from './extendedMetricStore';
export * from './getTenantLogo200Response';
export * from './logoCreationRequest';
export * from './metricStore';
export * from './metricStoreEndpoint';
export * from './metricStoreEndpointAuth';
export * from './metricStoreEndpointAuthBasic';
export * from './modelError';
export * from './roles';
export * from './submissionError';
export * from './tenant';
export * from './tenantCreationRequest';
export * from './tenantCreationRequestMetricStore';
export * from './tenantCreationResponse';
export * from './tenantLogo';
export * from './tenantLogosResponse';
export * from './tenantUpdateRequest';

import * as fs from 'fs';

export interface RequestDetailedFile {
    value: Buffer;
    options?: {
        filename?: string;
        contentType?: string;
    }
}

export type RequestFile = string | Buffer | fs.ReadStream | RequestDetailedFile;


import { ComplianceIssues } from './complianceIssues';
import { ComplianceIssuesComplianceIssuesInner } from './complianceIssuesComplianceIssuesInner';
import { DeleteTenantById200Response } from './deleteTenantById200Response';
import { ExtendedMetricStore } from './extendedMetricStore';
import { GetTenantLogo200Response } from './getTenantLogo200Response';
import { LogoCreationRequest } from './logoCreationRequest';
import { MetricStore } from './metricStore';
import { MetricStoreEndpoint } from './metricStoreEndpoint';
import { MetricStoreEndpointAuth } from './metricStoreEndpointAuth';
import { MetricStoreEndpointAuthBasic } from './metricStoreEndpointAuthBasic';
import { ModelError } from './modelError';
import { Roles } from './roles';
import { SubmissionError } from './submissionError';
import { Tenant } from './tenant';
import { TenantCreationRequest } from './tenantCreationRequest';
import { TenantCreationRequestMetricStore } from './tenantCreationRequestMetricStore';
import { TenantCreationResponse } from './tenantCreationResponse';
import { TenantLogo } from './tenantLogo';
import { TenantLogosResponse } from './tenantLogosResponse';
import { TenantUpdateRequest } from './tenantUpdateRequest';

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
        "ExtendedMetricStore.TypeEnum": ExtendedMetricStore.TypeEnum,
        "Roles": Roles,
        "Tenant.StatusEnum": Tenant.StatusEnum,
        "Tenant.ContractTypeEnum": Tenant.ContractTypeEnum,
        "TenantCreationRequest.ContractTypeEnum": TenantCreationRequest.ContractTypeEnum,
}

let typeMap: {[index: string]: any} = {
    "ComplianceIssues": ComplianceIssues,
    "ComplianceIssuesComplianceIssuesInner": ComplianceIssuesComplianceIssuesInner,
    "DeleteTenantById200Response": DeleteTenantById200Response,
    "ExtendedMetricStore": ExtendedMetricStore,
    "GetTenantLogo200Response": GetTenantLogo200Response,
    "LogoCreationRequest": LogoCreationRequest,
    "MetricStore": MetricStore,
    "MetricStoreEndpoint": MetricStoreEndpoint,
    "MetricStoreEndpointAuth": MetricStoreEndpointAuth,
    "MetricStoreEndpointAuthBasic": MetricStoreEndpointAuthBasic,
    "ModelError": ModelError,
    "SubmissionError": SubmissionError,
    "Tenant": Tenant,
    "TenantCreationRequest": TenantCreationRequest,
    "TenantCreationRequestMetricStore": TenantCreationRequestMetricStore,
    "TenantCreationResponse": TenantCreationResponse,
    "TenantLogo": TenantLogo,
    "TenantLogosResponse": TenantLogosResponse,
    "TenantUpdateRequest": TenantUpdateRequest,
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
