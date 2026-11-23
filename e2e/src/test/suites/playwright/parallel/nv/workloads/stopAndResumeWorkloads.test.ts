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
    settings,
    config,
    workloadsUtilsV2,
    workloadUtils
} from '../../../../../../modules/helpers';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { TrainingPage } from '../../../../../../modules/ui/runaiV2/pages/TrainingPage';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test, expect, DEBUG_TAGS } from '../../../../../helpers/DebugInfoCollector';
import { DISTRIBUTED_TOOLS_LABELS } from '../../../../../../modules/ui/runaiV2/pages/TrainingPage';
import { EnvironmentPage } from '../../../../../../modules/ui/runaiV2/pages/EnvironmentPage';
import { kubeCli } from '../../../../../../modules/cli';
import { crds, CustomResourceDefinition } from '../../../../../../modules/helpers';
import { waitUtils } from '../../../../../../modules/helpers';
import { WorkloadStatus } from '../../../../../../modules/helpers/models/baseWorkload';
import { runaiCliV2 } from '../../../../../../modules/cli';
import { workloadsCrudUtils } from '../../../../../../modules/helpers';
import { Phase } from '../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import { WorkloadType } from '../../../../../../modules/helpers/utils/WorkloadsCrudUtils';
import { cliV2WorkloadType } from '../../../../../../modules/cli/tools/RunaiCliV2.constants';
import { SubmitFlagsMerged } from '../../../../../../modules/cli/tools/RunaiCliV2.constants';
import { WorkloadPage, EWorkloadStatus } from '../../../../../../modules/ui/runaiV2/pages/WorkloadPage';
import { HybridForm } from '../../../../../../modules/ui/runaiV2/pages/HybridSubmissionForm';

test.describe(`Workload suspending and resuming, tags: ${tags.resource.workloads} ${tags.teams.loong}, debugTags: ${DEBUG_TAGS.SYSTEM_LOGS}`, () => {
    supportedFrom(SupportedVersions.V2_23)(function () {
        const ERROR_MESSAGES = {
            // CLI error messages
            FAILED_TO_RESUME_WORKLOAD: 'failed to resume workload',
            RESUME_ONLY_FOR_STOPPED: 'resume request is permitted only for stopped workloads',
            // quote-free substring so it matches both the CLI stderr (literal quotes) and the
            // API HttpError body (JSON-escaped quotes)
            WORKLOAD_IN_RUNNING_PHASE: 'currently the workload is in Running phase',
            FAILED_TO_SUSPEND_WORKLOAD: 'failed to suspend workload',
            SUSPEND_NOT_PERMITTED_FOR_STOPPED: 'suspend request is not permitted for Stopped workloads',
            // Finite state error messages
            SUSPEND_NOT_PERMITTED_FOR_COMPLETED: 'suspend request is not permitted for Completed workloads',
            SUSPEND_NOT_PERMITTED_FOR_FAILED: 'suspend request is not permitted for Failed workloads'
        };

        function getConcreteDistributedJobCrd(workloadType: cliV2WorkloadType): CustomResourceDefinition {
            switch (workloadType) {
                case cliV2WorkloadType.mpi:
                    return crds.mpiJobCrd;
                case cliV2WorkloadType.pyTorch:
                    return crds.pyTorchJobCrd;
                case cliV2WorkloadType.tensorFlow:
                    return crds.tfJobCrd;
                case cliV2WorkloadType.xgBoost:
                    return crds.xgboostJobCrd;
                case cliV2WorkloadType.jax:
                    return crds.jaxJobCrd;
                default:
                    throw new Error(`Unsupported distributed workload type: ${workloadType}`);
            }
        }

        function getUiDistributedWorkloadType(type: DISTRIBUTED_TOOLS_LABELS): cliV2WorkloadType {
            switch (type) {
                case DISTRIBUTED_TOOLS_LABELS.MPI:
                    return cliV2WorkloadType.mpi;
                case DISTRIBUTED_TOOLS_LABELS.PY_TORCH:
                    return cliV2WorkloadType.pyTorch;
                case DISTRIBUTED_TOOLS_LABELS.TENSOR_FLOW:
                    return cliV2WorkloadType.tensorFlow;
                case DISTRIBUTED_TOOLS_LABELS.XG_BOOST:
                    return cliV2WorkloadType.xgBoost;
                case DISTRIBUTED_TOOLS_LABELS.JAX:
                    return cliV2WorkloadType.jax;
                default:
                    throw new Error(`Unsupported distributed workload type: ${type}`);
            }
        }

        function workloadStatusToPhase(status: WorkloadStatus): Phase {
            const phase = Object.values(Phase).find((value) => String(value) === String(status));
            if (!phase) {
                throw new Error(`Workload status ${status} has no workloads read-model phase`);
            }
            return phase as Phase;
        }

        test.describe('UI cases', () => {
            let browserEngine: BrowserEngine;
            let auth: AuthPage;
            let training: TrainingPage;
            let workloadPage: WorkloadPage;
            let hybridForm: HybridForm;

            let videoName: string;
            let projectName: string;
            let namespace: string;
            let project: Project;
            test.beforeAll(async function () {
                browserEngine = new BrowserEngine();
                projectName = objectGenerator.generateShortNameByPrefix('project-e2e-for-training');
                namespace = Utils.namespaceFromProjectName(projectName);
                project = await projectResourcesUtils.createProject(namespace, {
                    name: projectName,
                    deservedGpus: 0,
                    maxAllowedGpus: 0
                });
                runaiCliV2.loginAsUser(
                    flowRootState.users.l1_researcher.email,
                    flowRootState.users.l1_researcher.password
                );
                runaiCliV2.projectSet(projectName);
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
                workloadPage = new WorkloadPage(browserEngine.currentPage);
                hybridForm = new HybridForm(browserEngine.currentPage);
                await auth.signIn(
                    flowRootState.users.l1_researcher.email,
                    flowRootState.users.l1_researcher.password
                );

                // retrieve access token for 'workloadsCrudUtils' in beforeEach instead of beforeAll hook
                // to minimize the risk of token expiration since there are 40+ tests in this file
                workloadsCrudUtils.accessToken = flowRootState.users.l1_researcher.accessToken;
            });

            test.afterEach('stop browser', async ({}, testInfo) => {
                if (browserEngine.currentPage) {
                    await browserEngine.stopAndArchive(testInfo, 'training-' + videoName);
                }
            });

            async function waitForWorkloadUITransition(
                workloadName: string,
                isDistributed: boolean,
                toPhase: WorkloadStatus,
                crdToPhase: WorkloadStatus,
                options: {
                    timeout?: number;
                    invalidPhase?: WorkloadStatus;
                    action?: () => void;
                    uiOnly?: boolean;
                } = {}
            ): Promise<void> {
                const { timeout = 600, invalidPhase, action, uiOnly = false } = options;

                const crdPromise = uiOnly
                    ? Promise.resolve()
                    : isDistributed
                      ? workloadUtils.waitForWorkloadBackendPhase(
                            workloadName,
                            projectName,
                            workloadStatusToPhase(toPhase),
                            timeout
                        )
                      : waitUtils.waitForWrappedJobState(workloadName, namespace, crdToPhase, timeout);
                const uiPromise = training.waitForWorkloadUIStatus(workloadName, toPhase);

                if (action) {
                    await new Promise((resolve) => setTimeout(resolve, 1000));
                    action();
                }
                await Promise.all([crdPromise, uiPromise]);
            }

            async function createAndVerifyWorkloadUI(
                workloadName: string,
                projectName: string,
                options: {
                    isDistributed?: boolean;
                    distributedType?: DISTRIBUTED_TOOLS_LABELS;
                    image?: string;
                    command?: string;
                    workersCount?: string;
                } = {}
            ): Promise<void> {
                const {
                    isDistributed = false,
                    distributedType,
                    image = config.quickStartImage.ubuntu,
                    command = 'sleep infinity',
                    workersCount = '2'
                } = options;

                await test.step('Create workload via Run:ai UI', async () => {
                    await workloadPage.navigateToWorkloads();

                    if (isDistributed) {
                        await training.fillFirstNewTrainingPage(
                            projectName,
                            workloadName,
                            true,
                            distributedType,
                            true,
                            true
                        );
                        if (distributedType === DISTRIBUTED_TOOLS_LABELS.JAX) {
                            await training.fillHybridDistributedTrainingCommonPartSecondPage(image, {
                                command,
                                args: '',
                                computeLabel: 'cpu-only',
                                workersCount,
                                imagePullPolicy: 'Always pull the image from the registry'
                            });
                            await training.clickOnSubmitBtn(); // worker step
                            await training.page.waitForTimeout(1500);
                            await training.clickOnSubmitBtn(); // setup step
                        } else {
                            await training.fillHybridDistributedTrainingSecondPage(image, {
                                command,
                                args: '',
                                computeLabel: 'cpu-only',
                                workersCount,
                                imagePullPolicy: 'Always pull the image from the registry'
                            });
                        }
                    } else {
                        await hybridForm.fillAndSubmitTrainingWorkloadForm(
                            workloadName,
                            projectName,
                            flowRootState.cluster.name,
                            {
                                nodePoolList: ['default'],
                                computeTemplate: 'cpu-only',
                                image: image,
                                cmdInput: command
                            }
                        );
                    }
                });

                await test.step('Verify workload appears in UI and is Running', async () => {
                    if (isDistributed) {
                        const workloadType = getUiDistributedWorkloadType(distributedType);
                        await workloadUtils.waitForDistributedSubmissionRoot(workloadName, namespace);
                        await waitUtils.waitForCrd(
                            getConcreteDistributedJobCrd(workloadType),
                            workloadName,
                            namespace
                        );
                        await waitUtils.waitForPods(workloadName, namespace);
                    }
                    await waitForWorkloadUITransition(
                        workloadName,
                        isDistributed,
                        WorkloadStatus.RUNNING,
                        WorkloadStatus.RUNNING
                    );
                });
            }

            async function suspendWorkloadUI(
                workloadName: string,
                isDistributed: boolean = false
            ): Promise<void> {
                await test.step('Stop the workload by clicking STOP and check Stopping/Suspending state', async () => {
                    await waitForWorkloadUITransition(
                        workloadName,
                        isDistributed,
                        WorkloadStatus.STOPPING,
                        WorkloadStatus.SUSPENDING,
                        {
                            action: () => training.stopWorkloadByName(workloadName)
                        }
                    );
                });

                await test.step('Check workload state transitions to Stopped/Suspended', async () => {
                    await waitForWorkloadUITransition(
                        workloadName,
                        isDistributed,
                        WorkloadStatus.STOPPED,
                        WorkloadStatus.SUSPENDED,
                        {
                            invalidPhase: WorkloadStatus.RUNNING
                        }
                    );
                });

                await test.step('Verify k8s pods do not exist', async () => {
                    await waitUtils.waitForPodsDeletion(workloadName, namespace);
                });
            }

            async function resumeWorkloadUI(
                workloadName: string,
                isDistributed: boolean = false
            ): Promise<void> {
                await test.step('Resume the workload by clicking RUN and check Resuming state', async () => {
                    // After switching to the WorkloadsV2 API, itno longer reports a Resuming phase - it can go
                    // straight from Stopped to Initializing/Running on UI.
                    await waitForWorkloadUITransition(
                        workloadName,
                        isDistributed,
                        WorkloadStatus.RESUMING,
                        WorkloadStatus.RESUMING,
                        {
                            action: () => training.activateWorkloadByName(workloadName),
                            uiOnly: true
                        }
                    );
                });

                await test.step('Check workload state transitions to Running', async () => {
                    await waitForWorkloadUITransition(
                        workloadName,
                        isDistributed,
                        WorkloadStatus.RUNNING,
                        WorkloadStatus.RUNNING
                    );
                });

                await test.step('Verify k8s pods are recreated', async () => {
                    const pods = await kubeCli.listPodsByNamepsace(namespace);
                    expect(
                        pods.filter((p) => p.metadata?.name?.includes(workloadName)).length
                    ).toBeGreaterThan(0);
                });
            }

            async function runUISuspendResumeTest(
                workloadName: string,
                options: {
                    projectName: string;
                    isDistributed?: boolean;
                    distributedType?: DISTRIBUTED_TOOLS_LABELS;
                    image?: string;
                    command?: string;
                    workersCount?: string;
                }
            ): Promise<void> {
                await createAndVerifyWorkloadUI(workloadName, options.projectName, options);
                await suspendWorkloadUI(workloadName, options.isDistributed || false);
                await resumeWorkloadUI(workloadName, options.isDistributed || false);
                await training.deleteWorkload(workloadName);
            }

            test('T5167106: UI-Suspend and resume a single node workload', async () => {
                test.setTimeout(300 * 1000);
                videoName = 'suspend-resume-single-node-workload';

                const workloadName = objectGenerator.generateShortNameByPrefix('ui-training');
                await runUISuspendResumeTest(workloadName, { projectName });
            });

            test('T5187457: UI-Suspend and resume a MPI distributed workload', async () => {
                test.setTimeout(300 * 1000);
                videoName = 'mpi-suspend-resume';
                const workloadName = objectGenerator.generateShortNameByPrefix('ui-mpi');
                await runUISuspendResumeTest(workloadName, {
                    projectName,
                    isDistributed: true,
                    distributedType: DISTRIBUTED_TOOLS_LABELS.MPI,
                    image: config.quickStartImage.ubuntu,
                    workersCount: '2'
                });
            });

            test('T5187458: UI-Suspend and resume a PyTorch distributed workload', async () => {
                test.setTimeout(300 * 1000);
                videoName = 'pytorch-suspend-resume';
                const workloadName = objectGenerator.generateShortNameByPrefix('ui-pytorch');
                await runUISuspendResumeTest(workloadName, {
                    projectName,
                    isDistributed: true,
                    distributedType: DISTRIBUTED_TOOLS_LABELS.PY_TORCH,
                    image: config.quickStartImage.ubuntu,
                    workersCount: '2'
                });
            });

            test('T5187459: UI-Suspend and resume a TensorFlow distributed workload', async () => {
                test.setTimeout(300 * 1000);
                videoName = 'tensorflow-suspend-resume';
                const workloadName = objectGenerator.generateShortNameByPrefix('ui-tensorflow');
                await runUISuspendResumeTest(workloadName, {
                    projectName,
                    isDistributed: true,
                    distributedType: DISTRIBUTED_TOOLS_LABELS.TENSOR_FLOW,
                    image: config.quickStartImage.ubuntu,
                    workersCount: '2'
                });
            });

            test('T5187460: UI-Suspend and resume an XGBoost distributed workload', async () => {
                test.setTimeout(300 * 1000);
                videoName = 'xgboost-suspend-resume';
                const workloadName = objectGenerator.generateShortNameByPrefix('ui-xgboost');
                await runUISuspendResumeTest(workloadName, {
                    projectName,
                    isDistributed: true,
                    distributedType: DISTRIBUTED_TOOLS_LABELS.XG_BOOST,
                    image: config.quickStartImage.ubuntu,
                    workersCount: '2'
                });
            });

            test('T5193825: UI-Suspend and resume an JAX distributed workload', async () => {
                test.setTimeout(300 * 1000);
                videoName = 'jax-suspend-resume';
                const workloadName = objectGenerator.generateShortNameByPrefix('ui-jax');
                await runUISuspendResumeTest(workloadName, {
                    projectName,
                    isDistributed: true,
                    distributedType: DISTRIBUTED_TOOLS_LABELS.JAX,
                    image: config.quickStartImage.ubuntu,
                    workersCount: '2'
                });
            });

            test('T5187461: UI-Attempt to resume a Running workload', async () => {
                videoName = 'running-resume';
                const workloadName = objectGenerator.generateShortNameByPrefix('ui-running');

                await test.step('Create workload and wait for Running state', async () => {
                    await createAndVerifyWorkloadUI(workloadName, projectName, {
                        isDistributed: false,
                        image: config.quickStartImage.ubuntu,
                        command: 'sleep infinity'
                    });
                });

                await test.step('Toggle workload and check RUN button is disabled', async () => {
                    await workloadPage.toggleWorkload(workloadName);
                    const runButton = workloadPage.page.locator(workloadPage.activateWorkloadBtn);
                    expect(await runButton.isDisabled()).toBeTruthy();
                    await workloadPage.toggleWorkload(workloadName);
                });

                await test.step('Cleanup', async () => {
                    await waitUtils.sleep(1000);
                    await training.deleteWorkload(workloadName);
                });
            });

            test('T5187463: UI-Attempt to suspend an already Stopped workload', async () => {
                test.setTimeout(300 * 1000);
                videoName = 'stopped-suspend';
                const workloadName = objectGenerator.generateShortNameByPrefix('ui-stopped');

                await test.step('Create workload and wait for Running state', async () => {
                    await createAndVerifyWorkloadUI(workloadName, projectName, {
                        isDistributed: false,
                        image: config.quickStartImage.ubuntu,
                        command: 'sleep infinity'
                    });
                });

                await test.step('Suspend workload and wait for Stopped state', async () => {
                    training.stopWorkloadByName(workloadName);
                    await waitForWorkloadUITransition(
                        workloadName,
                        false,
                        WorkloadStatus.STOPPED,
                        WorkloadStatus.SUSPENDED
                    );
                });

                await test.step('Verify stop button is disabled', async () => {
                    await workloadPage.toggleWorkload(workloadName);
                    const stopButton = workloadPage.page.locator(workloadPage.stopWorkloadBtn);
                    const isDisabled = await stopButton.isDisabled();
                    expect(isDisabled).toBeTruthy();
                    await workloadPage.toggleWorkload(workloadName);
                });

                await test.step('Cleanup', async () => {
                    await waitUtils.sleep(1000);
                    await training.deleteWorkload(workloadName);
                });
            });

            async function runRapidSuspendResumeTest(endingState: WorkloadStatus): Promise<void> {
                test.setTimeout(600 * 1000);
                const prefix =
                    endingState === WorkloadStatus.STOPPED ? 'ui-rapid-stopped' : 'ui-rapid-running';
                const videoSuffix = endingState === WorkloadStatus.STOPPED ? 'stopped' : 'running';
                videoName = `rapid-suspend-resume-${videoSuffix}`;
                const workloadName = objectGenerator.generateShortNameByPrefix(prefix);

                await test.step('Create workload and wait for Running state', async () => {
                    await createAndVerifyWorkloadUI(workloadName, projectName, {
                        isDistributed: false,
                        image: config.quickStartImage.ubuntu,
                        command: 'sleep infinity'
                    });
                });

                for (let i = 0; i < 3; i++) {
                    await test.step(`Rapid suspend/resume cycle ${i + 1}`, async () => {
                        training.stopWorkloadByName(workloadName);
                        await training.waitForWorkloadUIStatus(workloadName, WorkloadStatus.STOPPED);
                        training.activateWorkloadByName(workloadName);
                        await training.waitForWorkloadUIStatus(workloadName, WorkloadStatus.RUNNING);
                    });
                }

                if (endingState === WorkloadStatus.STOPPED) {
                    await test.step('Final suspend to Stopped state', async () => {
                        training.stopWorkloadByName(workloadName);
                        await training.waitForWorkloadUIStatus(workloadName, WorkloadStatus.STOPPED);
                    });
                }

                await test.step(`Wait for 30 seconds and verify ${endingState} state`, async () => {
                    await new Promise((resolve) => setTimeout(resolve, 30000));
                    const expectedCrdState =
                        endingState === WorkloadStatus.STOPPED
                            ? WorkloadStatus.SUSPENDED
                            : WorkloadStatus.RUNNING;
                    await waitForWorkloadUITransition(workloadName, false, endingState, expectedCrdState);
                });

                await test.step('Cleanup', async () => {
                    await training.deleteWorkload(workloadName);
                });
            }

            test('T5187470: UI-Rapid suspend/resume operations ending in Stopped state', async () => {
                await runRapidSuspendResumeTest(WorkloadStatus.STOPPED);
            });

            test('T5187471: UI-Rapid suspend/resume operations ending in Running state', async () => {
                await runRapidSuspendResumeTest(WorkloadStatus.RUNNING);
            });

            async function runSuspendFiniteStateWorkloadTest(
                finalState: WorkloadStatus,
                command: string
            ): Promise<void> {
                test.setTimeout(300 * 1000);
                const stateName = finalState === WorkloadStatus.COMPLETED ? 'complete' : 'failed';
                const prefix =
                    finalState === WorkloadStatus.COMPLETED ? 'ui-suspend-complete' : 'ui-suspend-failed';
                videoName = `suspend-${stateName}`;
                let workloadName = objectGenerator.generateShortNameByPrefix(prefix);

                const stepDescription =
                    finalState === WorkloadStatus.COMPLETED
                        ? 'Create a short-lived workload via API'
                        : 'Create a failed workload via API with invalid command';

                await test.step(stepDescription, async () => {
                    const { body: workload } = await workloadsCrudUtils.createTraining(
                        workloadName,
                        false,
                        flowRootState.cluster.uuid,
                        String(project.id),
                        // backoffLimit: 0 -> a failing command reaches the terminal Failed state
                        // immediately (no retries), so the finite-state wait doesn't hang on backoff.
                        { command, image: config.quickStartImage.ubuntu, backoffLimit: 0 },
                        flowRootState.users.l1_researcher.accessToken
                    );
                    workloadName = workload.name;
                });

                await test.step(`Wait until the workload is in ${finalState} state`, async () => {
                    await workloadPage.navigateToWorkloads();
                    await training.waitForWorkloadUIStatus(workloadName, finalState);
                });

                await test.step('Verify RUN and STOP buttons are disabled', async () => {
                    await workloadPage.toggleWorkload(workloadName);
                    const runButton = workloadPage.page.locator(workloadPage.activateWorkloadBtn);
                    const stopButton = workloadPage.page.locator(workloadPage.stopWorkloadBtn);
                    expect(await runButton.isDisabled()).toBeTruthy();
                    expect(await stopButton.isDisabled()).toBeTruthy();
                    await workloadPage.toggleWorkload(workloadName);
                });

                await test.step('Cleanup', async () => {
                    await waitUtils.sleep(1000);
                    await training.deleteWorkload(workloadName);
                });
            }

            test('T5187483: UI-Suspend a completed workload via UI', async () => {
                await runSuspendFiniteStateWorkloadTest(WorkloadStatus.COMPLETED, 'sleep 1');
            });

            test('T5187484: UI-Suspend a failed workload via UI', async () => {
                await runSuspendFiniteStateWorkloadTest(WorkloadStatus.FAILED, 'invalidcmd');
            });

            test('T5187474: UI-UI suspend/resume, verify on CLI & API', async () => {
                test.setTimeout(300 * 1000);
                videoName = 'ui-suspend-resume-verify-cli-api';
                const workloadName = objectGenerator.generateShortNameByPrefix('ui-verify-all');
                let workloadId: string;

                await test.step('In the UI, create a single-node workload. Wait until its phase is Running', async () => {
                    await createAndVerifyWorkloadUI(workloadName, projectName, {
                        isDistributed: false,
                        image: config.quickStartImage.ubuntu,
                        command: 'sleep infinity'
                    });
                });

                await test.step('In the UI, verify the workload status is Running and the STOP button is visible', async () => {
                    await workloadPage.toggleWorkload(workloadName);
                    const stopButton = workloadPage.page.locator(workloadPage.stopWorkloadBtn);
                    await waitUtils.sleep(1000);
                    expect(await stopButton.isVisible()).toBeTruthy();
                    expect(await stopButton.isDisabled()).toBeFalsy();
                    await workloadPage.toggleWorkload(workloadName);
                });

                await test.step('In the CLI, run: runai training standard describe <workload-name>, verify phase is Running', async () => {
                    await runaiCliV2.waitForWorkloadPhaseWithSampling(
                        workloadName,
                        projectName,
                        WorkloadStatus.RUNNING,
                        cliV2WorkloadType.training,
                        30,
                        WorkloadStatus.STOPPED
                    );
                });

                await test.step('Get workload ID for API verification', async () => {
                    workloadId = await workloadsUtilsV2.getWorkloadIdByName(workloadName, project.id);
                });

                await test.step('In the API, verify phase is Running', async () => {
                    const apiResponse = await workloadsCrudUtils.getTraining(workloadId);
                    expect(apiResponse.body.actualPhase).toBe(Phase.Running);
                });

                await test.step('In the UI, click the STOP button, verify the workload status changes to Stopping->Stopped', async () => {
                    await suspendWorkloadUI(workloadName, false);
                });

                await test.step('In the CLI, verify phase transitions to Stopped', async () => {
                    await runaiCliV2.waitForWorkloadPhaseWithSampling(
                        workloadName,
                        projectName,
                        WorkloadStatus.STOPPED,
                        cliV2WorkloadType.training,
                        120
                    );
                });

                await test.step('In the API, verify phase transitions to Stopped', async () => {
                    const apiResponse = await workloadsCrudUtils.getTraining(workloadId);
                    expect(apiResponse.body.actualPhase).toBe(Phase.Stopped);
                });

                await test.step('In the UI, click the RUN button, verify the workload status changes to Resuming -> Running', async () => {
                    await resumeWorkloadUI(workloadName, false);
                });

                await test.step('In the CLI, verify phase transitions to Running', async () => {
                    await runaiCliV2.waitForWorkloadPhaseWithSampling(
                        workloadName,
                        projectName,
                        WorkloadStatus.RUNNING,
                        cliV2WorkloadType.training,
                        30,
                        WorkloadStatus.STOPPED
                    );
                });

                await test.step('In the API, verify phase transitions to Running', async () => {
                    const apiResponse = await workloadsCrudUtils.getTraining(workloadId);
                    expect(apiResponse.body.actualPhase).toBe(Phase.Running);
                });

                await test.step('Cleanup', async () => {
                    await training.deleteWorkload(workloadName);
                });
            });

            test('T5187475: API-API suspend/resume, verify on CLI & UI', async () => {
                test.setTimeout(300 * 1000);
                videoName = 'api-suspend-resume-verify-cli-ui';
                const workloadNamePrefix = 'api-verify-all';
                let workloadName: string;
                let workloadId: string;

                await test.step('Use the API to create a single-node workload. Wait until its phase is Running', async () => {
                    const clusterId = flowRootState.cluster.uuid;
                    const projectId = project.id;

                    const { body: workload } = await workloadsCrudUtils.createTraining(
                        workloadNamePrefix,
                        true,
                        clusterId,
                        String(projectId),
                        { command: 'sleep infinity', image: config.quickStartImage.ubuntu },
                        flowRootState.users.l1_researcher.accessToken
                    );
                    workloadId = workload.workloadId;
                    workloadName = workload.name;
                    await waitUtils.waitForCrd(crds.workloadRunnerCrd, workload.name, namespace);
                    await waitUtils.waitForWrappedJobState(
                        workload.name,
                        namespace,
                        WorkloadStatus.RUNNING,
                        120
                    );
                    const apiResponse = await workloadsCrudUtils.getTraining(workloadId);
                    expect(apiResponse.body.actualPhase).toBe(Phase.Running);
                });

                await test.step('In the CLI, verify phase is Running', async () => {
                    await runaiCliV2.waitForWorkloadPhaseWithSampling(
                        workloadName,
                        projectName,
                        WorkloadStatus.RUNNING,
                        cliV2WorkloadType.training,
                        300,
                        WorkloadStatus.STOPPED
                    );
                });

                await test.step('In the UI, verify the workload appears in the list as Running', async () => {
                    await workloadPage.navigateToWorkloads();
                    await training.waitForWorkloadUIStatus(workloadName, WorkloadStatus.RUNNING);
                });

                await test.step('Use the API to suspend the workload, verify phase transitions from Running to Stopping to Stopped', async () => {
                    await workloadsCrudUtils.suspendTraining(workloadId);
                    await workloadsCrudUtils.waitForWorkloadPhase(
                        workloadId,
                        Phase.Stopping,
                        WorkloadType.TRAINING
                    );
                    await workloadsCrudUtils.waitForWorkloadPhase(
                        workloadId,
                        Phase.Stopped,
                        WorkloadType.TRAINING
                    );
                });

                await test.step('In the CLI, verify phase transitions to Stopped', async () => {
                    await runaiCliV2.waitForWorkloadPhaseWithSampling(
                        workloadName,
                        projectName,
                        WorkloadStatus.STOPPED,
                        cliV2WorkloadType.training,
                        30
                    );
                });

                await test.step('In the UI, verify the status updates to Stopped', async () => {
                    await training.waitForWorkloadUIStatus(workloadName, WorkloadStatus.STOPPED);
                });

                await test.step('Use the API to resume the workload, verify phase transitions from Stopped to Resuming to Running', async () => {
                    await workloadsCrudUtils.resumeTraining(workloadId);
                    await workloadsCrudUtils.waitForWorkloadPhase(
                        workloadId,
                        Phase.Resuming,
                        WorkloadType.TRAINING
                    );
                    await workloadsCrudUtils.waitForWorkloadPhase(
                        workloadId,
                        Phase.Running,
                        WorkloadType.TRAINING
                    );
                });

                await test.step('In the CLI, verify phase transitions to Running', async () => {
                    await runaiCliV2.waitForWorkloadPhaseWithSampling(
                        workloadName,
                        projectName,
                        WorkloadStatus.RUNNING,
                        cliV2WorkloadType.training,
                        30
                    );
                });

                await test.step('In the UI, verify phase transitions to Running', async () => {
                    await training.waitForWorkloadUIStatus(workloadName, WorkloadStatus.RUNNING);
                });

                await test.step('Cleanup', async () => {
                    await workloadsCrudUtils.deleteTraining(workloadId);
                });
            });

            test('T5187476 UI-CLI suspend/resume, verify on API & UI', async () => {
                videoName = 'ui-cli-api-suspend-resume';
                const workloadName = objectGenerator.generateShortNameByPrefix('ui-cli-api-workload');
                let workloadId: string;

                await test.step('In the CLI, submit a training workload', async () => {
                    const submitFlags = {
                        [SubmitFlagsMerged.command]: '-- /bin/sh -c "sleep infinity"',
                        [SubmitFlagsMerged.project]: projectName,
                        [SubmitFlagsMerged.cpuCoreRequest]: '1',
                        [SubmitFlagsMerged.image]: config.quickStartImage.ubuntu
                    } as Partial<typeof SubmitFlagsMerged>;

                    // Added a CLI re-login as L1 researcher right before submitting the workload,
                    // to prevent other parallel tests from inadvertently switching the CLI session to a different user.
                    runaiCliV2.loginAsUser(
                        flowRootState.users.l1_researcher.email,
                        flowRootState.users.l1_researcher.password
                    );
                    const result = runaiCliV2.submit(submitFlags, workloadName, cliV2WorkloadType.training);
                    expect(result).toContain(workloadName);

                    workloadId = await workloadsUtilsV2.getWorkloadIdByName(workloadName, project.id);
                });

                await test.step('Verify workload is Running in CLI, API, and UI', async () => {
                    await runaiCliV2.waitForWorkloadPhaseWithSampling(
                        workloadName,
                        projectName,
                        WorkloadStatus.RUNNING,
                        cliV2WorkloadType.training,
                        120
                    );

                    await workloadsCrudUtils.waitForWorkloadPhase(
                        workloadId,
                        Phase.Running,
                        WorkloadType.TRAINING
                    );

                    await workloadPage.navigateToWorkloads();
                    await training.waitForWorkloadUIStatus(workloadName, WorkloadStatus.RUNNING);
                });

                await test.step('Suspend via CLI and verify in CLI, API, and UI', async () => {
                    const suspendResult = runaiCliV2.suspend(
                        { '-p': projectName } as any,
                        workloadName,
                        cliV2WorkloadType.training
                    );
                    expect(suspendResult).toContain(workloadName);

                    await runaiCliV2.waitForWorkloadPhaseWithSampling(
                        workloadName,
                        projectName,
                        WorkloadStatus.STOPPING,
                        cliV2WorkloadType.training,
                        30
                    );
                    await runaiCliV2.waitForWorkloadPhaseWithSampling(
                        workloadName,
                        projectName,
                        WorkloadStatus.STOPPED,
                        cliV2WorkloadType.training,
                        180
                    );

                    await workloadsCrudUtils.waitForWorkloadPhase(
                        workloadId,
                        Phase.Stopped,
                        WorkloadType.TRAINING
                    );

                    await training.waitForWorkloadUIStatus(workloadName, WorkloadStatus.STOPPED);
                });

                await test.step('Resume via CLI and verify in CLI, API, and UI', async () => {
                    const resumeResult = runaiCliV2.resume(
                        { '-p': projectName } as any,
                        workloadName,
                        cliV2WorkloadType.training
                    );
                    expect(resumeResult).toContain(workloadName);

                    await runaiCliV2.waitForWorkloadPhaseWithSampling(
                        workloadName,
                        projectName,
                        WorkloadStatus.RESUMING,
                        cliV2WorkloadType.training,
                        30
                    );
                    await runaiCliV2.waitForWorkloadPhaseWithSampling(
                        workloadName,
                        projectName,
                        WorkloadStatus.RUNNING,
                        cliV2WorkloadType.training,
                        180
                    );

                    await workloadsCrudUtils.waitForWorkloadPhase(
                        workloadId,
                        Phase.Running,
                        WorkloadType.TRAINING
                    );

                    await training.waitForWorkloadUIStatus(workloadName, WorkloadStatus.RUNNING);
                });

                await test.step('Cleanup', async () => {
                    await workloadsCrudUtils.deleteTraining(workloadId);
                });
            });

            test('T5187477 UI-CLI create, UI stop, API resume, verify cross-endpoint consistency', async () => {
                videoName = 'ui-cli-ui-stop-api-resume';
                const workloadName = objectGenerator.generateShortNameByPrefix('ui-cli-cross-endpoint');
                let workloadId: string;

                await test.step('In the CLI, submit a training workload', async () => {
                    const submitFlags = {
                        [SubmitFlagsMerged.command]: '-- /bin/sh -c "sleep infinity"',
                        [SubmitFlagsMerged.project]: projectName,
                        [SubmitFlagsMerged.cpuCoreRequest]: '1',
                        [SubmitFlagsMerged.image]: config.quickStartImage.ubuntu
                    } as Partial<typeof SubmitFlagsMerged>;

                    // Added a CLI re-login as L1 researcher right before submitting the workload,
                    // to prevent other parallel tests from inadvertently switching the CLI session to a different user.
                    runaiCliV2.loginAsUser(
                        flowRootState.users.l1_researcher.email,
                        flowRootState.users.l1_researcher.password
                    );
                    const result = runaiCliV2.submit(submitFlags, workloadName, cliV2WorkloadType.training);
                    expect(result).toContain(workloadName);

                    workloadId = await workloadsUtilsV2.getWorkloadIdByName(workloadName, project.id);
                });

                await test.step('Verify workload is Running in CLI, UI, and API', async () => {
                    await runaiCliV2.waitForWorkloadPhaseWithSampling(
                        workloadName,
                        projectName,
                        WorkloadStatus.RUNNING,
                        cliV2WorkloadType.training,
                        120
                    );

                    await workloadPage.navigateToWorkloads();
                    await training.waitForWorkloadUIStatus(workloadName, WorkloadStatus.RUNNING);

                    await workloadsCrudUtils.waitForWorkloadPhase(
                        workloadId,
                        Phase.Running,
                        WorkloadType.TRAINING
                    );
                });

                await test.step('Stop via UI and verify in UI, CLI, and API', async () => {
                    await workloadPage.navigateToWorkloads();
                    await suspendWorkloadUI(workloadName, false);

                    await runaiCliV2.waitForWorkloadPhaseWithSampling(
                        workloadName,
                        projectName,
                        WorkloadStatus.STOPPED,
                        cliV2WorkloadType.training,
                        30
                    );

                    await workloadsCrudUtils.waitForWorkloadPhase(
                        workloadId,
                        Phase.Stopped,
                        WorkloadType.TRAINING
                    );
                });

                await test.step('Resume via API and verify in API, CLI, and UI', async () => {
                    await workloadsCrudUtils.resumeTraining(workloadId);

                    await workloadsCrudUtils.waitForWorkloadPhase(
                        workloadId,
                        Phase.Resuming,
                        WorkloadType.TRAINING
                    );
                    await workloadsCrudUtils.waitForWorkloadPhase(
                        workloadId,
                        Phase.Running,
                        WorkloadType.TRAINING
                    );

                    await runaiCliV2.waitForWorkloadPhaseWithSampling(
                        workloadName,
                        projectName,
                        WorkloadStatus.RUNNING,
                        cliV2WorkloadType.training,
                        30
                    );

                    await training.waitForWorkloadUIStatus(workloadName, WorkloadStatus.RUNNING);
                });

                await test.step('Cleanup', async () => {
                    await workloadsCrudUtils.deleteTraining(workloadId);
                });
            });

            // Suspend/resume event and audit-log coverage remains disabled pending revalidation.
            test.skip(`T5187478: UI-Event and audit log generation for suspend/resume actions and phase transitions, debugTags: ${DEBUG_TAGS.PODS} ${DEBUG_TAGS.TRAINING_WORKLOAD_CRDS}`, async () => {
                test.setTimeout(300 * 1000);
                videoName = 'event-audit-log';
                const workloadName = objectGenerator.generateShortNameByPrefix('ui-event-log');

                await test.step('Create workload and wait for Running state', async () => {
                    await createAndVerifyWorkloadUI(workloadName, projectName, {
                        isDistributed: false,
                        image: config.quickStartImage.ubuntu,
                        command: '/bin/sh -c "echo hello && sleep infinity"'
                    });
                });

                await test.step('Perform suspend/resume to generate events', async () => {
                    training.stopWorkloadByName(workloadName);
                    await training.waitForWorkloadUIStatus(workloadName, WorkloadStatus.STOPPED);
                    training.activateWorkloadByName(workloadName);
                    await training.waitForWorkloadUIStatus(workloadName, WorkloadStatus.RUNNING);
                });

                await test.step('Verify event history includes correct order', async () => {
                    const text = await workloadPage.getEventHistoryText(workloadName);

                    const statusChangeLines = text
                        .split('\n')
                        .filter((line) => line.includes('Status changed'))
                        .map((line) => {
                            const parts = line.split('\t');
                            return parts[parts.length - 1].trim();
                        });

                    const chronologicalOrder = statusChangeLines.reverse();
                    logger.info(`Status change: ${chronologicalOrder.join(' → ')}`);
                    const keyStates = ['Running', 'Stopping...', 'Stopped', 'Resuming...'];
                    const filteredStates = chronologicalOrder.filter((status) => {
                        return keyStates.includes(status);
                    });

                    const expectedSequence = 'Running → Stopping... → Stopped → Resuming... → Running';
                    const actualSequence = filteredStates.join(' → ');
                    expect(actualSequence).toEqual(expectedSequence);
                });

                await test.step('Cleanup', async () => {
                    await training.deleteWorkload(workloadName);
                });
            });
        });
        test.describe('CLI cases', () => {
            let projectName: string;
            let namespace: string;
            let project: Project;

            test.beforeAll(async function () {
                projectName = objectGenerator.generateShortNameByPrefix('project-e2e-for-training');
                namespace = Utils.namespaceFromProjectName(projectName);
                project = await projectResourcesUtils.createProject(namespace, {
                    name: projectName,
                    deservedGpus: 0,
                    maxAllowedGpus: 0
                });
                runaiCliV2.loginAsUser(
                    flowRootState.users.l1_researcher.email,
                    flowRootState.users.l1_researcher.password
                );
                runaiCliV2.projectSet(projectName);
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

            async function waitForWorkloadTransition(
                workloadName: string,
                workloadType: cliV2WorkloadType,
                toPhase: WorkloadStatus,
                crdToPhase: WorkloadStatus,
                options: {
                    timeout?: number;
                    invalidPhase?: WorkloadStatus;
                    action?: () => void;
                } = {}
            ): Promise<void> {
                const { timeout = 120, invalidPhase, action } = options;

                const crdPromise =
                    workloadType === cliV2WorkloadType.training
                        ? waitUtils.waitForWrappedJobState(workloadName, namespace, crdToPhase, timeout)
                        : workloadUtils.waitForWorkloadBackendPhase(
                              workloadName,
                              projectName,
                              workloadStatusToPhase(toPhase),
                              timeout
                          );
                const cliPromise = runaiCliV2.waitForWorkloadPhaseWithSampling(
                    workloadName,
                    projectName,
                    toPhase,
                    workloadType,
                    timeout,
                    invalidPhase
                );

                if (action) {
                    await new Promise((resolve) => setTimeout(resolve, 1000));
                    action();
                }
                await Promise.all([cliPromise, crdPromise]);
            }

            async function createAndVerifyWorkloadCLI(
                workloadName: string,
                workloadType: cliV2WorkloadType,
                image: string,
                options: { workers?: number } = {}
            ): Promise<void> {
                const submitFlags = {
                    [SubmitFlagsMerged.command]: '-- /bin/sh -c "sleep infinity"',
                    [SubmitFlagsMerged.project]: projectName,
                    [SubmitFlagsMerged.cpuCoreRequest]: '1',
                    [SubmitFlagsMerged.image]: image
                } as Partial<typeof SubmitFlagsMerged>;

                if (options.workers) {
                    submitFlags[SubmitFlagsMerged.workers] = options.workers.toString();
                }

                const result = runaiCliV2.submit(submitFlags, workloadName, workloadType);
                expect(result).toContain(workloadName);

                await runaiCliV2.waitForWorkloadPhaseWithSampling(
                    workloadName,
                    projectName,
                    WorkloadStatus.RUNNING,
                    workloadType,
                    120
                );
                if (workloadType === cliV2WorkloadType.training) {
                    await waitUtils.waitForWrappedJobState(
                        workloadName,
                        namespace,
                        WorkloadStatus.RUNNING,
                        120
                    );
                } else {
                    await workloadUtils.waitForDistributedSubmissionRoot(workloadName, namespace);
                    await waitUtils.waitForCrd(
                        getConcreteDistributedJobCrd(workloadType),
                        workloadName,
                        namespace
                    );
                    await waitUtils.waitForPods(workloadName, namespace);
                }
            }

            async function suspendWorkloadCLI(
                workloadName: string,
                workloadType: cliV2WorkloadType
            ): Promise<void> {
                await test.step('Suspend the workload and check workload state transitions to Stopping/Suspending', async () => {
                    await waitForWorkloadTransition(
                        workloadName,
                        workloadType,
                        WorkloadStatus.STOPPING,
                        WorkloadStatus.SUSPENDING,
                        {
                            timeout: 30,
                            action: () =>
                                runaiCliV2.suspend({ '-p': projectName } as any, workloadName, workloadType)
                        }
                    );
                });

                await test.step('Check workload state transitions to Stopped/Suspended', async () => {
                    await waitForWorkloadTransition(
                        workloadName,
                        workloadType,
                        WorkloadStatus.STOPPED,
                        WorkloadStatus.SUSPENDED,
                        {
                            timeout: 180,
                            invalidPhase: WorkloadStatus.RUNNING
                        }
                    );
                });

                await test.step('Verify k8s pods do not exist', async () => {
                    await waitUtils.waitForPodsDeletion(workloadName, namespace, 120);
                });
            }

            async function resumeWorkloadCLI(
                workloadName: string,
                workloadType: cliV2WorkloadType
            ): Promise<void> {
                await test.step('Resume the workload and check workload state transitions to Resuming', async () => {
                    await waitForWorkloadTransition(
                        workloadName,
                        workloadType,
                        WorkloadStatus.RESUMING,
                        WorkloadStatus.RESUMING,
                        {
                            timeout: 30,
                            action: () =>
                                runaiCliV2.resume({ '-p': projectName } as any, workloadName, workloadType)
                        }
                    );
                });

                await test.step('Check workload state transitions to Running', async () => {
                    await waitForWorkloadTransition(
                        workloadName,
                        workloadType,
                        WorkloadStatus.RUNNING,
                        WorkloadStatus.RUNNING,
                        {
                            timeout: 180
                        }
                    );
                });

                await test.step('Verify k8s pods are recreated', async () => {
                    const pods = await kubeCli.listPodsByNamepsace(namespace);
                    expect(
                        pods.filter((p) => p.metadata?.name?.includes(workloadName)).length
                    ).toBeGreaterThan(0);
                });
            }

            async function deleteWorkloadCLI(
                workloadName: string,
                workloadType: cliV2WorkloadType
            ): Promise<void> {
                await test.step('Delete the workload', () => {
                    const deleteResult = runaiCliV2.delete(
                        { '-p': projectName } as any,
                        workloadName,
                        workloadType
                    );
                    expect(deleteResult).toContain(workloadName);
                });
            }

            async function runCliSuspendResumeTest(
                workloadType: cliV2WorkloadType,
                image: string,
                workers?: number
            ): Promise<void> {
                const maxRetries = 3;
                let lastError: any = null;
                let workloadName: string;
                let attempt = 0;
                for (attempt = 1; attempt <= maxRetries; attempt++) {
                    try {
                        const prefix =
                            workloadType === cliV2WorkloadType.training
                                ? 'cli-training'
                                : workloadType === cliV2WorkloadType.mpi
                                  ? 'cli-mpi'
                                  : workloadType === cliV2WorkloadType.pyTorch
                                    ? 'cli-pytorch'
                                    : workloadType === cliV2WorkloadType.tensorFlow
                                      ? 'cli-tensorflow'
                                      : workloadType === cliV2WorkloadType.xgBoost
                                        ? 'cli-xgboost'
                                        : workloadType === cliV2WorkloadType.jax
                                          ? 'cli-jax'
                                          : 'cli-training';
                        workloadName = objectGenerator.generateShortNameByPrefix(prefix);
                        await createAndVerifyWorkloadCLI(workloadName, workloadType, image, { workers });
                        await suspendWorkloadCLI(workloadName, workloadType);
                        await resumeWorkloadCLI(workloadName, workloadType);
                        await deleteWorkloadCLI(workloadName, workloadType);
                        return;
                    } catch (error) {
                        lastError = error;
                        const errorMessage = error.toString();
                        logger.error(`Attempt ${attempt}/${maxRetries} failed with error: ${errorMessage}`);

                        if (
                            errorMessage.includes('Phase missing: [Suspending]') ||
                            errorMessage.includes('Phase missing: [Resuming]')
                        ) {
                            logger.warn(
                                `Attempt ${attempt}/${maxRetries} failed with retryable error: ${errorMessage}`
                            );
                            try {
                                await deleteWorkloadCLI(workloadName, workloadType);
                            } catch (cleanupError) {
                                logger.warn(`Failed to cleanup workload ${workloadName}: ${cleanupError}`);
                            }
                        } else {
                            logger.error(
                                `Attempt ${attempt}/${maxRetries} failed with error: ${errorMessage}`
                            );
                            throw error;
                        }
                    }
                }

                throw new Error(`Test failed after ${maxRetries} attempts. Last error: ${lastError}`);
            }

            test('T5187451: CLI-Suspend and resume a single node workload', async () => {
                test.setTimeout(300 * 1000);
                await runCliSuspendResumeTest(cliV2WorkloadType.training, config.quickStartImage.ubuntu);
            });

            // MPI suspend/resume coverage remains disabled pending revalidation.
            test.skip('T5187453: CLI-Suspend and resume a MPI distributed node workload', async () => {
                test.setTimeout(300 * 1000);
                await runCliSuspendResumeTest(cliV2WorkloadType.mpi, config.quickStartImage.ubuntu, 10);
            });

            test('T5187454: CLI-Suspend and resume a pytorch distributed node workload', async () => {
                test.setTimeout(300 * 1000);
                await runCliSuspendResumeTest(cliV2WorkloadType.pyTorch, config.quickStartImage.ubuntu, 10);
            });

            test('T5187455: CLI-Suspend and resume a tensorflow distributed node workload', async () => {
                test.setTimeout(300 * 1000);
                await runCliSuspendResumeTest(
                    cliV2WorkloadType.tensorFlow,
                    config.quickStartImage.ubuntu,
                    10
                );
            });

            test('T5187456: CLI-Suspend and resume an xgboost distributed node workload', async () => {
                test.setTimeout(300 * 1000);
                await runCliSuspendResumeTest(cliV2WorkloadType.xgBoost, config.quickStartImage.ubuntu, 10);
            });

            test('T5193824: CLI-Suspend and resume an JAX distributed node workload', async () => {
                test.setTimeout(300 * 1000);
                await runCliSuspendResumeTest(cliV2WorkloadType.jax, config.quickStartImage.ubuntu, 10);
            });

            test('T5187464: CLI-Attempt to resume a workload in Running state', async () => {
                const workloadName = objectGenerator.generateShortNameByPrefix('cli-running');

                await test.step('Create and verify workload', () => {
                    return createAndVerifyWorkloadCLI(
                        workloadName,
                        cliV2WorkloadType.training,
                        config.quickStartImage.ubuntu
                    );
                });

                await test.step('Try to resume and check result', () => {
                    try {
                        runaiCliV2.resume(
                            { '-p': projectName } as any,
                            workloadName,
                            cliV2WorkloadType.training
                        );
                        throw new Error('Expected resume to fail but it succeeded');
                    } catch (error) {
                        const errorMessage = error.toString();
                        logger.info(`Expected error: ${errorMessage}`);
                        expect(errorMessage).toContain(ERROR_MESSAGES.FAILED_TO_RESUME_WORKLOAD);
                        expect(errorMessage).toContain(ERROR_MESSAGES.RESUME_ONLY_FOR_STOPPED);
                        expect(errorMessage).toContain(ERROR_MESSAGES.WORKLOAD_IN_RUNNING_PHASE);
                    }
                });

                await test.step('Cleanup', async () => {
                    await deleteWorkloadCLI(workloadName, cliV2WorkloadType.training);
                });
            });

            test('T5187465: CLI-Attempt to suspend a workload in Stopped state', async () => {
                const workloadName = objectGenerator.generateShortNameByPrefix('cli-stopped');

                await test.step('Create and verify workload', () => {
                    return createAndVerifyWorkloadCLI(
                        workloadName,
                        cliV2WorkloadType.training,
                        config.quickStartImage.ubuntu
                    );
                });

                await test.step('Suspend once', async () => {
                    await suspendWorkloadCLI(workloadName, cliV2WorkloadType.training);
                });

                await test.step('Suspend again and check result', () => {
                    try {
                        runaiCliV2.suspend(
                            { '-p': projectName } as any,
                            workloadName,
                            cliV2WorkloadType.training
                        );
                        throw new Error('Expected suspend to fail but it succeeded');
                    } catch (error) {
                        const errorMessage = error.toString();
                        expect(errorMessage).toContain(ERROR_MESSAGES.FAILED_TO_SUSPEND_WORKLOAD);
                        expect(errorMessage).toContain(ERROR_MESSAGES.SUSPEND_NOT_PERMITTED_FOR_STOPPED);
                    }
                });

                await test.step('Cleanup', async () => {
                    await deleteWorkloadCLI(workloadName, cliV2WorkloadType.training);
                });
            });
            async function runCliRapidSuspendResumeTest(endingState: WorkloadStatus): Promise<void> {
                test.setTimeout(300 * 1000);
                const stateName = endingState === WorkloadStatus.RUNNING ? 'running' : 'stopped';
                const workloadName = objectGenerator.generateShortNameByPrefix(`cli-rapid-${stateName}`);

                await test.step('Create and verify workload', () => {
                    return createAndVerifyWorkloadCLI(
                        workloadName,
                        cliV2WorkloadType.training,
                        config.quickStartImage.ubuntu
                    );
                });

                for (let i = 0; i < 3; i++) {
                    await test.step(`Rapid suspend/resume cycle ${i + 1}`, async () => {
                        runaiCliV2.suspend(
                            { '-p': projectName } as any,
                            workloadName,
                            cliV2WorkloadType.training
                        );
                        await runaiCliV2.waitForWorkloadPhaseWithSampling(
                            workloadName,
                            projectName,
                            WorkloadStatus.STOPPED,
                            cliV2WorkloadType.training,
                            120
                        );
                        runaiCliV2.resume(
                            { '-p': projectName } as any,
                            workloadName,
                            cliV2WorkloadType.training
                        );
                        await runaiCliV2.waitForWorkloadPhaseWithSampling(
                            workloadName,
                            projectName,
                            WorkloadStatus.RUNNING,
                            cliV2WorkloadType.training,
                            120
                        );
                    });
                }

                if (endingState === WorkloadStatus.STOPPED) {
                    await test.step('Final suspend to Stopped state', async () => {
                        runaiCliV2.suspend(
                            { '-p': projectName } as any,
                            workloadName,
                            cliV2WorkloadType.training
                        );
                        await runaiCliV2.waitForWorkloadPhaseWithSampling(
                            workloadName,
                            projectName,
                            WorkloadStatus.STOPPED,
                            cliV2WorkloadType.training,
                            120
                        );
                    });
                }

                await test.step(`Wait for 30 seconds and verify ${endingState} state`, async () => {
                    await new Promise((resolve) => setTimeout(resolve, 30000));
                    const expectedCrdState =
                        endingState === WorkloadStatus.STOPPED
                            ? WorkloadStatus.SUSPENDED
                            : WorkloadStatus.RUNNING;
                    await waitForWorkloadTransition(
                        workloadName,
                        cliV2WorkloadType.training,
                        endingState,
                        expectedCrdState
                    );
                });

                await test.step('Cleanup', async () => {
                    await deleteWorkloadCLI(workloadName, cliV2WorkloadType.training);
                });
            }

            test('T5187468: CLI-Rapid suspend/resume operations ending in Running state', async () => {
                await runCliRapidSuspendResumeTest(WorkloadStatus.RUNNING);
            });

            test('T5187469: CLI-Rapid suspend/resume operations ending in Stopped state', async () => {
                await runCliRapidSuspendResumeTest(WorkloadStatus.STOPPED);
            });
            test('T5187481: CLI-Suspend a completed workload via CLI', async () => {
                test.setTimeout(300 * 1000);
                const workloadName = objectGenerator.generateShortNameByPrefix('cli-suspend-complete');

                await test.step('Create a short-lived workload', () => {
                    const submitFlags = {
                        [SubmitFlagsMerged.command]: '-- /bin/sh -c "sleep 1"',
                        [SubmitFlagsMerged.project]: projectName,
                        [SubmitFlagsMerged.cpuCoreRequest]: '1',
                        [SubmitFlagsMerged.image]: config.quickStartImage.ubuntu
                    } as Partial<typeof SubmitFlagsMerged>;

                    const result = runaiCliV2.submit(submitFlags, workloadName, cliV2WorkloadType.training);
                    expect(result).toContain(workloadName);
                });

                await test.step('Wait until the workload is in Completed state', async function () {
                    await waitForWorkloadTransition(
                        workloadName,
                        cliV2WorkloadType.training,
                        WorkloadStatus.COMPLETED,
                        WorkloadStatus.COMPLETED
                    );
                });

                await test.step('Try to suspend the workload', () => {
                    try {
                        runaiCliV2.suspend(
                            { '-p': projectName } as any,
                            workloadName,
                            cliV2WorkloadType.training
                        );
                        throw new Error('Expected suspend to fail but it succeeded');
                    } catch (error) {
                        const errorMessage = error.toString();
                        expect(errorMessage).toContain(ERROR_MESSAGES.SUSPEND_NOT_PERMITTED_FOR_COMPLETED);
                    }
                });

                await test.step('Wait for 30 seconds and verify Completed state', async function () {
                    await new Promise((resolve) => setTimeout(resolve, 30000));
                    await waitForWorkloadTransition(
                        workloadName,
                        cliV2WorkloadType.training,
                        WorkloadStatus.COMPLETED,
                        WorkloadStatus.COMPLETED
                    );
                });

                await test.step('Cleanup', async () => {
                    await deleteWorkloadCLI(workloadName, cliV2WorkloadType.training);
                });
            });

            test('T5187482: CLI-Suspend a failed workload via CLI', async () => {
                test.setTimeout(300 * 1000);
                const workloadName = objectGenerator.generateShortNameByPrefix('cli-suspend-failed');

                await test.step('Create a short-lived workload with invalid command', () => {
                    const submitFlags = {
                        [SubmitFlagsMerged.command]: '-- /bin/sh -c "invalidcmd"',
                        [SubmitFlagsMerged.project]: projectName,
                        [SubmitFlagsMerged.cpuCoreRequest]: '1',
                        [SubmitFlagsMerged.image]: config.quickStartImage.ubuntu,
                        // --backoff-limit 0 -> the Job reaches the terminal Failed condition immediately
                        // (no retries), so the batch-Job Failed wait doesn't hang on backoff.
                        [SubmitFlagsMerged.backoffLimit]: '0'
                    } as Partial<typeof SubmitFlagsMerged>;

                    const result = runaiCliV2.submit(submitFlags, workloadName, cliV2WorkloadType.training);
                    expect(result).toContain(workloadName);
                });

                await test.step('Wait until the workload is in Failed state', async function () {
                    await waitForWorkloadTransition(
                        workloadName,
                        cliV2WorkloadType.training,
                        WorkloadStatus.FAILED,
                        WorkloadStatus.FAILED
                    );
                });

                await test.step('Try to suspend the workload', () => {
                    try {
                        runaiCliV2.suspend(
                            { '-p': projectName } as any,
                            workloadName,
                            cliV2WorkloadType.training
                        );
                        throw new Error('Expected suspend to fail but it succeeded');
                    } catch (error) {
                        const errorMessage = error.toString();
                        expect(errorMessage).toContain(ERROR_MESSAGES.SUSPEND_NOT_PERMITTED_FOR_FAILED);
                    }
                });

                await test.step('Wait for 30 seconds and verify Failed state', async function () {
                    await new Promise((resolve) => setTimeout(resolve, 30000));
                    await waitForWorkloadTransition(
                        workloadName,
                        cliV2WorkloadType.training,
                        WorkloadStatus.FAILED,
                        WorkloadStatus.FAILED
                    );
                });

                await test.step('Cleanup', async () => {
                    await deleteWorkloadCLI(workloadName, cliV2WorkloadType.training);
                });
            });
        });

        test.describe('API cases', () => {
            let projectName: string;
            let namespace: string;
            let project: Project;

            test.beforeAll(async function () {
                projectName = objectGenerator.generateShortNameByPrefix('project-e2e-for-training');
                namespace = Utils.namespaceFromProjectName(projectName);
                project = await projectResourcesUtils.createProject(namespace, {
                    name: projectName,
                    deservedGpus: 0,
                    maxAllowedGpus: 0
                });
                runaiCliV2.projectSet(projectName);
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

            async function createAndVerifyWorkspace(
                workspaceName: string,
                command: string = 'sleep infinity'
            ): Promise<{ workspace: any; workloadId: string }> {
                const clusterId = flowRootState.cluster.uuid;
                const projectId = project.id;

                workloadsCrudUtils.accessToken = flowRootState.users.l1_researcher.accessToken;
                const { body: workspace } = await workloadsCrudUtils.createWorkspace(
                    workspaceName,
                    true,
                    clusterId,
                    String(projectId),
                    { command, image: config.quickStartImage.ubuntu }
                );
                const workloadId = workspace.workloadId;
                await waitUtils.waitForCrd(crds.workloadRunnerCrd, workspace.name, namespace);
                await waitUtils.waitForWrappedJobState(
                    workspace.name,
                    namespace,
                    WorkloadStatus.RUNNING,
                    120
                );
                await workloadsCrudUtils.waitForWorkloadPhase(
                    workloadId,
                    Phase.Running,
                    WorkloadType.WORKSPACE
                );
                return { workspace, workloadId };
            }

            async function verifyWorkspaceState(
                workspace: any,
                workloadId: string,
                expectedPhase: Phase,
                expectedCrdPhase: WorkloadStatus
            ): Promise<void> {
                await workloadsCrudUtils.waitForWorkloadPhase(workloadId, expectedPhase);
                await waitUtils.waitForWrappedJobState(workspace.name, namespace, expectedCrdPhase, 120);
            }

            async function suspendWorkspace(workspace: any, workloadId: string): Promise<void> {
                await test.step('Suspend workspace and verify Stopping/Suspending', async () => {
                    // Suspending is a transient phase that is not persisted on the wrapped Job;
                    // it is verified via the API phase (Stopping) only.
                    const apiPromise = workloadsCrudUtils.waitForWorkloadPhase(workloadId, Phase.Stopping);
                    await new Promise((resolve) => setTimeout(resolve, 1000));
                    await workloadsCrudUtils.suspendWorkspace(workloadId);
                    await apiPromise;
                });

                await test.step('Wait for Stopped/Suspended', async () => {
                    await waitUtils.waitForWrappedJobState(
                        workspace.name,
                        namespace,
                        WorkloadStatus.SUSPENDED,
                        120
                    );
                    await workloadsCrudUtils.waitForWorkloadPhase(
                        workloadId,
                        Phase.Stopped,
                        WorkloadType.WORKSPACE,
                        Phase.Running
                    );
                });
            }

            async function resumeWorkspace(workspace: any, workloadId: string): Promise<void> {
                await test.step('Resume workspace and verify Resuming', async () => {
                    // Resuming is a transient phase that is not persisted on the wrapped Job;
                    // it is verified via the API phase (Resuming) only.
                    const apiPromise = workloadsCrudUtils.waitForWorkloadPhase(workloadId, Phase.Resuming);
                    await new Promise((resolve) => setTimeout(resolve, 1000));
                    await workloadsCrudUtils.resumeWorkspace(workloadId);
                    await apiPromise;
                });

                await test.step('Wait for Running', async () => {
                    await workloadsCrudUtils.waitForWorkloadPhase(workloadId, Phase.Running);
                    await waitUtils.waitForWrappedJobState(
                        workspace.name,
                        namespace,
                        WorkloadStatus.RUNNING,
                        120
                    );
                });
            }

            async function expectOperationToFail(
                operation: () => Promise<any>,
                operationType: 'resume' | 'suspend'
            ): Promise<void> {
                try {
                    await operation();
                    throw new Error('Expected operation to fail but it succeeded');
                } catch (error) {
                    const errorMessage = error.toString();
                    logger.info(`Expected error: ${errorMessage}`);

                    if (operationType === 'resume') {
                        expect(errorMessage).toContain(ERROR_MESSAGES.RESUME_ONLY_FOR_STOPPED);
                        expect(errorMessage).toContain(ERROR_MESSAGES.WORKLOAD_IN_RUNNING_PHASE);
                    } else {
                        expect(errorMessage).toContain(ERROR_MESSAGES.SUSPEND_NOT_PERMITTED_FOR_STOPPED);
                    }
                }
            }

            test('T5187452: API-Suspend and resume a workspace', async () => {
                const maxRetries = 3;
                let lastError: any = null;
                let workloadName: string;
                let workloadId: string;
                let attempt = 0;

                for (attempt = 1; attempt <= maxRetries; attempt++) {
                    try {
                        workloadName = objectGenerator.generateShortNameByPrefix('api-workspace');
                        const { workspace, workloadId: id } =
                            await test.step('Create workspace and verify Running', () =>
                                createAndVerifyWorkspace(workloadName));
                        workloadId = id;
                        await suspendWorkspace(workspace, workloadId);
                        await resumeWorkspace(workspace, workloadId);
                        await workloadsCrudUtils.deleteWorkspace(workloadId);
                        return;
                    } catch (error) {
                        lastError = error;
                        const errorMessage = error.toString();
                        logger.error(`Attempt ${attempt}/${maxRetries} failed with error: ${errorMessage}`);

                        if (
                            errorMessage.includes('Phase missing:[Suspending]') ||
                            errorMessage.includes('Phase missing:[Resuming]')
                        ) {
                            logger.warn(
                                `Attempt ${attempt}/${maxRetries} failed with retryable error: ${errorMessage}`
                            );
                            try {
                                if (workloadId) {
                                    await workloadsCrudUtils.deleteWorkspace(workloadId);
                                }
                            } catch (cleanupError) {
                                logger.warn(`Failed to cleanup workspace ${workloadName}: ${cleanupError}`);
                            }
                        } else {
                            logger.error(
                                `Attempt ${attempt}/${maxRetries} failed with error: ${errorMessage}`
                            );
                            throw error;
                        }
                    }
                }

                throw new Error(`Test failed after ${maxRetries} attempts. Last error: ${lastError}`);
            });

            test('T5187466: API-Attempt to resume a workload in Running state', async () => {
                const workspaceName = objectGenerator.generateShortNameByPrefix('api-running');
                const { workspace, workloadId } = await test.step('Create workspace and verify Running', () =>
                    createAndVerifyWorkspace(workspaceName));

                await test.step('Attempt to resume workspace in Running state', () =>
                    expectOperationToFail(() => workloadsCrudUtils.resumeWorkspace(workloadId), 'resume'));

                await test.step('Verify workspace is still in Running state', () =>
                    verifyWorkspaceState(workspace, workloadId, Phase.Running, WorkloadStatus.RUNNING));

                await workloadsCrudUtils.deleteWorkspace(workloadId);
            });

            test('T5187467: API-Attempt to suspend a workload in Stopped state', async () => {
                const workspaceName = objectGenerator.generateShortNameByPrefix('api-stopped');
                const { workspace, workloadId } = await test.step('Create workspace and verify Running', () =>
                    createAndVerifyWorkspace(workspaceName));

                await suspendWorkspace(workspace, workloadId);

                await test.step('Attempt to suspend workspace in Stopped state', () =>
                    expectOperationToFail(() => workloadsCrudUtils.suspendWorkspace(workloadId), 'suspend'));

                await test.step('Verify workspace is still in Stopped state', () =>
                    verifyWorkspaceState(workspace, workloadId, Phase.Stopped, WorkloadStatus.SUSPENDED));

                await workloadsCrudUtils.deleteWorkspace(workloadId);
            });

            async function runApiRapidSuspendResumeTest(endingPhase: Phase): Promise<void> {
                test.setTimeout(300 * 1000);
                const phaseName = endingPhase === Phase.Running ? 'running' : 'stopped';
                const workspaceName = objectGenerator.generateShortNameByPrefix(`api-rapid-${phaseName}`);
                const { workspace, workloadId } = await test.step('Create workspace and verify Running', () =>
                    createAndVerifyWorkspace(workspaceName));

                for (let i = 0; i < 3; i++) {
                    await test.step(`Rapid suspend/resume cycle ${i + 1}`, async () => {
                        await workloadsCrudUtils.suspendWorkspace(workloadId);
                        await workloadsCrudUtils.waitForWorkloadPhase(workloadId, Phase.Stopped);
                        await workloadsCrudUtils.resumeWorkspace(workloadId);
                        await workloadsCrudUtils.waitForWorkloadPhase(workloadId, Phase.Running);
                    });
                }

                if (endingPhase === Phase.Stopped) {
                    await workloadsCrudUtils.suspendWorkspace(workloadId);
                    await workloadsCrudUtils.waitForWorkloadPhase(workloadId, Phase.Stopped);
                }

                await test.step(`Wait for 30 seconds and verify ${endingPhase} state`, async () => {
                    await new Promise((resolve) => setTimeout(resolve, 30000));
                    const expectedCrdPhase =
                        endingPhase === Phase.Stopped ? WorkloadStatus.SUSPENDED : WorkloadStatus.RUNNING;
                    await verifyWorkspaceState(workspace, workloadId, endingPhase, expectedCrdPhase);
                });

                await workloadsCrudUtils.deleteWorkspace(workloadId);
            }

            test('T5187472: API-Rapid suspend/resume operations ending in Running state', async () => {
                await runApiRapidSuspendResumeTest(Phase.Running);
            });

            test('T5187473: API-Rapid suspend/resume operations ending in Stopped state', async () => {
                await runApiRapidSuspendResumeTest(Phase.Stopped);
            });

            async function runApiFiniteTerminalStateTest(finalPhase: Phase, command: string): Promise<void> {
                const phaseName = finalPhase === Phase.Completed ? 'complete' : 'failed';
                const workspaceName = objectGenerator.generateShortNameByPrefix(`api-suspend-${phaseName}`);
                const clusterId = flowRootState.cluster.uuid;
                const projectId = project.id;

                workloadsCrudUtils.accessToken = flowRootState.users.l1_researcher.accessToken;
                const result = await workloadsCrudUtils.createWorkspace(
                    workspaceName,
                    true,
                    clusterId,
                    String(projectId),
                    // backoffLimit: 0 -> a failing command reaches the terminal Failed condition
                    // immediately (no retries), so the finite-state wait doesn't hang on backoff.
                    { command, image: config.quickStartImage.ubuntu, backoffLimit: 0 }
                );
                const workspace = result.body;

                const stepDescription =
                    finalPhase === Phase.Completed
                        ? 'Wait for workspace to complete'
                        : 'Wait for workspace to fail';
                const expectedWorkloadStatus =
                    finalPhase === Phase.Completed ? WorkloadStatus.COMPLETED : WorkloadStatus.FAILED;

                await test.step(stepDescription, () =>
                    verifyWorkspaceState(workspace, workspace.workloadId, finalPhase, expectedWorkloadStatus)
                );

                await test.step(`Attempt to suspend ${phaseName}d workspace`, async () => {
                    try {
                        await workloadsCrudUtils.suspendWorkspace(workspace.workloadId);
                        throw new Error('Expected suspend to fail but it succeeded');
                    } catch (error) {
                        const errorMessage = error.toString();
                        const expectedError =
                            finalPhase === Phase.Completed
                                ? ERROR_MESSAGES.SUSPEND_NOT_PERMITTED_FOR_COMPLETED
                                : ERROR_MESSAGES.SUSPEND_NOT_PERMITTED_FOR_FAILED;
                        expect(errorMessage).toContain(expectedError);
                    }
                });

                await test.step(`Wait for 30 seconds and verify ${finalPhase} state`, async () => {
                    await new Promise((resolve) => setTimeout(resolve, 30000));
                    await verifyWorkspaceState(
                        workspace,
                        workspace.workloadId,
                        finalPhase,
                        expectedWorkloadStatus
                    );
                });

                await workloadsCrudUtils.deleteWorkspace(workspace.workloadId);
            }

            test('T5187479: API-Suspend a completed workload via API', async () => {
                await runApiFiniteTerminalStateTest(Phase.Completed, 'sleep 1');
            });

            test('T5187480: API-Suspend a failed workload via API', async () => {
                await runApiFiniteTerminalStateTest(Phase.Failed, 'invalidcmd');
            });
        });
    });
});

test.describe(`Training UI - sanity, tags: ${tags.resource.workloads} ${tags.teams.loong} ${tags.runOn.dgxc}`, () => {
    supportedFrom(SupportedVersions.V2_15)(function () {
        let browserEngine: BrowserEngine;
        let auth: AuthPage;
        let training: TrainingPage;
        let environment: EnvironmentPage;
        let videoName: string;
        let projectName: string;
        let namespace: string;
        let project: Project;
        let trainingName: string;
        let environmentName: string;

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
            environment = new EnvironmentPage(browserEngine.currentPage);
            trainingName = objectGenerator.generateShortNameByPrefix('training');
            environmentName = objectGenerator.generateShortNameByPrefix('environment');

            await auth.signIn(
                flowRootState.users.l1_researcher.email,
                flowRootState.users.l1_researcher.password
            );
        });

        test.afterEach('stop browser', async ({}, testInfo) => {
            await browserEngine.stopAndArchive(testInfo, 'training-' + videoName);
        });

        test('T3738316 & T3738319: L1 researcher can stop and resume training', async () => {
            videoName = 'stop-and-resume-training';
            if (settings.isSelfHosted) test.setTimeout(180 * 1000);
            await training.createFlexibleTraining(projectName, trainingName);
            await training.waitForWorkloadUIStatus(trainingName, EWorkloadStatus.Running);

            await training.stopWorkloadByName(trainingName);
            await training.waitForWorkloadUIStatus(trainingName, EWorkloadStatus.Stopped);

            await training.activateWorkloadByName(trainingName);
            await training.waitForWorkloadUIStatus(trainingName, EWorkloadStatus.Running);

            await training.deleteWorkload(trainingName);
        });

        test.describe('MPI distributed workload hybrid submission', () => {
            supportedFrom(SupportedVersions.V2_20)(function () {
                // skip this test as it is already covered by other tests
                test.skip('T4142370: L1 researcher can stop and run MPI distributed job', async () => {
                    videoName = 'stop-and-run-mpi-distributed-job';
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
                            command: 'sleep infinity',
                            args: '',
                            computeLabel: 'cpu-only',
                            workersCount: '2',
                            imagePullPolicy: 'Always pull the image from the registry'
                        }
                    );
                    await training.waitForWorkloadUIStatus(trainingName, EWorkloadStatus.Running);

                    await training.stopWorkloadByName(trainingName);
                    await training.waitForWorkloadUIStatus(trainingName, EWorkloadStatus.Stopped);

                    await training.activateWorkloadByName(trainingName);
                    await training.waitForWorkloadUIStatus(trainingName, EWorkloadStatus.Running);
                    await training.deleteWorkload(trainingName);
                });
            });
        });

        test.describe('Distributed training stop and run tests for different frameworks', () => {
            supportedFrom(SupportedVersions.V2_20)(function () {
                test.setTimeout(300 * 1000);

                const distributedFrameworkTests = [
                    {
                        testCase: 'T4142371: L1 researcher can stop and run pytorch distributed job',
                        videoName: 'stop-and-run-pytorch-distributed-job',
                        distributedType: DISTRIBUTED_TOOLS_LABELS.PY_TORCH,
                        image: config.quickStartImage.distributedPytorch,
                        command: '/bin/bash -c',
                        args: '"sleep 300"',
                        computeLabel: 'one-gpu',
                        workersCount: '1'
                    },
                    {
                        testCase: 'T4142372: L1 researcher can stop and run tensorflow job',
                        videoName: 'stop-and-run-tensorflow-distributed-job',
                        distributedType: DISTRIBUTED_TOOLS_LABELS.TENSOR_FLOW,
                        image: config.quickStartImage.distributedTf,
                        command: '/bin/bash -c',
                        args: '"sleep 300"',
                        computeLabel: 'one-gpu',
                        workersCount: '1'
                    },
                    {
                        testCase: 'T4142373: L1 researcher can stop and run xgboost distributed job',
                        videoName: 'stop-and-run-xgboost-distributed-job',
                        distributedType: DISTRIBUTED_TOOLS_LABELS.XG_BOOST,
                        image: 'docker.io/kubeflow/lightgbm-dist-py-test',
                        command: '/bin/bash -c',
                        args: '"sleep 300"',
                        computeLabel: 'one-gpu',
                        workersCount: '2'
                    }
                ];

                distributedFrameworkTests.forEach((testData) => {
                    test(testData.testCase, async () => {
                        videoName = testData.videoName;
                        const currentTrainingName = objectGenerator.generateShortNameByPrefix('training');

                        await test.step(`Create ${testData.distributedType} distributed job`, async () => {
                            await training.fillFirstNewTrainingPage(
                                projectName,
                                currentTrainingName,
                                true,
                                testData.distributedType,
                                true,
                                true
                            );
                            await training.fillHybridDistributedTrainingSecondPage(testData.image, {
                                command: testData.command,
                                args: testData.args,
                                computeLabel: testData.computeLabel,
                                workersCount: testData.workersCount,
                                imagePullPolicy: 'Pull the image only if it'
                            });
                            await training.waitForWorkloadUIStatus(
                                currentTrainingName,
                                EWorkloadStatus.Running
                            );
                        });

                        await test.step(`Stop ${testData.distributedType} distributed job`, async () => {
                            await training.stopWorkloadByName(currentTrainingName);
                            await training.waitForWorkloadUIStatus(
                                currentTrainingName,
                                EWorkloadStatus.Stopped
                            );
                        });

                        await test.step(`Resume ${testData.distributedType} distributed job`, async () => {
                            await training.activateWorkloadByName(currentTrainingName);
                            await training.waitForWorkloadUIStatus(
                                currentTrainingName,
                                EWorkloadStatus.Running
                            );
                        });

                        await test.step(`Delete ${testData.distributedType} distributed job`, async () => {
                            await training.deleteWorkload(currentTrainingName);
                        });
                    });
                });
            });
        });
    });
});
