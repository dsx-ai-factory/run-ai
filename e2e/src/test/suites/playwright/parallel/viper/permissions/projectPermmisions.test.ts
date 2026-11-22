/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    authUtils,
    config,
    flowRootState,
    logger,
    objectGenerator,
    projectUtils,
    settings,
    tags,
    usersManager,
    Utils,
    workloadsCrudUtils
} from '../../../../../../modules/helpers';
import { backendApi, backendClusterApi, projectsDepartmentsApi } from '../../../../../../modules/api';
import { createProject } from '../../../../../../modules/helpers/models/projects';

import { Training } from '../../../../../../modules/api/apis/workloadsApi/generated/model/training';

import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { Role } from '../../../../../../modules/api/apis/backendApi';
import { StatusCodes } from 'http-status-codes';
import { expect } from 'chai';
import { test } from '@playwright/test';
import { UserPostResponse } from '../../../../../../modules/api/apis/identityManagerApi/generated/model/userPostResponse';

test.describe(
    `Basic permission tests, tags: ${tags.teams.viper}`,
    { tag: [tags.services.authorization, tags.runOn.parallelCI] },
    () => {
        // This e2e test is used to verify that only authorized users can submit jobs

        const numberOfUsersToCreate = 2; // 2 users created are:
        // researcher A (assigned to the project), researcher B (not assigned)
        const userNames = ['researcher-with-permissions', 'researcher-without-permissions'];
        const testUsers = new Array<UserPostResponse>(0);
        const testUserTokenIds = new Array<string>(0);
        const researcherRole = [Role.Researcher];

        let project: Project;
        let namespace: string;
        let testWorkload: Training;

        test.beforeAll(async function () {
            if (!settings.enableAuth) {
                // this test is only interesting if auth is enabled
                this.skip();
            }
            logger.debug('Set up test environment');
            // set accessTokens
            backendClusterApi.jobsApi.accessToken = flowRootState.users.system_administrator.accessToken;
            backendApi.authenticationApi.accessToken = flowRootState.users.system_administrator.accessToken;

            // create users
            for (let index = 0; index < numberOfUsersToCreate; index++) {
                const researcherEmail = objectGenerator.generateEmailByPrefix(userNames[index]);
                logger.debug(`Creating user: ${userNames[index]}, with roles: ${researcherRole}`);
                const userRes = (
                    await usersManager.createLocalUser(
                        flowRootState.users.system_administrator.accessToken,
                        researcherEmail,
                        false
                    )
                ).body;

                const userToken = await authUtils.authenticateUser(userRes.username, userRes.tempPassword);

                testUsers.push(userRes);
                testUserTokenIds.push(userToken.idToken);
            }

            // create project and assign researcher A to it
            backendClusterApi.jobsApi.accessToken = flowRootState.users.system_administrator.accessToken;
            const options = {
                maxAllowedGpus: 1,
                deservedGpus: 1,
                permissions: { users: [testUsers[0].username] }
            };
            const projectName = objectGenerator.generateNameByPrefix('project');
            project = await createProject(projectName, options);
            namespace = Utils.namespaceFromProjectName(project.name);

            logger.info(`Project's name: ${project.name}, project's namespace: ${namespace}`);

            const projectResponse = (
                await projectsDepartmentsApi.projectsApi.getProjects(flowRootState.cluster.uuid)
            ).body.find((p) => p.id === project.id);

            expect(projectResponse.permissions.users).to.contains(testUsers[0].username);
            logger.debug(`Researcher ${testUsers[0].username} is assigned to project ${project.name}
        with namespace ${namespace}`);

            const runaiResearcherRoleBindingName = projectUtils.getRunaiResearcherRoleBindingName();

            await projectUtils.waitForRoleBindingForUserInProject(
                namespace,
                testUsers[0].username,
                runaiResearcherRoleBindingName
            );
        });

        test.afterAll(async function () {
            const promisesToAwait = [];
            testUsers.forEach((user) => {
                promisesToAwait.push(
                    usersManager.deleteLocalUser(
                        flowRootState.users.system_administrator.accessToken,
                        user.id
                    )
                );
            });
            await Promise.all(promisesToAwait);
        });

        test.describe('Workload submission by project assignments (users)', () => {
            const researcherWithPermissionsIndex = 0;
            const researcherWithoutPermissionsIndex = 1;

            async function submitTrainingAsUser(token: string): Promise<Training> {
                const workloadName = objectGenerator.generateNameByPrefix('w');
                const res = await workloadsCrudUtils.createTraining(
                    workloadName,
                    false,
                    project.clusterUuid,
                    project.id.toString(),
                    {
                        image: config.quickStartImage.ubuntu,
                        command: 'sleep infinity'
                    },
                    token
                );
                return res.body;
            }

            // happy submit workload flow - researcher
            test('Researcher with permissions to a project should be able to submit workload', async function () {
                testWorkload = await submitTrainingAsUser(testUserTokenIds[researcherWithPermissionsIndex]);

                expect(testWorkload).to.not.be.empty;
                expect(testWorkload.name).to.match(/^w-/);
                expect(testWorkload.projectId).to.equal(project.id.toString());
            });

            // negative submit workload flow - researcher
            test('Researcher without permissions to a project should NOT be able to submit workload', async function () {
                try {
                    testWorkload = await submitTrainingAsUser(
                        testUserTokenIds[researcherWithoutPermissionsIndex]
                    );
                    expect.fail(
                        'Researcher without permissions should not be able to submit jobs to project'
                    );
                } catch (error) {
                    expect(error.response.statusCode).to.be.oneOf([
                        StatusCodes.FORBIDDEN,
                        StatusCodes.UNAUTHORIZED
                    ]);
                }
            });

            // submit workload flow - researcher manager
            test('Researcher manager (with permissions to all projects) should be able to submit a workload', async function () {
                testWorkload = await submitTrainingAsUser(
                    flowRootState.users.research_manager.accessToken
                );

                expect(testWorkload).to.not.be.empty;
                expect(testWorkload.name).to.match(/^w-/);
                expect(testWorkload.projectId).to.equal(project.id.toString());
            });
        });
    }
);
