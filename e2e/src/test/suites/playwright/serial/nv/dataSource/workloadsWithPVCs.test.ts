/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    dataSourcesUtils,
    flowRootState,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    Utils,
    workloadsCrudUtils,
    config,
    workloadsUtilsV2,
    departmentUtils
} from '../../../../../../modules/helpers';
import { Scope } from '../../../../../../modules/api/apis/assetsApi';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test, expect } from '@playwright/test';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { Department as DepartmentOU } from '../../../../../../modules/api/apis/orgUnitApi';

import { PVCAsset } from '../../../../../../modules/api/apis/assetsApi/generated/model/pVCAsset';
import { Phase } from '../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import { WorkloadType } from '../../../../../../modules/helpers/utils/WorkloadsCrudUtils';

test.describe(`Storage data sources E2E Tests, tags: ${tags.tool.ui} ${tags.resource.datasource} ${tags.teams.loong} ${tags.runOn.dgxc}`, function () {
    supportedFrom(SupportedVersions.V2_20)(function () {
        let projectName: string;
        let namespace: string;
        let project: Project;

        const storageClass = 'standard';
        const command = 'sleep infinity';
        test.beforeAll(async () => {
            projectName = objectGenerator.generateNameByPrefix('e2e-project');
            namespace = Utils.namespaceFromProjectName(projectName);
            project = await projectResourcesUtils.createProject(namespace, {
                name: projectName,
                deservedGpus: 1,
                maxAllowedGpus: 1
            });
            workloadsCrudUtils.accessToken = flowRootState.users.system_administrator.accessToken;
        });

        test.afterAll(async function () {
            if (project && project.id) {
                try {
                    await operations.deleteProjectAndNamespace(project.id);
                    logger.info(`deleted project ${projectName} and namespace ${namespace}`);
                } catch {
                    logger.error('failed to delete project ' + projectName + ' and namespace ' + namespace);
                }
            }
        });

        test.describe('test pvc storage with different workloads', function () {
            const dataSets = [
                {
                    testCase:
                        'T3738305: L1 researcher can select PVC storage for the single node training workload',
                    workloadType: WorkloadType.TRAINING
                },
                {
                    testCase:
                        'T3738307: L1 researcher can select PVC storage for the single node workspace workload',
                    workloadType: WorkloadType.WORKSPACE
                }
            ];
            for (const dataSet of dataSets) {
                test(dataSet.testCase, async function () {
                    const createdPvcs: Array<PVCAsset> = [];
                    let workloadResponse;

                    const pvcs = [
                        {
                            name: objectGenerator.generateNameByPrefix('pvc1-'),
                            scope: Scope.Project,
                            project: project,
                            path: '/mnt/pvc1',
                            size: '1G',
                            storageClass: storageClass
                        },
                        {
                            name: objectGenerator.generateNameByPrefix('pvc2-'),
                            scope: Scope.Project,
                            project: project,
                            path: '/mnt/pvc2',
                            size: '100M',
                            storageClass: storageClass
                        }
                    ];

                    await test.step('Create 2 data sources (PVCs)', async () => {
                        for (const pvc of pvcs) {
                            const createdPvc = await dataSourcesUtils.createPvcAsset({
                                meta: {
                                    name: pvc.name,
                                    scope: pvc.scope,
                                    projectId: pvc.project.id
                                },
                                spec: {
                                    path: pvc.path,
                                    claimName: pvc.name,
                                    existingPvc: false,
                                    claimInfo: {
                                        size: pvc.size,
                                        accessModes: { readWriteOnce: true },
                                        storageClass: pvc.storageClass
                                    },
                                    readOnly: false
                                }
                            });
                            await dataSourcesUtils.validateExpectedStatus(pvc.name, 'No issues found');
                            createdPvcs.push(createdPvc);
                        }
                    });

                    await test.step('Create a workload with the 2 data sources', async () => {
                        const pvcInstances = createdPvcs.map((pvc) => ({
                            claimName: pvc.spec.claimName,
                            path: pvc.spec.path,
                            existingPvc: true,
                            readOnly: false
                        }));

                        const job = {
                            name: 'workload-pvc-test',
                            scope: Scope.Project,
                            project: project,
                            spec: {
                                image: config.quickStartImage.ubuntu,
                                command: command,
                                storage: {
                                    pvc: pvcInstances
                                }
                            }
                        };

                        if (dataSet.workloadType === WorkloadType.TRAINING) {
                            const { body: training } = await workloadsCrudUtils.createTraining(
                                job.name,
                                true,
                                project.clusterUuid,
                                project.id.toString(),
                                job.spec
                            );
                            await workloadsCrudUtils.waitForWorkloadPhase(
                                training.workloadId,
                                Phase.Running,
                                WorkloadType.TRAINING,
                                undefined,
                                120
                            );
                            workloadResponse = training;
                        } else if (dataSet.workloadType === WorkloadType.WORKSPACE) {
                            const { body: workspace } = await workloadsCrudUtils.createWorkspace(
                                job.name,
                                true,
                                project.clusterUuid,
                                project.id.toString(),
                                job.spec
                            );
                            await workloadsCrudUtils.waitForWorkloadPhase(
                                workspace.workloadId,
                                Phase.Running,
                                WorkloadType.WORKSPACE,
                                undefined,
                                120
                            );
                            workloadResponse = workspace;
                        }

                        const {
                            body: { pods }
                        } = await workloadsUtilsV2.getWorkloadPods(workloadResponse.workloadId);

                        for (const pod of pods) {
                            await dataSourcesUtils.verifyDataSourceRW(pod.name, namespace, '/mnt/pvc1');
                            await dataSourcesUtils.verifyDataSourceRW(pod.name, namespace, '/mnt/pvc2');
                        }
                    });

                    await test.step('Delete the created workload', async () => {
                        if (dataSet.workloadType === WorkloadType.TRAINING) {
                            await workloadsCrudUtils.deleteTraining(workloadResponse.workloadId);
                        } else if (dataSet.workloadType === WorkloadType.WORKSPACE) {
                            await workloadsCrudUtils.deleteWorkspace(workloadResponse.workloadId);
                        }
                    });

                    await test.step('Delete the created data sources', async () => {
                        await dataSourcesUtils.deleteAllPvcs();
                    });
                });
            }
        });

        test.describe('test workload using data source with prevent data modification enabled', function () {
            test('T4060320: L1 researcher can access the cluster workload with data source with prevent data modification enabled', async function () {
                const pvc = {
                    name: objectGenerator.generateNameByPrefix('pvc-'),
                    scope: Scope.Project,
                    project: project,
                    path: '/mnt/pvc1',
                    size: '1G',
                    storageClass: storageClass
                };
                let createdPvc;
                let workloadResponse;

                await test.step('Create a data source (PVC)', async () => {
                    createdPvc = await dataSourcesUtils.createPvcAsset({
                        meta: {
                            name: pvc.name,
                            scope: pvc.scope,
                            projectId: pvc.project.id
                        },
                        spec: {
                            path: pvc.path,
                            claimName: pvc.name,
                            existingPvc: false,
                            claimInfo: {
                                size: pvc.size,
                                accessModes: { readWriteOnce: true },
                                storageClass: pvc.storageClass
                            },
                            readOnly: true // set readOnly to true to prevent data modification
                        }
                    });
                    await dataSourcesUtils.validateExpectedStatus(pvc.name, 'No issues found');
                });

                await test.step('Create a workload with the data source', async () => {
                    const pvcInstance = {
                        claimName: createdPvc.spec.claimName,
                        path: createdPvc.spec.path,
                        existingPvc: true,
                        readOnly: true
                    };

                    const job = {
                        name: 'workload-pvc-test',
                        scope: Scope.Project,
                        project: project,
                        spec: {
                            image: config.quickStartImage.ubuntu,
                            command: command,
                            storage: {
                                pvc: [pvcInstance]
                            }
                        }
                    };

                    const { body: training } = await workloadsCrudUtils.createTraining(
                        job.name,
                        true,
                        project.clusterUuid,
                        project.id.toString(),
                        job.spec
                    );
                    await workloadsCrudUtils.waitForWorkloadPhase(
                        training.workloadId,
                        Phase.Running,
                        WorkloadType.TRAINING,
                        undefined,
                        120
                    );
                    workloadResponse = training;

                    const {
                        body: { pods }
                    } = await workloadsUtilsV2.getWorkloadPods(workloadResponse.workloadId);

                    for (const pod of pods) {
                        await dataSourcesUtils.verifyDataSourceRW(pod.name, namespace, '/mnt/pvc1', false);
                    }
                });

                await test.step('Delete the created workload', async () => {
                    await workloadsCrudUtils.deleteTraining(workloadResponse.workloadId);
                });

                await test.step('Delete the created data sources', async () => {
                    await dataSourcesUtils.deleteAllPvcs();
                });
            });
        });

        test.describe('workload can use the data source in proper scope', function () {
            test('T3990200: workload can use the data source in proper scope', async function () {
                const departmentName = objectGenerator.generateNameByPrefix('department-');
                const departmentProjectName = objectGenerator.generateNameByPrefix('project-');
                let newdepartment: DepartmentOU;
                let departmentProject: Project;
                let departmentId: number;
                let departmentProjectId: number;
                let projectPvcs: Array<PVCAsset> = [];
                let workloadResponse;

                await test.step('Create a department and create a project under the department', async () => {
                    const { body: department } =
                        await departmentUtils.createBackendDepartmentWithName(departmentName);
                    newdepartment = department;
                    departmentId = Number(department.id);
                    const dpproject = await projectResourcesUtils.createProject(namespace, {
                        name: departmentProjectName,
                        deservedGpus: 1,
                        maxAllowedGpus: 1,
                        departmentId: departmentId
                    });
                    departmentProject = dpproject;
                    departmentProjectId = dpproject.id;
                });

                await test.step('Create several data source (PVC) with different scopes', async () => {
                    const pvcDefaults = {
                        size: '100M',
                        storageClass: storageClass
                    };
                    const pvcs = [
                        {
                            ...pvcDefaults,
                            name: objectGenerator.generateNameByPrefix('pvc-default-project-'),
                            scope: Scope.Project,
                            project: project,
                            path: '/mnt/pvc1'
                        },
                        {
                            ...pvcDefaults,
                            name: objectGenerator.generateNameByPrefix('pvc-default-department-'),
                            scope: Scope.Department,
                            project: project,
                            path: '/mnt/pvc2'
                        },
                        {
                            ...pvcDefaults,
                            name: objectGenerator.generateNameByPrefix('pvc-new-department-'),
                            scope: Scope.Department,
                            project: departmentProject,
                            path: '/mnt/pvc3'
                        },
                        {
                            ...pvcDefaults,
                            name: objectGenerator.generateNameByPrefix('pvc-new-department-project-'),
                            scope: Scope.Project,
                            project: departmentProject,
                            path: '/mnt/pvc4'
                        },
                        {
                            ...pvcDefaults,
                            name: objectGenerator.generateNameByPrefix('pvc-cluster-'),
                            scope: Scope.Cluster,
                            project: project,
                            path: '/mnt/pvc5'
                        }
                    ];
                    for (const pvc of pvcs) {
                        const meta: any = {
                            name: pvc.name,
                            scope: pvc.scope
                        };
                        if (pvc.scope === Scope.Cluster) meta.clusterId = flowRootState.cluster.uuid;
                        if (pvc.scope === Scope.Project) meta.projectId = pvc.project.id;
                        if (pvc.scope === Scope.Department)
                            meta.departmentId = pvc.project.departmentId?.toString();

                        await dataSourcesUtils.createPvcAsset({
                            meta,
                            spec: {
                                path: pvc.path,
                                claimName: pvc.name,
                                existingPvc: false,
                                claimInfo: {
                                    size: pvc.size,
                                    accessModes: { readWriteOnce: true },
                                    storageClass: pvc.storageClass
                                },
                                readOnly: false
                            }
                        });
                        await dataSourcesUtils.validateExpectedStatus(pvc.name, 'No issues found');
                    }
                });

                await test.step('Verify PVC list contains correct data sources', async () => {
                    projectPvcs = await dataSourcesUtils.getPvcsByProjectId(project.id, 'Training');
                    const pvcNames = projectPvcs.map((pvc) => pvc.meta.name);
                    logger.info(`PVC names found: ${pvcNames}`);

                    ['pvc-default-project-', 'pvc-default-department-', 'pvc-cluster-'].forEach((prefix) => {
                        if (!pvcNames.some((name) => name.includes(prefix))) {
                            throw new Error(`Expected PVC with prefix '${prefix}' not found`);
                        }
                    });

                    ['pvc-new-department-', 'pvc-new-department-project-'].forEach((prefix) => {
                        if (pvcNames.some((name) => name.includes(prefix))) {
                            throw new Error(`Unexpected PVC with prefix '${prefix}' found`);
                        }
                    });
                });

                await test.step('Create a workload with the data sources', async () => {
                    const pvcInstances = projectPvcs.map((pvc) => ({
                        claimName: pvc.spec.claimName,
                        path: pvc.spec.path,
                        existingPvc: true,
                        readOnly: false
                    }));

                    const job = {
                        name: 'workload-pvc-test',
                        scope: Scope.Project,
                        project: project,
                        spec: {
                            image: config.quickStartImage.ubuntu,
                            command: command,
                            storage: {
                                pvc: pvcInstances
                            }
                        }
                    };

                    const { body: training } = await workloadsCrudUtils.createTraining(
                        job.name,
                        true,
                        project.clusterUuid,
                        project.id.toString(),
                        job.spec
                    );
                    await workloadsCrudUtils.waitForWorkloadPhase(
                        training.workloadId,
                        Phase.Running,
                        WorkloadType.TRAINING,
                        undefined,
                        180
                    );
                    workloadResponse = training;

                    const {
                        body: { pods }
                    } = await workloadsUtilsV2.getWorkloadPods(workloadResponse.workloadId);

                    for (const pod of pods) {
                        await dataSourcesUtils.verifyDataSourceRW(pod.name, namespace, '/mnt/pvc1');
                        await dataSourcesUtils.verifyDataSourceRW(pod.name, namespace, '/mnt/pvc2');
                        await dataSourcesUtils.verifyDataSourceRW(pod.name, namespace, '/mnt/pvc5');
                    }
                });

                await test.step('Delete the created workload', async () => {
                    await workloadsCrudUtils.deleteTraining(workloadResponse.workloadId);
                });

                await test.step('Delete the created data sources', async () => {
                    await dataSourcesUtils.deleteAllPvcs();
                });
            });
        });
    });
});
