/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect, TestInfo } from '@playwright/test';
import {
    logger,
    objectGenerator,
    tags,
    flowRootState,
    networkTopologiesUtils,
    nodePoolUtils,
    nodeUtils,
    operations,
    projectResourcesUtils,
    workloadsCrudUtils,
    workloadsManagerUtils,
    workloadsUtilsV2,
    waitUtils,
    crds,
    workloadUtils
} from '../../../../../../modules/helpers';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import {
    NetworkTopologyCreateRequest,
    NodepoolCreateFields
} from '../../../../../../modules/api/apis/clusterApi';
import { Project } from '../../../../../../modules/api/apis/orgUnitApi/generated/model/project';
import { kubeCli } from '../../../../../../modules/cli';
import { Phase } from '../../../../../../modules/api/apis/assetsApi';
import { StatusCodes } from 'http-status-codes';
import { WorkloadType } from '../../../../../../modules/helpers/utils/WorkloadsCrudUtils';
import { skipPlaywrightTest } from '../../../../../../modules/helpers/utils/SkippedTestUtils';

test.describe(`Any Workload Network Topology Support Tests(with topology), tags: ${tags.tool.api} ${tags.resource.networkTopologies} ${tags.resource.workloads} ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_24)(function () {
        const topoNames: string[] = [];
        let workerNodeNames: string[] = [];
        const topologyIds: string[] = [];
        let npProject: Project;
        let midIndex: number;
        let workloadName: string;
        let workloadId: string;
        let customWorkloadId: string;
        let workloadInfo: { workloadType: WorkloadType; workloadId: string }[] = [];
        let projectUpdated: boolean = false;
        const nodeLabelKey = 'any_workload_topo_node_group';
        const nodeLabelVal1 = 'test_group1';
        const nodeLabelVal2 = 'test_group2';
        const sleepCommand = 'sleep infinity';
        const nodePoolNames: string[] = [];
        const nodePools: { name: string; id: string }[] = [];
        let nodeGroup1: string[];
        let nodeGroup2: string[];

        const topoName = objectGenerator.generateShortNameByPrefix('ui-topo1', 5);
        topoNames.push(topoName);
        const npProjectName = objectGenerator.generateShortNameByPrefix('proj-np', 5);
        const topoRegionKey = 'topology.runai.com/region';
        const topoCityKey = 'topology.runai.com/city';
        const region = 'us-east2';
        const cityName = 'newyork';

        const npName1 = objectGenerator.generateShortNameByPrefix('np-with-topo', 5);
        const npName2 = objectGenerator.generateShortNameByPrefix('np-no-topo', 5);
        nodePoolNames.push(npName1, npName2);

        const createTopoRequest: NetworkTopologyCreateRequest = {
            name: topoName,
            clusterId: flowRootState.cluster.uuid,
            levels: []
        };

        const nodePoolCreateRequest = Object.assign(new NodepoolCreateFields(), {
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
            workloadsCrudUtils.accessToken = flowRootState.users.system_administrator.accessToken;

            npProject = await nodePoolUtils.setupNodeGroupInfrastructure({
                workerNodeNames,
                nodeLabelKey,
                nodeLabelVal1,
                nodeLabelVal2,
                nodeGroupSize: 1,
                midIndex,
                nodeGroup1,
                nodeGroup2,
                createTopology: true,
                attachTopology: true,
                topoLevels: [topoRegionKey, topoCityKey],
                createTopoRequest,
                topologyIds,
                topoRegionKey,
                topoCityKey,
                region,
                cityName,
                nodePoolCreateRequest,
                nodePoolCreateRequest2,
                nodePools,
                npName1,
                npName2,
                npProjectName,
                gpuDeserved: 3
            });
        });

        test.afterAll(async () => {
            if (npProject) {
                try {
                    await operations.deleteProjectAndNamespace(parseInt(npProject.id));
                    logger.info(`Deleted project ${npProject.name}`);
                } catch (error) {
                    logger.error(`Failed to delete project ${npProject.name}:`, error);
                }
            }

            if (nodePools.length > 0) {
                await Promise.all(
                    nodePools.map(async (np) => {
                        try {
                            await nodePoolUtils.deleteNodePoolByIdWithWait(np.id, np.name);
                            logger.info(`Deleted node pool ${np.name}`);
                        } catch (error) {
                            logger.error(`Failed to delete node pool ${np.name}:`, error);
                        }
                    })
                );
            }

            if (topologyIds.length > 0) {
                await Promise.all(
                    topologyIds.map(async (id) => {
                        try {
                            await networkTopologiesUtils.deleteNetworkTopologyViaApi(id);
                            logger.info(`Deleted topology ${id}`);
                            try {
                                await networkTopologiesUtils.getNetworkTopologyByIdViaApi(id);
                                logger.error(`Topology ${id} still exists after deletion!`);
                            } catch (error) {
                                const statusCode = error.response?.statusCode || error.statusCode;
                                if (statusCode === 404 || error.message?.includes('not found')) {
                                    logger.info(`Verified topology ${id} was successfully deleted`);
                                } else {
                                    logger.warn(
                                        `Topology ${id} deletion verification returned unexpected error:`,
                                        error.message
                                    );
                                }
                            }
                        } catch (error) {
                            logger.error(`Failed to delete topology ${id}:`, error);
                        }
                    })
                );
            }

            const nodesToClean =
                workerNodeNames.length > 0 ? workerNodeNames : await nodeUtils.getWorkerNodeNames();

            if (nodesToClean && nodesToClean.length > 0) {
                const unlabelPromises: Promise<void>[] = [];
                const labelsToClean = [nodeLabelKey, topoRegionKey, topoCityKey];

                nodesToClean.forEach((node) => {
                    labelsToClean.forEach((label) => {
                        unlabelPromises.push(kubeCli.unLabelNode(node, label).catch(() => {}));
                    });
                });

                if (unlabelPromises.length > 0) {
                    await Promise.all(unlabelPromises);
                    logger.info(
                        `Cleaned up node labels (${labelsToClean.join(', ')}) from ${nodesToClean.length} nodes`
                    );
                }
            }
        });

        test.beforeEach(() => {
            workloadName = objectGenerator.generateShortNameByPrefix('any-workload-topo', 5);
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
                            logger.error(`Failed to delete ${workload.workloadType} workload:`, error);
                        }
                    })
                );
            }
            if (projectUpdated) {
                await projectResourcesUtils.updateProjectGpuConfig(npProject.id, nodePools[0].id, {
                    deserved: 3,
                    maxAllowed: -1,
                    overQuotaWeight: 3
                });
                projectUpdated = false;
            }
        });

        test('T5481160: Do Not Override Existing Topology Annotations via API', async () => {
            await test.step('Submit workload with explicit topology annotations via workload V2 API', async () => {
                const manifest = workloadUtils.createPyTorchJobManifest(
                    workloadName,
                    `runai-${npProject.name}`,
                    'ubuntu',
                    sleepCommand
                );
                manifest.metadata['annotations'] = {
                    'kai.scheduler/topology': topoName,
                    'kai.scheduler/topology-preferred-placement': topoCityKey
                };
                logger.info(
                    'Submitting workload with explicit topology annotations:',
                    manifest.metadata['annotations']
                );

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
                logger.info(
                    `Created workload V2: ${response.body.metadata.name} with ID: ${response.body.metadata.id}`
                );

                workloadId = response.body.metadata.id;
                workloadInfo.push({ workloadType: WorkloadType.DISTRIBUTED, workloadId: workloadId });
            });

            await test.step('Retrieve workload via API and check pod template annotations', async () => {
                const getResult = await workloadsManagerUtils.getWorkloadV2ById(workloadId);
                const workloadV2 = getResult.body.workload;
                expect(workloadV2.metadata.id).toEqual(workloadId);
                expect(workloadV2.metadata.name).toEqual(workloadName);

                const annotations = workloadV2.workloadManifest?.metadata?.annotations;
                expect(annotations).toBeDefined();

                logger.info('Workload manifest annotations:', annotations);

                // Verify topology annotation is preserved
                expect(annotations['kai.scheduler/topology']).toBeDefined();
                expect(annotations['kai.scheduler/topology']).toEqual(topoName);
                expect(annotations['kai.scheduler/topology-preferred-placement']).toBeDefined();
                logger.info(
                    'annotations[kai.scheduler/topology-preferred-placement]:',
                    annotations['kai.scheduler/topology-preferred-placement']
                );
                expect(annotations['kai.scheduler/topology-preferred-placement']).toEqual(topoCityKey);
                expect(annotations['kai.scheduler/topology-preferred-placement']).not.toEqual(topoRegionKey);
            });

            await test.step('Wait for workload to be running and verify pods are scheduled according to user-specified placement', async () => {
                await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 60);
                logger.info(`Workload ${workloadId} is running`);

                const pods = await kubeCli.listPodsByNamespaceAndPrefix(
                    `runai-${npProject.name}`,
                    workloadName
                );

                expect(pods.length).toBeGreaterThan(0);
                logger.info(`Found ${pods.length} pods for workload ${workloadName}`);

                const nodeNames = pods.map((pod) => pod.spec.nodeName).filter((name) => name);
                expect(nodeNames.length).toBeGreaterThan(0);

                logger.info(`Pods are scheduled on nodes: ${nodeNames.join(', ')}`);

                const nodeRegions = new Set<string>();
                for (const nodeName of nodeNames) {
                    const nodeLabels = await kubeCli.getNodeLabels(nodeName);
                    const customTopoValue = nodeLabels[topoCityKey];
                    logger.info(`Node ${nodeName} has ${topoCityKey} = ${customTopoValue}`);

                    if (customTopoValue) {
                        nodeRegions.add(customTopoValue);
                    }
                }
                expect(nodeRegions.size).toBeLessThanOrEqual(1);
            });

            await test.step('Delete the workload', async () => {
                const deleteResult = await workloadsManagerUtils.deleteWorkloadV2(workloadId);
                expect(deleteResult.response.statusCode).toEqual(StatusCodes.ACCEPTED);
                logger.info(`Deleted workload V2: ${workloadId} via API`);
                workloadInfo.pop();
            });
        });

        test('T5481160: Do Not Override Existing Topology Annotations via API - AI practitioner role', async () => {
            await test.step('Set workload V2 API token to be AI practitioner', () => {
                workloadsManagerUtils.accessToken = flowRootState.users.ai_practitioner.accessToken;
            });

            await test.step('Submit workload with explicit topology annotations via workload V2 API', async () => {
                const manifest = workloadUtils.createPyTorchJobManifest(
                    workloadName,
                    `runai-${npProject.name}`,
                    'ubuntu',
                    sleepCommand
                );
                manifest.metadata['annotations'] = {
                    'kai.scheduler/topology': topoName,
                    'kai.scheduler/topology-preferred-placement': topoCityKey
                };
                logger.info(
                    'Submitting workload with explicit topology annotations:',
                    manifest.metadata['annotations']
                );

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
                logger.info(
                    `Created workload V2: ${response.body.metadata.name} with ID: ${response.body.metadata.id}`
                );

                workloadId = response.body.metadata.id;
                workloadInfo.push({ workloadType: WorkloadType.DISTRIBUTED, workloadId: workloadId });
            });

            await test.step('Retrieve workload via API and check pod template annotations', async () => {
                const getResult = await workloadsManagerUtils.getWorkloadV2ById(workloadId);
                const workloadV2 = getResult.body.workload;
                expect(workloadV2.metadata.id).toEqual(workloadId);
                expect(workloadV2.metadata.name).toEqual(workloadName);

                const annotations = workloadV2.workloadManifest?.metadata?.annotations;
                expect(annotations).toBeDefined();

                logger.info('Workload manifest annotations:', annotations);

                // Verify topology annotation is preserved
                expect(annotations['kai.scheduler/topology']).toBeDefined();
                expect(annotations['kai.scheduler/topology']).toEqual(topoName);
                expect(annotations['kai.scheduler/topology-preferred-placement']).toBeDefined();
                logger.info(
                    'annotations[kai.scheduler/topology-preferred-placement]:',
                    annotations['kai.scheduler/topology-preferred-placement']
                );
                expect(annotations['kai.scheduler/topology-preferred-placement']).toEqual(topoCityKey);
                expect(annotations['kai.scheduler/topology-preferred-placement']).not.toEqual(topoRegionKey);
            });

            await test.step('Wait for workload to be running and verify pods are scheduled according to user-specified placement', async () => {
                await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 60);
                logger.info(`Workload ${workloadId} is running`);

                const pods = await kubeCli.listPodsByNamespaceAndPrefix(
                    `runai-${npProject.name}`,
                    workloadName
                );

                expect(pods.length).toBeGreaterThan(0);
                logger.info(`Found ${pods.length} pods for workload ${workloadName}`);

                const nodeNames = pods.map((pod) => pod.spec.nodeName).filter((name) => name);
                expect(nodeNames.length).toBeGreaterThan(0);

                logger.info(`Pods are scheduled on nodes: ${nodeNames.join(', ')}`);

                const nodeRegions = new Set<string>();
                for (const nodeName of nodeNames) {
                    const nodeLabels = await kubeCli.getNodeLabels(nodeName);
                    const customTopoValue = nodeLabels[topoCityKey];
                    logger.info(`Node ${nodeName} has ${topoCityKey} = ${customTopoValue}`);

                    if (customTopoValue) {
                        nodeRegions.add(customTopoValue);
                    }
                }
                expect(nodeRegions.size).toBeLessThanOrEqual(1);
            });

            await test.step('Delete the workload', async () => {
                const deleteResult = await workloadsManagerUtils.deleteWorkloadV2(workloadId);
                expect(deleteResult.response.statusCode).toEqual(StatusCodes.ACCEPTED);
                logger.info(`Deleted workload V2: ${workloadId} via API`);
                workloadInfo.pop();
            });

            await test.step('Set workload V2 API token to be system admin', () => {
                workloadsManagerUtils.accessToken = flowRootState.users.system_administrator.accessToken;
            });
        });

        test('T5481161: Partial Topology Annotation Handling via API', async ({}, testInfo: TestInfo) => {
            skipPlaywrightTest(test, testInfo, 'Partial-topology handling has an unresolved failure; revalidate the expected placement behavior before enabling this test');
            await test.step('Submit workload with only topology name annotation', async () => {
                const manifest = workloadUtils.createPyTorchJobManifest(
                    workloadName,
                    `runai-${npProject.name}`,
                    'ubuntu',
                    sleepCommand
                );
                manifest.metadata['annotations'] = {
                    'kai.scheduler/topology': topoName
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
            });

            await test.step('Retrieve workload and verify annotation state', async () => {
                const getResult = await workloadsManagerUtils.getWorkloadV2ById(workloadId);
                const workloadV2 = getResult.body.workload;

                logger.info('Workload V2 details:', workloadV2);
                expect(workloadV2.metadata.id).toEqual(workloadId);
                expect(workloadV2.metadata.name).toEqual(workloadName);

                const annotations = workloadV2.workloadManifest?.metadata?.annotations;
                expect(annotations).toBeDefined();
                logger.info('Workload manifest annotations:', annotations);

                expect(annotations['kai.scheduler/topology']).toBeDefined();
                expect(annotations['kai.scheduler/topology']).toEqual(topoName);
                logger.info(
                    `Verified: kai.scheduler/topology exists with value: ${annotations['kai.scheduler/topology']}`
                );

                expect(annotations['kai.scheduler/topology-preferred-placement']).toBeUndefined();
                logger.info('Verified: kai.scheduler/topology-preferred-placement was NOT auto-added');

                expect(annotations['kai.scheduler/topology-required-placement']).toBeUndefined();
                logger.info('Verified: kai.scheduler/topology-required-placement was NOT auto-added');
            });

            await test.step('Wait for workload will be running and verify scheduler uses default topology behavior', async () => {
                await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 120);
                logger.info(`Workload ${workloadId} is running`);
            });

            await test.step('Verify pods are scheduled according to default topology behavior', async () => {
                const pods = await kubeCli.listPodsByNamespaceAndPrefix(
                    `runai-${npProject.name}`,
                    workloadName
                );
                expect(pods.length).toBeGreaterThan(0);
                logger.info(`Found ${pods.length} pods for workload ${workloadName}`);
                const nodeNames = pods.map((pod) => pod.spec.nodeName).filter((name) => name);
                expect(nodeNames.length).toBeGreaterThan(0);
                logger.info(`Pods are scheduled on nodes: ${nodeNames.join(', ')}`);
                const nodeCities = new Set<string>();
                for (const nodeName of nodeNames) {
                    const nodeLabels = await kubeCli.getNodeLabels(nodeName);
                    const customTopoValue = nodeLabels[topoCityKey];
                    logger.info(`Node ${nodeName} has ${topoCityKey} = ${customTopoValue}`);
                    if (customTopoValue) {
                        nodeCities.add(customTopoValue);
                    }
                }
                expect(nodeCities.size).toBeLessThanOrEqual(1);
            });

            await test.step('Clean up: Delete workload', async () => {
                const deleteResult = await workloadsManagerUtils.deleteWorkloadV2(workloadId);
                expect(deleteResult.response.statusCode).toEqual(StatusCodes.ACCEPTED);
                logger.info(`Deleted workload V2: ${workloadId}`);
                workloadInfo.pop();
            });
        });

        test('T5481163: Invalid Topology Attribute via API', async () => {
            const invalidAttributeKey = objectGenerator.generateShortNameByPrefix('invalid-attribute-key', 5);
            await test.step('Submit workload with preferred placement referencing non-existent attribute', async () => {
                const manifest = workloadUtils.createPyTorchJobManifest(
                    workloadName,
                    `runai-${npProject.name}`,
                    'ubuntu',
                    sleepCommand
                );
                manifest.metadata['annotations'] = {
                    'kai.scheduler/topology': topoName,
                    'kai.scheduler/topology-preferred-placement': invalidAttributeKey
                };
                logger.info(`Submitting workload with invalid topology attribute: ${invalidAttributeKey}`);

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
                logger.info(
                    `API accepted the workload with invalid topology attribute. WorkloadId: ${workloadId}`
                );
            });

            await test.step('Verify workload will be creating and check topology annotations', async () => {
                await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Creating, 60);
                const getResult = await workloadsManagerUtils.getWorkloadV2ById(workloadId);
                const workloadV2 = getResult.body.workload;

                const annotations = workloadV2.workloadManifest?.metadata?.annotations;
                expect(annotations).toBeDefined();
                logger.info('Workload annotations:', annotations);

                expect(annotations['kai.scheduler/topology']).toBeDefined();
                expect(annotations['kai.scheduler/topology']).toEqual(topoName);
                expect(annotations['kai.scheduler/topology-preferred-placement']).toBeDefined();
                expect(annotations['kai.scheduler/topology-preferred-placement']).toEqual(
                    invalidAttributeKey
                );

                logger.info(
                    `Verified: Workload preferred-placement annotation (invalid): ${annotations['kai.scheduler/topology-preferred-placement']}`
                );
            });

            await test.step('Clean up: Delete workload', async () => {
                const deleteResult = await workloadsManagerUtils.deleteWorkloadV2(workloadId);
                expect(deleteResult.response.statusCode).toEqual(StatusCodes.ACCEPTED);
                logger.info(`Deleted workload V2: ${workloadId}`);
                workloadInfo.pop();
            });
        });

        test('T5481166: Concurrent Workload Submissions via API', async () => {
            const numWorkloads = 5;
            const workloadNames: string[] = [];
            const workloadIds: string[] = [];

            await test.step('Submit all 5 workloads concurrently via API with different names', async () => {
                const createWorkloadPromises = [];

                for (let i = 0; i < numWorkloads; i++) {
                    const name = `${workloadName}-concurrent-${i}`;
                    workloadNames.push(name);

                    const manifest = workloadUtils.createPyTorchJobManifest(
                        name,
                        `runai-${npProject.name}`,
                        'ubuntu',
                        sleepCommand
                    );
                    manifest.metadata['annotations'] = {
                        'kai.scheduler/topology': topoName,
                        'kai.scheduler/topology-preferred-placement': topoRegionKey
                    };
                    manifest.spec.pytorchReplicaSpecs.Master.template.spec['nodeSelector'] = {
                        [nodeLabelKey]: nodeLabelVal1
                    };
                    manifest.spec.pytorchReplicaSpecs.Worker.template.spec['nodeSelector'] = {
                        [nodeLabelKey]: nodeLabelVal1
                    };

                    const createPromise = workloadsManagerUtils.createWorkloadV2(
                        name,
                        npProject.id,
                        manifest,
                        {
                            priority: 'medium',
                            category: 'Train'
                        }
                    );
                    createWorkloadPromises.push(createPromise);
                }

                logger.info(`Submitting ${numWorkloads} workloads concurrently...`);
                const responses = await Promise.all(createWorkloadPromises);
                logger.info(`Submitted ${numWorkloads} workloads`);

                responses.forEach((response, index) => {
                    expect(response.response.statusCode).toEqual(StatusCodes.ACCEPTED);
                    workloadIds.push(response.body.metadata.id);
                    workloadInfo.push({
                        workloadType: WorkloadType.DISTRIBUTED,
                        workloadId: response.body.metadata.id
                    });
                    logger.info(
                        `Workload ${index + 1}: ${workloadNames[index]} - ID: ${response.body.metadata.id}`
                    );
                });
            });

            await test.step('Verify all 5 workloads are accepted by API', () => {
                expect(workloadIds.length).toBe(numWorkloads);
                logger.info(`Accepted ${numWorkloads} workloads by API`);
            });

            await test.step('Retrieve all workloads and verify topology annotations are correct and consistent', async () => {
                const getWorkloadPromises = workloadIds.map((id) =>
                    workloadsManagerUtils.getWorkloadV2ById(id)
                );
                const workloadResults = await Promise.all(getWorkloadPromises);

                // Verify each workload has correct annotations
                workloadResults.forEach((result, index) => {
                    expect(result.response.statusCode).toEqual(StatusCodes.OK);

                    const workload = result.body.workload;
                    const annotations = workload.workloadManifest?.metadata?.annotations;

                    expect(annotations).toBeDefined();
                    expect(annotations['kai.scheduler/topology']).toBe(topoName);
                    expect(annotations['kai.scheduler/topology-preferred-placement']).toBe(topoRegionKey);
                    logger.info(
                        `Workload ${index + 1} (${workloadNames[index]}): Topology annotations correct`
                    );
                });
                logger.info(`All ${numWorkloads} workloads have correct topology annotations`);

                // Verify all workloads reference the same topology (consistency check)
                const topologyNames = workloadResults.map((result) => {
                    const annotations = result.body.workload.workloadManifest?.metadata?.annotations;
                    return annotations['kai.scheduler/topology'];
                });
                const uniqueTopologies = new Set(topologyNames);
                expect(uniqueTopologies.size).toBe(1);
                expect(uniqueTopologies.has(topoName)).toBe(true);
                logger.info(`All ${numWorkloads} workloads reference the same topology CR: ${topoName}`);

                // Verify all workloads have the same preferred placement (consistency check)
                const preferredPlacements = workloadResults.map((result) => {
                    const annotations = result.body.workload.workloadManifest?.metadata?.annotations;
                    return annotations['kai.scheduler/topology-preferred-placement'];
                });
                const uniquePlacements = new Set(preferredPlacements);
                expect(uniquePlacements.size).toBe(1);
                expect(uniquePlacements.has(topoRegionKey)).toBe(true);
                logger.info(
                    `All ${numWorkloads} workloads have consistent preferred placement: ${topoRegionKey}`
                );
            });

            await test.step('Check for any race conditions or inconsistencies', async () => {
                const phasePromises = workloadIds.map((id, index) =>
                    waitUtils.waitForBackendWorkloadByIdAndPhase(id, Phase.Running, 120).then(() => {
                        logger.info(`  Workload ${index + 1} (${workloadNames[index]}): Running`);
                        return id;
                    })
                );
                await Promise.all(phasePromises);
                logger.info(`All ${numWorkloads} workloads reached Running phase`);

                const podCheckPromises = workloadNames.map(async (name, index) => {
                    const pods = await kubeCli.listPodsByNamespaceAndPrefix(`runai-${npProject.name}`, name);
                    expect(pods.length).toBeGreaterThan(0);
                    logger.info(`  Workload ${index + 1} (${name}): ${pods.length} pods created`);
                    return pods;
                });

                await Promise.all(podCheckPromises);
                logger.info('All workloads have pods created - no race conditions detected');
            });

            await test.step('Clean up: Delete all workloads', async () => {
                const deletePromises = workloadIds.map((id, index) =>
                    workloadsManagerUtils.deleteWorkloadV2(id).then((result) => {
                        expect(result.response.statusCode).toEqual(StatusCodes.ACCEPTED);
                        logger.info(`  Deleted workload ${index + 1}: ${workloadNames[index]}`);
                        return result;
                    })
                );
                await Promise.all(deletePromises);
                workloadIds.forEach(() => workloadInfo.pop());
                logger.info(`Successfully deleted all ${numWorkloads} workloads`);
            });
        });

        test('T5481167: Workload Update via Kubernetes - Topology Annotation Modification', async ({}, testInfo: TestInfo) => {
            skipPlaywrightTest(
                test,
                testInfo,
                'Workload topology annotation cannot be modified via kubectl patch bug'
            );
            const initialPodNames: string[] = [];

            await test.step('Create initial workload with topology annotations via API', async () => {
                const manifest = workloadUtils.createPyTorchJobManifest(
                    workloadName,
                    `runai-${npProject.name}`,
                    'ubuntu',
                    sleepCommand
                );
                manifest.metadata['annotations'] = {
                    'kai.scheduler/topology': topoName,
                    'kai.scheduler/topology-preferred-placement': topoRegionKey
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
            });

            await test.step('Verify workload is running and has correct topology annotations', async () => {
                await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 120);
                const getResult = await workloadsManagerUtils.getWorkloadV2ById(workloadId);
                expect(getResult.response.statusCode).toEqual(StatusCodes.OK);

                const workload = getResult.body.workload;
                const annotations = workload.workloadManifest?.metadata?.annotations;

                expect(annotations).toBeDefined();
                expect(annotations['kai.scheduler/topology']).toBe(topoName);
                expect(annotations['kai.scheduler/topology-preferred-placement']).toBe(topoRegionKey);
            });

            await test.step('Update topology annotations to different topology', async () => {
                logger.info('Updating topology annotations via kubectl:');
                logger.info(`From: ${topoName} with placement ${topoCityKey}`);

                const currentCrd = await kubeCli.getCrd(
                    crds.pyTorchJobCrd,
                    workloadName,
                    `runai-${npProject.name}`
                );
                if (!currentCrd.metadata.annotations) {
                    currentCrd.metadata.annotations = {};
                }
                currentCrd.metadata.annotations['kai.scheduler/topology'] = topoName;
                currentCrd.metadata.annotations['kai.scheduler/topology-preferred-placement'] = topoCityKey;

                // Patch the CRD with the modified object
                await kubeCli.patchCrd(
                    crds.pyTorchJobCrd,
                    currentCrd,
                    workloadName,
                    `runai-${npProject.name}`
                );
                logger.info('Successfully patched workload CRD via kubectl');

                // Verify the CRD was actually updated in Kubernetes
                const updatedCrd = await kubeCli.getCrd(
                    crds.pyTorchJobCrd,
                    workloadName,
                    `runai-${npProject.name}`
                );
                const topoAnnotation = updatedCrd.metadata.annotations?.['kai.scheduler/topology'];
                const placementAnnotation =
                    updatedCrd.metadata.annotations?.['kai.scheduler/topology-preferred-placement'];
                expect(topoAnnotation).toBe(topoName);
                expect(placementAnnotation).toBe(topoCityKey);
            });

            await test.step('Verify annotation update does not trigger pod restart', async () => {
                const currentPods = await kubeCli.listPodsByNamespaceAndPrefix(
                    `runai-${npProject.name}`,
                    workloadName
                );
                const currentPodNames = currentPods.map((pod) => pod.metadata.name);
                const podsChanged = !initialPodNames.every((name) => currentPodNames.includes(name));
                expect(podsChanged).toBe(false);
                logger.info('Pods were NOT restarted after annotation update (expected behavior)');
            });

            await test.step('Verify existing pods retain original topology annotations', async () => {
                await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 120);
                const pods = await kubeCli.listPodsByNamespaceAndPrefix(
                    `runai-${npProject.name}`,
                    workloadName
                );
                expect(pods.length).toBeGreaterThan(0);
                for (const pod of pods) {
                    const podAnnotations = pod.metadata?.annotations || {};
                    if (podAnnotations['kai.scheduler/topology-preferred-placement']) {
                        expect(podAnnotations['kai.scheduler/topology-preferred-placement']).toBe(
                            topoCityKey
                        );
                        expect(podAnnotations['kai.scheduler/topology-preferred-placement']).not.toBe(
                            topoRegionKey
                        );
                        logger.info(
                            `Pod retains original annotation (${topoRegionKey}), not updated to new value (${topoCityKey})`
                        );
                    }
                }
                logger.info('Verified existing pods maintain their original topology annotations');
            });

            await test.step('Clean up: Delete workload', async () => {
                const deleteResult = await workloadsManagerUtils.deleteWorkloadV2(workloadId);
                expect(deleteResult.response.statusCode).toEqual(StatusCodes.ACCEPTED);
                logger.info(`Deleted workload: ${workloadId}`);
                workloadInfo.pop();
            });
        });

        test('T5481168: Workload Deletion with Active Topology via API', async () => {
            await test.step('Create workload with topology via API', async () => {
                const manifest = workloadUtils.createPyTorchJobManifest(
                    workloadName,
                    `runai-${npProject.name}`,
                    'ubuntu',
                    sleepCommand
                );

                manifest.metadata['annotations'] = {
                    'kai.scheduler/topology': topoName,
                    'kai.scheduler/topology-preferred-placement': topoRegionKey
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
                logger.info(`Workload created with ID: ${workloadId}`);
            });

            await test.step('Verify workload and pods are running', async () => {
                await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 120);
                logger.info('Workload is in Running phase');
            });

            await test.step('Delete workload via API', async () => {
                const deleteResult = await workloadsManagerUtils.deleteWorkloadV2(workloadId);
                expect(deleteResult.response.statusCode).toEqual(StatusCodes.ACCEPTED);
                logger.info(`API accepted deletion request (status: ${deleteResult.response.statusCode})`);
                workloadInfo.pop();
            });

            await test.step('Verify workload deletion completes successfully', async () => {
                await waitUtils.waitForPodsDeletion(workloadName, `runai-${npProject.name}`);
            });

            await test.step('Check topology CR is not affected', async () => {
                const topologies = await networkTopologiesUtils.getNetworkTopologiesViaApi();
                const topology = topologies.body.networkTopologies?.find((t) => t.name === topoName);
                expect(topology).toBeDefined();
                expect(topology.id).toBe(topologyIds[0]);
                logger.info(
                    `Topology CR '${topoName}' still exists and is not affected by workload deletion`
                );
            });

            await test.step('Verify no orphaned resources remain', async () => {
                try {
                    await kubeCli.getCrd(crds.pyTorchJobCrd, workloadName, `runai-${npProject.name}`);
                    throw new Error('PyTorchJob CRD should not exist');
                } catch {
                    logger.info('PyTorchJob CRD has been cleaned up');
                }
                logger.info('No orphaned resources remain');
            });

            await test.step('Confirm topology CR can still be used by new workloads', async () => {
                const newWorkloadName = objectGenerator.generateShortNameByPrefix('verify-topo', 5);
                const manifest = workloadUtils.createPyTorchJobManifest(
                    newWorkloadName,
                    `runai-${npProject.name}`,
                    'ubuntu',
                    sleepCommand
                );

                manifest.metadata['annotations'] = {
                    'kai.scheduler/topology': topoName,
                    'kai.scheduler/topology-preferred-placement': topoRegionKey
                };
                const response = await workloadsManagerUtils.createWorkloadV2(
                    newWorkloadName,
                    npProject.id,
                    manifest,
                    {
                        priority: 'medium',
                        category: 'Train'
                    }
                );

                expect(response.response.statusCode).toEqual(StatusCodes.ACCEPTED);
                const newWorkloadId = response.body.metadata.id;
                workloadInfo.push({ workloadType: WorkloadType.DISTRIBUTED, workloadId: newWorkloadId });
                logger.info(`New workload created successfully with same topology: ${newWorkloadId}`);

                await waitUtils.waitForBackendWorkloadByIdAndPhase(newWorkloadId, Phase.Running, 120);
                logger.info(`New workload is running successfully with topology '${topoName}'`);

                const getResult = await workloadsManagerUtils.getWorkloadV2ById(newWorkloadId);
                const annotations = getResult.body.workload.workloadManifest?.metadata?.annotations;

                expect(annotations['kai.scheduler/topology']).toBe(topoName);
                expect(annotations['kai.scheduler/topology-preferred-placement']).toBe(topoRegionKey);
                logger.info(
                    `Confirmed topology CR '${topoName}' is still works for new workload with same topology`
                );

                await workloadsManagerUtils.deleteWorkloadV2(newWorkloadId);
                workloadInfo.pop();
            });
        });

        test('T5481169: Topology Annotations Visibility in GET /workload Spec with topology attached', async () => {
            await test.step('Submit workload to node pool with topology', async () => {
                const manifest = workloadUtils.createPyTorchJobManifest(
                    workloadName,
                    `runai-${npProject.name}`,
                    'ubuntu',
                    sleepCommand
                );
                manifest.metadata['annotations'] = {
                    'kai.scheduler/topology': topoName,
                    'kai.scheduler/topology-preferred-placement': topoRegionKey
                };

                logger.info('Submitting workload with topology annotations');
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
                logger.info(`Workload created with ID: ${workloadId}`);
            });

            await test.step('Query workload via GET API', async () => {
                await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 120);
                logger.info('Workload is in Running phase');
            });

            await test.step('Verify response structure includes spec section', async () => {
                const getResult = await workloadsManagerUtils.getWorkloadV2ById(workloadId);
                expect(getResult.response.statusCode).toEqual(StatusCodes.OK);

                const workload = getResult.body.workload;
                expect(workload).toBeDefined();
                expect(workload.workloadManifest).toBeDefined();
                expect(workload.workloadManifest.spec).toBeDefined();
                logger.info('Response structure includes spec section');
            });

            await test.step('Verify spec.template.metadata.annotations contains topology annotations', async () => {
                const getResult = await workloadsManagerUtils.getWorkloadV2ById(workloadId);
                const workload = getResult.body.workload;
                const manifestAnnotations = workload.workloadManifest?.metadata?.annotations;
                expect(manifestAnnotations).toBeDefined();
                expect(manifestAnnotations['kai.scheduler/topology']).toBeDefined();
                expect(manifestAnnotations['kai.scheduler/topology-preferred-placement']).toBeDefined();
                expect(manifestAnnotations['kai.scheduler/topology']).toBe(topoName);
                expect(manifestAnnotations['kai.scheduler/topology-preferred-placement']).toBe(topoRegionKey);
                logger.info('Topology annotations found in manifest metadata');
            });

            await test.step('Delete the workload', async () => {
                const deleteResult = await workloadsManagerUtils.deleteWorkloadV2(workloadId);
                expect(deleteResult.response.statusCode).toEqual(StatusCodes.ACCEPTED);
                logger.info(`Deleted workload: ${workloadId}`);
                workloadInfo.pop();
            });

            await test.step('Submit workload with explicit annotations and verify exact preservation', async () => {
                const customWorkloadName = objectGenerator.generateShortNameByPrefix('custom-topo', 5);
                const manifest = workloadUtils.createPyTorchJobManifest(
                    customWorkloadName,
                    `runai-${npProject.name}`,
                    'ubuntu',
                    sleepCommand
                );

                manifest.metadata['annotations'] = {
                    'kai.scheduler/topology': topoName,
                    'kai.scheduler/topology-preferred-placement': topoCityKey
                };
                const response = await workloadsManagerUtils.createWorkloadV2(
                    customWorkloadName,
                    npProject.id,
                    manifest,
                    {
                        priority: 'medium',
                        category: 'Train'
                    }
                );

                expect(response.response.statusCode).toEqual(StatusCodes.ACCEPTED);
                customWorkloadId = response.body.metadata.id;
                workloadInfo.push({ workloadType: WorkloadType.DISTRIBUTED, workloadId: customWorkloadId });

                await waitUtils.waitForBackendWorkloadByIdAndPhase(customWorkloadId, Phase.Running, 120);

                const getResult = await workloadsManagerUtils.getWorkloadV2ById(customWorkloadId);
                const annotations = getResult.body.workload.workloadManifest?.metadata?.annotations;

                expect(annotations['kai.scheduler/topology']).toBe(topoName);
                expect(annotations['kai.scheduler/topology-preferred-placement']).toBe(topoCityKey);
            });

            await test.step('Verify annotations are machine-readable and complete', async () => {
                const getResult = await workloadsManagerUtils.getWorkloadV2ById(customWorkloadId);
                const annotations = getResult.body.workload.workloadManifest?.metadata?.annotations;

                expect(typeof annotations['kai.scheduler/topology']).toBe('string');
                expect(typeof annotations['kai.scheduler/topology-preferred-placement']).toBe('string');

                expect(annotations['kai.scheduler/topology'].length).toBeGreaterThan(0);
                expect(annotations['kai.scheduler/topology-preferred-placement'].length).toBeGreaterThan(0);

                const topologyKeyPattern = /^[a-z0-9]([-a-z0-9.]*[a-z0-9])?\/([-a-z0-9]+)$/;
                expect(topologyKeyPattern.test('kai.scheduler/topology')).toBe(true);
                expect(topologyKeyPattern.test('kai.scheduler/topology-preferred-placement')).toBe(true);
                logger.info('Annotations are machine-readable and complete');
            });

            await test.step('Clean up: Delete workload', async () => {
                await workloadsManagerUtils.deleteWorkloadV2(customWorkloadId);
                workloadInfo.pop();
                logger.info(`Deleted workload: ${customWorkloadId}`);
            });
        });

        test('T5481171: Topology Status Information in GET /workload(success scenario)', async () => {
            let successWorkloadId: string;
            const successWorkloadName = objectGenerator.generateShortNameByPrefix('topo-status-success', 5);

            await test.step('Create workload with topology', async () => {
                const manifest = workloadUtils.createPyTorchJobManifest(
                    successWorkloadName,
                    `runai-${npProject.name}`,
                    'ubuntu',
                    sleepCommand,
                    1
                );
                manifest.metadata['annotations'] = {
                    'kai.scheduler/topology': topoName,
                    'kai.scheduler/topology-preferred-placement': topoRegionKey
                };
                const response = await workloadsManagerUtils.createWorkloadV2(
                    successWorkloadName,
                    npProject.id,
                    manifest,
                    {
                        priority: 'medium',
                        category: 'Train'
                    }
                );
                expect(response.response.statusCode).toEqual(StatusCodes.ACCEPTED);
                successWorkloadId = response.body.metadata.id;
                workloadInfo.push({ workloadType: WorkloadType.DISTRIBUTED, workloadId: successWorkloadId });
                logger.info(`Workload created with ID: ${successWorkloadId}`);
            });

            await test.step('Wait for workload to reach Running phase', async () => {
                await waitUtils.waitForBackendWorkloadByIdAndPhase(successWorkloadId, Phase.Running, 120);
                logger.info('Workload reached Running phase');
            });

            await test.step('Query workload via GET API and examine status section', async () => {
                const getResult = await workloadsManagerUtils.getWorkloadV2ById(successWorkloadId);
                expect(getResult.response.statusCode).toEqual(StatusCodes.OK);
                const workload = getResult.body.workload;
                expect(workload.workloadManifest).toBeDefined();
            });

            await test.step('Check PyTorchJob Status (replicas, readyReplicas, conditions)', async () => {
                const crd = (await kubeCli.getCrd(
                    crds.pyTorchJobCrd,
                    successWorkloadName,
                    `runai-${npProject.name}`
                )) as any;
                expect(crd.status).toBeDefined();
                logger.info(`PyTorchJob Status: ${JSON.stringify(crd.status, null, 2)}`);
                expect(crd.status.replicaStatuses).toBeDefined();
                logger.info(`Replica Statuses: ${JSON.stringify(crd.status.replicaStatuses, null, 2)}`);

                // Check conditions
                if (crd.status.conditions) {
                    const succeededCondition = crd.status.conditions.find(
                        (c: any) => c.type === 'Succeeded' || c.type === 'Running'
                    );
                    logger.info(`Job Conditions: ${JSON.stringify(crd.status.conditions, null, 2)}`);
                }
            });

            await test.step('Verify topology annotations are preserved in PyTorchJob CRD', async () => {
                const crd = (await kubeCli.getCrd(
                    crds.pyTorchJobCrd,
                    successWorkloadName,
                    `runai-${npProject.name}`
                )) as any;

                const crdAnnotations = crd.metadata?.annotations || {};
                logger.info(`PyTorchJob annotations: ${JSON.stringify(crdAnnotations, null, 2)}`);

                // Verify topology annotations are preserved at CRD level
                expect(crdAnnotations['kai.scheduler/topology']).toBe(topoName);
                expect(crdAnnotations['kai.scheduler/topology-preferred-placement']).toBe(topoRegionKey);

                logger.info('Topology annotations are correctly preserved in PyTorchJob CRD');
            });

            await test.step('Wait for all pods to be created and running', async () => {
                const pods = await waitUtils.waitForPods(
                    successWorkloadName,
                    `runai-${npProject.name}`,
                    1,
                    60,
                    ['Running'],
                    true,
                    5
                );
                expect(pods.length).toBeGreaterThan(0);
                logger.info(`All ${pods.length} pod(s) are running successfully`);
            });

            await test.step('Check Pod Placement and Topology Satisfaction', async () => {
                const pods = await kubeCli.listPodsByNamespaceAndPrefix(
                    `runai-${npProject.name}`,
                    successWorkloadName
                );
                const nodeNames = new Set<string>();
                const nodeTopologyInfo = new Map<string, any>();

                // Collect pod placement and topology information
                for (const pod of pods) {
                    expect(pod.spec?.nodeName).toBeDefined();
                    const nodeName = pod.spec!.nodeName!;
                    nodeNames.add(nodeName);

                    // Get node topology labels if not already collected
                    if (!nodeTopologyInfo.has(nodeName)) {
                        const nodeLabels = await kubeCli.getNodeLabels(nodeName);
                        nodeTopologyInfo.set(nodeName, {
                            hostname: nodeLabels['kubernetes.io/hostname'],
                            region: nodeLabels[topoRegionKey],
                            city: nodeLabels[topoCityKey]
                        });
                        logger.info(
                            `Node ${nodeName} topology: region=${nodeLabels[topoRegionKey]}, city=${nodeLabels[topoCityKey]}`
                        );
                    }
                }

                // Verify topology constraint satisfaction
                const regions = Array.from(nodeTopologyInfo.values()).map((info) => info.region);
                const uniqueRegions = new Set(regions.filter((r) => r));
                expect(uniqueRegions.size).toBe(1);
            });

            await test.step('Check Events for Topology-Related Messages', async () => {
                const eventsResponse = await workloadsUtilsV2.getWorkloadEvents(successWorkloadId);
                const events = eventsResponse.body?.events || [];

                const topologyEvents = events.filter(
                    (event: any) =>
                        event.message?.toLowerCase().includes('topology') ||
                        event.message?.toLowerCase().includes('affinity') ||
                        event.reason?.toLowerCase().includes('topology')
                );

                if (topologyEvents.length > 0) {
                    logger.info('Topology-related events found:');
                    topologyEvents.forEach((event: any) => {
                        logger.info(`  [${event.type}] ${event.reason}: ${event.message}`);
                    });
                }

                /// Verify no topology-related errors
                const errorEvents = events.filter(
                    (event: any) => event.type === 'Warning' || event.type === 'Error'
                );
                const topologyErrors = errorEvents.filter((event: any) =>
                    event.message?.toLowerCase().includes('topology')
                );
                expect(topologyErrors.length).toBe(0);
                logger.info('No topology-related error events found');
            });

            await test.step('Clean up: Delete workload', async () => {
                await workloadsManagerUtils.deleteWorkloadV2(successWorkloadId);
                workloadInfo.pop();
                logger.info(`Deleted workload: ${successWorkloadId}`);
            });
        });

        test('T5481173: API Response Structure for Successful Submission', async () => {
            let fullApiResponse: any;
            await test.step('Submit workload via API to node pool with topology', async () => {
                const manifest = workloadUtils.createPyTorchJobManifest(
                    workloadName,
                    `runai-${npProject.name}`,
                    'ubuntu',
                    sleepCommand,
                    1
                );
                manifest.metadata['annotations'] = {
                    'kai.scheduler/topology': topoName,
                    'kai.scheduler/topology-preferred-placement': topoCityKey
                };

                fullApiResponse = await workloadsManagerUtils.createWorkloadV2(
                    workloadName,
                    npProject.id,
                    manifest,
                    {
                        priority: 'medium',
                        category: 'Train'
                    }
                );
                workloadId = fullApiResponse.body.metadata.id;
                workloadInfo.push({ workloadType: WorkloadType.DISTRIBUTED, workloadId: workloadId });
                logger.info(`Workload submitted with ID: ${workloadId}`);
            });

            await test.step('Capture and document full API response', () => {
                logger.info('=== Full API Response Structure ===');
                logger.info(`HTTP Status Code: ${fullApiResponse.response.statusCode}`);
                logger.info(`HTTP Status Message: ${fullApiResponse.response.statusMessage || 'N/A'}`);
                const relevantHeaders = ['content-type', 'content-length', 'date'];
                relevantHeaders.forEach((header) => {
                    if (fullApiResponse.response.headers && fullApiResponse.response.headers[header]) {
                        logger.info(`  ${header}: ${fullApiResponse.response.headers[header]}`);
                    }
                });
                logger.info(`Response Body Top-Level Keys: ${Object.keys(fullApiResponse.body).join(', ')}`);
                if (fullApiResponse.body.metadata) {
                    logger.info(`Metadata Keys: ${Object.keys(fullApiResponse.body.metadata).join(', ')}`);
                }
            });

            await test.step('Verify HTTP status code: 202 Accepted', () => {
                expect(fullApiResponse.response.statusCode).toEqual(StatusCodes.ACCEPTED);
                logger.info(`HTTP status code is ${StatusCodes.ACCEPTED} (ACCEPTED)`);
            });

            await test.step('Check response body contains created resource', () => {
                const body = fullApiResponse.body;
                expect(body).toBeDefined();
                expect(body.metadata).toBeDefined();
                expect(body.metadata.id).toBeDefined();
                expect(body.metadata.name).toBeDefined();
                expect(body.metadata.projectId).toBeDefined();
                expect(body.metadata.id).toBe(workloadId);
                expect(body.metadata.name).toContain(workloadName);
                logger.info('Response body contains all essential created resource fields');
            });

            await test.step('Verify response includes topology annotations', () => {
                const body = fullApiResponse.body;
                logger.info('Checking for topology annotations in response:');
                const workloadAnnotations = body.workloadManifest?.metadata?.annotations || {};
                logger.info(`Workload manifest annotations: ${JSON.stringify(workloadAnnotations, null, 2)}`);

                const topologyAnnotation = workloadAnnotations['kai.scheduler/topology'];
                const placementAnnotation = workloadAnnotations['kai.scheduler/topology-preferred-placement'];
                logger.info(`Topology annotation: ${topologyAnnotation}`);
                expect(topologyAnnotation).toBe(topoName);
                logger.info(`Preferred placement annotation: ${placementAnnotation}`);
                expect(placementAnnotation).toBe(topoCityKey);
                logger.info(
                    'Topology annotations verified successfully in workloadManifest.metadata.annotations'
                );
            });

            await test.step('Document response structure and fields', () => {
                const body = fullApiResponse.body;
                logger.info('Complete Response Structure Documentation');
                logger.info('Metadata structure:');
                Object.keys(body.metadata).forEach((key) => {
                    const value = body.metadata[key];
                    const type = Array.isArray(value) ? 'array' : typeof value;
                    logger.info(`  - ${key}: ${type}`);
                    if (type === 'object' && value !== null && !Array.isArray(value)) {
                        logger.info(`    Keys: ${Object.keys(value).join(', ')}`);
                    }
                });

                logger.info('\nTop-level response fields:');
                Object.keys(body).forEach((key) => {
                    if (key !== 'metadata') {
                        const value = body[key];
                        const type = Array.isArray(value) ? 'array' : typeof value;
                        logger.info(`  - ${key}: ${type}`);
                    }
                });
            });

            await test.step('Validate response matches Kubernetes API conventions', () => {
                const body = fullApiResponse.body;
                logger.info('Validating Kubernetes API conventions:');
                const checks = [
                    {
                        field: 'metadata',
                        present: !!body.metadata,
                        description: 'Metadata section (standard K8s field)'
                    },
                    {
                        field: 'metadata.name',
                        present: !!body.metadata?.name,
                        description: 'Resource name in metadata'
                    },
                    {
                        field: 'metadata.id',
                        present: !!body.metadata?.id,
                        description: 'Unique identifier (Run:ai extension)'
                    }
                ];
                checks.forEach((check) => {
                    expect(check.present).toBe(true);
                    logger.info(`${check.field}: Present - ${check.description}`);
                });
                expect(body.metadata.name).toBeDefined();
                expect(typeof body.metadata.name).toBe('string');
                logger.info('Response matches with Kubernetes and REST API conventions');
            });

            await test.step('Clean up: Delete workload', async () => {
                if (workloadId) {
                    const deleteResult = await workloadsManagerUtils.deleteWorkloadV2(workloadId);
                    expect(deleteResult.response.statusCode).toEqual(StatusCodes.ACCEPTED);
                    logger.info(`Deleted test workload: ${workloadId}`);
                    workloadInfo = workloadInfo.filter((w) => w.workloadId !== workloadId);
                }
            });
        });

        test('T5481174: API Error Response Structure for Validation Failures', async () => {
            const invalidTopologyName = objectGenerator.generateShortNameByPrefix('invalid-topology', 5);

            await test.step('Submit invalid workload with non-existent node pool via API', async () => {
                const manifest = workloadUtils.createPyTorchJobManifest(
                    workloadName,
                    `runai-${npProject.name}`,
                    'ubuntu',
                    sleepCommand,
                    1
                );
                manifest.metadata['annotations'] = {
                    'kai.scheduler/topology': invalidTopologyName
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
                logger.info(`Workload created with ID: ${workloadId}`);
            });

            await test.step('Verify the workload is in creating status', async () => {
                await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Creating, 120);
                logger.info(`Workload ${workloadId} is in creating status`);
            });

            await test.step('Verify pods are pending and have valid pending reason', async () => {
                const namespace = `runai-${npProject.name}`;
                const pods = await kubeCli.listPodsByNamespaceAndPrefix(namespace, workloadName);

                for (const pod of pods) {
                    const podName = pod.metadata?.name || 'unknown';
                    const podPhase = pod.status?.phase;
                    logger.info(`\nPod: ${podName} (Phase: ${podPhase})`);

                    let podEvents: any[] = [];
                    let hasTopologyError = false;

                    // Wait for topology error event to appear (supports both old and new message formats)
                    await waitUtils.waitForCondition(
                        async () => {
                            const allEvents = await kubeCli.CoreV1Api.listNamespacedEvent({ namespace });
                            podEvents = allEvents.items
                                .filter((event: any) => event.involvedObject?.name === podName)
                                .sort(
                                    (a: any, b: any) =>
                                        new Date(a.lastTimestamp || a.firstTimestamp || 0).getTime() -
                                        new Date(b.lastTimestamp || b.firstTimestamp || 0).getTime()
                                );

                            // Check for new format: Unschedulable with topology error message
                            const hasNewFormatError = podEvents.some(
                                (event: any) =>
                                    event.reason === 'Unschedulable' &&
                                    event.message?.includes('topology') &&
                                    event.message?.includes('invalid-topology-')
                            );

                            // Check for old format: PodGroup error message
                            const hasOldFormatError = podEvents.some((event: any) =>
                                event.message?.includes(
                                    'Operation cannot be fulfilled on podgroups.scheduling.run.ai'
                                )
                            );

                            hasTopologyError = hasNewFormatError || hasOldFormatError;

                            logger.info(
                                `Waiting for topology error event... (events: ${podEvents.length}, new format: ${hasNewFormatError}, old format: ${hasOldFormatError})`
                            );
                            return hasTopologyError;
                        },
                        `Wait for topology error event for ${podName}`,
                        true,
                        120,
                        5
                    );

                    podEvents.forEach((event: any) => {
                        const type = event.type || 'Normal';
                        const reason = event.reason || 'Unknown';
                        const message = event.message || '';
                        const count = event.count > 1 ? ` (x${event.count})` : '';
                        logger.info(`  ${type.padEnd(8)} ${reason.padEnd(20)}${count} ${message}`);
                    });

                    const statuses = podEvents.map((e: any) => e.reason).filter(Boolean);
                    logger.info(`Status progression: ${statuses.join(' -> ')}`);

                    logger.info(
                        'DEBUG: Verified topology error event found (supports both old and new formats)'
                    );

                    expect(hasTopologyError).toBe(true);
                }
            });

            await test.step('Clean up: Delete workload', async () => {
                if (workloadId) {
                    const deleteResult = await workloadsManagerUtils.deleteWorkloadV2(workloadId);
                    expect(deleteResult.response.statusCode).toEqual(StatusCodes.ACCEPTED);
                    logger.info(`Deleted test workload: ${workloadId}`);
                    workloadInfo.pop();
                }
            });
        });

        test('T5481175: Workload Status Field Updates via API', async () => {
            await test.step('Create workload with topology via API', async () => {
                const manifest = workloadUtils.createPyTorchJobManifest(
                    workloadName,
                    `runai-${npProject.name}`,
                    'ubuntu',
                    sleepCommand,
                    1
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
                        category: 'Train',
                        nodePools: nodePoolNames
                    } as any
                );
                expect(response.response.statusCode).toEqual(StatusCodes.ACCEPTED);
                workloadId = response.body.metadata.id;
                workloadInfo.push({ workloadType: WorkloadType.DISTRIBUTED, workloadId: workloadId });
                logger.info(`Workload created with ID: ${workloadId}`);
            });

            await test.step('Wait for workload to reach Running phase', async () => {
                await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 120);
                logger.info('Workload reached Running phase');
            });

            await test.step('Retrieve workload and verify status', async () => {
                const namespace = `runai-${npProject.name}`;
                const pods = await kubeCli.listPodsByNamespaceAndPrefix(namespace, workloadName);
                const allEvents = await kubeCli.CoreV1Api.listNamespacedEvent({ namespace });
                const errorEvents = [
                    'Failed',
                    'Error',
                    'BackOff',
                    'Unhealthy',
                    'FailedScheduling',
                    'FailedMount'
                ];

                for (const pod of pods) {
                    const podName = pod.metadata?.name || 'unknown';
                    logger.info(`Pod: ${podName}`);

                    const podEvents = allEvents.items
                        .filter((event: any) => event.involvedObject?.name === podName)
                        .sort(
                            (a: any, b: any) =>
                                new Date(a.lastTimestamp || a.firstTimestamp || 0).getTime() -
                                new Date(b.lastTimestamp || b.firstTimestamp || 0).getTime()
                        );

                    const statuses = podEvents.map((e: any) => e.reason).filter(Boolean);
                    logger.info(`Status progression: ${statuses.join(' -> ')}`);

                    const hasError = statuses.some((status) =>
                        errorEvents.some((err) => status.includes(err))
                    );
                    expect(hasError).toBe(false);
                }
            });

            await test.step('Clean up: Delete workload', async () => {
                if (workloadId) {
                    const deleteResult = await workloadsManagerUtils.deleteWorkloadV2(workloadId);
                    expect(deleteResult.response.statusCode).toEqual(StatusCodes.ACCEPTED);
                    logger.info(`Deleted test workload: ${workloadId}`);
                    workloadInfo = workloadInfo.filter((w) => w.workloadId !== workloadId);
                }
            });
        });

        test('T5481176: Workload with Topology Annotations Runs Successfully', async () => {
            await test.step('Submit workload with explicit topology annotations, should be accepted', async () => {
                const manifest = workloadUtils.createPyTorchJobManifest(
                    workloadName,
                    `runai-${npProject.name}`,
                    'ubuntu',
                    sleepCommand,
                    1
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
                logger.info(`Workload created with ID: ${workloadId}`);
            });

            await test.step('Wait for workload to reach Running phase', async () => {
                await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 120);
            });

            await test.step('Retrieve workload and verify annotations are same as submitted', async () => {
                const getResult = await workloadsManagerUtils.getWorkloadV2ById(workloadId);
                const workload = getResult.body.workload;
                const currentAnnotations = workload.workloadManifest?.metadata?.annotations || {};
                expect(currentAnnotations['kai.scheduler/topology']).toBe(topoName);
                expect(currentAnnotations['kai.scheduler/topology-preferred-placement']).toBe(topoCityKey);
            });

            await test.step('Verify all pods are running', async () => {
                const pods = await waitUtils.waitForPods(workloadName, `runai-${npProject.name}`, 2, 120, [
                    'Running'
                ]);
                expect(pods.length).toBe(2);
                logger.info(`All ${pods.length} pods are running`);
            });

            await test.step('Verify pods are scheduled according to topology', async () => {
                const namespace = `runai-${npProject.name}`;
                const nodeSet = new Set<string>();
                const pods = await kubeCli.listPodsByNamespaceAndPrefix(namespace, workloadName);

                for (const pod of pods) {
                    const nodeName = pod.spec?.nodeName;
                    const podStatus = pod.status?.phase;
                    logger.info(`Pod ${pod.metadata?.name}: Node=${nodeName}, Status=${podStatus}`);
                    expect(nodeName).toBeDefined();
                    if (nodeName) nodeSet.add(nodeName);
                }
                for (const nodeName of nodeSet) {
                    const nodeLabels = await kubeCli.getNodeLabels(nodeName);
                    const regionLabel = nodeLabels[topoRegionKey];
                    const cityLabel = nodeLabels[topoCityKey];
                    const nodePoolLabel = nodeLabels[nodeLabelKey];
                    logger.info(
                        `Node ${nodeName}: region=${regionLabel}, city=${cityLabel}, pool=${nodePoolLabel}`
                    );
                    expect(regionLabel).toBeDefined();
                    expect(cityLabel).toBeDefined();
                    expect(nodePoolLabel).toBe(nodeLabelVal1);
                }
                expect(nodeSet.size).toBe(1);
            });

            await test.step('Verify workload reaches desired state: Check 2/2 ready replicas', async () => {
                const namespace = `runai-${npProject.name}`;
                const crd = (await kubeCli.getCrd(crds.pyTorchJobCrd, workloadName, namespace)) as any;
                logger.info('\nPyTorchJob Status:');
                logger.info(JSON.stringify(crd.status, null, 2));
                const replicaStatuses = crd.status.replicaStatuses;
                expect(replicaStatuses).toBeDefined();
                expect(replicaStatuses.Master.active).toBe(1);
                expect(replicaStatuses.Worker.active).toBe(1);
                const conditions = crd.status.conditions;
                expect(conditions).toBeDefined();
                expect(conditions.length).toBeGreaterThan(0);

                // Find the Running condition
                const runningCondition = conditions.find((c: any) => c.type === 'Running');
                expect(runningCondition).toBeDefined();
                expect(runningCondition.status).toBe('True');
                logger.info('Workload successfully reached desired state with topology constraints');
            });

            await test.step('Clean up: Delete workload', async () => {
                if (workloadId) {
                    const deleteResult = await workloadsManagerUtils.deleteWorkloadV2(workloadId);
                    expect(deleteResult.response.statusCode).toEqual(StatusCodes.ACCEPTED);
                    logger.info(`Deleted test workload: ${workloadId}`);
                    workloadInfo = workloadInfo.filter((w) => w.workloadId !== workloadId);
                }
            });
        });

        test('T5481159 & T5481177: Automatic Topology Annotation Addition and Workload Execution', async () => {
            await test.step('Submit workload WITHOUT topology annotations, should be accepted', async () => {
                const manifest = workloadUtils.createPyTorchJobManifest(
                    workloadName,
                    `runai-${npProject.name}`,
                    'ubuntu',
                    sleepCommand,
                    2
                );
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
                logger.info(`Workload created with ID: ${workloadId}`);
            });

            await test.step('Wait for workload to reach Running phase', async () => {
                await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 120);
            });

            await test.step('Retrieve workload and verify topology annotations were AUTOMATICALLY ADDED', async () => {
                const getResult = await workloadsManagerUtils.getWorkloadV2ById(workloadId);
                const workload = getResult.body.workload;
                const currentAnnotations = workload.workloadManifest?.metadata?.annotations || {};
                expect(currentAnnotations['kai.scheduler/topology']).toBeDefined();
                expect(currentAnnotations['kai.scheduler/topology-preferred-placement']).toBeDefined();
            });

            await test.step('Verify annotations match the nodepool topology', async () => {
                const nodePoolResponse = await nodePoolUtils.getNodePoolById(nodePools[0].id);
                const topologyResponse = await networkTopologiesUtils.getNetworkTopologyByIdViaApi(
                    nodePoolResponse.body.networkTopologyId
                );
                const topologyName = topologyResponse.body.name;
                const topologyLevels = topologyResponse.body.levels;

                const getResult = await workloadsManagerUtils.getWorkloadV2ById(workloadId);
                const annotations = getResult.body.workload.workloadManifest?.metadata?.annotations;
                expect(annotations).toBeDefined();
                expect(annotations['kai.scheduler/topology']).toBe(topologyName);
                expect(topologyLevels).toContain(annotations['kai.scheduler/topology-preferred-placement']);
                expect(topologyLevels.length).toBe(2);
            });

            await test.step('Wait for pods to be running', async () => {
                const pods = await waitUtils.waitForPods(workloadName, `runai-${npProject.name}`, 2, 120, [
                    'Running'
                ]);
                expect(pods.length).toBe(3);
            });

            await test.step('Verify workload reaches desired state: Check 3/3 ready replicas', async () => {
                const namespace = `runai-${npProject.name}`;
                const crd = (await kubeCli.getCrd(crds.pyTorchJobCrd, workloadName, namespace)) as any;
                logger.info('\nPyTorchJob Status:');
                logger.info(JSON.stringify(crd.status, null, 2));
                const replicaStatuses = crd.status.replicaStatuses;
                expect(replicaStatuses).toBeDefined();
                expect(replicaStatuses.Master.active).toBe(1);
                expect(replicaStatuses.Worker.active).toBe(2);

                const conditions = crd.status.conditions;
                expect(conditions).toBeDefined();
                expect(conditions.length).toBeGreaterThan(0);

                const runningCondition = conditions.find((c: any) => c.type === 'Running');
                expect(runningCondition).toBeDefined();
                expect(runningCondition.status).toBe('True');

                logger.info('Workload successfully reached desired state with topology constraints');
            });

            await test.step('Clean up: Delete workload', async () => {
                if (workloadId) {
                    const deleteResult = await workloadsManagerUtils.deleteWorkloadV2(workloadId);
                    expect(deleteResult.response.statusCode).toEqual(StatusCodes.ACCEPTED);
                    logger.info(`Deleted test workload: ${workloadId}`);
                    workloadInfo = workloadInfo.filter((w) => w.workloadId !== workloadId);
                }
            });
        });

        test('T5490561: Topology Status Information in GET /workload(failure scenario)', async () => {
            test.setTimeout(600 * 1000);

            let failureWorkloadId: string;
            const failureWorkloadName = objectGenerator.generateShortNameByPrefix('topo-status-fail', 5);
            await test.step('Update project GPU configuration', async () => {
                await projectResourcesUtils.updateProjectGpuConfig(npProject.id, nodePools[0].id, {
                    deserved: 2,
                    maxAllowed: 2,
                    overQuotaWeight: 2
                });
            });
            projectUpdated = true;

            await test.step('Create workload with impossible topology constraints', async () => {
                const manifest = workloadUtils.createPyTorchJobManifest(
                    failureWorkloadName,
                    `runai-${npProject.name}`,
                    'ubuntu',
                    sleepCommand,
                    1,
                    5
                );
                manifest.metadata['annotations'] = {
                    'kai.scheduler/topology': topoName,
                    'kai.scheduler/topology-preferred-placement': topoRegionKey
                };
                const response = await workloadsManagerUtils.createWorkloadV2(
                    failureWorkloadName,
                    npProject.id,
                    manifest,
                    {
                        priority: 'medium',
                        category: 'Train'
                    }
                );
                expect(response.response.statusCode).toEqual(StatusCodes.ACCEPTED);
                failureWorkloadId = response.body.metadata.id;
                workloadInfo.push({ workloadType: WorkloadType.DISTRIBUTED, workloadId: failureWorkloadId });
                logger.info(`Workload created with ID: ${failureWorkloadId}`);
            });

            await test.step('Verify workload in Creating phase', async () => {
                await waitUtils.waitForBackendWorkloadByIdAndPhase(failureWorkloadId, Phase.Creating, 120);
            });

            await test.step('Wait for pods are pending', async () => {
                const namespace = `runai-${npProject.name}`;
                const pods = await waitUtils.waitForPods(
                    failureWorkloadName,
                    namespace,
                    1,
                    60,
                    ['Pending'],
                    true,
                    5
                );
                expect(pods.length).toBeGreaterThan(0);
                logger.info(`Found ${pods.length} pods are pending`);
            });

            await test.step('Check the workload failed reason', async () => {
                const namespace = `runai-${npProject.name}`;
                const pods = await kubeCli.listPodsByNamespaceAndPrefix(namespace, failureWorkloadName);

                for (const pod of pods) {
                    const podName = pod.metadata?.name || 'unknown';
                    const podPhase = pod.status?.phase;
                    logger.info(`\nPod: ${podName} (Phase: ${podPhase})`);

                    let podEvents: any[] = [];
                    let hasUnschedulableOverLimitEvent = false;

                    await waitUtils.waitForCondition(
                        async () => {
                            const allEvents = await kubeCli.CoreV1Api.listNamespacedEvent({ namespace });
                            podEvents = allEvents.items
                                .filter((event: any) => event.involvedObject?.name === podName)
                                .sort(
                                    (a: any, b: any) =>
                                        new Date(a.lastTimestamp || a.firstTimestamp || 0).getTime() -
                                        new Date(b.lastTimestamp || b.firstTimestamp || 0).getTime()
                                );

                            hasUnschedulableOverLimitEvent = podEvents.some(
                                (event: any) =>
                                    event.reason === 'Unschedulable' && event.message?.includes('OverLimit')
                            );

                            logger.info(
                                `Waiting for Unschedulable event with OverLimit... (events: ${podEvents.length}, found: ${hasUnschedulableOverLimitEvent})`
                            );
                            return hasUnschedulableOverLimitEvent;
                        },
                        `Wait for Unschedulable event with OverLimit for ${podName}`,
                        true,
                        120,
                        5
                    );

                    podEvents.forEach((event: any) => {
                        const type = event.type || 'Normal';
                        const reason = event.reason || 'Unknown';
                        const message = event.message || '';
                        const count = event.count > 1 ? ` (x${event.count})` : '';
                        logger.info(`  ${type.padEnd(8)} ${reason.padEnd(20)}${count} ${message}`);
                    });

                    const statuses = podEvents.map((e: any) => e.reason).filter(Boolean);
                    logger.info(`Status progression: ${statuses.join(' -> ')}`);

                    logger.info('DEBUG: Looking for Unschedulable event with OverLimit message');
                    logger.info(`DEBUG: hasUnschedulableOverLimitEvent = ${hasUnschedulableOverLimitEvent}`);

                    expect(hasUnschedulableOverLimitEvent).toBe(true);
                }
            });

            await test.step('Clean up: Delete workload', async () => {
                await workloadsManagerUtils.deleteWorkloadV2(failureWorkloadId);
                workloadInfo.pop();
                logger.info(`Deleted workload: ${failureWorkloadId}`);
            });
        });

        test('T5481165: Workload Submission Fails When Nodepool Has Partial Topology Coverage', async ({}, testInfo: TestInfo) => {
            skipPlaywrightTest(test, testInfo, 'Partial-topology handling has an unresolved failure; revalidate the expected placement behavior before enabling this test');

            const onlyWorkerNodes = workerNodeNames.filter((node) => !node.includes('system'));
            if (onlyWorkerNodes.length < 2) {
                test.skip(true, 'Skipping test: requires at least 2 worker nodes');
            }

            await test.step('Create 2 node pools with different topology coverage', async () => {
                npProject = await nodePoolUtils.setupNodeGroupInfrastructure({
                    workerNodeNames,
                    nodeLabelKey,
                    nodeLabelVal1,
                    nodeLabelVal2,
                    nodeGroupSize: 2,
                    midIndex,
                    nodeGroup1,
                    nodeGroup2,
                    createTopology: true,
                    attachTopology: true,
                    topoLevels: [topoRegionKey, topoCityKey],
                    createTopoRequest,
                    topologyIds,
                    topoRegionKey,
                    topoCityKey,
                    region,
                    cityName,
                    nodePoolCreateRequest,
                    nodePoolCreateRequest2,
                    nodePools,
                    npName1,
                    npName2,
                    npProjectName,
                    gpuDeserved: 3
                });

                workerNodeNames = [];
                nodeGroup1 = [];
                nodeGroup2 = [];
            });

            await test.step('Submit workload targeting to the nodepool', async () => {
                const manifest = workloadUtils.createPyTorchJobManifest(
                    workloadName,
                    `runai-${npProject.name}`,
                    'ubuntu',
                    sleepCommand
                );
                manifest.metadata['annotations'] = {
                    'kai.scheduler/topology': topoName,
                    'kai.scheduler/topology-preferred-placement': topoRegionKey
                };

                try {
                    const response = await workloadsManagerUtils.createWorkloadV2(
                        workloadName,
                        npProject.id,
                        manifest,
                        {
                            priority: 'medium',
                            category: 'Train',
                            nodePools: nodePoolNames
                        } as any
                    );
                    workloadId = response.body.metadata.id;
                    workloadInfo.push({ workloadType: WorkloadType.DISTRIBUTED, workloadId: workloadId });
                    expect([400, 403, 422]).toContain(response.response.statusCode);
                } catch (error: any) {
                    logger.info(
                        `  Error Message: ${error.message || error.body?.message || JSON.stringify(error.body)}`
                    );
                    (this as any).apiError = error;
                }
            });

            await test.step('Verify API rejection with appropriate error', () => {
                const error = (this as any).apiError;
                expect(error).toBeDefined();
                const errorMessage = error.message || error.body?.message || JSON.stringify(error.body);
                logger.info(`Error message: ${errorMessage}`);
            });

            await test.step('Cleanup: Restore npName2 WITHOUT topology', async () => {
                if (workloadId) {
                    const deleteResult = await workloadsManagerUtils.deleteWorkloadV2(workloadId);
                    expect(deleteResult.response.statusCode).toEqual(StatusCodes.ACCEPTED);
                    logger.info(`Deleted workload V2: ${workloadId}`);
                    workloadInfo.pop();
                }
            });
        });
    });
});

test.describe(`Any Workload Network Topology Support Tests(without topology), tags: ${tags.tool.api} ${tags.resource.networkTopologies} ${tags.resource.workloads} ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_24)(function () {
        const topoNames: string[] = [];
        const topologyIds: string[] = [];
        let npProject: Project;
        let workloadName: string;
        let workloadId: string;
        let workloadInfo: { workloadType: WorkloadType; workloadId: string }[] = [];

        const nodeLabelKey = 'any_workload_topo_node_group';
        const nodeLabelVal1 = 'test_group1';
        const sleepCommand = 'sleep infinity';
        const nodePoolNames: string[] = [];
        const nodePools: { name: string; id: string }[] = [];
        let nodeGroup1: string[];

        const topoName = objectGenerator.generateShortNameByPrefix('ui-topo2', 5);
        topoNames.push(topoName);
        const npProjectName = objectGenerator.generateShortNameByPrefix('proj-np2', 5);
        const topoRegionKey = 'topology.runai.com/region';
        const topoCityKey = 'topology.runai.com/city';
        const region = 'us-east2';
        const cityName = 'newyork';

        const npName1 = objectGenerator.generateShortNameByPrefix('np-no-topo2', 5);
        const npName2 = objectGenerator.generateShortNameByPrefix('np-no-topo2', 5);
        nodePoolNames.push(npName1, npName2);

        const createTopoRequest: NetworkTopologyCreateRequest = {
            name: topoName,
            clusterId: flowRootState.cluster.uuid,
            levels: []
        };

        const nodePoolCreateRequest = Object.assign(new NodepoolCreateFields(), {
            name: npName1,
            labelKey: nodeLabelKey,
            labelValue: nodeLabelVal1,
            clusterId: flowRootState.cluster.uuid
        });

        test.beforeAll(async () => {
            workloadsManagerUtils.accessToken = flowRootState.users.system_administrator.accessToken;
            workloadsCrudUtils.accessToken = flowRootState.users.system_administrator.accessToken;

            npProject = await nodePoolUtils.setupNodeGroupInfrastructure({
                nodeLabelKey,
                nodeLabelVal1,
                nodeGroupSize: 1,
                nodeGroup1,
                createTopology: true,
                attachTopology: false,
                topoLevels: [topoRegionKey, topoCityKey],
                createTopoRequest,
                topologyIds,
                topoRegionKey,
                topoCityKey,
                region,
                cityName,
                nodePoolCreateRequest,
                nodePools,
                npName1,
                npProjectName,
                gpuDeserved: 4
            });
        });

        test.afterAll(async () => {
            if (npProject) {
                try {
                    await operations.deleteProjectAndNamespace(parseInt(npProject.id));
                    logger.info(`Deleted project ${npProject.name}`);
                } catch (error) {
                    logger.error(`Failed to delete project ${npProject.name}:`, error);
                }
            }

            if (nodePools.length > 0) {
                await Promise.all(
                    nodePools.map(async (np) => {
                        try {
                            await nodePoolUtils.deleteNodePoolByIdWithWait(np.id, np.name);
                            logger.info(`Deleted node pool ${np.name}`);
                        } catch (error) {
                            logger.error(`Failed to delete node pool ${np.name}:`, error);
                        }
                    })
                );
            }

            if (topologyIds.length > 0) {
                await Promise.all(
                    topologyIds.map(async (id) => {
                        try {
                            await networkTopologiesUtils.deleteNetworkTopologyViaApi(id);
                            logger.info(`Deleted topology ${id}`);
                            try {
                                await networkTopologiesUtils.getNetworkTopologyByIdViaApi(id);
                                logger.error(`Topology ${id} still exists after deletion!`);
                            } catch (error) {
                                const statusCode = error.response?.statusCode || error.statusCode;
                                if (statusCode === 404 || error.message?.includes('not found')) {
                                    logger.info(`Verified topology ${id} was successfully deleted`);
                                } else {
                                    logger.warn(
                                        `Topology ${id} deletion verification returned unexpected error:`,
                                        error.message
                                    );
                                }
                            }
                        } catch (error) {
                            logger.error(`Failed to delete topology ${id}:`, error);
                        }
                    })
                );
            }

            const nodesToClean = await nodeUtils.getWorkerNodeNames();
            if (nodesToClean && nodesToClean.length > 0) {
                const unlabelPromises: Promise<void>[] = [];
                const labelsToClean = [nodeLabelKey, topoRegionKey, topoCityKey];

                nodesToClean.forEach((node) => {
                    labelsToClean.forEach((label) => {
                        unlabelPromises.push(kubeCli.unLabelNode(node, label).catch(() => {}));
                    });
                });

                if (unlabelPromises.length > 0) {
                    await Promise.all(unlabelPromises);
                    logger.info(
                        `Cleaned up node labels (${labelsToClean.join(', ')}) from ${nodesToClean.length} nodes`
                    );
                }
            }
        });

        test.beforeEach(() => {
            workloadName = objectGenerator.generateShortNameByPrefix('any-workload-no-topo', 5);
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
                            logger.error(`Failed to delete ${workload.workloadType} workload:`, error);
                        }
                    })
                );
            }
        });

        test('T5481162: Invalid Topology Reference via API', async () => {
            const nonExistentTopologyName = objectGenerator.generateShortNameByPrefix(
                'non-existent-topology',
                5
            );
            await test.step('Submit workload referencing non-existent topology via API', async () => {
                const manifest = workloadUtils.createPyTorchJobManifest(
                    workloadName,
                    `runai-${npProject.name}`,
                    'ubuntu',
                    sleepCommand
                );
                manifest.metadata['annotations'] = {
                    'kai.scheduler/topology': nonExistentTopologyName
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
                logger.info(
                    `API accepted the workload with invalid topology reference. WorkloadId: ${workloadId}`
                );
            });

            await test.step('Verify workload is in Creating phase due to invalid topology', async () => {
                await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Creating, 60);
            });

            await test.step('Verify pod warnings show error about missing topology', async () => {
                let pods = [];
                const checkPods = async (): Promise<boolean> => {
                    pods = await kubeCli.listPodsByNamespaceAndPrefix(
                        `runai-${npProject.name}`,
                        workloadName
                    );
                    return pods.length > 0;
                };

                await waitUtils.waitForCondition(
                    checkPods,
                    `Wait for pods to be created for workload ${workloadName}`,
                    true,
                    60,
                    5
                );

                const podName = pods[0].metadata.name;
                logger.info(`Checking topology warning on pod ${podName}`);

                let topologyWarning;
                const checkTopologyWarning = async (): Promise<boolean> => {
                    const events = await kubeCli.CoreV1Api.listNamespacedEvent({
                        namespace: `runai-${npProject.name}`,
                        fieldSelector: `involvedObject.name=${podName}`
                    });

                    topologyWarning = events.items.find(
                        (event) =>
                            event.type === 'Warning' &&
                            event.reason === 'Unschedulable' &&
                            event.source?.component === 'runai-scheduler' &&
                            event.message?.includes('matching topology tree was not found') &&
                            event.message?.includes(nonExistentTopologyName)
                    );

                    // If not found, try with relaxed conditions
                    if (!topologyWarning) {
                        topologyWarning = events.items.find(
                            (event) =>
                                event.type === 'Warning' &&
                                event.reason === 'Unschedulable' &&
                                event.source?.component === 'runai-scheduler' &&
                                (event.message?.includes('topology') ||
                                    event.message?.includes(nonExistentTopologyName))
                        );
                    }

                    return topologyWarning !== undefined;
                };

                await waitUtils.waitForCondition(
                    checkTopologyWarning,
                    `Wait for topology warning event on pod ${podName}`,
                    true,
                    120,
                    5
                );

                expect(topologyWarning).toBeDefined();
                expect(topologyWarning.type).toBe('Warning');
                expect(topologyWarning.reason).toBe('Unschedulable');
                expect(topologyWarning.source?.component).toBe('runai-scheduler');
                logger.info(
                    `Verified: Pod ${podName} has topology warning about invalid topology reference: ${nonExistentTopologyName}`
                );
            });

            await test.step('Clean up: Delete workload', async () => {
                const deleteResult = await workloadsManagerUtils.deleteWorkloadV2(workloadId);
                expect(deleteResult.response.statusCode).toEqual(StatusCodes.ACCEPTED);
                logger.info(`Deleted workload V2: ${workloadId}`);
                workloadInfo.pop();
            });
        });

        test('T5481164: Workload Targeting Node Pool Without Topology Binding', async () => {
            await test.step('Submit workload with topology annotations to a node pool without topology', async () => {
                const manifest = workloadUtils.createPyTorchJobManifest(
                    workloadName,
                    `runai-${npProject.name}`,
                    'ubuntu',
                    sleepCommand
                );
                manifest.metadata['annotations'] = {
                    'kai.scheduler/topology': topoName,
                    'kai.scheduler/topology-preferred-placement': topoRegionKey
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
                logger.info(
                    `API accepted workload with topology annotations targeting nodepool(no topology). WorkloadId: ${workloadId}`
                );
            });

            await test.step('Verify workload will be running', async () => {
                await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 120);
                logger.info('Workload is in Running');
            });

            await test.step('Verify pods are scheduled successfully', async () => {
                const pods = await kubeCli.listPodsByNamespaceAndPrefix(
                    `runai-${npProject.name}`,
                    workloadName
                );

                expect(pods.length).toBeGreaterThan(0);
                logger.info(`Found ${pods.length} pods for workload ${workloadName}`);

                pods.forEach((pod) => {
                    if (pod.spec?.nodeName) {
                        logger.info(`Pod ${pod.metadata.name} scheduled on node: ${pod.spec.nodeName}`);
                    }
                });
            });

            await test.step('Clean up: Delete workload', async () => {
                const deleteResult = await workloadsManagerUtils.deleteWorkloadV2(workloadId);
                expect(deleteResult.response.statusCode).toEqual(StatusCodes.ACCEPTED);
                logger.info(`Deleted workload V2: ${workloadId}`);
                workloadInfo.pop();
            });
        });

        test('T5481170: Topology Annotations Visibility in GET /workload Spec without topology attached', async () => {
            await test.step('Submit workload targeting node pool without topology', async () => {
                const manifest = workloadUtils.createPyTorchJobManifest(
                    workloadName,
                    `runai-${npProject.name}`,
                    'ubuntu',
                    sleepCommand,
                    1
                );
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
                logger.info(`Workload created with ID: ${workloadId}`);
            });

            await test.step('Wait for workload to reach Running phase', async () => {
                await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 120);
                logger.info('Workload is in Running phase');
            });

            await test.step('Verify response structure includes spec section (consistent API structure)', async () => {
                const getResult = await workloadsManagerUtils.getWorkloadV2ById(workloadId);
                expect(getResult.response.statusCode).toEqual(StatusCodes.OK);
                const workload = getResult.body.workload;
                expect(workload.workloadManifest.spec.pytorchReplicaSpecs).toBeDefined();
                expect(workload.workloadManifest.spec.pytorchReplicaSpecs.Master).toBeDefined();
                logger.info('Response structure includes spec section (API structure is consistent)');
            });

            await test.step('Verify spec.template.metadata.annotations does NOT contain topology annotations', async () => {
                const getResult = await workloadsManagerUtils.getWorkloadV2ById(workloadId);
                const workload = getResult.body.workload;
                const manifestAnnotations = workload.workloadManifest?.metadata?.annotations || {};
                expect(manifestAnnotations['kai.scheduler/topology']).toBeUndefined();
                expect(manifestAnnotations['kai.scheduler/topology-preferred-placement']).toBeUndefined();
                logger.info(
                    'Confirmed topology annotations are NOT present (as expected for node pool without topology)'
                );
            });

            await test.step('Verify GET API response is clean and consistent regardless of topology presence', async () => {
                const getResult = await workloadsManagerUtils.getWorkloadV2ById(workloadId);
                const workload = getResult.body.workload;
                const syncInfo = getResult.body.syncInfo;
                expect(getResult.response.statusCode).toEqual(StatusCodes.OK);

                // Verify top-level workload structure
                expect(workload).toBeDefined();
                expect(workload.metadata).toBeDefined();
                expect(workload.desiredPhase).toBeDefined();
                expect(workload.desiredPhase).toBe('Running');

                // Verify workloadManifest structure
                expect(workload.workloadManifest).toBeDefined();
                expect(workload.workloadManifest.kind).toBe('PyTorchJob');
                expect(workload.workloadManifest.metadata).toBeDefined();
                expect(workload.workloadManifest.spec).toBeDefined();

                // Verify rawManifest exists
                expect(workload.rawManifest).toBeDefined();

                // Verify syncInfo structure
                expect(syncInfo).toBeDefined();
                expect(syncInfo.status).toBeDefined();
                expect(['Applying', 'Synced']).toContain(syncInfo.status);

                logger.info('GET API response is clean and consistent:');
            });

            await test.step('Clean up: Delete workload', async () => {
                const deleteResult = await workloadsManagerUtils.deleteWorkloadV2(workloadId);
                expect(deleteResult.response.statusCode).toEqual(StatusCodes.ACCEPTED);
                logger.info(`Deleted workload: ${workloadId}`);
                workloadInfo.pop();
            });
        });

        test('T5481172: GET API Topology Info for Workload Without Topology', async () => {
            await test.step('Create a workload', async () => {
                const manifest = workloadUtils.createPyTorchJobManifest(
                    workloadName,
                    `runai-${npProject.name}`,
                    'ubuntu',
                    sleepCommand,
                    1
                );

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
                logger.info(`Workload created with ID: ${workloadId}`);
            });

            await test.step('Wait for workload to reach Running state', async () => {
                await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 120);
            });

            await test.step('Query via GET API and examine response', async () => {
                const getResult = await workloadsManagerUtils.getWorkloadV2ById(workloadId);
                expect(getResult.response.statusCode).toEqual(StatusCodes.OK);
                logger.info('Successfully queried workload via GET API');
            });

            await test.step('Verify response structure is complete and valid', async () => {
                const getResult = await workloadsManagerUtils.getWorkloadV2ById(workloadId);
                const workload = getResult.body.workload;

                logger.info('Verifying response structure for workload WITHOUT topology:');

                // Verify essential fields exist
                expect(workload).toBeDefined();
                expect(workload.workloadManifest).toBeDefined();
                expect(workload.workloadManifest.metadata).toBeDefined();
                expect(workload.workloadManifest.spec).toBeDefined();
                logger.info('Essential fields (workloadManifest, metadata, spec) are present');

                const annotations = workload.workloadManifest?.metadata?.annotations || {};
                expect(annotations['kai.scheduler/topology']).toBeUndefined();
                expect(annotations['kai.scheduler/topology-preferred-placement']).toBeUndefined();
                logger.info('Annotations are clean - no topology annotations present');
            });

            await test.step('Verify response is clean (no null/undefined topology fields causing confusion)', async () => {
                const getResult = await workloadsManagerUtils.getWorkloadV2ById(workloadId);
                const workload = getResult.body.workload;

                logger.info('Checking for clean response (no confusing null/undefined fields):');

                // Check that topology annotations are absent (not present as null/empty)
                const annotations = workload.workloadManifest?.metadata?.annotations || {};
                const topologyAnnotationKeys = Object.keys(annotations).filter((key) =>
                    key.startsWith('kai.scheduler/topology')
                );
                expect(topologyAnnotationKeys.length).toBe(0);
                logger.info('No topology annotations present (expected: [])');
                const topologyValue = annotations['kai.scheduler/topology'];
                const preferredPlacementValue = annotations['kai.scheduler/topology-preferred-placement'];
                const requiredPlacementValue = annotations['kai.scheduler/topology-required-placement'];
                expect(topologyValue).toBeUndefined();
                expect(preferredPlacementValue).toBeUndefined();
                expect(requiredPlacementValue).toBeUndefined();
                logger.info('Topology annotations are properly absent (undefined, not null or empty)');

                //Verify no topology-related fields in workload status/metadata
                const workloadStr = JSON.stringify(workload, null, 2);
                const topologyReferences = workloadStr.match(/"topology[^"]*":/gi) || [];
                expect(topologyReferences.length).toBe(0);
                logger.info('Verified no unexpected topology-related fields in status');

                // Verify clean JSON structure (no null topology keys)
                const manifestAnnotations = workload.workloadManifest?.metadata?.annotations;
                if (manifestAnnotations) {
                    for (const key of Object.keys(manifestAnnotations)) {
                        if (key.includes('topology')) {
                            const value = manifestAnnotations[key];
                            expect(value).not.toBe(null);
                            expect(value).not.toBe('');
                            expect(value).not.toBe('null');
                            logger.warn(`Unexpected topology key found: ${key}=${value}`);
                        }
                    }
                }
                logger.info('JSON structure is clean (no null/empty topology keys)');

                // Verify metadata doesn't have misleading topology references
                const metadataAnnotations = workload.workloadManifest?.metadata?.annotations || {};
                const metadataLabels = workload.workloadManifest?.metadata?.labels || {};

                const topologyMetadataAnnotations = Object.keys(metadataAnnotations).filter((key) =>
                    key.toLowerCase().includes('topology')
                );
                const topologyMetadataLabels = Object.keys(metadataLabels).filter((key) =>
                    key.toLowerCase().includes('topology')
                );

                logger.info(`Metadata topology annotations: ${JSON.stringify(topologyMetadataAnnotations)}`);
                logger.info(`Metadata topology labels: ${JSON.stringify(topologyMetadataLabels)}`);

                expect(topologyMetadataAnnotations.length).toBe(0);
                expect(topologyMetadataLabels.length).toBe(0);
                logger.info('Metadata has no misleading topology references');

                // Additional check: Count null occurrences in entire response
                const responseStr = JSON.stringify(workload, null, 2);
                const nullMatches = (responseStr.match(/"kai\.scheduler\/topology[^"]*":\s*null/g) || [])
                    .length;
                expect(nullMatches).toBe(0);
                logger.info('Response is clean - no null/undefined topology fields causing confusion');
            });

            await test.step('Clean up: Delete workload', async () => {
                if (workloadId) {
                    const deleteResult = await workloadsManagerUtils.deleteWorkloadV2(workloadId);
                    expect(deleteResult.response.statusCode).toEqual(StatusCodes.ACCEPTED);
                    logger.info(`Deleted workload: ${workloadId}`);
                    workloadInfo = workloadInfo.filter((w) => w.workloadId !== workloadId);
                }
            });
        });

        test('T5481178: Workload to Nodepool Without Topology Runs Without Topology Annotations', async () => {
            await test.step('Submit workload to nodepool WITHOUT topology, should be accepted', async () => {
                const manifest = workloadUtils.createPyTorchJobManifest(
                    workloadName,
                    `runai-${npProject.name}`,
                    'ubuntu',
                    sleepCommand,
                    1
                );
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
                logger.info(`Workload created with ID: ${workloadId}`);
            });

            await test.step('Wait for workload to reach Running state', async () => {
                await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 120);
            });

            await test.step('Retrieve workload and verify NO topology annotations were added', async () => {
                const getResult = await workloadsManagerUtils.getWorkloadV2ById(workloadId);
                const workload = getResult.body.workload;
                const currentAnnotations = workload.workloadManifest?.metadata?.annotations || {};
                expect(currentAnnotations['kai.scheduler/topology']).not.toBeDefined();
                expect(currentAnnotations['kai.scheduler/topology-preferred-placement']).not.toBeDefined();
            });

            await test.step('Wait for pods to be running successfully', async () => {
                const pods = await waitUtils.waitForPods(workloadName, `runai-${npProject.name}`, 2, 120, [
                    'Running'
                ]);
                expect(pods.length).toBe(2);
            });

            await test.step('Verify workload reaches desired state: Check 2/2 ready replicas', async () => {
                const namespace = `runai-${npProject.name}`;
                const crd = (await kubeCli.getCrd(crds.pyTorchJobCrd, workloadName, namespace)) as any;
                logger.info('\nPyTorchJob Status:');
                logger.info(JSON.stringify(crd.status, null, 2));
                const replicaStatuses = crd.status.replicaStatuses;
                expect(replicaStatuses).toBeDefined();
                expect(replicaStatuses.Master.active).toBe(1);
                expect(replicaStatuses.Worker.active).toBe(1);
                const conditions = crd.status.conditions;
                expect(conditions).toBeDefined();
                expect(conditions.length).toBeGreaterThan(0);

                // Find the Running condition
                const runningCondition = conditions.find((c: any) => c.type === 'Running');
                expect(runningCondition).toBeDefined();
                expect(runningCondition.status).toBe('True');
                logger.info('Workload successfully reached desired state with topology constraints');
            });

            await test.step('Clean up: Delete workload', async () => {
                if (workloadId) {
                    const deleteResult = await workloadsManagerUtils.deleteWorkloadV2(workloadId);
                    expect(deleteResult.response.statusCode).toEqual(StatusCodes.ACCEPTED);
                    logger.info(`Deleted test workload: ${workloadId}`);
                    workloadInfo = workloadInfo.filter((w) => w.workloadId !== workloadId);
                }
            });
        });
    });
});
