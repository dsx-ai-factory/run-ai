/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test } from '@playwright/test';
import {
    flowRootState,
    logger,
    objectGenerator,
    orgUnitUtils,
    tags,
    workloadActionsUtils,
    workloadsManagerUtils
} from '../../../../../../../modules/helpers';
import { Project } from '../../../../../../../modules/api/apis/orgUnitApi';
import { WorkloadActionType } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/workloadActionType';
import { isNimOperatorInstalled } from '../../../../../../../modules/helpers/utils/IsInstalledUtils';
import {
    createDefaultNimServiceSpec,
    createNgcAuthSecret,
    createNimService,
    updateNimServiceSpec
} from '../../../../../../../modules/helpers/utils/nimServiceUtils';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';

test.describe(`Workload action history through the NIM API ${tags.resource.inferenceWorkload} ${tags.tool.api} ${tags.teams.viper}`, function () {
    supportedFrom(SupportedVersions.V3_1)(function () {
        let project: Project;
        let workloadId: string | undefined;

        test.beforeAll('create project', async function () {
            test.setTimeout(120 * 1000);
            if (!(await isNimOperatorInstalled())) {
                logger.warn('k8s-nim-operator is not installed on the cluster, skipping test');
                test.skip();
            }

            const systemAdministrator = flowRootState.users.system_administrator;
            project = await orgUnitUtils.createProjectAndValidate({
                name: objectGenerator.generateShortNameByPrefix('wl-actions-nim'),
                clusterId: flowRootState.cluster.uuid,
                resources: [{ gpu: { deserved: 0, limit: 0 } }]
            });
            await createNgcAuthSecret(project.status.namespace);
            workloadsManagerUtils.accessToken = systemAdministrator.accessToken;
        });

        test.afterAll('delete workload and project', async function () {
            if (project === undefined) {
                return;
            }

            if (workloadId !== undefined) {
                try {
                    await workloadsManagerUtils.deleteWorkloadV2(workloadId);
                } catch (error) {
                    logger.error(`Failed to delete NIM workload ${workloadId}; error ${error}`);
                }
            }

            try {
                await orgUnitUtils.deleteProjectAndNamespace(project.id, project.name);
            } catch (error) {
                logger.error(`Failed to delete project ${project.name}; error ${error}`);
            }
        });

        test('records workload actions initiated through NIM service operations', async function () {
            test.setTimeout(300 * 1000);
            const acceptedAfter = new Date();
            const workloadName = objectGenerator.generateShortNameByPrefix('nim-action');
            const expectedActions: WorkloadActionType[] = [];
            const spec = createDefaultNimServiceSpec();
            spec.servingPort.exposeExternally = false;

            // Creating a NIM service creates an underlying workload whose accepted actions are recorded in workload action history.
            const id = await test.step('Create a NIM service through the NIM API', async function () {
                const created = await createNimService(workloadName, project, spec);
                workloadId = created.metadata.id;
                return workloadId;
            });

            await test.step('Verify create action history', async function () {
                expectedActions.push(WorkloadActionType.Create);
                await workloadActionsUtils.waitForActions(id, workloadName, expectedActions, acceptedAfter);
            });

            await test.step('Update NIM service', async function () {
                await updateNimServiceSpec(id, { ...spec, replicas: 2 });
            });

            await test.step('Verify update action history', async function () {
                expectedActions.push(WorkloadActionType.Update);
                await workloadActionsUtils.waitForActions(id, workloadName, expectedActions, acceptedAfter);
            });

            await test.step('Delete NIM service', async function () {
                await workloadsManagerUtils.deleteWorkloadV2(id);
                workloadId = undefined;
            });

            await test.step('Verify delete action history', async function () {
                expectedActions.push(WorkloadActionType.Delete);
                await workloadActionsUtils.waitForActions(id, workloadName, expectedActions, acceptedAfter);
            });
        });
    });
});
