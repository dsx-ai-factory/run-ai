/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    Utils,
    flowRootState,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    config,
    waitUtils,
    settings,
    gpuTypes,
    environmentUtils,
    workloadsCrudUtils
} from '../../../../../../modules/helpers';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { TrainingPage } from '../../../../../../modules/ui/runaiV2/pages/TrainingPage';
import { EWorkloadStatus, WorkloadPage } from '../../../../../../modules/ui/runaiV2/pages/WorkloadPage';
import { EnvironmentPage } from '../../../../../../modules/ui/runaiV2/pages/EnvironmentPage';
import {
    supportedFrom,
    SupportedVersions,
    shouldRunTests
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { expect, test } from '@playwright/test';
import { DISTRIBUTED_TOOLS_LABELS } from '../../../../../../modules/ui/runaiV2/pages/TrainingPage';
import {
    WorkloadType,
    EnvironmentAsset,
    Phase,
    WorkloadSupportedTypes,
    Scope
} from '../../../../../../modules/api/apis/assetsApi';
import { WorkloadType as apiWorkloadType } from '../../../../../../modules/helpers/utils/WorkloadsCrudUtils';
import { DistributedFramework } from '../../../../../../modules/api/apis/workloadsApi/generated/model/distributedFramework';
import { ProjectPage } from '../../../../../../modules/ui/runaiV2/pages/ProjectPage';
import { kubeCli } from '../../../../../../modules/cli';
import { runaiCliV2 } from '../../../../../../modules/cli';
import { cliV2WorkloadType, execFlags } from '../../../../../../modules/cli/tools/RunaiCliV2.constants';
import { WorkloadDetailsDrawer } from '../../../../../../modules/ui/runaiV2/pages/WorkloadDetailsDrawer';
import { HybridForm } from '../../../../../../modules/ui/runaiV2/pages/HybridSubmissionForm';
import { CredentialPage } from '../../../../../../modules/ui/runaiV2/pages/CredentialPage';
import { V1Pod } from '@kubernetes/client-node';

test.describe(`Training UI - sanity, tags: ${tags.resource.workloads} ${tags.teams.loong} ${tags.runOn.dgxc}`, () => {
    let browserEngine: BrowserEngine;
    let auth: AuthPage;
    let training: TrainingPage;
    let projectPage: ProjectPage;
    let workloads: WorkloadPage;
    let videoName: string;
    let projectName: string;
    let namespace: string;
    let project: Project;
    let trainingName: string;
    let initialGpuCount: number;
    let initialCpuCount: number;
    let detailsDrawer: WorkloadDetailsDrawer;
    let credentialsPage: CredentialPage;

    test.beforeAll(async function () {
        browserEngine = new BrowserEngine();
        projectName = objectGenerator.generateShortNameByPrefix('project-e2e-for-training');
        namespace = Utils.namespaceFromProjectName(projectName);
        project = await projectResourcesUtils.createProject(namespace, {
            name: projectName,
            deservedGpus: 3,
            maxAllowedGpus: 3
        });
    });

    test.afterAll('deleting project and namespace', async () => {
        try {
            await operations.deleteProjectAndNamespace(project.id);
            logger.info(`deleted project ${projectName} and namespace ${namespace}`);
        } catch (error) {
            logger.error(`failed to delete project ${projectName} and namespace ${namespace}`);
            throw error;
        }
    });

    test.beforeEach(async function () {
        await browserEngine.start();
        auth = new AuthPage(browserEngine.currentPage);
        training = new TrainingPage(browserEngine.currentPage);
        workloads = new WorkloadPage(browserEngine.currentPage);
        projectPage = new ProjectPage(browserEngine.currentPage);
        credentialsPage = new CredentialPage(browserEngine.currentPage);
        trainingName = objectGenerator.generateShortNameByPrefix('training');
        detailsDrawer = new WorkloadDetailsDrawer(browserEngine.currentPage);
        await auth.signIn(
            flowRootState.users.l1_researcher.email,
            flowRootState.users.l1_researcher.password
        );
        initialGpuCount = await projectPage.getAllocatedGpuCount(projectName);
        initialCpuCount = await projectPage.getAllocatedCpuCount(projectName);
    });

    test.afterEach('stop browser', async ({}, testInfo) => {
        await browserEngine.stopAndArchive(testInfo, 'training-' + videoName);
    });

    test.describe('MPI distributed workload hybrid submission', () => {
        supportedFrom(SupportedVersions.V2_20)(function () {
            test.setTimeout(480 * 1000);
            test('T3738302: L1 researcher can create distributed MPI Training workload', async () => {
                videoName = 'mpi-distributed-training';
                await test.step('Create MPI distributed job', async () => {
                    await training.fillFirstNewTrainingPage(
                        projectName,
                        trainingName,
                        true,
                        DISTRIBUTED_TOOLS_LABELS.MPI,
                        true,
                        true
                    );
                    await training.fillHybridDistributedTrainingSecondPage(
                        'ubuntu',
                        {
                            command: 'sleep infinity',
                            args: '',
                            computeLabel: 'one-gpu',
                            workersCount: '2',
                            imagePullPolicy: 'Pull the image only if it'
                        },
                        {
                            image: config.quickStartImage.ubuntu,
                            opts: {
                                command: 'sleep infinity',
                                computeLabel: 'cpu-only'
                            }
                        } // masterOpts
                    );
                    await training.waitForWorkloadUIStatus(trainingName, EWorkloadStatus.Running, 300 * 1000);

                    await workloads.searchByName(trainingName);

                    await training.checkDisplayedWorkloadDetails({
                        jobName: trainingName,
                        requestedGpu: '2',
                        type: WorkloadType.Distributed,
                        projectName,
                        status: EWorkloadStatus.Running,
                        createdBy: flowRootState.users.l1_researcher.email,
                        pods: '3/3'
                    });
                });

                await test.step('Deleting the job', async () => {
                    await training.deleteWorkload(trainingName);
                    await waitUtils.waitForElementToNotExist(
                        training,
                        'td:has-text(' + `"${trainingName}"` + ')'
                    );
                });

                await test.step('Verifying pods are deleted', async () => {
                    await waitUtils.waitForPodsDeletion(trainingName, namespace, 120);
                });

                await test.step('Verifying resources are released', async () => {
                    await projectPage.goToPage();
                    const finalGpuCount = await projectPage.getAllocatedGpuCount(projectName);
                    const finalCpuCount = await projectPage.getAllocatedCpuCount(projectName);
                    expect(finalGpuCount).toBe(initialGpuCount);
                    expect(finalCpuCount).toBe(initialCpuCount);
                });
            });

            test('T3738309: L1 researcher can store and use secrets on Run:ai in distributed job', async () => {
                test.setTimeout(480 * 1000);
                videoName = 'T3738309-mpi-distributed-training-with-secrets';
                const credentialName = objectGenerator.generateNameByPrefix('cred');
                const secretName = `password-${credentialName}`;
                const username = 'username';
                const password = 'abc1234';

                await test.step('Create project-scoped username and password credential', async () => {
                    await credentialsPage.createUserPassCredential(
                        credentialName,
                        username,
                        password,
                        {
                            scope: Scope.Project,
                            scopeEntityName: projectName
                        }
                    );
                    await waitUtils.waitForCondition(
                        async () => {
                            const secrets = await kubeCli.listSecretsByNamepsace(namespace);
                            return secrets.items.some((secret) => secret.metadata.name === secretName);
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

                const trainingJobName = objectGenerator.generateNameByPrefix('secret-test');
                await test.step('Create distributed training job with secrets', async () => {
                    await training.fillFirstNewTrainingPage(
                        projectName,
                        trainingJobName,
                        true,
                        DISTRIBUTED_TOOLS_LABELS.MPI,
                        true,
                        true
                    );
                    await training.fillHybridDistributedTrainingSecondPage(
                        config.quickStartImage.distributed,
                        {
                            command: 'sleep infinity',
                            args: '',
                            computeLabel: 'cpu-only',
                            workersCount: '2',
                            imagePullPolicy: 'Pull the image only if it',
                            envVars: [
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
                            ]
                        },
                        {
                            image: config.quickStartImage.distributed,
                            opts: {
                                command: 'sleep infinity',
                                computeLabel: 'cpu-only'
                            }
                        } // masterOpts
                    );
                    logger.info('Waiting for distributed training job to be running');
                    await training.waitForWorkloadUIStatus(
                        trainingJobName,
                        EWorkloadStatus.Running,
                        300 * 1000
                    );
                });

                test.step('Verify environment variables in the pod', () => {
                    logger.info(`Checking environment variables in pod for job ${trainingJobName}`);
                    runaiCliV2.projectSet(projectName);
                    const execFlag = {
                        [execFlags.command]: 'env'
                    } as Partial<typeof execFlags>;
                    const execOutput = runaiCliV2.exec(execFlag, trainingJobName, cliV2WorkloadType.mpi);
                    const envOutput = execOutput.stdout;
                    logger.info('Environment variables output retrieved, filtering for target variables');

                    const usernameMatch = envOutput.match(/MYUSERNAME=(.*)/);
                    const passwordMatch = envOutput.match(/MYPASSWORD=(.*)/);

                    expect(usernameMatch).not.toBeNull();
                    expect(passwordMatch).not.toBeNull();
                    expect(usernameMatch[1].trim()).toBe(username);
                    expect(passwordMatch[1].trim()).toBe(password);
                });

                await test.step('Delete credential and workload', async () => {
                    await training.deleteWorkload(trainingJobName);
                    await credentialsPage.goToPage();
                    await credentialsPage.deleteCredential(credentialName);
                });
            });

            test('T3751342: Verify a L1 researcher can create distributed Training workload with different setup for master and worker', async () => {
                test.setTimeout(480 * 1000);
                videoName = 'T3751342-distributed-training-different-master-worker';
                const trainingJobName = objectGenerator.generateNameByPrefix('diff-master-worker');

                await test.step('Create training job with different master and worker setup', async () => {
                    await training.fillFirstNewTrainingPage(
                        projectName,
                        trainingJobName,
                        true,
                        DISTRIBUTED_TOOLS_LABELS.MPI,
                        true,
                        true
                    );
                    await training.fillHybridDistributedTrainingSecondPage(
                        config.quickStartImage.distributed,
                        {
                            command: 'sleep infinity',
                            computeLabel: 'one-gpu',
                            workersCount: '2',
                            imagePullPolicy: 'Pull the image only if it'
                        },
                        {
                            image: config.quickStartImage.ubuntu,
                            opts: {
                                command: 'sleep infinity',
                                computeLabel: 'cpu-only'
                            }
                        } // masterOpts
                    );
                    logger.info('Waiting for distributed training job to be running');
                    await training.waitForWorkloadUIStatus(
                        trainingJobName,
                        EWorkloadStatus.Running,
                        300 * 1000
                    );
                });

                await test.step('Verify workload details', async () => {
                    await workloads.checkDisplayedWorkloadDetails({
                        jobName: trainingJobName,
                        requestedGpu: '2',
                        type: WorkloadType.Distributed,
                        projectName,
                        status: EWorkloadStatus.Running,
                        createdBy: flowRootState.users.l1_researcher.email,
                        pods: '3/3'
                    });
                });

                await test.step('Delete workload', async () => {
                    await workloads.searchByName(trainingJobName);
                    await training.deleteWorkload(trainingJobName);
                });
            });

            test('T3753003: Verify L1 reasercher can create distributed training Pytorch workload', async () => {
                test.setTimeout(480 * 1000);
                videoName = 'T3753003-distributed-pytorch-training';
                // For PyTorch distributed training:
                // - PET_NNODES = total number of nodes
                // - PET_NPROC_PER_NODE = processes per node (typically 1 for mixed GPU/CPU setup)
                // - WORLD_SIZE = total number of processes across all nodes
                const expectedNNodes = 3;
                const expectedNProcPerNode = 1;
                const expectedWorldSize = 3;
                const pytorchDistributedTestImage = config.quickStartImage.distributed;

                await test.step('Create Pytorch distributed training workload', async () => {
                    await training.fillFirstNewTrainingPage(
                        projectName,
                        trainingName,
                        true,
                        DISTRIBUTED_TOOLS_LABELS.PY_TORCH,
                        true,
                        true
                    );
                    await training.fillHybridDistributedTrainingSecondPage(
                        pytorchDistributedTestImage,
                        {
                            command: 'sleep infinity',
                            args: '',
                            computeLabel: 'one-gpu',
                            workersCount: '2',
                            imagePullPolicy: 'Pull the image only if it'
                        },
                        {
                            image: pytorchDistributedTestImage,
                            opts: {
                                command: 'sleep infinity',
                                computeLabel: 'cpu-only'
                            }
                        } // masterOpts
                    );
                    await training.waitForWorkloadUIStatus(trainingName, EWorkloadStatus.Running, 300 * 1000);

                    await training.checkDisplayedWorkloadDetails({
                        jobName: trainingName,
                        requestedGpu: '2',
                        type: WorkloadType.Distributed,
                        projectName,
                        status: EWorkloadStatus.Running,
                        createdBy: flowRootState.users.l1_researcher.email,
                        pods: '3/3'
                    });
                });

                await test.step('Verify PyTorch environment variables in pods', async () => {
                    logger.info(
                        `Checking PyTorch environment variables in pods for workload ${trainingName}`
                    );
                    runaiCliV2.projectSet(projectName);

                    const pods = await kubeCli.listPodsByNamespaceAndPrefix(namespace, trainingName);
                    expect(pods.length).toBeGreaterThan(0);
                    logger.info(`Found ${pods.length} pods for training workload ${trainingName}`);

                    for (const pod of pods) {
                        logger.info(`Checking environment variables in pod: ${pod.metadata.name}`);

                        const podYaml = await kubeCli.getPodYaml(pod.metadata.name, namespace);
                        const container = (podYaml as V1Pod).spec?.containers[0];
                        const envVars = container.env || [];

                        const worldSizeEnv = envVars.find((env) => env.name === 'WORLD_SIZE');
                        const nProcPerNodeEnv = envVars.find((env) => env.name === 'PET_NPROC_PER_NODE');
                        const nNodesEnv = envVars.find((env) => env.name === 'PET_NNODES');

                        logger.info(
                            `Pod ${pod.metadata.name} has environment variables: WORLD_SIZE=${worldSizeEnv.value}, PET_NPROC_PER_NODE=${nProcPerNodeEnv.value}, PET_NNODES=${nNodesEnv.value}`
                        );

                        expect(worldSizeEnv.value).toBe(expectedWorldSize.toString());
                        expect(nProcPerNodeEnv.value).toBe(expectedNProcPerNode.toString());
                        expect(nNodesEnv.value).toBe(expectedNNodes.toString());
                    }
                });

                await test.step('Deleting the workload', async () => {
                    await training.deleteWorkload(trainingName);
                    await waitUtils.waitForElementToNotExist(
                        training,
                        'td:has-text(' + `"${trainingName}"` + ')'
                    );
                });

                await test.step('Verifying pods are deleted', async () => {
                    await waitUtils.waitForPodsDeletion(trainingName, namespace);
                });

                await test.step('Verifying resources are released', async () => {
                    await projectPage.goToPage();
                    const finalGpuCount = await projectPage.getAllocatedGpuCount(projectName);
                    const finalCpuCount = await projectPage.getAllocatedCpuCount(projectName);
                    expect(finalGpuCount).toBe(initialGpuCount);
                    expect(finalCpuCount).toBe(initialCpuCount);
                });
            });

            test('T3738311: L1 researcher can get the log of distributed job', async () => {
                videoName = 'log-of-distributed-job';
                test.setTimeout(480 * 1000);
                await test.step('Create MPI distributed job', async () => {
                    await training.fillFirstNewTrainingPage(
                        projectName,
                        trainingName,
                        true,
                        DISTRIBUTED_TOOLS_LABELS.MPI,
                        true,
                        true
                    );
                    await training.fillHybridDistributedTrainingSecondPage(
                        config.quickStartImage.distributedMpi,
                        {
                            computeLabel: 'one-gpu',
                            workersCount: '2',
                            imagePullPolicy: 'Always pull the image from the registry'
                        },
                        {
                            image: config.quickStartImage.ubuntu,
                            opts: {
                                command: 'sleep infinity',
                                computeLabel: 'cpu-only'
                            }
                        } // masterOpts
                    );
                    await training.waitForWorkloadUIStatus(trainingName, EWorkloadStatus.Running, 300 * 1000);

                    await detailsDrawer.clickShowDetailsBtn(trainingName);
                    await detailsDrawer.clickOnLogsTab();

                    // Verify the log info for each worker pod and launcher pod
                    const pods = await kubeCli.listPodsOfJob(namespace, trainingName);
                    for (const pod of pods) {
                        await detailsDrawer.selectWorkerPod(pod.metadata.name);
                        await detailsDrawer.waitForLogContentToAppear();
                        const logContent = await detailsDrawer.getLogContent();

                        const isMasterPod = pod.metadata.name.includes('launcher');
                        if (isMasterPod) {
                            expect(logContent).toContain('no logs for this workload yet');
                        } else {
                            expect(logContent).toContain('Server listening on 0.0.0.0 port 22');
                        }
                        logger.info(`logContent of ${pod.metadata.name} : ${logContent}`);
                    }
                });
            });

            test('T5073920: L1 researcher can create distributed JAX Training workload', async () => {
                logger.info(
                    'Skip this test case because there is 400 error when creating JAX distributed training workload'
                );
                test.skip();
                test.setTimeout(480 * 1000);
                videoName = 'T5073920-distributed-jax-training';
                let workloadId: string;
                const jaxTrainingName = objectGenerator.generateShortNameByPrefix('jax-dist-training');
                workloadsCrudUtils.accessToken = flowRootState.users.l1_researcher.accessToken;

                await test.step('Create JAX distributed training workload', async () => {
                    const { body: training } = await workloadsCrudUtils.createDistributed(
                        jaxTrainingName,
                        true,
                        flowRootState.cluster.uuid,
                        project.id.toString(),
                        {
                            image: 'ghcr.io/nvidia/jax:jax',
                            command: 'sleep infinity',
                            numWorkers: 2,
                            distributedFramework: DistributedFramework.Jax
                        },
                        true
                    );
                    workloadId = training.workloadId;
                    await workloadsCrudUtils.waitForWorkloadPhase(
                        workloadId,
                        Phase.Running,
                        apiWorkloadType.TRAINING,
                        undefined,
                        300
                    );
                });

                await test.step('Stop the workload', async () => {
                    await workloadsCrudUtils.suspendTraining(workloadId);
                });

                await test.step('Check workload is stopped and pods are deleted', async () => {
                    await workloadsCrudUtils.waitForWorkloadPhase(
                        workloadId,
                        Phase.Stopped,
                        apiWorkloadType.TRAINING
                    );
                });

                await test.step('Resume the workload', async () => {
                    await workloadsCrudUtils.resumeTraining(workloadId);
                    await workloadsCrudUtils.waitForWorkloadPhase(
                        workloadId,
                        Phase.Running,
                        apiWorkloadType.TRAINING
                    );
                });

                await test.step('Delete the workload', async () => {
                    await workloadsCrudUtils.deleteTraining(workloadId);
                });
            });
        });
    });

    test.describe('Distributed training with different frameworks', () => {
        shouldRunTests(settings.gpuType === gpuTypes.real)(function () {
            test('T3753004: Verify a L1 researcher can create distributed TensorFlow Training workload', async () => {
                videoName = 'T3753004-distributed-tensorflow-training';
                const trainingJobName = objectGenerator.generateNameByPrefix('dist-tensorflow');
                let environmentAsset: EnvironmentAsset;
                await test.step('Create environment', async () => {
                    environmentAsset = await environmentUtils.createEnvironmentWithArbitraryFields(
                        project.id,
                        '/bin/bash -c',
                        WorkloadSupportedTypes.DistFrameworkEnum.Tf,
                        config.quickStartImage.distributedTf,
                        '"sleep 300"'
                    );
                });
                await test.step('Create distributed TensorFlow training job', async () => {
                    await training.goToPage();
                    await training.fillFirstNewTrainingPage(
                        projectName,
                        trainingJobName,
                        true,
                        DISTRIBUTED_TOOLS_LABELS.TENSOR_FLOW,
                        false,
                        true
                    );
                    const hybridForm = new HybridForm(browserEngine.currentPage);
                    await hybridForm.openEnvironmentsDrawer();
                    await hybridForm.clickOnDrawerItemByLabel(environmentAsset.meta.name);
                    await hybridForm.fillComputeSection('one-gpu', 2);
                    await training.clickOnSubmitBtn(); // initial step
                    await training.waitForSubmitButton();
                    await training.clickOnSubmitBtn(); // setup step
                    logger.info('Waiting for distributed TensorFlow training job to be running');
                    await training.waitForWorkloadUIStatus(trainingJobName, EWorkloadStatus.Running);
                });

                await test.step('Verify workload details', async () => {
                    await workloads.checkDisplayedWorkloadDetails({
                        jobName: trainingJobName,
                        requestedGpu: '2',
                        type: WorkloadType.Distributed,
                        projectName,
                        status: EWorkloadStatus.Running,
                        createdBy: flowRootState.users.l1_researcher.email,
                        pods: '2/2'
                    });
                });

                await test.step('Delete workload and verify deletion', async () => {
                    await training.deleteWorkload(trainingJobName);
                    await waitUtils.waitForElementToNotExist(
                        training,
                        'td:has-text(' + `"${trainingJobName}"` + ')'
                    );
                    logger.info(`Verified workload ${trainingJobName} is deleted in UI`);
                    await waitUtils.waitForPodsDeletion(trainingJobName, namespace);
                    logger.info(`Verified pods for workload ${trainingJobName} are deleted in Kubernetes`);
                });

                await test.step('Delete environment', async () => {
                    await environmentUtils.delete(environmentAsset.meta.id);
                });
            });

            test('T3753005: Verify a L1 researcher can create distributed XGBoost Training workload', async () => {
                videoName = 'T3753005-distributed-xgboost-training';
                const distributedEnvironmentName = 'dist-xgboost-env';
                const trainingJobName = objectGenerator.generateNameByPrefix('dist-xgboost');
                let environmentAsset: EnvironmentAsset;
                await test.step('Create environment for XGBoost', async () => {
                    environmentAsset = await environmentUtils.createEnvironmentWithArbitraryFields(
                        project.id,
                        '',
                        WorkloadSupportedTypes.DistFrameworkEnum.XgBoost,
                        'docker.io/kubeflow/lightgbm-dist-py-test',
                        '--job_type=Train --boosting_type=gbdt --objective=binary --metric=binary_logloss,auc --metric_freq=1 --is_training_metric=true --is_training_metric=true --data=data/binary.train --valid_data=data/binary.test --valid_data=data/binary.test --valid_data=data/binary.test --valid_data=data/binary.test --valid_data=data/binary.test --feature_fraction=0.8 --bagging_freq=5 --bagging_fraction=0.8 --min_data_in_leaf=50 --min_sum_hessian_in_leaf=50 --is_enable_sparse=true --use_two_round_loading=false --is_save_binary_file=false'
                    );
                    logger.info(`XGBoost environment ${distributedEnvironmentName} created successfully`);
                });

                await test.step('Create distributed XGBoost training job', async () => {
                    await training.goToPage();
                    await training.fillFirstNewTrainingPage(
                        projectName,
                        trainingJobName,
                        true,
                        DISTRIBUTED_TOOLS_LABELS.XG_BOOST,
                        true,
                        true
                    );

                    const hybridForm = new HybridForm(browserEngine.currentPage);
                    await hybridForm.openEnvironmentsDrawer();
                    await hybridForm.clickOnDrawerItemByLabel(environmentAsset.meta.name);
                    await hybridForm.fillComputeSection('one-gpu', 2);
                    await training.clickOnSubmitBtn(); // worker form
                    await hybridForm.waitForDistributedMasterSubmitButton();
                    await hybridForm.clickOnSubmitMasterForm(); // master form
                    await training.waitForSubmitButton();
                    await training.clickOnSubmitBtn(); // setup form
                    logger.info('Waiting for distributed XGBoost training job to be running');
                    await training.waitForWorkloadUIStatus(trainingJobName, EWorkloadStatus.Running);
                });

                await test.step('Verify workload details', async () => {
                    await workloads.checkDisplayedWorkloadDetails({
                        jobName: trainingJobName,
                        requestedGpu: '3',
                        type: WorkloadType.Distributed,
                        projectName,
                        status: EWorkloadStatus.Running,
                        createdBy: flowRootState.users.l1_researcher.email,
                        pods: '3/3'
                    });
                });

                await test.step('Delete workload and verify deletion', async () => {
                    await training.deleteWorkload(trainingJobName);
                    await waitUtils.waitForElementToNotExist(
                        training,
                        'td:has-text(' + `"${trainingJobName}"` + ')'
                    );
                    logger.info(`Verified workload ${trainingJobName} is deleted in UI`);
                    await waitUtils.waitForPodsDeletion(trainingJobName, namespace);
                    logger.info(`Verified pods for workload ${trainingJobName} are deleted in Kubernetes`);
                });

                await test.step('Delete environment', async () => {
                    await environmentUtils.delete(environmentAsset.meta.id);
                });
            });
        });
    });
});
