/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { WorkloadPage } from '../../../../../../modules/ui/runaiV2/pages/WorkloadPage';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import {
    flowRootState,
    objectGenerator,
    logger,
    tags,
    projectResourcesUtils,
    operations,
    unleashClient,
    environmentUtils
} from '../../../../../../modules/helpers';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { test } from '@playwright/test';
import { HybridForm } from '../../../../../../modules/ui/runaiV2/pages/HybridSubmissionForm';
import {
    WorkloadType,
    EnvironmentAsset,
    WorkloadSupportedTypes
} from '../../../../../../modules/api/apis/assetsApi';
import { EFeatureToggleKeys } from '../../../../../../modules/helpers/utils/Unleash/featureToggle';
import {
    DISTRIBUTED_TOOLS_LABELS,
    TrainingPage
} from '../../../../../../modules/ui/runaiV2/pages/TrainingPage';

const testParams = [
    {
        framework: WorkloadSupportedTypes.DistFrameworkEnum.PyTorch,
        feFramework: DISTRIBUTED_TOOLS_LABELS.PY_TORCH,
        runWorkerOnly: true
    },
    {
        framework: WorkloadSupportedTypes.DistFrameworkEnum.Tf,
        feFramework: DISTRIBUTED_TOOLS_LABELS.TENSOR_FLOW,
        runWorkerOnly: true
    },
    {
        framework: WorkloadSupportedTypes.DistFrameworkEnum.Mpi,
        feFramework: DISTRIBUTED_TOOLS_LABELS.MPI,
        runWorkerOnly: false
    },
    {
        framework: WorkloadSupportedTypes.DistFrameworkEnum.XgBoost,
        feFramework: DISTRIBUTED_TOOLS_LABELS.XG_BOOST,
        runWorkerOnly: false
    }
];

test.describe(`UI sanity hybrid distributed submission, tags: ${tags.tool.ui} ${tags.resource.workloads} ${tags.teams.viper}`, function () {
    let project: Project;
    let videoName: string;

    let hybridForm: HybridForm;
    let browserEngine: BrowserEngine;
    let auth: AuthPage;
    let workload: WorkloadPage;
    let training: TrainingPage;
    let workloadName: string;

    let firstEnv: EnvironmentAsset;
    let secondEnv: EnvironmentAsset;

    async function cleanUpCreatedAssets(): Promise<void> {
        try {
            await Promise.all([
                environmentUtils.delete(firstEnv.meta.id),
                environmentUtils.delete(firstEnv.meta.id),
                environmentUtils.waitForEnvironmentToBeDeleted(firstEnv.meta.name),
                environmentUtils.waitForEnvironmentToBeDeleted(firstEnv.meta.name)
            ]);
            logger.info('Clean up created assets completed');
        } catch {
            logger.error('Failed to clean up created assets');
        }
    }

    async function cleanUpProjectAndNamespace(): Promise<void> {
        try {
            await operations.deleteProjectAndNamespace(project.id);
            logger.info(`deleted project ${project.name} an namespace ${project.namespace}`);
        } catch (error) {
            logger.error(`failed to delete project ${project.name} an namespace ${project.namespace}`);
            throw error;
        }
    }

    async function createEnvironment(
        projectId: number,
        framework: WorkloadSupportedTypes.DistFrameworkEnum
    ): Promise<EnvironmentAsset> {
        try {
            return await environmentUtils.createEnvironmentWithArbitraryFields(
                projectId,
                undefined,
                framework
            );
        } catch (e) {
            logger.error(`failed to create environment ,error: ${e}`);
            throw e;
        }
    }

    async function mountBrowserAndLogin(): Promise<void> {
        await browserEngine.start();
        auth = new AuthPage(browserEngine.currentPage);
        hybridForm = new HybridForm(browserEngine.currentPage);
        workload = new WorkloadPage(browserEngine.currentPage);
        training = new TrainingPage(browserEngine.currentPage);

        await auth.signIn(
            flowRootState.users.l2_researcher.email,
            flowRootState.users.l2_researcher.password
        );
    }

    async function fillWorkloadCreationFirstPage(
        framework: DISTRIBUTED_TOOLS_LABELS,
        workersOnly = false
    ): Promise<void> {
        await workload.openTrainingForm();

        await workload.selectClusterFromDropDown(flowRootState.cluster.name);
        await workload.workloadFormSelectProject(project.name);

        await training.page.waitForTimeout(1000);
        await training.selectedMultiNodeOption();
        await workload.workloadFormSelectTemplate();
        await training.selectedDistributedTool(framework);
        if (workersOnly) await training.selectedWorkersOnlyOption();

        await hybridForm.fillTrainingName(workloadName);
        await hybridForm.waitForWorkloadNameCount(workloadName);
        await hybridForm.clickOnContinue(WorkloadType.Training);
    }

    async function fillReplicaForm(opts?: {
        image?: string;
        imagePullPolicy?: string;
        toolName?: string;
        command?: string;
        args?: string;
    }): Promise<void> {
        await hybridForm.fillImageInput(opts?.image || 'jupyter/scipy-notebook');

        await hybridForm.clickOnSubSectionByLabel(hybridForm.imagePullSection);
        await hybridForm.clickOnRadioOptionByLabel(
            opts?.imagePullPolicy || 'Always pull the image from the registry'
        );

        await hybridForm.clickOnSubSectionByLabel('Tools');
        await hybridForm.addToolTypeByLabel(opts?.toolName || 'Jupyter');

        await hybridForm.clickOnSubSectionByLabel('Runtime settings');
        await hybridForm.clickOnCommandAndArguments();
        await hybridForm.fillCmdInput(opts?.command || 'start-notebook.sh');
        await hybridForm.fillArgsInput(
            opts?.args || '--NotebookApp.base_url=/${RUNAI_PROJECT}/${RUNAI_JOB_NAME} --NotebookApp.token=""'
        );

        await hybridForm.clickOnSubSectionByLabel('Security');
    }

    test.beforeAll('creating project and namespace', async function () {
        browserEngine = new BrowserEngine();
        project = await projectResourcesUtils.createProjectAndNamespace('h-p');
        logger.info(`created project ${project.name} ,namespace: "${project.namespace}"`);
    });

    test.afterAll('deleting project and namespace', async function () {
        await cleanUpProjectAndNamespace();
    });

    test.beforeEach('start browser', async function () {
        await mountBrowserAndLogin();
        workloadName = objectGenerator.generateShortNameByPrefix('hybrid');
        await workload.navigateToWorkloads();
    });

    test.afterEach('stop browser', async ({}, testInfo) => {
        try {
            await browserEngine.stopAndArchive(testInfo, videoName);
        } catch (e) {
            logger.error(e);
        }
    });

    testParams.forEach((params) => {
        test.describe(`${params.framework}`, function () {
            test.beforeAll('creating environments', async function () {
                firstEnv = await createEnvironment(project.id, params.framework);
                secondEnv = await createEnvironment(project.id, params.framework);
            });

            test.afterAll('cleaning up created assets', async function () {
                await cleanUpCreatedAssets();
            });

            test.describe(`${params.framework} master same as worker request`, function () {
                test(`${params.framework} create from scratch`, async function () {
                    videoName = `${params.framework}-hybrid-dist-scratch-same-as-worker`;
                    await fillWorkloadCreationFirstPage(params.feFramework);

                    await fillReplicaForm();

                    // continue to master spec
                    await hybridForm.clickOnSubmit();
                    // submit as is without changing master
                    await hybridForm.waitForDistributedMasterSubmitButton();
                    await hybridForm.clickOnSubmitMasterForm();
                    await training.waitForSubmitButton();
                    await hybridForm.clickOnSubmit(); // setup step

                    await workload.waitForWorkloadUIStatus(workloadName);
                    await workload.deleteWorkload(workloadName);
                });

                test(`${params.framework} create from scratch using asset templates`, async function () {
                    videoName = `${params.framework}-hybrid-dist-templates-same-as-worker`;

                    await fillWorkloadCreationFirstPage(params.feFramework);

                    await hybridForm.openEnvironmentsDrawer();
                    await hybridForm.clickOnDrawerItemByLabel(firstEnv.meta.name);

                    await hybridForm.openComputesDrawer();
                    await hybridForm.clickOnDrawerItemByLabel('cpu-only');

                    // continue to master spec
                    await hybridForm.clickOnSubmit();
                    // submit as is without changing master
                    await hybridForm.waitForDistributedMasterSubmitButton();
                    await hybridForm.clickOnSubmitMasterForm();
                    await training.waitForSubmitButton();
                    await hybridForm.clickOnSubmit(); // setup step

                    await workload.waitForWorkloadUIStatus(workloadName);
                    await workload.deleteWorkload(workloadName);
                });
            });

            test.describe(`${params.framework} different master and worker configurations`, function () {
                test('create from scratch', async function () {
                    videoName = `${params.framework}-hybrid-dist-scratch-different-master`;
                    await fillWorkloadCreationFirstPage(params.feFramework);

                    await fillReplicaForm();

                    await hybridForm.clickOnSubmit();

                    await hybridForm.page.waitForTimeout(1500);

                    await hybridForm.fillImageInput('alpine');
                    await hybridForm.clickOnSubSectionByLabel('Runtime settings');
                    await hybridForm.fillCmdInput('sleep 100');
                    await hybridForm.fillArgsInput('');

                    await hybridForm.waitForDistributedMasterSubmitButton();
                    await hybridForm.clickOnSubmitMasterForm();
                    await training.waitForSubmitButton();
                    await hybridForm.clickOnSubmit(); // setup step

                    await workload.waitForWorkloadUIStatus(workloadName);
                    await workload.deleteWorkload(workloadName);
                });

                test(`${params.framework} create from scratch using asset templates`, async function () {
                    videoName = `${params.framework}-hybrid-dist-templates-different-master`;

                    await fillWorkloadCreationFirstPage(params.feFramework);

                    await hybridForm.openEnvironmentsDrawer();
                    await hybridForm.clickOnDrawerItemByLabel(firstEnv.meta.name);

                    await hybridForm.openComputesDrawer();
                    await hybridForm.clickOnDrawerItemByLabel('one-gpu');

                    await hybridForm.clickOnSubmit();

                    await hybridForm.page.waitForTimeout(1500);

                    await hybridForm.openEnvironmentsDrawer();
                    await hybridForm.clickOnDrawerItemByLabel(secondEnv.meta.name);

                    await hybridForm.openComputesDrawer();
                    await hybridForm.clickOnDrawerItemByLabel('cpu-only');

                    await hybridForm.waitForDistributedMasterSubmitButton();
                    await hybridForm.clickOnSubmitMasterForm();
                    await training.waitForSubmitButton();
                    await hybridForm.clickOnSubmit(); // setup step

                    await workload.waitForWorkloadUIStatus(workloadName);
                    await workload.deleteWorkload(workloadName);
                });
            });

            test.describe(`${params.framework} no master step and copy and edit`, function () {
                if (params.runWorkerOnly) {
                    test(`${params.framework} create from scratch - no master step`, async function () {
                        test.setTimeout(210 * 1000);

                        videoName = `${params.framework} hybrid-dist-scratch-no-master-step`;
                        await fillWorkloadCreationFirstPage(params.feFramework, true);

                        await fillReplicaForm();

                        // no master step, submitting worker form and then setup step
                        await hybridForm.clickOnSubmit();
                        await training.waitForSubmitButton();
                        await hybridForm.clickOnSubmit(); // setup step

                        await workload.waitForWorkloadUIStatus(workloadName);
                        await workload.deleteWorkload(workloadName);
                    });
                }

                test(`${params.framework} copy and edit using assets templates`, async function () {
                    test.setTimeout(260 * 1000);
                    videoName = `${params.framework} hybrid-copy-and-edit`;

                    // create a workload
                    await fillWorkloadCreationFirstPage(params.feFramework);

                    await hybridForm.openEnvironmentsDrawer();
                    await hybridForm.clickOnDrawerItemByLabel(firstEnv.meta.name);

                    await hybridForm.openComputesDrawer();
                    await hybridForm.clickOnDrawerItemByLabel('cpu-only');

                    await hybridForm.clickOnSubmit();
                    await hybridForm.waitForDistributedMasterSubmitButton();
                    await hybridForm.clickOnSubmitMasterForm();
                    await training.waitForSubmitButton();
                    await hybridForm.clickOnSubmit(); // setup step

                    await workload.waitForWorkloadUIStatus(workloadName);

                    // copy and edit the workload recently created
                    const copiedWorkloadName = objectGenerator.generateShortNameByPrefix('copy-dist');
                    await workload.toggleWorkload(workloadName);
                    await workload.clickOnCopyAndEditWorkload();
                    await hybridForm.fillTrainingName(copiedWorkloadName);
                    await hybridForm.clickOnSubmit();
                    await hybridForm.waitForDistributedMasterSubmitButton();
                    await hybridForm.clickOnSubmitMasterForm();
                    await training.waitForSubmitButton();
                    await hybridForm.clickOnSubmit(); // setup step

                    await workload.waitForWorkloadUIStatus(copiedWorkloadName, undefined, 300000);

                    await workload.deleteWorkload(workloadName);
                    await workload.deleteWorkload(copiedWorkloadName);
                });
            });
        });
    });
});
