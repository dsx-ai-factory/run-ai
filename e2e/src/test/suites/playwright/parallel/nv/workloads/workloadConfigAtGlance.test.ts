/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    Utils,
    flowRootState,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    config,
    workloadsCrudUtils,
    workloadsUtilsV2,
    waitUtils
} from '../../../../../../modules/helpers';
import { WorkloadType as ApiWorkloadType } from '../../../../../../modules/helpers/utils/WorkloadsCrudUtils';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { EWorkloadStatus, WorkloadPage } from '../../../../../../modules/ui/runaiV2/pages/WorkloadPage';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { expect, test } from '@playwright/test';
import { Phase } from '../../../../../../modules/api/apis/assetsApi';
import { DistributedFramework } from '../../../../../../modules/api/apis/workloadsApi/generated/model/distributedFramework';
import { runaiCliV2 } from '../../../../../../modules/cli';
import {
    cliV2WorkloadType,
    SubmitFlagsMerged
} from '../../../../../../modules/cli/tools/RunaiCliV2.constants';
import { WorkloadDetailsDrawer } from '../../../../../../modules/ui/runaiV2/pages/WorkloadDetailsDrawer';
import { getBaseWorkloadSpec } from '../../viper/inference/distributed/dinfUtils';
import { createDistributedInference } from '../../viper/inference/inferenceUtils';

test.describe(`Workload Config at Glance, tags: ${tags.resource.workloads} ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_24)(function () {
        let browserEngine: BrowserEngine;
        let auth: AuthPage;
        let workloads: WorkloadPage;
        let videoName: string;
        let projectName: string;
        let namespace: string;
        let project: Project;
        let trainingName: string;
        let inferenceName: string;
        let detailsDrawer: WorkloadDetailsDrawer;
        let createdWorkloadIds: Array<{ id: string; type: ApiWorkloadType }>;

        test.beforeAll(async function () {
            browserEngine = new BrowserEngine();
            projectName = objectGenerator.generateShortNameByPrefix('project-e2e-for-training');
            namespace = Utils.namespaceFromProjectName(projectName);
            project = await projectResourcesUtils.createProject(namespace, {
                name: projectName,
                deservedGpus: 3,
                maxAllowedGpus: 3
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

        test.beforeEach(async function () {
            createdWorkloadIds = [];
            await browserEngine.start();
            auth = new AuthPage(browserEngine.currentPage);
            workloads = new WorkloadPage(browserEngine.currentPage);
            trainingName = objectGenerator.generateShortNameByPrefix('glance-training-');
            detailsDrawer = new WorkloadDetailsDrawer(browserEngine.currentPage);
            await auth.signIn(
                flowRootState.users.system_administrator.email,
                flowRootState.users.system_administrator.password
            );
            workloadsCrudUtils.accessToken = flowRootState.users.system_administrator.accessToken;
        });

        test.afterEach('cleanup workloads and stop browser', async function () {
            for (const workload of createdWorkloadIds) {
                try {
                    await workloadsCrudUtils.deleteWorkload(workload.type, workload.id);
                    logger.info(`Deleted workload: ${workload.id} of type ${workload.type}`);
                } catch (error) {
                    logger.warn(`Failed to delete workload ${workload.id}: ${error.message}`);
                }
            }
            await browserEngine.stopAndArchive(this, 'workload-config-at-glance-' + videoName);
        });

        test('T5541617: Verify Details Tab is Visible and Content Loads successfully', async () => {
            videoName = 'details-tab-is-visible';
            await test.step('Create a training workload with API', async () => {
                const { body: training } = await workloadsCrudUtils.createTraining(
                    trainingName,
                    true,
                    flowRootState.cluster.uuid,
                    project.id.toString(),
                    {
                        image: config.quickStartImage.ubuntu,
                        command: 'sleep',
                        args: 'infinity',
                        compute: { gpuDevicesRequest: 1 }
                    }
                );
                trainingName = training.name;
                createdWorkloadIds.push({ id: training.workloadId, type: ApiWorkloadType.TRAINING });
                logger.info(`Created training workload: ${trainingName} with ID: ${training.workloadId}`);
            });

            await test.step('Wait for workload to be running', async () => {
                await workloads.goToPage();
                await workloads.waitForWorkloadUIStatus(trainingName, EWorkloadStatus.Running);
                logger.info(`Workload ${trainingName} is running`);
            });

            await test.step('Navigate to workload detail page and verify Details tab exists', async () => {
                await detailsDrawer.clickShowDetailsBtn(trainingName);
                await workloads.validateDetailsTabContentExists();
                await detailsDrawer.clickOnDetailsTab();
            });

            await test.step('Verify content is displayed in the Details tab', async () => {
                // Verify General Information section
                const generalInfo = await detailsDrawer.getGeneralInformation();
                expect(generalInfo.project).toBe(projectName);
                expect(generalInfo.category).toBeTruthy();
                expect(generalInfo.type).toBeTruthy();
                logger.info(
                    `General Info - Project: ${generalInfo.project}, Category: ${generalInfo.category}, Type: ${generalInfo.type}`
                );

                // Verify Environment section
                const environmentInfo = await detailsDrawer.getEnvironmentInfo();
                expect(environmentInfo.imageUrl).toContain(config.quickStartImage.ubuntu);
                expect(environmentInfo.command).toContain('sleep');
                expect(environmentInfo.arguments).toContain('infinity');
                logger.info(
                    `Environment Info - Image: ${environmentInfo.imageUrl}, Command: ${environmentInfo.command}, Arguments: ${environmentInfo.arguments}`
                );

                // Verify Compute section
                const computeInfo = await detailsDrawer.getComputeInfo({
                    gpuResources: true,
                    cpuResources: false,
                    pods: true
                });
                expect(computeInfo.nodePools).toBeTruthy();
                expect(computeInfo.gpuComputeRequest).toBeTruthy();
                expect(computeInfo.gpuComputeRequest).toContain('1');
                expect(computeInfo.numberOfPods).toBeUndefined();
                logger.info(
                    `Compute Info - Node Pools: ${computeInfo.nodePools}, GPU Request: ${computeInfo.gpuComputeRequest}, Pods: ${computeInfo.numberOfPods}`
                );
            });

            await test.step('Verify Running / requested pods on Structure tab', async () => {
                await detailsDrawer.clickOnStructureTab();
                const runningRequestedPods = await detailsDrawer.getStructureRunningRequestedPods();
                expect(runningRequestedPods).toBe('1/1');
                logger.info(`Structure tab Running / requested pods: ${runningRequestedPods}`);
            });
        });

        test('T5541618: Verify General Information Section Display', async () => {
            videoName = 'general-information-verification';
            await test.step('Create a training workload', async () => {
                const { body: training } = await workloadsCrudUtils.createTraining(
                    trainingName,
                    true,
                    flowRootState.cluster.uuid,
                    project.id.toString(),
                    {
                        image: config.quickStartImage.ubuntu,
                        command: 'sleep infinity',
                        compute: { gpuDevicesRequest: 1 }
                    }
                );
                trainingName = training.name;
                createdWorkloadIds.push({ id: training.workloadId, type: ApiWorkloadType.TRAINING });
                logger.info(`Created training workload: ${trainingName} with ID: ${training.workloadId}`);
            });

            await test.step('Wait for workload to be running', async () => {
                await workloads.goToPage();
                await workloads.waitForWorkloadUIStatus(trainingName, EWorkloadStatus.Running);
                logger.info(`Workload ${trainingName} is running`);
            });

            await test.step('Navigate to workload Details tab and locate General Information section', async () => {
                await detailsDrawer.clickShowDetailsBtn(trainingName);
                await detailsDrawer.clickOnDetailsTab();
            });

            await test.step('Verify general information section, all fields are present with valid values', async () => {
                const generalInfo = await detailsDrawer.getGeneralInformation();
                expect(generalInfo.cluster).toBeTruthy();
                expect(generalInfo.cluster).not.toBe('');
                logger.info(`Cluster: ${generalInfo.cluster}`);

                expect(generalInfo.department).toBeTruthy();
                expect(generalInfo.department).not.toBe('');
                logger.info(`Department: ${generalInfo.department}`);

                expect(generalInfo.project).toBeTruthy();
                expect(generalInfo.project).not.toBe('');
                logger.info(`Project: ${generalInfo.project}`);

                expect(generalInfo.category).toBeTruthy();
                expect(generalInfo.category).not.toBe('');
                logger.info(`Category: ${generalInfo.category}`);

                expect(generalInfo.type).toBeTruthy();
                expect(generalInfo.type).not.toBe('');
                logger.info(`Type: ${generalInfo.type}`);

                expect(generalInfo.priority).toBeTruthy();
                expect(generalInfo.priority).not.toBe('');
                logger.info(`Priority: ${generalInfo.priority}`);

                expect(generalInfo.preemptible).toBeDefined();
                expect(generalInfo.preemptible).not.toBe('');
                logger.info(`Preemptible: ${generalInfo.preemptible}`);

                expect(generalInfo.creationTime).toBeTruthy();
                expect(generalInfo.creationTime).not.toBe('');
                expect(generalInfo.creationTime).toMatch(/^\d{2}\/\d{2}\/\d{4}$/);
                logger.info(`Creation Time: ${generalInfo.creationTime}`);

                expect(generalInfo.createdBy).toBeTruthy();
                expect(generalInfo.createdBy).not.toBe('');
                logger.info(`Created By: ${generalInfo.createdBy}`);
                logger.info('General information section is verified successfully');
            });
        });

        test('T5541619: Verify General Information Values are Accurate', async () => {
            videoName = 'general-information-values-accurate';
            await test.step('Create a new workload with specific known values', async () => {
                const { body: training } = await workloadsCrudUtils.createTraining(
                    trainingName,
                    true,
                    flowRootState.cluster.uuid,
                    project.id.toString(),
                    {
                        image: config.quickStartImage.ubuntu,
                        command: 'sleep infinity',
                        compute: { gpuDevicesRequest: 1 }
                    }
                );
                trainingName = training.name;
                createdWorkloadIds.push({ id: training.workloadId, type: ApiWorkloadType.TRAINING });
                logger.info(`Created training workload: ${trainingName} with ID: ${training.workloadId}`);
            });

            await test.step('Wait for workload to start', async () => {
                await workloads.goToPage();
                await workloads.waitForWorkloadUIStatus(trainingName, EWorkloadStatus.Running);
                logger.info(`Workload ${trainingName} is running`);
            });

            await test.step('Navigate to workload Details tab', async () => {
                await detailsDrawer.clickShowDetailsBtn(trainingName);
                await detailsDrawer.clickOnDetailsTab();
            });

            await test.step('Verify General Information section values are accurate', async () => {
                const generalInfo = await detailsDrawer.getGeneralInformation();
                expect(generalInfo.cluster).toBe(flowRootState.cluster.name);
                logger.info(`Cluster verified: ${generalInfo.cluster} matches ${flowRootState.cluster.name}`);

                expect(generalInfo.project).toBe(projectName);
                logger.info(`Project verified: ${generalInfo.project} matches ${projectName}`);

                expect(generalInfo.createdBy).toBe(flowRootState.users.system_administrator.email);
                logger.info(
                    `Created By verified: ${generalInfo.createdBy} matches ${flowRootState.users.system_administrator.email}`
                );

                const today = new Date();
                const day = String(today.getDate()).padStart(2, '0');
                const month = String(today.getMonth() + 1).padStart(2, '0');
                const year = today.getFullYear();
                const todayFormatted = `${day}/${month}/${year}`;
                expect(generalInfo.creationTime).toBe(todayFormatted);
                logger.info(
                    `Creation Time verified: ${generalInfo.creationTime} matches today's date ${todayFormatted}`
                );
                logger.info('All General Information values are verified as accurate');
            });
        });

        test('T5541620: Verify Category and Type for Different Workload Types', async () => {
            videoName = 'category-type-verification';
            let standardTrainingName: string;
            let standardTrainingId: string;
            let distributedTrainingName: string;
            let distributedTrainingId: string;
            let inferenceName: string;
            let inferenceId: string;
            await test.step('Create a standard Training workload', async () => {
                standardTrainingName = objectGenerator.generateShortNameByPrefix('std-training');
                const { body: training } = await workloadsCrudUtils.createTraining(
                    standardTrainingName,
                    true,
                    flowRootState.cluster.uuid,
                    project.id.toString(),
                    {
                        image: config.quickStartImage.ubuntu,
                        command: 'sleep infinity',
                        compute: { cpuCoreRequest: 1, cpuMemoryRequest: '1G' }
                    }
                );
                standardTrainingName = training.name;
                standardTrainingId = training.workloadId;
                createdWorkloadIds.push({ id: standardTrainingId, type: ApiWorkloadType.TRAINING });
                logger.info(`Created standard training workload: ${standardTrainingName}`);
            });

            await test.step('Wait for standard training workload to be running', async () => {
                await workloads.goToPage();
                await workloads.waitForWorkloadUIStatus(standardTrainingName, EWorkloadStatus.Running);
                logger.info(`Standard training workload ${standardTrainingName} is running`);
            });

            await test.step('Navigate to standard training Details tab and verify Category and Type', async () => {
                await detailsDrawer.clickShowDetailsBtn(standardTrainingName);
                await detailsDrawer.clickOnDetailsTab();
                const generalInfo = await detailsDrawer.getGeneralInformation();

                expect(generalInfo.category).toBe('Train');
                expect(generalInfo.type).toContain('Training');
                logger.info(
                    `Standard Training - Category: ${generalInfo.category}, Type: ${generalInfo.type} verified`
                );
            });

            await test.step('Delete the standard training workload', async () => {
                await workloadsCrudUtils.deleteWorkload(ApiWorkloadType.TRAINING, standardTrainingId);
                logger.info(
                    `Deleted standard training workload: ${standardTrainingName} (ID: ${standardTrainingId})`
                );
                createdWorkloadIds.shift();
            });

            await test.step('Create a distributed Training workload', async () => {
                distributedTrainingName = objectGenerator.generateShortNameByPrefix('dis-training');
                const { body: training } = await workloadsCrudUtils.createDistributed(
                    distributedTrainingName,
                    true,
                    flowRootState.cluster.uuid,
                    project.id.toString(),
                    {
                        image: config.quickStartImage.ubuntu,
                        command: 'sleep infinity',
                        compute: { gpuDevicesRequest: 1 },
                        numWorkers: 2,
                        distributedFramework: DistributedFramework.PyTorch
                    }
                );
                distributedTrainingName = training.name;
                distributedTrainingId = training.workloadId;
                createdWorkloadIds.push({ id: distributedTrainingId, type: ApiWorkloadType.DISTRIBUTED });
                logger.info(`Created distributed training workload: ${distributedTrainingName}`);
            });

            await test.step('Wait for distributed training workload to be running', async () => {
                await detailsDrawer.hideDetailsDrawer();
                await workloads.waitForWorkloadUIStatus(distributedTrainingName, EWorkloadStatus.Running);
                logger.info(`Distributed training workload ${distributedTrainingName} is running`);
            });

            await test.step('Navigate to distributed training Details tab and verify Category and Type', async () => {
                await detailsDrawer.clickShowDetailsBtn(distributedTrainingName);
                await detailsDrawer.clickOnDetailsTab();
                const generalInfo = await detailsDrawer.getGeneralInformation();

                expect(generalInfo.category).toBe('Train');
                expect(generalInfo.type).toContain('Distributed');
                logger.info(
                    `Distributed Training - Category: ${generalInfo.category}, Type: ${generalInfo.type} verified`
                );
            });

            await test.step('Delete the distributed training workload', async () => {
                await workloadsCrudUtils.deleteWorkload(ApiWorkloadType.DISTRIBUTED, distributedTrainingId);
                logger.info(
                    `Deleted distributed training workload: ${distributedTrainingName} (ID: ${distributedTrainingId})`
                );
                createdWorkloadIds.shift();
            });

            await test.step('Create a Distributed Inference workload (workers=0)', async () => {
                inferenceName = objectGenerator.generateShortNameByPrefix('inference');

                // Create distributed inference with workers=0 (single leader only) - ensures true Running state
                logger.info('Creating distributed inference with single leader (workers=0)');
                const inferSpec = getBaseWorkloadSpec();
                inferSpec.workers = 0; // Only leader, no workers
                inferSpec.replicas = 1;
                delete inferSpec.worker; // Remove worker spec when workers=0 to avoid validation error

                [inferenceId, inferenceName] = await createDistributedInference(
                    inferenceName,
                    inferSpec,
                    project
                );

                createdWorkloadIds.push({ id: inferenceId, type: ApiWorkloadType.DISTRIBUTED_INFERENCE });
                logger.info(`Created distributed inference workload: ${inferenceName} (ID: ${inferenceId})`);
            });

            await test.step('Wait for inference workload to be Running', async () => {
                await waitUtils.waitForBackendWorkloadByIdAndPhase(inferenceId, Phase.Running, 120);
                logger.info(`Inference workload ${inferenceName} is in Running phase`);
            });

            await test.step('Navigate to inference Details tab and verify Category and Type', async () => {
                await detailsDrawer.hideDetailsDrawer();
                await workloads.clearGridSearchInput();
                await workloads.searchByName(inferenceName);
                await detailsDrawer.clickShowDetailsBtn(inferenceName);
                await detailsDrawer.clickOnDetailsTab();
                const generalInfo = await detailsDrawer.getGeneralInformation();

                expect(generalInfo.category).toBe('Deploy');
                expect(generalInfo.type).toContain('Inference');
                logger.info(
                    `Inference - Category: ${generalInfo.category}, Type: ${generalInfo.type} verified`
                );

                logger.info('All workload categories and types are verified successfully');
            });
        });

        test('T5541621: Verify Environment Section Display', async () => {
            videoName = 'environment-section-verification';
            await test.step('Create a training workload with command and arguments', async () => {
                const { body: training } = await workloadsCrudUtils.createTraining(
                    trainingName,
                    true,
                    flowRootState.cluster.uuid,
                    project.id.toString(),
                    {
                        image: config.quickStartImage.ubuntu,
                        command: 'bash -c',
                        args: 'echo "Hello World"',
                        compute: { gpuDevicesRequest: 1 }
                    }
                );
                trainingName = training.name;
                createdWorkloadIds.push({ id: training.workloadId, type: ApiWorkloadType.TRAINING });
                logger.info(`Created training workload: ${trainingName} with ID: ${training.workloadId}`);
            });

            await test.step('Wait for workload to be running', async () => {
                await workloads.goToPage();
                await workloads.waitForWorkloadUIStatus(trainingName, EWorkloadStatus.Completed);
                logger.info(`Workload ${trainingName} is Completed`);
            });

            await test.step('Navigate to workload Details tab', async () => {
                await detailsDrawer.clickShowDetailsBtn(trainingName);
                await detailsDrawer.clickOnDetailsTab();
            });

            await test.step('Verify Environment section is displayed', async () => {
                const environmentSection = detailsDrawer.getEnvironmentSection();
                await expect(environmentSection).toBeVisible();
                logger.info('Environment section is visible');
            });

            await test.step('Verify Image URL field is present and shows container image', async () => {
                const imageUrl = await detailsDrawer.getImageUrl();
                expect(imageUrl).toBeTruthy();
                expect(imageUrl).toContain(config.quickStartImage.ubuntu);
                logger.info(`Image URL field verified: ${imageUrl}`);
            });

            await test.step('Verify Command field is present and shows executable', async () => {
                const command = await detailsDrawer.getCommand();
                expect(command).toBeTruthy();
                expect(command).toContain('bash -c');
                logger.info(`Command field verified: ${command}`);
            });

            await test.step('Verify Arguments field is present and shows parameters', async () => {
                const args = await detailsDrawer.getArguments();
                expect(args).toBeTruthy();
                expect(args).toContain('Hello World');
                logger.info(`Arguments field verified: ${args}`);
            });
        });

        test('T5541622: Verify Environment Section for Workload Without Custom Command', async () => {
            videoName = 'environment-without-custom-command';
            await test.step('Create a workload without custom command (uses container default)', async () => {
                const { body: training } = await workloadsCrudUtils.createTraining(
                    trainingName,
                    true,
                    flowRootState.cluster.uuid,
                    project.id.toString(),
                    {
                        image: config.quickStartImage.ubuntu,
                        compute: { gpuDevicesRequest: 1 }
                    }
                );
                trainingName = training.name;
                createdWorkloadIds.push({ id: training.workloadId, type: ApiWorkloadType.TRAINING });
                logger.info(
                    `Created training workload without custom command: ${trainingName} with ID: ${training.workloadId}`
                );
            });

            await test.step('Wait for workload to be Completed', async () => {
                await workloads.goToPage();
                await workloads.waitForWorkloadUIStatus(trainingName, EWorkloadStatus.Completed);
                logger.info(`Workload ${trainingName} is Completed`);
            });

            await test.step('Navigate to Details tab → Environment section', async () => {
                await detailsDrawer.clickShowDetailsBtn(trainingName);
                await detailsDrawer.clickOnDetailsTab();
            });

            await test.step('Verify Image URL is displayed', async () => {
                const imageUrl = await detailsDrawer.getImageUrl();
                expect(imageUrl).toBeTruthy();
                expect(imageUrl).toContain(config.quickStartImage.ubuntu);
                logger.info(`Image URL is displayed: ${imageUrl}`);
            });

            await test.step('Check Command field - should not exist when no custom command specified', async () => {
                const command = await detailsDrawer.getCommand();
                expect(command).toBeUndefined();
                logger.info('Command field is not displayed (no custom command specified)');
            });

            await test.step('Check Arguments field - should not exist when no custom arguments specified', async () => {
                const args = await detailsDrawer.getArguments();
                expect(args).toBeUndefined();
                logger.info('Arguments field is not displayed (no custom arguments specified)');
            });

            await test.step('Verify Environment section only contains Image URL', async () => {
                const environmentSection = detailsDrawer.getEnvironmentSection();
                const itemCount = await environmentSection.locator('.runai-details-panel-row').count();
                expect(itemCount).toBe(1);
                logger.info(
                    `Environment section contains ${itemCount} item (Image URL only) - Command and Arguments are not displayed`
                );
            });
        });

        test('T5541623: Verify Network Section Display', async () => {
            videoName = 'network-section-display';
            let inferenceId: string;

            await test.step('Create a Distributed Inference workload with external URL (Jupyter notebook)', async () => {
                inferenceName = objectGenerator.generateShortNameByPrefix('inference-jupyter');

                logger.info('Creating distributed inference with external URL (workers=0)');
                const inferSpec = getBaseWorkloadSpec();
                inferSpec.workers = 0;
                inferSpec.replicas = 1;
                inferSpec.servingPort = {
                    port: 8888
                };
                inferSpec.leader = {
                    image: 'jupyter/base-notebook:latest',
                    compute: { gpuDevicesRequest: 1 }
                };
                delete inferSpec.worker;

                [inferenceId, inferenceName] = await createDistributedInference(
                    inferenceName,
                    inferSpec,
                    project
                );

                createdWorkloadIds.push({ id: inferenceId, type: ApiWorkloadType.DISTRIBUTED_INFERENCE });
                logger.info(
                    `Created distributed inference workload with external URL: ${inferenceName} (ID: ${inferenceId})`
                );
            });

            await test.step('Wait for workload to be Running', async () => {
                await waitUtils.waitForBackendWorkloadByIdAndPhase(inferenceId, Phase.Running, 120);
                logger.info(`Workload ${inferenceName} is Running`);

                // Navigate to workload page for UI verification
                await workloads.goToPage();
                await workloads.searchByName(inferenceName);
            });

            await test.step('Navigate to Details tab', async () => {
                await detailsDrawer.clickShowDetailsBtn(inferenceName);
                await detailsDrawer.clickOnDetailsTab();
                logger.info('Navigated to Details tab');
            });

            await test.step('Verify Network section is displayed', async () => {
                const networkSection = detailsDrawer.getNetworkSection();
                await expect(networkSection).toBeVisible();
                logger.info('Network section is visible');
            });

            await test.step('Verify connection tool is displayed', async () => {
                const connectionTool = await detailsDrawer.getConnectionTool();
                expect(connectionTool).toBeTruthy();
                expect(connectionTool).toBe('custom');
                logger.info(`Connection tool: ${connectionTool}`);
            });

            await test.step('Verify connection type is displayed', async () => {
                const connectionType = await detailsDrawer.getConnectionType();
                expect(connectionType).toBeTruthy();
                expect(connectionType).toBe('Serving endpoint');
                logger.info(`Connection type: ${connectionType}`);
            });

            await test.step('Verify connection access is displayed', async () => {
                const connectionAccess = await detailsDrawer.getConnectionAccess();
                expect(connectionAccess).toBeTruthy();
                expect(connectionAccess).toBe('Public');
                logger.info(`Connection access: ${connectionAccess}`);
            });

            await test.step('Verify connection port is displayed', async () => {
                const connectionPort = await detailsDrawer.getConnectionPort();
                expect(connectionPort).toBeTruthy();
                expect(connectionPort).toBe('8888');
                logger.info(`Connection port: ${connectionPort}`);
            });
        });

        test('T5541624: Verify Network Section Display Without External Connections', async () => {
            videoName = 'network-section-no-connections';

            await test.step('Create a simple training workload without any external connections', async () => {
                const { body: training } = await workloadsCrudUtils.createTraining(
                    trainingName,
                    true,
                    flowRootState.cluster.uuid,
                    project.id.toString(),
                    {
                        image: 'ubuntu:20.04',
                        compute: { gpuDevicesRequest: 1 },
                        command: 'sleep',
                        args: 'infinity'
                    }
                );
                trainingName = training.name;
                createdWorkloadIds.push({ id: training.workloadId, type: ApiWorkloadType.TRAINING });
                logger.info(
                    `Created training workload without external connections: ${trainingName} (ID: ${training.workloadId})`
                );
            });

            await test.step('Wait for workload to be Running', async () => {
                await workloads.goToPage();
                await workloads.waitForWorkloadUIStatus(trainingName, EWorkloadStatus.Running);
                logger.info(`Workload ${trainingName} is Running`);
            });

            await test.step('Navigate to Details tab', async () => {
                await detailsDrawer.clickShowDetailsBtn(trainingName);
                await detailsDrawer.clickOnDetailsTab();
                logger.info('Navigated to Details tab');
            });

            await test.step('Verify Network section is not displayed for workloads without external connections', async () => {
                const networkInfo = await detailsDrawer.getNetworkSectionInfo();
                expect(networkInfo.isVisible).toBe(false);
                logger.info('Network section is not displayed as expected');
            });
        });

        test('T5541625: Verify Compute Section Display correctly', async () => {
            videoName = 'compute-section-display';

            await test.step('Create a workload requesting 2 GPUs', async () => {
                const { body: training } = await workloadsCrudUtils.createTraining(
                    trainingName,
                    true,
                    flowRootState.cluster.uuid,
                    project.id.toString(),
                    {
                        image: 'ubuntu:20.04',
                        compute: { gpuDevicesRequest: 2 },
                        command: 'sleep',
                        args: 'infinity'
                    }
                );
                trainingName = training.name;
                createdWorkloadIds.push({ id: training.workloadId, type: ApiWorkloadType.TRAINING });
                logger.info(
                    `Created training workload with 2 GPUs: ${trainingName} (ID: ${training.workloadId})`
                );
            });

            await test.step('Wait for workload to be Running', async () => {
                await workloads.goToPage();
                await workloads.waitForWorkloadUIStatus(trainingName, EWorkloadStatus.Running);
                logger.info(`Workload ${trainingName} is Running`);
            });

            await test.step('Navigate to Details tab', async () => {
                await detailsDrawer.clickShowDetailsBtn(trainingName);
                await detailsDrawer.clickOnDetailsTab();
                logger.info('Navigated to Details tab');
            });

            await test.step('Verify Compute section is displayed', async () => {
                const computeSection = detailsDrawer['getComputeSection']();
                await expect(computeSection).toBeVisible();
                await computeSection.scrollIntoViewIfNeeded();
                logger.info('Compute section is visible');
            });

            await test.step('Verify Compute section fields are displayed with correct values', async () => {
                const nodePools = await detailsDrawer.getNodePools();
                expect(nodePools).toBeTruthy();
                logger.info(`Node pools: ${nodePools}`);

                const gpuCompute = await detailsDrawer.getGpuComputeRequest();
                expect(gpuCompute).toBeTruthy();
                expect(gpuCompute).toContain('2');
                logger.info(`GPU compute: Request ${gpuCompute} (matches requested 2 GPUs)`);

                const numberOfPodsInDetails = await detailsDrawer.getNumberOfPods();
                expect(numberOfPodsInDetails).toBeUndefined();
            });

            await test.step('Verify Running / requested pods on Structure tab', async () => {
                await detailsDrawer.clickOnStructureTab();
                const numberOfPods = await detailsDrawer.getStructureRequestedPodsCount();
                expect(numberOfPods).toBeTruthy();
                expect(numberOfPods).toBe('1');
                logger.info(`Number of pods (Structure): ${numberOfPods}`);
            });
        });

        test('T5541626: Verify Compute Display correctly for CPU only workload', async () => {
            videoName = 'compute-section-cpu-only';

            await test.step('Create a workload requesting only CPU', async () => {
                const { body: training } = await workloadsCrudUtils.createTraining(
                    trainingName,
                    true,
                    flowRootState.cluster.uuid,
                    project.id.toString(),
                    {
                        image: 'ubuntu:20.04',
                        compute: {
                            cpuCoreRequest: 1,
                            cpuMemoryRequest: '1G'
                        },
                        command: 'sleep',
                        args: 'infinity'
                    }
                );
                trainingName = training.name;
                createdWorkloadIds.push({ id: training.workloadId, type: ApiWorkloadType.TRAINING });
                logger.info(
                    `Created CPU-only training workload: ${trainingName} (ID: ${training.workloadId})`
                );
            });

            await test.step('Wait for workload to be Running', async () => {
                await workloads.goToPage();
                await workloads.waitForWorkloadUIStatus(trainingName, EWorkloadStatus.Running);
                logger.info(`Workload ${trainingName} is Running`);
            });

            await test.step('Navigate to Details tab', async () => {
                await detailsDrawer.clickShowDetailsBtn(trainingName);
                await detailsDrawer.clickOnDetailsTab();
                logger.info('Navigated to Details tab');
            });

            await test.step('Verify Compute section is displayed', async () => {
                const computeSection = detailsDrawer['getComputeSection']();
                await expect(computeSection).toBeVisible();
                await computeSection.scrollIntoViewIfNeeded();
                logger.info('Compute section is visible');
            });

            await test.step('Verify Compute section fields for CPU-only workload', async () => {
                const nodePools = await detailsDrawer.getNodePools();
                expect(nodePools).toBeTruthy();
                logger.info(`Node pools: ${nodePools}`);

                const cpuCompute = await detailsDrawer.getCpuComputeRequest();
                expect(cpuCompute).toBeTruthy();
                expect(cpuCompute).toContain('1');
                logger.info(`CPU compute: Request ${cpuCompute} (matches requested 1 core)`);

                const cpuMemory = await detailsDrawer.getCpuMemoryRequest();
                expect(cpuMemory).toBeTruthy();
                expect(cpuMemory).toMatch(/1\s*GB/i);
                logger.info(`CPU memory: Request ${cpuMemory} (matches requested 1G)`);

                const numberOfPodsInDetails = await detailsDrawer.getNumberOfPods();
                expect(numberOfPodsInDetails).toBeUndefined();
            });

            await test.step('Verify Running / requested pods on Structure tab', async () => {
                await detailsDrawer.clickOnStructureTab();
                const numberOfPods = await detailsDrawer.getStructureRequestedPodsCount();
                expect(numberOfPods).toBeTruthy();
                expect(numberOfPods).toBe('1');
                logger.info(`Number of pods (Structure): ${numberOfPods}`);
            });

            await test.step('Verify no GPU information is shown for CPU-only workload', async () => {
                await detailsDrawer.clickOnDetailsTab();
                const gpuCompute = await detailsDrawer.getGpuComputeRequest();
                expect(gpuCompute).toBeUndefined();
                logger.info('GPU compute is not displayed (expected for CPU-only workload)');
            });
        });

        test('T5541627: Verify Number of Pods for Multi-Pod Workload', async () => {
            videoName = 'number-of-pods-multi-pod';
            let distributedTrainingName: string;

            await test.step('Create a distributed training workload with 2 workers', async () => {
                distributedTrainingName = objectGenerator.generateShortNameByPrefix('dis-training-pods');
                const { body: training } = await workloadsCrudUtils.createDistributed(
                    distributedTrainingName,
                    true,
                    flowRootState.cluster.uuid,
                    project.id.toString(),
                    {
                        image: 'ubuntu:20.04',
                        command: 'sleep',
                        args: 'infinity',
                        compute: { gpuDevicesRequest: 1 },
                        numWorkers: 2,
                        distributedFramework: DistributedFramework.PyTorch
                    }
                );
                distributedTrainingName = training.name;
                createdWorkloadIds.push({ id: training.workloadId, type: ApiWorkloadType.DISTRIBUTED });
                logger.info(
                    `Created distributed training workload with 2 workers: ${distributedTrainingName}`
                );
            });

            await test.step('Wait for workload to start', async () => {
                await workloads.goToPage();
                await workloads.waitForWorkloadUIStatus(distributedTrainingName, EWorkloadStatus.Running);
                logger.info(`Workload ${distributedTrainingName} is Running`);
            });

            await test.step('Navigate to Details tab', async () => {
                await detailsDrawer.clickShowDetailsBtn(distributedTrainingName);
                await detailsDrawer.clickOnDetailsTab();
                logger.info('Navigated to Details tab');
            });

            await test.step('Verify Compute section and scroll to it', async () => {
                const computeSection = detailsDrawer['getComputeSection']();
                await expect(computeSection).toBeVisible();
                await computeSection.scrollIntoViewIfNeeded();
                logger.info('Compute section is visible');
            });

            await test.step('Verify GPU Compute Number and Pods Number', async () => {
                const gpuCompute = await detailsDrawer.getGpuComputeRequest();
                expect(gpuCompute).toBeTruthy();
                expect(gpuCompute).toMatch(/2G/i);
                logger.info(`GPU compute: Request ${gpuCompute} (matches requested GPU)`);

                const numberOfPodsInDetails = await detailsDrawer.getNumberOfPods();
                expect(numberOfPodsInDetails).toBeTruthy();
                expect(numberOfPodsInDetails).toBe('2');
                logger.info(`Number of pods in Details: ${numberOfPodsInDetails} (expected: 2)`);
            });
        });

        test('T5541628: Verify CLI Command Section Display', async () => {
            test.setTimeout(60000);
            videoName = 'cli-command-section-display';
            let cliWorkloadName: string;

            await test.step('Submit a workload using CLI', async () => {
                cliWorkloadName = objectGenerator.generateShortNameByPrefix('cli-training');
                const submitFlags = {
                    [SubmitFlagsMerged.project]: project.name,
                    [SubmitFlagsMerged.image]: 'ubuntu:20.04',
                    [SubmitFlagsMerged.gpuRequest]: '1',
                    [SubmitFlagsMerged.command]: '-- sleep infinity'
                } as Partial<typeof SubmitFlagsMerged>;

                runaiCliV2.submit(submitFlags, cliWorkloadName, cliV2WorkloadType.training);
                logger.info(`Submitted workload via CLI: ${cliWorkloadName}`);

                const workloadId = await workloadsUtilsV2.getWorkloadIdByName(cliWorkloadName, project.id);
                createdWorkloadIds.push({ id: workloadId, type: ApiWorkloadType.TRAINING });
                logger.info(`Added CLI-submitted workload to cleanup list (ID: ${workloadId})`);
            });

            await test.step('Wait for workload to be Running', async () => {
                await workloads.goToPage();
                await workloads.waitForWorkloadUIStatus(cliWorkloadName, EWorkloadStatus.Running);
                logger.info(`Workload ${cliWorkloadName} is Running`);
            });

            await test.step('Navigate to Details tab', async () => {
                await detailsDrawer.clickShowDetailsBtn(cliWorkloadName);
                await detailsDrawer.clickOnDetailsTab();
                logger.info('Navigated to Details tab');
            });

            await test.step('Scroll to the bottom to locate CLI command section', async () => {
                const cliSection = detailsDrawer['getCliCommandSection']();
                await cliSection.scrollIntoViewIfNeeded();
                logger.info('Scrolled to CLI command section');
            });

            await test.step('Verify CLI command section is displayed', async () => {
                const cliInfo = await detailsDrawer.getCliCommandSectionInfo();
                expect(cliInfo.isVisible).toBe(true);
                logger.info('CLI command section is visible');
            });

            await test.step('Verify message text is displayed', async () => {
                const cliInfo = await detailsDrawer.getCliCommandSectionInfo();
                expect(cliInfo.messageText).toBeTruthy();
                expect(cliInfo.messageText).toContain(
                    'CLI command is available only for workloads submitted via the CLI'
                );
                expect(cliInfo.messageText).toContain('To learn how to submit a workload via CLI');
                expect(cliInfo.messageText).toContain('CLI Commands Reference');
                logger.info('Message text is displayed correctly');
            });

            await test.step('Verify "CLI Commands Reference" is a clickable link', async () => {
                const cliInfo = await detailsDrawer.getCliCommandSectionInfo();
                expect(cliInfo.docLink).toBeDefined();
                expect(cliInfo.docLink?.text).toContain('CLI Commands Reference');
                expect(cliInfo.docLink?.href).toBeTruthy();
                logger.info(`"CLI Commands Reference" is a clickable link: ${cliInfo.docLink?.href}`);
            });
        });

        test('T5541629: Verify CLI Command Display for CLI-Submitted Workload', async () => {
            videoName = 'cli-command-display-cli-submitted';
            let cliWorkloadName: string;

            await test.step('Submit a workload using CLI', async () => {
                cliWorkloadName = objectGenerator.generateShortNameByPrefix('cli-training');
                const submitFlags = {
                    [SubmitFlagsMerged.project]: project.name,
                    [SubmitFlagsMerged.image]: 'ubuntu:20.04',
                    [SubmitFlagsMerged.gpuRequest]: '1',
                    [SubmitFlagsMerged.command]: '-- sleep infinity'
                } as Partial<typeof SubmitFlagsMerged>;

                runaiCliV2.submit(submitFlags, cliWorkloadName, cliV2WorkloadType.training);
                logger.info(`Submitted workload via CLI: ${cliWorkloadName}`);

                const workloadId = await workloadsUtilsV2.getWorkloadIdByName(cliWorkloadName, project.id);
                createdWorkloadIds.push({ id: workloadId, type: ApiWorkloadType.TRAINING });
                logger.info(`Added CLI-submitted workload to cleanup list (ID: ${workloadId})`);
            });

            await test.step('Wait for workload to appear in UI', async () => {
                await workloads.goToPage();
                await workloads.waitForWorkloadUIStatus(cliWorkloadName, EWorkloadStatus.Running);
                logger.info(`Workload ${cliWorkloadName} is visible in UI and Running`);
            });

            await test.step('Navigate to Details tab', async () => {
                await detailsDrawer.clickShowDetailsBtn(cliWorkloadName);
                await detailsDrawer.clickOnDetailsTab();
                logger.info('Navigated to Details tab');
            });

            await test.step('Scroll to CLI command section', async () => {
                const cliSection = detailsDrawer['getCliCommandSection']();
                await cliSection.scrollIntoViewIfNeeded();
                logger.info('Scrolled to CLI command section');
            });

            await test.step('Verify CLI command section displays actual CLI command', async () => {
                const cliInfo = await detailsDrawer.getCliCommandSectionInfo();
                expect(cliInfo.isVisible).toBe(true);
                expect(cliInfo.cliCommand).toBeTruthy();
                expect(cliInfo.cliCommand).toContain(cliWorkloadName);
                expect(cliInfo.cliCommand).toContain('training');
                logger.info(`CLI command is displayed: ${cliInfo.cliCommand}`);
            });

            await test.step('Verify CLI command can be selected and copied', async () => {
                const cliInfo = await detailsDrawer.getCliCommandSectionInfo();
                const command = cliInfo.cliCommand || '';
                expect(command.trim().length).toBeGreaterThan(0);
                expect(command).toMatch(/runai.*training.*submit/);
                logger.info('CLI command is properly formatted and can be selected for copying');
            });
        });

        test('T5541630: Verify CLI Command Section is Not Displayed for UI-Submitted Workload', async () => {
            videoName = 'cli-command-ui-submitted';

            await test.step('Create a training workload via API (UI-submitted)', async () => {
                const { body: training } = await workloadsCrudUtils.createTraining(
                    trainingName,
                    true,
                    flowRootState.cluster.uuid,
                    project.id.toString(),
                    {
                        image: 'ubuntu:20.04',
                        compute: { gpuDevicesRequest: 1 },
                        command: 'sleep',
                        args: 'infinity'
                    }
                );
                trainingName = training.name;
                createdWorkloadIds.push({ id: training.workloadId, type: ApiWorkloadType.TRAINING });
                logger.info(`Created training workload: ${trainingName} with ID: ${training.workloadId}`);
            });

            await test.step('Wait for workload to be Running', async () => {
                await workloads.goToPage();
                await workloads.waitForWorkloadUIStatus(trainingName, EWorkloadStatus.Running);
                logger.info(`Workload ${trainingName} is Running`);
            });

            await test.step('Navigate to Details tab', async () => {
                await detailsDrawer.clickShowDetailsBtn(trainingName);
                await detailsDrawer.clickOnDetailsTab();
                logger.info('Navigated to Details tab');
            });

            await test.step('Verify CLI command section is not displayed', async () => {
                const cliInfo = await detailsDrawer.getCliCommandSectionInfo();
                expect(cliInfo.isVisible).toBe(false);
                logger.info('CLI command section is not displayed for UI-submitted workload');
            });
        });

        test('T5541631: Select multi workloads to view Details', async () => {
            videoName = 'select-multi-workloads-details';
            let workload1Name: string;
            let workload2Name: string;

            await test.step('Create first workload', async () => {
                workload1Name = objectGenerator.generateShortNameByPrefix('training-1');
                const { body: training1 } = await workloadsCrudUtils.createTraining(
                    workload1Name,
                    true,
                    flowRootState.cluster.uuid,
                    project.id.toString(),
                    {
                        image: 'ubuntu:20.04',
                        command: 'sleep',
                        args: 'infinity',
                        compute: { gpuDevicesRequest: 1 }
                    }
                );
                workload1Name = training1.name;
                createdWorkloadIds.push({ id: training1.workloadId, type: ApiWorkloadType.TRAINING });
                logger.info(`Created first training workload: ${workload1Name}`);
            });

            await test.step('Create second workload', async () => {
                workload2Name = objectGenerator.generateShortNameByPrefix('training-2');
                const { body: training2 } = await workloadsCrudUtils.createTraining(
                    workload2Name,
                    true,
                    flowRootState.cluster.uuid,
                    project.id.toString(),
                    {
                        image: 'ubuntu:20.04',
                        command: 'sleep',
                        args: 'infinity',
                        compute: { gpuDevicesRequest: 1 }
                    }
                );
                workload2Name = training2.name;
                createdWorkloadIds.push({ id: training2.workloadId, type: ApiWorkloadType.TRAINING });
                logger.info(`Created second training workload: ${workload2Name}`);
            });

            await test.step('Wait for both workloads to be Running', async () => {
                await workloads.goToPage();
                await workloads.waitForWorkloadUIStatus(workload1Name, EWorkloadStatus.Running);
                await workloads.waitForWorkloadUIStatus(workload2Name, EWorkloadStatus.Running);
                logger.info(`Both workloads are Running: ${workload1Name}, ${workload2Name}`);
            });

            await test.step('Select both workloads', async () => {
                await workloads.clearGridSearchInput();
                await workloads.selectWorkloadsByNames([workload1Name, workload2Name]);
            });

            await test.step('Click show details button', async () => {
                await detailsDrawer.clickShowDetailsBtn();
            });

            await test.step('Verify multi-selection prompt is displayed', async () => {
                const isPromptVisible = await detailsDrawer.verifyMultiSelectionPrompt();
                expect(isPromptVisible).toBe(true);
                logger.info(
                    'Multi-selection prompt "Select a single workload to view its details." is displayed'
                );
            });
        });

        test('T5541632: Verify Master/Worker View Switching for Distributed Workloads', async () => {
            videoName = 'distributed-master-worker-switching';
            let workloadName: string;
            let masterCommand: string;
            let masterImageUrl: string;

            await test.step('Create distributed training workload with 2 workers', async () => {
                workloadName = objectGenerator.generateShortNameByPrefix('distributed');
                const { body: distributed } = await workloadsCrudUtils.createDistributed(
                    workloadName,
                    true,
                    flowRootState.cluster.uuid,
                    project.id.toString(),
                    {
                        image: 'ubuntu:20.04',
                        command: 'sleep',
                        args: 'infinity',
                        compute: { gpuDevicesRequest: 1 },
                        numWorkers: 2,
                        distributedFramework: DistributedFramework.PyTorch
                    },
                    true,
                    flowRootState.users.system_administrator.accessToken
                );
                workloadName = distributed.name;
                createdWorkloadIds.push({ id: distributed.workloadId, type: ApiWorkloadType.TRAINING });
                logger.info(`Created distributed workload: ${workloadName} with 2 workers and 1 master`);
            });

            await test.step('Wait for workload to reach Running state', async () => {
                await workloads.goToPage();
                await workloads.waitForWorkloadUIStatus(workloadName, EWorkloadStatus.Running);
                logger.info(`Workload ${workloadName} is Running`);
            });

            await test.step('Navigate to workload Details tab', async () => {
                await detailsDrawer.clickShowDetailsBtn(workloadName);
                await detailsDrawer.clickOnDetailsTab();
            });

            await test.step('Verify pod selector dropdown is visible', async () => {
                const isPodSelectorVisible = await detailsDrawer.isPodSelectorVisible();
                expect(isPodSelectorVisible).toBe(true);
                logger.info('Pod selector dropdown is visible');
            });

            await test.step('Verify pod selector options', async () => {
                let podOptions: string[] = [];
                await waitUtils.waitForCondition(
                    async () => {
                        podOptions = await detailsDrawer.getPodSelectorOptions();
                        return podOptions.length > 0;
                    },
                    'Waiting for pod selector options to be available',
                    true,
                    60,
                    5
                );

                expect(podOptions.length).toBeGreaterThan(0);
                logger.info(`Pod selector has ${podOptions.length} options: ${podOptions.join(', ')}`);

                const hasMaster = podOptions.some((option) => option.toLowerCase().includes('master'));
                const hasWorkers = podOptions.some((option) => option.toLowerCase().includes('worker'));
                expect(hasMaster).toBe(true);
                expect(hasWorkers).toBe(true);
                expect(podOptions.length).toBe(2);
            });

            await test.step('Select Master pod and verify General Information', async () => {
                const podOptions = await detailsDrawer.getPodSelectorOptions();
                const masterOption = podOptions.find((option) => option.toLowerCase().includes('master'));
                expect(masterOption).toBeDefined();
                await detailsDrawer.selectWorkerPod(masterOption!);
                logger.info(`Selected Master pod: ${masterOption}`);

                // Verify General Information fields
                const cluster = await detailsDrawer.getClusterName();
                expect(cluster).toBe(flowRootState.cluster.name);
                logger.info(`Cluster: ${cluster}`);

                const department = await detailsDrawer.getDepartment();
                expect(department).toBeTruthy();
                logger.info(`Department: ${department}`);

                const projectName = await detailsDrawer.getProject();
                expect(projectName).toBe(project.name);
                logger.info(`Project: ${projectName}`);

                const category = await detailsDrawer.getCategory();
                expect(category).toBe('Train');
                logger.info(`Category: ${category}`);

                const type = await detailsDrawer.getType();
                expect(type).toBe('Distributed');
                logger.info(`Type: ${type}`);

                const priority = await detailsDrawer.getPriority();
                expect(priority).toBe('low');
                logger.info(`Priority: ${priority}`);

                const preemptible = await detailsDrawer.getPreemptible();
                expect(preemptible).toBeTruthy();
                logger.info(`Preemptible: ${preemptible}`);

                const creationTime = await detailsDrawer.getCreationTime();
                expect(creationTime).toBeDefined();
                logger.info(`Creation time: ${creationTime}`);

                const createdBy = await detailsDrawer.getCreatedBy();
                expect(createdBy).toBe(flowRootState.users.system_administrator.email);
                logger.info(`Created by: ${createdBy}`);
            });

            await test.step('Verify Compute section shows workload-level aggregated information', async () => {
                const nodePools = await detailsDrawer.getNodePools();
                expect(nodePools).toBeTruthy();
                logger.info(`Node pools: ${nodePools}`);

                const gpuCompute = await detailsDrawer.getGpuComputeRequest();
                expect(gpuCompute).toContain('Request 3G');
                logger.info(`GPU compute: ${gpuCompute}`);

                const numberOfPodsInDetails = await detailsDrawer.getNumberOfPods();
                expect(numberOfPodsInDetails).toBe('3');
                logger.info(`Number of pods in Details: ${numberOfPodsInDetails} (expected: 3)`);
            });

            await test.step('Note Master component-specific information', async () => {
                masterImageUrl = await detailsDrawer.getImageUrl();
                expect(masterImageUrl).toBeTruthy();
                logger.info(`Master Image URL: ${masterImageUrl}`);

                masterCommand = (await detailsDrawer.getCommand()) || '';
                logger.info(`Master Command: ${masterCommand}`);

                const masterArguments = await detailsDrawer.getArguments();
                logger.info(`Master Arguments: ${masterArguments}`);
            });

            await test.step('Switch to Worker pod', async () => {
                const podOptions = await detailsDrawer.getPodSelectorOptions();
                const workerOption = podOptions.find((option) => option.toLowerCase().includes('worker'));
                expect(workerOption).toBeDefined();
                await detailsDrawer.selectWorkerPod(workerOption!);
                logger.info(`Switched to Worker pod: ${workerOption}`);
            });

            await test.step('Verify General Information remains consistent for Worker', async () => {
                const type = await detailsDrawer.getType();
                expect(type).toBe('Distributed');
                logger.info(`Type (Worker view): ${type}`);

                const category = await detailsDrawer.getCategory();
                expect(category).toBe('Train');
                logger.info(`Category (Worker view): ${category}`);

                const projectName = await detailsDrawer.getProject();
                expect(projectName).toBe(project.name);
                logger.info(`Project (Worker view): ${projectName}`);

                const cluster = await detailsDrawer.getClusterName();
                expect(cluster).toBe(flowRootState.cluster.name);
                logger.info(`Cluster (Worker view): ${cluster}`);
            });

            await test.step('Verify Compute section workload-level information remains consistent', async () => {
                const nodePools = await detailsDrawer.getNodePools();
                expect(nodePools).toBeTruthy();
                logger.info(`Node pools (Worker view): ${nodePools}`);

                const gpuCompute = await detailsDrawer.getGpuComputeRequest();
                expect(gpuCompute).toContain('Request 3G');
                logger.info(`GPU compute (Worker view): ${gpuCompute}`);

                const numberOfPodsInDetails = await detailsDrawer.getNumberOfPods();
                expect(numberOfPodsInDetails).toBe('3');
                logger.info(
                    `Number of pods in Details (Worker view): ${numberOfPodsInDetails} (expected: 3)`
                );
            });

            await test.step('Verify Worker component-specific information', async () => {
                const workerImageUrl = await detailsDrawer.getImageUrl();
                expect(workerImageUrl).toBeTruthy();
                logger.info(`Worker Image URL: ${workerImageUrl}`);

                const workerCommand = await detailsDrawer.getCommand();
                logger.info(`Worker Command: ${workerCommand}`);

                const workerArguments = await detailsDrawer.getArguments();
                logger.info(`Worker Arguments: ${workerArguments}`);
            });
        });
    });
});
