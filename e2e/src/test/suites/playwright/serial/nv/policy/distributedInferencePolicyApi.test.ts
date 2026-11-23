/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import {
    Utils,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    flowRootState,
    policyUtils,
    departmentUtils
} from '../../../../../../modules/helpers';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { ScopeType } from '../../../../../../modules/api/apis/policyApi';
import { WorkloadType } from '../../../../../../modules/helpers/utils/policyUtils';
import { StatusCodes } from 'http-status-codes';

test.describe(`Distributed Inference Workload Policy API Tests, tags: ${tags.tool.api} ${tags.resource.policy} ${tags.teams.loong} ${tags.teams.viper}`, () => {
    supportedFrom(SupportedVersions.V2_23)(function () {
        let depmtId: number;
        let defaultDepmtId: number;
        let deptName: string;
        let projectNameList: Array<string>;
        let namespaceList: Array<string>;
        let projectIdList: Array<string>;
        let createdDistInfPolicyList: Array<{ scope: string; scopeId: string }> = [];

        const defaults = {
            nodePools: ['default'],
            category: 'string',
            priorityClass: 'string',
            restartPolicy: 'RecreateGroupOnPodRestart',
            servingPort: {
                port: 8080,
                protocol: 'http',
                authorizationType: 'public',
                exposeExternally: true,
                exposedUrl: 'string'
            },
            startupPolicy: 'LeaderCreated',
            workers: 4,
            replicas: 2,
            leader: {
                command: 'sleep infinity',
                image: 'ubuntu',
                compute: {
                    gpuDevicesRequest: 0,
                    cpuCoreRequest: 0.5,
                    cpuCoreLimit: 2,
                    cpuMemoryRequest: '20M',
                    cpuMemoryLimit: '30M',
                    largeShmRequest: false
                }
            },
            worker: {
                command: 'sleep infinity',
                image: 'alpine:latest',
                compute: {
                    gpuDevicesRequest: 0,
                    cpuCoreRequest: 0.5,
                    cpuCoreLimit: 2,
                    cpuMemoryRequest: '20M',
                    cpuMemoryLimit: '30M',
                    largeShmRequest: false
                }
            }
        };

        const rules = {
            nodePools: {
                required: true,
                canEdit: true
            },
            workers: {
                required: true,
                canEdit: true,
                step: 0,
                defaultFrom: {
                    field: 'string',
                    factor: 1
                }
            },
            replicas: {
                required: true,
                canEdit: true,
                step: 0,
                defaultFrom: {
                    field: 'string',
                    factor: 1
                }
            },
            leader: {
                command: {
                    required: true,
                    canEdit: false
                },
                image: {
                    required: true,
                    canEdit: false
                }
            },
            worker: {
                command: {
                    required: true,
                    canEdit: false
                },
                image: {
                    required: true,
                    canEdit: false
                }
            }
        };

        const patchedDefaults = {
            replicas: 1,
            workers: 2
        };

        const patchedRules = {
            replicas: {
                canEdit: false
            }
        };

        const updatedDefaults = {
            ...defaults,
            replicas: patchedDefaults.replicas,
            workers: patchedDefaults.workers
        };

        const updatedRules = {
            ...rules,
            replicas: {
                ...rules.replicas,
                canEdit: patchedRules.replicas.canEdit
            }
        };

        test.beforeAll(async function () {
            const depmt = await departmentUtils.createBackendDepartmentWithName(
                objectGenerator.generateShortNameByPrefix('dist-inf-policy-dept', 5)
            );
            depmtId = Number(depmt.body.id);
            deptName = depmt.body.name;
            defaultDepmtId = Number(
                (await departmentUtils.getDefaultDepartmentByCluster(flowRootState.cluster.uuid)).id
            );

            projectNameList = Array.from({ length: 2 }, () =>
                objectGenerator.generateShortNameByPrefix('dist-inf-policy-proj', 5)
            );
            namespaceList = projectNameList.map((name) => Utils.namespaceFromProjectName(name));

            const projectsToCreate = [
                { namespace: namespaceList[0], name: projectNameList[0], departmentId: depmtId },
                { namespace: namespaceList[1], name: projectNameList[1] }
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
            projectIdList = projResponse.map((proj) => proj.id.toString());
        });

        test.afterEach('deleting policies', async () => {
            await Promise.all(
                createdDistInfPolicyList.map(async (data) => {
                    try {
                        await policyUtils.deleteWorkloadPolicyV2(WorkloadType.DistributedInference, {
                            scope: data.scope.toString(),
                            scopeId: data.scopeId.toString()
                        });
                    } catch (error) {
                        logger.error(`failed to delete policy:\n ${error.message}`);
                    }
                })
            );
        });

        test.afterAll('deleting project and namespace', async () => {
            await Promise.all(
                projectIdList.map(async (projectId, i) => {
                    try {
                        await operations.deleteProjectAndNamespace(parseInt(projectId));
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

            await departmentUtils.deleteDepartment(depmtId);
        });

        test('T5306743: test cluster scoped distributed inference policy CRUD', async () => {
            await test.step('Create a cluster scoped inference policy', async function () {
                const response = await policyUtils.createPolicyV2(
                    flowRootState.cluster.name,
                    flowRootState.cluster.uuid,
                    WorkloadType.DistributedInference,
                    defaults,
                    rules,
                    ScopeType.Cluster
                );

                expect(response.response.statusCode).toEqual(StatusCodes.OK);

                createdDistInfPolicyList.push({
                    scope: ScopeType.Cluster.toString(),
                    scopeId: flowRootState.cluster.uuid
                });
            });

            await test.step('List the created inference policy for cluster scope', async function () {
                await policyUtils.validatePolicyBody(
                    WorkloadType.DistributedInference,
                    {
                        scope: ScopeType.Cluster.toString(),
                        scopeId: flowRootState.cluster.uuid
                    },
                    defaults,
                    rules
                );
            });

            await test.step('List the created inference policy for department scope', async function () {
                // 'policy' section is empty, and 'effective' section same as cluster's policy
                await policyUtils.validateSubDomainPolicyBody(
                    WorkloadType.DistributedInference,
                    {
                        scope: ScopeType.Department.toString(),
                        scopeId: depmtId.toString()
                    },
                    defaults,
                    rules
                );
            });

            await test.step('List the created inference policy for project scope', async function () {
                // 'policy' section is empty, and 'effective' section same as cluster's policy
                await policyUtils.validateSubDomainPolicyBody(
                    WorkloadType.DistributedInference,
                    {
                        scope: ScopeType.Project.toString(),
                        scopeId: projectIdList[0]
                    },
                    defaults,
                    rules
                );
            });

            await test.step('Update the inference policy', async function () {
                const response = await policyUtils.updatePolicyV2(
                    flowRootState.cluster.name,
                    flowRootState.cluster.uuid,
                    WorkloadType.DistributedInference,
                    updatedDefaults,
                    updatedRules,
                    ScopeType.Cluster
                );

                expect(response.response.statusCode).toEqual(StatusCodes.OK);
            });

            await test.step('Verify the updated inference policy', async function () {
                await policyUtils.validatePolicyBody(
                    WorkloadType.DistributedInference,
                    {
                        scope: ScopeType.Cluster.toString(),
                        scopeId: flowRootState.cluster.uuid
                    },
                    updatedDefaults,
                    updatedRules
                );
            });

            await test.step('Delete the inference policy', async function () {
                const statusCode = await policyUtils.deleteWorkloadPolicyV2(
                    WorkloadType.DistributedInference,
                    {
                        scope: ScopeType.Cluster.toString(),
                        scopeId: flowRootState.cluster.uuid
                    }
                );

                expect(statusCode).toEqual(StatusCodes.NO_CONTENT);

                createdDistInfPolicyList = createdDistInfPolicyList.filter((data) => {
                    data.scopeId !== flowRootState.cluster.uuid;
                });
            });

            await test.step('Verify there is no inference policy in project after testing', async function () {
                await policyUtils.validateNoWorkloadPolicy(WorkloadType.DistributedInference, {
                    scope: ScopeType.Cluster.toString(),
                    scopeId: flowRootState.cluster.uuid
                });
            });
        });

        test('T5306744: test department scoped distributed inference policy CRUD', async () => {
            await test.step('Create a department scoped inference policy', async function () {
                const response = await policyUtils.createPolicyV2(
                    deptName,
                    depmtId.toString(),
                    WorkloadType.DistributedInference,
                    defaults,
                    rules,
                    ScopeType.Department
                );

                expect(response.response.statusCode).toEqual(StatusCodes.OK);

                createdDistInfPolicyList.push({
                    scope: ScopeType.Department.toString(),
                    scopeId: depmtId.toString()
                });
            });

            await test.step('List the created inference policy under the department', async function () {
                await policyUtils.validatePolicyBody(
                    WorkloadType.DistributedInference,
                    {
                        scope: ScopeType.Department.toString(),
                        scopeId: depmtId.toString()
                    },
                    defaults,
                    rules
                );
            });

            await test.step('List the created inference policy out of department scope', async function () {
                // should not be able to see the policy
                await policyUtils.validateNoWorkloadPolicy(WorkloadType.DistributedInference, {
                    scope: ScopeType.Department.toString(),
                    scopeId: defaultDepmtId.toString()
                });
            });

            await test.step('List the created inference policy for project scope under the department', async function () {
                // 'policy' section is empty, and 'effective' section same as cluster's policy
                await policyUtils.validateSubDomainPolicyBody(
                    WorkloadType.DistributedInference,
                    {
                        scope: ScopeType.Project.toString(),
                        scopeId: projectIdList[0]
                    },
                    defaults,
                    rules
                );
            });

            await test.step('List the created inference policy for project scope out of the department', async function () {
                // should not be able to see the policy
                await policyUtils.validateNoWorkloadPolicy(WorkloadType.DistributedInference, {
                    scope: ScopeType.Project.toString(),
                    scopeId: projectIdList[0]
                });
            });

            await test.step('Update the inference policy', async function () {
                const response = await policyUtils.updatePolicyV2(
                    deptName,
                    depmtId.toString(),
                    WorkloadType.DistributedInference,
                    updatedDefaults,
                    updatedRules,
                    ScopeType.Department
                );

                expect(response.response.statusCode).toEqual(StatusCodes.OK);
            });

            await test.step('Verify the updated inference policy', async function () {
                await policyUtils.validatePolicyBody(
                    WorkloadType.DistributedInference,
                    {
                        scope: ScopeType.Department.toString(),
                        scopeId: depmtId.toString()
                    },
                    updatedDefaults,
                    updatedRules
                );
            });

            await test.step('Delete the inference policy', async function () {
                const statusCode = await policyUtils.deleteWorkloadPolicyV2(
                    WorkloadType.DistributedInference,
                    {
                        scope: ScopeType.Department.toString(),
                        scopeId: depmtId.toString()
                    }
                );

                expect(statusCode).toEqual(StatusCodes.NO_CONTENT);

                createdDistInfPolicyList = createdDistInfPolicyList.filter((data) => {
                    data.scopeId !== depmtId.toString();
                });
            });

            await test.step('Verify there is no inference policy in project after testing', async function () {
                await policyUtils.validateNoWorkloadPolicy(WorkloadType.DistributedInference, {
                    scope: ScopeType.Department.toString(),
                    scopeId: depmtId.toString()
                });
            });
        });

        test('T5306745: test project scoped distributed inference policy CRUD', async () => {
            await test.step('Create a project scoped inference policy', async function () {
                const response = await policyUtils.createPolicyV2(
                    projectNameList[0],
                    projectIdList[0],
                    WorkloadType.DistributedInference,
                    defaults,
                    rules,
                    ScopeType.Project
                );

                expect(response.response.statusCode).toEqual(StatusCodes.OK);

                createdDistInfPolicyList.push({
                    scope: ScopeType.Project.toString(),
                    scopeId: projectIdList[0]
                });
            });

            await test.step('List the created inference policy under the project', async function () {
                await policyUtils.validatePolicyBody(
                    WorkloadType.DistributedInference,
                    {
                        scope: ScopeType.Project.toString(),
                        scopeId: projectIdList[0]
                    },
                    defaults,
                    rules
                );
            });

            await test.step('List the created inference policy out of project scope', async function () {
                // should not be able to see the policy
                await policyUtils.validateNoWorkloadPolicy(WorkloadType.DistributedInference, {
                    scope: ScopeType.Project.toString(),
                    scopeId: projectIdList[1]
                });
            });

            await test.step('Update the inference policy', async function () {
                const response = await policyUtils.updatePolicyV2(
                    projectNameList[0],
                    projectIdList[0],
                    WorkloadType.DistributedInference,
                    updatedDefaults,
                    updatedRules,
                    ScopeType.Project
                );

                expect(response.response.statusCode).toEqual(StatusCodes.OK);
            });

            await test.step('Verify the updated inference policy', async function () {
                await policyUtils.validatePolicyBody(
                    WorkloadType.DistributedInference,
                    {
                        scope: ScopeType.Project.toString(),
                        scopeId: projectIdList[0]
                    },
                    updatedDefaults,
                    updatedRules
                );
            });

            await test.step('Delete the inference policy', async function () {
                const statusCode = await policyUtils.deleteWorkloadPolicyV2(
                    WorkloadType.DistributedInference,
                    {
                        scope: ScopeType.Project.toString(),
                        scopeId: projectIdList[0]
                    }
                );

                expect(statusCode).toEqual(StatusCodes.NO_CONTENT);

                createdDistInfPolicyList = createdDistInfPolicyList.filter((data) => {
                    data.scopeId !== projectIdList[0];
                });
            });

            await test.step('Verify there is no inference policy in project after testing', async function () {
                await policyUtils.validateNoWorkloadPolicy(WorkloadType.DistributedInference, {
                    scope: ScopeType.Project.toString(),
                    scopeId: projectIdList[0]
                });
            });
        });

        test('T5306746: test project-level distributed inference policy overrides cluster-level policy', async () => {
            await test.step('Create a project scoped inference policy', async function () {
                const response = await policyUtils.createPolicyV2(
                    projectNameList[0],
                    projectIdList[0],
                    WorkloadType.DistributedInference,
                    defaults,
                    rules,
                    ScopeType.Project
                );

                expect(response.response.statusCode).toEqual(StatusCodes.OK);

                createdDistInfPolicyList.push({
                    scope: ScopeType.Project.toString(),
                    scopeId: projectIdList[0]
                });
            });

            await test.step('Create a cluster scoped inference policy', async function () {
                const response = await policyUtils.createPolicyV2(
                    flowRootState.cluster.name,
                    flowRootState.cluster.uuid,
                    WorkloadType.DistributedInference,
                    updatedDefaults,
                    updatedRules,
                    ScopeType.Cluster
                );

                expect(response.response.statusCode).toEqual(StatusCodes.OK);

                createdDistInfPolicyList.push({
                    scope: ScopeType.Cluster.toString(),
                    scopeId: flowRootState.cluster.uuid
                });
            });

            await test.step('List the created inference policy under the project', async function () {
                await policyUtils.validatePolicyBody(
                    WorkloadType.DistributedInference,
                    {
                        scope: ScopeType.Project.toString(),
                        scopeId: projectIdList[0]
                    },
                    defaults,
                    rules
                );
            });

            await test.step('Delete the inference policies', async function () {
                const scopesToDelete = [
                    { scope: ScopeType.Project.toString(), scopeId: projectIdList[0] },
                    { scope: ScopeType.Cluster.toString(), scopeId: flowRootState.cluster.uuid }
                ];

                for (const { scope, scopeId } of scopesToDelete) {
                    const statusCode = await policyUtils.deleteWorkloadPolicyV2(
                        WorkloadType.DistributedInference,
                        { scope, scopeId }
                    );
                    expect(statusCode).toEqual(StatusCodes.NO_CONTENT);

                    createdDistInfPolicyList = createdDistInfPolicyList.filter(
                        (data) => data.scopeId !== scopeId
                    );
                }
            });

            await test.step('Verify there is no inference policy in project after testing', async function () {
                await policyUtils.validateNoWorkloadPolicy(WorkloadType.DistributedInference, {
                    scope: ScopeType.Project.toString(),
                    scopeId: projectIdList[0]
                });
            });
        });

        test('T5306747: test only allow 1 distributed inference policy per cluster/department/project', async () => {
            const testDatasets = [
                {
                    scope: ScopeType.Cluster,
                    scopeName: flowRootState.cluster.name,
                    scopeId: flowRootState.cluster.uuid
                },
                {
                    scope: ScopeType.Department,
                    scopeName: deptName,
                    scopeId: depmtId.toString()
                },
                {
                    scope: ScopeType.Project,
                    scopeName: projectNameList[0],
                    scopeId: projectIdList[0]
                }
            ];

            for (const data of testDatasets) {
                await test.step(`Create a ${data.scope.toString()} scoped inference policy`, async function () {
                    const response = await policyUtils.createPolicyV2(
                        data.scopeName,
                        data.scopeId,
                        WorkloadType.DistributedInference,
                        defaults,
                        rules,
                        data.scope
                    );

                    expect(response.response.statusCode).toEqual(StatusCodes.OK);

                    createdDistInfPolicyList.push({
                        scope: data.scope.toString(),
                        scopeId: data.scopeId
                    });
                });

                await test.step(`List inference policy under the ${data.scope.toString()} and verify`, async function () {
                    await policyUtils.validatePolicyBody(
                        WorkloadType.DistributedInference,
                        {
                            scope: data.scope.toString(),
                            scopeId: data.scopeId
                        },
                        defaults,
                        rules
                    );
                });

                await test.step(`Create another ${data.scope.toString()} scoped inference policy`, async function () {
                    const response = await policyUtils.createPolicyV2(
                        data.scopeName,
                        data.scopeId,
                        WorkloadType.DistributedInference,
                        updatedDefaults,
                        updatedRules,
                        data.scope
                    );

                    expect(response.response.statusCode).toEqual(StatusCodes.OK);

                    createdDistInfPolicyList.push({
                        scope: data.scope.toString(),
                        scopeId: data.scopeId
                    });
                });

                await test.step(`List inference policy under the ${data.scope.toString()} and verify the later created policy is taking effect`, async function () {
                    await policyUtils.validatePolicyBody(
                        WorkloadType.DistributedInference,
                        {
                            scope: data.scope.toString(),
                            scopeId: data.scopeId
                        },
                        updatedDefaults,
                        updatedRules
                    );
                });

                await test.step('Delete the inference policy', async function () {
                    const statusCode = await policyUtils.deleteWorkloadPolicyV2(
                        WorkloadType.DistributedInference,
                        {
                            scope: data.scope.toString(),
                            scopeId: data.scopeId
                        }
                    );

                    expect(statusCode).toEqual(StatusCodes.NO_CONTENT);

                    createdDistInfPolicyList = createdDistInfPolicyList.filter((item) => {
                        item.scopeId !== data.scopeId;
                    });
                });
            }
        });
    });
});
