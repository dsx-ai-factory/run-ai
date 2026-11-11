import localVarRequest from 'request';

export * from './accessKey';
export * from './accessKeyAdministration';
export * from './accessKeyCreationRequest';
export * from './accessKeyPostResponse';
export * from './accessKeys';
export * from './accessKeysAdministration';
export * from './appTokenRequest';
export * from './appTokenResponse';
export * from './application';
export * from './applicationCreationRequest';
export * from './applicationPatchRequest';
export * from './applicationPostResponse';
export * from './authConfiguration';
export * from './bulkUserCreationRequest';
export * from './bulkUserCreationResponse';
export * from './bulkUserCreationResponseResultsInner';
export * from './bulkUserCreationResponseResultsInnerError';
export * from './changeUserPassword200Response';
export * from './client';
export * from './clientAttributeCreationRequest';
export * from './clientCreationRequest';
export * from './clientRedirectUrisUpdateRequest';
export * from './clientType';
export * from './codeChallengeMethod';
export * from './complianceIssues';
export * from './complianceIssuesComplianceIssuesInner';
export * from './configSamlResponse';
export * from './countUsers200Response';
export * from './createInternalClient201Response';
export * from './getAutoRedirectSsoSetting';
export * from './getAutoRedirectSsoSettingValue';
export * from './getBrowserSessionTimeoutSetting';
export * from './getBrowserSessionTimeoutSettingValue';
export * from './getClientSecret200Response';
export * from './getLogoutRedirectUriSetting';
export * from './getLogoutRedirectUriSettingValue';
export * from './idp';
export * from './idpCreationRequest';
export * from './idpCreationRequestTypeEnum';
export * from './idpPostResponse';
export * from './idpResources';
export * from './internalApplicationCreationRequest';
export * from './internalServiceAccountCreationRequest';
export * from './internalUserCreationRequest';
export * from './mandatoryClaim';
export * from './mappers';
export * from './modelError';
export * from './oAuth2Error';
export * from './oAuth2TokenResponse';
export * from './ocpCreationData';
export * from './ocpData';
export * from './oidcCreationData';
export * from './oidcData';
export * from './openIDConfiguration';
export * from './realmCreationRequest';
export * from './realmCreationResponse';
export * from './realmReconcileRequest';
export * from './reconcilerType';
export * from './regenerateAccessKeySecret200Response';
export * from './regenerateApplicationSecret200Response';
export * from './regenerateServiceAccountSecretResponse';
export * from './resetUserPassword200Response';
export * from './samlCreationData';
export * from './samlCreationDataMetadataXmlTypeEnum';
export * from './samlData';
export * from './serviceAccount';
export * from './serviceAccountCreationRequest';
export * from './serviceAccountPatchRequest';
export * from './serviceAccountPostResponse';
export * from './settingsGetResponse';
export * from './settingsKeyEnum';
export * from './submissionError';
export * from './tokenRequest';
export * from './tokenRequestGrantTypeEnum';
export * from './tokenResponse';
export * from './updateAutoRedirectSsoSettingValue';
export * from './updateBrowserSessionTimeoutSettingValue';
export * from './updateLogoutRedirectUriSettingValue';
export * from './updateSecuritySettingByKeyRequest';
export * from './user';
export * from './userApp';
export * from './userAppAdministration';
export * from './userAppCreationRequest';
export * from './userAppPostResponse';
export * from './userChangePasswordRequest';
export * from './userCreationRequest';
export * from './userCreationRequestBulk';
export * from './userMinimal';
export * from './userPatchRequest';
export * from './userPostResponse';
export * from './usersFilterSortFields';

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
import { AccessKeyAdministration } from './accessKeyAdministration';
import { AccessKeyCreationRequest } from './accessKeyCreationRequest';
import { AccessKeyPostResponse } from './accessKeyPostResponse';
import { AccessKeys } from './accessKeys';
import { AccessKeysAdministration } from './accessKeysAdministration';
import { AppTokenRequest } from './appTokenRequest';
import { AppTokenResponse } from './appTokenResponse';
import { Application } from './application';
import { ApplicationCreationRequest } from './applicationCreationRequest';
import { ApplicationPatchRequest } from './applicationPatchRequest';
import { ApplicationPostResponse } from './applicationPostResponse';
import { AuthConfiguration } from './authConfiguration';
import { BulkUserCreationRequest } from './bulkUserCreationRequest';
import { BulkUserCreationResponse } from './bulkUserCreationResponse';
import { BulkUserCreationResponseResultsInner } from './bulkUserCreationResponseResultsInner';
import { BulkUserCreationResponseResultsInnerError } from './bulkUserCreationResponseResultsInnerError';
import { ChangeUserPassword200Response } from './changeUserPassword200Response';
import { Client } from './client';
import { ClientAttributeCreationRequest } from './clientAttributeCreationRequest';
import { ClientCreationRequest } from './clientCreationRequest';
import { ClientRedirectUrisUpdateRequest } from './clientRedirectUrisUpdateRequest';
import { ClientType } from './clientType';
import { CodeChallengeMethod } from './codeChallengeMethod';
import { ComplianceIssues } from './complianceIssues';
import { ComplianceIssuesComplianceIssuesInner } from './complianceIssuesComplianceIssuesInner';
import { ConfigSamlResponse } from './configSamlResponse';
import { CountUsers200Response } from './countUsers200Response';
import { CreateInternalClient201Response } from './createInternalClient201Response';
import { GetAutoRedirectSsoSetting } from './getAutoRedirectSsoSetting';
import { GetAutoRedirectSsoSettingValue } from './getAutoRedirectSsoSettingValue';
import { GetBrowserSessionTimeoutSetting } from './getBrowserSessionTimeoutSetting';
import { GetBrowserSessionTimeoutSettingValue } from './getBrowserSessionTimeoutSettingValue';
import { GetClientSecret200Response } from './getClientSecret200Response';
import { GetLogoutRedirectUriSetting } from './getLogoutRedirectUriSetting';
import { GetLogoutRedirectUriSettingValue } from './getLogoutRedirectUriSettingValue';
import { Idp } from './idp';
import { IdpCreationRequest } from './idpCreationRequest';
import { IdpCreationRequestTypeEnum } from './idpCreationRequestTypeEnum';
import { IdpPostResponse } from './idpPostResponse';
import { IdpResources } from './idpResources';
import { InternalApplicationCreationRequest } from './internalApplicationCreationRequest';
import { InternalServiceAccountCreationRequest } from './internalServiceAccountCreationRequest';
import { InternalUserCreationRequest } from './internalUserCreationRequest';
import { MandatoryClaim } from './mandatoryClaim';
import { Mappers } from './mappers';
import { ModelError } from './modelError';
import { OAuth2Error } from './oAuth2Error';
import { OAuth2TokenResponse } from './oAuth2TokenResponse';
import { OcpCreationData } from './ocpCreationData';
import { OcpData } from './ocpData';
import { OidcCreationData } from './oidcCreationData';
import { OidcData } from './oidcData';
import { OpenIDConfiguration } from './openIDConfiguration';
import { RealmCreationRequest } from './realmCreationRequest';
import { RealmCreationResponse } from './realmCreationResponse';
import { RealmReconcileRequest } from './realmReconcileRequest';
import { ReconcilerType } from './reconcilerType';
import { RegenerateAccessKeySecret200Response } from './regenerateAccessKeySecret200Response';
import { RegenerateApplicationSecret200Response } from './regenerateApplicationSecret200Response';
import { RegenerateServiceAccountSecretResponse } from './regenerateServiceAccountSecretResponse';
import { ResetUserPassword200Response } from './resetUserPassword200Response';
import { SamlCreationData } from './samlCreationData';
import { SamlCreationDataMetadataXmlTypeEnum } from './samlCreationDataMetadataXmlTypeEnum';
import { SamlData } from './samlData';
import { ServiceAccount } from './serviceAccount';
import { ServiceAccountCreationRequest } from './serviceAccountCreationRequest';
import { ServiceAccountPatchRequest } from './serviceAccountPatchRequest';
import { ServiceAccountPostResponse } from './serviceAccountPostResponse';
import { SettingsGetResponse } from './settingsGetResponse';
import { SettingsKeyEnum } from './settingsKeyEnum';
import { SubmissionError } from './submissionError';
import { TokenRequest } from './tokenRequest';
import { TokenRequestGrantTypeEnum } from './tokenRequestGrantTypeEnum';
import { TokenResponse } from './tokenResponse';
import { UpdateAutoRedirectSsoSettingValue } from './updateAutoRedirectSsoSettingValue';
import { UpdateBrowserSessionTimeoutSettingValue } from './updateBrowserSessionTimeoutSettingValue';
import { UpdateLogoutRedirectUriSettingValue } from './updateLogoutRedirectUriSettingValue';
import { UpdateSecuritySettingByKeyRequest } from './updateSecuritySettingByKeyRequest';
import { User } from './user';
import { UserApp } from './userApp';
import { UserAppAdministration } from './userAppAdministration';
import { UserAppCreationRequest } from './userAppCreationRequest';
import { UserAppPostResponse } from './userAppPostResponse';
import { UserChangePasswordRequest } from './userChangePasswordRequest';
import { UserCreationRequest } from './userCreationRequest';
import { UserCreationRequestBulk } from './userCreationRequestBulk';
import { UserMinimal } from './userMinimal';
import { UserPatchRequest } from './userPatchRequest';
import { UserPostResponse } from './userPostResponse';
import { UsersFilterSortFields } from './usersFilterSortFields';

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
        "ClientType": ClientType,
        "CodeChallengeMethod": CodeChallengeMethod,
        "Idp.TypeEnum": Idp.TypeEnum,
        "IdpCreationRequestTypeEnum": IdpCreationRequestTypeEnum,
        "IdpResources.TypeEnum": IdpResources.TypeEnum,
        "OAuth2Error.ErrorEnum": OAuth2Error.ErrorEnum,
        "ReconcilerType": ReconcilerType,
        "SamlCreationDataMetadataXmlTypeEnum": SamlCreationDataMetadataXmlTypeEnum,
        "SettingsKeyEnum": SettingsKeyEnum,
        "TokenRequestGrantTypeEnum": TokenRequestGrantTypeEnum,
        "UsersFilterSortFields": UsersFilterSortFields,
}

let typeMap: {[index: string]: any} = {
    "AccessKey": AccessKey,
    "AccessKeyAdministration": AccessKeyAdministration,
    "AccessKeyCreationRequest": AccessKeyCreationRequest,
    "AccessKeyPostResponse": AccessKeyPostResponse,
    "AccessKeys": AccessKeys,
    "AccessKeysAdministration": AccessKeysAdministration,
    "AppTokenRequest": AppTokenRequest,
    "AppTokenResponse": AppTokenResponse,
    "Application": Application,
    "ApplicationCreationRequest": ApplicationCreationRequest,
    "ApplicationPatchRequest": ApplicationPatchRequest,
    "ApplicationPostResponse": ApplicationPostResponse,
    "AuthConfiguration": AuthConfiguration,
    "BulkUserCreationRequest": BulkUserCreationRequest,
    "BulkUserCreationResponse": BulkUserCreationResponse,
    "BulkUserCreationResponseResultsInner": BulkUserCreationResponseResultsInner,
    "BulkUserCreationResponseResultsInnerError": BulkUserCreationResponseResultsInnerError,
    "ChangeUserPassword200Response": ChangeUserPassword200Response,
    "Client": Client,
    "ClientAttributeCreationRequest": ClientAttributeCreationRequest,
    "ClientCreationRequest": ClientCreationRequest,
    "ClientRedirectUrisUpdateRequest": ClientRedirectUrisUpdateRequest,
    "ComplianceIssues": ComplianceIssues,
    "ComplianceIssuesComplianceIssuesInner": ComplianceIssuesComplianceIssuesInner,
    "ConfigSamlResponse": ConfigSamlResponse,
    "CountUsers200Response": CountUsers200Response,
    "CreateInternalClient201Response": CreateInternalClient201Response,
    "GetAutoRedirectSsoSetting": GetAutoRedirectSsoSetting,
    "GetAutoRedirectSsoSettingValue": GetAutoRedirectSsoSettingValue,
    "GetBrowserSessionTimeoutSetting": GetBrowserSessionTimeoutSetting,
    "GetBrowserSessionTimeoutSettingValue": GetBrowserSessionTimeoutSettingValue,
    "GetClientSecret200Response": GetClientSecret200Response,
    "GetLogoutRedirectUriSetting": GetLogoutRedirectUriSetting,
    "GetLogoutRedirectUriSettingValue": GetLogoutRedirectUriSettingValue,
    "Idp": Idp,
    "IdpCreationRequest": IdpCreationRequest,
    "IdpPostResponse": IdpPostResponse,
    "IdpResources": IdpResources,
    "InternalApplicationCreationRequest": InternalApplicationCreationRequest,
    "InternalServiceAccountCreationRequest": InternalServiceAccountCreationRequest,
    "InternalUserCreationRequest": InternalUserCreationRequest,
    "MandatoryClaim": MandatoryClaim,
    "Mappers": Mappers,
    "ModelError": ModelError,
    "OAuth2Error": OAuth2Error,
    "OAuth2TokenResponse": OAuth2TokenResponse,
    "OcpCreationData": OcpCreationData,
    "OcpData": OcpData,
    "OidcCreationData": OidcCreationData,
    "OidcData": OidcData,
    "OpenIDConfiguration": OpenIDConfiguration,
    "RealmCreationRequest": RealmCreationRequest,
    "RealmCreationResponse": RealmCreationResponse,
    "RealmReconcileRequest": RealmReconcileRequest,
    "RegenerateAccessKeySecret200Response": RegenerateAccessKeySecret200Response,
    "RegenerateApplicationSecret200Response": RegenerateApplicationSecret200Response,
    "RegenerateServiceAccountSecretResponse": RegenerateServiceAccountSecretResponse,
    "ResetUserPassword200Response": ResetUserPassword200Response,
    "SamlCreationData": SamlCreationData,
    "SamlData": SamlData,
    "ServiceAccount": ServiceAccount,
    "ServiceAccountCreationRequest": ServiceAccountCreationRequest,
    "ServiceAccountPatchRequest": ServiceAccountPatchRequest,
    "ServiceAccountPostResponse": ServiceAccountPostResponse,
    "SettingsGetResponse": SettingsGetResponse,
    "SubmissionError": SubmissionError,
    "TokenRequest": TokenRequest,
    "TokenResponse": TokenResponse,
    "UpdateAutoRedirectSsoSettingValue": UpdateAutoRedirectSsoSettingValue,
    "UpdateBrowserSessionTimeoutSettingValue": UpdateBrowserSessionTimeoutSettingValue,
    "UpdateLogoutRedirectUriSettingValue": UpdateLogoutRedirectUriSettingValue,
    "UpdateSecuritySettingByKeyRequest": UpdateSecuritySettingByKeyRequest,
    "User": User,
    "UserApp": UserApp,
    "UserAppAdministration": UserAppAdministration,
    "UserAppCreationRequest": UserAppCreationRequest,
    "UserAppPostResponse": UserAppPostResponse,
    "UserChangePasswordRequest": UserChangePasswordRequest,
    "UserCreationRequest": UserCreationRequest,
    "UserCreationRequestBulk": UserCreationRequestBulk,
    "UserMinimal": UserMinimal,
    "UserPatchRequest": UserPatchRequest,
    "UserPostResponse": UserPostResponse,
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
