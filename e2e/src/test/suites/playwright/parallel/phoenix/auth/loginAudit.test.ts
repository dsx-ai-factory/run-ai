/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import {
    accessKeyUtils,
    auditLogsUtils,
    authorizationUtils,
    authUtils,
    flowRootState,
    objectGenerator,
    tags,
    usersManager
} from '../../../../../../modules/helpers';
import { UserPostResponse } from '../../../../../../modules/api/apis/identityManagerApi/generated/api';
import { AuditLog } from '../../../../../../modules/api/apis/auditApi';
import { IncomingMessage } from 'http';

test.describe(
    `Login audit tags: ${tags.teams.phoenix} ${tags.suites.legacySphinx} ${tags.tool.api}`,
    { tag: [tags.services.auditService, tags.runOn.parallelCI] },
    () => {
        test('should audit a password (Login API) authentication', async () => {
            const user: { response: IncomingMessage; body: UserPostResponse } =
                await usersManager.createLocalUser(
                    flowRootState.users.system_administrator.accessToken,
                    objectGenerator.generateEmailByPrefix('loginauditpwd'),
                    false
                );
            await authorizationUtils.createAccessRuleForUser(
                flowRootState.users.system_administrator.accessToken,
                user.body.username,
                flowRootState.users.system_administrator.role.id
            );

            try {
                await authUtils.authenticateTokenApi(user.body.username, user.body.tempPassword);

                const auditLog = await auditLogsUtils.findLoginAudit(
                    flowRootState.users.system_administrator.accessToken,
                    user.body.username,
                    'password'
                );
                expect(auditLog.subjectType, 'password login audit should record subjectType as User').toBe(AuditLog.SubjectTypeEnum.User);
                expect(auditLog.subject, 'password login audit subject should match the authenticated username').toBe(user.body.username);
            } finally {
                await usersManager.deleteLocalUser(
                    flowRootState.users.system_administrator.accessToken,
                    user.body.id
                );
            }
        });

        test('should audit a client_credentials (Login Application) authentication', async () => {
            const adminToken = flowRootState.users.system_administrator.accessToken;
            const accessKey = await accessKeyUtils.createAccessKey(adminToken, {
                name: objectGenerator.generateShortNameByPrefix('loginauditapp')
            });
            const clientId = (await accessKeyUtils.getAccessKeyById(adminToken, accessKey.body.id)).body
                .clientId;

            try {
                await authUtils.authenticateAppAndGetToken(clientId, accessKey.body.clientSecret);

                const auditLog = await auditLogsUtils.findLoginAudit(
                    adminToken,
                    clientId,
                    'client_credentials'
                );
                expect(auditLog.subjectType, 'client_credentials login audit should record subjectType as ServiceAccount').toBe(
                    'ServiceAccount' as unknown as AuditLog.SubjectTypeEnum
                );
                expect(auditLog.subject, 'client_credentials login audit subject should match the access key clientId').toBe(clientId);
            } finally {
                await accessKeyUtils.deleteAccessKey(adminToken, accessKey.body.id);
            }
        });

        test('should audit a failed password (Login API) authentication', async () => {
            const user: { response: IncomingMessage; body: UserPostResponse } =
                await usersManager.createLocalUser(
                    flowRootState.users.system_administrator.accessToken,
                    objectGenerator.generateEmailByPrefix('loginauditpwdfail'),
                    false
                );

            try {
                await expect(
                    authUtils.authenticateTokenApi(user.body.username, 'WrongPassword!1'),
                    'authentication with wrong password should throw'
                ).rejects.toThrow();

                const auditLog = await auditLogsUtils.findLoginAudit(
                    flowRootState.users.system_administrator.accessToken,
                    user.body.username,
                    'password',
                    AuditLog.ResultEnum.Failed
                );
                expect(auditLog.subjectType, 'failed password login audit should record subjectType as User').toBe(AuditLog.SubjectTypeEnum.User);
                expect(auditLog.subject, 'failed password login audit subject should match the username').toBe(user.body.username);
            } finally {
                await usersManager.deleteLocalUser(
                    flowRootState.users.system_administrator.accessToken,
                    user.body.id
                );
            }
        });

        test('should audit a failed client_credentials (Login Application) authentication', async () => {
            const adminToken = flowRootState.users.system_administrator.accessToken;
            const accessKey = await accessKeyUtils.createAccessKey(adminToken, {
                name: objectGenerator.generateShortNameByPrefix('loginauditappfail')
            });
            const clientId = (await accessKeyUtils.getAccessKeyById(adminToken, accessKey.body.id)).body
                .clientId;

            try {
                await expect(
                    authUtils.authenticateAppAndGetToken(clientId, 'wrong-secret'),
                    'authentication with wrong client secret should throw'
                ).rejects.toThrow();

                const auditLog = await auditLogsUtils.findLoginAudit(
                    adminToken,
                    clientId,
                    'client_credentials',
                    AuditLog.ResultEnum.Failed
                );
                expect(auditLog.subjectType, 'failed client_credentials login audit should record subjectType as ServiceAccount').toBe(
                    'ServiceAccount' as unknown as AuditLog.SubjectTypeEnum
                );
                expect(auditLog.subject, 'failed client_credentials login audit subject should match the access key clientId').toBe(clientId);
            } finally {
                await accessKeyUtils.deleteAccessKey(adminToken, accessKey.body.id);
            }
        });
    }
);
