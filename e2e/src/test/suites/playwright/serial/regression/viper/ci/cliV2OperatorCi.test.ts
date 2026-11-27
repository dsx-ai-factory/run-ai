/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    config,
    flowRootState,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    settings,
    tags,
    waitUtils
} from '../../../../../../../modules/helpers';
import { bashCli, runaiCliV2 } from '../../../../../../../modules/cli';
import { Project } from '../../../../../../../modules/api/apis/projectsDepartmentsApi';
import { assert } from 'chai';
import { Phase } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import {
    SubmitFlagsMerged,
    cliV2WorkloadType,
    loginWithAccessKey,
    portForwardFlags,
    execFlags,
    cliLoginAsSystemAdmin
} from '../../../../../../../modules/cli/tools/RunaiCliV2.constants';
import { assertExpectedValue } from '../../../../../../../conditionPolling/assertUsingPolling';
import { ChildProcess } from 'child_process';
import { BrowserEngine } from '../../../../../../../modules/ui/runaiV2/pages/BrowserEngine';

import { test, expect } from '@playwright/test';

test.describe(
    `CLI V2 operator ci tests  ${tags.teams.viper} ${tags.runOn.serialCI}`,
    { tag: [tags.runOn.serialCI, tags.services.clusterApi] },
    function () {
        let project: Project;
        let videoName: string;
        let browserEngine: BrowserEngine;
        test.beforeAll(async function () {
            project = await projectResourcesUtils.createProjectAndNamespace('cli-login');
        });
        test.afterAll(async function () {
            try {
                await browserEngine.stopAndArchive(this, 'workspaces-' + videoName);
            } catch (e) {
                logger.error(e);
            }
            await operations.forceCleanup();
        });
        ['access-key', 'user'].forEach(function (loginType) {
            test.describe(`login with ${loginType}`, function () {
                test.beforeAll(async function () {
                    if (loginType === 'access-key') {
                        await loginWithAccessKey();
                    } else if (loginType === 'user') {
                        cliLoginAsSystemAdmin();
                    } else if (loginType === 'sso') {
                        videoName = 'sso-login';
                        browserEngine = new BrowserEngine();
                        await browserEngine.start();
                        await runaiCliV2.loginSSOv2(browserEngine, flowRootState.users.system_administrator);
                    }

                    runaiCliV2.projectSet(project.name);
                });

                test.describe('Validate workload logs', function () {
                    const workspaceName = objectGenerator.generateNameByPrefix('workspace-params');
                    const trainingName = objectGenerator.generateNameByPrefix('training-params');
                    const followName = objectGenerator.generateNameByPrefix('follow-params');
                    const randomLog = objectGenerator.generateNameByPrefix('wl-logs-test');

                    test.beforeAll(async function () {
                        const submitFlag = {
                            [SubmitFlagsMerged.image]: settings.isArm
                                ? config.quickStartImage.training
                                : config.quickStartImage.sleepInfinity
                        } as Partial<typeof SubmitFlagsMerged>;

                        runaiCliV2.submit(submitFlag, followName, cliV2WorkloadType.workspace);

                        submitFlag[SubmitFlagsMerged.image] = config.quickStartImage.ubuntu;
                        submitFlag[SubmitFlagsMerged.command] =
                            `-- bash -c 'echo ${randomLog} && sleep 999999'`;
                        runaiCliV2.submit(submitFlag, workspaceName, cliV2WorkloadType.workspace);
                        runaiCliV2.submit(submitFlag, trainingName, cliV2WorkloadType.training);

                        await Promise.all([
                            waitUtils.waitForBackendWorkloadPhase(workspaceName, project.id, Phase.Running),
                            waitUtils.waitForBackendWorkloadPhase(trainingName, project.id, Phase.Running),
                            waitUtils.waitForBackendWorkloadPhase(followName, project.id, Phase.Running)
                        ]);
                    });

                    test('validate workspace logs with follow', async function () {
                        const sleepTime = 3;
                        const workspaceLogs = await runaiCliV2.workloadLogsFollow(
                            followName,
                            cliV2WorkloadType.workspace,
                            sleepTime
                        );
                        const numOfLogs = workspaceLogs.length;
                        // the image used in this workload, outputs 2 logs per second.
                        expect(
                            numOfLogs,
                            `there should be at least ${sleepTime * 2 - 3} logs messages, got instead ${numOfLogs}`
                        ).toBeGreaterThanOrEqual(sleepTime * 2 - 3);
                        expect(
                            numOfLogs,
                            `there should be no more ${sleepTime * 2 + 2} logs messages, got instead ${numOfLogs}`
                        ).toBeLessThanOrEqual(sleepTime * 2 + 2);
                    });

                    test('validate workspace logs', function () {
                        const workspaceLogs = runaiCliV2.workspaceLogs(workspaceName, {});
                        expect(
                            workspaceLogs.stdout.includes(randomLog),
                            'workspace logs should include the random log'
                        ).toBeTruthy();
                    });

                    test('validate training logs', function () {
                        const workspaceLogs = runaiCliV2.trainingLogs(trainingName, {});
                        expect(
                            workspaceLogs.stdout.includes(randomLog),
                            'training logs should include the random log'
                        ).toBeTruthy();
                    });
                });

                test.describe('Validate port forward and exec command', function () {
                    const trainingName = objectGenerator.generateNameByPrefix('test-port-forward');
                    const localPort = '30003';
                    const podListenPort = '8080';
                    let process: ChildProcess;
                    const submitFlag = {
                        [SubmitFlagsMerged.image]: settings.isArm
                            ? 'nginx'
                            : config.quickStartImage.inferenceHelloWorld
                    } as Partial<typeof SubmitFlagsMerged>;
                    test.beforeAll(async function () {
                        runaiCliV2.submit(submitFlag, trainingName, cliV2WorkloadType.training);

                        await waitUtils.waitForBackendWorkloadPhase(trainingName, project.id, Phase.Running);
                    });

                    test.afterAll(function () {
                        try {
                            process.kill();
                        } catch (e) {
                            console.error(e);
                        }
                    });

                    test('test port forward', async function () {
                        const portforwardFlag = {
                            [portForwardFlags.port]: `${localPort}:${podListenPort}`
                        } as Partial<typeof portForwardFlags>;
                        process = (
                            await runaiCliV2.portForward(
                                portforwardFlag,
                                trainingName,
                                cliV2WorkloadType.training
                            )
                        )[1];
                        assertExpectedValue(bashCli, bashCli.exec, 'Hello World!\n', [
                            `curl 127.0.0.1:${localPort}`
                        ]);
                        process.kill();
                        assert.throw(
                            () => bashCli.exec(`curl 127.0.0.1:${localPort}`),
                            'Failed to connect to'
                        );
                    });

                    test('test exec', () => {
                        const execFlag = {
                            [execFlags.command]: 'ls'
                        } as Partial<typeof execFlags>;
                        const execOutput = runaiCliV2.exec(
                            execFlag,
                            trainingName,
                            cliV2WorkloadType.training
                        );
                        expect(execOutput.status).toBe(0); //ATM exec command doesnt output to stdout
                    });
                });
            });
        });
    }
);
