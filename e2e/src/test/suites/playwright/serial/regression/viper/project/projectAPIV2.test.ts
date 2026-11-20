/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    apiUtils,
    crds,
    crdUtils,
    flowRootState,
    logger,
    nodePoolUtils,
    objectGenerator,
    orgUnitUtils,
    projectResourcesUtils,
    tags,
    versionUtils,
    waitUtils
} from '../../../../../../../modules/helpers';
import { expect } from '@playwright/test';
import { StatusCodes } from 'http-status-codes';
import {
    Project,
    ProjectCreationRequest,
    Resources,
    ResourcesNodePool,
    DepartmentForList
} from '../../../../../../../modules/api/apis/orgUnitApi';
import { NodePool } from '../../../../../../../modules/api/apis/backendClusterApi';
import { MIN_OVER_QUOTA_WEIGHT } from '../../../../../../../modules/helpers/utils/OrgUnitUtils';
import { validateOrgUnitProject } from '../../../../../../../modules/helpers/Validations';
import { SupportedVersions } from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test } from '@playwright/test';

const crdTimeout = 60;

test.describe(
    `Projects advanced API V2, tags:  ${tags.resource.projects} ${tags.tool.api} ${tags.teams.viper}`,
    { tag: [tags.services.projectController, tags.services.orgUnitService, tags.runOn.serialCI] },
    () => {
        let defaultNodePoolId: string;
        let testProject: Project;
        let projectName: string;
        let createdNp: NodePool;
        const deservedGpus = 1;

        test.beforeAll(async () => {
            await nodePoolUtils.deleteAllNodePools();
            defaultNodePoolId = String((await nodePoolUtils.getDefaultNodePoolFromBackend()).id);

            await apiUtils.disableProjectOverQuotaFeature();
            await apiUtils.disableCpuQuotaFeature();

            projectName = objectGenerator.generateNameByPrefix('p');

            logger.debug(`Creating project ${projectName}`);
            testProject = await orgUnitUtils.createProjectAndValidate({
                schedulingRules: {
                    interactiveJobTimeLimitSeconds: 2,
                    interactiveJobMaxIdleDurationSeconds: 3,
                    interactiveJobPreemptIdleDurationSeconds: 4,
                    trainingJobMaxIdleDurationSeconds: 5,
                    trainingJobTimeLimitSeconds: 6
                },
                resources: [
                    {
                        nodePool: { id: defaultNodePoolId, name: 'default' },
                        gpu: {
                            deserved: deservedGpus,
                            limit: deservedGpus
                        }
                    }
                ],
                name: projectName,
                clusterId: flowRootState.cluster.uuid
            });
        });

        test.afterAll(async () => {
            try {
                if (createdNp) {
                    await nodePoolUtils.deleteNodePool(createdNp);
                }
            } catch {
                logger.error('Failed to delete node pool');
            }

            createdNp = null;
        });

        test('Add/delete nodepool and validate project resources extended', async () => {
            const npName = objectGenerator.generateNameByPrefix('np');
            const createdNpResponse = await nodePoolUtils.createBackendNodePool(npName, 'key-A', 'value-B');
            expect(createdNpResponse.response.statusCode).toEqual(StatusCodes.CREATED);
            createdNp = createdNpResponse.body;

            let getResponse = await orgUnitUtils.getProject(String(testProject.id));
            expect(getResponse.resources.length).toEqual(2);
            const resourcesSorted = getResponse.resources.sort((a, b) =>
                a.nodePool.id <= b.nodePool.id ? -1 : 1
            );
            expect(resourcesSorted[0].nodePool.id).toEqual(defaultNodePoolId);
            expect(resourcesSorted[1].nodePool.name).toEqual(createdNp.name);
            expect(resourcesSorted[1].gpu.deserved).toEqual(0);
            expect(resourcesSorted[1].gpu.limit).toEqual(-1);
            // adjustable weights off + deserved 0 → API floors overQuotaWeight to MIN_OVER_QUOTA_WEIGHT
            expect(resourcesSorted[1].gpu.overQuotaWeight).toEqual(MIN_OVER_QUOTA_WEIGHT);

            await nodePoolUtils.waitForNodePoolCrd(createdNp.name);

            await nodePoolUtils.deleteNodePool(
                { id: createdNp.id, name: createdNp.name, isDefault: false },
                false
            );
            createdNp = null;

            // Projects are updated synchronously on delete nodepool, once it become async need to update the test to have a pooling time for the project.
            getResponse = await orgUnitUtils.getProject(String(testProject.id));
            expect(getResponse.resources.length).toEqual(1);
            expect(getResponse.resources[0].nodePool.id).toEqual(defaultNodePoolId);
        });

        test.describe('GPU/CPU/MEMORY quota management V2', () => {
            let projectsCreated: Project[] = [];
            let nodePoolCreated: NodePool[] = [];
            let department: DepartmentForList;
            let defaultNodePool: NodePool;
            let isCpuFeatureOn = false;
            let isOverQuotaFeatureOn = false;
            const deservedGpu = 2;
            const deservedCpu = 4;
            const deservedMemory = 5;

            function initCreated(): void {
                projectsCreated = [];
                nodePoolCreated = [];
            }

            test.beforeAll(async () => {
                defaultNodePool = await nodePoolUtils.getDefaultNodePoolFromBackend();
                const departments = await orgUnitUtils.getDepartments();
                department = departments.find((department) => department.name === 'default');
            });

            async function testProjectQuota(
                resources: Resources[],
                expectedBackendResources: Map<string, Resources>,
                expectedQueueResources: any,
                nodePoolNames: string[]
            ): Promise<void> {
                const projectName = objectGenerator.generateNameByPrefix('p');

                // create project
                const projectRequest = {
                    name: projectName,
                    clusterId: flowRootState.cluster.uuid,
                    parentId: department.id,
                    resources: resources
                } as ProjectCreationRequest;

                logger.info(
                    'sending POST projects to org unit with request: ' + JSON.stringify(projectRequest)
                );

                const postResponse = await orgUnitUtils.createProjectAndValidate(projectRequest);
                projectsCreated.push(postResponse);

                const projectId = postResponse.id;
                // check backend project
                const getResponse = await orgUnitUtils.getProject(String(projectId));
                expect(getResponse).not.toBeNull();

                logger.info('validating backend project ' + projectName);
                validateOrgUnitProject(getResponse, projectName, department.id, expectedBackendResources);

                // check queues
                logger.info('validating default queue ' + projectName);
                const defaultQueue = await waitUtils.waitForCrd(crds.queueCrd, projectName, '', crdTimeout);
                expect(defaultQueue).not.toBeUndefined();
                expect(defaultQueue['metadata'].labels[nodePoolUtils.NodePoolNameLabel]).toBeUndefined();
                expect(defaultQueue['metadata'].labels['project']).toEqual(projectName);

                expectedQueueResources['default'] = crdUtils.convertQueueSpecToCrdVersion(
                    expectedQueueResources['default'],
                    isOverQuotaFeatureOn
                );
                if (versionUtils.isVersionSupported(flowRootState.cluster.version, SupportedVersions.V2_13)) {
                    defaultQueue['spec'].resources = crdUtils.roundMemoryQuotaToThreeDigitsAfterDecimal(
                        defaultQueue['spec'].resources
                    );
                }

                expect(defaultQueue['spec'].resources).toEqual(expectedQueueResources['default']);
                logger.info('done validating default queue ' + projectName);

                for (const nodePoolName of nodePoolNames) {
                    logger.debug('validating queue ' + projectName + '-' + nodePoolName);
                    const queue = await waitUtils.waitForCrd(
                        crds.queueCrd,
                        projectName + '-' + nodePoolName,
                        '',
                        crdTimeout
                    );
                    expect(queue).not.toBeUndefined();
                    expect(queue['metadata'].labels[nodePoolUtils.NodePoolNameLabel]).toEqual(nodePoolName);
                    expectedQueueResources[nodePoolName] = crdUtils.convertQueueSpecToCrdVersion(
                        expectedQueueResources[nodePoolName],
                        isOverQuotaFeatureOn
                    );
                    if (
                        versionUtils.isVersionSupported(
                            flowRootState.cluster.version,
                            SupportedVersions.V2_13
                        )
                    ) {
                        queue['spec'].resources = crdUtils.roundMemoryQuotaToThreeDigitsAfterDecimal(
                            queue['spec'].resources
                        );
                    }
                    expect(queue['spec'].resources).toEqual(expectedQueueResources[nodePoolName]);
                    logger.info('done validating queue ' + projectName + '-' + nodePoolName);
                }
            }

            test.describe('cpu resources feature off, overQuota feature off', () => {
                test.beforeAll(async () => {
                    await apiUtils.disableProjectOverQuotaFeature();
                    await apiUtils.disableCpuQuotaFeature();
                });

                test.beforeEach(() => {
                    initCreated();
                });

                test.afterEach(async () => {
                    try {
                        for (const project of projectsCreated) {
                            await orgUnitUtils.deleteProjectAndNamespace(project.id, project.name);
                        }
                    } catch (error) {
                        logger.error('error deleting project: ' + error);
                    }
                });

                async function testAllFeatureToggleOff(overQuotaWeight: number): Promise<void> {
                    const defaultNodePoolResources = [
                        orgUnitUtils.createResources(
                            isCpuFeatureOn,
                            isOverQuotaFeatureOn,
                            {
                                id: defaultNodePool.id.toString(),
                                name: defaultNodePool.name
                            } as ResourcesNodePool,
                            overQuotaWeight,
                            deservedGpu,
                            undefined,
                            undefined
                        )
                    ];
                    const expectedResources = orgUnitUtils.getExpectedResources(
                        isCpuFeatureOn,
                        isOverQuotaFeatureOn,
                        {
                            id: defaultNodePool.id.toString(),
                            name: defaultNodePool.name
                        } as ResourcesNodePool,
                        overQuotaWeight,
                        deservedGpu,
                        undefined,
                        undefined
                    );
                    const expectedQueueResources =
                        projectResourcesUtils.getExpectedQueueResourcesOverQuotaOff(
                            isCpuFeatureOn,
                            overQuotaWeight,
                            deservedGpu,
                            undefined,
                            undefined
                        );
                    await testProjectQuota(
                        defaultNodePoolResources,
                        new Map<string, Resources>([['default', expectedResources]]),
                        { default: expectedQueueResources },
                        []
                    );
                }

                test('create project with overQuota=0 with resources with default node pool resources', async () => {
                    await testAllFeatureToggleOff(0);
                });
                test('create project with overQuota=1 with resources with default node pool resources', async () => {
                    await testAllFeatureToggleOff(1);
                });
            });

            test.describe('cpu resources feature on, overQuota feature off, default node-pool resources', () => {
                test.beforeAll(async () => {
                    await apiUtils.enableCpuQuotaFeature();
                    await apiUtils.disableProjectOverQuotaFeature();
                    isCpuFeatureOn = true;
                });

                test.afterAll(async () => {
                    try {
                        await apiUtils.disableCpuQuotaFeature();
                        isCpuFeatureOn = false;
                    } catch (error) {
                        logger.error('error disabling cpu quota feature: ' + error);
                    }
                });

                test.beforeEach(() => {
                    initCreated();
                });

                test.afterEach(async () => {
                    try {
                        for (const project of projectsCreated) {
                            await orgUnitUtils.deleteProjectAndNamespace(project.id, project.name);
                        }
                    } catch (error) {
                        logger.error('error deleting project: ' + error);
                    }
                });

                async function testCpuQuotaFeatureOn(overQuotaWeight: number): Promise<void> {
                    const defaultNodePoolResources = [
                        orgUnitUtils.createResources(
                            isCpuFeatureOn,
                            isOverQuotaFeatureOn,
                            {
                                id: defaultNodePool.id.toString(),
                                name: defaultNodePool.name
                            } as ResourcesNodePool,
                            overQuotaWeight,
                            deservedGpu,
                            deservedCpu,
                            deservedMemory
                        )
                    ];
                    const expectedResources = orgUnitUtils.getExpectedResources(
                        isCpuFeatureOn,
                        isOverQuotaFeatureOn,
                        {
                            id: defaultNodePool.id.toString(),
                            name: defaultNodePool.name
                        } as ResourcesNodePool,
                        overQuotaWeight,
                        deservedGpu,
                        deservedCpu,
                        deservedMemory
                    );
                    const expectedQueueResources =
                        projectResourcesUtils.getExpectedQueueResourcesOverQuotaOff(
                            isCpuFeatureOn,
                            overQuotaWeight,
                            deservedGpu,
                            deservedCpu,
                            deservedMemory
                        );
                    await testProjectQuota(
                        defaultNodePoolResources,
                        new Map<string, Resources>([['default', expectedResources]]),
                        { default: expectedQueueResources },
                        []
                    );
                }

                test('create project with overQuota=0 with default node pool resources', async () => {
                    await testCpuQuotaFeatureOn(0);
                });
                test('create project with overQuota=1 with default node pool resources', async () => {
                    await testCpuQuotaFeatureOn(1);
                });
            });

            test.describe('cpu resources feature off, overQuota feature on, default node-pool resources', () => {
                test.beforeAll(async () => {
                    await apiUtils.enableProjectOverQuotaFeature();
                    await apiUtils.disableCpuQuotaFeature();
                    isOverQuotaFeatureOn = true;
                });

                test.afterAll(async () => {
                    try {
                        await apiUtils.disableProjectOverQuotaFeature();
                        isOverQuotaFeatureOn = false;
                    } catch (error) {
                        logger.error('error disabling project over quota feature: ' + error);
                    }
                });

                test.beforeEach(() => {
                    initCreated();
                });

                test.afterEach(async () => {
                    try {
                        for (const project of projectsCreated) {
                            await orgUnitUtils.deleteProjectAndNamespace(project.id, project.name);
                        }
                    } catch (error) {
                        logger.error('error deleting project: ' + error);
                    }
                });

                async function testProjectOverQuotaFeatureOn(overQuotaWeight: number): Promise<void> {
                    const deservedGpu = 2;
                    const defaultNodePoolResources = [
                        orgUnitUtils.createResources(
                            isCpuFeatureOn,
                            isOverQuotaFeatureOn,
                            {
                                id: defaultNodePool.id.toString(),
                                name: defaultNodePool.name
                            } as ResourcesNodePool,
                            overQuotaWeight,
                            deservedGpu,
                            undefined,
                            undefined
                        )
                    ];
                    const expectedResources = orgUnitUtils.getExpectedResources(
                        isCpuFeatureOn,
                        isOverQuotaFeatureOn,
                        {
                            id: defaultNodePool.id.toString(),
                            name: defaultNodePool.name
                        } as ResourcesNodePool,
                        overQuotaWeight,
                        deservedGpu,
                        undefined,
                        undefined
                    );
                    const expectedQueueResources =
                        projectResourcesUtils.getExpectedQueueResourcesWithOverQuotaOn(
                            isCpuFeatureOn,
                            overQuotaWeight,
                            deservedGpu,
                            undefined,
                            undefined
                        );
                    await testProjectQuota(
                        defaultNodePoolResources,
                        new Map<string, Resources>([['default', expectedResources]]),
                        { default: expectedQueueResources },
                        []
                    );
                }

                test('create project with overQuota with resources with default node pool resources', async () => {
                    await testProjectOverQuotaFeatureOn(1);
                });
                test('create project with overQuota no resources with default node pool resources', async () => {
                    await testProjectOverQuotaFeatureOn(0);
                });
            });

            test.describe('cpu resources feature on, overQuota feature on, default node-pool resources', () => {
                test.beforeAll(async () => {
                    await apiUtils.enableProjectOverQuotaFeature();
                    await apiUtils.enableCpuQuotaFeature();
                    isOverQuotaFeatureOn = true;
                    isCpuFeatureOn = true;
                });

                test.afterAll(async () => {
                    try {
                        await apiUtils.disableProjectOverQuotaFeature();
                        await apiUtils.disableCpuQuotaFeature();
                        isCpuFeatureOn = false;
                        isOverQuotaFeatureOn = false;
                    } catch (error) {
                        logger.error('error disabling project over quota feature: ' + error);
                    }
                });

                test.beforeEach(() => {
                    initCreated();
                });

                test.afterEach(async () => {
                    try {
                        for (const project of projectsCreated) {
                            await orgUnitUtils.deleteProjectAndNamespace(project.id, project.name);
                        }
                    } catch (error) {
                        logger.error('error deleting project: ' + error);
                    }
                });

                async function testProjectCpuQuotaOverQuotaFeatureOn(overQuotaWeight: number): Promise<void> {
                    const defaultNodePoolResources = [
                        orgUnitUtils.createResources(
                            isCpuFeatureOn,
                            isOverQuotaFeatureOn,
                            {
                                id: defaultNodePool.id.toString(),
                                name: defaultNodePool.name
                            } as ResourcesNodePool,
                            overQuotaWeight,
                            deservedGpu,
                            deservedCpu,
                            deservedMemory
                        )
                    ];
                    const expectedResources = orgUnitUtils.getExpectedResources(
                        isCpuFeatureOn,
                        isOverQuotaFeatureOn,
                        {
                            id: defaultNodePool.id.toString(),
                            name: defaultNodePool.name
                        } as ResourcesNodePool,
                        overQuotaWeight,
                        deservedGpu,
                        deservedCpu,
                        deservedMemory
                    );
                    const expectedQueueResources =
                        projectResourcesUtils.getExpectedQueueResourcesWithOverQuotaOn(
                            isCpuFeatureOn,
                            overQuotaWeight,
                            deservedGpu,
                            deservedCpu,
                            deservedMemory
                        );
                    await testProjectQuota(
                        defaultNodePoolResources,
                        new Map<string, Resources>([['default', expectedResources]]),
                        { default: expectedQueueResources },
                        []
                    );
                }

                test('create project with overQuota with default node pool resources', async () => {
                    await testProjectCpuQuotaOverQuotaFeatureOn(1);
                });
            });

            test.describe('cpu resources feature on, overQuota feature off, default & non-default node-pools', () => {
                test.beforeAll(async () => {
                    await apiUtils.disableProjectOverQuotaFeature();
                    await apiUtils.enableCpuQuotaFeature();
                    isCpuFeatureOn = true;
                });

                test.afterAll(async () => {
                    try {
                        await apiUtils.disableCpuQuotaFeature();
                        isCpuFeatureOn = false;
                    } catch (error) {
                        logger.error('error disabling project over quota feature: ' + error);
                    }
                });

                test.beforeEach(() => {
                    initCreated();
                });

                test.afterEach(async () => {
                    try {
                        for (const project of projectsCreated) {
                            await orgUnitUtils.deleteProjectAndNamespace(project.id, project.name);
                        }
                    } catch (error) {
                        logger.error('error deleting project: ' + error);
                    }
                    try {
                        for (const nodePool of nodePoolCreated) {
                            -(await nodePoolUtils.deleteNodePool(nodePool));
                        }
                        nodePoolCreated = [];
                    } catch (error) {
                        logger.error('error deleting nodepool: ' + error);
                    }
                });

                async function testCpuMemoryFeatureOn(overQuotaWeight: number): Promise<void> {
                    const nodePool = await nodePoolUtils.createNodePool(
                        objectGenerator.generateNameByPrefix('np'),
                        objectGenerator.generateNameByPrefix('labelKey'),
                        objectGenerator.generateNameByPrefix('labelValue')
                    );
                    nodePoolCreated.push(nodePool);
                    const nodePoolName = nodePool.name;

                    const defaultNodePoolResources = orgUnitUtils.createResources(
                        isCpuFeatureOn,
                        isOverQuotaFeatureOn,
                        {
                            id: defaultNodePool.id.toString(),
                            name: defaultNodePool.name
                        } as ResourcesNodePool,
                        overQuotaWeight,
                        deservedGpu,
                        deservedCpu,
                        deservedMemory
                    );
                    const newNodePoolResources = orgUnitUtils.createResources(
                        isCpuFeatureOn,
                        isOverQuotaFeatureOn,
                        {
                            id: nodePool.id.toString(),
                            name: nodePool.name
                        } as ResourcesNodePool,
                        overQuotaWeight,
                        0,
                        1,
                        2
                    );
                    const defaultExpectedResources = orgUnitUtils.getExpectedResources(
                        isCpuFeatureOn,
                        isOverQuotaFeatureOn,
                        {
                            id: defaultNodePool.id.toString(),
                            name: defaultNodePool.name
                        } as ResourcesNodePool,
                        overQuotaWeight,
                        deservedGpu,
                        deservedCpu,
                        deservedMemory
                    );
                    const newExpectedResources = orgUnitUtils.getExpectedResources(
                        isCpuFeatureOn,
                        isOverQuotaFeatureOn,
                        {
                            id: nodePool.id.toString(),
                            name: nodePool.name
                        } as ResourcesNodePool,
                        overQuotaWeight,
                        0,
                        1,
                        2
                    );
                    const expectedQueueResources = {};
                    expectedQueueResources['default'] =
                        projectResourcesUtils.getExpectedQueueResourcesOverQuotaOff(
                            isCpuFeatureOn,
                            overQuotaWeight,
                            deservedGpu,
                            deservedCpu,
                            deservedMemory
                        );
                    expectedQueueResources[nodePoolName] =
                        projectResourcesUtils.getExpectedQueueResourcesOverQuotaOff(
                            isCpuFeatureOn,
                            overQuotaWeight,
                            0,
                            1,
                            2
                        );
                    await testProjectQuota(
                        [defaultNodePoolResources, newNodePoolResources],
                        new Map<string, Resources>([
                            ['default', defaultExpectedResources],
                            [nodePoolName, newExpectedResources]
                        ]),
                        expectedQueueResources,
                        [nodePoolName]
                    );
                }

                test('create project with overQuota=0', async () => {
                    await testCpuMemoryFeatureOn(0);
                });

                test('create project with overQuota=1', async () => {
                    await testCpuMemoryFeatureOn(1);
                });
            });

            test.describe('cpu resources feature off, overQuota feature on, default & non-default node-pools', () => {
                test.beforeAll(async () => {
                    await apiUtils.enableProjectOverQuotaFeature();
                    isOverQuotaFeatureOn = true;
                    await apiUtils.disableCpuQuotaFeature();
                });

                test.afterAll(async () => {
                    try {
                        await apiUtils.disableProjectOverQuotaFeature();
                        isOverQuotaFeatureOn = false;
                    } catch (error) {
                        logger.error('error disabling project over quota feature: ' + error);
                    }
                });

                test.beforeEach(() => {
                    initCreated();
                });

                test.afterEach(async () => {
                    try {
                        for (const project of projectsCreated) {
                            await orgUnitUtils.deleteProjectAndNamespace(project.id, project.name);
                        }
                    } catch (error) {
                        logger.error('error deleting project: ' + error);
                    }
                    try {
                        for (const nodePool of nodePoolCreated) {
                            -(await nodePoolUtils.deleteNodePool(nodePool));
                        }
                        nodePoolCreated = [];
                    } catch (error) {
                        logger.error('error deleting nodepool: ' + error);
                    }
                });

                async function testNodePoolOverQuotaFeatureOn(overQuotaWeight: number): Promise<void> {
                    const nodePool = await nodePoolUtils.createNodePool(
                        objectGenerator.generateNameByPrefix('np'),
                        objectGenerator.generateNameByPrefix('labelKey'),
                        objectGenerator.generateNameByPrefix('labelValue')
                    );
                    nodePoolCreated.push(nodePool);
                    const nodePoolName = nodePool.name;

                    const defaultNodePoolResources = orgUnitUtils.createResources(
                        isCpuFeatureOn,
                        isOverQuotaFeatureOn,
                        {
                            id: defaultNodePool.id.toString(),
                            name: defaultNodePool.name
                        } as ResourcesNodePool,
                        overQuotaWeight,
                        deservedGpu,
                        undefined,
                        undefined
                    );
                    const newNodePoolResources = orgUnitUtils.createResources(
                        isCpuFeatureOn,
                        isOverQuotaFeatureOn,
                        {
                            id: nodePool.id.toString(),
                            name: nodePool.name
                        } as ResourcesNodePool,
                        overQuotaWeight,
                        0,
                        undefined,
                        undefined
                    );
                    const defaultExpectedResources = orgUnitUtils.getExpectedResources(
                        isCpuFeatureOn,
                        isOverQuotaFeatureOn,
                        {
                            id: defaultNodePool.id.toString(),
                            name: defaultNodePool.name
                        } as ResourcesNodePool,
                        overQuotaWeight,
                        deservedGpu,
                        undefined,
                        undefined
                    );
                    const newExpectedResources = orgUnitUtils.getExpectedResources(
                        isCpuFeatureOn,
                        isOverQuotaFeatureOn,
                        {
                            id: nodePool.id.toString(),
                            name: nodePool.name
                        } as ResourcesNodePool,
                        overQuotaWeight,
                        0,
                        undefined,
                        undefined
                    );
                    const expectedQueueResources = {};
                    expectedQueueResources['default'] =
                        projectResourcesUtils.getExpectedQueueResourcesWithOverQuotaOn(
                            isCpuFeatureOn,
                            overQuotaWeight,
                            deservedGpu,
                            undefined,
                            undefined
                        );
                    expectedQueueResources[nodePoolName] =
                        projectResourcesUtils.getExpectedQueueResourcesWithOverQuotaOn(
                            isCpuFeatureOn,
                            overQuotaWeight,
                            0,
                            undefined,
                            undefined
                        );
                    await testProjectQuota(
                        [defaultNodePoolResources, newNodePoolResources],
                        new Map<string, Resources>([
                            ['default', defaultExpectedResources],
                            [nodePoolName, newExpectedResources]
                        ]),
                        expectedQueueResources,
                        [nodePoolName]
                    );
                }

                test('create project with overQuota', async () => {
                    await testNodePoolOverQuotaFeatureOn(2);
                });
            });

            test.describe('cpu resources feature on, overQuota feature on, default & non-default node-pools', () => {
                test.beforeAll(async () => {
                    await apiUtils.enableProjectOverQuotaFeature();
                    isOverQuotaFeatureOn = true;
                    await apiUtils.enableCpuQuotaFeature();
                    isCpuFeatureOn = true;
                });

                test.afterAll(async () => {
                    try {
                        await apiUtils.disableProjectOverQuotaFeature();
                        isOverQuotaFeatureOn = false;
                        await apiUtils.disableCpuQuotaFeature();
                        isCpuFeatureOn = false;
                    } catch (error) {
                        logger.error('error disabling project over quota feature: ' + error);
                    }
                });

                test.beforeEach(() => {
                    initCreated();
                });

                test.afterEach(async () => {
                    try {
                        for (const project of projectsCreated) {
                            await orgUnitUtils.deleteProjectAndNamespace(project.id, project.name);
                        }
                    } catch (error) {
                        logger.error('error deleting project: ' + error);
                    }
                    try {
                        for (const nodePool of nodePoolCreated) {
                            -(await nodePoolUtils.deleteNodePool(nodePool));
                        }
                        nodePoolCreated = [];
                    } catch (error) {
                        logger.error('error deleting nodepool: ' + error);
                    }
                });

                async function testAllFeatureTogglesOn(overQuotaWeight: number): Promise<void> {
                    const nodePool = await nodePoolUtils.createNodePool(
                        objectGenerator.generateNameByPrefix('np'),
                        objectGenerator.generateNameByPrefix('labelKey'),
                        objectGenerator.generateNameByPrefix('labelValue')
                    );
                    nodePoolCreated.push(nodePool);
                    const nodePoolName = nodePool.name;

                    const defaultNodePoolResources = orgUnitUtils.createResources(
                        isCpuFeatureOn,
                        isOverQuotaFeatureOn,
                        {
                            id: defaultNodePool.id.toString(),
                            name: defaultNodePool.name
                        } as ResourcesNodePool,
                        overQuotaWeight,
                        deservedGpu,
                        deservedCpu,
                        deservedMemory
                    );
                    const newNodePoolResources = orgUnitUtils.createResources(
                        isCpuFeatureOn,
                        isOverQuotaFeatureOn,
                        {
                            id: nodePool.id.toString(),
                            name: nodePool.name
                        } as ResourcesNodePool,
                        overQuotaWeight,
                        0,
                        2,
                        1
                    );
                    const defaultExpectedResources = orgUnitUtils.getExpectedResources(
                        isCpuFeatureOn,
                        isOverQuotaFeatureOn,
                        {
                            id: defaultNodePool.id.toString(),
                            name: defaultNodePool.name
                        } as ResourcesNodePool,
                        overQuotaWeight,
                        deservedGpu,
                        deservedCpu,
                        deservedMemory
                    );
                    const newExpectedResources = orgUnitUtils.getExpectedResources(
                        isCpuFeatureOn,
                        isOverQuotaFeatureOn,
                        {
                            id: nodePool.id.toString(),
                            name: nodePool.name
                        } as ResourcesNodePool,
                        overQuotaWeight,
                        0,
                        2,
                        1
                    );
                    const expectedQueueResources = {};
                    expectedQueueResources['default'] =
                        projectResourcesUtils.getExpectedQueueResourcesWithOverQuotaOn(
                            isCpuFeatureOn,
                            overQuotaWeight,
                            deservedGpu,
                            deservedCpu,
                            deservedMemory
                        );
                    expectedQueueResources[nodePoolName] =
                        projectResourcesUtils.getExpectedQueueResourcesWithOverQuotaOn(
                            isCpuFeatureOn,
                            overQuotaWeight,
                            0,
                            2,
                            1
                        );
                    await testProjectQuota(
                        [defaultNodePoolResources, newNodePoolResources],
                        new Map<string, Resources>([
                            ['default', defaultExpectedResources],
                            [nodePoolName, newExpectedResources]
                        ]),
                        expectedQueueResources,
                        [nodePoolName]
                    );
                }

                test('create project with overQuota with resources', async () => {
                    await testAllFeatureTogglesOn(1);
                });
            });
        });
    }
);
