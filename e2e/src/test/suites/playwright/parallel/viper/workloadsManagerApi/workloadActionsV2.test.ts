/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test } from '@playwright/test';
import { expect } from 'chai';
import { StatusCodes } from 'http-status-codes';
import * as yaml from 'js-yaml';
import {
    flowRootState,
    logger,
    objectGenerator,
    orgUnitUtils,
    tags,
    waitUtils,
    workloadActionsUtils,
    workloadsManagerUtils
} from '../../../../../../modules/helpers';
import { workloadsManagerService } from '../../../../../../modules/api';
import { Project } from '../../../../../../modules/api/apis/orgUnitApi';
import { RunState } from '../../../../../../modules/api/apis/workloadsManagerApi/generated/model/runState';
import { Phase } from '../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import { WorkloadActionType } from '../../../../../../modules/api/apis/workloadsApi/generated/model/workloadActionType';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';

test.describe(`Direct v2 workload action history ${tags.resource.workloadsV2} ${tags.tool.api} ${tags.teams.viper}`, function () {
    supportedFrom(SupportedVersions.V3_1)(function () {
        let project: Project;

        test.beforeAll('create project', async function () {
            const systemAdministrator = flowRootState.users.system_administrator;
            project = await orgUnitUtils.createProjectAndValidate({
                name: objectGenerator.generateShortNameByPrefix('wl-actions-v2'),
                clusterId: flowRootState.cluster.uuid,
                resources: [{ gpu: { deserved: 0, limit: 0 } }]
            });
            workloadsManagerUtils.accessToken = systemAdministrator.accessToken;
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

        test('records direct YAML Job actions', async function () {
            test.setTimeout(300 * 1000);
            const acceptedAfter = new Date();
            const workloadName = objectGenerator.generateShortNameByPrefix('yaml-action');
            const manifest = {
                apiVersion: 'batch/v1',
                kind: 'Job',
                metadata: { name: workloadName },
                spec: {
                    suspend: false,
                    backoffLimit: 0,
                    template: {
                        spec: {
                            restartPolicy: 'Never',
                            containers: [
                                {
                                    name: 'job',
                                    image: 'busybox',
                                    command: ['sh', '-c', 'sleep 3600']
                                }
                            ]
                        }
                    }
                }
            };
            const manifestBase64Encoded = Buffer.from(yaml.dump(manifest)).toString('base64');
            const expectedActions: WorkloadActionType[] = [];

            const workloadId = await test.step('Create YAML Job', async function () {
                const result = await workloadsManagerUtils.createWorkloadV2(
                    workloadName,
                    String(project.id),
                    undefined,
                    undefined,
                    manifestBase64Encoded
                );
                expect(result.response.statusCode).to.equal(StatusCodes.ACCEPTED);
                return result.body.metadata.id;
            });

            await test.step('Verify create action history', async function () {
                expectedActions.push(WorkloadActionType.Create);
                await workloadActionsUtils.waitForActions(
                    workloadId,
                    workloadName,
                    expectedActions,
                    acceptedAfter
                );
            });

            await test.step('Wait for YAML Job to reach Running', async function () {
                await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 120);
            });

            await test.step('Update YAML Job', async function () {
                const result = await workloadsManagerUtils.updateWorkloadV2Metadata(workloadId, manifest, {
                    category: 'Train'
                });
                expect(result.response.statusCode).to.equal(StatusCodes.ACCEPTED);
            });

            await test.step('Verify update action history', async function () {
                expectedActions.push(WorkloadActionType.Update);
                await workloadActionsUtils.waitForActions(
                    workloadId,
                    workloadName,
                    expectedActions,
                    acceptedAfter
                );
            });

            await test.step('Wait for updated YAML Job to reach Running', async function () {
                await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 120);
            });

            await test.step('Suspend YAML Job', async function () {
                const result = await workloadsManagerService.workloadsV2Api.setWorkloadRunState(workloadId, {
                    value: RunState.Suspend
                });
                expect(result.response.statusCode).to.equal(StatusCodes.ACCEPTED);
            });

            await test.step('Verify suspend action history', async function () {
                expectedActions.push(WorkloadActionType.Suspend);
                await workloadActionsUtils.waitForActions(
                    workloadId,
                    workloadName,
                    expectedActions,
                    acceptedAfter
                );
            });

            await test.step('Wait for YAML Job to reach Stopped', async function () {
                await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Stopped, 120);
            });

            await test.step('Resume YAML Job', async function () {
                const result = await workloadsManagerService.workloadsV2Api.setWorkloadRunState(workloadId, {
                    value: RunState.Resume
                });
                expect(result.response.statusCode).to.equal(StatusCodes.ACCEPTED);
            });

            await test.step('Verify resume action history', async function () {
                expectedActions.push(WorkloadActionType.Resume);
                await workloadActionsUtils.waitForActions(
                    workloadId,
                    workloadName,
                    expectedActions,
                    acceptedAfter
                );
            });

            await test.step('Wait for resumed YAML Job to reach Running', async function () {
                await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 120);
            });

            await test.step('Delete YAML Job', async function () {
                const result = await workloadsManagerUtils.deleteWorkloadV2(workloadId);
                expect(result.response.statusCode).to.equal(StatusCodes.ACCEPTED);
            });

            await test.step('Verify delete action history', async function () {
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
