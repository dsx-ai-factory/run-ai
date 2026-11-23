/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    flowRootState,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    waitUtils
} from '../../../../../../modules/helpers';
import { assetsApi } from '../../../../../../modules/api';
import { GenericSecret, Scope } from '../../../../../../modules/api/apis/assetsApi';
import { test } from '@playwright/test';

const scopes = [Scope.Project, Scope.Department, Scope.Cluster];

test.describe(
    `Generic Secrets API, tags: ${tags.tool.api} ${tags.resource.credential} ${tags.teams.viper}`,
    { tag: [tags.services.assets, tags.services.k8sObjectsTracker, tags.runOn.parallelCI] },
    function () {
        let project: Project;
        let secretName: string;

        test.beforeAll(async () => {
            project = await projectResourcesUtils.createProjectAndNamespace('test-generic-secrets');
            assetsApi.credentialsApi.accessToken = flowRootState.users.system_administrator.accessToken;
        });
        test.afterAll(async () => {
            await operations.deleteProjectAndNamespace(project.id);
        });
        test.beforeEach(() => {
            secretName = objectGenerator.generateNameByPrefix('generic-secret');
        });

        async function _createGenericSecret(scope: Scope): Promise<GenericSecret> {
            const secret = {
                meta: {
                    name: secretName,
                    scope: scope
                },
                spec: {
                    keyValuePairs: [
                        { key: 'username', value: 'dG9t' },
                        { key: 'password', value: 'dG9t' }
                    ]
                }
            };

            if (scope === Scope.Project) {
                secret.meta['projectId'] = project.id;
            } else if (scope === Scope.Department) {
                secret.meta['departmentId'] = project.departmentId.toString();
            } else if (scope === Scope.Cluster) {
                secret.meta['clusterId'] = flowRootState.cluster.uuid;
            }

            return (await assetsApi.credentialsApi.createGenericSecret(secret)).body;
        }

        scopes.forEach(function (scope) {
            test(`${scope} scope`, async () => {
                const genericSecret = await _createGenericSecret(scope);

                // validate the secret that was created on the cluster
                await waitUtils.waitForSecret(`genericsecret-${secretName}`, project.namespace, 210);
                if (scope === Scope.Cluster)
                    await waitUtils.waitForSecret(`genericsecret-${secretName}`, 'runai', 90);

                await assetsApi.credentialsApi.deleteGenericSecret(genericSecret.meta.id);
            });
        });
    }
);
