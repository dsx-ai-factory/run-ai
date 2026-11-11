/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { assetsApi } from '../../../../../../modules/api';
import {
    credentialUtils,
    flowRootState,
    logger,
    ngcCredentialVerificationUtils,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    Utils,
    waitUtils,
    workloadsUtilsV2
} from '../../../../../../modules/helpers';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { CredentialPage } from '../../../../../../modules/ui/runaiV2/pages/CredentialPage';
import { Scope } from '../../../../../../modules/api/apis/assetsApi';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test, expect } from '@playwright/test';
import { WorkloadPage } from '../../../../../../modules/ui/runaiV2/pages/WorkloadPage';
import { InferencePage } from '../../../../../../modules/ui/runaiV2/pages/InferencePage';
import { InferenceFlatUIUtil } from '../../../../../../modules/ui/utils/inference/inferenceFlatUI.util';
import { HybridForm } from '../../../../../../modules/ui/runaiV2/pages/HybridSubmissionForm';
import { WorkloadType } from '../../../../../../modules/helpers/utils/WorkloadsUtilsV2';
import {
    ECredentialsSource,
    ECredentialsSourceType,
    IInferenceType
} from '../../../../../../modules/helpers/models/inferenceBaseFlatUI.models';
import { LocalUserSettings } from '../../../../../../modules/ui/runaiV2/pages/localUserSettings';
import { LoginUi } from '../../../../../../modules/ui/runaiV2/pages/LoginUiUtils';

test.describe(`NGC API Key Credentials UI Test, tags: ${tags.resource.inferenceWorkload} ${tags.resource.credential} ${tags.teams.loong}`, function () {
    supportedFrom(SupportedVersions.V2_24)(function () {
        let browserEngine: BrowserEngine;
        let authPage: AuthPage;
        let credentialsPage: CredentialPage;
        let workloadPage: WorkloadPage;
        let inferencePage: InferencePage;
        let inferenceFlatUIUtil: InferenceFlatUIUtil;
        let userSettingsPage: LocalUserSettings;
        let loginUi: LoginUi;

        let videoName: string;
        let projectName1: string;
        let projectName2: string;
        let namespace1: string;
        let namespace2: string;
        let project1: Project;
        let project2: Project;

        let ngcApiKeyName: string;
        let ngcApiKey: string;
        let ngcApiKeyDescription: string;
        let inferenceName: string;
        let secondInferenceName: string;
        let workloadId: string;
        let createdWorkloadIds: Array<{ id: string; type: WorkloadType }> = [];
        let createdCredentialName: string | null = null;
        let createdUserCredentialName: string | null = null;
        let authToken: string;

        const nimModelName = 'meta/llama-3.1-8b-instruct-dgx-spark';

        test.beforeAll(async function () {
            browserEngine = new BrowserEngine();

            projectName1 = objectGenerator.generateShortNameByPrefix('proj-ngc', 5);
            namespace1 = Utils.namespaceFromProjectName(projectName1);
            project1 = await projectResourcesUtils.createProject(namespace1, {
                name: projectName1,
                deservedGpus: 1,
                maxAllowedGpus: 1
            });
            authToken = flowRootState.users.system_administrator.accessToken;

            ngcApiKey = process.env.NGC_API_KEY;

            // skip the tests until the ENV variable 'NGC_API_KEY' is added to the pipeline
            test.skip();
        });

        test.afterAll('deleting project and namespace', async () => {
            if (project1 && project1.id) {
                try {
                    await operations.deleteProjectAndNamespace(project1.id);
                    logger.info(`deleted project ${projectName1} and namespace ${namespace1}`);
                } catch {
                    logger.error('failed to delete project ' + projectName1 + ' and namespace ' + namespace1);
                }
            }
        });

        test.beforeEach(async function () {
            await browserEngine.start();
            authPage = new AuthPage(browserEngine.currentPage);
            credentialsPage = new CredentialPage(browserEngine.currentPage);
            workloadPage = new WorkloadPage(browserEngine.currentPage);
            inferencePage = new InferencePage(browserEngine.currentPage);
            inferenceFlatUIUtil = new InferenceFlatUIUtil(browserEngine.currentPage);
            userSettingsPage = new LocalUserSettings(browserEngine.currentPage);
            loginUi = new LoginUi(browserEngine.currentPage);

            await authPage.signIn(
                flowRootState.users.system_administrator.email,
                flowRootState.users.system_administrator.password
            );

            ngcApiKeyName = objectGenerator.generateShortNameByPrefix('ngc-api-key', 5);
            ngcApiKeyDescription = objectGenerator.generateShortNameByPrefix('ngc-desc', 5);
            inferenceName = objectGenerator.generateShortNameByPrefix('inference-nim');
            secondInferenceName = objectGenerator.generateShortNameByPrefix('second-inference-nim');
            createdCredentialName = null;
        });

        test.afterEach('Delete workloads and stop browser', async function () {
            // Delete personal NGC API key credentials via UI (API cannot list/delete personal credentials)
            if (createdUserCredentialName) {
                try {
                    await userSettingsPage.openLocalSettingsPage();
                    await userSettingsPage.deletePersonalCredential(createdUserCredentialName);
                    logger.info(
                        `Successfully deleted personal NGC API key credential via UI: ${createdUserCredentialName}`
                    );
                } catch (error) {
                    logger.warn(
                        `Failed to delete personal NGC API key credential ${createdUserCredentialName}:`,
                        error
                    );
                }
                createdUserCredentialName = null;
            }
            await browserEngine.stopAndArchive(this, videoName);

            if (createdWorkloadIds.length > 0) {
                try {
                    for (const workload of createdWorkloadIds) {
                        try {
                            await workloadsUtilsV2.deleteWorkload(
                                workload.id,
                                workload.type,
                                flowRootState.users.system_administrator.accessToken
                            );
                            logger.info(
                                `Successfully deleted workload with ID: ${workload.id}, type: ${workload.type}`
                            );
                        } catch (error) {
                            logger.warn(`Failed to delete workload with ID ${workload.id}:`, error);
                        }
                    }
                } catch (error) {
                    logger.warn('Failed to delete workloads:', error);
                }
                createdWorkloadIds = [];
            }

            // Delete project/department scope NGC API key credentials via API
            if (createdCredentialName) {
                try {
                    const credentials = await credentialUtils.list();
                    const credential = credentials.find((cred) => cred.meta.name === createdCredentialName);

                    if (credential) {
                        await credentialUtils.deleteNgcApiKeyAsset(
                            credential.meta.id,
                            flowRootState.users.system_administrator.accessToken
                        );
                        logger.info(
                            `Successfully deleted project/department scope NGC API key credential via API: ${createdCredentialName}`
                        );
                    } else {
                        logger.warn(
                            `Project/department scope credential ${createdCredentialName} not found, may already be deleted`
                        );
                    }
                } catch (error) {
                    logger.warn(
                        `Failed to delete project/department scope NGC API key credential ${createdCredentialName}:`,
                        error
                    );
                }
                createdCredentialName = null;
            }

            if (project2 && project2.id) {
                try {
                    await operations.deleteProjectAndNamespace(project2.id);
                    logger.info(`Successfully deleted project2 ${projectName2} and namespace ${namespace2}`);
                    project2 = null;
                    projectName2 = null;
                    namespace2 = null;
                } catch (error) {
                    logger.warn(
                        `Failed to delete project2 ${projectName2} and namespace ${namespace2}:`,
                        error
                    );
                }
            }
        });

        test('T5530741: Create, Edit, Delete NGC API key', async () => {
            videoName = 'ngc-api-key-crud-project-scope';
            const editedDesc = `edited-${ngcApiKeyDescription}`;

            await test.step('Create NGC API Key Credential with Project Scope', async () => {
                await credentialsPage.goToPage();
                await credentialsPage.createNgcApiKeyCredential(
                    ngcApiKeyName,
                    ngcApiKey,
                    {
                        scope: Scope.Project,
                        scopeEntityName: project1.name
                    },
                    ngcApiKeyDescription
                );
                createdCredentialName = ngcApiKeyName;
                logger.info(`NGC API key credential created: ${ngcApiKeyName}`);
            });

            await test.step('Verify the key is successfully created and all columns are correctly displayed', async () => {
                await credentialsPage.setColumn('Description', true);
                await credentialsPage.verifyCredentialDetails({
                    credentialName: ngcApiKeyName,
                    type: 'NGC API key',
                    status: 'No issues found',
                    createdBy: flowRootState.users.system_administrator.email,
                    description: ngcApiKeyDescription
                });
            });

            await test.step('Edit the description of the key and submit', async () => {
                await credentialsPage.updateCredentialDesc(ngcApiKeyName, editedDesc);
                logger.info(`Updated NGC API key description to: ${editedDesc}`);
            });

            await test.step('Verify the key is successfully updated and all columns are correctly displayed', async () => {
                await credentialsPage.verifyCredentialDetails({
                    credentialName: ngcApiKeyName,
                    type: 'NGC API key',
                    status: 'No issues found',
                    description: editedDesc,
                    createdBy: flowRootState.users.system_administrator.email
                });
            });

            await test.step('Delete the key', async () => {
                await credentialsPage.goToPage();
                await credentialsPage.deleteCredential(ngcApiKeyName);
                createdCredentialName = null;
                logger.info(`Deleted NGC API key credential: ${ngcApiKeyName}`);
            });
        });

        test('T5530742: Create nim workload with NGC API key', async () => {
            test.setTimeout(300 * 1000);
            videoName = 'nim-workload-with-ngc-api-key';
            let secretInfo: { secretName: string; secretKey: string };

            await test.step('Create NGC API Key Credential with Project Scope', async () => {
                await credentialsPage.goToPage();
                await credentialsPage.createNgcApiKeyCredential(
                    ngcApiKeyName,
                    ngcApiKey,
                    {
                        scope: Scope.Project,
                        scopeEntityName: project1.name
                    },
                    ngcApiKeyDescription
                );
                createdCredentialName = ngcApiKeyName;
                logger.info(`NGC API key credential created: ${ngcApiKeyName}`);
            });

            await test.step('Verify credential is ready before using it', async () => {
                await credentialsPage.verifyCredentialDetails({
                    credentialName: ngcApiKeyName,
                    type: 'NGC API key',
                    status: 'No issues found',
                    createdBy: flowRootState.users.system_administrator.email
                });
                logger.info(`Credential ${ngcApiKeyName} is ready to use`);
            });

            await test.step('Create NIM inference workload using UI', async () => {
                await workloadPage.goToPage();
                await inferenceFlatUIUtil.createNimFlatWorkload({
                    clusterName: flowRootState.cluster.name,
                    projectName: project1.name,
                    workloadName: inferenceName,
                    modelSection: {
                        modelName: nimModelName,
                        credentials: {
                            source: ECredentialsSource.Shared,
                            sourceType: ECredentialsSourceType.NgcApiKey,
                            secretName: ngcApiKeyName
                        }
                    }
                });
                logger.info(`NIM inference workload created via UI: ${inferenceName}`);

                // Get and store workload ID
                workloadId = await workloadsUtilsV2.getWorkloadIdByName(inferenceName, project1.id);
                createdWorkloadIds.push({ id: workloadId, type: WorkloadType.INFERENCE });
                logger.info(`Workload ID ${workloadId} added to cleanup list`);
            });

            await test.step('Verify the workload is created successfully', async () => {
                await workloadsUtilsV2.waitForBackendWorkloadByNameAndProjectName(
                    inferenceName,
                    project1.name,
                    60
                );
                logger.info(`Workload ${inferenceName} is successfully created`);
            });

            await test.step('Get pod name and verify image pull succeeded', async () => {
                await ngcCredentialVerificationUtils.waitForPodAndVerifyImagePullMessage(
                    namespace1,
                    inferenceName,
                    inferenceName
                );
            });

            await test.step('Verify NGC_API_KEY environment variable in pod', async () => {
                secretInfo = await ngcCredentialVerificationUtils.verifyNgcApiKeyEnvInPod(
                    namespace1,
                    inferenceName
                );
            });

            await test.step('Verify secret value matches NGC API key', async () => {
                await ngcCredentialVerificationUtils.verifyNgcApiKeySecret(
                    namespace1,
                    secretInfo.secretName,
                    secretInfo.secretKey,
                    ngcApiKey
                );
            });
        });

        test('T5530743: Create NIM workloads with Cluster scope NGC API key in multiple projects', async () => {
            test.setTimeout(300 * 1000);
            videoName = 'nim-workloads-with-cluster-ngc-api-key';

            await test.step('Create project2 for this test', async () => {
                projectName2 = objectGenerator.generateShortNameByPrefix('proj-ngc', 5);
                namespace2 = Utils.namespaceFromProjectName(projectName2);
                project2 = await projectResourcesUtils.createProject(namespace2, {
                    name: projectName2,
                    deservedGpus: 1,
                    maxAllowedGpus: 1
                });
                logger.info(`Created project2: ${projectName2} with namespace: ${namespace2}`);
            });

            await test.step('Create NGC API Key Credential with Cluster Scope', async () => {
                await credentialsPage.goToPage();
                await credentialsPage.createNgcApiKeyCredential(
                    ngcApiKeyName,
                    ngcApiKey,
                    {
                        scope: Scope.Cluster,
                        scopeEntityName: flowRootState.cluster.name
                    },
                    ngcApiKeyDescription
                );
                createdCredentialName = ngcApiKeyName;
                logger.info(`Created NGC API key with cluster scope: ${ngcApiKeyName}`);
            });

            await test.step('Verify credential is ready before using it', async () => {
                await credentialsPage.verifyCredentialDetails({
                    credentialName: ngcApiKeyName,
                    type: 'NGC API key',
                    status: 'No issues found',
                    createdBy: flowRootState.users.system_administrator.email
                });
                logger.info(`Credential ${ngcApiKeyName} is ready to use`);
            });

            await test.step('Create NIM inference workload in project1', async () => {
                await workloadPage.goToPage();
                await inferenceFlatUIUtil.createNimFlatWorkload({
                    clusterName: flowRootState.cluster.name,
                    projectName: project1.name,
                    workloadName: inferenceName,
                    modelSection: {
                        modelName: nimModelName,
                        credentials: {
                            source: ECredentialsSource.Shared,
                            sourceType: ECredentialsSourceType.NgcApiKey,
                            secretName: ngcApiKeyName
                        }
                    }
                });
                logger.info(`NIM inference workload created in project1: ${inferenceName}`);

                const workloadId1 = await workloadsUtilsV2.getWorkloadIdByName(inferenceName, project1.id);
                createdWorkloadIds.push({ id: workloadId1, type: WorkloadType.INFERENCE });
                logger.info(`Workload ID ${workloadId1} added to cleanup list`);
            });

            await test.step('Verify the workload is created successfully', async () => {
                await workloadsUtilsV2.waitForBackendWorkloadByNameAndProjectName(
                    inferenceName,
                    project1.name,
                    60
                );
                logger.info(`Workload ${inferenceName} is successfully created`);
            });

            await test.step('Get pod name and verify image pull succeeded', async () => {
                await ngcCredentialVerificationUtils.waitForPodAndVerifyImagePullMessage(
                    namespace1,
                    inferenceName,
                    inferenceName
                );
            });

            await test.step('Verify NGC_API_KEY in project1 pod', async () => {
                const secretInfo = await ngcCredentialVerificationUtils.verifyNgcApiKeyEnvInPod(
                    namespace1,
                    inferenceName,
                    'Project1'
                );
                await ngcCredentialVerificationUtils.verifyNgcApiKeySecret(
                    namespace1,
                    secretInfo.secretName,
                    secretInfo.secretKey,
                    ngcApiKey,
                    'Project1'
                );
            });

            await test.step('Delete workload in project1', async () => {
                await workloadsUtilsV2.deleteWorkloadByNameAndProjectId(
                    inferenceName,
                    project1.id,
                    WorkloadType.INFERENCE,
                    flowRootState.users.system_administrator.accessToken
                );
                logger.info(`Workload ${inferenceName} deleted from project1`);
            });

            await test.step('Create NIM inference workload in project2', async () => {
                await browserEngine.currentPage.reload();
                await workloadPage.goToPage();
                await inferenceFlatUIUtil.createNimFlatWorkload({
                    clusterName: flowRootState.cluster.name,
                    projectName: project2.name,
                    workloadName: secondInferenceName,
                    modelSection: {
                        modelName: nimModelName,
                        credentials: {
                            source: ECredentialsSource.Shared,
                            sourceType: ECredentialsSourceType.NgcApiKey,
                            secretName: ngcApiKeyName
                        }
                    }
                });
                logger.info(`NIM inference workload created in project2: ${secondInferenceName}`);

                const workloadId2 = await workloadsUtilsV2.getWorkloadIdByName(
                    secondInferenceName,
                    project2.id
                );
                createdWorkloadIds.push({ id: workloadId2, type: WorkloadType.INFERENCE });
                logger.info(`Workload ID ${workloadId2} added to cleanup list`);
            });

            await test.step('Verify the workload is created successfully', async () => {
                await workloadsUtilsV2.waitForBackendWorkloadByNameAndProjectName(
                    secondInferenceName,
                    project2.name,
                    60
                );
                logger.info(`Workload ${secondInferenceName} is successfully created`);
            });

            await test.step('Verify image pull in project2', async () => {
                await ngcCredentialVerificationUtils.waitForPodAndVerifyImagePullMessage(
                    namespace2,
                    secondInferenceName,
                    secondInferenceName
                );
                logger.info('Project2 - Image pull verified');
            });

            await test.step('Verify NGC_API_KEY is correctly set in project2 pod', async () => {
                const secretInfo = await ngcCredentialVerificationUtils.verifyNgcApiKeyEnvInPod(
                    namespace2,
                    secondInferenceName,
                    'Project2'
                );
                await ngcCredentialVerificationUtils.verifyNgcApiKeySecret(
                    namespace2,
                    secretInfo.secretName,
                    secretInfo.secretKey,
                    ngcApiKey,
                    'Project2'
                );
            });
        });

        test('T5530744: NGC API key of project scope cannot be used by other project', async () => {
            videoName = 'ngc-api-key-project-scope-isolation';

            await test.step('Create project2 for this test', async () => {
                projectName2 = objectGenerator.generateShortNameByPrefix('proj-ngc', 5);
                namespace2 = Utils.namespaceFromProjectName(projectName2);
                project2 = await projectResourcesUtils.createProject(namespace2, {
                    name: projectName2,
                    deservedGpus: 1,
                    maxAllowedGpus: 1
                });
                logger.info(`Created project2: ${projectName2} with namespace: ${namespace2}`);
            });

            await test.step('Create NGC API Key Credential with Project1 Scope', async () => {
                await credentialsPage.goToPage();
                await credentialsPage.createNgcApiKeyCredential(
                    ngcApiKeyName,
                    ngcApiKey,
                    {
                        scope: Scope.Project,
                        scopeEntityName: project1.name
                    },
                    ngcApiKeyDescription
                );
                createdCredentialName = ngcApiKeyName;
                logger.info(`NGC API key credential created under project1: ${ngcApiKeyName}`);
            });

            await test.step('Navigate to create NIM inference workload in project2 and verify NGC credential', async () => {
                await browserEngine.currentPage.goBack();
                await browserEngine.currentPage.reload();
                await workloadPage.goToPage();
                await workloadPage.openInferenceForm();
                await inferencePage.fillTheFirstStep(
                    flowRootState.cluster.name,
                    project2.name,
                    inferenceName,
                    IInferenceType.Nim
                );

                await inferenceFlatUIUtil.fillNimModelData({
                    modelName: nimModelName
                });

                const hybridForm = new HybridForm(browserEngine.currentPage);
                await hybridForm.selectNimAccessFromCredentials();
                await hybridForm.selectFromDropDown(
                    hybridForm.hybridNimModelSection.credentials.sourceSelect,
                    'Shared secret'
                );
                await hybridForm.selectFromDropDown(
                    hybridForm.hybridNimModelSection.credentials.sourceTypeSelect,
                    'NGC API key'
                );
                await browserEngine.currentPage
                    .locator(hybridForm.hybridNimModelSection.credentials.secretNameSelect)
                    .click();
                const credentialOptions = await browserEngine.currentPage
                    .locator('[role="option"]')
                    .allTextContents();
                const isKeyAvailable = credentialOptions.some((option) =>
                    option.trim().includes(ngcApiKeyName)
                );
                expect(isKeyAvailable).toBe(false);
                logger.info(
                    `Verified: NGC API key '${ngcApiKeyName}' from project1 is NOT available in project2`
                );
            });

            await test.step('Close workload creation form', async () => {
                await workloadPage.closeWorkloadForm();
            });
        });

        test('T5530747: Create, Edit, Delete personal NGC API key', async () => {
            videoName = 'personal-ngc-api-key-crud';

            await test.step('Create personal NGC API key', async () => {
                await userSettingsPage.openLocalSettingsPage();
                await userSettingsPage.createPersonalNgcApiKeyCredential(
                    ngcApiKeyName,
                    ngcApiKey,
                    ngcApiKeyDescription
                );
                createdUserCredentialName = ngcApiKeyName;
                logger.info(`Personal NGC API key credential created: ${ngcApiKeyName}`);
            });

            await test.step('Verify the key is successfully created and all fields are correctly displayed', async () => {
                const verificationResult = await userSettingsPage.verifyPersonalNgcApiKeyCredential(
                    ngcApiKeyName,
                    ngcApiKeyDescription
                );
                expect(verificationResult.nameMatch).toBe(true);
                expect(verificationResult.typeMatch).toBe(true);
                expect(verificationResult.descriptionMatch).toBe(true);
                logger.info(
                    `Verified: Personal NGC API key '${ngcApiKeyName}' is correctly displayed with all fields`
                );
            });

            await test.step('Delete the personal NGC API key', async () => {
                await userSettingsPage.deletePersonalCredential(ngcApiKeyName);
                createdUserCredentialName = null;
                logger.info(`Deleted personal NGC API key credential: ${ngcApiKeyName}`);
            });
        });

        test('T5530748: User2 cannot access personal api key of User1', async () => {
            videoName = 'personal-ngc-api-key-isolation';
            const user1PersonalKeyName = `${ngcApiKeyName}-user1`;
            const user1 = flowRootState.users.system_administrator;
            const user2 = flowRootState.users.ml_engineer;

            await test.step('User1 (System Admin) creates a personal NGC API key', async () => {
                await userSettingsPage.openLocalSettingsPage();
                await userSettingsPage.createPersonalNgcApiKeyCredential(
                    user1PersonalKeyName,
                    ngcApiKey,
                    'User1 personal NGC API key'
                );
                createdUserCredentialName = user1PersonalKeyName;
                logger.info(
                    `User1 (${user1.email}) created personal NGC API key: ${user1PersonalKeyName}`
                );
            });

            await test.step('Verify personal NGC API key of User1 is created', async () => {
                const verificationResult = await userSettingsPage.verifyPersonalNgcApiKeyCredential(
                    user1PersonalKeyName,
                    'User1 personal NGC API key'
                );
                expect(verificationResult.nameMatch).toBe(true);
                expect(verificationResult.typeMatch).toBe(true);
                expect(verificationResult.descriptionMatch).toBe(true);
                logger.info(
                    `Verified: personal NGC API key '${user1PersonalKeyName}' of User1 is correctly displayed`
                );
            });

            await test.step('User1 logs out and switch to User2', async () => {
                await loginUi.logoutToLoginPage();
                await authPage.signIn(user2.email, user2.password);
                logger.info(`User1 logged out successfully and User2 (${user2.email}) logged in`);
            });

            await test.step('User2 tries to create inference workload and cannot see personal credential of user1', async () => {
                await workloadPage.goToPage();
                await workloadPage.openInferenceForm();
                await inferencePage.fillTheFirstStep(
                    flowRootState.cluster.name,
                    projectName1,
                    inferenceName,
                    IInferenceType.Nim
                );

                await inferenceFlatUIUtil.fillNimModelData({
                    modelName: nimModelName
                });

                const hybridForm = new HybridForm(browserEngine.currentPage);
                await hybridForm.selectNimAccessFromCredentials();
                await hybridForm.selectFromDropDown(
                    hybridForm.hybridNimModelSection.credentials.sourceSelect,
                    'Shared secret'
                );
                await hybridForm.selectFromDropDown(
                    hybridForm.hybridNimModelSection.credentials.sourceTypeSelect,
                    'NGC API key'
                );

                await browserEngine.currentPage
                    .locator(hybridForm.hybridNimModelSection.credentials.secretNameSelect)
                    .click();
                const allCredentialOptions = await browserEngine.currentPage
                    .locator('[role="option"]')
                    .allTextContents();
                const isUser1KeyAvailable = allCredentialOptions.some((option) =>
                    option.trim().includes(user1PersonalKeyName)
                );

                expect(isUser1KeyAvailable).toBe(false);
                logger.info(
                    `Verified: User2 cannot see User1's personal NGC API key '${user1PersonalKeyName}' in the credentials list`
                );
            });

            await test.step('Close workload creation form and switch to User1', async () => {
                await workloadPage.closeWorkloadForm();
                await loginUi.logoutToLoginPage();
                await authPage.signIn(user1.email, user1.password);
            });
        });

        test('T5530749: Create nim workload with personal NGC API key', async () => {
            test.setTimeout(300 * 1000);
            videoName = 'nim-workload-with-personal-ngc-api-key';
            const personalKeyName = `${ngcApiKeyName}-personal`;
            const personalInferenceName = `${inferenceName}-personal`;

            await test.step('Login and create personal NGC API key', async () => {
                await userSettingsPage.openLocalSettingsPage();
                await userSettingsPage.createPersonalNgcApiKeyCredential(
                    personalKeyName,
                    ngcApiKey,
                    'Personal NGC API key for workload creation'
                );
                createdUserCredentialName = personalKeyName;
                logger.info(`Personal NGC API key credential created: ${personalKeyName}`);
            });

            await test.step('Wait for personal NGC API key credential to be ready', async () => {
                const authHeaders = { headers: { Authorization: `Bearer ${authToken}` } };

                // Wait for credential to appear in the list and be ready
                await waitUtils.waitForCondition(
                    async () => {
                        try {
                            const response = await assetsApi.credentialsApi.listNgcApiKey(
                                personalKeyName,
                                undefined,
                                undefined,
                                undefined,
                                undefined,
                                false,
                                true,
                                true,
                                false,
                                authHeaders
                            );

                            const ngcApiKeys = response.body.entries;
                            if (!ngcApiKeys || ngcApiKeys.length === 0) {
                                logger.debug(`Personal NGC API key ${personalKeyName} not found in list yet`);
                                return false;
                            }

                            const personalNgcApiKey = ngcApiKeys[0];
                            const status = personalNgcApiKey.status?.status?.toString();
                            logger.debug(`Personal NGC API key ${personalKeyName} status: ${status}`);

                            return status === 'No issues found';
                        } catch (error) {
                            logger.debug(
                                `Error checking personal NGC API key ${personalKeyName} status: ${error.message}`
                            );
                            return false;
                        }
                    },
                    `Waiting for personal NGC API key ${personalKeyName} to be ready`,
                    true
                );
                logger.info(`Personal NGC API key '${personalKeyName}' is ready to use`);
            });

            await test.step('Close user settings and navigate to workloads page', async () => {
                await browserEngine.currentPage.goBack();
                await browserEngine.currentPage.waitForLoadState('networkidle');
                logger.info('Navigated back to close user settings');
            });

            await test.step('Create NIM inference workload with personal NGC API key', async () => {
                await workloadPage.goToPage();
                await inferenceFlatUIUtil.createNimFlatWorkload({
                    clusterName: flowRootState.cluster.name,
                    projectName: project1.name,
                    workloadName: personalInferenceName,
                    modelSection: {
                        modelName: nimModelName,
                        credentials: {
                            source: ECredentialsSource.Private,
                            sourceType: ECredentialsSourceType.NgcApiKey,
                            secretName: personalKeyName
                        }
                    }
                });
                logger.info(
                    `NIM inference workload created with personal NGC API key: ${personalInferenceName}`
                );

                workloadId = await workloadsUtilsV2.getWorkloadIdByName(personalInferenceName, project1.id);
                createdWorkloadIds.push({ id: workloadId, type: WorkloadType.INFERENCE });
                logger.info(`Workload ID ${workloadId} added to cleanup list`);
            });

            await test.step('Verify the workload is created successfully', async () => {
                await workloadsUtilsV2.waitForBackendWorkloadByNameAndProjectName(
                    personalInferenceName,
                    project1.name,
                    60
                );
                logger.info(`Workload ${personalInferenceName} is successfully created`);
            });

            await test.step('Check the pod event shows the image is pulled successfully', async () => {
                await ngcCredentialVerificationUtils.waitForPodAndVerifyImagePullMessage(
                    namespace1,
                    personalInferenceName,
                    personalInferenceName
                );
            });

            await test.step('Check environment value NGC_API_KEY is correctly configured', async () => {
                const secretInfo = await ngcCredentialVerificationUtils.verifyNgcApiKeyEnvInPod(
                    namespace1,
                    personalInferenceName
                );
                await ngcCredentialVerificationUtils.verifyNgcApiKeySecret(
                    namespace1,
                    secretInfo.secretName,
                    secretInfo.secretKey,
                    ngcApiKey
                );
            });
        });
    });
});
