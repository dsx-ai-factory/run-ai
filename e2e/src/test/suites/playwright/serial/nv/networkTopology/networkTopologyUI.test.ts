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
    Utils,
    projectResourcesUtils,
    distributedUtils,
    workloadsCrudUtils,
    waitUtils,
    workloadsUtilsV2
} from '../../../../../../modules/helpers';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { ClusterPage } from '../../../../../../modules/ui/runaiV2/pages/ClusterPage';
import { NetworkTopologyModal } from '../../../../../../modules/ui/runaiV2/pages/NetworkTopologyModal';
import {
    NetworkTopologyCreateRequest,
    NodepoolCreateFields,
    NodepoolPhase
} from '../../../../../../modules/api/apis/clusterApi';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { bashCli, kubeCli } from '../../../../../../modules/cli';
import { NodePool } from '../../../../../../modules/api/apis/backendClusterApi';
import { NodePoolPage } from '../../../../../../modules/ui/runaiV2/pages/NodePoolPage';
import { DistributedFramework } from '../../../../../../modules/api/apis/policyApi';
import { StatusCodes } from 'http-status-codes';
import { Phase } from '../../../../../../modules/api/apis/assetsApi';
import { ProjectUpdateRequest } from '../../../../../../modules/api/apis/orgUnitApi';
import {
    DISTRIBUTED_TOOLS_LABELS,
    TrainingPage
} from '../../../../../../modules/ui/runaiV2/pages/TrainingPage';
import { HybridForm } from '../../../../../../modules/ui/runaiV2/pages/HybridSubmissionForm';
import { workloadsCrudService } from '../../../../../../modules/api';

test.describe(`Network Topology UI Tests, tags: ${tags.tool.ui} ${tags.resource.networkTopologies} ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_23)(function () {
        let browserEngine: BrowserEngine;
        let authPage: AuthPage;
        let clusterPage: ClusterPage;
        let trainingPage: TrainingPage;
        let hybirdForm: HybridForm;
        let topoModal: NetworkTopologyModal;
        let nodePoolPage: NodePoolPage;
        let topoNames: string[] = [];
        let workerNodeNames: string[];
        let videoName: string;
        let midIndex: number;
        let nodeGroup1: string[];
        let nodeGroup2: string[];

        const image = 'alpine:latest';
        const cmd = 'sleep infinity';

        const nodeLabelKey = objectGenerator.generateShortNameByPrefix('test-ng', 5);
        const nodeLabelVal1 = 'group1';
        const nodeLabelVal2 = 'group2';
        const topoRegionKey = 'cloud.provider.com/region';
        const topoRackKey = 'cloud.provider.com/rack';
        const topoBlockKey = 'cloud.provider.com/block';
        const regrionName = 'us-east1';
        const rackName = 'rack-1';
        const blockName = 'block-1';
        const mnnvlPodAffinityKey = 'nvidia.com/gpu.clique';

        test.beforeAll(async () => {
            browserEngine = new BrowserEngine();

            workerNodeNames = await nodeUtils.getWorkerNodeNames();

            // edit 2 worker nodes with topology labels
            midIndex = Math.floor(workerNodeNames.length / 2) + 1;
            nodeGroup1 = workerNodeNames.slice(1, midIndex);
            nodeGroup2 = workerNodeNames.slice(midIndex);

            // label nodes for topologies
            // Add region label to all nodes in both groups
            await Promise.all([
                ...nodeGroup1.map((node) => kubeCli.labelNode(node, topoRegionKey, regrionName)),
                ...nodeGroup2.map((node) => kubeCli.labelNode(node, topoRegionKey, regrionName))
            ]);

            // Add rack label to all nodes in nodeGroup1
            await Promise.all(nodeGroup1.map((node) => kubeCli.labelNode(node, topoRackKey, rackName)));

            // Add block label to all nodes in nodeGroup2
            await Promise.all(nodeGroup2.map((node) => kubeCli.labelNode(node, topoBlockKey, blockName)));
        });

        test.afterAll(async () => {
            // Remove region label from all nodes in both groups
            await Promise.all(
                [...nodeGroup1, ...nodeGroup2].map((node) => kubeCli.unLabelNode(node, topoRegionKey))
            );

            // Remove rack label from all nodes in nodeGroup1
            await Promise.all(nodeGroup1.map((node) => kubeCli.unLabelNode(node, topoRackKey)));

            // Remove block label from all nodes in nodeGroup2
            await Promise.all(nodeGroup2.map((node) => kubeCli.unLabelNode(node, topoBlockKey)));
        });

        test.beforeEach(async () => {
            await browserEngine.start();
            authPage = new AuthPage(browserEngine.currentPage);
            clusterPage = new ClusterPage(browserEngine.currentPage);
            topoModal = new NetworkTopologyModal(browserEngine.currentPage);
            nodePoolPage = new NodePoolPage(browserEngine.currentPage);
            trainingPage = new TrainingPage(browserEngine.currentPage);
            hybirdForm = new HybridForm(browserEngine.currentPage);

            await authPage.signIn(
                flowRootState.users.system_administrator.email,
                flowRootState.users.system_administrator.password
            );

            topoNames = [];
        });

        test.afterEach('stop browser', async ({}, testInfo) => {
            await browserEngine.stopAndArchive(testInfo, videoName);

            await Promise.all(
                topoNames.map(async (topoName) => {
                    const response = await networkTopologiesUtils.getNetworkTopologiesViaApi([
                        `name==${topoName}`
                    ]);
                    const topo = response.body.networkTopologies;
                    if (topo && topo.length > 0) {
                        try {
                            await networkTopologiesUtils.deleteNetworkTopologyViaApi(topo[0].id);
                        } catch (error) {
                            logger.error(`Failed to delete topology ${topoName}`, error);
                        }
                    }
                })
            );
        });

        test.describe('test topology CRUD', function () {
            test('T5404990: test system admin user can create/view/delete network topology', async function () {
                videoName = 'topo-crud-ui';
                const topoName = objectGenerator.generateShortNameByPrefix('ui-crud-topo', 5);
                topoNames.push(topoName);

                await test.step('create a network topology', async function () {
                    await clusterPage.goToPage();
                    await clusterPage.openNetworkTopoCreationModal(flowRootState.cluster.name);
                    await topoModal.createTopo(topoName, [topoRackKey, topoRegionKey]);
                });

                await test.step('verify the topology is created successfully', async function () {
                    await clusterPage.clickNeworkTopologyBtn();
                    const topoList = await topoModal.get_topo_name_list();
                    logger.info(`topo list: ${topoList}`);
                    expect(topoList.some((item) => item === topoName)).toBeTruthy();
                });

                await test.step('delete the created topology', async function () {
                    await topoModal.deleteTopoByName(topoName);
                    const topoList = await topoModal.get_topo_name_list();
                    logger.info(`topo list: ${topoList}`);
                    expect(topoList).not.toContain(topoName);
                    topoNames = topoNames.filter((item) => item !== topoName);
                });
            });
        });

        test.describe('test topology - node pools E2E workflow', () => {
            const nodePoolNames: string[] = [];
            let npId: string;
            let np2Id: string;
            const topoIdList: string[] = [];
            let projectName: string;
            let namespace: string;
            let project: Project;
            let workloadName: string;
            let workloadId: string;
            const workloadIdList: string[] = [];
            let networkTopoColumnEnabled = false;
            let defaultNodePool: NodePool;
            let defaultNodePoolId: string;
            let projectUpdateReq: ProjectUpdateRequest;

            const topoName1 = objectGenerator.generateShortNameByPrefix('ui-np-topo1', 5);
            const topoName2 = objectGenerator.generateShortNameByPrefix('ui-np-topo2', 5);
            const npName = objectGenerator.generateShortNameByPrefix('np', 5);
            const npName2 = objectGenerator.generateShortNameByPrefix('np-2', 5);
            nodePoolNames.push(npName, npName2);

            const createRequest1: NetworkTopologyCreateRequest = {
                name: topoName1,
                clusterId: flowRootState.cluster.uuid,
                levels: [topoRackKey, topoRegionKey]
            };

            const createRequest2: NetworkTopologyCreateRequest = {
                name: topoName2,
                clusterId: flowRootState.cluster.uuid,
                levels: [topoBlockKey, topoRegionKey]
            };

            const nodePoolCreateRequest = Object.assign(new NodepoolCreateFields(), {
                name: npName,
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
                defaultNodePool = await nodePoolUtils.getDefaultNodePoolFromBackend();
                defaultNodePoolId = defaultNodePool.id.toString();

                // create 2 topologies
                const topoRsp1 = await networkTopologiesUtils.createNetworkTopologyViaApi(createRequest1);
                const topoRsp2 = await networkTopologiesUtils.createNetworkTopologyViaApi(createRequest2);
                topoIdList.push(topoRsp1.body.id);
                topoIdList.push(topoRsp2.body.id);

                await networkTopologiesUtils.waitForNetworkTopologies([`name==${topoName1}`]);
                await networkTopologiesUtils.waitForNetworkTopologies([`name==${topoName2}`]);

                // label nodes
                const nodeGroupsWithLabels: [string[], string][] = [
                    [nodeGroup1, nodeLabelVal1],
                    [nodeGroup2, nodeLabelVal2]
                ];

                nodeGroupsWithLabels.forEach(([group, labelVal]) => {
                    bashCli.exec(
                        `kubectl label node ${group.join(' ')} ${nodeLabelKey}=${labelVal} --overwrite`
                    );
                });

                const npRsp = await nodePoolUtils.createClusterNodePoolWithRequest(nodePoolCreateRequest);
                const npRsp2 = await nodePoolUtils.createClusterNodePoolWithRequest(nodePoolCreateRequest2);
                npId = npRsp.body.id;
                np2Id = npRsp2.body.id;

                // Wait for node pools to be ready
                await Promise.all(
                    [npRsp.body.id, npRsp2.body.id].map((id) =>
                        nodePoolUtils.waitForClusterNodePoolPhase(
                            id,
                            [NodepoolPhase.Ready, NodepoolPhase.Empty],
                            30
                        )
                    )
                );

                projectName = objectGenerator.generateShortNameByPrefix('proj-topo-ui', 5);
                namespace = Utils.namespaceFromProjectName(projectName);
                project = await projectResourcesUtils.createProject(namespace, {
                    name: projectName,
                    deservedGpus: 0,
                    maxAllowedGpus: 0
                });

                projectUpdateReq = {
                    defaultNodePools: [npName],
                    resources: [
                        {
                            nodePool: {
                                id: defaultNodePoolId,
                                name: 'default'
                            },
                            gpu: {
                                deserved: 0,
                                limit: -1
                            }
                        },
                        {
                            nodePool: {
                                id: npId,
                                name: npName
                            },
                            gpu: {
                                deserved: 0,
                                limit: -1
                            }
                        },
                        {
                            nodePool: {
                                id: np2Id,
                                name: npName2
                            },
                            gpu: {
                                deserved: 0,
                                limit: -1
                            }
                        }
                    ]
                };
            });

            test.afterAll('clean up test data', async ({}, testInfo) => {
                // delete node pools
                await Promise.all(
                    [
                        { name: npName, id: npId },
                        { name: npName2, id: np2Id }
                    ].map(async (np) => {
                        try {
                            await nodePoolUtils.deleteNodePoolByIdWithWait(np.id, np.name);
                        } catch (error) {
                            logger.error(error);
                        }
                    })
                );

                // unlabel nodes
                [nodeGroup1, nodeGroup2].forEach((group) => {
                    try {
                        bashCli.exec(`kubectl label node ${group.join(' ')} ${nodeLabelKey}-`);
                    } catch (error) {
                        logger.error(error);
                    }
                });

                // delete topologies
                await Promise.all(
                    topoIdList.map(async (topoId) => {
                        try {
                            await networkTopologiesUtils.deleteNetworkTopologyViaApi(topoId);
                        } catch (error) {
                            logger.error(`Failed to delete topo: ${topoId}`, error);
                        }
                    })
                );

                // delete project and namespace
                try {
                    await operations.deleteProjectAndNamespace(project.id);
                    logger.info(`deleted project ${projectName} and namespace ${namespace}`);
                } catch (error) {
                    logger.error(`failed to delete project ${projectName} and namespace ${namespace}`, error);
                }
            });

            test.beforeEach(async () => {
                workloadName = objectGenerator.generateShortNameByPrefix('topo-wl', 5);

                if (!networkTopoColumnEnabled) {
                    await nodePoolPage.goToPage();
                    await nodePoolPage.setColumn('Network topology', true);
                    networkTopoColumnEnabled = true;
                }
            });

            test.afterEach(async ({}, testInfo) => {
                await Promise.all(
                    workloadIdList.map(async (id) => {
                        try {
                            await workloadsCrudUtils.deleteDistributed(id);
                            await waitUtils.waitForWorkloadPodNotExist(namespace, workloadName);
                        } catch (error) {
                            logger.error(`Failed to delete workload: ${workloadName}`, error);
                        }
                    })
                );

                // hide 'Network topology' column on node pool page
                if (networkTopoColumnEnabled) {
                    await nodePoolPage.goToPage();
                    await nodePoolPage.setColumn('Network topology', false);
                    networkTopoColumnEnabled = false;
                }
            });

            test('T5429408: test clicking the Network topologies views in cluster grid and validate data in cluster topo modal', async () => {
                videoName = 'test-cluster-topo-modal';

                await test.step('go to cluster page and let the network topologies column to display', async () => {
                    await clusterPage.goToPage();
                    await clusterPage.setColumn('Network topologies', true);
                });

                await test.step('open the cluster topologies view modal and validate that created topologies info', async () => {
                    await clusterPage.openClusterNetworkTopologiesModal(flowRootState.cluster.name);

                    const topo1 = await clusterPage.networkTopoModal.getTopoData(topoName1);
                    const topo2 = await clusterPage.networkTopoModal.getTopoData(topoName2);

                    expect(topo1['Topology']).toEqual(topoName1);
                    expect(topo1['Labels']).toEqual(`${topoRackKey}, ${topoRegionKey}`);

                    expect(topo2['Topology']).toEqual(topoName2);
                    expect(topo2['Labels']).toEqual(`${topoBlockKey}, ${topoRegionKey}`);

                    await clusterPage.networkTopoModal.closeModal();
                });
            });

            test('T5429409: test create new network topology via node pool', async () => {
                videoName = 'create_topo_via_node_pool';
                const topoName = objectGenerator.generateShortNameByPrefix('new-topo-via-np', 5);
                const labels = ['cloud.provider.com/building', 'cloud.provider.com/region'];

                await test.step('go to node pool page and open node pool edit form', async () => {
                    await nodePoolPage.goToPage();
                    await nodePoolPage.openEditForm(npName);
                });

                await test.step('create a new network topology via node pool', async () => {
                    await nodePoolPage.createTopo(topoName, labels);
                    topoNames.push(topoName);

                    await nodePoolPage.clickCancelBtn();
                });

                await test.step('go to clusters page and validate the created topo data', async () => {
                    await clusterPage.goToPage();
                    await clusterPage.setColumn('Network topologies', true);

                    await clusterPage.openClusterNetworkTopologiesModal(flowRootState.cluster.name);

                    const topo = await clusterPage.networkTopoModal.getTopoData(topoName);

                    expect(topo['Topology']).toEqual(topoName);
                    expect(topo['Labels']).toEqual(`${labels.join(', ')}`);

                    await clusterPage.networkTopoModal.closeModal();
                });
            });

            test('T5404993: test attaching a topology to an existing node pool and then detaching it', async () => {
                videoName = 'attach-detach-topo';

                await test.step(`attach topology ${topoName1} to nodepool ${npName}`, async () => {
                    await nodePoolPage.goToPage();
                    await nodePoolPage.openEditForm(npName);
                    await nodePoolPage.attachTopology(topoName1);
                    await nodePoolPage.submitNodePoolForm();
                    await nodePoolUtils.waitForClusterNodePoolPhase(npId, [NodepoolPhase.Ready], 60);
                });

                await test.step('verify the topology is successfully attached to the node pool', async () => {
                    const data = await nodePoolPage.getNodePoolData(npName);
                    logger.info(`node pool ${npName} data: ${data}`);
                    expect(data['Network topology']).toEqual(topoName1);
                });

                await test.step('verify the Network topology column is clickable and validate the popup modal data', async () => {
                    await nodePoolPage.openNodePoolTopoModal(npName);
                    const labels = await nodePoolPage.getTopoLabels();
                    expect(labels.slice().sort()).toEqual([topoRackKey, topoRegionKey].slice().sort());
                    await nodePoolPage.page.getByRole('button', { name: /^close$/i }).click();
                });

                await test.step('detach the topology from the node pool', async () => {
                    await nodePoolPage.openEditForm(npName);
                    await nodePoolPage.detachTopology();
                    await nodePoolPage.submitNodePoolForm();
                    await nodePoolUtils.waitForClusterNodePoolPhase(npId, [NodepoolPhase.Ready], 60);
                });

                await test.step('verify the topology is successfully detached from the node pool', async () => {
                    const data = await nodePoolPage.getNodePoolData(npName);
                    logger.info(`node pool ${npName} data: ${data}`);
                    expect(data['Network topology']).toEqual('-');
                });
            });

            test('T5404995: test editing topology that attaches to a node pool', async () => {
                videoName = 'edit-topo';

                await test.step(`attach topology ${topoName1} to nodepool ${npName}`, async () => {
                    await nodePoolPage.goToPage();
                    await nodePoolPage.openEditForm(npName);
                    await nodePoolPage.attachTopology(topoName1);
                    await nodePoolPage.submitNodePoolForm();
                    await nodePoolUtils.waitForClusterNodePoolPhase(
                        npId,
                        [NodepoolPhase.Ready, NodepoolPhase.Empty],
                        30
                    );
                });

                await test.step('verify the topology is successfully attached to the node pool', async () => {
                    const data = await nodePoolPage.getNodePoolData(npName);
                    logger.info(`node pool ${npName} data: ${data}`);
                    expect(data['Network topology']).toEqual(topoName1);
                });

                await test.step(`edit the nodepool ${npName} and attach another topology ${topoName2} to it`, async () => {
                    await nodePoolPage.openEditForm(npName);
                    await nodePoolPage.attachTopology(topoName2);
                    await nodePoolPage.submitNodePoolForm();
                    await nodePoolUtils.waitForClusterNodePoolPhase(
                        npId,
                        [NodepoolPhase.Ready, NodepoolPhase.Empty],
                        30
                    );
                });

                await test.step('verify the edit was successful', async () => {
                    const data = await nodePoolPage.getNodePoolData(npName);
                    logger.info(`node pool ${npName} data: ${data}`);
                    expect(data['Network topology']).toEqual(topoName2);
                });
            });

            test('T5404996: test attaching a topology to multiple node pools', async () => {
                videoName = 'attach-topo-to-multi-nps';

                await test.step(`attach topology ${topoName1} to nodepool ${npName}`, async () => {
                    await nodePoolPage.goToPage();
                    await nodePoolPage.openEditForm(npName);
                    await nodePoolPage.attachTopology(topoName1);
                    await nodePoolPage.submitNodePoolForm();
                    await nodePoolUtils.waitForClusterNodePoolPhase(npId, [NodepoolPhase.Ready], 60);
                });

                await test.step('verify the topology is successfully attached to the node pool', async () => {
                    const data = await nodePoolPage.getNodePoolData(npName);
                    logger.info(`node pool ${npName} data: ${data}`);
                    expect(data['Network topology']).toEqual(topoName1);
                });

                await test.step(`attach topology ${topoName1} to nodepool default`, async () => {
                    await nodePoolPage.goToPage();
                    await nodePoolPage.openEditForm('default');
                    await nodePoolPage.attachTopology(topoName1);
                    await nodePoolPage.submitNodePoolForm();
                    await nodePoolUtils.waitForClusterNodePoolPhase(
                        defaultNodePoolId,
                        [NodepoolPhase.Ready],
                        30
                    );
                });

                await test.step(`verify the topology ${topoName1} is successfully attached to the node pool default`, async () => {
                    const data = await nodePoolPage.getNodePoolData('default');
                    logger.info(`node pool ${npName} data: ${data}`);
                    expect(data['Network topology']).toEqual(topoName1);
                });

                await test.step('update the project and set its default node pool includes node pool 1', async () => {
                    await orgUnitUtils.updateProject(project.id.toString(), projectUpdateReq);
                    await orgUnitUtils.waitForProjectFromBackendToBeReady(projectName, project.id.toString());
                });

                await test.step('create a distributed PyTorch training using L1 Researcher and wait it to be running', async () => {
                    const response = await distributedUtils.createDistributedTraining(
                        project.id.toString(),
                        workloadName,
                        {
                            distributedFramework: DistributedFramework.PyTorch,
                            image: 'ubuntu',
                            command: 'sleep infinity',
                            numWorkers: 2
                        },
                        undefined
                    );

                    expect(response.response.statusCode).toEqual(StatusCodes.ACCEPTED);
                    workloadId = response.body.workloadId;

                    await waitUtils.waitForCondition(
                        async () =>
                            (await workloadsCrudUtils.getDistributed(workloadId)).body.actualPhase ===
                            Phase.Running,
                        `Wait for workload ${response.body.name} to be Running`,
                        true,
                        120
                    );
                });

                await test.step('Delete the workload', async function () {
                    await workloadsCrudUtils.deleteDistributed(workloadId);
                    await waitUtils.waitForWorkloadPodNotExist(namespace, workloadName);
                    workloadIdList.filter((id) => id !== workloadId);
                });
            });

            test('T5404997: test starting a stopped distributed training after network topology is attached to the node pool it is using', async () => {
                test.setTimeout(5 * 60 * 1000);
                videoName = 'start-stopped-workload-with-topo-attached';

                await test.step('update the project and set its default node pool to np1 which has no topo attached', async () => {
                    await orgUnitUtils.updateProject(project.id.toString(), projectUpdateReq);
                    await orgUnitUtils.waitForProjectFromBackendToBeReady(projectName, project.id.toString());
                });

                await test.step('create a distributed training under the updated project (default nodepool has no topo attached)', async () => {
                    const response = await distributedUtils.createDistributedTraining(
                        project.id.toString(),
                        workloadName,
                        {
                            distributedFramework: DistributedFramework.PyTorch,
                            image: 'alpine',
                            command: 'sleep infinity',
                            numWorkers: 1,
                            nodePools: [npName2]
                        },
                        undefined
                    );

                    expect(response.response.statusCode).toEqual(StatusCodes.ACCEPTED);
                    workloadId = response.body.workloadId;

                    await waitUtils.waitForCondition(
                        async () =>
                            (await workloadsCrudUtils.getDistributed(workloadId)).body.actualPhase ===
                            Phase.Running,
                        `Wait for workload ${workloadName} to be Running`,
                        true,
                        120
                    );
                });

                await test.step('stop the distributed training', async () => {
                    const response = await workloadsCrudUtils.suspendDistributedTraining(workloadId);
                    expect(response.response.statusCode).toEqual(StatusCodes.ACCEPTED);

                    await waitUtils.waitForCondition(
                        async () =>
                            (await workloadsCrudUtils.getDistributed(workloadId)).body.actualPhase ===
                            Phase.Stopped,
                        `Wait for workload ${workloadName} to be Stopped`,
                        true,
                        120
                    );
                });

                await test.step(`attach topology ${topoName1} to nodepool ${npName}`, async () => {
                    await nodePoolPage.goToPage();
                    await nodePoolPage.openEditForm(npName);
                    await nodePoolPage.attachTopology(topoName1);
                    await nodePoolPage.submitNodePoolForm();
                    await nodePoolUtils.waitForClusterNodePoolPhase(npId, [NodepoolPhase.Ready], 60);
                });

                await test.step('resume the distributed training and it should be running again', async () => {
                    const response = await workloadsCrudUtils.resumeDistributedTraining(workloadId);
                    expect(response.response.statusCode).toEqual(StatusCodes.ACCEPTED);

                    await waitUtils.waitForCondition(
                        async () =>
                            (await workloadsCrudUtils.getDistributed(workloadId)).body.actualPhase ===
                            Phase.Running,
                        `Wait for workload ${workloadName} to be Running`,
                        true,
                        120
                    );
                });

                await test.step('delete the workload', async () => {
                    const response = await workloadsCrudUtils.deleteDistributed(workloadId);
                    expect(response.response.statusCode).toEqual(StatusCodes.ACCEPTED);
                    await waitUtils.waitForWorkloadPodNotExist(namespace, workloadName);
                });

                await test.step(`detach topology ${topoName1} to nodepool ${npName}`, async () => {
                    await nodePoolPage.goToPage();
                    await nodePoolPage.openEditForm(npName);
                    await nodePoolPage.detachTopology();
                    await nodePoolPage.submitNodePoolForm();
                    await nodePoolUtils.waitForClusterNodePoolPhase(npId, [NodepoolPhase.Ready], 60);
                });
            });

            test('T5404994: test deleting a topology that is attached to a node pool', async () => {
                videoName = 'delete-topo-attached-to-np';

                await test.step(`attach topology ${topoName1} to nodepool ${npName}`, async () => {
                    await nodePoolPage.goToPage();
                    await nodePoolPage.openEditForm(npName);
                    await nodePoolPage.attachTopology(topoName1);
                    await nodePoolPage.submitNodePoolForm();
                    await nodePoolUtils.waitForClusterNodePoolPhase(npId, [NodepoolPhase.Ready], 60);
                });

                await test.step('verify the topology is successfully attached to the node pool', async () => {
                    const data = await nodePoolPage.getNodePoolData(npName);
                    logger.info(`node pool ${npName} data: ${data}`);
                    expect(data['Network topology']).toEqual(topoName1);
                });

                await test.step(`delete topology ${topoName1}`, async () => {
                    await networkTopologiesUtils.deleteNetworkTopologyViaApi(topoIdList[0]);
                    topoIdList.splice(0, 1);
                });

                await test.step(`verify the topology ${topoName1} is successfully detached to the node pool default`, async () => {
                    await expect
                        .poll(
                            async () => {
                                const data = await nodePoolPage.getNodePoolData(npName);
                                logger.info(`node pool ${npName} data: ${JSON.stringify(data)}`);
                                return data['Network topology'];
                            },
                            {
                                timeout: 60000,
                                intervals: [1000]
                            }
                        )
                        .toEqual('-');
                });
            });

            test('T5404998: test both MNNVL (general topology) and the network topology are enabled, verify network topoloy rule override MNNVL rule', async () => {
                videoName = 'topo-takes-precedence-over-mnnvl';

                projectUpdateReq.defaultNodePools = [npName2];
                await test.step('update the project and set its default node pool to np2', async () => {
                    await orgUnitUtils.updateProject(project.id.toString(), projectUpdateReq);
                    await orgUnitUtils.waitForProjectFromBackendToBeReady(projectName, project.id.toString());
                });

                await test.step('update node pool with topo attached and enable the MNNVL', async () => {
                    await nodePoolPage.goToPage();
                    await nodePoolPage.openEditForm(npName2);
                    await nodePoolPage.attachTopology(topoName2);
                    await nodePoolPage.clickMnnvlIsPresent();
                    await nodePoolPage.submitNodePoolForm();
                    await nodePoolUtils.waitForClusterNodePoolPhase(np2Id, [NodepoolPhase.Ready], 60);
                });

                await test.step('create a distributed training and set workload topology to rack level)', async () => {
                    await trainingPage.goToPage();
                    await trainingPage.fillFlexTrainingFirstPage(
                        projectName,
                        workloadName,
                        true,
                        DISTRIBUTED_TOOLS_LABELS.PY_TORCH,
                        true
                    );

                    await hybirdForm.fillImageInput(image);
                    await hybirdForm.fillRuntimeSettings(cmd);
                    await hybirdForm.selectExistingComputeResource('cpu-only');
                    await hybirdForm.fillInputBySelector(hybirdForm.hybridComputeSection.workersInput, '2');
                    await hybirdForm.openNodesSection();
                    // there is no node with the mnnvlPodAffinityKey label, use it to test topology scheduling rule take precedence over MNNVL pod affinity.
                    await hybirdForm.setTopology(mnnvlPodAffinityKey, 'Preferred');
                    await hybirdForm.clickOnSubmit();
                    await hybirdForm.page.waitForLoadState('load');
                    // sometimes there is UI rendering delay(DOM state is ready to click but UI button is grayed out) when clicking the submit button
                    await hybirdForm.waitForDistributedMasterSubmitButton();
                    await hybirdForm.clickOnSubmitMasterForm();

                    await trainingPage.waitForSubmitButton();
                    await hybirdForm.clickOnSubmit();

                    // need to add a delay since sometimes the get workload api is fired before workload is created in the backend
                    await hybirdForm.page.waitForTimeout(500);
                    workloadId = await workloadsUtilsV2.getWorkloadIdByName(workloadName, project.id);
                    workloadsCrudService.distributedApi.accessToken =
                        flowRootState.users.system_administrator.accessToken;
                    await waitUtils.waitForCondition(
                        async () =>
                            (await workloadsCrudUtils.getDistributed(workloadId)).body.actualPhase ===
                            Phase.Running,
                        `Wait for workload ${workloadName} to be Running`,
                        true,
                        120
                    );
                });

                await test.step('verify the pods are running on nodes within nodeGroup2 (np2)', async () => {
                    const pods = await kubeCli.listPodsByNamepsace(namespace);
                    // only get workload pods in current test
                    const workloadPods = pods.filter((pod) => pod.metadata.name.includes(workloadName));
                    expect(workloadPods.length).toBeGreaterThan(0);

                    const oWideCmd = `kubectl get pod -n ${namespace} -o wide`;
                    logger.info(`-o wide cmd: \n${oWideCmd}`);
                    logger.warn(`kgp -o wide: \n${bashCli.exec(oWideCmd)}`);

                    for (const pod of workloadPods) {
                        expect(nodeGroup2).toContain(pod.spec?.nodeName);
                    }
                });

                await test.step('delete the workload', async () => {
                    const response = await workloadsCrudUtils.deleteDistributed(workloadId);
                    expect(response.response.statusCode).toEqual(StatusCodes.ACCEPTED);
                    await waitUtils.waitForWorkloadPodNotExist(namespace, workloadName);
                });
            });

            test('T5405000: update node pool topology and verify the topology column in the grid and topo modal also updated', async () => {
                videoName = 'update-topo-and-verify';

                await test.step(`update the attached topology in node pool ${npName}`, async () => {
                    await nodePoolPage.goToPage();
                    await nodePoolPage.openEditForm(npName);
                    await nodePoolPage.attachTopology(topoName2);
                    await nodePoolPage.submitNodePoolForm();
                    await nodePoolUtils.waitForClusterNodePoolPhase(np2Id, [NodepoolPhase.Ready], 60);
                });

                await test.step('verify the topology is successfully attached to the node pool', async () => {
                    const data = await nodePoolPage.getNodePoolData(npName);
                    logger.info(`node pool ${npName} data: ${data}`);
                    expect(data['Network topology']).toEqual(topoName2);
                });

                await test.step('verify the Network topology column is clickable and validate the popup modal data', async () => {
                    await nodePoolPage.openNodePoolTopoModal(npName);
                    const labels = await nodePoolPage.getTopoLabels();
                    expect(labels.slice().sort()).toEqual([topoBlockKey, topoRegionKey].slice().sort());
                    await nodePoolPage.page.getByRole('button', { name: /^close$/i }).click();
                });
            });
        });
    });
});
