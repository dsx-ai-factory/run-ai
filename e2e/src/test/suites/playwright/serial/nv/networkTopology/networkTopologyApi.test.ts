/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import {
    logger,
    objectGenerator,
    tags,
    flowRootState,
    networkTopologiesUtils,
    nodeUtils,
    nodePoolUtils,
    orgUnitUtils,
    operations,
    distributedUtils,
    workloadsCrudUtils,
    waitUtils,
    config,
    workloadUtils,
    workloadsManagerUtils,
    workloadsUtilsV2,
    versionUtils,
    apiUtils
} from '../../../../../../modules/helpers';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import {
    NetworkTopologyCreateRequest,
    NodepoolCreateFields,
    NodepoolPhase
} from '../../../../../../modules/api/apis/clusterApi';
import { Project } from '../../../../../../modules/api/apis/orgUnitApi';
import { kubeCli } from '../../../../../../modules/cli';
import { DistributedFramework, ServingPortProtocol } from '../../../../../../modules/api/apis/policyApi';
import { Phase } from '../../../../../../modules/api/apis/assetsApi';
import { StatusCodes } from 'http-status-codes';
import { DistributedInferenceSpecSpec } from '../../../../../../modules/api/apis/workloadsApi/generated/api';
import { getBaseWorkloadSpec } from '../../../parallel/viper/inference/distributed/dinfUtils';
import { createDistributedInference } from '../../../parallel/viper/inference/inferenceUtils.js';
import { workloadsCrudService } from '../../../../../../modules/api';
import { WorkloadType } from '../../../../../../modules/helpers/utils/WorkloadsCrudUtils';
import { isEmpty } from 'lodash-es';

test.describe(`Network Topology API E2E Tests, tags: ${tags.tool.api} ${tags.resource.networkTopologies} ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_23)(function () {
        const topoNames: string[] = [];
        let workerNodeNames: string[] = [];
        let inferSpec: DistributedInferenceSpecSpec;
        let nodeGroup1: string[];
        let nodeGroup2: string[];
        const nodePools: { name: string; id: string }[] = [];
        const nodePoolNames: string[] = [];
        const topologyIds: string[] = [];
        let npProject: Project;
        let workloadName: string;
        let workloadId: string;
        let workloadInfo: { workloadType: WorkloadType; workloadId: string }[] = [];

        const nodeLabelKey = 'test_node_group';
        const nodeLabelVal1 = 'group1';
        const nodeLabelVal2 = 'group2';
        const alpineImage = 'alpine:latest';
        const sleepCommand = 'sleep infinity';

        const topoRegionKey = 'cloud.provider.com/region';
        const topoCityKey = 'cloud.provider.com/city';
        const region = 'us-east1';
        const cityName = 'chicago';

        const topoName = objectGenerator.generateShortNameByPrefix('ui-topo1', 5);
        topoNames.push(topoName);

        const npName1 = objectGenerator.generateShortNameByPrefix('np', 5);
        const npName2 = objectGenerator.generateShortNameByPrefix('np', 5);
        nodePoolNames.push(npName1, npName2);

        const npProjectName = objectGenerator.generateShortNameByPrefix('proj-np', 5);

        const createTopoRequest: NetworkTopologyCreateRequest = {
            name: topoName,
            clusterId: flowRootState.cluster.uuid,
            levels: [topoCityKey, topoRegionKey]
        };

        const nodePoolCreateRequest1 = Object.assign(new NodepoolCreateFields(), {
            name: npName1,
            labelKey: nodeLabelKey,
            labelValue: nodeLabelVal1,
            clusterId: flowRootState.cluster.uuid
        });

        const nodePoolCreateRequest2 = Object.assign(new NodepoolCreateFields(), {
            name: npName2,
            labelKey: nodeLabelKey,
            labelValue: nodeLabelVal2,
            clusterId: flowRootState.cluster.uuid
        });

        test.beforeAll(async () => {
            workloadsManagerUtils.accessToken = flowRootState.users.system_administrator.accessToken;
            await apiUtils.disableLimitOverQuotaSubscription();

            workerNodeNames = await nodeUtils.getWorkerNodeNames();
            inferSpec = getBaseWorkloadSpec();

            // edit 3 worker nodes with topology labels
            const nodesToLabel = workerNodeNames.slice(0, 3);
            const labels = [
                { key: topoRegionKey, value: region },
                { key: topoCityKey, value: cityName }
            ];

            await Promise.all(
                nodesToLabel.flatMap((node) =>
                    labels.map((label) => kubeCli.labelNode(node, label.key, label.value))
                )
            );

            // create topo
            const topoRsp = await networkTopologiesUtils.createNetworkTopologyViaApi(createTopoRequest);
            topologyIds.push(topoRsp.body.id);

            nodeGroup1 = workerNodeNames.slice(0, 1);
            nodeGroup2 = workerNodeNames.slice(1, 3);

            // label nodes
            const nodeGroupsWithLabels: [string[], string][] = [
                [nodeGroup1, nodeLabelVal1],
                [nodeGroup2, nodeLabelVal2]
            ];

            await Promise.all(
                nodeGroupsWithLabels.flatMap(([group, labelVal]) =>
                    group.map((node) => kubeCli.labelNode(node, nodeLabelKey, labelVal))
                )
            );

            nodePoolCreateRequest1.networkTopologyId = topoRsp.body.id;
            nodePoolCreateRequest2.networkTopologyId = topoRsp.body.id;

            const npRsp1 = await nodePoolUtils.createClusterNodePoolWithRequest(nodePoolCreateRequest1);
            const npRsp2 = await nodePoolUtils.createClusterNodePoolWithRequest(nodePoolCreateRequest2);

            // Wait for node pools to be ready
            await Promise.all(
                [npRsp1.body.id, npRsp2.body.id].map((id) =>
                    nodePoolUtils.waitForClusterNodePoolPhase(
                        id,
                        [NodepoolPhase.Ready, NodepoolPhase.Empty],
                        30
                    )
                )
            );

            nodePools.push({ name: npName1, id: npRsp1.body.id }, { name: npName2, id: npRsp2.body.id });

            const res = [npRsp1.body, npRsp2.body].map((np) => ({
                nodePool: { name: np.name, id: np.id.toString() },
                gpu: { deserved: 8, limit: -1 }
            }));

            res.push({
                nodePool: { name: 'default', id: '100000' },
                gpu: { deserved: 2, limit: -1 }
            });

            npProject = await orgUnitUtils.createProjectAndValidate({
                clusterId: flowRootState.cluster.uuid,
                resources: res,
                name: npProjectName,
                defaultNodePools: nodePoolNames
            });
        });

        test.afterAll(async () => {
            //delete project and namespace
            try {
                await operations.deleteProjectAndNamespace(parseInt(npProject.id));
                logger.info(`deleted project ${npProjectName}`);
            } catch (error) {
                logger.error(`failed to delete project ${npProjectName}`, error);
            }

            // delete node pools
            await Promise.all(
                nodePools.map(async (np) => {
                    try {
                        await nodePoolUtils.deleteNodePoolByIdWithWait(np.id, np.name);
                    } catch (error) {
                        logger.error(error);
                    }
                })
            );

            // delete topologies
            await Promise.all(
                topologyIds.map(async (id) => {
                    try {
                        await networkTopologiesUtils.deleteNetworkTopologyViaApi(id);
                    } catch (error) {
                        logger.error(`Failed to delete topology ${id}`, error);
                    }
                })
            );

            const unlabelPromises: Promise<void>[] = [];

            [nodeGroup1, nodeGroup2].forEach((group) => {
                group.forEach((node) => {
                    unlabelPromises.push(kubeCli.unLabelNode(node, nodeLabelKey));
                });
            });

            workerNodeNames.slice(0, 3).forEach((node) => {
                unlabelPromises.push(kubeCli.unLabelNode(node, topoRegionKey));
                unlabelPromises.push(kubeCli.unLabelNode(node, topoCityKey));
            });

            await Promise.all(unlabelPromises);
        });

        test.beforeEach(() => {
            workloadName = objectGenerator.generateShortNameByPrefix('topo-workload', 5);
            workloadInfo = [];
        });

        test.afterEach(async () => {
            if (workloadInfo.length !== 0) {
                await Promise.all(
                    workloadInfo.map(async (workload) => {
                        try {
                            await workloadsCrudUtils.deleteWorkload(
                                workload.workloadType,
                                workload.workloadId
                            );
                        } catch (error) {
                            logger.error(`failed to delete ${workload.workloadType} workload`, error);
                        }
                    })
                );
            }
        });

        test('T5404991: test L1 Researcher user to distributed training across multiple node pools with different topologies', async () => {
            await test.step('create a distributed PyTorch training and wait it to be running', async () => {
                const response = await distributedUtils.createDistributedTraining(
                    npProject.id,
                    workloadName,
                    {
                        distributedFramework: DistributedFramework.PyTorch,
                        image: 'ubuntu',
                        command: sleepCommand,
                        numWorkers: 2,
                        nodePools: nodePoolNames
                    },
                    undefined
                );

                expect(response.response.statusCode).toEqual(StatusCodes.ACCEPTED);

                workloadId = response.body.workloadId;
                workloadInfo.push({ workloadType: WorkloadType.DISTRIBUTED, workloadId: workloadId });

                await waitUtils.waitForCondition(
                    async () =>
                        (await workloadsCrudUtils.getDistributed(workloadId)).body.actualPhase ===
                        Phase.Running,
                    `Wait for workload ${response.body.name} to be Running`,
                    true,
                    60
                );
            });

            await test.step('verify that all pods are running on a same node', async () => {
                const pods = await kubeCli.listPodsByNamespaceAndPrefix(
                    `runai-${npProjectName}`,
                    workloadName
                );
                logger.info(pods);
            });

            await test.step('delete the distributed training', async () => {
                await workloadsCrudUtils.deleteDistributed(workloadId);
                workloadInfo.pop();
            });
        });

        test('T5404992: test ML engineer to create distributed inference across multiple node pools with different topologies', async () => {
            workloadsCrudUtils.accessToken = flowRootState.users.ml_engineer.accessToken;
            await test.step('create a distributed inference and wait it to be running', async () => {
                inferSpec.workers = 2;
                inferSpec.replicas = 1;

                // create the workload
                npProject['namespace'] = `runai-${npProjectName}`;

                [workloadId, workloadName] = await createDistributedInference(
                    workloadName,
                    inferSpec,
                    npProject
                );

                workloadInfo.push({
                    workloadType: WorkloadType.DISTRIBUTED_INFERENCE,
                    workloadId: workloadId
                });

                // wait for the workload to be in phase Running
                await waitUtils.waitForBackendWorkloadPhase(
                    workloadName,
                    parseInt(npProject.id),
                    Phase.Running,
                    120
                );
            });

            await test.step('verify that all pods are running on a same node', async () => {
                const pods = await kubeCli.listPodsByNamespaceAndPrefix(
                    `runai-${npProjectName}`,
                    workloadName
                );
                logger.info(pods);
            });

            await test.step('delete the distributed inference', async () => {
                await workloadsCrudUtils.deleteDistributedInference(workloadId);
                workloadInfo.pop();
            });
        });

        test('T5418029: test creating a single node training that uses a node pool with topoloies attached', async () => {
            workloadsCrudService.trainingsApi.accessToken =
                flowRootState.users.system_administrator.accessToken;

            await test.step('create a training that uses a node pool with topologies attached', async () => {
                const { body: workload } = await workloadsCrudUtils.createTraining(
                    workloadName,
                    false,
                    flowRootState.cluster.uuid,
                    npProject.id,
                    { image: alpineImage, command: sleepCommand }
                );

                workloadId = workload.workloadId;
                workloadInfo.push({ workloadType: WorkloadType.TRAINING, workloadId: workloadId });

                await waitUtils.waitForCondition(
                    async () =>
                        (await workloadsCrudUtils.getTraining(workload.workloadId)).body.actualPhase ===
                        Phase.Running,
                    `Wait for workload ${workload.name} to be Running`,
                    true,
                    90
                );
            });

            await test.step('delete the training', async () => {
                await workloadsCrudUtils.deleteTraining(workloadId);
                workloadInfo.pop();
            });
        });

        test('T5418030: test creating a workpsace that uses a node pool with topoloies attached', async () => {
            workloadsCrudService.workspacesApi.accessToken =
                flowRootState.users.system_administrator.accessToken;

            await test.step('create a workspace that uses a node pool with topologies attached', async () => {
                const { body: workload } = await workloadsCrudUtils.createWorkspace(
                    workloadName,
                    false,
                    flowRootState.cluster.uuid,
                    npProject.id,
                    { image: alpineImage, command: sleepCommand }
                );

                workloadId = workload.workloadId;
                workloadInfo.push({ workloadType: WorkloadType.WORKSPACE, workloadId: workloadId });

                await waitUtils.waitForCondition(
                    async () =>
                        (await workloadsCrudUtils.getWorkspace(workload.workloadId)).body.actualPhase ===
                        Phase.Running,
                    `Wait for workload ${workload.name} to be Running`,
                    true,
                    90
                );
            });

            await test.step('delete the workspace', async () => {
                await workloadsCrudUtils.deleteWorkspace(workloadId);
                workloadInfo.pop();
            });
        });

        test('T5418031: test creating an inference that uses a node pool with topoloies attached', async () => {
            workloadsCrudService.inferencesApi.accessToken =
                flowRootState.users.system_administrator.accessToken;

            await test.step('create an inference that uses a node pool with topologies attached', async () => {
                const { body: workload } = await workloadsCrudUtils.createInference(
                    workloadName,
                    false,
                    flowRootState.cluster.uuid,
                    npProject.id,
                    {
                        image: config.quickStartImage.inferenceHelloWorld,
                        servingPort: {
                            container: 8000,
                            protocol: ServingPortProtocol.Http
                        }
                    }
                );

                workloadId = workload.workloadId;
                workloadInfo.push({ workloadType: WorkloadType.INFERENCE, workloadId: workloadId });

                await waitUtils.waitForCondition(
                    async () =>
                        (await workloadsCrudUtils.getInference(workload.workloadId)).body.actualPhase ===
                        Phase.Running,
                    `Wait for workload ${workload.name} to be Running`,
                    true,
                    90
                );
            });

            await test.step('delete the inference', async () => {
                await workloadsCrudUtils.deleteInference(workloadId);
                workloadInfo.pop();
            });
        });

        supportedFrom(SupportedVersions.V2_24)(function () {
            test('T5404999: test submitting a distributed training and 3 new topology fields shown in workloads GET API', async () => {
                workloadsCrudUtils.accessToken = flowRootState.users.system_administrator.accessToken;

                await test.step('create an external distributed PyTorch training with topo annotations', async () => {
                    const manifest = workloadUtils.createPyTorchJobManifest(
                        workloadName,
                        `runai-${npProject.name}`,
                        'ubuntu',
                        sleepCommand,
                        2,
                        0
                    );

                    manifest.metadata['annotations'] = {
                        'kai.scheduler/topology': topoName,
                        'kai.scheduler/topology-preferred-placement': topoCityKey
                    };

                    const response = await workloadsManagerUtils.createWorkloadV2(
                        workloadName,
                        npProject.id,
                        manifest,
                        {
                            priority: 'medium',
                            category: 'Train'
                        }
                    );

                    expect(response.response.statusCode).toEqual(StatusCodes.ACCEPTED);

                    workloadId = response.body.metadata.id;
                    workloadInfo.push({ workloadType: WorkloadType.DISTRIBUTED, workloadId: workloadId });

                    await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 120);
                });

                await test.step('verify that 3 topology fields shown in workloads GET API', async () => {
                    const response = await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId);
                    networkTopologiesUtils.validateNetworkTopologyDataInWorkloadsApi(
                        response.body.topology,
                        topoName,
                        topoCityKey,
                        undefined
                    );
                });

                await test.step('delete the distributed training', async () => {
                    await workloadsManagerUtils.deleteWorkloadV2(workloadId);
                    await waitUtils.waitForPodsDeletion(workloadName, `runai-${npProject.name}`, 120);
                    workloadInfo.pop();
                });
            });

            test('Verify actualPlacement (LCA) is populated in workloads GET API for a running distributed training with topology', async () => {
                const clusterVersion = flowRootState.cluster.version;
                if (!versionUtils.isVersionSupported(clusterVersion, SupportedVersions.V2_25)) {
                    return;
                }

                workloadsCrudUtils.accessToken = flowRootState.users.system_administrator.accessToken;

                await test.step('create an external distributed PyTorch training with topo annotations', async () => {
                    const manifest = workloadUtils.createPyTorchJobManifest(
                        workloadName,
                        `runai-${npProject.name}`,
                        'ubuntu',
                        sleepCommand,
                        2,
                        0
                    );

                    manifest.metadata['annotations'] = {
                        'kai.scheduler/topology': topoName,
                        'kai.scheduler/topology-preferred-placement': topoCityKey
                    };

                    const response = await workloadsManagerUtils.createWorkloadV2(
                        workloadName,
                        npProject.id,
                        manifest,
                        {
                            priority: 'medium',
                            category: 'Train'
                        }
                    );

                    expect(response.response.statusCode).toEqual(StatusCodes.ACCEPTED);

                    workloadId = response.body.metadata.id;
                    workloadInfo.push({ workloadType: WorkloadType.DISTRIBUTED, workloadId: workloadId });

                    await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 120);
                });

                await test.step('verify that actualPlacement is populated in workloads GET API', async () => {
                    let response;
                    await waitUtils.waitForCondition(
                        async () => {
                            response = await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId);
                            return !isEmpty(response.body.topology?.actualPlacement);
                        },
                        `Wait for actualPlacement to be defined for workload ${workloadId}`,
                        true,
                        60
                    );
                    const topology = response.body.topology;

                    expect(topology).toBeDefined();
                    expect(topology.name).toEqual(topoName);
                    expect(topology.actualPlacement).toBeDefined();
                    expect(topology.actualPlacement).not.toBeNull();
                    expect(topology.actualPlacement.level).toEqual(topoRegionKey);
                    expect(topology.actualPlacement.value).toEqual('us-east1');
                    logger.info(
                        `actualPlacement: level=${topology.actualPlacement.level}, value=${topology.actualPlacement.value}`
                    );
                });

                await test.step('delete the distributed training', async () => {
                    await workloadsManagerUtils.deleteWorkloadV2(workloadId);
                    workloadInfo.pop();
                });
            });

            test('T5550743: Test topology info is shown in node pool API when a node pool has a topology attached', async () => {
                const response = await nodePoolUtils.getNodePoolById(nodePools[0].id);
                expect(response.response.statusCode).toEqual(StatusCodes.OK);
                expect(response.body.networkTopologyName).toEqual(topoName);
                expect(response.body.networkTopologyId).toEqual(topologyIds[0]);
            });
        });
    });
});
