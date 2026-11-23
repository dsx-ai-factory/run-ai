/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    flowRootState,
    objectGenerator,
    tags,
    projectUtils,
    logger,
    operations,
} from '../../../../../../modules/helpers';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test, expect } from '@playwright/test';
import { runaiCliV2 } from '../../../../../../modules/cli';
import { ProjectPageV2 } from '../../../../../../modules/ui/runaiV2/pages/ProjectPageV2';

test.describe(`Department E2E Tests, tags: ${tags.teams.loong} ${tags.tool.ui} ${tags.resource.department}`, () => {
    supportedFrom(SupportedVersions.V2_20)(function () {
        let browserEngine: BrowserEngine;
        let authPage: AuthPage;
        let projectPage: ProjectPageV2;
        let projectId: number;
        let videoName: string;
        let createdProjectIds: number[] = [];

        test.beforeAll(function () {
            browserEngine = new BrowserEngine();
        });

        test.beforeEach(async function () {
            await browserEngine.start();
            authPage = new AuthPage(browserEngine.currentPage);
            projectPage = new ProjectPageV2(browserEngine.currentPage);
        });

        async function trackCreatedProject(projectName: string): Promise<void> {
            try {
                projectId = await projectUtils.getProjectIdByProjectName(projectName);
                createdProjectIds.push(projectId);
                logger.info(`project ${projectName} created with ID: ${projectId}`);
            } catch (error) {
                logger.error(`failed to get project ID for ${projectName}: ${error}`);
            }
        }

        test.afterEach('stop browser and cleanup projects', async ({}, testInfo) => {
            await browserEngine.stopAndArchive(testInfo, videoName);

            for (const projectId of createdProjectIds) {
                if (projectId) {
                    try {
                        await operations.deleteProjectAndNamespace(projectId);
                        logger.info(`deleted project with ID ${projectId} and its namespace`);
                    } catch (error) {
                        logger.error(
                            `failed to delete project with ID ${projectId} and its namespace: ${error}`
                        );
                    }
                }
            }
            createdProjectIds = [];
        });

        test('T3929487: Department administrator can list all projects under the department', async () => {
            videoName = 'dept-admin-list-projects';
            const projectNames: string[] = [];

            await test.step('login as department administrator', async () => {
                await authPage.signIn(
                    flowRootState.users.department_admin.email,
                    flowRootState.users.department_admin.password
                );
            });

            await test.step('Create multiple projects under default department', async () => {
                for (let i = 1; i <= 3; i++) {
                    const projName = objectGenerator.generateShortNameByPrefix(`dept-proj-${i}`, 3);
                    projectNames.push(projName);

                    await projectPage.goToPage();
                    await projectPage.createProject(projName, undefined, [{ default: 1 }]);

                    await trackCreatedProject(projName);
                }
            });

            await test.step('Configure user as department administrator for default department', () => {
                runaiCliV2.loginAsUser(
                    flowRootState.users.department_admin.email,
                    flowRootState.users.department_admin.password
                );
            });

            await test.step('List projects with CLI and verify output contains all projects', () => {
                const listOutput = runaiCliV2.list('project');
                for (const projectName of projectNames) {
                    expect(listOutput.stdout).toContain(projectName);
                }
            });

            await test.step('Cleanup: Delete all created projects', async () => {
                for (const projectName of projectNames) {
                    try {
                        await projectPage.goToPage();
                        await projectPage.deleteProject(projectName);
                    } catch (error) {
                        console.warn(`Failed to delete project ${projectName}:`, error);
                    }
                }
            });
        });
    });
});
