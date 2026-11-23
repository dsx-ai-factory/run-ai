/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import {
    flowRootState,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    projectUtils,
    settings,
    tags,
    trainingUtils,
    Utils,
    workloadsUtilsV2,
    workloadUtils
} from '../../../../../../modules/helpers';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { randomUUID } from 'crypto';
import { test } from '@playwright/test';

test.describe(`Bulk Suspend & Resume API ${tags.teams.viper} ${tags.tool.api}`, function () {
    const phaseTimeout = settings.isSelfHosted ? 90 : 180;
    let projectName: string;
    let namespace: string;
    let project: Project;

    test.beforeAll(async function () {
        projectName = objectGenerator.generateShortNameByPrefix('project-e2e-workloads-bulk-deletion');
        namespace = Utils.namespaceFromProjectName(projectName);
        project = await projectResourcesUtils.createProject(namespace, {
            name: projectName,
            deservedGpus: 0,
            maxAllowedGpus: 0
        });

        const roleBindingName = projectUtils.getSystemAdministratorRoleBindingName();

        await projectUtils.waitForRoleBindingForUserInProject(
            namespace,
            flowRootState.users.system_administrator.email,
            roleBindingName,
            true
        );
    });

    test.afterAll('deleting project and namespace', async function () {
        try {
            await operations.deleteProjectAndNamespace(project.id);
            logger.info(`deleted project ${projectName} an namespace ${namespace}`);
        } catch (error) {
            logger.error(`failed to delete project ${projectName} an namespace ${namespace}`);
            throw error;
        }
    });

    test('Bulk suspend and resume workloads', async function () {
        const workloads = await test.step('Create 3 workloads', async () => {
            const created = await trainingUtils.createTrainingWorkloads(3, project);
            expect(created).to.have.lengthOf(3);
            return created;
        });

        const workloadsIds = workloads.map((workload) => workload.meta.workloadId);
        const idNonExistWorkload = randomUUID();

        await test.step('Wait for workloads to reach "Running" state', async () => {
            await workloadsUtilsV2.waitForWorkloadsPhase(workloadsIds, 'Running', phaseTimeout);
        });

        await test.step('Suspend all workloads including a non-existent one', async () => {
            const res = await workloadUtils.suspendMultipleWorkloads([...workloadsIds, idNonExistWorkload]);
            expect(res.failed).to.have.lengthOf(1);
            expect(res.failed[0].id).to.equal(idNonExistWorkload);
            expect(res.failed[0].code).to.equal(404);
            expect(res.succeeded).to.have.lengthOf(3);
            for (const id of workloadsIds) {
                expect(res.succeeded.includes(id)).to.be.true;
            }
        });

        await test.step('Wait for workloads to reach "Stopped" state', async () => {
            await workloadsUtilsV2.waitForWorkloadsPhase(workloadsIds, 'Stopped', phaseTimeout);
        });

        await test.step('Resume all workloads including the same non-existent one', async () => {
            const res = await workloadUtils.resumeMultipleWorkloads([...workloadsIds, idNonExistWorkload]);
            expect(res.failed).to.have.lengthOf(1);
            expect(res.failed[0].id).to.equal(idNonExistWorkload);
            expect(res.failed[0].code).to.equal(404);
            expect(res.succeeded).to.have.lengthOf(3);
            for (const id of workloadsIds) {
                expect(res.succeeded.includes(id)).to.be.true;
            }
        });

        await test.step('Wait for workloads to reach "Running" state', async () => {
            await workloadsUtilsV2.waitForWorkloadsPhase(workloadsIds, 'Running', phaseTimeout);
        });
    });
});
