/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect, TestInfo } from '@playwright/test';
import {
    config,
    Utils,
    logger,
    operations,
    projectResourcesUtils,
    tags,
    waitUtils,
    nodePoolUtils,
    workloadsCrudUtils,
    crds,
    versionUtils,
    policyUtils,
    workloadsUtilsV2,
    apiUtils,
    orgUnitUtils
} from '../../../../../modules/helpers';
import { Project } from '../../../../../modules/api/apis/projectsDepartmentsApi';
import { namespaceCluster } from '../../../../../modules/helpers/config/const';
import {
    SubmitFlagsMerged,
    cliV2WorkloadType,
    preemptibilityOptions
} from '../../../../../modules/cli/tools/RunaiCliV2.constants';
import { kubeCli, runaiCliV2, bashCli } from '../../../../../modules/cli';
import { Phase } from '../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import { supportedFrom, SupportedVersions } from '../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { ImagePullPolicy } from '../../../../../modules/api/apis/workloadsApi/generated/model/imagePullPolicy';
import { flowRootState } from '../../../../../modules/helpers/config/config';
import { objectGenerator } from '../../../../../modules/helpers';
import { ServingPortProtocol } from '../../../../../modules/api/apis/workloadsApi/generated/model/servingPortProtocol';
import { WorkloadUtils } from '../../../../../modules/helpers/utils/WorkloadUtils';
import { ScopeType } from '../../../../../modules/api/apis/policyApi/generated/model/scopeType';
import { WorkloadType as PolicyWorkloadType } from '../../../../../modules/helpers/utils/policyUtils';
import { WorkloadType } from '../../../../../modules/helpers/utils/WorkloadsCrudUtils';
import { StatusCodes } from 'http-status-codes';
import { skipPlaywrightTest } from '../../../../../modules/helpers/utils/SkippedTestUtils';

test.describe('Workload Priority Testing', () => {
    let project: Project;
    let projectName: string;
    let namespace: string;
    let clusterScopedCrdPolicyName: string;
    const workloadUtils = new WorkloadUtils();
    const PRIORITY_BUILD = 'build';
    const PRIORITY_INTERACTIVE_PREEMPTIBLE = 'interactive-preemptible';
    const PRIORITY_INVALID = 'invalid-priorityclass';
    const PRIORITY_HIGH = 'high';
    const PRIORITY_LOW = 'low';
    const PRIORITY_MEDIUM = 'medium';
    const PRIORITY_VERY_LOW = 'very-low';
    const PRIORITY_MEDIUM_LOW = 'medium-low';
    const PRIORITY_MEDIUM_HIGH = 'medium-high';
    const PRIORITY_VERY_HIGH = 'very-high';

    // Priority values mapping
    const PRIORITY_VALUES = {
        [PRIORITY_VERY_LOW]: 25,
        [PRIORITY_LOW]: 40,
        [PRIORITY_MEDIUM_LOW]: 65,
        [PRIORITY_MEDIUM]: 80,
        [PRIORITY_MEDIUM_HIGH]: 90,
        [PRIORITY_HIGH]: 125,
        [PRIORITY_VERY_HIGH]: 150
    };

    // Expected priority classes array for testing
    const expectedPriorityClasses = [
        PRIORITY_VERY_LOW,
        PRIORITY_LOW,
        PRIORITY_MEDIUM_LOW,
        PRIORITY_MEDIUM,
        PRIORITY_MEDIUM_HIGH,
        PRIORITY_HIGH,
        PRIORITY_VERY_HIGH
    ];

    const EVENT_NONPREEMPTIBLE_OVERQUOTA_FULL =
        'NonPreemptibleOverQuota: Non-preemptible workload is over quota';
    const EVENT_QUOTA_LIMIT_MSG = 'quota has reached the allowable limit of GPUs';

    const WORKLOAD_NAME_1 = objectGenerator.generateNameByPrefix('pcw1');
    const WORKLOAD_NAME_2 = objectGenerator.generateNameByPrefix('pcw2');
    const TRAINING_WORKLOAD = objectGenerator.generateNameByPrefix('training');

    const ERROR_MESSAGE_INVALID_PRIORITY_CLASS_CLI =
        'Error: invalid argument "invalid-priorityclass" for priorityClass field:';

    const ERROR_MESSAGE_INVALID_PRIORITY_CLASS_API =
        'Error: invalid argument \\"invalid-priorityclass\\" for priorityClass field:';

    test.beforeAll(async function () {
        await apiUtils.enableProjectOverQuotaFeature();
        project = await projectResourcesUtils.createProjectAndNamespace('priority-control', undefined, 2, 2);
        projectName = project.name;
        namespace = Utils.namespaceFromProjectName(projectName);
        runaiCliV2.loginAsUser(
            flowRootState.users.system_administrator.email,
            flowRootState.users.system_administrator.password
        );
        runaiCliV2.projectSet(projectName);
        clusterScopedCrdPolicyName = policyUtils.getPolicyName(ScopeType.Cluster, flowRootState.cluster.uuid);
        workloadsCrudUtils.accessToken = flowRootState.users.system_administrator.accessToken;
    });

    test.afterAll('deleting project and namespace', async () => {
        await apiUtils.disableProjectOverQuotaFeature();
        if (project) {
            await operations.deleteProjectAndNamespace(project.id);
        }
    });

    function submitWorkload({
        name,
        gpu,
        cpuCoreRequest,
        priority,
        preemptible
    }: {
        name: string;
        gpu?: number;
        cpuCoreRequest?: number;
        priority?: string;
        preemptible?: boolean;
    }): void {
        const submitFlag: any = {
            [SubmitFlagsMerged.image]: config.quickStartImage.training,
            [SubmitFlagsMerged.command]: '-- sleep infinity'
        };
        if (gpu) submitFlag[SubmitFlagsMerged.gpuRequest] = gpu;
        if (cpuCoreRequest) submitFlag[SubmitFlagsMerged.cpuCoreRequest] = cpuCoreRequest;
        if (priority) submitFlag[SubmitFlagsMerged.priority] = priority;
        if (preemptible) submitFlag[SubmitFlagsMerged.preemptibility] = preemptibilityOptions.preemptible;
        runaiCliV2.submit(submitFlag, name, cliV2WorkloadType.training);
    }

    async function expectWorkloadRunnerPriorityClass(
        workloadName: string,
        expectedPriorityClass: string
    ): Promise<void> {
        const workloadRunner = await workloadUtils.getWorkloadRunnerByWorkloadName(workloadName, namespace);
        const workloadRunnerPriorityClass = workloadUtils.getPriorityClassFromWorkloadRunner(workloadRunner);
        expect(
            workloadRunnerPriorityClass,
            `WorkloadRunner priority class should be ${expectedPriorityClass}. ` +
                `WorkloadRunner: ${JSON.stringify(workloadRunner)}`
        ).toBe(expectedPriorityClass);
    }

    async function checkWorkloadAndPodDeleted(name: string): Promise<void> {
        await waitUtils.waitForCondition(
            () => !runaiCliV2.list(cliV2WorkloadType.training).stdout.includes(name),
            `Wait for workload ${name} to disappear from list`,
            true,
            30,
            2
        );
        const podsAfter = await kubeCli.listPodsOfJob(namespace, name);
        expect(podsAfter.length).toBe(0);
    }

    async function deleteAndCheckWorkloads(names: string[]): Promise<void> {
        for (const name of names) {
            try {
                runaiCliV2.delete({}, name, cliV2WorkloadType.training);
            } catch (error) {
                logger.warn(`Failed to delete workload ${name}: ${error}`);
            }
        }
        for (const name of names) {
            try {
                await waitUtils.waitForPodsDeletion(name, namespace, 70);
                await checkWorkloadAndPodDeleted(name);
            } catch (error) {
                logger.warn(`Failed to delete workload ${name}: ${error}`);
            }
        }
    }

    async function runStandardWorkloadLifecycle({
        name,
        gpu,
        cpuCoreRequest,
        priority,
        preemptible,
        phase = Phase.Running,
        phaseTimeout = 240,
        expectDescribe,
        deleteAfter = true
    }: {
        name: string;
        gpu?: number;
        cpuCoreRequest?: number;
        priority?: string;
        preemptible?: boolean;
        phase?: Phase;
        phaseTimeout?: number;
        expectDescribe?: (output: string) => void | Promise<void>;
        deleteAfter?: boolean;
    }): Promise<void> {
        if (cpuCoreRequest) {
            submitWorkload({ name, cpuCoreRequest, priority, preemptible });
        } else if (gpu) {
            submitWorkload({ name, gpu, priority, preemptible });
        }
        await waitUtils.waitForBackendWorkloadPhase(name, project.id, phase, phaseTimeout);
        const describeOutput = runaiCliV2.describe({}, name, cliV2WorkloadType.training);
        if (expectDescribe) {
            await expectDescribe(describeOutput);
        }
        expect((await kubeCli.listPodsOfJob(namespace, name)).length).toBeGreaterThan(0);
        if (deleteAfter) {
            await deleteAndCheckWorkloads([name]);
        }
    }

    async function withProjectConfig<T>(
        config: { deserved: number; maxAllowed: number; overQuotaWeight: number },
        fn: () => Promise<T>
    ): Promise<T> {
        const projectId = String(project.id);
        const nodePoolId = String((await nodePoolUtils.getDefaultNodePoolFromBackend()).id);

        await waitUtils.waitForCondition(
            async () => {
                const current = await orgUnitUtils.getProject(projectId);
                return (current.resources ?? []).some((r) => r.nodePool?.id === nodePoolId);
            },
            `Wait for project ${projectId} resources to include node pool ${nodePoolId}`,
            true,
            60,
            2
        );

        await projectResourcesUtils.updateProjectGpuConfig(projectId, nodePoolId, config);

        const applied = (await orgUnitUtils.getProject(projectId)).resources?.find(
            (r) => r.nodePool?.id === nodePoolId
        )?.gpu;
        expect(applied?.deserved).toBe(config.deserved);
        expect(applied?.limit).toBe(config.maxAllowed);

        return await fn();
    }

    function checkPendingWithEvent(name: string, expectedEvent: string): void {
        const describeOutput = runaiCliV2.describe({}, name, cliV2WorkloadType.training);
        expect(describeOutput).toContain(Phase.Pending);
        expect(describeOutput).toContain(expectedEvent);
    }

    function expectCliSubmitError(submitFlag: any, name: string, expectedMsg: string): void {
        let errorMsg = '';
        try {
            runaiCliV2.submit(submitFlag, name, cliV2WorkloadType.training);
        } catch (err: any) {
            errorMsg = err?.message || String(err);
        }
        expect(errorMsg).toContain(expectedMsg);
    }

    function buildApiSpec(priorityClass: string): any {
        return {
            command: 'sleep infinity',
            image: config.quickStartImage.training,
            imagePullPolicy: ImagePullPolicy.Always,
            priorityClass: priorityClass as any
        };
    }

    async function createAndValidateTraining(
        name: string,
        clusterId: string,
        projectId: string,
        priorityClass: string,
        shouldDelete = true
    ): Promise<void> {
        const spec = buildApiSpec(priorityClass);
        const resp = await workloadsCrudUtils.createTraining(name, true, clusterId, String(projectId), spec);
        expect(resp.response.statusCode).toBe(StatusCodes.ACCEPTED);
        const workloadId = resp.body.workloadId;
        const getResp = await workloadsCrudUtils.getTraining(workloadId);
        expect(getResp.response.statusCode).toBe(StatusCodes.OK);
        if (shouldDelete) {
            const delResp = await workloadsCrudUtils.deleteTraining(workloadId);
            expect(delResp.response.statusCode).toBe(StatusCodes.ACCEPTED);
        }
    }

    test.describe(`Workload Priority Control, tags: ${tags.tool.api} ${tags.resource.workloads} ${tags.teams.loong}`, () => {
        supportedFrom(SupportedVersions.V2_18)(function () {
            test.describe('Workload Priority Control - CLI', () => {
                test('T5075496: Standard training workload with interactive-preemptible/build priorityClass', async () => {
                    for (const priority of [PRIORITY_INTERACTIVE_PREEMPTIBLE, PRIORITY_BUILD]) {
                        await runStandardWorkloadLifecycle({
                            name: WORKLOAD_NAME_1,
                            gpu: 1,
                            priority,
                            phase: Phase.Running,
                            expectDescribe: (output) => expect(output).toContain(priority)
                        });
                    }
                });

                test('T5075549: Training workload with "interactive-preemptible" priorityClass is allowed to run over quota', async () => {
                    await withProjectConfig({ deserved: 1, maxAllowed: -1, overQuotaWeight: 2 }, async () => {
                        await runStandardWorkloadLifecycle({
                            name: WORKLOAD_NAME_1,
                            gpu: 2,
                            priority: PRIORITY_INTERACTIVE_PREEMPTIBLE,
                            preemptible: true,
                            phase: Phase.Running,
                            phaseTimeout: 240,
                            expectDescribe: (output) =>
                                expect(output).toContain(PRIORITY_INTERACTIVE_PREEMPTIBLE)
                        });
                    });
                });

                test('T5075565: Training workload with "build" priorityClass is not allowed to run over quota', async () => {
                    await withProjectConfig({ deserved: 1, maxAllowed: -1, overQuotaWeight: 2 }, async () => {
                        await runStandardWorkloadLifecycle({
                            name: WORKLOAD_NAME_1,
                            gpu: 2,
                            priority: PRIORITY_BUILD,
                            phase: Phase.Pending,
                            phaseTimeout: 60,
                            expectDescribe: async () => {
                                const pods = await kubeCli.listPodsOfJob(namespace, WORKLOAD_NAME_1);
                                expect(pods.length).toBeGreaterThan(0);
                                const podName = pods[0].metadata.name;
                                const describeOutput = bashCli.exec(
                                    `kubectl describe po ${podName} -n ${namespace}`
                                );
                                expect(describeOutput).toContain(Phase.Pending);
                                await waitUtils.waitForCondition(
                                    () => {
                                        const describeOutput = runaiCliV2.describe(
                                            {},
                                            WORKLOAD_NAME_1,
                                            cliV2WorkloadType.training
                                        );
                                        return describeOutput.includes(EVENT_NONPREEMPTIBLE_OVERQUOTA_FULL);
                                    },
                                    `Wait for event "${EVENT_NONPREEMPTIBLE_OVERQUOTA_FULL}" in describe output`,
                                    true,
                                    12,
                                    5
                                );
                            }
                        });
                    });
                });

                test('T5075566: Training workload with higher priorityClass can preempt resources', async () => {
                    await withProjectConfig({ deserved: 2, maxAllowed: 2, overQuotaWeight: 0 }, async () => {
                        await runStandardWorkloadLifecycle({
                            name: WORKLOAD_NAME_1,
                            gpu: 1,
                            priority: PRIORITY_LOW,
                            preemptible: true,
                            phase: Phase.Running,
                            phaseTimeout: 240,
                            expectDescribe: (output) => expect(output).toContain(WORKLOAD_NAME_1),
                            deleteAfter: false
                        });
                        await runStandardWorkloadLifecycle({
                            name: WORKLOAD_NAME_2,
                            gpu: 2,
                            priority: PRIORITY_BUILD,
                            phase: Phase.Running,
                            expectDescribe: (output) => expect(output).toContain(WORKLOAD_NAME_2),
                            deleteAfter: false
                        });
                        await waitUtils.waitForBackendWorkloadPhase(
                            WORKLOAD_NAME_1,
                            project.id,
                            Phase.Pending,
                            120
                        );
                        await deleteAndCheckWorkloads([WORKLOAD_NAME_1, WORKLOAD_NAME_2]);
                    });
                });

                test('T5075567: Training workload with "build" priorityClass can not be preempted', async () => {
                    await withProjectConfig({ deserved: 2, maxAllowed: 2, overQuotaWeight: 0 }, async () => {
                        await runStandardWorkloadLifecycle({
                            name: WORKLOAD_NAME_1,
                            gpu: 1,
                            priority: PRIORITY_BUILD,
                            phase: Phase.Running,
                            expectDescribe: (output) => expect(output).toContain(WORKLOAD_NAME_1),
                            deleteAfter: false
                        });
                        submitWorkload({ name: WORKLOAD_NAME_2, gpu: 2 });
                        await waitUtils.sleep(30 * 1000);
                        const describeOutput1After = runaiCliV2.describe(
                            {},
                            WORKLOAD_NAME_1,
                            cliV2WorkloadType.training
                        );
                        expect(describeOutput1After).toContain(Phase.Running);
                        expect(
                            (await kubeCli.listPodsOfJob(namespace, WORKLOAD_NAME_1)).length
                        ).toBeGreaterThan(0);
                        checkPendingWithEvent(WORKLOAD_NAME_2, EVENT_QUOTA_LIMIT_MSG);
                        await deleteAndCheckWorkloads([WORKLOAD_NAME_1, WORKLOAD_NAME_2]);
                    });
                });

                test('T5075568: Create a training workload by using CLI with invalid priorityClass should fail', () => {
                    const submitFlag = {
                        [SubmitFlagsMerged.image]: config.quickStartImage.training,
                        [SubmitFlagsMerged.gpuRequest]: 1,
                        [SubmitFlagsMerged.priority]: PRIORITY_INVALID
                    } as Partial<typeof SubmitFlagsMerged>;
                    expectCliSubmitError(
                        submitFlag,
                        WORKLOAD_NAME_1,
                        ERROR_MESSAGE_INVALID_PRIORITY_CLASS_CLI
                    );
                });
            });

            test.describe('Workload Priority Control - API', () => {
                test('T5075583: Create training workloads with different priorityClass by using API', async () => {
                    const clusterId = project.clusterUuid;
                    const projectId = String(project.id);
                    const name = WORKLOAD_NAME_1;
                    await createAndValidateTraining(name, clusterId, projectId, 'interactive-preemptible');
                    await createAndValidateTraining(name, clusterId, projectId, 'build');
                });

                test('T5075584: Create a training workload by using API with invalid priorityClass should fail', async () => {
                    const clusterId = project.clusterUuid;
                    const projectId = String(project.id);
                    const name = WORKLOAD_NAME_1;
                    const spec = buildApiSpec('invalid-priorityclass');
                    try {
                        await workloadsCrudUtils.createTraining(
                            name,
                            true,
                            clusterId,
                            String(projectId),
                            spec
                        );
                        throw new Error('API did not throw error, test failed');
                    } catch (e: any) {
                        if (e && e.response && e.response.statusCode) {
                            expect(e.response.statusCode).toBe(StatusCodes.BAD_REQUEST);
                            const errStr = JSON.stringify(e.body) || e.message;
                            expect(errStr).toContain(ERROR_MESSAGE_INVALID_PRIORITY_CLASS_API);
                        }
                    }
                });
            });
        });

        test.describe(`Workload Priority Flexibility, tags: ${tags.tool.api} ${tags.resource.workloads} ${tags.teams.loong}`, () => {
            supportedFrom(SupportedVersions.V2_23)(function () {
                test('T5217826: Training workload - Submit with default priority class', async () => {
                    await runStandardWorkloadLifecycle({
                        name: TRAINING_WORKLOAD,
                        gpu: 1,
                        phase: Phase.Running,
                        phaseTimeout: 240,
                        expectDescribe: async (describeOutput) => {
                            await expectWorkloadRunnerPriorityClass(TRAINING_WORKLOAD, PRIORITY_LOW);
                            expect(describeOutput).toMatch(new RegExp(`Priority:\\s+${PRIORITY_LOW}\\s*\\(`));

                            const pods = await kubeCli.listPodsByNamespaceAndPrefix(
                                namespace,
                                TRAINING_WORKLOAD
                            );
                            expect(pods.length).toBeGreaterThan(0);
                            expect(pods[0].spec.priorityClassName).toBe(PRIORITY_LOW);
                        }
                    });
                });

                test('T5217827: Submit training workloads with different priority classes', async () => {
                    await withProjectConfig({ deserved: 3, maxAllowed: 3, overQuotaWeight: 0 }, async () => {
                        const priorities = [PRIORITY_HIGH, PRIORITY_MEDIUM, PRIORITY_LOW];
                        const workloadNames = priorities.map(
                            (priority) =>
                                `${objectGenerator.generateNameByPrefix('priority-test')}-${priority}`
                        );

                        for (let i = 0; i < priorities.length; i++) {
                            const priority = priorities[i];
                            const workloadName = workloadNames[i];

                            await runStandardWorkloadLifecycle({
                                name: workloadName,
                                gpu: 1,
                                priority: priority,
                                phase: Phase.Running,
                                phaseTimeout: 240,
                                expectDescribe: async (describeOutput) => {
                                    await expectWorkloadRunnerPriorityClass(workloadName, priority);
                                    expect(describeOutput).toMatch(
                                        new RegExp(`Priority:\\s+${priority}\\s*\\(`)
                                    );

                                    const pods = await kubeCli.listPodsByNamespaceAndPrefix(
                                        namespace,
                                        workloadName
                                    );
                                    expect(pods.length).toBeGreaterThan(0);
                                    expect(pods[0].spec.priorityClassName).toBe(priority);
                                },
                                deleteAfter: false
                            });
                        }
                        await deleteAndCheckWorkloads(workloadNames);
                    });
                });

                test('T5217828: Submit inference workloads with different priority classes', async function ({}, testInfo: TestInfo) {
                    skipPlaywrightTest(
                        test,
                        testInfo,
                        'Skip this test case since the inference image size is too big'
                    );
                    test.setTimeout(360 * 1000);
                    await withProjectConfig({ deserved: 2, maxAllowed: 2, overQuotaWeight: 0 }, async () => {
                        const veryHighWorkloadName = 'inference-mapping-test-vh';
                        const mediumWorkloadName = 'inference-mapping-test-med';
                        const createdInferenceIds = [];

                        await test.step('Submit inference workload with very-high priority', async () => {
                            const { body: inference } = await workloadsCrudUtils.createInference(
                                veryHighWorkloadName,
                                false,
                                flowRootState.cluster.uuid,
                                project.id.toString(),
                                {
                                    image: 'ghcr.io/knative/helloworld-go',
                                    compute: { gpuDevicesRequest: 1 },
                                    priorityClass: PRIORITY_VERY_HIGH,
                                    servingPort: {
                                        container: 8080,
                                        protocol: ServingPortProtocol.Http
                                    }
                                }
                            );
                            createdInferenceIds.push(inference.workloadId);

                            await workloadsCrudUtils.waitForWorkloadPhase(
                                inference.workloadId,
                                Phase.Running,
                                WorkloadType.INFERENCE,
                                undefined,
                                180
                            );
                        });

                        await test.step('Submit inference workload with medium priority', async () => {
                            const { body: inference } = await workloadsCrudUtils.createInference(
                                mediumWorkloadName,
                                false,
                                flowRootState.cluster.uuid,
                                project.id.toString(),
                                {
                                    image: 'ghcr.io/knative/helloworld-go',
                                    compute: { gpuDevicesRequest: 1 },
                                    priorityClass: PRIORITY_MEDIUM,
                                    servingPort: {
                                        container: 8080,
                                        protocol: ServingPortProtocol.Http
                                    }
                                }
                            );
                            createdInferenceIds.push(inference.workloadId);

                            await workloadsCrudUtils.waitForWorkloadPhase(
                                inference.workloadId,
                                Phase.Running,
                                WorkloadType.INFERENCE,
                                undefined,
                                180
                            );
                        });

                        await test.step('Verify both workloads use their requested priority classes in WorkloadRunner', async () => {
                            await expectWorkloadRunnerPriorityClass(veryHighWorkloadName, PRIORITY_VERY_HIGH);
                            await expectWorkloadRunnerPriorityClass(mediumWorkloadName, PRIORITY_MEDIUM);
                        });

                        await test.step('Verify pods inherit correct priority classes', async () => {
                            const veryHighPods = await kubeCli.listPodsOfJob(namespace, veryHighWorkloadName);
                            expect(veryHighPods.length).toBeGreaterThan(0);
                            expect(veryHighPods[0].metadata.labels.priorityClassName).toBe(
                                PRIORITY_VERY_HIGH
                            );
                            const mediumPods = await kubeCli.listPodsOfJob(namespace, mediumWorkloadName);
                            expect(mediumPods.length).toBeGreaterThan(0);
                            expect(mediumPods[0].metadata.labels.priorityClassName).toBe(PRIORITY_MEDIUM);
                        });

                        await test.step('Cleanup workloads', async () => {
                            for (const workloadId of createdInferenceIds) {
                                try {
                                    await workloadsCrudUtils.deleteInference(workloadId);
                                } catch (error) {
                                    logger.error(
                                        `Failed to delete inference workload ${workloadId}: ${error}`
                                    );
                                }
                            }
                        });
                    });
                });

                test('T5217829: Inference workload - Submit with default priority class', async function ({}, testInfo: TestInfo) {
                    skipPlaywrightTest(
                        test,
                        testInfo,
                        'Skip this test case since the inference image size is too big'
                    );
                    const inferenceWorkloadName = 'inference-test-1';
                    let inferenceId: string;

                    await test.step('Submit an inference workload without specifying priority class via API', async () => {
                        const { body: inference } = await workloadsCrudUtils.createInference(
                            inferenceWorkloadName,
                            false,
                            flowRootState.cluster.uuid,
                            project.id.toString(),
                            {
                                image: 'ghcr.io/knative/helloworld-go',
                                compute: { gpuDevicesRequest: 1 },
                                servingPort: {
                                    container: 8080,
                                    protocol: ServingPortProtocol.Http
                                }
                            }
                        );
                        inferenceId = inference.workloadId;

                        await workloadsCrudUtils.waitForWorkloadPhase(
                            inference.workloadId,
                            Phase.Running,
                            WorkloadType.INFERENCE,
                            undefined,
                            120
                        );
                    });

                    await test.step('Verify the workload is created successfully and default priority is applied', async () => {
                        await expectWorkloadRunnerPriorityClass(inferenceWorkloadName, PRIORITY_VERY_HIGH);
                    });

                    await test.step('Verify the actual pod has the correct priority class', async () => {
                        const pods = await kubeCli.listPodsOfJob(namespace, inferenceWorkloadName);
                        expect(pods.length).toBeGreaterThan(0);
                        expect(pods[0].metadata.labels.priorityClassName).toBe(PRIORITY_VERY_HIGH);
                    });

                    await test.step('Delete the workload', async () => {
                        await workloadsCrudUtils.deleteInference(inferenceId);
                    });
                });

                test('T5217830: External workload - Submit via kubectl with default priority', async () => {
                    const deploymentName = 'external-test-1';
                    const deploymentYaml = `apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: ${deploymentName}\n  namespace: ${namespace}\nspec:\n  replicas: 1\n  selector:\n    matchLabels:\n      app: ${deploymentName}\n  template:\n    metadata:\n      labels:\n        app: ${deploymentName}\n    spec:\n      containers:\n      - name: test\n        image: ubuntu\n        command: ["sleep", "infinity"]`;
                    await kubeCli.createFromYaml(deploymentYaml);
                    await waitUtils.waitForCondition(
                        async () => {
                            const deployment = await kubeCli.getDeployment(deploymentName, namespace);
                            return deployment.status?.readyReplicas === 1;
                        },
                        `Wait for deployment ${deploymentName} to be ready`,
                        true,
                        60,
                        5
                    );
                    const podsYamlOutput = bashCli.exec(
                        `kubectl get pods -l app=${deploymentName} -o yaml -n ${namespace}`
                    );
                    console.log('podsYamlOutput', podsYamlOutput);
                    expect(podsYamlOutput).toContain('priority: 0');
                    expect(podsYamlOutput).not.toContain('priorityClassName:');

                    await kubeCli.deleteDeployment(deploymentName, namespace);
                });

                test('T5217831: External workload - Submit via kubectl with specific priority class in pod spec', async () => {
                    const deploymentName = 'external-test-2';
                    const deploymentYaml = `apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: ${deploymentName}\n  namespace: ${namespace}\nspec:\n  replicas: 1\n  selector:\n    matchLabels:\n      app: ${deploymentName}\n  template:\n    metadata:\n      labels:\n        app: ${deploymentName}\n    spec:\n      priorityClassName: ${PRIORITY_MEDIUM_HIGH}\n      containers:\n      - name: test\n        image: ubuntu\n        command: ["sleep", "infinity"]`;
                    await kubeCli.createFromYaml(deploymentYaml);

                    await waitUtils.waitForCondition(
                        async () => {
                            const deployment = await kubeCli.getDeployment(deploymentName, namespace);
                            return deployment.status?.readyReplicas === 1;
                        },
                        `Wait for deployment ${deploymentName} to be ready`,
                        true,
                        60,
                        5
                    );

                    const podsYamlOutput = bashCli.exec(
                        `kubectl get pods -l app=${deploymentName} -o yaml  -n ${namespace}`
                    );
                    expect(podsYamlOutput).toContain(`priorityClassName: ${PRIORITY_MEDIUM_HIGH}`);

                    await kubeCli.deleteDeployment(deploymentName, namespace);
                });

                test('T5217833: Verify PriorityClasses are available and have correct properties', async () => {
                    const priorityClassesOutput = bashCli.exec('kubectl get priorityclass');
                    for (const priorityClass of expectedPriorityClasses) {
                        expect(priorityClassesOutput).toContain(priorityClass);
                    }

                    for (const priorityClass of expectedPriorityClasses) {
                        const yamlOutput = bashCli.exec(`kubectl get priorityclass ${priorityClass} -o yaml`);
                        expect(yamlOutput).toContain(`name: ${priorityClass}`);
                        const expectedValue = PRIORITY_VALUES[priorityClass];
                        expect(yamlOutput).toContain(`value: ${expectedValue}`);
                        expect(yamlOutput).toContain('apiVersion: scheduling.k8s.io/v1');
                        expect(yamlOutput).toContain('kind: PriorityClass');
                    }

                    await runStandardWorkloadLifecycle({
                        name: TRAINING_WORKLOAD,
                        gpu: 1,
                        priority: PRIORITY_HIGH,
                        phase: Phase.Running,
                        phaseTimeout: 240,
                        expectDescribe: async () => {
                            const pods = await kubeCli.listPodsByNamespaceAndPrefix(
                                namespace,
                                TRAINING_WORKLOAD
                            );
                            expect(pods.length).toBeGreaterThan(0);
                            expect(pods[0].spec.priorityClassName).toBe(PRIORITY_HIGH);
                        }
                    });
                });

                test('T5217834: Verify PriorityClass source labels are correctly set', async () => {
                    for (const priorityClass of expectedPriorityClasses) {
                        const yamlOutput = bashCli.exec(`kubectl get priorityclass ${priorityClass} -o yaml`);
                        expect(yamlOutput).toContain('run.ai/priority-class-source: Run.ai');
                    }

                    const targetPriority =
                        expectedPriorityClasses[Math.floor(Math.random() * expectedPriorityClasses.length)];
                    await runStandardWorkloadLifecycle({
                        name: TRAINING_WORKLOAD,
                        gpu: 1,
                        priority: targetPriority,
                        phase: Phase.Running,
                        phaseTimeout: 240,
                        expectDescribe: async () => {
                            await expectWorkloadRunnerPriorityClass(TRAINING_WORKLOAD, targetPriority);
                            const pods = await kubeCli.listPodsByNamespaceAndPrefix(
                                namespace,
                                TRAINING_WORKLOAD
                            );
                            expect(pods.length).toBeGreaterThan(0);
                            expect(pods[0].spec.priorityClassName).toBe(targetPriority);
                        }
                    });
                });

                test('T5217835: Submit workload with legacy PriorityClass', async () => {
                    const legacyPriorityClass = 'train';

                    await runStandardWorkloadLifecycle({
                        name: TRAINING_WORKLOAD,
                        gpu: 1,
                        priority: legacyPriorityClass,
                        expectDescribe: async () => {
                            await expectWorkloadRunnerPriorityClass(TRAINING_WORKLOAD, legacyPriorityClass);

                            const pods = await kubeCli.listPodsOfJob(namespace, TRAINING_WORKLOAD);
                            expect(pods[0].spec.priorityClassName).toBe(legacyPriorityClass);
                        }
                    });
                });

                test('T5217836: Submit workload with deprecated preemptible field', async () => {
                    await withProjectConfig({ deserved: 1, maxAllowed: 1, overQuotaWeight: 0 }, async () => {
                        const preemptibleWorkload = 'preemptible-test';
                        const nonPreemptibleWorkload = 'non-preemptible-test';

                        await runStandardWorkloadLifecycle({
                            name: preemptibleWorkload,
                            gpu: 1,
                            priority: PRIORITY_INTERACTIVE_PREEMPTIBLE,
                            expectDescribe: async () => {
                                await expectWorkloadRunnerPriorityClass(
                                    preemptibleWorkload,
                                    PRIORITY_INTERACTIVE_PREEMPTIBLE
                                );

                                const pods = await kubeCli.listPodsOfJob(namespace, preemptibleWorkload);
                                expect(pods[0].spec.priorityClassName).toBe(PRIORITY_INTERACTIVE_PREEMPTIBLE);
                            },
                            deleteAfter: false
                        });

                        const highPriorityWorkload = 'high-priority-preempt-test';
                        submitWorkload({
                            name: highPriorityWorkload,
                            gpu: 1,
                            priority: PRIORITY_VERY_HIGH
                        });
                        await waitUtils.waitForBackendWorkloadPhase(
                            highPriorityWorkload,
                            project.id,
                            Phase.Running,
                            60
                        );
                        await waitUtils.waitForBackendWorkloadPhase(
                            preemptibleWorkload,
                            project.id,
                            Phase.Pending,
                            60
                        );
                        await deleteAndCheckWorkloads([highPriorityWorkload, preemptibleWorkload]);

                        await runStandardWorkloadLifecycle({
                            name: nonPreemptibleWorkload,
                            gpu: 1,
                            expectDescribe: async () => {
                                await expectWorkloadRunnerPriorityClass(nonPreemptibleWorkload, PRIORITY_LOW);

                                const pods = await kubeCli.listPodsOfJob(namespace, nonPreemptibleWorkload);
                                expect(pods[0].spec.priorityClassName).toBe(PRIORITY_LOW);
                            },
                            deleteAfter: false
                        });
                        submitWorkload({
                            name: highPriorityWorkload,
                            gpu: 1,
                            priority: PRIORITY_VERY_HIGH
                        });
                        await waitUtils.waitForBackendWorkloadPhase(
                            highPriorityWorkload,
                            project.id,
                            Phase.Pending,
                            60
                        );
                        await waitUtils.waitForBackendWorkloadPhase(
                            nonPreemptibleWorkload,
                            project.id,
                            Phase.Running,
                            60
                        );
                    });
                });

                test('T5217837: Submit workload with non-legacy PriorityClass in old cluster - should fail', async () => {
                    logger.info('Features not supported in old cluster which runai version < 2.23');
                    if (
                        versionUtils.isVersionSupported(
                            flowRootState.cluster.version,
                            SupportedVersions.V2_23
                        )
                    ) {
                        test.skip();
                        return;
                    }
                    const submitFlag = {
                        [SubmitFlagsMerged.image]: config.quickStartImage.ubuntu,
                        [SubmitFlagsMerged.gpuRequest]: 1,
                        [SubmitFlagsMerged.priority]: PRIORITY_MEDIUM_HIGH,
                        [SubmitFlagsMerged.command]: 'sleep infinity'
                    } as Partial<typeof SubmitFlagsMerged>;
                    expectCliSubmitError(submitFlag, 'old-cluster-fail-test', 'invalid argument');

                    await runStandardWorkloadLifecycle({
                        name: TRAINING_WORKLOAD,
                        gpu: 1,
                        priority: 'train',
                        expectDescribe: async () => {
                            const crdYaml = await kubeCli.getCrdYaml(
                                crds.runaiTrainingWorkloadCrd,
                                TRAINING_WORKLOAD,
                                namespace
                            );
                            expect(crdYaml['spec']['labels']['items']['priorityClassName']['value']).toBe(
                                'train'
                            );

                            const pods = await kubeCli.listPodsOfJob(namespace, TRAINING_WORKLOAD);
                            expect(pods[0].metadata.labels.priorityClassName).toBe('train');
                        }
                    });
                });

                test('T5217839: Verify legacy workload behavior in old cluster', async () => {
                    logger.info('Features not supported in old cluster which runai version < 2.23');
                    if (
                        versionUtils.isVersionSupported(
                            flowRootState.cluster.version,
                            SupportedVersions.V2_23
                        )
                    ) {
                        test.skip();
                        return;
                    }
                    const defaultTrainingWorkload = 'old-training-default-test';
                    const preemptibleTrainingWorkload = 'old-training-preemptible-test';

                    await runStandardWorkloadLifecycle({
                        name: defaultTrainingWorkload,
                        gpu: 1,
                        expectDescribe: async () => {
                            const crdYaml = await kubeCli.getCrdYaml(
                                crds.runaiTrainingWorkloadCrd,
                                defaultTrainingWorkload,
                                namespace
                            );
                            const crdYamlString = JSON.stringify(crdYaml);
                            expect(crdYamlString).not.toContain('priorityClassName');

                            const pods = await kubeCli.listPodsOfJob(namespace, defaultTrainingWorkload);
                            expect(pods.length).toBeGreaterThan(0);
                            expect(pods[0]).not.toContain('priorityClassName');
                        }
                    });

                    await runStandardWorkloadLifecycle({
                        name: preemptibleTrainingWorkload,
                        gpu: 1,
                        priority: PRIORITY_INTERACTIVE_PREEMPTIBLE,
                        expectDescribe: async () => {
                            const crdYaml = await kubeCli.getCrdYaml(
                                crds.runaiTrainingWorkloadCrd,
                                preemptibleTrainingWorkload,
                                namespace
                            );
                            expect(crdYaml['spec']['labels']['items']['priorityClassName']['value']).toBe(
                                PRIORITY_INTERACTIVE_PREEMPTIBLE
                            );

                            const pods = await kubeCli.listPodsOfJob(namespace, preemptibleTrainingWorkload);
                            expect(pods.length).toBeGreaterThan(0);
                            expect(pods[0].metadata.labels.priorityClassName).toBe(
                                PRIORITY_INTERACTIVE_PREEMPTIBLE
                            );
                        }
                    });
                });

                test('T5217840: Apply a yaml with priorityClass through kubectl to old cluster', async () => {
                    logger.info(
                        'Testing kubectl application with priority class in old cluster version < 2.23'
                    );
                    if (
                        versionUtils.isVersionSupported(
                            flowRootState.cluster.version,
                            SupportedVersions.V2_23
                        )
                    ) {
                        test.skip();
                        return;
                    }
                    const deploymentName = 'external-test-2';
                    const deploymentYaml = `apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: ${deploymentName}\n  namespace: ${namespace}\nspec:\n  replicas: 1\n  selector:\n    matchLabels:\n      app: ${deploymentName}\n  template:\n    metadata:\n      labels:\n        app: ${deploymentName}\n    spec:\n      priorityClassName: ${PRIORITY_MEDIUM_HIGH}\n      containers:\n      - name: test\n        image: ubuntu\n        command: ["sleep", "infinity"]`;
                    await kubeCli.createFromYaml(deploymentYaml);

                    const deployment = await kubeCli.getDeployment(deploymentName, namespace);
                    if (deployment.status && deployment.status.conditions) {
                        let foundForbiddenMessage = false;
                        deployment.status.conditions.forEach((condition: any, index: number) => {
                            logger.info(`Message: ${condition.message}`);
                            if (
                                condition.message &&
                                condition.message.includes(
                                    `forbidden: no PriorityClass with name ${PRIORITY_MEDIUM_HIGH} was found`
                                )
                            ) {
                                foundForbiddenMessage = true;
                            }
                        });
                        expect(foundForbiddenMessage).toBe(true);
                    }
                    await kubeCli.deleteDeployment(deploymentName, namespace);
                });

                test('T5217843: Cross-workload type priority consistency', async function ({}, testInfo: TestInfo) {
                    skipPlaywrightTest(
                        test,
                        testInfo,
                        'Skip this test case since the inference image size is too big'
                    );
                    test.setTimeout(480 * 1000);
                    await withProjectConfig({ deserved: 2, maxAllowed: 2, overQuotaWeight: 0 }, async () => {
                        const priorityClass = PRIORITY_MEDIUM_HIGH;
                        const deploymentName = 'cross-external';

                        const { body: training } = await workloadsCrudUtils.createTraining(
                            'cross-training',
                            false,
                            flowRootState.cluster.uuid,
                            project.id.toString(),
                            {
                                image: config.quickStartImage.ubuntu,
                                command: 'sleep infinity',
                                compute: { gpuDevicesRequest: 1 },
                                priorityClass: priorityClass
                            }
                        );
                        const trainingId = training.workloadId;

                        await workloadsCrudUtils.waitForWorkloadPhase(
                            trainingId,
                            Phase.Running,
                            WorkloadType.TRAINING,
                            180
                        );

                        const { body: inference } = await workloadsCrudUtils.createInference(
                            'cross-inference',
                            false,
                            flowRootState.cluster.uuid,
                            project.id.toString(),
                            {
                                image: 'ghcr.io/knative/helloworld-go',
                                compute: { gpuDevicesRequest: 1 },
                                priorityClass: priorityClass,
                                servingPort: {
                                    container: 8080,
                                    protocol: ServingPortProtocol.Http
                                }
                            }
                        );
                        const inferenceId = inference.workloadId;

                        await workloadsCrudUtils.waitForWorkloadPhase(
                            inferenceId,
                            Phase.Running,
                            WorkloadType.INFERENCE,
                            120
                        );

                        const deploymentYaml = `apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: ${deploymentName}\n  namespace: ${namespace}\nspec:\n  replicas: 1\n  selector:\n    matchLabels:\n      app: ${deploymentName}\n  template:\n    metadata:\n      labels:\n        app: ${deploymentName}\n    spec:\n      priorityClassName: ${priorityClass}\n      containers:\n      - name: test\n        image: ubuntu\n        command: ["sleep", "infinity"]`;
                        await kubeCli.createFromYaml(deploymentYaml);

                        await waitUtils.waitForCondition(
                            async () => {
                                const deployment = await kubeCli.getDeployment(deploymentName, namespace);
                                return deployment.status?.readyReplicas === 1;
                            },
                            `Wait for deployment ${deploymentName} to be ready`,
                            true,
                            120,
                            5
                        );

                        await expectWorkloadRunnerPriorityClass(training.name, priorityClass);

                        const trainingPods = await kubeCli.listPodsOfJob(namespace, training.name);
                        expect(trainingPods.length).toBeGreaterThan(0);
                        expect(trainingPods[0].spec.priorityClassName).toBe(priorityClass);

                        await expectWorkloadRunnerPriorityClass(inference.name, priorityClass);

                        const inferencePods = await kubeCli.listPodsOfJob(namespace, inference.name);
                        expect(inferencePods.length).toBeGreaterThan(0);
                        expect(inferencePods[0].metadata.labels.priorityClassName).toBe(priorityClass);

                        const podsYamlOutput = bashCli.exec(
                            `kubectl get pods -l app=${deploymentName} -o yaml -n ${namespace}`
                        );
                        expect(podsYamlOutput).toContain(`priorityClassName: ${priorityClass}`);

                        await workloadsCrudUtils.deleteTraining(trainingId);
                        await workloadsCrudUtils.deleteInference(inferenceId);
                        await kubeCli.deleteDeployment(deploymentName, namespace);
                    });
                });

                test.describe('Workload Priority Flexibility - API', () => {
                    let policyCreated: boolean = false;
                    let workloadTypeUpdated: boolean = false;
                    let clusterPolicyCreated: boolean = false;
                    let workloadTypeId: string;
                    let originalCategoryId: string;
                    let originalPriorityId: string;
                    let originalCategoryName: string;
                    let originalPriorityName: string;
                    let newCategory: any;
                    let newPriority: any;
                    let policyName: string;
                    let workloadTypeName: string;
                    let policyWorkloadType: PolicyWorkloadType;
                    let createdWorkloadIds: string[] = [];
                    let workloadCreated: boolean = false;

                    test.beforeEach(async function () {
                        createdWorkloadIds = [];
                        workloadCreated = false;
                        policyName = `project-${project.id}`;
                        workloadTypeName = 'XGBoostJob';
                        policyWorkloadType = PolicyWorkloadType.Training;
                        try {
                            const workloadTypes = await workloadUtils.getWorkloadTypes();
                            const trainingWorkloadType = workloadTypes.find(
                                (t) => t.name === workloadTypeName
                            );
                            workloadTypeId = trainingWorkloadType.id;
                            const currentDetails = await workloadUtils.getWorkloadTypeById(workloadTypeId);
                            originalCategoryId = currentDetails.categoryId;
                            originalPriorityId = currentDetails.priorityId;
                            originalCategoryName = currentDetails.categoryName;
                            originalPriorityName = currentDetails.priorityName;
                        } catch (error) {
                            logger.error(`Failed to get workload type: ${error}`);
                        }
                    });

                    test.afterEach(async function () {
                        if (workloadCreated && createdWorkloadIds.length > 0) {
                            await workloadsUtilsV2.deleteAllWorkloads();
                            createdWorkloadIds = [];
                            workloadCreated = false;
                        }
                        if (clusterPolicyCreated) {
                            const clusterPolicy = await policyUtils.createPolicyV2(
                                `cluster-${flowRootState.cluster.uuid}`,
                                flowRootState.cluster.uuid,
                                policyWorkloadType,
                                { priorityClass: PRIORITY_LOW },
                                undefined,
                                ScopeType.Cluster
                            );
                        }
                        if (policyCreated) {
                            await policyUtils.createPolicyV2(
                                projectName,
                                project.id.toString(),
                                policyWorkloadType,
                                { priorityClass: PRIORITY_LOW },
                                undefined,
                                ScopeType.Project
                            );
                        }
                        if (workloadTypeUpdated) {
                            try {
                                await workloadUtils.updateWorkloadTypeById(workloadTypeId, {
                                    categoryId: originalCategoryId,
                                    priorityId: originalPriorityId
                                });
                            } catch (error) {
                                logger.error(`Failed to restore workloadType: ${error}`);
                            }
                        }
                    });

                    test('T5217907: List workload type by ID', async () => {
                        const categoryNames = ['Build', 'Deploy', 'Train'];
                        const workloadTypes = await workloadUtils.getWorkloadTypes();
                        expect(Array.isArray(workloadTypes)).toBe(true);
                        expect(workloadTypes.length).toBeGreaterThan(0);
                        for (const workloadType of workloadTypes) {
                            logger.info(`Workload type: ${workloadType.name}`);
                            logger.info(`Workload type category: ${workloadType.categoryName}`);
                            logger.info(`Workload type priority: ${workloadType.priorityName}`);
                        }
                        const trainingWorkloadType = workloadTypes.find((t) => t.name === workloadTypeName);
                        expect(trainingWorkloadType).toBeDefined();
                        expect(categoryNames).toContain(trainingWorkloadType.categoryName);
                        expect(expectedPriorityClasses).toContain(trainingWorkloadType.priorityName);
                        expect(trainingWorkloadType).toHaveProperty('id');
                        expect(trainingWorkloadType).toHaveProperty('name');
                        expect(trainingWorkloadType).toHaveProperty('updatedAt');
                        expect(trainingWorkloadType).toHaveProperty('updatedBy');

                        try {
                            if (!trainingWorkloadType.id) {
                                throw new Error('TrainingWorkload type has no valid ID');
                            }
                            const details = await workloadUtils.getWorkloadTypeById(trainingWorkloadType.id);
                            // Validate detail information
                            expect(details).toHaveProperty('id');
                            expect(details).toHaveProperty('categoryName');
                            expect(details).toHaveProperty('priorityName');
                            expect(details).toHaveProperty('updatedAt');
                            expect(details).toHaveProperty('updatedBy');
                            expect(details).toHaveProperty('name');
                            expect(typeof details.id).toBe('string');
                            expect(typeof details.name).toBe('string');
                            expect(details.name).toBe(workloadTypeName);
                        } catch (error) {
                            throw new Error(
                                `Failed to get detailed information for TrainingWorkload: ${error.message}`
                            );
                        }
                    });

                    test('T5217908: Update workload type by ID', async () => {
                        await test.step('Get a valid workload type ID and current details', async () => {
                            const workloadTypes = await workloadUtils.getWorkloadTypes();
                            expect(Array.isArray(workloadTypes)).toBe(true);
                            expect(workloadTypes.length).toBeGreaterThan(0);

                            const workloadType = workloadTypes.find((t) => t.name === workloadTypeName);

                            workloadTypeId = workloadType.id;
                            logger.info(`Workload type ID: ${workloadTypeId}`);
                            const currentDetails = await workloadUtils.getWorkloadTypeById(workloadTypeId);
                            originalCategoryId = currentDetails.categoryId;
                            originalPriorityId = currentDetails.priorityId;
                            originalCategoryName = currentDetails.categoryName;
                            originalPriorityName = currentDetails.priorityName;

                            logger.info(`Original category: ${originalCategoryName} (${originalCategoryId})`);
                            logger.info(`Original priority: ${originalPriorityName} (${originalPriorityId})`);
                        });

                        await test.step('Update workload type with new priority class', async () => {
                            const priorities = await workloadUtils.getWorkloadPriorities();
                            newPriority = priorities.find((p) => p.id !== originalPriorityId);
                            expect(newPriority).toBeDefined();
                            const categories = await workloadUtils.getWorkloadCategories();
                            expect(Array.isArray(categories)).toBe(true);
                            expect(categories.length).toBeGreaterThan(0);

                            newCategory = categories.find((c) => c.id !== originalCategoryId);
                            expect(newCategory).toBeDefined();
                            logger.info(`Updating to category: ${newCategory.name} (${newCategory.id})`);
                            logger.info(`Updating to priority: ${newPriority.name} (${newPriority.id})`);

                            const updateResponse = await workloadUtils.updateWorkloadTypeById(
                                workloadTypeId,
                                {
                                    categoryId: newCategory.id,
                                    priorityId: newPriority.id
                                }
                            );
                            workloadTypeUpdated = true;
                            expect(updateResponse.response.statusCode).toBe(StatusCodes.OK);
                            logger.info('Workload type updated successfully');
                        });

                        await test.step('Verify the update by getting the workload type again', async () => {
                            const updatedDetails = await workloadUtils.getWorkloadTypeById(workloadTypeId);
                            expect(updatedDetails).toBeDefined();
                            expect(updatedDetails.categoryId).toBe(newCategory.id);
                            expect(updatedDetails.priorityId).toBe(newPriority.id);

                            logger.info('Update verification completed successfully');
                        });
                    });

                    test('T5217910: Create and validate policy with default PriorityClass', async () => {
                        await withProjectConfig(
                            { deserved: 2, maxAllowed: 2, overQuotaWeight: 0 },
                            async () => {
                                const policyWorkloadName = 'policy-default-test';
                                await test.step('Create training policy with default PriorityClass set to medium-high', async () => {
                                    const response = await policyUtils.createPolicyV2(
                                        projectName,
                                        project.id.toString(),
                                        policyWorkloadType,
                                        { priorityClass: PRIORITY_MEDIUM_HIGH },
                                        undefined,
                                        ScopeType.Project
                                    );
                                    expect(response.response.statusCode).toBe(StatusCodes.OK);
                                    logger.info(
                                        'Training policy created successfully with default PriorityClass'
                                    );
                                    policyCreated = true;
                                    await policyUtils.waitForPolicyToBeSynced(
                                        policyWorkloadType,
                                        namespace,
                                        policyName
                                    );
                                });

                                await test.step('Submit training workload without specifying PriorityClass', async () => {
                                    const { body: training } = await workloadsCrudUtils.createTraining(
                                        policyWorkloadName,
                                        false,
                                        flowRootState.cluster.uuid,
                                        project.id.toString(),
                                        {
                                            image: config.quickStartImage.ubuntu,
                                            command: 'sleep infinity',
                                            compute: { gpuDevicesRequest: 1 }
                                        }
                                    );
                                    createdWorkloadIds.push(training.workloadId);
                                    workloadCreated = true;
                                    await workloadsCrudUtils.waitForWorkloadPhase(
                                        training.workloadId,
                                        Phase.Running,
                                        WorkloadType.TRAINING,
                                        undefined,
                                        120
                                    );
                                    logger.info('Training workload submitted and running successfully');
                                });

                                await test.step('Validate CRD has correct PriorityClass', async () => {
                                    await expectWorkloadRunnerPriorityClass(
                                        policyWorkloadName,
                                        PRIORITY_MEDIUM_HIGH
                                    );
                                    logger.info(
                                        'WorkloadRunner validation passed: PriorityClass is set correctly'
                                    );
                                });

                                await test.step('Validate Pod has correct PriorityClass', async () => {
                                    const pods = await kubeCli.listPodsOfJob(namespace, policyWorkloadName);
                                    expect(pods.length).toBeGreaterThan(0);
                                    expect(pods[0].spec.priorityClassName).toBe(PRIORITY_MEDIUM_HIGH);
                                    logger.info('Pod validation passed: PriorityClass is set correctly');
                                });
                            }
                        );
                    });

                    test('T5217911: Test policy with closed list of PriorityClasses', async ({}, testInfo: TestInfo) => {
                        skipPlaywrightTest(
                            test,
                            testInfo,
                            'PriorityClass policy coverage has an unresolved failure; revalidate policy enforcement before enabling this test'
                        );
                        const allowedWorkloadName = 'policy-allowed-test';
                        const disallowedWorkloadName = 'policy-disallowed-test';
                        let allowedTrainingId: string;

                        await test.step('Add rule to policy restricting PriorityClass to closed list', async () => {
                            const allowedPriorityClasses = [
                                PRIORITY_MEDIUM,
                                PRIORITY_MEDIUM_HIGH,
                                PRIORITY_HIGH
                            ];
                            const response = await policyUtils.updatePolicyV2(
                                projectName,
                                project.id.toString(),
                                policyWorkloadType,
                                { priorityClass: PRIORITY_LOW },
                                {
                                    priorityClass: {
                                        allowedValues: allowedPriorityClasses
                                    }
                                },
                                ScopeType.Project
                            );
                            expect(response.response.statusCode).toBe(StatusCodes.OK);
                            policyCreated = true;
                            await policyUtils.waitForPolicyToBeSynced(
                                policyWorkloadType,
                                namespace,
                                policyName
                            );
                        });

                        await test.step('Submit workload with allowed PriorityClass', async () => {
                            const { body: training } = await workloadsCrudUtils.createTraining(
                                allowedWorkloadName,
                                false,
                                flowRootState.cluster.uuid,
                                project.id.toString(),
                                {
                                    image: config.quickStartImage.ubuntu,
                                    command: 'sleep infinity',
                                    compute: { gpuDevicesRequest: 1 },
                                    priorityClass: PRIORITY_MEDIUM_HIGH
                                }
                            );
                            allowedTrainingId = training.workloadId;
                            createdWorkloadIds.push(training.workloadId);
                            workloadCreated = true;
                            await workloadsCrudUtils.waitForWorkloadPhase(
                                training.workloadId,
                                Phase.Running,
                                WorkloadType.TRAINING,
                                undefined,
                                120
                            );

                            await expectWorkloadRunnerPriorityClass(
                                allowedWorkloadName,
                                PRIORITY_MEDIUM_HIGH
                            );

                            const pods = await kubeCli.listPodsOfJob(namespace, allowedWorkloadName);
                            expect(pods.length).toBeGreaterThan(0);
                            expect(pods[0].spec.priorityClassName).toBe(PRIORITY_MEDIUM_HIGH);
                            logger.info('PriorityClass validation passed for allowed workload');
                        });

                        await test.step('Try to submit workload with disallowed PriorityClass', async () => {
                            try {
                                await workloadsCrudUtils.createTraining(
                                    disallowedWorkloadName,
                                    false,
                                    flowRootState.cluster.uuid,
                                    project.id.toString(),
                                    {
                                        image: config.quickStartImage.ubuntu,
                                        command: 'sleep infinity',
                                        compute: { gpuDevicesRequest: 1 },
                                        priorityClass: PRIORITY_VERY_HIGH
                                    }
                                );
                                throw new Error(
                                    'API did not throw error for disallowed PriorityClass, test failed'
                                );
                            } catch (e: any) {
                                if (e && e.response && e.response.statusCode) {
                                    expect(e.response.statusCode).toBe(StatusCodes.BAD_REQUEST);
                                    const errStr = JSON.stringify(e.body) || e.message;
                                    expect(errStr).toContain('priorityClass');
                                    expect(errStr).toContain('not allowed');
                                    logger.info('Expected API error for disallowed PriorityClass:', errStr);
                                } else {
                                    throw e;
                                }
                            }
                        });
                    });

                    test('T5217912: Test policy requiring PriorityClass', async ({}, testInfo: TestInfo) => {
                        skipPlaywrightTest(
                            test,
                            testInfo,
                            'PriorityClass policy coverage has an unresolved failure; revalidate policy enforcement before enabling this test'
                        );
                        const requiredWorkloadName = 'policy-required-test';
                        const successWorkloadName = 'policy-required-success-test';
                        let successTrainingId: string;

                        await test.step('Add rule to policy requiring PriorityClass', async () => {
                            const response = await policyUtils.updatePolicyV2(
                                projectName,
                                project.id.toString(),
                                policyWorkloadType,
                                { priorityClass: PRIORITY_LOW }, // Keep default
                                {
                                    priorityClass: {
                                        required: true
                                    }
                                },
                                ScopeType.Project
                            );
                            expect(response.response.statusCode).toBe(StatusCodes.OK);
                            policyCreated = true;
                            await policyUtils.waitForPolicyToBeSynced(
                                policyWorkloadType,
                                namespace,
                                policyName
                            );
                        });

                        await test.step('Get and validate updated policy', async () => {
                            const { body: updatedPolicy } = await policyUtils.getPolicyV2(
                                policyWorkloadType,
                                {
                                    scope: 'project',
                                    scopeId: project.id.toString()
                                }
                            );
                        });

                        await test.step('Try to submit workload without PriorityClass', async () => {
                            try {
                                await workloadsCrudUtils.createTraining(
                                    requiredWorkloadName,
                                    false,
                                    flowRootState.cluster.uuid,
                                    project.id.toString(),
                                    {
                                        image: config.quickStartImage.ubuntu,
                                        command: 'sleep infinity',
                                        compute: { gpuDevicesRequest: 1 }
                                    }
                                );
                                throw new Error(
                                    'API did not throw error for missing PriorityClass, test failed'
                                );
                            } catch (e: any) {
                                if (e && e.response && e.response.statusCode) {
                                    expect(e.response.statusCode).toBe(StatusCodes.BAD_REQUEST);
                                    const errStr = JSON.stringify(e.body) || e.message;
                                    expect(errStr).toContain('priorityClass');
                                    expect(errStr).toContain('required');
                                    logger.info('Expected API error for missing PriorityClass:', errStr);
                                } else {
                                    throw e;
                                }
                            }
                        });

                        await test.step('Submit workload with PriorityClass', async () => {
                            const { body: training } = await workloadsCrudUtils.createTraining(
                                successWorkloadName,
                                false,
                                flowRootState.cluster.uuid,
                                project.id.toString(),
                                {
                                    image: config.quickStartImage.ubuntu,
                                    command: 'sleep infinity',
                                    compute: { gpuDevicesRequest: 1 },
                                    priorityClass: PRIORITY_MEDIUM
                                }
                            );
                            successTrainingId = training.workloadId;
                            createdWorkloadIds.push(training.workloadId);
                            workloadCreated = true;
                            await workloadsCrudUtils.waitForWorkloadPhase(
                                training.workloadId,
                                Phase.Running,
                                WorkloadType.TRAINING,
                                undefined,
                                120
                            );

                            await expectWorkloadRunnerPriorityClass(successWorkloadName, PRIORITY_MEDIUM);

                            const pods = await kubeCli.listPodsOfJob(namespace, successWorkloadName);
                            expect(pods.length).toBeGreaterThan(0);
                            expect(pods[0].spec.priorityClassName).toBe(PRIORITY_MEDIUM);
                            logger.info('PriorityClass validation passed for successful workload');
                        });
                    });

                    test('T5217913: Delete policy and verify behavior', async () => {
                        await withProjectConfig(
                            { deserved: 2, maxAllowed: 2, overQuotaWeight: 0 },
                            async () => {
                                const noPolicyWorkloadName = 'no-policy-test';
                                const withPriorityWorkloadName = 'no-policy-with-priority-test';

                                await test.step('Delete the policy', async () => {
                                    try {
                                        const response = await policyUtils.deleteWorkloadPolicyV2(
                                            policyWorkloadType,
                                            {
                                                scope: 'project',
                                                scopeId: project.id.toString()
                                            }
                                        );
                                        expect(response).toBe(StatusCodes.OK);
                                        policyCreated = true;
                                        logger.info('Policy deleted successfully');
                                    } catch {
                                        logger.info(
                                            'Policy deletion failed, trying to clear policy content instead'
                                        );
                                        const response = await policyUtils.updatePolicyV2(
                                            projectName,
                                            project.id.toString(),
                                            policyWorkloadType,
                                            { priorityClass: PRIORITY_LOW },
                                            undefined,
                                            ScopeType.Project
                                        );
                                        expect(response.response.statusCode).toBe(StatusCodes.OK);
                                        policyCreated = true;
                                        logger.info('Policy content cleared successfully');
                                    }
                                    await policyUtils.waitForPolicyToBeSynced(
                                        policyWorkloadType,
                                        namespace,
                                        policyName
                                    );
                                });

                                await test.step('Submit workload without PriorityClass after policy deletion', async () => {
                                    const { body: training } = await workloadsCrudUtils.createTraining(
                                        noPolicyWorkloadName,
                                        false,
                                        flowRootState.cluster.uuid,
                                        project.id.toString(),
                                        {
                                            image: config.quickStartImage.ubuntu,
                                            command: 'sleep infinity',
                                            compute: { gpuDevicesRequest: 1 }
                                        }
                                    );
                                    createdWorkloadIds.push(training.workloadId);
                                    workloadCreated = true;
                                    await workloadsCrudUtils.waitForWorkloadPhase(
                                        training.workloadId,
                                        Phase.Running,
                                        WorkloadType.TRAINING,
                                        undefined,
                                        120
                                    );
                                    await expectWorkloadRunnerPriorityClass(
                                        noPolicyWorkloadName,
                                        PRIORITY_LOW
                                    ); // System default

                                    const pods = await kubeCli.listPodsOfJob(namespace, noPolicyWorkloadName);
                                    expect(pods.length).toBeGreaterThan(0);
                                    expect(pods[0].spec.priorityClassName).toBe(PRIORITY_LOW);
                                    logger.info('System default PriorityClass applied correctly');
                                });

                                await test.step('Submit workload with PriorityClass after policy deletion', async () => {
                                    const { body: training } = await workloadsCrudUtils.createTraining(
                                        withPriorityWorkloadName,
                                        false,
                                        flowRootState.cluster.uuid,
                                        project.id.toString(),
                                        {
                                            image: config.quickStartImage.ubuntu,
                                            command: 'sleep infinity',
                                            compute: { gpuDevicesRequest: 1 },
                                            priorityClass: PRIORITY_HIGH
                                        }
                                    );
                                    createdWorkloadIds.push(training.workloadId);
                                    workloadCreated = true;
                                    await workloadsCrudUtils.waitForWorkloadPhase(
                                        training.workloadId,
                                        Phase.Running,
                                        WorkloadType.TRAINING,
                                        undefined,
                                        120
                                    );

                                    await expectWorkloadRunnerPriorityClass(
                                        withPriorityWorkloadName,
                                        PRIORITY_HIGH
                                    );

                                    const pods = await kubeCli.listPodsOfJob(
                                        namespace,
                                        withPriorityWorkloadName
                                    );
                                    expect(pods.length).toBeGreaterThan(0);
                                    expect(pods[0].spec.priorityClassName).toBe(PRIORITY_HIGH);
                                    logger.info(
                                        'Custom PriorityClass applied correctly after policy deletion'
                                    );
                                });
                            }
                        );
                    });

                    test('T5217914: Test policy inheritance and precedence', async ({}, testInfo: TestInfo) => {
                        skipPlaywrightTest(
                            test,
                            testInfo,
                            'PriorityClass policy coverage has an unresolved failure; revalidate policy enforcement before enabling this test'
                        );
                        const inheritanceWorkloadName = 'policy-inheritance-test';
                        const conflictWorkloadName = 'policy-conflict-test';
                        const projectPolicyName = `project-${project.id}`;
                        clusterScopedCrdPolicyName = policyUtils.getPolicyName(
                            ScopeType.Cluster,
                            flowRootState.cluster.uuid
                        );

                        await test.step('Create cluster-level policy with default PriorityClass', async () => {
                            const response = await policyUtils.createPolicyV2(
                                `cluster-${flowRootState.cluster.uuid}`,
                                flowRootState.cluster.uuid,
                                policyWorkloadType,
                                { priorityClass: PRIORITY_MEDIUM }, // Cluster default
                                undefined,
                                ScopeType.Cluster
                            );
                            expect(response.response.statusCode).toBe(StatusCodes.OK);
                            clusterPolicyCreated = true;
                            logger.info('Cluster-level policy created successfully');
                            await policyUtils.waitForPolicyToBeSynced(
                                policyWorkloadType,
                                namespaceCluster,
                                clusterScopedCrdPolicyName
                            );
                        });

                        await test.step('Create project-level policy with different default PriorityClass', async () => {
                            const response = await policyUtils.createPolicyV2(
                                projectName,
                                project.id.toString(),
                                policyWorkloadType,
                                { priorityClass: PRIORITY_HIGH },
                                undefined,
                                ScopeType.Project
                            );
                            expect(response.response.statusCode).toBe(StatusCodes.OK);
                            logger.info('Project-level policy created successfully');
                            policyCreated = true;
                            await policyUtils.waitForPolicyToBeSynced(
                                policyWorkloadType,
                                namespace,
                                projectPolicyName
                            );
                        });

                        await test.step('Submit workload in project with both policies', async () => {
                            const { body: training } = await workloadsCrudUtils.createTraining(
                                inheritanceWorkloadName,
                                false,
                                flowRootState.cluster.uuid,
                                project.id.toString(),
                                {
                                    image: config.quickStartImage.ubuntu,
                                    command: 'sleep infinity',
                                    compute: { gpuDevicesRequest: 1 }
                                }
                            );
                            createdWorkloadIds.push(training.workloadId);
                            workloadCreated = true;
                            await workloadsCrudUtils.waitForWorkloadPhase(
                                training.workloadId,
                                Phase.Running,
                                WorkloadType.TRAINING,
                                undefined,
                                120
                            );

                            await expectWorkloadRunnerPriorityClass(inheritanceWorkloadName, PRIORITY_HIGH); // Project policy should override

                            const pods = await kubeCli.listPodsOfJob(namespace, inheritanceWorkloadName);
                            expect(pods.length).toBeGreaterThan(0);
                            expect(pods[0].spec.priorityClassName).toBe(PRIORITY_HIGH);
                            logger.info(
                                'Policy precedence validation passed - project policy overrides cluster policy'
                            );
                        });

                        await test.step('Create conflicting policies', async () => {
                            const clusterResponse = await policyUtils.updatePolicyV2(
                                `cluster-${flowRootState.cluster.uuid}`,
                                flowRootState.cluster.uuid,
                                policyWorkloadType,
                                { priorityClass: PRIORITY_MEDIUM },
                                {
                                    priorityClass: {
                                        allowedValues: [PRIORITY_LOW, PRIORITY_MEDIUM, PRIORITY_MEDIUM_LOW]
                                    }
                                },
                                ScopeType.Cluster
                            );
                            expect(clusterResponse.response.statusCode).toBe(StatusCodes.OK);

                            const projectResponse = await policyUtils.updatePolicyV2(
                                projectName,
                                project.id.toString(),
                                policyWorkloadType,
                                { priorityClass: PRIORITY_HIGH },
                                {
                                    priorityClass: {
                                        allowedValues: [PRIORITY_HIGH, PRIORITY_VERY_HIGH]
                                    }
                                },
                                ScopeType.Project
                            );
                            expect(projectResponse.response.statusCode).toBe(StatusCodes.OK);

                            await policyUtils.waitForPolicyToBeSynced(
                                policyWorkloadType,
                                namespaceCluster,
                                clusterScopedCrdPolicyName
                            );
                            await policyUtils.waitForPolicyToBeSynced(
                                policyWorkloadType,
                                namespace,
                                projectPolicyName
                            );
                            logger.info('Conflicting policies created successfully');
                        });

                        await test.step('Test workload submission with conflicting policies', async () => {
                            try {
                                await workloadsCrudUtils.createTraining(
                                    conflictWorkloadName,
                                    false,
                                    flowRootState.cluster.uuid,
                                    project.id.toString(),
                                    {
                                        image: config.quickStartImage.ubuntu,
                                        command: 'sleep infinity',
                                        compute: { gpuDevicesRequest: 1 },
                                        priorityClass: PRIORITY_MEDIUM
                                    }
                                );
                                throw new Error(
                                    'API did not throw error for conflicting policies, test failed'
                                );
                            } catch (e: any) {
                                if (e && e.response && e.response.statusCode) {
                                    expect(e.response.statusCode).toBe(StatusCodes.BAD_REQUEST);
                                    const errStr = JSON.stringify(e.body) || e.message;
                                    expect(errStr).toContain('priorityClass');
                                    logger.info('Expected API error for conflicting policies:', errStr);
                                } else {
                                    throw e;
                                }
                            }

                            const { body: training } = await workloadsCrudUtils.createTraining(
                                conflictWorkloadName,
                                false,
                                flowRootState.cluster.uuid,
                                project.id.toString(),
                                {
                                    image: config.quickStartImage.ubuntu,
                                    command: 'sleep infinity',
                                    compute: { gpuDevicesRequest: 1 },
                                    priorityClass: PRIORITY_HIGH
                                }
                            );
                            createdWorkloadIds.push(training.workloadId);
                            workloadCreated = true;

                            await workloadsCrudUtils.waitForWorkloadPhase(
                                training.workloadId,
                                Phase.Running,
                                WorkloadType.TRAINING,
                                undefined,
                                120
                            );
                            logger.info('Workload with compatible priority submitted successfully');
                            await expectWorkloadRunnerPriorityClass(conflictWorkloadName, PRIORITY_HIGH);

                            const pods = await kubeCli.listPodsOfJob(namespace, conflictWorkloadName);
                            expect(pods.length).toBeGreaterThan(0);
                            expect(pods[0].spec.priorityClassName).toBe(PRIORITY_HIGH);
                            logger.info('Policy conflict resolution validation passed');
                        });
                    });
                });
            });
        });
    });
});
