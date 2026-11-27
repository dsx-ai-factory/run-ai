/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    dataSourcesUtils,
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
import { CredentialPage } from '../../../../../../modules/ui/runaiV2/pages/CredentialPage';
import { DataSourcePage } from '../../../../../../modules/ui/runaiV2/pages/DataSourcePage';
import { TrainingPage } from '../../../../../../modules/ui/runaiV2/pages/TrainingPage';
import { Scope } from '../../../../../../modules/api/apis/assetsApi';
import { EWorkloadStatus } from '../../../../../../modules/ui/runaiV2/pages/WorkloadPage';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test, expect } from '@playwright/test';

test.describe(`Datasource Secret E2E Test, tags: ${tags.tool.ui} ${tags.resource.credential} ${tags.teams.loong}`, function () {
    supportedFrom(SupportedVersions.V2_19)(function () {
        let browserEngine: BrowserEngine;
        let auth: AuthPage;
        let credentialsPage: CredentialPage;
        let dataSourcePage: DataSourcePage;
        let trainingPage: TrainingPage;

        let videoName: string;
        let projectName: string;
        let namespace: string;
        let project: Project;

        let credentialName: string;
        let username: string;
        let password: string;
        let dataSourceName: string;
        let trainingName: string;

        test.beforeAll(async function () {
            browserEngine = new BrowserEngine();
            projectName = objectGenerator.generateShortNameByPrefix('project-e2e-for-secret-ds');
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
                logger.info(`deleted project ${projectName} and namespace ${namespace}`);
            } catch (error) {
                logger.error(`failed to delete project ${projectName} and namespace ${namespace}`);
                throw error;
            }
        });

        test.beforeEach(async function () {
            await browserEngine.start();
            auth = new AuthPage(browserEngine.currentPage);
            credentialsPage = new CredentialPage(browserEngine.currentPage);
            dataSourcePage = new DataSourcePage(browserEngine.currentPage);
            trainingPage = new TrainingPage(browserEngine.currentPage);

            await auth.signIn(
                flowRootState.users.l1_researcher.email,
                flowRootState.users.l1_researcher.password
            );
        });

        test.afterEach('stop browser', async ({}, testInfo) => {
            await browserEngine.stopAndArchive(testInfo, videoName);
        });

        async function deleteDataSourceByName(name: string): Promise<void> {
            await dataSourcePage.goToPage();
            await dataSourcePage.deleteDataSource(name);
            await dataSourcesUtils.waitForDataSourceToBeDeleted(name);
        }

        test.describe('test secret data source - username & password credentials', function () {
            test.beforeEach(function () {
                credentialName = objectGenerator.generateNameByPrefix('cred');
                dataSourceName = objectGenerator.generateNameByPrefix('ds');
                username = objectGenerator.generateNameByPrefix('user');
                password = objectGenerator.generateNameByPrefix('psswd');
                trainingName = objectGenerator.generateShortNameByPrefix('training');
            });

            test('T5013106: test secret data source - username & password credentials - project scope', async () => {
                videoName = 'username-password-credentials';
                const mountPath = '/mnt/data';

                await test.step('create username & password secret', async () => {
                    await credentialsPage.createUserPassCredential(credentialName, username, password, {
                        scope: Scope.Project,
                        scopeEntityName: project.name
                    });
                });

                await test.step('create data source using username & password secret', async () => {
                    await dataSourcePage.createSecretDataSource(
                        dataSourceName,
                        Scope.Project,
                        project.name,
                        credentialName,
                        mountPath
                    );
                });

                await test.step('create training workload with the data source', async () => {
                    await trainingPage.createFlexibleTraining(projectName, trainingName, {
                        dsName: dataSourceName,
                        containerPath: mountPath
                    });
                    await trainingPage.waitForWorkloadUIStatus(trainingName, EWorkloadStatus.Running);
                });

                await test.step('verify the secrect in the pod', async () => {
                    const pods = await kubeCli.listPodsOfJob(namespace, trainingName);
                    for (const pod of pods) {
                        const getUser = bashCli.exec(
                            `kubectl exec -it ${pod.metadata.name} -n ${namespace} -- cat ${mountPath}/User`
                        );
                        const getPasswd = bashCli.exec(
                            `kubectl exec -it ${pod.metadata.name} -n ${namespace} -- cat ${mountPath}/Password`
                        );
                        logger.info(
                            `cat ${mountPath}/User: ${getUser}\ncat ${mountPath}/Password: ${getPasswd}`
                        );
                        expect(getUser).toEqual(username);
                        expect(getPasswd).toEqual(password);
                    }
                });

                await test.step('doing cleanup', async () => {
                    await trainingPage.deleteWorkload(trainingName);
                    await deleteDataSourceByName(dataSourceName);
                    await credentialsPage.goToPage();
                    await credentialsPage.deleteCredential(credentialName);
                });
            });
        });
    });
});
