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
    tags,
    unleashClient,
    assetsApiV2Utils,
    workloadsCrudUtils,
    environmentUtils,
    computeResourceUtils,
    settings,
    workloadsUtilsV2
} from '../../../../../../modules/helpers';
import { WorkloadType as CrudWorkloadType } from '../../../../../../modules/helpers/utils/WorkloadsUtilsV2';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { CredentialPage } from '../../../../../../modules/ui/runaiV2/pages/CredentialPage';
import {
    AssetKind,
    Scope,
    GenericSecretCreationRequest,
    GenericSecret,
    DockerRegistryCreationRequest,
    DockerRegistry,
    WorkloadType
} from '../../../../../../modules/api/apis/assetsApi';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test, expect } from '@playwright/test';
import { LocalUserSettings } from '../../../../../../modules/ui/runaiV2/pages/localUserSettings';
import { SecretType } from '../../../../../../modules/api/apis/k8sObjectsTrackerApi/generated/api';
import { EFeatureToggleKeys } from '../../../../../../modules/helpers/utils/Unleash/featureToggle';
import { StatusCodes } from 'http-status-codes';
import { WorkloadPage } from '../../../../../../modules/ui/runaiV2/pages/WorkloadPage';
import { HybridForm } from '../../../../../../modules/ui/runaiV2/pages/HybridSubmissionForm';
import { ECredentialsSource } from '../../../../../../modules/helpers/models/inferenceBaseFlatUI.models';

import { kubeCli } from '../../../../../../modules/cli';
import { V1Pod } from '@kubernetes/client-node';

test.describe(`Personal credentials, tags: ${tags.tool.ui} ${tags.resource.credential} ${tags.teams.viper}`, { tag: [tags.services.assets, tags.services.k8sObjectsTracker, tags.services.frontend, tags.runOn.parallelCI] }, function () {
    supportedFrom(SupportedVersions.V2_23)(function () {
        let browserEngine: BrowserEngine;
        let auth: AuthPage;
        let credential: CredentialPage;
        let userSettingsPage: LocalUserSettings;

        let videoName: string;
        let credentialName: string;

        async function initBrowserAndLogin(): Promise<void> {
            browserEngine = new BrowserEngine();

            await browserEngine.start();

            auth = new AuthPage(browserEngine.currentPage);
            userSettingsPage = new LocalUserSettings(browserEngine.currentPage);
            credential = new CredentialPage(browserEngine.currentPage);

            await auth.signIn(
                flowRootState.users.system_administrator.email,
                flowRootState.users.system_administrator.password
            );
        }

        test.describe('create and delete', function () {
            test.beforeEach(async () => {
                await initBrowserAndLogin();
            });

            test.afterEach(async ({}, testInfo) => {
                await browserEngine.stopAndArchive(testInfo, 'credential-' + videoName);
            });

            test('should create personal credential of type docker-registry and then delete it', async () => {
                videoName = 'create-personal-docker-registry-credential';
                credentialName = objectGenerator.generateNameByPrefix('dr');

                await userSettingsPage.openLocalSettingsPage();
                await userSettingsPage.toggleCredentialsSection();

                await userSettingsPage.openNewPersonalCredentialModal('docker-registry');
                await credential.fillCredentialCreationDescriptionInput(
                    'Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam'
                );

                await credential.fillDockerRegistryName(credentialName);
                await credential.fillDockerRegistryUsernameInput('dr-username');
                await credential.fillDockerRegistryPasswordInput('dr-password');
                await credential.fillDockerRegistryUrlInput('https://my.url.com');
                await userSettingsPage.clickOnSubmitCredentialFormBtn();

                await browserEngine.currentPage.waitForSelector(`[aid="delete-credential-${credentialName}"]`);

                await userSettingsPage.deletePersonalCredential(credentialName);
            });

            test('should create personal credential of type generic-secret and then delete it', async () => {
                videoName = 'create-personal-generic-secret-credential';
                credentialName = objectGenerator.generateNameByPrefix('gs');

                await userSettingsPage.openLocalSettingsPage();
                await userSettingsPage.toggleCredentialsSection();

                await userSettingsPage.openNewPersonalCredentialModal('generic-secret');
                await credential.fillCredentialCreationDescriptionInput(
                    'Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam'
                );
                await credential.fillGenericSecretNameInput(credentialName);
                await credential.fillGenericSecretKeyInput('test-key');
                await credential.fillGenericSecretValueInput('test-val');

                await userSettingsPage.clickOnSubmitCredentialFormBtn();

                await browserEngine.currentPage.waitForSelector(`[aid="delete-credential-${credentialName}"]`);

                await userSettingsPage.deletePersonalCredential(credentialName);
            });
        });

        test.describe('Submit a workload with personal credentials', function () {
            let project: Project;
            let workload: WorkloadPage;
            let hybridForm: HybridForm;
            let workloadName: string;

            const COMPUTE_RESOURCE = 'cpu-only';
            const IMAGE_NAME = 'alpine';
            const COMMAND = 'sleep 1000';
            const RUNTIME_SETTINGS = 'Runtime settings';

            const assetToCrudType: Record<string, CrudWorkloadType> = {
                [WorkloadType.Workspace]: CrudWorkloadType.WORKSPACE,
                [WorkloadType.Training]: CrudWorkloadType.TRAINING
            };

            test.beforeAll('creating project and namespace', async function () {
                project = await projectResourcesUtils.createProjectAndNamespace('h-p');
            });

            test.afterAll('deleting project and namespace', async function () {
                await operations.deleteProjectAndNamespace(project.id);
            });

            async function getWorkloadPod(): Promise<V1Pod> {
                const pods = await kubeCli.listPodsOfJob(project.namespace, workloadName);
                expect(pods.length).toBe(1);

                const pod = pods[0];
                expect(pod.metadata.name).toContain(workloadName);

                return pod;
            }

            test.describe('Generic Secret Credentials', function () {
                let genericSecret: GenericSecret;
                let currentCrudType: CrudWorkloadType;

                const ENV_VAR_NAME = 'PERSONAL_SECRET';
                const SECRET_KEY = 'SECRET_TOKEN';
                const SECRET_VALUE = 'test-secret-token-value';
                const ENV_VAR_SOURCE = 'My credentials';

                test.beforeEach(
                    'setup browser and create credential of type generic-secret',
                    async function () {
                        workloadsCrudUtils.accessToken =
                            flowRootState.users.system_administrator.accessToken;
                        assetsApiV2Utils.accessToken = flowRootState.users.system_administrator.accessToken;

                        await initBrowserAndLogin();

                        hybridForm = new HybridForm(browserEngine.currentPage);
                        workload = new WorkloadPage(browserEngine.currentPage);
                        workloadName = objectGenerator.generateNameByPrefix('hybrid-w');

                        credentialName = objectGenerator.generateNameByPrefix('test-personal-secret');

                        const genericSecretPayload: GenericSecretCreationRequest = {
                            meta: {
                                name: credentialName,
                                description: 'Test personal generic secret for API workload submission',
                                scope: Scope.Tenant,
                                _private: true
                            },
                            spec: {
                                keyValuePairs: [{ key: SECRET_KEY, value: SECRET_VALUE }]
                            }
                        };

                        genericSecret = await credentialUtils.createGenericSecretAsset(genericSecretPayload);
                        await credentialUtils.waitForGenericSecretAssetToBeReady(genericSecret);
                    }
                );

                test.afterEach('cleanup browser and workload', async ({}, testInfo) => {
                    await Promise.allSettled([
                        workloadsUtilsV2.deleteWorkloadByNameAndProjectId(
                            workloadName,
                            project.id,
                            currentCrudType
                        ).catch((e) => logger.info(`Workload cleanup skipped: ${e.message}`)),
                        credentialUtils.deleteGenericSecretAsset(genericSecret.meta.id)
                    ]);

                    await browserEngine.stopAndArchive(testInfo, 'credential-' + videoName);
                });

                async function assertPodHasCorrectPersonalSecretEnvVar(): Promise<void> {
                    await test.step('verify the personal secret environment variable in the pod and the value is correct', async () => {
                        const pod = await getWorkloadPod();

                        const personalSecretEnvVar = pod.spec.containers[0].env.find(
                            (envVar) => envVar.name === ENV_VAR_NAME
                        );

                        expect(personalSecretEnvVar).toBeDefined();
                        expect(personalSecretEnvVar.name).toBe(ENV_VAR_NAME);
                        expect(personalSecretEnvVar.valueFrom).toBeDefined();
                        expect(personalSecretEnvVar.valueFrom.secretKeyRef).toBeDefined();
                        expect(personalSecretEnvVar.valueFrom.secretKeyRef.key).toBe(SECRET_KEY);
                        expect(personalSecretEnvVar.valueFrom.secretKeyRef.name).toContain(credentialName);

                        const actualSecretName = personalSecretEnvVar.valueFrom.secretKeyRef.name;

                        const secret = await kubeCli.getSecret(actualSecretName, project.namespace);
                        expect(secret.data[SECRET_KEY]).toBeDefined();

                        const decodedValue = Buffer.from(secret.data[SECRET_KEY], 'base64').toString();
                        expect(decodedValue).toBe(SECRET_VALUE);
                    });
                }

                const workloadNameToApiType = {
                    [WorkloadType.Workspace]: WorkloadType.Workspace,
                    [WorkloadType.Training]: WorkloadType.Training
                };

                const workloadTestData = [WorkloadType.Workspace, WorkloadType.Training];

                workloadTestData.forEach((workloadTypeName) => {
                    test(`should submit a ${workloadTypeName} with a personal generic secret credential`, async () => {
                        currentCrudType = assetToCrudType[workloadTypeName];
                        videoName = `personal-generic-secret-${workloadTypeName}-submission`;

                        await workload.navigateToWorkloads();
                        await hybridForm.fillWorkloadCreationFirstPage(
                            flowRootState.cluster.name,
                            project.name,
                            workloadName,
                            workloadNameToApiType[workloadTypeName]
                        );

                        await hybridForm.fillImageInput(IMAGE_NAME);
                        await hybridForm.clickOnSubSectionByLabel(RUNTIME_SETTINGS);
                        await hybridForm.clickOnCommandAndArguments();
                        await hybridForm.fillCmdInput(COMMAND);
                        await hybridForm.clickOnAddEnvVarBtn();

                        await hybridForm.fillEnvVarName(ENV_VAR_NAME);
                        await hybridForm.selectEnvVarSource(ENV_VAR_SOURCE);
                        await hybridForm.selectEnvVarPersonalSecret(credentialName);
                        await hybridForm.fillEnvVarPersonalSecretKey(SECRET_KEY);

                        await hybridForm.openComputesDrawer();
                        await hybridForm.clickOnDrawerItemByLabel(COMPUTE_RESOURCE);

                        await hybridForm.clickOnSubmit();

                        await workload.waitForWorkloadUIStatus(workloadName);

                        await assertPodHasCorrectPersonalSecretEnvVar();
                    });
                });
            });

            test.describe('Docker Registry Credentials', function () {
                let dockerRegistry: DockerRegistry;
                let currentCrudType: CrudWorkloadType;

                const DR_USER = 'user@example.com';
                const DR_PASSWORD = 'dummy-registry-password';
                const DR_URL = 'https://my.url.com';

                test.beforeEach('setup browser and create docker registry credential', async function () {
                    workloadsCrudUtils.accessToken = flowRootState.users.system_administrator.accessToken;
                    assetsApiV2Utils.accessToken = flowRootState.users.system_administrator.accessToken;

                    await initBrowserAndLogin();

                    hybridForm = new HybridForm(browserEngine.currentPage);
                    workload = new WorkloadPage(browserEngine.currentPage);
                    workloadName = objectGenerator.generateNameByPrefix('hybrid-w');
                    credentialName = objectGenerator.generateNameByPrefix('test-dr-cred');

                    const dockerRegistryPayload: DockerRegistryCreationRequest = {
                        meta: {
                            name: credentialName,
                            description: 'Test personal docker registry for API workload submission',
                            scope: Scope.Tenant,
                            _private: true
                        },
                        spec: {
                            existingSecretName: null,
                            user: DR_USER,
                            password: DR_PASSWORD,
                            url: DR_URL
                        }
                    };

                    dockerRegistry = await credentialUtils.createDockerRegistryAsset(dockerRegistryPayload);
                    await credentialUtils.waitForDockerRegistryAssetToBeReady(dockerRegistry);
                });

                test.afterEach('cleanup browser and docker registry credential', async ({}, testInfo) => {
                    await Promise.allSettled([
                        workloadsUtilsV2.deleteWorkloadByNameAndProjectId(
                            workloadName,
                            project.id,
                            currentCrudType
                        ).catch((e) => logger.info(`Workload cleanup skipped: ${e.message}`)),
                        credentialUtils.deleteDockerRegistryAsset(dockerRegistry.meta.id)
                    ]);

                    await browserEngine.stopAndArchive(testInfo, 'credential-' + videoName);
                });

                async function verifyImagePullSecret(): Promise<void> {
                    await test.step('verify the docker registry image pull secret is correctly set in the pod', async () => {
                        const pod = await getWorkloadPod();

                        expect(pod.spec.imagePullSecrets).toBeDefined();
                        expect(pod.spec.imagePullSecrets.length).toBeGreaterThan(0);

                        const imagePullSecret = pod.spec.imagePullSecrets.find((secret) =>
                            secret.name.includes(credentialName.toLowerCase())
                        );

                        expect(imagePullSecret).toBeDefined();
                        expect(imagePullSecret.name).toContain(credentialName.toLowerCase());

                        const secret = await kubeCli.getSecret(imagePullSecret.name, project.namespace);
                        expect(secret.type).toBe('kubernetes.io/dockerconfigjson');
                        expect(secret.data['.dockerconfigjson']).toBeDefined();
                    });
                }

                const workloadNameToApiType = {
                    [WorkloadType.Workspace]: WorkloadType.Workspace,
                    [WorkloadType.Training]: WorkloadType.Training
                };

                const workloadTestData = [WorkloadType.Workspace, WorkloadType.Training];

                workloadTestData.forEach((workloadTypeName) => {
                    test(`should submit a ${workloadTypeName} with a personal docker registry credential`, async () => {
                        currentCrudType = assetToCrudType[workloadTypeName];
                        videoName = `personal-docker-registry-${workloadTypeName}-submission`;

                        await workload.navigateToWorkloads();
                        await hybridForm.fillWorkloadCreationFirstPage(
                            flowRootState.cluster.name,
                            project.name,
                            workloadName,
                            workloadNameToApiType[workloadTypeName]
                        );

                        await hybridForm.fillImageInput(IMAGE_NAME);
                        await hybridForm.clickOnSubSectionByLabel(RUNTIME_SETTINGS);
                        await hybridForm.clickOnCommandAndArguments();
                        await hybridForm.fillCmdInput(COMMAND);
                        await hybridForm.clickOnSubSectionByLabel('Image pull');
                        await hybridForm.selectImagePullSecretWithSource(
                            ECredentialsSource.Private,
                            credentialName
                        );

                        await hybridForm.openComputesDrawer();
                        await hybridForm.clickOnDrawerItemByLabel(COMPUTE_RESOURCE);

                        await hybridForm.clickOnSubmit();

                        await workload.waitForWorkloadUIStatus(workloadName);

                        await verifyImagePullSecret();
                    });
                });
            });
        });
    });
});
