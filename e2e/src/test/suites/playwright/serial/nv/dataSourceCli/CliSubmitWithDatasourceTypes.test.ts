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
    credentialUtils,
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
    cliV2TrainingType
} from '../../../../../../modules/cli/tools/RunaiCliV2.constants';

test.describe(`CLI Submit with Datasource - Datasource Types, tags: ${tags.resource.datasource} ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_24)(function () {
        let projectName: string;
        let namespace: string;
        let project: Project;

        test.beforeAll(async function () {
            projectName = objectGenerator.generateShortNameByPrefix('proj-ds-types', 5);
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

        // Helper function: Create datasource
        async function createDatasource(
            type: string,
            dsName: string
        ): Promise<{
            dataSourceId: string;
            additionalIds?: { credentialId?: string; configMapName?: string };
        }> {
            let dataSourceId: string;
            const additionalIds: { credentialId?: string; configMapName?: string } = {};

            switch (type) {
                case 'pvc': {
                    const pvcAsset = await dataSourcesUtils.createPVC(dsName, Scope.Project, project);
                    dataSourceId = pvcAsset.meta.id;
                    break;
                }

                case 'nfs': {
                    const nfsAsset = await dataSourcesUtils.createNfsAsset(
                        {
                            meta: {
                                name: dsName,
                                scope: Scope.Project,
                                projectId: project.id,
                                clusterId: flowRootState.cluster.uuid
                            },
                            spec: {
                                server: 'nfs.example.com',
                                path: '/exports/data',
                                mountPath: '/mnt/nfs'
                            }
                        },
                        flowRootState.users.system_administrator.idToken
                    );
                    dataSourceId = nfsAsset.meta.id;
                    break;
                }

                case 's3': {
                    const s3Asset = await dataSourcesUtils.createS3Asset(
                        {
                            meta: {
                                name: dsName,
                                scope: Scope.Project,
                                projectId: project.id,
                                clusterId: flowRootState.cluster.uuid
                            },
                            spec: {
                                url: 'https://s3.amazonaws.com',
                                bucket: 'my-ml-bucket',
                                path: '/mnt/s3'
                            }
                        },
                        flowRootState.users.system_administrator.idToken
                    );
                    dataSourceId = s3Asset.meta.id;
                    break;
                }

                case 'git': {
                    const gitAsset = await dataSourcesUtils.createGitAsset(
                        {
                            meta: {
                                name: dsName,
                                scope: Scope.Project,
                                projectId: project.id,
                                clusterId: flowRootState.cluster.uuid
                            },
                            spec: {
                                repository: 'https://github.com/pytorch/examples.git',
                                branch: 'main',
                                path: '/mnt/git'
                            }
                        },
                        flowRootState.users.system_administrator.idToken
                    );
                    dataSourceId = gitAsset.meta.id;
                    break;
                }

                case 'hostPath': {
                    const hostPathAsset = await dataSourcesUtils.createHostPathAsset(
                        {
                            meta: {
                                name: dsName,
                                scope: Scope.Project,
                                projectId: project.id,
                                clusterId: flowRootState.cluster.uuid
                            },
                            spec: {
                                path: '/mnt/shared-data',
                                mountPath: '/data'
                            }
                        },
                        flowRootState.users.system_administrator.idToken
                    );
                    dataSourceId = hostPathAsset.meta.id;
                    break;
                }

                case 'configMap': {
                    const configMapMetadata = await dataSourcesUtils.createConfigMapResource(
                        'app-config',
                        namespace
                    );
                    additionalIds.configMapName = configMapMetadata.name;
                    logger.info(`Created K8s ConfigMap: ${additionalIds.configMapName}`);

                    const configMapAsset = await dataSourcesUtils.createConfigMapAsset(
                        {
                            meta: {
                                name: dsName,
                                scope: Scope.Project,
                                projectId: project.id,
                                clusterId: flowRootState.cluster.uuid
                            },
                            spec: {
                                configMap: additionalIds.configMapName,
                                mountPath: '/config'
                            }
                        },
                        flowRootState.users.system_administrator.idToken
                    );
                    dataSourceId = configMapAsset.meta.id;
                    break;
                }

                case 'secretVolume': {
                    const secretName = objectGenerator.generateShortNameByPrefix('app-secrets', 5);
                    const credentialAsset = await credentialUtils.createGenericSecretAsset({
                        meta: {
                            name: secretName,
                            scope: Scope.Project,
                            projectId: project.id,
                            clusterId: flowRootState.cluster.uuid
                        },
                        spec: {
                            existingSecretName: null,
                            keyValuePairs: [
                                { key: 'username', value: 'testuser' },
                                { key: 'password', value: 'testpassword' }
                            ]
                        }
                    });
                    additionalIds.credentialId = credentialAsset.meta.id;
                    logger.info(`Created generic secret credential: ${secretName}`);

                    const secretAsset = await dataSourcesUtils.createSecretAsset(
                        {
                            meta: {
                                name: dsName,
                                scope: Scope.Project,
                                projectId: project.id,
                                clusterId: flowRootState.cluster.uuid
                            },
                            spec: {
                                credentialAssetId: additionalIds.credentialId,
                                mountPath: '/secrets'
                            }
                        },
                        flowRootState.users.system_administrator.idToken
                    );
                    dataSourceId = secretAsset.meta.id;
                    break;
                }

                default:
                    throw new Error(`Unknown datasource type: ${type}`);
            }

            logger.info(`Created ${type} datasource: ${dsName}`);
            return { dataSourceId, additionalIds };
        }

        // Helper function: Delete datasource
        async function deleteDatasource(type: string, dataSourceId: string): Promise<void> {
            switch (type) {
                case 'pvc':
                    await dataSourcesUtils.deletePvcAsset(
                        dataSourceId,
                        flowRootState.users.system_administrator.idToken
                    );
                    break;
                case 'nfs':
                    await dataSourcesUtils.deleteNfsAsset(
                        dataSourceId,
                        flowRootState.users.system_administrator.idToken
                    );
                    break;
                case 's3':
                    await dataSourcesUtils.deleteS3Asset(
                        dataSourceId,
                        flowRootState.users.system_administrator.idToken
                    );
                    break;
                case 'git':
                    await dataSourcesUtils.deleteGitAsset(
                        dataSourceId,
                        flowRootState.users.system_administrator.idToken
                    );
                    break;
                case 'hostPath':
                    await dataSourcesUtils.deleteHostPathAsset(
                        dataSourceId,
                        flowRootState.users.system_administrator.idToken
                    );
                    break;
                case 'configMap':
                    await dataSourcesUtils.deleteConfigMapAsset(
                        dataSourceId,
                        flowRootState.users.system_administrator.idToken
                    );
                    break;
                case 'secretVolume':
                    await dataSourcesUtils.deleteSecretAsset(
                        dataSourceId,
                        flowRootState.users.system_administrator.idToken
                    );
                    break;
            }
        }

        // Helper function: Cleanup workload and datasource
        async function cleanupWorkloadAndDatasource(
            jobName: string,
            type: string,
            dataSourceId: string
        ): Promise<void> {
            try {
                runaiCliV2.delete({}, jobName, cliV2WorkloadType.training);
                await deleteDatasource(type, dataSourceId);
                logger.info('Cleaned up workload and datasource');
            } catch (error) {
                logger.error('Cleanup error:', error);
            }
        }

        // Test configuration array
        const datasourceTestCases = [
            {
                id: 'T5585466',
                name: 'PVC datasource',
                type: 'pvc',
                prefix: 'pvc'
            },
            {
                id: 'T5585467',
                name: 'NFS datasource',
                type: 'nfs',
                prefix: 'nfs'
            },
            {
                id: 'T5585468',
                name: 'S3 datasource',
                type: 's3',
                prefix: 's3'
            },
            {
                id: 'T5585469',
                name: 'Git datasource',
                type: 'git',
                prefix: 'git'
            },
            {
                id: 'T5585470',
                name: 'hostPath datasource',
                type: 'hostPath',
                prefix: 'hostpath'
            },
            {
                id: 'T5585471',
                name: 'configMap datasource',
                type: 'configMap',
                prefix: 'configmap'
            },
            {
                id: 'T5585472',
                name: 'secretVolume datasource',
                type: 'secretVolume',
                prefix: 'secret'
            }
        ];

        // Generate test cases from configuration
        datasourceTestCases.forEach((testCase) => {
            test(`${testCase.id}: Submit pytorch training with ${testCase.name}`, async function () {
                const dsName = objectGenerator.generateShortNameByPrefix(`ds-${testCase.prefix}`, 5);
                const jobName = objectGenerator.generateShortNameByPrefix(`job-${testCase.prefix}`, 5);
                let dataSourceId: string;

                await test.step(`Create ${testCase.name} via API`, async function () {
                    const result = await createDatasource(testCase.type, dsName);
                    dataSourceId = result.dataSourceId;
                });

                await test.step(`Submit pytorch training with ${testCase.name}`, function () {
                    logger.info(`Submitting pytorch training with ${testCase.type}: ${jobName}`);
                    const stdout = runaiCliV2.submit(
                        {
                            '--image': 'pytorch/pytorch:latest',
                            '-g': '1',
                            '--datasource': [`type=${testCase.type},name=${dsName}`]
                        } as any,
                        jobName,
                        cliV2WorkloadType.training,
                        cliV2TrainingType.pyTorch
                    );
                    logger.info(`Submit output: ${stdout}`);
                    expect(stdout).toBeTruthy();
                    logger.info(`Pytorch training with ${testCase.name} submitted successfully`);
                });

                await test.step('Verify workload is created', async function () {
                    await waitUtils.waitForBackendWorkload(jobName, String(project.id), 60);
                    logger.info('Verified: Workload created successfully');
                });

                await test.step('Cleanup', async function () {
                    await cleanupWorkloadAndDatasource(jobName, testCase.type, dataSourceId);
                });
            });
        });
    });
});
