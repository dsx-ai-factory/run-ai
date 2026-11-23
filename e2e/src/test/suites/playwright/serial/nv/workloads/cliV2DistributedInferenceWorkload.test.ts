/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import {
    config,
    flowRootState,
    logger,
    objectGenerator,
    orgUnitUtils,
    tags,
    waitUtils,
    workloadsCrudUtils,
    workloadsUtilsV2,
    workloadUtils,
    dataSourcesUtils
} from '../../../../../../modules/helpers';
import { Project } from '../../../../../../modules/api/apis/orgUnitApi';
import { Phase } from '../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import { isLwsInstalled } from '../../../../../../modules/helpers/utils/IsInstalledUtils';
import { kubeCli, runaiCliV2 } from '../../../../../../modules/cli';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test } from '@playwright/test';
import {
    cliV2WorkloadType,
    SubmitFlagsMerged,
    execFlags,
    portForwardFlags,
    bashFlags
} from '../../../../../../modules/cli/tools/RunaiCliV2.constants';
import {
    createBasicSubmitFlags,
    submitDistributedInferenceWorkload,
    identifyLeaderAndWorkerPods,
    getWorkloadIdAndWaitForPhase,
    describeWorkload,
    describeWorkloadAsync,
    waitForWorkloadByNameAndProjectId,
    verifyBasicWorkloadInfo,
    verifyWorkerReplicas,
    verifyReplicasCount,
    verifyGpuCompute,
    verifyImage,
    verifyCpuCores,
    verifyCpuMemory,
    verifyPhase,
    verifyEnvironmentVariable,
    verifyServingPort,
    verifyRunningPods
} from '../../../../../../modules/cli/helpers/DistributedInferenceTestHelpers';

test.describe(
    `Distributed Inference tests ${tags.teams.viper} ${tags.resource.distInferenceWorkload} ${tags.teams.loong}`,
    { tag: [tags.resource.distInferenceWorkload] },
    function () {
        supportedFrom(SupportedVersions.V2_24)(function () {
            let project: Project;
            let workloadId: string;
            let workloadName: string;
            let namespace: string;
            let createdWorkloadNames: string[] = [];
            let describeResult: string;
            let negativeWorkersWorkloadName: string;
            let workerPodName: string;
            let leaderPodName: string;
            const createdPvcNames: string[] = [];
            let portForwardProcess: any;

            test.beforeAll(async function () {
                if (!(await isLwsInstalled())) {
                    logger.warn('LeaderWorkerSet (LWS) is not installed on the cluster, skipping test');
                    test.skip();
                }

                project = await orgUnitUtils.createProjectAndValidate({
                    name: objectGenerator.generateShortNameByPrefix('project-inference-distributed-cli'),
                    clusterId: flowRootState.cluster.uuid,
                    resources: [{ gpu: { deserved: 8, limit: 8 } }]
                });
                namespace = project.status.namespace;
                runaiCliV2.loginAsUser(
                    flowRootState.users.system_administrator.email,
                    flowRootState.users.system_administrator.password
                );
                runaiCliV2.projectSet(project.name);
                workloadsCrudUtils.accessToken = flowRootState.users.system_administrator.accessToken;
            });

            test.afterAll(async function () {
                if (project) {
                    try {
                        await workloadsUtilsV2.deleteWorkloadsByProjectId(project.id, 120);
                        await orgUnitUtils.deleteProjectAndNamespace(project.id, project.name);
                    } catch (error) {
                        logger.error(
                            `Failed to delete workloads and project ${project.name} namespace ${project.status.namespace} with error ${error}`
                        );
                    }
                }
            });

            test.beforeEach(function () {
                workloadName = objectGenerator.generateShortNameByPrefix('distributed-inference-cli');
                negativeWorkersWorkloadName = objectGenerator.generateShortNameByPrefix('di-neg-workload');
            });

            test.afterEach(async function () {
                if (portForwardProcess && !portForwardProcess.killed) {
                    try {
                        logger.info('Cleaning up port-forward process in afterEach');
                        portForwardProcess.kill('SIGTERM');
                        await new Promise((resolve) => setTimeout(resolve, 500));
                        if (!portForwardProcess.killed) {
                            logger.info('Port-forward process still running, sending SIGKILL');
                            portForwardProcess.kill('SIGKILL');
                        }
                        logger.info('Port-forward process terminated in afterEach');
                    } catch (error) {
                        logger.error(`Failed to kill port-forward process: ${error}`);
                    }
                    portForwardProcess = null;
                }

                if (createdWorkloadNames.length > 0) {
                    logger.info(
                        `Cleaning up ${createdWorkloadNames.length} distributed inference workload(s)`
                    );
                    for (const workload of createdWorkloadNames) {
                        try {
                            const workloadId = await workloadsUtilsV2.getWorkloadIdByName(
                                workload,
                                project.id
                            );
                            await workloadsCrudUtils.deleteDistributedInference(workloadId);
                            logger.info(`Deleted distributed inference workload: ${workload}`);
                        } catch (error) {
                            const msg = (error as Error)?.message ?? String(error);
                            const isNotFound = /not found|404|does not exist/i.test(msg);
                            if (isNotFound) {
                                logger.warn(
                                    `Workload ${workload} not found during cleanup (may not have been created or already deleted)`
                                );
                            } else {
                                logger.error(`Failed to delete workload ${workload}: ${error}`);
                            }
                        }
                    }
                    createdWorkloadNames.length = 0;
                }

                if (createdPvcNames.length > 0) {
                    logger.info(`Cleaning up ${createdPvcNames.length} PVC(s)`);
                    for (const pvcName of createdPvcNames) {
                        try {
                            await kubeCli.CoreV1Api.deleteNamespacedPersistentVolumeClaim({
                                name: pvcName,
                                namespace
                            });
                            logger.info(`Deleted PVC: ${pvcName}`);
                        } catch (error) {
                            logger.error(`Failed to delete PVC ${pvcName}: ${error}`);
                        }
                    }
                    createdPvcNames.length = 0;
                }
            });

            test.describe('Help & Documentation', function () {
                test('T5576421: Verify Distributed Inference Help Command', async function () {
                    await test.step('Execute help command for distributed inference', () => {
                        const helpCommand = `${runaiCliV2.cliPath} inference distributed --help`;
                        const helpOutput = runaiCliV2.bashCli.exec(helpCommand);
                        logger.info(`Distributed inference help output: ${helpOutput}`);

                        expect(helpOutput).to.include(
                            'Runs multiple coordinated inference processes across multiple nodes'
                        );
                        logger.info('Verified distributed inference command description');

                        const expectedCommands = [
                            'bash',
                            'delete',
                            'exec',
                            'list',
                            'logs',
                            'port-forward',
                            'scale',
                            'submit'
                        ];
                        expectedCommands.forEach((command) => {
                            expect(helpOutput).to.include(command);
                        });
                        logger.info('Verified all expected commands are listed');

                        expect(helpOutput).to.match(/Global Flags|Flags:/i);
                        logger.info('Verified global flags section is present');
                    });

                    await test.step('Execute help command for inference', () => {
                        const inferenceHelpCommand = `${runaiCliV2.cliPath} inference --help`;
                        const inferenceHelpOutput = runaiCliV2.bashCli.exec(inferenceHelpCommand);
                        logger.info(`Inference help output: ${inferenceHelpOutput}`);

                        expect(inferenceHelpOutput).to.match(/Inference Architecture/i);
                        expect(inferenceHelpOutput).to.include('distributed');
                        expect(inferenceHelpOutput).to.include('standard');
                        logger.info(
                            'Verified Inference Architecture section lists both distributed and standard'
                        );
                    });
                });

                test('T5608160: Verify General Inference Help Command', async function () {
                    let inferenceHelpOutput: string;
                    await test.step('Execute general inference help command', () => {
                        const helpCommand = `${runaiCliV2.cliPath} inference --help`;
                        inferenceHelpOutput = runaiCliV2.bashCli.exec(helpCommand);
                        logger.info(`General inference help output: ${inferenceHelpOutput}`);
                    });

                    await test.step('Verify help contains inference management description', () => {
                        expect(inferenceHelpOutput).to.include('inference management');
                        logger.info('Verified inference management description is present');
                    });

                    await test.step('Verify Inference Architecture section lists both types', () => {
                        expect(inferenceHelpOutput).to.match(/Inference Architecture:/i);
                        expect(inferenceHelpOutput).to.include('distributed');
                        expect(inferenceHelpOutput).to.include('standard');
                        logger.info(
                            'Verified Inference Architecture section lists both distributed and standard types'
                        );
                    });

                    await test.step('Verify distributed architecture description', () => {
                        expect(inferenceHelpOutput).to.match(
                            /distributed.*multiple.*nodes|distributed.*coordinated.*processes/i
                        );
                        logger.info('Verified distributed architecture is properly documented');
                    });

                    await test.step('Verify standard architecture description', () => {
                        expect(inferenceHelpOutput).to.match(/standard.*single.*node/i);
                        logger.info('Verified standard architecture is properly documented');
                    });

                    await test.step('Verify Inference General Commands section', () => {
                        expect(inferenceHelpOutput).to.match(/Inference General Commands:/i);
                        expect(inferenceHelpOutput).to.include('list');
                        logger.info('Verified Inference General Commands section with list command');
                    });

                    await test.step('Verify Inference Standard Commands are listed', () => {
                        expect(inferenceHelpOutput).to.match(/Inference Standard Commands:/i);
                        const expectedStandardCommands = [
                            'bash',
                            'delete',
                            'describe',
                            'exec',
                            'logs',
                            'port-forward',
                            'submit',
                            'update'
                        ];
                        expectedStandardCommands.forEach((command) => {
                            expect(inferenceHelpOutput).to.include(command);
                            logger.info(`Verified standard command '${command}' is listed`);
                        });
                        logger.info('Verified all expected Inference Standard Commands are listed');
                    });

                    await test.step('Verify Flags section is present', () => {
                        expect(inferenceHelpOutput).to.match(/Flags:/i);
                        expect(inferenceHelpOutput).to.include('-h, --help');
                        logger.info('Verified Flags section is present');
                    });

                    await test.step('Verify Global Flags section is present', () => {
                        expect(inferenceHelpOutput).to.match(/Global Flags:/i);
                        expect(inferenceHelpOutput).to.include('--config-file');
                        expect(inferenceHelpOutput).to.include('--debug');
                        expect(inferenceHelpOutput).to.include('--quiet');
                        expect(inferenceHelpOutput).to.include('--verbose');
                        logger.info('Verified Global Flags section with expected flags');
                    });
                });

                test('T5576422: Verify Submit Command Help Documentation', async function () {
                    let submitHelpOutput: string;
                    await test.step('Execute submit help command', () => {
                        const submitHelpCommand = `${runaiCliV2.cliPath} inference distributed submit --help`;
                        submitHelpOutput = runaiCliV2.bashCli.exec(submitHelpCommand);
                        logger.info(`Distributed inference submit help output: ${submitHelpOutput}`);
                    });

                    await test.step('Verify "Distributed" section contains required flags', () => {
                        expect(submitHelpOutput).to.match(/Distributed/i);
                        expect(submitHelpOutput).to.include('--replicas');
                        expect(submitHelpOutput).to.include('--workers');
                        expect(submitHelpOutput).to.include('--worker-as-leader');
                        expect(submitHelpOutput).to.include('--startup-policy');
                        expect(submitHelpOutput).to.include('--worker-annotation');
                        expect(submitHelpOutput).to.include('--worker-label');
                        logger.info('Verified Distributed section contains required flags');
                    });

                    await test.step('Verify "Container" section contains both leader and worker flags', () => {
                        expect(submitHelpOutput).to.match(/Container/i);
                        expect(submitHelpOutput).to.include('--image');
                        expect(submitHelpOutput).to.include('--worker-image');
                        expect(submitHelpOutput).to.include('--environment');
                        expect(submitHelpOutput).to.include('--worker-environment');
                        expect(submitHelpOutput).to.include('--command');
                        expect(submitHelpOutput).to.include('--worker-command');
                        expect(submitHelpOutput).to.include('--arguments');
                        expect(submitHelpOutput).to.include('--worker-arguments');
                        logger.info('Verified Container section contains both leader and worker flags');
                    });

                    await test.step('Verify "Compute" section contains worker compute flags', () => {
                        expect(submitHelpOutput).to.match(/Compute/i);
                        expect(submitHelpOutput).to.include('--worker-gpu-devices-request');
                        expect(submitHelpOutput).to.include('--worker-cpu-core-request');
                        expect(submitHelpOutput).to.include('--worker-cpu-core-limit');
                        expect(submitHelpOutput).to.include('--worker-cpu-memory-request');
                        expect(submitHelpOutput).to.include('--worker-cpu-memory-limit');
                        logger.info('Verified Compute section contains worker compute flags');
                    });

                    await test.step('Verify "Examples" section contains examples', () => {
                        expect(submitHelpOutput).to.match(/Examples?:/i);
                        logger.info('Verified examples section is present');
                    });

                    await test.step('Verify "Flag Inheritance" section contains flag inheritance explanation', () => {
                        expect(submitHelpOutput).to.match(/inherit|inheritance|override/i);
                        logger.info('Verified flag inheritance explanation is present');
                    });
                });

                test('T5576423: Verify Scale Command Help Documentation', async function () {
                    let scaleHelpOutput: string;

                    await test.step('Execute scale help command', () => {
                        const scaleHelpCommand = `${runaiCliV2.cliPath} inference distributed scale --help`;
                        scaleHelpOutput = runaiCliV2.bashCli.exec(scaleHelpCommand);
                        logger.info(`Scale command help output: ${scaleHelpOutput}`);
                    });

                    await test.step('Verify command description', () => {
                        expect(scaleHelpOutput).to.include('scale a distributed inference workload');
                        logger.info('Verified scale command description');
                    });

                    await test.step('Verify usage format', () => {
                        expect(scaleHelpOutput).to.match(
                            /runai inference distributed scale \[WORKLOAD[_-]NAME\] \[flags\]/i
                        );
                        logger.info('Verified usage format');
                    });

                    await test.step('Verify --replicas flag is documented', () => {
                        expect(scaleHelpOutput).to.include('--replicas');
                        expect(scaleHelpOutput).to.match(/--replicas.*int32/i);
                        logger.info('Verified --replicas flag is documented as required');
                    });

                    await test.step('Verify example demonstrates proper usage', () => {
                        expect(scaleHelpOutput).to.match(/runai inference distributed scale.*--replicas.*5/i);
                        logger.info('Verified example is shown');
                    });

                    await test.step('Verify --project flag is documented', () => {
                        expect(scaleHelpOutput).to.match(/-p,?\s*--project/i);
                        logger.info('Verified -p, --project flag is documented');
                    });
                });
            });

            test.describe('Submit Operations', function () {
                test.describe('Basic Submit', function () {
                    test('T5576424: Submit Basic Distributed Inference Workload', async function () {
                        await test.step('Submit basic distributed workload using CLI', () => {
                            const submitFlags = createBasicSubmitFlags();

                            submitDistributedInferenceWorkload(
                                submitFlags,
                                workloadName,
                                createdWorkloadNames
                            );
                        });

                        await test.step('Verify workload created successfully via describe', async () => {
                            const describeResult = await describeWorkloadAsync(workloadName, project.id);
                            expect(describeResult).to.include(workloadName);
                            expect(describeResult).to.include('DistributedInference');
                            expect(describeResult).to.match(/Workers per Replica\s*:\s*2/i);
                            expect(describeResult).to.include(config.quickStartImage.inferenceHelloWorld);
                            logger.info('Verified CLI parameters correctly passed to backend');
                        });
                    });
                });

                test.describe('Worker-as-Leader Configuration', function () {
                    test('T5576429: Submit with Default Worker-as-Leader Inheritance', async function () {
                        await test.step('Submit workload without explicit worker-as-leader flag (defaults to true)', () => {
                            const submitFlags = createBasicSubmitFlags({
                                [SubmitFlagsMerged.envVar]: 'ROLE=LEADER'
                            });

                            submitDistributedInferenceWorkload(
                                submitFlags,
                                workloadName,
                                createdWorkloadNames,
                                'Submitted without explicit --worker-as-leader flag (defaults to true)'
                            );
                        });

                        await test.step('Wait for workload to be running', async () => {
                            workloadId = await getWorkloadIdAndWaitForPhase(
                                workloadName,
                                project.id,
                                Phase.Running,
                                120
                            );
                        });

                        await test.step('Verify default worker-as-leader inheritance behavior via describe', () => {
                            describeResult = describeWorkload(workloadName);

                            verifyBasicWorkloadInfo(describeResult, workloadName);
                            verifyWorkerReplicas(describeResult, 2);
                            verifyGpuCompute(describeResult, '3.0');
                            verifyEnvironmentVariable(describeResult, 'ROLE', 'LEADER');
                            logger.info('Verified default behavior: workers inherit leader configuration');
                        });
                    });

                    test('T5576430: Submit with Explicit Worker-as-Leader True', async function () {
                        await test.step('Submit with explicit --worker-as-leader flag', () => {
                            const submitFlags = createBasicSubmitFlags({
                                [SubmitFlagsMerged.envVar]: 'APP_ENV=production',
                                '--worker-as-leader': ''
                            });

                            submitDistributedInferenceWorkload(
                                submitFlags,
                                workloadName,
                                createdWorkloadNames,
                                'Submitted with explicit --worker-as-leader flag'
                            );
                        });

                        await test.step('Wait for workload to be running', async () => {
                            workloadId = await getWorkloadIdAndWaitForPhase(
                                workloadName,
                                project.id,
                                Phase.Running,
                                120
                            );
                        });

                        await test.step('Verify explicit flag accepted and behavior matches default', () => {
                            describeResult = describeWorkload(workloadName);

                            verifyBasicWorkloadInfo(describeResult, workloadName);
                            verifyGpuCompute(describeResult, '3.0');
                            verifyEnvironmentVariable(describeResult, 'APP_ENV', 'production');
                            logger.info(
                                'Verified explicit --worker-as-leader flag accepted and works same as default'
                            );
                        });
                    });

                    test('T5576433: Submit with Worker-as-Leader False', async function () {
                        await test.step('Submit with --worker-as-leader=false and explicit worker config', () => {
                            const submitFlags = createBasicSubmitFlags({
                                [SubmitFlagsMerged.gpuRequest]: '2',
                                [SubmitFlagsMerged.envVar]: 'LEADER_VAR=leader_value',
                                '--worker-as-leader=false': '',
                                '--worker-image': 'python:3.11',
                                '--worker-command': '"sleep infinity"',
                                '--worker-gpu-devices-request': '1',
                                '--worker-environment': 'WORKER_VAR=worker_value'
                            });

                            submitDistributedInferenceWorkload(
                                submitFlags,
                                workloadName,
                                createdWorkloadNames,
                                'Submitted with --worker-as-leader=false'
                            );
                        });

                        await test.step('Verify leader and worker configurations are independent via pod inspection', async () => {
                            const pods = await waitUtils.waitForPods(
                                workloadName,
                                project.status.namespace,
                                3,
                                120,
                                ['Running'],
                                true
                            );
                            const { leaderPod, workerPod } = identifyLeaderAndWorkerPods(pods);
                            expect(leaderPod).to.exist;
                            expect(workerPod).to.exist;
                            const leaderImage = leaderPod!.spec?.containers?.[0]?.image;
                            const workerImage = workerPod!.spec?.containers?.[0]?.image;
                            expect(leaderImage).to.include(config.quickStartImage.inferenceHelloWorld);
                            expect(workerImage).to.include('python:3.11');
                            logger.info(
                                `Leader image: ${config.quickStartImage.inferenceHelloWorld}, Worker image: python:3.11`
                            );

                            // Verify different environments
                            const leaderEnv = leaderPod!.spec?.containers?.[0]?.env?.find(
                                (e) => e.name === 'LEADER_VAR'
                            );
                            const workerEnv = workerPod!.spec?.containers?.[0]?.env?.find(
                                (e) => e.name === 'WORKER_VAR'
                            );
                            const leaderVarInWorker = workerPod!.spec?.containers?.[0]?.env?.find(
                                (e) => e.name === 'LEADER_VAR'
                            );
                            expect(leaderEnv?.value).to.equal('leader_value');
                            expect(workerEnv?.value).to.equal('worker_value');
                            expect(leaderVarInWorker).to.not.exist;
                            logger.info('Worker does not inherit leader environment');

                            // Verify different GPU allocations
                            const leaderGpu =
                                leaderPod!.spec?.containers?.[0]?.resources?.limits?.['nvidia.com/gpu'];
                            const workerGpu =
                                workerPod!.spec?.containers?.[0]?.resources?.limits?.['nvidia.com/gpu'];
                            expect(leaderGpu).to.equal('2');
                            expect(workerGpu).to.equal('1');
                            logger.info('Leader GPU: 2, Worker GPU: 1 (independent configurations)');
                            logger.info(
                                'Verified --worker-as-leader=false: NO inheritance, configurations are independent'
                            );
                        });
                    });

                    test('T5576434: Submit with Worker-as-Leader False Missing Worker Config', async function () {
                        await test.step('Submit with --worker-as-leader=false but no explicit worker config', () => {
                            const submitFlags = createBasicSubmitFlags({
                                [SubmitFlagsMerged.gpuRequest]: '2',
                                '--worker-as-leader=false': ''
                            });

                            submitDistributedInferenceWorkload(
                                submitFlags,
                                workloadName,
                                createdWorkloadNames,
                                'Submitted with --worker-as-leader=false but missing worker config'
                            );
                        });

                        await test.step('Verify system handles missing worker config gracefully', async () => {
                            describeResult = await describeWorkloadAsync(workloadName, project.id);

                            verifyBasicWorkloadInfo(describeResult, workloadName);
                            logger.info(
                                'Submission accepted: system provides defaults or inherits when worker config missing'
                            );

                            const pods = await waitUtils.waitForPods(
                                workloadName,
                                project.status.namespace,
                                3,
                                120,
                                ['Running'],
                                true
                            );
                            const { leaderPod, workerPod } = identifyLeaderAndWorkerPods(pods);

                            const leaderImage = leaderPod?.spec?.containers?.[0]?.image;
                            const workerImage = workerPod?.spec?.containers?.[0]?.image;
                            expect(workerImage).to.equal(leaderImage);
                            logger.info(
                                'Documented behavior: worker inherits leader config despite --worker-as-leader=false'
                            );
                        });
                    });
                });

                test.describe('Worker Configuration Overrides', function () {
                    test('T5576438: Submit with Worker Storage Configuration', async function () {
                        let testLeaderPvcName: string;
                        let testWorkerPvcName: string;

                        await test.step('Create PVCs for storage test', async () => {
                            const leaderPvc = await dataSourcesUtils.createClusterPvc(
                                'leader-pvc',
                                namespace
                            );
                            testLeaderPvcName = leaderPvc.metadata.name;
                            createdPvcNames.push(testLeaderPvcName);
                            logger.info(`Created leader PVC: ${testLeaderPvcName}`);

                            const workerPvc = await dataSourcesUtils.createClusterPvc(
                                'worker-pvc',
                                namespace
                            );
                            testWorkerPvcName = workerPvc.metadata.name;
                            createdPvcNames.push(testWorkerPvcName);
                            logger.info(`Created worker PVC: ${testWorkerPvcName}`);
                        });

                        await test.step('Submit with different worker storage', () => {
                            const submitFlags = createBasicSubmitFlags({
                                '--existing-pvc': `claimname=${testLeaderPvcName},path=/leader-data`,
                                '--worker-existing-pvc': `claimname=${testWorkerPvcName},path=/worker-data`
                            });

                            submitDistributedInferenceWorkload(
                                submitFlags,
                                workloadName,
                                createdWorkloadNames,
                                `Submitted distributed inference workload ${workloadName} with worker storage override`
                            );
                        });

                        await test.step('Wait for workload to be running for describe verification', async () => {
                            workloadId = await getWorkloadIdAndWaitForPhase(
                                workloadName,
                                project.id,
                                Phase.Running,
                                120
                            );
                        });

                        await test.step('Verify worker storage configuration via describe command', () => {
                            describeResult = describeWorkload(workloadName);

                            expect(describeResult).to.include('Leader Configuration');
                            expect(describeResult).to.include(testLeaderPvcName);
                            expect(describeResult).to.match(
                                new RegExp(`${testLeaderPvcName}\\s+/leader-data`, 'i')
                            );
                            logger.info(
                                `Verified Leader storage: PVC ${testLeaderPvcName} mounted at /leader-data`
                            );

                            expect(describeResult).to.include('Worker Configuration');
                            expect(describeResult).to.include(testWorkerPvcName);
                            expect(describeResult).to.match(
                                new RegExp(`${testWorkerPvcName}\\s+/worker-data`, 'i')
                            );
                            logger.info(
                                `Verified Worker storage: PVC ${testWorkerPvcName} mounted at /worker-data`
                            );

                            verifyRunningPods(describeResult, 3, 1, 2);
                            expect(describeResult).to.match(
                                /Requested Pods\s*:\s*Leaders:\s*1,\s*Workers:\s*2,\s*Total:\s*3/i
                            );
                            logger.info('Verified pod status: 3/3 running (1 leader + 2 workers)');

                            expect(describeResult).to.match(/Leader.*Running/i);
                            expect(describeResult).to.match(/Worker.*Running/i);
                            logger.info(
                                'Verified worker storage override: different PVCs for leader and workers with correct mount paths'
                            );
                        });
                    });

                    test('T5576464: Submit with Worker-as-Leader True and Worker Overrides', async function () {
                        await test.step('Submit workload with worker image override', () => {
                            workloadName = objectGenerator.generateShortNameByPrefix('di-worker-override');

                            const submitFlags = createBasicSubmitFlags({
                                '--worker-image': 'python:3.11'
                            });

                            submitDistributedInferenceWorkload(
                                submitFlags,
                                workloadName,
                                createdWorkloadNames,
                                `Submitted distributed inference workload ${workloadName} with leader image python:3.10 and worker image python:3.11`
                            );
                        });

                        await test.step('Verify worker image override via CLI describe', async () => {
                            const describeResult = await describeWorkloadAsync(workloadName, project.id);

                            verifyBasicWorkloadInfo(describeResult, workloadName);
                            verifyImage(describeResult, 'python:3.11');
                            logger.info(
                                'Verified worker image override (python:3.11) in CLI describe output'
                            );
                        });
                    });

                    test('T5585151: Submit with Selective Worker Configuration Override', async function () {
                        await test.step('Submit workload with selective worker overrides', () => {
                            workloadName = objectGenerator.generateShortNameByPrefix('di-selective-override');

                            const submitFlags = createBasicSubmitFlags({
                                '--command': '"sleep infinity"',
                                [SubmitFlagsMerged.envVar]: 'ROLE=LEADER',
                                [SubmitFlagsMerged.cpuCoreRequest]: '0.5',
                                '--cpu-memory-request': '100M',
                                '--worker-image': 'python:3.11-slim',
                                '--worker-environment': 'ROLE=WORKER',
                                '--worker-command': '"sleep 9999"',
                                '--worker-gpu-devices-request': '1',
                                '--worker-cpu-core-request': '0.25',
                                '--worker-cpu-memory-request': '200M'
                            });

                            submitDistributedInferenceWorkload(
                                submitFlags,
                                workloadName,
                                createdWorkloadNames,
                                `Submitted distributed inference workload ${workloadName} with selective worker overrides`
                            );
                        });

                        await test.step('Wait for workload to be running', async () => {
                            workloadId = await getWorkloadIdAndWaitForPhase(
                                workloadName,
                                project.id,
                                Phase.Initializing,
                                120
                            );
                        });

                        await test.step('Verify leader and worker configuration via API', async () => {
                            const distributedInference =
                                await workloadsCrudUtils.getDistributedInference(workloadId);
                            const leaderSpec = distributedInference.body.spec.leader;
                            const workerSpec = distributedInference.body.spec.worker;

                            logger.info('Verify leader configuration');
                            expect(leaderSpec.image).to.equal(config.quickStartImage.inferenceHelloWorld);
                            expect(leaderSpec.command).to.include('sleep infinity');
                            expect(leaderSpec.compute.gpuDevicesRequest).to.equal(1);
                            expect(leaderSpec.compute.cpuCoreRequest).to.equal(0.5);
                            expect(leaderSpec.compute.cpuMemoryRequest).to.equal('100M');
                            logger.info(
                                'Verified leader config via API: image, command, GPU (1), CPU (0.5), memory (100M)'
                            );

                            logger.info('Verify worker configuration (overrides)');
                            expect(workerSpec.image).to.equal('python:3.11-slim');
                            expect(workerSpec.command).to.include('sleep 9999');
                            expect(workerSpec.compute.gpuDevicesRequest).to.equal(1);
                            expect(workerSpec.compute.cpuCoreRequest).to.equal(0.25);
                            expect(workerSpec.compute.cpuMemoryRequest).to.equal('200M');
                            logger.info(
                                'Verified worker config via API: image (python:3.11-slim), command (sleep 9999), GPU (1), CPU (0.25), memory (200M)'
                            );
                        });

                        await test.step('Verify configuration via CLI describe', () => {
                            const describeResult = describeWorkload(workloadName);

                            // Verify basic workload info
                            verifyBasicWorkloadInfo(describeResult, workloadName);
                            verifyWorkerReplicas(describeResult, 2);

                            // Verify images (leader and worker)
                            verifyImage(describeResult, config.quickStartImage.inferenceHelloWorld);
                            verifyImage(describeResult, 'python:3.11-slim');

                            // Verify leader command and environment
                            expect(describeResult).to.match(/Command\s*:\s*sleep infinity/i);
                            verifyEnvironmentVariable(describeResult, 'ROLE', 'LEADER');

                            // Verify total compute resources (leader + workers)
                            verifyGpuCompute(describeResult, '3.0+'); // 1 leader + 2×1 workers = 3
                            verifyCpuCores(describeResult, '1.0+'); // 0.5 leader + 2×0.25 workers = 1.0
                            verifyCpuMemory(describeResult, '500M'); // 100M leader + 2×200M workers = 500M

                            logger.info(
                                'Verified CLI describe shows correct leader config and aggregated resources'
                            );
                        });
                    });
                });
            });

            test.describe('Scaling Operations', function () {
                test('T5576440: Scale Distributed Inference Replicas', async function () {
                    await test.step('Create initial workload with 1 replica', () => {
                        const submitFlags = createBasicSubmitFlags();

                        submitDistributedInferenceWorkload(
                            submitFlags,
                            workloadName,
                            createdWorkloadNames,
                            `Submitted distributed inference workload ${workloadName} with 1 replica and 2 workers`
                        );
                    });

                    await test.step('Wait for workload to be running for describe verification', async () => {
                        workloadId = await getWorkloadIdAndWaitForPhase(
                            workloadName,
                            project.id,
                            Phase.Running,
                            120
                        );

                        describeResult = describeWorkload(workloadName);
                        verifyBasicWorkloadInfo(describeResult, workloadName);
                        verifyReplicasCount(describeResult, 1);
                        verifyWorkerReplicas(describeResult, 2);
                        logger.info('Verified initial state: 1 replica, 2 workers per replica');
                    });

                    await test.step('Verify initial pod count: 3 pods (1 leader + 2 workers)', async () => {
                        const pods = await waitUtils.waitForPods(
                            workloadName,
                            project.status.namespace,
                            3,
                            120,
                            ['Running'],
                            true
                        );
                        expect(pods).to.have.lengthOf(3);

                        const leaderPods = pods.filter((pod) =>
                            pod.spec?.containers?.some((container) => container.name === 'leader')
                        );
                        const workerPods = pods.filter((pod) =>
                            pod.spec?.containers?.some((container) => container.name === 'worker')
                        );

                        expect(leaderPods).to.have.lengthOf(1);
                        expect(workerPods).to.have.lengthOf(2);
                        logger.info('Verified initial pod count: 1 leader pod, 2 worker pods');
                    });

                    await test.step('Scale to 2 replicas', () => {
                        const scaleCommand = `${runaiCliV2.cliPath} inference distributed scale ${workloadName} --replicas 2`;
                        runaiCliV2.bashCli.exec(scaleCommand);
                        logger.info('Executed scale command to 2 replicas');
                    });

                    await test.step('Verify scaling via describe', () => {
                        describeResult = describeWorkload(workloadName);
                        verifyReplicasCount(describeResult, 2);
                        logger.info('Verified via describe: replicas scaled to 2');
                    });

                    await test.step('Verify new pod count: 6 pods (2 × (1 leader + 2 workers))', async () => {
                        await waitUtils.waitForCondition(
                            () => {
                                const output = runaiCliV2.describe(
                                    {},
                                    workloadName,
                                    cliV2WorkloadType.inference,
                                    true
                                );
                                logger.debug(`CLI describe result: ${output}`);

                                const runningPodsMatch = output.match(
                                    /Running Pods\s*:\s*(\d+)\/(\d+)\s*\((\d+)\s+leaders?,\s*(\d+)\s+workers?\)/i
                                );
                                if (!runningPodsMatch) {
                                    logger.debug('Running Pods info not found in describe output');
                                    return false;
                                }
                                const runningPodCount = parseInt(runningPodsMatch[1], 10);
                                const totalPodCount = parseInt(runningPodsMatch[2], 10);
                                const leaderCount = parseInt(runningPodsMatch[3], 10);
                                const workerCount = parseInt(runningPodsMatch[4], 10);

                                logger.debug(
                                    `Current pod count: ${runningPodCount}/${totalPodCount} (${leaderCount} leaders, ${workerCount} workers)`
                                );

                                return (
                                    runningPodCount === 6 &&
                                    totalPodCount === 6 &&
                                    leaderCount === 2 &&
                                    workerCount === 4
                                );
                            },
                            'Wait for pod count to scale up to 6 (2 replicas)',
                            true,
                            120,
                            5
                        );
                        logger.info(
                            'Verified scaled pod count: 2 leader pods, 4 worker pods (2 replicas × 2 workers)'
                        );
                    });

                    await test.step('Scale down to 1 replica', () => {
                        const scaleCommand = `${runaiCliV2.cliPath} inference distributed scale ${workloadName} --replicas 1`;
                        runaiCliV2.bashCli.exec(scaleCommand);
                        logger.info('Executed scale command to 1 replica');
                    });

                    await test.step('Verify scale down via describe', async () => {
                        await waitUtils.waitForCondition(
                            () => {
                                const output = runaiCliV2.describe(
                                    {},
                                    workloadName,
                                    cliV2WorkloadType.inference,
                                    true
                                );
                                logger.info(`CLI describe result: ${output}`);

                                if (!output.match(/Replicas \(Requested\)\s*:\s*1/i)) {
                                    logger.debug('Replicas not yet scaled down to 1 in describe output');
                                    return false;
                                }

                                const runningPodsMatch = output.match(/Running Pods\s*:\s*(\d+)\/(\d+)/i);
                                if (!runningPodsMatch) {
                                    logger.debug('Running Pods info not found in describe output');
                                    return false;
                                }

                                const runningPodCount = parseInt(runningPodsMatch[1], 10);
                                const totalPodCount = parseInt(runningPodsMatch[2], 10);
                                logger.debug(
                                    `Current running pod count from describe output: ${runningPodCount}/${totalPodCount}`
                                );

                                return runningPodCount === 3 && totalPodCount === 3;
                            },
                            'Wait for pod count to scale down to 3 (1 replica)',
                            true,
                            120,
                            5
                        );
                        logger.info('Verified scale down: 1 replica, 3 pods total');
                    });
                });

                test('T5576442: Rapid Sequential Scaling', async function () {
                    await test.step('Create workload with 1 replica', () => {
                        workloadName = objectGenerator.generateShortNameByPrefix('di-rapid-scale');
                        const submitFlags = createBasicSubmitFlags({
                            '--replicas': '1'
                        });
                        submitDistributedInferenceWorkload(
                            submitFlags,
                            workloadName,
                            createdWorkloadNames,
                            `Submitted distributed inference workload ${workloadName} with 1 replica`
                        );
                    });

                    await test.step('Scale to 2 replicas', async () => {
                        await waitForWorkloadByNameAndProjectId(workloadName, project.id);
                        const scaleCommand = `${runaiCliV2.cliPath} inference distributed scale ${workloadName} --replicas 2`;
                        logger.info('Scaling to 2 replicas');
                        runaiCliV2.bashCli.exec(scaleCommand);
                        logger.info('Executed scale command to 2 replicas');
                    });

                    await test.step('Verify replica count changed to 2 via CLI describe', async () => {
                        const describeResult = await describeWorkloadAsync(workloadName, project.id);

                        verifyBasicWorkloadInfo(describeResult, workloadName);
                        verifyReplicasCount(describeResult, 2);
                        logger.info('Verified replica count changed to 2 after scaling');
                    });
                });
            });

            test.describe('List & Describe Operations', function () {
                test('T5576444: List Distributed Inference Workloads', async function () {
                    const workloadNames: string[] = [];
                    const workloadConfigs = [
                        { name: 'di-list-1', image: 'python:3.10', workers: 2 },
                        { name: 'di-list-2', image: 'python:3.11', workers: 3 },
                        { name: 'di-list-3', image: 'python:3.9', workers: 1 }
                    ];

                    await test.step('Create multiple distributed workloads', () => {
                        for (const config of workloadConfigs) {
                            const workloadName = objectGenerator.generateNameByPrefix(config.name);
                            const submitFlags = {
                                [SubmitFlagsMerged.image]: config.image,
                                [SubmitFlagsMerged.workers]: config.workers.toString(),
                                [SubmitFlagsMerged.gpuRequest]: '1',
                                [SubmitFlagsMerged.servingPort]: '80'
                            } as Partial<typeof SubmitFlagsMerged>;

                            runaiCliV2.submit(
                                submitFlags,
                                workloadName,
                                cliV2WorkloadType.inference,
                                undefined,
                                true
                            );
                            workloadNames.push(workloadName);
                            createdWorkloadNames.push(workloadName);
                            logger.info(
                                `Created distributed inference workload ${workloadName} with ${config.workers} workers`
                            );
                        }
                        logger.info(
                            `Created ${workloadNames.length} distributed inference workloads for list test`
                        );
                    });

                    await test.step('Wait for workloads to appear in project', async () => {
                        for (const name of workloadNames) {
                            await waitForWorkloadByNameAndProjectId(name, project.id);
                        }
                    });

                    await test.step('List all distributed workloads', () => {
                        const listResult = runaiCliV2.list('inference distributed');
                        const listOutput = listResult.stdout;
                        logger.info(`CLI list output: ${listOutput}`);

                        for (const workloadName of workloadNames) {
                            expect(listOutput).to.include(workloadName);
                            logger.info(`Verified workload ${workloadName} appears in list output`);
                        }
                    });

                    await test.step('Verify list shows key information', () => {
                        const listResult = runaiCliV2.list('inference distributed');
                        const listOutput = listResult.stdout;

                        for (const workloadName of workloadNames) {
                            expect(listOutput).to.include(workloadName);
                        }
                        expect(listOutput).to.include('DistributedInference');
                        const hasStatus = /Running|Creating|Initializing|Pending|Failed|Degraded/.test(
                            listOutput
                        );
                        expect(hasStatus).to.be.true;
                        const hasPodInfo = /Running\/Req\. Pods|\d+\/\d+/.test(listOutput);
                        expect(hasPodInfo).to.be.true;
                        logger.info(
                            'Verified list output contains workload names, type, status, and pod information'
                        );
                    });
                });

                test('T5576445: List with No Distributed Workloads', async function () {
                    await test.step('Ensure no distributed workloads exist in project', async () => {
                        const allWorkloads = (
                            await workloadsUtilsV2.getWorkloadsByProjectId(project.id.toString())
                        ).body.workloads;

                        const distributedInferenceWorkloads =
                            allWorkloads?.filter((w) => w.type === 'DistributedInference') || [];

                        if (distributedInferenceWorkloads.length > 0) {
                            // Bulk delete - wrap in try-catch to handle 404s (workloads may already be deleted)
                            const ids = distributedInferenceWorkloads.map((w) => w.id);
                            try {
                                const result = await workloadUtils.deleteMultipleWorkloads(ids);
                                logger.info(
                                    `Bulk delete result: ${result.succeeded.length} succeeded, ${result.failed.length} failed`
                                );
                            } catch (error: any) {
                                // 404 is expected if workloads were already deleted by another test/process
                                if (error.statusCode === 404 || error.message?.includes('404')) {
                                    logger.info(
                                        'Bulk delete returned 404 - workloads may already be deleted, continuing...'
                                    );
                                } else {
                                    throw error;
                                }
                            }

                            // Wait for all distributed inference workloads to be actually deleted
                            await waitUtils.waitForCondition(
                                async () => {
                                    const remaining =
                                        (
                                            await workloadsUtilsV2.getWorkloadsByProjectId(
                                                project.id.toString()
                                            )
                                        ).body.workloads?.filter((w) => w.type === 'DistributedInference') ||
                                        [];
                                    return remaining.length === 0;
                                },
                                'wait for all distributed inference workloads to be deleted',
                                true,
                                30,
                                2
                            );
                            logger.info(
                                `Deleted ${distributedInferenceWorkloads.length} distributed inference workload(s)`
                            );
                        }
                        logger.info('Project is clean, no distributed inference workloads exist');
                    });

                    await test.step('Execute list command and verify no workloads found', () => {
                        const listResult = runaiCliV2.list('inference distributed');
                        const listOutput = listResult.stdout;
                        logger.info(`CLI list output: ${listOutput}`);

                        const hasWorkloadEntries = listOutput.split('\n').some((line) => {
                            return (
                                line.includes('DistributedInference') &&
                                (line.includes('Creating') ||
                                    line.includes('Running') ||
                                    line.includes('Pending') ||
                                    line.includes('Failed'))
                            );
                        });

                        expect(hasWorkloadEntries).to.be.false;
                        logger.info('Verified list output indicates no distributed workloads found');
                    });
                });

                test('T5608159: List All Inference Workloads (Both Distributed and Standard)', async function () {
                    let distributedWorkloadName: string;
                    let standardWorkloadName: string;

                    await test.step('Create a distributed inference workload', () => {
                        distributedWorkloadName = objectGenerator.generateShortNameByPrefix('di-mixed-');
                        const submitFlags = {
                            [SubmitFlagsMerged.image]: 'python:3.10',
                            [SubmitFlagsMerged.workers]: '2',
                            [SubmitFlagsMerged.gpuRequest]: '1',
                            [SubmitFlagsMerged.servingPort]: '80'
                        } as Partial<typeof SubmitFlagsMerged>;

                        runaiCliV2.submit(
                            submitFlags,
                            distributedWorkloadName,
                            cliV2WorkloadType.inference,
                            undefined,
                            true
                        );
                        createdWorkloadNames.push(distributedWorkloadName);
                        logger.info(`Created distributed inference workload: ${distributedWorkloadName}`);
                    });

                    await test.step('Create a standard inference workload', () => {
                        standardWorkloadName = objectGenerator.generateShortNameByPrefix('std-inf-');
                        const submitFlags = {
                            [SubmitFlagsMerged.image]: 'python:3.10',
                            [SubmitFlagsMerged.gpuRequest]: '1',
                            [SubmitFlagsMerged.servingPort]: '80'
                        } as Partial<typeof SubmitFlagsMerged>;

                        runaiCliV2.submit(
                            submitFlags,
                            standardWorkloadName,
                            cliV2WorkloadType.inference,
                            undefined,
                            false
                        );
                        createdWorkloadNames.push(standardWorkloadName);
                        logger.info(`Created standard inference workload: ${standardWorkloadName}`);
                    });

                    await test.step('List all inference workloads without type filter', async () => {
                        await waitForWorkloadByNameAndProjectId(distributedWorkloadName, project.id);
                        await waitForWorkloadByNameAndProjectId(standardWorkloadName, project.id);
                        const listResult = runaiCliV2.list('inference');
                        const listOutput = listResult.stdout;
                        logger.info(`CLI list output: ${listOutput}`);
                        expect(listOutput).to.include(distributedWorkloadName);
                        expect(listOutput).to.include(standardWorkloadName);
                        logger.info(
                            'Verified both distributed and standard inference workloads appear in list output'
                        );
                    });

                    await test.step('Verify list shows both workload types', () => {
                        const listResult = runaiCliV2.list('inference');
                        const listOutput = listResult.stdout;
                        expect(listOutput).to.include('DistributedInference');
                        expect(listOutput).to.include('Inference');
                        logger.info(
                            'Verified list output contains both DistributedInference and Inference types'
                        );
                    });
                });

                test('T5576446: Describe Distributed Inference Workload Details', async function () {
                    await test.step('Submit comprehensive distributed workload with multiple parameters', () => {
                        workloadName = objectGenerator.generateShortNameByPrefix('di-comprehensive-test');

                        const submitFlags = createBasicSubmitFlags({
                            '--replicas': '1',
                            [SubmitFlagsMerged.workers]: '1',
                            [SubmitFlagsMerged.cpuCoreRequest]: '0.1',
                            '--cpu-core-limit': '0.5',
                            '--cpu-memory-request': '100M',
                            '--cpu-memory-limit': '200M',
                            [SubmitFlagsMerged.envVar]: 'LOG_LEVEL=INFO'
                        });

                        submitDistributedInferenceWorkload(
                            submitFlags,
                            workloadName,
                            createdWorkloadNames,
                            `Submitted comprehensive distributed inference workload ${workloadName} with 1 replica, 1 worker per replica`
                        );
                    });

                    await test.step('Verify submission via list command', async () => {
                        await waitForWorkloadByNameAndProjectId(workloadName, project.id);
                        const listResult = runaiCliV2.list('inference distributed');
                        const listOutput = listResult.stdout;
                        logger.info(`CLI list output: ${listOutput}`);

                        expect(listOutput).to.include(workloadName);
                        logger.info(`Verified workload ${workloadName} appears in list output`);
                    });

                    await test.step('Wait for workload to be running', async () => {
                        workloadId = await getWorkloadIdAndWaitForPhase(
                            workloadName,
                            project.id,
                            Phase.Running,
                            120
                        );
                    });

                    await test.step('Describe workload and verify comprehensive details', () => {
                        describeResult = describeWorkload(workloadName);

                        verifyBasicWorkloadInfo(describeResult, workloadName);
                        verifyPhase(describeResult, 'Running');
                        logger.info(
                            'Verified basic workload info: name, type (DistributedInference), and running phase'
                        );
                    });

                    await test.step('Verify distributed configuration (replicas and workers)', () => {
                        expect(describeResult).to.include('Distributed Inference Configuration');
                        verifyReplicasCount(describeResult, 1);
                        expect(describeResult).to.match(/Replicas \(Ready\)\s*:\s*1/i);
                        verifyWorkerReplicas(describeResult, 1);
                        logger.info(
                            'Verified distributed configuration: 1 replica (requested and ready), 1 worker per replica'
                        );
                    });

                    await test.step('Verify total pod count: 1 replica × (1 leader + 1 worker) = 2 pods', () => {
                        verifyRunningPods(describeResult, 2, 1, 1);
                        expect(describeResult).to.match(
                            /Requested Pods\s*:\s*Leaders:\s*1,\s*Workers:\s*1,\s*Total:\s*2/i
                        );
                        logger.info(
                            'Verified pod distribution: 2 total pods (1 leader + 1 worker) in 1 replica'
                        );
                    });

                    await test.step('Verify resource allocation in aggregated compute resources', () => {
                        expect(describeResult).to.include('Compute Resources');
                        verifyGpuCompute(describeResult, '2.0');
                        logger.info('Verified total GPU allocation: 2.0 (2 pods × 1 GPU)');
                        verifyCpuCores(describeResult, '0.2');
                        logger.info('Verified total CPU cores request: 0.2 (2 pods × 0.1 core)');
                        verifyCpuMemory(describeResult, '200M');
                        logger.info('Verified total CPU memory request: 200M (2 pods × 100M)');
                    });

                    await test.step('Verify image configuration', () => {
                        verifyImage(describeResult, config.quickStartImage.inferenceHelloWorld);
                        logger.info(`Verified image: ${config.quickStartImage.inferenceHelloWorld}`);
                    });

                    await test.step('Verify environment variables', () => {
                        verifyEnvironmentVariable(describeResult, 'LOG_LEVEL', 'INFO');
                        logger.info('Verified environment variable: LOG_LEVEL=INFO');
                    });

                    await test.step('Verify serving endpoint configuration', () => {
                        verifyServingPort(describeResult, 8080);
                        logger.info('Verified serving endpoint: port 8080 and serving URL present');
                    });

                    await test.step('Verify pod details show correct roles and running state', () => {
                        // Verify pod summary shows correct cou
                        verifyRunningPods(describeResult, 2, 1, 1);
                        expect(describeResult).to.match(
                            /Requested Pods\s*:\s*Leaders:\s*1,\s*Workers:\s*1,\s*Total:\s*2/i
                        );
                        logger.info('Verified pod summary: 2/2 running (1 leader, 1 worker)');

                        // Verify pod table shows Leader and Worker with Running status
                        const allLines = describeResult.split('\n');
                        const podTableLines = allLines.filter(
                            (line) => line.includes('Leader') || line.includes('Worker')
                        );

                        const leaderLine = podTableLines.find(
                            (line) => line.includes('Leader') && line.includes('Running')
                        );
                        const workerLine = podTableLines.find(
                            (line) => line.includes('Worker') && line.includes('Running')
                        );

                        expect(leaderLine).to.exist;
                        expect(workerLine).to.exist;
                        logger.info('Verified pod table: 1 leader pod and 1 worker pod in Running state');
                    });

                    await test.step('Verify pod distribution via API', async () => {
                        const pods = await waitUtils.waitForPods(
                            workloadName,
                            namespace,
                            2,
                            120,
                            ['Running'],
                            true
                        );
                        expect(pods).to.have.lengthOf(2);

                        const leaderPods = pods.filter((pod) => {
                            const name = pod.metadata?.name || '';
                            return name.endsWith('-0') && !name.match(/-0-\d+$/);
                        });
                        expect(leaderPods).to.have.lengthOf(1);
                        logger.info('Verified 1 leader pod (replica-0)');

                        const workerPods = pods.filter((pod) => {
                            const name = pod.metadata?.name || '';
                            return name.match(/-0-\d+$/);
                        });
                        expect(workerPods).to.have.lengthOf(1);
                        logger.info('Verified 1 worker pod (replica-0)');
                        logger.info(
                            'Verified pod distribution: 1 replica with 1 leader + 1 worker (2 pods total)'
                        );
                    });
                });

                test('T5576447: Describe Non-Existent Workload', async function () {
                    await test.step('Attempt to describe non-existent workload and verify error', () => {
                        const nonExistentWorkloadName = 'non-existent-workload-' + Date.now();
                        logger.info(
                            `Attempting to describe non-existent workload: ${nonExistentWorkloadName}`
                        );

                        try {
                            runaiCliV2.describe(
                                {},
                                nonExistentWorkloadName,
                                cliV2WorkloadType.inference,
                                true
                            );
                            expect.fail('Expected describe to fail for non-existent workload');
                        } catch (error) {
                            const errorMessage = error.message || error.toString();
                            logger.info(`CLI error message: ${errorMessage}`);

                            expect(errorMessage).to.include('failed with code 1');
                            logger.info('Verified command failed with exit code 1');
                            expect(errorMessage).to.match(/workload not found|Describe workload failed/i);
                            logger.info('Verified error message indicates workload not found');
                        }
                    });
                });
            });

            test.describe('Runtime Operations', function () {
                test('T5576448: Execute Logs Command on Distributed Workload', async function () {
                    await test.step('Create distributed workload with different leader and worker commands', () => {
                        const submitFlags = createBasicSubmitFlags({
                            '--command': '"sh -c \'while true; do echo Leader log message; sleep 5; done\'"',
                            '--worker-command':
                                '"sh -c \'while true; do echo Worker log message; sleep 5; done\'"'
                        });

                        logger.info(`Submit flags: ${JSON.stringify(submitFlags, null, 2)}`);
                        submitDistributedInferenceWorkload(
                            submitFlags,
                            workloadName,
                            createdWorkloadNames,
                            `Submitted distributed inference workload ${workloadName} with custom leader and worker commands`
                        );
                    });

                    await test.step('Wait for workload to start running', async () => {
                        workloadId = await getWorkloadIdAndWaitForPhase(
                            workloadName,
                            project.id,
                            Phase.Initializing
                        );
                    });

                    await test.step('Wait for all pods to be running', async () => {
                        const pods = await waitUtils.waitForPods(
                            workloadName,
                            project.status.namespace,
                            3,
                            120,
                            ['Running'],
                            true
                        );
                        expect(pods).to.have.lengthOf(3);
                        logger.info('All 3 pods (1 leader + 2 workers) are in Running state');
                    });

                    await test.step('Execute logs command without pod specification', () => {
                        const logsCommand = `${runaiCliV2.cliPath} inference distributed logs ${workloadName}`;
                        const logsOutput = runaiCliV2.bashCli.exec(logsCommand);
                        logger.info(`CLI logs output: ${logsOutput}`);
                        expect(logsOutput).to.not.be.empty;
                        expect(logsOutput.length).to.be.greaterThan(0);

                        const hasLeaderLogs = logsOutput.includes('Leader log message');
                        const hasWorkerLogs = logsOutput.includes('Worker log message');
                        expect(hasLeaderLogs || hasWorkerLogs).to.be.true;

                        if (hasLeaderLogs) {
                            logger.info('Logs displayed are from leader pod (contains "Leader log message")');
                        } else if (hasWorkerLogs) {
                            logger.info('Logs displayed are from worker pod (contains "Worker log message")');
                        }
                        logger.info(
                            'Verified logs command displays output from a pod (random pod as per design)'
                        );
                    });
                });

                test('T5576449: Execute Logs with Specific Pod Target', async function () {
                    await test.step('Create distributed workload with distinct leader and worker log messages', () => {
                        const submitFlags = createBasicSubmitFlags({
                            '--command': '"sh -c \'while true; do echo LEADER: $(date); sleep 3; done\'"',
                            '--worker-command':
                                '"sh -c \'while true; do echo WORKER: $(date); sleep 3; done\'"'
                        });

                        logger.info(`Submit flags: ${JSON.stringify(submitFlags, null, 2)}`);
                        submitDistributedInferenceWorkload(
                            submitFlags,
                            workloadName,
                            createdWorkloadNames,
                            `Submitted distributed inference workload ${workloadName} with distinct leader/worker log messages`
                        );
                    });

                    await test.step('Wait for workload to be running', async () => {
                        workloadId = await getWorkloadIdAndWaitForPhase(
                            workloadName,
                            project.id,
                            Phase.Initializing,
                            120
                        );
                    });

                    await test.step('Wait for all pods to be running and get pod names', async () => {
                        const pods = await waitUtils.waitForPods(
                            workloadName,
                            project.status.namespace,
                            3,
                            120,
                            ['Running'],
                            true
                        );
                        expect(pods).to.have.lengthOf(3);
                        logger.info('All 3 pods (1 leader + 2 workers) are in Running state');
                        const {
                            leaderPod,
                            workerPod,
                            leaderPodName: leaderName,
                            workerPodName: workerName
                        } = identifyLeaderAndWorkerPods(pods);
                        expect(leaderPod).to.exist;
                        expect(workerPod).to.exist;
                        leaderPodName = leaderName!;
                        workerPodName = workerName!;
                    });

                    await test.step('Target leader pod and verify LEADER logs', () => {
                        const logsCommand = `${runaiCliV2.cliPath} inference distributed logs ${workloadName} --pod ${leaderPodName}`;
                        const logsOutput = runaiCliV2.bashCli.exec(logsCommand);
                        logger.info(`Leader pod logs output: ${logsOutput}`);

                        expect(logsOutput).to.not.be.empty;
                        expect(logsOutput).to.include('LEADER:');
                        expect(logsOutput).to.not.include('WORKER:');
                        logger.info(`Verified leader pod ${leaderPodName} logs contain "LEADER:" messages`);
                    });

                    await test.step('Target worker pod and verify WORKER logs', () => {
                        const logsCommand = `${runaiCliV2.cliPath} inference distributed logs ${workloadName} --pod ${workerPodName}`;
                        const logsOutput = runaiCliV2.bashCli.exec(logsCommand);
                        logger.info(`Worker pod logs output: ${logsOutput}`);
                        expect(logsOutput).to.not.be.empty;
                        expect(logsOutput).to.include('WORKER:');
                        expect(logsOutput).to.not.include('LEADER:');
                        logger.info(`Verified worker pod ${workerPodName} logs contain "WORKER:" messages`);
                    });
                });

                test('T5576450: Execute Bash Command on Distributed Workload', async function () {
                    test.skip(
                        process.platform !== 'linux',
                        'PTY-based interactive CLI coverage runs on Linux only'
                    );

                    await test.step('Create distributed workload', () => {
                        const submitFlags = createBasicSubmitFlags({
                            [SubmitFlagsMerged.image]: 'nginx:latest',
                            [SubmitFlagsMerged.servingPort]: '80'
                        });

                        submitDistributedInferenceWorkload(
                            submitFlags,
                            workloadName,
                            createdWorkloadNames,
                            `Submitted distributed inference workload ${workloadName}`
                        );
                    });

                    await test.step('Wait for workload to be running', async () => {
                        workloadId = await getWorkloadIdAndWaitForPhase(
                            workloadName,
                            project.id,
                            Phase.Running,
                            120
                        );
                    });

                    await test.step('Get pod names from describe output', async () => {
                        const pods = await waitUtils.waitForPods(
                            workloadName,
                            project.status.namespace,
                            3,
                            120,
                            ['Running'],
                            true
                        );
                        expect(pods).to.have.lengthOf(3);
                        logger.info('All 3 pods (1 leader + 2 workers) are in Running state');
                        const { workerPod, workerPodName: workerName } = identifyLeaderAndWorkerPods(pods);
                        expect(workerPod).to.exist;
                        workerPodName = workerName!;
                    });

                    await test.step('Execute bash command to verify hostname on specific worker pod', async () => {
                        logger.info(`Executing bash command on worker pod: ${workerPodName}`);
                        const [output, childProcess] = await runaiCliV2.bashInteractive(
                            {
                                [bashFlags.pod]: workerPodName
                            },
                            workloadName,
                            cliV2WorkloadType.inference,
                            true
                        );
                        childProcess.stdin.write('hostname\n');
                        await waitUtils.waitForCondition(
                            () => output.stdout.join('').includes(workerPodName),
                            `bash output to contain worker pod hostname ${workerPodName}`,
                            true,
                            30
                        );
                        const stdoutStr = output.stdout.join('');
                        logger.info(`Bash hostname output: ${stdoutStr}`);
                        expect(stdoutStr).to.include(workerPodName);
                        logger.info(`Verified hostname matches worker pod: ${workerPodName}`);
                        childProcess.kill();
                    });
                });

                test('T5576451: Execute Exec Command on Distributed Workload', async function () {
                    await test.step('Create distributed workload with leader and worker environment variables', () => {
                        const submitFlags = createBasicSubmitFlags({
                            [SubmitFlagsMerged.envVar]: 'LEADER_ID=leader-001',
                            '--worker-environment': 'WORKER_ID=worker-001'
                        });

                        submitDistributedInferenceWorkload(
                            submitFlags,
                            workloadName,
                            createdWorkloadNames,
                            `Submitted distributed inference workload ${workloadName} with leader and worker environment variables`
                        );
                    });

                    await test.step('Wait for workload to be running', async () => {
                        workloadId = await getWorkloadIdAndWaitForPhase(
                            workloadName,
                            project.id,
                            Phase.Running,
                            120
                        );
                    });

                    await test.step('Get pod names', async () => {
                        const pods = await waitUtils.waitForPods(
                            workloadName,
                            project.status.namespace,
                            3,
                            120,
                            ['Running'],
                            true
                        );
                        expect(pods).to.have.lengthOf(3);
                        logger.info('All 3 pods (1 leader + 2 workers) are in Running state');
                        const {
                            leaderPod,
                            workerPod,
                            leaderPodName: leaderName,
                            workerPodName: workerName
                        } = identifyLeaderAndWorkerPods(pods);
                        expect(leaderPod).to.exist;
                        expect(workerPod).to.exist;
                        leaderPodName = leaderName!;
                        workerPodName = workerName!;
                    });

                    await test.step('Execute command on leader pod and verify LEADER_ID', () => {
                        const execOutput = runaiCliV2.exec(
                            {
                                [execFlags.pod]: leaderPodName,
                                [execFlags.command]: 'env'
                            },
                            workloadName,
                            cliV2WorkloadType.inference,
                            true
                        );
                        logger.info(`Leader pod env output: ${execOutput.stdout}`);
                        expect(execOutput.stdout).to.include('LEADER_ID=leader-001');
                        expect(execOutput.stdout).to.not.include('WORKER_ID=worker-001');
                        logger.info(
                            `Verified leader pod ${leaderPodName} has LEADER_ID=leader-001 and no WORKER_ID`
                        );
                    });

                    await test.step('Execute command on worker pod and verify WORKER_ID', () => {
                        const execOutput = runaiCliV2.exec(
                            {
                                [execFlags.pod]: workerPodName,
                                [execFlags.command]: 'env'
                            },
                            workloadName,
                            cliV2WorkloadType.inference,
                            true
                        );
                        logger.info(`Worker pod env output: ${execOutput.stdout}`);
                        expect(execOutput.stdout).to.include('WORKER_ID=worker-001');
                        logger.info(`Verified worker pod ${workerPodName} has WORKER_ID=worker-001`);
                    });
                });

                test('T5576452: Execute Port-Forward on Distributed Workload', async function () {
                    await test.step('Create distributed workload with web server', () => {
                        const submitFlags = createBasicSubmitFlags({
                            [SubmitFlagsMerged.image]: 'python:3.11-slim',
                            '--command': '"python -m http.server 8080"'
                        });

                        submitDistributedInferenceWorkload(
                            submitFlags,
                            workloadName,
                            createdWorkloadNames,
                            `Submitted distributed inference workload ${workloadName} with python http.server`
                        );
                    });

                    await test.step('Wait for workload to be running', async () => {
                        workloadId = await getWorkloadIdAndWaitForPhase(
                            workloadName,
                            project.id,
                            Phase.Running,
                            120
                        );
                    });

                    await test.step('Get leader pod name', async () => {
                        const pods = await waitUtils.waitForPods(
                            workloadName,
                            project.status.namespace,
                            3,
                            120,
                            ['Running'],
                            true
                        );
                        expect(pods).to.have.lengthOf(3);
                        logger.info('All 3 pods (1 leader + 2 workers) are in Running state');

                        const { leaderPod, leaderPodName: leaderName } = identifyLeaderAndWorkerPods(pods);
                        expect(leaderPod).to.exist;
                        leaderPodName = leaderName!;
                    });

                    await test.step('Start port-forward to leader pod', async () => {
                        logger.info(
                            `Starting port-forward to ${workloadName} pod ${leaderPodName} 8888:8080`
                        );
                        const [childOutput, childProc] = await runaiCliV2.portForward(
                            {
                                [portForwardFlags.pod]: leaderPodName,
                                [portForwardFlags.port]: '8888:8080'
                            },
                            workloadName,
                            cliV2WorkloadType.inference,
                            true
                        );
                        portForwardProcess = childProc;
                        if (childOutput.stdout && childOutput.stdout.length > 0) {
                            childOutput.stdout.forEach((line: string) => {
                                logger.info(`Port-forward stdout: ${line}`);
                            });
                        }
                        if (childOutput.stderr) {
                            logger.info(`Port-forward stderr: ${childOutput.stderr}`);
                        }
                    });

                    await test.step('Test connection and verify content through port-forward', () => {
                        logger.info('Testing connection via curl to localhost:8888');
                        const statusCommand =
                            'curl -s -o /dev/null -w "%{http_code}" http://localhost:8888 --max-time 10';
                        const httpCode = runaiCliV2.bashCli.exec(statusCommand);
                        logger.info(`HTTP response code: ${httpCode}`);
                        expect(httpCode.trim()).to.equal('200');

                        // Get HTML content and verify it's accessible
                        const curlCommand = 'curl -s http://localhost:8888 --max-time 10';
                        const htmlContent = runaiCliV2.bashCli.exec(curlCommand);
                        logger.info(
                            `HTML content received (first 200 chars): ${htmlContent.substring(0, 200)}`
                        );
                        expect(htmlContent).to.include('Directory listing');
                        logger.info(
                            'Verified connection successful and HTML content is accessible through port-forward'
                        );
                    });
                });
            });

            test.describe('Delete Operations', function () {
                test('T5576454: Delete Distributed Inference Workload', async function () {
                    await test.step('Create distributed inference workload', () => {
                        const submitFlags = createBasicSubmitFlags();
                        submitDistributedInferenceWorkload(
                            submitFlags,
                            workloadName,
                            createdWorkloadNames,
                            `Submitted distributed inference workload ${workloadName} for deletion test`
                        );
                    });

                    await test.step('Verify workload exists via list command', async () => {
                        await waitForWorkloadByNameAndProjectId(workloadName, project.id);
                        const listResult = runaiCliV2.list('inference distributed');
                        const listOutput = listResult.stdout;
                        logger.info(`CLI list output: ${listOutput}`);

                        expect(listOutput).to.include(workloadName);
                        logger.info(`Verified workload ${workloadName} appears in list output`);
                    });

                    await test.step('Delete workload via CLI', () => {
                        const deleteOutput = runaiCliV2.delete(
                            undefined,
                            workloadName,
                            cliV2WorkloadType.inference,
                            true,
                            true
                        );
                        logger.info(`Delete command output: ${deleteOutput}`);
                        createdWorkloadNames = createdWorkloadNames.filter((name) => name !== workloadName);
                    });

                    await test.step('Wait for workload to be deleted and verify removal', async () => {
                        await waitUtils.waitForCondition(
                            () => {
                                const listResult = runaiCliV2.list('inference distributed');
                                const listOutput = listResult.stdout;
                                return !listOutput.includes(workloadName);
                            },
                            `Wait for workload ${workloadName} to be deleted from list`,
                            true,
                            60,
                            2
                        );
                        logger.info(`Verified workload ${workloadName} is no longer in list output`);
                    });
                });

                test('T5576455: Delete Non-Existent Workload', async function () {
                    await test.step('Verify delete non-existent workload returns error', () => {
                        const nonExistentWorkloadName = 'non-existent-workload-' + Date.now();
                        logger.info(`Attempting to delete non-existent workload: ${nonExistentWorkloadName}`);

                        try {
                            runaiCliV2.delete(
                                undefined,
                                nonExistentWorkloadName,
                                cliV2WorkloadType.inference,
                                true,
                                true
                            );
                            expect.fail('Expected delete to fail for non-existent workload');
                        } catch (error) {
                            const errorMessage = error.message || error.toString();
                            expect(errorMessage).to.include('failed to search for requested workload');
                            expect(errorMessage).to.include('workload not found');
                            logger.info(
                                'Verified that deleting non-existent workload returns appropriate error'
                            );
                        }
                    });
                });
            });

            test.describe('Advanced Submit', function () {
                test('T5576457: Distributed Workload with Startup Policy', async function () {
                    await test.step('Submit workload with LeaderCreated startup policy', () => {
                        workloadName = objectGenerator.generateShortNameByPrefix('di-startup-created');
                        const submitFlags = createBasicSubmitFlags({
                            '--startup-policy': 'LeaderCreated'
                        });

                        submitDistributedInferenceWorkload(
                            submitFlags,
                            workloadName,
                            createdWorkloadNames,
                            `Submitted distributed inference workload ${workloadName} with LeaderCreated startup policy`
                        );
                    });

                    await test.step('Verify LeaderCreated startup policy via CLI describe', async () => {
                        const describeResult = await describeWorkloadAsync(workloadName, project.id);

                        verifyBasicWorkloadInfo(describeResult, workloadName);
                        expect(describeResult).to.match(/Startup Policy\s*:\s*LeaderCreated/i);
                        logger.info('Verified LeaderCreated startup policy in CLI describe output');
                    });
                });

                test('T5585150: Submit Comprehensive Distributed Workload with Multiple Parameters', async function () {
                    await test.step('Submit comprehensive distributed workload', () => {
                        const submitFlags = createBasicSubmitFlags({
                            '--replicas': '2',
                            [SubmitFlagsMerged.cpuCoreRequest]: '0.1',
                            '--cpu-core-limit': '0.5',
                            '--cpu-memory-request': '100M',
                            '--cpu-memory-limit': '200M',
                            [SubmitFlagsMerged.servingPort]:
                                'container=8080,authorization-type=public,protocol=http',
                            [SubmitFlagsMerged.envVar]: 'ROLE=LEADER'
                        });

                        submitDistributedInferenceWorkload(
                            submitFlags,
                            workloadName,
                            createdWorkloadNames,
                            `Submitted comprehensive distributed inference workload ${workloadName} with full serving port format`
                        );
                    });

                    await test.step('Wait for workload to be running', async () => {
                        workloadId = await getWorkloadIdAndWaitForPhase(
                            workloadName,
                            project.id,
                            Phase.Running,
                            120
                        );
                    });

                    await test.step('Verify basic workload information via CLI describe', () => {
                        describeResult = describeWorkload(workloadName);

                        verifyBasicWorkloadInfo(describeResult, workloadName);
                        logger.info('Verified workload name and type');
                    });

                    await test.step('Verify replicas and workers', () => {
                        expect(describeResult).to.match(/Replicas.*:\s*2/i);
                        verifyWorkerReplicas(describeResult, 2);
                        logger.info('Verified replicas (2) and workers per replica (2)');
                    });

                    await test.step('Verify image', () => {
                        verifyImage(describeResult, config.quickStartImage.inferenceHelloWorld);
                        logger.info(`Verified image: ${config.quickStartImage.inferenceHelloWorld}`);
                    });

                    await test.step('Verify environment variable', () => {
                        verifyEnvironmentVariable(describeResult, 'ROLE', 'LEADER');
                        logger.info('Verified environment variable: ROLE=LEADER');
                    });

                    await test.step('Verify serving port', () => {
                        verifyServingPort(describeResult, 8080);
                        logger.info(
                            'Verified serving port: 8080 (submitted with full format: container=8080,authorization-type=public,protocol=http)'
                        );
                    });

                    await test.step('Verify GPU allocation', () => {
                        verifyGpuCompute(describeResult, '6.0+');
                        logger.info('Verified total GPU allocation: 6.0 (2 replicas × 3 pods × 1 GPU)');
                    });

                    await test.step('Verify CPU allocation', () => {
                        verifyCpuCores(describeResult, '0.6+');
                        logger.info(
                            'Verified total CPU request: 0.6 cores (2 replicas × 3 pods × 0.1 cores)'
                        );
                    });

                    await test.step('Verify memory allocation', () => {
                        verifyCpuMemory(describeResult, '600M');
                        logger.info('Verified total memory request: 600M (2 replicas × 3 pods × 100M)');
                    });
                    logger.info('All configurations verified successfully via CLI describe');
                });
            });

            test.describe('Restart Policy', function () {
                test('T5576458: Distributed Workload with RecreateGroupOnPodRestart Policy', async function () {
                    await test.step('Create workload with RecreateGroupOnPodRestart policy', () => {
                        workloadName = objectGenerator.generateShortNameByPrefix('di-restart-group');
                        const submitFlags = createBasicSubmitFlags({
                            '--restart-policy': 'RecreateGroupOnPodRestart'
                        });
                        submitDistributedInferenceWorkload(
                            submitFlags,
                            workloadName,
                            createdWorkloadNames,
                            `Submitted distributed inference workload ${workloadName} with RecreateGroupOnPodRestart policy`
                        );
                    });

                    await test.step('Verify restart policy via CLI describe', async () => {
                        const describeResult = await describeWorkloadAsync(workloadName, project.id);

                        verifyBasicWorkloadInfo(describeResult, workloadName);
                        expect(describeResult).to.match(/Restart Policy\s*:\s*RecreateGroupOnPodRestart/i);
                        logger.info(
                            'Verified RecreateGroupOnPodRestart restart policy in CLI describe output'
                        );
                    });
                });

                test('T5582967: Distributed Workload with None Restart Policy', async function () {
                    await test.step('Create workload with None restart policy', () => {
                        workloadName = objectGenerator.generateShortNameByPrefix('di-restart-none');
                        const submitFlags = createBasicSubmitFlags({
                            '--restart-policy': 'None'
                        });

                        submitDistributedInferenceWorkload(
                            submitFlags,
                            workloadName,
                            createdWorkloadNames,
                            `Submitted distributed inference workload ${workloadName} with None restart policy`
                        );
                    });

                    await test.step('Verify restart policy via CLI describe', async () => {
                        const describeResult = await describeWorkloadAsync(workloadName, project.id);

                        verifyBasicWorkloadInfo(describeResult, workloadName);
                        expect(describeResult).to.match(/Restart Policy\s*:\s*None/i);
                        logger.info('Verified None restart policy in CLI describe output');
                    });
                });
            });

            test.describe('Negative Tests', function () {
                test('T5576459: Submit Without Required Workers Parameter', async function () {
                    await test.step('Submit workload without --workers parameter', () => {
                        const submitFlags = {
                            [SubmitFlagsMerged.image]: config.quickStartImage.inferenceHelloWorld,
                            [SubmitFlagsMerged.gpuRequest]: '1',
                            [SubmitFlagsMerged.servingPort]: '8080'
                        } as Partial<typeof SubmitFlagsMerged>;

                        logger.info('Submitting distributed inference workload without --workers parameter');
                        submitDistributedInferenceWorkload(
                            submitFlags,
                            workloadName,
                            createdWorkloadNames,
                            `Successfully submitted workload ${workloadName} without --workers parameter`
                        );
                    });

                    await test.step('Verify workers count is 0 via CLI describe', async () => {
                        const describeResult = await describeWorkloadAsync(workloadName, project.id);

                        verifyBasicWorkloadInfo(describeResult, workloadName);
                        expect(describeResult).to.include('Workers per Replica');
                        verifyWorkerReplicas(describeResult, 0);
                        logger.info('Verified that Workers per Replica is 0 in CLI describe output');
                    });
                });

                test('T5576460: Submit with Invalid Workers Value', async function () {
                    await test.step('Submit with negative workers value (-1)', () => {
                        const submitFlags = createBasicSubmitFlags({
                            [SubmitFlagsMerged.workers]: '-1'
                        });
                        logger.info('Submitting with negative workers value (-1)');

                        try {
                            runaiCliV2.submit(
                                submitFlags,
                                negativeWorkersWorkloadName,
                                cliV2WorkloadType.inference,
                                undefined,
                                true
                            );
                            createdWorkloadNames.push(negativeWorkersWorkloadName);
                            expect.fail(
                                'CLI accepted negative workers value (-1), but should have rejected it (workers must be >= 0)'
                            );
                        } catch (error) {
                            if (error.message && error.message.includes('CLI accepted negative workers')) {
                                throw error;
                            }
                            logger.info('CLI correctly rejected negative workers value');
                        }
                    });

                    await test.step('Verify non-numeric workers value is rejected', () => {
                        const workloadName = objectGenerator.generateShortNameByPrefix('di-invalid-workers');

                        const submitFlags = createBasicSubmitFlags({
                            [SubmitFlagsMerged.workers]: 'abc'
                        });
                        logger.info('Attempting to submit with non-numeric workers value (abc)');

                        try {
                            runaiCliV2.submit(
                                submitFlags,
                                workloadName,
                                cliV2WorkloadType.inference,
                                undefined,
                                true
                            );
                            expect.fail('Expected submit to fail with non-numeric workers value');
                        } catch (error) {
                            const errorMessage = error.message || error.toString();
                            expect(errorMessage).to.include('invalid argument');
                            expect(errorMessage).to.include('--workers');
                            logger.info('Verified non-numeric workers value is rejected');
                        }
                    });
                });

                test('T5576461: Submit with Missing Required Image', async function () {
                    await test.step('Verify submit without image is rejected', () => {
                        const submitFlags = {
                            [SubmitFlagsMerged.workers]: '2',
                            [SubmitFlagsMerged.gpuRequest]: '1',
                            [SubmitFlagsMerged.servingPort]: '8080'
                        } as Partial<typeof SubmitFlagsMerged>;

                        logger.info('Attempting to submit without image parameter');

                        try {
                            runaiCliV2.submit(
                                submitFlags,
                                workloadName,
                                cliV2WorkloadType.inference,
                                undefined,
                                true
                            );
                            expect.fail('Expected submit to fail without image parameter');
                        } catch (error) {
                            const errorMessage = error.message || error.toString();
                            expect(errorMessage).to.include('failed to submit');
                            expect(errorMessage).to.match(/'image' must be provided/);
                            logger.info('Verified that image parameter is required');
                        }
                    });
                });

                test('T5576462: Submit with Invalid Serving Port', async function () {
                    await test.step('Verify invalid port number (70000) is rejected', () => {
                        const submitFlags = createBasicSubmitFlags({
                            [SubmitFlagsMerged.servingPort]: '70000'
                        });
                        logger.info('Attempting to submit with invalid port number (70000)');

                        try {
                            runaiCliV2.submit(
                                submitFlags,
                                workloadName,
                                cliV2WorkloadType.inference,
                                undefined,
                                true
                            );
                            expect.fail('Expected submit to fail with invalid port number');
                        } catch (error) {
                            const errorMessage = error.message || error.toString();
                            expect(errorMessage).to.include('failed to submit');
                            expect(errorMessage).to.include('must be between 1 and 65535, got 70000');
                            expect(errorMessage).to.include('invalid serving port range');
                            logger.info('Verified that invalid port number (70000) is rejected');
                        }
                    });

                    await test.step('Verify negative port number (-1) is rejected', () => {
                        const submitFlags = createBasicSubmitFlags({
                            [SubmitFlagsMerged.servingPort]: '-1'
                        });

                        logger.info('Attempting to submit with negative port number (-1)');

                        try {
                            runaiCliV2.submit(
                                submitFlags,
                                workloadName,
                                cliV2WorkloadType.inference,
                                undefined,
                                true
                            );
                            expect.fail('Expected submit to fail with negative port number');
                        } catch (error) {
                            const errorMessage = error.message || error.toString();
                            expect(errorMessage).to.include('failed to submit');
                            expect(errorMessage).to.include('must be between 1 and 65535, got -1');
                            expect(errorMessage).to.include('invalid serving port range');
                            logger.info('Verified that negative port number (-1) is rejected');
                        }
                    });
                });

                test('T5576463: Scale with Missing Replicas Parameter', async function () {
                    await test.step('Create distributed inference workload', () => {
                        const submitFlags = createBasicSubmitFlags();

                        submitDistributedInferenceWorkload(
                            submitFlags,
                            workloadName,
                            createdWorkloadNames,
                            `Submitted distributed inference workload ${workloadName}`
                        );
                    });

                    await test.step('Verify scale without --replicas parameter is rejected', async () => {
                        await waitForWorkloadByNameAndProjectId(workloadName, project.id);
                        const scaleCommand = `${runaiCliV2.cliPath} inference distributed scale ${workloadName}`;
                        logger.info('Attempting to scale without --replicas parameter');

                        try {
                            runaiCliV2.bashCli.exec(scaleCommand);
                            expect.fail('Expected scale to fail without --replicas parameter');
                        } catch (error) {
                            const errorMessage = error.message || error.toString();
                            expect(errorMessage).to.include('failed to scale');
                            expect(errorMessage).to.include('--replicas flag is required');
                            logger.info('Verified that --replicas parameter is required for scale command');
                        }
                    });
                });
            });
        });
    }
);
