/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import {
    objectGenerator,
    tags,
    waitUtils,
    config,
    operations,
    Utils,
    logger,
    projectResourcesUtils,
    versionUtils,
    flowRootState
} from '../../../../../../modules/helpers';
import { kubeCli } from '../../../../../../modules/cli';
import { InteractiveWorkload } from '../../../../../../modules/helpers';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { test } from '@playwright/test';
import { shouldRunNativeWorkloadCrdTests } from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';

shouldRunNativeWorkloadCrdTests(() => {
test.describe(`Create multiple node ports: ${tags.tool.api} ${tags.teams.viper} ${tags.runOn.parallelCI}`, function () {
    let projectName: string, namespace: string, project: Project;
    let jobName1: string, jobName2: string;
    let ports: object;
    let workload1, workload2;

    test.beforeAll(async function () {
        projectName = objectGenerator.generateNameByPrefix('p');
        namespace = Utils.namespaceFromProjectName(projectName);
        project = await projectResourcesUtils.createProject(namespace, {
            name: projectName,
            deservedGpus: 2,
            maxAllowedGpus: 2
        });
        jobName1 = objectGenerator.generateNameByPrefix('job');
        jobName2 = objectGenerator.generateNameByPrefix('job');
        ports = {
            port1: {
                container: 8080,
                external: 30568
            }
        };

        workload1 = new InteractiveWorkload(config.quickStartImage.interactive, namespace, jobName1);
        workload1.serviceType = 'NodePort';
        workload1.ports = ports;

        workload2 = new InteractiveWorkload(config.quickStartImage.interactive, namespace, jobName2);
        workload2.serviceType = 'NodePort';
        workload2.ports = ports;
    });

    test.afterAll(async function () {
        try {
            await operations.deleteProjectAndNamespace(project.id);
        } catch {
            logger.error(`Failed to delete project ${projectName} and namespace ${namespace}`);
        }
    });

    test('Should reject the workload which attempts to reuse the port', async function () {
        await kubeCli.createWorkload(workload1);
        if (!versionUtils.isVersionSupported(flowRootState.cluster.version, '2.16')) {
            await waitUtils.waitForService(jobName1, namespace);
        } else {
            const serviceName = `iw-${jobName1}-0-nodeport`;
            await waitUtils.waitForService(serviceName, namespace);
        }

        try {
            await kubeCli.createWorkload(workload2);
        } catch (e) {
            expect(e.message).to.contain('provided port is already allocated');
        }
    });
});
});
