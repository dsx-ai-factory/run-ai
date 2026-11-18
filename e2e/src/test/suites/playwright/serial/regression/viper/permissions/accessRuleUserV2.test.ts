/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { StatusCodes } from 'http-status-codes';
import { UserPostResponse } from '../../../../../../../modules/api/apis/identityManagerApi/generated/api';
import { Project } from '../../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    accessRulesClusterUtils,
    authorizationUtils,
    authUtils,
    flowRootState,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    settings,
    tags,
    TokenObject,
    usersManager,
    Utils
} from '../../../../../../../modules/helpers';
import { shouldRunTests } from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { expect } from 'chai';
import { test } from '@playwright/test';

test.describe(`Access Rule User: ${tags.teams.phoenix} ${tags.suites.legacySphinx} ${tags.tool.api} ${tags.resource.accessRules} ${tags.resource.users}`, function () {
    shouldRunTests(settings.isSelfHosted)(function () {
        let userV2: UserPostResponse;
        let userEmail: string;
        let userPassword: string;
        let projectName: string;
        let namespace: string;
        let user: TokenObject;
        let project: Project;

        test.beforeEach('Create user with view access rule', async function () {
            userEmail = objectGenerator.generateEmailByPrefix(`accessruleview-${Date.now()}`);
            projectName = objectGenerator.generateNameByPrefix(`projectviewer-${Date.now()}`);
            namespace = Utils.namespaceFromProjectName(projectName);
            userV2 = (
                await usersManager.createLocalUser(
                    flowRootState.users.system_administrator.accessToken,
                    userEmail,
                    false
                )
            ).body;
            userPassword = userV2.tempPassword;
            logger.debug(`User with email ${userEmail} was created`);
            const accessRuleViewer = await authorizationUtils.createAccessRuleForUser(
                flowRootState.users.system_administrator.accessToken,
                userV2.username,
                flowRootState.users.viewer.role.id
            );
            logger.debug(
                `Access rule ${accessRuleViewer.roleName} for user with email ${userEmail} was created`
            );
            user = await authUtils.authenticateUser(userEmail, userPassword);
        });

        test('should fail to create project when user has viewer access rule only', async function () {
            logger.info('User with viewer access rule should fail to create project');
            try {
                await projectResourcesUtils.createProject(
                    namespace,
                    {
                        name: projectName,
                        deservedGpus: 2,
                        maxAllowedGpus: 2
                    },
                    user.idToken
                );
            } catch (error) {
                expect(error.response.statusCode).equal(StatusCodes.FORBIDDEN);
            }
            logger.debug(`User ${userEmail} with viewer access rule failed to create project ${projectName}`);
        });

        test('should successfully create project after upgrading user from viewer to editor access', async function () {
            logger.info(
                'Should successfully create project after upgrading user from viewer to editor access'
            );
            const accessRuleEditor = await authorizationUtils.createAccessRuleForUser(
                flowRootState.users.system_administrator.accessToken,
                userV2.username,
                flowRootState.users.editor.role.id
            );
            logger.debug(
                `Access rule ${accessRuleEditor.roleName} for user with email ${userEmail} was created`
            );
            project = await projectResourcesUtils.createProject(
                namespace,
                {
                    name: projectName,
                    deservedGpus: 2,
                    maxAllowedGpus: 2
                },
                user.idToken
            );
            logger.debug(
                `User ${userEmail} with editor access rule successfully created project ${projectName}`
            );
            expect(project.createdAt).not.equal(undefined);
            expect(accessRuleEditor.roleName).equal(flowRootState.users.editor.role.name);
            const userAccessRulesValidation = await accessRulesClusterUtils.getAccessRulesByUser(
                flowRootState.users.system_administrator.accessToken,
                userV2.username
            );
            expect(userAccessRulesValidation).equal(2);
        });

        test.afterAll('Clean up: delete test user and created projects', async function () {
            if (!settings.isSelfHosted) {
                return logger.debug('this after does not run on ocp');
            }
            await usersManager.deleteLocalUser(flowRootState.users.system_administrator.accessToken, userV2.id);
            logger.debug(`User with email ${userEmail} was deleted`);
            await operations.deleteProjectAndNamespace(
                project.id,
                flowRootState.users.system_administrator.accessToken
            );
        });
    });
});
