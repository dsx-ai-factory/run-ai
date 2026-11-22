/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Scope } from '../../../../../../modules/api/apis/assetsApi';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    credentialUtils,
    dataSourcesUtils,
    flowRootState,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    storageUtils,
    tags,
    waitUtils
} from '../../../../../../modules/helpers';
import { kubeCli } from '../../../../../../modules/cli';
import { expect, test } from '@playwright/test';

const DENY_PVC_BODY =
    'username kubernetes-admin is not permitted to do any action on PersistentVolumeClaim with the run.ai/asset label';
const DENY_SECRET_BODY =
    'username kubernetes-admin is not permitted to do any action on Secret with the run.ai/asset label';

test.describe(
    `validate-secrets-pvc webhook deny, tags: ${tags.teams.viper} ${tags.resource.datasource} ${tags.resource.credential}`,
    { tag: [tags.services.agent, tags.runOn.parallelCI] },
    function () {
        let project: Project;
        let adminToken: string;

        test.beforeAll(async () => {
            adminToken = flowRootState.users.system_administrator.accessToken;
            project = await projectResourcesUtils.createProjectAndNamespace('admission-deny');
        });

        test.afterAll(async () => {
            await operations.deleteProjectAndNamespace(project.id);
        });

        test.afterEach(async function ({}, testInfo) {
            if (testInfo.status === 'failed') {
                logger.info(`test ${testInfo.title} failed, agent logs:`);
                await kubeCli.printLogs('agent');
            }
        });

        test('deny creating a PVC with run.ai/asset label', async function () {
            const pvcName = objectGenerator.generateShortNameByPrefix('pvc-deny-create');

            await expect(async () => {
                await storageUtils.createPvc(
                    storageUtils.buildLabeledPvc(pvcName, project.namespace)
                );
            }).rejects.toMatchObject({
                body: expect.stringContaining(DENY_PVC_BODY)
            });
        });

        test('deny creating a Secret with run.ai/asset label', async function () {
            const secretName = objectGenerator.generateShortNameByPrefix('secret-deny-create');

            await expect(async () => {
                await credentialUtils.createLabeledSecret(
                    credentialUtils.buildLabeledSecret(secretName, project.namespace)
                );
            }).rejects.toMatchObject({
                body: expect.stringContaining(DENY_SECRET_BODY)
            });
        });

        test('deny adding run.ai/asset label to an existing PVC', async function () {
            const pvcName = objectGenerator.generateShortNameByPrefix('pvc-deny-update');

            await test.step('Create unlabeled PVC via Kubernetes API (E2E kubeconfig)', async () => {
                await storageUtils.createPvc(
                    storageUtils.buildUnlabeledPvc(pvcName, project.namespace)
                );
            });

            await test.step('Patch PVC to add run.ai/asset label is denied', async () => {
                await expect(
                    kubeCli.patchPvc(
                        pvcName,
                        [{ op: 'add', path: '/metadata/labels/run.ai~1asset', value: 'pvc' }],
                        project.namespace
                    )
                ).rejects.toThrow();
                const pvc = await kubeCli.CoreV1Api.readNamespacedPersistentVolumeClaim({
                    name: pvcName,
                    namespace: project.namespace
                });
                expect(pvc.metadata?.labels?.['run.ai/asset']).toBeUndefined();
            });

            await test.step('Delete unlabeled PVC', async () => {
                await storageUtils.deletePvc(pvcName, project.namespace);
            });
        });

        test('deny adding run.ai/asset label to an existing Secret', async function () {
            const secretName = objectGenerator.generateShortNameByPrefix('secret-deny-update');

            await test.step('Create unlabeled Secret via Kubernetes API (E2E kubeconfig)', async () => {
                await credentialUtils.createLabeledSecret(
                    credentialUtils.buildUnlabeledSecret(secretName, project.namespace)
                );
            });

            await test.step('Patch Secret to add run.ai/asset label is denied', async () => {
                await expect(
                    kubeCli.patchSecret(
                        secretName,
                        [{ op: 'add', path: '/metadata/labels/run.ai~1asset', value: 'generic' }],
                        project.namespace
                    )
                ).rejects.toThrow();
                const secret = await kubeCli.getSecret(secretName, project.namespace);
                expect(secret.metadata?.labels?.['run.ai/asset']).toBeUndefined();
            });

            await test.step('Delete unlabeled Secret', async () => {
                await credentialUtils.deleteNamespacedSecret(secretName, project.namespace);
            });
        });

        test('deny deleting a labeled PVC', async function () {
            const pvcName = objectGenerator.generateShortNameByPrefix('pvc-deny-delete');
            let createdPvc: Awaited<ReturnType<typeof dataSourcesUtils.createPvcAsset>>;

            await test.step('Create PVC asset via Run:ai assets API', async () => {
                createdPvc = await dataSourcesUtils.createPvcAsset(
                    {
                        meta: {
                            name: pvcName,
                            scope: Scope.Project,
                            clusterId: flowRootState.cluster.uuid,
                            projectId: project.id
                        },
                        spec: {
                            path: '/home',
                            claimName: pvcName,
                            existingPvc: false,
                            claimInfo: {
                                size: '1M',
                                accessModes: { readWriteOnce: true },
                                storageClass: 'standard'
                            },
                            readOnly: false
                        }
                    },
                    adminToken
                );
            });

            await test.step('Wait for agent to create labeled PVC in cluster', async () => {
                await waitUtils.waitForPvc(pvcName, project.namespace);
                const pvc = await kubeCli.CoreV1Api.readNamespacedPersistentVolumeClaim({
                    name: pvcName,
                    namespace: project.namespace
                });
                expect(pvc.metadata?.labels?.['run.ai/asset']).toBe('pvc');
            });

            await test.step('Delete labeled PVC via Kubernetes API is denied', async () => {
                await expect(async () => {
                    await storageUtils.deletePvc(pvcName, project.namespace);
                }).rejects.toMatchObject({
                    body: expect.stringContaining(DENY_PVC_BODY)
                });
            });

            await test.step('Delete PVC asset via Run:ai assets API', async () => {
                await dataSourcesUtils.deletePvcAsset(createdPvc.meta.id, adminToken);
            });
        });

        test('deny deleting a labeled Secret', async function () {
            const secretName = objectGenerator.generateNameByPrefix('generic-deny-delete');
            const k8sSecretName = `genericsecret-${secretName}`;
            let genericSecret: Awaited<ReturnType<typeof credentialUtils.createGenericSecretAsset>>;

            await test.step('Create generic Secret asset via Run:ai assets API', async () => {
                genericSecret = await credentialUtils.createGenericSecretAsset(
                    {
                        meta: { name: secretName, scope: Scope.Project, projectId: project.id },
                        spec: {
                            keyValuePairs: [
                                { key: 'username', value: 'dG9t' },
                                { key: 'password', value: 'dG9t' }
                            ]
                        }
                    },
                    adminToken
                );
            });

            await test.step('Wait for agent to create labeled Secret in cluster', async () => {
                await waitUtils.waitForSecret(k8sSecretName, project.namespace, 210);
                const secret = await kubeCli.getSecret(k8sSecretName, project.namespace);
                expect(secret.metadata?.labels?.['run.ai/asset']).toBe('generic');
            });

            await test.step('Delete labeled Secret via Kubernetes API is denied', async () => {
                await expect(async () => {
                    await kubeCli.CoreV1Api.deleteNamespacedSecret({
                        name: k8sSecretName,
                        namespace: project.namespace
                    });
                }).rejects.toMatchObject({
                    body: expect.stringContaining(DENY_SECRET_BODY)
                });
            });

            await test.step('Delete Secret asset via Run:ai assets API', async () => {
                await credentialUtils.deleteGenericSecretAsset(genericSecret.meta.id, adminToken);
            });
        });
    }
);
