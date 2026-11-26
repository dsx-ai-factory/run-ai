/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'playwright/test';
import {
    crds,
    flowRootState,
    kubeConfigUtils,
    logger,
    objectGenerator,
    operations,
    tags,
    waitUtils
} from '../../../../../../../../../modules/helpers';
import { kubeCli } from '../../../../../../../../../modules/cli';
import { externalWorkloadUtils, Utils } from '../../../../../../../../../modules/helpers';
import { RayClusterWorkload } from '../../../../../../../../../modules/helpers/models/externalWorkload';
import { Project } from '../../../../../../../../../modules/api/apis/projectsDepartmentsApi';
import { projectResourcesUtils } from '../../../../../../../../../modules/helpers';
import { Phase } from '../../../../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import { installRayOperator } from '../../../../../../../../../modules/helpers/utils/IsInstalledUtils';

import { test } from '@playwright/test';

test.describe(
    `Ray job cluster submission rbac test  ${tags.teams.viper}`,
    { tag: [tags.services.externalWorkloadIntegrator] },
    function () {
        let project: Project;
        let workloadName: string;
        let workloadNamespace: string;
        let workloadId: string;
        let initialKubeConfig: string;
        let testConfig: string;

        async function configureL1ResearcherToken(): Promise<void> {
            testConfig = kubeConfigUtils.setUserToken(
                testConfig,
                flowRootState.users.l1_researcher.accessToken,
                flowRootState.users.l1_researcher.refreshToken
            );
            await kubeConfigUtils.setKubeConfig(testConfig);
        }

        async function setTestConfigContext(): Promise<void> {
            initialKubeConfig = await kubeConfigUtils.getKubeConfig();
            testConfig = initialKubeConfig;
            const contextName = objectGenerator.generateNameByPrefix('ray-cluster');
            testConfig = kubeConfigUtils.addUser(testConfig);
            testConfig = kubeConfigUtils.addContext(testConfig, contextName);
            testConfig = kubeConfigUtils.setContext(testConfig, contextName);
        }

        test.beforeAll(async function () {
            await setTestConfigContext();
            await installRayOperator();

            project = await projectResourcesUtils.createProjectAndNamespace('cli-ray', undefined, 5, 5);
            workloadNamespace = Utils.namespaceFromProjectName(project.name);
        });

        test.afterAll(async function () {
            if (project) {
                await operations.deleteProjectAndNamespace(project.id);
            }
            if (initialKubeConfig) {
                await kubeConfigUtils.setKubeConfig(initialKubeConfig);
            }
        });

        test('should submit a Ray job using kubectl apply and validate it is running', async function () {
            await configureL1ResearcherToken();

            workloadName = objectGenerator.generateNameByPrefix('ray-job');

            const rayCluster = new RayClusterWorkload(
                workloadName,
                workloadNamespace,
                undefined,
                'rayproject/ray:2.9.3',
                '2.9.3',
                '500m',
                '500m',
                '2Gi',
                '2Gi',
                '1',
                '1',
                4,
                1,
                2
            );

            workloadId = await externalWorkloadUtils.submitWorkloadAndGetWorkloadId(rayCluster, project.id);
            expect(workloadId, 'Ray workload ID should be returned').toBeTruthy();

            const workload = await waitUtils.waitForBackendWorkloadPhase(
                workloadName,
                project.id,
                Phase.Running,
                300
            );
            expect(workload).toBeTruthy();
            expect(workload.name).toEqual(workloadName);
            expect(workload.phase === Phase.Running).toBeTruthy();

            try {
                await kubeCli.deleteRayCluster(workload.name, workload.namespace);
                await kubeCli.waitForCrdToBeDeleted(crds.rayClusterCrd, workload.name, workload.namespace);
            } catch (error) {
                logger.error(`Failed to delete Ray cluster: ${error}`);
            }
        });
    }
);
