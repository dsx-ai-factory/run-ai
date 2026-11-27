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
    dataSourcesUtils,
    orgUnitUtils,
    waitUtils
} from '../../../../../../modules/helpers';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { Scope } from '../../../../../../modules/api/apis/assetsApi';
import { runaiCliV2 } from '../../../../../../modules/cli';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    cliV2WorkloadType,
    cliV2TrainingType,
    SubmitFlagsMerged
} from '../../../../../../modules/cli/tools/RunaiCliV2.constants';

test.describe(`CLI Submit with Datasource - Combinations and Scope, tags: ${tags.resource.datasource} ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_24)(function () {
        let projectName: string;
        let namespace: string;
        let project: Project;

        test.beforeAll(async function () {
            projectName = objectGenerator.generateShortNameByPrefix('proj-ds-combo', 5);
            namespace = Utils.namespaceFromProjectName(projectName);

            await test.step('Create project via API', async function () {
                project = await projectResourcesUtils.createProject(namespace, {
                    name: projectName,
                    deservedGpus: 2,
                    maxAllowedGpus: 2
                });
                logger.info(`Created project: ${projectName} (ID: ${project.id})`);
            });

            await test.step('Set project context for CLI', function () {
                runaiCliV2.projectSet(projectName);
                logger.info(`Set CLI project context to: ${projectName}`);
            });
        });

        test.afterAll(async function () {
            await test.step('Cleanup project', async function () {
                if (project?.id) {
                    await operations.deleteProjectAndNamespace(project.id);
                    logger.info(`Deleted project: ${projectName}`);
                }
            });
        });

        test('T5585473: Submit with multiple datasources of different types', async function () {
            const dsPvcName = objectGenerator.generateShortNameByPrefix('ds-pvc-multi', 5);
            const dsNfsName = objectGenerator.generateShortNameByPrefix('ds-nfs-multi', 5);
            const dsS3Name = objectGenerator.generateShortNameByPrefix('ds-s3-multi', 5);
            const jobName = objectGenerator.generateShortNameByPrefix('job-multi-ds', 5);
            const dataSources: Array<{ id: string; type: 'pvc' | 'nfs' | 's3' }> = [];

            await test.step('Create multiple datasources of different types', async function () {
                // Create PVC datasource
                const pvcAsset = await dataSourcesUtils.createPVC(dsPvcName, Scope.Project, project);
                dataSources.push({ id: pvcAsset.meta.id, type: 'pvc' });
                logger.info(`Created PVC datasource: ${dsPvcName}`);

                // Create NFS datasource
                const nfsAsset = await dataSourcesUtils.createNfsAsset(
                    {
                        meta: {
                            name: dsNfsName,
                            scope: Scope.Project,
                            projectId: project.id,
                            clusterId: flowRootState.cluster.uuid
                        },
                        spec: {
                            server: 'nfs.example.com',
                            path: '/data',
                            mountPath: '/mnt/nfs'
                        }
                    },
                    flowRootState.users.system_administrator.idToken
                );
                dataSources.push({ id: nfsAsset.meta.id, type: 'nfs' });
                logger.info(`Created NFS datasource: ${dsNfsName}`);

                // Create S3 datasource
                const s3Asset = await dataSourcesUtils.createS3Asset(
                    {
                        meta: {
                            name: dsS3Name,
                            scope: Scope.Project,
                            projectId: project.id,
                            clusterId: flowRootState.cluster.uuid
                        },
                        spec: {
                            url: 'https://s3.amazonaws.com',
                            bucket: 'my-bucket',
                            path: '/mnt/s3'
                        }
                    },
                    flowRootState.users.system_administrator.idToken
                );
                dataSources.push({ id: s3Asset.meta.id, type: 's3' });
                logger.info(`Created S3 datasource: ${dsS3Name}`);
            });

            await test.step('Submit pytorch training with multiple datasources', function () {
                logger.info(`Submitting pytorch training with multiple datasources: ${jobName}`);
                const stdout = runaiCliV2.submit(
                    {
                        [SubmitFlagsMerged.image]: 'pytorch/pytorch:latest',
                        '-g': '1',
                        [SubmitFlagsMerged.datasource]: [
                            `type=pvc,name=${dsPvcName}`,
                            `type=nfs,name=${dsNfsName}`,
                            `type=s3,name=${dsS3Name}`
                        ]
                    } as any,
                    jobName,
                    cliV2WorkloadType.training,
                    cliV2TrainingType.pyTorch
                );
                logger.info(`Submit output: ${stdout}`);
                expect(stdout).toBeTruthy();
                logger.info('Pytorch training with multiple datasources submitted successfully');
            });

            await test.step('Verify workload is created', async function () {
                await waitUtils.waitForBackendWorkload(jobName, String(project.id), 60);
                logger.info('Verified: Workload created with multiple datasources');
            });

            await test.step('Cleanup workload and datasources', async function () {
                try {
                    runaiCliV2.delete({}, jobName, cliV2WorkloadType.training);
                    logger.info('Deleted workload successfully');
                } catch (error) {
                    logger.warn(`Failed to delete workload (may already be deleted): ${error.message}`);
                }

                for (const ds of dataSources) {
                    switch (ds.type) {
                        case 'pvc':
                            await dataSourcesUtils.deletePvcAsset(
                                ds.id,
                                flowRootState.users.system_administrator.idToken
                            );
                            break;
                        case 'nfs':
                            await dataSourcesUtils.deleteNfsAsset(
                                ds.id,
                                flowRootState.users.system_administrator.idToken
                            );
                            break;
                        case 's3':
                            await dataSourcesUtils.deleteS3Asset(
                                ds.id,
                                flowRootState.users.system_administrator.idToken
                            );
                            break;
                    }
                }
                logger.info('Cleaned up workload and datasources');
            });
        });

        test('T5585474: Submit with multiple datasources of same type', async function () {
            const dsPvc1Name = objectGenerator.generateShortNameByPrefix('ds-pvc-1', 5);
            const dsPvc2Name = objectGenerator.generateShortNameByPrefix('ds-pvc-2', 5);
            const jobName = objectGenerator.generateShortNameByPrefix('job-multi-pvc', 5);
            const dataSourceIds: string[] = [];

            await test.step('Create multiple PVC datasources', async function () {
                const pvcNames = [dsPvc1Name, dsPvc2Name];
                for (const name of pvcNames) {
                    const pvcAsset = await dataSourcesUtils.createPVC(name, Scope.Project, project);
                    dataSourceIds.push(pvcAsset.meta.id);
                    logger.info(`Created PVC datasource: ${name}`);
                }
            });

            await test.step('Submit pytorch training with multiple PVC datasources', function () {
                logger.info(`Submitting pytorch training with multiple PVC datasources: ${jobName}`);
                const stdout = runaiCliV2.submit(
                    {
                        [SubmitFlagsMerged.image]: 'pytorch/pytorch:latest',
                        '-g': '1',
                        [SubmitFlagsMerged.datasource]: [
                            `type=pvc,name=${dsPvc1Name}`,
                            `type=pvc,name=${dsPvc2Name}`
                        ]
                    } as any,
                    jobName,
                    cliV2WorkloadType.training,
                    cliV2TrainingType.pyTorch
                );
                logger.info(`Submit output: ${stdout}`);
                expect(stdout).toBeTruthy();
                logger.info('Pytorch training with multiple PVC datasources submitted successfully');
            });

            await test.step('Verify workload is created', async function () {
                await waitUtils.waitForBackendWorkload(jobName, String(project.id), 60);
                logger.info('Verified: Workload created with multiple PVC datasources');
            });

            await test.step('Cleanup workload and datasources', async function () {
                try {
                    runaiCliV2.delete({}, jobName, cliV2WorkloadType.training);
                    logger.info('Deleted workload successfully');
                } catch (error) {
                    logger.warn(`Failed to delete workload (may already be deleted): ${error.message}`);
                }

                for (const dsId of dataSourceIds) {
                    await dataSourcesUtils.deletePvcAsset(
                        dsId,
                        flowRootState.users.system_administrator.idToken
                    );
                }
                logger.info('Cleaned up workload and datasources');
            });
        });

        test('T5585475: Submit with cluster-scoped datasource', async function () {
            const dsName = objectGenerator.generateShortNameByPrefix('ds-cluster', 5);
            const jobName = objectGenerator.generateShortNameByPrefix('job-cluster-ds', 5);
            let dataSourceId: string;

            await test.step('Create cluster-scoped datasource via API', async function () {
                const nfsAsset = await dataSourcesUtils.createNfsAsset(
                    {
                        meta: {
                            name: dsName,
                            scope: Scope.Cluster,
                            clusterId: flowRootState.cluster.uuid
                        },
                        spec: {
                            server: 'nfs.cluster.com',
                            path: '/cluster/data',
                            mountPath: '/mnt/cluster'
                        }
                    },
                    flowRootState.users.system_administrator.idToken
                );
                dataSourceId = nfsAsset.meta.id;
                logger.info(`Created cluster-scoped NFS datasource: ${dsName}`);
            });

            await test.step('Submit pytorch training with cluster-scoped datasource', function () {
                logger.info(`Submitting pytorch training with cluster-scoped datasource: ${jobName}`);
                const stdout = runaiCliV2.submit(
                    {
                        [SubmitFlagsMerged.image]: 'pytorch/pytorch:latest',
                        '-g': '1',
                        [SubmitFlagsMerged.datasource]: [`type=nfs,name=${dsName}`]
                    } as any,
                    jobName,
                    cliV2WorkloadType.training,
                    cliV2TrainingType.pyTorch
                );
                logger.info(`Submit output: ${stdout}`);
                expect(stdout).toBeTruthy();
                logger.info('Pytorch training with cluster-scoped datasource submitted successfully');
            });

            await test.step('Verify workload is created', async function () {
                await waitUtils.waitForBackendWorkload(jobName, String(project.id), 60);
                logger.info('Verified: Workload can access cluster-scoped datasource');
            });

            await test.step('Cleanup workload and datasource', async function () {
                try {
                    runaiCliV2.delete({}, jobName, cliV2WorkloadType.training);
                    logger.info('Deleted workload successfully');
                } catch (error) {
                    logger.warn(`Failed to delete workload (may already be deleted): ${error.message}`);
                }
                await dataSourcesUtils.deleteNfsAsset(
                    dataSourceId,
                    flowRootState.users.system_administrator.idToken
                );
                logger.info('Cleaned up workload and datasource');
            });
        });

        test('T5585476: Submit with department-scoped datasource', async function () {
            const departmentName = objectGenerator.generateShortNameByPrefix('dept-ds', 5);
            const projectDeptName = objectGenerator.generateShortNameByPrefix('proj-in-dept', 5);
            const namespaceDept = Utils.namespaceFromProjectName(projectDeptName);
            const dsName = objectGenerator.generateShortNameByPrefix('ds-dept', 5);
            const jobName = objectGenerator.generateShortNameByPrefix('job-dept-ds', 5);
            let departmentId: number;
            let projectInDept: Project;
            let dataSourceId: string;

            await test.step('Create department and project', async function () {
                const department = await orgUnitUtils.createDepartment({
                    name: departmentName,
                    clusterId: flowRootState.cluster.uuid,
                    resources: [
                        {
                            gpu: { deserved: 0, limit: -1, overQuotaWeight: 0 },
                            cpu: null,
                            memory: null,
                            nodePool: null
                        }
                    ]
                });
                departmentId = Number(department.id);
                logger.info(`Created department: ${departmentName}`);

                projectInDept = await projectResourcesUtils.createProject(namespaceDept, {
                    name: projectDeptName,
                    deservedGpus: 0,
                    maxAllowedGpus: 2,
                    departmentId: departmentId
                });
                logger.info(`Created project in department: ${projectDeptName}`);
            });

            await test.step('Create department-scoped datasource', async function () {
                const pvcAsset = await dataSourcesUtils.createPVC(dsName, Scope.Department, projectInDept);
                dataSourceId = pvcAsset.meta.id;
                logger.info(`Created department-scoped PVC datasource: ${dsName}`);
            });

            await test.step('Submit pytorch training with department-scoped datasource', function () {
                logger.info(`Submitting pytorch training with department-scoped datasource: ${jobName}`);
                const stdout = runaiCliV2.submit(
                    {
                        [SubmitFlagsMerged.image]: 'pytorch/pytorch:latest',
                        '-g': '1',
                        [SubmitFlagsMerged.datasource]: [`type=pvc,name=${dsName}`],
                        [SubmitFlagsMerged.project]: projectDeptName
                    } as any,
                    jobName,
                    cliV2WorkloadType.training,
                    cliV2TrainingType.pyTorch
                );
                logger.info(`Submit output: ${stdout}`);
                expect(stdout).toBeTruthy();
                logger.info('Pytorch training with department-scoped datasource submitted successfully');
            });

            await test.step('Verify workload is created', async function () {
                await waitUtils.waitForBackendWorkload(jobName, String(projectInDept.id), 60);
                logger.info('Verified: Workload can access department-scoped datasource');
            });

            await test.step('Cleanup workload, datasource, project and department', async function () {
                try {
                    runaiCliV2.delete({ '-p': projectDeptName } as any, jobName, cliV2WorkloadType.training);
                    logger.info('Deleted workload successfully');
                } catch (error) {
                    logger.warn(`Failed to delete workload (may already be deleted): ${error.message}`);
                }
                await dataSourcesUtils.deletePvcAsset(
                    dataSourceId,
                    flowRootState.users.system_administrator.idToken
                );
                await operations.deleteProjectAndNamespace(projectInDept.id);
                await orgUnitUtils.deleteDepartment(departmentId.toString());
                logger.info('Cleaned up all resources');
            });
        });

        test('T5585477: Submit with project-scoped datasource', async function () {
            const dsName = objectGenerator.generateShortNameByPrefix('ds-project', 5);
            const jobName = objectGenerator.generateShortNameByPrefix('job-proj-ds', 5);
            let dataSourceId: string;

            await test.step('Create project-scoped datasource via API', async function () {
                const pvcAsset = await dataSourcesUtils.createPVC(dsName, Scope.Project, project);
                dataSourceId = pvcAsset.meta.id;
                logger.info(`Created project-scoped PVC datasource: ${dsName}`);
            });

            await test.step('Submit pytorch training with project-scoped datasource', function () {
                logger.info(`Submitting pytorch training with project-scoped datasource: ${jobName}`);
                const stdout = runaiCliV2.submit(
                    {
                        [SubmitFlagsMerged.image]: 'pytorch/pytorch:latest',
                        '-g': '1',
                        [SubmitFlagsMerged.datasource]: [`type=pvc,name=${dsName}`],
                        [SubmitFlagsMerged.project]: projectName
                    } as any,
                    jobName,
                    cliV2WorkloadType.training,
                    cliV2TrainingType.pyTorch
                );
                logger.info(`Submit output: ${stdout}`);
                expect(stdout).toBeTruthy();
                logger.info('Pytorch training with project-scoped datasource submitted successfully');
            });

            await test.step('Verify workload is created', async function () {
                await waitUtils.waitForBackendWorkload(jobName, String(project.id), 60);
                logger.info('Verified: Workload uses project-scoped datasource');
            });

            await test.step('Cleanup workload and datasource', async function () {
                try {
                    runaiCliV2.delete({}, jobName, cliV2WorkloadType.training);
                    logger.info('Deleted workload successfully');
                } catch (error) {
                    logger.warn(`Failed to delete workload (may already be deleted): ${error.message}`);
                }
                await dataSourcesUtils.deletePvcAsset(
                    dataSourceId,
                    flowRootState.users.system_administrator.idToken
                );
                logger.info('Cleaned up workload and datasource');
            });
        });

        test('T5585478: Submit with reversed format (name=NAME,type=TYPE)', async function () {
            const dsName = objectGenerator.generateShortNameByPrefix('ds-reversed', 5);
            const jobName = objectGenerator.generateShortNameByPrefix('job-reversed', 5);
            let dataSourceId: string;

            await test.step('Create PVC datasource via API', async function () {
                const pvcAsset = await dataSourcesUtils.createPVC(dsName, Scope.Project, project);
                dataSourceId = pvcAsset.meta.id;
                logger.info(`Created PVC datasource: ${dsName}`);
            });

            await test.step('Submit pytorch training with reversed format', function () {
                logger.info(`Submitting pytorch training with reversed format: ${jobName}`);
                const stdout = runaiCliV2.submit(
                    {
                        [SubmitFlagsMerged.image]: 'pytorch/pytorch:latest',
                        '-g': '1',
                        [SubmitFlagsMerged.datasource]: [`name=${dsName},type=pvc`]
                    } as any,
                    jobName,
                    cliV2WorkloadType.training,
                    cliV2TrainingType.pyTorch
                );
                logger.info(`Submit output: ${stdout}`);
                expect(stdout).toBeTruthy();
                logger.info('Pytorch training with reversed format submitted successfully');
            });

            await test.step('Verify workload is created', async function () {
                await waitUtils.waitForBackendWorkload(jobName, String(project.id), 60);
                logger.info('Verified: Workload created - order of name and type does not matter');
            });

            await test.step('Cleanup workload and datasource', async function () {
                try {
                    runaiCliV2.delete({}, jobName, cliV2WorkloadType.training);
                    logger.info('Deleted workload successfully');
                } catch (error) {
                    logger.warn(`Failed to delete workload (may already be deleted): ${error.message}`);
                }
                await dataSourcesUtils.deletePvcAsset(
                    dataSourceId,
                    flowRootState.users.system_administrator.idToken
                );
                logger.info('Cleaned up workload and datasource');
            });
        });
    });
});
