/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { BrowserEngine } from '../../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { WorkspacePage } from '../../../../../../../modules/ui/runaiV2/pages/WorkspacePage';
import { WorkloadPage, EWorkloadStatus } from '../../../../../../../modules/ui/runaiV2/pages/WorkloadPage';
import { EnvironmentPage } from '../../../../../../../modules/ui/runaiV2/pages/EnvironmentPage';
import { ComputeResourcePage } from '../../../../../../../modules/ui/runaiV2/pages/ComputeResourcePage';
import { AuthPage } from '../../../../../../../modules/ui/runaiV2/pages/AuthPage';
import {
    flowRootState,
    objectGenerator,
    logger,
    tags,
    waitUtils,
    projectResourcesUtils,
    Utils,
    operations,
    environmentUtils,
    computeResourceUtils,
    workspaceTemplateUtils,
    settings,
    unleashClient,
    apiUtils,
    versionUtils
} from '../../../../../../../modules/helpers';
import type { WorkloadTemplateCreationRequest } from '../../../../../../../modules/api/apis/assetsApi';
import { Scope } from '../../../../../../../modules/api/apis/assetsApi';
import { Project } from '../../../../../../../modules/api/apis/projectsDepartmentsApi';
import { test } from '@playwright/test';
import { EFeatureToggleKeys } from '../../../../../../../modules/helpers/utils/Unleash/featureToggle';
import {
    AnyWorkloadSubmissionType,
    isAnyWorkloadSubmissionEnabled
} from '../../../../../../../modules/helpers/utils/AnyWorkloadSubmissionUtils';
import { kubeCli } from '../../../../../../../modules/cli';

test.describe(`Workspaces UI, tags: ${tags.tool.ui} ${tags.resource.workspace} ${tags.teams.viper}`, function () {
    let projectName: string;
    let namespace: string;
    let project: Project;
    let videoName: string;

    let browserEngine: BrowserEngine;
    let auth: AuthPage;
    let workspace: WorkspacePage;
    let workload: WorkloadPage;
    let environment: EnvironmentPage;
    let computeResource: ComputeResourcePage;
    let workspaceName: string;
    let claimName: string;
    let anyWorkloadSubmissionEnabled: boolean;
    let kubernetesServerVersion: string | undefined;
    let isAffectedByJobResumeBug = false;

    test.beforeAll('creating project and namespace', async function () {
        browserEngine = new BrowserEngine();
        projectName = objectGenerator.generateNameByPrefix('project-e2e-for-workspace');
        namespace = Utils.namespaceFromProjectName(projectName);
        project = await projectResourcesUtils.createProject(namespace, {
            name: projectName,
            deservedGpus: 1,
            maxAllowedGpus: 1
        });
        if (!settings.isSelfHosted) {
            await unleashClient.disableFeature(EFeatureToggleKeys.SHOW_HYBRID_SUBMISSION);
        }
        await apiUtils.disableHybridSubmission();
        anyWorkloadSubmissionEnabled = await isAnyWorkloadSubmissionEnabled(
            AnyWorkloadSubmissionType.Workspace
        );
        if (anyWorkloadSubmissionEnabled) {
            kubernetesServerVersion = await kubeCli.getServerVersion();
            isAffectedByJobResumeBug =
                versionUtils.isKubernetesVersionAffectedByJobResumeBug(kubernetesServerVersion);
        }
        logger.info(`created project ${projectName}`);
    });

    test.afterAll('deleting project and namespace', async function () {
        try {
            await operations.deleteProjectAndNamespace(project.id);
            if (!settings.isSelfHosted) {
                await unleashClient.enableFeature(EFeatureToggleKeys.SHOW_HYBRID_SUBMISSION);
            }
            await apiUtils.enableHybridSubmission();
            logger.info(`deleted project ${projectName} an namespace ${namespace}`);
        } catch (error) {
            logger.error(`failed to delete project ${projectName} an namespace ${namespace}`);
            throw error;
        }
    });

    test.beforeEach('start browser', async function () {
        await browserEngine.start();
        auth = new AuthPage(browserEngine.currentPage);
        workspace = new WorkspacePage(browserEngine.currentPage);
        workload = new WorkloadPage(browserEngine.currentPage);
        await auth.signIn(
            flowRootState.users.system_administrator.email,
            flowRootState.users.system_administrator.password
        );
        workspaceName = objectGenerator.generateNameByPrefix('workspace');
    });

    test.afterEach('stop browser', async ({}, testInfo) => {
        try {
            await browserEngine.stopAndArchive(testInfo, 'workspaces-' + videoName);
        } catch (e) {
            logger.error(e);
        }
    });

    async function fillFirstNewWorkspacePage(): Promise<void> {
        await workload.navigateToWorkloads();
        await workload.openWorkspaceForm();

        await workload.selectClusterFromDropDown(flowRootState.cluster.name);
        await workload.workloadFormSelectProject(projectName);
        await workspace.fillWorkspaceName(workspaceName);
        await workspace.waitForWorkspaceNameCount(workspaceName);
        await workspace.selectAssetsSubmissionOption();
        await workspace.clickOnContinue();
    }

    async function createWorkspaceWithWorkloadVolume(isEphemeralVolume?: boolean): Promise<void> {
        await fillFirstNewWorkspacePage();
        const envName = (await environmentUtils.presetNames()).tensorboard;
        await workspace.selectEnvironment(envName);
        await workspace.selectComputeResource('cpu-only');
        await workspace.openVolumeSection();
        await workspace.addWorkloadVolume(isEphemeralVolume);
        await workspace.clickOnCreateWorkspaceBtn();
    }

    async function waitForPvc(workspaceName: string, namespace: string): Promise<void> {
        claimName = await waitUtils.getFullClaimName(workspaceName, namespace);
        await waitUtils.getPvc(claimName, namespace);
    }

    async function waitForWorkspaceResource(shouldExist = true): Promise<void> {
        if (anyWorkloadSubmissionEnabled) {
            await waitUtils.waitForBatchJobToExist(workspaceName, namespace, shouldExist);
        } else {
            await waitUtils.waitForInteractiveWorkloadToExist(workspaceName, namespace, shouldExist);
        }
    }

    test.describe('Create new workspace', function () {
        let environmentName: string;
        let computeName: string;

        test.beforeEach('setup', function () {
            environment = new EnvironmentPage(browserEngine.currentPage);
            computeResource = new ComputeResourcePage(browserEngine.currentPage);
            environmentName = objectGenerator.generateNameByPrefix('environment');
            computeName = objectGenerator.generateNameByPrefix('compute');
        });

        async function createTemplate(TemplateName: string): Promise<string> {
            const workspaceTemplateCreationRequest: WorkloadTemplateCreationRequest = {
                meta: {
                    name: TemplateName,
                    scope: Scope.Project,
                    projectId: project.id,
                    clusterId: project.clusterUuid
                },
                spec: {
                    assets: {
                        environment: await environmentUtils.getId('tensorboard'),
                        compute: await computeResourceUtils.getId('cpu-only')
                    },
                    specificEnv: {
                        environmentVariables: []
                    }
                }
            };
            const template = await workspaceTemplateUtils.create(workspaceTemplateCreationRequest);
            return template.meta.id;
        }

        test('should create new workspace from scratch', async function () {
            videoName = 'create-workspace-from-scratch';
            if (settings.isSelfHosted) test.setTimeout(180 * 1000);
            await fillFirstNewWorkspacePage();

            await workspace.clickOnNewEnvironmentBtn();
            await environment.createNewEnvironment(environmentName, projectName, {
                waitForProjectScope: true,
                command: 'sleep infinity'
            });

            await workspace.clickOnNewComputeBtn();
            await computeResource.createNewCompute(computeName, projectName);

            await workspace.clickOnCreateWorkspaceBtn();
            await waitForWorkspaceResource();
            await workload.waitForWorkloadUIStatus(workspaceName, EWorkloadStatus.Running);

            //cleanup
            await workload.deleteWorkload(workspaceName);
            await waitForWorkspaceResource(false);
            await computeResource.goToPage();
            await computeResource.removeCompute(computeName);
            await environment.goToPage();
            await environment.removeEnvironment(environmentName);
        });

        test('copy and edit workspace', async function () {
            videoName = 'copy-and-edit-workspace ';
            const copiedWorkspace = objectGenerator.generateShortNameByPrefix('copy-workspace');
            await fillFirstNewWorkspacePage();
            const envName = (await environmentUtils.presetNames()).tensorboard;
            await environment.selectEnvironment(envName);
            await computeResource.selectComputeResource('cpu-only');
            await workspace.clickOnCreateWorkspaceBtn();
            await workspace.waitForWorkloadUIStatus(workspaceName, EWorkloadStatus.Running);

            //copy and edit
            await workspace.toggleWorkload(workspaceName);
            await workspace.clickOnCopyAndEditWorkload();
            await workspace.fillWorkspaceName(copiedWorkspace);
            await workspace.clickOnCreateWorkspaceBtn();
            await workspace.waitForWorkloadUIStatus(copiedWorkspace, EWorkloadStatus.Running);

            await workspace.deleteWorkload(workspaceName);
            await workspace.deleteWorkload(copiedWorkspace);
        });

        test('should create new workspace from existing template', async function () {
            videoName = 'create-workspace-from-template';
            const templateName = objectGenerator.generateNameByPrefix('template');
            const templateId = await createTemplate(templateName);

            await workload.navigateToWorkloads();
            await workload.openWorkspaceForm();
            await workload.selectClusterFromDropDown(flowRootState.cluster.name);
            await workload.workloadFormSelectProject(projectName);
            await workspace.selectTemplate(templateName);
            await workspace.fillWorkspaceName(workspaceName);
            await workspace.waitForWorkspaceNameCount(workspaceName);
            await workspace.clickOnContinue();
            await workspace.clickOnCreateWorkspaceBtn();
            await workload.waitForWorkloadUIStatus(workspaceName, EWorkloadStatus.Running);

            // cleanup
            await workload.deleteWorkload(workspaceName);
            await workspaceTemplateUtils.delete(templateId);
        });
    });

    test('should create workspace with Ephemeral workload volumes and make sure it behave as expected', async () => {
        videoName = 'create-workspace-with-ephemral-volume';
        await createWorkspaceWithWorkloadVolume(true);

        await workload.waitForWorkloadUIStatus(workspaceName, EWorkloadStatus.Running);
        // Checking that the pvc is created when the workspace is active
        await waitForPvc(workspaceName, namespace);
        logger.info(`workload volume for ${workspaceName} has created.`);

        await workload.stopWorkloadByName(workspaceName);
        await workload.waitForWorkloadUIStatus(workspaceName, EWorkloadStatus.Stopped);

        await test.step('verify the workspace and its PVC recover after activation', async function (step) {
            step.skip(
                anyWorkloadSubmissionEnabled && isAffectedByJobResumeBug,
                `Kubernetes ${kubernetesServerVersion} is affected by the Job resume bug`
            );

            await workload.activateWorkloadByName(workspaceName);
            await workload.waitForWorkloadUIStatus(workspaceName, EWorkloadStatus.Running);

            // Checking that the pvc is created again when the workspace is re-activated
            await waitForPvc(workspaceName, namespace);
            logger.info(`workload volume for ${workspaceName} has created again.`);
        });

        await workload.deleteWorkload(workspaceName);
        await waitForWorkspaceResource(false);
        // Checking that the pvc is deleted after the workspace is deleted
        await waitUtils.waitForPvc(claimName, namespace, 90, false);
    });

    test('should create workspace with Persistent workload volumes and make sure it behave as expected', async function () {
        videoName = 'create-workspace-with-persistent-volumes';
        test.skip(
            anyWorkloadSubmissionEnabled,
            'Persistent PVC assets conflict with WorkloadRunner ownership'
        );

        await createWorkspaceWithWorkloadVolume(false);
        await workload.waitForWorkloadUIStatus(workspaceName, EWorkloadStatus.Running);

        // Checking that the pvc is created when the workspace is active
        await waitForPvc(workspaceName, namespace);

        await workload.stopWorkloadByName(workspaceName);
        await workload.waitForWorkloadUIStatus(workspaceName, EWorkloadStatus.Stopped);

        await waitForPvc(workspaceName, namespace);
        logger.info(`workload volume for ${workspaceName} exists.`);

        // cleanup
        await workload.deleteWorkload(workspaceName);
        // Checking that the pvc is deleted when the workspace is deleted
        await waitUtils.waitForPvc(claimName, namespace, 90, false);
    });
});
