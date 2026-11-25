/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import {
    objectGenerator,
    Utils,
    logger,
    projectResourcesUtils,
    flowRootState,
    operations,
    waitUtils,
    projectUtils,
    templateUtils
} from '../../../../../../modules/helpers';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { WorkloadPage } from '../../../../../../modules/ui/runaiV2/pages/WorkloadPage';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { tags } from '../../../../../../modules/helpers/config/tags';
import { supportedFrom, SupportedVersions } from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { TrainingPage } from '../../../../../../modules/ui/runaiV2/pages/TrainingPage';
import { HybridForm } from '../../../../../../modules/ui/runaiV2/pages/HybridSubmissionForm';
import { ProjectPage } from '../../../../../../modules/ui/runaiV2/pages/ProjectPage';
import { Scope, WorkloadType } from '../../../../../../modules/api/apis/assetsApi';
import { TemplatePageV2 } from '../../../../../../modules/ui/runaiV2/pages/TemplatePageV2';

test.describe(`Incorporate keyboard actions across system, tags: ${tags.tool.ui} ${tags.resource.workspace} ${tags.resource.templateV2} ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_23)(function () {
        let browserEngine: BrowserEngine;
        let auth: AuthPage;
        let workloads: WorkloadPage;
        let training: TrainingPage;
        let templateName: string;
        let templatePageV2: TemplatePageV2;
        let trainingName: string;
        let projectName: string;
        let namespace: string;
        let project;
        let videoName: string;
        let hybridForm: HybridForm;
        let projectPage: ProjectPage;
        let createdTemplates: any[] = [];
        let createdWorkloads: any[] = [];
        let createdProjects: string[] = [];

        test.beforeAll(async function () {
            browserEngine = new BrowserEngine();
            projectName = objectGenerator.generateNameByPrefix('keyboard-project');
            namespace = Utils.namespaceFromProjectName(projectName);
            project = await projectResourcesUtils.createProject(namespace, {
                name: projectName,
                deservedGpus: 1,
                maxAllowedGpus: 1
            });
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

        test.beforeEach(async () => {
            createdTemplates = [];
            createdWorkloads = [];
            createdProjects = [];
            browserEngine = new BrowserEngine();
            await browserEngine.start();
            auth = new AuthPage(browserEngine.currentPage);
            workloads = new WorkloadPage(browserEngine.currentPage);
            training = new TrainingPage(browserEngine.currentPage);
            templatePageV2 = new TemplatePageV2(browserEngine.currentPage);
            hybridForm = new HybridForm(browserEngine.currentPage);
            projectPage = new ProjectPage(browserEngine.currentPage);
            trainingName = objectGenerator.generateNameByPrefix('keyboard-training');
            templateName = objectGenerator.generateNameByPrefix('keyboard-template');
        });

        test.afterEach(async ({}, testInfo) => {
            // Delete templates
            for (const template of createdTemplates) {
                try {
                    const templateId = await templateUtils.getIdByName(template);
                    await templateUtils.deleteWorkloadTemplateV2(
                        templateId,
                        WorkloadType.Workspace,
                        flowRootState.users.system_administrator.idToken
                    );
                    logger.info(`Successfully deleted template: ${template}`);
                } catch (e) {
                    logger.error('Error deleting template:', template, e);
                }
            }

            // Delete workloads
            for (const workload of createdWorkloads) {
                try {
                    await workloads.deleteWorkload(workload);
                    logger.info(`Successfully deleted workload: ${workload}`);
                } catch (e) {
                    logger.error('Error deleting workload:', workload, e);
                }
            }

            for (const projectName of createdProjects) {
                try {
                    const projectId = await projectUtils.getProjectIdByProjectName(projectName);
                    if (projectId) {
                        await operations.deleteProjectAndNamespace(projectId);
                        logger.info(`Successfully deleted project: ${projectName}`);
                    }
                } catch (e) {
                    logger.error('Error deleting project:', projectName, e);
                }
            }

            await browserEngine.stopAndArchive(testInfo, videoName || 'incorporate-keyboard-test');
        });

        test('T5397030: Enter Key for New workload Creation', async () => {
            videoName = 'enter-key-for-new-workload';
            await auth.signIn(
                flowRootState.users.l1_researcher.email,
                flowRootState.users.l1_researcher.password
            );
            // Create a Training workload
            await training.navigateToWorkloads();
            await training.openTrainingForm();
            await training.selectClusterFromDropDown(flowRootState.cluster.name);
            await training.workloadFormSelectProject(projectName);
            await training.workloadFormSelectTemplate();
            await training.fillTrainingName(trainingName);
            await training.waitForTrainingNameCount(trainingName);
            await training.focusTrainingName();
            await training.pressEnterOnTrainingName();

            await hybridForm.fillImageSection('ubuntu', 'Always pull the image from the registry');
            await hybridForm.fillRuntimeSettings('sleep infinity', '');
            await hybridForm.fillComputeSection('one-gpu');
            await hybridForm.focusGpuDevicesInput();
            await hybridForm.pressEnterOnGpuDevicesInput();
            createdWorkloads.push(trainingName);

            await waitUtils.waitForElementToExist(workloads, `td:has-text("${trainingName}")`);
        });

        test('T5397031: ENTER Key Functionality for Textarea', async () => {
            videoName = 'enter-key-for-textarea';
            const localProjectName = objectGenerator.generateNameByPrefix('enter-key-project');
            await auth.signIn(
                flowRootState.users.system_administrator.email,
                flowRootState.users.system_administrator.password
            );
            await projectPage.goToPage();
            await projectPage.clickOnAddNewProjectBtn();
            await projectPage.fillInputBySelector(projectPage.projectNameInput, localProjectName);
            // Verify a new line created in the textarea after pressing Enter
            await projectPage.fillProjectDescriptionInput('');
            await projectPage.pressEnterOnProjectDescriptionInput();

            const value = await projectPage.getProjectDescriptionInputValueWithEscapes();
            const lineCount = await projectPage.getProjectDescriptionLineCount();
            expect(value).toBe('\\n');
            expect(lineCount).toBe(2);

            // Verify the form is not submitted
            await projectPage.locator(projectPage.projectNameInput).isVisible();
        });

        test('T5397032: ENTER Key Functionality when not focus on input', async () => {
            videoName = 'enter-key-not-focus-on-input';
            const localProjectName = objectGenerator.generateNameByPrefix('enter-key-project');
            await auth.signIn(
                flowRootState.users.system_administrator.email,
                flowRootState.users.system_administrator.password
            );
            await projectPage.goToPage();
            await projectPage.clickOnAddNewProjectBtn();
            await projectPage.fillInputBySelector(projectPage.projectNameInput, localProjectName);
            await projectPage.pressEnterKey();
            await projectPage.locator(projectPage.projectNameInput).isVisible();
        });

        test('T5397033: ENTER Key for Browser "Leave Site?" Dialog', async () => {
            test.setTimeout(300 * 1000);
            videoName = 'enter-key-leave-site-dialog';
            await auth.signIn(
                flowRootState.users.l1_researcher.email,
                flowRootState.users.l1_researcher.password
            );

            await test.step('Open training workload form and make changes', async () => {
                await training.navigateToWorkloads();
                await training.openTrainingForm();
                await training.selectClusterFromDropDown(flowRootState.cluster.name);
                await training.workloadFormSelectProject(projectName);
                await training.fillTrainingName(trainingName);
            });

            await test.step('Focus on cancel button and press Enter Key', async () => {
                await workloads.page.focus(workloads.cancelBtn);
                await workloads.pressEnterKey();
            });

            await test.step('Verify the leave confirmation dialog appears', async () => {
                await workloads.verifyLeavePageDialogAppears();
            });
        });

        test('T5397034: Esc Key for any open dialog', async () => {
            videoName = 'esc-key-dialog';
            await auth.signIn(
                flowRootState.users.l1_researcher.email,
                flowRootState.users.l1_researcher.password
            );

            await test.step('Open training workload form and make changes', async () => {
                await training.navigateToWorkloads();
                await training.openTrainingForm();
                await training.selectClusterFromDropDown(flowRootState.cluster.name);
                await training.workloadFormSelectProject(projectName);
                await training.fillTrainingName(trainingName);
            });

            await test.step('Click cancel button to trigger leave confirmation dialog', async () => {
                await workloads.page.locator(workloads.cancelBtn).click();
            });

            await test.step('Press Enter Key to close the dialog', async () => {
                await workloads.pressEnterKey();
            });

            await test.step('Verify the leave confirmation dialog was closed', async () => {
                const leaveDialog = workloads.page.locator(workloads.leavePageDialog);
                await expect(leaveDialog).not.toBeVisible();
            });
        });

        test('T5397035: Esc Key for Dropdown', async () => {
            videoName = 'esc-key-dropdown';
            await auth.signIn(
                flowRootState.users.system_administrator.email,
                flowRootState.users.system_administrator.password
            );
            await test.step('Create a training workload', async () => {
                await training.navigateToWorkloads();
                await training.openTrainingForm();
                await training.selectClusterFromDropDown(flowRootState.cluster.name);
                await training.workloadFormSelectProject(projectName);
                await training.workloadFormSelectTemplate();
                await training.fillTrainingName(trainingName);
                await training.waitForTrainingNameCount(trainingName);
                await training.focusTrainingName();
                await training.pressEnterOnTrainingName();
            });

            await test.step('Press esc key when selecting tool on Tools section', async () => {
                await hybridForm.clickOnSubSectionByLabel('Tools');
                await hybridForm.clickOnAddNewToolButton();
                await hybridForm.page
                    .locator(hybridForm.hybridEnvironmentSection.toolTypeSelect)
                    .nth(0)
                    .click();
                await hybridForm.pressEscKey();
            });

            await test.step('Verify the dropdown of tools was closed', async () => {
                const pytorchOption = hybridForm.page.getByRole('option', { name: 'PyTorch' });
                await expect(pytorchOption).not.toBeVisible();
            });
        });

        test('T5397036: Esc Key for Drawer', async () => {
            videoName = 'esc-key-drawer';
            await auth.signIn(
                flowRootState.users.l1_researcher.email,
                flowRootState.users.l1_researcher.password
            );
            await workloads.navigateToWorkloads();
            await workloads.clickOnNewWorkloadButton();
            await workloads.pressEscKey();
            await expect(workloads.page.locator(workloads.newTrainingBtn)).not.toBeVisible();
        });

        test('T5397037: Enter Key for training workload first page', async () => {
            videoName = 'enter-key-training-workload-first-page';
            await auth.signIn(
                flowRootState.users.l1_researcher.email,
                flowRootState.users.l1_researcher.password
            );

            await test.step('Open training workload form and fill first page', async () => {
                await training.navigateToWorkloads();
                await training.openTrainingForm();
                await training.selectClusterFromDropDown(flowRootState.cluster.name);
                await training.workloadFormSelectProject(projectName);
                await training.workloadFormSelectTemplate();
                await training.fillTrainingName(trainingName);
                await training.waitForTrainingNameCount(trainingName);
            });

            await test.step('Press Enter Key to navigate to second page', async () => {
                await training.focusTrainingName();
                await training.pressEnterOnTrainingName();
            });

            await test.step('Verify navigated to second page', async () => {
                await expect(
                    hybridForm.page.locator(hybridForm.hybridEnvironmentSection.sectionSelector)
                ).toBeVisible();
            });
        });

        test('T5397038: Enter Key for template creation', async () => {
            videoName = 'enter-key-template-creation';
            await auth.signIn(
                flowRootState.users.system_administrator.email,
                flowRootState.users.system_administrator.password
            );

            await test.step('Create a template', async () => {
                await templatePageV2.goToPage();
                await templatePageV2.openCreationForm(WorkloadType.Workspace);
                await templatePageV2.assetScopeSection.selectScope({
                    scope: Scope.Project,
                    scopeEntityName: projectName
                });
                await templatePageV2.fillName(templateName);
                await waitUtils.sleep(2000);
                await templatePageV2.pressEnterKey();

                await templatePageV2.hybridForm.selectExistingEnv('tensorboard-tensorflow');
                await templatePageV2.hybridForm.selectExistingComputeResource('one-gpu');
            });

            await test.step('Press Enter Key', async () => {
                await templatePageV2.hybridForm.focusGpuDevicesInput();
                await templatePageV2.pressEnterKey();
            });

            await test.step('Verify template was created', async () => {
                await templatePageV2.waitForTemplateExist(templateName);
            });
            createdTemplates.push(templateName);
        });
    });
});
