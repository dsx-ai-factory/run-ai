/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    environmentUtils,
    flowRootState,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    projectUtils,
    tags,
    Utils
} from '../../../../../../modules/helpers';
import { kubeCli, bashCli } from '../../../../../../modules/cli';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { DataSourcePage } from '../../../../../../modules/ui/runaiV2/pages/DataSourcePage';
import { WorkspacePage } from '../../../../../../modules/ui/runaiV2/pages/WorkspacePage';
import { TrainingPage } from '../../../../../../modules/ui/runaiV2/pages/TrainingPage';
import { CredentialPage } from '../../../../../../modules/ui/runaiV2/pages/CredentialPage';
import { EnvironmentAsset, Scope } from '../../../../../../modules/api/apis/assetsApi';
import { EWorkloadStatus } from '../../../../../../modules/ui/runaiV2/pages/WorkloadPage';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test, expect } from '@playwright/test';

test.describe(`Git Datasource E2E Test, tags: ${tags.tool.ui} ${tags.resource.datasource} ${tags.teams.loong} ${tags.runOn.dgxc}`, function () {
    supportedFrom(SupportedVersions.V2_19)(function () {
        let browserEngine: BrowserEngine;
        let auth: AuthPage;
        let dataSourcePage: DataSourcePage;
        let workspacePage: WorkspacePage;
        let trainingPage: TrainingPage;
        let credentialsPage: CredentialPage;
        let videoName: string;
        let projectName: string;
        let namespace: string;
        let project: Project;

        test.beforeAll(async function () {
            browserEngine = new BrowserEngine();
            projectName = objectGenerator.generateShortNameByPrefix('project-e2e-for-git-ds');
            namespace = Utils.namespaceFromProjectName(projectName);
            project = await projectResourcesUtils.createProject(namespace, {
                name: projectName,
                deservedGpus: 0,
                maxAllowedGpus: 0
            });
            const roleBindingName = projectUtils.getSystemAdministratorRoleBindingName();

            await projectUtils.waitForRoleBindingForUserInProject(
                namespace,
                flowRootState.users.system_administrator.email,
                roleBindingName,
                true
            );
        });

        test.afterAll('deleting project and namespace', async () => {
            try {
                await operations.deleteProjectAndNamespace(project.id);
                logger.info(`deleted project ${projectName} an namespace ${namespace}`);
            } catch (error) {
                logger.error(`failed to delete project ${projectName} an namespace ${namespace}`);
                throw error;
            }
        });

        test.beforeEach(async function () {
            await browserEngine.start();
            auth = new AuthPage(browserEngine.currentPage);
            dataSourcePage = new DataSourcePage(browserEngine.currentPage);
            workspacePage = new WorkspacePage(browserEngine.currentPage);
            trainingPage = new TrainingPage(browserEngine.currentPage);
            credentialsPage = new CredentialPage(browserEngine.currentPage);
            await auth.signIn(
                flowRootState.users.l1_researcher.email,
                flowRootState.users.l1_researcher.password
            );
        });

        test.afterEach('stop browser', async ({}, testInfo) => {
            await browserEngine.stopAndArchive(testInfo, videoName);
        });

        test('T4042313: Create git data source with public git repository', async ({}) => {
            videoName = 'T4042313-public-git-ds';
            const mountPath = '/mnt/data';
            const publicGitDataSourceName = objectGenerator.generateNameByPrefix('public-git-ds');
            const publicGitTrainingName = objectGenerator.generateShortNameByPrefix('public-git-training');

            await test.step('create public git data source', async () => {
                await dataSourcePage.createGit(publicGitDataSourceName, {
                    scope: Scope.Project,
                    scopeEntityName: project.name,
                    repository: 'https://github.com/octocat/Hello-World',
                    branch: 'master',
                    containerPath: mountPath
                });
            });

            await test.step('create training workload with the public git data source', async () => {
                await trainingPage.createFlexibleTraining(projectName, publicGitTrainingName, {
                    dsName: publicGitDataSourceName
                });
                await trainingPage.waitForWorkloadUIStatus(publicGitTrainingName, EWorkloadStatus.Running);
            });

            await test.step('verify the public git repo content in the pod', async () => {
                const pods = await kubeCli.listPodsOfJob(namespace, publicGitTrainingName);
                for (const pod of pods) {
                    const repoContent = bashCli.exec(
                        `kubectl exec -it ${pod.metadata.name} -n ${namespace} -- ls ${mountPath}/Hello-World`
                    );
                    logger.info(`ls ${mountPath}/Hello-World: ${repoContent}`);
                    expect(repoContent).toContain('README');

                    const readmeContent = bashCli.exec(
                        `kubectl exec -it ${pod.metadata.name} -n ${namespace} -- cat ${mountPath}/Hello-World/README`
                    );
                    logger.info(`cat ${mountPath}/Hello-World/README: ${readmeContent}`);
                    expect(readmeContent).toContain('Hello World!');
                }
            });

            await test.step('doing cleanup for public git test', async () => {
                await trainingPage.deleteWorkload(publicGitTrainingName);
                await dataSourcePage.goToPage();
                await dataSourcePage.deleteDataSource(publicGitDataSourceName);
            });
        });
    });
});
