/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Project } from '../../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    Utils,
    flowRootState,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    waitUtils,
    projectUtils,
    apiUtils,
    environmentUtils,
    settings,
    unleashClient
} from '../../../../../../../modules/helpers';
import { AuthPage } from '../../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { BrowserEngine } from '../../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import {
    DISTRIBUTED_TOOLS_LABELS,
    TrainingPage
} from '../../../../../../../modules/ui/runaiV2/pages/TrainingPage';
import { WorkloadPage, EWorkloadStatus } from '../../../../../../../modules/ui/runaiV2/pages/WorkloadPage';
import { ComputeResourcePage } from '../../../../../../../modules/ui/runaiV2/pages/ComputeResourcePage';
import { EnvironmentPage } from '../../../../../../../modules/ui/runaiV2/pages/EnvironmentPage';
import { WorkloadType } from '../../../../../../../modules/helpers/config/types';
import { WorkspacePage } from '../../../../../../../modules/ui/runaiV2/pages/WorkspacePage';
import { expect } from 'chai';
import { Response } from 'playwright';
import {
    EnvironmentAsset,
    EnvironmentCreationRequest,
    Scope,
    WorkloadSupportedTypes
} from '../../../../../../../modules/api/apis/assetsApi';
import { shouldRunTests } from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { DataSourcePage } from '../../../../../../../modules/ui/runaiV2/pages/DataSourcePage';
import { test } from '@playwright/test';
import { EFeatureToggleKeys } from '../../../../../../../modules/helpers/utils/Unleash/featureToggle';

type TSupportedWorkloadTypes = WorkloadType.training | WorkloadType.distributed | WorkloadType.workspace;

const workloadSubmissionUrlsMap: Record<TSupportedWorkloadTypes, string> = {
    [WorkloadType.training]: '/api/v2/training',
    [WorkloadType.distributed]: '/api/v2/training',
    [WorkloadType.workspace]: '/api/v2/workspace'
};

test.describe(`Data source overrides during workload submission, tags: ${tags.tool.ui} ${tags.resource.workloads} ${tags.teams.viper} `, () => {
    shouldRunTests(settings.isSelfHosted)(function () {
        // infra
        let browserEngine: BrowserEngine;
        let auth: AuthPage;
        // pages
        let training: TrainingPage;
        let workloads: WorkloadPage;
        let workspace: WorkspacePage;
        let environment: EnvironmentPage;
        let computeResource: ComputeResourcePage;
        let dataSource: DataSourcePage;

        let videoName: string;
        let projectName: string;
        let namespace: string;
        let project: Project;
        let distributedEnvironment: EnvironmentAsset;
        let jupyterPresetName: string;

        async function createDistributedEnvironment(projectId: number): Promise<EnvironmentAsset> {
            const environmentName = objectGenerator.generateNameByPrefix('e2e-ds-overrides');
            const environmentCreationRequest: EnvironmentCreationRequest = {
                meta: {
                    name: environmentName,
                    scope: Scope.Project,
                    projectId,
                    workloadSupportedTypes: {
                        training: true,
                        distributed: true,
                        distFramework: WorkloadSupportedTypes.DistFrameworkEnum.PyTorch
                    }
                },
                spec: {
                    image: 'alpine',
                    command: 'sleep infinity'
                }
            };
            return await environmentUtils.create(environmentCreationRequest);
        }

        async function fillTrainingForm(name: string): Promise<void> {
            await workloads.openTrainingForm();
            await workloads.workloadFormSelectProject(projectName);
            await training.fillTrainingName(name);
            await training.waitForTrainingNameCount(name);
            await training.selectAssetsSubmissionOption();
            await training.clickOnContinueBtn();

            await environment.selectEnvironment(jupyterPresetName);
            await computeResource.selectComputeResource('cpu-only');
        }

        async function fillDistributedForm(name: string): Promise<void> {
            await workloads.openTrainingForm();
            await workloads.workloadFormSelectProject(projectName);

            await waitUtils.sleep(1000);
            await training.selectedMultiNodeOption();
            await training.selectedDistributedTool(DISTRIBUTED_TOOLS_LABELS.PY_TORCH);
            await training.selectedWorkersOnlyOption();

            await training.fillTrainingName(name);
            await training.waitForTrainingNameCount(name);
            await training.selectAssetsSubmissionOption();
            await training.clickOnContinueBtn();

            await environment.selectEnvironment(distributedEnvironment.meta.name);
            await computeResource.selectComputeResource('cpu-only');
        }

        async function fillWorkspaceForm(name: string): Promise<void> {
            await workloads.openWorkspaceForm();
            await workloads.workloadFormSelectProject(projectName);
            await workspace.fillWorkspaceName(name);
            await workspace.waitForWorkspaceNameCount(name);
            await workspace.selectAssetsSubmissionOption();
            await workspace.clickOnContinue();

            await environment.selectEnvironment(jupyterPresetName);
            await computeResource.selectComputeResource('cpu-only');
        }

        async function submitWorkload(workloadType: WorkloadType): Promise<void> {
            switch (workloadType) {
                case WorkloadType.workspace:
                    await workspace.clickOnCreateWorkspaceBtn();
                    break;
                case WorkloadType.training:
                    await training.clickOnSubmitBtn();
                    break;
                case WorkloadType.distributed:
                    await training.clickOnSubmitBtn();
                    break;
            }
        }

        async function fillWorkloadForm(
            workloadType: WorkloadType,
            name: string,
            navigateToPage = true
        ): Promise<void> {
            logger.debug(`creating workload of type: ${workloadType}, name: ${name}`);
            if (navigateToPage) await workloads.navigateToWorkloads();

            switch (workloadType) {
                case WorkloadType.workspace:
                    await fillWorkspaceForm(name);
                    break;
                case WorkloadType.training:
                    await fillTrainingForm(name);
                    break;
                case WorkloadType.distributed:
                    await fillDistributedForm(name);
                    break;
                default:
                    logger.error(`Unsupported workload type: ${workloadType} for the test`);
                    return;
            }

            logger.debug(`Successfully created workload of type: ${workloadType}, name: ${name}`);
        }

        async function validateResponsePayload(url: string, containerPath: string): Promise<void> {
            const response: Response = await browserEngine.page.waitForResponse((resp) => {
                return resp.url().includes(url);
            });

            expect(response.status()).to.equal(200);
            expect(response.request().method()).to.equal('POST');

            const responseBody = await response.json();
            expect(responseBody.assets.datasources.length).to.equal(1);
            expect(responseBody.assets.datasources[0].overrides.containerPath).to.equal(containerPath);
        }

        async function cleanUpCreatedAssets(): Promise<void> {
            try {
                await environmentUtils.delete(distributedEnvironment.meta.id);
                await environmentUtils.waitForEnvironmentToBeDeleted(distributedEnvironment.meta.name);
                logger.info('Clean up created assets completed');
            } catch {
                logger.error('Failed to clean up created assets');
            }
        }

        async function initBrowserAndLogin(): Promise<void> {
            await apiUtils.setUIWorkloadV2SubmissionFeature(true);
            await browserEngine.start();
            auth = new AuthPage(browserEngine.currentPage);
            training = new TrainingPage(browserEngine.currentPage);
            workspace = new WorkspacePage(browserEngine.currentPage);
            workloads = new WorkloadPage(browserEngine.currentPage);
            computeResource = new ComputeResourcePage(browserEngine.currentPage);
            environment = new EnvironmentPage(browserEngine.currentPage);
            dataSource = new DataSourcePage(browserEngine.currentPage);

            await auth.signIn(
                flowRootState.users.system_administrator.email,
                flowRootState.users.system_administrator.password
            );
        }

        test.beforeAll(async function () {
            jupyterPresetName = (await environmentUtils.presetNames()).jupyter;
            browserEngine = new BrowserEngine();
            projectName = objectGenerator.generateShortNameByPrefix('project-e2e-ds-overrides');
            namespace = Utils.namespaceFromProjectName(projectName);
            project = await projectResourcesUtils.createProject(namespace, {
                name: projectName,
                deservedGpus: 4,
                maxAllowedGpus: 4
            });

            distributedEnvironment = await createDistributedEnvironment(project.id);
            const roleBindingName = projectUtils.getSystemAdministratorRoleBindingName();

            await projectUtils.waitForRoleBindingForUserInProject(
                namespace,
                flowRootState.users.system_administrator.email,
                roleBindingName,
                true
            );
            await unleashClient.disableFeature(EFeatureToggleKeys.SHOW_HYBRID_SUBMISSION);
            await apiUtils.disableHybridSubmission();

            await initBrowserAndLogin();
        });

        test.afterAll('deleting project and namespace', async ({}, testInfo) => {
            await browserEngine.stopAndArchive(testInfo, 'data-source-overrides-' + videoName);
            await unleashClient.enableFeature(EFeatureToggleKeys.SHOW_HYBRID_SUBMISSION);
            await apiUtils.enableHybridSubmission();
            try {
                await apiUtils.setUIWorkloadV2SubmissionFeature(true);
                await cleanUpCreatedAssets();
                await operations.deleteProjectAndNamespace(project.id);
                logger.info(`deleted project ${projectName} an namespace ${namespace}`);
            } catch (error) {
                logger.error(`failed to delete project ${projectName} an namespace ${namespace}`);
                throw error;
            }
        });

        [WorkloadType.workspace, WorkloadType.training, WorkloadType.distributed].forEach(function (
            workloadType: WorkloadType
        ) {
            test(`Submit a ${workloadType} with data source and override its container path`, async function () {
                const workloadName: string = objectGenerator.generateNameByPrefix(
                    `${workloadType.toLocaleLowerCase()}-v2`
                );

                const dataSourceName = objectGenerator.generateNameByPrefix('hp-override');

                await test.step(`Go to the ${workloadType} page and fill it`, async function () {
                    await fillWorkloadForm(workloadType, workloadName);
                });

                await test.step(`Create a data source directly from the ${workloadType} page`, async function () {
                    await training.addNewHostPath();
                    await dataSource.fillHostPathCreationForm(dataSourceName);
                });

                await test.step(`Override and container path of the created data source in the ${workloadType} page, and submit the workload`, async function () {
                    await training.fillContainerPathOverrideInput('/hp-container-path-override');
                    await submitWorkload(workloadType);
                });

                await test.step(`Validate ${workloadType} response`, async () => {
                    await validateResponsePayload(
                        workloadSubmissionUrlsMap[workloadType],
                        '/hp-container-path-override'
                    );
                });

                await test.step(`Wait for ${workloadType} workload status and delete the workload`, async () => {
                    await workloads.waitForWorkloadUIStatus(workloadName, EWorkloadStatus.Running);
                    await workloads.deleteWorkload(workloadName);
                });
            });
        });
    });
});
