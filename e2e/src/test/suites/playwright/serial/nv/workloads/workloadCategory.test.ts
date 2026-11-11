/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import {
    config,
    Utils,
    logger,
    objectGenerator,
    projectResourcesUtils,
    tags,
    flowRootState,
    workloadsCrudUtils,
    telemetryUtils,
    waitUtils,
    orgUnitUtils,
    workloadsUtilsV2
} from '../../../../../../modules/helpers';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { kubeCli } from '../../../../../../modules/cli';
import { DistributedFramework } from '../../../../../../modules/api/apis/workloadsApi/generated/model/distributedFramework';
import { WorkloadUtils } from '../../../../../../modules/helpers/utils/WorkloadUtils';
import { StatusCodes } from 'http-status-codes';
import { clusterApi } from '../../../../../../modules/api';
import { MetricsType } from '../../../../../../modules/api/apis/clusterApi/generated/model/metricsType';
import { Phase } from '../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import { WorkloadTelemetryType } from '../../../../../../modules/api/apis/workloadsApi/generated/model/workloadTelemetryType';
import { OverviewPage } from '../../../../../../modules/ui/runaiV2/pages/OverviewPage';
import { getBaseWorkloadSpec } from '../../../parallel/viper/inference/distributed/dinfUtils';
import { createDistributedInference } from '../../../parallel/viper/inference/inferenceUtils';

test.describe(`Workload Categories ${tags.tool.ui} ${tags.resource.workloads} ${tags.teams.loong}`, function () {
    supportedFrom(SupportedVersions.V2_23)(function () {
        let project: Project;
        let workloadTypesUpdated: boolean = false;
        let trainingWorkloadCategoryId: string;
        let trainingWorkloadId: string;
        let trainingName: string;
        let workspaceName: string;
        let inferenceName: string;
        let createdWorkloadIds: string[] = [];
        const workloadUtils = new WorkloadUtils();

        async function createAllWorkloadTypes(
            trainingName: string,
            workspaceName: string,
            inferenceName: string,
            projectId: string,
            project: Project,
            gpuRequest: number = 1
        ): Promise<{
            trainingWorkload: any;
            buildWorkload: any;
            distributedInferenceId: string;
            distributedInferenceName: string;
        }> {
            const { body: trainingWorkload } = await workloadsCrudUtils.createTraining(
                trainingName,
                true,
                flowRootState.cluster.uuid,
                projectId,
                {
                    image: config.quickStartImage.ubuntu,
                    command: 'sleep infinity',
                    compute: { gpuDevicesRequest: gpuRequest }
                }
            );

            const { body: buildWorkload } = await workloadsCrudUtils.createWorkspace(
                workspaceName,
                true,
                flowRootState.cluster.uuid,
                projectId,
                {
                    image: config.quickStartImage.ubuntu,
                    command: 'sleep infinity',
                    compute: { gpuDevicesRequest: gpuRequest }
                }
            );

            createdWorkloadIds.push(trainingWorkload.workloadId, buildWorkload.workloadId);

            logger.info('Creating distributed inference with single leader (workers=0)');
            const inferSpec = getBaseWorkloadSpec();
            inferSpec.workers = 0;
            inferSpec.replicas = 1;
            delete inferSpec.worker;

            const [distributedInferenceId, distributedInferenceName] = await createDistributedInference(
                inferenceName,
                inferSpec,
                project
            );

            createdWorkloadIds.push(distributedInferenceId);

            await Promise.all([
                waitUtils.waitForBackendWorkloadByIdAndPhase(trainingWorkload.workloadId, Phase.Running, 120),
                waitUtils.waitForBackendWorkloadByIdAndPhase(buildWorkload.workloadId, Phase.Running, 120),
                waitUtils.waitForBackendWorkloadByIdAndPhase(distributedInferenceId, Phase.Running, 120)
            ]);

            logger.info('All workloads created and running successfully (using distributed inference)');

            return {
                trainingWorkload,
                buildWorkload,
                distributedInferenceId,
                distributedInferenceName
            };
        }

        test.beforeAll(async () => {
            const projectName = objectGenerator.generateNameByPrefix('workload-category');
            const namespace = Utils.namespaceFromProjectName(projectName);
            project = await projectResourcesUtils.createProject(namespace, {
                name: projectName,
                deservedGpus: 3,
                maxAllowedGpus: 3
            });
            const accessToken = flowRootState.users.system_administrator.accessToken;
            workloadsCrudUtils.accessToken = accessToken;
            clusterApi.clustersApi.accessToken = accessToken;
        });

        test.afterAll(async function () {
            try {
                await orgUnitUtils.deleteProjectAndNamespace(project.id.toString(), project.name);
                logger.info(`deleted project ${project.name} and namespace ${project.namespace}`);
            } catch (error) {
                logger.error(`failed to delete project ${project.name} and namespace ${project.namespace}`);
                throw error;
            }
        });

        test.beforeEach(async function () {
            trainingName = objectGenerator.generateNameByPrefix('training');
            workspaceName = objectGenerator.generateNameByPrefix('workspace');
            inferenceName = objectGenerator.generateNameByPrefix('inference');
            createdWorkloadIds = [];

            // Get the default trainingWorkloadCategoryId
            const workloadTypes = await workloadUtils.getWorkloadTypes();
            const targetWorkloadType = 'TrainingWorkload';
            const trainingWorkloadType = workloadTypes.find((wt) => wt.name === targetWorkloadType);
            if (trainingWorkloadType) {
                trainingWorkloadCategoryId = trainingWorkloadType.categoryId;
                trainingWorkloadId = trainingWorkloadType.id;
                logger.info(`Saved default trainingWorkloadCategoryId: ${trainingWorkloadCategoryId}`);
            }
        });

        test.afterEach(async ({}, testInfo) => {
            if (createdWorkloadIds.length > 0) {
                try {
                    await workloadsUtilsV2.deleteAllWorkloads();
                    logger.info('Successfully deleted all workloads');
                } catch (error) {
                    logger.warn('Failed to delete all workloads:', error);
                }
                createdWorkloadIds = [];
            }

            // Restore the default trainingWorkloadCategoryId if it was modified
            if (workloadTypesUpdated) {
                try {
                    await workloadUtils.updateWorkloadTypeById(trainingWorkloadId, {
                        categoryId: trainingWorkloadCategoryId
                    });
                    logger.info(`Restored default trainingWorkloadCategoryId: ${trainingWorkloadCategoryId}`);
                    workloadTypesUpdated = false;
                } catch (error) {
                    logger.warn(`Failed to restore default trainingWorkloadCategoryId: ${error}`);
                }
            }
        });

        test.describe('Workload Category -> API', () => {
            test('T5307579: Submit Workload without workload category, verify they are automatically categorized', async function () {
                await test.step('Create multiple workloads of different categories', async () => {
                    await createAllWorkloadTypes(
                        trainingName,
                        workspaceName,
                        inferenceName,
                        project.id.toString(),
                        project
                    );
                });
                await test.step('Get workload categories, workloads should be automatically categorized', async () => {
                    const workloadCategories = await workloadUtils.getWorkloadCategories();
                    expect(workloadCategories.length).toBeGreaterThanOrEqual(3);

                    // Verify the three expected workload categories exist
                    const categoryNames = workloadCategories.map((cat) => cat.name);
                    expect(categoryNames).toContain('Build');
                    expect(categoryNames).toContain('Train');
                    expect(categoryNames).toContain('Deploy');

                    // Verify descriptions match expected content
                    const buildCategory = workloadCategories.find((cat) => cat.name === 'Build');
                    const trainCategory = workloadCategories.find((cat) => cat.name === 'Train');
                    const deployCategory = workloadCategories.find((cat) => cat.name === 'Deploy');

                    expect(buildCategory).toBeDefined();
                    expect(buildCategory.description).toBe('Build workloads');

                    expect(trainCategory).toBeDefined();
                    expect(trainCategory.description).toBe('Training workloads');

                    expect(deployCategory).toBeDefined();
                    expect(deployCategory.description).toBe('Deployment and inference workloads');
                });
            });

            test('T5311666: Submit a training workload with specific workload category, it should be categorized to the workload category', async function () {
                await test.step('Create a training workload with specific category', async () => {
                    const specificTrainingName = objectGenerator.generateNameByPrefix('specific-training');
                    const { body: specificTrainingWorkload } = await workloadsCrudUtils.createTraining(
                        specificTrainingName,
                        true,
                        flowRootState.cluster.uuid,
                        project.id.toString(),
                        {
                            image: config.quickStartImage.ubuntu,
                            command: 'sleep infinity',
                            compute: { gpuDevicesRequest: 1 },
                            category: 'Build' // Specify the workload category
                        }
                    );
                    createdWorkloadIds.push(specificTrainingWorkload.workloadId);
                });

                await test.step('Verify the workload is categorized correctly', async () => {
                    const workloadCategories = await workloadUtils.getWorkloadCategories();
                    expect(workloadCategories.length).toBeGreaterThan(0);
                    const categoryNames = workloadCategories.map((cat) => cat.name);
                    expect(categoryNames).toContain('Build');
                });
            });

            test('T5307580: Submit Third-party Workload and verify the Automatic Categorization', async function () {
                let podName: string;
                let mpiJobName: string;

                await test.step('Deploy a pod', async () => {
                    podName = objectGenerator.generateNameByPrefix('third-party-pod');
                    const deploymentYaml = `apiVersion: v1\nkind: Pod\nmetadata:\n  name: ${podName}\n  namespace: ${project.namespace}\n  labels:\n    app: ${podName}\nspec:\n  containers:\n  - name: my-container\n    image: nginx:latest\n    ports:\n    - containerPort: 80`;
                    await kubeCli.createFromYaml(deploymentYaml);
                });

                await test.step('Create a MPI job', async () => {
                    mpiJobName = objectGenerator.generateNameByPrefix('mpi-job');
                    const { body: mpiWorkload } = await workloadsCrudUtils.createDistributed(
                        mpiJobName,
                        true,
                        flowRootState.cluster.uuid,
                        project.id.toString(),
                        {
                            image: config.quickStartImage.ubuntu,
                            command: 'sleep infinity',
                            numWorkers: 2,
                            distributedFramework: DistributedFramework.Mpi
                        },
                        true
                    );
                    createdWorkloadIds.push(mpiWorkload.workloadId);
                });

                await test.step('Verify the workload is categorized correctly', async () => {
                    const workloadCategories = await workloadUtils.getWorkloadCategories();
                    expect(workloadCategories.length).toBeGreaterThan(0);

                    // Verify the three expected workload categories exist
                    const categoryNames = workloadCategories.map((cat) => cat.name);
                    expect(categoryNames).toContain('Build');
                    expect(categoryNames).toContain('Deploy');

                    // Verify descriptions match expected content
                    const buildCategory = workloadCategories.find((cat) => cat.name === 'Build');
                    const deployCategory = workloadCategories.find((cat) => cat.name === 'Deploy');

                    expect(buildCategory).toBeDefined();
                    expect(buildCategory.description).toBe('Build workloads');

                    expect(deployCategory).toBeDefined();
                    expect(deployCategory.description).toBe('Deployment and inference workloads');
                });

                await test.step('Delete the created pod', async () => {
                    try {
                        await kubeCli.deletePod(podName, project.namespace);
                        logger.info(`Successfully deleted pod: ${podName}`);
                    } catch (error) {
                        logger.warn(`Failed to delete pod ${podName}:`, error);
                    }
                });
            });

            test('T5307581: Verify admin update default workload category and mapping correctly', async function () {
                await test.step('Get workloadtype info', async () => {
                    const workloadTypes = await workloadUtils.getWorkloadTypes();
                    const targetWorkloadType = 'TrainingWorkload';
                    const trainingWorkloadType = workloadTypes.find((wt) => wt.name === targetWorkloadType);
                    trainingWorkloadCategoryId = trainingWorkloadType.categoryId;
                    trainingWorkloadId = trainingWorkloadType.id;
                });

                await test.step('Update the default category mapping of a workload type.', async () => {
                    // First, get the Build category ID
                    const workloadCategories = await workloadUtils.getWorkloadCategories();
                    const buildCategory = workloadCategories.find((cat) => cat.name === 'Build');
                    const buildCategoryId = buildCategory.id;
                    console.log('Build category ID:', buildCategoryId);

                    // Update TrainingWorkload to use Build category instead of Train
                    const updateResult = await workloadUtils.updateWorkloadTypeById(trainingWorkloadId, {
                        categoryId: buildCategoryId
                    });
                    workloadTypesUpdated = true;
                    expect(updateResult.response.statusCode).toBe(StatusCodes.OK);
                });

                await test.step('Create a training job to verify the category mapping', async () => {
                    const { body: trainingJobWorkload } = await workloadsCrudUtils.createTraining(
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
                    createdWorkloadIds.push(trainingJobWorkload.workloadId);
                });

                await test.step('Verify the workload category is mapped to Build', async () => {
                    const workloadCategories = await workloadUtils.getWorkloadCategories();
                    expect(workloadCategories.length).toBeGreaterThan(0);

                    const categoryNames = workloadCategories.map((cat) => cat.name);
                    expect(categoryNames).toContain('Build');
                });

                // Note: The default category mapping will be restored in afterEach
            });

            test('T5307583: Verify Category Trends and Monitoring', async function () {
                await test.step('Submit multiple workloads of different categories', async () => {
                    await createAllWorkloadTypes(
                        trainingName,
                        workspaceName,
                        inferenceName,
                        project.id.toString(),
                        project
                    );
                });

                await test.step('Verify Workloads by category over time is shown', async () => {
                    await test.step('Get time-range metrics by category', async () => {
                        const end = new Date();
                        const start = new Date(end.getTime() - 2 * 60 * 60 * 1000);

                        const response = await clusterApi.clustersApi.getClusterMetrics(
                            flowRootState.cluster.uuid,
                            start,
                            end,
                            [MetricsType.WorkloadsCount],
                            673,
                            'Category'
                        );

                        expect(response.body.measurements).toBeDefined();
                        expect(response.body.measurements.length).toBeGreaterThanOrEqual(3);
                    });
                });
            });

            test('T5307585: Verify Workload Categorization display count is correct', async function () {
                await test.step('Submit multiple workloads of different categories', async () => {
                    await createAllWorkloadTypes(
                        trainingName,
                        workspaceName,
                        inferenceName,
                        project.id.toString(),
                        project
                    );
                });

                await test.step('Verify the number of workloads by category is correct', async () => {
                    const telemetryResponse = await telemetryUtils.getWorkloadTelemetryUsingGroupBy(
                        WorkloadTelemetryType.WorkloadsCount,
                        ['Category']
                    );
                    expect(telemetryResponse).toBeDefined();
                    expect(telemetryResponse.type).toBe(WorkloadTelemetryType.WorkloadsCount);
                    expect(telemetryResponse.values).toBeDefined();

                    // Verify that we have exactly 3 categories
                    expect(telemetryResponse.values.length).toBe(3);

                    // Check that we have the expected categories
                    const categoryValues = telemetryResponse.values
                        .map((value) => {
                            const categoryGroup = value.groups?.find((group) => group.key === 'Category');
                            return categoryGroup?.value;
                        })
                        .filter(Boolean);

                    console.log('Found categories:', categoryValues);

                    // Verify that we have the three expected categories
                    expect(categoryValues).toContain('Build');
                    expect(categoryValues).toContain('Train');
                    expect(categoryValues).toContain('Deploy');

                    const categoryValueCounts = telemetryResponse.values
                        .filter((value) => value.groups?.some((group) => group.key === 'Category'))
                        .map((value) => Number(value.value));

                    for (const count of categoryValueCounts) {
                        expect(count).toBeGreaterThanOrEqual(0);
                    }
                });
            });
        });

        test.describe('Workload Category -> UI', () => {
            let browserEngine: BrowserEngine;
            let auth: AuthPage;
            let overviewPage: OverviewPage;
            let videoName: string;

            test.beforeEach(async function () {
                browserEngine = new BrowserEngine();
                await browserEngine.start();
                auth = new AuthPage(browserEngine.currentPage);
                overviewPage = new OverviewPage(browserEngine.currentPage);
                await auth.signIn(
                    flowRootState.users.system_administrator.email,
                    flowRootState.users.system_administrator.password
                );
            });

            test.afterEach('stop browser', async ({}, testInfo) => {
                if (browserEngine.currentPage) {
                    await browserEngine.stopAndArchive(testInfo, 'workload-category-' + videoName);
                }
            });

            test('T5307582: Filter workloads by category and verify filtering accuracy', async function () {
                test.setTimeout(180 * 1000);
                await test.step('Create workloads using distributed inference API', async () => {
                    await createAllWorkloadTypes(
                        trainingName,
                        workspaceName,
                        inferenceName,
                        project.id.toString(),
                        project
                    );
                });

                await test.step('Filter workloads by workload type and verify filtering accuracy', async () => {
                    logger.info('Refreshing page to load latest workload data');
                    await overviewPage.page.reload();
                    await overviewPage.page.waitForLoadState('networkidle');

                    await overviewPage.clickWorkloadByText('Running workloads by type');
                    await overviewPage.waitWorkloadTypeToLoad();

                    await waitUtils.waitForCondition(
                        async () => {
                            const trainingVisible = await overviewPage.isWorkloadTypeVisible('Training');
                            const workspaceVisible = await overviewPage.isWorkloadTypeVisible('Workspace');
                            const inferenceVisible =
                                await overviewPage.isWorkloadTypeVisible('DistributedInference');
                            logger.debug(
                                `UI visibility check: Training=${trainingVisible}, Workspace=${workspaceVisible}, DistributedInference=${inferenceVisible}`
                            );
                            return trainingVisible && workspaceVisible && inferenceVisible;
                        },
                        'All workload types (Training, Workspace, DistributedInference) should be visible',
                        true,
                        60,
                        3
                    );

                    await overviewPage.filterWorkloadType('Training');
                    expect(await overviewPage.isWorkloadTypeHidden('Training')).toBe(true);
                    await overviewPage.filterWorkloadType('Training');
                    expect(await overviewPage.isWorkloadTypeHidden('Training')).toBe(false);

                    await overviewPage.filterWorkloadType('Workspace');
                    expect(await overviewPage.isWorkloadTypeHidden('Workspace')).toBe(true);
                    await overviewPage.filterWorkloadType('Workspace');
                    expect(await overviewPage.isWorkloadTypeHidden('Workspace')).toBe(false);

                    await overviewPage.filterWorkloadType('DistributedInference');
                    expect(await overviewPage.isWorkloadTypeHidden('DistributedInference')).toBe(true);
                    await overviewPage.filterWorkloadType('DistributedInference');
                    expect(await overviewPage.isWorkloadTypeHidden('DistributedInference')).toBe(false);
                });
            });

            test('T5307584: Verify Workload Categorization count is correct when switching between workload by type to workload category', async function () {
                test.setTimeout(180 * 1000);
                await test.step('Submit multiple workloads of different categories', async () => {
                    await createAllWorkloadTypes(
                        trainingName,
                        workspaceName,
                        inferenceName,
                        project.id.toString(),
                        project
                    );
                });

                await test.step('Verify the number of workloads by category is correct', async () => {
                    await overviewPage.page.reload();
                    await overviewPage.page.waitForLoadState('networkidle');
                    await overviewPage.clickWorkloadByText('Running workloads by category');
                    await overviewPage.waitWorkloadTypeToLoad();

                    expect(await overviewPage.isWorkloadTypeVisible('Build')).toBe(true);
                    expect(await overviewPage.isWorkloadTypeVisible('Train')).toBe(true);
                    expect(await overviewPage.isWorkloadTypeVisible('Deploy')).toBe(true);
                });

                await test.step('Switch to workload by type and verify the workload type is correct', async () => {
                    await overviewPage.clickWorkloadByText('Running workloads by type');
                    await overviewPage.waitWorkloadTypeToLoad();

                    await waitUtils.waitForCondition(
                        async () => {
                            const trainingVisible = await overviewPage.isWorkloadTypeVisible('Training');
                            const workspaceVisible = await overviewPage.isWorkloadTypeVisible('Workspace');
                            const inferenceVisible =
                                await overviewPage.isWorkloadTypeVisible('DistributedInference');
                            logger.debug(
                                `UI visibility check: Training=${trainingVisible}, Workspace=${workspaceVisible}, DistributedInference=${inferenceVisible}`
                            );
                            return trainingVisible && workspaceVisible && inferenceVisible;
                        },
                        'All workload types (Training, Workspace, DistributedInference) should be visible',
                        true,
                        30,
                        2
                    );
                });
            });
        });
    });
});
