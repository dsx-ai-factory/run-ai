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
    serviceAccountUtils,
    settings,
    tags,
    waitUtils,
    workloadsUtilsV2
} from '../../../../../../../modules/helpers';
import { runaiCliV2 } from '../../../../../../../modules/cli';
import { Project } from '../../../../../../../modules/api/apis/projectsDepartmentsApi';
import { assert } from 'chai';
import { Phase } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import {
    SubmitFlagsMerged,
    cliV2WorkloadType,
    loginWithAccessKey,
    cliLoginAsSystemAdmin
} from '../../../../../../../modules/cli/tools/RunaiCliV2.constants';
import { BrowserEngine } from '../../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { test, expect } from '@playwright/test';
import {
    AnyWorkloadSubmissionType,
    isAnyWorkloadSubmissionEnabled
} from '../../../../../../../modules/helpers/utils/AnyWorkloadSubmissionUtils';
import { skipPlaywrightTest } from '../../../../../../../modules/helpers/utils/SkippedTestUtils';

test.describe(`CLI V2 Control Plane Tests  ${tags.teams.viper} ${tags.runOn.serialCI}`, { tag: [tags.runOn.serialCI, tags.services.runaiCli] }, function () {
    let project: Project;
    let clusterName: string;
    let videoName: string;
    let browserEngine: BrowserEngine;

    test.beforeAll(async function () {
        project = await projectResourcesUtils.createProjectAndNamespace('cli-login');
    });

    test.afterAll(async function () {
        try {
            await browserEngine.stopAndArchive(this, 'CLI v2 Control plane-' + videoName);
        } catch (e) {
            logger.error(e);
        }
        await operations.forceCleanup();
    });

    test.describe('basic commands', function () {
        ['access-key', 'user', 'sso'].forEach(function (loginType) {
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
                });
                test('cluster list', function () {
                    const cmdOutput = runaiCliV2.list('cluster');
                    try {
                        clusterName = cmdOutput.stdout
                            .split('\n')
                            .filter((x) => x.startsWith(` ${config.clusterNamePrefix}`))[0]
                            .split(' ')[1];
                    } catch {
                        throw new Error('List cluster should return a cluster with the host prefix');
                    }
                });

                test('cluster set', function () {
                    runaiCliV2.clusterSet(clusterName);
                });
                test('project set', function () {
                    runaiCliV2.projectSet(project.name);
                });
                test('project list', function () {
                    const cmdOutput = runaiCliV2.list('project');
                    expect(
                        cmdOutput.stdout.toString().includes(project.name),
                        'project list should include the project'
                    ).toBeTruthy();
                });
                test('workload list', function () {
                    const cmdOutput = runaiCliV2.list('workload');
                    expect(cmdOutput.stdout.toString().startsWith(' Workload')).toBeTruthy();
                });
                test('nodes list', function () {
                    const cmdOutput = runaiCliV2.list('node');
                    flowRootState.nodesResources.forEach(function (params) {
                        cmdOutput.stdout.toString().includes(params.name);
                    });
                });
                test('version', function () {
                    const cmdOutput = runaiCliV2.version();
                    expect(cmdOutput.stdout.toString().startsWith('runai')).toBeTruthy();
                });

                test('help', function () {
                    const cmdOutput = runaiCliV2.help();
                    expect(
                        cmdOutput
                            .toString()
                            .startsWith('runai - The Run:ai Researcher Command Line Interface')
                    ).toBeTruthy();
                });
            });
        });

        for (const wlType of [cliV2WorkloadType.workspace, cliV2WorkloadType.training]) {
            const image = settings.isArm
                ? config.quickStartImage.training
                : config.quickStartImage.sleepInfinity;
            const submissionType =
                wlType === cliV2WorkloadType.workspace
                    ? AnyWorkloadSubmissionType.Workspace
                    : AnyWorkloadSubmissionType.Training;
            test.describe(`${wlType} submission tests`, function () {
                test.describe.configure({ mode: 'serial' });
                const wlTypeName = objectGenerator.generateNameByPrefix(wlType);

                test.beforeAll(function () {
                    runaiCliV2.projectSet(project.name);
                });

                test(`submit running ${wlType}`, async function () {
                    const submitFlag = {
                        [SubmitFlagsMerged.image]: image
                    } as Partial<typeof SubmitFlagsMerged>;
                    runaiCliV2.submit(submitFlag, wlTypeName, wlType);
                    await waitUtils.waitForBackendWorkloadPhase(wlTypeName, project.id, Phase.Running);
                });
                test('validate copy and edit api call', async function ({}, testInfo) {
                    if (await isAnyWorkloadSubmissionEnabled(submissionType)) {
                        skipPlaywrightTest(
                            test,
                            testInfo,
                            'Redirected CLI workloads do not preserve additionalFields.cliCommand'
                        );
                    }
                    const expectedCliCommand = `${runaiCliV2.cliPath} ${wlType} submit ${wlTypeName} --image ${image}`;
                    const allWorkloadsResponse = await workloadsUtilsV2.getWorkloadsByProjectId(
                        project.id.toString()
                    );
                    const createdWorkload = allWorkloadsResponse.body.workloads.find(
                        (wl) => wl.name === wlTypeName
                    );
                    expect(
                        createdWorkload.additionalFields['cliCommand'],
                        'The workload details  should have the same command that was used to create him'
                    ).toBe(expectedCliCommand);
                });
                test('Validate workload list', function () {
                    const isWorkloadInList = runaiCliV2
                        .list('workload')
                        .stdout.toString()
                        .includes(wlTypeName);
                    expect(isWorkloadInList, 'The running workload should show on the list').toBeTruthy();
                });
                test(`suspend running ${wlType}`, async function () {
                    runaiCliV2.suspend({}, wlTypeName, wlType);
                    await waitUtils.waitForBackendWorkloadPhase(wlTypeName, project.id, Phase.Stopped);
                });
                test(`resume suspended ${wlType}`, async function () {
                    runaiCliV2.resume({}, wlTypeName, wlType);
                    await waitUtils.waitForBackendWorkloadPhase(wlTypeName, project.id, Phase.Running);
                });
                test(`delete ${wlType}`, async function () {
                    runaiCliV2.delete({}, wlTypeName, wlType);
                    await waitUtils.waitForCondition(
                        () => runaiCliV2.list('workload').stdout.toString().includes(wlTypeName),
                        'wait the workload to be deleted',
                        false,
                        40
                    );
                });
            });
        }

        test.describe('Validate unauthorized msg', function () {
            const testParams = [
                'nodepool',
                'cluster',
                'project',
                'workload',
                'node',
                'training',
                'workspace'
            ];
            test.beforeAll(async function () {
                const serviceAccountName = objectGenerator.generateShortNameByPrefix('cli-v2');
                const serviceAccount = await serviceAccountUtils.createServiceAccount(
                    flowRootState.users.system_administrator.idToken,
                    { name: serviceAccountName }
                );
                runaiCliV2.loginWithAccessKey(
                    serviceAccount.body.clientId,
                    serviceAccount.body.clientSecret
                );
                project = await projectResourcesUtils.createProjectAndNamespace('cli-login');
            });
            testParams.forEach(function (param) {
                test(`list ${param}`, function () {
                    const insufficientPermissionsMsg =
                        'You have insufficient permissions. Ensure you have the required role/permissions to run this command';
                    assert.throw(() => runaiCliV2.list(param), insufficientPermissionsMsg);
                });
            });
        });
    });
});
