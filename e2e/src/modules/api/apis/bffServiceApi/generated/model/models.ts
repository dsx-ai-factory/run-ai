import localVarRequest from 'request';

export * from './appearanceSettingsGetAllResponse';
export * from './appearanceSettingsKeyEnum';
export * from './brandingSettingsGetAllResponse';
export * from './brandingSettingsKeyEnum';
export * from './complianceIssues';
export * from './complianceIssuesComplianceIssuesInner';
export * from './getAdminMessage';
export * from './getAdminMessageValue';
export * from './getAllowGPT2Model';
export * from './getAllowGPT2ModelValue';
export * from './getBannerConfig';
export * from './getBannerConfigValue';
export * from './getDeletedWorkloadsView';
export * from './getDeletedWorkloadsViewValue';
export * from './getEnableModelCatalog';
export * from './getEnableModelCatalogValue';
export * from './getEnableNIM';
export * from './getEnableNIMValue';
export * from './getLogo';
export * from './getLogoValue';
export * from './getSystemNotification';
export * from './modelError';
export * from './notificationsSettingsGetAllResponse';
export * from './notificationsSettingsKeyEnum';
export * from './submissionError';
export * from './updateAdminMessage';
export * from './updateAllowGPT2Model';
export * from './updateAppearanceSettingByKeyRequest';
export * from './updateBannerConfig';
export * from './updateBrandingSettingByKeyRequest';
export * from './updateDeletedWorkloadsView';
export * from './updateEnableModelCatalog';
export * from './updateEnableNIM';
export * from './updateLogo';
export * from './updateNotificationsSettingByKeyRequest';
export * from './updateSystemNotification';
export * from './updateWorkloadsSettingByKeyRequest';
export * from './versionResponse';
export * from './workloadsSettingsGetAllResponse';
export * from './workloadsSettingsKeyEnum';

import * as fs from 'fs';

export interface RequestDetailedFile {
    value: Buffer;
    options?: {
        filename?: string;
        contentType?: string;
    }
}

export type RequestFile = string | Buffer | fs.ReadStream | RequestDetailedFile;


import { AppearanceSettingsGetAllResponse } from './appearanceSettingsGetAllResponse';
import { AppearanceSettingsKeyEnum } from './appearanceSettingsKeyEnum';
import { BrandingSettingsGetAllResponse } from './brandingSettingsGetAllResponse';
import { BrandingSettingsKeyEnum } from './brandingSettingsKeyEnum';
import { ComplianceIssues } from './complianceIssues';
import { ComplianceIssuesComplianceIssuesInner } from './complianceIssuesComplianceIssuesInner';
import { GetAdminMessage } from './getAdminMessage';
import { GetAdminMessageValue } from './getAdminMessageValue';
import { GetAllowGPT2Model } from './getAllowGPT2Model';
import { GetAllowGPT2ModelValue } from './getAllowGPT2ModelValue';
import { GetBannerConfig } from './getBannerConfig';
import { GetBannerConfigValue } from './getBannerConfigValue';
import { GetDeletedWorkloadsView } from './getDeletedWorkloadsView';
import { GetDeletedWorkloadsViewValue } from './getDeletedWorkloadsViewValue';
import { GetEnableModelCatalog } from './getEnableModelCatalog';
import { GetEnableModelCatalogValue } from './getEnableModelCatalogValue';
import { GetEnableNIM } from './getEnableNIM';
import { GetEnableNIMValue } from './getEnableNIMValue';
import { GetLogo } from './getLogo';
import { GetLogoValue } from './getLogoValue';
import { GetSystemNotification } from './getSystemNotification';
import { ModelError } from './modelError';
import { NotificationsSettingsGetAllResponse } from './notificationsSettingsGetAllResponse';
import { NotificationsSettingsKeyEnum } from './notificationsSettingsKeyEnum';
import { SubmissionError } from './submissionError';
import { UpdateAdminMessage } from './updateAdminMessage';
import { UpdateAllowGPT2Model } from './updateAllowGPT2Model';
import { UpdateAppearanceSettingByKeyRequest } from './updateAppearanceSettingByKeyRequest';
import { UpdateBannerConfig } from './updateBannerConfig';
import { UpdateBrandingSettingByKeyRequest } from './updateBrandingSettingByKeyRequest';
import { UpdateDeletedWorkloadsView } from './updateDeletedWorkloadsView';
import { UpdateEnableModelCatalog } from './updateEnableModelCatalog';
import { UpdateEnableNIM } from './updateEnableNIM';
import { UpdateLogo } from './updateLogo';
import { UpdateNotificationsSettingByKeyRequest } from './updateNotificationsSettingByKeyRequest';
import { UpdateSystemNotification } from './updateSystemNotification';
import { UpdateWorkloadsSettingByKeyRequest } from './updateWorkloadsSettingByKeyRequest';
import { VersionResponse } from './versionResponse';
import { WorkloadsSettingsGetAllResponse } from './workloadsSettingsGetAllResponse';
import { WorkloadsSettingsKeyEnum } from './workloadsSettingsKeyEnum';

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
        "AppearanceSettingsKeyEnum": AppearanceSettingsKeyEnum,
        "BrandingSettingsKeyEnum": BrandingSettingsKeyEnum,
        "NotificationsSettingsKeyEnum": NotificationsSettingsKeyEnum,
        "WorkloadsSettingsKeyEnum": WorkloadsSettingsKeyEnum,
}

let typeMap: {[index: string]: any} = {
    "AppearanceSettingsGetAllResponse": AppearanceSettingsGetAllResponse,
    "BrandingSettingsGetAllResponse": BrandingSettingsGetAllResponse,
    "ComplianceIssues": ComplianceIssues,
    "ComplianceIssuesComplianceIssuesInner": ComplianceIssuesComplianceIssuesInner,
    "GetAdminMessage": GetAdminMessage,
    "GetAdminMessageValue": GetAdminMessageValue,
    "GetAllowGPT2Model": GetAllowGPT2Model,
    "GetAllowGPT2ModelValue": GetAllowGPT2ModelValue,
    "GetBannerConfig": GetBannerConfig,
    "GetBannerConfigValue": GetBannerConfigValue,
    "GetDeletedWorkloadsView": GetDeletedWorkloadsView,
    "GetDeletedWorkloadsViewValue": GetDeletedWorkloadsViewValue,
    "GetEnableModelCatalog": GetEnableModelCatalog,
    "GetEnableModelCatalogValue": GetEnableModelCatalogValue,
    "GetEnableNIM": GetEnableNIM,
    "GetEnableNIMValue": GetEnableNIMValue,
    "GetLogo": GetLogo,
    "GetLogoValue": GetLogoValue,
    "GetSystemNotification": GetSystemNotification,
    "ModelError": ModelError,
    "NotificationsSettingsGetAllResponse": NotificationsSettingsGetAllResponse,
    "SubmissionError": SubmissionError,
    "UpdateAdminMessage": UpdateAdminMessage,
    "UpdateAllowGPT2Model": UpdateAllowGPT2Model,
    "UpdateAppearanceSettingByKeyRequest": UpdateAppearanceSettingByKeyRequest,
    "UpdateBannerConfig": UpdateBannerConfig,
    "UpdateBrandingSettingByKeyRequest": UpdateBrandingSettingByKeyRequest,
    "UpdateDeletedWorkloadsView": UpdateDeletedWorkloadsView,
    "UpdateEnableModelCatalog": UpdateEnableModelCatalog,
    "UpdateEnableNIM": UpdateEnableNIM,
    "UpdateLogo": UpdateLogo,
    "UpdateNotificationsSettingByKeyRequest": UpdateNotificationsSettingByKeyRequest,
    "UpdateSystemNotification": UpdateSystemNotification,
    "UpdateWorkloadsSettingByKeyRequest": UpdateWorkloadsSettingByKeyRequest,
    "VersionResponse": VersionResponse,
    "WorkloadsSettingsGetAllResponse": WorkloadsSettingsGetAllResponse,
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
