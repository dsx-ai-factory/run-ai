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

            test.describe(`workloads table, tags: ${tags.teams.viper}`, () => {
                test('validate the presence of the Workload page', async () => {
                    videoName = 'table-exists';
                    await workloadTable.validateCurrentPageIsWorkload();
                });

                test('submit and validate Interactive workload with fraction of GPU', async () => {
                    videoName = 'submit-and-validate-interactive-workload';
                    const environmentName = await environmentUtils.resolveEnvPresetName('jupyter-lab');
                    const workspaceName = objectGenerator.generateShortNameByPrefix('half-gpu-workspace');

                    await pageFilterUtils.selectCluster(workloadTable.page, flowRootState.cluster.name);
                    await workspacePage.submitWorkspaceWorkloadHybridUI(project.name, workspaceName, {
                        environment: environmentName,
                        computeResource: 'half-gpu'
                    });
                    await workloadsUtilsV2.waitForStatusInWorkload(
                        workspaceName,
                        Phase.Running,
                        180,
                        Number(project.id)
                    );
                    await workloadTable.validateCurrentPageIsWorkload();

                    //validate requested gpu is correct
                    const workloadRowIndex = await workloadTable.getWorkloadRowIndexByName(workspaceName);
                    await workloadTable.isRequestedGpuCorrect(workloadRowIndex, '0.5');

                    await workloadTable.deleteWorkload(workspaceName);
                    logger.info(`workload ${workspaceName} deleted.`);
                    await waitUtils.waitForElementToNotExist(
                        workloadTable,
                        `td:has-text("${workspaceName}")`
                    );
                });

                test('submit and validate Training workload with 1 GPU and copy it', async () => {
                    videoName = 'submit-and-validate-training-workload';
                    const environmentName = await environmentUtils.resolveEnvPresetName('jupyter-lab');
                    const trainingName = objectGenerator.generateShortNameByPrefix('one-gpu-training');

                    await pageFilterUtils.selectCluster(workloadTable.page, flowRootState.cluster.name);
                    await trainingPage.createFlexibleTraining(
                        project.name,
                        trainingName,
                        undefined,
                        undefined,
                        environmentName,
                        'one-gpu'
                    );
                    await workloadTable.validateCurrentPageIsWorkload();

                    //validate requested gpu is correct
                    const workloadRowIndex = await workloadTable.getWorkloadRowIndexByName(trainingName);
                    await workloadTable.isRequestedGpuCorrect(workloadRowIndex, '1');

                    //copy and edit
                    await workloadTable.toggleWorkload(trainingName);
                    await workloadTable.clickOnCopyAndEditWorkload();

                    const copiedTrainingName = objectGenerator.generateShortNameByPrefix('copied');
                    await trainingPage.fillTrainingName(copiedTrainingName);

                    //submit form
                    await hybridForm.clickOnSubmit();
                    await workloadsUtilsV2.waitForStatusInWorkload(
                        trainingName,
                        Phase.Running,
                        180,
                        Number(project.id)
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

                    await workloadTable.deleteWorkload(trainingName);
                    logger.info(`workload ${trainingName} deleted.`);
                    //validate removed from the table
                    await workloadTable.isTableNotIncludeWorkload(trainingName);
                });

                test('should create a training and validate show details drawer', async () => {
                    videoName = 'submit-and-validate-training-workload-with-show-details-validation';
                    const environmentName = await environmentUtils.resolveEnvPresetName('jupyter-lab');
                    const trainingName = objectGenerator.generateShortNameByPrefix('one-gpu-training');

                    await pageFilterUtils.selectCluster(workloadTable.page, flowRootState.cluster.name);
                    await trainingPage.createFlexibleTraining(
                        project.name,
                        trainingName,
                        undefined,
                        undefined,
                        environmentName,
                        'one-gpu'
                    );
                    await workloadTable.validateCurrentPageIsWorkload();
                    const workloadIndex = await workloadTable.getWorkloadRowIndexByName(trainingName);
                    await workloadTable.validateRunningPhase(workloadIndex);
                    logger.info(`workload ${trainingName} is running.`);
                    //validate drawer tabs
                    await workloadTable.toggleWorkloadDrawerContent(trainingName);
                    await workloadTable.validateEventsHistoryTabContentExists();
                    await workloadTable.validateMetricTabAllPodsContentExists();
                    await workloadTable.validateLogsTabContentExists();
                    await workloadTable.closeWorkloadDrawerContent();
                    logger.info(`workload ${trainingName} drawer content validated.`);

                    //wait for actions to hide
                    await workloadTable.deleteWorkload(trainingName);
                    logger.info(`training ${trainingName} deleted.`);
                    await workloadTable.isTableNotIncludeWorkload(trainingName);
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
