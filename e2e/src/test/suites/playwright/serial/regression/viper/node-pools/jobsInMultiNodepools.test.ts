/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { backendClusterApi, clusterApi, orgUnitService } from '../../../../../../../modules/api';
import {
    apiUtils,
    config,
    flowRootState,
    inferenceClientCreator,
    logger,
    nodePoolUtils,
    objectGenerator,
    operations,
    projectUtils,
    tags,
    Utils,
    waitUtils,
    workloadsUtilsV2
} from '../../../../../../../modules/helpers';
import { kubeCli, runaiCliV2 } from '../../../../../../../modules/cli';
import { cliV2WorkloadType, SubmitFlagsMerged } from '../../../../../../../modules/cli/tools/RunaiCliV2.constants';
import { Project } from '../../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    validateDeploymentCreatedSuccessfullyInCluster,
    waitAndValidateDeploymentCreatedSuccessfullyInBackend,
    waitAndValidateDeploymentInBackendStatusAvailable
} from '../../../../../../../modules/helpers/Validations';
import { WorkloadType } from '../../../../../../../modules/helpers/utils/WorkloadsUtilsV2';
import { Department } from '../../../../../../../modules/api/apis/orgUnitApi';
const trainingImagePath = config.quickStartImage.training;
const interactiveImagePath = config.quickStartImage.interactive;
const inferenceImagePath = config.quickStartImage.inference;
const numberOfNodePoolsToCreateOfEachType = 1;
const LONGER_TEST_TIMEOUT_MS = 300000;

type workloadNameTypeAndProjectId = {
    workloadName: string;
    workloadType: WorkloadType;
    projectId: string | number;
};

test.describe(
    `Submitting Jobs in Multi Node Pools tests, tags:  ${tags.resource.nodepools} ${tags.operations.submit} ${tags.teams.viper}`,
    { tag: [tags.services.nodepoolController, tags.services.clusterService, tags.services.podGroupAssigner, tags.runOn.serialCI] },
    () => {
        let numOfGpus: number;
        let numOfNodesWithGpus = 0;

        let defaultNodePoolId: string;
        let projectCreated: Project;
        let department: Department;
        let namespace: string;
        let numOfGpusOnNodeInNodePoolWithQuota: number;
        let nodePoolsWithQuotaCreated = new Array(0);
        let nodePoolsNoQuotaCreated = new Array(0);
        let emptyNodePoolsCreated = new Array(0);
        const allTestsScopedNodePoolResources = new Array(0);
        let currentTestScopedNodePoolResources = new Array(0);
        let nodePoolResourcesForProjectCreation = new Array(0);
        let labeledNodes = new Array(0);
        let labeledNodeNames = new Array(0);
        let defaultNodePoolsForProject: string[] = [];

        let nodeTypeLabel: string;
        let anotherNodeInNodePool: string;
        let numOfGpusOnAnotherNodeInNodePool: number;
        let workloadsToDelete: workloadNameTypeAndProjectId[] = [];

        function prerequisites(
            numOfGpusRequiredForThisTest: number,
            numOfNodesWithGpusRequired: number,
            shouldLog = true
        ): boolean {
            return nodePoolUtils.prerequisitesForTestsWithGpus(
                numOfGpusRequiredForThisTest,
                numOfGpus,
                numOfNodesWithGpusRequired,
                numOfNodesWithGpus,
                shouldLog
            );
        }

        test.beforeAll(async () => {
            await apiUtils.disableCpuQuotaFeature();
            clusterApi.nodepoolsApi.accessToken = flowRootState.users.system_administrator.accessToken;
            backendClusterApi.jobsApi.accessToken = flowRootState.users.system_administrator.accessToken;

            await nodePoolUtils.deleteAllClusterNodePools();

            await operations.reduceSchedulerResources();

            const defaultNodePool = await nodePoolUtils.getDefaultNodePoolFromCluster();
            defaultNodePoolId = defaultNodePool.id;

            numOfGpus = await kubeCli.getNumOfGpus();
            numOfNodesWithGpus = await kubeCli.getNumOfNodesWithGpus();
            logger.debug(`numOfGpus: ${numOfGpus}`);
            logger.debug(`numOfNodesWithGpus: ${numOfNodesWithGpus}`);

            // create some empty node pools (don't label any node) with quota to work with
            for (let index = 0; index < 3; index++) {
                const nodePoolLabelKeyValue = nodePoolUtils.generateKeyValueLabels();
                const { nodePoolName, nodePoolId, errorStr } = await nodePoolUtils.createEmptyClusterNodePool(
                    nodePoolLabelKeyValue.key,
                    nodePoolLabelKeyValue.value
                );
                emptyNodePoolsCreated.push({
                    name: nodePoolName,
                    labelKey: nodePoolLabelKeyValue.key,
                    labelValue: nodePoolLabelKeyValue.value,
                    id: nodePoolId
                });
                allTestsScopedNodePoolResources.push({
                    nodePoolName: nodePoolName,
                    nodePoolId: nodePoolId,
                    deservedGPUs: 2
                });

                if (errorStr) {
                    throw new Error(`Failed creating empty node pool: ${errorStr}`);
                }
            }

            allTestsScopedNodePoolResources.push({
                nodePoolName: nodePoolUtils.DefaultNodePoolName,
                nodePoolId: defaultNodePoolId,
                deservedGPUs: 2
            });
        });

        test.afterAll(async () => {
            const promisesToAwait = [];
            for (const nodePool of emptyNodePoolsCreated) {
                promisesToAwait.push(
                    nodePoolUtils.cleanupDeleteClusterNodePool({
                        nodePoolName: nodePool.name,
                        nodePoolId: nodePool.id
                    })
                );
            }
            await Promise.all(promisesToAwait);
            emptyNodePoolsCreated = [];

            try {
                await nodePoolUtils.deleteAllClusterNodePools();
            } catch (error) {
                logger.error(`Failed to delete all nodepools in after all, error: ${error}`);
            }

            await operations.revertSchedulerResourcesToDefault();
        });

        async function beforeEach(): Promise<void> {
            nodePoolResourcesForProjectCreation = allTestsScopedNodePoolResources.concat(
                currentTestScopedNodePoolResources
            );
            department = await nodePoolUtils.createDepartmentWithAllNodePools(
                nodePoolResourcesForProjectCreation
            );
            projectCreated = await nodePoolUtils.createProjectWithSomeNodePoolResources(
                nodePoolResourcesForProjectCreation,
                { defaultNodePools: defaultNodePoolsForProject, departmentId: department.id }
            );
            namespace = Utils.namespaceFromProjectName(projectCreated.name);
            await projectUtils.waitForProjectCrdNonEmptyStatus(projectCreated.name, 'Ready');
        }

        async function afterEach(): Promise<void> {
            try {
                for (const workloadToDelete of workloadsToDelete) {
                    await workloadsUtilsV2.deleteWorkloadByNameAndProjectId(
                        workloadToDelete.workloadName,
                        workloadToDelete.projectId,
                        workloadToDelete.workloadType
                    );
                    await waitUtils.waitForWorkloadToBeDeleted(workloadToDelete.workloadName);
                }
            } catch (error) {
                logger.error(`Failed to delete workload in after each hook, error: ${error}`);
            }
            try {
                workloadsToDelete = [];
                await operations.deleteProjectAndNamespace(projectCreated.id);
            } catch (error) {
                logger.error(`Failed to cleanup in after each hook, error: ${error}`);
            }
            try {
                await orgUnitService.departmentApi.deleteDepartment(department.id);
            } catch (error) {
                logger.error(`Failed to cleanup in after each hook, error: ${error}`);
            }
        }

        test.describe('Submit jobs - 2 node pools, one with quota in project and one without', () => {
            test.beforeAll(async () => {
                if (!prerequisites(2, 2)) {
                    test.skip();
                }

                // create some node pools with nodes and quota to work with
                // we are not using a promises.all await because I need too much info in every iteration
                for (let index = 0; index < numberOfNodePoolsToCreateOfEachType; index++) {
                    const deservedGpusList = [0, 20];
                    for (const deservedGpus of deservedGpusList) {
                        const nodePoolLabelKeyValue = nodePoolUtils.generateKeyValueLabels();
                        const { nodePoolName, nodePoolId, nodeName } =
                            await nodePoolUtils.createClusterNodePoolAndLabelNodeToAssignToNodePool(
                                nodePoolLabelKeyValue.key,
                                nodePoolLabelKeyValue.value,
                                labeledNodeNames
                            );
                        currentTestScopedNodePoolResources.push({
                            nodePoolName: nodePoolName,
                            nodePoolId: nodePoolId,
                            deservedGPUs: deservedGpus
                        });
                        labeledNodes.push({
                            nodeName: nodeName,
                            labelKey: nodePoolLabelKeyValue.key
                        });
                        labeledNodeNames.push(nodeName);
                        const nodePoolCreated = {
                            name: nodePoolName,
                            labelKey: nodePoolLabelKeyValue.key,
                            labelValue: nodePoolLabelKeyValue.value,
                            id: nodePoolId
                        };

                        if (deservedGpus > 0) {
                            nodePoolsWithQuotaCreated.push(nodePoolCreated);
                            numOfGpusOnNodeInNodePoolWithQuota = await kubeCli.getNumOfGpusOnNode(nodeName);
                        } else {
                            nodePoolsNoQuotaCreated.push(nodePoolCreated);
                        }
                    }
                }

                defaultNodePoolsForProject = [
                    emptyNodePoolsCreated[2].name,
                    emptyNodePoolsCreated[0].name,
                    nodePoolsNoQuotaCreated[0].name,
                    nodePoolUtils.DefaultNodePoolName
                ];
            });

            test.afterAll(async () => {
                try {
                    defaultNodePoolsForProject = [];
                    let promisesToAwait = labeledNodes.map((labeledNode) =>
                        kubeCli.unLabelNode(labeledNode.nodeName, labeledNode.labelKey)
                    );
                    await Promise.all(promisesToAwait);
                    labeledNodes = [];
                    labeledNodeNames = [];

                    const allTestCreatedNodePools = nodePoolsWithQuotaCreated.concat(nodePoolsNoQuotaCreated);
                    promisesToAwait = [];
                    for (const nodePool of allTestCreatedNodePools) {
                        promisesToAwait.push(
                            nodePoolUtils.cleanupDeleteClusterNodePool({
                                nodePoolName: nodePool.name,
                                nodePoolId: nodePool.id
                            })
                        );
                    }
                    await Promise.all(promisesToAwait);
                    nodePoolsWithQuotaCreated = [];
                    nodePoolsNoQuotaCreated = [];
                    currentTestScopedNodePoolResources = [];
                } catch (error) {
                    logger.error(`Failed to cleanup in after hook, error: ${error}`);
                }
            });

            test.beforeEach(async () => {
                await beforeEach();
            });

            test.afterEach(async () => {
                await afterEach();
            });

            test('1: Submit train on multiple node pools by project default - schedule on the first non-empty one, debugTags: @debug-nodepool-controller-logs @debug-pods', async () => {
                logger.debug(
                    'Starting test: Submit train on multiple node pools - schedule on the first non-empty one'
                );
                const nodePoolsForJob = defaultNodePoolsForProject;
                const jobName = nodePoolUtils.submitJobWithMultiNodePools(
                    projectCreated.name,
                    trainingImagePath
                );
                workloadsToDelete.push({
                    workloadName: jobName,
                    projectId: projectCreated.id,
                    workloadType: WorkloadType.TRAINING
                });

                await nodePoolUtils.waitForUnschedulableOnNodePoolsInCorrectOrder(
                    jobName,
                    namespace,
                    nodePoolsForJob.slice(0, 2),
                    nodePoolsForJob.length
                );

                let pods = await waitUtils.waitForPods(jobName, namespace, 1, waitUtils.waitForPodTimeout, [
                    'Running'
                ]);
                expect(pods.length).toBe(1);
                const jobsFirstPod = pods[0];
                await waitUtils.waitForUnschedulableInNamespace(jobName, namespace, 60, '', false, false);

                await nodePoolUtils.waitForJobAndRunningPodsLabeledCorrectly(
                    jobName,
                    namespace,
                    nodePoolsNoQuotaCreated[0].name
                );

                await nodePoolUtils.consistentlyNodePoolReady(nodePoolsNoQuotaCreated[0].name);

                const podGroups = await waitUtils.waitForPodGroups(jobName, namespace);
                podGroups.forEach((podGroup) => {
                    expect(nodePoolUtils.getMarkUnschedulableOfPodGroup(podGroup)).toBe(false);
                    expect(nodePoolUtils.getSchedulingBackoffOfPodGroup(podGroup)).toBe(
                        nodePoolUtils.SingleSchedulingBackoff
                    );
                });

                // delete the pod
                await kubeCli.deletePod(jobsFirstPod.metadata.name, namespace);
                await waitUtils.waitForPodToBeDeleted(jobsFirstPod.metadata.name, namespace);

                // expect the new pod created to be scheduled
                pods = await waitUtils.waitForPods(jobName, namespace, 1, waitUtils.waitForPodTimeout, [
                    'Running'
                ]);
                // just to validate a new pod was created
                pods.forEach((pod) => {
                    expect(pod.metadata.name).not.toBe(jobsFirstPod.metadata.name);

                    const nodePoolName = nodePoolUtils.getObjectsNodePool(pod);
                    expect(nodePoolName).toBe(nodePoolsNoQuotaCreated[0].name);
                });

                await nodePoolUtils.waitForBackendWorkloadAndPodsWithCorrectNodePool(
                    jobName,
                    projectCreated.id,
                    nodePoolsNoQuotaCreated[0].name,
                    60
                );
            });

            test('2: Sanity - single node pool - interactive override project default', async () => {
                logger.debug('Starting test: Sanity - single node pool - interactive');
                const jobName = nodePoolUtils.submitJobWithMultiNodePools(
                    projectCreated.name,
                    interactiveImagePath,
                    [nodePoolsWithQuotaCreated[0].name],
                    cliV2WorkloadType.workspace
                );
                workloadsToDelete.push({
                    workloadName: jobName,
                    projectId: projectCreated.id,
                    workloadType: WorkloadType.WORKSPACE
                });

                await waitUtils.waitForPods(jobName, namespace, 1, waitUtils.waitForPodTimeout, ['Running']);
                await nodePoolUtils.waitForJobAndRunningPodsLabeledCorrectly(
                    jobName,
                    namespace,
                    nodePoolsWithQuotaCreated[0].name
                );

                const podGroups = await waitUtils.waitForPodGroups(jobName, namespace);
                podGroups.forEach((podGroup) => {
                    expect(nodePoolUtils.getMarkUnschedulableOfPodGroup(podGroup)).toBe(true);
                    expect(nodePoolUtils.getSchedulingBackoffOfPodGroup(podGroup)).toBe(
                        nodePoolUtils.InfiniteSchedulingBackoff
                    );
                });
            });

            test('3: Submit interactive on multiple node pools override project default - scheduled on the first from the list with quota', async () => {
                logger.debug(
                    'Starting test: Submit interactive on multiple node pools - scheduled on the first from the list with quota'
                );
                const nodePoolsForJob = [
                    nodePoolsNoQuotaCreated[0].name,
                    emptyNodePoolsCreated[0].name,
                    emptyNodePoolsCreated[1].name,
                    nodePoolsWithQuotaCreated[0].name
                ];
                const jobName = nodePoolUtils.submitJobWithMultiNodePools(
                    projectCreated.name,
                    interactiveImagePath,
                    nodePoolsForJob,
                    cliV2WorkloadType.workspace
                );
                workloadsToDelete.push({
                    workloadName: jobName,
                    projectId: projectCreated.id,
                    workloadType: WorkloadType.WORKSPACE
                });

                await nodePoolUtils.waitForUnschedulableOnNodePoolsInCorrectOrder(
                    jobName,
                    namespace,
                    nodePoolsForJob.slice(0, 2),
                    nodePoolsForJob.length
                );

                await waitUtils.waitForPods(jobName, namespace, 1, waitUtils.waitForPodTimeout, ['Running']);
                await waitUtils.waitForUnschedulableInNamespace(jobName, namespace, 60, '', false, false);

                await nodePoolUtils.waitForJobAndRunningPodsLabeledCorrectly(
                    jobName,
                    namespace,
                    nodePoolsWithQuotaCreated[0].name
                );

                const podGroups = await waitUtils.waitForPodGroups(jobName, namespace);
                podGroups.forEach((podGroup) => {
                    expect(nodePoolUtils.getMarkUnschedulableOfPodGroup(podGroup)).toBe(true);
                    expect(nodePoolUtils.getSchedulingBackoffOfPodGroup(podGroup)).toBe(
                        nodePoolUtils.SingleSchedulingBackoff
                    );
                });
            });

            test('4: Fill node with train jobs in over-quota, then submit interactive - train moves to next node pool - override project default', async () => {
                test.setTimeout(LONGER_TEST_TIMEOUT_MS);

                logger.debug(
                    'Starting test: Fill node with train jobs in over-quota, then submit interactive - train moves to next node pool'
                );
                const nodePoolsForTrainJobs = [
                    nodePoolsWithQuotaCreated[0].name,
                    nodePoolsNoQuotaCreated[0].name
                ];

                // submit train jobs to fill the node in the first node pool
                const submittedJobsNames = new Array(0);
                let gpusToSubmitJobsWith;
                if (numOfGpusOnNodeInNodePoolWithQuota > 1) {
                    gpusToSubmitJobsWith = [String(numOfGpusOnNodeInNodePoolWithQuota - 1), '1'];
                } else {
                    gpusToSubmitJobsWith = ['1'];
                }

                for (const gpusToSubmitJobWith of gpusToSubmitJobsWith) {
                    const jobName = nodePoolUtils.submitJobWithMultiNodePools(
                        projectCreated.name,
                        trainingImagePath,
                        nodePoolsForTrainJobs,
                        cliV2WorkloadType.training,
                        gpusToSubmitJobWith
                    );
                    submittedJobsNames.push(jobName);
                    workloadsToDelete.push({
                        workloadName: jobName,
                        projectId: projectCreated.id,
                        workloadType: WorkloadType.TRAINING
                    });

                    // wait for the job to be running - because the test relies on the order of the jobs.
                    // if they are submitted too close to each other,
                    // then sometimes the scheduler picks them up in "reverse" order - which makes sense, if the diff in submission time is very small -
                    // and then the preempted train is not the first one...
                    await nodePoolUtils.waitForJobsToRunInNodePool(
                        [jobName],
                        namespace,
                        nodePoolsForTrainJobs[0]
                    );
                }

                // now submit an interactive
                const interactiveJobName = nodePoolUtils.submitJobWithMultiNodePools(
                    projectCreated.name,
                    interactiveImagePath,
                    [nodePoolsWithQuotaCreated[0].name, nodePoolUtils.DefaultNodePoolName],
                    cliV2WorkloadType.workspace
                );
                workloadsToDelete.push({
                    workloadName: interactiveJobName,
                    projectId: projectCreated.id,
                    workloadType: WorkloadType.WORKSPACE
                });

                await waitUtils.waitForPods(interactiveJobName, namespace, 1, waitUtils.waitForPodTimeout, [
                    'Running'
                ]);
                // expect it to run on the same node of node pool with quota
                await nodePoolUtils.waitForJobAndRunningPodsLabeledCorrectly(
                    interactiveJobName,
                    namespace,
                    nodePoolsWithQuotaCreated[0].name
                );

                // the last train job was supposed to be evicted and moved to next node pool
                const lastTrainJobName = submittedJobsNames[submittedJobsNames.length - 1];
                await waitUtils.waitForPods(lastTrainJobName, namespace, 1, waitUtils.waitForPodTimeout, [
                    'Running'
                ]);
                await nodePoolUtils.waitForJobAndRunningPodsLabeledCorrectly(
                    lastTrainJobName,
                    namespace,
                    nodePoolsForTrainJobs[1]
                );
            });
        });

        test.describe('Submit jobs - 2 node pools, both with quota in project', () => {
            test.beforeEach(async () => {
                await beforeEach();
            });

            test.afterEach(async () => {
                await afterEach();
            });

            // assuming only 1 node in each node pool
            let nodeInNodePoolNumOfGpus = {};

            test.beforeAll(async () => {
                if (!prerequisites(2, 2)) {
                    test.skip();
                }

                // create some node pools with nodes and quota to work with
                for (let index = 0; index < numberOfNodePoolsToCreateOfEachType * 2; index++) {
                    const nodePoolLabelKeyValue = nodePoolUtils.generateKeyValueLabels();
                    const { nodePoolName, nodePoolId, nodeName } =
                        await nodePoolUtils.createClusterNodePoolAndLabelNodeToAssignToNodePool(
                            nodePoolLabelKeyValue.key,
                            nodePoolLabelKeyValue.value,
                            labeledNodeNames
                        );
                    currentTestScopedNodePoolResources.push({
                        nodePoolName: nodePoolName,
                        nodePoolId: nodePoolId,
                        deservedGPUs: 20
                    });
                    labeledNodes.push({
                        nodeName: nodeName,
                        labelKey: nodePoolLabelKeyValue.key
                    });
                    labeledNodeNames.push(nodeName);
                    const nodePoolCreated = {
                        name: nodePoolName,
                        labelKey: nodePoolLabelKeyValue.key,
                        labelValue: nodePoolLabelKeyValue.value,
                        id: nodePoolId
                    };

                    nodeInNodePoolNumOfGpus[nodePoolName] = await kubeCli.getNumOfGpusOnNode(nodeName);
                    nodePoolsWithQuotaCreated.push(nodePoolCreated);
                    numOfGpusOnNodeInNodePoolWithQuota = await kubeCli.getNumOfGpusOnNode(nodeName);
                }
            });

            test.afterAll(async () => {
                nodeInNodePoolNumOfGpus = {};
                try {
                    let promisesToAwait = labeledNodes.map((labeledNode) =>
                        kubeCli.unLabelNode(labeledNode.nodeName, labeledNode.labelKey)
                    );
                    await Promise.all(promisesToAwait);
                    labeledNodes = [];
                    labeledNodeNames = [];

                    promisesToAwait = [];
                    for (const nodePool of nodePoolsWithQuotaCreated) {
                        promisesToAwait.push(
                            nodePoolUtils.cleanupDeleteClusterNodePool({
                                nodePoolName: nodePool.name,
                                nodePoolId: nodePool.id
                            })
                        );
                    }
                    await Promise.all(promisesToAwait);
                    nodePoolsWithQuotaCreated = [];
                    currentTestScopedNodePoolResources = [];
                    numOfGpusOnNodeInNodePoolWithQuota = 0;
                } catch (error) {
                    logger.error(`Failed to cleanup in after hook, error: ${error}`);
                }
            });

            test('5: Test round-robin - preempted train should move to next node pool which is the first one in the list', async () => {
                test.setTimeout(LONGER_TEST_TIMEOUT_MS);

                logger.debug(
                    'Starting test: Test round-robin - preempted train should move to next node pool which is the first one in the list'
                );
                const nodePoolsForTrainJobs = [
                    nodePoolsWithQuotaCreated[1].name,
                    nodePoolsWithQuotaCreated[0].name
                ];
                const nodePoolsForInteractiveJobs = [
                    nodePoolsWithQuotaCreated[0].name,
                    nodePoolsWithQuotaCreated[1].name
                ];

                // submit train jobs to fill the node in the first node pool
                let submittedJobsNames = new Array(0);
                let gpusToSubmitJobsWith;
                if (nodeInNodePoolNumOfGpus[nodePoolsForTrainJobs[0]] > 1) {
                    gpusToSubmitJobsWith = [
                        String(nodeInNodePoolNumOfGpus[nodePoolsForTrainJobs[0]] - 1),
                        '1'
                    ];
                } else {
                    gpusToSubmitJobsWith = ['1'];
                }

                for (const gpusToSubmitJobWith of gpusToSubmitJobsWith) {
                    const jobName = nodePoolUtils.submitJobWithMultiNodePools(
                        projectCreated.name,
                        trainingImagePath,
                        nodePoolsForTrainJobs,
                        cliV2WorkloadType.training,
                        gpusToSubmitJobWith
                    );
                    submittedJobsNames.push(jobName);
                    workloadsToDelete.push({
                        workloadName: jobName,
                        projectId: projectCreated.id,
                        workloadType: WorkloadType.TRAINING
                    });

                    // wait for the job to be running - because the test relies on the order of the jobs.
                    // if they are submitted too close to each other,
                    // then sometimes the scheduler picks them up in "reverse" order - which makes sense, if the diff in submission time is very small -
                    // and then the preempted train is not the first one...
                    await nodePoolUtils.waitForJobsToRunInNodePool(
                        [jobName],
                        namespace,
                        nodePoolsForTrainJobs[0]
                    );
                }
                const trainJobToDeleteLater = submittedJobsNames[submittedJobsNames.length - 1];

                // now submit train jobs to fill the node in the second node pool
                submittedJobsNames = [];
                if (nodeInNodePoolNumOfGpus[nodePoolsForTrainJobs[1]] > 1) {
                    gpusToSubmitJobsWith = [
                        String(nodeInNodePoolNumOfGpus[nodePoolsForTrainJobs[1]] - 1),
                        '1'
                    ];
                } else {
                    gpusToSubmitJobsWith = ['1'];
                }

                for (const gpusToSubmitJobWith of gpusToSubmitJobsWith) {
                    const jobName = nodePoolUtils.submitJobWithMultiNodePools(
                        projectCreated.name,
                        trainingImagePath,
                        nodePoolsForTrainJobs,
                        cliV2WorkloadType.training,
                        gpusToSubmitJobWith
                    );
                    submittedJobsNames.push(jobName);
                    workloadsToDelete.push({
                        workloadName: jobName,
                        projectId: projectCreated.id,
                        workloadType: WorkloadType.TRAINING
                    });

                    await nodePoolUtils.waitForJobsToRunInNodePool(
                        [jobName],
                        namespace,
                        nodePoolsForTrainJobs[1]
                    );
                }

                // now delete the train job that runs on 1 GPU in the first node pool
                await workloadsUtilsV2.deleteWorkloadByNameAndProjectId(
                    trainJobToDeleteLater,
                    projectCreated.id,
                    WorkloadType.TRAINING
                );
                workloadsToDelete = workloadsToDelete.filter(
                    (workloadToDelete) => workloadToDelete.workloadName !== trainJobToDeleteLater
                );
                await waitUtils.waitForPodsDeletion(trainJobToDeleteLater, namespace);

                // now submit an interactive
                const interactiveJobName = nodePoolUtils.submitJobWithMultiNodePools(
                    projectCreated.name,
                    interactiveImagePath,
                    nodePoolsForInteractiveJobs,
                    cliV2WorkloadType.workspace
                );
                workloadsToDelete.push({
                    workloadName: interactiveJobName,
                    projectId: projectCreated.id,
                    workloadType: WorkloadType.WORKSPACE
                });

                await waitUtils.waitForPods(interactiveJobName, namespace, 1, waitUtils.waitForPodTimeout, [
                    'Running'
                ]);
                // expect it to run on its first node pool - which is the second node pool of the train jobs
                await nodePoolUtils.waitForJobAndRunningPodsLabeledCorrectly(
                    interactiveJobName,
                    namespace,
                    nodePoolsForInteractiveJobs[0]
                );

                // the last train job was supposed to be evicted and moved to next node pool - which is its first node pool -
                // and we made sure there is 1 GPU available for it
                const lastTrainJobName = submittedJobsNames[submittedJobsNames.length - 1];
                await waitUtils.waitForPods(lastTrainJobName, namespace, 1, waitUtils.waitForPodTimeout, [
                    'Running'
                ]);
                await nodePoolUtils.waitForJobAndRunningPodsLabeledCorrectly(
                    lastTrainJobName,
                    namespace,
                    nodePoolsForTrainJobs[0]
                );

                // validate both node pools are consistently ready
                const promisesToAwait = [];
                for (const nodePoolName of nodePoolsForTrainJobs) {
                    promisesToAwait.push(nodePoolUtils.consistentlyNodePoolReady(nodePoolName, 20));
                }
                await Promise.all(promisesToAwait);
            });
        });

        test.describe('Submit jobs with multi node pools - no need for GPUs', () => {
            test.beforeEach(async () => {
                await beforeEach();
            });

            test.afterEach(async () => {
                await afterEach();
            });

            test('6: Sanity - all nodepools Unschedulable', async () => {
                test.setTimeout(LONGER_TEST_TIMEOUT_MS);

                logger.debug('Starting test: Sanity - all nodepools Unschedulable');
                const nodePoolsForJob = [
                    emptyNodePoolsCreated[0].name,
                    emptyNodePoolsCreated[2].name,
                    emptyNodePoolsCreated[1].name
                ];
                const jobName = nodePoolUtils.submitJobWithMultiNodePools(
                    projectCreated.name,
                    trainingImagePath,
                    nodePoolsForJob
                );
                workloadsToDelete.push({
                    workloadName: jobName,
                    projectId: projectCreated.id,
                    workloadType: WorkloadType.TRAINING
                });

                await waitUtils.waitForUnschedulableInNamespace(jobName, namespace);

                await nodePoolUtils.waitForUnschedulableOnNodePoolsInCorrectOrder(
                    jobName,
                    namespace,
                    nodePoolsForJob,
                    nodePoolsForJob.length
                );

                const podGroups = await waitUtils.waitForPodGroups(jobName, namespace);
                podGroups.forEach((podGroup) => {
                    expect(nodePoolUtils.getMarkUnschedulableOfPodGroup(podGroup)).toBe(true);
                    expect(nodePoolUtils.getSchedulingBackoffOfPodGroup(podGroup)).toBe(
                        nodePoolUtils.SingleSchedulingBackoff
                    );
                });

                // when the Job is Pending - node pool field should be empty
                await nodePoolUtils.waitForBackendWorkloadAndPodsWithCorrectNodePool(
                    jobName,
                    projectCreated.id,
                    undefined,
                    60,
                    false
                );
            });
        });

        test.describe('Submit jobs with multi node pools - a node pool with 2 nodes', () => {
            test.beforeEach(async () => {
                await beforeEach();
            });

            test.afterEach(async () => {
                await afterEach();
            });

            test.beforeAll(async () => {
                if (!prerequisites(2, 2)) {
                    test.skip();
                }

                // create a node pool with nodes and quota to work with
                const nodePoolLabelKeyValue = nodePoolUtils.generateKeyValueLabels();
                const { nodePoolName, nodePoolId, nodeName } =
                    await nodePoolUtils.createClusterNodePoolAndLabelNodeToAssignToNodePool(
                        nodePoolLabelKeyValue.key,
                        nodePoolLabelKeyValue.value,
                        labeledNodeNames
                    );
                currentTestScopedNodePoolResources.push({
                    nodePoolName: nodePoolName,
                    nodePoolId: nodePoolId,
                    deservedGPUs: 20
                });
                labeledNodes.push({
                    nodeName: nodeName,
                    labelKey: nodePoolLabelKeyValue.key
                });
                labeledNodeNames.push(nodeName);
                const nodePoolCreated = {
                    name: nodePoolName,
                    labelKey: nodePoolLabelKeyValue.key,
                    labelValue: nodePoolLabelKeyValue.value,
                    id: nodePoolId
                };
                nodePoolsWithQuotaCreated.push(nodePoolCreated);

                // label another node to be included in nodePool
                anotherNodeInNodePool = await nodePoolUtils.labelAGPUNode(
                    nodePoolLabelKeyValue.key,
                    nodePoolLabelKeyValue.value,
                    labeledNodeNames
                );
            });

            test.afterAll(async () => {
                try {
                    let promisesToAwait = labeledNodes.map((labeledNode) =>
                        kubeCli.unLabelNode(labeledNode.nodeName, labeledNode.labelKey)
                    );
                    await Promise.all(promisesToAwait);
                    labeledNodes = [];
                    labeledNodeNames = [];

                    promisesToAwait = [];
                    for (const nodePool of nodePoolsWithQuotaCreated) {
                        promisesToAwait.push(
                            nodePoolUtils.cleanupDeleteClusterNodePool({
                                nodePoolName: nodePool.name,
                                nodePoolId: nodePool.id
                            })
                        );
                    }
                    await Promise.all(promisesToAwait);
                    nodePoolsWithQuotaCreated = [];
                    currentTestScopedNodePoolResources = [];
                } catch (error) {
                    logger.error(`Failed to cleanup in after hook, error: ${error}`);
                }
            });

            test.describe('Submit jobs to multi-nodepools, in a project that also has node-affinity', () => {
                let projectCreatedWithAffinity;
                let namespaceOfProjectCreatedWithAffinity;

                test.beforeAll(async () => {
                    // label the same node with node-type
                    nodeTypeLabel = objectGenerator.generateNameByPrefix('nodetype');
                    const runaiNodeTypeLabelKey = await nodePoolUtils.labelNodeAsRunaiNodeType(
                        anotherNodeInNodePool,
                        nodeTypeLabel
                    );
                    labeledNodes.push({
                        nodeName: anotherNodeInNodePool,
                        labelKey: runaiNodeTypeLabelKey
                    });
                    labeledNodeNames.push(anotherNodeInNodePool);

                    numOfGpusOnAnotherNodeInNodePool =
                        await kubeCli.getNumOfGpusOnNode(anotherNodeInNodePool);
                });

                test.beforeEach(async () => {
                    nodePoolResourcesForProjectCreation = allTestsScopedNodePoolResources.concat(
                        currentTestScopedNodePoolResources
                    );
                    projectCreatedWithAffinity =
                        await nodePoolUtils.createProjectWithNodePoolResourcesAndTrainNodeAffinity(
                            nodePoolResourcesForProjectCreation,
                            [nodeTypeLabel]
                        );
                    namespaceOfProjectCreatedWithAffinity = Utils.namespaceFromProjectName(
                        projectCreatedWithAffinity.name
                    );
                });

                test.afterEach(async () => {
                    try {
                        const allNodeTypes = await nodePoolUtils.getAllNodeTypes();
                        const promisesToAwait = [];
                        for (const nodeType of allNodeTypes) {
                            promisesToAwait.push(nodePoolUtils.deleteNodeType(nodeType.id));
                        }
                        await Promise.all(promisesToAwait);

                        try {
                            for (const workloadToDelete of workloadsToDelete) {
                                if (workloadToDelete.projectId === projectCreatedWithAffinity.id) {
                                    await workloadsUtilsV2.deleteWorkloadByNameAndProjectId(
                                        workloadToDelete.workloadName,
                                        workloadToDelete.projectId,
                                        workloadToDelete.workloadType
                                    );
                                    await waitUtils.waitForWorkloadToBeDeleted(workloadToDelete.workloadName);
                                }
                            }
                            workloadsToDelete = workloadsToDelete.filter(
                                (workloadToDelete) =>
                                    workloadToDelete.projectId !== projectCreatedWithAffinity.id
                            );
                        } catch (error) {
                            logger.error(`Failed to delete workload in after each hook, error: ${error}`);
                        }

                        await operations.deleteProjectAndNamespace(projectCreatedWithAffinity.id);
                    } catch (error) {
                        logger.error(`Failed to cleanup in after each hook, error: ${error}`);
                    }
                });

                test('7: Submit job in project with node affinity and multi node pools', async () => {
                    logger.debug(
                        'Starting test: Submit job in project with node affinity and multi node pools'
                    );
                    const nodePoolsForJob = [
                        emptyNodePoolsCreated[0].name,
                        nodePoolsWithQuotaCreated[0].name,
                        emptyNodePoolsCreated[1].name
                    ];

                    // submit a train job
                    const jobName = nodePoolUtils.submitJobWithMultiNodePools(
                        projectCreatedWithAffinity.name,
                        trainingImagePath,
                        nodePoolsForJob,
                        cliV2WorkloadType.training,
                        String(numOfGpusOnAnotherNodeInNodePool)
                    );
                    workloadsToDelete.push({
                        workloadName: jobName,
                        projectId: projectCreatedWithAffinity.id,
                        workloadType: WorkloadType.TRAINING
                    });

                    // validate the job is running - in the second node pool
                    await waitUtils.waitForPods(
                        jobName,
                        namespaceOfProjectCreatedWithAffinity,
                        1,
                        waitUtils.waitForPodTimeout,
                        ['Running']
                    );
                    await nodePoolUtils.waitForJobAndRunningPodsLabeledCorrectly(
                        jobName,
                        namespaceOfProjectCreatedWithAffinity,
                        nodePoolsForJob[1]
                    );

                    // validate the submitted job is running on the node we labeled with node-affinity (runai/node-type)
                    const pods = await waitUtils.waitForPods(
                        jobName,
                        namespaceOfProjectCreatedWithAffinity,
                        1,
                        waitUtils.waitForPodTimeout,
                        ['Running']
                    );
                    // there should only be 1 pod
                    expect(pods.length).toBe(1);
                    const pod = pods[0];
                    expect(pod.spec.nodeName).toBe(anotherNodeInNodePool);

                    // to avoid cases where the pods are randomly scheduled on the same node anyway... validate the node affinity
                    const nodeSelectorTerms =
                        pod.spec.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution
                            .nodeSelectorTerms;
                    for (const nodeSelectorTerm of nodeSelectorTerms) {
                        const matchExpressions = nodeSelectorTerm.matchExpressions;
                        let nodeTypeAffinityFound = false;
                        for (const matchExpression of matchExpressions) {
                            if (
                                matchExpression.key === nodePoolUtils.RunaiNodeTypeLabelKey &&
                                matchExpression.operator === 'In' &&
                                matchExpression.values.length === 1 &&
                                matchExpression.values[0] === nodeTypeLabel
                            ) {
                                nodeTypeAffinityFound = true;
                                break;
                            }
                        }

                        expect(nodeTypeAffinityFound).toBe(true);
                    }

                    // submit another train job to the same node pool list
                    const anotherJobName = nodePoolUtils.submitJobWithMultiNodePools(
                        projectCreatedWithAffinity.name,
                        trainingImagePath,
                        nodePoolsForJob
                    );
                    workloadsToDelete.push({
                        workloadName: anotherJobName,
                        projectId: projectCreatedWithAffinity.id,
                        workloadType: WorkloadType.TRAINING
                    });

                    // it should be Unschedulable - as the other node we have is not labeled correctly!
                    await waitUtils.waitForUnschedulableInNamespace(
                        anotherJobName,
                        namespaceOfProjectCreatedWithAffinity
                    );

                    await nodePoolUtils.waitForUnschedulableOnNodePoolsInCorrectOrder(
                        anotherJobName,
                        namespaceOfProjectCreatedWithAffinity,
                        nodePoolsForJob,
                        nodePoolsForJob.length
                    );

                    // validate first job is still running
                    await waitUtils.waitForPods(
                        jobName,
                        namespaceOfProjectCreatedWithAffinity,
                        1,
                        waitUtils.waitForPodTimeout,
                        ['Running']
                    );

                    // submit another train job to the same node pool list - but for the other project that's without the node affinity!
                    const noAffinityJobName = nodePoolUtils.submitJobWithMultiNodePools(
                        projectCreated.name,
                        trainingImagePath,
                        nodePoolsForJob
                    );
                    workloadsToDelete.push({
                        workloadName: noAffinityJobName,
                        projectId: projectCreated.id,
                        workloadType: WorkloadType.TRAINING
                    });

                    // validate the job is running - in the same node pool - but on the other node
                    const noAffinityJobsPods = await waitUtils.waitForPods(
                        noAffinityJobName,
                        namespace,
                        1,
                        waitUtils.waitForPodTimeout,
                        ['Running']
                    );
                    expect(noAffinityJobsPods.length).toBe(1);
                    const noAffinityJobsPod = noAffinityJobsPods[0];
                    expect(noAffinityJobsPod.spec.nodeName).not.toBe(anotherNodeInNodePool);
                    await nodePoolUtils.waitForJobAndRunningPodsLabeledCorrectly(
                        noAffinityJobName,
                        namespace,
                        nodePoolsForJob[1]
                    );
                });
            });
        });

        test.describe('Submit jobs - only 1 node with GPUs required', () => {
            test.beforeEach(async () => {
                await beforeEach();
            });

            test.afterEach(async () => {
                await afterEach();
            });

            test.beforeAll(() => {
                if (!prerequisites(1, 1)) {
                    test.skip();
                }
            });

            test('8: Sanity - submit to only default node pool', async () => {
                logger.debug('Starting test: Sanity - submit to only default node pool');
                const jobName = nodePoolUtils.submitJobWithMultiNodePools(
                    projectCreated.name,
                    trainingImagePath,
                    [nodePoolUtils.DefaultNodePoolName]
                );
                workloadsToDelete.push({
                    workloadName: jobName,
                    projectId: projectCreated.id,
                    workloadType: WorkloadType.TRAINING
                });

                const pods = await waitUtils.waitForPods(jobName, namespace, 1, 240, ['Running']);
                expect(pods.length).toBe(1);
                const pod = pods[0];

                await nodePoolUtils.waitForJobAndRunningPodsLabeledCorrectly(
                    jobName,
                    namespace,
                    nodePoolUtils.DefaultNodePoolName
                );

                const podGroups = await waitUtils.waitForPodGroups(jobName, namespace);
                podGroups.forEach((podGroup) => {
                    expect(nodePoolUtils.getMarkUnschedulableOfPodGroup(podGroup)).toBe(true);
                    expect(nodePoolUtils.getSchedulingBackoffOfPodGroup(podGroup)).toBe(
                        nodePoolUtils.InfiniteSchedulingBackoff
                    );
                });

                const nodeSelectorTerms =
                    pod.spec.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution
                        .nodeSelectorTerms;
                expect(nodeSelectorTerms.length).toBe(1);
                const matchExpressions = nodeSelectorTerms[0].matchExpressions;
                expect(matchExpressions.length).toBe(1);
                const matchExpression = matchExpressions[0];
                expect(matchExpression.key).toBe('runai/node-pool');
                expect(matchExpression.operator).toBe('DoesNotExist');
            });

            test('9: Sanity - submit without node pools', async () => {
                logger.debug('Starting test: Sanity - submit without node pools');
                const jobName = objectGenerator.generateNameByPrefix('job');
                const submitFlag = {
                    [SubmitFlagsMerged.image]: trainingImagePath,
                    [SubmitFlagsMerged.project]: projectCreated.name,
                    [SubmitFlagsMerged.gpuRequest]: '1'
                } as Partial<typeof SubmitFlagsMerged>;
                runaiCliV2.submit(
                    submitFlag,
                    jobName,
                    cliV2WorkloadType.training
                );

                const pods = await waitUtils.waitForPods(jobName, namespace, 1, 240, ['Running']);
                expect(pods.length).toBe(1);

                await nodePoolUtils.waitForJobAndRunningPodsLabeledCorrectly(
                    jobName,
                    namespace,
                    nodePoolUtils.DefaultNodePoolName
                );

                const podGroups = await waitUtils.waitForPodGroups(jobName, namespace);
                podGroups.forEach((podGroup) => {
                    expect(nodePoolUtils.getMarkUnschedulableOfPodGroup(podGroup)).toBe(true);
                    expect(nodePoolUtils.getSchedulingBackoffOfPodGroup(podGroup)).toBe(
                        nodePoolUtils.InfiniteSchedulingBackoff
                    );
                });
            });

            test.describe('Submit deployment', () => {
                let deploymentName: string;
                let replicas: number;

                test.beforeEach(() => {
                    deploymentName = objectGenerator.generateNameByPrefix('deployment');
                    replicas = 1;
                });

                test.afterEach(async () => {
                    try {
                        await operations.deleteDeployment(
                            deploymentName,
                            namespace,
                            flowRootState.cluster.uuid
                        );
                    } catch (warn) {
                        logger.debug('deployment failed to be deleted', warn);
                    }
                });

                test('10: Submit deployment yaml to multi node pools', async () => {
                    logger.debug('Starting test: Submit deployment yaml to multi node pools');

                    const nodePoolsForDeployment = [
                        emptyNodePoolsCreated[0],
                        { name: nodePoolUtils.DefaultNodePoolName, labelKey: '', labelValue: '' },
                        emptyNodePoolsCreated[1]
                    ];

                    const result = await kubeCli.createDeployment(
                        inferenceClientCreator.getDeploymentWithMultiNodePools(
                            namespace,
                            projectCreated.name,
                            deploymentName,
                            replicas,
                            inferenceImagePath,
                            nodePoolsForDeployment
                        )
                    );
                    let deployment = result;

                    validateDeploymentCreatedSuccessfullyInCluster(
                        deployment,
                        deploymentName,
                        namespace,
                        projectCreated.name,
                        replicas,
                        inferenceImagePath
                    );
                    // the "UnschedulableOnNodePool" condition only lives while the deployment is still
                    // pending on the first (empty) node pool - the scheduler clears it once the pods are
                    // bound, so it has to be validated before waiting for the deployment to be available
                    await nodePoolUtils.waitForUnschedulableOnNodePoolsInCorrectOrder(
                        deploymentName,
                        namespace,
                        [nodePoolsForDeployment[0].name],
                        nodePoolsForDeployment.length
                    );
                    await waitUtils.waitForDeploymentToBeAvailable(deploymentName, namespace);
                    await waitAndValidateDeploymentCreatedSuccessfullyInBackend(
                        flowRootState.cluster.uuid,
                        deployment,
                        replicas
                    );
                    deployment = await operations.getDeployment(namespace, deploymentName);
                    await waitAndValidateDeploymentInBackendStatusAvailable(
                        flowRootState.cluster.uuid,
                        deployment
                    );
                    await waitUtils.waitForPods(
                        deploymentName,
                        namespace,
                        1,
                        waitUtils.waitForPodTimeout,
                        ['Running'],
                        true
                    );
                    await waitUtils.waitForUnschedulableInNamespace(
                        deploymentName,
                        namespace,
                        60,
                        '',
                        false,
                        false,
                        true
                    );

                    await nodePoolUtils.waitForJobAndRunningPodsLabeledCorrectly(
                        deploymentName,
                        namespace,
                        nodePoolsForDeployment[1].name
                    );

                    const podGroups = await waitUtils.waitForPodGroups(deploymentName, namespace);
                    podGroups.forEach((podGroup) => {
                        expect(nodePoolUtils.getMarkUnschedulableOfPodGroup(podGroup)).toBe(false);
                        expect(nodePoolUtils.getSchedulingBackoffOfPodGroup(podGroup)).toBe(
                            nodePoolUtils.SingleSchedulingBackoff
                        );
                    });
                });
            });
        });
    }
);

/*
********* Submit jobs - 2 node pools, each with 1 node, one node pool with quota in project and one without
*********   -> required: 2 nodes with GPUs
1. 1: Submit train on multiple node pools by project default - schedule on the first non-empty one
    a. define a project with default node pool list
    b. submit train without node pools - should schedule on the first available, which is not the first one
    c. validate pod is not "Unschedulable"
    d. validate job is running in node pool - pod group's and pod's labels are correct
    e. validate node pool is ready consistently for 1 minute
    f. validate pod group has correct spec - markUnschedulable and schedulingBackoff
    g. delete the pod - see the new pod created is scheduled on the same node pool + labeled correctly
    h. validate job and pods from backend have correct node pool value
2. 2: Sanity - single node pool - interactive override project default
    a. define a project with default node pool list
    b. submit interactive to a list that contains 1 node pool, ignore the project default list
    c. validate job is running in node pool - pod group's and pod's labels are correct
    d. validate pod group has correct spec - markUnschedulable and schedulingBackoff
3. 3: Submit interactive on multiple node pools override project default - scheduled on the first from the list with quota
    a. define a project with default node pool list
    b. submit interactive to multi node pools, ignore the project default list - should schedule on the first node pool that has quota, which is not the first one
    c. validate the "UnschedulableOnNodePool" scheduling conditions appeared on the Pod Group in correct order
    d. validate pod is not "Unschedulable"
    e. validate job is running in node pool - pod group's and pod's labels are correct
    f. validate pod group has correct spec - markUnschedulable and schedulingBackoff
4. 4: Fill node with train jobs in over-quota, then submit interactive - train moves to next node pool - override project default
    a. define a project with default node pool list
    b. submit 2 train jobs with multi node pools, ignore the project default list - should schedule on the node in node pool with quota
        1. first job with (numOfGpus-1) GPUs, second job with 1 GPU.
    c. submit an interactive with multi nodepools, ignore the project default list  - should trigger an eviction and schedule on the same node
    d. the second train (with 1 GPU) should move to the other node pool - validating it

********* Submit jobs - 2 node pools, each with 1 node, both with quota in project
*********   -> required: 2 nodes with GPUs
5. 5: Test round-robin - preempted train should move to next node pool which is the first one in the list
    a. submit train jobs to fill both nodes of both node pools
        1. 2 jobs with (numOfGpus-1) GPUs, 2 jobs with 1 GPU.
        2. this also tests "fill resources in one np - next job to schedule on next node pool"
    b. delete the train job that runs on 1 GPU in the first node pool - to make a GPU avaialble for later
    c. submit an interactive on multi node pools - but this time its first node pool in the list is train's second node pool
    d. this should trigger a preemption of the train job with 1 GPU on the second node pool
    e. now, round-robin: this train should now be scheduled on the first node pool!
    f. validate both node pools are consistently ready for 1 minute

********* Submit jobs with multi node pools - no need for GPUs
*********   -> required: no GPUs required
6. 6: Sanity - all nodepools Unschedulable
    a. submit train to a list of node pools, all of them are empty
    b. see the pod is "Unschedulable"
    c. validate the "UnschedulableOnNodePool" scheduling conditions appeared on the Pod Group in correct order
    d. validate pod group has correct spec - markUnschedulable and schedulingBackoff
    e. validate job and pods from backend have correct node pool value (empty value)

********* Submit jobs - a node pool with 2 nodes with GPUs
*********   -> required: 2 nodes with GPUs
7. 7: Submit job in project with node affinity and multi node pools
    a. (before hook) create a project with node affinity
        1. and label only 1 of the nodes in the node pool with this node-type ("runai/node-type")
    b. submit 1 train job to a list of node pools, should schedule on the one node pool we have
        1. submit with (numOfGpus) GPUs to fill the whole node
    c. validate it is scheduled on the node that we have labeled with the node-type
    d. submit another train job exactly the same - should be left pending as the second node doesn't have the node affinity!
    e. submit another train job to same node pools but to a project without node affinity -
        - should be scheduled on the same node pool, but on the other node!

********* Submit jobs - only 1 node with GPUs required
*********   -> required: 1 node with GPUs
8. 8: Sanity - submit to only default node pool
    a. submit a job with only 'default' node pool in the list
    b. see it's scheduled with correct labels and all
9. 9: Sanity - submit without node pools
    a. submit a job without node pools at all
    b. see it's scheduled to 'default' node pool with correct labels and all
10. 10: Submit deployment yaml to multi node pools
    a. submit a deployment with a list of 3 node pools - should schedule on the second, which is 'default' node pool
    b. validate UnschedulableOnNodePool on pod groups for the first node pool, while it's still pending
    c. validate deployment is "available" in both backend and cluster
    d. validate no "Unschedulable" event on pods
    e. validate pods and pod groups have correct node pool label
    f. validate pod group objects have correct spec - markUnschedulable and schedulingBackoff
11. 11: Submit mpi to multi node pools
    a. submit mpi to multi node pools - should schedule on the first available, which is the second one
    b. validate pod is not "Unschedulable"
    c. validate correct order of UnschedulableOnNodePool scheduling conditions on pod group
    d. validate job is running in node pool - pod group's and pod's labels are correct
    e. validate job and pods from backend have correct node pool value


LEFT TO DO:
* default np from project when submitted without node pools
 */
