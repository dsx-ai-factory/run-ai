/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import {
    authorizationUtils,
    authUtils,
    config,
    flowRootState,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    projectUtils,
    serviceAccountUtils,
    tags,
    Utils,
    workloadsCrudUtils
} from '../../../../../../modules/helpers';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { ServiceAccountPostResponse } from '../../../../../../modules/api/apis/identityManagerApi/generated/api';
import { IncomingMessage } from 'http';
import { ScopeType } from '../../../../../../modules/api/apis/authorizationApi/generated/api';
import { test } from '@playwright/test';

test.describe(
    `Service Account Submit Job, tags: ${tags.teams.phoenix} ${tags.suites.legacySphinx} ${tags.tool.api} ${tags.runOn.parallelCI}`,
    { tag: [tags.services.identityManager] },
    function () {
        let serviceAccountName: string;
        let projectName: string;
        let namespace: string;
        let projectId: number;
        let project: Project;
        let serviceAccount: { response: IncomingMessage; body: ServiceAccountPostResponse };

        test.beforeAll(function () {
            serviceAccountName = objectGenerator.generateShortNameByPrefix('serviceaccount');
            projectName = objectGenerator.generateNameByPrefix('saproject');
            namespace = Utils.namespaceFromProjectName(projectName);
        });

        test.afterAll(async function () {
            try {
                await operations.deleteProjectAndNamespace(
                    projectId,
                    flowRootState.users.system_administrator.accessToken
                );
                logger.debug(`project ${projectName} and ns deleted`);
            } catch (error) {
                logger.error(`Failed to delete project ${projectName}`, error);
            }
            try {
                await serviceAccountUtils.deleteServiceAccount(
                    flowRootState.users.system_administrator.accessToken,
                    serviceAccount.body.id
                );
                logger.debug(`service account ${serviceAccountName} deleted`);
            } catch (error) {
                logger.error(`Failed to delete service account ${serviceAccountName}`, error);
            }
        });

        test('should submit a job using a service account token', async function () {
            serviceAccount = await serviceAccountUtils.createServiceAccount(
                flowRootState.users.system_administrator.accessToken,
                { name: serviceAccountName }
            );
            const scopeID = flowRootState.cluster.tenantId;
            await authorizationUtils.createAccessRuleForServiceAccount(
                serviceAccount.body.clientId,
                flowRootState.users.l1_researcher.role.name,
                ScopeType.Tenant,
                scopeID.toString()
            );
            await authorizationUtils.createAccessRuleForServiceAccount(
                serviceAccount.body.clientId,
                flowRootState.users.editor.role.name,
                ScopeType.Tenant,
                scopeID.toString()
            );
            const serviceAccountToken = await authUtils.authenticateAppAndGetToken(
                serviceAccount.body.clientId,
                serviceAccount.body.clientSecret
            );
            logger.debug(
                `service account created: ${serviceAccount.body.name} with clientId: ${serviceAccount.body.clientId} and token: ${serviceAccountToken}`
            );
            expect(serviceAccountToken, 'service account authentication should return a token object').not.null;
            project = await projectResourcesUtils.createProject(
                namespace,
                {
                    name: projectName,
                    maxAllowedGpus: 2,
                    deservedGpus: 2,
                    permissions: { applications: [serviceAccount.body.name], groups: [] }
                },
                serviceAccountToken.accessToken
            );
            logger.debug(`project created: ${project.name} with id: ${project.id}`);
            logger.debug(` version: ${flowRootState.cluster.version}`);

            const runaiResearcherRoleBindingName = projectUtils.getRunaiResearcherRoleBindingName();
            await projectUtils.waitForRoleBindingForUserInProject(
                namespace,
                serviceAccount.body.name,
                runaiResearcherRoleBindingName,
                true,
                120
            );

            projectId = project.id;
            const submitRes = await workloadsCrudUtils.createTraining(
                serviceAccountName,
                false,
                project.clusterUuid,
                project.id.toString(),
                {
                    image: config.quickStartImage.training,
                    command: 'sleep infinity'
                },
                serviceAccountToken.accessToken
            );
            const job = submitRes.body;
            expect(job.name, 'submitted job name should match the service account name').equal(serviceAccountName);
            expect(job.projectId, 'submitted job projectId should match the created project id').equal(project.id.toString());
            expect(job.spec.image, 'submitted job image should match the configured quick start image').equal(config.quickStartImage.training);
        });
    }
);
