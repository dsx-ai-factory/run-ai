/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    assetsApiV2Utils,
    dataSourcesUtils,
    environmentUtils,
    flowRootState,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    waitUtils,
    workloadsCrudUtils
} from '../../../../../../modules/helpers';
import { assetsApi } from '../../../../../../modules/api';
import {
    AccessKey,
    AssetKind,
    S3Asset,
    S3CreationRequest,
    Scope,
    TrainingV2
} from '../../../../../../modules/api/apis/assetsApi';
import { expect } from 'chai';
import { test } from '@playwright/test';
import { kubeCli } from '../../../../../../modules/cli';

const scopes = [Scope.Project, Scope.Department];
const CONTAINER_PATH = '/test/path';

test.describe(
    `Access key API, tags: ${tags.tool.api} ${tags.resource.credential} ${tags.teams.viper}`,
    { tag: [tags.services.assets, tags.runOn.parallelCI] },
    function () {
        let project: Project;
        let accessKeyName: string;
        let s3Name: string;
        let trainingName: string;
        let createdS3: S3Asset;
        let serviceIp: string;

        async function _createTraining(): Promise<TrainingV2> {
            const workloadPayload = {
                meta: {
                    name: trainingName,
                    projectId: project.id.toString(),
                    clusterId: flowRootState.cluster.uuid
                },
                assets: {
                    environment: await environmentUtils.getId('tensorboard'),
                    datasources: [{ id: createdS3.meta.id, kind: AssetKind.S3 }]
                }
            };
            return (await assetsApiV2Utils.createTraining(workloadPayload)).body;
        }

        async function setupMinioAndBucket(): Promise<void> {
            await kubeCli.deployMinio(project.namespace);
            await kubeCli.createMinioService(project.namespace);
            await waitUtils.waitForService('minio-service', project.namespace);
            serviceIp = await kubeCli.getMinioServiceIP(project.namespace);
            await kubeCli.createS3Bucket('my-new-bucket', project.namespace);
        }

        async function teardownMinio(): Promise<void> {
            const results = await Promise.allSettled([
                kubeCli.deletePod('my-new-bucket', project.namespace),
                kubeCli.deleteMinioService(project.namespace),
                kubeCli.deleteMinio(project.namespace)
            ]);
            for (const result of results) {
                if (result.status === 'rejected') {
                    logger.warn(`MinIO cleanup failed: ${result.reason}`);
                }
            }
            try {
                await waitUtils.waitForCondition(
                    async () => {
                        const [pods, services, deployments] = await Promise.all([
                            kubeCli.CoreV1Api.listNamespacedPod({ namespace: project.namespace }),
                            kubeCli.CoreV1Api.listNamespacedService({ namespace: project.namespace }),
                            kubeCli.AppsV1Api.listNamespacedDeployment({ namespace: project.namespace })
                        ]);
                        return (
                            !pods.items.some(
                                (pod) =>
                                    pod.metadata?.name === 'my-new-bucket' ||
                                    pod.metadata?.labels?.app === 'minio-server'
                            ) &&
                            !services.items.some((service) => service.metadata?.name === 'minio-service') &&
                            !deployments.items.some(
                                (deployment) => deployment.metadata?.name === 'minio-server'
                            )
                        );
                    },
                    'MinIO resources to be removed before the next access-key case',
                    true,
                    30
                );
            } catch (error) {
                logger.warn(`MinIO cleanup did not finish: ${error.message}`);
            }
        }

        async function _createAccessKeyForMinio(scope: Scope): Promise<AccessKey> {
            const secretReq = {
                meta: {
                    name: accessKeyName,
                    scope: scope
                },
                spec: {
                    // minio's default creds
                    accessKeyId: 'minioadmin',
                    secretAccessKey: 'minioadmin'
                }
            };

            if (scope === Scope.Project) {
                secretReq.meta['projectId'] = project.id;
            } else if (scope === Scope.Department) {
                secretReq.meta['departmentId'] = project.departmentId.toString();
            }
            return (await assetsApi.credentialsApi.createAccessKey(secretReq)).body;
        }

        async function _createS3AssetMinio(secret: AccessKey, scope: Scope): Promise<S3Asset> {
            const s3Req: S3CreationRequest = {
                meta: {
                    name: s3Name,
                    scope: scope,
                    clusterId: scope === Scope.Project ? flowRootState.cluster.uuid : undefined,
                    projectId: scope === Scope.Project ? project.id : undefined,
                    departmentId: scope === Scope.Department ? project.departmentId.toString() : undefined
                },
                spec: {
                    bucket: 'my-s3',
                    path: CONTAINER_PATH,
                    url: `http://${serviceIp}:9000`,
                    accessKeyAssetId: secret.meta.id
                }
            };
            return await dataSourcesUtils.createS3Asset(s3Req);
        }

        test.beforeAll(async () => {
            project = await projectResourcesUtils.createProjectAndNamespace('test-ak');
            assetsApi.credentialsApi.accessToken = flowRootState.users.system_administrator.accessToken;
        });
        test.afterAll(async () => {
            await operations.deleteProjectAndNamespace(project.id);
        });
        test.beforeEach(() => {
            accessKeyName = objectGenerator.generateNameByPrefix('ak');
            s3Name = 's3-0';
            trainingName = objectGenerator.generateShortNameByPrefix('training');

            workloadsCrudUtils.accessToken = flowRootState.users.system_administrator.accessToken;
            assetsApiV2Utils.accessToken = flowRootState.users.system_administrator.accessToken;
        });

        scopes.forEach((scope) => {
            test(`create ${scope} scoped access-key`, async function ({}, testInfo) {
                test.setTimeout(210 * 1000);
                try {
                    // set up local minio and bucket to provide a valid S3 endpoint with correct access credentials for the test
                    await setupMinioAndBucket();

                    const secret = await _createAccessKeyForMinio(scope);
                    await waitUtils.waitForSecret(`accesskey-${accessKeyName}`, project.namespace, 120);

                    createdS3 = await _createS3AssetMinio(secret, scope);
                    const training = await _createTraining();

                    await waitUtils.waitForPodPrefix(training.meta.name, project.namespace);
                    const pod = (
                        await kubeCli.listPodsByNamespaceAndPrefix(project.namespace, training.meta.name)
                    )[0];
                    const container = pod.spec.containers[0];

                    const matchingMount = container.volumeMounts.find(
                        (vm) => vm.mountPath === CONTAINER_PATH
                    );
                    expect(matchingMount).to.not.be.undefined;
                    expect(matchingMount.name).to.equal(s3Name);

                    const volume = pod.spec.volumes.find((v) => v.name === s3Name);
                    expect(volume).to.not.be.undefined;

                    await workloadsCrudUtils.deleteTraining(training.meta.workloadId);
                    await dataSourcesUtils.deleteS3Asset(createdS3.meta.id);
                    await assetsApi.credentialsApi.deleteAccessKey(secret.meta.id);
                } catch (error) {
                    await kubeCli.logS3Diagnostics(project.namespace);
                    throw error;
                } finally {
                    await teardownMinio();
                }
            });
        });
    }
);
