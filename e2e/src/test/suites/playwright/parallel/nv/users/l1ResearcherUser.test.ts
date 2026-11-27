/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    flowRootState,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    Utils,
    waitUtils,
    config,
    usersManager,
    authUtils,
    usersUtils
} from '../../../../../../modules/helpers';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test, expect } from '@playwright/test';
import { MENU_SUBJECT_NAMES } from '../../../../../../modules/ui/utils/sidenavUtils';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { ProjectPage } from '../../../../../../modules/ui/runaiV2/pages/ProjectPage';
import { kubeCli, bashCli, runaiCliV2 } from '../../../../../../modules/cli';
import { KubeCli } from '../../../../../../modules/cli/tools/KubeCli';
import { ScopeType } from '../../../../../../modules/api/apis/authorizationApi/generated/model/scopeType';
import { User } from '../../../../../../modules/helpers/config/types';
import { cliV2WorkloadType } from '../../../../../../modules/cli/tools/RunaiCliV2.constants';

test.describe(`L1 Researcher User - access control, tags: ${tags.teams.loong} ${tags.tool.ui} ${tags.resource.users}`, () => {
    supportedFrom(SupportedVersions.V2_18)(function () {
        let browserEngine: BrowserEngine;
        let authPage: AuthPage;
        let projectPage: ProjectPage;
        let projectName: string;
        let namespace: string;
        let project: Project;

        let videoName: string;

        test.beforeAll(async function () {
            browserEngine = new BrowserEngine();

            projectName = objectGenerator.generateShortNameByPrefix('proj-l1-researcher', 5);
            namespace = Utils.namespaceFromProjectName(projectName);
            project = await projectResourcesUtils.createProject(namespace, {
                name: projectName,
                deservedGpus: 0,
                maxAllowedGpus: 0
            });
        });

        test.beforeEach(async function () {
            await browserEngine.start();
            authPage = new AuthPage(browserEngine.currentPage);
            projectPage = new ProjectPage(browserEngine.currentPage);

            await authPage.signIn(
                flowRootState.users.l1_researcher.email,
                flowRootState.users.l1_researcher.password
            );
        });

        test.afterEach('stop browser', async ({}, testInfo) => {
            await browserEngine.stopAndArchive(testInfo, videoName);
        });

        test.afterAll('deleting project and namespace', async () => {
            try {
                await operations.deleteProjectAndNamespace(project.id);
                logger.info(`deleted project ${projectName} and namespace ${namespace}`);
            } catch (error) {
                logger.error(`failed to delete project ${projectName} and namespace ${namespace}`);
                throw error;
            }
        });

        test('T3738083: L1 researcher user cannot access nodes page', async () => {
            videoName = 'l1-researcher-no-nodes-access';

            await test.step('verify L1 researcher cannot see Resources menu', async () => {
                await expect(
                    authPage.page.locator(`[aid="${MENU_SUBJECT_NAMES.Resources}"]`)
                ).not.toBeVisible();
            });
        });

        test('T3738075: L1 Researcher user is not able to manage users', async () => {
            videoName = 'l1-researcher-no-manage-users-permission';

            await test.step('verify L1 Researcher user is not able to see the Access menu', async () => {
                await expect(authPage.page.locator(`[aid="${MENU_SUBJECT_NAMES.Access}"]`)).not.toBeVisible();
            });
        });

        test('T3738081: L1 Researcher user can not create/edit/delete a project', async () => {
            videoName = 'l1-researcher-cannot-manipulate-project';

            await test.step('verify L1 Researcher user is able to open the project page', async () => {
                await projectPage.goToPage();
            });

            await test.step('verify view only badge is shown on the project page', async () => {
                await waitUtils.waitForCondition(
                    async () => {
                        return await projectPage.viewOnlyBadgeVisible();
                    },
                    'view only badge becomes visible on the page',
                    true,
                    20
                );
            });

            await test.step('verify L1 Researcher user can not create a project', async () => {
                await expect(projectPage.locator(projectPage.addNewProjectBtn)).toBeDisabled();
            });

            await test.step('verify L1 Researcher user can not edit a project', async () => {
                await projectPage.clickOnElementByText('td', projectName);
                await expect(projectPage.locator(projectPage.editProjectBtn)).toBeDisabled();
            });

            await test.step('verify L1 Researcher user can not delete a project', async () => {
                await expect(projectPage.locator(projectPage.deleteProjectBtn)).toBeDisabled();
            });
        });
    });
});

test.describe(`L1 researchers permission check of kubetl and runai, tags: ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_18)(function () {
        let projectName1: string;
        let projectName2: string;
        let namespace1: string;
        let namespace2: string;
        let project1: Project;
        let project2: Project;
        let user1: User;
        let user2: User;
        const runaiCliPath = config.runaiV2CliPath;
        let originalPath: string;

        test.beforeAll(function () {
            projectName1 = objectGenerator.generateShortNameByPrefix('project-test1-');
            projectName2 = objectGenerator.generateShortNameByPrefix('project-test2-');
            namespace1 = Utils.namespaceFromProjectName(projectName1);
            namespace2 = Utils.namespaceFromProjectName(projectName2);
            bashCli.exec(`cp ${config.kubeconfigPath} ${config.kubeconfigPath}.bak`);
            originalPath = process.env.PATH;
            process.env.PATH = `${runaiCliPath.substring(0, runaiCliPath.lastIndexOf('/'))}:${process.env.PATH}`;
        });

        test.afterAll('deleting project and namespace', async () => {
            process.env.PATH = originalPath;
            bashCli.exec(`cp ${config.kubeconfigPath}.bak ${config.kubeconfigPath}`);

            const usersToDelete = [user1, user2].filter((user) => user && user.userId);
            await Promise.all(
                usersToDelete.map(async (user) => {
                    try {
                        await usersManager.deleteLocalUser(
                            flowRootState.users.system_administrator.accessToken,
                            user.userId
                        );
                        logger.info(`deleted user ${user.email}`);
                    } catch (error) {
                        logger.error(`failed to delete user ${user.email}`, error);
                    }
                })
            );
            await Promise.all(
                [project1.id, project2.id].map(async (projectId, i) => {
                    try {
                        await operations.deleteProjectAndNamespace(projectId);
                        logger.info(
                            `deleted project ${[projectName1, projectName2][i]} and namespace ${[namespace1, namespace2][i]}`
                        );
                    } catch (error) {
                        logger.error(
                            `failed to delete project ${[projectName1, projectName2][i]} and namespace ${[namespace1, namespace2][i]}`,
                            error
                        );
                    }
                })
            );
        });

        test('T3738097: access K8s APIs associated with my workloads / Projects / namespaces with federated SF identity', async function () {
            test.setTimeout(300 * 1000);
            await test.step('Create 2 projects', async () => {
                project1 = await projectResourcesUtils.createProject(namespace1, {
                    name: projectName1,
                    deservedGpus: 0,
                    maxAllowedGpus: -1
                });
                project2 = await projectResourcesUtils.createProject(namespace2, {
                    name: projectName2,
                    deservedGpus: 0,
                    maxAllowedGpus: -1
                });
            });

            await test.step('Create user1 under project1 and user2 under project2', async () => {
                const l1ResearcherRole = flowRootState.users.l1_researcher.role;
                user1 = await usersManager.createUserWithRole(
                    flowRootState.users.system_administrator.accessToken,
                    l1ResearcherRole,
                    'user1',
                    project1.id.toString(),
                    ScopeType.Project
                );
                user2 = await usersManager.createUserWithRole(
                    flowRootState.users.system_administrator.accessToken,
                    l1ResearcherRole,
                    'user2',
                    project2.id.toString(),
                    ScopeType.Project
                );
                for (const user of [user1, user2]) {
                    const token = await authUtils.authenticateUser(user.email, user.password);
                    await usersUtils.changeUserPassword(
                        token.idToken,
                        user.password,
                        config.controlPlaneAdminPassword
                    );
                    const tokenNew = await authUtils.authenticateUser(
                        user.email,
                        config.controlPlaneAdminPassword
                    );
                    expect(tokenNew).not.toBeNull();
                }
            });

            await test.step('Test kubectl permissions', async () => {
                runaiCliV2.loginAsUser(user1.email, config.controlPlaneAdminPassword);
                runaiCliV2.projectSet(project1.name);

                runaiCliV2.kubeconfigSet();

                logger.info('Recreate kubeCli instance with updated kubeconfig');
                const updatedKubeCli = new KubeCli(config.kubeconfigPath, config.kubeconfigClusterName);

                logger.info('Wait for kubeconfig to be set and then test permissions');
                await waitUtils.waitForCondition(
                    async () => {
                        try {
                            const namespaces = await updatedKubeCli.listNamespaces();
                            const namespaceNames = namespaces.items.map((ns) => ns.metadata.name);
                            logger.info(`kubectl get ns output: ${namespaceNames.join(', ')}`);
                            return namespaceNames.includes(namespace1) && namespaceNames.includes(namespace2);
                        } catch (error) {
                            logger.error(`kubectl get ns error: ${error}`);
                            return false;
                        }
                    },
                    'wait for kubectl get ns to show namespaces (up to 2min)',
                    true,
                    180
                );

                logger.info('Test listPodsByNamepsace for namespace1 (should succeed)');
                try {
                    const pods1 = await updatedKubeCli.listPodsByNamepsace(namespace1);
                    logger.info(`kubectl get pods ns1: found ${pods1.length} pods`);
                    expect(pods1.length).toBe(0); // Should be empty
                } catch (error) {
                    logger.error(`kubectl get pods ns1 error: ${error}`);
                    expect(error.message).toContain(`No resources found in ${namespace1} namespace`);
                }

                logger.info('Test listPodsByNamepsace for namespace2 (should fail with forbidden)');
                try {
                    await updatedKubeCli.listPodsByNamepsace(namespace2);
                    throw new Error('Expected forbidden error but got success');
                } catch (error) {
                    logger.info(`kubectl get pods ns2 error: ${error.message}`);
                    expect(error.message.toLowerCase()).toMatch(
                        new RegExp(`forbidden.*${user1.email}.*${namespace2}`)
                    );
                }

                logger.info('Test listPodsByNamepsace for kyverno namespace (should fail with forbidden)');
                try {
                    await updatedKubeCli.listPodsByNamepsace('kyverno');
                    throw new Error('Expected forbidden error but got success');
                } catch (error) {
                    logger.info(`kubectl get pods kyverno error: ${error.message}`);
                    expect(error.message.toLowerCase()).toMatch(
                        new RegExp(`forbidden.*${user1.email}.*kyverno`)
                    );
                }

                logger.info('Test getNodes (should succeed)');
                const nodes = await updatedKubeCli.getNodes('nvidia.com/gpu.present=true');
                const nodeName = nodes.items[0].metadata?.name;
                logger.info(`kubectl get nodes: found ${nodes.items.length} nodes`);
                expect(nodes.items.length).toBeGreaterThan(0);
                expect(nodeName).toBeDefined();

                logger.info('Test getNodeLabels (should succeed)');
                const nodeLabels = await updatedKubeCli.getNodeLabels(nodeName);
                logger.info(
                    `kubectl describe node output: node has ${Object.keys(nodeLabels).length} labels`
                );
                expect(Object.keys(nodeLabels).length).toBeGreaterThanOrEqual(0);

                logger.info('Test cordon operation (should fail with forbidden)');
                try {
                    await updatedKubeCli.patchNode(nodeName, { spec: { unschedulable: true } });
                    throw new Error('Expected forbidden error but got success');
                } catch (error) {
                    logger.info(`kubectl cordon error: ${error.message}`);
                    expect(error.message.toLowerCase()).toContain('forbidden');
                }
            });

            await test.step('Test runai CLI permissions', () => {
                const workloadName = objectGenerator.generateShortNameByPrefix('workload-test-');
                const workloadName2 = objectGenerator.generateShortNameByPrefix('workload-test-2-');
                const image = 'ghcr.io/knative/helloworld-go';

                logger.info('Test project list - should only show own project');
                const projectList = runaiCliV2.list('project');
                logger.info(`runai project list output: ${projectList.stdout}`);
                expect(projectList.stdout).toContain(project1.name);
                expect(projectList.stdout).not.toContain(project2.name);

                logger.info('Test submit to own project (should succeed)');
                runaiCliV2.submit(
                    { '--image': image, '--project': project1.name } as any,
                    workloadName,
                    cliV2WorkloadType.training
                );
                const workloadListOwn = runaiCliV2.list('workload', { '--project': project1.name } as any);
                logger.info(`runai workload list (own project) output: ${workloadListOwn.stdout}`);
                expect(workloadListOwn.stdout).toContain(workloadName);

                logger.info('Test workload list for other project (should fail)');
                try {
                    const workloadListOther = runaiCliV2.list('workload', {
                        '--project': project2.name
                    } as any);
                    logger.info(`runai workload list (other project) output: ${workloadListOther.stderr}`);
                    expect(workloadListOther.stderr).toContain('Workloads list failed');
                } catch (error) {
                    logger.info(`runai workload list (other project) error: ${error.message}`);
                    expect(error.message).toContain('Workloads list failed');
                }

                logger.info('Test submit to other project (should fail)');
                try {
                    runaiCliV2.submit(
                        { '--image': image, '--project': project2.name } as any,
                        workloadName2,
                        cliV2WorkloadType.training
                    );
                    throw new Error('Expected error but got success');
                } catch (error) {
                    logger.info(`runai training submit (other project) error: ${error.message}`);
                    expect(error.message).toMatch('failed to submit');
                }
            });
        });
    });
});
