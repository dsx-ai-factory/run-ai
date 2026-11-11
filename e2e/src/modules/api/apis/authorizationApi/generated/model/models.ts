import localVarRequest from 'request';

export * from './accessRule';
export * from './accessRuleAllOf';
export * from './accessRuleById';
export * from './accessRuleByIdAllOf';
export * from './accessRuleClusterPhase';
export * from './accessRuleCreationFields';
export * from './accessRuleDeletionBatchResponse';
export * from './accessRuleDeletionBatchResponseFailureItem';
export * from './accessRuleForSync';
export * from './accessRuleForSyncAllOf';
export * from './accessRuleMetadataFields';
export * from './accessRulePhase';
export * from './accessRuleScopesInner';
export * from './accessRuleStatusRequest';
export * from './accessRuleStatusResponse';
export * from './accessRuleSubjectsInner';
export * from './accessRuleSync';
export * from './accessRuleSyncAllOf';
export * from './accessRuleSyncAllOfPredefinedRole';
export * from './accessRulesBatchAction';
export * from './accessRulesBatchCreateFields';
export * from './accessRulesBatchCreateFieldsPayload';
export * from './accessRulesBatchCreateResponse';
export * from './accessRulesBatchCreateResponseFailedInner';
export * from './accessRulesBatchFields';
export * from './accessRulesMigrateDeprecatedFailureItem';
export * from './accessRulesMigrateDeprecatedFields';
export * from './accessRulesMigrateDeprecatedResponse';
export * from './accessRulesMigrateDeprecatedResponseSucceeded';
export * from './accessRulesSortFilterFields';
export * from './action';
export * from './allowedInternalCreateRoles';
export * from './authorizationSettingsGetResponse';
export * from './authorizationSettingsKeyEnum';
export * from './backendEvaluationRequest';
export * from './batchResponse';
export * from './batchResponseFailureItem';
export * from './complianceIssues';
export * from './complianceIssuesComplianceIssuesInner';
export * from './countAccessRules200Response';
export * from './enableScopedAuthorizationSetting';
export * from './enrichmentDataFields';
export * from './getAccessRules200Response';
export * from './getAccessRulesSync200Response';
export * from './getGroups200Response';
export * from './getPermissionSets200Response';
export * from './getPermittedScopesRequest';
export * from './getRoles200Response';
export * from './internalCreateAccessRuleRequest';
export * from './internalCreateAccessRuleRequestScope';
export * from './kubernetesPermissions';
export * from './meAccessRule';
export * from './meAccessRuleAllOf';
export * from './meAccessRulesResponse';
export * from './meAccessRulesResponseAccessRules';
export * from './modelError';
export * from './permission';
export * from './permissionAction';
export * from './permissionSet';
export * from './permissionSetAllOf';
export * from './permittedScopes';
export * from './permittedScopesActions';
export * from './postPermission';
export * from './resourceType';
export * from './resourceTypeGroupId';
export * from './role';
export * from './roleActiveFields';
export * from './roleCalculatedFields';
export * from './roleCreationFields';
export * from './roleCreationFieldsInternal';
export * from './roleCreationFieldsInternalAllOf';
export * from './roleCreationFieldsKubernetesPermissions';
export * from './roleEnabledDisabledFields';
export * from './rolePermission';
export * from './rolePermissionSet';
export * from './roleV1';
export * from './rolesSortFilterFields';
export * from './scopeType';
export * from './subjectType';
export * from './submissionError';
export * from './tenantIsCspManagedSetting';
export * from './updateAuthorizationSettingByKeyRequest';
export * from './updateScopeAccessRulesRequest';

import * as fs from 'fs';

export interface RequestDetailedFile {
    value: Buffer;
    options?: {
        filename?: string;
        contentType?: string;
    }
}

export type RequestFile = string | Buffer | fs.ReadStream | RequestDetailedFile;


import { AccessRule } from './accessRule';
import { AccessRuleAllOf } from './accessRuleAllOf';
import { AccessRuleById } from './accessRuleById';
import { AccessRuleByIdAllOf } from './accessRuleByIdAllOf';
import { AccessRuleClusterPhase } from './accessRuleClusterPhase';
import { AccessRuleCreationFields } from './accessRuleCreationFields';
import { AccessRuleDeletionBatchResponse } from './accessRuleDeletionBatchResponse';
import { AccessRuleDeletionBatchResponseFailureItem } from './accessRuleDeletionBatchResponseFailureItem';
import { AccessRuleForSync } from './accessRuleForSync';
import { AccessRuleForSyncAllOf } from './accessRuleForSyncAllOf';
import { AccessRuleMetadataFields } from './accessRuleMetadataFields';
import { AccessRulePhase } from './accessRulePhase';
import { AccessRuleScopesInner } from './accessRuleScopesInner';
import { AccessRuleStatusRequest } from './accessRuleStatusRequest';
import { AccessRuleStatusResponse } from './accessRuleStatusResponse';
import { AccessRuleSubjectsInner } from './accessRuleSubjectsInner';
import { AccessRuleSync } from './accessRuleSync';
import { AccessRuleSyncAllOf } from './accessRuleSyncAllOf';
import { AccessRuleSyncAllOfPredefinedRole } from './accessRuleSyncAllOfPredefinedRole';
import { AccessRulesBatchAction } from './accessRulesBatchAction';
import { AccessRulesBatchCreateFields } from './accessRulesBatchCreateFields';
import { AccessRulesBatchCreateFieldsPayload } from './accessRulesBatchCreateFieldsPayload';
import { AccessRulesBatchCreateResponse } from './accessRulesBatchCreateResponse';
import { AccessRulesBatchCreateResponseFailedInner } from './accessRulesBatchCreateResponseFailedInner';
import { AccessRulesBatchFields } from './accessRulesBatchFields';
import { AccessRulesMigrateDeprecatedFailureItem } from './accessRulesMigrateDeprecatedFailureItem';
import { AccessRulesMigrateDeprecatedFields } from './accessRulesMigrateDeprecatedFields';
import { AccessRulesMigrateDeprecatedResponse } from './accessRulesMigrateDeprecatedResponse';
import { AccessRulesMigrateDeprecatedResponseSucceeded } from './accessRulesMigrateDeprecatedResponseSucceeded';
import { AccessRulesSortFilterFields } from './accessRulesSortFilterFields';
import { Action } from './action';
import { AllowedInternalCreateRoles } from './allowedInternalCreateRoles';
import { AuthorizationSettingsGetResponse } from './authorizationSettingsGetResponse';
import { AuthorizationSettingsKeyEnum } from './authorizationSettingsKeyEnum';
import { BackendEvaluationRequest } from './backendEvaluationRequest';
import { BatchResponse } from './batchResponse';
import { BatchResponseFailureItem } from './batchResponseFailureItem';
import { ComplianceIssues } from './complianceIssues';
import { ComplianceIssuesComplianceIssuesInner } from './complianceIssuesComplianceIssuesInner';
import { CountAccessRules200Response } from './countAccessRules200Response';
import { EnableScopedAuthorizationSetting } from './enableScopedAuthorizationSetting';
import { EnrichmentDataFields } from './enrichmentDataFields';
import { GetAccessRules200Response } from './getAccessRules200Response';
import { GetAccessRulesSync200Response } from './getAccessRulesSync200Response';
import { GetGroups200Response } from './getGroups200Response';
import { GetPermissionSets200Response } from './getPermissionSets200Response';
import { GetPermittedScopesRequest } from './getPermittedScopesRequest';
import { GetRoles200Response } from './getRoles200Response';
import { InternalCreateAccessRuleRequest } from './internalCreateAccessRuleRequest';
import { InternalCreateAccessRuleRequestScope } from './internalCreateAccessRuleRequestScope';
import { KubernetesPermissions } from './kubernetesPermissions';
import { MeAccessRule } from './meAccessRule';
import { MeAccessRuleAllOf } from './meAccessRuleAllOf';
import { MeAccessRulesResponse } from './meAccessRulesResponse';
import { MeAccessRulesResponseAccessRules } from './meAccessRulesResponseAccessRules';
import { ModelError } from './modelError';
import { Permission } from './permission';
import { PermissionAction } from './permissionAction';
import { PermissionSet } from './permissionSet';
import { PermissionSetAllOf } from './permissionSetAllOf';
import { PermittedScopes } from './permittedScopes';
import { PermittedScopesActions } from './permittedScopesActions';
import { PostPermission } from './postPermission';
import { ResourceType } from './resourceType';
import { ResourceTypeGroupId } from './resourceTypeGroupId';
import { Role } from './role';
import { RoleActiveFields } from './roleActiveFields';
import { RoleCalculatedFields } from './roleCalculatedFields';
import { RoleCreationFields } from './roleCreationFields';
import { RoleCreationFieldsInternal } from './roleCreationFieldsInternal';
import { RoleCreationFieldsInternalAllOf } from './roleCreationFieldsInternalAllOf';
import { RoleCreationFieldsKubernetesPermissions } from './roleCreationFieldsKubernetesPermissions';
import { RoleEnabledDisabledFields } from './roleEnabledDisabledFields';
import { RolePermission } from './rolePermission';
import { RolePermissionSet } from './rolePermissionSet';
import { RoleV1 } from './roleV1';
import { RolesSortFilterFields } from './rolesSortFilterFields';
import { ScopeType } from './scopeType';
import { SubjectType } from './subjectType';
import { SubmissionError } from './submissionError';
import { TenantIsCspManagedSetting } from './tenantIsCspManagedSetting';
import { UpdateAuthorizationSettingByKeyRequest } from './updateAuthorizationSettingByKeyRequest';
import { UpdateScopeAccessRulesRequest } from './updateScopeAccessRulesRequest';

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
        "AccessRuleClusterPhase": AccessRuleClusterPhase,
        "AccessRulePhase": AccessRulePhase,
        "AccessRulesBatchAction": AccessRulesBatchAction,
        "AccessRulesMigrateDeprecatedFailureItem.StepEnum": AccessRulesMigrateDeprecatedFailureItem.StepEnum,
        "AccessRulesSortFilterFields": AccessRulesSortFilterFields,
        "Action": Action,
        "AllowedInternalCreateRoles": AllowedInternalCreateRoles,
        "AuthorizationSettingsKeyEnum": AuthorizationSettingsKeyEnum,
        "GetPermittedScopesRequest.ActionEnum": GetPermittedScopesRequest.ActionEnum,
        "PermissionAction": PermissionAction,
        "ResourceType": ResourceType,
        "ResourceTypeGroupId": ResourceTypeGroupId,
        "Role.ScopeTypeEnum": Role.ScopeTypeEnum,
        "RoleActiveFields.ScopeTypeEnum": RoleActiveFields.ScopeTypeEnum,
        "RoleCreationFields.ScopeTypeEnum": RoleCreationFields.ScopeTypeEnum,
        "RoleCreationFieldsInternal.ScopeTypeEnum": RoleCreationFieldsInternal.ScopeTypeEnum,
        "RoleEnabledDisabledFields.ScopeTypeEnum": RoleEnabledDisabledFields.ScopeTypeEnum,
        "RoleV1.ScopeTypeEnum": RoleV1.ScopeTypeEnum,
        "RolesSortFilterFields": RolesSortFilterFields,
        "ScopeType": ScopeType,
        "SubjectType": SubjectType,
}

let typeMap: {[index: string]: any} = {
    "AccessRule": AccessRule,
    "AccessRuleAllOf": AccessRuleAllOf,
    "AccessRuleById": AccessRuleById,
    "AccessRuleByIdAllOf": AccessRuleByIdAllOf,
    "AccessRuleCreationFields": AccessRuleCreationFields,
    "AccessRuleDeletionBatchResponse": AccessRuleDeletionBatchResponse,
    "AccessRuleDeletionBatchResponseFailureItem": AccessRuleDeletionBatchResponseFailureItem,
    "AccessRuleForSync": AccessRuleForSync,
    "AccessRuleForSyncAllOf": AccessRuleForSyncAllOf,
    "AccessRuleMetadataFields": AccessRuleMetadataFields,
    "AccessRuleScopesInner": AccessRuleScopesInner,
    "AccessRuleStatusRequest": AccessRuleStatusRequest,
    "AccessRuleStatusResponse": AccessRuleStatusResponse,
    "AccessRuleSubjectsInner": AccessRuleSubjectsInner,
    "AccessRuleSync": AccessRuleSync,
    "AccessRuleSyncAllOf": AccessRuleSyncAllOf,
    "AccessRuleSyncAllOfPredefinedRole": AccessRuleSyncAllOfPredefinedRole,
    "AccessRulesBatchCreateFields": AccessRulesBatchCreateFields,
    "AccessRulesBatchCreateFieldsPayload": AccessRulesBatchCreateFieldsPayload,
    "AccessRulesBatchCreateResponse": AccessRulesBatchCreateResponse,
    "AccessRulesBatchCreateResponseFailedInner": AccessRulesBatchCreateResponseFailedInner,
    "AccessRulesBatchFields": AccessRulesBatchFields,
    "AccessRulesMigrateDeprecatedFailureItem": AccessRulesMigrateDeprecatedFailureItem,
    "AccessRulesMigrateDeprecatedFields": AccessRulesMigrateDeprecatedFields,
    "AccessRulesMigrateDeprecatedResponse": AccessRulesMigrateDeprecatedResponse,
    "AccessRulesMigrateDeprecatedResponseSucceeded": AccessRulesMigrateDeprecatedResponseSucceeded,
    "AuthorizationSettingsGetResponse": AuthorizationSettingsGetResponse,
    "BackendEvaluationRequest": BackendEvaluationRequest,
    "BatchResponse": BatchResponse,
    "BatchResponseFailureItem": BatchResponseFailureItem,
    "ComplianceIssues": ComplianceIssues,
    "ComplianceIssuesComplianceIssuesInner": ComplianceIssuesComplianceIssuesInner,
    "CountAccessRules200Response": CountAccessRules200Response,
    "EnableScopedAuthorizationSetting": EnableScopedAuthorizationSetting,
    "EnrichmentDataFields": EnrichmentDataFields,
    "GetAccessRules200Response": GetAccessRules200Response,
    "GetAccessRulesSync200Response": GetAccessRulesSync200Response,
    "GetGroups200Response": GetGroups200Response,
    "GetPermissionSets200Response": GetPermissionSets200Response,
    "GetPermittedScopesRequest": GetPermittedScopesRequest,
    "GetRoles200Response": GetRoles200Response,
    "InternalCreateAccessRuleRequest": InternalCreateAccessRuleRequest,
    "InternalCreateAccessRuleRequestScope": InternalCreateAccessRuleRequestScope,
    "KubernetesPermissions": KubernetesPermissions,
    "MeAccessRule": MeAccessRule,
    "MeAccessRuleAllOf": MeAccessRuleAllOf,
    "MeAccessRulesResponse": MeAccessRulesResponse,
    "MeAccessRulesResponseAccessRules": MeAccessRulesResponseAccessRules,
    "ModelError": ModelError,
    "Permission": Permission,
    "PermissionSet": PermissionSet,
    "PermissionSetAllOf": PermissionSetAllOf,
    "PermittedScopes": PermittedScopes,
    "PermittedScopesActions": PermittedScopesActions,
    "PostPermission": PostPermission,
    "Role": Role,
    "RoleActiveFields": RoleActiveFields,
    "RoleCalculatedFields": RoleCalculatedFields,
    "RoleCreationFields": RoleCreationFields,
    "RoleCreationFieldsInternal": RoleCreationFieldsInternal,
    "RoleCreationFieldsInternalAllOf": RoleCreationFieldsInternalAllOf,
    "RoleCreationFieldsKubernetesPermissions": RoleCreationFieldsKubernetesPermissions,
    "RoleEnabledDisabledFields": RoleEnabledDisabledFields,
    "RolePermission": RolePermission,
    "RolePermissionSet": RolePermissionSet,
    "RoleV1": RoleV1,
    "SubmissionError": SubmissionError,
    "TenantIsCspManagedSetting": TenantIsCspManagedSetting,
    "UpdateAuthorizationSettingByKeyRequest": UpdateAuthorizationSettingByKeyRequest,
    "UpdateScopeAccessRulesRequest": UpdateScopeAccessRulesRequest,
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
