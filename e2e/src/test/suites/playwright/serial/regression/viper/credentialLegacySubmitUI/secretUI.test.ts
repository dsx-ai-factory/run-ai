/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Project } from '../../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    dataSourcesUtils,
    environmentUtils,
    flowRootState,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    projectUtils,
    tags,
    unleashClient,
    apiUtils,
    Utils
} from '../../../../../../../modules/helpers';
import { AllureUtils } from '../../../../../../../modules/helpers/utils/AllureUtils';
import { AuthPage } from '../../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { BrowserEngine } from '../../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { CredentialPage } from '../../../../../../../modules/ui/runaiV2/pages/CredentialPage';
import { DataSourcePage } from '../../../../../../../modules/ui/runaiV2/pages/DataSourcePage';
import { WorkspacePage } from '../../../../../../../modules/ui/runaiV2/pages/WorkspacePage';
import {
    EnvironmentAsset,
    EnvironmentCreationRequest,
    Scope
} from '../../../../../../../modules/api/apis/assetsApi';
import { EWorkloadStatus, WorkloadPage } from '../../../../../../../modules/ui/runaiV2/pages/WorkloadPage';
import { test } from '@playwright/test';
import { EFeatureToggleKeys } from '../../../../../../../modules/helpers/utils/Unleash/featureToggle';

test.describe(`Secret as a Datasource UI, tags: ${tags.tool.ui} ${tags.resource.credential} ${tags.teams.viper}`, function () {
    let browserEngine: BrowserEngine;
    let auth: AuthPage;
    let credential: CredentialPage;
    let dataSource: DataSourcePage;
    let workspace: WorkspacePage;
    let workload: WorkloadPage;

    let videoName: string;
    let projectName: string;
    let namespace: string;
    let project: Project;
    let allureUtils: AllureUtils;

    let credentialName: string;
    let credentialId: string;
    let credentialSecret: string;
    let dataSourceName: string;
    let workspaceName: string;

    async function initBrowserAndLogin(): Promise<void> {
        await browserEngine.start();
        auth = new AuthPage(browserEngine.currentPage);
        credential = new CredentialPage(browserEngine.currentPage);
        dataSource = new DataSourcePage(browserEngine.currentPage);
        workspace = new WorkspacePage(browserEngine.currentPage);
        workload = new WorkloadPage(browserEngine.currentPage);
        await auth.signIn(
            flowRootState.users.system_administrator.email,
            flowRootState.users.system_administrator.password
        );
    }

    test.beforeAll(async function () {
        browserEngine = new BrowserEngine();
        allureUtils = new AllureUtils();
        projectName = objectGenerator.generateNameByPrefix('e2e-project-cred');
        namespace = Utils.namespaceFromProjectName(projectName);

        try {
            project = await projectResourcesUtils.createProject(namespace, {
                name: projectName,
                deservedGpus: 1,
                maxAllowedGpus: 1
            });
            projectName = project.name;
            namespace = project.namespace;
        } catch (error) {
            logger.error(`Failed to create project: ${error.message}`);
            logger.error(`Error details: ${JSON.stringify(error, null, 2)}`);

            // Attach error details to Allure report
            const { attachment } = await import('allure-js-commons');
            await attachment(
                'Project Creation Error Details',
                JSON.stringify(
                    {
                        message: error.message,
                        statusCode: error.statusCode,
                        url: error.options?.uri || error.options?.url,
                        method: error.options?.method,
                        response: error.body,
                        timestamp: new Date().toISOString()
                    },
                    null,
                    2
                ),
                'application/json'
            );

            // Collect backend logs for 502 debugging
            logger.info('Collecting backend service logs for 502 debugging...');
            await allureUtils.attachSystemComponentLogs(
                'org-unit-service',
                'Org-Unit Service Logs',
                'runai-backend',
                1000
            );
            await allureUtils.attachSystemComponentLogs(
                'cluster-service',
                'Cluster Service Logs',
                'runai-backend',
                500
            );
            await allureUtils.attachSystemComponentLogs(
                'backend-nginx',
                'API Gateway Logs',
                'runai-backend',
                500
            );
            throw error;
        }

        await unleashClient.disableFeature(EFeatureToggleKeys.SHOW_HYBRID_SUBMISSION);
        await apiUtils.disableHybridSubmission();
        try {
            await projectUtils.waitForRoleBindingForUserInProject(
                namespace,
                flowRootState.users.system_administrator.email,
                projectUtils.newRbacRunaiSystemAdministratorRoleBindingName,
                true
            );
        } catch (e) {
            logger.error(e);
        }
    });

    test.afterAll('deleting project and namespace', async function () {
        try {
            await operations.deleteProjectAndNamespace(project.id);
            await unleashClient.enableFeature(EFeatureToggleKeys.SHOW_HYBRID_SUBMISSION);
            await apiUtils.enableHybridSubmission();
        } catch (error) {
            logger.error(
                `failed to delete project ${projectName} and namespace ${namespace}: ${error.message}`
            );
            throw error;
        }
    });
    test.beforeEach(async () => {
        credentialId = objectGenerator.generateNameByPrefix('ak-id');
        credentialSecret = objectGenerator.generateNameByPrefix('ak-secret');
        dataSourceName = objectGenerator.generateNameByPrefix('data-source');
        workspaceName = objectGenerator.generateNameByPrefix('workspace');
        credentialName = objectGenerator.generateNameByPrefix('ak');
        await initBrowserAndLogin();
    });

    test.afterEach('stop browser', async function ({}, testInfo) {
        // Collect logs if test failed
        if (testInfo.status === 'failed' || testInfo.status === 'timedOut') {
            try {
                // Collect backend service logs
                await allureUtils.attachSystemComponentLogs(
                    'org-unit-service',
                    'Org-Unit Service Logs',
                    'runai-backend',
                    1000
                );
                await allureUtils.attachSystemComponentLogs(
                    'cluster-service',
                    'Cluster Service Logs',
                    'runai-backend',
                    500
                );

                // Collect workspace pod logs if workspace was created
                if (workspaceName && namespace) {
                    allureUtils.attachPodsLogs([workspaceName], namespace, 500);
                }
            } catch (error) {
                logger.warn(`Failed to collect failure logs: ${error.message}`);
            }
        }

        await browserEngine.stopAndArchive(testInfo, 'credential-' + videoName);
    });

    async function createWorkspaceWithDataSource(
        workspaceName: string,
        environmentName: string,
        computeName: string,
        dataSourceName: string
    ): Promise<void> {
        await workload.navigateToWorkloads();
        await workload.openWorkspaceForm();
        await workload.workloadFormSelectProject(projectName);
        await workspace.fillWorkspaceName(workspaceName);
        await workspace.waitForWorkspaceNameCount(workspaceName);
        await workspace.selectAssetsSubmissionOption();
        await workspace.clickOnContinue();
        await workspace.selectEnvironment(environmentName);
        await workspace.selectComputeResource(computeName);
        await workspace.openDataSourceSection();
        await workspace.selectDataSourceName(dataSourceName);
        await workspace.clickOnCreateWorkspaceBtn();
    }

    //async function deleteDataSourceByName(name: string): Promise<void> {
    async function deleteDataSourceByName(name: string): Promise<void> {
        await dataSource.goToPage();
        await dataSource.deleteDataSource(name);
        await dataSourcesUtils.waitForDataSourceToBeDeleted(name);
        logger.info(`Data source deleted successfully: ${name}`);
    }

    async function createEnvironment(projectId: number): Promise<EnvironmentAsset> {
        // Creating a light environment to reduce the time of the workload to get ready.
        const environmentName = objectGenerator.generateNameByPrefix('e2e-test-environment');
        const environmentCreationRequest: EnvironmentCreationRequest = {
            meta: {
                name: environmentName,
                scope: Scope.Project,
                projectId,
                workloadSupportedTypes: {
                    workspace: true,
                    training: true
                }
            },
            spec: {
                image: 'alpine',
                command: 'sleep infinity'
            }
        };
        return await environmentUtils.create(environmentCreationRequest);
    }

    test('create secret with access key credential - project scope', async () => {
        videoName = 'create-secret-with-credential';
        const environmentAsset = await createEnvironment(project.id);

        await credential.createAccessKeyCredential(credentialName, credentialId, credentialSecret, {
            scope: Scope.Project,
            scopeEntityName: project.name
        });
        // Creating Secret with the new credential
        await dataSource.createSecretDataSource(
            dataSourceName,
            Scope.Project,
            project.name,
            credentialName,
            '/tmp/'
        );
        // Wait for datasource to be ready
        await dataSourcesUtils.validateExpectedStatus(dataSourceName, 'No issues found');

        // Creating workspace with the new data source using the new credential
        await createWorkspaceWithDataSource(
            workspaceName,
            environmentAsset.meta.name,
            'cpu-only',
            dataSourceName
        );
        await workload.waitForWorkloadUIStatus(workspaceName, EWorkloadStatus.Running);
        logger.debug('workspace created successfully');

        // cleanup
        await workload.deleteWorkload(workspaceName);
        await environmentUtils.delete(environmentAsset.meta.id);
        await deleteDataSourceByName(dataSourceName);
        await credential.goToPage();
        await credential.deleteCredential(credentialName);
    });
});
