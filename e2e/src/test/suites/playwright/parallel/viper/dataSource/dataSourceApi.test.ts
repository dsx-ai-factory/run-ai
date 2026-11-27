/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    credentialUtils,
    dataSourcesUtils,
    flowRootState,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags
} from '../../../../../../modules/helpers';
import { AssetKind, Scope } from '../../../../../../modules/api/apis/assetsApi';
import { test } from '@playwright/test';

test.describe(
    `Data source API, tags: ${tags.resource.datasource} ${tags.teams.viper} ${tags.runOn.k8sFlavCertKit} ${tags.runOn.k8sFlavCertKitTemp} `,
    { tag: [tags.services.assets, tags.runOn.parallelCI] },
    () => {
        let project: Project;
        let dataSourceName: string;

        const dataSourceTypes = [
            {
                type: AssetKind.Pvc,
                createMethod: dataSourcesUtils.createPvcAsset,
                getPayload: dataSourcesUtils.getPvcCreationPayload.bind(dataSourcesUtils),
                input: 'fillPvcNameInput',
                dataSourceLowerCase: 'pvc',
                deleteMethod: dataSourcesUtils.deletePvcAsset
            },
            {
                type: AssetKind.Git,
                createMethod: dataSourcesUtils.createGitAsset,
                getPayload: dataSourcesUtils.getGitCreationPayload.bind(dataSourcesUtils),
                input: 'fillGitNameInput',
                dataSourceLowerCase: 'git',
                deleteMethod: dataSourcesUtils.deleteGitAsset
            },
            {
                type: AssetKind.S3,
                createMethod: dataSourcesUtils.createS3Asset,
                getPayload: dataSourcesUtils.getS3CreationPayload.bind(dataSourcesUtils),
                input: 'fillS3NameInput',
                dataSourceLowerCase: 's3',
                deleteMethod: dataSourcesUtils.deleteS3Asset
            }
        ];
        const noIssuesFoundStatus = 'No issues found';

        test.beforeAll(async function () {
            project = await projectResourcesUtils.createProjectAndNamespace('project-for-data-sources');
        });

        test.afterAll(async function () {
            await operations.deleteProjectAndNamespace(project.id);
        });

        async function createCredentials(assetKind: AssetKind): Promise<string> {
            switch (assetKind) {
                case AssetKind.Git: {
                    //create user-pass credentials
                    return (await credentialUtils.createUserPasswordWithArbitraryFields()).meta.id;
                }
                case AssetKind.S3: {
                    //create access-key credentials
                    return (await credentialUtils.createAccessKeyWithArbitraryFields()).meta.id;
                }
                default: {
                    return undefined;
                }
            }
        }

        async function deleteCredentials(assetKind: AssetKind, id: string): Promise<void> {
            switch (assetKind) {
                case AssetKind.Git: {
                    return await credentialUtils.deletePasswordAsset(id);
                }
                case AssetKind.S3: {
                    return await credentialUtils.deleteAccessKeyAsset(id);
                }
            }
        }

        dataSourceTypes.forEach(function (params) {
            test(`${params.dataSourceLowerCase} - create and check status`, async function () {
                dataSourceName = objectGenerator.generateShortNameByPrefix(params.dataSourceLowerCase);
                let createRequestPayload;
                const credentialsId = await createCredentials(params.type);

                if (params.type === AssetKind.Pvc) {
                    createRequestPayload = params.getPayload({
                        name: dataSourceName,
                        scope: Scope.Cluster,
                        scopeId: flowRootState.cluster.uuid
                    });
                } else {
                    //need to send additional credentials param
                    createRequestPayload = params.getPayload(
                        {
                            name: dataSourceName,
                            scope: Scope.Cluster,
                            scopeId: flowRootState.cluster.uuid
                        },
                        credentialsId
                    );
                }
                const dataSource = await params.createMethod(createRequestPayload);
                await dataSourcesUtils.validateExpectedStatus(dataSourceName, noIssuesFoundStatus);
                await params.deleteMethod(dataSource.meta.id);
                await deleteCredentials(params.type, credentialsId);
            });
        });
    }
);
