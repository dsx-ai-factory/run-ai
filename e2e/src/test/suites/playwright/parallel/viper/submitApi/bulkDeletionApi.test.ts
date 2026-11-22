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
    tags,
    trainingUtils,
    Utils,
    workloadUtils
} from '../../../../../../modules/helpers';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { randomUUID } from 'crypto';
import { test } from '@playwright/test';

test.describe(
    `Bulk Deletion API ${tags.teams.viper} ${tags.tool.api}`,
    { tag: [tags.services.workloads, tags.runOn.parallelCI] },
    function () {
        let projectName: string;
        let namespace: string;
        let project: Project;

        test.beforeAll(async function () {
            projectName = objectGenerator.generateShortNameByPrefix('project-e2e-workloads-bulk-deletion');
            namespace = Utils.namespaceFromProjectName(projectName);
            project = await projectResourcesUtils.createProject(namespace, {
                name: projectName,
                deservedGpus: 4,
                maxAllowedGpus: 4
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

        test('should succeed to delete multiple workloads in a single request', async function () {
            const workloads = await trainingUtils.createTrainingWorkloads(3, project);
            const workloadsIds = workloads.map((workload) => workload.meta.workloadId);
            workloadsIds.push(randomUUID());

            const res = await workloadUtils.deleteMultipleWorkloads(workloadsIds);

            expect(res.succeeded).to.have.lengthOf(3);
            expect(res.succeeded.includes(workloadsIds[0])).to.be.true;
            expect(res.succeeded.includes(workloadsIds[1])).to.be.true;
            expect(res.succeeded.includes(workloadsIds[2])).to.be.true;
        });

        test('should succeed to delete multiple workloads in a single request, and report on failure of deletion', async function () {
            const workloads = await trainingUtils.createTrainingWorkloads(2, project);
            const workloadsIds = workloads.map((workload) => workload.meta.workloadId);
            workloadsIds.push(randomUUID());

            const res = await workloadUtils.deleteMultipleWorkloads(workloadsIds);

            expect(res.succeeded).to.have.lengthOf(2);
            expect(res.succeeded.includes(workloadsIds[0])).to.be.true;
            expect(res.succeeded.includes(workloadsIds[1])).to.be.true;
            expect(res.succeeded.includes(workloadsIds[2])).to.be.false;
            expect(res.failed).to.have.lengthOf(1);
            expect(res.failed[0].id).to.equal(workloadsIds[2]);
            expect(res.failed[0].code).to.equal(404);
        });
    }
);
