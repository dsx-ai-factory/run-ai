/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import {
    objectGenerator,
    Utils,
    waitUtils,
    logger,
    projectResourcesUtils,
    flowRootState,
    operations,
    apiUtils,
    unleashClient,
    credentialUtils
} from '../../../../../../modules/helpers';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { WorkloadPage, EWorkloadStatus } from '../../../../../../modules/ui/runaiV2/pages/WorkloadPage';
import { TrainingPage } from '../../../../../../modules/ui/runaiV2/pages/TrainingPage';
import { WorkspacePage } from '../../../../../../modules/ui/runaiV2/pages/WorkspacePage';
import { ProjectPage } from '../../../../../../modules/ui/runaiV2/pages/ProjectPage';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { CredentialPage } from '../../../../../../modules/ui/runaiV2/pages/CredentialPage';
import { UserRole } from '../../../../../../modules/helpers/config/roleMapping';
import { tags } from '../../../../../../modules/helpers/config/tags';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { WorkloadType, Scope } from '../../../../../../modules/api/apis/assetsApi';
import { runaiCliV2 } from '../../../../../../modules/cli';
import { cliV2WorkloadType, execFlags } from '../../../../../../modules/cli/tools/RunaiCliV2.constants';
import { EFeatureToggleKeys } from '../../../../../../modules/helpers/utils/Unleash/featureToggle';
import { kubeCli } from '../../../../../../modules/cli';

test.describe(`Create single node workloads (Training & Workspace), tags: ${tags.tool.ui} ${tags.resource.workloads} ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_15)(function () {
        let browserEngine: BrowserEngine;
        let auth: AuthPage;
        let workloads: WorkloadPage;
        let training: TrainingPage;
        let workspace: WorkspacePage;
        let projectPage: ProjectPage;
        let trainingName: string;
        let workspaceName: string;
        let projectName: string;
        let namespace: string;
        let project;
        let initialGpuCount: number;
        let initialCpuCount: number;
        const environmentName = 'tensorboard-tensorflow';

        test.beforeAll(async () => {
            await unleashClient.disableFeature(EFeatureToggleKeys.SHOW_HYBRID_SUBMISSION);
            await apiUtils.disableHybridSubmission();
        });

        test.afterAll(async () => {
            await unleashClient.enableFeature(EFeatureToggleKeys.SHOW_HYBRID_SUBMISSION);
            await apiUtils.enableHybridSubmission();
            await apiUtils.setFlexTemplateFeature(true);
        });

        test.beforeEach(async () => {
            browserEngine = new BrowserEngine();
            await browserEngine.start();
            auth = new AuthPage(browserEngine.currentPage);
            workloads = new WorkloadPage(browserEngine.currentPage);
            training = new TrainingPage(browserEngine.currentPage);
            workspace = new WorkspacePage(browserEngine.currentPage);
            projectPage = new ProjectPage(browserEngine.currentPage);
            trainingName = objectGenerator.generateNameByPrefix('e2e-training-job');
            workspaceName = objectGenerator.generateNameByPrefix('e2e-workspace');
            projectName = objectGenerator.generateNameByPrefix('e2e-project');
            namespace = Utils.namespaceFromProjectName(projectName);
            project = await projectResourcesUtils.createProject(namespace, {
                name: projectName,
                deservedGpus: 1,
                maxAllowedGpus: 1
            });
        });

        test.afterEach(async ({}, testInfo) => {
            await browserEngine.stopAndArchive(testInfo, 'single-node-workload');
            if (project && project.id) {
                try {
                    await operations.deleteProjectAndNamespace(project.id);
                    logger.info(`deleted project ${projectName} and namespace ${namespace}`);
                } catch {
                    logger.error('failed to delete project ' + projectName + ' and namespace ' + namespace);
                }
            }
        });

        test('T3738301: L1 researcher can create single node training workload', async () => {
            test.setTimeout(300 * 1000); // this test may take longer than the default timeout
            await auth.signIn(
                flowRootState.users.l1_researcher.email,
                flowRootState.users.l1_researcher.password
            );
            // Get initial GPU/CPU allocation
            initialGpuCount = await projectPage.getAllocatedGpuCount(projectName);
            initialCpuCount = await projectPage.getAllocatedCpuCount(projectName);

            // Create Training workload
            await training.submitTrainingWorkloadUI(projectName, trainingName, {
                environment: environmentName,
                computeResource: 'one-gpu'
            });
            await workloads.waitForWorkloadUIStatus(trainingName, EWorkloadStatus.Running);

            // Check workload status
            await workloads.checkDisplayedWorkloadDetails({
                jobName: trainingName,
                requestedGpu: '1',
                type: WorkloadType.Training,
                projectName,
                status: EWorkloadStatus.Running,
                createdBy: flowRootState.users.l1_researcher.email,
                pods: '1/1'
            });

            // Delete workload
            await workloads.deleteWorkload(trainingName);
            await waitUtils.waitForElementToNotExist(workloads, 'td:has-text(' + `"${trainingName}"` + ')');

            // Verify pod is deleted
            await waitUtils.waitForPodsDeletion(trainingName, namespace);

            // Check GPU/CPU allocation is released
            const finalGpuCount = await projectPage.getAllocatedGpuCount(projectName);
            const finalCpuCount = await projectPage.getAllocatedCpuCount(projectName);
            expect(finalGpuCount).toBe(initialGpuCount);
            expect(finalCpuCount).toBe(initialCpuCount);
        });

        test('T3738303: L1 researcher can create single node workspace workload', async () => {
            test.setTimeout(300 * 1000); // this test may take longer than the default timeout
            await auth.signIn(
                flowRootState.users.l1_researcher.email,
                flowRootState.users.l1_researcher.password
            );
            // Get initial GPU/CPU allocation
            initialGpuCount = await projectPage.getAllocatedGpuCount(projectName);
            initialCpuCount = await projectPage.getAllocatedCpuCount(projectName);

            // Create Workspace
            await workspace.submitWorkspaceWorkloadUI(projectName, workspaceName, {
                environment: environmentName,
                computeResource: 'one-gpu'
            });
            await workspace.waitForWorkspaceStatus(workspaceName, 'Running');

            // Check workload status
            await workloads.checkDisplayedWorkloadDetails({
                jobName: workspaceName,
                requestedGpu: '1',
                type: WorkloadType.Workspace,
                projectName,
                status: EWorkloadStatus.Running,
                createdBy: flowRootState.users.l1_researcher.email,
                pods: '1/1'
            });

            // Delete Workspace
            await workloads.deleteWorkload(workspaceName);
            await waitUtils.waitForElementToNotExist(workloads, 'td:has-text(' + `"${workspaceName}"` + ')');

            // Verify pod is deleted
            await waitUtils.waitForPodsDeletion(workspaceName, namespace);

            // Check GPU/CPU allocation is released
            const finalGpuCount = await projectPage.getAllocatedGpuCount(projectName);
            const finalCpuCount = await projectPage.getAllocatedCpuCount(projectName);
            expect(finalGpuCount).toBe(initialGpuCount);
            expect(finalCpuCount).toBe(initialCpuCount);
        });

        test('T4990202: L1 researcher can not create duplicated workload', async () => {
            test.setTimeout(300 * 1000);
            await auth.signIn(
                flowRootState.users.l1_researcher.email,
                flowRootState.users.l1_researcher.password
            );
            // Create a training workload
            await training.submitTrainingWorkloadUI(projectName, trainingName, {
                environment: environmentName,
                computeResource: 'one-gpu'
            });

            // Copy the training workload with the same name
            await workloads.toggleWorkload(trainingName);
            await workloads.clickOnCopyAndEditWorkload();

            // The second workload should fail due to the name already exists
            const errorFieldSelector = '.q-field--error.runai-name-validation';
            await expect(workloads.page.locator(errorFieldSelector)).toBeVisible();
            const fieldErrorSelector = '.q-field--error [role="alert"]';
            const errorText = await workloads.page.locator(fieldErrorSelector).innerText();
            expect(errorText).toBe('This name already exists');
        });

        test.describe(`Department/Application admin can create single node workloads (Training, GPU & CPU), tags: ${tags.runOn.dgxc}`, () => {
            async function testRoleCanCreateGpuAndCpuWorkload({
                auth,
                workloads,
                training,
                projectName,
                clusterName,
                role,
                namePrefix
            }: {
                auth: AuthPage;
                workloads: WorkloadPage;
                training: TrainingPage;
                projectName: string;
                clusterName: string;
                role: UserRole;
                namePrefix: string;
            }): Promise<void> {
                // Login as specified role
                await auth.signIn(
                    flowRootState.users[role].email,
                    flowRootState.users[role].password
                );
                // GPU workload
                const gpuTrainingName = objectGenerator.generateNameByPrefix(`gpu-wl-${namePrefix}`);
                await training.submitTrainingWorkloadUI(projectName, gpuTrainingName, {
                    environment: environmentName,
                    computeResource: 'one-gpu'
                });
                await workloads.waitForWorkloadUIStatus(gpuTrainingName, EWorkloadStatus.Running);
                // CPU workload
                const cpuTrainingName = objectGenerator.generateNameByPrefix(`cpu-wl-${namePrefix}`);
                await training.submitTrainingWorkloadUI(projectName, cpuTrainingName, {
                    environment: environmentName,
                    computeResource: 'cpu-only'
                });
                await workloads.waitForWorkloadUIStatus(cpuTrainingName, EWorkloadStatus.Running);
            }

            test('T3794043: Department administrator can create single node GPU and CPU training workload', async () => {
                await testRoleCanCreateGpuAndCpuWorkload({
                    auth,
                    workloads,
                    training,
                    projectName,
                    clusterName: flowRootState.cluster.name,
                    role: 'department_admin',
                    namePrefix: 'deptadmin'
                });
            });

            test('T3795606: Verify a failed workspace can be deleted', async () => {
                test.setTimeout(900 * 1000);

                await auth.signIn(
                    flowRootState.users.l1_researcher.email,
                    flowRootState.users.l1_researcher.password
                );

                // Get initial GPU/CPU allocation
                initialGpuCount = await projectPage.getAllocatedGpuCount(projectName);
                initialCpuCount = await projectPage.getAllocatedCpuCount(projectName);

                // Create Workspace with invalid command to force failure
                await workspace.submitWorkspaceWorkloadUI(projectName, workspaceName, {
                    environment: environmentName,
                    computeResource: 'one-gpu',
                    cmd: 'abc'
                });

                // Wait for workspace to fail
                await waitUtils.waitForCondition(
                    async () => {
                        try {
                            await workloads.waitForWorkloadUIStatus(workspaceName, EWorkloadStatus.Failed);
                            return true;
                        } catch {
                            return false;
                        }
                    },
                    'wait for workload to fail (up to 10min)',
                    true,
                    600
                );

                // Open details and check failure reason
                await workloads.toggleWorkloadDrawerContent(workspaceName);
                const expectedFailureFragments = [
                    'failed to create containerd task',
                    'OCI runtime create failed',
                    'runc create failed',
                    'unable to start container process',
                    'exec: "abc": executable file not found in $PATH'
                ];
                await waitUtils.waitForCondition(
                    async () => {
                        const el = workloads.page.locator('.runai-drawer-content');
                        const visible = await el.isVisible();
                        const text = visible ? await el.innerText() : '';
                        return (
                            visible && expectedFailureFragments.every((fragment) => text.includes(fragment))
                        );
                    },
                    'wait for all failure fragments in .runai-drawer-content',
                    true,
                    60
                );

                await workloads.closeWorkloadDrawerContent();

                // Delete the failed workspace
                await workloads.deleteWorkload(workspaceName);
                await waitUtils.waitForElementToNotExist(
                    workloads,
                    'td:has-text(' + `"${workspaceName}"` + ')'
                );

                // Verify pod is deleted
                await waitUtils.waitForPodsDeletion(workspaceName, namespace);

                // Check GPU/CPU allocation is released
                const finalGpuCount = await projectPage.getAllocatedGpuCount(projectName);
                const finalCpuCount = await projectPage.getAllocatedCpuCount(projectName);
                expect(finalGpuCount).toBe(initialGpuCount);
                expect(finalCpuCount).toBe(initialCpuCount);
            });
        });
    });
});

test.describe(`L1 Researcher can submit single node workload on Run:ai with hybrid submission, tags: ${tags.tool.ui} ${tags.resource.workloads} ${tags.teams.loong} ${tags.runOn.dgxc}`, () => {
    supportedFrom(SupportedVersions.V2_20)(function () {
        let browserEngine: BrowserEngine;
        let auth: AuthPage;
        let workloads: WorkloadPage;
        let training: TrainingPage;
        let projectPage: ProjectPage;
        let projectName: string;
        let namespace: string;
        let project;
        let credentialsPage: CredentialPage;
        let videoName: string;
        test.beforeAll(async () => {
            logger.info('Enabling hybrid submission feature');
            await unleashClient.enableFeature(EFeatureToggleKeys.SHOW_HYBRID_SUBMISSION);
            await apiUtils.enableHybridSubmission();

            projectName = objectGenerator.generateNameByPrefix('e2e-project');
            namespace = Utils.namespaceFromProjectName(projectName);
            project = await projectResourcesUtils.createProject(namespace, {
                name: projectName,
                deservedGpus: 0,
                maxAllowedGpus: 0
            });
        });

        test.afterAll(async () => {
            logger.info('Disabling hybrid submission feature');

            if (project && project.id) {
                try {
                    await operations.deleteProjectAndNamespace(project.id);
                    logger.info(`deleted project ${projectName} and namespace ${namespace}`);
                } catch {
                    logger.error('failed to delete project ' + projectName + ' and namespace ' + namespace);
                }
            }
        });

        test.beforeEach(async () => {
            browserEngine = new BrowserEngine();
            await browserEngine.start();
            auth = new AuthPage(browserEngine.currentPage);
            workloads = new WorkloadPage(browserEngine.currentPage);
            training = new TrainingPage(browserEngine.currentPage);
            projectPage = new ProjectPage(browserEngine.currentPage);
            credentialsPage = new CredentialPage(browserEngine.currentPage);
            await auth.signIn(
                flowRootState.users.l1_researcher.email,
                flowRootState.users.l1_researcher.password
            );
        });

        test.afterEach(async ({}, testInfo) => {
            await browserEngine.stopAndArchive(testInfo, 'tenant-user-secret-workload');
        });

        test('T3738308: L1 Researcher can store and use secrets on Run:ai in single node job', async () => {
            test.setTimeout(300 * 1000);
            videoName = 'T3738308-single-node-job-with-secrets';
            const secretName = 'my-secret';
            const namespace = `runai-${projectName}`;
            const fakeUsername = 'fake-username';
            const fakePassword = 'fake-abc1234';
            const secretData = {
                username: Buffer.from(fakeUsername).toString('base64'),
                password: Buffer.from(fakePassword).toString('base64')
            };

            await kubeCli.createSecret(secretName, namespace, secretData, 'Opaque', {
                'run.ai/resource': 'generic'
            });
            logger.info(`Created secret ${secretName} in namespace ${namespace}`);

            const secret = await kubeCli.getSecret(secretName, namespace);
            logger.info(
                `Secret verification: Secret ${secretName} exists with keys: ${Object.keys(secret.data).join(', ')}`
            );

            const trainingJobName = objectGenerator.generateNameByPrefix('secret-test');
            await training.createFlexibleTraining(projectName, trainingJobName, null, [
                { name: 'MYUSERNAME', sourcetype: 'Secret', sourcename: secretName, sourcevalue: 'username' },
                { name: 'MYPASSWORD', sourcetype: 'Secret', sourcename: secretName, sourcevalue: 'password' }
            ]);

            await workloads.waitForWorkloadUIStatus(trainingJobName, EWorkloadStatus.Running);

            logger.info(`Checking environment variables in pod for job ${trainingJobName}`);
            runaiCliV2.projectSet(projectName);
            const execFlag = {
                [execFlags.command]: 'env'
            } as Partial<typeof execFlags>;
            const execOutput = runaiCliV2.exec(execFlag, trainingJobName, cliV2WorkloadType.training);
            const envOutput = execOutput.stdout;
            logger.info('Environment variables output retrieved, filtering for target variables');

            const usernameMatch = envOutput.match(/MYUSERNAME=(.*)/);
            const passwordMatch = envOutput.match(/MYPASSWORD=(.*)/);

            expect(usernameMatch).not.toBeNull();
            expect(passwordMatch).not.toBeNull();
            expect(usernameMatch[1].trim()).toBe(fakeUsername);
            expect(passwordMatch[1].trim()).toBe(fakePassword);

            await test.step('Delete secret and workload', async () => {
                await kubeCli.deleteSecret(secretName, namespace);
                await workloads.deleteWorkload(trainingJobName);
                await waitUtils.waitForElementToNotExist(
                    workloads,
                    'td:has-text(' + `"${trainingJobName}"` + ')'
                );
                await waitUtils.waitForPodsDeletion(trainingJobName, namespace);
            });
        });

        test('T3795579: Verify a L1 Researcher can create username and password credentials', async () => {
            test.setTimeout(300 * 1000);
            videoName = 'T3795579-single-node-job-with-username-password-credential';
            const credentialName = objectGenerator.generateNameByPrefix('cred');
            const secretName = `password-${credentialName}`;
            const username = 'username';
            const password = 'password';
            const trainingJobName = objectGenerator.generateNameByPrefix('cred-test');

            await test.step('Create username & password credential', async () => {
                await credentialsPage.createUserPassCredential(credentialName, username, password, {
                    scope: Scope.Project,
                    scopeEntityName: projectName
                });
            });

            await test.step('Check secret exists in Kubernetes', async () => {
                logger.info(`Waiting for secret ${secretName} to be created...`);
                await waitUtils.waitForCondition(
                    async () => {
                        const secrets = await kubeCli.listSecretsByNamepsace(namespace);
                        return secrets.items.some((s) => s.metadata.name === secretName);
                    },
                    `Wait for secret ${secretName} to be created in namespace ${namespace}`,
                    true,
                    60
                );
                const secret = await kubeCli.getSecret(secretName, namespace);
                logger.info(
                    `Secret verification: Secret ${secretName} exists with keys: ${Object.keys(secret.data).join(', ')}`
                );
            });

            await test.step('Create single node job with credential environment variables', async () => {
                await training.createFlexibleTraining(projectName, trainingJobName, null, [
                    {
                        name: 'MYUSERNAME',
                        sourcetype: 'Secret',
                        sourcename: secretName,
                        sourcevalue: 'User'
                    },
                    {
                        name: 'MYPASSWORD',
                        sourcetype: 'Secret',
                        sourcename: secretName,
                        sourcevalue: 'Password'
                    }
                ]);
            });
            await test.step('Wait for job to be running', async () => {
                await workloads.waitForWorkloadUIStatus(trainingJobName, EWorkloadStatus.Running);
            });

            test.step('Verify environment variables in the pod', () => {
                logger.info(`Checking environment variables in pod for job ${trainingJobName}`);
                runaiCliV2.projectSet(projectName);

                const execFlag = {
                    [execFlags.command]: 'env'
                } as Partial<typeof execFlags>;
                const execOutput = runaiCliV2.exec(execFlag, trainingJobName, cliV2WorkloadType.training);
                const envOutput = execOutput.stdout;
                logger.info('Environment variables output retrieved, filtering for target variables');

                const usernameMatch = envOutput.match(/MYUSERNAME=(.*)/);
                const passwordMatch = envOutput.match(/MYPASSWORD=(.*)/);

                expect(usernameMatch).not.toBeNull();
                expect(passwordMatch).not.toBeNull();
                expect(usernameMatch[1].trim()).toBe(username);
                expect(passwordMatch[1].trim()).toBe(password);
            });

            await test.step('Delete secret and workload', async () => {
                const credentialId = (await credentialUtils.list()).find(
                    (c) => c.meta.name === credentialName
                )?.meta.id;
                if (credentialId) await credentialUtils.deletePasswordAsset(credentialId);
                await workloads.deleteWorkload(trainingJobName);
                await waitUtils.waitForElementToNotExist(
                    workloads,
                    'td:has-text(' + `"${trainingJobName}"` + ')'
                );
                await waitUtils.waitForPodsDeletion(trainingJobName, namespace);
            });
        });

        test('T3795580: Verify a L1 Researcher can create access key credentials', async () => {
            test.setTimeout(300 * 1000);
            videoName = 'T3795580-single-node-job-with-access-key-credential';
            const credentialName = objectGenerator.generateNameByPrefix('access-cred');

            const secretName = `accesskey-${credentialName}`;
            const accessKey = 'accesskey';
            const accessSecret = 'accessvalue';
            const trainingJobName = objectGenerator.generateNameByPrefix('access-test');

            await test.step('Create access key credential', async () => {
                await credentialsPage.createAccessKeyCredential(credentialName, accessKey, accessSecret, {
                    scope: Scope.Project,
                    scopeEntityName: projectName
                });
            });

            await test.step('Check secret exists in Kubernetes', async () => {
                logger.info(`Waiting for secret ${secretName} to be created...`);
                await waitUtils.waitForCondition(
                    async () => {
                        const secrets = await kubeCli.listSecretsByNamepsace(namespace);
                        return secrets.items.some((s) => s.metadata.name === secretName);
                    },
                    `Wait for secret ${secretName} to be created in namespace ${namespace}`,
                    true,
                    60
                );
                logger.info(`Checking for secret ${secretName} in namespace ${namespace}`);
                const secret = await kubeCli.getSecret(secretName, namespace);
                const secretStr = JSON.stringify(secret, null, 2);
                logger.info(
                    `Secret verification: Secret ${secretName} exists with keys: ${Object.keys(secret.data).join(', ')}`
                );
                expect(secret).toBeTruthy();
                expect(secretStr).toContain(secretName);
                expect(Object.keys(secret.data)).toContain('AccessKeyId');
                expect(Object.keys(secret.data)).toContain('SecretKey');
            });

            await test.step('Create single node job with credential environment variables', async () => {
                await training.createFlexibleTraining(projectName, trainingJobName, null, [
                    {
                        name: 'MYKEY',
                        sourcetype: 'Secret',
                        sourcename: `${secretName}`,
                        sourcevalue: 'AccessKeyId'
                    },
                    {
                        name: 'MYVALUE',
                        sourcetype: 'Secret',
                        sourcename: `${secretName}`,
                        sourcevalue: 'SecretKey'
                    }
                ]);
            });

            await test.step('Wait for job to be running', async () => {
                await workloads.waitForWorkloadUIStatus(trainingJobName, EWorkloadStatus.Running);
            });

            test.step('Verify environment variables in the pod', () => {
                logger.info(`Checking environment variables in pod for job ${trainingJobName}`);
                runaiCliV2.projectSet(projectName);

                const execFlag = {
                    [execFlags.command]: 'env'
                } as Partial<typeof execFlags>;
                const execOutput = runaiCliV2.exec(execFlag, trainingJobName, cliV2WorkloadType.training);
                const envOutput = execOutput.stdout;
                logger.info('Environment variables output retrieved: ' + envOutput);

                const keyMatch = envOutput.match(/MYKEY=(.*)/);
                const valueMatch = envOutput.match(/MYVALUE=(.*)/);

                expect(keyMatch).not.toBeNull();
                expect(valueMatch).not.toBeNull();
                expect(keyMatch[1].trim()).toBe(accessKey);
                expect(valueMatch[1].trim()).toBe(accessSecret);
            });

            await test.step('Delete secret and workload', async () => {
                const credentialId = (await credentialUtils.list()).find(
                    (c) => c.meta.name === credentialName
                )?.meta.id;
                if (credentialId) await credentialUtils.deleteAccessKeyAsset(credentialId);
                await workloads.deleteWorkload(trainingJobName);
                await waitUtils.waitForElementToNotExist(
                    workloads,
                    'td:has-text(' + `"${trainingJobName}"` + ')'
                );
                await waitUtils.waitForPodsDeletion(trainingJobName, namespace);
            });
        });
    });
});
