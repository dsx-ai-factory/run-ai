/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect, test } from '@playwright/test';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { BrowserEngine } from '../../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { AuthPage } from '../../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { WorkloadPage } from '../../../../../../../modules/ui/runaiV2/pages/WorkloadPage';
import { WorkspacePage } from '../../../../../../../modules/ui/runaiV2/pages/WorkspacePage';
import { Project } from '../../../../../../../modules/api/apis/orgUnitApi';
import { Utils } from '../../../../../../../modules/helpers/utils/utils';
import {
    apiUtils,
    flowRootState,
    logger,
    objectGenerator,
    operations,
    orgUnitUtils,
    projectUtils,
    waitUtils,
    workloadsUtilsV2
} from '../../../../../../../modules/helpers';
import { WorkloadType } from '../../../../../../../modules/helpers/utils/WorkloadsUtilsV2';
import {
    AnyWorkloadSubmissionType,
    isAnyWorkloadSubmissionEnabled
} from '../../../../../../../modules/helpers/utils/AnyWorkloadSubmissionUtils';
test.describe('Workload submission with NGC catalog image', () => {
    supportedFrom(SupportedVersions.V2_23)(function () {
        let browserEngine: BrowserEngine;
        let auth: AuthPage;

        let workloads: WorkloadPage;
        let workspace: WorkspacePage;

        let projectName: string;
        let namespace: string;
        let project: Project;

        const ngcImage = 'nvidia/pytorch';

        async function initBrowserAndLogin(): Promise<void> {
            await browserEngine.start();
            auth = new AuthPage(browserEngine.currentPage);
            workspace = new WorkspacePage(browserEngine.currentPage);
            workloads = new WorkloadPage(browserEngine.currentPage);

            await auth.signIn(
                flowRootState.users.system_administrator.email,
                flowRootState.users.system_administrator.password
            );
        }

        test.beforeAll(async function () {
            try {
                await operations.deleteAllProjects();
            } catch (e) {
                logger.error(e);
            }
            browserEngine = new BrowserEngine();
            projectName = objectGenerator.generateShortNameByPrefix('project-ff');
            namespace = Utils.namespaceFromProjectName(projectName);
            project = await orgUnitUtils.createProject({
                name: projectName,
                clusterId: flowRootState.cluster.uuid,
                resources: []
            });

            await apiUtils.enableNgcCatalog();

            const roleBindingName = projectUtils.getSystemAdministratorRoleBindingName();
            await projectUtils.waitForRoleBindingForUserInProject(
                namespace,
                flowRootState.users.system_administrator.email,
                roleBindingName,
                true
            );

            await initBrowserAndLogin();
        });

        test.afterAll(async function () {
            await apiUtils.disableNgcCatalog();
            await operations.deleteProjectAndNamespace(Number(project.id));
        });

        test('should submit a workload with NGC catalog image', async () => {
            await workloads.goToPage();
            await workloads.openWorkspaceForm();

            // First Page
            const workspaceName = objectGenerator.generateShortNameByPrefix('workspace');
            await workloads.workloadFormSelectProject(projectName);
            await workloads.workloadFormSelectTemplate();
            await workspace.fillWorkspaceName(workspaceName);
            await workspace.waitForWorkspaceNameCount(workspaceName);
            await workspace.selectAssetsSubmissionOption();
            await workspace.clickOnContinue();

            await workspace.clickOnElementByAidName('radio-option-Select from the NGC catalog');
            await workspace.selectNgcImage(ngcImage);
            await workspace.waitForImageTagsToLoad();
            await workspace.selectNgcTag();
            await workspace.clickOnCreateWorkspaceBtn();
            await workspace.waitForWorkloadToAppear(workspaceName);
            let workload = await workloadsUtilsV2.waitForBackendWorkloadByNameAndProjectName(
                workspaceName,
                projectName
            );

            if (await isAnyWorkloadSubmissionEnabled(AnyWorkloadSubmissionType.Workspace)) {
                await waitUtils.waitForCondition(
                    async () => {
                        workload = await workloadsUtilsV2.getWorkloadByNameAndProjectName(
                            workspaceName,
                            projectName
                        );
                        return (
                            workload?.images?.some((image) => image.includes(`nvcr.io/${ngcImage}`)) ?? false
                        );
                    },
                    `wait for workload ${workspaceName} NGC image to be reported`,
                    true,
                    60,
                    2
                );
            }

            expect(workload.images[0]).toContain(`nvcr.io/${ngcImage}`);

            // Cleanup - deleting workload
            try {
                await workloadsUtilsV2.deleteWorkload(workload.id, WorkloadType.WORKSPACE);
            } catch (e) {
                // We only log the error, and then in the afterAll when the project and namespace are deleted, the workload will be deleted as well.
                logger.error(`Failed to delete workload ${workload.name} with id ${workload.id}`, e);
            }
        });
    });
});
