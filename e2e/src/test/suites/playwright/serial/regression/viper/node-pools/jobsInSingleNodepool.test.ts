/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    BaseWorkload,
    CliError,
    config,
    flowRootState,
    InferenceWorkload,
    InteractiveWorkload,
    logger,
    nodePoolUtils,
    objectGenerator,
    operations,
    tags,
    TrainingWorkload,
    Utils,
    waitUtils,
    workloadUtils,
    inferenceClientCreator,
    versionUtils,
    workloadsUtilsV2,
    orgUnitUtils
} from '../../../../../../../modules/helpers';
import { backendClusterApi, clusterApi } from '../../../../../../../modules/api';
import { kubeCli } from '../../../../../../../modules/cli';
import { test, expect } from '@playwright/test';
import { NodePool } from '../../../../../../../modules/api/apis/backendClusterApi';
import { PlacementStrategy } from '../../../../../../../modules/api/apis/clusterApi';
import { submitFractionWorkloadsAndGetGpuIndices } from '../../../../../../../modules/helpers/models/workloads';
import {
    waitAndValidateDeploymentCreatedSuccessfullyInBackend,
    validateDeploymentCreatedSuccessfullyInCluster,
    waitAndValidateDeploymentInBackendStatusAvailable
} from '../../../../../../../modules/helpers/Validations';
import {
    shouldRunNativeWorkloadCrdTests,
    shouldRunTests,
    supportedFrom,
    SupportedVersions
} from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';

test.describe(
    `Submitting Jobs in Single Node Pool tests, tags:  ${tags.resource.nodepools} ${tags.tool.api} ${tags.operations.submit} ${tags.teams.viper}`,
    { tag: [tags.services.nodepoolController, tags.services.clusterService, tags.services.podGroupAssigner] },
    () => {
        const defaultNodePoolName = 'default';
        const imagePath = config.quickStartImage.ubuntu;
        const imageArguments = 'sleep infinity';

        let numOfGpus = 0;
        let numOfGpusRequired = 0;
        let numOfGpusRequiredWithOQ = 0;
        let numOfNodesWithGpus = 0;
        let numOfNodesWithGpusRequired = 0;
        let extraNodePoolCreated;

        let defaultNodePool;
        let nodePoolLabelKey;
        let nodePoolLabelValue;
        let nodePoolName;
        let nodePoolId;
        let nodeName;
        let nodePoolsWithResources;
        let project;
        let namespace;

        let createdWorkloads;

        let hasOnlyWorkloads: boolean;

        function prerequisites(numOfGpusRequiredForThisTest: number, shouldLog = true): boolean {
            return nodePoolUtils.prerequisitesForTestsWithGpus(
                numOfGpusRequiredForThisTest,
                numOfGpus,
                numOfNodesWithGpusRequired,
                numOfNodesWithGpus,
                shouldLog
            );
        }

        async function afterEachCleanup(): Promise<void> {
            let promisesToAwait = [];
            for (const workload of createdWorkloads) {
                promisesToAwait.push(workloadUtils.delete(workload));
            }
            try {
                await Promise.all(promisesToAwait);
            } catch {
                //continue
            }

            promisesToAwait = [];
            for (const workload of createdWorkloads) {
                promisesToAwait.push(
                    waitUtils.waitForPodsDeletion(workload.name, workload.testingSpecificFields.namespace)
                );
            }
            await Promise.all(promisesToAwait);
            createdWorkloads = [];

            await nodePoolUtils.cleanupDeleteClusterNodePool(extraNodePoolCreated);
            extraNodePoolCreated = undefined;
        }

        function _getWorkload(namespace: string, nodePoolName: string, type = 'Interactive'): BaseWorkload {
            const jobName = objectGenerator.generateNameByPrefix('job');
            let testJob;

            switch (type) {
                case 'Interactive':
                    testJob = new InteractiveWorkload(imagePath, namespace, jobName);
                    break;
                case 'Training':
                    testJob = new TrainingWorkload(imagePath, namespace, jobName);
                    break;
                case 'Inference':
                    testJob = new InferenceWorkload(namespace, jobName, {
                        image: config.quickStartImage.inferenceHelloWorld
                    });
                    break;
            }

            testJob.gpu = '1';
            if (type !== 'Inference') {
                testJob.command = imageArguments;
            }

            testJob.nodePools = [nodePoolName].join(' ');
            return testJob;
        }

        function getInteractiveWorkload(namespace: string, nodePoolName: string): InteractiveWorkload {
            return _getWorkload(namespace, nodePoolName, 'Interactive') as InteractiveWorkload;
        }

        function getTrainingWorkload(namespace: string, nodePoolName: string): TrainingWorkload {
            return _getWorkload(namespace, nodePoolName, 'Training') as TrainingWorkload;
        }

        test.beforeAll(async () => {
            clusterApi.nodepoolsApi.accessToken = flowRootState.users.system_administrator.accessToken;
            backendClusterApi.jobsApi.accessToken = flowRootState.users.system_administrator.accessToken;

            await nodePoolUtils.deleteAllClusterNodePools();

            await operations.reduceSchedulerResources();

            defaultNodePool = await nodePoolUtils.getDefaultNodePoolFromCluster();

            numOfGpus = await kubeCli.getNumOfGpus();
            numOfNodesWithGpus = await kubeCli.getNumOfNodesWithGpus();

            logger.debug(`numOfGpus: ${numOfGpus}`);
            logger.debug(`numOfNodesWithGpus: ${numOfNodesWithGpus}`);

            const generatedKeyValue = nodePoolUtils.generateKeyValueLabels();
            nodePoolLabelKey = generatedKeyValue.key;
            nodePoolLabelValue = generatedKeyValue.value;

            const createdNodePool = await nodePoolUtils.createClusterNodePoolAndLabelNodeToAssignToNodePool(
                nodePoolLabelKey,
                nodePoolLabelValue
            );
            nodePoolName = createdNodePool.nodePoolName;
            nodePoolId = createdNodePool.nodePoolId;
            nodeName = createdNodePool.nodeName;

            // departments are enabled by default so we need to set quota for the new nodepool
            const defaultDepartment = await orgUnitUtils.getDefaultDepartment();
            const defaultDepartmentResources = defaultDepartment.resources;

            const nodePoolResources = defaultDepartmentResources.find(
                (resource) => resource.nodePool.name === nodePoolName
            );
            if (nodePoolResources) {
                nodePoolResources.gpu.deserved = 10;
                nodePoolResources.gpu.overQuotaWeight = 10;
                await orgUnitUtils.patchDepartmentResources(defaultDepartment.id, [nodePoolResources]);
            }

            nodePoolsWithResources = [
                {
                    nodePoolName: nodePoolName,
                    nodePoolId: nodePoolId,
                    deservedGPUs: 1
                },
                {
                    nodePoolName: defaultNodePoolName,
                    nodePoolId: defaultNodePool.id,
                    deservedGPUs: 1
                }
            ];
            numOfGpusRequired = 2; // 1 for each deserved
            numOfGpusRequiredWithOQ = 4; // 1 for each deserved plus another two to test allow OQ
            numOfNodesWithGpusRequired = 2; // 1 for each node pool we are going to use

            project = await nodePoolUtils.createProjectWithSomeNodePoolResources(nodePoolsWithResources);
            namespace = Utils.namespaceFromProjectName(project.name);
            hasOnlyWorkloads = versionUtils.isVersionSupported(
                flowRootState.cluster.version,
                SupportedVersions.V2_23
            );
        });

        test.afterAll(async () => {
            try {
                if (project) {
                    await operations.deleteProjectAndNamespace(project.id);
                }

                if (nodePoolName) {
                    await nodePoolUtils.cleanupDeleteClusterNodePool({
                        nodePoolName: nodePoolName,
                        nodePoolId: nodePoolId
                    });
                    await kubeCli.unLabelNode(nodeName, nodePoolLabelKey);
                }
            } catch (error) {
                logger.error('Failed to cleanup node pool resources', error);
            }

            try {
                await nodePoolUtils.deleteAllClusterNodePools();
            } catch (error) {
                logger.error(`Failed to delete all nodepools in after all, error: ${error}`);
            }

            try {
                await operations.revertSchedulerResourcesToDefault();
            } catch (error) {
                logger.error(`Failed to revert scheduler resources to default, error: ${error}`);
            }
        });

        shouldRunNativeWorkloadCrdTests(() =>
        test.describe('Nodepools with some running jobs - submission fails', () => {
            test.beforeEach(() => {
                createdWorkloads = [];
            });

            test.afterEach(async () => {
                await afterEachCleanup();
            });

            test('Nodepool does not exist', async () => {
                    const testJob = getTrainingWorkload(
                        namespace,
                        objectGenerator.generateNameByPrefix('no-np')
                    );
                createdWorkloads.push(testJob);

                try {
                    await workloadUtils.submit(testJob);
                } catch (e) {
                    expect(e).toBeInstanceOf(CliError);
                }
            });

            test('Nodepool empty', async () => {
                const result = nodePoolUtils.generateKeyValueLabels();
                const createRes = await nodePoolUtils.createClusterNodePoolAndValidate(
                    result.key,
                    result.value
                );
                extraNodePoolCreated = createRes;
                await nodePoolUtils.waitForNodePoolCrdStatus(extraNodePoolCreated.nodePoolName, [
                    NodePool.StatusEnum.Empty
                ]);

                const testJob = getTrainingWorkload(namespace, extraNodePoolCreated.nodePoolName);
                createdWorkloads.push(testJob);

                // submit should fail cause the nodepool is empty
                try {
                    await workloadUtils.submit(testJob);
                } catch (e) {
                    expect(e).toBeInstanceOf(CliError);
                }
            });
            })
        );

        shouldRunNativeWorkloadCrdTests(() =>
        test.describe('Nodepools with some running jobs - required 4 GPUs', () => {
            shouldRunTests(prerequisites(numOfGpusRequiredWithOQ))(() => {
                test.beforeEach(() => {
                    createdWorkloads = [];
                });

                test.afterEach(async () => {
                    await afterEachCleanup();
                });

                test('Sanity submission and over-quota', async () => {
                    // submit interactive until quota of deserved GPUs - for each nodepool
                    let promisesToAwait = [];
                    for (const nodePool of nodePoolsWithResources) {
                        for (let i = 0; i < nodePool.deservedGPUs; i++) {
                            const testJob = getInteractiveWorkload(namespace, nodePool.nodePoolName);
                            createdWorkloads.push(testJob);
                            promisesToAwait.push(workloadUtils.submit(testJob));
                            promisesToAwait.push(
                                waitUtils.waitForPods(
                                    testJob.name,
                                    namespace,
                                    1,
                                    waitUtils.waitForPodTimeout,
                                    ['Running']
                                )
                            );
                        }
                    }
                    await Promise.all(promisesToAwait);

                    // submit another interactive job in each nodepool - should be Unschedulable
                    promisesToAwait = [];
                    for (const nodePool of nodePoolsWithResources) {
                        const jobNoQuota = getInteractiveWorkload(namespace, nodePool.nodePoolName);
                        createdWorkloads.push(jobNoQuota);
                        promisesToAwait.push(workloadUtils.submit(jobNoQuota));
                        logger.debug(
                            `waiting for unschedulable event on job ${jobNoQuota.name} in nodepool ${nodePool.nodePoolName}`
                        );
                        promisesToAwait.push(
                            waitUtils.waitForUnschedulable(
                                jobNoQuota.name,
                                project.name,
                                60,
                                `Failed to find unschedulable event for job ${jobNoQuota.name} in nodepool ${nodePool.nodePoolName}`
                            )
                        );
                    }
                    await Promise.all(promisesToAwait);

                    // submit training for each nodepool - should schedule, cause OQ allowed
                    promisesToAwait = [];
                    for (const nodePool of nodePoolsWithResources) {
                        const testJob = getTrainingWorkload(namespace, nodePool.nodePoolName);
                        createdWorkloads.push(testJob);
                        promisesToAwait.push(workloadUtils.submit(testJob));
                        promisesToAwait.push(
                                waitUtils.waitForPods(
                                    testJob.name,
                                    namespace,
                                    1,
                                    waitUtils.waitForPodTimeout,
                                    ['Running']
                                )
                        );
                    }
                    await Promise.all(promisesToAwait);
                });
            });
            })
        );

        shouldRunNativeWorkloadCrdTests(() =>
        test.describe('Nodepools with some running jobs - required 2 GPUs', () => {
            shouldRunTests(prerequisites(numOfGpusRequiredWithOQ))(() => {
                test.beforeEach(() => {
                    createdWorkloads = [];
                });

                test.afterEach(async () => {
                    await afterEachCleanup();
                });

                test('Validate NodePool field for backend job and pod objects - interactive', async () => {
                    // submit interactive and training jobs in each nodepool
                    let promisesToAwait = [];
                    for (const nodePool of nodePoolsWithResources) {
                        const testJob = getInteractiveWorkload(namespace, nodePool.nodePoolName);

                        createdWorkloads.push(testJob);
                        promisesToAwait.push(workloadUtils.submit(testJob));
                        promisesToAwait.push(
                                waitUtils.waitForPods(
                                    testJob.name,
                                    namespace,
                                    1,
                                    waitUtils.waitForPodTimeout,
                                    ['Running']
                                )
                        );
                    }
                    await Promise.all(promisesToAwait);

                    // validate job and pods objects from backend have correct nodepool value
                    promisesToAwait = [];
                    for (const workload of createdWorkloads) {
                        promisesToAwait.push(
                            nodePoolUtils.waitForBackendWorkloadAndPodsWithCorrectNodePool(
                                workload.name,
                                project.id,
                                workload.nodePools,
                                60
                            )
                        );
                    }
                    await Promise.all(promisesToAwait);
                });

                test('Validate NodePool field for backend job and pod objects -  training fraction', async () => {
                    // submit interactive and training jobs in each nodepool
                    let promisesToAwait = [];
                    for (const nodePool of nodePoolsWithResources) {
                        const testJob = getTrainingWorkload(namespace, nodePool.nodePoolName);
                        testJob.gpu = '0.5';

                        createdWorkloads.push(testJob);
                        promisesToAwait.push(workloadUtils.submit(testJob));
                        promisesToAwait.push(
                                waitUtils.waitForPods(
                                    testJob.name,
                                    namespace,
                                    1,
                                    waitUtils.waitForPodTimeout,
                                    ['Running']
                                )
                        );
                    }
                    await Promise.all(promisesToAwait);

                    // validate job and pods objects from backend have correct nodepool value
                    promisesToAwait = [];
                    for (const workload of createdWorkloads) {
                        promisesToAwait.push(
                            nodePoolUtils.waitForBackendWorkloadAndPodsWithCorrectNodePool(
                                workload.name,
                                project.id,
                                workload.nodePools,
                                60
                            )
                        );
                    }
                    await Promise.all(promisesToAwait);
                });
            });
            })
        );

        shouldRunNativeWorkloadCrdTests(() =>
        test.describe('Nodepools with some running jobs - required 1 GPUs', () => {
            const shouldRun =
                !prerequisites(numOfGpusRequired, false) ||
                nodePoolUtils.prerequisitesForTestsWithGpus(1, numOfGpus, 1, numOfNodesWithGpus);
            shouldRunTests(shouldRun)(() => {
                test.beforeEach(() => {
                    createdWorkloads = [];
                });

                test.afterEach(async () => {
                    await afterEachCleanup();
                });

                test('Sanity if only 1 GPU - interactive fraction', async () => {
                    // submit 1 interactive for the 1 nodepool we have "Ready"
                    const testJob = getInteractiveWorkload(namespace, nodePoolName);
                    testJob.gpu = '0.5';
                    createdWorkloads.push(testJob);
                    await workloadUtils.submit(testJob);
                    await waitUtils.waitForPods(testJob.name, namespace, 1, waitUtils.waitForPodTimeout, [
                        'Running'
                    ]);

                    if (hasOnlyWorkloads) {
                        const workload = await nodePoolUtils.waitForBackendWorkloadWithNodePool(
                            testJob.name,
                            project.id,
                            nodePoolName
                        );
                        expect(workload).toBeDefined();
                        const pods = await workloadsUtilsV2.getWorkloadPods(workload.id);
                        for (const pod of pods.body.pods) {
                            expect(pod.currentNodePool).toBe(nodePoolName);
                        }
                    } else {
                        const backendJob = await nodePoolUtils.waitForBackendJobWithNodePool(
                            testJob.name,
                            nodePoolName
                        );
                        if (!backendJob) {
                            throw new Error('Failed to get backend job');
                        }
                        const backendJobPods = (
                            await backendClusterApi.jobsApi.getPodsByJobId(
                                backendJob.podGroupId,
                                flowRootState.cluster.uuid
                            )
                        ).body;
                        for (const pod of backendJobPods) {
                            expect(pod.nodePool).toBe(nodePoolName);
                        }
                    }
                });

                test('Sanity if only 1 GPU - train', async () => {
                    // submit 1 train for the 1 nodepool we have "Ready"
                    const testJob = getTrainingWorkload(namespace, nodePoolName);
                    createdWorkloads.push(testJob);
                    await workloadUtils.submit(testJob);
                    await waitUtils.waitForPods(testJob.name, namespace, 1, waitUtils.waitForPodTimeout, [
                        'Running'
                    ]);

                    if (hasOnlyWorkloads) {
                        const workload = await nodePoolUtils.waitForBackendWorkloadWithNodePool(
                            testJob.name,
                            project.id,
                            nodePoolName
                        );
                        if (!workload) {
                            throw new Error('Failed to get workload');
                        }
                        const pods = await workloadsUtilsV2.getWorkloadPods(workload.id);
                        for (const pod of pods.body.pods) {
                            expect(pod.currentNodePool).toBe(nodePoolName);
                        }
                    } else {
                        const backendJob = await nodePoolUtils.waitForBackendJobWithNodePool(
                            testJob.name,
                            nodePoolName
                        );
                        if (!backendJob) {
                            throw new Error('Failed to get backend job');
                        }
                        const backendJobPods = (
                            await backendClusterApi.jobsApi.getPodsByJobId(
                                backendJob.podGroupId,
                                flowRootState.cluster.uuid
                            )
                        ).body;
                        for (const pod of backendJobPods) {
                            expect(pod.nodePool).toBe(nodePoolName);
                        }
                    }
                });
            });
            })
        );

        test.describe('Nodepools with deployment - required 1 GPUs', () => {
            let deploymentName: string;
            let replicas: number;
            let clusterId: string;

            shouldRunTests(nodePoolUtils.prerequisitesForTestsWithGpus(1, numOfGpus, 1, numOfNodesWithGpus))(
                () => {
                    test.beforeEach(() => {
                        replicas = 1;
                        clusterId = flowRootState.cluster.uuid;
                        deploymentName = objectGenerator.generateNameByPrefix('deployment');
                    });

                    test.afterEach(async () => {
                        try {
                            await operations.deleteDeployment(deploymentName, namespace, clusterId);
                        } catch (warn) {
                            logger.debug('deployment fails to be deleted', warn);
                        }
                    });

                    test.describe('NodePool + Apply old deployment via yaml', () => {
                        test('should apply old deployment', async () => {
                            const result = await kubeCli.createDeployment(
                                inferenceClientCreator.getDeployment(
                                    namespace,
                                    project.name,
                                    deploymentName,
                                    replicas,
                                    imagePath,
                                    nodePoolName
                                )
                            );
                            let deployment = result;

                            // validate deployment creation in cluster
                            validateDeploymentCreatedSuccessfullyInCluster(
                                deployment,
                                deploymentName,
                                namespace,
                                project.name,
                                replicas,
                                imagePath,
                                nodePoolName
                            );

                            // validate deployment creation in backend
                            await waitAndValidateDeploymentCreatedSuccessfullyInBackend(
                                clusterId,
                                deployment,
                                replicas
                            );

                            // wait for deployment status to be available in cluster and see it was updated in BE
                            await waitUtils.waitForDeploymentToBeAvailable(deploymentName, namespace);

                            deployment = await operations.getDeployment(namespace, deploymentName);

                            await waitAndValidateDeploymentInBackendStatusAvailable(clusterId, deployment);
                        });
                    });
                }
            );
        });

        // Device-level (GPU-level) placement strategy: with a single-node pool that has >= 2 GPUs,
        // two GPU fractions either pack onto the same physical GPU (binpack) or spread across
        // different GPUs of that node (spread), controlled by gpuPodsPlacementStrategy.device.
        shouldRunNativeWorkloadCrdTests(() =>
        test.describe('Nodepools with GPU device-level placement strategy', () => {
            // device-level strategy can only diverge from node-level on cluster >= 2.26 (older
            // clusters reject gpuPodsPlacementStrategy.device); also need >= 2 GPUs on a GPU node so
            // binpack ("same GPU") and spread ("different GPUs") are distinguishable
            supportedFrom(SupportedVersions.V2_26)(() => {
                shouldRunTests(
                    nodePoolUtils.prerequisitesForTestsWithGpus(2, numOfGpus, 1, numOfNodesWithGpus)
                )(() => {
                    test.beforeEach(() => {
                        createdWorkloads = [];
                    });

                    test.afterEach(async () => {
                        await afterEachCleanup();
                    });

                    test('Device-level Binpack - two GPU fractions land on the same GPU', async () => {
                        await nodePoolUtils.setGpuDevicePlacementStrategyAndWait(
                            nodePoolId,
                            nodePoolName,
                            nodePoolLabelKey,
                            nodePoolLabelValue,
                            PlacementStrategy.Binpack
                        );
                            const { workloads, gpuIndices } = await submitFractionWorkloadsAndGetGpuIndices(
                                project.name,
                                nodePoolName,
                                2,
                                0.5
                            );
                        createdWorkloads.push(...workloads);
                        expect(gpuIndices[0]).toBe(gpuIndices[1]);
                    });

                    test('Device-level Spread - two GPU fractions land on different GPUs of the same node', async () => {
                        await nodePoolUtils.setGpuDevicePlacementStrategyAndWait(
                            nodePoolId,
                            nodePoolName,
                            nodePoolLabelKey,
                            nodePoolLabelValue,
                            PlacementStrategy.Spread
                        );
                            const { workloads, gpuIndices } = await submitFractionWorkloadsAndGetGpuIndices(
                                project.name,
                                nodePoolName,
                                2,
                                0.5
                            );
                        createdWorkloads.push(...workloads);
                        expect(gpuIndices[0]).not.toBe(gpuIndices[1]);
                    });
                });
            });
            })
        );
    }
);
