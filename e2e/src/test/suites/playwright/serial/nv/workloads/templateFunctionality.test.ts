/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import {
    dataSourcesUtils,
    objectGenerator,
    Utils,
    waitUtils,
    logger,
    projectResourcesUtils,
    flowRootState,
    operations,
    environmentUtils,
    unleashClient,
    apiUtils,
    config
} from '../../../../../../modules/helpers';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { WorkloadPage, EWorkloadStatus } from '../../../../../../modules/ui/runaiV2/pages/WorkloadPage';
import { TemplatePage } from '../../../../../../modules/ui/runaiV2/pages/TemplatePage';
import { WorkspacePage } from '../../../../../../modules/ui/runaiV2/pages/WorkspacePage';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { tags } from '../../../../../../modules/helpers/config/tags';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import {
    WorkloadType,
    Scope,
    WorkloadSupportedTypes,
    EnvironmentAsset
} from '../../../../../../modules/api/apis/assetsApi';
import { kubeCli, bashCli } from '../../../../../../modules/cli';
import { DataSourcePage } from '../../../../../../modules/ui/runaiV2/pages/DataSourcePage';
import {
    TrainingPage,
    DISTRIBUTED_TOOLS_LABELS
} from '../../../../../../modules/ui/runaiV2/pages/TrainingPage';
import { HybridForm } from '../../../../../../modules/ui/runaiV2/pages/HybridSubmissionForm';
import { EFeatureToggleKeys } from '../../../../../../modules/helpers/utils/Unleash/featureToggle';

test.describe(`Run:ai template functionalities, tags: ${tags.tool.ui} ${tags.resource.workspaceTemplate} ${tags.teams.loong} ${tags.runOn.dgxc}`, () => {
    supportedFrom(SupportedVersions.V2_19)(function () {
        logger.info('Template functionality need nfs installed, skipping for now...');
        test.skip();
        let browserEngine: BrowserEngine;
        let auth: AuthPage;
        let workloads: WorkloadPage;
        let workspace: WorkspacePage;
        let training: TrainingPage;
        let templateName: string;
        let workspaceName: string;
        let trainingName: string;
        let projectName: string;
        let namespace: string;
        let project;
        let videoName: string;
        let pvcName: string;
        let templatePage: TemplatePage;
        let dataSourcePage: DataSourcePage;
        let copiedTemplateName: string;
        let deletedPvcName: string;
        let createdTemplates: any[] = [];
        const MOUNT_PATH = '/home/local/data';
        const claimSize = '2';
        const claimUnit = 'GB';
        const storageClass = 'nfs-csi';
        const tensorboardEnvName = 'tensorboard-tensorflow';

        test.beforeAll(async () => {
            await apiUtils.enableHybridSubmission();
            await apiUtils.setFlexTemplateFeature(false);
            browserEngine = new BrowserEngine();
        });

        test.afterAll(async () => {
            await apiUtils.setFlexTemplateFeature(true);
        });

        test.beforeEach(async () => {
            createdTemplates = [];
            await browserEngine.start();
            auth = new AuthPage(browserEngine.currentPage);
            workloads = new WorkloadPage(browserEngine.currentPage);
            workspace = new WorkspacePage(browserEngine.currentPage);
            training = new TrainingPage(browserEngine.currentPage);
            dataSourcePage = new DataSourcePage(browserEngine.currentPage);
            templatePage = new TemplatePage(browserEngine.currentPage);
            workspaceName = objectGenerator.generateNameByPrefix('e2e-workspace');
            trainingName = objectGenerator.generateNameByPrefix('e2e-training');
            projectName = objectGenerator.generateNameByPrefix('e2e-project');
            namespace = Utils.namespaceFromProjectName(projectName);
            project = await projectResourcesUtils.createProject(namespace, {
                name: projectName,
                deservedGpus: 1,
                maxAllowedGpus: 1
            });
            templateName = objectGenerator.generateNameByPrefix('e2e-template');
            pvcName = objectGenerator.generateNameByPrefix('e2e-pvc');
            await auth.signIn(
                flowRootState.users.l1_researcher.email,
                flowRootState.users.l1_researcher.password
            );
        });

        test.afterEach(async ({}, testInfo) => {
            for (const template of createdTemplates) {
                try {
                    await templatePage.navigateToTemplates();
                    await templatePage.deleteTemplate(template);
                } catch (e) {
                    logger.error('Error deleting template:', template, e);
                }
            }
            try {
                await dataSourcePage.goToPage();
                await dataSourcePage.deleteDataSource(pvcName);
            } catch (e) {
                logger.error('Error deleting PVC:', pvcName, e);
            }
            await browserEngine.stopAndArchive(testInfo, videoName || 'template-functionality-test');
            if (project && project.id) {
                try {
                    await operations.deleteProjectAndNamespace(project.id);
                    logger.info(`deleted project ${projectName} and namespace ${namespace}`);
                } catch {
                    logger.error('failed to delete project ' + projectName + ' and namespace ' + namespace);
                }
            }
        });

        test('T3738313: L1 researcher can leverage the Run:ai template functionalities to create interactive workloads', async () => {
            test.setTimeout(600 * 1000);
            videoName = 'L1-researcher-template-functionalities';

            await test.step('Create a PVC for template', async () => {
                await dataSourcePage.createPVC({
                    name: pvcName,
                    scope: Scope.Project,
                    scopeEntityName: projectName,
                    storageClass: storageClass,
                    accessMode: 'Read-write by many nodes',
                    claimSize: claimSize,
                    claimUnit: claimUnit,
                    volumeMode: 'Filesystem',
                    containerPath: MOUNT_PATH
                });
                await dataSourcesUtils.waitForDataSourceStatusInTable(
                    pvcName,
                    'No issues found',
                    dataSourcePage
                );
            });

            await test.step('Create a template', async () => {
                await templatePage.createWorkloadTemplateUI({
                    scopeType: Scope.Project,
                    scopeEntityName: projectName,
                    templateName,
                    environment: tensorboardEnvName,
                    computeResource: 'one-gpu',
                    dataSourceName: pvcName
                });

                await waitUtils.waitForCondition(
                    async () => {
                        const templateNames = await templatePage.getTemplateNames();
                        return templateNames.includes(templateName);
                    },
                    `Wait for template '${templateName}' to be created`,
                    true,
                    12,
                    5
                );
                logger.info(`Template ${templateName} created successfully in beforeAll`);
            });
            createdTemplates.push(templateName);

            await test.step('Navigate to Workloads and create workspace using template', async () => {
                await workspace.submitWorkspaceWorkloadUI(projectName, workspaceName, {
                    templateName: templateName
                });
                logger.info(`Created workspace: ${workspaceName} using template: ${templateName}`);
            });

            await test.step('Check the job is running as expected', async () => {
                await workloads.waitForWorkloadUIStatus(workspaceName, EWorkloadStatus.Running);
                await workloads.checkDisplayedWorkloadDetails({
                    jobName: workspaceName,
                    requestedGpu: '1',
                    type: WorkloadType.Workspace,
                    projectName,
                    status: EWorkloadStatus.Running,
                    createdBy: flowRootState.users.l1_researcher.email,
                    pods: '1/1'
                });
                logger.info(`Workspace ${workspaceName} is running successfully`);
            });

            await test.step('Login the pod and check mounted data sources exist', async () => {
                const pods = await kubeCli.listPodsOfJob(namespace, workspaceName);
                expect(pods.length).toBeGreaterThan(0);
                for (const pod of pods) {
                    const dfOutput = bashCli.exec(
                        `kubectl exec -it ${pod.metadata.name} -n ${namespace} -- df -h`
                    );
                    logger.info(`df -h output for pod ${pod.metadata.name}: ${dfOutput}`);
                    expect(dfOutput).toContain(MOUNT_PATH);
                    logger.info(
                        `Verified that data source is mounted at ${MOUNT_PATH} in pod ${pod.metadata.name}`
                    );
                }
            });
        });

        test('T3738314: L1 researcher can leverage the Run:ai template functionalities to create training workloads', async () => {
            test.setTimeout(600 * 1000);
            videoName = `L1-researcher-template-functionalities-${trainingName}`;
            await test.step('Create a PVC for template', async () => {
                await dataSourcePage.createPVC({
                    name: pvcName,
                    scope: Scope.Project,
                    scopeEntityName: projectName,
                    storageClass: storageClass,
                    accessMode: 'Read-write by many nodes',
                    claimSize: claimSize,
                    claimUnit: claimUnit,
                    volumeMode: 'Filesystem',
                    containerPath: MOUNT_PATH
                });
                await dataSourcesUtils.waitForDataSourceStatusInTable(
                    pvcName,
                    'No issues found',
                    dataSourcePage
                );
            });

            await test.step('Create a template', async () => {
                await templatePage.createWorkloadTemplateUI({
                    scopeType: Scope.Project,
                    scopeEntityName: projectName,
                    templateName,
                    environment: tensorboardEnvName,
                    computeResource: 'one-gpu',
                    dataSourceName: pvcName
                });

                await waitUtils.waitForCondition(
                    async () => {
                        const templateNames = await templatePage.getTemplateNames();
                        return templateNames.includes(templateName);
                    },
                    `Wait for template '${templateName}' to be created`,
                    true,
                    12,
                    5
                );
                logger.info(`Template ${templateName} created successfully in beforeAll`);
            });
            createdTemplates.push(templateName);

            await test.step('Navigate to Workloads and create training using template', async () => {
                await training.submitTrainingWorkloadUI(projectName, trainingName, {
                    templateName: templateName
                });
                logger.info(`Created training: ${trainingName} using template: ${templateName}`);
            });

            await test.step('Check the job is running as expected', async () => {
                await workloads.waitForWorkloadUIStatus(trainingName, EWorkloadStatus.Running);
                await workloads.checkDisplayedWorkloadDetails({
                    jobName: trainingName,
                    requestedGpu: '1',
                    type: WorkloadType.Training,
                    projectName,
                    status: EWorkloadStatus.Running,
                    createdBy: flowRootState.users.l1_researcher.email,
                    pods: '1/1'
                });
                logger.info(`Training ${trainingName} is running successfully`);
            });

            await test.step('Login the pod and check mounted data sources exist', async () => {
                const pods = await kubeCli.listPodsOfJob(namespace, trainingName);
                expect(pods.length).toBeGreaterThan(0);
                for (const pod of pods) {
                    const dfOutput = bashCli.exec(
                        `kubectl exec -it ${pod.metadata.name} -n ${namespace} -- df -h`
                    );
                    logger.info(`df -h output for pod ${pod.metadata.name}: ${dfOutput}`);
                    expect(dfOutput).toContain(MOUNT_PATH);
                    logger.info(
                        `Verified that data source is mounted at ${MOUNT_PATH} in pod ${pod.metadata.name}`
                    );
                }
            });
        });

        test('T4066888: Verify L1 researcher can copy the Run:ai template', async () => {
            test.setTimeout(300 * 1000);
            videoName = 'template-copy-functionality';
            deletedPvcName = objectGenerator.generateNameByPrefix('e2e-deleted-pvc');
            copiedTemplateName = objectGenerator.generateNameByPrefix('copied-template');

            await test.step('Create a PVC for template', async () => {
                await dataSourcePage.createPVC({
                    name: pvcName,
                    scope: Scope.Project,
                    scopeEntityName: projectName,
                    storageClass: storageClass,
                    accessMode: 'Read-write by many nodes',
                    claimSize: claimSize,
                    claimUnit: claimUnit,
                    volumeMode: 'Filesystem',
                    containerPath: MOUNT_PATH
                });
                await dataSourcesUtils.waitForDataSourceStatusInTable(
                    pvcName,
                    'No issues found',
                    dataSourcePage
                );
                logger.info(`Created PVC ${pvcName} successfully`);
            });

            await test.step('Create a template with environment variable', async () => {
                await templatePage.createWorkloadTemplateUI({
                    scopeType: Scope.Project,
                    scopeEntityName: projectName,
                    templateName,
                    environment: tensorboardEnvName,
                    computeResource: 'one-gpu',
                    dataSourceName: pvcName
                });

                await waitUtils.waitForCondition(
                    async () => {
                        const templateNames = await templatePage.getTemplateNames();
                        return templateNames.includes(templateName);
                    },
                    `Wait for template '${templateName}' to be created`,
                    true,
                    12,
                    5
                );
                logger.info(`Template ${templateName} created successfully`);
            });
            createdTemplates.push(templateName);

            await test.step('Create another PVC', async () => {
                await dataSourcePage.createPVC({
                    name: deletedPvcName,
                    scope: Scope.Project,
                    scopeEntityName: projectName,
                    storageClass: storageClass,
                    accessMode: 'Read-write by many nodes',
                    claimSize: claimSize,
                    claimUnit: claimUnit,
                    volumeMode: 'Filesystem',
                    containerPath: MOUNT_PATH
                });
                await dataSourcesUtils.waitForDataSourceStatusInTable(
                    deletedPvcName,
                    'No issues found',
                    dataSourcePage
                );
                logger.info(`Created PVC ${deletedPvcName} successfully`);
            });

            await test.step('Delete the second PVC', async () => {
                await dataSourcePage.goToPage();
                await dataSourcePage.deleteDataSource(deletedPvcName);
                await waitUtils.waitForElementToNotExist(dataSourcePage, `td:has-text('${deletedPvcName}')`);
                logger.info(`Deleted PVC ${deletedPvcName}`);
            });

            await test.step('Select the created template, click "MAKE A COPY" to create another template', async () => {
                logger.info(`Verify that deleted PVC ${deletedPvcName} is not available in copied template`);
                await templatePage.navigateToTemplates();
                await templatePage.toggleTemplate(templateName);
                await templatePage.clickOnCopyTemplate();
                await templatePage.fillTemplateName(copiedTemplateName);

                // Check the deleted PVC is not available when copy template
                await templatePage.page.locator(templatePage.dataSourceExpansion).click();
                expect(
                    await templatePage.page.locator(templatePage.getCardSelector(deletedPvcName)).count()
                ).toEqual(0);

                await templatePage.clickOnCreateTemplateBtn();
                await waitUtils.waitForCondition(
                    async () => {
                        const templateNames = await templatePage.getTemplateNames();
                        return templateNames.includes(copiedTemplateName);
                    },
                    `Wait for copied template '${copiedTemplateName}' to be created`,
                    true,
                    12,
                    5
                );
                logger.info(`Copied template ${copiedTemplateName} created successfully`);
            });
            createdTemplates.push(copiedTemplateName);

            await test.step('Check that new template can be used for workloads', async () => {
                // Create a workspace using the copied template
                await workspace.submitWorkspaceWorkloadUI(projectName, workspaceName, {
                    templateName: copiedTemplateName
                });
                logger.info(
                    `Created workspace: ${workspaceName} using copied template: ${copiedTemplateName}`
                );

                // Check the job is running as expected
                await workloads.waitForWorkloadUIStatus(workspaceName, EWorkloadStatus.Running);
                await workloads.checkDisplayedWorkloadDetails({
                    jobName: workspaceName,
                    requestedGpu: '1',
                    type: WorkloadType.Workspace,
                    projectName,
                    status: EWorkloadStatus.Running,
                    createdBy: flowRootState.users.l1_researcher.email,
                    pods: '1/1'
                });
                logger.info(`Workspace ${workspaceName} is running successfully using copied template`);
            });

            await test.step('Login the pod of the workspace and check data sources can be used', async () => {
                const pods = await kubeCli.listPodsOfJob(namespace, workspaceName);
                expect(pods.length).toBeGreaterThan(0);
                for (const pod of pods) {
                    const dfOutput = bashCli.exec(
                        `kubectl exec -it ${pod.metadata.name} -n ${namespace} -- df -h`
                    );
                    logger.info(`df -h output for pod ${pod.metadata.name}: ${dfOutput}`);
                    expect(dfOutput).toContain(MOUNT_PATH);
                    logger.info(
                        `Verified that data source is mounted at ${MOUNT_PATH} in pod ${pod.metadata.name}`
                    );
                }
            });
        });
    });
});

test.describe(`Distributed training workloads with templates, tags: ${tags.tool.ui} ${tags.resource.workloads} ${tags.teams.loong} ${tags.runOn.dgxc}`, () => {
    supportedFrom(SupportedVersions.V2_19)(function () {
        let browserEngine: BrowserEngine;
        let auth: AuthPage;
        let workloads: WorkloadPage;
        let training: TrainingPage;
        let templatePage: TemplatePage;
        let dataSourcePage: DataSourcePage;
        let projectName: string;
        let namespace: string;
        let project;
        let videoName: string;
        let environmentAsset: EnvironmentAsset;
        let templateName: string;
        let trainingName: string;
        let pvcName: string;
        const MOUNT_PATH = '/home/local/data';
        const claimSize = '2';
        const claimUnit = 'GB';
        const storageClass = 'nfs-csi';

        test.beforeAll(async () => {
            logger.info('Template functionality need nfs installed, skipping for now...');
            test.skip();
            logger.info('Enabling hybrid submission feature');
            await unleashClient.enableFeature(EFeatureToggleKeys.SHOW_HYBRID_SUBMISSION);
            await apiUtils.enableHybridSubmission();

            projectName = objectGenerator.generateNameByPrefix('e2e-project');
            namespace = Utils.namespaceFromProjectName(projectName);
            project = await projectResourcesUtils.createProject(namespace, {
                name: projectName,
                deservedGpus: 3,
                maxAllowedGpus: 3
            });

            templateName = objectGenerator.generateNameByPrefix('e2e-template');
            pvcName = objectGenerator.generateNameByPrefix('e2e-pvc');
        });

        test.afterAll(async () => {
            logger.info('Disabling hybrid submission feature');
            await apiUtils.disableHybridSubmission();
            await unleashClient.disableFeature(EFeatureToggleKeys.SHOW_HYBRID_SUBMISSION);

            if (project && project.id) {
                try {
                    await operations.deleteProjectAndNamespace(project.id);
                    logger.info(`deleted project ${projectName} and namespace ${namespace}`);
                } catch {
                    logger.error('failed to delete project ' + projectName + ' and namespace ' + namespace);
                }
            }
        });

        test.beforeEach(async () => {
            browserEngine = new BrowserEngine();
            await browserEngine.start();
            auth = new AuthPage(browserEngine.currentPage);
            workloads = new WorkloadPage(browserEngine.currentPage);
            training = new TrainingPage(browserEngine.currentPage);
            dataSourcePage = new DataSourcePage(browserEngine.currentPage);
            templatePage = new TemplatePage(browserEngine.currentPage);
            trainingName = objectGenerator.generateNameByPrefix('e2e-training');
            await auth.signIn(
                flowRootState.users.l1_researcher.email,
                flowRootState.users.l1_researcher.password
            );
        });

        test.afterEach(async ({}, testInfo) => {
            try {
                await templatePage.navigateToTemplates();
                await templatePage.deleteTemplate(templateName);
            } catch (e) {
                logger.error('Error deleting template:', templateName, e);
            }

            try {
                await environmentUtils.delete(environmentAsset.meta.id);
            } catch (e) {
                logger.error('Error deleting environment:', environmentAsset?.meta?.name, e);
            }

            try {
                await dataSourcePage.goToPage();
                await dataSourcePage.deleteDataSource(pvcName);
            } catch (e) {
                logger.error('Error deleting PVC:', pvcName, e);
            }
            await browserEngine.stopAndArchive(testInfo, 'template-functionality-distributed-training');
        });

        test('T3751344: L1 researcher can leverage the Run:ai template functionalities to distribute training workloads', async () => {
            test.setTimeout(300 * 1000);
            videoName = 'L1-researcher-template-functionalities-distributed-training';

            await test.step('Create environment that supports distributed job', async () => {
                environmentAsset = await environmentUtils.createEnvironmentWithArbitraryFields(
                    project.id,
                    '/bin/bash -c',
                    WorkloadSupportedTypes.DistFrameworkEnum.Tf,
                    config.quickStartImage.distributedTf,
                    '"sleep 300"'
                );
                logger.info(`Created environment ${environmentAsset.meta.name} for distributed training`);
            });

            await test.step('Create a PVC for template', async () => {
                await dataSourcePage.createPVC({
                    name: pvcName,
                    scope: Scope.Project,
                    scopeEntityName: projectName,
                    storageClass: storageClass,
                    accessMode: 'Read-write by many nodes',
                    claimSize: claimSize,
                    claimUnit: claimUnit,
                    volumeMode: 'Filesystem',
                    containerPath: MOUNT_PATH
                });
                await dataSourcesUtils.waitForDataSourceStatusInTable(
                    pvcName,
                    'No issues found',
                    dataSourcePage
                );
                logger.info(`Created PVC ${pvcName} successfully`);
            });

            await test.step('Create a template', async () => {
                await templatePage.createWorkloadTemplateUI({
                    scopeType: Scope.Project,
                    scopeEntityName: projectName,
                    templateName,
                    environment: environmentAsset.meta.name,
                    computeResource: 'one-gpu',
                    dataSourceName: pvcName
                });

                await waitUtils.waitForCondition(
                    async () => {
                        const templateNames = await templatePage.getTemplateNames();
                        return templateNames.includes(templateName);
                    },
                    `Wait for template '${templateName}' to be created`,
                    true,
                    12,
                    5
                );
                logger.info(`Template ${templateName} created successfully`);
            });

            await test.step('Navigate to Workloads and create training using template', async () => {
                await test.step('Create distributed TensorFlow training job', async () => {
                    await training.goToPage();
                    await training.fillFirstNewTrainingPage(
                        projectName,
                        trainingName,
                        true,
                        DISTRIBUTED_TOOLS_LABELS.TENSOR_FLOW,
                        true,
                        true,
                        templateName
                    );
                    const hybridForm = new HybridForm(browserEngine.currentPage);
                    await hybridForm.fillComputeSection(undefined, 2);
                    await training.clickOnSubmitBtn(); // worker step
                    await hybridForm.waitForDistributedMasterSubmitButton();
                    await hybridForm.clickOnSubmitMasterForm();
                    await training.waitForSubmitButton();
                    await training.clickOnSubmitBtn(); // setup step
                });
            });

            await test.step('Check the job is running as expected', async () => {
                await workloads.waitForWorkloadUIStatus(trainingName, EWorkloadStatus.Running);
                await workloads.checkDisplayedWorkloadDetails({
                    jobName: trainingName,
                    requestedGpu: '3',
                    type: WorkloadType.Training,
                    projectName,
                    status: EWorkloadStatus.Running,
                    createdBy: flowRootState.users.l1_researcher.email,
                    pods: '3/3'
                });
                logger.info(`Training ${trainingName} is running successfully`);
            });

            await test.step('Login the launcher and worker pod of the distributed job and check mounted data sources exist', async () => {
                const pods = await kubeCli.listPodsOfJob(namespace, trainingName);
                expect(pods.length).toBeGreaterThan(0);
                for (const pod of pods) {
                    const dfOutput = bashCli.exec(
                        `kubectl exec -it ${pod.metadata.name} -n ${namespace} -- df -h`
                    );
                    logger.info(`df -h output for pod ${pod.metadata.name}: ${dfOutput}`);
                    expect(dfOutput).toContain(MOUNT_PATH);
                    logger.info(
                        `Verified that data source is mounted at ${MOUNT_PATH} in pod ${pod.metadata.name}`
                    );
                }
            });
        });
    });
});
