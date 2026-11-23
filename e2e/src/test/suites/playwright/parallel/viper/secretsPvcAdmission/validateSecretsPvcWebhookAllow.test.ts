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
import { expect } from 'chai';
import { test } from '@playwright/test';
import { kubeCli } from '../../../../../../modules/cli';

test.describe(
    `validate-secrets-pvc webhook allow, tags: ${tags.teams.viper} ${tags.resource.datasource} ${tags.resource.credential}`,
    { tag: [tags.services.agent, tags.runOn.parallelCI] },
    function () {
        let project: Project;
        let adminToken: string;

        test.beforeAll(async () => {
            adminToken = flowRootState.users.system_administrator.accessToken;
            project = await projectResourcesUtils.createProjectAndNamespace('admission-allow');
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

        test('allow unlabeled PVC create and delete', async function () {
            const pvcName = objectGenerator.generateShortNameByPrefix('pvc-unlabeled-webhook');

            await test.step('Create unlabeled PVC via Kubernetes API (E2E kubeconfig)', async () => {
                await storageUtils.createPvc(
                    storageUtils.buildUnlabeledPvc(pvcName, project.namespace)
                );
            });

            await test.step('Delete unlabeled PVC via Kubernetes API (E2E kubeconfig)', async () => {
                await storageUtils.deletePvc(pvcName, project.namespace);
            });
        });

        test('allow unlabeled Secret create and delete', async function () {
            const secretName = objectGenerator.generateShortNameByPrefix('secret-unlabeled-webhook');

            await test.step('Create unlabeled Secret via Kubernetes API (E2E kubeconfig)', async () => {
                await credentialUtils.createLabeledSecret(
                    credentialUtils.buildUnlabeledSecret(secretName, project.namespace)
                );
            });

            await test.step('Delete unlabeled Secret via Kubernetes API (E2E kubeconfig)', async () => {
                await credentialUtils.deleteNamespacedSecret(secretName, project.namespace);
            });
        });
    }
);
