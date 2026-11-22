/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { authUtils, flowRootState, logger, objectGenerator, usersManager, waitUtils } from '../index';
import { TokenObject } from '../models/tokenObject';
import { UserPostResponse } from '../../api/apis/identityManagerApi/generated/api';
import { authorizationApi } from '../../api';
import { expect } from 'chai';
import { StatusCodes } from 'http-status-codes';
import { AccessRule } from '../../api/apis/authorizationApi/generated/model/accessRule';
import { GetAccessRules200Response } from '../../api/apis/authorizationApi/generated/model/getAccessRules200Response';
import { SubjectType } from '../../api/apis/authorizationApi/generated/model/subjectType';
import { ScopeType } from '../../api/apis/authorizationApi/generated/model/scopeType';
import { Role } from '../../api/apis/authorizationApi/generated/model/role';
import { Permission } from '../../api/apis/authorizationApi/generated/model/permission';
import { Action } from '../../api/apis/authorizationApi/generated/model/action';
import { RoleCreationFields } from '../../api/apis/authorizationApi/generated/model/roleCreationFields';
import { KubernetesPermissions } from '../../api/apis/authorizationApi/generated/model/kubernetesPermissions';
import { RoleActiveFields } from '../../api/apis/authorizationApi/generated/model/roleActiveFields';
import { TenantApi } from '../../api/apis/backendApi/generated/api/tenantApi';
import { EFeatureFlagsKeys } from '../models/featureFlags';
import { IncomingMessage } from 'http';
import { AccessRulesBatchCreateFields } from '../../api/apis/authorizationApi/generated/model/accessRulesBatchCreateFields';
import { AccessRulesBatchCreateResponse } from '../../api/apis/authorizationApi/generated/model/accessRulesBatchCreateResponse';
import { RoleCreationFieldsInternal } from '../../api/apis/authorizationApi/generated/model/roleCreationFieldsInternal';
import { RolePermissionSet } from '../../api/apis/authorizationApi/generated/model/rolePermissionSet';
import { ResourceType } from '../../api/apis/authorizationApi/generated/model/resourceType';

export const aiPractitionerRoleName = 'AI practitioner';
export const projectManagerRoleName = 'Project administrator';

export class AuthorizationUtils {
    async getRoles(
        token: string = flowRootState.users.system_administrator.accessToken,
        rolesApi = authorizationApi.rolesApi,
        includeInternal?: boolean
    ): Promise<Role[]> {
        authorizationApi.rolesApi.accessToken = token;
        const pageSize = 500;
        const roles: Role[] = [];
        let offset: number | undefined = 0;
        while (offset !== undefined) {
            const response = await rolesApi.getRoles(
                pageSize,
                offset,
                null,
                null,
                null,
                null,
                includeInternal
            );
            roles.push(...response.body.roles);
            offset = response.body.next ?? undefined;
        }
        return roles;
    }

    async getRolesV2(
        token: string = flowRootState.users.system_administrator.accessToken,
        rolesApi = authorizationApi.rolesApi,
        filter: Array<string> = null,
        includeInternal?: boolean
    ): Promise<Role[]> {
        authorizationApi.rolesApi.accessToken = token;
        const response = await rolesApi.getRoles(
            null,
            null,
            null,
            null,
            filter,
            null,
            includeInternal
        );
        return response?.body.roles;
    }

    async getRoleV2(
        roleId: number,
        token: string = flowRootState.users.system_administrator.accessToken,
        rolesApi = authorizationApi.rolesApi
    ): Promise<Role> {
        authorizationApi.rolesApi.accessToken = token;
        const response = await rolesApi.getRole(roleId);
        return response?.body;
    }

    async getAccessRules(
        filters: any = {},
        token: string = flowRootState.users.system_administrator.accessToken
    ): Promise<GetAccessRules200Response> {
        authorizationApi.accessRulesApi.accessToken = token;
        // Join filterBy into a single comma-separated string because the Go controller
        // parses filterBy with query.Get() + strings.Split(",") — repeated query params are ignored.
        const filterBy = filters.filterBy?.length ? [filters.filterBy.join(',')] : filters.filterBy;
        const response = await authorizationApi.accessRulesApi.getAccessRules(
            filters.subjectType,
            filters.subjectIds,
            filters.limit,
            filters.offset,
            filters.lastUpdated,
            filters.includeDeleted,
            filters.clusterId,
            filters.scopeId,
            filters.sortOrder,
            filters.sortBy,
            filterBy,
            filters.search
        );
        expect(response.response.statusCode).equal(StatusCodes.OK);

        return response?.body;
    }

    async deleteAccessRule(
        accessRule: AccessRule,
        token: string = flowRootState.users.system_administrator.accessToken,
        accessRulesApi = authorizationApi.accessRulesApi
    ): Promise<void> {
        for (const user of Object.values(flowRootState.users)) {
            if (user.accessRules?.some((infraAccessRules) => infraAccessRules.id === accessRule.id)) {
                return;
            }
        }

        accessRulesApi.accessToken = token;
        const response = await accessRulesApi.deleteAccessRule(accessRule.id);
        expect(response.response.statusCode).equal(StatusCodes.NO_CONTENT);
    }

    async waitForAccessRule(
        accessRuleId: number,
        shouldExist: boolean,
        token: string = flowRootState.users.system_administrator.accessToken,
        timeout = 60
    ): Promise<void> {
        authorizationApi.accessRulesApi.accessToken = token;

        const doesAccessRuleExist = async (): Promise<boolean> => {
            try {
                const res = (await authorizationApi.accessRulesApi.getAccessRule(accessRuleId)).body;
                return !!res;
            } catch {
                return false;
            }
        };

        await waitUtils.waitForCondition(
            doesAccessRuleExist,
            `waitForAccessRule ${accessRuleId} to ${shouldExist ? 'exist' : 'not exist'}`,
            shouldExist,
            timeout
        );
    }

    async createAccessRuleWithDefaults(
        subjectId: string,
        subjectType: SubjectType,
        roleId: number,
        scopeId?: string,
        scopeType?: ScopeType,
        token: string = flowRootState.users.system_administrator.accessToken
    ): Promise<AccessRule> {
        authorizationApi.accessRulesApi.accessToken = token;
        const response = await authorizationApi.accessRulesApi.createAccessRule({
            subjectId,
            subjectType,
            roleId,
            scopeId: scopeId ?? flowRootState.cluster.tenantId.toString(),
            scopeType: scopeType ?? ScopeType.Tenant
        });
        expect(response.response.statusCode).equal(StatusCodes.CREATED);

        return response.body;
    }

    async createAccessRulesForUsers(
        userNameToRoleId: Map<string, number>,
        scopeId: string,
        scopeType: ScopeType,
        subjectType: SubjectType
    ): Promise<AccessRule[]> {
        logger.info(
            `Creating accessrules in scope ${scopeType}/${scopeId} for: userNameToRoleId: ${JSON.stringify(
                Object.fromEntries(userNameToRoleId)
            )}`
        );
        const promises = [];
        for (const [userName, roleId] of Array.from(userNameToRoleId.entries())) {
            promises.push(
                authorizationApi.accessRulesApi.createAccessRule({
                    roleId: roleId,
                    scopeId: scopeId,
                    scopeType: scopeType,
                    subjectId: userName,
                    subjectType: subjectType
                })
            );
        }
        const accessRulesResponses = await Promise.all(promises);
        return accessRulesResponses.map((accessRulesResponse) => accessRulesResponse.body);
    }

    async createAccessRuleForUser(
        accessToken: string,
        userName: string,
        roleId: number,
        subjectType?: SubjectType,
        scopeId = flowRootState.cluster.tenantId.toString(),
        scopeType = ScopeType.Tenant,
        accessRulesApi = authorizationApi.accessRulesApi
    ): Promise<AccessRule> {
        accessRulesApi.accessToken = accessToken;
        const response = await accessRulesApi.createAccessRule({
            roleId: roleId,
            scopeId: scopeId,
            scopeType: scopeType,
            subjectId: userName,
            subjectType: subjectType ?? SubjectType.User
        });
        expect(response.response.statusCode).equal(StatusCodes.CREATED);

        return response.body;
    }

    async createAccessRuleForGroup(
        accessToken: string,
        groupName: string,
        roleId: number,
        scopeType: ScopeType = ScopeType.Tenant,
        scopeId = flowRootState.cluster.tenantId.toString()
    ): Promise<AccessRule> {
        authorizationApi.accessRulesApi.accessToken = accessToken;
        const response = await authorizationApi.accessRulesApi.createAccessRule({
            roleId: roleId,
            scopeId: scopeId,
            scopeType: scopeType,
            subjectId: groupName,
            subjectType: SubjectType.Group
        });
        expect(response.response.statusCode).equal(StatusCodes.CREATED);

        return response.body;
    }

    async createAccessRuleForApp(
        appName: string,
        roleName: string,
        scopeType: ScopeType,
        scopeId: string,
        accessToken: string = flowRootState.users.system_administrator.accessToken
    ): Promise<AccessRule> {
        authorizationApi.accessRulesApi.accessToken = accessToken;
        const roles = await this.getRoles(accessToken);
        const roleId = roles.find((role) => role.name === roleName).id;
        const response = await authorizationApi.accessRulesApi.createAccessRule({
            roleId: roleId,
            scopeId: scopeId,
            scopeType: scopeType,
            subjectId: appName,
            subjectType: SubjectType.App
        });
        expect(response.response.statusCode).equal(StatusCodes.CREATED);

        return response.body;
    }

    async createAccessRuleForServiceAccount(
        serviceAccountClientId: string,
        roleName: string,
        scopeType: ScopeType,
        scopeId: string,
        accessToken: string = flowRootState.users.system_administrator.accessToken
    ): Promise<AccessRule> {
        authorizationApi.accessRulesApi.accessToken = accessToken;
        const roles = await this.getRoles(accessToken);
        const roleId = roles.find((role) => role.name === roleName).id;
        const response = await authorizationApi.accessRulesApi.createAccessRule({
            roleId: roleId,
            scopeId: scopeId,
            scopeType: scopeType,
            subjectId: serviceAccountClientId,
            // 'service-account' is supported by the backend; cast to bypass outdated client enum
            subjectType: 'service-account' as any
        });
        expect(response.response.statusCode).equal(StatusCodes.CREATED);

        return response.body;
    }

    async createAccessRuleForUserByRoleName(
        userName: string,
        roleName: string,
        scopeType: ScopeType,
        scopeId: string,
        accessToken: string = flowRootState.users.system_administrator.accessToken
    ): Promise<AccessRule> {
        authorizationApi.accessRulesApi.accessToken = accessToken;
        const roles = await this.getRoles(accessToken);
        const roleId = roles.find((role) => role.name === roleName).id;

        const response = await authorizationApi.accessRulesApi.createAccessRule({
            roleId: roleId,
            scopeId: scopeId,
            scopeType: scopeType,
            subjectId: userName,
            subjectType: SubjectType.User
        });
        expect(response.response.statusCode).equal(StatusCodes.CREATED);

        return response.body;
    }

    async bindUserWithRoles(
        user: UserPostResponse,
        roleNames: string[],
        options: { scopeId: string; scopeType: ScopeType },
        accessToken: string = flowRootState.users.system_administrator.accessToken
    ): Promise<void> {
        const roles = await this.getRoles(accessToken);
        for (const roleName of roleNames) {
            const role = roles.find((r) => r.name === roleName);
            if (!role) throw new Error(`role ${roleName} not defined!`);

            await this.createAccessRuleWithDefaults(
                user.username,
                SubjectType.User,
                role.id,
                options.scopeId,
                options.scopeType,
                accessToken
            );
        }
    }

    async getAccessRuleById(
        accessRuleId: number,
        accessToken: string = flowRootState.users.system_administrator.accessToken
    ): Promise<AccessRule> {
        authorizationApi.accessRulesApi.accessToken = accessToken;
        const response = await authorizationApi.accessRulesApi.getAccessRule(accessRuleId);
        return response.body;
    }

    async getUserPermissions(userAccessToken: string): Promise<Permission[]> {
        authorizationApi.permissionsApi.accessToken = userAccessToken;
        const response = await authorizationApi.permissionsApi.getPermissions();
        expect(response.response.statusCode).equal(StatusCodes.OK);

        return response.body;
    }

    getActionsForResource(permissions: Permission[], resource: ResourceType): string[] {
        return (
            (permissions.find((permission) => permission.resourceType === resource)?.actions ??
                []) as unknown[]
        ).map(String);
    }

    async createUserWithCustomRole(opts: {
        adminToken: string;
        rolePrefix: string;
        description: string;
        permissions: Permission[];
        tenantId: string;
        logLabel: string;
    }): Promise<{ user: UserPostResponse; token: TokenObject; roleId: number }> {
        const role = await this.createCustomRole(
            objectGenerator.generateShortNameByPrefix(opts.rolePrefix),
            opts.description,
            opts.permissions,
            undefined,
            opts.adminToken,
            RoleCreationFields.ScopeTypeEnum.Tenant,
            opts.tenantId,
            true
        );
        const userResponse = await usersManager.createLocalUser(
            opts.adminToken,
            objectGenerator.generateEmailByPrefixWithRandom(role.name, 5)
        );
        const user = userResponse.body;
        await this.createAccessRuleForUser(opts.adminToken, user.username, role.id, undefined, opts.tenantId);
        const token = await authUtils.authenticateUser(user.username, user.tempPassword);
        logger.info(`${opts.logLabel} created: username=${user.username}, roleId=${role.id}`);
        return { user, token, roleId: role.id };
    }

    buildRolePermissions(
        availablePermissions: Permission[],
        readResources: Set<ResourceType>,
        writeResources: Set<ResourceType>
    ): Permission[] {
        return availablePermissions.flatMap((permission): Permission[] => {
            if (!permission.actions?.length) return [];
            if (writeResources.has(permission.resourceType)) {
                return [{ ...permission, actions: permission.actions } as Permission];
            }
            if (readResources.has(permission.resourceType) && permission.actions.includes(Action.Read)) {
                return [{ ...permission, actions: [Action.Read] } as Permission];
            }
            return [];
        });
    }

    async getErrorStatusCode(call: () => Promise<unknown>): Promise<number | undefined> {
        try {
            await call();
            return undefined;
        } catch (error) {
            return error?.statusCode ?? error?.response?.statusCode ?? error?.response?.status;
        }
    }

    async getAccessRulesByUserName(
        userName: string,
        accessToken: string = flowRootState.users.system_administrator.accessToken
    ): Promise<AccessRule[]> {
        authorizationApi.accessRulesApi.accessToken = accessToken;
        // get all access rules and return the ones with the specified userName
        const response = await this.getAccessRules(
            {
                subjectType: SubjectType.User,
                filterBy: [`subjectIdFilter=@${userName}`]
            },
            accessToken
        );
        return response.accessRules;
    }

    async createCustomRole(
        roleName: string,
        description: string,
        permissions: Permission[],
        kubernetesPermissions: KubernetesPermissions = { predefinedRole: null },
        accessToken: string = flowRootState.users.system_administrator.accessToken,
        scopeType?: RoleCreationFields.ScopeTypeEnum,
        scopeId?: string,
        enabled?: boolean,
        rolesApi = authorizationApi.rolesApi
    ): Promise<Role> {
        rolesApi.accessToken = accessToken;

        const response = await rolesApi.createRole({
            name: roleName,
            description: description,
            permissions: permissions as any,
            kubernetesPermissions: kubernetesPermissions,
            scopeType: scopeType,
            scopeId: scopeId,
            enabled: enabled
        });
        expect(response.response.statusCode).equal(StatusCodes.CREATED);

        return response.body;
    }

    async createRoleWithPermissionSets(
        roleName: string,
        description: string,
        permissionSetIds: string[],
        scopeType: RoleCreationFieldsInternal.ScopeTypeEnum = RoleCreationFieldsInternal.ScopeTypeEnum.Tenant,
        scopeId: string = null,
        enabled: boolean = true,
        accessToken: string = flowRootState.users.system_administrator.accessToken,
        rolesApi = authorizationApi.rolesApi
    ): Promise<Role> {
        rolesApi.accessToken = accessToken;

        const finalScopeId =
            scopeType === RoleCreationFieldsInternal.ScopeTypeEnum.Tenant
                ? scopeId || flowRootState.cluster.tenantId.toString()
                : scopeType === RoleCreationFieldsInternal.ScopeTypeEnum.System
                  ? scopeId || flowRootState.cluster.tenantId.toString()
                  : null;

        const roleCreationFields: RoleCreationFieldsInternal = {
            name: roleName,
            description,
            scopeType,
            scopeId: finalScopeId,
            enabled,
            permissionSets: permissionSetIds.map((id) => ({ id }) as RolePermissionSet)
        };

        const response = await rolesApi.createRole(roleCreationFields);
        expect(response.response.statusCode).equal(StatusCodes.CREATED);

        return response.body;
    }

    async updateRole(
        role: Role,
        roleCreationFields: RoleCreationFields,
        accessToken: string = flowRootState.users.system_administrator.accessToken,
        rolesApi = authorizationApi.rolesApi
    ): Promise<Role> {
        rolesApi.accessToken = accessToken;
        const response = await rolesApi.updateRole(role.id, roleCreationFields);
        expect(response.response.statusCode).equal(StatusCodes.OK);

        return response.body;
    }

    async deleteRole(
        roleId: number,
        accessToken: string = flowRootState.users.system_administrator.accessToken,
        rolesApi = authorizationApi.rolesApi
    ): Promise<void> {
        rolesApi.accessToken = accessToken;
        const response = await rolesApi.deleteRole(roleId);
        expect(response.response.statusCode).equal(StatusCodes.NO_CONTENT);
    }

    async updateRolesK8sPermissions(
        costumeRole: Role,
        k8sPermissions: KubernetesPermissions,
        accessToken: string = flowRootState.users.system_administrator.accessToken
    ): Promise<Role> {
        authorizationApi.rolesApi.accessToken = accessToken;
        const roleCreationFields = new RoleCreationFields();

        const roleId = costumeRole.id;
        roleCreationFields.name = costumeRole.name;
        roleCreationFields.description = costumeRole.description;
        roleCreationFields.scopeType = costumeRole.scopeType;
        roleCreationFields.scopeId = costumeRole.scopeId;
        roleCreationFields.enabled = costumeRole.enabled;
        // for now, we only have tests with permissions. when role will be with permission sets this step can be changes to take permission sets OR permissions from the original role
        roleCreationFields['permissions'] = costumeRole.permissions;

        roleCreationFields.kubernetesPermissions = k8sPermissions;

        const response = await authorizationApi.rolesApi.updateRole(roleId, roleCreationFields);
        expect(response.response.statusCode).equal(StatusCodes.OK);
        return response.body;
    }

    async activateRole(
        role: Role,
        roleActiveFields: RoleActiveFields,
        accessToken: string = flowRootState.users.system_administrator.accessToken,
        rolesApi = authorizationApi.rolesApi
    ): Promise<void> {
        rolesApi.accessToken = accessToken;
        const response = await rolesApi.roleActive(role.id, roleActiveFields);
        expect(response.response.statusCode).equal(StatusCodes.NO_CONTENT);
    }

    async bindAccessRuleToLocalUser(
        accessRulesBatchCreateFields: AccessRulesBatchCreateFields,
        accessToken?: string
    ): Promise<{
        response: IncomingMessage;
        body: AccessRulesBatchCreateResponse;
    }> {
        authorizationApi.accessRulesApi.accessToken = accessToken
            ? accessToken
            : flowRootState.users.system_administrator.accessToken;
        return await authorizationApi.accessRulesApi.accessRulesBatchCreate(accessRulesBatchCreateFields);
    }

    static verifyStatusCode(response: any, expectedCode: number): void {
        expect(response.response.statusCode).equal(expectedCode);
    }

    static extractErrorMessage(errorBody: any): string {
        if (typeof errorBody === 'string') {
            return errorBody.toLowerCase();
        }
        return (
            errorBody?.message ||
            errorBody?.error ||
            errorBody?.details ||
            errorBody?.errorMessage ||
            ''
        ).toLowerCase();
    }
}
