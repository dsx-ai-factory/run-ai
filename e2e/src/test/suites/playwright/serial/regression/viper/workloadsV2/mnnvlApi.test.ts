/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    config,
    crds,
    flowRootState,
    logger,
    nodePoolUtils,
    objectGenerator,
    orgUnitUtils,
    tags,
    Utils,
    waitUtils,
    mnnvlUtils
} from '../../../../../../../modules/helpers';
import { workloadsManagerService } from '../../../../../../../modules/api';
import { kubeCli } from '../../../../../../../modules/cli';
import { test, expect } from '@playwright/test';
import { StatusCodes } from 'http-status-codes';
import { WorkloadV2CreateRequest } from '../../../../../../../modules/api/apis/workloadsManagerApi/generated/model/workloadV2CreateRequest';
import { Department, Project } from '../../../../../../../modules/api/apis/orgUnitApi';
import { Phase } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import { MnnvlMode } from '../../../../../../../modules/api/apis/workloadsManagerApi/generated/model/mnnvlMode';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { ComputeDomain } from '../../../../../../../modules/helpers/models/computeDomain';
import { HttpError } from '../../../../../../../modules/api/apis/workloadsManagerApi/generated/api';
import { MnnvlUtils } from '../../../../../../../modules/helpers/utils/mnnvlUtils';
import { skipPlaywrightTest } from '../../../../../../../modules/helpers/utils/SkippedTestUtils';
import { NodePool } from '../../../../../../../modules/api/apis/backendClusterApi';

supportedFrom(SupportedVersions.V2_24)(function () {
    test.describe(`MNNVL API tests: ${tags.resource.workloads} ${tags.tool.api} ${tags.teams.viper} ${tags.teams.viper}`, function () {
        let mnnvlProject: Project;
        let nonMnnvlProject: Project;
        let testDepartment: Department;
        let createdWorkloadIds: string[] = [];
        let defaultNodePoolId: string;
        let mnnvlNodePoolId: string;
        let mnnvlNodePoolName: string;

        // Create a Deployment manifest with GPU resources for MNNVL testing
        function createDeploymentManifest(name: string, replicas: number = 2): { [key: string]: any } {
            return {
                apiVersion: 'apps/v1',
                kind: 'Deployment',
                metadata: {
                    name: name
                },
                spec: {
                    replicas: replicas,
                    selector: {
                        matchLabels: {
                            app: name
                        }
                    },
                    template: {
                        metadata: {
                            labels: {
                                app: name
                            }
                        },
                        spec: {
                            containers: [
                                {
                                    name: 'main',
                                    image: config.quickStartImage.ubuntu,
                                    command: ['sleep', 'infinity'],
                                    resources: {
                                        limits: {
                                            'nvidia.com/gpu': '1'
                                        }
                                    }
                                }
                            ]
                        }
                    }
                }
            };
        }

        test.beforeAll(async ({}, testInfo) => {
            if (!mnnvlUtils.isDraSupportedInCluster()) {
                skipPlaywrightTest(
                    test,
                    testInfo,
                    'The k8s environment does not support dra - skipping as it is a prerequisite for MNNVL'
                );
            }
            await test.step('Install MNNVL operator', async () => {
                await mnnvlUtils.installMnnvlOperator();
            });

            await test.step('Setup authentication', () => {
                workloadsManagerService.workloadsV2Api.accessToken =
                    flowRootState.users.system_administrator.idToken;
            });

            await test.step('Get default node pool for non-MNNVL project', async () => {
                const defaultNodePool = await nodePoolUtils.getDefaultNodePoolFromBackend();
                defaultNodePoolId = String(defaultNodePool.id);
                logger.info(
                    `Using default node pool for non-MNNVL project: ${defaultNodePool.name} with ID: ${defaultNodePoolId}`
                );
            });

            await test.step('Create MNNVL-capable node pool', async () => {
                const mnnvlNodePoolLabelKey = objectGenerator.generateNameByPrefix('mnnvl-key');
                const mnnvlNodePoolLabelValue = objectGenerator.generateNameByPrefix('mnnvl-val');
                const mnnvlNodePoolNamePrefix = objectGenerator.generateNameByPrefix('mnnvl-np');

                await nodePoolUtils.labelAGPUNode(mnnvlNodePoolLabelKey, mnnvlNodePoolLabelValue);

                const createdNodePool = await mnnvlUtils.createMnnvlNodePool(
                    mnnvlNodePoolLabelKey,
                    mnnvlNodePoolLabelValue,
                    mnnvlNodePoolNamePrefix
                );
                mnnvlNodePoolId = String(createdNodePool.nodePoolId);
                mnnvlNodePoolName = createdNodePool.nodePoolName;
                logger.info(
                    `Created MNNVL-capable node pool: ${mnnvlNodePoolName} with ID: ${mnnvlNodePoolId}`
                );
            });

            await test.step('Enable GPU network acceleration (runaiconfig + node pool)', async () => {
                // Atomic with the afterAll disable: flips the cluster-wide runaiconfig flag and
                // marks the MNNVL node pool as acceleration-capable, then waits for detection.
                await mnnvlUtils.enableGpuNetworkAcceleration(mnnvlNodePoolId);
            });

            await test.step('Wait for MNNVL node pool CRD to sync', async () => {
                await nodePoolUtils.waitForNodePoolCrdStatus(
                    mnnvlNodePoolName,
                    [NodePool.StatusEnum.Ready],
                    120
                );
                logger.info(`MNNVL node pool ${mnnvlNodePoolName} CRD is Ready`);
            });

            await test.step('Create test department with quota on MNNVL node pool', async () => {
                testDepartment = await orgUnitUtils.createDepartment({
                    name: objectGenerator.generateNameByPrefix('mnnvl-dep'),
                    clusterId: flowRootState.cluster.uuid,
                    resources: [
                        {
                            nodePool: { id: defaultNodePoolId, name: 'default' },
                            gpu: { deserved: 1, limit: -1 }
                        },
                        {
                            nodePool: { id: mnnvlNodePoolId, name: mnnvlNodePoolName },
                            gpu: { deserved: 2, limit: -1 }
                        }
                    ]
                });
                logger.info(
                    `Created test department: ${testDepartment.name} with ID: ${testDepartment.id}`
                );
            });

            await test.step('Create non-MNNVL project on default node pool', async () => {
                nonMnnvlProject = await orgUnitUtils.createProjectAndValidate({
                    resources: [
                        {
                            nodePool: { id: defaultNodePoolId, name: 'default' },
                            gpu: { deserved: 1 }
                        }
                    ],
                    name: objectGenerator.generateNameByPrefix('non-mnnvl-proj'),
                    clusterId: flowRootState.cluster.uuid,
                    parentId: testDepartment.id
                });
                logger.info(
                    `Created non-MNNVL project: ${nonMnnvlProject.name} with ID: ${nonMnnvlProject.id}`
                );
            });

            await test.step('Create MNNVL project on MNNVL node pool', async () => {
                logger.info(
                    `Creating MNNVL project with node pool: ${mnnvlNodePoolName} (ID: ${mnnvlNodePoolId})`
                );
                const mnnvlProjectRequest = {
                    resources: [
                        {
                            nodePool: { id: mnnvlNodePoolId, name: mnnvlNodePoolName },
                            gpu: { deserved: 2 }
                        }
                    ],
                    name: objectGenerator.generateNameByPrefix('mnnvl-proj'),
                    clusterId: flowRootState.cluster.uuid,
                    parentId: testDepartment.id,
                    defaultNodePools: [mnnvlNodePoolName]
                };
                logger.info(
                    `MNNVL project creation request: ${JSON.stringify(mnnvlProjectRequest, null, 2)}`
                );
                mnnvlProject = await orgUnitUtils.createProjectAndValidate(mnnvlProjectRequest);
                logger.info(`Created MNNVL project: ${mnnvlProject.name} with ID: ${mnnvlProject.id}`);
                logger.info(`MNNVL project resources: ${JSON.stringify(mnnvlProject.resources, null, 2)}`);
            });

            await test.step('Wait for MNNVL project CRD defaultNodePools to sync', async () => {
                await waitUtils.waitForCondition(
                    async () => {
                        try {
                            const projectCrd = await kubeCli.getCrd(crds.runaiProjectCrd, mnnvlProject.name, '');
                            const defaultNodePools = projectCrd?.['spec']?.['defaultNodePools'] as string[] | undefined;
                            return Array.isArray(defaultNodePools) && defaultNodePools.includes(mnnvlNodePoolName);
                        } catch (e) {
                            return false;
                        }
                    },
                    `project ${mnnvlProject.name} CRD to have defaultNodePools=[${mnnvlNodePoolName}]`,
                    true,
                    60
                );
                logger.info(`MNNVL project ${mnnvlProject.name} CRD defaultNodePools synced`);
            });
        });

        test.afterEach(async function () {
            for (const workloadId of createdWorkloadIds) {
                try {
                    await workloadsManagerService.workloadsV2Api.deleteWorkloadV2(workloadId);
                    logger.debug(`Cleaned up workload: ${workloadId}`);
                } catch (error) {
                    logger.warn(`Failed to clean up workload ${workloadId}: ${error.message}`);
                }
            }
            createdWorkloadIds = [];
        });

        test.afterAll(async function () {
            try {
                // Revert first (node pool detection + cluster-wide runaiconfig flag) so the
                // cluster never stays half-configured, before the node pool is deleted below.
                if (mnnvlNodePoolId) {
                    await mnnvlUtils.disableGpuNetworkAcceleration(mnnvlNodePoolId);
                    logger.info('Reverted GPU network acceleration');
                }
                if (mnnvlProject) {
                    await orgUnitUtils.deleteProject(mnnvlProject.id);
                    logger.info(`Deleted MNNVL project: ${mnnvlProject.name}`);
                }
                if (nonMnnvlProject) {
                    await orgUnitUtils.deleteProject(nonMnnvlProject.id);
                    logger.info(`Deleted non-MNNVL project: ${nonMnnvlProject.name}`);
                }
                if (testDepartment) {
                    await orgUnitUtils.deleteDepartment(testDepartment.id);
                    logger.info(`Deleted test department: ${testDepartment.name}`);
                }
                if (mnnvlNodePoolId && mnnvlNodePoolName && mnnvlNodePoolId !== defaultNodePoolId) {
                    await nodePoolUtils.cleanupDeleteNodePool({
                        nodePoolId: mnnvlNodePoolId,
                        nodePoolName: mnnvlNodePoolName
                    });
                    logger.info(`Deleted MNNVL node pool: ${mnnvlNodePoolName}`);
                }
            } catch (error) {
                logger.error(`Failed to cleanup: ${error}`);
            }
        });

        test('mnnvlMode=none: no ComputeDomain created', async function ({}, testInfo) {
            const workloadName = objectGenerator.generateNameByPrefix('mnnvl-none');
            let workloadId: string;

            await test.step('Create workload with mnnvlMode=None', async () => {
                const manifest = createDeploymentManifest(workloadName, 2);
                const createRequest: WorkloadV2CreateRequest = {
                    metadata: {
                        name: workloadName,
                        projectId: String(mnnvlProject.id),
                        configuration: {
                            mnnvl: MnnvlMode.None
                        }
                    },
                    manifest
                };

                logger.info(`Creating workload with mnnvlMode=None: ${workloadName}`);
                const result = await workloadsManagerService.workloadsV2Api.createWorkloadV2(createRequest);

                expect(result.response.statusCode).toBe(StatusCodes.ACCEPTED);
                expect(result.body.metadata.id).toBeDefined();

                workloadId = result.body.metadata.id;
                createdWorkloadIds.push(workloadId);
            });

            await test.step('Wait for workload to reach Running phase', async () => {
                await waitUtils.waitForBackendWorkloadByIdAndPhase(
                    workloadId,
                    // With mnnvlMode=None the pod has no MNNVL requirements and
                    // schedules normally on the labeled node.
                    Phase.Running,
                    120
                );
            });

            await test.step('Verify no ComputeDomain was created', async () => {
                const namespace = Utils.namespaceFromProjectName(mnnvlProject.name);
                const computeDomains = (await kubeCli.getCrds(
                    crds.computeDomainCrd,
                    namespace
                )) as ComputeDomain[];

                const workloadComputeDomains = computeDomains.filter((cd) =>
                    cd.metadata?.name?.includes(workloadName)
                );

                expect(workloadComputeDomains).toHaveLength(0);

                logger.info(`Verified no ComputeDomain created for workload: ${workloadName}`);
            });
        });

        test('mnnvlMode=required with MNNVL nodepool: ComputeDomain created', async function ({}, testInfo) {
            const workloadName = objectGenerator.generateNameByPrefix('mnnvl-req');
            const replicas = 2;
            let workloadId: string;
            let namespace: string;
            let computeDomain: ComputeDomain;
            let resourceClaimTemplateName: string;

            await test.step('Create workload with mnnvlMode=Required', async () => {
                const manifest = createDeploymentManifest(workloadName, replicas);

                const createRequest: WorkloadV2CreateRequest = {
                    metadata: {
                        name: workloadName,
                        projectId: String(mnnvlProject.id),
                        configuration: {
                            mnnvl: MnnvlMode.Required
                        }
                    },
                    manifest
                };

                logger.info(`Creating workload with mnnvlMode=Required: ${workloadName}`);
                const result = await workloadsManagerService.workloadsV2Api.createWorkloadV2(createRequest);

                expect(result.response.statusCode).toBe(StatusCodes.ACCEPTED);
                expect(result.body.metadata.id).toBeDefined();

                workloadId = result.body.metadata.id;
                createdWorkloadIds.push(workloadId);
            });

            await test.step('Wait for ComputeDomain to be created', async () => {
                // We can't wait for the workload to reach Running: the DRA driver on a
                // non-GB200 cluster cannot allocate the ComputeDomain resource claims,
                // so pods stay Unschedulable and the workload stays Initializing. Wait
                // for the ComputeDomain CRD instead - that's what this test validates.
                namespace = Utils.namespaceFromProjectName(mnnvlProject.name);
                await waitUtils.waitForCondition(
                    async () => {
                        const cds = (await kubeCli.getCrds(
                            crds.computeDomainCrd,
                            namespace
                        )) as ComputeDomain[];
                        return cds.some((cd) => cd.metadata?.name?.includes(workloadName));
                    },
                    `ComputeDomain for workload ${workloadName} to be created`,
                    true,
                    120
                );
            });

            await test.step('Wait for workload pod to be created', async () => {
                await waitUtils.waitForCondition(
                    async () => {
                        const pods = await kubeCli.listPodsByNamepsace(namespace);
                        return pods.some((pod) => pod.metadata.name.includes(workloadName));
                    },
                    `pod for workload ${workloadName} to be created`,
                    true,
                    120
                );
            });

            await test.step('Verify ComputeDomain was created with correct spec', async () => {
                const computeDomains = (await kubeCli.getCrds(
                    crds.computeDomainCrd,
                    namespace
                )) as ComputeDomain[];

                const workloadComputeDomains = computeDomains.filter((cd) =>
                    cd.metadata?.name?.includes(workloadName)
                );

                expect(workloadComputeDomains).toHaveLength(1);

                computeDomain = workloadComputeDomains[0];
                logger.info(`Found ComputeDomain: ${computeDomain.metadata?.name}`);

                expect(computeDomain.spec).toBeDefined();
                expect(computeDomain.spec.numNodes).toBe(replicas);

                expect(computeDomain.metadata?.ownerReferences).toHaveLength(1);
                expect(computeDomain.metadata?.ownerReferences?.[0].kind).toBe('WorkloadRunner');
                expect(computeDomain.metadata?.ownerReferences?.[0].controller).toBe(true);

                resourceClaimTemplateName = computeDomain.spec.channel.resourceClaimTemplate.name;
                logger.info(`Verified ComputeDomain spec for workload: ${workloadName}`);
            });

            await test.step('Verify pod has resource claims for ComputeDomain', async () => {
                const pods = await kubeCli.listPodsByNamepsace(namespace);
                const workloadPods = pods.filter((pod) => pod.metadata.name.includes(workloadName));

                expect(workloadPods.length).toBeGreaterThan(0);

                const pod = workloadPods[0];
                logger.info(`Checking pod: ${pod.metadata.name} has resource claims for ComputeDomain`);

                expect(pod.spec.resourceClaims?.length ?? 0).toBeGreaterThan(0);
                const hasComputeDomainClaim = pod.spec.resourceClaims.some((claim: any) =>
                    claim.resourceClaimTemplateName?.includes(
                        computeDomain.spec.channel.resourceClaimTemplate.name
                    )
                );
                expect(hasComputeDomainClaim).toBe(true);
                expect(pod.spec.containers).toHaveLength(1);
                expect(pod.spec.containers[0].resources?.claims).toHaveLength(1);
                expect(pod.spec.containers[0].resources?.claims[0].name).toBe(
                    pod.spec.resourceClaims[0].name
                );

                logger.info('Verified pod has resource claims for ComputeDomain');
            });

            await test.step('Verify pod has node affinity for GPU network acceleration', async () => {
                const pods = await kubeCli.listPodsByNamepsace(namespace);
                const workloadPods = pods.filter((pod) => pod.metadata.name.includes(workloadName));
                const pod = workloadPods[0];

                expect(pod.spec.affinity).toBeDefined();
                expect(pod.spec.affinity.nodeAffinity).toBeDefined();

                const nodeAffinity = pod.spec.affinity.nodeAffinity;
                const hasGpuAccelerationAffinity =
                    nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution?.nodeSelectorTerms?.some(
                        (term: any) =>
                            term.matchExpressions?.some(
                                (expr: any) =>
                                    expr.key === MnnvlUtils.mnnvlLabel && expr.operator === 'Exists'
                            )
                    );

                expect(hasGpuAccelerationAffinity).toBe(true);

                logger.info('Verified pod has node affinity for GPU network acceleration');
            });
        });

        test('mnnvlMode=required with non-MNNVL nodepool: returns 400', async function () {
            const workloadName = objectGenerator.generateNameByPrefix('mnnvl-invalid');

            await test.step('Attempt to create workload on non-MNNVL nodepool - expect 400', async () => {
                const manifest = createDeploymentManifest(workloadName, 2);

                const createRequest: WorkloadV2CreateRequest = {
                    metadata: {
                        name: workloadName,
                        projectId: String(nonMnnvlProject.id),
                        configuration: {
                            mnnvl: MnnvlMode.Required
                        }
                    },
                    manifest
                };

                logger.info(
                    `Attempting to create workload with mnnvlMode=Required on non-MNNVL nodepool: ${workloadName}`
                );

                let error: HttpError;
                try {
                    await workloadsManagerService.workloadsV2Api.createWorkloadV2(createRequest);
                } catch (err) {
                    error = err;
                }
                expect(error, 'Expected request to fail with 400 Bad Request').toBeDefined();
                expect(error.statusCode).toBe(StatusCodes.BAD_REQUEST);

                logger.info(
                    `Verified 400 error for mnnvlMode=Required with non-MNNVL nodepool: ${error.message}`
                );
            });
        });
    });
});
