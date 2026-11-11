/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import {
    flowRootState,
    logger,
    tags,
    operations,
    projectResourcesUtils,
    workloadsCrudUtils,
    workloadsManagerUtils,
    waitUtils,
    crds,
    versionUtils,
    config
} from '../../../../../../modules/helpers';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { WorkloadStatus } from '../../../../../../modules/helpers/models/baseWorkload';
import { StatusCodes } from 'http-status-codes';
import { isKnativeInstalled } from '../../../../../../modules/helpers/utils/IsInstalledUtils';
import { test } from '@playwright/test';
import { skipPlaywrightTest } from '../../../../../../modules/helpers/utils/SkippedTestUtils';
import {
    AnyWorkloadSubmissionType,
    isAnyWorkloadSubmissionEnabled
} from '../../../../../../modules/helpers/utils/AnyWorkloadSubmissionUtils';
import { WorkloadType } from '../../../../../../modules/helpers/utils/WorkloadsCrudUtils';
import { Phase } from '../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import { WorkloadDesiredPhase } from '../../../../../../modules/api/apis/workloadsApi/generated/model/workloadDesiredPhase';
import { DesiredPhase } from '../../../../../../modules/api/apis/workloadsManagerApi/generated/model/desiredPhase';
import { kubeCli } from '../../../../../../modules/cli';

let project: Project;
let workloadId: string;
let workloadName: string;

const workloadSpec = {
    image: 'ubuntu',
    command: 'sleep',
    args: '500'
};

const workloadSpecInference = {
    image: config.quickStartImage.inferenceHelloWorld,
    servingPort: {
        container: 8080
    },
    autoscaling: {
        minReplicas: 2,
        maxReplicas: 3,
        metric: 'throughput',
        metricThreshold: 100
    }
};

const WorkloadTypeParams = [
    {
        workloadType: 'training',
        apiWorkloadType: WorkloadType.TRAINING,
        workloadName: 'test-training',
        getFunction: workloadsCrudUtils.getTraining,
        getOtherWorkloadTypeFunction: workloadsCrudUtils.getWorkspace,
        createFunction: workloadsCrudUtils.createTraining,
        suspendFunction: workloadsCrudUtils.suspendTraining,
        resumeFunction: workloadsCrudUtils.resumeTraining,
        deleteFunction: workloadsCrudUtils.deleteTraining,
        jobKind: crds.runaiTrainingWorkloadCrd.kind,
        workloadSpec: workloadSpec
    },
    {
        workloadType: 'interactive',
        apiWorkloadType: WorkloadType.WORKSPACE,
        workloadName: 'test-interactive',
        getFunction: workloadsCrudUtils.getWorkspace,
        getOtherWorkloadTypeFunction: workloadsCrudUtils.getTraining,
        createFunction: workloadsCrudUtils.createWorkspace,
        suspendFunction: workloadsCrudUtils.suspendWorkspace,
        resumeFunction: workloadsCrudUtils.resumeWorkspace,
        deleteFunction: workloadsCrudUtils.deleteWorkspace,
        jobKind: crds.runaiInteractiveWorkloadCrd.kind,
        workloadSpec: workloadSpec
    },
    {
        workloadType: 'inference',
        apiWorkloadType: WorkloadType.INFERENCE,
        workloadName: 'test-inference',
        getFunction: workloadsCrudUtils.getInference,
        getOtherWorkloadTypeFunction: workloadsCrudUtils.getWorkspace,
        createFunction: workloadsCrudUtils.createInference,
        deleteFunction: workloadsCrudUtils.deleteInference,
        jobKind: crds.runaiInferenceWorkloadCrd.kind,
        workloadSpec: workloadSpecInference
    }
];

WorkloadTypeParams.forEach(function (params) {
    let teamTag = tags.teams.viper;
    if (params.workloadType === 'inference') {
        teamTag = tags.teams.viper;
    }
    test.describe(`${params.workloadType} basic API, tags: ${tags.runOn.k8sFlavCertKit}${tags.operations.submit}  ${teamTag}`, function () {
        test.beforeAll('creating project', async function () {
            if (params.workloadType === 'inference') {
                if (!(await isKnativeInstalled())) {
                    logger.warn('knative is not installed on the cluster, skipping test');
                    test.skip();
                }
            }

            project = await projectResourcesUtils.createProjectAndNamespace('project');
            workloadsCrudUtils.accessToken = flowRootState.users.system_administrator.accessToken;
        });

        test.afterAll('Cleanup', async function () {
            try {
                await operations.deleteProjectAndNamespace(project.id);
            } catch {
                logger.error(`Failed to delete project ${project.name} and namespace ${project.namespace}`);
            }
        });

        test(`Create ${params.workloadType} workload and validate it can be managed using the APIs`, async function ({}, testInfo) {
            if (params.workloadType === 'inference') {
                skipPlaywrightTest(
                    test,
                    testInfo,
                    'test depends on the legacy InferenceWorkload CR, which the any-workload path no longer creates (it produces a WorkloadRunner instead)'
                );
            }
            const isRedirectedSubmission =
                params.workloadType !== 'inference' &&
                (await isAnyWorkloadSubmissionEnabled(
                    params.workloadType === 'interactive'
                        ? AnyWorkloadSubmissionType.Workspace
                        : AnyWorkloadSubmissionType.Training
                ));
            let kubernetesServerVersion: string | undefined;
            let isAffectedByJobResumeBug = false;
            if (isRedirectedSubmission) {
                kubernetesServerVersion = await kubeCli.getServerVersion();
                isAffectedByJobResumeBug =
                    versionUtils.isKubernetesVersionAffectedByJobResumeBug(kubernetesServerVersion);
            }

            const getWorkloadState = async (): Promise<{
                response: { statusCode?: number };
                desiredPhase: WorkloadDesiredPhase | DesiredPhase;
                actualPhase: Phase | undefined;
            }> => {
                if (!isRedirectedSubmission) {
                    const result = await params.getFunction(workloadId);
                    return {
                        response: result.response,
                        desiredPhase: result.body.desiredPhase,
                        actualPhase: result.body.actualPhase
                    };
                }

                const [v2Result, v1Result] = await Promise.all([
                    workloadsManagerUtils.getWorkloadV2ById(workloadId),
                    params.getFunction(workloadId)
                ]);
                return {
                    response: v2Result.response,
                    desiredPhase: v2Result.body.workload.desiredPhase,
                    // Workloads-manager V2 does not expose the calculated actual phase.
                    actualPhase: v1Result.body.actualPhase
                };
            };

            await test.step(`Create ${params.workloadType} workload`, async function () {
                const result = await params.createFunction(
                    params.workloadName,
                    true /*add suffix*/,
                    flowRootState.cluster.uuid,
                    String(project.id),
                    params.workloadSpec
                );

                expect(result.response.statusCode).to.equal(StatusCodes.ACCEPTED);
                expect(result.body.desiredPhase).to.equal(WorkloadStatus.RUNNING);

                workloadId = result.body.workloadId;
                workloadName = result.body.name;

                let numPods = 1;
                if (params.workloadType === 'inference') {
                    numPods = params.workloadSpec['autoscaling'].minReplicas;
                }
                const pods = await waitUtils.waitForPods(workloadName, project.namespace, numPods, 60, [
                    'Running'
                ]);
                expect(pods.length).to.equal(numPods);
            });

            await test.step(`Get after create ${params.workloadType} workload should show running`, async function () {
                if (isRedirectedSubmission) {
                    await workloadsCrudUtils.waitForWorkloadPhase(
                        workloadId,
                        Phase.Running,
                        params.apiWorkloadType
                    );
                }

                const workloadState = await getWorkloadState();
                expect(workloadState.response.statusCode).to.equal(StatusCodes.OK);
                expect(workloadState.desiredPhase).to.equal(WorkloadStatus.RUNNING);

                // allowing INITIALIZING as in inference it may take the workload more time to reach RUNNING
                const expectedActualPhases = [WorkloadStatus.RUNNING];
                if (params.workloadType === 'inference') {
                    expectedActualPhases.push(WorkloadStatus.INITIALIZING);
                }
                expect(workloadState.actualPhase).to.be.oneOf(expectedActualPhases);

                // try getting the other workload type
                try {
                    await params.getOtherWorkloadTypeFunction(workloadId);
                } catch (error) {
                    expect(error.statusCode).to.equal(StatusCodes.NOT_FOUND); // because it is not a workspace!
                }
            });

            await test.step(`Suspend ${params.workloadType} workload`, async function () {
                if (params.suspendFunction === undefined) {
                    return;
                }
                const suspendResult = await params.suspendFunction(workloadId);
                expect(suspendResult.response.statusCode).to.equal(StatusCodes.ACCEPTED);

                if (isRedirectedSubmission) {
                    await workloadsCrudUtils.waitForWorkloadPhase(
                        workloadId,
                        Phase.Stopped,
                        params.apiWorkloadType
                    );
                } else {
                    await waitUtils.waitForJobToBeSuspend(workloadName, project.namespace, params.jobKind);
                }

                const workloadState = await getWorkloadState();
                expect(workloadState.response.statusCode).to.equal(StatusCodes.OK);
                expect(workloadState.desiredPhase).to.equal(WorkloadStatus.STOPPED);
                expect(workloadState.actualPhase).to.equal(WorkloadStatus.STOPPED);
            });

            await test.step(`Wait for ${params.workloadType} workload pods to disappear`, async function () {
                if (params.suspendFunction === undefined) {
                    return;
                }

                await waitUtils.waitForWorkloadPodNotExist(project.namespace, workloadName);
            });

            await test.step(`Request resume for ${params.workloadType} workload`, async function () {
                if (params.resumeFunction === undefined) {
                    return;
                }

                const resumeResult = await params.resumeFunction(workloadId);
                expect(resumeResult.response.statusCode).to.equal(StatusCodes.ACCEPTED);
            });

            await test.step(`Verify ${params.workloadType} workload resumed`, async function (step) {
                if (params.resumeFunction === undefined) {
                    return;
                }

                step.skip(
                    isRedirectedSubmission && isAffectedByJobResumeBug,
                    `Kubernetes ${kubernetesServerVersion} is affected by the Job resume bug`
                );

                if (isRedirectedSubmission) {
                    // Any-workload submissions materialize as batch/v1 Jobs, not Interactive/Training CRDs.
                    await workloadsCrudUtils.waitForWorkloadPhase(
                        workloadId,
                        Phase.Running,
                        params.apiWorkloadType
                    );
                } else {
                    await waitUtils.waitForJobToBeRunning(workloadName, project.namespace, params.jobKind);
                }

                const pods = await waitUtils.waitForPods(workloadName, project.namespace, 1, 60, ['Running']);
                expect(pods.length).to.equal(1);

                const workloadState = await getWorkloadState();
                expect(workloadState.response.statusCode).to.equal(StatusCodes.OK);
                expect(workloadState.desiredPhase).to.equal(WorkloadStatus.RUNNING);
                expect(workloadState.actualPhase).to.equal(WorkloadStatus.RUNNING);
            });

            await test.step(`Delete ${params.workloadType} workload`, async function () {
                const deleteResult = await params.deleteFunction(workloadId);
                expect(deleteResult.response.statusCode).to.equal(StatusCodes.ACCEPTED);

                await waitUtils.waitForWorkloadToBeDeleted(workloadName);

                if (isRedirectedSubmission) {
                    const getResult = await workloadsManagerUtils.getWorkloadV2ById(workloadId);
                    expect(getResult.response.statusCode).to.equal(StatusCodes.OK);
                    expect(getResult.body.workload.desiredPhase).to.equal(DesiredPhase.Deleted);
                } else {
                    try {
                        await params.getFunction(workloadId);
                    } catch (error) {
                        expect(error.statusCode).to.equal(StatusCodes.NOT_FOUND);
                    }
                }
            });
        });
    });
});
