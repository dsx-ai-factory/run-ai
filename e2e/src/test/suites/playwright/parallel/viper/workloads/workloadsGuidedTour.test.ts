/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import {
    flowRootState,
    objectGenerator,
    logger,
    tags,
    projectResourcesUtils,
    operations,
    settings,
    environmentUtils
} from '../../../../../../modules/helpers';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { test } from '@playwright/test';

import {
    shouldRunTests,
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import {
    GuidedTour,
    getWorkloadTourKeyTextPerStep
} from '../../../../../../modules/ui/runaiV2/pages/GuidedTour';

test.describe(`Workloads guided tour, tags: ${tags.tool.ui} ${tags.resource.workspace} ${tags.teams.viper}`, function () {
    supportedFrom(SupportedVersions.V2_23)(function () {
        let project: Project;
        let videoName: string;

        let guidedTour: GuidedTour;
        let browserEngine: BrowserEngine;
        let auth: AuthPage;

        let workloadName: string;

        test.beforeAll('creating project and namespace', async function () {
            browserEngine = new BrowserEngine();
            project = await projectResourcesUtils.createProjectAndNamespace('a1-guided-tour');
            logger.info(`created project ${project.name} ,namespace: "${project.namespace}"`);
        });

        test.afterAll('deleting project and namespace', async function () {
            try {
                await operations.deleteProjectAndNamespace(project.id);
                logger.info(`deleted project ${project.name} an namespace ${project.namespace}`);
            } catch (error) {
                logger.error(`failed to delete project ${project.name} an namespace ${project.namespace}`);
                throw error;
            }
        });

        test.beforeEach('start browser', async function () {
            await browserEngine.start();
            auth = new AuthPage(browserEngine.currentPage);
            guidedTour = new GuidedTour(browserEngine.currentPage);

            await auth.signIn(
                flowRootState.users.l2_researcher.email,
                flowRootState.users.l2_researcher.password
            );
            workloadName = objectGenerator.generateNameByPrefix('gtw');
        });

        test.afterEach('stop browser', async ({}, testInfo) => {
            try {
                await browserEngine.stopAndArchive(testInfo, 'guided-tour' + videoName);
            } catch (e) {
                logger.error(e);
            }
        });

        shouldRunTests(settings.isSelfHosted)(function () {
            test.describe('Trigger the guided tour from the help menu and submit a workspace', function () {
                let templateName: string;
                let tourSteps: ReturnType<typeof getWorkloadTourKeyTextPerStep>;

                test.beforeAll(async function () {
                    templateName = await environmentUtils.resolveTemplatePresetName('jupyter-lab-one-gpu');
                    tourSteps = getWorkloadTourKeyTextPerStep(templateName);
                });

                test('positive scenario expected flow', async function () {
                    videoName = 'gt-quick-submit';

                    await test.step('start the guided tour by clicking its button in the help menu', async () => {
                        await guidedTour.clickOnGuidedTourButtonInTheHelpMenu();
                        await guidedTour.clickOnGuidedStepsOptionJupyterNotebook();
                    });

                    await test.step('click create new workload', async () => {
                        await guidedTour.performGuidedTourStepLikeAHuman(
                            tourSteps.newWorkload,
                            async () => {
                                await guidedTour.clickOnNewWorkloadButton();
                            }
                        );
                    });

                    await test.step('select workspace option', async () => {
                        await guidedTour.performGuidedTourStepLikeAHuman(
                            tourSteps.selectWorkspace,
                            async () => {
                                await guidedTour.clickOnNewWorkspaceOption();
                            }
                        );
                    });

                    await test.step('select a project', async () => {
                        await guidedTour.performGuidedTourStepLikeAHuman(
                            tourSteps.selectProject,
                            async () => {
                                await guidedTour.workloadFormSelectProject(project.name, false);
                            }
                        );
                    });

                    await test.step('select a template', async () => {
                        await guidedTour.performGuidedTourStepLikeAHuman(
                            tourSteps.selectTemplate,
                            async () => {
                                await guidedTour.workloadFormSelectTemplateV2(templateName, false);
                            }
                        );
                    });

                    await test.step('fill the name input and click create workspace', async () => {
                        await guidedTour.performGuidedTourStepLikeAHuman(
                            tourSteps.nameAndCreateWorkspace,
                            async () => {
                                await guidedTour.fillWorkspaceName(workloadName);
                                await guidedTour.quickSubmitMainButton();
                            }
                        );
                    });

                    await test.step('click the checkbox of the created workspace in the table', async () => {
                        await guidedTour.performGuidedTourStepLikeAHuman(
                            tourSteps.selectCreatedWorkspace,
                            async () => {
                                await guidedTour.clickOnFirstWorkloadCheckbox();
                            }
                        );
                    });

                    await test.step('click connect action button', async () => {
                        await guidedTour.performGuidedTourStepLikeAHuman(
                            tourSteps.connect,
                            async () => {
                                await guidedTour.waitForWorkloadUIStatus(
                                    workloadName,
                                    'Running',
                                    undefined,
                                    false
                                );
                                await guidedTour.clickOnConnectWorkloadButton();
                            }
                        );
                    });

                    await test.step('wait for the guided tour to announce that it is completed', async () => {
                        await guidedTour.waitForCompletionModal();
                        await guidedTour.closeCompletionModal();
                    });
                });

                test('positive scenario - cancel guided tour', async function () {
                    videoName = 'gt-cancel';

                    await guidedTour.clickOnGuidedTourButtonInTheHelpMenu();
                    await guidedTour.clickOnGuidedStepsOptionJupyterNotebook();

                    await test.step('click create new workload', async () => {
                        await guidedTour.performGuidedTourStepLikeAHuman(
                            tourSteps.newWorkload,
                            async () => {
                                await guidedTour.clickOnNewWorkloadButton();
                            }
                        );
                    });

                    await test.step('open the exit modal and confirm exit', async () => {
                        await guidedTour.openExitTourModal();
                        await guidedTour.confirmExitTourModal();
                    });
                });
            });
        });
    });
});
