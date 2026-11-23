/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import {
    Utils,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    flowRootState,
    templateUtils
} from '../../../../../../modules/helpers';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { ScopeType, WorkloadType } from '../../../../../../modules/api/apis/workloadsApi/generated/api';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { WorkspacePage } from '../../../../../../modules/ui/runaiV2/pages/WorkspacePage';
import { TrainingPage } from '../../../../../../modules/ui/runaiV2/pages/TrainingPage';
import { WorkloadPage, EWorkloadStatus } from '../../../../../../modules/ui/runaiV2/pages/WorkloadPage';
import { HybridForm } from '../../../../../../modules/ui/runaiV2/pages/HybridSubmissionForm';
import { DISTRIBUTED_TOOLS_LABELS } from '../../../../../../modules/ui/runaiV2/pages/TrainingPage';

test.describe(`Templates Quickstart Capabilities UI Tests , tags: ${tags.tool.ui} ${tags.resource.templateV2} ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_23)(function () {
        let browserEngine: BrowserEngine;
        let authPage: AuthPage;
        let workspacePage: WorkspacePage;
        let trainingPage: TrainingPage;
        let workloadPage: WorkloadPage;
        let hybridForm: HybridForm;
        let videoName: string;
        let projectName: string;
        let namespace: string;
        let projectId: number;
        let runnableTemplateId: string;
        let nonRunnableTemplateId: string;
        let runnableTemplateName: string;
        let nonRunnableTemplateName: string;
        let workspaceName: string;
        let trainingName: string;

        const ubuntuImage = 'ubuntu:latest';
        const alpineImage = 'alpine:latest';
        const sleepInfinity = 'sleep infinity';
        const cpuOnly = 'cpu-only';
        const tensorboardEnvName = 'tensorboard-tensorflow';

        const baseTemplateSpec = {
            imagePullPolicy: 'IfNotPresent',
            backoffLimit: 6,
            command: sleepInfinity,
            existingComputeResName: cpuOnly
        };

        const runnableTemplateSpec = {
            ...baseTemplateSpec,
            image: ubuntuImage
        };

        const nonRunnableTemplateSpec = {
            ...baseTemplateSpec
        };

        test.beforeAll(async function () {
            // Create project and namespace
            projectName = objectGenerator.generateShortNameByPrefix('project-for-quickstart-test');
            namespace = Utils.namespaceFromProjectName(projectName);

            const projectResponse = await projectResourcesUtils.createProject(namespace, {
                name: projectName,
                deservedGpus: 0,
                maxAllowedGpus: 0
            });
            projectId = projectResponse.id;
        });

        test.afterAll('cleanup resources', async () => {
            // Delete project and namespace
            if (projectId) {
                try {
                    await operations.deleteProjectAndNamespace(projectId);
                    logger.info(`Deleted project ${projectName} and namespace ${namespace}`);
                } catch (error) {
                    logger.error(`Failed to delete project ${projectName} and namespace ${namespace}`, error);
                }
            }
        });

        test.beforeEach(async function () {
            browserEngine = new BrowserEngine();
            await browserEngine.start();
            authPage = new AuthPage(browserEngine.currentPage);
            trainingPage = new TrainingPage(browserEngine.currentPage);
            workspacePage = new WorkspacePage(browserEngine.currentPage);
            workloadPage = new WorkloadPage(browserEngine.currentPage);
            hybridForm = new HybridForm(browserEngine.currentPage);
            await authPage.signIn(
                flowRootState.users.l1_researcher.email,
                flowRootState.users.l1_researcher.password
            );

            trainingName = objectGenerator.generateShortNameByPrefix('quickstart-training');
            workspaceName = objectGenerator.generateShortNameByPrefix('quickstart-workspace');
        });

        test.describe(`Templates Quickstart Capabilities UI Tests - Workspace, tags: ${tags.tool.ui} ${tags.resource.templateV2} ${tags.teams.loong}`, () => {
            test.beforeAll(async function () {
                // Create runnable workspace template
                runnableTemplateName = objectGenerator.generateShortNameByPrefix(
                    'runnable-workspace-template'
                );
                const runnableMetaConfig = {
                    templateName: runnableTemplateName,
                    workloadType: WorkloadType.Workspace,
                    scopeType: ScopeType.Project,
                    scopeId: projectId.toString(),
                    description: 'Runnable workspace template for quick start testing'
                };

                const runnableResponse = await templateUtils.createWorkloadTemplateV2(
                    runnableMetaConfig,
                    runnableTemplateSpec
                );
                runnableTemplateId = runnableResponse.body.meta.id;
                logger.info(
                    `Created runnable template: ${runnableTemplateName} with ID: ${runnableTemplateId}`
                );

                // Create non-runnable workspace template (missing image)
                nonRunnableTemplateName = objectGenerator.generateShortNameByPrefix(
                    'non-runnable-workspace-template'
                );
                const nonRunnableMetaConfig = {
                    templateName: nonRunnableTemplateName,
                    workloadType: WorkloadType.Workspace,
                    scopeType: ScopeType.Project,
                    scopeId: projectId.toString(),
                    description: 'Non-runnable workspace template for testing'
                };

                const nonRunnableResponse = await templateUtils.createWorkloadTemplateV2(
                    nonRunnableMetaConfig,
                    nonRunnableTemplateSpec
                );
                nonRunnableTemplateId = nonRunnableResponse.body.meta.id;
                logger.info(
                    `Created non-runnable template: ${nonRunnableTemplateName} with ID: ${nonRunnableTemplateId}`
                );
            });

            test.afterAll('cleanup templates', async () => {
                // Delete templates
                if (runnableTemplateId) {
                    try {
                        await templateUtils.deleteWorkloadTemplateV2(
                            runnableTemplateId,
                            WorkloadType.Workspace
                        );
                        logger.info(`Deleted runnable template: ${runnableTemplateName}`);
                    } catch (error) {
                        logger.error(`Failed to delete runnable template: ${runnableTemplateName}`, error);
                    }
                }

                if (nonRunnableTemplateId) {
                    try {
                        await templateUtils.deleteWorkloadTemplateV2(
                            nonRunnableTemplateId,
                            WorkloadType.Workspace
                        );
                        logger.info(`Deleted non-runnable template: ${nonRunnableTemplateName}`);
                    } catch (error) {
                        logger.error(
                            `Failed to delete non-runnable template: ${nonRunnableTemplateName}`,
                            error
                        );
                    }
                }
            });

            test.afterEach('cleanup test resources and stop browser', async ({}, testInfo) => {
                await browserEngine.stopAndArchive(testInfo, videoName);
            });

            test('T5331846: UI-Quick start submission with runnable template - Workspace', async function () {
                videoName = 'quickstart-submission-runnable-template-workspace';
                await test.step('Create a workspace with runnable template and choose CREATE-WORKSPACE option', async function () {
                    await workspacePage.fillFlexWorkspaceFirstPage(
                        projectName,
                        workspaceName,
                        runnableTemplateName,
                        false // create workspace directly
                    );
                });

                await test.step('Verify workspace is running', async function () {
                    await workspacePage.waitForWorkspaceStatus(workspaceName, 'Running');
                });

                await test.step('Delete workspace', async function () {
                    await workspacePage.deleteWorkload(workspaceName);
                });
            });

            test('T5331876: UI-Quick start submission with runnable template - Advanced setup option - Workspace', async function () {
                videoName = 'quickstart-submission-runnable-template-workspace-advanced-setup';
                await test.step('Create a workspace with runnable template and choose Advanced setup option', async function () {
                    await workspacePage.fillFlexWorkspaceFirstPage(
                        projectName,
                        workspaceName,
                        runnableTemplateName,
                        true // create workspace with advanced setup
                    );
                });

                await test.step('Fill the second page of the workspace form and submit', async function () {
                    await hybridForm.fillHybridFormByLoadingExistingConfig(tensorboardEnvName, cpuOnly);
                    await hybridForm.clickOnSubmit();
                });

                await test.step('Verify workspace is running', async function () {
                    await workspacePage.waitForWorkspaceStatus(workspaceName, 'Running');
                });

                await test.step('Delete workspace', async function () {
                    await workspacePage.deleteWorkload(workspaceName);
                });
            });

            test('T5331877: UI-Quick start submission with non-runnable template - Workspace', async function () {
                videoName = 'quickstart-submission-non-runnable-template-workspace-advanced-setup';

                await test.step('Create a workspace with non-runnable template and choose Advanced setup option', async function () {
                    await workspacePage.navigateToWorkloads();
                    await workspacePage.openWorkspaceForm();
                    await workspacePage.selectClusterFromDropDown(flowRootState.cluster.name);
                    await workspacePage.workloadFormSelectProject(projectName);

                    await workspacePage.workloadFormSelectTemplateV2(nonRunnableTemplateName);
                    await workspacePage.page.waitForTimeout(500);

                    await workspacePage.fillWorkspaceName(workspaceName);
                    await workspacePage.waitForWorkspaceNameCount(workspaceName);
                    await hybridForm.openQuickSubmitDropdown();
                    // verify submit option is disabled
                    expect(await hybridForm.isQuickSubmitDropdownSubmitOptionDisabled()).toBe(true);
                    // click advanced setup option
                    await hybridForm.clickOnQuickSubmitDropdownContinueOption();
                });

                await test.step('Fill the second page of the workspace form and submit', async function () {
                    await hybridForm.fillHybridFormByLoadingExistingConfig(tensorboardEnvName, cpuOnly);
                    await hybridForm.clickOnSubmit();
                });

                await test.step('Verify workspace is running', async function () {
                    await workspacePage.waitForWorkspaceStatus(workspaceName, 'Running');
                });

                await test.step('Delete workspace', async function () {
                    await workspacePage.deleteWorkload(workspaceName);
                });
            });

            test('T5331880: UI-Template switching behavior between runnable and non-runnable templates - Workspace', async function () {
                videoName = 'template-switching-behavior-workspace';

                await test.step('Navigate to workspace creation page', async function () {
                    await workspacePage.navigateToWorkloads();
                    await workspacePage.openWorkspaceForm();
                });

                await test.step('Select a project with both runnable and non-runnable templates', async function () {
                    await workspacePage.selectClusterFromDropDown(flowRootState.cluster.name);
                    await workspacePage.workloadFormSelectProject(projectName);
                });

                await test.step('Select a runnable template, verify main button shows "CREATE WORKSPACE"', async function () {
                    await workspacePage.workloadFormSelectTemplateV2(runnableTemplateName);
                    await workspacePage.page.waitForTimeout(500);
                    await hybridForm.openQuickSubmitDropdown();
                    expect(await hybridForm.isQuickSubmitDropdownSubmitOptionDisabled()).toBe(false);
                    expect(await hybridForm.isQuickSubmitDropdownContinueOptionDisabled()).toBe(false);
                });

                await test.step('Switch to a non-runnable template, verify main button shows "ADVANCED SETUP"', async function () {
                    await workspacePage.workloadFormSelectTemplateV2(nonRunnableTemplateName);
                    await workspacePage.page.waitForTimeout(500);
                    await hybridForm.openQuickSubmitDropdown();
                    expect(await hybridForm.isQuickSubmitDropdownSubmitOptionDisabled()).toBe(true);
                    expect(await hybridForm.isQuickSubmitDropdownContinueOptionDisabled()).toBe(false);
                });

                await test.step('Switch back to runnable template, verify main button shows "CREATE WORKSPACE" again', async function () {
                    await workspacePage.workloadFormSelectTemplateV2(runnableTemplateName);
                    await workspacePage.page.waitForTimeout(500);
                    await hybridForm.openQuickSubmitDropdown();
                    // Verify both options are enabled again
                    expect(await hybridForm.isQuickSubmitDropdownSubmitOptionDisabled()).toBe(false);
                    expect(await hybridForm.isQuickSubmitDropdownContinueOptionDisabled()).toBe(false);
                });

                await test.step('Switch to no template selection, verify main button shows "Continue"', async function () {
                    await workspacePage.workloadFromScratchTemplateV2();
                    await workspacePage.page.waitForTimeout(500);
                    // Verify continue button is visible when no template is selected
                    expect(await workspacePage.isContinueButtonVisible()).toBe(true);
                });
            });
        });

        test.describe(`Templates Quickstart Capabilities UI Tests - Training, tags: ${tags.tool.ui} ${tags.resource.templateV2} ${tags.teams.loong}`, () => {
            test.beforeAll(async function () {
                // Create runnable training template
                runnableTemplateName = objectGenerator.generateShortNameByPrefix(
                    'runnable-training-template'
                );
                const runnableMetaConfig = {
                    templateName: runnableTemplateName,
                    workloadType: WorkloadType.Training,
                    scopeType: ScopeType.Project,
                    scopeId: projectId.toString(),
                    description: 'Runnable training template for quick start testing'
                };

                const runnableResponse = await templateUtils.createWorkloadTemplateV2(
                    runnableMetaConfig,
                    runnableTemplateSpec
                );
                runnableTemplateId = runnableResponse.body.meta.id;
                logger.info(
                    `Created runnable template: ${runnableTemplateName} with ID: ${runnableTemplateId}`
                );

                // Create non-runnable training template (missing image)
                nonRunnableTemplateName = objectGenerator.generateShortNameByPrefix(
                    'non-runnable-training-template'
                );
                const nonRunnableMetaConfig = {
                    templateName: nonRunnableTemplateName,
                    workloadType: WorkloadType.Training,
                    scopeType: ScopeType.Project,
                    scopeId: projectId.toString(),
                    description: 'Non-runnable training template for testing'
                };

                const nonRunnableResponse = await templateUtils.createWorkloadTemplateV2(
                    nonRunnableMetaConfig,
                    nonRunnableTemplateSpec
                );
                nonRunnableTemplateId = nonRunnableResponse.body.meta.id;
                logger.info(
                    `Created non-runnable template: ${nonRunnableTemplateName} with ID: ${nonRunnableTemplateId}`
                );
            });

            test.afterAll('cleanup templates', async () => {
                // Delete templates
                if (runnableTemplateId) {
                    try {
                        await templateUtils.deleteWorkloadTemplateV2(
                            runnableTemplateId,
                            WorkloadType.Workspace
                        );
                        logger.info(`Deleted runnable template: ${runnableTemplateName}`);
                    } catch (error) {
                        logger.error(`Failed to delete runnable template: ${runnableTemplateName}`, error);
                    }
                }
                if (nonRunnableTemplateId) {
                    try {
                        await templateUtils.deleteWorkloadTemplateV2(
                            nonRunnableTemplateId,
                            WorkloadType.Workspace
                        );
                        logger.info(`Deleted non-runnable template: ${nonRunnableTemplateName}`);
                    } catch (error) {
                        logger.error(
                            `Failed to delete non-runnable template: ${nonRunnableTemplateName}`,
                            error
                        );
                    }
                }
            });

            test.afterEach('cleanup test resources and stop browser', async ({}, testInfo) => {
                await browserEngine.stopAndArchive(testInfo, videoName);
            });

            test('T5331881: UI-Quick start submission with runnable template - Training', async function () {
                videoName = 'quickstart-submission-runnable-template-training';

                await test.step('Create a training with runnable template and choose CREATE-TRAINING option', async function () {
                    await trainingPage.fillFlexTrainingFirstPage(
                        projectName,
                        trainingName,
                        false,
                        undefined,
                        false,
                        runnableTemplateName,
                        false // create training directly
                    );
                });

                await test.step('Verify training is running', async function () {
                    await workloadPage.waitForWorkloadUIStatus(trainingName, EWorkloadStatus.Running);
                });

                await test.step('Delete training', async function () {
                    await trainingPage.deleteWorkload(trainingName);
                });
            });

            test('T5331882: UI-Quick start submission with runnable template - Advanced setup option - Training', async function () {
                videoName = 'quickstart-submission-runnable-template-training-advanced-setup';
                await test.step('Create a training with runnable template and choose Advanced setup option', async function () {
                    await trainingPage.fillFlexTrainingFirstPage(
                        projectName,
                        trainingName,
                        false,
                        undefined,
                        false,
                        runnableTemplateName,
                        true // create training with advanced setup
                    );
                });

                await test.step('Fill the second page of the training form and submit', async function () {
                    await hybridForm.fillHybridFormByLoadingExistingConfig(tensorboardEnvName, cpuOnly);
                    await hybridForm.clickOnSubmit();
                });

                await test.step('Verify training is running', async function () {
                    await trainingPage.waitForTrainingStatus(trainingName, 'Running');
                });

                await test.step('Delete training', async function () {
                    await trainingPage.deleteWorkload(trainingName);
                });
            });

            test('T5331883: UI-Quick start submission with non-runnable template - Training', async function () {
                videoName = 'quickstart-submission-non-runnable-template-training-advanced-setup';

                await test.step('Create a training with non-runnable template and choose Advanced setup option', async function () {
                    await trainingPage.navigateToWorkloads();
                    await trainingPage.openTrainingForm();
                    await trainingPage.selectClusterFromDropDown(flowRootState.cluster.name);
                    await trainingPage.workloadFormSelectProject(projectName);

                    await trainingPage.workloadFormSelectTemplateV2(nonRunnableTemplateName);
                    await trainingPage.page.waitForTimeout(500);

                    await trainingPage.fillTrainingName(trainingName);
                    await trainingPage.waitForTrainingNameCount(trainingName);
                    await hybridForm.openQuickSubmitDropdown();
                    // verify submit option is disabled
                    expect(await hybridForm.isQuickSubmitDropdownSubmitOptionDisabled()).toBe(true);
                    // click advanced setup option
                    await hybridForm.clickOnQuickSubmitDropdownContinueOption();
                });

                await test.step('Fill the second page of the training form and submit', async function () {
                    await hybridForm.fillHybridFormByLoadingExistingConfig(tensorboardEnvName, cpuOnly);
                    await hybridForm.clickOnSubmit();
                });

                await test.step('Verify training is running', async function () {
                    await trainingPage.waitForTrainingStatus(trainingName, 'Running');
                });

                await test.step('Delete training', async function () {
                    await trainingPage.deleteWorkload(trainingName);
                });
            });

            test('T5331886: UI-Template switching behavior between runnable and non-runnable templates - Training', async function () {
                videoName = 'template-switching-behavior-training';

                await test.step('Navigate to training creation page', async function () {
                    await trainingPage.navigateToWorkloads();
                    await trainingPage.openTrainingForm();
                });

                await test.step('Select a project with both runnable and non-runnable templates', async function () {
                    await trainingPage.selectClusterFromDropDown(flowRootState.cluster.name);
                    await trainingPage.workloadFormSelectProject(projectName);
                });

                await test.step('Select a runnable template, verify main button shows "CREATE TRAINING"', async function () {
                    await trainingPage.workloadFormSelectTemplateV2(runnableTemplateName);
                    await trainingPage.page.waitForTimeout(500);
                    await hybridForm.openQuickSubmitDropdown();
                    expect(await hybridForm.isQuickSubmitDropdownSubmitOptionDisabled()).toBe(false);
                    expect(await hybridForm.isQuickSubmitDropdownContinueOptionDisabled()).toBe(false);
                });

                await test.step('Switch to a non-runnable template, verify main button shows "ADVANCED SETUP"', async function () {
                    await trainingPage.workloadFormSelectTemplateV2(nonRunnableTemplateName);
                    await trainingPage.page.waitForTimeout(500);
                    await hybridForm.openQuickSubmitDropdown();
                    expect(await hybridForm.isQuickSubmitDropdownSubmitOptionDisabled()).toBe(true);
                    expect(await hybridForm.isQuickSubmitDropdownContinueOptionDisabled()).toBe(false);
                });

                await test.step('Switch back to runnable template, verify main button shows "CREATE TRAINING" again', async function () {
                    await trainingPage.workloadFormSelectTemplateV2(runnableTemplateName);
                    await trainingPage.page.waitForTimeout(500);
                    await hybridForm.openQuickSubmitDropdown();
                    // Verify both options are enabled again
                    expect(await hybridForm.isQuickSubmitDropdownSubmitOptionDisabled()).toBe(false);
                    expect(await hybridForm.isQuickSubmitDropdownContinueOptionDisabled()).toBe(false);
                });

                await test.step('Switch to no template selection, verify main button shows "Continue"', async function () {
                    await trainingPage.workloadFromScratchTemplateV2();
                    await trainingPage.page.waitForTimeout(500);
                    // Verify continue button is visible when no template is selected
                    expect(await hybridForm.isContinueButtonVisible()).toBe(true);
                });
            });
        });

        test.describe(`Templates Quickstart Capabilities UI Tests - Distributed Training, tags: ${tags.tool.ui} ${tags.resource.templateV2} ${tags.teams.loong}`, () => {
            test.afterEach('cleanup test resources and stop browser', async ({}, testInfo) => {
                await browserEngine.stopAndArchive(testInfo, videoName);
                // Delete templates
                if (runnableTemplateId) {
                    try {
                        await templateUtils.deleteWorkloadTemplateV2(
                            runnableTemplateId,
                            WorkloadType.Distributed
                        );
                        logger.info(`Deleted runnable distributed template: ${runnableTemplateName}`);
                    } catch (error) {
                        logger.error(
                            `Failed to delete runnable distributed template: ${runnableTemplateName}`,
                            error
                        );
                    }
                }
            });

            const baseDistSpec = {
                numWorkers: 1,
                image: alpineImage,
                imagePullPolicy: 'IfNotPresent',
                backoffLimit: 6,
                command: sleepInfinity,
                compute: cpuOnly
            };

            const distMPISpec = {
                ...baseDistSpec,
                distributedFramework: 'MPI'
            };

            const distPyTorchSpec = {
                ...baseDistSpec,
                distributedFramework: 'PyTorch'
            };

            const distXGBoostSpec = {
                ...baseDistSpec,
                distributedFramework: 'XGBoost'
            };

            const distTensorFlowSpec = {
                ...baseDistSpec,
                distributedFramework: 'TF'
            };

            const distJAXSpec = {
                ...baseDistSpec,
                distributedFramework: 'JAX'
            };

            const dataSets = [
                {
                    testCase: 'T5331893: UI-Distributed template with PyTorch framework - runnable',
                    workerSpec: distPyTorchSpec,
                    framework: DISTRIBUTED_TOOLS_LABELS.PY_TORCH,
                    templatePrefix: 'pytorch-distributed-template',
                    description: 'PyTorch distributed template for quick start testing'
                },
                {
                    testCase: 'T5331895: UI-Distributed template with XGBoost framework - runnable',
                    workerSpec: distXGBoostSpec,
                    framework: DISTRIBUTED_TOOLS_LABELS.XG_BOOST,
                    templatePrefix: 'xgboost-distributed-template',
                    description: 'XGBoost distributed template for quick start testing'
                },
                {
                    testCase: 'T5331897: UI-Distributed template with MPI framework - runnable',
                    workerSpec: distMPISpec,
                    framework: DISTRIBUTED_TOOLS_LABELS.MPI,
                    templatePrefix: 'mpi-distributed-template',
                    description: 'MPI distributed template for quick start testing'
                },
                {
                    testCase: 'T5331899: UI-Distributed template with TensorFlow framework - runnable',
                    workerSpec: distTensorFlowSpec,
                    framework: DISTRIBUTED_TOOLS_LABELS.TENSOR_FLOW,
                    templatePrefix: 'tensorflow-distributed-template',
                    description: 'TensorFlow distributed template for quick start testing'
                },
                {
                    testCase: 'T5331901: UI-Distributed template with JAX framework - runnable',
                    workerSpec: distJAXSpec,
                    framework: DISTRIBUTED_TOOLS_LABELS.JAX,
                    templatePrefix: 'jax-distributed-template',
                    description: 'JAX distributed template for quick start testing'
                }
            ];

            dataSets.forEach((caseData) => {
                test(`${caseData.testCase}`, async function () {
                    runnableTemplateName = objectGenerator.generateShortNameByPrefix(caseData.templatePrefix);
                    videoName = 'quickstart-submission-runnable-template-training-advanced-setup';
                    await test.step('create a distributed template', async function () {
                        const metaConfig = {
                            workloadType: WorkloadType.Distributed,
                            scopeType: ScopeType.Project,
                            scopeId: projectId.toString(),
                            templateName: runnableTemplateName,
                            description: caseData.description
                        };

                        const distConfig =
                            caseData.framework === DISTRIBUTED_TOOLS_LABELS.JAX
                                ? undefined
                                : { masterSpecSameAsWorker: true };

                        const response = await templateUtils.createWorkloadTemplateV2(
                            metaConfig,
                            caseData.workerSpec,
                            distConfig
                        );
                        const templateId = response.body.meta.id;
                        logger.info(
                            `Created ${caseData.framework} distributed template: ${runnableTemplateName} with ID: ${templateId}`
                        );
                        runnableTemplateId = templateId;
                    });

                    await test.step('create a distributed training with runnable template and verify create training button is visible', async function () {
                        await trainingPage.navigateToWorkloads();
                        await trainingPage.openTrainingForm();
                        await trainingPage.selectClusterFromDropDown(flowRootState.cluster.name);
                        await trainingPage.workloadFormSelectProject(projectName);
                        await trainingPage.selectedMultiNodeOption();

                        await trainingPage.selectedDistributedTool(caseData.framework);
                        await trainingPage.workloadFormSelectTemplateV2(runnableTemplateName);
                        await trainingPage.page.waitForTimeout(500);

                        await trainingPage.fillTrainingName(trainingName);
                        await trainingPage.waitForTrainingNameCount(trainingName);
                        // verify create training button is visible
                        expect(await hybridForm.isCreateTrainingButtonVisible()).toBe(true);
                    });
                });
            });
        });

        test.describe(`Templates Quickstart Capabilities UI Tests - Non-Runnable Distributed Training, tags: ${tags.tool.ui} ${tags.resource.templateV2} ${tags.teams.loong}`, () => {
            test.afterEach('cleanup test resources and stop browser', async ({}, testInfo) => {
                await browserEngine.stopAndArchive(testInfo, videoName);
                // Delete templates
                if (nonRunnableTemplateId) {
                    try {
                        await templateUtils.deleteWorkloadTemplateV2(
                            nonRunnableTemplateId,
                            WorkloadType.Distributed
                        );
                        logger.info(`Deleted non-runnable distributed template: ${nonRunnableTemplateName}`);
                    } catch (error) {
                        logger.error(
                            `Failed to delete non-runnable distributed template: ${nonRunnableTemplateName}`,
                            error
                        );
                    }
                }
            });

            const baseNonRunnableDistSpec = {
                numWorkers: 1,
                imagePullPolicy: 'IfNotPresent',
                backoffLimit: 6,
                command: sleepInfinity,
                compute: cpuOnly
            };

            const distMPINonRunnableSpec = {
                ...baseNonRunnableDistSpec,
                distributedFramework: 'MPI'
            };

            const distPyTorchNonRunnableSpec = {
                ...baseNonRunnableDistSpec,
                distributedFramework: 'PyTorch'
            };

            const distXGBoostNonRunnableSpec = {
                ...baseNonRunnableDistSpec,
                distributedFramework: 'XGBoost'
            };

            const distTensorFlowNonRunnableSpec = {
                ...baseNonRunnableDistSpec,
                distributedFramework: 'TF'
            };

            const distJAXNonRunnableSpec = {
                ...baseNonRunnableDistSpec,
                distributedFramework: 'JAX'
            };

            const dataSets = [
                {
                    testCase: 'T5331894: UI-Distributed template with PyTorch framework - non-runnable',
                    workerSpec: distPyTorchNonRunnableSpec,
                    framework: DISTRIBUTED_TOOLS_LABELS.PY_TORCH,
                    templatePrefix: 'pytorch-distributed-template-non-runnable',
                    description: 'PyTorch distributed non-runnable template for quick start testing'
                },
                {
                    testCase: 'T5331896: UI-Distributed template with XGBoost framework - non-runnable',
                    workerSpec: distXGBoostNonRunnableSpec,
                    framework: DISTRIBUTED_TOOLS_LABELS.XG_BOOST,
                    templatePrefix: 'xgboost-distributed-template-non-runnable',
                    description: 'XGBoost distributed non-runnable template for quick start testing'
                },
                {
                    testCase: 'T5331898: UI-Distributed template with MPI framework - non-runnable',
                    workerSpec: distMPINonRunnableSpec,
                    framework: DISTRIBUTED_TOOLS_LABELS.MPI,
                    templatePrefix: 'mpi-distributed-template-non-runnable',
                    description: 'MPI distributed non-runnable template for quick start testing'
                },
                {
                    testCase: 'T5331900: UI-Distributed template with TensorFlow framework - non-runnable',
                    workerSpec: distTensorFlowNonRunnableSpec,
                    framework: DISTRIBUTED_TOOLS_LABELS.TENSOR_FLOW,
                    templatePrefix: 'tensorflow-distributed-template-non-runnable',
                    description: 'TensorFlow distributed non-runnable template for quick start testing'
                },
                {
                    testCase: 'T5331902: UI-Distributed template with JAX framework - non-runnable',
                    workerSpec: distJAXNonRunnableSpec,
                    framework: DISTRIBUTED_TOOLS_LABELS.JAX,
                    templatePrefix: 'jax-distributed-template-non-runnable',
                    description: 'JAX distributed non-runnable template for quick start testing'
                }
            ];

            dataSets.forEach((caseData) => {
                test(`${caseData.testCase}`, async function () {
                    nonRunnableTemplateName = objectGenerator.generateShortNameByPrefix(
                        caseData.templatePrefix
                    );
                    videoName = 'quickstart-submission-non-runnable-template-training-advanced-setup';
                    await test.step('create a non-runnable distributed template', async function () {
                        const metaConfig = {
                            workloadType: WorkloadType.Distributed,
                            scopeType: ScopeType.Project,
                            scopeId: projectId.toString(),
                            templateName: nonRunnableTemplateName,
                            description: caseData.description
                        };

                        const distConfig =
                            caseData.framework === DISTRIBUTED_TOOLS_LABELS.JAX
                                ? undefined
                                : { masterSpecSameAsWorker: true };
                        const response = await templateUtils.createWorkloadTemplateV2(
                            metaConfig,
                            caseData.workerSpec,
                            distConfig
                        );
                        const templateId = response.body.meta.id;
                        logger.info(
                            `Created ${caseData.framework} non-runnable distributed template: ${nonRunnableTemplateName} with ID: ${templateId}`
                        );
                        nonRunnableTemplateId = templateId;
                    });

                    await test.step('create a distributed training with non-runnable template and verify advanced setup button is visible', async function () {
                        await trainingPage.navigateToWorkloads();
                        await trainingPage.openTrainingForm();
                        await trainingPage.selectClusterFromDropDown(flowRootState.cluster.name);
                        await trainingPage.workloadFormSelectProject(projectName);

                        await trainingPage.selectedMultiNodeOption();
                        await trainingPage.selectedDistributedTool(caseData.framework);
                        await trainingPage.workloadFormSelectTemplateV2(nonRunnableTemplateName);
                        await trainingPage.page.waitForTimeout(500);

                        await trainingPage.fillTrainingName(trainingName);
                        await trainingPage.waitForTrainingNameCount(trainingName);
                        // verify advanced setup button is visible for non-runnable template
                        expect(await hybridForm.isAdvancedSetupButtonVisible()).toBe(true);
                        // verify create training option is disabled
                        await hybridForm.openQuickSubmitDropdown();
                        expect(await hybridForm.isQuickSubmitDropdownSubmitOptionDisabled()).toBe(true);
                        expect(await hybridForm.isQuickSubmitDropdownContinueOptionDisabled()).toBe(false);
                    });
                });
            });
        });
    });
});
