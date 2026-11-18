/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    apiUtils,
    config,
    flowRootState,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    waitUtils,
    workloadsUtilsV2
} from '../../../../../../../modules/helpers';
import { Project } from '../../../../../../../modules/api/apis/projectsDepartmentsApi';
import { bashCli, runaiCliV2 } from '../../../../../../../modules/cli';
import {
    cliLoginAsSystemAdmin,
    cliV2TrainingType,
    cliV2WorkloadType,
    execFlags,
    listFlags,
    portForwardFlags,
    SubmitFlagsMerged,
    workloadLog
} from '../../../../../../../modules/cli/tools/RunaiCliV2.constants';
import { Phase } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import { assert } from 'chai';
import { ChildProcess } from 'child_process';
import { assertExpectedValueInclude } from '../../../../../../../conditionPolling/assertUsingPolling';
import { test, expect, TestInfo } from '@playwright/test';
import {
    AnyWorkloadSubmissionType,
    isAnyWorkloadSubmissionEnabled
} from '../../../../../../../modules/helpers/utils/AnyWorkloadSubmissionUtils';
import { skipPlaywrightTest } from '../../../../../../../modules/helpers/utils/SkippedTestUtils';

test.describe(`Workload unify tests  ${tags.teams.viper}`, function () {
    const workloadListLimit = 100;
    const jobNamePrefix = objectGenerator.generateShortNameByPrefix('w', 5);
    const failCommand = '-- cat nonExistCommand';
    const randomLog = objectGenerator.generateNameByPrefix('wl-logs-test');
    const runningCommand = `-- bash -c 'echo ${randomLog} && sleep 999999'`;
    const logsFollowCommand = '-- /bin/bash -c "c=0; while true; do echo \\$c; ((c++)); sleep 1; done"';
    const portforwardCommand = '-- /bin/bash -c "service nginx start && sleep 100000"';
    const runningSuffix = 'running';
    const failSuffix = 'fail';
    const completedSuffix = 'completed';
    const gpuSuffix = '-gpu';
    const workloadsCounter = {};
    const logsFollowSuffix = 'logs-follow';
    const portForwardSuffix = 'port-forward';
    const cleanPodPolicyAll = 'All';
    let redirectedTrainingEnabled = false;

    // Returns GPU portions up to `maxPortion` while keeping their total within `maxTotal`.
    function generateGpuPortions(count: number, maxTotal: number = 16, maxPortion: number = 1): string[] {
        const portions: string[] = [];
        let allocated = 0;
        for (let i = 0; i < count; i++) {
            // Minimum budget needed for all subsequent workloads (each gets at least 0.5)
            const minForRemaining = (count - i - 1) * 0.5;
            const maxForThis = maxTotal - allocated - minForRemaining;
            const portion = maxPortion >= 1 && maxForThis >= 1 && Math.random() < 0.5 ? 1 : 0.5;
            portions.push(String(portion));
            allocated += portion;
        }
        return portions;
    }

    function submitWorkload(
        jobName: string,
        type: cliV2TrainingType,
        submitFlag: Partial<typeof SubmitFlagsMerged>,
        gpuPortion: string
    ): void {
        runaiCliV2.submit(submitFlag, jobName, cliV2WorkloadType.training, type as cliV2TrainingType);
        jobName = `${jobName}${gpuSuffix}`;
        // submitFlag[SubmitFlagsMerged.gpuRequest] = 1;

        submitFlag[SubmitFlagsMerged.gpuPortionRequest] = gpuPortion;
        runaiCliV2.submit(submitFlag, jobName, cliV2WorkloadType.training, type as cliV2TrainingType);
    }

    function generateWorkloads(): void {
        const gpuPortions = generateGpuPortions(5 * Object.values(cliV2TrainingType).length, 16);
        let gpuIndex = 0;
        let jobName: string;
        const submitFlag = {
            [SubmitFlagsMerged.image]: config.quickStartImage.ubuntu
        } as Partial<typeof SubmitFlagsMerged>;

        Object.values(cliV2TrainingType).forEach((type: cliV2TrainingType) => {
            jobName = `${jobNamePrefix}-${type}-${completedSuffix}`;
            submitWorkload(jobName, type, structuredClone(submitFlag), gpuPortions[gpuIndex++]);
            workloadsCounter[type] += 2;
        });

        Object.values(cliV2TrainingType).forEach((type) => {
            const currentFlags = structuredClone(submitFlag);
            currentFlags[SubmitFlagsMerged.command] = runningCommand;
            jobName = `${jobNamePrefix}-${type}-${runningSuffix}`;
            submitWorkload(jobName, type, currentFlags, gpuPortions[gpuIndex++]);
            workloadsCounter[type] += 2;
        });

        Object.values(cliV2TrainingType).forEach((type) => {
            const currentFlags = structuredClone(submitFlag);
            currentFlags[SubmitFlagsMerged.command] = logsFollowCommand;
            jobName = `${jobNamePrefix}-${type}-${logsFollowSuffix}`;
            submitWorkload(jobName, type, currentFlags, gpuPortions[gpuIndex++]);
            workloadsCounter[type] += 2;
        });

        Object.values(cliV2TrainingType).forEach((type) => {
            const currentFlags = structuredClone(submitFlag);
            currentFlags[SubmitFlagsMerged.backoffLimit] = 1;
            currentFlags[SubmitFlagsMerged.command] = failCommand;
            jobName = `${jobNamePrefix}-${type}-${failSuffix}`;
            submitWorkload(jobName, type, currentFlags, gpuPortions[gpuIndex++]);
            workloadsCounter[type] += 2;
        });

        Object.values(cliV2TrainingType).forEach((type) => {
            const currentFlags = structuredClone(submitFlag);
            currentFlags[SubmitFlagsMerged.command] = portforwardCommand;
            currentFlags[SubmitFlagsMerged.image] = 'nginx';
            jobName = `${jobNamePrefix}-${type}-${portForwardSuffix}`;
            submitWorkload(jobName, type, currentFlags, gpuPortions[gpuIndex++]);
            workloadsCounter[type] += 2;
        });
    }

    test.beforeAll(async function () {
        await apiUtils.setResourceOptimization(true);
        await operations.forceCleanup();
        cliLoginAsSystemAdmin();
        redirectedTrainingEnabled = await isAnyWorkloadSubmissionEnabled(
            AnyWorkloadSubmissionType.Training
        );
    });

    test.describe('Workloads submit and list validation', function () {
        let project: Project;

        function validateListValue(listFlagsParams, shouldBeZero = true, addFrameworkFlag = true): void {
            for (const type of Object.values(cliV2TrainingType)) {
                if (addFrameworkFlag) {
                    listFlagsParams[listFlags.framework] = type;
                }
                const cmdType = addFrameworkFlag ? '' : type;
                const listResult = runaiCliV2.list(`training ${cmdType}`, listFlagsParams).stdout;
                const expectedValue = shouldBeZero
                    ? '{\n' + '  "workloads": []\n' + '}\n'
                    : workloadsCounter[type];
                const listResultLength = shouldBeZero
                    ? listResult
                    : JSON.parse(listResult)['workloads'].length;
                expect(expectedValue === listResultLength, `workloads type ${type} count`).toBeTruthy();
            }
        }

        test.beforeAll(async function () {
            for (const type of Object.values(cliV2TrainingType)) {
                workloadsCounter[type] = 0;
            }
            project = await projectResourcesUtils.createProjectAndNamespace('unify', undefined, 30, 30);
            runaiCliV2.projectSet(project.name);
            generateWorkloads();
        });

        test.afterAll(async function () {
            try {
                await operations.deleteProjectAndNamespace(project.id);
            } catch (error) {
                logger.error(`Failed to delete project ${project.name}`);
            }
        });

        // Before each test, log the workload list and verify no workload is stuck in Pending.
        // Workloads briefly pass through Pending on startup, so we poll up to 60 s for them to transition.
        // If any workload is still Pending after the timeout it is likely stalled due to GPU exhaustion.
        test.beforeEach(async function () {
            const result = await workloadsUtilsV2.getWorkloadsByProjectId(String(project.id));
            const workloads = result.body.workloads;
            logger.info(
                `Workloads in project ${project.id}: ${JSON.stringify(workloads.map((w) => ({ name: w.name, phase: w.phase })))}`
            );
            await waitUtils.waitForCondition(
                async () => {
                    const r = await workloadsUtilsV2.getWorkloadsByProjectId(String(project.id));
                    return !r.body.workloads.some((w) => w.phase === Phase.Pending);
                },
                'workloads to leave Pending state — possible GPU exhaustion',
                true,
                60
            );
        });

        test('validate completed workloads', async function () {
            for (const type of Object.values(cliV2TrainingType)) {
                await waitUtils.waitForBackendWorkloadPhase(
                    `${jobNamePrefix}-${type}-${completedSuffix}`,
                    project.id,
                    Phase.Completed,
                    undefined,
                    workloadListLimit
                );
                await waitUtils.waitForBackendWorkloadPhase(
                    `${jobNamePrefix}-${type}-${completedSuffix}${gpuSuffix}`,
                    project.id,
                    Phase.Completed,
                    undefined,
                    workloadListLimit
                );
            }
        });
        test('validate running workloads', async function () {
            for (const suffix of [runningSuffix, logsFollowSuffix]) {
                for (const type of Object.values(cliV2TrainingType)) {
                    await waitUtils.waitForBackendWorkloadPhase(
                        `${jobNamePrefix}-${type}-${suffix}`,
                        project.id,
                        Phase.Running,
                        undefined,
                        workloadListLimit
                    );
                    await waitUtils.waitForBackendWorkloadPhase(
                        `${jobNamePrefix}-${type}-${suffix}${gpuSuffix}`,
                        project.id,
                        Phase.Running,
                        undefined,
                        workloadListLimit
                    );
                }
            }
        });

        test('validate failed workloads', async function () {
            for (const type of Object.values(cliV2TrainingType)) {
                await waitUtils.waitForBackendWorkloadPhase(
                    `${jobNamePrefix}-${type}-${failSuffix}`,
                    project.id,
                    Phase.Failed,
                    undefined,
                    workloadListLimit
                );
                await waitUtils.waitForBackendWorkloadPhase(
                    `${jobNamePrefix}-${type}-${failSuffix}${gpuSuffix}`,
                    project.id,
                    Phase.Failed,
                    undefined,
                    workloadListLimit
                );
            }
        });
        test('validate training list by platform', function () {
            const listFlagsParams = {
                [listFlags.json]: ''
            } as Partial<typeof listFlags>;
            validateListValue(structuredClone(listFlagsParams), false, true);
        });
        test('validate training list with framework flag', function () {
            const listFlagsParams = {
                [listFlags.json]: ''
            } as Partial<typeof listFlags>;
            validateListValue(structuredClone(listFlagsParams), false, false);
        });

        test.describe('validate cluster commands', function () {
            const localPort = '8080';
            const podListenPort = '80';
            let process: ChildProcess;
            const randomFileName = objectGenerator.generateNameByPrefix('test');
            const execCommandFileCreate = `sh -c 'echo something > ${randomFileName}'`;
            const execCommandLsFile = 'ls';

            function skipRedirectedClusterCommand(testInfo: TestInfo): void {
                if (redirectedTrainingEnabled) {
                    skipPlaywrightTest(
                        test,
                        testInfo,
                        'CLI cluster commands cannot locate WLR-backed training workloads'
                    );
                }
            }

            test('validate workload logs', ({}, testInfo) => {
                skipRedirectedClusterCommand(testInfo);
                for (const gpu of ['', gpuSuffix]) {
                    for (const type of Object.values(cliV2TrainingType)) {
                        const wlLogs = runaiCliV2.unifyLogs(
                            `${jobNamePrefix}-${type}-${runningSuffix}${gpu}`,
                            type,
                            {}
                        );
                        expect(
                            wlLogs.stdout.includes(randomLog),
                            'workspace logs should include the random log'
                        ).toBeTruthy();
                    }
                }
            });
            test('validate workload logs follow', async ({}, testInfo) => {
                skipRedirectedClusterCommand(testInfo);
                const sleepTime = 5;
                for (const gpu of ['', gpuSuffix]) {
                    for (const type of Object.values(cliV2TrainingType)) {
                        const wlName = `${jobNamePrefix}-${type}-${logsFollowSuffix}${gpu}`;
                        const wlLogs = await runaiCliV2.workloadLogsFollow(wlName, type, sleepTime);
                        const numOfLogs = wlLogs.length;
                        expect(
                            numOfLogs,
                            `${wlName}\nthere should be at least 2 logs messages, got instead ${numOfLogs}`
                        ).toBeGreaterThanOrEqual(2);
                    }
                }
            });
            test('validate any workload logs', () => {
                for (const gpu of ['', gpuSuffix]) {
                    for (const type of Object.values(cliV2TrainingType)) {
                        const wlLogs = runaiCliV2.unifyLogs(
                            `${jobNamePrefix}-${type}-${runningSuffix}${gpu}`,
                            workloadLog.noun,
                            {}
                        );
                        expect(
                            wlLogs.stdout.includes(randomLog),
                            'workspace logs should include the random log'
                        ).toBeTruthy();
                    }
                }
            });
            test('validate any workload logs follow', async () => {
                const sleepTime = 3;
                for (const gpu of ['', gpuSuffix]) {
                    for (const type of Object.values(cliV2TrainingType)) {
                        const wlName = `${jobNamePrefix}-${type}-${logsFollowSuffix}${gpu}`;
                        const wlLogs = await runaiCliV2.workloadLogsFollow(
                            wlName,
                            workloadLog.noun,
                            sleepTime
                        );
                        const text = Buffer.concat(
                            wlLogs.map((c) => (Buffer.isBuffer(c) ? c : Buffer.from(String(c))))
                        ).toString();
                        const numOfLogs = text.split('\n').filter((l) => l.trim().length > 0).length;
                        expect(
                            numOfLogs,
                            `${wlName}\nthere should be at least ${sleepTime} log lines, got instead ${numOfLogs}`
                        ).toBeGreaterThanOrEqual(sleepTime);
                    }
                }
            });
            test('validate port forward', async function ({}, testInfo) {
                skipRedirectedClusterCommand(testInfo);
                const portforwardFlag = {
                    [portForwardFlags.port]: `${localPort}:${podListenPort}`
                } as Partial<typeof portForwardFlags>;
                for (const gpu of ['', gpuSuffix]) {
                    for (const type of Object.values(cliV2TrainingType)) {
                        process = (
                            await runaiCliV2.portForward(
                                portforwardFlag,
                                `${jobNamePrefix}-${type}-${portForwardSuffix}${gpu}`,
                                type
                            )
                        )[1];
                        assertExpectedValueInclude(bashCli, bashCli.exec, 'Welcome to nginx', [
                            `curl localhost:${localPort}`
                        ]);
                        process.kill();
                        assert.throw(
                            () => bashCli.exec(`curl localhost:${localPort}`),
                            'Failed to connect to'
                        );
                    }
                }
            });
            test('validate exec', function ({}, testInfo) {
                skipRedirectedClusterCommand(testInfo);
                const execFlag = {
                    [execFlags.command]: execCommandFileCreate
                } as Partial<typeof execFlags>;
                Object.values(cliV2TrainingType).map(function (type) {
                    execFlag[execFlags.command] = execCommandFileCreate;
                    runaiCliV2.exec(execFlag, `${jobNamePrefix}-${type}-${runningSuffix}`, type).stdout;
                    execFlag[execFlags.command] = execCommandLsFile;
                    const execLsOutput = runaiCliV2.exec(
                        execFlag,
                        `${jobNamePrefix}-${type}-${runningSuffix}`,
                        type
                    ).stdout;
                    expect(execLsOutput).toContain(randomFileName);
                });
            });
        });
        test.describe('change project and validate data leaks', function () {
            let secondProject: Project;
            test.beforeAll(async function () {
                secondProject = await projectResourcesUtils.createProjectAndNamespace('unify2');
                runaiCliV2.projectSet(secondProject.name);
            });

            test.afterAll(async function () {
                try {
                    await operations.deleteProjectAndNamespace(secondProject.id);
                } catch (error) {
                    logger.error(`Failed to delete project ${secondProject.name}`);
                }
            });
            test('validate training list by platform data leak', function () {
                const listFlagsParams = {
                    [listFlags.json]: ''
                } as Partial<typeof listFlags>;
                validateListValue(structuredClone(listFlagsParams), true, false);
            });
            test('validate training list with framework flag data leak', function () {
                const listFlagsParams = {
                    [listFlags.json]: ''
                } as Partial<typeof listFlags>;
                validateListValue(structuredClone(listFlagsParams), true, true);
            });
            test('validate unable to delete workload from different project', function () {
                for (const type of Object.values(cliV2TrainingType)) {
                    for (const phase of [completedSuffix, runningSuffix, failSuffix]) {
                        assert.throw(
                            () => runaiCliV2.delete(undefined, `${jobNamePrefix}-${type}-${phase}`, type),
                            'workload not found'
                        );
                        assert.throw(
                            () =>
                                runaiCliV2.delete(
                                    undefined,
                                    `${jobNamePrefix}-${type}-${phase}${gpuSuffix}`,
                                    type
                                ),
                            'workload not found'
                        );
                    }
                }
            });
        });
        test.describe('validate workload deletion', function () {
            test('validate workload deletion', async function () {
                runaiCliV2.projectSet(project.name);
                const testSuffixes = [
                    completedSuffix,
                    runningSuffix,
                    failSuffix,
                    logsFollowSuffix,
                    portForwardSuffix
                ];
                for (const type of Object.values(cliV2TrainingType)) {
                    for (const suffix of testSuffixes) {
                        runaiCliV2.delete(undefined, `${jobNamePrefix}-${type}-${suffix}`, type);
                        runaiCliV2.delete(undefined, `${jobNamePrefix}-${type}-${suffix}${gpuSuffix}`, type);
                    }
                }
                for (const type of Object.values(cliV2TrainingType)) {
                    for (const suffix of testSuffixes) {
                        await waitUtils.waitForCondition(
                            () =>
                                runaiCliV2
                                    .list('workload')
                                    .stdout.toString()
                                    .includes(`${jobNamePrefix}-${type}-${suffix}`) ||
                                runaiCliV2
                                    .list('workload')
                                    .stdout.toString()
                                    .includes(`${jobNamePrefix}-${type}-${suffix}${gpuSuffix}`),
                            `wait the workload ${`${jobNamePrefix}-${type}-${suffix}`} to be deleted`,
                            false,
                            20
                        );
                    }
                }
                const listFlagsParams = {
                    [listFlags.json]: ''
                } as Partial<typeof listFlags>;
                validateListValue(structuredClone(listFlagsParams), true, false);
            });
        });
    });
    test.describe('Distributed workload special flags', function () {
        let project: Project;
        test.beforeAll(async function () {
            project = await projectResourcesUtils.createProjectAndNamespace('unify2');
            runaiCliV2.projectSet(project.name);
        });

        test.afterAll(async function () {
            try {
                await operations.deleteProjectAndNamespace(project.id);
            } catch (error) {
                logger.error(`Failed to delete project ${project.name}`);
            }
        });
        test('Validate MPI flag submission', async function () {
            const submitFlag = {
                [SubmitFlagsMerged.image]: config.quickStartImage.ubuntu,
                [SubmitFlagsMerged.slotsPerWorker]: 1
            } as Partial<typeof SubmitFlagsMerged>;
            const mpiJob = objectGenerator.generateShortNameByPrefix('mpi-job', 5);
            const pyTorch = objectGenerator.generateShortNameByPrefix('py-job', 5);
            runaiCliV2.submit(submitFlag, mpiJob, cliV2WorkloadType.training, cliV2TrainingType.mpi);
            await waitUtils.waitForBackendWorkloadPhase(mpiJob, project.id, Phase.Completed);
            assert.throw(
                () =>
                    runaiCliV2.submit(
                        submitFlag,
                        pyTorch,
                        cliV2WorkloadType.training,
                        cliV2TrainingType.pyTorch
                    ),
                `unknown flag: ${SubmitFlagsMerged.slotsPerWorker}`
            );
        });
        test('Validate CleanPodPolicy flag submission', async function () {
            const submitFlag = {
                [SubmitFlagsMerged.image]: config.quickStartImage.ubuntu,
                [SubmitFlagsMerged.cleanPodPolicy]: cleanPodPolicyAll
            } as Partial<typeof SubmitFlagsMerged>;
            const mpiJob = objectGenerator.generateShortNameByPrefix('mpi-job', 5);
            const pyTorch = objectGenerator.generateShortNameByPrefix('py-job', 5);
            runaiCliV2.submit(submitFlag, mpiJob, cliV2WorkloadType.training, cliV2TrainingType.mpi);
            runaiCliV2.submit(submitFlag, pyTorch, cliV2WorkloadType.training, cliV2TrainingType.pyTorch);

            const waitTasks = [
                waitUtils.waitForBackendWorkloadPhase(mpiJob, project.id, Phase.Completed),
                waitUtils.waitForBackendWorkloadPhase(pyTorch, project.id, Phase.Completed)
            ];
            await Promise.all(waitTasks);
        });
        test('Validate pyTorch/Tensor Flow flag submission', async function () {
            const tfJob = objectGenerator.generateShortNameByPrefix('tf-job', 5);
            const pyTorch = objectGenerator.generateShortNameByPrefix('py-job', 5);
            const mpiJob = objectGenerator.generateShortNameByPrefix('mpi-job', 5);
            const submitFlag = {
                [SubmitFlagsMerged.image]: config.quickStartImage.ubuntu,
                [SubmitFlagsMerged.noMaster]: ''
            } as Partial<typeof SubmitFlagsMerged>;
            runaiCliV2.submit(submitFlag, tfJob, cliV2WorkloadType.training, cliV2TrainingType.tensorFlow);
            assert.throw(
                () =>
                    runaiCliV2.submit(submitFlag, mpiJob, cliV2WorkloadType.training, cliV2TrainingType.mpi),
                `unknown flag: ${SubmitFlagsMerged.noMaster}`
            );
            delete submitFlag[SubmitFlagsMerged.noMaster];
            submitFlag[SubmitFlagsMerged.maxReplicas] = 2;
            assert.throw(
                () =>
                    runaiCliV2.submit(submitFlag, mpiJob, cliV2WorkloadType.training, cliV2TrainingType.mpi),
                `unknown flag: ${SubmitFlagsMerged.maxReplicas}`
            );
            delete submitFlag[SubmitFlagsMerged.maxReplicas];
            submitFlag[SubmitFlagsMerged.minReplicas] = 1;
            assert.throw(
                () =>
                    runaiCliV2.submit(submitFlag, mpiJob, cliV2WorkloadType.training, cliV2TrainingType.mpi),
                `unknown flag: ${SubmitFlagsMerged.minReplicas}`
            );
            submitFlag[SubmitFlagsMerged.maxReplicas] = 2;
            submitFlag[SubmitFlagsMerged.noMaster] = '';
            runaiCliV2.submit(submitFlag, pyTorch, cliV2WorkloadType.training, cliV2TrainingType.pyTorch);
            assert.throw(
                () =>
                    runaiCliV2.submit(
                        submitFlag,
                        tfJob,
                        cliV2WorkloadType.training,
                        cliV2TrainingType.tensorFlow
                    ),
                'replica'
            );
            await waitUtils.waitForBackendWorkloadPhase(pyTorch, project.id, Phase.Completed);
        });
    });
});
