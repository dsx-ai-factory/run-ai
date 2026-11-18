/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test } from '@playwright/test';
import { expect } from 'chai';
import { StatusCodes } from 'http-status-codes';
import {
    config,
    flowRootState,
    logger,
    objectGenerator,
    orgUnitUtils,
    tags,
    waitUtils,
    workloadActionsUtils,
    workloadsCrudUtils
} from '../../../../../../../modules/helpers';
import { Project } from '../../../../../../../modules/api/apis/orgUnitApi';
import { Phase } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import { WorkloadActionType } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/workloadActionType';
import {
    isKnativeInstalled,
    isKubeflowInstalled,
    isLwsInstalled
} from '../../../../../../../modules/helpers/utils/IsInstalledUtils';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { WorkloadType } from '../../../../../../../modules/helpers/utils/WorkloadsCrudUtils';

const basicSpec = {
    image: 'ubuntu',
    command: 'sleep',
    args: '500'
};

const inferenceSpec = {
    image: config.quickStartImage.inferenceHelloWorld,
    servingPort: {
        container: 8080
    },
    autoscaling: {
        minReplicas: 1,
        maxReplicas: 1,
        metric: 'throughput',
        metricThreshold: 100
    }
};

const distributedSpec = {
    image: 'ubuntu',
    command: 'sleep infinity',
    numWorkers: 1,
    distributedFramework: 'PyTorch'
};

const distributedInferenceSpec = {
    workers: 1,
    replicas: 1,
    servingPort: {
        port: 8080
    },
    leader: {
        image: config.quickStartImage.inferenceHelloWorld
    },
    worker: {
        image: config.quickStartImage.inferenceHelloWorld
    }
};

const WorkloadTypeParams = [
    {
        workloadType: WorkloadType.WORKSPACE,
        workloadName: 'action-workspace',
        spec: basicSpec,
        create: workloadsCrudUtils.createWorkspace.bind(workloadsCrudUtils),
        suspend: workloadsCrudUtils.suspendWorkspace.bind(workloadsCrudUtils),
        resume: workloadsCrudUtils.resumeWorkspace.bind(workloadsCrudUtils),
        delete: workloadsCrudUtils.deleteWorkspace.bind(workloadsCrudUtils)
    },
    {
        workloadType: WorkloadType.TRAINING,
        workloadName: 'action-training',
        spec: basicSpec,
        create: workloadsCrudUtils.createTraining.bind(workloadsCrudUtils),
        suspend: workloadsCrudUtils.suspendTraining.bind(workloadsCrudUtils),
        resume: workloadsCrudUtils.resumeTraining.bind(workloadsCrudUtils),
        delete: workloadsCrudUtils.deleteTraining.bind(workloadsCrudUtils)
    },
    {
        workloadType: WorkloadType.DISTRIBUTED,
        workloadName: 'action-distributed',
        spec: distributedSpec,
        create: workloadsCrudUtils.createDistributed.bind(workloadsCrudUtils),
        suspend: workloadsCrudUtils.suspendDistributedTraining.bind(workloadsCrudUtils),
        resume: workloadsCrudUtils.resumeDistributedTraining.bind(workloadsCrudUtils),
        delete: workloadsCrudUtils.deleteDistributed.bind(workloadsCrudUtils)
    },
    {
        workloadType: WorkloadType.INFERENCE,
        workloadName: 'action-inference',
        spec: inferenceSpec,
        create: workloadsCrudUtils.createInference.bind(workloadsCrudUtils),
        update: (workloadId: string): ReturnType<typeof workloadsCrudUtils.updateInferenceSpec> =>
            workloadsCrudUtils.updateInferenceSpec(workloadId, { category: 'Deploy' }),
        delete: workloadsCrudUtils.deleteInference.bind(workloadsCrudUtils)
    },
    {
        workloadType: WorkloadType.DISTRIBUTED_INFERENCE,
        workloadName: 'action-distributed-inference',
        spec: distributedInferenceSpec,
        create: workloadsCrudUtils.createDistributedInference.bind(workloadsCrudUtils),
        update: (workloadId: string): ReturnType<typeof workloadsCrudUtils.updateDistributedInferenceSpec> =>
            workloadsCrudUtils.updateDistributedInferenceSpec(workloadId, { spec: { replicas: 2 } }),
        delete: workloadsCrudUtils.deleteDistributedInference.bind(workloadsCrudUtils)
    }
] as const;

const workloadTypesWithUpdate = [WorkloadType.INFERENCE, WorkloadType.DISTRIBUTED_INFERENCE] as const;
const workloadTypesWithSuspendResume = [
    WorkloadType.WORKSPACE,
    WorkloadType.TRAINING,
    WorkloadType.DISTRIBUTED
] as const;
const workloadTypesRequiringKnative = [WorkloadType.INFERENCE] as const;
const workloadTypesRequiringLws = [WorkloadType.DISTRIBUTED_INFERENCE] as const;
const workloadTypesRequiringKubeflow = [WorkloadType.DISTRIBUTED] as const;
const workloadTypesUsingBackendPhaseWait = [WorkloadType.DISTRIBUTED] as const;

WorkloadTypeParams.forEach(function (params) {
    test.describe(`${params.workloadType} v1 action history ${tags.resource.workloads} ${tags.tool.api} ${tags.teams.viper}`, function () {
        supportedFrom(SupportedVersions.V3_1)(function () {
            let project: Project;

            test.beforeAll('create project', async function () {
                const systemAdministrator = flowRootState.users.system_administrator;
                if (
                    workloadActionsUtils.isWorkloadTypeIn(params, workloadTypesRequiringKnative) &&
                    !(await isKnativeInstalled())
                ) {
                    logger.warn('knative is not installed on the cluster, skipping test');
                    test.skip();
                }
                if (
                    workloadActionsUtils.isWorkloadTypeIn(params, workloadTypesRequiringLws) &&
                    !(await isLwsInstalled())
                ) {
                    logger.warn('LeaderWorkerSet (LWS) is not installed on the cluster, skipping test');
                    test.skip();
                }
                if (
                    workloadActionsUtils.isWorkloadTypeIn(params, workloadTypesRequiringKubeflow) &&
                    !(await isKubeflowInstalled())
                ) {
                    logger.warn('Kubeflow is not installed on the cluster, skipping test');
                    test.skip();
                }

                project = await orgUnitUtils.createProjectAndValidate({
                    name: objectGenerator.generateShortNameByPrefix('wl-actions'),
                    clusterId: flowRootState.cluster.uuid,
                    resources: [{ gpu: { deserved: 0, limit: 0 } }]
                });
                workloadsCrudUtils.accessToken = systemAdministrator.accessToken;
            });

            test.afterAll('delete project and namespace', async function () {
                if (project === undefined) {
                    return;
                }

                try {
                    await orgUnitUtils.deleteProjectAndNamespace(project.id, project.name);
                } catch (error) {
                    logger.error(`Failed to delete project ${project.name}; error ${error}`);
                }
            });

            test(`records accepted ${params.workloadType} actions`, async function () {
                test.setTimeout(300 * 1000);
                const acceptedAfter = new Date();
                const expectedActions: WorkloadActionType[] = [];

                const { workloadId, workloadName } =
                    await test.step(`Create ${params.workloadType}`, async function () {
                        const result = await params.create(
                            params.workloadName,
                            true,
                            flowRootState.cluster.uuid,
                            String(project.id),
                            params.spec
                        );
                        expect(result.response.statusCode).to.equal(StatusCodes.ACCEPTED);

                        const workloadId = result.body.workloadId;
                        const workloadName = result.body.name;
                        expectedActions.push(WorkloadActionType.Create);
                        await workloadActionsUtils.waitForActions(
                            workloadId,
                            workloadName,
                            expectedActions,
                            acceptedAfter
                        );

                        return { workloadId, workloadName };
                    });

                if (workloadActionsUtils.isWorkloadTypeIn(params, workloadTypesWithUpdate)) {
                    await test.step(`Update ${params.workloadType}`, async function () {
                        const result = await params.update(workloadId);
                        expect(result.response.statusCode).to.equal(StatusCodes.ACCEPTED);
                        expectedActions.push(WorkloadActionType.Update);
                        await workloadActionsUtils.waitForActions(
                            workloadId,
                            workloadName,
                            expectedActions,
                            acceptedAfter
                        );
                    });
                }

                if (workloadActionsUtils.isWorkloadTypeIn(params, workloadTypesWithSuspendResume)) {
                    const waitForPhase = async (phase: Phase): Promise<void> => {
                        if (
                            workloadActionsUtils.isWorkloadTypeIn(params, workloadTypesUsingBackendPhaseWait)
                        ) {
                            await waitUtils.waitForBackendWorkloadPhase(workloadName, project.id, phase, 180);
                            return;
                        }
                        await workloadsCrudUtils.waitForWorkloadPhase(
                            workloadId,
                            phase,
                            params.workloadType,
                            undefined,
                            120
                        );
                    };

                    await test.step(`Wait for ${params.workloadType} to reach Running`, async function () {
                        await waitForPhase(Phase.Running);
                    });

                    await test.step(`Suspend ${params.workloadType}`, async function () {
                        const result = await params.suspend(workloadId);
                        expect(result.response.statusCode).to.equal(StatusCodes.ACCEPTED);
                        expectedActions.push(WorkloadActionType.Suspend);
                        await workloadActionsUtils.waitForActions(
                            workloadId,
                            workloadName,
                            expectedActions,
                            acceptedAfter
                        );
                    });

                    if (!workloadActionsUtils.isWorkloadTypeIn(params, workloadTypesUsingBackendPhaseWait)) {
                        await test.step(`Wait for ${params.workloadType} to reach Stopped`, async function () {
                            await waitForPhase(Phase.Stopped);
                        });
                    }

                    await test.step(`Resume ${params.workloadType}`, async function () {
                        if (
                            workloadActionsUtils.isWorkloadTypeIn(params, workloadTypesUsingBackendPhaseWait)
                        ) {
                            await workloadActionsUtils.waitForAcceptedOperation(
                                () => params.resume(workloadId),
                                `${params.workloadType} workload to accept resume`
                            );
                        } else {
                            const result = await params.resume(workloadId);
                            expect(result.response.statusCode).to.equal(StatusCodes.ACCEPTED);
                        }
                        expectedActions.push(WorkloadActionType.Resume);
                        await workloadActionsUtils.waitForActions(
                            workloadId,
                            workloadName,
                            expectedActions,
                            acceptedAfter
                        );
                    });
                }

                await test.step(`Delete ${params.workloadType}`, async function () {
                    const result = await params.delete(workloadId);
                    expect(result.response.statusCode).to.equal(StatusCodes.ACCEPTED);
                    expectedActions.push(WorkloadActionType.Delete);
                    await workloadActionsUtils.waitForActions(
                        workloadId,
                        workloadName,
                        expectedActions,
                        acceptedAfter
                    );
                });
            });
        });
    });
});
