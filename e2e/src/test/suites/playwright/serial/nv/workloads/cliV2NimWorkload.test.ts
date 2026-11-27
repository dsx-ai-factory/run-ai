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
    workloadsUtilsV2
} from '../../../../../../modules/helpers';
import { Project } from '../../../../../../modules/api/apis/orgUnitApi';
import { Phase } from '../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import { isNimOperatorInstalled } from '../../../../../../modules/helpers/utils/IsInstalledUtils';
import { kubeCli, runaiCliV2 } from '../../../../../../modules/cli';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test } from '@playwright/test';
import {
    cliV2WorkloadType,
    execFlags,
    NimSubmitFlagsMerged,
    portForwardFlags
} from '../../../../../../modules/cli/tools/RunaiCliV2.constants';
import {
    createBasicNimSubmitFlagsWithPvc,
    createNimWithExistingPvcFlags,
    createNimWithAutoscalingFlags,
    createNimWithWorkersFlags,
    createNimWithFixedReplicasFlags,
    submitNimWorkload,
    describeNimWorkload,
    describeNimWorkloadJson,
    scaleNimWorkload,
    updateNimWorkload,
    deleteNimWorkload,
    getWorkloadIdAndWaitForPhase,
    verifyNimBasicInfo,
    verifyNimImage,
    verifyNimServingPort,
    verifyNimPhase,
    verifyNimReplicas,
    verifyNimAutoscaling,
    verifyNimMultiNode,
    verifyNimModelStore,
    verifyNimRunningPods,
    NGC_AUTH_SECRET_NAME
} from '../../../../../../modules/cli/helpers/NimCliTestHelpers';

test.describe(
    `NIM CLI tests ${tags.teams.viper} ${tags.resource.mimService} ${tags.tool.runaiCli}`,
    { tag: [tags.resource.mimService, tags.tool.runaiCli] },
    function () {
        supportedFrom(SupportedVersions.V2_25)(function () {
            const PHASE_WAIT_SEC = 180;
            const MULTI_NODE_PHASE_WAIT_SEC = 300;
            const TEST_TIMEOUT_MS = 300000;
            const INTERACTIVE_TEST_TIMEOUT_MS = 360000;
            let project: Project;
            let workloadId: string;
            let workloadName: string;
            let projectName: string;
            let namespace: string;
            const createdWorkloadNames: string[] = [];
            let describeResult: string;
            const createdPvcNames: string[] = [];
            let portForwardProcess: any;

            async function createNgcAuthSecret(ns: string): Promise<void> {
                const secretName = NGC_AUTH_SECRET_NAME;
                const secretData = {
                    NGC_API_KEY: 'a2V5Cg=='
                };

                try {
                    await kubeCli.createSecret(secretName, ns, secretData);
                    logger.info(`Created NGC auth secret ${secretName} in namespace ${ns}`);
                } catch (error) {
                    if ((error as any)?.body?.reason === 'AlreadyExists') {
                        logger.info(`NGC auth secret ${secretName} already exists in namespace ${ns}`);
                    } else {
                        throw error;
                    }
                }
            }

            test.beforeAll(async function () {
                if (!(await isNimOperatorInstalled())) {
                    logger.warn('NIM operator is not installed on the cluster, skipping test');
                    test.skip();
                }

                project = await orgUnitUtils.createProjectAndValidate({
                    name: objectGenerator.generateShortNameByPrefix('project-nim-cli'),
                    clusterId: flowRootState.cluster.uuid,
                    resources: [{ gpu: { deserved: 4, limit: 4 } }]
                });
                projectName = project.name;
                namespace = project.status.namespace;

                runaiCliV2.loginAsUser(
                    flowRootState.users.system_administrator.email,
                    flowRootState.users.system_administrator.password
                );
                runaiCliV2.projectSet(project.name);
                workloadsCrudUtils.accessToken = flowRootState.users.system_administrator.accessToken;

                await createNgcAuthSecret(namespace);
            });

            test.afterAll(async function () {
                if (project) {
                    try {
                        await workloadsUtilsV2.deleteWorkloadsByProjectId(project.id, 45);
                        await orgUnitUtils.deleteProjectAndNamespace(project.id, project.name);
                    } catch (error) {
                        logger.error(
                            `Failed to delete workloads and project ${project.name} namespace ${project.status.namespace} with error ${error}`
                        );
                    }
                }
            });

            test.beforeEach(function () {
                workloadName = objectGenerator.generateShortNameByPrefix('nim-cli');
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
                    logger.info(`Cleaning up ${createdWorkloadNames.length} NIM workload(s)`);
                    for (const workload of createdWorkloadNames) {
                        try {
                            const wlId = await workloadsUtilsV2.getWorkloadIdByName(
                                workload,
                                project.id
                            );
                            await workloadsCrudUtils.deleteInference(wlId);
                            logger.info(`Deleted NIM workload: ${workload}`);
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

            async function createPvcForTest(): Promise<string> {
                const pvcName = objectGenerator.generateShortNameByPrefix('nim-pvc');
                await kubeCli.CoreV1Api.createNamespacedPersistentVolumeClaim({
                    namespace,
                    body: {
                        apiVersion: 'v1',
                        kind: 'PersistentVolumeClaim',
                        metadata: { name: pvcName },
                        spec: {
                            accessModes: ['ReadWriteOnce'],
                            resources: { requests: { storage: '100Mi' } }
                        }
                    }
                });
                createdPvcNames.push(pvcName);
                logger.info(`Created PVC ${pvcName} for test`);
                return pvcName;
            }

            test.describe('Help & Documentation', function () {
                test('Verify NIM Help Command', async function () {
                    await test.step('Execute help command for inference nim', () => {
                        const helpCommand = `${runaiCliV2.cliPath} inference nim --help`;
                        const helpOutput = runaiCliV2.bashCli.exec(helpCommand);
                        logger.info(`NIM help output: ${helpOutput}`);

                        expect(helpOutput).to.match(/NIM|NVIDIA/i);
                        logger.info('Verified NIM command description');

                        const expectedCommands = ['delete', 'describe', 'list', 'scale', 'submit', 'update'];
                        expectedCommands.forEach((command) => {
                            expect(helpOutput).to.include(command);
                        });
                        logger.info('Verified all expected commands are listed');
                    });

                    await test.step('Execute help command for inference nim submit', () => {
                        const submitHelpCommand = `${runaiCliV2.cliPath} inference nim submit --help`;
                        const submitHelpOutput = runaiCliV2.bashCli.exec(submitHelpCommand);
                        logger.info(`NIM submit help output: ${submitHelpOutput}`);

                        const expectedFlags = [
                            '--model-existing-pvc',
                            '--model-new-pvc',
                            '--serving-port',
                            '--ngc-auth-secret',
                            '--min-replicas',
                            '--max-replicas',
                            '--workers'
                        ];
                        expectedFlags.forEach((flag) => {
                            expect(submitHelpOutput).to.include(flag);
                        });
                        logger.info('Verified all expected submit flags are listed');
                    });
                });
            });

            test.describe('Submit Operations', function () {
                test('Submit NIM with PVC', async function () {
                    test.setTimeout(TEST_TIMEOUT_MS);

                    const defaultPvcName = await createPvcForTest();

                    await test.step('Submit NIM workload with model store PVC', () => {
                        const submitFlags = createBasicNimSubmitFlagsWithPvc(defaultPvcName);
                        submitNimWorkload(submitFlags, workloadName, createdWorkloadNames);
                        logger.info(`Submitted NIM workload ${workloadName} with PVC ${defaultPvcName}`);
                    });

                    await test.step('Wait for workload to reach Running phase', async () => {
                        workloadId = await getWorkloadIdAndWaitForPhase(
                            workloadName,
                            project.id,
                            Phase.Running,
                            PHASE_WAIT_SEC
                        );
                        logger.info(`NIM workload ${workloadName} is running with ID ${workloadId}`);
                    });

                    await test.step('Verify describe output', () => {
                        describeResult = describeNimWorkload(workloadName, project.name);
                        verifyNimBasicInfo(describeResult, workloadName);
                        verifyNimImage(describeResult, config.quickStartImage.nim);
                        verifyNimServingPort(describeResult, 8000);
                        verifyNimPhase(describeResult, 'Running');
                        verifyNimModelStore(describeResult);
                        logger.info('Verified NIM workload describe output');
                    });
                });

                test('Submit NIM with existing PVC', async function () {
                    test.setTimeout(TEST_TIMEOUT_MS);

                    const pvcName = await createPvcForTest();

                    await test.step('Wait for PVC to be available', async () => {
                        await waitUtils.waitForCondition(
                            async () => {
                                const pvc = await kubeCli.CoreV1Api.readNamespacedPersistentVolumeClaim({
                                    name: pvcName,
                                    namespace
                                });
                                return pvc.status?.phase === 'Bound' || pvc.status?.phase === 'Pending';
                            },
                            'Wait for PVC to be available',
                            true,
                            60,
                            2
                        );
                    });

                    await test.step('Submit NIM workload with existing PVC', () => {
                        const submitFlags = createNimWithExistingPvcFlags(pvcName);
                        submitNimWorkload(submitFlags, workloadName, createdWorkloadNames);
                        logger.info(`Submitted NIM workload ${workloadName} with existing PVC ${pvcName}`);
                    });

                    await test.step('Wait for workload to reach Running phase', async () => {
                        workloadId = await getWorkloadIdAndWaitForPhase(
                            workloadName,
                            project.id,
                            Phase.Running,
                            PHASE_WAIT_SEC
                        );
                        logger.info(`NIM workload ${workloadName} is running with ID ${workloadId}`);
                    });

                    await test.step('Verify describe output shows existing PVC', () => {
                        describeResult = describeNimWorkload(workloadName, project.name);
                        verifyNimBasicInfo(describeResult, workloadName);
                        expect(describeResult).to.include(pvcName);
                        logger.info('Verified NIM workload uses existing PVC');
                    });
                });

                test('Submit NIM with autoscaling', async function () {
                    test.setTimeout(TEST_TIMEOUT_MS);

                    const defaultPvcName = await createPvcForTest();

                    await test.step('Submit NIM workload with autoscaling configuration', () => {
                        const submitFlags = createNimWithAutoscalingFlags(
                            defaultPvcName,
                            1,
                            3,
                            'num_requests_running',
                            10
                        );
                        submitNimWorkload(submitFlags, workloadName, createdWorkloadNames);
                        logger.info(`Submitted NIM workload ${workloadName} with autoscaling`);
                    });

                    await test.step('Wait for workload to reach Running phase', async () => {
                        workloadId = await getWorkloadIdAndWaitForPhase(
                            workloadName,
                            project.id,
                            Phase.Running,
                            PHASE_WAIT_SEC
                        );
                        logger.info(`NIM workload ${workloadName} is running with ID ${workloadId}`);
                    });

                    await test.step('Verify HPA is created', async () => {
                        await waitUtils.waitForCondition(
                            async () => {
                                try {
                                    const hpa = await kubeCli.getHPA(workloadName, namespace);
                                    return hpa !== null;
                                } catch {
                                    return false;
                                }
                            },
                            'Wait for HPA to be created',
                            true,
                            60,
                            5
                        );
                        logger.info('Verified HPA was created for autoscaling NIM workload');
                    });

                    await test.step('Verify describe output shows autoscaling', () => {
                        describeResult = describeNimWorkload(workloadName, project.name);
                        verifyNimBasicInfo(describeResult, workloadName);
                        verifyNimAutoscaling(describeResult, 1, 3, 'num_requests_running');
                        logger.info('Verified NIM workload describe shows autoscaling configuration');
                    });
                });
            });

            test.describe('Multi-Node Submit', function () {
                test('Submit NIM with workers (multi-node)', async function () {
                    test.setTimeout(TEST_TIMEOUT_MS);

                    const defaultPvcName = await createPvcForTest();

                    await test.step('Submit NIM workload with 1 worker (1 leader + 1 worker)', () => {
                        const submitFlags = createNimWithWorkersFlags(defaultPvcName, 1, {
                            [NimSubmitFlagsMerged.gpuRequest]: '1'
                        });
                        submitNimWorkload(submitFlags, workloadName, createdWorkloadNames);
                        logger.info(`Submitted NIM workload ${workloadName} with 1 worker`);
                    });

                    await test.step('Wait for workload to reach Running phase', async () => {
                        workloadId = await getWorkloadIdAndWaitForPhase(
                            workloadName,
                            project.id,
                            Phase.Running,
                            MULTI_NODE_PHASE_WAIT_SEC
                        );
                        logger.info(`NIM workload ${workloadName} is running with ID ${workloadId}`);
                    });

                    await test.step('Verify describe output shows multi-node configuration', () => {
                        describeResult = describeNimWorkload(workloadName, project.name);
                        verifyNimBasicInfo(describeResult, workloadName);
                        verifyNimMultiNode(describeResult, 1);
                        logger.info('Verified NIM workload describe shows multi-node configuration (1 worker)');
                    });

                    await test.step('Verify leader and worker pods created', async () => {
                        const pods = await waitUtils.waitForPods(
                            workloadName,
                            namespace,
                            2,
                            PHASE_WAIT_SEC,
                            ['Running'],
                            true
                        );
                        expect(pods.length).to.be.at.least(2);
                        logger.info(`Verified ${pods.length} pods are running for multi-node NIM workload (1 leader + 1 worker)`);
                    });
                });
            });

            test.describe('Scale Operations', function () {
                test('Scale fixed replicas NIM workload', async function () {
                    test.setTimeout(TEST_TIMEOUT_MS);

                    const defaultPvcName = await createPvcForTest();

                    await test.step('Submit NIM workload with fixed replicas', () => {
                        const submitFlags = createNimWithFixedReplicasFlags(defaultPvcName, 1);
                        submitNimWorkload(submitFlags, workloadName, createdWorkloadNames);
                        logger.info(`Submitted NIM workload ${workloadName} with 1 replica`);
                    });

                    await test.step('Wait for workload to reach Running phase', async () => {
                        workloadId = await getWorkloadIdAndWaitForPhase(
                            workloadName,
                            project.id,
                            Phase.Running,
                            PHASE_WAIT_SEC
                        );
                        logger.info(`NIM workload ${workloadName} is running with ID ${workloadId}`);
                    });

                    await test.step('Scale workload to 2 replicas', () => {
                        scaleNimWorkload(workloadName, 2, false, projectName);
                        logger.info(`Scaled NIM workload ${workloadName} to 2 replicas`);
                    });

                    await test.step('Verify pod count increases', async () => {
                        const pods = await waitUtils.waitForPods(
                            workloadName,
                            namespace,
                            2,
                            PHASE_WAIT_SEC,
                            ['Running'],
                            true
                        );
                        expect(pods.length).to.be.at.least(2);
                        logger.info('Verified NIM workload scaled to 2 replicas');
                    });

                    await test.step('Verify describe shows updated replicas via JSON', () => {
                        const parsed = describeNimWorkloadJson(workloadName);
                        expect(parsed.nim.replicasReady).to.be.at.least(2);
                        logger.info('Verified NIM workload describe shows 2 replicas');
                    });
                });

                test('Scale autoscaling NIM workload with force flag', async function () {
                    test.setTimeout(TEST_TIMEOUT_MS);

                    const defaultPvcName = await createPvcForTest();

                    await test.step('Submit NIM workload with autoscaling', () => {
                        const submitFlags = createNimWithAutoscalingFlags(
                            defaultPvcName,
                            1,
                            3,
                            'num_requests_running',
                            10
                        );
                        submitNimWorkload(submitFlags, workloadName, createdWorkloadNames);
                        logger.info(`Submitted NIM workload ${workloadName} with autoscaling`);
                    });

                    await test.step('Wait for workload to reach Running phase', async () => {
                        workloadId = await getWorkloadIdAndWaitForPhase(
                            workloadName,
                            project.id,
                            Phase.Running,
                            PHASE_WAIT_SEC
                        );
                        logger.info(`NIM workload ${workloadName} is running with ID ${workloadId}`);
                    });

                    await test.step('Scale workload with --force flag to 1 replica', () => {
                        scaleNimWorkload(workloadName, 1, true, projectName);
                        logger.info(`Scaled autoscaling NIM workload ${workloadName} to 1 fixed replica`);
                    });

                    await test.step('Verify autoscaling is disabled', async () => {
                        await waitUtils.waitForCondition(
                            () => {
                                const parsed = describeNimWorkloadJson(workloadName);
                                return parsed.nim.autoscalingEnabled === false;
                            },
                            'autoscaling to be disabled after scale with force',
                            true,
                            60,
                            5
                        );
                        logger.info('Verified autoscaling is disabled after scale with force');
                    });
                });
            });

            test.describe('Update Operations', function () {
                test('Update NIM compute resources', async function () {
                    test.setTimeout(TEST_TIMEOUT_MS);

                    const defaultPvcName = await createPvcForTest();

                    await test.step('Submit NIM workload', () => {
                        const submitFlags = createBasicNimSubmitFlagsWithPvc(defaultPvcName);
                        submitNimWorkload(submitFlags, workloadName, createdWorkloadNames);
                        logger.info(`Submitted NIM workload ${workloadName}`);
                    });

                    await test.step('Wait for workload to reach Running phase', async () => {
                        workloadId = await getWorkloadIdAndWaitForPhase(
                            workloadName,
                            project.id,
                            Phase.Running,
                            PHASE_WAIT_SEC
                        );
                        logger.info(`NIM workload ${workloadName} is running`);
                    });

                    await test.step('Update compute resources', () => {
                        updateNimWorkload(workloadName, {
                            [NimSubmitFlagsMerged.cpuCoreRequest]: '0.2',
                            [NimSubmitFlagsMerged.cpuMemoryRequest]: '200M'
                        });
                        logger.info('Updated NIM workload compute resources');
                    });

                    await test.step('Wait for workload to be ready after update', async () => {
                        await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, PHASE_WAIT_SEC);
                    });

                    await test.step('Verify updated resources in describe', async () => {
                        await waitUtils.waitForCondition(
                            () => {
                                describeResult = describeNimWorkload(workloadName, project.name);
                                return /0\.2|200m/i.test(describeResult);
                            },
                            'describe output to show updated CPU/memory (0.2 or 200m)',
                            true,
                            60,
                            5
                        );
                        logger.info('Verified NIM workload compute resources were updated');
                    });
                });

                test('Update NIM environment variables', async function () {
                    test.setTimeout(TEST_TIMEOUT_MS);

                    const defaultPvcName = await createPvcForTest();

                    await test.step('Submit NIM workload', () => {
                        const submitFlags = createBasicNimSubmitFlagsWithPvc(defaultPvcName);
                        submitNimWorkload(submitFlags, workloadName, createdWorkloadNames);
                        logger.info(`Submitted NIM workload ${workloadName}`);
                    });

                    await test.step('Wait for workload to reach Running phase', async () => {
                        workloadId = await getWorkloadIdAndWaitForPhase(
                            workloadName,
                            project.id,
                            Phase.Running,
                            PHASE_WAIT_SEC
                        );
                        logger.info(`NIM workload ${workloadName} is running`);
                    });

                    await test.step('Update environment variables', () => {
                        updateNimWorkload(workloadName, {
                            [NimSubmitFlagsMerged.envVar]: 'TEST_VAR=test_value'
                        });
                        logger.info('Updated NIM workload environment variables');
                    });

                    await test.step('Wait for workload to be ready after update', async () => {
                        await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, PHASE_WAIT_SEC);
                    });

                    await test.step('Verify workload is still healthy after env var update', () => {
                        const parsed = describeNimWorkloadJson(workloadName);
                        expect(parsed.general).to.exist;
                        expect(parsed.general.phase).to.equal('Running');
                        logger.info('Verified NIM workload is healthy after environment variable update');
                    });
                });

                test('Transition from fixed replicas to autoscaling', async function () {
                    test.setTimeout(TEST_TIMEOUT_MS);

                    const defaultPvcName = await createPvcForTest();

                    await test.step('Submit NIM workload with fixed replicas', () => {
                        const submitFlags = createNimWithFixedReplicasFlags(defaultPvcName, 1);
                        submitNimWorkload(submitFlags, workloadName, createdWorkloadNames);
                        logger.info(`Submitted NIM workload ${workloadName} with fixed replicas`);
                    });

                    await test.step('Wait for workload to reach Running phase', async () => {
                        workloadId = await getWorkloadIdAndWaitForPhase(
                            workloadName,
                            project.id,
                            Phase.Running,
                            PHASE_WAIT_SEC
                        );
                        logger.info(`NIM workload ${workloadName} is running`);
                    });

                    await test.step('Update to autoscaling configuration', () => {
                        updateNimWorkload(workloadName, {
                            [NimSubmitFlagsMerged.minReplicas]: '1',
                            [NimSubmitFlagsMerged.maxReplicas]: '2',
                            [NimSubmitFlagsMerged.metric]: 'concurrency',
                            [NimSubmitFlagsMerged.metricThreshold]: '10'
                        });
                        logger.info('Updated NIM workload to autoscaling configuration');
                    });

                    await test.step('Verify HPA is created', async () => {
                        await waitUtils.waitForCondition(
                            async () => {
                                try {
                                    const hpa = await kubeCli.getHPA(workloadName, namespace);
                                    return hpa !== null;
                                } catch {
                                    return false;
                                }
                            },
                            'Wait for HPA to be created',
                            true,
                            60,
                            5
                        );
                        logger.info('Verified HPA was created after transitioning to autoscaling');
                    });

                    await test.step('Verify autoscaling in describe', () => {
                        describeResult = describeNimWorkload(workloadName, project.name);
                        verifyNimAutoscaling(describeResult, 1, 2, 'concurrency');
                        logger.info('Verified NIM workload transitioned to autoscaling');
                    });
                });
            });

            test.describe('List & Describe', function () {
                test('List shows NIM workloads', async function () {
                    test.setTimeout(TEST_TIMEOUT_MS);

                    const defaultPvcName = await createPvcForTest();

                    await test.step('Submit NIM workload', () => {
                        const submitFlags = createBasicNimSubmitFlagsWithPvc(defaultPvcName);
                        submitNimWorkload(submitFlags, workloadName, createdWorkloadNames);
                        logger.info(`Submitted NIM workload ${workloadName}`);
                    });

                    await test.step('Wait for workload to appear', async () => {
                        workloadId = await getWorkloadIdAndWaitForPhase(
                            workloadName,
                            project.id,
                            Phase.Running,
                            PHASE_WAIT_SEC
                        );
                        logger.info(`NIM workload ${workloadName} is running`);
                    });

                    await test.step('Verify inference list includes NIM workload', () => {
                        const nimListCommand = `${runaiCliV2.cliPath} inference list -p ${projectName}`;
                        const nimListOutput = runaiCliV2.bashCli.exec(nimListCommand);
                        expect(nimListOutput).to.include(workloadName);
                        logger.info('Verified NIM workload appears in inference list');
                    });

                    await test.step('Verify inference nim list shows workload', () => {
                        const nimListCommand = `${runaiCliV2.cliPath} inference nim list -p ${projectName}`;
                        const nimListOutput = runaiCliV2.bashCli.exec(nimListCommand);
                        expect(nimListOutput).to.include(workloadName);
                        logger.info('Verified NIM workload appears in inference nim list');
                    });
                });

                test('Describe shows NIM-specific fields', async function () {
                    test.setTimeout(TEST_TIMEOUT_MS);

                    const defaultPvcName = await createPvcForTest();

                    await test.step('Submit NIM workload', () => {
                        const submitFlags = createBasicNimSubmitFlagsWithPvc(defaultPvcName);
                        submitNimWorkload(submitFlags, workloadName, createdWorkloadNames);
                        logger.info(`Submitted NIM workload ${workloadName}`);
                    });

                    await test.step('Wait for workload to reach Running phase', async () => {
                        workloadId = await getWorkloadIdAndWaitForPhase(
                            workloadName,
                            project.id,
                            Phase.Running,
                            PHASE_WAIT_SEC
                        );
                        logger.info(`NIM workload ${workloadName} is running`);
                    });

                    await test.step('Verify describe output contains NIM-specific fields', () => {
                        describeResult = describeNimWorkload(workloadName, project.name);
                        verifyNimBasicInfo(describeResult, workloadName);
                        verifyNimImage(describeResult, config.quickStartImage.nim);
                        verifyNimServingPort(describeResult, 8000);
                        verifyNimModelStore(describeResult);
                        expect(describeResult).to.match(/Serving|URL/i);
                        logger.info('Verified NIM describe output contains all expected fields');
                    });

                    await test.step('Verify JSON output format', () => {
                        const parsed = describeNimWorkloadJson(workloadName);
                        expect(parsed).to.have.property('general');
                        expect(parsed).to.have.property('nim');
                        expect(parsed.general.name).to.equal(workloadName);
                        expect(parsed.general.type).to.equal('NIMService');
                        logger.info('Verified NIM describe JSON output format');
                    });
                });

                test('Describe Non-Existent NIM Workload', async function () {
                    await test.step('Attempt to describe non-existent workload and verify error', () => {
                        const nonExistentWorkloadName = 'non-existent-nim-' + Date.now();
                        logger.info(
                            `Attempting to describe non-existent workload: ${nonExistentWorkloadName}`
                        );

                        try {
                            runaiCliV2.describe(
                                {},
                                nonExistentWorkloadName,
                                cliV2WorkloadType.inferenceNim
                            );
                            expect.fail('Expected describe to fail for non-existent workload');
                        } catch (error) {
                            const errorMessage = (error as Error).message || String(error);
                            logger.info(`CLI error message: ${errorMessage}`);

                            expect(errorMessage).to.include('failed with code 1');
                            logger.info('Verified command failed with exit code 1');
                            expect(errorMessage).to.match(/workload not found|Describe workload failed/i);
                            logger.info('Verified error message indicates workload not found');
                        }
                    });
                });
            });

            test.describe('Interactive commands (port-forward, exec, bash, logs)', function () {
                const NIM_SERVING_PORT = 8000;
                const PORT_FORWARD_LOCAL = 18000;

                test('NIM logs, exec, port-forward, and bash commands run successfully', async function () {
                    test.skip(
                        process.platform !== 'linux',
                        'PTY-based interactive CLI coverage runs on Linux only'
                    );

                    test.setTimeout(INTERACTIVE_TEST_TIMEOUT_MS);

                    const defaultPvcName = await createPvcForTest();

                    await test.step('Submit NIM workload with serving port', () => {
                        const submitFlags = createBasicNimSubmitFlagsWithPvc(defaultPvcName);
                        submitNimWorkload(submitFlags, workloadName, createdWorkloadNames);
                        logger.info(`Submitted NIM workload ${workloadName}`);
                    });

                    await test.step('Wait for workload to reach Running phase', async () => {
                        workloadId = await getWorkloadIdAndWaitForPhase(
                            workloadName,
                            project.id,
                            Phase.Running,
                            PHASE_WAIT_SEC
                        );
                        logger.info(`NIM workload ${workloadName} is running`);
                    });

                    await test.step('Run NIM logs (CLI) and verify it succeeds', () => {
                        const logsResult = runaiCliV2.unifyLogs(
                            workloadName,
                            cliV2WorkloadType.inferenceNim,
                            {}
                        );
                        expect(logsResult.status).to.equal(0, 'runai inference nim logs should exit 0');
                        logger.info('NIM logs command completed successfully');
                    });

                    await test.step('Run NIM exec (CLI) and verify it succeeds', () => {
                        const execResult = runaiCliV2.exec(
                            { [execFlags.command]: 'echo ok' },
                            workloadName,
                            cliV2WorkloadType.inferenceNim
                        );
                        expect(execResult.status).to.equal(0, 'runai inference nim exec should exit 0');
                        logger.info('NIM exec command completed successfully');
                    });

                    await test.step('Run NIM port-forward (CLI) and verify it starts', async () => {
                        const [, childProc] = await runaiCliV2.portForward(
                            { [portForwardFlags.port]: `${PORT_FORWARD_LOCAL}:${NIM_SERVING_PORT}` },
                            workloadName,
                            cliV2WorkloadType.inferenceNim
                        );
                        portForwardProcess = childProc;
                        expect(childProc).to.be.ok;
                        logger.info('NIM port-forward started successfully');
                    });

                    await test.step('Run NIM bash (CLI) and verify it starts', async () => {
                        const [, ptyWrapper] = await runaiCliV2.bashInteractive(
                            {},
                            workloadName,
                            cliV2WorkloadType.inferenceNim,
                            undefined,
                            2000
                        );
                        try {
                            expect(ptyWrapper).to.be.ok;
                            expect(ptyWrapper.kill).to.be.a('function');
                            logger.info('NIM bash command started successfully');
                        } finally {
                            ptyWrapper.kill();
                        }
                    });
                });
            });

            test.describe('Delete Operations', function () {
                test('Delete NIM workload', async function () {
                    test.setTimeout(TEST_TIMEOUT_MS);

                    const defaultPvcName = await createPvcForTest();

                    await test.step('Submit NIM workload', () => {
                        const submitFlags = createBasicNimSubmitFlagsWithPvc(defaultPvcName);
                        submitNimWorkload(submitFlags, workloadName, createdWorkloadNames);
                        logger.info(`Submitted NIM workload ${workloadName}`);
                    });

                    await test.step('Wait for workload to reach Running phase', async () => {
                        workloadId = await getWorkloadIdAndWaitForPhase(
                            workloadName,
                            project.id,
                            Phase.Running,
                            PHASE_WAIT_SEC
                        );
                        logger.info(`NIM workload ${workloadName} is running`);
                    });

                    await test.step('Delete workload via CLI', () => {
                        deleteNimWorkload(workloadName, createdWorkloadNames);
                        logger.info(`Deleted NIM workload ${workloadName}`);
                    });

                    await test.step('Verify workload is removed', async () => {
                        await waitUtils.waitForCondition(
                            async () => {
                                try {
                                    const { body } = await workloadsUtilsV2.getWorkloadByName(
                                        workloadName,
                                        project.id
                                    );
                                    const found = body.workloads?.some((w) => w.name === workloadName);
                                    return !found;
                                } catch {
                                    return true;
                                }
                            },
                            'Wait for workload to be deleted',
                            true,
                            60,
                            5
                        );
                        logger.info('Verified NIM workload was deleted');
                    });
                });

                test('Delete Non-Existent NIM Workload', async function () {
                    await test.step('Verify delete non-existent workload returns error', () => {
                        const nonExistentWorkloadName = 'non-existent-nim-' + Date.now();
                        logger.info(`Attempting to delete non-existent workload: ${nonExistentWorkloadName}`);

                        try {
                            runaiCliV2.delete(
                                undefined,
                                nonExistentWorkloadName,
                                cliV2WorkloadType.inferenceNim,
                                true,
                                true
                            );
                            expect.fail('Expected delete to fail for non-existent workload');
                        } catch (error) {
                            const errorMessage = (error as Error).message || String(error);
                            expect(errorMessage).to.include('failed to search for requested workload');
                            expect(errorMessage).to.include('workload not found');
                            logger.info(
                                'Verified that deleting non-existent workload returns appropriate error'
                            );
                        }
                    });
                });
            });
        });
    }
);
