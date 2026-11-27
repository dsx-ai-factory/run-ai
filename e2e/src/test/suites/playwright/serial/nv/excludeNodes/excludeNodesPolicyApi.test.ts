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
    operations,
    projectResourcesUtils,
    nodeUtils,
    tags,
    flowRootState,
    workloadsCrudUtils,
    waitUtils,
    trainingUtils,
    policyUtils,
    workspaceUtils,
    inferenceUtils,
    distributedUtils,
    departmentUtils,
    UtilsError,
    workloadsUtilsV2
} from '../../../../../../modules/helpers';
import { hostnameLabel, namespaceCluster } from '../../../../../../modules/helpers/config/const';
import { Phase } from '../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import { Department, Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { ServingPortProtocol } from '../../../../../../modules/api/apis/workloadsApi/generated/api';
import { MatchExpressionOperator } from '../../../../../../modules/api/apis/workloadsApi/generated/model/matchExpressionOperator';
import { verifyNodeExclusion, verifyNodeInclusion } from '../../../../../../modules/helpers/Validations';
import { ScopeType } from '../../../../../../modules/api/apis/policyApi';
import { WorkloadType } from '../../../../../../modules/helpers/utils/policyUtils';

test.describe(`Exclude Nodes Policy API Tests, tags: ${tags.tool.api} ${tags.resource.workloads} ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_23)(function () {
        let projectName: string;
        let namespace: string;
        let project: Project;
        let workloadName: string;
        let workloadId: string;
        let workerNodeNames: string[];
        let policyWorkloadType: WorkloadType;
        let policyCreated: boolean;
        let clusterScopedCrdPolicyName: string;

        const alpineImage = 'alpine:latest';
        const sleepCmd = 'sleep infinity';
        const prohibitMsg = 'The workload request does not comply with the policy set by the administrator';

        test.beforeAll(async function () {
            projectName = objectGenerator.generateShortNameByPrefix('proj-exclude-nodes', 5);
            namespace = Utils.namespaceFromProjectName(projectName);
            project = await projectResourcesUtils.createProject(namespace, {
                name: projectName,
                deservedGpus: 0,
                maxAllowedGpus: 0
            });

            workerNodeNames = await nodeUtils.getWorkerNodeNames();
            clusterScopedCrdPolicyName = policyUtils.getPolicyName(
                ScopeType.Cluster,
                flowRootState.cluster.uuid
            );
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

        test.beforeEach(function () {
            workloadName = objectGenerator.generateShortNameByPrefix('exclude-node-workload', 5);
            policyCreated = false;
        });

        test.describe('Test cluster scoped workload policies', function () {
            /**
             * 1. randomly choose a worker node and create cluster scoped workload policy to let workload run on it;
             * 2. if canEdit=true, workload with node exclusion overwrite workload policy rules; workload w/o node exclusion uses workload policy rules;
             * 3. if canEdit=false, workload with node exclusion will be blocked from creation; workload w/o node exclusion uses workload policy rules;
             * 4. cover 'standard training', 'distributed training', 'workspace' and 'inference' workloads testing.
             */
            test.setTimeout(5 * 60 * 1000);
            let nodeNames: string[];

            const inferenceSpec = {
                image: config.quickStartImage.inferenceHelloWorld,
                servingPort: { container: 8033, protocol: ServingPortProtocol.Http }
            };

            test.beforeEach(function () {
                nodeNames = Utils.getRandomElements(workerNodeNames, 1);
            });

            test.afterEach(async function () {
                if (policyCreated) {
                    await policyUtils.deleteWorkloadPolicyV2(policyWorkloadType);
                }
            });

            test('T5186252: test standard training workload node affinity policy with canEdit set to true', async function () {
                policyWorkloadType = WorkloadType.Training;

                await test.step('Create cluster scoped standard training node affinity policy with canEdit set to true', async function () {
                    const response = await policyUtils.createPolicyV2(
                        flowRootState.cluster.name,
                        flowRootState.cluster.uuid,
                        policyWorkloadType,
                        policyUtils.getNodeSelectorTerm(MatchExpressionOperator.In, nodeNames),
                        policyUtils.getRules(true),
                        ScopeType.Cluster
                    );
                    expect(response.response.statusCode).toEqual(200);
                    policyCreated = true;
                    await policyUtils.waitForPolicyToBeSynced(
                        policyWorkloadType,
                        namespaceCluster,
                        clusterScopedCrdPolicyName
                    );
                });

                await test.step('Verify the created policy', async function () {
                    await policyUtils.validatePolicyBody(
                        policyWorkloadType,
                        {
                            scope: ScopeType.Cluster.toString(),
                            scopeId: flowRootState.cluster.uuid
                        },
                        policyUtils.getNodeSelectorTerm(MatchExpressionOperator.In, nodeNames),
                        policyUtils.getRules(true)
                    );
                });

                await test.step('Create standard training workload with 1 node excluded', async function () {
                    const response = await trainingUtils.createTrainingV2(
                        project.id.toString(),
                        workloadName,
                        {
                            image: alpineImage,
                            command: sleepCmd
                        },
                        nodeNames
                    );

                    expect(response.response.statusCode).toEqual(202);
                    workloadId = response.body.workloadId;

                    await waitUtils.waitForCondition(
                        async () =>
                            (await workloadsCrudUtils.getTraining(workloadId)).body.actualPhase ===
                            Phase.Running,
                        `Wait for workload ${response.body.name} to be Running`,
                        true,
                        60
                    );
                });

                await test.step('Verify workload node exclusion rule override workload policy rule and workload is not running on excluded node', async function () {
                    await verifyNodeExclusion(nodeNames, namespace, workloadName);
                });

                await test.step('Delete workload', async function () {
                    await workloadsCrudUtils.deleteTraining(workloadId);
                    await waitUtils.waitForWorkloadPodNotExist(namespace, workloadName);
                });

                await test.step('Create standard training workload without node exclusion', async function () {
                    const response = await trainingUtils.createTrainingV2(
                        project.id.toString(),
                        workloadName,
                        {
                            image: alpineImage,
                            command: sleepCmd
                        }
                    );

                    expect(response.response.statusCode).toEqual(202);
                    workloadId = response.body.workloadId;

                    await waitUtils.waitForCondition(
                        async () =>
                            (await workloadsCrudUtils.getTraining(workloadId)).body.actualPhase ===
                            Phase.Running,
                        `Wait for workload ${response.body.name} to be Running`,
                        true,
                        60
                    );
                });

                await test.step('Verify workload policy is taking effect and workload is running on chosen node', async function () {
                    await verifyNodeInclusion(nodeNames, namespace, workloadName);
                });

                await test.step('Delete workload', async function () {
                    await workloadsCrudUtils.deleteTraining(workloadId);
                    await waitUtils.waitForWorkloadPodNotExist(namespace, workloadName);
                });
            });

            test('T5186253: test standard training workload node affinity policy with canEdit set to false', async function () {
                policyWorkloadType = WorkloadType.Training;
                // TO-DO: temp workaround for canEdit=false workpolicy fails to prevent workload with ndoe exclusion from creating problem.
                const succWorkloadName = `${workloadName}-2`;

                await test.step('Create cluster scoped standard training node affinity policy with canEdit set to false', async function () {
                    const response = await policyUtils.createPolicyV2(
                        flowRootState.cluster.name,
                        flowRootState.cluster.uuid,
                        policyWorkloadType,
                        policyUtils.getNodeSelectorTerm(MatchExpressionOperator.In, nodeNames),
                        policyUtils.getRules(false),
                        ScopeType.Cluster
                    );
                    expect(response.response.statusCode).toEqual(200);
                    policyCreated = true;
                    await policyUtils.waitForPolicyToBeSynced(
                        policyWorkloadType,
                        namespaceCluster,
                        clusterScopedCrdPolicyName
                    );
                });

                await test.step('Verify the created policy', async function () {
                    await policyUtils.validatePolicyBody(
                        policyWorkloadType,
                        {
                            scope: ScopeType.Cluster.toString(),
                            scopeId: flowRootState.cluster.uuid
                        },
                        policyUtils.getNodeSelectorTerm(MatchExpressionOperator.In, nodeNames),
                        policyUtils.getRules(false)
                    );
                });

                await test.step('Verify standard training with node exclusion is blocked by policy', async function () {
                    try {
                        await trainingUtils.createTrainingV2(
                            project.id.toString(),
                            workloadName,
                            {
                                image: alpineImage,
                                command: sleepCmd
                            },
                            nodeNames
                        );
                    } catch (error) {
                        expect(error.response.statusCode).toEqual(400);
                        expect(error.response.body.message).toEqual(prohibitMsg);
                    }
                });

                await test.step('Create standard training workload without node exclusion', async function () {
                    const response = await trainingUtils.createTrainingV2(
                        project.id.toString(),
                        succWorkloadName,
                        {
                            image: alpineImage,
                            command: sleepCmd
                        }
                    );

                    expect(response.response.statusCode).toEqual(202);
                    workloadId = response.body.workloadId;

                    await waitUtils.waitForCondition(
                        async () =>
                            (await workloadsCrudUtils.getTraining(workloadId)).body.actualPhase ===
                            Phase.Running,
                        `Wait for workload ${response.body.name} to be Running`,
                        true,
                        60
                    );
                });

                await test.step('Verify workload policy is taking effect and workload is running on chosen node', async function () {
                    await verifyNodeInclusion(nodeNames, namespace, succWorkloadName);
                });

                await test.step('Delete workload', async function () {
                    await workloadsCrudUtils.deleteTraining(workloadId);
                    await waitUtils.waitForWorkloadPodNotExist(namespace, succWorkloadName);
                });
            });

            test('T5186254: test distributed training workload node affinity policy with canEdit set to true', async function () {
                policyWorkloadType = WorkloadType.Distributed;

                await test.step('Create cluster scoped distributed training node affinity policy with canEdit set to true', async function () {
                    const response = await policyUtils.createPolicyV2(
                        flowRootState.cluster.name,
                        flowRootState.cluster.uuid,
                        policyWorkloadType,
                        {
                            worker: policyUtils.getNodeSelectorTerm(MatchExpressionOperator.In, nodeNames),
                            master: policyUtils.getNodeSelectorTerm(MatchExpressionOperator.In, nodeNames)
                        },
                        {
                            worker: policyUtils.getRules(true),
                            master: policyUtils.getRules(true)
                        },
                        ScopeType.Cluster
                    );
                    expect(response.response.statusCode).toEqual(200);
                    policyCreated = true;
                    await policyUtils.waitForPolicyToBeSynced(
                        policyWorkloadType,
                        namespaceCluster,
                        clusterScopedCrdPolicyName
                    );
                });

                await test.step('Verify the created policy', async function () {
                    await policyUtils.validatePolicyBody(
                        policyWorkloadType,
                        {
                            scope: ScopeType.Cluster.toString(),
                            scopeId: flowRootState.cluster.uuid
                        },
                        {
                            worker: policyUtils.getNodeSelectorTerm(MatchExpressionOperator.In, nodeNames),
                            master: policyUtils.getNodeSelectorTerm(MatchExpressionOperator.In, nodeNames)
                        },
                        {
                            worker: policyUtils.getRules(true),
                            master: policyUtils.getRules(true)
                        }
                    );
                });

                await test.step('Create distributed training workload with 1 node excluded', async function () {
                    const response = await distributedUtils.createDistributedTraining(
                        project.id.toString(),
                        workloadName,
                        undefined,
                        undefined,
                        nodeNames
                    );

                    expect(response.response.statusCode).toEqual(202);
                    workloadId = response.body.workloadId;

                    await waitUtils.waitForCondition(
                        async () =>
                            (await workloadsCrudUtils.getDistributed(workloadId)).body.actualPhase ===
                            Phase.Running,
                        `Wait for workload ${response.body.name} to be Running`,
                        true,
                        60
                    );
                });

                await test.step('Verify workload node exclusion rule override workload policy rule and workload is not running on excluded node', async function () {
                    await verifyNodeExclusion(nodeNames, namespace, workloadName);
                });

                await test.step('Delete workload', async function () {
                    await workloadsCrudUtils.deleteDistributed(workloadId);
                    await waitUtils.waitForWorkloadPodNotExist(namespace, workloadName);
                });

                await test.step('Create distributed training workload without node exclusion', async function () {
                    const response = await distributedUtils.createDistributedTraining(
                        project.id.toString(),
                        workloadName
                    );

                    expect(response.response.statusCode).toEqual(202);
                    workloadId = response.body.workloadId;

                    await waitUtils.waitForCondition(
                        async () =>
                            (await workloadsCrudUtils.getDistributed(workloadId)).body.actualPhase ===
                            Phase.Running,
                        `Wait for workload ${response.body.name} to be Running`,
                        true,
                        60
                    );
                });

                await test.step('Verify workload policy is taking effect and workload is running on chosen node', async function () {
                    await verifyNodeInclusion(nodeNames, namespace, workloadName);
                });

                await test.step('Delete workload', async function () {
                    await workloadsCrudUtils.deleteDistributed(workloadId);
                    await waitUtils.waitForWorkloadPodNotExist(namespace, workloadName);
                });
            });

            test('T5186255: test distributed training workload node affinity policy with canEdit set to false', async function () {
                policyWorkloadType = WorkloadType.Distributed;
                // TO-DO: temp workaround for canEdit=false workpolicy fails to prevent workload with ndoe exclusion from creating problem.
                const succWorkloadName = `${workloadName}-2`;

                await test.step('Create cluster scoped distributed training node affinity policy with canEdit set to false', async function () {
                    const response = await policyUtils.createPolicyV2(
                        flowRootState.cluster.name,
                        flowRootState.cluster.uuid,
                        policyWorkloadType,
                        {
                            worker: policyUtils.getNodeSelectorTerm(MatchExpressionOperator.In, nodeNames),
                            master: policyUtils.getNodeSelectorTerm(MatchExpressionOperator.In, nodeNames)
                        },
                        {
                            worker: policyUtils.getRules(false),
                            master: policyUtils.getRules(false)
                        },
                        ScopeType.Cluster
                    );
                    expect(response.response.statusCode).toEqual(200);
                    policyCreated = true;
                    await policyUtils.waitForPolicyToBeSynced(
                        policyWorkloadType,
                        namespaceCluster,
                        clusterScopedCrdPolicyName
                    );
                });

                await test.step('Verify the created policy', async function () {
                    await policyUtils.validatePolicyBody(
                        policyWorkloadType,
                        {
                            scope: ScopeType.Cluster.toString(),
                            scopeId: flowRootState.cluster.uuid
                        },
                        {
                            worker: policyUtils.getNodeSelectorTerm(MatchExpressionOperator.In, nodeNames),
                            master: policyUtils.getNodeSelectorTerm(MatchExpressionOperator.In, nodeNames)
                        },
                        {
                            worker: policyUtils.getRules(false),
                            master: policyUtils.getRules(false)
                        }
                    );
                });

                await test.step('Verify distributed workload with node exclusion is blocked by policy', async function () {
                    try {
                        await distributedUtils.createDistributedTraining(
                            project.id.toString(),
                            workloadName,
                            undefined,
                            undefined,
                            nodeNames
                        );
                    } catch (error) {
                        expect(error.response.statusCode).toEqual(400);
                        expect(error.response.body.message).toEqual(prohibitMsg);
                    }
                });

                await test.step('Create distributed training workload without node exclusion', async function () {
                    const response = await distributedUtils.createDistributedTraining(
                        project.id.toString(),
                        succWorkloadName
                    );

                    expect(response.response.statusCode).toEqual(202);
                    workloadId = response.body.workloadId;

                    await waitUtils.waitForCondition(
                        async () =>
                            (await workloadsCrudUtils.getDistributed(workloadId)).body.actualPhase ===
                            Phase.Running,
                        `Wait for workload ${response.body.name} to be Running`,
                        true,
                        60
                    );
                });

                await test.step('Verify workload policy is taking effect and workload is running on chosen node', async function () {
                    await verifyNodeInclusion(nodeNames, namespace, succWorkloadName);
                });

                await test.step('Delete workload', async function () {
                    await workloadsCrudUtils.deleteDistributed(workloadId);
                    await waitUtils.waitForWorkloadPodNotExist(namespace, succWorkloadName);
                });
            });

            test('T5186256: test workspace workload node affinity policy with canEdit set to true', async function () {
                policyWorkloadType = WorkloadType.Interactive;

                await test.step('Create cluster scoped workspace workload node affinity policy with canEdit set to true', async function () {
                    const response = await policyUtils.createPolicyV2(
                        flowRootState.cluster.name,
                        flowRootState.cluster.uuid,
                        policyWorkloadType,
                        policyUtils.getNodeSelectorTerm(MatchExpressionOperator.In, nodeNames),
                        policyUtils.getRules(true),
                        ScopeType.Cluster
                    );
                    expect(response.response.statusCode).toEqual(200);
                    policyCreated = true;
                    await policyUtils.waitForPolicyToBeSynced(
                        policyWorkloadType,
                        namespaceCluster,
                        clusterScopedCrdPolicyName
                    );
                });

                await test.step('Verify the created policy', async function () {
                    await policyUtils.validatePolicyBody(
                        policyWorkloadType,
                        {
                            scope: ScopeType.Cluster.toString(),
                            scopeId: flowRootState.cluster.uuid
                        },
                        policyUtils.getNodeSelectorTerm(MatchExpressionOperator.In, nodeNames),
                        policyUtils.getRules(true)
                    );
                });

                await test.step('Create workspace workload with 1 node excluded', async function () {
                    const response = await workspaceUtils.createWorkspace(
                        project.id.toString(),
                        workloadName,
                        {
                            image: alpineImage,
                            command: sleepCmd
                        },
                        nodeNames
                    );

                    expect(response.response.statusCode).toEqual(202);
                    workloadId = response.body.workloadId;

                    await waitUtils.waitForCondition(
                        async () =>
                            (await workloadsCrudUtils.getWorkspace(workloadId)).body.actualPhase ===
                            Phase.Running,
                        `Wait for workload ${response.body.name} to be Running`,
                        true,
                        60
                    );
                });

                await test.step('Verify workload node exclusion rule override workload policy rule and workload is not running on excluded node', async function () {
                    await verifyNodeExclusion(nodeNames, namespace, workloadName);
                });

                await test.step('Delete workload', async function () {
                    await workloadsCrudUtils.deleteWorkspace(workloadId);
                    await waitUtils.waitForWorkloadPodNotExist(namespace, workloadName);
                });

                await test.step('Create workspace workload without node exclusion', async function () {
                    const response = await workspaceUtils.createWorkspace(
                        project.id.toString(),
                        workloadName,
                        {
                            image: alpineImage,
                            command: sleepCmd
                        }
                    );

                    expect(response.response.statusCode).toEqual(202);
                    workloadId = response.body.workloadId;

                    await waitUtils.waitForCondition(
                        async () =>
                            (await workloadsCrudUtils.getWorkspace(workloadId)).body.actualPhase ===
                            Phase.Running,
                        `Wait for workload ${response.body.name} to be Running`,
                        true,
                        60
                    );
                });

                await test.step('Verify workload policy is taking effect and workload is running on chosen node', async function () {
                    await verifyNodeInclusion(nodeNames, namespace, workloadName);
                });

                await test.step('Delete workload', async function () {
                    await workloadsCrudUtils.deleteWorkspace(workloadId);
                    await waitUtils.waitForWorkloadPodNotExist(namespace, workloadName);
                });
            });

            test('T5186257: test workspace workload node affinity policy with canEdit set to false', async function () {
                policyWorkloadType = WorkloadType.Interactive;
                // TO-DO: temp workaround for canEdit=false workpolicy fails to prevent workload with ndoe exclusion from creating problem.
                const succWorkloadName = `${workloadName}-2`;

                await test.step('Create cluster scoped workspace node affinity policy with canEdit set to false', async function () {
                    const response = await policyUtils.createPolicyV2(
                        flowRootState.cluster.name,
                        flowRootState.cluster.uuid,
                        policyWorkloadType,
                        policyUtils.getNodeSelectorTerm(MatchExpressionOperator.In, nodeNames),
                        policyUtils.getRules(false),
                        ScopeType.Cluster
                    );
                    expect(response.response.statusCode).toEqual(200);
                    policyCreated = true;
                    await policyUtils.waitForPolicyToBeSynced(
                        policyWorkloadType,
                        namespaceCluster,
                        clusterScopedCrdPolicyName
                    );
                });

                await test.step('Verify the created policy', async function () {
                    await policyUtils.validatePolicyBody(
                        policyWorkloadType,
                        {
                            scope: ScopeType.Cluster.toString(),
                            scopeId: flowRootState.cluster.uuid
                        },
                        policyUtils.getNodeSelectorTerm(MatchExpressionOperator.In, nodeNames),
                        policyUtils.getRules(false)
                    );
                });

                await test.step('Verify workspace workload with node exclusion is blocked by policy', async function () {
                    try {
                        await workspaceUtils.createWorkspace(
                            project.id.toString(),
                            workloadName,
                            {
                                image: alpineImage,
                                command: sleepCmd
                            },
                            nodeNames
                        );
                    } catch (error) {
                        expect(error.response.statusCode).toEqual(400);
                        expect(error.response.body.message).toEqual(prohibitMsg);
                    }
                });

                await test.step('Create workspace workload without node exclusion', async function () {
                    const response = await workspaceUtils.createWorkspace(
                        project.id.toString(),
                        succWorkloadName,
                        {
                            image: alpineImage,
                            command: sleepCmd
                        }
                    );

                    expect(response.response.statusCode).toEqual(202);
                    workloadId = response.body.workloadId;

                    await waitUtils.waitForCondition(
                        async () =>
                            (await workloadsCrudUtils.getWorkspace(workloadId)).body.actualPhase ===
                            Phase.Running,
                        `Wait for workload ${response.body.name} to be Running`,
                        true,
                        60
                    );
                });

                await test.step('Verify workload policy is taking effect and workload is running on chosen node', async function () {
                    await verifyNodeInclusion(nodeNames, namespace, succWorkloadName);
                });

                await test.step('Delete workload', async function () {
                    await workloadsCrudUtils.deleteWorkspace(workloadId);
                    await waitUtils.waitForWorkloadPodNotExist(namespace, succWorkloadName);
                });
            });

            test('T5186258: test inference workload node affinity policy with canEdit set to true', async function () {
                policyWorkloadType = WorkloadType.Inference;

                await test.step('Create cluster scoped inference workload node affinity policy with canEdit set to true', async function () {
                    const response = await policyUtils.createPolicyV2(
                        flowRootState.cluster.name,
                        flowRootState.cluster.uuid,
                        policyWorkloadType,
                        policyUtils.getNodeSelectorTerm(MatchExpressionOperator.In, nodeNames),
                        policyUtils.getRules(true),
                        ScopeType.Cluster
                    );
                    expect(response.response.statusCode).toEqual(200);
                    policyCreated = true;
                    await policyUtils.waitForPolicyToBeSynced(
                        policyWorkloadType,
                        namespaceCluster,
                        clusterScopedCrdPolicyName
                    );
                });

                await test.step('Verify the created policy', async function () {
                    await policyUtils.validatePolicyBody(
                        policyWorkloadType,
                        {
                            scope: ScopeType.Cluster.toString(),
                            scopeId: flowRootState.cluster.uuid
                        },
                        policyUtils.getNodeSelectorTerm(MatchExpressionOperator.In, nodeNames),
                        policyUtils.getRules(true)
                    );
                });

                await test.step('Create inference workload with 1 node excluded', async function () {
                    inferenceSpec['nodeAffinityRequired'] = policyUtils.getNodeSelectorTerm(
                        MatchExpressionOperator.NotIn,
                        nodeNames
                    ).nodeAffinityRequired;

                    [workloadId, workloadName] = await inferenceUtils.createInference(
                        workloadName,
                        inferenceSpec,
                        project.id.toString()
                    );

                    await waitUtils.waitForCondition(
                        async () =>
                            (await workloadsCrudUtils.getInference(workloadId)).body.actualPhase ===
                            Phase.Running,
                        `Wait for workload ${workloadName} to be Running`,
                        true,
                        60
                    );

                    inferenceSpec['nodeAffinityRequired'] = undefined;
                });

                await test.step('Verify workload node exclusion rule override workload policy rule and workload is not running on excluded node', async function () {
                    await verifyNodeExclusion(nodeNames, namespace, workloadName);
                });

                await test.step('Delete workload', async function () {
                    await workloadsCrudUtils.deleteInference(workloadId);
                    await waitUtils.waitForWorkloadPodNotExist(namespace, workloadName);
                });

                await test.step('Create inference workload without node exclusion', async function () {
                    [workloadId, workloadName] = await inferenceUtils.createInference(
                        workloadName,
                        inferenceSpec,
                        project.id.toString()
                    );

                    await waitUtils.waitForCondition(
                        async () =>
                            (await workloadsCrudUtils.getInference(workloadId)).body.actualPhase ===
                            Phase.Running,
                        `Wait for workload ${workloadName} to be Running`,
                        true,
                        60
                    );
                });

                await test.step('Verify workload policy is taking effect and workload is running on chosen node', async function () {
                    await verifyNodeInclusion(nodeNames, namespace, workloadName);
                });

                await test.step('Delete workload', async function () {
                    await workloadsCrudUtils.deleteInference(workloadId);
                    await waitUtils.waitForWorkloadPodNotExist(namespace, workloadName);
                });
            });

            test('T5186259: test inference workload node affinity policy with canEdit set to false', async function () {
                policyWorkloadType = WorkloadType.Inference;
                // TO-DO: temp workaround for canEdit=false workpolicy fails to prevent workload with ndoe exclusion from creating problem.
                const succWorkloadName = `${workloadName}-2`;

                await test.step('Create cluster scoped inference node affinity policy with canEdit set to false', async function () {
                    const response = await policyUtils.createPolicyV2(
                        flowRootState.cluster.name,
                        flowRootState.cluster.uuid,
                        policyWorkloadType,
                        policyUtils.getNodeSelectorTerm(MatchExpressionOperator.In, nodeNames),
                        policyUtils.getRules(false),
                        ScopeType.Cluster
                    );
                    expect(response.response.statusCode).toEqual(200);
                    policyCreated = true;
                    await policyUtils.waitForPolicyToBeSynced(
                        policyWorkloadType,
                        namespaceCluster,
                        clusterScopedCrdPolicyName
                    );
                });

                await test.step('Verify the created policy', async function () {
                    await policyUtils.validatePolicyBody(
                        policyWorkloadType,
                        {
                            scope: ScopeType.Cluster.toString(),
                            scopeId: flowRootState.cluster.uuid
                        },
                        policyUtils.getNodeSelectorTerm(MatchExpressionOperator.In, nodeNames),
                        policyUtils.getRules(false)
                    );
                });

                await test.step('Verify inference workload with node exclusion is blocked by policy', async function () {
                    try {
                        inferenceSpec['nodeAffinityRequired'] = policyUtils.getNodeSelectorTerm(
                            MatchExpressionOperator.NotIn,
                            nodeNames
                        ).nodeAffinityRequired;

                        [workloadId, workloadName] = await inferenceUtils.createInference(
                            workloadName,
                            inferenceSpec,
                            project.id.toString()
                        );
                    } catch (error) {
                        expect(error.response.statusCode).toEqual(400);
                        expect(error.response.body.message).toEqual(prohibitMsg);
                    } finally {
                        inferenceSpec['nodeAffinityRequired'] = undefined;
                    }
                });

                await test.step('Create inference workload without node exclusion', async function () {
                    [workloadId, workloadName] = await inferenceUtils.createInference(
                        succWorkloadName,
                        inferenceSpec,
                        project.id.toString()
                    );

                    await waitUtils.waitForCondition(
                        async () =>
                            (await workloadsCrudUtils.getInference(workloadId)).body.actualPhase ===
                            Phase.Running,
                        `Wait for workload ${workloadName} to be Running`,
                        true,
                        60
                    );
                });

                await test.step('Verify workload policy is taking effect and workload is running on chosen node', async function () {
                    await verifyNodeInclusion(nodeNames, namespace, succWorkloadName);
                });

                await test.step('Delete workload', async function () {
                    await workloadsCrudUtils.deleteInference(workloadId);
                    await waitUtils.waitForWorkloadPodNotExist(namespace, succWorkloadName);
                });
            });
        });

        test.describe('Test node affinity workload policy action scopes', function () {
            test.setTimeout(5 * 60 * 1000);
            let nodeNames: string[];
            let depmtId: number;
            let projectNameList: Array<string>;
            let namespaceList: Array<string>;
            let projectIdList: Array<number>;
            let scope: string;
            let scopeId: string;
            let defaultDepmt: Department;

            test.beforeAll(async function () {
                const dept = await departmentUtils.createBackendDepartmentWithName(
                    objectGenerator.generateShortNameByPrefix('d1')
                );
                depmtId = Number(dept.body.id);

                projectNameList = Array.from({ length: 2 }, () =>
                    objectGenerator.generateShortNameByPrefix('proj-policy-scope', 5)
                );
                namespaceList = projectNameList.map((name) => Utils.namespaceFromProjectName(name));

                // create 1 project under dept, 1 project under the 'default'
                const projectsToCreate = [
                    { namespace: namespaceList[0], name: projectNameList[0] },
                    { namespace: namespaceList[1], name: projectNameList[1], departmentId: depmtId }
                ];

                const projResponse = await Promise.all(
                    projectsToCreate.map((config) =>
                        projectResourcesUtils.createProject(config.namespace, {
                            name: config.name,
                            deservedGpus: 0,
                            maxAllowedGpus: 0,
                            ...(config.departmentId && { departmentId: config.departmentId })
                        })
                    )
                );
                projectIdList = projResponse.map((proj) => proj.id);

                defaultDepmt = await departmentUtils.getDefaultDepartmentByCluster(
                    flowRootState.cluster.uuid
                );
            });

            test.afterAll('deleting project and namespace', async () => {
                // delete all projects and ns
                await Promise.all(
                    projectIdList.map(async (projectId, i) => {
                        try {
                            await operations.deleteProjectAndNamespace(projectId);
                            logger.info(
                                `deleted project ${projectNameList[i]} and namespace ${namespaceList[i]}`
                            );
                        } catch (error) {
                            logger.error(
                                `failed to delete project ${projectNameList[i]} and namespace ${namespaceList[i]}`,
                                error
                            );
                        }
                    })
                );

                // delete department
                await departmentUtils.deleteDepartment(depmtId);
            });

            test.beforeEach(function () {
                nodeNames = Utils.getRandomElements(workerNodeNames, 1);
            });

            test.afterEach(async function () {
                if (policyCreated) {
                    await policyUtils.deleteWorkloadPolicyV2(policyWorkloadType, {
                        scope: scope,
                        scopeId: scopeId
                    });
                }
            });

            test('T5186260: Test department action scope for workload policy', async function () {
                policyWorkloadType = WorkloadType.Training;
                const scopeType = ScopeType.Department;
                scope = scopeType.toString();
                scopeId = defaultDepmt.id.toString();

                await test.step('Create cluster scoped standard training node affinity policy with canEdit set to false', async function () {
                    const response = await policyUtils.createPolicyV2(
                        defaultDepmt.name,
                        scopeId,
                        policyWorkloadType,
                        policyUtils.getNodeSelectorTerm(MatchExpressionOperator.In, nodeNames),
                        policyUtils.getRules(false),
                        scopeType
                    );
                    expect(response.response.statusCode).toEqual(200);
                    policyCreated = true;
                    const depmtScopedCrdPolicyName = policyUtils.getPolicyName(scopeType, scopeId);
                    await policyUtils.waitForPolicyToBeSynced(
                        policyWorkloadType,
                        namespaceCluster,
                        depmtScopedCrdPolicyName
                    );
                });

                await test.step('Verify the created policy', async function () {
                    await policyUtils.validatePolicyBody(
                        policyWorkloadType,
                        {
                            scope: scope,
                            scopeId: scopeId
                        },
                        policyUtils.getNodeSelectorTerm(MatchExpressionOperator.In, nodeNames),
                        policyUtils.getRules(false)
                    );
                });

                await test.step('Verify training workload with node exclusion within policy scope is blocked by policy', async function () {
                    try {
                        await trainingUtils.createTrainingV2(
                            projectIdList[0].toString(),
                            workloadName,
                            {
                                image: alpineImage,
                                command: sleepCmd
                            },
                            nodeNames
                        );
                    } catch (error) {
                        expect(error.response.statusCode).toEqual(400);
                        expect(error.response.body.message).toEqual(prohibitMsg);
                    }
                });

                await test.step('Verify training workload with node exclusion out of policy scope can be created', async function () {
                    const response = await trainingUtils.createTrainingV2(
                        projectIdList[1].toString(),
                        workloadName,
                        {
                            image: alpineImage,
                            command: sleepCmd
                        },
                        nodeNames
                    );

                    expect(response.response.statusCode).toEqual(202);
                    workloadId = response.body.workloadId;

                    await waitUtils.waitForCondition(
                        async () =>
                            (await workloadsCrudUtils.getTraining(workloadId)).body.actualPhase ===
                            Phase.Running,
                        `Wait for workload ${response.body.name} to be Running`,
                        true,
                        60
                    );
                });

                await test.step('Verify workload is not running on excluded node', async function () {
                    await verifyNodeExclusion(nodeNames, namespaceList[1], workloadName);
                });

                await test.step('Delete workload', async function () {
                    await workloadsCrudUtils.deleteTraining(workloadId);
                    await waitUtils.waitForWorkloadPodNotExist(namespaceList[1], workloadName);
                });
            });

            test('T5186261: Test project action scope for workload policy', async function () {
                policyWorkloadType = WorkloadType.Training;
                const scopeType = ScopeType.Project;
                scope = scopeType.toString();
                scopeId = projectIdList[1].toString();

                await test.step('Create cluster scoped standard training node affinity policy with canEdit set to false', async function () {
                    const response = await policyUtils.createPolicyV2(
                        projectNameList[1],
                        scopeId,
                        policyWorkloadType,
                        policyUtils.getNodeSelectorTerm(MatchExpressionOperator.In, nodeNames),
                        policyUtils.getRules(false),
                        scopeType
                    );
                    expect(response.response.statusCode).toEqual(200);
                    policyCreated = true;
                    const projectScopedCrdPolicyName = policyUtils.getPolicyName(scopeType, scopeId);
                    await policyUtils.waitForPolicyToBeSynced(
                        policyWorkloadType,
                        namespaceList[1],
                        projectScopedCrdPolicyName
                    );
                });

                await test.step('Verify the created policy', async function () {
                    await policyUtils.validatePolicyBody(
                        policyWorkloadType,
                        {
                            scope: scope,
                            scopeId: scopeId
                        },
                        policyUtils.getNodeSelectorTerm(MatchExpressionOperator.In, nodeNames),
                        policyUtils.getRules(false)
                    );
                });

                await test.step('Verify training workload with node exclusion within policy scope is blocked by policy', async function () {
                    try {
                        await trainingUtils.createTrainingV2(
                            projectIdList[1].toString(),
                            workloadName,
                            {
                                image: alpineImage,
                                command: sleepCmd
                            },
                            nodeNames
                        );
                    } catch (error) {
                        expect(error.response.statusCode).toEqual(400);
                        expect(error.response.body.message).toEqual(prohibitMsg);
                    }
                });

                await test.step('Verify training workload with node exclusion out of policy scope can be created', async function () {
                    const response = await trainingUtils.createTrainingV2(
                        projectIdList[0].toString(),
                        workloadName,
                        {
                            image: alpineImage,
                            command: sleepCmd
                        },
                        nodeNames
                    );

                    expect(response.response.statusCode).toEqual(202);
                    workloadId = response.body.workloadId;

                    await waitUtils.waitForCondition(
                        async () =>
                            (await workloadsCrudUtils.getTraining(workloadId)).body.actualPhase ===
                            Phase.Running,
                        `Wait for workload ${response.body.name} to be Running`,
                        true,
                        60
                    );
                });

                await test.step('Verify workload is not running on excluded node', async function () {
                    await verifyNodeExclusion(nodeNames, namespaceList[0], workloadName);
                });

                await test.step('Delete workload', async function () {
                    await workloadsCrudUtils.deleteTraining(workloadId);
                    await waitUtils.waitForWorkloadPodNotExist(namespaceList[0], workloadName);
                });
            });
        });

        test.describe('Test edit workload policy', function () {
            test.afterEach(async function () {
                if (policyCreated) {
                    await policyUtils.deleteWorkloadPolicyV2(policyWorkloadType);
                }
            });

            test('T5186901: Edit node affinity policy and it populates the new settings correctly', async function () {
                policyWorkloadType = WorkloadType.Training;
                const originalNodeNames = workerNodeNames.slice(0, 1);
                const editedNodeNames = workerNodeNames.slice(1, 2);
                // TO-DO: temp workaround for canEdit=false workpolicy fails to prevent workload with ndoe exclusion from creating problem.
                const succWorkloadName = `${workloadName}-2`;

                await test.step('Create cluster scoped standard training node affinity policy with canEdit set to false', async function () {
                    const response = await policyUtils.createPolicyV2(
                        flowRootState.cluster.name,
                        flowRootState.cluster.uuid,
                        policyWorkloadType,
                        policyUtils.getNodeSelectorTerm(MatchExpressionOperator.In, originalNodeNames),
                        policyUtils.getRules(false),
                        ScopeType.Cluster
                    );
                    expect(response.response.statusCode).toEqual(200);
                    policyCreated = true;
                    await policyUtils.waitForPolicyToBeSynced(
                        policyWorkloadType,
                        namespaceCluster,
                        clusterScopedCrdPolicyName
                    );
                });

                await test.step('Verify the created policy', async function () {
                    await policyUtils.validatePolicyBody(
                        policyWorkloadType,
                        {
                            scope: ScopeType.Cluster.toString(),
                            scopeId: flowRootState.cluster.uuid
                        },
                        policyUtils.getNodeSelectorTerm(MatchExpressionOperator.In, originalNodeNames),
                        policyUtils.getRules(false)
                    );
                });

                await test.step('Verify standard training with node exclusion is blocked by policy', async function () {
                    try {
                        await trainingUtils.createTrainingV2(
                            project.id.toString(),
                            workloadName,
                            {
                                image: alpineImage,
                                command: sleepCmd
                            },
                            originalNodeNames
                        );
                    } catch (error) {
                        expect(error.response.statusCode).toEqual(400);
                        expect(error.response.body.message).toEqual(prohibitMsg);
                    }
                });

                await test.step('Edit the workload policy and change the canEdit to true', async function () {
                    const response = await policyUtils.createPolicyV2(
                        flowRootState.cluster.name,
                        flowRootState.cluster.uuid,
                        policyWorkloadType,
                        policyUtils.getNodeSelectorTerm(MatchExpressionOperator.In, editedNodeNames),
                        policyUtils.getRules(true),
                        ScopeType.Cluster
                    );
                    expect(response.response.statusCode).toEqual(200);
                    await policyUtils.waitForPolicyToBeSynced(
                        policyWorkloadType,
                        namespaceCluster,
                        clusterScopedCrdPolicyName
                    );
                });

                await test.step('Verify the edited policy', async function () {
                    await policyUtils.validatePolicyBody(
                        policyWorkloadType,
                        {
                            scope: ScopeType.Cluster.toString(),
                            scopeId: flowRootState.cluster.uuid
                        },
                        policyUtils.getNodeSelectorTerm(MatchExpressionOperator.In, editedNodeNames),
                        policyUtils.getRules(true)
                    );
                });

                await test.step('Create standard training workload without node exclusion', async function () {
                    const response = await trainingUtils.createTrainingV2(
                        project.id.toString(),
                        succWorkloadName,
                        {
                            image: alpineImage,
                            command: sleepCmd
                        }
                    );
                    expect(response.response.statusCode).toEqual(202);
                    workloadId = response.body.workloadId;
                    await waitUtils.waitForCondition(
                        async () =>
                            (await workloadsCrudUtils.getTraining(workloadId)).body.actualPhase ===
                            Phase.Running,
                        `Wait for workload ${response.body.name} to be Running`,
                        true,
                        60
                    );
                });

                await test.step('Verify workload node exclusion rule override workload policy rule and workload is not running on excluded node', async function () {
                    await verifyNodeInclusion(editedNodeNames, namespace, succWorkloadName);
                });

                await test.step('Delete workload', async function () {
                    await workloadsCrudUtils.deleteTraining(workloadId);
                    await waitUtils.waitForWorkloadPodNotExist(namespace, succWorkloadName);
                });
            });
        });

        test.describe('Negative tests for workload policy', function () {
            test.afterEach(async function () {
                if (policyCreated) {
                    await policyUtils.deleteWorkloadPolicyV2(policyWorkloadType);
                }
            });

            test('T5189369: test invalid node affinity when creating training policy', async function () {
                const nodeNames = Utils.getRandomElements(workerNodeNames, 1);

                await test.step('submit invalid node affinity rule and it should fail with error return code', async function () {
                    try {
                        await policyUtils.createPolicyV2(
                            flowRootState.cluster.name,
                            flowRootState.cluster.uuid,
                            WorkloadType.Training,
                            policyUtils.getNodeSelectorTerm(MatchExpressionOperator.In, nodeNames),
                            {
                                // use a typo of 'nodeAffinityRequire' here
                                nodeAffRequired: {
                                    canEdit: true
                                }
                            },
                            ScopeType.Cluster,
                            true
                        );
                    } catch (error) {
                        expect(error.response.statusCode).toEqual(400);
                        expect(error.body.message).toContain('unknown field');
                    }
                });
            });

            test('T5189370: test user defined node affinities have contradictions', async function () {
                policyWorkloadType = WorkloadType.Training;
                const nodeNames = Utils.getRandomElements(workerNodeNames, 1);

                await test.step('Create cluster scoped training policy with contradictory rules', async function () {
                    const nodeSelectorTerm = policyUtils.getNodeSelectorTerm(
                        MatchExpressionOperator.In,
                        nodeNames
                    );
                    nodeSelectorTerm.nodeAffinityRequired.nodeSelectorTerms[0].matchExpressions.push({
                        key: hostnameLabel,
                        operator: MatchExpressionOperator.NotIn,
                        values: nodeNames
                    });

                    const response = await policyUtils.createPolicyV2(
                        flowRootState.cluster.name,
                        flowRootState.cluster.uuid,
                        policyWorkloadType,
                        nodeSelectorTerm,
                        policyUtils.getRules(false),
                        ScopeType.Cluster
                    );
                    expect(response.response.statusCode).toEqual(200);
                    policyCreated = true;
                    await policyUtils.waitForPolicyToBeSynced(
                        policyWorkloadType,
                        namespaceCluster,
                        clusterScopedCrdPolicyName
                    );
                });

                await test.step('Create standard training workload without node exclusion and it should be stuck at pending', async function () {
                    const response = await trainingUtils.createTrainingV2(
                        project.id.toString(),
                        workloadName,
                        {
                            image: alpineImage,
                            command: sleepCmd
                        }
                    );

                    expect(response.response.statusCode).toEqual(202);
                    workloadId = response.body.workloadId;

                    try {
                        await waitUtils.waitForCondition(
                            async () =>
                                (await workloadsCrudUtils.getTraining(workloadId)).body.actualPhase ===
                                Phase.Running,
                            `Wait for workload ${response.body.name} to be Running`,
                            true,
                            60
                        );
                    } catch (error) {
                        if (error instanceof UtilsError) {
                            const workload = await workloadsUtilsV2.getWorkloadByNameAndProjectName(
                                workloadName,
                                projectName
                            );
                            expect(workload.phase).toBe('Pending');
                        } else {
                            throw error;
                        }
                    }
                });

                await test.step('Delete workload', async function () {
                    await workloadsCrudUtils.deleteTraining(workloadId);
                    await waitUtils.waitForWorkloadPodNotExist(namespace, workloadName);
                });
            });
        });
    });
});
