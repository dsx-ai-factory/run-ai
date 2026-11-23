/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

//internal modules
import { orgUnitUtils, workloadsUtilsV2 } from '../../../../../../modules/helpers';
import {
    flowRootState,
    logger,
    objectGenerator,
    operations,
    tags,
    Utils,
    waitUtils,
    environmentUtils,
} from '../../../../../../modules/helpers';
import { Phase } from '../../../../../../modules/api/apis/workloadsApi/generated/model/phase';

//pages
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { WorkloadPage } from '../../../../../../modules/ui/runaiV2/pages/WorkloadPage';
import { Project } from '../../../../../../modules/api/apis/orgUnitApi';
import { WorkspacePage } from '../../../../../../modules/ui/runaiV2/pages/WorkspacePage';
import { TrainingPage } from '../../../../../../modules/ui/runaiV2/pages/TrainingPage';
import { InferencePage } from '../../../../../../modules/ui/runaiV2/pages/InferencePage';
import { EnvironmentPage } from '../../../../../../modules/ui/runaiV2/pages/EnvironmentPage';
import { pageFilterUtils } from '../../../../../../modules/ui/utils/pageFilterUtils';
import { test } from '@playwright/test';
import { InferenceFlatUIUtil } from '../../../../../../modules/ui/utils/inference/inferenceFlatUI.util';
import { HybridForm } from '../../../../../../modules/ui/runaiV2/pages/HybridSubmissionForm';
import {
    aiPractitionerRoleName,
    projectManagerRoleName
} from '../../../../../../modules/helpers/utils/AuthorizationUtils';
import { PROPERTY_TO_ROLE_MAPPING } from '../../../../../../modules/helpers/config/roleMapping';

interface CreateWorkloadOptions {
    clusterName: string;
    workloadName: string;
    environmentSelector?: string;
    projectId: number;
    projectName: string;
    environmentData?: {
        name: string;
        image?: string;
        containerPort?: number;
    };
    computeResourceData?: {
        name: string;
    };
}
test.describe(`Workloads UI, tags: ${tags.tool.ui} ${tags.resource.workloads} ${tags.teams.viper}`, () => {
    let browserEngine: BrowserEngine;
    let authPage: AuthPage;
    let workloadTable: WorkloadPage;
    let project: Project;
    let projectNamespace: string;
    let projectName: string;
    let workspacePage: WorkspacePage;
    let trainingPage: TrainingPage;
    let inferencePage: InferencePage;
    let environmentPage: EnvironmentPage;
    let inferenceFlatUIUtil: InferenceFlatUIUtil;
    let hybridForm: HybridForm;
    let videoName;
    const roles = [PROPERTY_TO_ROLE_MAPPING.system_administrator, aiPractitionerRoleName, projectManagerRoleName];

    roles.forEach((role) => {
        test.describe(`Workloads UI for role: ${role}`, () => {
            const startSession = async (): Promise<void> => {
                browserEngine = new BrowserEngine();
                await browserEngine.start();
                authPage = new AuthPage(browserEngine.currentPage);
                workloadTable = new WorkloadPage(browserEngine.currentPage);
                workspacePage = new WorkspacePage(browserEngine.currentPage);
                trainingPage = new TrainingPage(browserEngine.currentPage);
                inferencePage = new InferencePage(browserEngine.currentPage);
                environmentPage = new EnvironmentPage(browserEngine.currentPage);
                inferenceFlatUIUtil = new InferenceFlatUIUtil(browserEngine.currentPage);
                hybridForm = new HybridForm(browserEngine.currentPage);
                await authPage.loginAsRole(role);
            };

            test.beforeEach(async () => {
                await startSession();
                await workloadTable.navigateToWorkloads();
            });

            test.afterEach('stop browser', async ({}, testInfo) => {
                await browserEngine.stopAndArchive(testInfo, 'workloads-' + videoName);
            });

            const createTraining = async (options: CreateWorkloadOptions): Promise<void> => {
                await pageFilterUtils.selectCluster(workloadTable.page, options.clusterName);
                await workloadTable.openTrainingForm();

                // fill the workspace form details
                await workloadTable.workloadFormSelectProject(projectName);
                await waitUtils.waitForElementToExist(trainingPage, trainingPage.createNameInput);
                await trainingPage.workloadFormSelectTemplate();
                await trainingPage.fillTrainingName(options.workloadName);
                await trainingPage.waitForTrainingNameCount(options.workloadName);
                await trainingPage.selectAssetsSubmissionOption();
                await trainingPage.clickOnContinueBtn();

                // choose env
                await hybridForm.openEnvironmentsDrawer();
                await hybridForm.clickOnDrawerItemByLabel(options.environmentData.name);

                // choose compute
                await hybridForm.openComputesDrawer();
                await hybridForm.clickOnDrawerItemByLabel(options.computeResourceData.name);

                //submit form
                await trainingPage.clickOnSubmitBtn();

                //validate return to workloads
                await workloadTable.validateCurrentPageIsWorkload();
            };

            const createWorkspace = async (options: CreateWorkloadOptions): Promise<void> => {
                await pageFilterUtils.selectCluster(workloadTable.page, options.clusterName);
                await workloadTable.openWorkspaceForm();

                // fill the workspace form details
                await workloadTable.workloadFormSelectProject(projectName);
                await waitUtils.waitForElementToExist(workspacePage, workspacePage.createWorkspaceNameInput);
                await workspacePage.workloadFormSelectTemplate();
                await workspacePage.fillWorkspaceName(options.workloadName);
                await workspacePage.waitForWorkspaceNameCount(options.workloadName);
                await workspacePage.selectAssetsSubmissionOption();
                await workspacePage.clickOnContinue();

                // choose env
                await hybridForm.openEnvironmentsDrawer();
                await hybridForm.clickOnDrawerItemByLabel(options.environmentData.name);

                // choose compute
                await hybridForm.openComputesDrawer();
                await hybridForm.clickOnDrawerItemByLabel(options.computeResourceData.name);

                // create workspace
                await workspacePage.clickOnCreateWorkspaceBtn();
                logger.debug('workspace created successfully');
                await workloadsUtilsV2.waitForStatusInWorkload(options.workloadName, Phase.Running, 180, options.projectId);

                //validate return to workloads
                await workloadTable.validateCurrentPageIsWorkload();
            };

            test.describe(`workloads table, tags: ${tags.teams.viper}`, () => {
                test('validate the presence of the Workload page', async () => {
                    videoName = 'table-exists';
                    await workloadTable.validateCurrentPageIsWorkload();
                });

                test('submit and validate Interactive workload with fraction of GPU', async () => {
                    videoName = 'submit-and-validate-interactive-workload';
                    const envName = await environmentUtils.resolveEnvPresetName('jupyter-lab');
                    const createWorkloadOptions: CreateWorkloadOptions = {
                        clusterName: flowRootState.cluster.name,
                        workloadName: objectGenerator.generateShortNameByPrefix('half-gpu-workspace'),
                        projectId: Number(project.id),
                        projectName: project.name,
                        environmentSelector: `[aid="${envName}-card"]`,
                        computeResourceData: {
                            name: 'half-gpu'
                        },
                        environmentData: {
                            name: envName
                        }
                    };

                    await createWorkspace(createWorkloadOptions);

                    //validate requested gpu is correct
                    const workloadRowIndex = await workloadTable.getWorkloadRowIndexByName(
                        createWorkloadOptions.workloadName
                    );
                    await workloadTable.isRequestedGpuCorrect(workloadRowIndex, '0.5');

                    await workloadTable.deleteWorkload(createWorkloadOptions.workloadName);
                    logger.info(`workload ${createWorkloadOptions.workloadName} deleted.`);
                    await waitUtils.waitForElementToNotExist(
                        workloadTable,
                        `td:has-text("${createWorkloadOptions.workloadName}")`
                    );
                });

                test('submit and validate Training workload with 1 GPU and copy it', async () => {
                    videoName = 'submit-and-validate-training-workload';
                    const trainingEnvName = await environmentUtils.resolveEnvPresetName('jupyter-lab');
                    const createWorkloadOptions: CreateWorkloadOptions = {
                        clusterName: flowRootState.cluster.name,
                        workloadName: objectGenerator.generateShortNameByPrefix('one-gpu-training'),
                        projectId: Number(project.id),
                        projectName: project.name,
                        environmentSelector: `[aid="${trainingEnvName}-card"]`,
                        computeResourceData: {
                            name: 'one-gpu'
                        },
                        environmentData: {
                            name: trainingEnvName
                        }
                    };
                    await createTraining(createWorkloadOptions);

                    //validate requested gpu is correct
                    const workloadRowIndex = await workloadTable.getWorkloadRowIndexByName(
                        createWorkloadOptions.workloadName
                    );
                    await workloadTable.isRequestedGpuCorrect(workloadRowIndex, '1');

                    //copy and edit
                    await workloadTable.toggleWorkload(createWorkloadOptions.workloadName);
                    await workloadTable.clickOnCopyAndEditWorkload();

                    const copiedTrainingName = objectGenerator.generateShortNameByPrefix('copied');
                    await trainingPage.fillTrainingName(copiedTrainingName);

                    //submit form
                    await trainingPage.clickOnSubmitBtn();
                    await workloadsUtilsV2.waitForStatusInWorkload(
                        createWorkloadOptions.workloadName,
                        Phase.Running,
                        180,
                        createWorkloadOptions.projectId
                    );
                    await waitUtils.waitForElementToExist(
                        workloadTable,
                        `td:has-text("${copiedTrainingName}")`
                    );
                    //validate requested gpu is correct
                    const workloadRowIndexCopied =
                        await workloadTable.getWorkloadRowIndexByName(copiedTrainingName);
                    await workloadTable.isRequestedGpuCorrect(workloadRowIndexCopied, '1');

                    //delete both workloads
                    await workloadTable.deleteWorkload(copiedTrainingName);
                    logger.info(`workload ${copiedTrainingName} deleted.`);

                    await workloadTable.deleteWorkload(createWorkloadOptions.workloadName);
                    logger.info(`workload ${createWorkloadOptions.workloadName} deleted.`);
                    //validate removed from the table
                    await workloadTable.isTableNotIncludeWorkload(createWorkloadOptions.workloadName);
                });

                test('should create a training and validate show details drawer', async () => {
                    videoName = 'submit-and-validate-training-workload-with-show-details-validation';
                    const drawerEnvName = await environmentUtils.resolveEnvPresetName('jupyter-lab');
                    const createWorkloadOptions: CreateWorkloadOptions = {
                        clusterName: flowRootState.cluster.name,
                        workloadName: objectGenerator.generateShortNameByPrefix('one-gpu-training'),
                        projectId: Number(project.id),
                        projectName: project.name,
                        environmentSelector: `[aid="${drawerEnvName}-card"]`,
                        computeResourceData: {
                            name: 'one-gpu'
                        },
                        environmentData: {
                            name: drawerEnvName
                        }
                    };
                    await createTraining(createWorkloadOptions);
                    const workloadIndex = await workloadTable.getWorkloadRowIndexByName(
                        createWorkloadOptions.workloadName
                    );
                    await workloadTable.validateRunningPhase(workloadIndex);
                    logger.info(`workload ${createWorkloadOptions.workloadName} is running.`);
                    //validate drawer tabs
                    await workloadTable.toggleWorkloadDrawerContent(createWorkloadOptions.workloadName);
                    await workloadTable.validateEventsHistoryTabContentExists();
                    await workloadTable.validateMetricTabAllPodsContentExists();
                    await workloadTable.validateLogsTabContentExists();
                    await workloadTable.closeWorkloadDrawerContent();
                    logger.info(`workload ${createWorkloadOptions.workloadName} drawer content validated.`);

                    //wait for actions to hide
                    await workloadTable.deleteWorkload(createWorkloadOptions.workloadName);
                    logger.info(`training ${createWorkloadOptions.workloadName} deleted.`);
                    await workloadTable.isTableNotIncludeWorkload(createWorkloadOptions.workloadName);
                });
            });
        });
    });

    const createProject = async (): Promise<void> => {
        projectName = objectGenerator.generateShortNameByPrefix('p-0');
        projectNamespace = Utils.namespaceFromProjectName(projectName);

        project = await orgUnitUtils.createProject({
            name: projectName,
            clusterId: flowRootState.cluster.uuid,
            resources: [
                {
                    gpu: {
                        deserved: 16,
                        limit: 16
                    }
                }
            ]
        });
        logger.info(`created project ${project.name}`);
    };

    test.beforeAll(async () => {
        await createProject();
    });

    test.afterAll(async () => {
        await operations.deleteProjectAndNamespace(Number(project.id));
    });
});
