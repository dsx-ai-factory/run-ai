/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    accessKeyUtils,
    auditLogsUtils,
    authorizationUtils,
    authUtils,
    config,
    flowRootState,
    objectGenerator,
    projectResourcesUtils,
    tags,
    TokenObject,
    usersManager,
    Utils,
    waitUtils
} from '../../../../../../modules/helpers';
import { UserPostResponse } from '../../../../../../modules/api/apis/identityManagerApi/generated/api';
import { expect } from 'chai';
import { test } from '@playwright/test';
import { IncomingMessage } from 'http';

test.describe(`Access Keys Audit tags: ${tags.resource.audit} ${tags.teams.phoenix} ${tags.suites.legacySphinx} ${tags.tool.api}`, function () {
    let editorUser: { response: IncomingMessage; body: UserPostResponse };
    let editorUserToken: TokenObject;

    test.beforeEach(async function () {
        editorUser = await usersManager.createLocalUser(
            flowRootState.users.system_administrator.accessToken,
            objectGenerator.generateEmailByPrefix('editoraudit'),
            false
        );
        await authorizationUtils.createAccessRuleForUser(
            flowRootState.users.system_administrator.accessToken,
            editorUser.body.username,
            flowRootState.users.editor.role.id
        );
        editorUserToken = await authUtils.authenticateUser(
            editorUser.body.username,
            editorUser.body.tempPassword
        );
    });

    test.afterEach(async function () {
        await usersManager.deleteLocalUser(
            flowRootState.users.system_administrator.accessToken,
            editorUser.body.id
        );
    });

    test('should make sure the audit data shows access key in metadata', async function () {
        const accessKeyName = objectGenerator.generateShortNameByPrefix('accesskeyaudit');
        await accessKeyUtils.createAccessKey(editorUserToken.accessToken, {
            name: accessKeyName
        });
        const e = new Date(new Date().getTime() + 1000 * 60);
        const s = new Date(new Date().getTime() - 1000 * 60 * 10);
        const filters = new Array<string>();
        filters.push('url==/api/v1/access-keys');
        await waitUtils.waitForCondition(
            async () =>
                (
                    await auditLogsUtils.getAuditLogs(
                        flowRootState.users.system_administrator.accessToken,
                        s,
                        e,
                        filters
                    )
                ).body.auditLogs !== undefined,
            'waiting for audit logs to be available',
            true,
            60,
            5
        );
        const auditLogs = await auditLogsUtils.getAuditLogs(
            flowRootState.users.system_administrator.accessToken,
            s,
            e,
            filters
        );
        expect(auditLogs.body.auditLogs, 'audit logs should contain at least one entry for the access key creation request').to.have.lengthOf.above(0);
    });

    test('should validate audit created by access key', async function () {
        const projectName = objectGenerator.generateNameByPrefix('aceesskeyprojectaudit');
        const projectNamespace = Utils.namespaceFromProjectName(projectName);
        const accessKeyName = objectGenerator.generateShortNameByPrefix('accesskeyaudit');
        const createdAccessKey = await accessKeyUtils.createAccessKey(editorUserToken.accessToken, {
            name: accessKeyName
        });
        const getAccessKey = await accessKeyUtils.getAccessKeyById(
            editorUserToken.accessToken,
            createdAccessKey.body.id
        );

        const accessKeyToken = await authUtils.authenticatePersonalApp(
            getAccessKey.body.clientId,
            createdAccessKey.body.clientSecret,
            config.tenantUrl
        );
        await projectResourcesUtils.createProject(
            projectNamespace,
            {
                name: projectName,
                maxAllowedGpus: 2,
                deservedGpus: 2,
                permissions: { applications: [accessKeyName], groups: [] }
            },
            accessKeyToken.accessToken
        );
        const e = new Date(new Date().getTime() + 1000 * 60);
        const s = new Date(new Date().getTime() - 1000 * 60 * 10);
        const filters = new Array<string>();
        filters.push('url==/api/v1/org-unit/projects');
        filters.push('subject_type==AccessKey');
        filters.push(`subject==${editorUser.body.username}`);
        await waitUtils.waitForCondition(
            async () => {
                const response = await auditLogsUtils.getAuditLogs(
                    flowRootState.users.system_administrator.accessToken,
                    s,
                    e,
                    filters
                );
                return response.body.auditLogs !== undefined && response.body.auditLogs.length > 0;
            },
            'waiting for audit logs to be available',
            true,
            60,
            5
        );
        const auditLogs = await auditLogsUtils.getAuditLogs(
            flowRootState.users.system_administrator.accessToken,
            s,
            e,
            filters
        );
        expect(auditLogs.body.auditLogs, 'audit logs should contain at least one entry for the project creation made via access key').to.have.lengthOf.above(0);
    });
});
