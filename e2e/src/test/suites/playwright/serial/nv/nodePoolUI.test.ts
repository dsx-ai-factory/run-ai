/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect, DEBUG_TAGS } from '../../../../helpers/DebugInfoCollector';
import { tags } from '../../../../../modules/helpers/config/tags';
import { supportedFrom, SupportedVersions } from '../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import {
    flowRootState,
    objectGenerator,
    operations,
    projectResourcesUtils,
    Utils,
    workloadUtils,
    orgUnitUtils,
    waitUtils,
    workloadsCrudUtils
} from '../../../../../modules/helpers';
import { AuthPage } from '../../../../../modules/ui/runaiV2/pages/AuthPage';
import { NodePoolPage } from '../../../../../modules/ui/runaiV2/pages/NodePoolPage';
import { WorkloadPage, EWorkloadStatus } from '../../../../../modules/ui/runaiV2/pages/WorkloadPage';
import { BrowserEngine } from '../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import {
    NodePool,
    NodePoolCreateRequest,
    PlacementStrategy
} from '../../../../../modules/api/apis/backendClusterApi';
import { Project as V1Project } from '../../../../../modules/api/apis/projectsDepartmentsApi';
import { Project, Department } from '../../../../../modules/api/apis/orgUnitApi';
import { Scope } from '../../../../../modules/api/apis/assetsApi';
import { HybridForm } from '../../../../../modules/ui/runaiV2/pages/HybridSubmissionForm';
import { NodeUtils } from '../../../../../modules/helpers/utils/NodeUtils';
import { LoginUi } from '../../../../../modules/ui/runaiV2/pages/LoginUiUtils';
import { TrainingWorkload, config, nodePoolUtils, nodeUtils } from '../../../../../modules/helpers';
import { logger } from '../../../../../modules/helpers/logger/logger';
import { kubeCli } from '../../../../../modules/cli';
import {
    NodepoolUpdateFields,
    SchedulingConfiguration,
    GpuPodsPlacementStrategy
} from '../../../../../modules/api/apis/clusterApi';
import { PlacementStrategy as PlacementStrategyClusterApi } from '../../../../../modules/api/apis/clusterApi';
import { Phase } from '../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import { WorkloadType } from '../../../../../modules/helpers/utils/WorkloadsCrudUtils';
import { DepartmentPageV2 } from '../../../../../modules/ui/runaiV2/pages/DepartmentPageV2';
import { ProjectPageV2 } from '../../../../../modules/ui/runaiV2/pages/ProjectPageV2';

test.describe(`Node Pool UI E2E Tests, tags: ${tags.tool.ui} ${tags.resource.workloads} ${tags.resource.nodepools} ${tags.teams.loong}, debugTags: ${DEBUG_TAGS.SYSTEM_LOGS}`, () => {
    supportedFrom(SupportedVersions.V2_18)(function () {
        let browserEngine: BrowserEngine;
        let authPage: AuthPage;
        let nodePoolPage: NodePoolPage;
        let departmentPage: DepartmentPageV2;
        let projectPage: ProjectPageV2;
        let workloadPage: WorkloadPage;
        let hybridForm: HybridForm;
        let gpuType: string;
        let gpuNodeCount: number;
        let projectName: string;
        let namespace: string;
        let project: V1Project;
        let createdNodePools: string[] = [];
        let currentWorkloadNames: string[] = [];
        let loginUi: LoginUi;
        let videoName = '';
        let createdDepartments: string[] = [];
        let createdProjects: string[] = [];

        const TEST_NODE_POOL_LABEL_KEY = 'nodepool';
        const TEST_NODE_POOL_LABEL_VALUE = 'nptest';
        const TEST_NODE_AFFINITY_LABEL_KEY = 'run.ai/node-type';

        async function loginAsUser(auth: AuthPage, loginUi: LoginUi, user: any): Promise<void> {
            await loginUi.logoutToLoginPage();
            const password = user.tempPassword ?? user.password;
            await auth.signIn(user.email, password);
        }

        test.beforeAll(async function () {
            gpuType = await NodeUtils.getGpuType();
            // Every test in this suite assumes a multi-GPU-node environment — placement
            // strategy (Spread/Binpack), node affinity, and priority-overflow tests are
            // either vacuous or impossible to satisfy with a single GPU node. Skip the
            // entire suite when the env can't host them.
            gpuNodeCount = (await nodeUtils.getGpuNodes(gpuType)).length;
            if (gpuNodeCount < 2) {
                logger.info(`Skipping suite — env has ${gpuNodeCount} GPU node(s); need at least 2`);
                test.skip();
                return;
            }
            browserEngine = new BrowserEngine();
            projectName = objectGenerator.generateShortNameByPrefix('project-nodepool');
            namespace = Utils.namespaceFromProjectName(projectName);
            project = await projectResourcesUtils.createProject(namespace, {
                name: projectName,
                deservedGpus: 3,
                maxAllowedGpus: 3
            });
            try {
                await orgUnitUtils.validateAndSetDefaultDepartmentGpu();
                logger.info('Successfully validated and set default department GPU quota');
            } catch (e) {
                logger.error(`Failed to validate and set default department GPU quota: ${e}`);
            }
        });

        test.beforeEach(async function () {
            await browserEngine.start();
            authPage = new AuthPage(browserEngine.currentPage);
            nodePoolPage = new NodePoolPage(browserEngine.currentPage);
            departmentPage = new DepartmentPageV2(browserEngine.currentPage);
            projectPage = new ProjectPageV2(browserEngine.currentPage);
            workloadPage = new WorkloadPage(browserEngine.currentPage);
            hybridForm = new HybridForm(browserEngine.currentPage);
            loginUi = new LoginUi(browserEngine.currentPage);
            await authPage.signIn(
                flowRootState.users.system_administrator.email,
                flowRootState.users.system_administrator.password
            );
        });

        test.afterEach(async ({}, testInfo) => {
            try {
                if (currentWorkloadNames.length > 0) {
                    const workloads = currentWorkloadNames.map(
                        (name) => new TrainingWorkload(config.quickStartImage.training, namespace, name)
                    );
                    try {
                        await operations.deleteWorkloads(workloads);
                        logger.info(
                            `Successfully cleaned up ${workloads.length} workloads using operations.deleteWorkloads`
                        );
                    } catch (deleteError) {
                        logger.warn(
                            `Failed to delete workloads using operations.deleteWorkloads: ${deleteError}`
                        );
                        logger.info('Attempting to delete workloads using UI as fallback...');

                        try {
                            await loginAsUser(authPage, loginUi, flowRootState.users.l1_researcher);
                            await workloadPage.navigateToWorkloads();

                            for (const workloadName of currentWorkloadNames) {
                                try {
                                    await workloadPage.deleteWorkload(workloadName);
                                    logger.info(`Successfully deleted workload ${workloadName} using UI`);
                                } catch (uiDeleteError) {
                                    logger.warn(
                                        `Failed to delete workload ${workloadName} using UI: ${uiDeleteError}`
                                    );
                                }
                            }
                        } catch (uiError) {
                            logger.error(`Failed to delete workloads using UI fallback: ${uiError}`);
                        }
                    }
                    currentWorkloadNames = [];
                }
                if (createdProjects.length > 0) {
                    for (const projectId of createdProjects) {
                        try {
                            await orgUnitUtils.deleteProjectAndNamespace(projectId, `project-${projectId}`);
                        } catch (e) {
                            logger.error(`Failed to delete project ${projectId}: ${e}`);
                        }
                    }
                    createdProjects = [];
                }
            } catch (e) {
                logger.error('Project cleanup failed:', e);
            }
            try {
                if (createdDepartments.length > 0) {
                    for (const departmentId of createdDepartments) {
                        try {
                            await orgUnitUtils.deleteDepartment(departmentId);
                        } catch (e) {
                            logger.error(`Failed to delete department ${departmentId}: ${e}`);
                        }
                    }
                    createdDepartments = [];
                }
            } catch (e) {
                logger.error('Department cleanup failed:', e);
            }
            try {
                const nodes = await kubeCli.getNodes('nvidia.com/gpu.present=true');
                for (const node of nodes.items) {
                    const nodeName = node.metadata?.name;
                    if (nodeName) {
                        try {
                            const nodeLabels = await kubeCli.getNodeLabels(nodeName);
                            for (const labelKey of [TEST_NODE_POOL_LABEL_KEY, TEST_NODE_AFFINITY_LABEL_KEY]) {
                                if (nodeLabels[labelKey]) {
                                    await kubeCli.unLabelNode(nodeName, labelKey);
                                    logger.info(
                                        `Successfully removed label ${labelKey} from node ${nodeName}`
                                    );
                                }
                            }
                        } catch (nodeError) {
                            logger.warn(`Failed to clean up labels for node ${nodeName}: ${nodeError}`);
                        }
                    }
                }
            } catch (e) {
                logger.error('Node label cleanup failed:', e);
            }
            try {
                if (createdNodePools.length > 0) {
                    await loginAsUser(authPage, loginUi, flowRootState.users.system_administrator);
                    for (const nodePoolName of createdNodePools) {
                        await nodePoolPage.deleteNodePoolAndWait(nodePoolName);
                    }
                    createdNodePools = [];
                }
            } catch (e) {
                logger.error('Node pool cleanup failed through UI:', e);
                if (createdNodePools.length > 0) {
                    logger.warn('Attempting to delete nodepools through API as fallback');
                    for (const nodePoolName of createdNodePools) {
                        try {
                            const nodePools = await nodePoolUtils.getAllBackendNodePoolsAndValidate();
                            const nodePool = nodePools.find((np) => np.name === nodePoolName);
                            if (nodePool) {
                                await nodePoolUtils.deleteNodePool(nodePool, false);
                            } else {
                                logger.error(`Could not find nodepool ${nodePoolName} for API deletion`);
                            }
                        } catch (apiError) {
                            logger.error(
                                `Failed to delete nodepool ${nodePoolName} through API: ${apiError}`
                            );
                        }
                    }
                }
                throw e;
            } finally {
                await browserEngine.stopAndArchive(testInfo, videoName);
            }
        });

        test.afterAll(async function () {
            if (project && project.id) {
                await operations.deleteProjectAndNamespace(project.id);
            }
        });

        async function createNodePoolWithStrategy(
            nodePoolName: string,
            labelKey: string,
            labelValue: string,
            strategy: PlacementStrategy.GpuEnum
        ): Promise<void> {
            const nodePoolRequest: NodePoolCreateRequest = {
                name: nodePoolName,
                labelKey,
                labelValue,
                placementStrategy: { gpu: strategy }
            };
            await nodePoolPage.navigateToNodePools();
            await nodePoolPage.createNodePoolWithGpuStrategy(nodePoolRequest);

            await nodePoolUtils.waitForNodePoolCrdProperty(
                nodePoolName,
                'spec.placementStrategy.gpu',
                (gpu) => gpu === strategy.toString(),
                60
            );
        }

        async function validatePlacementStrategy(
            nodePoolName: string,
            assertFn: (workloadNames: string[], namespace: string) => Promise<void>
        ): Promise<void> {
            workloadsCrudUtils.accessToken = flowRootState.users.l1_researcher.accessToken;
            const workloads: { [name: string]: string } = {};

            // Create workloads
            for (let i = 0; i < 3; i++) {
                logger.info(`Creating workload ${i + 1} of 3`);
                const { body: training } = await workloadsCrudUtils.createTraining(
                    objectGenerator.generateShortNameByPrefix('np-gpu-validate-job'),
                    true,
                    flowRootState.cluster.uuid,
                    project.id.toString(),
                    {
                        image: config.quickStartImage.ubuntu,
                        command: 'sleep 300',
                        nodePools: [nodePoolName],
                        compute: { gpuDevicesRequest: 1 }
                    }
                );
                workloads[training.name] = training.workloadId;
                currentWorkloadNames.push(training.name);
                await workloadsCrudUtils.waitForWorkloadPhase(
                    training.workloadId,
                    Phase.Running,
                    WorkloadType.TRAINING,
                    undefined,
                    120
                );
            }

            // Assert and cleanup
            await assertFn(Object.keys(workloads), namespace);
            await Promise.all(
                Object.entries(workloads).map(async ([name, id]) => {
                    await workloadsCrudUtils.deleteTraining(id);
                    const index = currentWorkloadNames.indexOf(name);
                    if (index > -1) {
                        currentWorkloadNames.splice(index, 1); // remove workload from currentWorkloadNames if it is deleted
                    }
                })
            );
        }

        async function createAndWaitWorkloads(
            hybridForm: HybridForm,
            workloadPage: WorkloadPage,
            count: number,
            nodePoolList: string[],
            projectName: string,
            clusterName: string,
            prefix: string,
            nodeAffinity?: string,
            envTemplate: string = 'jupyter-scipy',
            computeTemplate: string = 'two-gpus'
        ): Promise<string[]> {
            await workloadPage.navigateToWorkloads();
            const workloadNames: string[] = [];
            for (let i = 0; i < count; i++) {
                const workloadName = objectGenerator.generateShortNameByPrefix(`${prefix}${i}`);
                workloadNames.push(workloadName);
                await hybridForm.fillAndSubmitTrainingWorkloadForm(workloadName, projectName, clusterName, {
                    nodePoolList: nodePoolList,
                    nodeAffinity: nodeAffinity,
                    envTemplate: envTemplate,
                    computeTemplate: computeTemplate
                });
                logger.info(`Submitted ${i + 1} of ${count} workload ${workloadName}`);
            }
            for (const name of workloadNames) {
                await workloadPage.waitForWorkloadUIStatus(name, EWorkloadStatus.Running);
            }
            return workloadNames;
        }

        async function assertWorkloadsOnSameNode(workloadNames: string[], namespace: string): Promise<void> {
            const nodeNames: string[] = [];
            for (const name of workloadNames) {
                const nodeName = await workloadUtils.getNodeNameByJob(namespace, { name } as any);
                nodeNames.push(nodeName);
            }
            expect(new Set(nodeNames).size).toBe(1);
        }

        async function assertWorkloadsOnDifferentNodes(
            workloadNames: string[],
            namespace: string
        ): Promise<void> {
            const nodeNames: string[] = [];
            for (const name of workloadNames) {
                const nodeName = await workloadUtils.getNodeNameByJob(namespace, { name } as any);
                nodeNames.push(nodeName);
            }
            expect(new Set(nodeNames).size).toBeGreaterThan(1);
        }

        async function assertWorkloadsOnSpecificNode(
            workloadNames: string[],
            namespace: string,
            expectedNodeName: string
        ): Promise<void> {
            for (const name of workloadNames) {
                const nodeName = await workloadUtils.getNodeNameByJob(namespace, { name } as any);
                expect(nodeName).toBe(expectedNodeName);
            }
        }

        async function runNodePoolWorkloadTest({
            nodePoolPrefix,
            placementStrategy,
            workloadPrefix,
            assertFn
        }: {
            nodePoolPrefix: string;
            placementStrategy: PlacementStrategy.GpuEnum;
            workloadPrefix: string;
            assertFn: (workloadNames: string[], namespace: string) => Promise<void>;
        }): Promise<void> {
            const nodePoolName = objectGenerator.generateNameByPrefix(nodePoolPrefix);
            createdNodePools.push(nodePoolName);
            await createNodePoolWithStrategy(
                nodePoolName,
                nodePoolPage.nodePoolLabelKey,
                gpuType,
                placementStrategy
            );
            await loginAsUser(authPage, loginUi, flowRootState.users.l1_researcher);
            await workloadPage.navigateToWorkloads();
            currentWorkloadNames = await createAndWaitWorkloads(
                hybridForm,
                workloadPage,
                3,
                [nodePoolName],
                projectName,
                flowRootState.cluster.name,
                workloadPrefix
            );
            await assertFn(currentWorkloadNames, namespace);
        }

        test('T4059998: node pool with GPU Binpack placement strategy', async function () {
            test.setTimeout(300 * 1000); // this test may take longer than the default timeout
            videoName = 'nodepool-gpu-binpack';
            await runNodePoolWorkloadTest({
                nodePoolPrefix: 'np-binpack',
                placementStrategy: PlacementStrategy.GpuEnum.Binpack,
                workloadPrefix: 'np-gpu-binpack-job',
                assertFn: assertWorkloadsOnSameNode
            });
        });

        test('T4059999: node pool with GPU Spread placement strategy', async function () {
            test.setTimeout(300 * 1000); // this test may take longer than the default timeout
            videoName = 'nodepool-gpu-spread';
            await runNodePoolWorkloadTest({
                nodePoolPrefix: 'np-spread',
                placementStrategy: PlacementStrategy.GpuEnum.Spread,
                workloadPrefix: 'np-gpu-spread-job',
                assertFn: assertWorkloadsOnDifferentNodes
            });
        });

        async function fillNodePoolWithWorkloads(
            nodePoolName: string,
            labelValue: string,
            labelKey: string,
            workloadCount: number,
            randomNode: any
        ): Promise<void> {
            await createNodePoolWithStrategy(
                nodePoolName,
                labelKey,
                labelValue,
                PlacementStrategy.GpuEnum.Binpack
            );
            createdNodePools.push(nodePoolName);
            await loginAsUser(authPage, loginUi, flowRootState.users.l1_researcher);
            currentWorkloadNames = await createAndWaitWorkloads(
                hybridForm,
                workloadPage,
                workloadCount,
                [nodePoolName, 'default'],
                projectName,
                flowRootState.cluster.name,
                'npworkload',
                '',
                'jupyter-scipy',
                'two-gpus'
            );

            for (const name of currentWorkloadNames) {
                const nodeName = await workloadUtils.getNodeNameByJob(namespace, { name } as any);
                expect(nodeName).toBe(randomNode.metadata?.name);
            }
        }

        test('T4060001:Node pool prioity set when submitting the job --- The first pool has free resources', async function () {
            test.setTimeout(300 * 1000);
            videoName = 'T4060001-nodepool-free-resource';

            const selectedNode = await nodeUtils.getIdleGpuNode();
            const workloadCount = Math.floor(
                parseInt(selectedNode.status?.capacity?.['nvidia.com/gpu'] ?? '0') / 2
            );

            logger.info(
                `Selected idle GPU node: ${selectedNode.metadata?.name}, GPU count: ${workloadCount}`
            );

            await test.step('Label one node with custom label', async () => {
                logger.info(
                    `Labeling node ${selectedNode.metadata?.name} with ${TEST_NODE_POOL_LABEL_KEY}=${TEST_NODE_POOL_LABEL_VALUE}`
                );
                await kubeCli.labelNode(
                    selectedNode.metadata?.name,
                    TEST_NODE_POOL_LABEL_KEY,
                    TEST_NODE_POOL_LABEL_VALUE
                );
            });

            await test.step('Fill node pool with workloads', async () => {
                logger.info(`Creating ${workloadCount} workloads to fill the node pool`);
                await fillNodePoolWithWorkloads(
                    'nodepool',
                    TEST_NODE_POOL_LABEL_VALUE,
                    TEST_NODE_POOL_LABEL_KEY,
                    workloadCount,
                    selectedNode
                );
            });
        });

        test('T4060002:Node pool priority set when submitting the job --- The first pool has no free resources', async function () {
            test.setTimeout(300 * 1000);
            videoName = 'T4060002-nodepool-no-free-resource';
            const nodePoolName = objectGenerator.generateNameByPrefix('np-resource');

            // Get a random GPU node for testing
            const nodes = await kubeCli.getNodes('nvidia.com/gpu.present=true');
            const randomNode = nodes.items[Math.floor(Math.random() * nodes.items.length)];
            const workloadCount = Math.floor(
                parseInt(randomNode.status?.capacity?.['nvidia.com/gpu'] ?? '0') / 2
            );
            logger.info(`Selected GPU node: ${randomNode.metadata?.name}, GPU count: ${workloadCount}`);

            await test.step('Label one node with custom label', async () => {
                logger.info(
                    `Labeling node ${randomNode.metadata?.name} with ${TEST_NODE_POOL_LABEL_KEY}=${TEST_NODE_POOL_LABEL_VALUE}`
                );
                await kubeCli.labelNode(
                    randomNode.metadata?.name,
                    TEST_NODE_POOL_LABEL_KEY,
                    TEST_NODE_POOL_LABEL_VALUE
                );
            });

            await test.step('Fill node pool with workloads', async () => {
                logger.info(`Creating ${workloadCount} workloads to fill the node pool completely`);
                await fillNodePoolWithWorkloads(
                    nodePoolName,
                    TEST_NODE_POOL_LABEL_VALUE,
                    TEST_NODE_POOL_LABEL_KEY,
                    workloadCount,
                    randomNode
                );
            });

            await test.step('Submit new workload when first pool is full', async () => {
                const additionalWorkloadName =
                    objectGenerator.generateShortNameByPrefix('npworkload-additional');
                logger.info(
                    `Submitting additional workload: ${additionalWorkloadName} when first pool has no free resources`
                );
                await hybridForm.fillAndSubmitTrainingWorkloadForm(
                    additionalWorkloadName,
                    projectName,
                    flowRootState.cluster.name,
                    {
                        nodePoolList: [nodePoolName, 'default'],
                        envTemplate: 'jupyter-scipy',
                        computeTemplate: 'two-gpus'
                    }
                );
                await workloadPage.waitForWorkloadUIStatus(additionalWorkloadName, EWorkloadStatus.Running);
                currentWorkloadNames.push(additionalWorkloadName);
            });

            await test.step('Check that new workload runs on different node', async () => {
                const additionalWorkloadName = currentWorkloadNames[currentWorkloadNames.length - 1];
                logger.info(
                    `Verifying workload ${additionalWorkloadName} runs on different node than ${randomNode.metadata?.name}`
                );
                const additionalNodeName = await workloadUtils.getNodeNameByJob(namespace, {
                    name: additionalWorkloadName
                } as any);
                const validNodeNames = nodes.items.map((node) => node.metadata?.name);
                expect(validNodeNames).toContain(additionalNodeName);
                expect(additionalNodeName).not.toBe(randomNode.metadata?.name);
            });
        });

        test('T4060005: Node pool with node affinity test', async function () {
            test.setTimeout(300 * 1000);
            videoName = 'T4060005-nodepool-node-affinity';
            let workloadNames: string[] = [];
            const NODE_AFFINITY_LABEL_KEY = TEST_NODE_AFFINITY_LABEL_KEY;
            const NODE_AFFINITY_LABEL_VALUE = 'nodepooltest';
            const WORKLOAD_COUNT = 3;

            // Get GPU nodes for testing
            const nodes = await kubeCli.getNodes('nvidia.com/gpu.present=true');
            const nodePoolName = objectGenerator.generateNameByPrefix('np-affinity');
            const availableNodes = nodes.items.slice(0, 2); // Take first 2 nodes
            const [node1, node2] = availableNodes;

            logger.info(`Selected nodes for testing: ${node1.metadata?.name}, ${node2.metadata?.name}`);

            await test.step('Label 2 nodes with nodepool=nptest', async () => {
                logger.info(
                    `Labeling nodes ${node1.metadata?.name} and ${node2.metadata?.name} with ${TEST_NODE_POOL_LABEL_KEY}=${TEST_NODE_POOL_LABEL_VALUE}`
                );
                await kubeCli.labelNode(
                    node1.metadata?.name,
                    TEST_NODE_POOL_LABEL_KEY,
                    TEST_NODE_POOL_LABEL_VALUE
                );
                await kubeCli.labelNode(
                    node2.metadata?.name,
                    TEST_NODE_POOL_LABEL_KEY,
                    TEST_NODE_POOL_LABEL_VALUE
                );
            });

            await test.step('Add node affinity label for one of the 2 nodes', async () => {
                logger.info(
                    `Adding affinity label ${NODE_AFFINITY_LABEL_KEY}=${NODE_AFFINITY_LABEL_VALUE} to node ${node2.metadata?.name}`
                );
                await kubeCli.labelNode(
                    node2.metadata?.name,
                    NODE_AFFINITY_LABEL_KEY,
                    NODE_AFFINITY_LABEL_VALUE
                );
            });

            await test.step('Create a node pool with Admin user from UI', async () => {
                logger.info('Creating node pool with Admin user');
                await createNodePoolWithStrategy(
                    nodePoolName,
                    TEST_NODE_POOL_LABEL_KEY,
                    TEST_NODE_POOL_LABEL_VALUE,
                    PlacementStrategy.GpuEnum.Spread
                );
                createdNodePools.push(nodePoolName);
            });
            await test.step('Login UI with tenant user and create workloads with node affinity', async () => {
                logger.info('Logging in as tenant user and creating workloads with node affinity');
                await loginAsUser(authPage, loginUi, flowRootState.users.l1_researcher);
                await workloadPage.navigateToWorkloads();

                workloadNames = await createAndWaitWorkloads(
                    hybridForm,
                    workloadPage,
                    WORKLOAD_COUNT,
                    [nodePoolName],
                    projectName,
                    flowRootState.cluster.name,
                    'np-affinity-workload',
                    NODE_AFFINITY_LABEL_VALUE
                );
            });

            await test.step('Check that all workloads are running on the node with affinity', async () => {
                logger.info('Verifying all workloads run on the node with affinity label');
                for (const workloadName of workloadNames) {
                    const nodeName = await workloadUtils.getNodeNameByJob(namespace, {
                        name: workloadName
                    } as any);
                    expect(nodeName).toBe(node2.metadata?.name);
                    logger.info(`Workload ${workloadName} is running on expected node: ${nodeName}`);
                }
            });
        });

        test('T4060000: Department node pool priority', async function () {
            test.setTimeout(300 * 1000);
            videoName = 'nodepool-department-priority';

            let workerNodeName: string;
            let nodePoolName: string;
            let createdDepartment: Department;
            let priorityProjectName: string;
            let priorityNamespace: string;
            const customLabelKey = 'nodepool1';
            const customLabelValue = 'nptest1';

            await test.step('Setup: Label worker node for node pool affinity', async () => {
                const nodes = await nodeUtils.getAllNodes();
                workerNodeName = nodes.body.nodes.find((node) => node.name.includes('worker'))?.name;
                expect(workerNodeName).not.toBeUndefined();

                await kubeCli.labelNode(workerNodeName, customLabelKey, customLabelValue);

                const nodeLabels = await kubeCli.getNodeLabels(workerNodeName);
                expect(nodeLabels[customLabelKey]).toBe(customLabelValue);
            });

            try {
                await test.step('Create node pool with custom node affinity', async () => {
                    nodePoolName = 'np-priority';
                    nodePoolName = objectGenerator.generateNameByPrefix(nodePoolName);
                    createdNodePools.push(nodePoolName);

                    await nodePoolPage.navigateToNodePools();
                    const nodePoolRequest: NodePoolCreateRequest = {
                        name: nodePoolName,
                        labelKey: customLabelKey,
                        labelValue: customLabelValue
                    };
                    await nodePoolPage.createBasicNodePool(nodePoolRequest);
                    expect(await nodePoolPage.isTableIncludeNodePool(nodePoolName)).toBe(true);

                    const createdNodePool = (await waitUtils.waitForObject(
                        async () => {
                            const allNodePools = await nodePoolUtils.getAllBackendNodePoolsAndValidate();
                            return allNodePools.find((np) => np.name === nodePoolName);
                        },
                        `wait for nodePool ${nodePoolName} to be created`,
                        30
                    )) as NodePool;

                    expect(createdNodePool).toBeDefined();
                    expect(createdNodePool.id).toBeDefined();

                    expect(createdNodePool.labelKey).toBe(customLabelKey);
                    expect(createdNodePool.labelValue).toBe(customLabelValue);
                });

                await test.step('Create department with node pool priority configuration', async () => {
                    const departmentName = objectGenerator.generateNameByPrefix('dept-priority');

                    await departmentPage.goToPage();
                    await departmentPage.createDepartment(departmentName, undefined, [{ [nodePoolName]: 3 }]);

                    createdDepartment = (await waitUtils.waitForObject(
                        async () => {
                            const allDepartments = await orgUnitUtils.getDepartments();
                            return allDepartments.find((dept) => dept.name === departmentName);
                        },
                        `wait for department ${departmentName} to be created`,
                        30
                    )) as Department;

                    expect(createdDepartment).toBeDefined();
                    expect(createdDepartment.id).toBeDefined();
                    createdDepartments.push(createdDepartment.id);
                });

                await test.step('Create project under department with node pool at top priority', async () => {
                    priorityProjectName = objectGenerator.generateNameByPrefix('proj-priority');

                    await projectPage.goToPage();
                    await projectPage.createProject(
                        priorityProjectName,
                        undefined,
                        [{ [nodePoolName]: 3 }],
                        undefined,
                        undefined,
                        undefined,
                        undefined,
                        {
                            scope: Scope.Department,
                            scopeEntityName: createdDepartment.name
                        }
                    );

                    const createdProject = (await waitUtils.waitForObject(
                        async () => {
                            const allProjects = await orgUnitUtils.getProjectsByClusterId();
                            return allProjects.find((proj) => proj.name === priorityProjectName);
                        },
                        `wait for project ${priorityProjectName} to be created`,
                        30
                    )) as Project;

                    expect(createdProject).toBeDefined();
                    expect(createdProject.id).toBeDefined();
                    createdProjects.push(createdProject.id);
                });

                await test.step('Submit workloads as L1 researcher', async () => {
                    await loginAsUser(authPage, loginUi, flowRootState.users.l1_researcher);

                    currentWorkloadNames = await createAndWaitWorkloads(
                        hybridForm,
                        workloadPage,
                        3,
                        [nodePoolName],
                        priorityProjectName,
                        flowRootState.cluster.name,
                        'priority-np-job'
                    );
                });

                await test.step('Verify workloads are scheduled on the correct node', async () => {
                    priorityNamespace = Utils.namespaceFromProjectName(priorityProjectName);
                    await assertWorkloadsOnSpecificNode(
                        currentWorkloadNames,
                        priorityNamespace,
                        workerNodeName
                    );
                    logger.info(
                        `Successfully verified that all ${currentWorkloadNames.length} workloads are running on node ${workerNodeName}`
                    );
                });
            } finally {
                await test.step('Cleanup: Remove node labels', async () => {
                    await kubeCli.unLabelNode(workerNodeName, customLabelKey);
                });
            }
        });

        test('T5073840: Edit node pool GPU replacement strategy', async function () {
            test.setTimeout(300 * 1000);
            videoName = 'T4060006-nodepool-gpu-replacement-strategy';
            const nodePoolName = objectGenerator.generateNameByPrefix('np-gpu');
            let nodePoolId;

            await test.step('Create a node pool with GPU Spread and check that the workloads are running on different nodes', async () => {
                await createNodePoolWithStrategy(
                    nodePoolName,
                    nodePoolPage.nodePoolLabelKey,
                    gpuType,
                    PlacementStrategy.GpuEnum.Spread
                );
                createdNodePools.push(nodePoolName);

                const allNodePools = await nodePoolUtils.getAllNodePoolsByClusterAndValidate();
                const targetNodePool = allNodePools.find((np) => np.name === nodePoolName);
                expect(targetNodePool).toBeDefined();
                expect(targetNodePool.id).toBeDefined();
                nodePoolId = targetNodePool.id;

                await validatePlacementStrategy(nodePoolName, assertWorkloadsOnDifferentNodes);
            });
            await test.step('Edit the node pool GPU replacement strategy to to binpack', async () => {
                const updateFields = new NodepoolUpdateFields();
                const gpuPodsPlacementStrategy = new GpuPodsPlacementStrategy();
                gpuPodsPlacementStrategy.node = PlacementStrategyClusterApi.Binpack;
                const schedulingConfiguration = new SchedulingConfiguration();
                schedulingConfiguration.gpuPodsPlacementStrategy = gpuPodsPlacementStrategy;
                updateFields.schedulingConfiguration = schedulingConfiguration;
                updateFields.labelKey = nodePoolPage.nodePoolLabelKey;
                updateFields.labelValue = gpuType;

                await nodePoolUtils.updateClusterNodePool(nodePoolId, updateFields);
            });

            await test.step('Check that the workloads are running on the same node', async () => {
                await validatePlacementStrategy(nodePoolName, assertWorkloadsOnSameNode);
            });

            await test.step('Edit the node pool GPU replacement strategy to to spread', async () => {
                const updateFields = new NodepoolUpdateFields();
                const gpuPodsPlacementStrategy = new GpuPodsPlacementStrategy();
                gpuPodsPlacementStrategy.node = PlacementStrategyClusterApi.Spread;
                const schedulingConfiguration = new SchedulingConfiguration();
                schedulingConfiguration.gpuPodsPlacementStrategy = gpuPodsPlacementStrategy;
                updateFields.schedulingConfiguration = schedulingConfiguration;
                updateFields.labelKey = nodePoolPage.nodePoolLabelKey;
                updateFields.labelValue = gpuType;

                await nodePoolUtils.updateClusterNodePool(nodePoolId, updateFields);
            });

            await test.step('Check that the workloads are running on different nodes', async () => {
                await validatePlacementStrategy(nodePoolName, assertWorkloadsOnDifferentNodes);
            });
        });
    });
});
