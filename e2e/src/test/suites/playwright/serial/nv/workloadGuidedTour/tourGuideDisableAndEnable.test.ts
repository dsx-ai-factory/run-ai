/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import {
    flowRootState,
    logger,
    tags,
    projectResourcesUtils,
    projectUtils,
    operations
} from '../../../../../../modules/helpers';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { test } from '@playwright/test';

import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { GuidedTour } from '../../../../../../modules/ui/runaiV2/pages/GuidedTour';

test.describe(`UI-Tour guide disable and enable, tags: ${tags.tool.ui} ${tags.resource.workspace} ${tags.teams.loong}`, function () {
    supportedFrom(SupportedVersions.V2_23)(function () {
        let project: Project;
        let videoName: string = '';

        let guidedTour: GuidedTour;
        let browserEngine: BrowserEngine;
        let auth: AuthPage;

        test.beforeAll('cleanup and setup', async function () {
            browserEngine = new BrowserEngine();

            // Clean up all existing projects and namespaces
            try {
                const projects = await projectUtils.getProjects();
                for (const proj of projects) {
                    try {
                        await operations.deleteProjectAndNamespace(proj.id);
                        logger.info(`deleted project ${proj.name} and namespace ${proj.namespace}`);
                    } catch (error) {
                        logger.warn(`failed to delete project ${proj.name}: ${error}`);
                    }
                }
            } catch (error) {
                logger.warn(`failed to cleanup projects: ${error}`);
            }
        });

        test.afterAll('deleting project and namespace', async function () {
            if (project && project.id) {
                try {
                    await operations.deleteProjectAndNamespace(project.id);
                    logger.info(`deleted project ${project.name} an namespace ${project.namespace}`);
                } catch (error) {
                    logger.error(
                        `failed to delete project ${project.name} and namespace ${project.namespace}`
                    );
                    throw error;
                }
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
        });

        test.afterEach('stop browser', async ({}, testInfo) => {
            try {
                await browserEngine.stopAndArchive(testInfo, videoName || 'tour-guide-disable-enable');
            } catch (e) {
                logger.error(e);
            }
        });

        test('T5357201: UI-Tour guide disable and enable', async function () {
            videoName = 'tour-guide-disable-enable';
            await test.step('verify the tour button from the help button at the top right corner of the app is disabled status', async () => {
                await guidedTour.verifyGuidedTourButtonState(true);
            });

            await test.step('create one project', async () => {
                project = await projectResourcesUtils.createProjectAndNamespace(
                    'a1-tour-guide-disable-enable'
                );
                logger.info(`created project ${project.name} ,namespace: "${project.namespace}"`);
            });

            await test.step('refresh the page', async () => {
                await guidedTour.page.reload();
                await guidedTour.page.waitForLoadState('networkidle');
            });

            await test.step('verify the tour button status change to enabled', async () => {
                await guidedTour.verifyGuidedTourButtonState(false);
            });
        });
    });
});
