/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    authorizationUtils,
    authUtils,
    config,
    flowRootState,
    logger,
    objectGenerator,
    orgUnitUtils,
    projectResourcesUtils,
    tags,
    TokenObject,
    usersManager,
    Utils,
    accessKeyUtils
} from '../../../../../../modules/helpers';
import {
    AccessKeyPostResponse,
    UserPostResponse
} from '../../../../../../modules/api/apis/identityManagerApi/generated/api';
import { IncomingMessage } from 'http';
import { AccessRule } from '../../../../../../modules/api/apis/authorizationApi/generated/api';
import { expect } from 'chai';
import { test } from '@playwright/test';

test.describe(
    `Access Keys Access tags: ${tags.teams.phoenix} ${tags.suites.legacySphinx} ${tags.tool.api} ${tags.runOn.k8sFlavCertKit} ${tags.runOn.k8sFlavCertKitTemp}`,
    { tag: [tags.services.identityManager] },
    function () {
        let accessKeyName: string;
        let createdAccessKey: { response: IncomingMessage; body: AccessKeyPostResponse };
        let editorUser: { response: IncomingMessage; body: UserPostResponse };
        let editorUserToken: TokenObject;
        let userAccessRule: AccessRule;
        let initialProjectName: string;
        let projectAfterAccessChangeName: string;
        let projectWithoutAccessName: string;
        let initialNamespace: string;
        let namespaceAfterAccessChange: string;
        let namespaceWithoutAccess: string;
        let initialProject: any;
        let projectAfterAccessChange: any;

        test.beforeEach(async function () {
            initialProjectName = objectGenerator.generateNameByPrefix('initialproject');
            projectAfterAccessChangeName = objectGenerator.generateNameByPrefix('projectafteraccesschange');
            initialNamespace = Utils.namespaceFromProjectName(initialProjectName);
            namespaceAfterAccessChange = Utils.namespaceFromProjectName(projectAfterAccessChangeName);
            projectWithoutAccessName = objectGenerator.generateNameByPrefix('projectwithoutaccess');
            namespaceWithoutAccess = Utils.namespaceFromProjectName(projectWithoutAccessName);

            editorUser = await usersManager.createLocalUser(
                flowRootState.users.system_administrator.accessToken,
                objectGenerator.generateEmailByPrefix('editoracces'),
                false,
                false
            );

            userAccessRule = await authorizationUtils.createAccessRuleForUser(
                flowRootState.users.system_administrator.accessToken,
                editorUser.body.username,
                flowRootState.users.editor.role.id
            );

            editorUserToken = await authUtils.authenticateUser(
                editorUser.body.username,
                editorUser.body.tempPassword
            );

            accessKeyName = objectGenerator.generateShortNameByPrefix('accesskey');
            createdAccessKey = await accessKeyUtils.createAccessKey(editorUserToken.accessToken, {
                name: accessKeyName
            });
        });

        test.afterEach(async function () {
            try {
                await orgUnitUtils.deleteProject(
                    initialProject.id,
                    flowRootState.users.system_administrator.accessToken
                );
                await orgUnitUtils.deleteProject(
                    projectAfterAccessChange.id,
                    flowRootState.users.system_administrator.accessToken
                );
                logger.debug(`projects: ${initialProject.id} , ${projectAfterAccessChange.id} deleted`);
                await authUtils.deleteUser(
                    editorUser.body.id,
                    flowRootState.users.system_administrator.accessToken
                );
                logger.debug(`Deleted user with id: ${editorUser.body.id}`);
            } catch (error) {
                logger.error(`Failed to clean up personal application with id: ${JSON.stringify(error)}`);
            }
        });

        test('should reflect user access rule changes by invalidating existing tokens and enforcing new permissions', async function () {
            const fetchedAccessRule = await authorizationUtils.getAccessRuleById(userAccessRule.id);
            expect(fetchedAccessRule.roleId, 'fetched access rule should have the editor role id').to.equal(flowRootState.users.editor.role.id);

            const appAuthToken = await authUtils.authenticatePersonalApp(
                createdAccessKey.body.clientId,
                createdAccessKey.body.clientSecret,
                config.tenantUrl
            );

            initialProject = await projectResourcesUtils.createProject(
                initialNamespace,
                {
                    name: initialProjectName,
                    maxAllowedGpus: 2,
                    deservedGpus: 2,
                    permissions: { applications: [accessKeyName], groups: [] }
                },
                appAuthToken.accessToken
            );

            logger.debug(`Project created: ${initialProject.name} with id: ${initialProject.id}`);
            expect(initialProject, 'project creation with valid access key token should succeed').not.to.be.null;

            await authorizationUtils.deleteAccessRule(
                userAccessRule,
                flowRootState.users.system_administrator.accessToken
            );

            try {
                await projectResourcesUtils.createProject(
                    namespaceWithoutAccess,
                    {
                        name: projectWithoutAccessName,
                        maxAllowedGpus: 2,
                        deservedGpus: 2,
                        permissions: { applications: [accessKeyName], groups: [] }
                    },
                    appAuthToken.accessToken
                );
                expect.fail('Project creation succeeded without valid access rules');
            } catch (error) {
                expect(error.statusCode, 'project creation should fail with 403 after user access rule is revoked').to.equal(403);
            }

            await authorizationUtils.createAccessRuleForUser(
                flowRootState.users.system_administrator.accessToken,
                editorUser.body.username,
                flowRootState.users.editor.role.id
            );

            const appAuthTokenAfterAccessChange = await authUtils.authenticatePersonalApp(
                createdAccessKey.body.clientId,
                createdAccessKey.body.clientSecret,
                config.tenantUrl
            );

            projectAfterAccessChange = await projectResourcesUtils.createProject(
                namespaceAfterAccessChange,
                {
                    name: projectAfterAccessChangeName,
                    maxAllowedGpus: 2,
                    deservedGpus: 2,
                    permissions: { applications: [accessKeyName], groups: [] }
                },
                appAuthTokenAfterAccessChange.accessToken
            );

            logger.debug(
                `Project created after access changed: ${projectAfterAccessChange.name} with id: ${projectAfterAccessChange.id}`
            );
            expect(projectAfterAccessChange, 'project creation should succeed after access rule is re-granted').not.to.be.null;
        });
    }
);
