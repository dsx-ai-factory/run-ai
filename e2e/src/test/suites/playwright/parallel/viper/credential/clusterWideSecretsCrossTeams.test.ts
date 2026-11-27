/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import {
    flowRootState,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    Utils,
    waitUtils
} from '../../../../../../modules/helpers';
import { assetsApi } from '../../../../../../modules/api';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { StatusCodes } from 'http-status-codes';
import {
    AccessKey,
    AccessKeyCreationRequest,
    AccessKeyCreationSpec,
    AssetCreationRequest,
    Password,
    PasswordCreationRequest,
    PasswordCreationSpec,
    Scope
} from '../../../../../../modules/api/apis/assetsApi';
import { V1Secret } from '@kubernetes/client-node/';
import { kubeCli } from '../../../../../../modules/cli';

test.describe(`Cluster-wide secrets, tags:  ${tags.resource.projects} ${tags.teams.viper}`, () => {
    const numberOfProjects = 3;
    const projects: Project[] = [];
    const runaiNamespace = 'runai';

    let passwordSecretName: string;
    let accessKeySecretName: string;

    async function waitForSecretCreationInCluster(
        name: string,
        namespace: string,
        desiredResult = true
    ): Promise<V1Secret> {
        let foundSecret;
        const doesSecretExist = async (): Promise<boolean> => {
            const res = await kubeCli.CoreV1Api.listNamespacedSecret({ namespace });
            const secrets = res.items;
            foundSecret = secrets.find((secret) => secret.metadata.name === name);
            return foundSecret !== undefined;
        };

        await waitUtils.waitForCondition(
            doesSecretExist,
            `expected secret ${namespace}/${name} to exist, desiredResult: ${desiredResult}`,
            desiredResult,
            60,
            3
        );
        return foundSecret;
    }

    test.beforeAll(async () => {
        assetsApi.credentialsApi.accessToken = flowRootState.users.system_administrator.accessToken;

        for (let index = 0; index < numberOfProjects; index++) {
            const project = await projectResourcesUtils.createProjectAndNamespace('p');
            projects.push(project);
        }
    });

    test.afterAll(async () => {
        const deleteProject = async (project: Project): Promise<void> => {
            try {
                await operations.deleteProjectAndNamespace(project.id);
                logger.debug(`Successfully deleted project and namespace: ${project.name}`);
            } catch (e: any) {
                logger.warn(`Failed to delete project ${project.name} (ID: ${project.id}): ${e.message}`);
            }
        };

        // Delete secrets created in the test
        logger.debug(`Deleting password secret ${passwordSecretName} from namespace ${runaiNamespace}`);
        try {
            await kubeCli.CoreV1Api.deleteNamespacedSecret({
                name: `password-${passwordSecretName}`,
                namespace: runaiNamespace
            });
        } catch (err) {
            logger.warn(`Failed to delete password secret: ${err}`);
        }

        logger.debug(`Deleting access key secret ${accessKeySecretName} from namespace ${runaiNamespace}`);
        try {
            await kubeCli.CoreV1Api.deleteNamespacedSecret({
                name: `accesskey-${accessKeySecretName}`,
                namespace: runaiNamespace
            });
        } catch (err) {
            logger.warn(`Failed to delete access key secret: ${err}`);
        }

        for (const project of projects) {
            await deleteProject(project);
        }
    });

    test('Create password, create access key and validate replication', async () => {
        let testPassword: Password;
        let testAccessKey: AccessKey;

        await test.step(`Create a password with clusterWideLabel, ${tags.teams.viper}`, async () => {
            const passwordToCreate = new PasswordCreationRequest();
            passwordToCreate.meta = new AssetCreationRequest();
            passwordToCreate.spec = new PasswordCreationSpec();
            passwordToCreate.meta.name = objectGenerator.generateShortNameByPrefix('pass');
            passwordToCreate.meta.scope = Scope.Tenant;
            passwordToCreate.spec.user = 'someuser';
            passwordToCreate.spec.password = 'somepass';
            passwordSecretName = passwordToCreate.meta.name;

            logger.debug(`Creating tenant password "${passwordSecretName}"`);
            const res = await assetsApi.credentialsApi.createPassword(passwordToCreate);
            testPassword = res.body;

            expect(res.response.statusCode).toBe(StatusCodes.ACCEPTED);
        });

        await test.step(`Validate password secret replication for projects, ${tags.teams.viper}`, async () => {
            for (const project of projects) {
                await waitForSecretCreationInCluster(`password-${testPassword.meta.name}`, project.namespace);
            }
        });

        await test.step(`Create an access key with clusterWideLabel, ${tags.teams.viper}`, async () => {
            const accessKeyToCreate = new AccessKeyCreationRequest();
            accessKeyToCreate.meta = new AssetCreationRequest();
            accessKeyToCreate.spec = new AccessKeyCreationSpec();
            accessKeyToCreate.meta.name = objectGenerator.generateShortNameByPrefix('ac');
            accessKeyToCreate.meta.scope = Scope.Tenant;
            accessKeyToCreate.spec.accessKeyId = 'someid';
            accessKeyToCreate.spec.secretAccessKey = 'somekey';
            accessKeySecretName = accessKeyToCreate.meta.name;

            logger.debug(`Creating tenant access-key "${accessKeySecretName}"`);
            const res = await assetsApi.credentialsApi.createAccessKey(accessKeyToCreate);
            testAccessKey = res.body;

            expect(res.response.statusCode).toBe(StatusCodes.ACCEPTED);
        });

        await test.step(`Validate access key secret replication for projects, ${tags.teams.viper}`, async () => {
            for (const project of projects) {
                await waitForSecretCreationInCluster(
                    `accesskey-${testAccessKey.meta.name}`,
                    Utils.namespaceFromProjectName(project.name)
                );
            }
        });
    });
});
