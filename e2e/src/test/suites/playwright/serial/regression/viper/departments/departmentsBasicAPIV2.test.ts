/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    apiUtils,
    crds,
    departmentUtils,
    flowRootState,
    logger,
    nodePoolUtils,
    objectGenerator,
    operations,
    orgUnitUtils,
    tags,
    waitUtils
} from '../../../../../../../modules/helpers';
import { expect } from '@playwright/test';
import { test } from '@playwright/test';
import { Department } from '../../../../../../../modules/api/apis/orgUnitApi';

test.describe(
    `Departments Basic API v2, tags:  ${tags.resource.department} ${tags.teams.viper}`,
    { tag: [tags.services.orgUnitService, tags.runOn.serialCI] },
    () => {
        let defaultNodePoolId: string;

        test.beforeAll(async () => {
            defaultNodePoolId = String((await nodePoolUtils.getDefaultNodePoolFromBackend()).id);
            await apiUtils.disableProjectOverQuotaFeature();
        });

        test.afterAll(async () => {
            try {
                await operations.forceCleanup();
            } catch (error) {
                logger.error('Failed to delete projects created in test', error);
            }
        });

        test.describe('Basic department CRUD validations', () => {
            let departmentName: string;
            let departmentId: string;
            let queueName: string;

            test.describe('CRUD department and validate crd+backend department fields', () => {
                test('Create department and validate', async () => {
                    departmentName = objectGenerator.generateNameByPrefix('d');

                    const res = await orgUnitUtils.createDepartment({
                        name: departmentName,
                        clusterId: flowRootState.cluster.uuid,
                        resources: [
                            {
                                nodePool: { id: defaultNodePoolId, name: 'default' },
                                gpu: { deserved: 10 }
                            }
                        ]
                    });
                    departmentId = res.id;
                    queueName = 'q-' + departmentId;

                    orgUnitUtils.validateDepartmentResponse(
                        res,
                        departmentName,
                        {
                            nodePool: { id: defaultNodePoolId, name: 'default' },
                            gpu: { deserved: 10 },
                            rank: 'MediumLow'
                        },
                        1
                    );

                    await departmentUtils.waitForDepartmentCrd(res.name);
                    await orgUnitUtils.validateDepartmentQueueCrd(queueName, res);
                });

                test('Update department and validate', async () => {
                    const depUpdateRes = await orgUnitUtils.updateDepartment(departmentId, {
                        resources: [
                            {
                                nodePool: { id: defaultNodePoolId, name: 'default' },
                                gpu: { deserved: 20 }
                            }
                        ]
                    });
                    const depGet = await orgUnitUtils.getDepartment(departmentId);

                    for (const dep of [depUpdateRes, depGet]) {
                        orgUnitUtils.validateDepartmentResponse(
                            dep,
                            departmentName,
                            {
                                nodePool: { id: defaultNodePoolId, name: 'default' },
                                gpu: { deserved: 20 },
                                rank: 'MediumLow'
                            },
                            1
                        );
                    }

                    await orgUnitUtils.validateDepartmentQueueCrd(queueName, depUpdateRes, 20, 30, true);
                });

                test('Patch department resources and validate', async () => {
                    const depPatchRes = await orgUnitUtils.patchDepartmentResources(departmentId, [
                        {
                            nodePool: { id: defaultNodePoolId, name: 'default' },
                            gpu: { deserved: 30, limit: 30 }
                        }
                    ]);
                    orgUnitUtils.validateResources(1, depPatchRes, {
                        nodePool: { id: defaultNodePoolId, name: 'default' },
                        gpu: { deserved: 30, limit: 30 }
                    });

                    await orgUnitUtils.validateDepartmentQueueCrd(
                        queueName,
                        {
                            id: departmentId,
                            description: '',
                            name: departmentName,
                            resources: [
                                {
                                    nodePool: { id: defaultNodePoolId, name: 'default' },
                                    gpu: { deserved: 30, limit: 30, overQuotaWeight: 30 },
                                    rank: 'MediumLow'
                                }
                            ],
                            clusterId: '',
                            createdAt: undefined,
                            totalResources: undefined,
                            updatedAt: undefined
                        },
                        30,
                        30,
                        true
                    );
                });

                test('Update department resources and validate', async () => {
                    const depUpdateRes = await orgUnitUtils.updateDepartmentResources(departmentId, [
                        {
                            nodePool: { id: defaultNodePoolId, name: 'default' },
                            gpu: { deserved: 40, limit: -1 }
                        }
                    ]);
                    orgUnitUtils.validateResources(1, depUpdateRes, {
                        nodePool: { id: defaultNodePoolId, name: 'default' },
                        gpu: { deserved: 40, limit: -1 }
                    });

                    await orgUnitUtils.validateDepartmentQueueCrd(
                        queueName,
                        {
                            id: departmentId,
                            description: '',
                            name: departmentName,
                            resources: [
                                {
                                    nodePool: { id: defaultNodePoolId, name: 'default' },
                                    gpu: { deserved: 40, limit: -1, overQuotaWeight: 40 },
                                    rank: 'MediumLow'
                                }
                            ],
                            clusterId: '',
                            createdAt: undefined,
                            totalResources: undefined,
                            updatedAt: undefined
                        },
                        40,
                        30,
                        true
                    );
                });

                test('Delete department and validate deleted', async () => {
                    let departments = await orgUnitUtils.getDepartments();
                    const numOfDepartmentBeforeDeletion = departments.length;
                    expect(numOfDepartmentBeforeDeletion).toBeGreaterThan(0);
                    const dep = departments.find((department) => department.name === departmentName);
                    expect(dep).toBeDefined();
                    const depForTest = {
                        ...dep,
                        description: ''
                    } as Department;

                    orgUnitUtils.validateDepartmentResponse(
                        depForTest,
                        departmentName,
                        {
                            nodePool: { id: defaultNodePoolId, name: 'default' },
                            gpu: { deserved: 40 },
                            rank: 'MediumLow'
                        },
                        1
                    );

                    await orgUnitUtils.deleteDepartment(departmentId);
                    await waitUtils.waitForCrd(crds.queueCrd, queueName, undefined, 60, false);

                    departments = await orgUnitUtils.getDepartments();
                    expect(departments.length).toBe(numOfDepartmentBeforeDeletion - 1);
                });
            });
        });
    }
);
