/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    credentialUtils,
    flowRootState,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    projectUtils,
    tags
} from '../../../../../../modules/helpers';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { CredentialPage } from '../../../../../../modules/ui/runaiV2/pages/CredentialPage';
import { AssetKind, Scope } from '../../../../../../modules/api/apis/assetsApi';
import { test } from '@playwright/test';

test.describe(`Credential UI, tags: ${tags.tool.ui} ${tags.resource.credential} ${tags.teams.viper}`, function () {
    let browserEngine: BrowserEngine;
    let auth: AuthPage;
    let credential: CredentialPage;

    let videoName: string;
    let projectName: string;
    let namespace: string;
    let project: Project;
    let credentialName: string;

    async function initBrowserAndLogin(): Promise<void> {
        await browserEngine.start();
        auth = new AuthPage(browserEngine.currentPage);
        credential = new CredentialPage(browserEngine.currentPage);
        await auth.signIn(
            flowRootState.users.system_administrator.email,
            flowRootState.users.system_administrator.password
        );
    }

    test.beforeAll(async function () {
        browserEngine = new BrowserEngine();
        project = await projectResourcesUtils.createProjectAndNamespace('p-e2e-for-cred');
        projectName = project.name;
        namespace = project.namespace;

        await projectUtils.waitForRoleBindingForUserInProject(
            namespace,
            flowRootState.users.system_administrator.email,
            projectUtils.newRbacRunaiSystemAdministratorRoleBindingName,
            true
        );
    });

    test.afterAll('deleting project and namespace', async function () {
        try {
            await operations.deleteProjectAndNamespace(project.id);
            logger.info(`deleted project ${projectName} and namespace ${namespace}`);
        } catch (error) {
            logger.error(`failed to delete project ${projectName} and namespace ${namespace}`);
            throw error;
        }
    });

    test.describe('should create credential', function () {
        let scopeEntityName: string;

        test.beforeEach(async ({}, testInfo) => {
            if (testInfo.title.includes(Scope.Department.toString())) {
                scopeEntityName = 'default';
            } else if (testInfo.title.includes(Scope.Project.toString())) {
                scopeEntityName = project.name;
            } else {
                scopeEntityName = 'runai';
            }
            await initBrowserAndLogin();
        });

        test.afterEach(async ({}, testInfo) => {
            await browserEngine.stopAndArchive(testInfo, 'credential-' + videoName);
        });

        [Scope.Department, Scope.Project].forEach(function (scope) {
            test(`access-key - ${scope} scope`, async () => {
                videoName = `create-ak-credential-${scope}-scope`;
                credentialName = objectGenerator.generateNameByPrefix('ak');

                await credential.createAccessKeyCredential(credentialName, 'ak-id', 'ak-secret', {
                    scope: scope,
                    scopeEntityName: scopeEntityName
                });

                await credential.deleteCredential(credentialName);
            });
        });

        [Scope.Department, Scope.Project].forEach(function (scope) {
            test(`docker-registry - ${scope} scope`, async () => {
                videoName = `create-dr-credential-${scope}-scope`;
                credentialName = objectGenerator.generateNameByPrefix('dr');

                // create
                await credential.createDockerRegistryCredential(
                    credentialName,
                    'dr-username',
                    'dr-password',
                    'https://my.url.com',
                    {
                        scope: scope,
                        scopeEntityName: scopeEntityName
                    }
                );

                await credential.deleteCredential(credentialName);
            });
        });

        [Scope.Project, Scope.Department].forEach(function (scope) {
            test(`username & password - ${scope} scope`, async () => {
                videoName = `create-userpass-credential-${scope}-scope`;
                credentialName = objectGenerator.generateNameByPrefix('up');

                await credential.createUserPassCredential(credentialName, 'up-username', 'up-password', {
                    scope: scope,
                    scopeEntityName: scopeEntityName
                });
                logger.debug('starting cleanup');
                await credential.deleteCredential(credentialName);
            });
        });
    });

    test.describe('should update credential description', function () {
        const credentialTypes = [
            AssetKind.DockerRegistry,
            AssetKind.Password,
            AssetKind.AccessKey,
            AssetKind.GenericSecret
        ];

        const credentialNameByType = {
            [AssetKind.DockerRegistry]: 'docker-registry',
            [AssetKind.Password]: 'password',
            [AssetKind.AccessKey]: 'access-key',
            [AssetKind.GenericSecret]: 'generic-secret'
        };

        const credentialCreationFunctions = {
            [AssetKind.DockerRegistry]: credentialUtils.createDockerRegistryAsset,
            [AssetKind.Password]: credentialUtils.createPasswordAsset,
            [AssetKind.AccessKey]: credentialUtils.createAccessKeyAsset,
            [AssetKind.GenericSecret]: credentialUtils.createGenericSecretAsset
        };

        const credentialDeleteFunctions = {
            [AssetKind.DockerRegistry]: credentialUtils.deleteDockerRegistryAsset,
            [AssetKind.Password]: credentialUtils.deletePasswordAsset,
            [AssetKind.AccessKey]: credentialUtils.deleteAccessKeyAsset,
            [AssetKind.GenericSecret]: credentialUtils.deleteGenericSecretAsset
        };

        const credentialCreationPayloads = {
            [AssetKind.DockerRegistry]: credentialUtils.getDockerRegistryAssetPayload,
            [AssetKind.Password]: credentialUtils.getPasswordAssetPayload,
            [AssetKind.AccessKey]: credentialUtils.getAccessKeyAssetPayload,
            [AssetKind.GenericSecret]: credentialUtils.getGenericSecretAssetPayload
        };

        test.beforeEach(async () => {
            await initBrowserAndLogin();
        });

        test.afterEach('stop browser', async ({}, testInfo) => {
            await browserEngine.stopAndArchive(testInfo, 'credential-' + videoName);
        });

        credentialTypes.forEach((credentialType: AssetKind) => {
            test(`should update ${credentialType} credential`, async () => {
                videoName = `update-${credentialType}`;
                const credentialName = objectGenerator.generateLettersOnlyByPrefix(
                    credentialNameByType[credentialType]
                );
                const creationPayload = credentialCreationPayloads[credentialType](credentialName);
                const createdCredential = await credentialCreationFunctions[credentialType](creationPayload);

                await credential.goToPage();
                const newDesc = objectGenerator.generateLettersOnlyByPrefix(
                    'new-description' + credentialName
                );
                await credential.updateCredentialDesc(credentialName, newDesc);
                await credentialDeleteFunctions[credentialType](createdCredential.meta.id);
            });
        });
    });
});
