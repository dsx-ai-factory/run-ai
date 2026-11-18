/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    config,
    crds,
    flowRootState,
    TrainingWorkload,
    logger,
    nodePoolUtils,
    objectGenerator,
    operations,
    projectResourcesUtils,
    projectUtils,
    tags,
    Utils,
    waitUtils,
    workloadsUtilsV2
} from '../../../../../../../modules/helpers';
import { kubeCli } from '../../../../../../../modules/cli';
import { test, expect, TestInfo } from '@playwright/test';
import { Phase } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import {
    shouldRunNativeWorkloadCrdTests,
    supportedFrom,
    SupportedVersions
} from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import {
    NodepoolCreateFields,
    NodepoolPhase,
    SchedulingConfiguration
} from '../../../../../../../modules/api/apis/clusterApi';
import { clusterApi } from '../../../../../../../modules/api';
import { Nodepool } from '../../../../../../../modules/api/apis/clusterApi';
import { StatusCodes } from 'http-status-codes';
import { get } from 'lodash-es';

test.describe(`workload api advanced parallel tests: ${tags.tool.api} ${tags.resource.workloads} ${tags.teams.viper}`, () => {
    supportedFrom(SupportedVersions.V2_23)(() => {
        // global setup's initCrds mutations don't survive the worker fork - re-apply per worker
        test.beforeAll(() => {
            crds.initCrds(flowRootState.cluster.version);
        });

        function createNodepoolRequest(
            minGuaranteedRuntime: string,
            nodePoolLabelKey: string,
            nodePoolLabelValue: string,
            clusterId: string
        ): {
            nodePoolCreateRequest: NodepoolCreateFields;
        } {
            const nodePoolName = objectGenerator.generateNameByPrefix('np');

            const nodePoolCreateRequest = new NodepoolCreateFields();
            nodePoolCreateRequest.name = nodePoolName;
            nodePoolCreateRequest.labelKey = nodePoolLabelKey;
            nodePoolCreateRequest.labelValue = nodePoolLabelValue;
            nodePoolCreateRequest.clusterId = clusterId;
            if (minGuaranteedRuntime) {
                // Create scheduling configuration
                const schedulingConfiguration = new SchedulingConfiguration();
                schedulingConfiguration.minGuaranteedRuntime = minGuaranteedRuntime;
                nodePoolCreateRequest.schedulingConfiguration = schedulingConfiguration;
            }
            return { nodePoolCreateRequest };
        }

        test.describe(`minGuaranteedRuntime tests ${tags.tool.api} ${tags.resource.workloads} ${tags.teams.viper}`, () => {
            const nodePoolLabelKey = 'test-label-key-1';
            const nodePoolLabelValue = 'test-label-value-1';
            test.describe(`valid minGuaranteedRuntime values ${tags.tool.api} ${tags.resource.workloads} ${tags.teams.viper}`, () => {
                const validTestCases = [
                    {
                        minGuaranteedRuntime: '1h30m45s',
                        description: 'valid duration with hours, minutes and seconds'
                    },
                    {
                        minGuaranteedRuntime: undefined,
                        description: 'not set value'
                    },
                    {
                        minGuaranteedRuntime: '0m',
                        description: 'valid format - zero duration'
                    }
                ];

                for (const testCase of validTestCases) {
                    test(`should create node pool with ${testCase.description}`, async () => {
                        const { nodePoolCreateRequest } = createNodepoolRequest(
                            testCase.minGuaranteedRuntime,
                            nodePoolLabelKey,
                            nodePoolLabelValue,
                            flowRootState.cluster.uuid
                        );

                        const createResponse =
                            await nodePoolUtils.createClusterNodePoolWithRequest(nodePoolCreateRequest);

                        // always clean up - a leftover nodepool holds the shared label and 409s the next case
                        try {
                            expect(createResponse.response.statusCode).toBe(StatusCodes.CREATED);
                            expect(createResponse.body).toBeTruthy();
                            expect(createResponse.body.id).toBeTruthy();
                            expect(createResponse.body.name).toBe(nodePoolCreateRequest.name);

                            const nodepoolResponse = createResponse.body;
                            // Wait for node pool to be ready
                            await nodePoolUtils.waitForClusterNodePoolPhase(
                                nodepoolResponse.id,
                                [NodepoolPhase.Ready, NodepoolPhase.Empty],
                                20
                            );

                            expect(nodepoolResponse.schedulingConfiguration).toBeTruthy();
                            expect(nodepoolResponse.schedulingConfiguration.minGuaranteedRuntime).toBe(
                                testCase.minGuaranteedRuntime
                            );

                            const minRuntimePath = nodePoolUtils.schedulingShardMinRuntimePath();
                            const schedulingShard = await nodePoolUtils.waitForSchedulingShardCrdArgs(
                                nodePoolCreateRequest.name,
                                `${minRuntimePath}.preemptMinRuntime`,
                                testCase.minGuaranteedRuntime
                            );
                            const schedulingShardCrd = schedulingShard as any;
                            expect(schedulingShardCrd).toBeTruthy();
                            expect(schedulingShardCrd['spec']).toBeTruthy();
                            // the min runtime section only exists once the node pool has a
                            // minGuaranteedRuntime, so the "not set" case has no section at all
                            expect(get(schedulingShardCrd, `${minRuntimePath}.preemptMinRuntime`)).toBe(
                                testCase.minGuaranteedRuntime
                            );
                            expect(get(schedulingShardCrd, `${minRuntimePath}.reclaimMinRuntime`)).toBe(
                                testCase.minGuaranteedRuntime
                            );

                            logger.info(
                                `Successfully created node pool ${nodePoolCreateRequest.name} with minGuaranteedRuntime: ${testCase.minGuaranteedRuntime}`
                            );
                        } finally {
                            await nodePoolUtils
                                .deleteNodePoolByIdWithWait(createResponse.body.id, createResponse.body.name)
                                .catch((e) =>
                                    logger.warn(
                                        `Cleanup: failed to delete nodepool ${nodePoolCreateRequest.name}: ${e.message}`
                                    )
                                );
                        }
                    });
                }
            });

            test.describe(`invalid minGuaranteedRuntime values ${tags.tool.api} ${tags.resource.workloads} ${tags.teams.viper}`, () => {
                const invalidTestCases = [
                    { minGuaranteedRuntime: '34', description: 'invalid format - just number' },
                    { minGuaranteedRuntime: 'abc', description: 'invalid format - letters only' },
                    { minGuaranteedRuntime: '-30m', description: 'invalid format - negative duration' }
                ];
                for (const testCase of invalidTestCases) {
                    test(`should reject node pool creation with ${testCase.description}`, async () => {
                        const { nodePoolCreateRequest } = createNodepoolRequest(
                            testCase.minGuaranteedRuntime,
                            nodePoolLabelKey,
                            nodePoolLabelValue,
                            flowRootState.cluster.uuid
                        );

                        let errorCaught = false;
                        let responseStatusCode = null;
                        let errorMessage = '';

                        try {
                            const createResponse =
                                await nodePoolUtils.createClusterNodePoolWithRequest(nodePoolCreateRequest);
                            expect(createResponse).toBe(undefined);
                            // If we get here, the request succeeded when it should have failed
                            // Clean up and fail the test
                            await clusterApi.nodepoolsApi.deleteNodepool(createResponse.body.id);
                        } catch (error) {
                            errorCaught = true;

                            logger.debug(
                                `Error caught for invalid minGuaranteedRuntime "${testCase.minGuaranteedRuntime}":`,
                                error
                            );

                            // Check if error has response property
                            if (error.response) {
                                responseStatusCode = error.response.statusCode;
                                logger.debug(`Response status code: ${responseStatusCode}`);

                                // Log response body if available
                                if (error.response.body) {
                                    logger.debug(`Response body: ${JSON.stringify(error.response.body)}`);
                                }
                            } else if (error.statusCode) {
                                responseStatusCode = error.statusCode;
                                logger.debug(`Error status code: ${responseStatusCode}`);
                            } else {
                                // Log the error structure for debugging
                                logger.debug(`Error structure: ${JSON.stringify(error)}`);
                            }

                            // Store error message
                            errorMessage = error.message || error.toString();
                            logger.debug(`Error message: ${errorMessage}`);
                        }

                        expect(errorCaught).toBe(true);

                        expect(responseStatusCode).toBe(StatusCodes.BAD_REQUEST);

                        logger.debug(
                            `Correctly rejected node pool creation with invalid minGuaranteedRuntime: ${testCase.minGuaranteedRuntime}`
                        );
                    });
                }
            });
        });

        shouldRunNativeWorkloadCrdTests(() =>
        test.describe(`guaranteedRuntimeEndsAt validation ${tags.tool.api} ${tags.resource.workloads} ${tags.teams.viper}`, () => {
            let nodePoolLabelKey: string, nodePoolLabelValue: string;
            const testCases = [
                {
                    minGuaranteedRuntime: '30m',
                    minGuaranteedRuntimeMilliseconds: 30 * 60 * 1000,
                    description: 'valid duration',
                    guaranteedRuntimeEndsAtValueExists: true,
                    isValid: true
                },
                {
                    minGuaranteedRuntime: undefined,
                    minGuaranteedRuntimeMilliseconds: 0,
                    description: 'undefined value',
                    guaranteedRuntimeEndsAtValueExists: false,
                    isValid: true
                }
            ];

            for (const testCase of testCases) {
                test.beforeAll(function ({}, testInfo: TestInfo) {
                    nodePoolLabelKey = objectGenerator.generateNameByPrefix('np-key');
                    nodePoolLabelValue = objectGenerator.generateNameByPrefix('np-value');
                });

                test(`should handle ${testCase.description}`, async () => {
                    const { nodePoolCreateRequest } = createNodepoolRequest(
                        testCase.minGuaranteedRuntime,
                        nodePoolLabelKey,
                        nodePoolLabelValue,
                        flowRootState.cluster.uuid
                    );

                    const nodePoolName = nodePoolCreateRequest.name;
                        logger.info(
                            `creating nodepool ${nodePoolName}, cluster ${flowRootState.cluster.uuid}`
                        );
                    const createResponse =
                        await nodePoolUtils.createClusterNodePoolWithRequest(nodePoolCreateRequest);

                    expect(createResponse.response.statusCode).toBe(StatusCodes.CREATED);
                    expect(createResponse.body).toBeTruthy();
                    expect(createResponse.body.id).toBeTruthy();
                    const createdNodepool = createResponse.body as Nodepool;
                    const nodePoolId = String(createdNodepool.id);
                    expect(nodePoolId).toBeTruthy();

                    // Wait for node pool to be ready
                    await nodePoolUtils.waitForClusterNodePoolPhase(
                        nodePoolId,
                        [NodepoolPhase.Ready, NodepoolPhase.Empty],
                        20
                    );
                    logger.info(`nodepool ${nodePoolName} is running`);
                    // Label a node to assign it to the node pool
                    const nodeName = await nodePoolUtils.labelAGPUNode(
                        nodePoolLabelKey,
                        nodePoolLabelValue,
                        []
                    );
                    expect(nodeName).toBeTruthy();
                        logger.info(
                            `Labeled node ${nodeName} with ${nodePoolLabelKey}=${nodePoolLabelValue}`
                        );

                    // Wait until the nodepool status reflects the labeled node
                    await nodePoolUtils.waitForClusterNodePoolProperty(
                        nodePoolId,
                        'status.nodes',
                        (nodes: string[]) => Array.isArray(nodes) && nodes.includes(nodeName)
                    );

                    // Wait for node pool to be ready with the node
                        await nodePoolUtils.waitForClusterNodePoolPhase(
                            nodePoolId,
                            [NodepoolPhase.Ready],
                            60
                        );

                    // Create a project that uses this node pool
                    const projectName = objectGenerator.generateNameByPrefix('p');
                    const projectNamespace = Utils.namespaceFromProjectName(projectName);

                    const project = await projectResourcesUtils.createProject(projectNamespace, {
                        name: projectName,
                        deservedGpus: 1,
                        maxAllowedGpus: 1,
                        nodePoolsResources: [
                            {
                                nodePool: { id: Number(nodePoolId), name: nodePoolName },
                                gpu: { deserved: 1, maxAllowed: 1 }
                            }
                        ]
                    });

                    // Wait for project to be ready
                    await projectUtils.waitForProjectFromBackendNonEmptyStatus(
                        projectName,
                        project.id.toString()
                    );

                    // Submit a workload on this node pool
                    const workloadName = objectGenerator.generateNameByPrefix('train');
                    const workload = new TrainingWorkload(
                        config.quickStartImage.ubuntu,
                        projectNamespace,
                        workloadName
                    );
                    workload.arguments = 'sleep 60';
                    workload.cpu = '50m';
                    workload.gpu = '1';
                    workload.memory = '100M';
                    workload.nodePools = nodePoolName;

                    await kubeCli.createWorkload(workload);

                    // Wait for workload to be running
                    logger.info(
                        `Waiting for workload ${workloadName} to be running on node pool ${nodePoolName}`
                    );
                    await waitUtils.waitForJobToBeRunning(
                        workloadName,
                        projectNamespace,
                        crds.runaiTrainingWorkloadCrd.kind
                    );

                    // Get the workload from API
                    const trainingWorkload = await kubeCli.getTrainWorkloadOfRunaiJob(
                        workloadName,
                        projectNamespace
                    );
                    const workloadId = trainingWorkload.metadata.uid;

                    logger.info(`Workload ${workloadName} created with ID ${workloadId}`);

                    // Wait for workload to be running and get API workload
                    await workloadsUtilsV2.waitForApiWorkloadWithPhase(
                        workloadId,
                        Phase.Running,
                        workloadName
                    );
                    const apiWorkload = (await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId)).body;

                    logger.info(`API workload ${apiWorkload.name} is in phase ${apiWorkload.phase}`);
                        logger.info(
                            `API workload node pools: ${JSON.stringify(apiWorkload.currentNodePools)}`
                        );

                    // Validate that workload is running on the correct node pool
                    expect(apiWorkload.currentNodePools).toContain(nodePoolName);
                    logger.info(`Workload ${apiWorkload.name} is running on node pool ${nodePoolName}`);

                    // Validate guaranteedRuntimeEndsAt
                    if (testCase.isValid) {
                        // For valid duration, guaranteedRuntimeEndsAt should be set
                        if (testCase.guaranteedRuntimeEndsAtValueExists) {
                            expect(apiWorkload.guaranteedRuntimeEndsAt).toBeTruthy();
                            // Calculate expected end time (creation time + minGuaranteedRuntimeMilliseconds)
                            const creationTime = new Date(apiWorkload.createdAt);
                            const expectedEndTime = new Date(
                                creationTime.getTime() + testCase.minGuaranteedRuntimeMilliseconds
                            );
                            const actualEndTime = new Date(apiWorkload.guaranteedRuntimeEndsAt);

                            // Allow for small time differences (within milliseconds)
                            const timeDifference = Math.abs(
                                expectedEndTime.getTime() - actualEndTime.getTime()
                            );
                            expect(timeDifference).toBeLessThan(1000000);
                        } else {
                            expect(apiWorkload.guaranteedRuntimeEndsAt).toBeFalsy();
                        }

                        logger.info(
                            `guaranteedRuntimeEndsAt validation passed for workload ${apiWorkload.name}`
                        );
                    } else {
                        // For invalid values, guaranteedRuntimeEndsAt should be null/undefined
                        expect(apiWorkload.guaranteedRuntimeEndsAt).toBeFalsy();
                        logger.info(
                            `guaranteedRuntimeEndsAt is null/undefined as expected for workload ${apiWorkload.name}`
                        );
                    }

                    logger.info(
                        `cleaning up. deleting workload ${apiWorkload.name}, project ${projectName}, nodepool ${nodePoolName}`
                    );
                    await operations.deleteWorkloads([workload]);
                    await operations.deleteProjectAndNamespace(project.id);
                        await nodePoolUtils.deleteNodePoolByIdWithWait(
                            createdNodepool.id,
                            createdNodepool.name
                        );

                    // Remove the label from the node to clean up
                    if (nodeName) {
                        await kubeCli.unLabelNode(nodeName, nodePoolLabelKey);
                        logger.info(`Removed label ${nodePoolLabelKey} from node ${nodeName}`);
                    }
                });
            }
            })
        );
    });
});
