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
    projectUtils,
    environmentUtils,
    settings
} from '../../../../../../../modules/helpers';
import { AuthPage } from '../../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { BrowserEngine } from '../../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import {
    DISTRIBUTED_TOOLS_LABELS,
    TrainingPage
} from '../../../../../../../modules/ui/runaiV2/pages/TrainingPage';
import { WorkloadPage, EWorkloadStatus } from '../../../../../../../modules/ui/runaiV2/pages/WorkloadPage';
import { WorkloadType } from '../../../../../../../modules/helpers/config/types';
import {
    HybridForm,
    EDataSourceOption
} from '../../../../../../../modules/ui/runaiV2/pages/HybridSubmissionForm';
import { WorkloadType as FormType } from '../../../../../../../modules/api/apis/assetsApi';
import { expect } from 'chai';
import {
    EnvironmentAsset,
    EnvironmentCreationRequest,
    Scope,
    WorkloadSupportedTypes
} from '../../../../../../../modules/api/apis/assetsApi';
import { shouldRunTests } from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { DataSourcePage } from '../../../../../../../modules/ui/runaiV2/pages/DataSourcePage';
import { test } from '@playwright/test';

type TSupportedWorkloadTypes = WorkloadType.training | WorkloadType.distributed | WorkloadType.workspace;

const workloadSubmissionUrlsMap: Record<TSupportedWorkloadTypes, string> = {
    [WorkloadType.training]: '/api/v1/workloads/trainings',
    [WorkloadType.distributed]: '/api/v1/workloads/distributed',
    [WorkloadType.workspace]: '/api/v1/workloads/workspaces'
};

test.describe(`Data source overrides during workload submission, tags: ${tags.tool.ui} ${tags.resource.workloads} ${tags.teams.viper} `, () => {
    shouldRunTests(settings.isSelfHosted)(function () {
        // infra
        let browserEngine: BrowserEngine;
        let auth: AuthPage;
        // pages
        let training: TrainingPage;
        let workloads: WorkloadPage;
        let hybridForm: HybridForm;
        let dataSource: DataSourcePage;

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

        async function fillWorkloadForm(workloadType: TSupportedWorkloadTypes, name: string): Promise<void> {
            const isWorkspace = workloadType === WorkloadType.workspace;
            await workloads.navigateToWorkloads();
            if (isWorkspace) await workloads.openWorkspaceForm();
            else await workloads.openTrainingForm();
            await workloads.workloadFormSelectProject(projectName);
            if (workloadType === WorkloadType.distributed) {
                await training.selectedMultiNodeOption();
                await training.selectedDistributedTool(DISTRIBUTED_TOOLS_LABELS.PY_TORCH);
                await training.selectedWorkersOnlyOption();
            }
            await workloads.workloadFormSelectTemplate();
            if (isWorkspace) await hybridForm.fillWorkspaceName(name);
            else await hybridForm.fillTrainingName(name);
            await hybridForm.waitForWorkloadNameCount(name);
            await hybridForm.clickOnContinue(isWorkspace ? FormType.Workspace : FormType.Training);
            const environmentName =
                workloadType === WorkloadType.distributed
                    ? distributedEnvironment.meta.name
                    : jupyterPresetName;
            await hybridForm.fillHybridFormByLoadingExistingConfig(environmentName, 'cpu-only');
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
            await browserEngine.start();
            auth = new AuthPage(browserEngine.currentPage);
            training = new TrainingPage(browserEngine.currentPage);
            hybridForm = new HybridForm(browserEngine.currentPage);
            workloads = new WorkloadPage(browserEngine.currentPage);
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

            await initBrowserAndLogin();
        });

        test.afterAll('deleting project and namespace', async ({}, testInfo) => {
            await browserEngine.stopAndArchive(testInfo, 'data-source-overrides');
            try {
                await cleanUpCreatedAssets();
                await operations.deleteProjectAndNamespace(project.id);
                logger.info(`deleted project ${projectName} an namespace ${namespace}`);
            } catch (error) {
                logger.error(`failed to delete project ${projectName} an namespace ${namespace}`);
                throw error;
            }
        });

        ([WorkloadType.workspace, WorkloadType.training, WorkloadType.distributed] as const).forEach(
            function (workloadType: TSupportedWorkloadTypes) {
                test(`Submit a ${workloadType} with data source and override its container path`, async function () {
                    const workloadName: string = objectGenerator.generateNameByPrefix(
                        `${workloadType.toLocaleLowerCase()}-v2`
                    );

                    const dataSourceName = objectGenerator.generateNameByPrefix('hp-override');

                    await test.step(`Go to the ${workloadType} page and fill it`, async function () {
                        await fillWorkloadForm(workloadType, workloadName);
                    });

                    await test.step(`Create a data source directly from the ${workloadType} page`, async function () {
                        await hybridForm.openDataSourcesSection();
                        await hybridForm.openDataSourcesDrawer();
                        await hybridForm.clickOnAddNewDropdownBtnInDrawer();
                        await hybridForm.selectDataSourceInDrawer(EDataSourceOption.HostPath);
                        await dataSource.fillHostPathCreationForm(dataSourceName);
                        await hybridForm.clickOnDrawerItemByLabel(dataSourceName, false);
                        await hybridForm.clickOnLoadFromDrawer();
                    });

                    await test.step(`Override the mount path and submit the ${workloadType}`, async () => {
                        const mountPath = '/hp-container-path-override';
                        await hybridForm.fillContainerPathOverrideInput(mountPath);
                        const [response] = await Promise.all([
                            browserEngine.page.waitForResponse(
                                (resp) =>
                                    new URL(resp.url()).pathname ===
                                        workloadSubmissionUrlsMap[workloadType] &&
                                    resp.request().method() === 'POST'
                            ),
                            hybridForm.clickOnSubmit()
                        ]);
                        expect(response.status()).to.equal(202);
                        const workload = await response.json();
                        expect(workload.spec.storage.hostPath).to.have.lengthOf(1);
                        expect(workload.spec.storage.hostPath[0].mountPath).to.equal(mountPath);
                    });

                    await test.step(`Wait for ${workloadType} workload status and delete the workload`, async () => {
                        await workloads.waitForWorkloadUIStatus(workloadName, EWorkloadStatus.Running);
                        await workloads.deleteWorkload(workloadName);
                    });
                });
            }
        );
    });
});
