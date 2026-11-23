/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import { test } from '@playwright/test';
import { StatusCodes } from 'http-status-codes';
import { IncomingMessage } from 'http';
import {
    authorizationUtils,
    flowRootState,
    objectGenerator,
    serviceAccountUtils,
    tags
} from '../../../../../../modules/helpers';
import { ScopeType } from '../../../../../../modules/api/apis/authorizationApi/generated/model/scopeType';
import { PROPERTY_TO_ROLE_MAPPING } from '../../../../../../modules/helpers/config/roleMapping';
import { ServiceAccountPostResponse } from '../../../../../../modules/api/apis/identityManagerApi/generated/api';
import { AccessRule } from '../../../../../../modules/api/apis/authorizationApi/generated/model/accessRule';
import { cleanupAccessRuleAndServiceAccount } from '../../../../../../modules/helpers/utils/ServiceAccountAccessRulesUtils';

test.describe(
    'Service Account Access Rules - basic API test',
    {
        tag: [
            tags.teams.phoenix, tags.suites.legacySphinx,
            tags.teams.viper,
            tags.tool.api,
            tags.resource.accessRules,
            tags.runOn.parallelCI
        ]
    },
    () => {
        let serviceAccountName: string;
        let serviceAccount: { response: IncomingMessage; body: ServiceAccountPostResponse } | undefined;
        let createdAccessRule: AccessRule | undefined;

        test.beforeEach(async () => {
            createdAccessRule = undefined;
            serviceAccountName = objectGenerator.generateShortNameByPrefix(`sa-accessrule-${Date.now()}`);
            serviceAccount = await serviceAccountUtils.createServiceAccount(
                flowRootState.users.system_administrator.accessToken,
                { name: serviceAccountName }
            );
            expect(
                serviceAccount.response.statusCode,
                'service account creation in beforeEach should return 201 Created'
            ).to.equal(StatusCodes.CREATED);
            expect(
                serviceAccount.body.clientId,
                'service account should have a non-empty clientId after creation'
            ).to.not.be.empty;
        });

        test.afterEach(async () => {
            await cleanupAccessRuleAndServiceAccount(createdAccessRule, serviceAccount);
        });

        test('should bind access rule to service account using subjectType=service-account', async () => {
            const clientId = serviceAccount.body.clientId as string;

            createdAccessRule = await authorizationUtils.createAccessRuleForServiceAccount(
                clientId,
                PROPERTY_TO_ROLE_MAPPING.system_administrator,
                ScopeType.Cluster,
                flowRootState.cluster.uuid,
                flowRootState.users.system_administrator.accessToken
            );

            expect(createdAccessRule, 'createAccessRuleForServiceAccount should return a defined access rule')
                .to.not.be.undefined;
            expect(
                createdAccessRule.subjectId,
                'access rule subjectId should match the service account clientId'
            ).to.equal(clientId);
            expect(
                String((createdAccessRule as any).subjectType),
                'access rule subjectType should be service-account'
            ).to.equal('service-account');

            await authorizationUtils.waitForAccessRule(createdAccessRule.id, true);

            const result = await authorizationUtils.getAccessRules({
                filterBy: [`subjectIdFilter=@${clientId}`, 'subjectType==service-account'],
                limit: 10
            });

            expect(
                result.accessRules.some((ar) => ar.id === createdAccessRule?.id),
                'created access rule should appear when filtering by subjectType==service-account'
            ).to.be.true;
        });

        test('should filter access rules by subjectType service-account', async () => {
            const clientId = serviceAccount.body.clientId as string;

            createdAccessRule = await authorizationUtils.createAccessRuleForServiceAccount(
                clientId,
                PROPERTY_TO_ROLE_MAPPING.system_administrator,
                ScopeType.Cluster,
                flowRootState.cluster.uuid,
                flowRootState.users.system_administrator.accessToken
            );
            await authorizationUtils.waitForAccessRule(createdAccessRule.id, true);

            const bySa = await authorizationUtils.getAccessRules({
                filterBy: [`subjectIdFilter=@${clientId}`, 'subjectType==service-account'],
                limit: 10
            });
            expect(
                bySa.accessRules.map((r) => r.id),
                'filtering by subjectType==service-account should include the service-account rule'
            ).to.include(createdAccessRule.id);

            const notSa = await authorizationUtils.getAccessRules({
                filterBy: [`subjectIdFilter=@${clientId}`, 'subjectType!=service-account'],
                limit: 10
            });
            expect(
                notSa.accessRules.map((r) => r.id),
                'filtering by subjectType!=service-account should exclude the service-account rule'
            ).to.not.include(createdAccessRule.id);
        });

        test('should bind access rule to service account at tenant scope', async () => {
            const clientId = serviceAccount.body.clientId as string;

            createdAccessRule = await authorizationUtils.createAccessRuleForServiceAccount(
                clientId,
                PROPERTY_TO_ROLE_MAPPING.system_administrator,
                ScopeType.Tenant,
                flowRootState.cluster.tenantId.toString(),
                flowRootState.users.system_administrator.accessToken
            );
            expect(
                createdAccessRule.scopeType,
                'access rule created at tenant scope should have scopeType Tenant'
            ).to.equal(ScopeType.Tenant);
            await authorizationUtils.waitForAccessRule(createdAccessRule.id, true);

            const list = await authorizationUtils.getAccessRules({
                filterBy: [`subjectIdFilter=@${clientId}`, `scopeType==${ScopeType.Tenant}`],
                limit: 10
            });
            expect(
                list.accessRules.some((r) => r.id === createdAccessRule?.id),
                'tenant-scoped access rule should appear when filtering by tenant scope'
            ).to.be.true;
        });

        test('should delete service-account access rule and verify removal', async () => {
            const clientId = serviceAccount.body.clientId as string;
            const rule = await authorizationUtils.createAccessRuleForServiceAccount(
                clientId,
                PROPERTY_TO_ROLE_MAPPING.system_administrator,
                ScopeType.Cluster,
                flowRootState.cluster.uuid,
                flowRootState.users.system_administrator.accessToken
            );
            await authorizationUtils.waitForAccessRule(rule.id, true);

            await authorizationUtils.deleteAccessRule(rule);
            await authorizationUtils.waitForAccessRule(rule.id, false);

            const list = await authorizationUtils.getAccessRules({
                filterBy: [`subjectIdFilter=@${clientId}`],
                limit: 10
            });
            expect(
                list.accessRules.some((r) => r.id === rule.id),
                'deleted access rule should not appear in the list'
            ).to.be.false;
        });
    }
);
