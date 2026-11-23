/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect, test } from '@playwright/test';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { runaiCliV2 } from '../../../../../../modules/cli';
import {
    cliLoginAsSystemAdmin,
    cliV2TrainingType,
    cliV2WorkloadType,
    cliV2workloadNoun,
    DeleteFlagsMerged,
    listFlags,
    preemptibilityOptions,
    SubmitFlagsMerged
} from '../../../../../../modules/cli/tools/RunaiCliV2.constants';
import {
    config,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    waitUtils
} from '../../../../../../modules/helpers';
import { Phase } from '../../../../../../modules/api/apis/assetsApi';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { skipPlaywrightTest } from '../../../../../../modules/helpers/utils/SkippedTestUtils';

// Workload configuration interface
interface WorkloadConfig {
    kind: string;
    namePrefix: string;
    workloadType: cliV2WorkloadType;
    trainingType?: cliV2TrainingType;
    gpuPortion: string;
    preemptible?: boolean;
    workers?: string;
}

test.describe(
    `Workloads info tests CLIv2, tags: ${tags.resource.workloads} ${tags.teams.viper} ${tags.runOn.parallelCI}`,
    { tag: [tags.services.runaiCli] },
    function () {
        const imagePath = config.quickStartImage.sleepInfinity;
        const runningCommand = '-- sleep infinity';

        // Workload kind constants
        const WorkloadKindWorkspace = 'workspace';
        const WorkloadKindWorkspacePreemptible = 'workspace-preemptible';
        const WorkloadKindTraining = 'training';
        const WorkloadKindPytorch = 'pytorch';

        // Base workload configurations (non-preemptible, available from V2_18)
        const workloadConfigs: WorkloadConfig[] = [
            {
                kind: WorkloadKindWorkspace,
                namePrefix: 'ws',
                workloadType: cliV2WorkloadType.workspace,
                gpuPortion: '0.2'
            },
            {
                kind: WorkloadKindTraining,
                namePrefix: 'tr',
                workloadType: cliV2WorkloadType.training,
                gpuPortion: '0.3'
            },
            {
                kind: WorkloadKindPytorch,
                namePrefix: 'pytorch',
                workloadType: cliV2WorkloadType.pyTorch,
                gpuPortion: '0.2',
                workers: '2'
            }
        ];

        let projectName: string;
        let project: Project;

        const workloadNames: { [kind: string]: string } = {};

        const deleteFlag = {
            [DeleteFlagsMerged.byPassDialog]: ''
        } as Partial<typeof DeleteFlagsMerged>;

        test.beforeAll(async function () {
            cliLoginAsSystemAdmin();

            // Create project
            project = await projectResourcesUtils.createProjectAndNamespace(
                'wl-info-cliv2',
                undefined,
                3,
                -1
            );
            projectName = project.name;

            // Generate workload names and submit workloads
            const nameSuffix = objectGenerator.generateNameByPrefix('-wl');

            for (const wlConfig of workloadConfigs) {
                // Generate name
                workloadNames[wlConfig.kind] = `${wlConfig.namePrefix}${nameSuffix}`;

                // Build submit flags
                const submitFlags = {
                    [SubmitFlagsMerged.image]: imagePath,
                    [SubmitFlagsMerged.project]: projectName,
                    [SubmitFlagsMerged.command]: runningCommand,
                    [SubmitFlagsMerged.gpuPortionRequest]: wlConfig.gpuPortion
                } as Partial<typeof SubmitFlagsMerged>;

                if (wlConfig.preemptible) {
                    submitFlags[SubmitFlagsMerged.preemptibility] = preemptibilityOptions.preemptible;
                }
                if (wlConfig.workers) {
                    submitFlags[SubmitFlagsMerged.workers] = wlConfig.workers;
                }

                // Submit workload
                runaiCliV2.submit(
                    submitFlags,
                    workloadNames[wlConfig.kind],
                    wlConfig.workloadType,
                    wlConfig.trainingType
                );
            }

            // Wait for all workloads to be running
            await Promise.all(
                workloadConfigs.map((wlConfig) =>
                    waitUtils.waitForBackendWorkloadPhase(
                        workloadNames[wlConfig.kind],
                        project.id,
                        Phase.Running
                    )
                )
            );
        });

        test.afterAll(async function () {
            try {
                // Delete workloads
                deleteFlag[DeleteFlagsMerged.project] = projectName;
                for (const workloadName of Object.values(workloadNames)) {
                    runaiCliV2.delete(deleteFlag, workloadName, cliV2workloadNoun, false);
                }

                // Delete project
                await operations.deleteProjectAndNamespace(project.id);
                logger.info(`Deleted project ${projectName}`);
            } catch (error) {
                logger.error(`Failed to cleanup project ${projectName} or workloads: ${error}`);
            }
        });

        test('list workspaces should show workspace', function () {
            const result = runaiCliV2.list('workspace', {
                [listFlags.project]: projectName
            } as Partial<typeof listFlags>);

            const output = result.stdout;
            logger.info(`Workspace list output:\n${output}`);

            // Verify workspace is listed
            expect(output).toContain(workloadNames[WorkloadKindWorkspace]);
        });

        test('describe workspace should show correct info', function () {
            const output = runaiCliV2.describeRawFlags(
                workloadNames[WorkloadKindWorkspace],
                cliV2WorkloadType.workspace,
                `-p ${projectName}`
            );

            logger.info(`Workspace describe output:\n${output}`);

            // Verify workload name appears
            expect(output).toContain(workloadNames[WorkloadKindWorkspace]);
        });

        test('describe training should show correct info', function () {
            const output = runaiCliV2.describeRawFlags(
                workloadNames[WorkloadKindTraining],
                cliV2WorkloadType.training,
                `-p ${projectName}`
            );

            logger.info(`Training describe output:\n${output}`);

            // Verify workload name appears
            expect(output).toContain(workloadNames[WorkloadKindTraining]);
        });

        test('describe distributed PyTorch workload should show correct info', function ({}, testInfo) {
            skipPlaywrightTest(
                test,
                testInfo,
                'runai pytorch describe returns "workload not found" for a Running distributed PyTorch workload'
            );
            const pytorchConfig = workloadConfigs.find((c) => c.kind === WorkloadKindPytorch);
            const pytorchGpuPortion = pytorchConfig?.gpuPortion ?? '0.2';
            const pytorchWorkers = parseInt(pytorchConfig?.workers ?? '2');

            const output = runaiCliV2.describeRawFlags(
                workloadNames[WorkloadKindPytorch],
                cliV2WorkloadType.pyTorch,
                `-p ${projectName}`
            );

            logger.info(`PyTorch describe output:\n${output}`);

            // Verify workload name appears
            expect(output).toContain(workloadNames[WorkloadKindPytorch]);
            // Verify it's recognized as PyTorch distributed workload
            expect(output.toLowerCase()).toContain('pytorch');
            // Verify GPU allocation is shown - check for GPU-related field
            expect(output).toMatch(/gpu/i);
            // Expected total GPU: gpuPortion * workers (or per-worker portion shown)
            const expectedTotalGpu = (parseFloat(pytorchGpuPortion) * pytorchWorkers).toFixed(1);
            expect(output).toMatch(new RegExp(`(${pytorchGpuPortion}|${expectedTotalGpu})`, 'i'));
        });

        test('list all workloads (-A) should include all submitted workloads', function () {
            const result = runaiCliV2.list('workload', {
                [listFlags.allScopes]: ''
            } as Partial<typeof listFlags>);

            const output = result.stdout;
            logger.info(`All workloads list output:\n${output}`);

            // Verify all our workloads are included
            expect(output).toContain(workloadNames[WorkloadKindWorkspace]);
            expect(output).toContain(workloadNames[WorkloadKindTraining]);
            expect(output).toContain(workloadNames[WorkloadKindPytorch]);
        });

        test('list workloads in project should show all workloads', function () {
            const result = runaiCliV2.list('workload', {
                [listFlags.project]: projectName
            } as Partial<typeof listFlags>);

            const output = result.stdout;
            logger.info(`Project workloads list output:\n${output}`);

            // Verify all our workloads are included
            expect(output).toContain(workloadNames[WorkloadKindWorkspace]);
            expect(output).toContain(workloadNames[WorkloadKindTraining]);
            expect(output).toContain(workloadNames[WorkloadKindPytorch]);
        });

        // Preemptibility tests - require V2_24+ for --preemptibility flag
        test.describe('Preemptibility tests', function () {
            supportedFrom(SupportedVersions.V2_24)(function () {
                let preemptibleWorkloadName: string;

                test.beforeAll(async function () {
                    const nameSuffix = objectGenerator.generateNameByPrefix('-wl');
                    preemptibleWorkloadName = `wsp${nameSuffix}`;
                    workloadNames[WorkloadKindWorkspacePreemptible] = preemptibleWorkloadName;

                    const submitFlags = {
                        [SubmitFlagsMerged.image]: imagePath,
                        [SubmitFlagsMerged.project]: projectName,
                        [SubmitFlagsMerged.command]: runningCommand,
                        [SubmitFlagsMerged.gpuPortionRequest]: '0.2',
                        [SubmitFlagsMerged.preemptibility]: preemptibilityOptions.preemptible
                    } as Partial<typeof SubmitFlagsMerged>;

                    runaiCliV2.submit(submitFlags, preemptibleWorkloadName, cliV2WorkloadType.workspace);

                    await waitUtils.waitForBackendWorkloadPhase(
                        preemptibleWorkloadName,
                        project.id,
                        Phase.Running
                    );
                });

                test.afterAll(function () {
                    try {
                        const deleteFlags = {
                            [DeleteFlagsMerged.byPassDialog]: '',
                            [DeleteFlagsMerged.project]: projectName
                        } as Partial<typeof DeleteFlagsMerged>;
                        runaiCliV2.delete(deleteFlags, preemptibleWorkloadName, cliV2workloadNoun, false);
                    } catch (error) {
                        logger.error(`Failed to cleanup preemptible workload: ${error}`);
                    }
                });

                test('list workspaces with preemptible column should show preemptibility status', function () {
                    const result = runaiCliV2.list('workspace', {
                        [listFlags.project]: projectName,
                        [listFlags.columns]: 'name,preemptible'
                    } as Partial<typeof listFlags>);

                    const output = result.stdout;
                    logger.info(`Workspace list with preemptible column:\n${output}`);

                    // Verify both workspaces are listed
                    expect(output).toContain(workloadNames[WorkloadKindWorkspace]);
                    expect(output).toContain(preemptibleWorkloadName);

                    // Verify preemptible column header is shown
                    expect(output.toLowerCase()).toContain('preemptible');
                });

                test('describe workspace should show non-preemptible status', function () {
                    test.skip(true, 'Depends on CLI changes');

                    const output = runaiCliV2.describeRawFlags(
                        workloadNames[WorkloadKindWorkspace],
                        cliV2WorkloadType.workspace,
                        `-p ${projectName}`
                    );

                    logger.info(`Workspace describe output:\n${output}`);

                    // Verify preemptibility field is shown as non-preemptible
                    expect(output).toMatch(/Preemptibility:\s+Non-preemptible/);
                });

                test('describe preemptible workspace should show preemptible status', function () {
                    test.skip(true, 'Depends on CLI changes');

                    const output = runaiCliV2.describeRawFlags(
                        preemptibleWorkloadName,
                        cliV2WorkloadType.workspace,
                        `-p ${projectName}`
                    );

                    logger.info(`Preemptible workspace describe output:\n${output}`);

                    // Verify workload name appears
                    expect(output).toContain(preemptibleWorkloadName);

                    // Verify preemptibility field is shown as preemptible
                    expect(output).toMatch(/Preemptibility:\s+Preemptible/);
                });
            });
        });
    }
);
