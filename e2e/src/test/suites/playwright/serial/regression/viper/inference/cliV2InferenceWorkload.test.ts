/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import {
    config,
    flowRootState,
    logger,
    objectGenerator,
    operations,
    orgUnitUtils,
    tags,
    waitUtils
} from '../../../../../../../modules/helpers';
import { Project } from '../../../../../../../modules/api/apis/orgUnitApi';
import { runaiCliV2 } from '../../../../../../../modules/cli';
import {
    cliLoginAsSystemAdmin,
    cliV2WorkloadType,
    listFlags
} from '../../../../../../../modules/cli/tools/RunaiCliV2.constants';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';

test.describe(
    `Inference tests  ${tags.teams.viper} ${tags.resource.inferenceWorkload}`,
    { tag: [tags.resource.inferenceWorkload] },
    () => {
        async function waitForWorkloadCount(expectedCount: number): Promise<void> {
            await waitUtils.waitForCondition(
                () => {
                    const listFlagsParams = { [listFlags.json]: '' } as Partial<typeof listFlags>;
                    const stdoutCustom = runaiCliV2.list(cliV2WorkloadType.inference, listFlagsParams).stdout;
                    const parsedOutput = JSON.parse(stdoutCustom);

                    const workloads = parsedOutput?.workloads;
                    if (!workloads) {
                        logger.debug('No workloads found in parsed output');
                        return expectedCount === 0;
                    }

                    const workloadCount = workloads.length;
                    logger.debug(`workloadCount: ${workloadCount}, expectedCount: ${expectedCount}`);
                    return workloadCount === expectedCount;
                },
                'wait for workloads count',
                true,
                300
            );
        }

        function validateInferenceWorkload(workloadName: string, expectedValues: Record<string, any>): void {
            const res = runaiCliV2.describeRawFlags(workloadName, cliV2WorkloadType.inference, '');
            const parsedRes = res.split('\n').reduce(
                (acc, line) => {
                    const parts = line.split(':').map((p) => p.trim());
                    if (parts.length === 2) {
                        acc[parts[0]] = isNaN(Number(parts[1])) ? parts[1] : Number(parts[1]);
                    }
                    return acc;
                },
                {} as Record<string, any>
            );

            for (const key in expectedValues) {
                expect(parsedRes[key], `Mismatch in ${key}`).toEqual(expectedValues[key]);
            }
        }

        supportedFrom(SupportedVersions.V2_20)(() => {
            const randomLog = objectGenerator.generateNameByPrefix('wlog');
            const runningCommand = `-- bash -c 'echo ${randomLog} && sleep 999999'`;
            const workloadName = objectGenerator.generateNameByPrefix('inf');
            let project: Project;
            const listFlagsParams = { [listFlags.json]: '' } as Partial<typeof listFlags>;

            test.beforeAll(async () => {
                await operations.forceCleanup();
                cliLoginAsSystemAdmin();
            });

            test.beforeAll(async () => {
                project = await orgUnitUtils.createProjectAndValidate({
                    name: objectGenerator.generateShortNameByPrefix('cli-inf'),
                    clusterId: flowRootState.cluster.uuid,
                    resources: [{ gpu: { deserved: 30, limit: 30 } }]
                });
                runaiCliV2.projectSet(project.name);
            });

            test.afterAll(async () => {
                try {
                    await orgUnitUtils.deleteProjectAndNamespace(project.id, project.name);
                } catch (error) {
                    logger.error(`Failed to delete project ${project.name}`);
                }
            });

            test('Inference workloads submit, describe, update, knative, delete validations', async () => {
                await test.step('Submit and list inference workload', async () => {
                    let submitCommand = `--image=${config.quickStartImage.ubuntu} --serving-port=8080 --max-replicas=5 --min-replicas=3 --metric=latency --metric-threshold=10`;

                    supportedFrom(SupportedVersions.V2_23)(() => {
                        submitCommand =
                            submitCommand +
                            ' --initialization-timeout-seconds=180 --request-timeout-seconds=30';
                    });
                    submitCommand = submitCommand + ` ${runningCommand}`;

                    runaiCliV2.submitRawFlags(workloadName, cliV2WorkloadType.inference, submitCommand);

                    await waitForWorkloadCount(1);
                    const stdoutCustom = runaiCliV2.list(cliV2WorkloadType.inference, listFlagsParams).stdout;
                    expect(JSON.parse(stdoutCustom).workloads.length).toBe(1);
                });

                await test.step('Initial inference workload describe', () => {
                    const expectedValues = {
                        Name: workloadName,
                        Type: 'Inference',
                        Project: project.name,
                        Image: config.quickStartImage.ubuntu,
                        'Minimum Replicas': 3,
                        'Maximum Replicas': 5,
                        Metric: 'latency',
                        'Metric Threshold': 10
                    };
                    supportedFrom(SupportedVersions.V2_23)(() => {
                        expectedValues['Initialization Timeout Seconds'] = 180;
                        expectedValues['Request Timeout Seconds'] = 30;
                    });

                    validateInferenceWorkload(workloadName, expectedValues);
                });

                await test.step('Describe created inference workload', () => {
                    const expectedValues = {
                        Name: workloadName,
                        Type: 'Inference',
                        Project: project.name,
                        Image: config.quickStartImage.ubuntu,
                        'Minimum Replicas': 3,
                        'Maximum Replicas': 5,
                        Metric: 'latency',
                        'Metric Threshold': 10
                    };

                    supportedFrom(SupportedVersions.V2_23)(() => {
                        expectedValues['Initialization Timeout Seconds'] = 180;
                        expectedValues['Request Timeout Seconds'] = 30;
                    });

                    validateInferenceWorkload(workloadName, expectedValues);
                });

                await test.step('Wait for Knative service update', async () => {
                    await waitUtils.waitForKnativeRevision(
                        workloadName + '-00001',
                        project.status.namespace,
                        120
                    );
                });

                await test.step('Update inference workload', () => {
                    let updateCommand =
                        '--image=new-img --image-pull-policy=Always --working-dir=/a/b/c --activation-replicas=2 --concurrency-hard-limit=0 --initial-replicas=2  --max-replicas=5 --metric=throughput --metric-threshold=6 --metric-threshold-percentage=70 --min-replicas=0 --scale-down-delay-seconds=8 --scale-to-zero-retention-seconds=9 --cpu-core-limit=2 --cpu-core-request=1 --cpu-memory-limit=4G --cpu-memory-request=2G --extended-resource=resource_name=2 --gpu-devices-request=1 --large-shm --node-pools default';

                    supportedFrom(SupportedVersions.V2_23)(() => {
                        updateCommand =
                            updateCommand +
                            ' --initialization-timeout-seconds=300 --request-timeout-seconds=60';
                    });

                    const res = runaiCliV2.updateRawFlags(
                        workloadName,
                        cliV2WorkloadType.inference,
                        updateCommand
                    );
                    expect(res).toContain('updated successfully');
                });

                await test.step('Wait for Knative service update', async () => {
                    await waitUtils.waitForKnativeRevision(
                        workloadName + '-00002',
                        project.status.namespace,
                        120
                    );
                });

                await test.step('Describe updated inference workload', () => {
                    const expectedValues = {
                        Name: workloadName,
                        Type: 'Inference',
                        Project: project.name,
                        Image: 'new-img',
                        'Minimum Replicas': 0,
                        'Maximum Replicas': 5,
                        'Initial Replicas': 2,
                        Metric: 'throughput',
                        'Metric Threshold': 6,
                        'Metric Threshold Percentage': 70,
                        'Scale Down Delay Seconds': 8,
                        'Scale To Zero Retention Seconds': 9,
                        'Hard Limit': 0
                    };

                    supportedFrom(SupportedVersions.V2_23)(() => {
                        expectedValues['Initialization Timeout Seconds'] = 300;
                        expectedValues['Request Timeout Seconds'] = 60;
                    });

                    validateInferenceWorkload(workloadName, expectedValues);
                });

                await test.step('Validate Knative service configuration', () => {
                    const revisionOutput = runaiCliV2.runKubectlCommand(
                        `get ksvc ${workloadName} -n ${project.status.namespace} -o jsonpath='{.status.latestCreatedRevisionName}'`
                    );
                    const revisionName = revisionOutput.stdout.trim();
                    expect(revisionName).not.toBe('');

                    supportedFrom(SupportedVersions.V2_23)(() => {
                        const revisionConfig = runaiCliV2.runKubectlCommand(
                            `get revision ${revisionName} -n ${project.status.namespace} -o jsonpath='{.spec.timeoutSeconds}'`
                        );
                        const requestTimeout = parseInt(revisionConfig.stdout.trim());
                        expect(requestTimeout).toBe(60);

                        const serviceConfig = runaiCliV2.runKubectlCommand(
                            `get ksvc ${workloadName} -n ${project.status.namespace} -o jsonpath='{.spec.template.spec.timeoutSeconds}'`
                        );
                        const serviceTimeout = parseInt(serviceConfig.stdout.trim());
                        expect(serviceTimeout).toBe(60);

                        const initTimeoutConfig = runaiCliV2.runKubectlCommand(
                            `get ksvc ${workloadName} -n ${project.status.namespace} -o jsonpath='{.spec.template.metadata.annotations.serving\\.knative\\.dev/progress-deadline}'`
                        );

                        expect(initTimeoutConfig.stdout.trim()).toBe('300s');
                    });
                });

                await test.step('Delete inference workload and verify count', async () => {
                    await waitForWorkloadCount(1);

                    runaiCliV2.delete(undefined, workloadName, cliV2WorkloadType.inference);

                    await waitForWorkloadCount(0);

                    const stdoutCustom = runaiCliV2.list(cliV2WorkloadType.inference, listFlagsParams).stdout;
                    const result = JSON.parse(stdoutCustom).workloads;
                    expect(result === undefined || result.length === 0).toBeTruthy();
                });
            });
        });
    }
);
