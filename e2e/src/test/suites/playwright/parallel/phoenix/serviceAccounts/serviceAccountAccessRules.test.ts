/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import {
    authorizationUtils,
    flowRootState,
    logger,
    objectGenerator,
    serviceAccountUtils,
    tags
} from '../../../../../../modules/helpers';
import { test } from '@playwright/test';
import { StatusCodes } from 'http-status-codes';
import { ScopeType } from '../../../../../../modules/api/apis/authorizationApi/generated/model/scopeType';
import { PROPERTY_TO_ROLE_MAPPING } from '../../../../../../modules/helpers/config/roleMapping';
import { ServiceAccountPostResponse } from '../../../../../../modules/api/apis/identityManagerApi/generated/api';
import { AccessRule } from '../../../../../../modules/api/apis/authorizationApi/generated/model/accessRule';
import { IncomingMessage } from 'http';
import { authorizationApi } from '../../../../../../modules/api';
import { SubjectType } from '../../../../../../modules/api/apis/authorizationApi/generated/model/subjectType';
import { supportedFrom, SupportedVersions } from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';

test.describe(`Service Account Access Rules - basic API test, tags: ${tags.teams.phoenix} ${tags.suites.legacySphinx} ${tags.teams.viper} ${tags.tool.api} ${tags.resource.accessRules}`, function () {
    let serviceAccountName: string;
    let serviceAccount: { response: IncomingMessage; body: ServiceAccountPostResponse } | undefined;
    let createdAccessRule: AccessRule | undefined;

    test.beforeEach(async function () {
        serviceAccountName = objectGenerator.generateShortNameByPrefix(`sa-accessrule-${Date.now()}`);
        serviceAccount = await serviceAccountUtils.createServiceAccount(
            flowRootState.users.system_administrator.accessToken,
            { name: serviceAccountName }
        );
        expect(serviceAccount.response.statusCode, 'service account creation in beforeEach should return 201 Created').to.equal(StatusCodes.CREATED);
        expect(serviceAccount.body.clientId, 'service account should have a non-empty clientId after creation').to.not.be.empty;
    });

    test.afterEach(async function () {
        // Cleanup access rule if created
        if (createdAccessRule) {
            try {
                await authorizationUtils.deleteAccessRule(createdAccessRule);
            } catch (e) {
                logger.error(`Failed to delete access rule id=${createdAccessRule.id}: ${JSON.stringify(e)}`);
            }
        }

        // Cleanup service account
        if (serviceAccount) {
            try {
                await serviceAccountUtils.deleteServiceAccount(
                    flowRootState.users.system_administrator.accessToken,
                    serviceAccount.body.id
                );
            } catch (e) {
                logger.error(
                    `Failed to delete Service Account with id: ${serviceAccount.body.id} ${JSON.stringify(e)}`
                );
            }
        }
    });

    supportedFrom(SupportedVersions.V2_24)(function () {
        test('should bind access rule to service account using subjectType=service-account', async function () {
            const clientId = serviceAccount.body.clientId as string;

            createdAccessRule = await authorizationUtils.createAccessRuleForServiceAccount(
                clientId,
                PROPERTY_TO_ROLE_MAPPING.system_administrator,
                ScopeType.Cluster,
                flowRootState.cluster.uuid,
                flowRootState.users.system_administrator.accessToken
            );

            expect(createdAccessRule, 'createAccessRuleForServiceAccount should return a defined access rule').to.not.be.undefined;
            expect(createdAccessRule.subjectId, 'access rule subjectId should match the service account clientId').to.equal(clientId);
            expect(String((createdAccessRule as any).subjectType), 'access rule subjectType should be service-account').to.equal('service-account');

            await authorizationUtils.waitForAccessRule(createdAccessRule.id, true);

            const result = await authorizationUtils.getAccessRules({
                filterBy: [`subjectIdFilter=@${clientId}`, 'subjectType==service-account'],
                limit: 10
            });

            expect(result.accessRules.some((ar) => ar.id === createdAccessRule?.id), 'created access rule should appear when filtering by subjectType==service-account').to.be.true;
        });
    });

    test('should support backward-compatible filtering for app and service-account', async function () {
        const clientId = serviceAccount.body.clientId as string;
        let createdRuleSa: AccessRule | undefined;
        let createdRuleApp: AccessRule | undefined;
        try {
            // Create service-account subjectType rule
            createdRuleSa = await authorizationUtils.createAccessRuleForServiceAccount(
                clientId,
                PROPERTY_TO_ROLE_MAPPING.system_administrator,
                ScopeType.Cluster,
                flowRootState.cluster.uuid,
                flowRootState.users.system_administrator.accessToken
            );
            await authorizationUtils.waitForAccessRule(createdRuleSa.id, true);

            // Create legacy app subjectType rule for the same clientId
            authorizationApi.accessRulesApi.accessToken =
                flowRootState.users.system_administrator.accessToken;
            const legacyCreate = await authorizationApi.accessRulesApi.createAccessRule({
                roleId: flowRootState.users.system_administrator.role.id,
                scopeId: flowRootState.cluster.uuid,
                scopeType: ScopeType.Cluster,
                subjectId: clientId,
                subjectType: SubjectType.App
            });
            createdRuleApp = legacyCreate.body;
            expect(legacyCreate.response.statusCode, 'legacy app subjectType access rule creation should return 201 Created').to.equal(StatusCodes.CREATED);
            await authorizationUtils.waitForAccessRule(createdRuleApp.id, true);

            // Filter by subjectType==app should return both
            const byApp = await authorizationUtils.getAccessRules({
                filterBy: [`subjectIdFilter=@${clientId}`, 'subjectType==app'],
                limit: 10
            });
            const byAppIds = byApp.accessRules.map((r) => r.id);
            expect(byAppIds, 'filtering by subjectType==app should include the service-account rule').to.include(createdRuleSa.id);
            expect(byAppIds, 'filtering by subjectType==app should include the legacy app rule').to.include(createdRuleApp.id);

            // Filter by subjectType==service-account should also return both
            const bySa = await authorizationUtils.getAccessRules({
                filterBy: [`subjectIdFilter=@${clientId}`, 'subjectType==service-account'],
                limit: 10
            });
            const bySaIds = bySa.accessRules.map((r) => r.id);
            expect(bySaIds, 'filtering by subjectType==service-account should include the service-account rule').to.include(createdRuleSa.id);
            expect(bySaIds, 'filtering by subjectType==service-account should include the legacy app rule').to.include(createdRuleApp.id);

            // Filter by subjectType!=app should exclude both
            const notApp = await authorizationUtils.getAccessRules({
                filterBy: [`subjectIdFilter=@${clientId}`, 'subjectType!=app'],
                limit: 10
            });
            const notAppIds = notApp.accessRules.map((r) => r.id);
            expect(notAppIds, 'filtering by subjectType!=app should exclude the service-account rule').to.not.include(createdRuleSa.id);
            expect(notAppIds, 'filtering by subjectType!=app should exclude the legacy app rule').to.not.include(createdRuleApp.id);
        } finally {
            // Cleanup
            if (createdRuleSa) {
                try {
                    await authorizationUtils.deleteAccessRule(createdRuleSa);
                } catch (e) {
                    logger.warn(
                        `Ignoring delete error for service-account rule id=${createdRuleSa.id}: ${JSON.stringify(e)}`
                    );
                }
            }
            if (createdRuleApp) {
                try {
                    await authorizationUtils.deleteAccessRule(createdRuleApp);
                } catch (e) {
                    logger.warn(
                        `Ignoring delete error for legacy app rule id=${createdRuleApp.id}: ${JSON.stringify(e)}`
                    );
                }
            }
        }
    });

    test('should bind access rule to service account at tenant scope', async function () {
        const clientId = serviceAccount.body.clientId as string;
        let rule: AccessRule | undefined;
        try {
            rule = await authorizationUtils.createAccessRuleForServiceAccount(
                clientId,
                PROPERTY_TO_ROLE_MAPPING.system_administrator,
                ScopeType.Tenant,
                flowRootState.cluster.tenantId.toString(),
                flowRootState.users.system_administrator.accessToken
            );
            expect(rule.scopeType, 'access rule created at tenant scope should have scopeType Tenant').to.equal(ScopeType.Tenant);
            await authorizationUtils.waitForAccessRule(rule.id, true);

            const list = await authorizationUtils.getAccessRules({
                filterBy: [`subjectIdFilter=@${clientId}`, `scopeType==${ScopeType.Tenant}`],
                limit: 10
            });
            expect(list.accessRules.some((r) => r.id === rule?.id), 'tenant-scoped access rule should appear when filtering by tenant scope').to.be.true;
        } finally {
            if (rule) {
                await authorizationUtils.deleteAccessRule(rule);
            }
        }
    });

    test('should delete service-account access rule and verify removal', async function () {
        const clientId = serviceAccount.body.clientId as string;
        // Create
        const rule = await authorizationUtils.createAccessRuleForServiceAccount(
            clientId,
            PROPERTY_TO_ROLE_MAPPING.system_administrator,
            ScopeType.Cluster,
            flowRootState.cluster.uuid,
            flowRootState.users.system_administrator.accessToken
        );
        await authorizationUtils.waitForAccessRule(rule.id, true);

        // Delete
        await authorizationUtils.deleteAccessRule(rule);
        await authorizationUtils.waitForAccessRule(rule.id, false);

        // Verify not listed
        const list = await authorizationUtils.getAccessRules({
            filterBy: [`subjectIdFilter=@${clientId}`],
            limit: 10
        });
        expect(list.accessRules.some((r) => r.id === rule.id), 'deleted access rule should not appear in the list').to.be.false;
    });
});
