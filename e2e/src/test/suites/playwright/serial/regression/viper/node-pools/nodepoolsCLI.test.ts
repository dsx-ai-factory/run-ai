/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import { clusterApi } from '../../../../../../../modules/api';
import {
    config,
    flowRootState,
    logger,
    nodePoolUtils,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    Utils,
    waitUtils
} from '../../../../../../../modules/helpers';

import { runaiCliV2 } from '../../../../../../../modules/cli';
import { skipPlaywrightTest } from '../../../../../../../modules/helpers/utils/SkippedTestUtils';
import { Project } from '../../../../../../../modules/api/apis/projectsDepartmentsApi';
import {test} from '@playwright/test';
import {
    cliV2WorkloadType,
    describeFlags,
    SubmitFlagsMerged
} from '../../../../../../../modules/cli/tools/RunaiCliV2.constants';

test.describe(`Nodepools CLI, tags:  ${tags.resource.nodepools} ${tags.teams.viper}`, function () {
    let imagePath: string;
    let projectName: string;
    let namespace: string;
    let projectCreated: Project;
    let nodePoolLabelKeyValue: {
        key: string;
        value: string;
    };
    let nodePoolCreated: {
        nodePoolName: string;
        nodePoolId: string;
    };

    test.beforeAll(async function () {
        imagePath = config.quickStartImage.ubuntu;
        projectName = objectGenerator.generateNameByPrefix('project');
        namespace = Utils.namespaceFromProjectName(projectName);
        await operations.reduceSchedulerResources();

        try {
            projectCreated = await projectResourcesUtils.createProject(namespace, {
                name: projectName,
                resources: {
                    gpu: {
                        deserved: 2,
                        maxAllowed: 2,
                        overQuotaWeight: null
                    },
                    cpu: {
                        deserved: undefined,
                        maxAllowed: undefined,
                        overQuotaWeight: null
                    },
                    memory: {
                        deserved: undefined,
                        maxAllowed: undefined,
                        overQuotaWeight: null
                    }
                }
            });
        } catch (error) {
            logger.error(`Failed to create project, error: ${error}`);
        }

        clusterApi.nodepoolsApi.accessToken = flowRootState.users.system_administrator.accessToken;
        nodePoolLabelKeyValue = nodePoolUtils.generateKeyValueLabels();
        const { nodePoolName, nodePoolId } = await nodePoolUtils.createClusterNodePoolAndValidate(
            nodePoolLabelKeyValue.key,
            nodePoolLabelKeyValue.value
        );
        nodePoolCreated = { nodePoolName, nodePoolId };
        await nodePoolUtils.waitForNodePoolCrdReadyOrEmpty(nodePoolName);
    });

    test.afterAll(async function () {
        try {
            await nodePoolUtils.cleanupDeleteClusterNodePool(nodePoolCreated);
        } catch (error) {
            logger.error(`Failed to cleanup delete nodepool, error: ${error}`);
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

        try {
            await operations.deleteProjectAndNamespace(projectCreated.id);
        } catch (error) {
            logger.error(`Failed to delete project and namespace, error: ${error}`);
        }
    });

    test.describe('Nodepools enabled in backend and cluster', function () {
        test.describe('Create job with multi node pools', function () {
            async function CreateJobWithMultiNodePools(
                wlType: cliV2WorkloadType = cliV2WorkloadType.training,
                image = imagePath
            ): Promise<void> {
                const jobName = nodePoolUtils.submitJobWithMultiNodePools(
                    projectCreated.name,
                    image,
                    [nodePoolUtils.DefaultNodePoolName, nodePoolCreated.nodePoolName],
                    wlType
                );

                const clusterPods = await waitUtils.waitForPods(
                    jobName,
                    namespace,
                    1,
                    waitUtils.waitForPodTimeout,
                    [],
                    true
                );

                expect(clusterPods).to.be.not.empty;
                clusterPods.forEach((pod) => {
                    const nodeSelectorTerms =
                        pod.spec.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution
                            .nodeSelectorTerms;
                    expect(nodeSelectorTerms[0].matchExpressions[0].key).to.be.equal(
                        nodePoolUtils.NodePoolNameLabel
                    );
                    expect(nodeSelectorTerms[0].matchExpressions[0].operator).to.be.equal('DoesNotExist');
                    expect(nodeSelectorTerms[1].matchExpressions[0].key).to.be.equal(
                        nodePoolLabelKeyValue.key
                    );
                    expect(nodeSelectorTerms[1].matchExpressions[0].operator).to.be.equal('In');
                    expect(nodeSelectorTerms[1].matchExpressions[0].values[0]).to.be.equal(
                        nodePoolLabelKeyValue.value
                    );
                });
            }

            test('Create interactive job with multi node pools', async function () {
                await CreateJobWithMultiNodePools(cliV2WorkloadType.workspace);
            });
            test('Create train job with multi node pools', async function () {
                await CreateJobWithMultiNodePools();
            });
        });

        test.describe('Wrong submit should fail', function () {
            test('Node pool not exists', function () {
                const jobName = objectGenerator.generateNameByPrefix('job');
                const nodePoolName = 'somenodepool';
                try {
                    const submitFlag = {
                        [SubmitFlagsMerged.image]: imagePath,
                        [SubmitFlagsMerged.project]: projectCreated.name,
                        [SubmitFlagsMerged.nodePools]: nodePoolName
                    } as Partial<typeof SubmitFlagsMerged>;
                    runaiCliV2.submit(submitFlag, jobName, cliV2WorkloadType.training);
                    expect.fail('Should not submit job to not existing node-pool');
                } catch (error) {
                    expect(error.message).to.contain(
                        `the following node pools do not exist: ${nodePoolName}`
                    );
                }
            });
            test('Node pools list with duplicated node pool', async function ({}, testInfo) {
                skipPlaywrightTest(
                    test,
                    testInfo,
                    'duplicate node-pool submission validation was lost in the Engine to KAI cutover'
                );
                const jobName = objectGenerator.generateNameByPrefix('job');
                const submitFlag = {
                    [SubmitFlagsMerged.image]: imagePath,
                    [SubmitFlagsMerged.project]: projectCreated.name,
                    [SubmitFlagsMerged.nodePools]: [
                        nodePoolUtils.DefaultNodePoolName,
                        nodePoolUtils.DefaultNodePoolName
                    ]
                } as Partial<typeof SubmitFlagsMerged>;

                const stdout = runaiCliV2.submit(submitFlag, jobName, cliV2WorkloadType.training);
                expect(stdout).not.to.contain('Error');

                await waitUtils.waitForCondition(
                    () => {
                        try {
                            const describeOutput = runaiCliV2.describe(
                                {[describeFlags.project]: projectCreated.name},
                                jobName,
                                cliV2WorkloadType.training
                            );
                            return describeOutput.includes('node pools names must be unique');
                        } catch (error) {
                            logger.error(`Error while running describe command: ${error}`);
                            return false;
                        }
                    },
                    'Wait for duplicate node pools issue to appear in describe output',
                    true,
                    60,
                    5
                );
            });
        });
    });
});
